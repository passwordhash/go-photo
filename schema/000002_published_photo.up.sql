CREATE TABLE "published_photo_info"
(
    "id"           int PRIMARY KEY,
    "published_at" timestamp   DEFAULT (CURRENT_TIMESTAMP),
    "public_token" varchar(64) DEFAULT (gen_random_uuid()),
    FOREIGN KEY ("id") REFERENCES "photos" ("id")
);