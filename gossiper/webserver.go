package gossiper

import (
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
				allMsg = allMsg + rm.Text + "\n"
			}
		}
	}
	g.receivedMessagesLock.RUnlock()
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

// Handler functions
func (g *Gossiper) peerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		peer := r.FormValue("peer")
		g.addPeerToListIfApplicable(peer)
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
		message := r.FormValue("message")
		g.ClientMessageHandler(utils.Message{Text: message})

	case http.MethodGet:
		utils.MarshalAndWrite(w, g.GetAllReceivedMessages())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) getId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.MarshalAndWrite(w, g.name)
		return
	}
	utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
}

func (g *Gossiper) indexFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// fileName := r.FormValue("filename")
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) getKnownOrigins(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.MarshalAndWrite(w, g.GetAllOrigins())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) BootstrapUI() {
	log.Println("Starting UI on " + fmt.Sprintf("%s:%s", g.Address.IP, "8080"))
	r := mux.NewRouter()
	r.HandleFunc("/message", g.handleMessage).Methods("POST", "GET")
	r.HandleFunc("/node", g.peerHandler).Methods("POST", "GET")
	r.HandleFunc("/id", g.getId).Methods("GET")
	r.HandleFunc("/origins", g.getKnownOrigins).Methods("GET")
	r.HandleFunc("/file", g.indexFile).Methods("POST")
	r.Handle("/", http.FileServer(http.Dir("webpage"))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("webpage/js/"))))
	utils.HandleError(http.ListenAndServe(fmt.Sprintf("%s:%s", g.Address.IP, "8080"), r))
}
