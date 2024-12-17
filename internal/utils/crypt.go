package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func (u *Utils) EncryptPassword(publicKey *string, password string) (string, error) {
	block, _ := pem.Decode([]byte(*publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return "", fmt.Errorf("%w: failed to parse PEM block containing the public key", InvalidPublicKeyError)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("%w: %s", InvalidPublicKeyError, err.Error())
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("%w: not an RSA public key", InvalidPublicKeyError)
	}

	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(password))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
