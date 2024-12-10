// Package security - пакет для генерации ключей
package security

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"go.uber.org/zap"
)

const (
	PrivKeyFileName string = "private_key.pem"
	PubKeyFileName  string = "public_key.pem"
	label           string = ""
	lengthofKey     int    = 2048
)

func SignSendData(src, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(src)
	return h.Sum(nil)
}

func CompareSign(sign1, sign2 []byte) bool {
	return hmac.Equal(sign1, sign2)
}

func GenerateKeys(logger zap.SugaredLogger) error {
	var err error
	var pemPrivFile, pemPubFile *os.File
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	pemPrivFile, err = os.Create(PrivKeyFileName)
	if err != nil {
		return fmt.Errorf("error when create private pem file %w", err)
	}
	defer func() {
		err = pemPrivFile.Close()
		if err != nil {
			logger.Errorf("error when closing private pem file %w", err)
		}
	}()

	pemPubFile, err = os.Create(PubKeyFileName)
	if err != nil {
		return fmt.Errorf("error when create public pem file %w", err)
	}
	defer func() {
		err = pemPubFile.Close()
		if err != nil {
			logger.Errorf("error when closing public pem file %w", err)
		}
	}()

	privateKey, err = rsa.GenerateKey(rand.Reader, lengthofKey)
	if err != nil {
		return fmt.Errorf("error when generate priv key %w", err)
	}

	publicKey = &privateKey.PublicKey

	blockPriv := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	blockPub := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	}

	if err := pem.Encode(pemPrivFile, blockPriv); err != nil {
		return fmt.Errorf("error when write pem format priv key to file %w", err)
	}

	if err := pem.Encode(pemPubFile, blockPub); err != nil {
		return fmt.Errorf("error when write pem format pib key to file %w", err)
	}
	return nil
}

func ReadPrivedKey(filename string, logger zap.SugaredLogger) (*rsa.PrivateKey, error) {
	var key *rsa.PrivateKey

	KeyFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error when read priv key file %w", err)
	}
	defer func() {
		err = KeyFile.Close()
		if err != nil {
			logger.Errorf("error when closing private key file %w", err)
		}
	}()

	pemfileinfo, _ := KeyFile.Stat()
	pembytes := make([]byte, pemfileinfo.Size())
	buffer := bufio.NewReader(KeyFile)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return nil, fmt.Errorf("error when read priv key file %w", err)
	}
	block, _ := pem.Decode(pembytes)

	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("err when parse priv key %w", err)
	}
	return key, nil
}

func ReadPublicKey(filename string, logger zap.SugaredLogger) (*rsa.PublicKey, error) {
	var key *rsa.PublicKey

	KeyFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error when read priv key file %w", err)
	}
	defer func() {
		err = KeyFile.Close()
		if err != nil {
			logger.Errorf("error when closing public key file %w", err)
		}
	}()

	pemfileinfo, _ := KeyFile.Stat()
	pembytes := make([]byte, pemfileinfo.Size())
	buffer := bufio.NewReader(KeyFile)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return nil, fmt.Errorf("error when read priv key file %w", err)
	}
	block, _ := pem.Decode(pembytes)

	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	key, err = x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("err when parse public key %w", err)
	}
	return key, nil
}

func EncryptMetic(metric []byte, pubKey *rsa.PublicKey) ([]byte, error) {

	rng := rand.Reader

	encMetric, err := rsa.EncryptOAEP(sha256.New(), rng, pubKey, metric, []byte(label))
	if err != nil {
		return nil, fmt.Errorf("error from encryption: %w", err)
	}
	return encMetric, nil
}

func DecryptMetric(buf *bytes.Buffer, privKey *rsa.PrivateKey) error {

	metric, err := rsa.DecryptOAEP(sha256.New(), nil, privKey, buf.Bytes(), []byte(label))
	if err != nil {
		return fmt.Errorf("error from decryption: %w", err)
	}

	// в отчищенную переменную buf записываются распакованные данные
	buf.Reset()
	_, err = buf.ReadFrom(bytes.NewReader(metric))
	if err != nil {
		return fmt.Errorf("error when read decrypt metric %w", err)
	}
	return nil
}
