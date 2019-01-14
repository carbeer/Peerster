package utils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

// Returns boolean value whether Peerster should continue rumormongering with a new peer
func FlipCoin() bool {
	newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2)
	if newRand == 0 {
		return true
	}
	return false
}

func CeilIntDiv(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

func MinUint64(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

// Logging of errors
func HandleError(e error) {
	if e != nil {
		log.Println("General error", e)
		fmt.Println("General error", e)
		debug.PrintStack()
	}
}

func MarshalAndWrite(w http.ResponseWriter, msg interface{}) {
	bytes, e := json.Marshal(msg)
	HandleError(e)
	w.Write(bytes)
}

func GetNextDataChunk(file *os.File, step int) []byte {
	bytes := make([]byte, step)
	_, e := file.Read(bytes)
	if e != nil {
		if e == io.EOF {
			return nil
		}
		fmt.Println("Error in next data chunk", e)
		debug.PrintStack()
	}
	return bytes
}

// Returns hash no. index of the metafile
func GetHashAtIndex(metaFile []byte, index int) []byte {
	begin := HASH_LENGTH * index
	end := begin + HASH_LENGTH

	if len(metaFile) >= end {
		return metaFile[begin:end]
	}
	return nil
}

// The message must be no longer than the length of the public modulus minus twice the hash length, minus a further 2.
func GetMaxEncodedChunkLength(pubKey *rsa.PublicKey) (int, error) {
	ret := int(float64(pubKey.Size())/8 - float64(HASH_ALGO.Size())*2/8 - 2)
	if ret <= 0 {
		log.Printf("pubKey.Size() / 8 - HASH_ALGO.Size() / 8 - 2:\n %d / 8 - %d * 2 / 8 - 2 = %d\n", pubKey.Size(), HASH_ALGO.Size(), ret)
		return 0, errors.New("You cannot encode any message with this setup.")
	}
	return ret, nil
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func ContainsKey(s map[string]interface{}, key string) bool {
	for k, _ := range s {
		if k == key {
			return true
		}
	}
	return false
}

func ContainsValue(s map[interface{}]string, key string) bool {
	for _, v := range s {
		if v == key {
			return true
		}
	}
	return false
}

func GetRandomNonce() (arr [32]byte) {
	token := make([]byte, 32)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(token)
	copy(arr[:], token)
	return
}

func NextNonce(arr [32]byte) [32]byte {
	// Increment the nonce by 1
	for i := 0; i < len(arr); i++ {
		if arr[i] < 255 {
			arr[i]++
			break
		} else {
			arr[i] = 0
		}
	}
	return arr
}

func GenerateRandomByteArr(length int) []byte {
	key := make([]byte, length)
	_, e := rand.Read(key)
	HandleError(e)
	return key
}

func EqualByteArr(arr1 []byte, arr2 []byte) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, _ := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}
