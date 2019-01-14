package gossiper

import (
	"errors"
	"log"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Downloads a private file, if possible returning the locally stored file,
// otherwise by fetching one of the remote replications.
func (g *Gossiper) DownloadPrivateFile(msg utils.Message) error {
	f := g.getPrivateFile(msg.Request)
	name := msg.FileName
	if len(g.getStoredChunk(utils.StringHash(f.MetafileHash))) != 0 {
		m := utils.Message{FileName: name, Request: utils.StringHash(f.MetafileHash), Destination: g.Name}
		return g.DownloadFile(m)
	} else {
		r := g.getDistributedReplica(f)
		if r.NodeID == "" {
			return errors.New("No distributed version of the private file available.")
		}
		return g.downloadReplica(r, name)
	}
}

// Retrieve replica from peer
func (g *Gossiper) downloadReplica(r utils.Replica, name string) error {
	var err error
	m := utils.Message{FileName: name, Request: r.Metafilehash, Destination: r.NodeID}
	g.fileDownloadChannel[r.Metafilehash] = make(chan bool, 1024)
	g.sendDataRequest(m, m.Destination)
	timeout := time.After(utils.PRIV_FILE_REQ_TIMEOUT)
	select {
	case <-timeout:
		err = errors.New("Timeout, couldn't download file")
		break
	case <-g.fileDownloadChannel[r.Metafilehash]:
		utils.AESDecryptFile(r, name)
	}
	delete(g.fileDownloadChannel, r.Metafilehash)
	if err == nil {
		log.Printf("Succesfully downloaded the file %s, you can find the decrypted file as decrypted_%s\n ", name, name)
	}
	return err
}
