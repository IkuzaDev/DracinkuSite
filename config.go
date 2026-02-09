package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig stores application configuration
type AppConfig struct {
	LicenseKey  string `json:"licenseKey"`
	Language    string `json:"language"`
	DownloadDir string `json:"downloadDir"`
}

// ConfigManager manages the application configuration
type ConfigManager struct {
	config  *AppConfig
	mu      sync.RWMutex
	dataDir string
}

func NewConfigManager() *ConfigManager {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".dramawave")
	os.MkdirAll(dataDir, 0755)

	cm := &ConfigManager{
		config: &AppConfig{
			Language: "id", // Default language
		},
		dataDir: dataDir,
	}
	cm.load()
	return cm
}

func (cm *ConfigManager) getConfigPath() string {
	return filepath.Join(cm.dataDir, "config.json")
}

func (cm *ConfigManager) load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, cm.config)
}

func (cm *ConfigManager) save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.getConfigPath(), data, 0644)
}

// GetLicenseKey returns the stored license key
func (cm *ConfigManager) GetLicenseKey() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.LicenseKey
}

// SetLicenseKey saves the license key
func (cm *ConfigManager) SetLicenseKey(key string) error {
	cm.mu.Lock()
	cm.config.LicenseKey = key
	cm.mu.Unlock()

	return cm.save()
}
