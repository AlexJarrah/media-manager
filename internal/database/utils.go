package database

import (
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

// AddArtist adds a new artist to the database
func (db *DB) AddArtist(artist *Artist) error {
	stmt, err := db.Prepare("INSERT INTO artists (name, bio, image_uri) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(artist.Name, artist.Bio, artist.ImageURI)
	if err != nil {
		return err
	}

	artist.ID, err = result.LastInsertId()
	return err
}

// GetArtist retrieves an artist from the database
func (db *DB) GetArtist(id int64) (*Artist, error) {
	var artist Artist
	err := db.QueryRow("SELECT artist_id, name, bio, image_uri FROM artists WHERE artist_id = ?", id).Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
	if err != nil {
		return nil, err
	}
	return &artist, nil
}

// UpdateArtist updates an existing artist in the database
func (db *DB) UpdateArtist(artist *Artist) error {
	_, err := db.Exec("UPDATE artists SET name = ?, bio = ?, image_uri = ? WHERE artist_id = ?", artist.Name, artist.Bio, artist.ImageURI, artist.ID)
	return err
}

// DeleteArtist deletes an artist from the database
func (db *DB) DeleteArtist(id int64) error {
	_, err := db.Exec("DELETE FROM artists WHERE artist_id = ?", id)
	return err
}

// AddAlbum adds a new album to the database
func (db *DB) AddAlbum(album *Album) error {
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

	return tx.Commit()
}

// GetAlbum retrieves an album from the database
func (db *DB) GetAlbum(id int64) (*Album, error) {
	var album Album
	err := db.QueryRow("SELECT album_id, name, release_date, image_uri FROM albums WHERE album_id = ?", id).Scan(&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN album_artists aa ON a.artist_id = aa.artist_id WHERE aa.album_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist Artist
		err := rows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
		if err != nil {
			return nil, err
		}
		album.Artists = append(album.Artists, artist)
	}

	return &album, nil
}

// UpdateAlbum updates an existing album in the database
func (db *DB) UpdateAlbum(album *Album) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE albums SET name = ?, release_date = ?, image_uri = ? WHERE album_id = ?", album.Name, album.ReleaseDate, album.ImageURI, album.ID)
	if err != nil {
		return err
	}

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

	return tx.Commit()
}

// DeleteAlbum deletes an album from the database
func (db *DB) DeleteAlbum(id int64) error {
	_, err := db.Exec("DELETE FROM albums WHERE album_id = ?", id)
	return err
}

// AddTrack adds a new track to the database
func (db *DB) AddTrack(track *Track) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tracks (album_id, name, duration, lyrics, is_explicit, file_path, sha512sum) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(track.AlbumID, track.Name, track.Duration, track.Lyrics, track.IsExplicit, track.FilePath, track.SHA512Sum)
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

	return tx.Commit()
}

// GetTrack retrieves a track from the database
func (db *DB) GetTrack(id int64) (*Track, error) {
	var track Track
	err := db.QueryRow("SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha512sum FROM tracks WHERE track_id = ?", id).Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN track_artists ta ON a.artist_id = ta.artist_id WHERE ta.track_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist Artist
		err := rows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
		if err != nil {
			return nil, err
		}
		track.Artists = append(track.Artists, artist)
	}

	rows, err = db.Query("SELECT t.tag_id, t.name FROM tags t JOIN track_tags tt ON t.tag_id = tt.tag_id WHERE tt.track_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, err
		}
		track.Tags = append(track.Tags, tag)
	}

	return &track, nil
}

// UpdateTrack updates an existing track in the database
func (db *DB) UpdateTrack(track *Track) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE tracks SET album_id = ?, name = ?, duration = ?, lyrics = ?, is_explicit = ?, file_path = ?, sha512sum = ? WHERE track_id = ?", track.AlbumID, track.Name, track.Duration, track.Lyrics, track.IsExplicit, track.FilePath, track.SHA512Sum, track.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM track_artists WHERE track_id = ?", track.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM track_tags WHERE track_id = ?", track.ID)
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

	return tx.Commit()
}

// DeleteTrack deletes a track from the database
func (db *DB) DeleteTrack(id int64) error {
	_, err := db.Exec("DELETE FROM tracks WHERE track_id = ?", id)
	return err
}

// AddUser adds a new user to the database
func (db *DB) AddUser(user *User) error {
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO users (name, preferences) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Name, preferencesJSON)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

// GetUser retrieves a user from the database
func (db *DB) GetUser(id int64) (*User, error) {
	var user User
	var preferencesJSON []byte
	err := db.QueryRow("SELECT user_id, name, preferences FROM users WHERE user_id = ?", id).Scan(&user.ID, &user.Name, &preferencesJSON)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(preferencesJSON, &user.Preferences)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user in the database
func (db *DB) UpdateUser(user *User) error {
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET name = ?, preferences = ? WHERE user_id = ?", user.Name, preferencesJSON, user.ID)
	return err
}

// DeleteUser deletes a user from the database
func (db *DB) DeleteUser(id int64) error {
	_, err := db.Exec("DELETE FROM users WHERE user_id = ?", id)
	return err
}

// AddListen adds a new listen event to the database
func (db *DB) AddListen(listen *Listen) error {
	stmt, err := db.Prepare("INSERT INTO listens (user_id, track_id, listen_time, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(listen.UserID, listen.TrackID, listen.ListenTime, listen.Timestamp)
	if err != nil {
		return err
	}

	listen.ID, err = result.LastInsertId()
	return err
}

// GetUserListens retrieves all listen events for a user from the database
func (db *DB) GetUserListens(userID int64) ([]Listen, error) {
	rows, err := db.Query("SELECT listen_id, user_id, track_id, listen_time, timestamp FROM listens WHERE user_id = ? ORDER BY timestamp DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listens []Listen
	for rows.Next() {
		var listen Listen
		err := rows.Scan(&listen.ID, &listen.UserID, &listen.TrackID, &listen.ListenTime, &listen.Timestamp)
		if err != nil {
			return nil, err
		}
		listens = append(listens, listen)
	}

	return listens, nil
}

// AddTag adds a new tag to the database
func (db *DB) AddTag(tag *Tag) error {
	stmt, err := db.Prepare("INSERT INTO tags (name) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(tag.Name)
	if err != nil {
		return err
	}

	tag.ID, err = result.LastInsertId()
	return err
}

// GetTag retrieves a tag from the database
func (db *DB) GetTag(id int64) (*Tag, error) {
	var tag Tag
	err := db.QueryRow("SELECT tag_id, name FROM tags WHERE tag_id = ?", id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// UpdateTag updates an existing tag in the database
func (db *DB) UpdateTag(tag *Tag) error {
	_, err := db.Exec("UPDATE tags SET name = ? WHERE tag_id = ?", tag.Name, tag.ID)
	return err
}

// DeleteTag deletes a tag from the database
func (db *DB) DeleteTag(id int64) error {
	_, err := db.Exec("DELETE FROM tags WHERE tag_id = ?", id)
	return err
}

// AddPlaylist adds a new playlist to the database
func (db *DB) AddPlaylist(playlist *Playlist) error {
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

	return tx.Commit()
}

// GetPlaylist retrieves a playlist from the database
func (db *DB) GetPlaylist(id int64) (*Playlist, error) {
	var playlist Playlist
	err := db.QueryRow("SELECT playlist_id, user_id, name, is_favorite FROM playlists WHERE playlist_id = ?", id).Scan(&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.IsFavorite)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha512sum FROM tracks t JOIN playlist_tracks pt ON t.track_id = pt.track_id WHERE pt.playlist_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum)
		if err != nil {
			return nil, err
		}
		playlist.Tracks = append(playlist.Tracks, track)
	}

	rows, err = db.Query("SELECT a.artist_id, a.name, a.bio, a.image_uri FROM artists a JOIN playlist_artists pa ON a.artist_id = pa.artist_id WHERE pa.playlist_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist Artist
		err := rows.Scan(&artist.ID, &artist.Name, &artist.Bio, &artist.ImageURI)
		if err != nil {
			return nil, err
		}
		playlist.Artists = append(playlist.Artists, artist)
	}

	rows, err = db.Query("SELECT a.album_id, a.name, a.release_date, a.image_uri FROM albums a JOIN playlist_albums pa ON a.album_id = pa.album_id WHERE pa.playlist_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var album Album
		err := rows.Scan(&album.ID, &album.Name, &album.ReleaseDate, &album.ImageURI)
		if err != nil {
			return nil, err
		}
		playlist.Albums = append(playlist.Albums, album)
	}

	return &playlist, nil
}

// UpdatePlaylist updates an existing playlist in the database
func (db *DB) UpdatePlaylist(playlist *Playlist) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE playlists SET user_id = ?, name = ?, is_favorite = ? WHERE playlist_id = ?", playlist.UserID, playlist.Name, playlist.IsFavorite, playlist.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM playlist_tracks WHERE playlist_id = ?", playlist.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM playlist_artists WHERE playlist_id = ?", playlist.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM playlist_albums WHERE playlist_id = ?", playlist.ID)
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

	return tx.Commit()
}

// DeletePlaylist deletes a playlist from the database
func (db *DB) DeletePlaylist(id int64) error {
	_, err := db.Exec("DELETE FROM playlists WHERE playlist_id = ?", id)
	return err
}

// SearchTracks searches for tracks based on a query string
func (db *DB) SearchTracks(query string) ([]Track, error) {
	rows, err := db.Query("SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha512sum FROM tracks WHERE name LIKE ? OR lyrics LIKE ?", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetTopTracks returns the top N most listened tracks
func (db *DB) GetTopTracks(limit int) ([]Track, error) {
	rows, err := db.Query(`
		SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha512sum, COUNT(*) as listen_count
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

	var tracks []Track
	for rows.Next() {
		var track Track
		var listenCount int
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum, &listenCount)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetRecentlyAddedTracks returns the N most recently added tracks
func (db *DB) GetRecentlyAddedTracks(limit int) ([]Track, error) {
	rows, err := db.Query(`
		SELECT track_id, album_id, name, duration, lyrics, is_explicit, file_path, sha512sum
		FROM tracks
		ORDER BY track_id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetUserFavoritePlaylists returns a user's favorite playlists
func (db *DB) GetUserFavoritePlaylists(userID int64) ([]Playlist, error) {
	rows, err := db.Query("SELECT playlist_id, user_id, name, is_favorite FROM playlists WHERE user_id = ? AND is_favorite = 1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.IsFavorite)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

// GetTracksByTag returns tracks associated with a specific tag
func (db *DB) GetTracksByTag(tagID int64) ([]Track, error) {
	rows, err := db.Query(`
		SELECT t.track_id, t.album_id, t.name, t.duration, t.lyrics, t.is_explicit, t.file_path, t.sha512sum
		FROM tracks t
		JOIN track_tags tt ON t.track_id = tt.track_id
		WHERE tt.tag_id = ?
	`, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.AlbumID, &track.Name, &track.Duration, &track.Lyrics, &track.IsExplicit, &track.FilePath, &track.SHA512Sum)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
