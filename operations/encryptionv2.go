package operations

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"

	"github.com/robocorp/rcc/common"
	"gopkg.in/square/go-jose.v2"
)

type EncryptionV2 struct {
	*ecdsa.PrivateKey
}

func GenerateEphemeralEccKey() (Ephemeral, error) {
	common.Timeline("start ephemeral key generation")
	defer common.Timeline("done ephemeral key generation")
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &EncryptionV2{key}, nil
}

func (it *EncryptionV2) PublicPEM() (string, error) {
	bytes, err := x509.MarshalPKIXPublicKey(it.Public())
	if err != nil {
		return "", err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte(bytes),
	}
	result := pem.EncodeToMemory(block)
	return string(result), nil
}

func (it *EncryptionV2) RequestObject(payload interface{}) ([]byte, error) {
	result := make(Token)
	encryption := make(Token)
	encryption["scheme"] = "rc-encryption-v2"
	envelope, err := it.PublicPEM()
	if err != nil {
		return nil, err
	}
	encryption["publicKey"] = envelope
	result["encryption"] = encryption
	if payload != nil {
		result["payload"] = payload
	}
	return json.Marshal(result)
}

func (it *EncryptionV2) RequestBody(payload interface{}) (io.Reader, error) {
	blob, err := it.RequestObject(payload)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(blob), nil
}

func (it *EncryptionV2) Decode(blob []byte) ([]byte, error) {
	jwe, err := jose.ParseEncrypted(string(blob))
	if err != nil {
		return nil, err
	}
	payload, err := jwe.Decrypt(it.PrivateKey)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
