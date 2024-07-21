package database

import "gitlab.com/AlexJarrah/media-manager/internal/filesystem"

func Initialize() error {
	db, err := NewDB()
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
