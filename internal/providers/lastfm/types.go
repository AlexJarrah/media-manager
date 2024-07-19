package lastfm

// Struct to parse the session key response
type SessionKeyResponse struct {
	Session struct {
		Key string `json:"key"`
	} `json:"session"`
}
