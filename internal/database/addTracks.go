package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/tag"
)

// Modes
const (
	// Only add tracks not already in the database
	AddNewTracks uint8 = iota

	// Update track metadata by title
	UpdateFromTitle

	// Update track metadata by file path
	UpdateFromFilePath

	// Update track metadata by hash
	UpdateFromHash
)

func (db *DB) LoadTracksFromDirectory(dirPath string, mode uint8) error {
	supportedFormats := map[string]bool{
		".mp3": true, ".m4a": true, ".m4b": true, ".m4p": true,
		".alac": true, ".flac": true, ".ogg": true, ".dsf": true,
	}

	existingHashes := make(map[string]struct{})

	// Do not populate existing hashes for updating modes as no track should be ignored in these modes
	if mode == AddNewTracks {
		existingTracks, _ := db.GetTracks("sha256sum", "")
		for _, t := range existingTracks {
			existingHashes[t.SHA256Sum] = struct{}{}
		}
	}

	var (
		tracks  []*Track
		artists = make(map[string]*Artist)
		albums  = make(map[string]*Album)
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	// Determine the number of concurrent goroutines based on the number of CPU cores
	numCPU := runtime.NumCPU()
	sem := make(chan struct{}, numCPU)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			ext := strings.ToLower(filepath.Ext(path))
			if !supportedFormats[ext] {
				return
			}

			track, artist, album, err := processFile(path, existingHashes)
			if err != nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if track != nil {
				tracks = append(tracks, track)
			}
			if artist != nil {
				artists[artist.Name] = artist
			}
			if album != nil {
				albums[album.Name] = album
			}
		}()

		return nil
	})

	wg.Wait()

	if err != nil {
		return err
	}

	// Convert maps to slices
	artistSlice := make([]*Artist, 0, len(artists))
	for _, a := range artists {
		artistSlice = append(artistSlice, a)
	}

	// Add artists to the database and update their IDs
	if err = db.AddArtists(artistSlice); err != nil {
		return err
	}

	// Update album and track relationships with the new artist IDs
	for _, album := range albums {
		for i, artist := range album.Artists {
			if updatedArtist, ok := artists[artist.Name]; ok {
				album.Artists[i] = *updatedArtist
			}
		}
	}

	for _, track := range tracks {
		for i, artist := range track.Artists {
			if updatedArtist, ok := artists[artist.Name]; ok {
				track.Artists[i] = *updatedArtist
			}
		}
	}

	// Convert albums map to slice
	albumSlice := make([]*Album, 0, len(albums))
	for _, a := range albums {
		albumSlice = append(albumSlice, a)
	}

	// Add albums and tracks to the database
	if err = db.AddAlbums(albumSlice); err != nil {
		return err
	}

	if mode == AddNewTracks {
		err = db.AddTracks(tracks)
		return err
	} else {
		for _, t := range tracks {
			keys := []string{
				"name",
				"duration",
				"lyrics",
				"is_explicit",
				"file_path",
				"sha256sum",
				"album_id",
			}

			var key, value string
			switch mode {
			case UpdateFromTitle:
				key = "title"
				value = t.Name
			case UpdateFromFilePath:
				key = "file_path"
				value = t.FilePath
			case UpdateFromHash:
				key = "sha256sum"
				value = t.SHA256Sum
			}

			if err = db.UpdateTrack(t, keys, key, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(path string, existingHashes map[string]struct{}) (*Track, *Artist, *Album, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, nil, nil, err
	}
	sha256sum := hex.EncodeToString(hash.Sum(nil))

	if _, exists := existingHashes[sha256sum]; exists {
		return nil, nil, nil, nil
	}

	if _, err = file.Seek(0, 0); err != nil {
		return nil, nil, nil, err
	}

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, nil, nil, err
	}

	artist := &Artist{Name: metadata.Artist()}
	albumArtist := &Artist{Name: metadata.AlbumArtist()}

	album := &Album{
		Name:        metadata.Album(),
		ReleaseDate: sql.NullTime{Time: time.Date(metadata.Year(), 1, 1, 0, 0, 0, 0, time.UTC), Valid: metadata.Year() != 0},
		Artists:     []Artist{*albumArtist},
	}

	track := &Track{
		Name:      metadata.Title(),
		Duration:  0,
		FilePath:  path,
		SHA256Sum: sha256sum,
		Artists:   []Artist{*artist},
		Album:     *album,
	}

	return track, artist, album, nil
}
