package compiler

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/onsi/gomega/gexec"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MakeRequest struct {
	PKey           string
	FingerPrint    string
	RemotePassword *string
}

type MakeResponse struct {
	FingerPrint string
}

var mutex sync.Mutex


func Make(req MakeRequest)(MakeResponse, error){
	mutex.Lock()
	defer mutex.Unlock()

	err := compile(req)
	if err != nil {
		return MakeResponse{}, err
	}

	if os.Getenv("LOG_KEY") == "true"{
		log.Println("\nkey: ", req.PKey, "\nfingerprint: ", req.FingerPrint)
	}

	return MakeResponse{FingerPrint: req.FingerPrint}, nil
}

var CustomGoPath string
var CustomPath string

func compile(req MakeRequest)error{
	ldFlags := fmt.Sprintf("-X main.FingerPrint=%s -X main.PrivateKey=%s -X main.SshHost=%s", req.FingerPrint, req.PKey, os.Getenv("FLOXY_SSH_HOST"))
	if req.RemotePassword != nil {
		ldFlags += fmt.Sprintf(" -X main.RemotePassword=%s", *req.RemotePassword)
	}

	if CustomPath == "" {
		CustomPath = "internal/cook/cook.go"
	}
	var err error
	var compStr string
	if CustomGoPath != "" {
		log.Println("using custom Gopath: ", CustomGoPath)
		compStr, err = gexec.BuildIn(CustomGoPath,CustomPath,"-ldflags",ldFlags)
	}else {
		compStr, err = gexec.Build(CustomPath,"-ldflags",ldFlags)
	}
	if err != nil {
		return err
	}



	newLocation := filepath.Join("internal", "home", "cooked_bin", req.FingerPrint, "floxy")

	path := filepath.Join("internal", "home", "cooked_bin",req.FingerPrint)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return err
		}
	}

	err = os.Rename(compStr, newLocation)
	if err != nil {
		return err
	}
	return nil
}

func getLdFlagFromKey(pKey *rsa.PrivateKey)string{
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(pKey)
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	blockStr := string(keyPem)
	blockStr = strings.Replace(blockStr,"\n","<br>",-1)
	blockStr = strings.Replace(blockStr," ","<p>",-1)
	return blockStr
}