CREATE TABLE IF NOT EXISTS "users" (
    "login" varchar(100) PRIMARY KEY,
    "encrypted_password" varchar(60) NOT NULL,
    "balance" numeric NOT NULL DEFAULT 0);