package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

type RSAKeyManager struct {
	privateKey *rsa.PrivateKey
}

func NewRSAKeyManagerFromBase64(base64Key string) (*RSAKeyManager, error) {
	derBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS8PrivateKey(derBytes)
	if err != nil {
		return nil, err
	}

	return &RSAKeyManager{
		privateKey: key.(*rsa.PrivateKey),
	}, nil
}

func (r *RSAKeyManager) Sign(data string) (string, error) {
	if r.privateKey == nil {
		return "", nil
	}

	hash := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (r *RSAKeyManager) GetPrivateKey() *rsa.PrivateKey {
	return r.privateKey
}
