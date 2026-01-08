-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,

    display_name TEXT NOT NULL,
    role TEXT NOT NULL,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

CREATE TABLE user_identities (
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL,

    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL,

    PRIMARY KEY(provider, provider_id),
    UNIQUE(provider, user_id)
);

-- CREATE TABLE users_settings (
--     id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
-- );

CREATE TABLE api_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name TEXT NOT NULL CHECK(name<>''),

    created INTEGER NOT NULL,
    updated INTEGER NOT NULL
);

-- +goose Down
DROP TABLE api_tokens; 

DROP TABLE users_settings; 
DROP TABLE users; 
