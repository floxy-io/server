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
	CreatedAt      time.Time `json:"createdAt"`
	Activated      bool `json:"isActive"`
	Port           int `json:"-"`
}

func (f Floxy) IsActive()bool{
	if time.Now().Sub(f.CreatedAt).Minutes() < 30 {
		return true
	}
	if f.Expiration.After(time.Now()){
		return false
	}
	if !f.Activated {
		return false
	}
	return true
}

func (f Floxy) ExpiredLink()bool{
	if time.Now().Sub(f.CreatedAt).Minutes() > 10 {
		return true
	}
	return false
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
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port,remotePassword,expireAt,createdAt FROM floxy WHERE fingerprint=?", fingerprint)
	if err != nil {
		return Floxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var remotePass *string
		var expireAt time.Time
		var createdAt time.Time
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port, &remotePass, &expireAt, &createdAt)
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
			CreatedAt:      createdAt,
			Port:           port,
		}, nil
	}
	return Floxy{}, fmt.Errorf("no scan")
}

func GetAll() ([]Floxy, error){
	sshAll := make([]Floxy, 0)
	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port,activated,createdAt,expireAt FROM floxy")
	if err != nil {
		return sshAll, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var activated bool
		var createdAt time.Time
		var expireAt  time.Time
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port, &activated, &createdAt, &expireAt)
		if err != nil {
			return sshAll, err
		}

		publicDec, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return sshAll, err
		}
		sshAll = append(sshAll, Floxy{
			PublicKey:      publicDec,
			Fingerprint:    fingerprint,
			RemotePassword: nil,
			Expiration:     expireAt,
			CreatedAt:      createdAt,
			Activated:      activated,
			Port:           port,
		})
	}
	return sshAll, nil
}

func SetPort(fingerprint string, port int)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE floxy SET port=? WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(port, fingerprint)
	if err != nil {
		return err
	}
	return nil
}

func Remove(fingerprint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("DELETE floxy WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(fingerprint)
	if err != nil {
		return err
	}
	return nil
}

func ActiveProxy(fingerprint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE floxy SET activated=true WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(fingerprint)
	if err != nil {
		return err
	}
	return nil
}
