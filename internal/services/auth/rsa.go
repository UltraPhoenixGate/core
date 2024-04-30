package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	GenerateKeyPair()
}

func GetPublicKey() *rsa.PublicKey {
	file, err := os.ReadFile("config/public.pem")
	if err != nil {
		logrus.WithError(err).Error("Failed to read public key file")
		return nil
	}

	block, _ := pem.Decode(file)
	if block == nil {
		logrus.Error("Failed to decode public key")
		return nil
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse public key")
		return nil
	}

	return pub.(*rsa.PublicKey)
}

func GetPrivateKey() *rsa.PrivateKey {
	file, err := os.ReadFile("config/private.pem")
	if err != nil {
		logrus.WithError(err).Error("Failed to read private key file")
		return nil
	}

	block, _ := pem.Decode(file)
	if block == nil {
		logrus.Error("Failed to decode private key")
		return nil
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse private key")
		return nil
	}

	return priv
}

func isKeyPairExist() bool {
	_, err := os.ReadFile("config/private.pem")
	if err != nil {
		return false
	}
	_, err = os.ReadFile("config/public.pem")
	return err == nil
}

func GenerateKeyPair() {
	// make sure config directory exists
	err := os.MkdirAll("config", 0755)
	if err != nil {
		logrus.WithError(err).Error("Failed to create config directory")
		return
	}

	if isKeyPairExist() {
		logrus.Debug("Key pair already exists, skipping...")
		return
	}

	logrus.Info("Generating RSA key pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate key pair")
		return
	}

	// To Pem
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal public key")
		return
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Save to file
	err = os.WriteFile("config/private.pem", privateKeyPEM, 0644)
	if err != nil {
		logrus.WithError(err).Error("Failed to write private key to file")
		return
	}

	err = os.WriteFile("config/public.pem", publicKeyPEM, 0644)
	if err != nil {
		logrus.WithError(err).Error("Failed to write public key to file")
		return
	}

	logrus.Info("Key pair generated successfully")
}
