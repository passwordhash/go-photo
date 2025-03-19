CREATE TABLE "published_photo_info"
(
    "photo_id"     int PRIMARY KEY,
    "published_at" timestamp   DEFAULT (CURRENT_TIMESTAMP),
    "public_token" varchar(16) DEFAULT (substring(replace(gen_random_uuid()::text, '-', '') from 1 for 10)),
    FOREIGN KEY ("photo_id") REFERENCES "photos" ("id")
);