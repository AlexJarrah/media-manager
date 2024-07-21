package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// AddArtists adds multiple new artists to the database
func (db *DB) AddArtists(artists []*Artist) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO artists (name, bio, image_uri) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, artist := range artists {
		result, err := stmt.Exec(artist.Name, artist.Bio, artist.ImageURI)
		if err != nil {
			return err
		}
		artist.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetArtists retrieves multiple artists from the database
func (db *DB) GetArtists(key, value string) ([]*Artist, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT artist_id, name, bio, image_uri FROM artists"
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT artist_id, name, bio, image_uri FROM artists WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []*Artist
	for rows.Next() {
		var artist Artist
		err := rows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
		if err != nil {
			return nil, err
		}
		artists = append(artists, &artist)
	}
	return artists, nil
}

// AddAlbums adds multiple new albums to the database
func (db *DB) AddAlbums(albums []*Album) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO albums (name, release_date, image_uri) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, album := range albums {
		result, err := stmt.Exec(album.Name, album.ReleaseDate, album.ImageURI)
		if err != nil {
			return err
		}
		album.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}

		for _, artist := range album.Artists {
			_, err = tx.Exec("INSERT INTO album_artists (album_id, artist_id) VALUES (?, ?)", album.ID, artist.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetAlbums retrieves multiple albums from the database
func (db *DB) GetAlbums(key, value string) ([]*Album, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT album_id, name, release_date, image_uri FROM albums"
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT album_id, name, release_date, image_uri FROM albums WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []*Album
	for rows.Next() {
		var album Album
		err := rows.Scan(&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
		if err != nil {
			return nil, err
		}

		artistRows, err := db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN album_artists aa ON a.artist_id = aa.artist_id WHERE aa.album_id = ?", album.ID)
		if err != nil {
			return nil, err
		}
		defer artistRows.Close()

		for artistRows.Next() {
			var artist Artist
			err := artistRows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
			if err != nil {
				return nil, err
			}
			album.Artists = append(album.Artists, artist)
		}

		albums = append(albums, &album)
	}

	return albums, nil
}

// AddTracks adds multiple new tracks to the database
func (db *DB) AddTracks(tracks []*Track) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tracks (album_id, name, duration, lyrics, is_explicit, file_path, sha256sum) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, track := range tracks {
		result, err := stmt.Exec(track.AlbumID, track.Name, track.Duration, track.Lyrics, track.IsExplicit, track.FilePath, track.SHA256Sum)
		if err != nil {
			return err
		}
		track.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}

		for _, artist := range track.Artists {
			_, err = tx.Exec("INSERT INTO track_artists (track_id, artist_id) VALUES (?, ?)", track.ID, artist.ID)
			if err != nil {
				return err
			}
		}

		for _, tag := range track.Tags {
			_, err = tx.Exec("INSERT INTO track_tags (track_id, tag_id) VALUES (?, ?)", track.ID, tag.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetTracks retrieves multiple tracks from the database
func (db *DB) GetTracks(key, value string) ([]*Track, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha256sum FROM tracks "
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha256sum FROM tracks WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum)
		if err != nil {
			return nil, err
		}

		artistRows, err := db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN track_artists ta ON a.artist_id = ta.artist_id WHERE ta.track_id = ?", track.ID)
		if err != nil {
			return nil, err
		}
		defer artistRows.Close()

		for artistRows.Next() {
			var artist Artist
			err := artistRows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
			if err != nil {
				return nil, err
			}
			track.Artists = append(track.Artists, artist)
		}

		tagRows, err := db.Query("SELECT t.tag_id, t.name FROM tags t JOIN track_tags tt ON t.tag_id = tt.tag_id WHERE tt.track_id = ?", track.ID)
		if err != nil {
			return nil, err
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var tag Tag
			err := tagRows.Scan(&tag.ID, &tag.Name)
			if err != nil {
				return nil, err
			}
			track.Tags = append(track.Tags, tag)
		}

		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// AddUsers adds multiple new users to the database
func (db *DB) AddUsers(users []*User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO users (name, preferences) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		preferencesJSON, err := json.Marshal(user.Preferences)
		if err != nil {
			return err
		}

		result, err := stmt.Exec(user.Name, preferencesJSON)
		if err != nil {
			return err
		}
		user.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetUsers retrieves multiple users from the database
func (db *DB) GetUsers(key, value string) ([]*User, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT user_id, name, preferences FROM users"
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT user_id, name, preferences FROM users WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var preferencesJSON []byte
		err := rows.Scan(&user.ID, &user.Name, &preferencesJSON)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(preferencesJSON, &user.Preferences)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

// AddListens adds multiple new listen events to the database
func (db *DB) AddListens(listens []*Listen) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO listens (user_id, track_id, listen_time, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, listen := range listens {
		result, err := stmt.Exec(listen.UserID, listen.TrackID, listen.ListenTime, listen.Timestamp)
		if err != nil {
			return err
		}
		listen.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetUserListens retrieves all listen events for a user from the database
func (db *DB) GetUserListens(userId int64, key, value string) ([]*Listen, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = fmt.Sprintf("SELECT listen_id, user_id, track_id, listen_time, timestamp FROM listens WHERE user_id = %d ORDER BY timestamp DESC", userId)
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT listen_id, user_id, track_id, listen_time, timestamp FROM listens WHERE user_id = %d AND %s = ? ORDER BY timestamp DESC", userId, key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listens []*Listen
	for rows.Next() {
		var listen Listen
		err := rows.Scan(&listen.ID, &listen.UserID, &listen.TrackID, &listen.ListenTime, &listen.Timestamp)
		if err != nil {
			return nil, err
		}
		listens = append(listens, &listen)
	}

	return listens, nil
}

// AddTags adds multiple new tags to the database
func (db *DB) AddTags(tags []*Tag) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tags (name) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tag := range tags {
		result, err := stmt.Exec(tag.Name)
		if err != nil {
			return err
		}
		tag.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTags retrieves multiple tags from the database
func (db *DB) GetTags(key, value string) ([]*Tag, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT tag_id, name FROM tags"
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT tag_id, name FROM tags WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}

	return tags, nil
}

// AddPlaylists adds multiple new playlists to the database
func (db *DB) AddPlaylists(playlists []*Playlist) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO playlists (user_id, name, is_favorite) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, playlist := range playlists {
		result, err := stmt.Exec(playlist.UserID, playlist.Name, playlist.IsFavorite)
		if err != nil {
			return err
		}
		playlist.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}

		for _, track := range playlist.Tracks {
			_, err = tx.Exec("INSERT INTO playlist_tracks (playlist_id, track_id) VALUES (?, ?)", playlist.ID, track.ID)
			if err != nil {
				return err
			}
		}

		for _, artist := range playlist.Artists {
			_, err = tx.Exec("INSERT INTO playlist_artists (playlist_id, artist_id) VALUES (?, ?)", playlist.ID, artist.ID)
			if err != nil {
				return err
			}
		}

		for _, album := range playlist.Albums {
			_, err = tx.Exec("INSERT INTO playlist_albums (playlist_id, album_id) VALUES (?, ?)", playlist.ID, album.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetPlaylists retrieves multiple playlists from the database
func (db *DB) GetPlaylists(key, value string) ([]*Playlist, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = "SELECT playlist_id, user_id, name, is_favorite FROM playlists"
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf("SELECT playlist_id, user_id, name, is_favorite FROM playlists WHERE %s = ?", key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*Playlist
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.IsFavorite)
		if err != nil {
			return nil, err
		}

		trackRows, err := db.Query("SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum FROM tracks t JOIN playlist_tracks pt ON t.track_id = pt.track_id WHERE pt.playlist_id = ?", playlist.ID)
		if err != nil {
			return nil, err
		}
		defer trackRows.Close()

		for trackRows.Next() {
			var track Track
			err := trackRows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum)
			if err != nil {
				return nil, err
			}
			playlist.Tracks = append(playlist.Tracks, track)
		}

		artistRows, err := db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN playlist_artists pa ON a.artist_id = pa.artist_id WHERE pa.playlist_id = ?", playlist.ID)
		if err != nil {
			return nil, err
		}
		defer artistRows.Close()

		for artistRows.Next() {
			var artist Artist
			err := artistRows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
			if err != nil {
				return nil, err
			}
			playlist.Artists = append(playlist.Artists, artist)
		}

		albumRows, err := db.Query("SELECT a.album_id, a.name, a.release_date, a.image_uri FROM albums a JOIN playlist_albums pa ON a.album_id = pa.album_id WHERE pa.playlist_id = ?", playlist.ID)
		if err != nil {
			return nil, err
		}
		defer albumRows.Close()

		for albumRows.Next() {
			var album Album
			err := albumRows.Scan(&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
			if err != nil {
				return nil, err
			}
			playlist.Albums = append(playlist.Albums, album)
		}

		playlists = append(playlists, &playlist)
	}

	return playlists, nil
}

// SearchTracks searches for tracks based on a query string
func (db *DB) SearchTracks(query string) ([]*Track, error) {
	rows, err := db.Query("SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha256sum FROM tracks WHERE name LIKE ? OR lyrics LIKE ?", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// GetTopTracks returns the top N most listened tracks
func (db *DB) GetTopTracks(limit int) ([]*Track, error) {
	rows, err := db.Query(`
		SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum, COUNT(*) as listen_count
		FROM tracks t
		JOIN listens l ON t.track_id = l.track_id
		GROUP BY t.track_id
		ORDER BY listen_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		var listenCount int
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum, &listenCount)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// GetRecentlyAddedTracks returns the N most recently added tracks
func (db *DB) GetRecentlyAddedTracks(limit int) ([]*Track, error) {
	rows, err := db.Query(`
		SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha256sum
		FROM tracks
		ORDER BY track_id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// GetUserFavoritePlaylists returns a user's favorite playlists
func (db *DB) GetUserFavoritePlaylists(userID int64) ([]*Playlist, error) {
	rows, err := db.Query("SELECT playlist_id, user_id, name, is_favorite FROM playlists WHERE user_id = ? AND is_favorite = 1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*Playlist
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.IsFavorite)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, &playlist)
	}

	return playlists, nil
}

// GetTracksByTag returns tracks associated with a specific tag
func (db *DB) GetTracksByTag(tagID int64) ([]*Track, error) {
	rows, err := db.Query(`
		SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum
		FROM tracks t
		JOIN track_tags tt ON t.track_id = tt.track_id
		WHERE tt.tag_id = ?
	`, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// AddTagsToTrack adds multiple tags to a track
func (db *DB) AddTagsToTrack(trackID int64, tagIDs []int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO track_tags (track_id, tag_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tagID := range tagIDs {
		_, err = stmt.Exec(trackID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveTagsFromTrack removes multiple tags from a track
func (db *DB) RemoveTagsFromTrack(trackID int64, tagIDs []int64) error {
	placeholders := make([]string, len(tagIDs))
	args := make([]interface{}, len(tagIDs)+1)
	args[0] = trackID
	for i, tagID := range tagIDs {
		placeholders[i] = "?"
		args[i+1] = tagID
	}

	query := fmt.Sprintf("DELETE FROM track_tags WHERE track_id = ? AND tag_id IN (%s)", strings.Join(placeholders, ","))
	_, err := db.Exec(query, args...)
	return err
}
