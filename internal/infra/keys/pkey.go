package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/danielsussa/floxy/internal/env"
)

func LoadKey() *rsa.PrivateKey {
	b, err := base64.StdEncoding.DecodeString(env.Get(env.PrivateSshKey))
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(b)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key
}
