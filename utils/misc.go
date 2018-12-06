package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

func ByteMetaHash(hash string) []byte {
	res, e := hex.DecodeString(hash)
	HandleError(e)
	return res
}

func StringHash(hash []byte) string {
	return hex.EncodeToString(hash)
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func MinUint64(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}
