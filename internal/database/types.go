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
	ID       int64
	Name     string
	Bio      string
	ImageURI string
}

// Album represents an album in the database
type Album struct {
	ID          int64
	Name        string
	ReleaseDate time.Time
	ImageURI    string
	Artists     []Artist
}

// Track represents a track in the database
type Track struct {
	ID         int64
	AlbumID    int64
	Name       string
	Duration   int
	Lyrics     string
	IsExplicit bool
	FilePath   string
	SHA512Sum  string
	Artists    []Artist
	Tags       []Tag
}

// User represents a user in the database
type User struct {
	ID          int64
	Name        string
	Preferences map[string]interface{}
}

// Listen represents a listen event in the database
type Listen struct {
	ID         int64
	UserID     int64
	TrackID    int64
	ListenTime int
	Timestamp  time.Time
}

// Tag represents a tag in the database
type Tag struct {
	ID   int64
	Name string
}

// Playlist represents a playlist in the database
type Playlist struct {
	ID         int64
	UserID     int64
	Name       string
	IsFavorite bool
	Tracks     []Track
	Artists    []Artist
	Albums     []Album
}
