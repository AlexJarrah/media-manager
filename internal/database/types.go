package database

import (
	"database/sql"
	"time"
)

// DB represents the database connection
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

// Artist represents an artist in the database
type Artist struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	ImageURI string `json:"image_uri"`
}

// Album represents an album in the database
type Album struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ReleaseDate time.Time `json:"release_date"`
	ImageURI    string    `json:"image_uri"`
	Artists     []Artist  `json:"artists"`
}

// Track represents a track in the database
type Track struct {
	ID         int64    `json:"id"`
	AlbumID    int64    `json:"album_id"`
	Name       string   `json:"name"`
	Duration   int      `json:"duration"`
	Lyrics     string   `json:"lyrics"`
	IsExplicit bool     `json:"is_explicit"`
	FilePath   string   `json:"file_path"`
	SHA512Sum  string   `json:"sha512sum"`
	Artists    []Artist `json:"artists"`
	Tags       []Tag    `json:"tags"`
}

// User represents a user in the database
type User struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Preferences map[string]interface{} `json:"preferences"`
}

// Listen represents a listen event in the database
type Listen struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TrackID    int64     `json:"track_id"`
	ListenTime int       `json:"listen_time"`
	Timestamp  time.Time `json:"timestamp"`
}

// Tag represents a tag in the database
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Playlist represents a playlist in the database
type Playlist struct {
	ID         int64    `json:"id"`
	UserID     int64    `json:"user_id"`
	Name       string   `json:"name"`
	IsFavorite bool     `json:"is_favorite"`
	Tracks     []Track  `json:"tracks"`
	Artists    []Artist `json:"artists"`
	Albums     []Album  `json:"albums"`
}
