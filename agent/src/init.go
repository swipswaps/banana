package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"enix.io/banana/src/models"
	"enix.io/banana/src/services"
)

// initCmd : Command implementation for 'init'
type initCmd struct {
	Organization string
	Name         string
	Token        string
}

// newInitCmd : Creates init command from command line args
func newInitCmd(args *launchArgs) (*initCmd, error) {
	if len(args.Values) < 4 {
		return nil, fmt.Errorf("usage: %s init <token> <company name> <agent name>", os.Args[0])
	}

	return &initCmd{
		Token:        args.Values[1],
		Organization: args.Values[2],
		Name:         args.Values[3],
	}, nil
}

// execute : Start the init
func (cmd *initCmd) execute(config *models.Config) error {
	config.BucketName = cmd.Name

	err := os.Mkdir("/etc/banana", 00755)
	if err != nil && os.IsPermission(err) {
		return err
	}

	services.Vault.Client.SetToken(cmd.Token)
	out, err := services.Vault.Client.Logical().Write(
		fmt.Sprintf("%s/%s/agents-pki/issue/default", config.Vault.RootPath, cmd.Organization),
		map[string]interface{}{
			"common_name": cmd.Name,
		},
	)
	if err != nil {
		return err
	}

	cert, _ := out.Data["certificate"].(string)
	privkey, _ := out.Data["private_key"].(string)

	schedule := config.ScheduledBackups
	config.ScheduledBackups = nil

	configRaw, _ := json.MarshalIndent(config, "", "  ")
	scheduleRaw, _ := json.MarshalIndent(schedule, "", "  ")
	if schedule == nil {
		scheduleRaw = []byte("{}")
	}

	writeFileWithoutOverwrite(config.CertPath, []byte(cert))
	writeFileWithoutOverwrite(config.PrivKeyPath, []byte(privkey))
	writeFileWithoutOverwrite("/etc/banana/banana.json", configRaw)
	writeFileWithoutOverwrite(config.ScheduleConfigPath, scheduleRaw)

	loadCredentialsToMem(config)
	sendMessageToMonitor("initialized", config, cmd, nil, "")
	return nil
}

// jsonMap : Convert struct to an anonymous map with given JSON keys
func (cmd *initCmd) jsonMap() (out map[string]interface{}) {
	raw, _ := json.Marshal(cmd)
	json.Unmarshal(raw, &out)
	return
}

func writeFileWithoutOverwrite(filename string, data []byte) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		ioutil.WriteFile(filename, data, 00600)
	} else {
		assert(fmt.Errorf("failed to initialize agent: %s: file already exists", filename))
	}
}
