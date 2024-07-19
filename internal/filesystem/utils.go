package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gitlab.com/AlexJarrah/media-manager/internal"
)

// Returns the config directory for the OS
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "AppData", "Roaming", internal.APP_ID), nil
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", internal.APP_ID), nil
	case "linux":
		return filepath.Join(homeDir, ".config", internal.APP_ID), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// Returns the data directory for the OS
func GetDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", internal.APP_ID), nil
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", internal.APP_ID), nil
	case "linux":
		return filepath.Join(homeDir, ".local", "share", internal.APP_ID), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func GetConfigFile() (config internal.Config, err error) {
	dir, err := GetConfigDir()
	if err != nil {
		return internal.Config{}, err
	}

	file, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		return internal.Config{}, err
	}

	if err = json.Unmarshal(file, &config); err != nil {
		return internal.Config{}, err
	}

	return config, nil
}

func WriteConfigFile(config internal.Config) (err error) {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	json_, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.json"), json_, 0677)
}
