CREATE TABLE folders (
  id SERIAL PRIMARY KEY,
  folder_path TEXT not null
);

-- TODO: добавить пользователя в бд

CREATE TABLE photos (
  id SERIAL PRIMARY KEY,
  user_uuid UUID not null ,
  filename VARCHAR(255) not null ,
  folder_id INTEGER,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (folder_id) REFERENCES Folders (id)
);

CREATE TYPE version_type_enum AS ENUM ('original', 'thumbnail', 'preview');

CREATE TABLE photo_versions (
  id SERIAL PRIMARY KEY,
  photo_id INTEGER not null ,
  version_type version_type_enum default 'original',
  filepath VARCHAR(255) not null ,
--     TODO: может стоит вынести в отдельную таблицу
--   width INTEGER not null,
--   height INTEGER not null,
  size INTEGER not null,
  FOREIGN KEY (photo_id) REFERENCES Photos (id)
);

CREATE INDEX idx_photos_user_uuid ON photos (user_uuid);
CREATE INDEX idx_photo_versions_photo_id ON photo_versions (photo_id);