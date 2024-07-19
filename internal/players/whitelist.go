package players

import (
	"log"

	"gitlab.com/AlexJarrah/media-manager/internal/filesystem"
)

func (p *Player) IsWhitelisted() bool {
	config, err := filesystem.GetConfigFile()
	if err != nil {
		log.Println(err)
		return false
	}

	for _, wp := range config.Players {
		if p.Name == wp {
			return true
		}
	}

	return false
}
