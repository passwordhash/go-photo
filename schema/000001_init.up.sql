CREATE TABLE Folders (
  id UUID PRIMARY KEY,
  folder_path TEXT
);

CREATE TABLE Photos (
  id UUID PRIMARY KEY,
  user_id UUID,
  filename VARCHAR(255),
  folder_id UUID,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (folder_id) REFERENCES Folders (id)
);

CREATE TYPE version_type_enum AS ENUM ('original', 'thumbnail', 'preview');

CREATE TABLE PhotoVersions (
  id UUID PRIMARY KEY,
  photo_id UUID,
  version_type version_type_enum,
  filepath VARCHAR(255),
  width INTEGER,
  height INTEGER,
  FOREIGN KEY (photo_id) REFERENCES Photos (id)
);