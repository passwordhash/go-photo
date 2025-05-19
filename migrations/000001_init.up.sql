CREATE TABLE photos
(
    id          SERIAL PRIMARY KEY,
    user_uuid   UUID         not null,
    filename    VARCHAR(255) not null,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE version_type_enum AS ENUM ('original', 'thumbnail', 'preview');

CREATE TABLE photo_versions
(
    id           SERIAL PRIMARY KEY,
    photo_id     INTEGER      not null,
    version_type version_type_enum default 'original',
    uuid_filename     VARCHAR(255) not null,
    size         INTEGER      not null,
    width        INTEGER      not null,
    height       INTEGER      not null,
    saved_at    TIMESTAMP DEFAULT null,

    FOREIGN KEY (photo_id) REFERENCES Photos (id)
);

CREATE INDEX idx_photos_user_uuid ON photos (user_uuid);
CREATE INDEX idx_photo_versions_photo_id ON photo_versions (photo_id);