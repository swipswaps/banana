package main

import (
	"encoding/json"
	"time"

	"enix.io/banana/src/models"
	"github.com/imdario/mergo"
	"k8s.io/klog"
)

// routineCmd : Command implementation for 'daemon'
type routineCmd struct{}

// newRoutineCmd : Creates routine command from command line args
func newRoutineCmd(*launchArgs) (*routineCmd, error) {
	return &routineCmd{}, nil
}

// execute : Start the routine using specified plugin
func (cmd *routineCmd) execute(config *models.Config) error {
	sendMessageToMonitor("routine_start", config, cmd, nil, "")
	klog.Info("starting banana routine")

	state := &State{}
	err := state.loadFromDisk(config)
	if err != nil {
		sendMessageToMonitor("routine_crashed", config, cmd, nil, err.Error())
		return err
	}

	err = cmd.runTasks(state, config)
	if err != nil {
		sendMessageToMonitor("routine_failed", config, cmd, nil, err.Error())
		return err
	}

	err = state.saveToDisk(config)
	if err != nil {
		sendMessageToMonitor("routine_crashed", config, cmd, nil, err.Error())
		return err
	}

	sendMessageToMonitor("routine_done", config, cmd, nil, "")
	return nil
}

// jsonMap : Convert struct to an anonymous map with given JSON keys
func (cmd *routineCmd) jsonMap() (out map[string]interface{}) {
	raw, _ := json.Marshal(cmd)
	json.Unmarshal(raw, &out)
	return
}

func (cmd *routineCmd) runTasks(state *State, config *models.Config) error {
	for name, schedule := range config.ScheduledBackups {
		backupState, exists := state.LastBackups[name]
		fullConfig := schedule
		mergo.Merge(&fullConfig.Config, config)
		if !exists || backupState.Status == "Failed" {
			_ = cmd.doBackup(name, state, &fullConfig)
			// TODO: do something with this error
			// at this point monitor should already knows that something fucked up
			// but I feel bad leaving an error unused
			// if err != nil {
			// 	return err
			// }
		} else {
			timeSinceLastBackup := time.Since(backupState.Time)
			interval := time.Duration(schedule.Interval * float32(time.Hour) * 24)
			if timeSinceLastBackup > interval {
				_ = cmd.doBackup(name, state, &fullConfig)
				// TODO: same as above
				// if err != nil {
				// 	return err
				// }
			}
		}
	}

	return nil
}

func (cmd *routineCmd) doBackup(name string, state *State, config *models.ScheduledBackupConfig) error {
	if state.LastBackups[name] == nil {
		state.LastBackups[name] = &BackupState{}
	}

	state.LastBackups[name].Time = time.Now()
	state.LastBackups[name].Status = "Failed"
	klog.Infof("backing up %s", name)

	typ := "incremental"
	if state.LastBackups[name].Type == "" || state.LastBackups[name].IncrCountSinceLastFull >= config.FullEvery-1 {
		typ = "full"
	}

	backupCmd := &backupCmd{
		Type:       typ,
		Name:       name,
		PluginArgs: config.PluginArgs,
	}
	err := backupCmd.execute(&config.Config)
	if err != nil {
		return err
	}

	state.LastBackups[name].Status = "Success"
	state.LastBackups[name].Type = typ
	if typ == "full" {
		state.LastBackups[name].IncrCountSinceLastFull = 0
	} else {
		state.LastBackups[name].IncrCountSinceLastFull++
	}

	return nil
}
