package internal

type Config struct {
	Players []string `json:"players"`
	LastFM  LastFM   `json:"lastfm"`
	Discord Discord  `json:"discord"`
}

type LastFM struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type Discord struct {
	ClientID string `json:"client_id"`
}
