package gossiper

import (
	"errors"
	"log"
	"time"

	"github.com/carbeer/Peerster/utils"
)

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
		utils.DecryptFile(r, name)
	}
	delete(g.fileDownloadChannel, r.Metafilehash)
	if err == nil {
		log.Println("Succesfully downloaded the file", name)
	}
	return err
}
