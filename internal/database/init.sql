-- Artists table
CREATE TABLE IF NOT EXISTS artists (
    artist_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    bio TEXT,
    image_uri TEXT
);

-- Albums table
CREATE TABLE IF NOT EXISTS albums (
    album_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    release_date DATE,
    image_uri TEXT
);

-- Album Artists table
CREATE TABLE IF NOT EXISTS album_artists (
    album_id INTEGER NOT NULL,
    artist_id INTEGER NOT NULL,
    PRIMARY KEY (album_id, artist_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id)
);

-- Tracks table
CREATE TABLE IF NOT EXISTS tracks (
    track_id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    duration INTEGER NOT NULL, -- seconds
    lyrics TEXT,
    is_explicit BOOLEAN DEFAULT 0,
    file_path TEXT NOT NULL UNIQUE, -- Path to the track file
    sha256sum TEXT NOT NULL UNIQUE, -- SHA-256 checksum of the track file
    FOREIGN KEY (album_id) REFERENCES albums (album_id)
);

-- Track Artists table
CREATE TABLE IF NOT EXISTS track_artists (
    track_id INTEGER NOT NULL,
    artist_id INTEGER NOT NULL,
    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id)
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    preferences TEXT -- JSON or text field for user preferences
);

-- Listens table
CREATE TABLE IF NOT EXISTS listens (
    listen_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    track_id INTEGER NOT NULL,
    listen_time INTEGER NOT NULL, -- seconds
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (user_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id)
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    tag_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Track Tags table
CREATE TABLE IF NOT EXISTS track_tags (
    track_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (track_id, tag_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Album Tags table
CREATE TABLE IF NOT EXISTS album_tags (
    album_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (album_id, tag_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Artist Tags table
CREATE TABLE IF NOT EXISTS artist_tags (
    artist_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (artist_id, tag_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Playlists table
CREATE TABLE IF NOT EXISTS playlists (
    playlist_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    is_favorite BOOLEAN DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);

-- Playlist Tracks table
CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id INTEGER NOT NULL,
    track_id INTEGER NOT NULL,
    PRIMARY KEY (playlist_id, track_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id)
);

-- Playlist Artists table
CREATE TABLE IF NOT EXISTS playlist_artists (
    playlist_id INTEGER NOT NULL,
    artist_id INTEGER NOT NULL,
    PRIMARY KEY (playlist_id, artist_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id)
);

-- Playlist Albums table
CREATE TABLE IF NOT EXISTS playlist_albums (
    playlist_id INTEGER NOT NULL,
    album_id INTEGER NOT NULL,
    PRIMARY KEY (playlist_id, album_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id)
);
