package gossiper

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/carbeer/Peerster/utils"
)

// Index a public file
func (g *Gossiper) indexFile(msg utils.Message) {
	fmt.Printf("REQUESTING INDEXING filename %s\n", msg.FileName)
	var hashFunc = crypto.SHA256.New()
	var chunkHashed []byte
	file, e := os.Open(filepath.Join(".", utils.SHARED_FOLDER, msg.FileName))
	utils.HandleError(e)

	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(utils.CHUNK_SIZE)))

	// Generate chunk hashes
	for i := 0; i < noChunks; i++ {
		hashFunc.Reset()
		chunk := utils.GetNextDataChunk(file, utils.CHUNK_SIZE)
		hashFunc.Write(chunk)
		temp := hashFunc.Sum(nil)
		g.addStoredChunk(hex.EncodeToString(temp), chunk)
		chunkHashed = append(chunkHashed, temp...)
	}
	file.Close()

	if len(chunkHashed) > utils.CHUNK_SIZE {
		fmt.Printf("CAN'T INDEX FILE %s: METAFILE LARGER THAN ALLOWED\n", msg.FileName)
		return
	}

	hashFunc.Reset()
	hashFunc.Write(chunkHashed)
	metaHash := hex.EncodeToString(hashFunc.Sum(nil))
	g.addStoredChunk(metaHash, chunkHashed)
	fmt.Printf("Indexed File with Metahash %s\n", metaHash)
	g.setStoredFile(utils.StringHash(hashFunc.Sum(nil)), utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(metaHash), Size: fileSize})
	// Publish indexed file
	g.txPublishHandler(utils.TxPublish{File: utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(metaHash), Size: fileSize}, HopLimit: utils.TX_PUBLISH_HOP_LIMIT + 1}, "")
}
