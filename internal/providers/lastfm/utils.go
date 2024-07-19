package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"sort"
)

// Generate API signature
func generateAPISignature(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var str string
	for _, key := range keys {
		str += key + params[key]
	}
	str += secret

	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// Make a POST request to the last.fm API
func makePostRequest(params map[string]string) ([]byte, error) {
	const apiURL = "https://ws.audioscrobbler.com/2.0/"

	data := url.Values{}
	for key, value := range params {
		data.Set(key, value)
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
