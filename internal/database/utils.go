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

// UpdateArtist updates an artist in the database
func (db *DB) UpdateArtist(artist *Artist, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"name":      artist.Name,
		"bio":       artist.Bio,
		"image_uri": artist.ImageURI,
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE artists SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err := db.Exec(queryBuilder.String(), args...)
	return err
}

// GetArtists retrieves multiple artists from the database
func (db *DB) GetArtists(key string, value any) ([]*Artist, error) {
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

// UpdateAlbum updates an album in the database
func (db *DB) UpdateAlbum(album *Album, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"name":         album.Name,
		"release_date": album.ReleaseDate,
		"image_uri":    album.ImageURI,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE albums SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err = tx.Exec(queryBuilder.String(), args...)
	if err != nil {
		return err
	}

	// Update album artists if specified
	if contains(keys, "artists") {
		_, err = tx.Exec("DELETE FROM album_artists WHERE album_id = ?", album.ID)
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
func (db *DB) GetAlbums(key string, value any) ([]*Album, error) {
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

	// Prepare statements
	stmtTrack, err := tx.Prepare("INSERT INTO tracks (album_id, name, duration, lyrics, is_explicit, file_path, sha256sum) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtTrack.Close()

	stmtAlbum, err := tx.Prepare("INSERT INTO albums (name, release_date, image_uri) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtAlbum.Close()

	for _, track := range tracks {
		// Check if album exists, if not, add it
		var albumID int64
		err := db.QueryRow("SELECT album_id FROM albums WHERE name = ?", track.Album.Name).Scan(&albumID)
		if err == sql.ErrNoRows {
			result, err := stmtAlbum.Exec(track.Album.Name, track.Album.ReleaseDate, track.Album.ImageURI)
			if err != nil {
				return err
			}
			albumID, err = result.LastInsertId()
			if err != nil {
				return err
			}
			track.Album.ID = albumID
		} else if err != nil {
			return err
		}

		// Insert track
		result, err := stmtTrack.Exec(albumID, track.Name, track.Duration, track.Lyrics, track.IsExplicit, track.FilePath, track.SHA256Sum)
		if err != nil {
			return err
		}
		track.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}

		// Add artists
		for _, artist := range track.Artists {
			_, err = tx.Exec("INSERT INTO track_artists (track_id, artist_id) VALUES (?, ?)", track.ID, artist.ID)
			if err != nil {
				return err
			}
		}

		// Add tags
		for _, tag := range track.Tags {
			_, err = tx.Exec("INSERT INTO track_tags (track_id, tag_id) VALUES (?, ?)", track.ID, tag.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// UpdateTrack updates a track in the database
func (db *DB) UpdateTrack(track *Track, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"name":        track.Name,
		"duration":    track.Duration,
		"lyrics":      track.Lyrics,
		"is_explicit": track.IsExplicit,
		"file_path":   track.FilePath,
		"sha256sum":   track.SHA256Sum,
		"album_id":    track.Album.ID,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE tracks SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err = tx.Exec(queryBuilder.String(), args...)
	if err != nil {
		return err
	}

	// Update track artists if specified
	if contains(keys, "artists") {
		_, err = tx.Exec("DELETE FROM track_artists WHERE track_id = ?", track.ID)
		if err != nil {
			return err
		}
		for _, artist := range track.Artists {
			_, err = tx.Exec("INSERT INTO track_artists (track_id, artist_id) VALUES (?, ?)", track.ID, artist.ID)
			if err != nil {
				return err
			}
		}
	}

	// Update track tags if specified
	if contains(keys, "tags") {
		_, err = tx.Exec("DELETE FROM track_tags WHERE track_id = ?", track.ID)
		if err != nil {
			return err
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
func (db *DB) GetTracks(key string, value any) ([]*Track, error) {
	var query string
	var rows *sql.Rows
	var err error

	if value == "" {
		query = `
      SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
      a.album_id, a.name, a.release_date, a.image_uri
      FROM tracks t
      JOIN albums a ON t.album_id = a.album_id
    `
		rows, err = db.Query(query)
	} else {
		query = fmt.Sprintf(`
      SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
        a.album_id, a.name, a.release_date, a.image_uri
      FROM tracks t
      JOIN albums a ON t.album_id = a.album_id
      WHERE t.%s = ?
    `, key)
		rows, err = db.Query(query, value)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		var album Album
		err := rows.Scan(
			&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
			&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI,
		)
		if err != nil {
			return nil, err
		}
		track.Album = album

		// Get track artists
		trackArtistRows, err := db.Query(`
      SELECT a.artist_id, a.name, a.bio, a.image_uri
      FROM artists a
      JOIN track_artists ta ON a.artist_id = ta.artist_id
      WHERE ta.track_id = ?
    `, track.ID)
		if err != nil {
			return nil, err
		}
		defer trackArtistRows.Close()

		for trackArtistRows.Next() {
			var artist Artist
			err := trackArtistRows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
			if err != nil {
				return nil, err
			}
			track.Artists = append(track.Artists, artist)
		}

		// Get album artists
		albumArtistRows, err := db.Query(`
            SELECT a.artist_id, a.name, a.bio, a.image_uri
            FROM artists a
            JOIN album_artists aa ON a.artist_id = aa.artist_id
            WHERE aa.album_id = ?
        `, track.Album.ID)
		if err != nil {
			return nil, err
		}
		defer albumArtistRows.Close()

		for albumArtistRows.Next() {
			var artist Artist
			err := albumArtistRows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
			if err != nil {
				return nil, err
			}
			track.Album.Artists = append(track.Album.Artists, artist)
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

// UpdateUser updates a user in the database
func (db *DB) UpdateUser(user *User, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"name": user.Name,
	}

	if contains(keys, "preferences") {
		preferencesJSON, err := json.Marshal(user.Preferences)
		if err != nil {
			return err
		}
		keyMap["preferences"] = preferencesJSON
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE users SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err := db.Exec(queryBuilder.String(), args...)
	return err
}

// GetUsers retrieves multiple users from the database
func (db *DB) GetUsers(key string, value any) ([]*User, error) {
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

// UpdateListen updates a listen event in the database
func (db *DB) UpdateListen(listen *Listen, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"user_id":     listen.UserID,
		"track_id":    listen.TrackID,
		"listen_time": listen.ListenTime,
		"timestamp":   listen.Timestamp,
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE listens SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err := db.Exec(queryBuilder.String(), args...)
	return err
}

// GetUserListens retrieves all listen events for a user from the database
func (db *DB) GetUserListens(userId int64, key string, value any) ([]*Listen, error) {
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

// UpdateTag updates a tag in the database
func (db *DB) UpdateTag(tag *Tag, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"name": tag.Name,
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE tags SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err := db.Exec(queryBuilder.String(), args...)
	return err
}

// GetTags retrieves multiple tags from the database
func (db *DB) GetTags(key string, value any) ([]*Tag, error) {
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

// UpdatePlaylist updates a playlist in the database
func (db *DB) UpdatePlaylist(playlist *Playlist, keys []string, updateKey string, updateValue any) error {
	keyMap := map[string]interface{}{
		"user_id":     playlist.UserID,
		"name":        playlist.Name,
		"is_favorite": playlist.IsFavorite,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE playlists SET ")
	args := []interface{}{}
	for i, key := range keys {
		if val, ok := keyMap[key]; ok {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(key + " = ?")
			args = append(args, val)
		}
	}

	queryBuilder.WriteString(" WHERE " + updateKey + " = ?")
	args = append(args, updateValue)

	_, err = tx.Exec(queryBuilder.String(), args...)
	if err != nil {
		return err
	}

	// Update playlist tracks if specified
	if contains(keys, "tracks") {
		_, err = tx.Exec("DELETE FROM playlist_tracks WHERE playlist_id = ?", playlist.ID)
		if err != nil {
			return err
		}
		for _, track := range playlist.Tracks {
			_, err = tx.Exec("INSERT INTO playlist_tracks (playlist_id, track_id) VALUES (?, ?)", playlist.ID, track.ID)
			if err != nil {
				return err
			}
		}
	}

	// Update playlist artists if specified
	if contains(keys, "artists") {
		_, err = tx.Exec("DELETE FROM playlist_artists WHERE playlist_id = ?", playlist.ID)
		if err != nil {
			return err
		}
		for _, artist := range playlist.Artists {
			_, err = tx.Exec("INSERT INTO playlist_artists (playlist_id, artist_id) VALUES (?, ?)", playlist.ID, artist.ID)
			if err != nil {
				return err
			}
		}
	}

	// Update playlist albums if specified
	if contains(keys, "albums") {
		_, err = tx.Exec("DELETE FROM playlist_albums WHERE playlist_id = ?", playlist.ID)
		if err != nil {
			return err
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
func (db *DB) GetPlaylists(key string, value any) ([]*Playlist, error) {
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

		trackRows, err := db.Query(`
      SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
        a.album_id, a.name, a.release_date, a.image_uri
      FROM tracks t
      JOIN playlist_tracks pt ON t.track_id = pt.track_id
      JOIN albums a ON t.album_id = a.album_id
      WHERE pt.playlist_id = ?
    `, playlist.ID)
		if err != nil {
			return nil, err
		}
		defer trackRows.Close()

		for trackRows.Next() {
			var track Track
			var album Album
			err := trackRows.Scan(&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
				&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
			if err != nil {
				return nil, err
			}
			track.Album = album
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
	rows, err := db.Query(`
    SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
      a.album_id, a.name, a.release_date, a.image_uri
    FROM tracks t
    JOIN albums a ON t.album_id = a.album_id
    WHERE t.name LIKE ? OR t.lyrics LIKE ?
  `, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		var album Album
		err := rows.Scan(&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
			&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
		if err != nil {
			return nil, err
		}
		track.Album = album
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// GetTopTracks returns the top N most listened tracks
func (db *DB) GetTopTracks(limit int) ([]*Track, error) {
	rows, err := db.Query(`
    SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
      a.album_id, a.name, a.release_date, a.image_uri,
      COUNT(*) as listen_count
    FROM tracks t
    JOIN albums a ON t.album_id = a.album_id
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
		var album Album
		var listenCount int
		err := rows.Scan(&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
			&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI,
			&listenCount)
		if err != nil {
			return nil, err
		}
		track.Album = album
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

// GetRecentlyAddedTracks returns the N most recently added tracks
func (db *DB) GetRecentlyAddedTracks(limit int) ([]*Track, error) {
	rows, err := db.Query(`
    SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
      a.album_id, a.name, a.release_date, a.image_uri
    FROM tracks t
    JOIN albums a ON t.album_id = a.album_id
    ORDER BY t.track_id DESC
    LIMIT ?
  `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var track Track
		var album Album
		err := rows.Scan(&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
			&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
		if err != nil {
			return nil, err
		}
		track.Album = album
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
    SELECT t.track_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha256sum,
      a.album_id, a.name, a.release_date, a.image_uri
    FROM tracks t
    JOIN albums a ON t.album_id = a.album_id
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
		var album Album
		err := rows.Scan(&track.ID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA256Sum,
			&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
		if err != nil {
			return nil, err
		}
		track.Album = album
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

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
