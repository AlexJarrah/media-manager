package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"gitlab.com/AlexJarrah/media-manager/internal/players"
)

func FetchMBID(player players.Player) (string, error) {
	// Construct the URL and encode parameters
	baseURL := "https://musicbrainz.org/ws/2/recording"
	query := fmt.Sprintf("recording:\"%s\" AND artist:\"%s\" AND release:\"%s\"",
		player.Title, player.Artists[0], player.Album)
	url := fmt.Sprintf("%s?query=%s&fmt=json", baseURL, url.QueryEscape(query))

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error making GET request: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("Error decoding JSON: %s", err.Error())
	}

	// Extract the recording ID
	recordings, ok := data["recordings"].([]interface{})
	if !ok || len(recordings) == 0 {
		return "", errors.New("No recordings found")
	}
	firstRecording, ok := recordings[0].(map[string]interface{})
	if !ok {
		return "", errors.New("Invalid format of recordings")
	}
	recordingID, ok := firstRecording["id"].(string)
	if !ok {
		return "", errors.New("Invalid format of recording ID")
	}

	// Output the recording ID
	log.Println("MusicBrainz recording ID:", recordingID)

	return recordingID, nil
}
