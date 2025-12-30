-- +goose Up
CREATE TABLE artists (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,

    name TEXT NOT NULL CHECK(name<>''),
    other_name TEXT,

    picture TEXT,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE albums (
    id TEXT PRIMARY KEY,

    name TEXT NOT NULL CHECK(name<>''),
    other_name TEXT,

    artist_id TEXT NOT NULL REFERENCES artists(id),

    cover_art TEXT,
    year INT,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE albums_featuring_artists (
    album_id TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    artist_id TEXT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,

    PRIMARY KEY(album_id, artist_id)
);

CREATE TABLE tracks (
    id TEXT PRIMARY KEY,

    filename TEXT NOT NULL,
    modified_time INT NOT NULL,
    media_type TEXT NOT NULL,

    name TEXT NOT NULL CHECK(name<>''),
    other_name TEXT,

    album_id TEXT NOT NULL REFERENCES albums(id),
    artist_id TEXT NOT NULL REFERENCES artists(id),

    duration INT NOT NULL,
    number INT,
    year INT,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE tracks_featuring_artists (
    track_id TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    artist_id TEXT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,

    PRIMARY KEY(track_id, artist_id)
);

CREATE TABLE tags (
    slug TEXT PRIMARY KEY
);

CREATE TABLE artists_tags  (
    artist_id TEXT REFERENCES artists(id) ON DELETE CASCADE,
    tag_slug TEXT REFERENCES tags(slug) ON DELETE CASCADE,

    PRIMARY KEY(artist_id, tag_slug)
);

CREATE TABLE albums_tags  (
    album_id TEXT REFERENCES albums(id) ON DELETE CASCADE,
    tag_slug TEXT REFERENCES tags(slug) ON DELETE CASCADE,

    PRIMARY KEY(album_id, tag_slug)
);

CREATE TABLE tracks_tags (
    track_id TEXT REFERENCES tracks(id) ON DELETE CASCADE,
    tag_slug TEXT REFERENCES tags(slug) ON DELETE CASCADE,

    PRIMARY KEY(track_id, tag_slug)
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL COLLATE NOCASE CHECK(username<>'') UNIQUE,
    password TEXT NOT NULL CHECK(password<>''),
    role TEXT NOT NULL,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE users_settings (
    id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT,

    quick_playlist TEXT REFERENCES playlists(id) ON DELETE SET NULL
);

CREATE TABLE playlists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK(name<>''),
    picture TEXT,

    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE playlist_items (
    playlist_id TEXT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,

    PRIMARY KEY(playlist_id, track_id)
);

CREATE TABLE taglists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK(name<>''),

    filter TEXT NOT NULL,

    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE api_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name TEXT NOT NULL CHECK(name<>''),

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

-- +goose Down
DROP TABLE api_tokens; 

DROP TABLE taglists; 
DROP TABLE playlist_items; 
DROP TABLE playlists; 

DROP TABLE users_settings; 
DROP TABLE users; 

DROP TABLE tracks_tags; 
DROP TABLE albums_tags ; 
DROP TABLE artists_tags ; 
DROP TABLE tags; 

DROP TABLE tracks_featuring_artists; 
DROP TABLE tracks_media; 
DROP TABLE tracks; 

DROP TABLE albums_featuring_artists; 
DROP TABLE albums; 

DROP TABLE artists; 

