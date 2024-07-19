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
    album_id INTEGER,
    artist_id INTEGER,
    PRIMARY KEY (album_id, artist_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id)
);

-- Tracks table
CREATE TABLE IF NOT EXISTS tracks (
    track_id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER,
    name TEXT NOT NULL,
    duration INTEGER, -- seconds
    lyrics TEXT,
    is_explicit BOOLEAN DEFAULT 0,
    file_path TEXT, -- Path to the track file
    sha512sum TEXT NOT NULL UNIQUE, -- SHA-512 checksum of the track file
    FOREIGN KEY (album_id) REFERENCES albums (album_id)
);

-- Track Artists table
CREATE TABLE IF NOT EXISTS track_artists (
    track_id INTEGER,
    artist_id INTEGER,
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
    user_id INTEGER,
    track_id INTEGER,
    listen_time INTEGER, -- seconds
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
    track_id INTEGER,
    tag_id INTEGER,
    PRIMARY KEY (track_id, tag_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Album Tags table
CREATE TABLE IF NOT EXISTS album_tags (
    album_id INTEGER,
    tag_id INTEGER,
    PRIMARY KEY (album_id, tag_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Artist Tags table
CREATE TABLE IF NOT EXISTS artist_tags (
    artist_id INTEGER,
    tag_id INTEGER,
    PRIMARY KEY (artist_id, tag_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id),
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id)
);

-- Playlists table
CREATE TABLE IF NOT EXISTS playlists (
    playlist_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    name TEXT NOT NULL,
    is_favorite BOOLEAN DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);

-- Playlist Tracks table
CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id INTEGER,
    track_id INTEGER,
    PRIMARY KEY (playlist_id, track_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (track_id) REFERENCES tracks (track_id)
);

-- Playlist Artists table
CREATE TABLE IF NOT EXISTS playlist_artists (
    playlist_id INTEGER,
    artist_id INTEGER,
    PRIMARY KEY (playlist_id, artist_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (artist_id) REFERENCES artists (artist_id)
);

-- Playlist Albums table
CREATE TABLE IF NOT EXISTS playlist_albums (
    playlist_id INTEGER,
    album_id INTEGER,
    PRIMARY KEY (playlist_id, album_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists (playlist_id),
    FOREIGN KEY (album_id) REFERENCES albums (album_id)
);
