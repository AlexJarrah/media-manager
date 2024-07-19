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

	_, err = db.Exec(sqlInit)
	return err
}
