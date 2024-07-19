package database

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

// TrackFile represents a music file with its metadata
type TrackFile struct {
	FilePath   string
	Name       string
	Artist     string
	Album      string
	Duration   int
	Lyrics     string
	IsExplicit bool
	SHA512Sum  string
}

// LoadTracksFromDirectory scans a directory for music files and loads their metadata
func (db *DB) LoadTracksFromDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".mp3" || ext == ".flac" || ext == ".m4a" || ext == ".wav" {
			track, err := loadTrackFile(path)
			if err != nil {
				return err
			}

			// Check for .lrc file
			lrcPath := strings.TrimSuffix(path, ext) + ".lrc"
			if lrcContent, err := os.ReadFile(lrcPath); err == nil {
				track.Lyrics = string(lrcContent)
			}

			// Add the track to the database
			if err := db.addTrackFromFile(track); err != nil {
				return err
			}
		}

		return nil
	})
}

func loadTrackFile(filePath string) (*TrackFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Calculate SHA512 sum
	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	sha512sum := hex.EncodeToString(hash.Sum(nil))

	// Rewind file for metadata reading
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	// Read metadata
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	duration := 0

	track := &TrackFile{
		FilePath:  filePath,
		Name:      metadata.Title(),
		Artist:    metadata.Artist(),
		Album:     metadata.Album(),
		Duration:  duration,
		SHA512Sum: sha512sum,
	}

	return track, nil
}

func (db *DB) addTrackFromFile(track *TrackFile) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if artist exists, if not add it
	var artistID int64
	err = tx.QueryRow("SELECT artist_id FROM artists WHERE name = ?", track.Artist).Scan(&artistID)
	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO artists (name) VALUES (?)", track.Artist)
		if err != nil {
			return err
		}
		artistID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Check if album exists, if not add it
	var albumID int64
	err = tx.QueryRow("SELECT album_id FROM albums WHERE name = ?", track.Album).Scan(&albumID)
	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO albums (name, release_date) VALUES (?, ?)", track.Album, time.Now())
		if err != nil {
			return err
		}
		albumID, err = result.LastInsertId()
		if err != nil {
			return err
		}

		// Link album to artist
		_, err = tx.Exec("INSERT INTO album_artists (album_id, artist_id) VALUES (?, ?)", albumID, artistID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Add track
	result, err := tx.Exec("INSERT INTO tracks (album_id, name, duration, lyrics, is_explicit, file_path, sha512sum) VALUES (?, ?, ?, ?, ?, ?, ?)",
		albumID, track.Name, track.Duration, track.Lyrics, track.IsExplicit, track.FilePath, track.SHA512Sum)
	if err != nil {
		return err
	}

	trackID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Link track to artist
	_, err = tx.Exec("INSERT INTO track_artists (track_id, artist_id) VALUES (?, ?)", trackID, artistID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
