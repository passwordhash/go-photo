CREATE DATABASE go_photo;


GRANT ALL PRIVILEGES ON DATABASE go_photo TO postgres;

GRANT CONNECT ON DATABASE go_photo TO yaroslav_dev;
GRANT USAGE ON SCHEMA public TO yaroslav_dev;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO yaroslav_dev;
