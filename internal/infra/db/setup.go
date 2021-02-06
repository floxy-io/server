package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var sshPairTable = `
CREATE TABLE IF NOT EXISTS floxy (
	fingerprint TEXT PRIMARY KEY,
	publicKey TEXT,
	port      NUMBER,
	createdAt DATE,
	activated BOOLEAN,
	expireAt  DATE,
	remotePassword TEXT,
	status TEXT NOT NULL
);
`

var binaryTable = `
CREATE TABLE IF NOT EXISTS floxy_binary (
	fingerprint TEXT PRIMARY KEY,
	parent      TEXT NOT NULL,
	kind        TEXT NOT NULL,
	os          TEXT NOT NULL,
	platform    TEXT NOT NULL
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
	_, err = db.Exec(binaryTable)
	if err != nil {
		return err
	}
	return nil
}
