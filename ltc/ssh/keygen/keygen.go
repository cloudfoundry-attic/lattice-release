package keygen

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fake_keygen/fake_keygen.go . KeyGenerator
type KeyGenerator interface {
	GenerateRSAPrivateKey(bits int) (pemEncodedPrivateKey string, err error)
	GenerateRSAKeyPair(bits int) (pemEncodedPrivateKey string, authorizedKey string, err error)
}

type keygen struct {
	randReader io.Reader
}

func NewKeyGenerator(randReader io.Reader) KeyGenerator {
	return &keygen{
		randReader: randReader,
	}
}

func (k *keygen) GenerateRSAPrivateKey(bits int) (string, error) {
	pk, err := rsa.GenerateKey(k.randReader, bits)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	})), nil
}

func (k *keygen) GenerateRSAKeyPair(bits int) (pemEncodedPrivateKey string, authorizedKey string, err error) {
	privateKey, err := rsa.GenerateKey(k.randReader, bits)
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
