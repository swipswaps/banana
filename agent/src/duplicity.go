package main

import (
	"fmt"
	"os"
)

// DuplicityBackend : BackupBackend implementation for duplicity
type DuplicityBackend struct{}

// Backup : BackupBackend's Backup call implementation for duplicity
func (d *DuplicityBackend) Backup(config *Config, cmd *BackupCmd) error {
	endpoint := fmt.Sprintf("s3://%s/%s/%s", config.StorageHost, config.BucketName, cmd.Name)
	fmt.Printf("%+v", config)
	fmt.Println(os.Getenv("AWS_ACCESS_KEY_ID"))
	fmt.Printf("+ duplicity --full-if-older-than 1W %s %s\n", cmd.Target, endpoint)
	return Execute("duplicity", "--full-if-older-than", "1W", cmd.Target, endpoint)
}
