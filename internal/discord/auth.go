package discord

import (
	"github.com/hugolgst/rich-go/client"
	"github.com/hugolgst/rich-go/ipc"

	"gitlab.com/AlexJarrah/media-manager/internal/filesystem"
)

func Login() error {
	config, err := filesystem.GetConfigFile()
	if err != nil {
		return err
	}

	return client.Login(config.Discord.ClientID)
}

func Logout() {
	client.Logout()
	ipc.GetIpcPath()
}
