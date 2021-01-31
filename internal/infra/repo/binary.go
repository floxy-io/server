package repo

import (
	"encoding/base64"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/db"
	"time"
)

var insertOnTable = `
INSERT INTO floxy (fingerprint, publicKey, port, createdAt, remotePassword, expireAt)
VALUES(?,?,?,?,?,?);
`

type Floxy struct {
	PublicKey      []byte `json:"-"`
	Fingerprint    string `json:"fingerPrint"`
	RemotePassword *string `json:"remotePassword"`
	Expiration     time.Time `json:"expireAt"`
	Port           int `json:"-"`
}

func AddNewFloxy(k Floxy)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(insertOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(k.Fingerprint,base64.StdEncoding.EncodeToString(k.PublicKey), k.Port, time.Now(), k.RemotePassword, k.Expiration)
	if err != nil {
		return err
	}
	return nil
}


func GetByUserAndKey(user string, public []byte)(Floxy, error){
	key := base64.StdEncoding.EncodeToString(public)

	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port FROM floxy WHERE fingerprint=? AND publicKey=?", user, key)
	if err != nil {
		return Floxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port)
		if err != nil {
			return Floxy{}, err
		}

		publicDec, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return Floxy{}, err
		}
		return Floxy{Fingerprint: fingerprint, PublicKey: publicDec, Port: port}, nil
	}
	return Floxy{}, fmt.Errorf("no scan")
}

func GetByFingerprint(fingerprint string)(Floxy, error){

	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port,remotePassword,expireAt FROM floxy WHERE fingerprint=?", fingerprint)
	if err != nil {
		return Floxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var remotePass *string
		var expireAt time.Time
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port, &remotePass, &expireAt)
		if err != nil {
			return Floxy{}, err
		}

		publicDec, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return Floxy{}, err
		}
		return Floxy{
			PublicKey:      publicDec,
			Fingerprint:    fingerprint,
			RemotePassword: remotePass,
			Expiration:     expireAt,
			Port:           port,
		}, nil
	}
	return Floxy{}, fmt.Errorf("no scan")
}

func GetAll() ([]Floxy, error){
	sshAll := make([]Floxy, 0)
	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port FROM floxy")
	if err != nil {
		return sshAll, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port)
		if err != nil {
			return sshAll, err
		}

		publicDec, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return sshAll, err
		}
		sshAll = append(sshAll, Floxy{Fingerprint: fingerprint, PublicKey: publicDec, Port: port})
	}
	return sshAll, nil
}

func SetPort(fingerprint string, port int)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE sshPair SET port=? WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(port, fingerprint)
	if err != nil {
		return err
	}
	return nil
}

func SetUpdatedAt(fingerprint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE sshPair SET updatedAd=? WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(time.Now(), fingerprint)
	if err != nil {
		return err
	}
	return nil
}