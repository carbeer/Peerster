package gossiper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
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
