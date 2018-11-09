package gossiper

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/gorilla/mux"
)

// Helper functions

func (g *Gossiper) GetAllReceivedMessages() string {
	allMsg := ""
	g.receivedMessagesLock.RLock()
	for _, v := range g.ReceivedMessages {
		for _, rm := range v {
			log.Println(rm)
			allMsg = allMsg + rm.Text + "\n"
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
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			g.ClientMessageHandler(utils.Message{Text: message})
			wg.Done()
		}()
		wg.Wait()
	case http.MethodGet:
		messages := g.GetAllReceivedMessages()
		utils.MarshalAndWrite(w, messages)
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

func (g *Gossiper) getKnownOrigins(w http.ResponseWriter, r *http.Request) {
	log.Print("blalalsdkjfkjdsf")
}

func (g *Gossiper) BootstrapUI() {
	log.Println("Starting UI on " + fmt.Sprintf("%s:%s", g.Address.IP, "8080"))
	r := mux.NewRouter()
	r.HandleFunc("/message", g.handleMessage).Methods("POST", "GET")
	r.HandleFunc("/node", g.peerHandler).Methods("POST", "GET")
	r.HandleFunc("/id", g.getId).Methods("GET")
	r.HandleFunc("/origins", g.getKnownOrigins).Methods("GET")
	r.Handle("/", http.FileServer(http.Dir("webpage"))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("webpage/js/"))))
	utils.HandleError(http.ListenAndServe(fmt.Sprintf("%s:%s", g.Address.IP, "8080"), r))
}
