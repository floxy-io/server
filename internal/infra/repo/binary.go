package repo

import (
	"encoding/base64"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/db"
	"time"
)

var insertOnTable = `
INSERT INTO floxy (fingerprint,status)
VALUES(?,'burning');
`

var updateOnTable = `
UPDATE floxy SET publicKey=?,port=?,createdAt=?, remotePassword=?, expireAt=?
WHERE fingerprint=?;
`

var setStatusOnTable = `
UPDATE floxy SET status=?
WHERE fingerprint=?;
`

type Floxy struct {
	PublicKey      []byte `json:"-"`
	Fingerprint    string `json:"fingerPrint"`
	RemotePassword *string `json:"remotePassword"`
	Expiration     time.Time `json:"expireAt"`
	CreatedAt      time.Time `json:"createdAt"`
	Activated      bool `json:"isActive"`
	Port           int `json:"-"`
	Status         string `json:"status"`
}

type FloxyBinary struct {
	Parent         string `json:"parent"`
	Fingerprint    string `json:"fingerPrint"`
	Kind           string `json:"kind"`
	Os             string `json:"os"`
	Platform       string `json:"platform"`
}

var addChildOnTable = `
INSERT INTO floxy_binary (parent,fingerprint,kind,os,platform)
VALUES(?,?,?,?,?);
`


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

func AddNewFloxy(fingerPrint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(insertOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(fingerPrint)
	if err != nil {
		return err
	}
	return nil
}



func AddFloxyBinary(bin FloxyBinary)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(addChildOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(bin.Parent, bin.Fingerprint, bin.Kind, bin.Os, bin.Platform)
	if err != nil {
		return err
	}
	return nil
}

func SetFailed(fingerPrint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(setStatusOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("failed", fingerPrint)
	if err != nil {
		return err
	}
	return nil
}

func SetActive(fingerPrint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(setStatusOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("active", fingerPrint)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFloxy(k Floxy)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(updateOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(base64.StdEncoding.EncodeToString(k.PublicKey), k.Port, time.Now(), k.RemotePassword, k.Expiration,k.Fingerprint)
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
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port,remotePassword,expireAt,createdAt,status FROM floxy WHERE fingerprint=?", fingerprint)
	if err != nil {
		return Floxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey *string
		var remotePass *string
		var expireAt *time.Time
		var createdAt *time.Time
		var port *int
		var status string
		err = row.Scan(&fingerprint, &publicKey, &port, &remotePass, &expireAt, &createdAt, &status)
		if err != nil {
			return Floxy{}, err
		}

		if status == "burning"{
			return Floxy{Status: status, Fingerprint: fingerprint}, nil
		}

		publicDec, err := base64.StdEncoding.DecodeString(*publicKey)
		if err != nil {
			return Floxy{}, err
		}
		return Floxy{
			PublicKey:      publicDec,
			Fingerprint:    fingerprint,
			RemotePassword: remotePass,
			Expiration:     *expireAt,
			CreatedAt:      *createdAt,
			Port:           *port,
			Status:         status,
		}, nil
	}
	return Floxy{}, fmt.Errorf("no scan")
}

func GetFloxyBinaries(parent string)([]FloxyBinary, error){
	binaries := make([]FloxyBinary, 0)
	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,kind,os,platform FROM floxy_binary WHERE parent=?", parent)
	if err != nil {
		return binaries, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var kind string
		var os string
		var plat string
		err = row.Scan(&fingerprint, &kind, &os, &plat)
		if err != nil {
			return binaries, err
		}
		binaries = append(binaries, FloxyBinary{
			Parent:      parent,
			Fingerprint: fingerprint,
			Kind:        kind,
			Os:          os,
			Platform:    plat,
		})
	}
	return binaries, nil
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
