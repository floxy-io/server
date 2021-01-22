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
)

type MakeRequest struct {
	PKey        *rsa.PrivateKey
	FingerPrint uuid.UUID
	Port        int
}

type MakeResponse struct {
	FingerPrint string
}

func Make(req MakeRequest)(MakeResponse, error){
	certFlag := getLdFlagFromCert(req.PKey)
	fingerPrint := req.FingerPrint.String()

	ldFlags := fmt.Sprintf("-X main.FingerPrint=%s -X main.PrivateKey=%s", fingerPrint, certFlag)

	compStr, err := gexec.Build("internal/cook/local/local.go","-ldflags",ldFlags)
	if err != nil {
		return MakeResponse{}, err
	}
	newLocation := filepath.Join("internal", "home", "cooked_bin", fingerPrint, "local")
	err = os.Mkdir(filepath.Join("internal", "home", "cooked_bin",fingerPrint), 0700)
	if err != nil {
		return MakeResponse{}, err
	}

	err = os.Rename(compStr, newLocation)
	if err != nil {
		return MakeResponse{}, err
	}

	return MakeResponse{FingerPrint: fingerPrint}, nil
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