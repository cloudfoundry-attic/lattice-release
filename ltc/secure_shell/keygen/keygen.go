package keygen

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	"golang.org/x/crypto/ssh"
)

type KeyGenerator struct {
	RandReader io.Reader
}

func (k *KeyGenerator) GenerateRSAPrivateKey(bits int) (string, error) {
	pk, err := rsa.GenerateKey(k.RandReader, bits)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	})), nil
}

func (k *KeyGenerator) GenerateRSAKeyPair(bits int) (pemEncodedPrivateKey string, authorizedKey string, err error) {
	privateKey, err := rsa.GenerateKey(k.RandReader, bits)
	if err != nil {
		return "", "", err
	}

	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privatePEM := pem.EncodeToMemory(privateKeyBlock)

	publicKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		return "", "", err
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	return string(privatePEM), string(publicKeyBytes[:len(publicKeyBytes)-1]), nil
}
