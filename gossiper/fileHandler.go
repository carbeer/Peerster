package gossiper

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) indexFile(msg utils.Message) {
	fmt.Printf("REQUESTING INDEXING filename %s\n", msg.FileName)
	var hashFunc = crypto.SHA256.New()
	var chunkHashed []byte
	file, e := os.Open(fmt.Sprintf(".%s_SharedFiles%s%s", string(os.PathSeparator), string(os.PathSeparator), msg.FileName))
	utils.HandleError(e)

	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(utils.GetChunkSize())))

	for i := 0; i < noChunks; i++ {
		hashFunc.Reset()
		chunk := utils.GetNextDataChunk(file, utils.GetChunkSize())
		hashFunc.Write(chunk)
		temp := hashFunc.Sum(nil)
		g.addStoredChunk(hex.EncodeToString(temp), chunk)
		chunkHashed = append(chunkHashed, temp...)
	}
	file.Close()

	if len(chunkHashed) > utils.GetChunkSize() {
		fmt.Printf("CAN'T INDEX FILE %s: METAFILE LARGER THAN ALLOWED\n", msg.FileName)
		return
	}

	hashFunc.Reset()
	hashFunc.Write(chunkHashed)
	metaHash := hex.EncodeToString(hashFunc.Sum(nil))
	g.addStoredChunk(metaHash, chunkHashed)
	fmt.Printf("Indexed File with Metahash %s\n", metaHash)
	g.setStoredFile(hex.EncodeToString(hashFunc.Sum(nil)), utils.File{FileName: msg.FileName, MetaHash: metaHash, FileSize: fileSize})
}

func (g *Gossiper) newDataReplyMessage(msg utils.DataRequest, sender string) {
	dataReplyMessage := utils.DataReply{Origin: g.name, Destination: msg.Origin, HopLimit: utils.GetHopLimitConstant(), HashValue: msg.HashValue, Data: g.getStoredChunk(hex.EncodeToString(msg.HashValue))}
	gossipMessage := utils.GossipPacket{DataReply: &dataReplyMessage}

	nextHop := g.getNextHop(dataReplyMessage.Destination)
	if nextHop.HighestID != 0 {
		fmt.Printf("Sending data reply %s to %s via %s\n", hex.EncodeToString(dataReplyMessage.HashValue), dataReplyMessage.Destination, nextHop.Address)
		g.sendToPeer(gossipMessage, nextHop.Address)
	} else {
		fmt.Printf("No next hop to %s. Sending data response for %s back to where it came from: %s\n", dataReplyMessage.Destination, hex.EncodeToString(dataReplyMessage.HashValue), sender)
		g.sendToPeer(gossipMessage, sender)
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
	if !utils.CheckDataValidity(msg.Data, msg.HashValue) {
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
