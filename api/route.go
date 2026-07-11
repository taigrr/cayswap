package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/taigrr/cayswap/auth"
	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/wg"
	"github.com/taigrr/cayswap/wg/parser"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"SendKey",
		strings.ToUpper("Post"),
		"/key",
		ReceiveKey,
	},
}

var (
	isAuthorized     = auth.IsAuthorized
	clientExists     = wg.ClientExists
	clientAdd        = wg.ClientAdd
	restartInterface = wg.RestartInterface
	restartEnabled   = true
	generateReq      = wg.GenerateReq
	reduceIP         = parser.ReduceIP
	marshalJSON      = defaultMarshalJSON
)

// SetRestartEnabled controls whether successful key exchanges reload the
// WireGuard interface after writing a new peer.
func SetRestartEnabled(enabled bool) {
	restartEnabled = enabled
}

func defaultMarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		handler := http.Handler(route.HandlerFunc)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

func ReceiveKey(w http.ResponseWriter, r *http.Request) {
	var req types.Request
	clientIP := remoteHost(r.RemoteAddr)
	log.Printf("Received req from %s", clientIP)
	if !isAuthorized(r.Header.Get("key")) {
		log.Printf("Unauthorized key exchange attempt from %s", clientIP)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding incoming body from %s: %v", clientIP, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if clientExists(req.PubKey, req.IPAddr) {
		log.Printf("Error: Client %s already exists (%s). Ignoring.", req.IPAddr, req.Comment)
		http.Error(w, http.StatusText(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}
	if err := clientAdd(req); err != nil {
		log.Printf("Error adding client %s (%s): %v", req.Comment, req.IPAddr, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("Success: Client %s added for (%s)", req.Comment, req.IPAddr)
	if restartEnabled {
		go func() {
			if err := restartInterface(); err != nil {
				log.Printf("Error restarting interface after adding %s: %v", req.Comment, err)
			}
		}()
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := generateReq()
	if err != nil {
		log.Printf("Error building response for %s: %v", clientIP, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resp.IPAddr = reduceIP(resp.IPAddr)
	jr, err := marshalJSON(resp)
	if err != nil {
		log.Printf("Error encoding response for %s: %v", clientIP, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jr); err != nil {
		log.Printf("Error writing response to %s: %v", clientIP, err)
	}
}

func remoteHost(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
