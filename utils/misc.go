package utils

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

func HandleError(e error) {
	if e != nil {
		log.Println("General error", e)
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
		fmt.Println("Error in next data chunk", e)
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

func FixedByteHash(hash string) (arr [32]byte) {
	b := make([]byte, 32)
	b = ByteMetaHash(hash)
	copy(arr[:], b)
	return
}

func StringHash(hash []byte) string {
	return hex.EncodeToString(hash)
}

func FixedStringHash(hash [32]byte) string {
	return StringHash(hash[:])
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

func ValidateBlockHash(block Block) bool {
	hash := block.Hash()
	for ix := 0; ix < LEADING_ZEROES; ix++ {
		if hash[ix] != 0 {
			return false
		}
	}
	return true
}

func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(StringHash(text)), nil
}

func (h *Hash) UnmarshalText(text []byte) error {
	var s string
	if e := json.Unmarshal(text, &s); e != nil {
		return e
	}
	*h = FixedByteHash(s)
	return nil
}

func CeilIntDiv(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
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
