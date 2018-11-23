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
		utils.MarshalAndWrite(w, g.getAllRumorMessages())
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
		utils.MarshalAndWrite(w, g.getAllPrivateMessages(keys[0]))
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
