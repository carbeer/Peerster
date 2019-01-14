package gossiper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"math"

	"github.com/carbeer/Peerster/utils"
)

// Encrypt text data using the hex-encoded public key `target` of the recipient
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

// Decrypt text using the private key
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

// Sign private message content
func (g *Gossiper) RSASignPM(msg utils.PrivateMessage) utils.PrivateMessage {
	hash := utils.HASH_ALGO.New()

	bytes, e := json.Marshal(msg)
	utils.HandleError(e)
	hash.Write(bytes)
	hashed := hash.Sum(nil)

	signature, e := rsa.SignPKCS1v15(rand.Reader, &g.privateKey, utils.HASH_ALGO, hashed)
	utils.HandleError(e)
	msg.Signature = signature
	return msg
}

// Verify signature of a private message
func (g *Gossiper) RSAVerifyPMSignature(msg utils.PrivateMessage) bool {
	hash := utils.HASH_ALGO.New()

	bytes, e := json.Marshal(msg)
	utils.HandleError(e)
	hash.Write(bytes)
	hashed := hash.Sum(nil)

	pubKeyBytes, e := hex.DecodeString(msg.Origin)
	utils.HandleError(e)
	pubKey, e := x509.ParsePKCS1PublicKey(pubKeyBytes)
	utils.HandleError(e)

	e = rsa.VerifyPKCS1v15(pubKey, utils.HASH_ALGO, hashed, msg.Signature)
	utils.HandleError(e)
	if e == nil {
		return true
	} else {
		return false
	}
}
