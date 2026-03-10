package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func PKeyPath(username string) string {
	return fmt.Sprintf("storage/cred/%s.key", username)
}

func PubKeyPath(username string) string {
	return fmt.Sprintf("storage/cred/%s.pub", username)
}

func ReadPrivKey(path string) (*rsa.PrivateKey, error) {
	cont, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(cont)
	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	return pkey, err
}
