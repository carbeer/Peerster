package utils

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func HandleError(e error) {
	if e != nil {
		log.Fatal(e)
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
