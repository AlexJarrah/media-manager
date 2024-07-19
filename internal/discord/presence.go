package discord

import (
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"

	"gitlab.com/AlexJarrah/media-manager/internal/players"
)

var elapsed = time.Now()

func UpdatePresence(player *players.Player) error {
	return client.SetActivity(client.Activity{
		Details:    player.Title,
		State:      strings.Join(player.Artists, ", "),
		LargeText:  player.Album,
		LargeImage: "image",
		SmallText:  strings.Join(player.Artists, ", "),
		SmallImage: player.Artists[0],
		Timestamps: &client.Timestamps{Start: &elapsed},
	})
}
