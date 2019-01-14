package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// Generate hash from Block
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

// Generate hash from TxPublish
func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}

// Checks whether blockhahs has enough leading zeroes for PoW
func ValidateBlockHash(block Block) bool {
	hash := block.Hash()
	for ix := 0; ix < LEADING_ZEROES; ix++ {
		if hash[ix] != 0 {
			return false
		}
	}
	return true
}

// Convert hex string to byte slice
func ByteMetaHash(hash string) []byte {
	res, e := hex.DecodeString(hash)
	HandleError(e)
	return res
}

// Convert hex string to byte array
func FixedByteHash(hash string) (arr [32]byte) {
	b := make([]byte, 32)
	b = ByteMetaHash(hash)
	copy(arr[:], b)
	return
}

// Convert byte hash to hex string (slice)
func StringHash(hash []byte) string {
	return hex.EncodeToString(hash)
}

// Convert byte hash to hex string (array)
func FixedStringHash(hash [32]byte) string {
	return StringHash(hash[:])
}
