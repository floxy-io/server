package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var sshPairTable = `
CREATE TABLE IF NOT EXISTS floxy (
	fingerprint TEXT PRIMARY KEY,
	publicKey TEXT NOT NULL,
	port      NUMBER,
	createdAt DATE,
	activated BOOLEAN,
	expireAt  DATE,
	remotePassword TEXT
);
`
func Get()*sql.DB{
	return db
}

func Setup()error{
	var err error
	db, err = sql.Open("sqlite3", "main.db")
	if err != nil {
		return err
	}
	_, err = db.Exec(sshPairTable)
	if err != nil {
		return err
	}
	return nil
}
