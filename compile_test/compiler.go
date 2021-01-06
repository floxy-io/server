package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega/gexec"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main(){
	pKey, _ := generateCerts()
	// "-X main.GitCommit=$GIT_COMMIT"
	fingerPrint := uuid.New().String()


	ldFlags := fmt.Sprintf("-X main.FingerPrint=%s -X main.PrivateKey=%s", fingerPrint, getLdFlagFromCert(pKey))

	compStr, err := gexec.Build("home_api/cook/local/local.go","-ldflags",ldFlags)
	if err != nil {
		panic(err)
	}
	newLocation := filepath.Join("cooked_bin",fingerPrint,"local")
	os.Mkdir(filepath.Join("cooked_bin",fingerPrint), 0700)
	err = os.Rename(compStr, newLocation)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(compStr)
}

func generateCerts()(*rsa.PrivateKey, error){
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privatekey, nil
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