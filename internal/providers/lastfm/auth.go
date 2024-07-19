package lastfm

import (
	"encoding/json"
	"fmt"

	"gitlab.com/AlexJarrah/media-manager/internal"
)

// Get last.fm session key
func Authenticate(config internal.LastFM) (sessionKey string, err error) {
	authParams := map[string]string{
		"method":   "auth.getMobileSession",
		"username": config.Username,
		"password": config.Password,
		"api_key":  config.APIKey,
	}
	authParams["api_sig"] = generateAPISignature(authParams, config.APISecret)
	authParams["format"] = "json"

	response, err := makePostRequest(authParams)
	if err != nil {
		return "", err
	}
	fmt.Println("Response:", string(response))

	// Parse the session key from the response
	var sessionKeyResponse SessionKeyResponse
	err = json.Unmarshal(response, &sessionKeyResponse)
	if err != nil {
		return "", fmt.Errorf("Error parsing session key: %s", err.Error())
	}

	sessionKey = sessionKeyResponse.Session.Key
	return sessionKey, nil
}
