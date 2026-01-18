package hytale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// BackupConfig holds backup configuration settings
type BackupConfig struct {
	Enabled  bool `json:"enabled"`
	Frequency int `json:"frequency"` // in minutes
}

// GetBackupConfigPath returns the path to the shared backup config file
func GetBackupConfigPath() string {
	return filepath.Join(GetSharedConfigDir(), "backup.json")
}

// ReadBackupConfig reads the backup configuration from the shared config file
func ReadBackupConfig() (*BackupConfig, error) {
	configPath := GetBackupConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist, return defaults
		return &BackupConfig{
			Enabled:  DefaultBackupEnabled,
			Frequency: DefaultBackupFrequency,
		}, nil
	}

	var config BackupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse backup config: %w", err)
	}

	return &config, nil
}

// WriteBackupConfig writes the backup configuration to the shared config file
func WriteBackupConfig(config *BackupConfig) error {
	configPath := GetBackupConfigPath()
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup config: %w", err)
	}

	return nil
}
