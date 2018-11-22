package gossiper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/carbeer/Peerster/utils"
	"github.com/gorilla/mux"
)

// Helper functions
func (g *Gossiper) GetAllReceivedMessages() string {
	allMsg := ""
	g.receivedMessagesLock.RLock()
	for _, v := range g.ReceivedMessages {
		for _, rm := range v {
			if rm.Text != "" {
				if rm.Origin == g.name {
					allMsg = fmt.Sprintf("%sYOU: %s\n", allMsg, rm.Text)
				} else {
					allMsg = fmt.Sprintf("%s%s: %s\n", allMsg, rm.Origin, rm.Text)
				}
			}
		}
	}
	g.receivedMessagesLock.RUnlock()
	return allMsg
}

func (g *Gossiper) GetAllReceivedPrivateMessages(dest string) string {
	allMsg := ""
	g.privateMessagesLock.RLock()
	for _, v := range g.PrivateMessages {
		for _, rm := range v {
			if rm.Text != "" {
				if rm.Destination == dest {
					allMsg = fmt.Sprintf("%sYOU: %s\n", allMsg, rm.Text)
				} else if rm.Origin == dest {
					allMsg = fmt.Sprintf("%s%s: %s\n", allMsg, rm.Origin, rm.Text)
				}
			}
		}
	}
	g.privateMessagesLock.RUnlock()
	return allMsg
}

func (g *Gossiper) GetAllPeers() string {
	allPeers := ""
	for _, peer := range g.peers {
		allPeers = allPeers + peer + "\n"
	}
	return allPeers
}

func (g *Gossiper) GetAllOrigins() string {
	allOrigins := ""
	g.nextHopLock.RLock()
	for k, _ := range g.nextHop {
		allOrigins = allOrigins + k + "\n"
	}
	g.nextHopLock.RUnlock()
	return allOrigins
}

func (g *Gossiper) unmarshalAndForward(r *http.Request) {
	var msg utils.Message
	e := json.NewDecoder(r.Body).Decode(&msg)
	utils.HandleError(e)
	g.ClientMessageHandler(msg)
}

// Handler functions
func (g *Gossiper) handlePeer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		peers := g.GetAllPeers()
		utils.MarshalAndWrite(w, peers)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		utils.MarshalAndWrite(w, g.GetAllReceivedMessages())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handlePrivateMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		keys, _ := r.URL.Query()["peer"]
		utils.MarshalAndWrite(w, g.GetAllReceivedPrivateMessages(keys[0]))
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.MarshalAndWrite(w, g.name)
		return
	}
	utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
}

func (g *Gossiper) handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleGetOrigin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.MarshalAndWrite(w, g.GetAllOrigins())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleDownload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "webpage/favicon.ico")
}

func (g *Gossiper) BootstrapUI() {
	log.Println("Starting UI on " + fmt.Sprintf("%s:%s", g.Address.IP, utils.GetUIPort()))
	r := mux.NewRouter()
	r.SkipClean(true)
	r.HandleFunc("/message", g.handleMessage).Methods("POST", "GET")
	r.HandleFunc("/privateMessage", g.handlePrivateMessage).Methods("POST", "GET")
	r.HandleFunc("/node", g.handlePeer).Methods("POST", "GET")
	r.HandleFunc("/id", g.handleId).Methods("GET")
	r.HandleFunc("/origins", g.handleGetOrigin).Methods("GET")
	r.HandleFunc("/file", g.handleFile).Methods("POST")
	r.HandleFunc("/download", g.handleDownload).Methods("POST")
	r.HandleFunc("/favicon.ico", serveFavicon)
	r.Handle("/", http.FileServer(http.Dir("webpage/"))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("webpage/js/"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("webpage/static/"))))

	utils.HandleError(http.ListenAndServe(fmt.Sprintf("%s:%s", utils.GetClientIp(), utils.GetUIPort()), r))
}
