package lastfm

import (
	"fmt"
	"log"

	"gitlab.com/AlexJarrah/media-manager/internal/players"
)

func Scrobble(player players.Player, apiKey, apiSecret, sessionKey string) error {
	scrobbleParams := map[string]string{
		"method":    "track.scrobble",
		"artist":    player.Artists[0],
		"track":     player.Title,
		"timestamp": fmt.Sprint(player.StartListeningTime.Unix()),
		"album":     player.Album,
		"mbid":      player.MBID,
		"duration":  fmt.Sprint(player.GetTotalPlayTime().Seconds()),
		"api_key":   apiKey,
		"sk":        sessionKey,
	}
	scrobbleParams["api_sig"] = generateAPISignature(scrobbleParams, apiSecret)
	scrobbleParams["format"] = "json"

	scrobbleResponse, err := makePostRequest(scrobbleParams)
	if err != nil {
		return err
	}

	log.Println("Last.fm response:", string(scrobbleResponse))
	return nil
}
