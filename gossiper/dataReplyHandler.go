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
			fmt.Printf("DATA FIELD WAS EMPTY - It seems like the peer doesn't have the requested file. Dropping requests.\n")
			g.sendToDataRequestChannel(stringHashValue, true)
		}
		return
	}
	destAvailable := g.getDesintationSpecified(reqChunk.MetaHash)

	g.sendToDataRequestChannel(stringHashValue, true)
	g.addStoredChunk(stringHashValue, msg.Data)
	if reqChunk.ChunkNr == 0 {
		fmt.Printf("DOWNLOADING metafile of %s from %s\n", reqChunk.FileName, msg.Origin)
		g.onMetaFileReception(msg.Data, msg.HashValue)
		reqChunk.MetaHash = stringHashValue
	} else {
		fmt.Printf("DOWNLOADING %s chunk %d from %s\n", reqChunk.FileName, reqChunk.ChunkNr, msg.Origin)
	}
	if reqChunk.NextHash == "" && reqChunk.ChunkNr != 0 {
		g.popRequestedChunks(stringHashValue)
		g.reconstructFile(reqChunk.MetaHash)
		return
	}
	if !destAvailable {
		g.sendDataRequest(utils.Message{Request: g.popRequestedChunks(stringHashValue).NextHash}, g.getChunkHolder(reqChunk.MetaHash, (reqChunk.ChunkNr+1)))
	} else {
		g.sendDataRequest(utils.Message{Destination: msg.Origin, Request: g.popRequestedChunks(stringHashValue).NextHash}, msg.Origin)
	}
}

func (g *Gossiper) onMetaFileReception(metaFile []byte, hashValue []byte) {
	file := g.getStoredFile(utils.StringHash(hashValue))
	prevChunk := utils.StringHash(hashValue)

	counter := 0
	for {
		temp := utils.GetHashAtIndex(metaFile, counter)
		if temp == nil {
			break
		}
		chunk := utils.StringHash(temp)
		g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, NextHash: chunk, MetaHash: utils.StringHash(hashValue), FileName: file.FileName})
		prevChunk = chunk
		counter = counter + 1
	}
	// Last element is metaHash
	g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, MetaHash: utils.StringHash(hashValue), FileName: file.FileName})
}

func (g *Gossiper) reconstructFile(metaHash string) {
	file, e := os.Create(fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), utils.GetDownloadFolder(), string(os.PathSeparator), g.getStoredFile(metaHash).FileName))
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
