package filesystem

import (
	"log"
	"os"
	"path/filepath"
)

func Initialize() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	dataDir, err := GetDataDir()
	if err != nil {
		return err
	}

	// Maps out the file's paths with their default data
	files := map[string][]byte{
		filepath.Join(dataDir, "data.db"):       []byte(""),
		filepath.Join(configDir, "config.json"): []byte(`{}`),
	}

	// Creates and writes the default data to the file if it does not exist
	for filename, data := range files {
		dir := filepath.Dir(filename)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		if _, err = os.Stat(filename); os.IsNotExist(err) {
			if err = os.WriteFile(filename, data, 0644); err != nil {
				return err
			}
		}
	}

	// Get & write config to update file with new/missing fields & format JSON
	config, err := GetConfigFile()
	if err == nil {
		WriteConfigFile(config)
	}

	// Wipes and sets logs.txt as the log output file
	logsPath := filepath.Join(dataDir, "logs.txt")
	setLogPath(logsPath)

	return nil
}

func setLogPath(filePath string) error {
	logsFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	log.SetOutput(logsFile)
	return nil
}
