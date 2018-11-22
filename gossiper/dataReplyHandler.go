package gossiper

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) dataReplyHandler(msg utils.DataReply) {
	if msg.Destination == g.name {
		g.receiveDataReply(msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{DataReply: &msg}
		// fmt.Printf("%d: Send the private message\n", time.Now().Second())
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

func (g *Gossiper) receiveDataReply(msg utils.DataReply) {
	stringHashValue := hex.EncodeToString(msg.HashValue)
	reqChunk := g.getRequestedChunks(stringHashValue)
	// Chunk was not requested
	if reqChunk.FileName == "" {
		fmt.Printf("CHUNK %s was not requested\n", stringHashValue)
		return
	}
	// Corrupted data
	if !CheckDataValidity(msg.Data, msg.HashValue) {
		if string(msg.Data) == "" {
			g.sendToDataRequestChannel(stringHashValue, true)
			fmt.Printf("DATA FIELD WAS EMPTY - It seems like the peer doesn't have the requested file. Dropping requests.\n")
		}
		return
	}
	g.sendToDataRequestChannel(stringHashValue, true)
	g.addStoredChunk(stringHashValue, msg.Data)
	if reqChunk.ChunkNr == 0 {
		fmt.Printf("DOWNLOADING metafile of %s from %s\n", reqChunk.FileName, msg.Origin)
		g.onMetaFileReception(msg.Data, msg.HashValue)
	} else {
		fmt.Printf("DOWNLOADING %s chunk %d from %s\n", reqChunk.FileName, reqChunk.ChunkNr, msg.Origin)
	}
	if reqChunk.MetaHash != "" {
		g.popRequestedChunks(stringHashValue)
		g.reconstructFile(reqChunk.MetaHash)
		return
	}
	g.sendDataRequest(utils.Message{Destination: msg.Origin, Request: g.popRequestedChunks(stringHashValue).NextHash})
}

func (g *Gossiper) onMetaFileReception(metaFile []byte, hashValue []byte) {
	stringHashValue := hex.EncodeToString(hashValue)
	file := g.getStoredFile(stringHashValue)
	prevChunk := stringHashValue

	counter := 0
	for {
		temp := utils.GetHashAtIndex(metaFile, counter)
		if temp == nil {
			break
		}
		chunk := hex.EncodeToString(temp)
		g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, NextHash: chunk, FileName: file.FileName})
		prevChunk = chunk
		counter = counter + 1
	}
	// Last element is metaHash
	g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, MetaHash: hex.EncodeToString(hashValue), FileName: file.FileName})
}

func (g *Gossiper) reconstructFile(metaHash string) {
	file, e := os.Create(fmt.Sprintf(".%s_Downloads%s%s", string(os.PathSeparator), string(os.PathSeparator), g.getStoredFile(metaHash).FileName))
	utils.HandleError(e)
	defer file.Close()
	metaFile := g.getStoredChunk(metaHash)
	counter := 0
	for {
		temp := utils.GetHashAtIndex(metaFile, counter)
		if temp == nil {
			break
		}
		file.Write(bytes.Trim(g.getStoredChunk(hex.EncodeToString(temp)), "\x00"))
		counter = counter + 1
	}
	storedFile := g.getStoredFile(metaHash)
	fileInfo, e := file.Stat()
	utils.HandleError(e)
	storedFile.FileSize = fileInfo.Size()
	g.setStoredFile(metaHash, storedFile)

	fmt.Printf("RECONSTRUCTED file %s\n", storedFile.FileName)
}

func CheckDataValidity(data []byte, hash []byte) bool {
	var hashFunc = crypto.SHA256.New()
	hashFunc.Write(data)
	expectedHash := hex.EncodeToString(hashFunc.Sum(nil))
	if expectedHash != hex.EncodeToString(hash) {
		fmt.Printf("DATA IS CORRUPTED! Expected %s and got %s\n", expectedHash, hex.EncodeToString(hash))
		return false
	}
	return true
}
