CREATE TABLE IF NOT EXISTS "orders" (
    "number"  bigint PRIMARY KEY,
	"user_login" varchar(100) NOT NULL REFERENCES users (login),
	"status" smallint NOT NULL DEFAULT 0,
	"accrual" numeric NOT NULL DEFAULT 0,
	"uploaded_at" timestamp NOT NULL);