DO $$ BEGIN
    CREATE TYPE lang AS ENUM ('en', 'ru');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'author', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;


CREATE TABLE IF NOT EXISTS Users(
    uid bigint PRIMARY KEY,
    name varchar(255) NOT NULL,
    language lang,
    role user_role NOT NULL DEFAULT 'user',
    banned bool NOT NULL DEFAULT false
);
