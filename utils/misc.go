package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

func HandleError(e error) {
	if e != nil {
		fmt.Println("General error ", e)
		debug.PrintStack()
	}
}

// Returns boolean value whether Peerster should continue rumormongering with a new peer
func FlipCoin() bool {
	newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2)
	if newRand == 0 {
		return true
	}
	return false
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
		fmt.Println("Error in next data chunk ", e)
		debug.PrintStack()
	}
	return bytes
}

// Returns hash no. index of the metafile
func GetHashAtIndex(metaFile []byte, index int) []byte {
	begin := 32 * (index)
	end := begin + 32

	if len(metaFile) >= end {
		return metaFile[begin:end]
	}
	return nil
}
