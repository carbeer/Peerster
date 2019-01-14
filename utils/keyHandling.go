package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/martinlindhe/base36"
)

// Generate RSA key pair
func GenerateKeyPair() *rsa.PrivateKey {
	privKey, e := rsa.GenerateKey(rand.Reader, RSA_KEY_LENGTH)
	HandleError(e)
	StoreKey(privKey)
	return privKey
}

// Return hex encoded public key from private key
func PublicKeyAsString(privKey *rsa.PrivateKey) string {
	pubKeyBytes := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	s := hex.EncodeToString(pubKeyBytes)
	return s
}

// Store private key as file in pem encoded format
func StoreKey(privKey *rsa.PrivateKey) {
	_ = os.Mkdir(KEY_FOLDER, os.ModePerm)
	s := PublicKeyAsString(privKey)
	cwd, _ := os.Getwd()
	f, e := os.Create(filepath.Join(cwd, KEY_FOLDER, fmt.Sprint(HexToBase36(s), ".pem")))
	HandleError(e)
	defer f.Close()

	obj := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	e = pem.Encode(f, obj)
	HandleError(e)
	fmt.Println("Generated and stored new RSA keypair. Public key:", s)
}

// Load private key from file
func LoadKey(pubKey string) (*rsa.PrivateKey, error) {
	cwd, _ := os.Getwd()
	b, e := ioutil.ReadFile(filepath.Join(cwd, KEY_FOLDER, fmt.Sprint(HexToBase36(pubKey), ".pem")))
	if e != nil {
		return nil, e
	}
	block, _ := pem.Decode(b)

	privKey, e := x509.ParsePKCS1PrivateKey(block.Bytes)
	if e != nil {
		return nil, e
	}
	return privKey, privKey.Validate()
}

// Converts hex string to (shorter) base36 encoded string
func HexToBase36(h string) string {
	b, e := hex.DecodeString(h)
	HandleError(e)
	return base36.EncodeBytes(b)
}
