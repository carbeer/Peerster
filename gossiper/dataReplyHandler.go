package gossiper

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carbeer/Peerster/utils"
)

// Sends out a new data reply message
func (g *Gossiper) newDataReplyMessage(msg utils.DataRequest, sender string) {
	dataReplyMessage := utils.DataReply{Origin: g.Name, Destination: msg.Origin, HopLimit: utils.HOPLIMIT_CONSTANT, HashValue: msg.HashValue, Data: g.getStoredChunk(hex.EncodeToString(msg.HashValue))}
	gossipMessage := utils.GossipPacket{DataReply: &dataReplyMessage}

	NextHop := g.getNextHop(dataReplyMessage.Destination)
	if NextHop.HighestID != 0 {
		fmt.Printf("Sending data reply %s to %s via %s\n", hex.EncodeToString(dataReplyMessage.HashValue), dataReplyMessage.Destination, NextHop.Address)
		g.sendToPeer(gossipMessage, NextHop.Address)
	} else {
		fmt.Printf("No next hop to %s. Sending data response for %s back to where it came from: %s\n", dataReplyMessage.Destination, hex.EncodeToString(dataReplyMessage.HashValue), sender)
		g.sendToPeer(gossipMessage, sender)
	}
}

// Handles DataReply messages
func (g *Gossiper) dataReplyHandler(msg utils.DataReply) {
	if msg.Destination == g.Name {
		g.receiveDataReply(msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			return
		}
		gossipMessage := utils.GossipPacket{DataReply: &msg}
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

// Handles incoming DataReplies
func (g *Gossiper) receiveDataReply(msg utils.DataReply) {
	stringHashValue := hex.EncodeToString(msg.HashValue)
	reqChunk := g.getRequestedChunks(stringHashValue)
	// Chunk was not requested
	if reqChunk.FileName == "" {
		fmt.Printf("CHUNK %s was not requested\n", stringHashValue)
		return
	}

	if !CheckDataValidity(msg.Data, msg.HashValue) {
		if string(msg.Data) == "" {
			fmt.Printf("DATA FIELD WAS EMPTY - It seems like the peer doesn't have the requested file. Dropping requests.\n")
			g.sendToDataRequestChannel(stringHashValue, true)
		}
		return
	}

	// Notify process
	g.sendToDataRequestChannel(stringHashValue, true)
	g.addStoredChunk(stringHashValue, msg.Data)
	// Metafile chunk
	if reqChunk.ChunkNr == 0 {
		fmt.Printf("DOWNLOADING metafile of %s from %s\n", reqChunk.FileName, msg.Origin)
		g.onMetaFileReception(msg.Data, msg.HashValue)
		reqChunk.MetaHash = stringHashValue
	} else {
		fmt.Printf("DOWNLOADING %s chunk %d from %s\n", reqChunk.FileName, reqChunk.ChunkNr, msg.Origin)
	}
	// Last chunk of the file
	if reqChunk.NextHash == "" && reqChunk.ChunkNr != 0 {
		g.popRequestedChunks(stringHashValue)
		g.reconstructFile(reqChunk.MetaHash)
		return
	}

	// Chek whether file is supposed to be from one peer or the entire network
	destAvailable := g.getDestinationSpecified(stringHashValue)
	if !destAvailable {
		g.sendDataRequest(utils.Message{Request: g.popRequestedChunks(stringHashValue).NextHash}, g.getChunkHolder(reqChunk.MetaHash, (reqChunk.ChunkNr+1)))
	} else {
		g.sendDataRequest(utils.Message{Destination: msg.Origin, Request: g.popRequestedChunks(stringHashValue).NextHash}, msg.Origin)
	}
}

// Reads out all the required chunk hashes from the metafile
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
		g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, NextHash: chunk, MetaHash: utils.StringHash(hashValue), FileName: file.Name})
		prevChunk = chunk
		counter = counter + 1
	}
	// Last element is metaHash
	g.addRequestedChunks(prevChunk, utils.ChunkInfo{ChunkNr: counter, MetaHash: utils.StringHash(hashValue), FileName: file.Name})
}

// Reconstructs file from all available chunks
func (g *Gossiper) reconstructFile(metaHash string) {
	_ = os.Mkdir(fmt.Sprintf(".%s%s", string(os.PathSeparator), utils.DOWNLOAD_FOLDER), os.ModePerm)
	file, e := os.Create(filepath.Join(".", utils.DOWNLOAD_FOLDER, g.getStoredFile(metaHash).Name))
	utils.HandleError(e)
	defer file.Close()
	metaFile := g.getStoredChunk(metaHash)
	counter := 0
	// Loop through all chunk hashes in the metafile
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
	storedFile.Size = fileInfo.Size()
	g.setStoredFile(metaHash, storedFile)

	fmt.Printf("RECONSTRUCTED file %s\n", storedFile.Name)
	// Notfies channel of successful download
	g.fileDownloadChannel[metaHash] <- true
}

// Check whether the hash and the data align with each other
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
