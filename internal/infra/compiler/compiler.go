package compiler

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega/gexec"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MakeRequest struct {
	PKey        *rsa.PrivateKey
	FingerPrint uuid.UUID
	Port        int
}

type MakeResponse struct {
	FingerPrint string
}

var mutex sync.Mutex


func Make(req MakeRequest)(MakeResponse, error){
	mutex.Lock()
	defer mutex.Unlock()

	{
		err := compile(req, "local")
		if err != nil {
			return MakeResponse{}, err
		}
	}

	{
		err := compile(req, "remote")
		if err != nil {
			return MakeResponse{}, err
		}
	}

	return MakeResponse{FingerPrint: req.FingerPrint.String()}, nil
}

func compile(req MakeRequest, k string)error{

	certFlag := getLdFlagFromCert(req.PKey)
	fingerPrint := req.FingerPrint.String()

	ldFlags := fmt.Sprintf("-X main.FingerPrint=%s -X main.PrivateKey=%s", fingerPrint, certFlag)

	compStr, err := gexec.Build("internal/cook/main.go","-ldflags",ldFlags)
	if err != nil {
		return err
	}

	newLocation := filepath.Join("internal", "home", "cooked_bin", fingerPrint, k)

	path := filepath.Join("internal", "home", "cooked_bin",fingerPrint)
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

func getLdFlagFromCert(pKey *rsa.PrivateKey)string{
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