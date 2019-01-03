package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func GenerateKeyPair() *rsa.PrivateKey {
	privKey, e := rsa.GenerateKey(rand.Reader, KEY_LENGTH)
	HandleError(e)
	StoreKey(privKey)
	return privKey
}

func PublicKeyAsString(privKey *rsa.PrivateKey) string {
	pubKeyBytes := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	s := base64.URLEncoding.EncodeToString(pubKeyBytes)
	return s
}

func StoreKey(privKey *rsa.PrivateKey) {
	_ = os.Mkdir(KEY_FOLDER, os.ModePerm)
	s := PublicKeyAsString(privKey)
	log.Println(s)
	f, e := os.Create(fmt.Sprint(KEY_FILE, s, ".pem"))
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

func LoadKey(pubKey string) (*rsa.PrivateKey, error) {
	b, e := ioutil.ReadFile(fmt.Sprint(KEY_FILE, pubKey, ".pem"))
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
