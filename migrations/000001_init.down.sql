DROP INDEX IF EXISTS idx_photo_versions_photo_id;
DROP INDEX IF EXISTS idx_photos_user_uuid;

DROP TABLE IF EXISTS photo_versions CASCADE; -- Удаляет зависимые объекты
DROP TABLE IF EXISTS photos CASCADE;
DROP TABLE IF EXISTS folders CASCADE;

DROP TYPE IF EXISTS version_type_enum;
