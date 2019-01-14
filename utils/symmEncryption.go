package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
)

// AES encrypt data byte slice using symmetric key
func AESEncrypt(data []byte, key []byte) []byte {
	block, e := aes.NewCipher(key)
	HandleError(e)
	gcm, e := cipher.NewGCM(block)
	HandleError(e)
	nonce := make([]byte, gcm.NonceSize())
	_, e = io.ReadFull(rand.Reader, nonce)
	HandleError(e)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

// AES decrypt data byte slie using symmetric key
func AESDecrypt(data []byte, key []byte) []byte {
	block, e := aes.NewCipher(key)
	HandleError(e)
	gcm, e := cipher.NewGCM(block)
	HandleError(e)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, e := gcm.Open(nil, nonce, ciphertext, nil)
	HandleError(e)
	return plaintext
}

// AES decrypt a locally stored file using the key from Replica r
func AESDecryptFile(r Replica, name string) {
	<-time.After(time.Second) // Wait for deferred file closing
	file, e := os.Open(filepath.Join(".", DOWNLOAD_FOLDER, name))
	HandleError(e)
	file_dec, e := os.Create(filepath.Join(".", DOWNLOAD_FOLDER, "decrypted_"+name))
	HandleError(e)

	fileInfo, e := file.Stat()
	HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(CHUNK_SIZE+AES_MARKUP)))

	for i := 0; i < noChunks; i++ {
		chunk := GetNextDataChunk(file, CHUNK_SIZE+AES_MARKUP)
		chunk = AESDecrypt(chunk, r.EncryptionKey)
		_, e = file_dec.Write(bytes.Trim(chunk, "\x00"))
		HandleError(e)
	}
	file.Close()
	fmt.Printf("DECRYPTED file %s\n", name)
}
