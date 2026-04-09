package sig_key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

const CredentialsDir = "storage/cred"

func PKeyPath(username string) string {
	return CredentialsDir + "/" + username + ".key"
}

func PubKeyPath(username string) string {
	return CredentialsDir + "/" + username + ".pub"
}

func GenerateKeys(username string) (*rsa.PrivateKey, error) {
	err := os.MkdirAll(CredentialsDir, 0700)
	if err != nil {
		return nil, err
	}
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	pkeyFile, err := os.Create(PKeyPath(username))
	if err != nil {
		return nil, err
	}
	err = pem.Encode(pkeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return nil, err
	}
	err = pkeyFile.Close()
	if err != nil {
		return nil, err
	}
	pubkeyFile, err := os.Create(PubKeyPath(username))
	if err != nil {
		return nil, err
	}
	pubkeyDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(pubkeyFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubkeyDER,
	})
	if err != nil {
		return nil, err
	}
	err = pubkeyFile.Close()
	return key, err
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
