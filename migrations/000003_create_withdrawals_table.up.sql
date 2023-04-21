CREATE TABLE IF NOT EXISTS "withdrawals" (
    "order_number" bigint PRIMARY KEY,
	"user_login" varchar(100) NOT NULL REFERENCES users (login),
	"sum" numeric NOT NULL,
	"processed_at" timestamp NOT NULL);