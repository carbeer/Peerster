package utils

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

func HandleError(e error) {
	if e != nil {
		log.Fatal("General error ", e)
		fmt.Println("General error ", e)
		debug.PrintStack()
	}
}

func FlipCoin() bool {
	newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2)
	if newRand == 0 {
		// Continue rumormongering
		return true
	}
	// Stop rumermongering
	return false
}

func TimeoutCounter(channel chan<- bool, frequency string) {
	duration, e := time.ParseDuration(frequency)
	HandleError(e)
	<-time.NewTicker(duration).C
	channel <- true
	close(channel)
}

func MarshalAndWrite(w http.ResponseWriter, msg interface{}) {
	bytes, e := json.Marshal(msg)
	HandleError(e)
	w.Write(bytes)
}

func CheckDataValidity(data []byte, hash []byte) bool {
	var hashFunc = crypto.SHA256.New()
	hashFunc.Write(data)
	expectedHash := hex.EncodeToString(hashFunc.Sum(nil))
	if expectedHash != hex.EncodeToString(hash) {
		log.Printf("DATA IS CORRUPTED! Expected %s and got %s\n", expectedHash, hex.EncodeToString(hash))
		return false
	}
	return true
}

func GetNextDataChunk(file *os.File, step int) []byte {
	bytes := make([]byte, step)
	_, e := file.Read(bytes)
	if e != nil {
		if e == io.EOF {
			return nil
		}
		log.Fatal("Error in next data chunk ", e)
		debug.PrintStack()
	}
	return bytes
}

func GetHashAtIndex(metaFile []byte, index int) []byte {
	begin := 32 * (index)
	end := begin + 32

	if len(metaFile) >= end {
		return metaFile[begin:end]
	}
	return nil
}
