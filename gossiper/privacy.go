package gossiper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"io"
	"log"
	"math"

	"github.com/carbeer/Peerster/utils"
)

func RSAEncryptText(target string, text string) string {

	textBytes := []byte(text)
	hash := utils.HASH_ALGO.New()
	pubKeyBytes, e := hex.DecodeString(target)
	utils.HandleError(e)
	pubKey, e := x509.ParsePKCS1PublicKey(pubKeyBytes)
	utils.HandleError(e)
	ctext := ""

	chunkSize, e := utils.GetMaxEncodedChunkLength(pubKey)
	utils.HandleError(e)

	for i := 0; i < len(textBytes); {
		j := int(math.Min(float64(len(textBytes)), float64(i+chunkSize)))
		el, e := rsa.EncryptOAEP(hash, rand.Reader, pubKey, textBytes[i:j], nil)
		utils.HandleError(e)
		ctext += string(el)
		i = j
	}
	return ctext
}

func (g *Gossiper) RSADecryptText(ctext string) string {
	hash := utils.HASH_ALGO.New()
	textBytes := []byte(ctext)
	text := ""

	for i := 0; i < len(textBytes); {
		j := int(math.Min(float64(len(textBytes)), float64(i+g.privateKey.Size())))
		textBytes, e := rsa.DecryptOAEP(hash, rand.Reader, &g.privateKey, textBytes[i:j], nil)
		utils.HandleError(e)
		text += string(textBytes)
		i = j
	}
	return text
}

func AESEncrypt(data []byte, key []byte) []byte {
	log.Println("Length of the data before:", len(data))
	block, e := aes.NewCipher(key)
	utils.HandleError(e)
	gcm, e := cipher.NewGCM(block)
	utils.HandleError(e)
	nonce := make([]byte, gcm.NonceSize())
	_, e = io.ReadFull(rand.Reader, nonce)
	utils.HandleError(e)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	log.Println("Length of the ciphertext:", len(ciphertext))
	return ciphertext
}

func (g *Gossiper) AESDecrypt(data []byte, key []byte) []byte {
	block, e := aes.NewCipher(key)
	utils.HandleError(e)
	gcm, e := cipher.NewGCM(block)
	utils.HandleError(e)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
