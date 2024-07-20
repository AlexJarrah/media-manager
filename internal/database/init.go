package database

import (
	"path/filepath"

	"gitlab.com/AlexJarrah/media-manager/internal/filesystem"
)

func Initialize() error {
	dataDir, err := filesystem.GetDataDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "data.db")
	db, err := NewDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err = db.Exec(sqlInit); err != nil {
		return err
	}

	config, err := filesystem.GetConfigFile()
	if err != nil {
		return err
	}

	for _, dir := range config.MediaDirectories {
		if err := db.LoadTracksFromDirectory(dir); err != nil {
			return err
		}
	}

	return err
}
