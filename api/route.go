package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/taigrr/cayswap/auth"
	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/wg"
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

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
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
	clientInterface := r.RemoteAddr
	clientIP := strings.Split(clientInterface, ":")[0]
	log.Printf("Received req from %s\n", clientIP)
	key := r.Header.Get("key")
	if !auth.IsAuthorized(key) {
		fmt.Printf("Invalid key `%s` received, ignoring request!\n", key)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("Error decoding incoming body: %v\n", err)
	}
	if wg.ClientExists(req.IPAddr) {
		log.Printf("Error: Client %s already exists. Ignoring.", req.IPAddr)
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
	wg.ClientAdd(req)
	go wg.RestartInterface()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	req = wg.GetServerReq()
	jr, _ := json.Marshal(req)
	w.WriteHeader(http.StatusOK)
	w.Write(jr)
}
