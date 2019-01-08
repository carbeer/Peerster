package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

func AESEncrypt(data []byte, key []byte) []byte {
	log.Println("Length of the data before:", len(data))
	block, e := aes.NewCipher(key)
	HandleError(e)
	gcm, e := cipher.NewGCM(block)
	HandleError(e)
	nonce := make([]byte, gcm.NonceSize())
	_, e = io.ReadFull(rand.Reader, nonce)
	HandleError(e)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	log.Println("Length of the ciphertext:", len(ciphertext))
	return ciphertext
}

func AESDecrypt(data []byte, key []byte) []byte {
	block, e := aes.NewCipher(key)
	HandleError(e)
	gcm, e := cipher.NewGCM(block)
	HandleError(e)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func DecryptFile(r Replica, name string) {
	file, e := os.Open(filepath.Join(".", DOWNLOAD_FOLDER, name))
	HandleError(e)

	fileInfo, e := file.Stat()
	HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(CHUNK_SIZE+AES_MARKUP)))

	for i := 0; i < noChunks; i++ {
		chunk := GetNextDataChunk(file, CHUNK_SIZE+AES_MARKUP)
		chunk = AESDecrypt(chunk, r.EncryptionKey)
		_, e = file.Write(chunk)
		HandleError(e)
	}
	file.Close()
	fmt.Printf("DECRYPTED file %s\n", name)
}
