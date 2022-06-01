package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/taigrr/cayswap/auth"
)

type Request struct {
	PubKey string `json:"PubKey"`
	IPAddr string `json:"IPAddr"`
}

func ReceiveKey(w http.ResponseWriter, r *http.Request) {
	var req Request
	clientInterface := r.RemoteAddr
	clientIP := strings.Split(clientInterface, ":")[0]
	log.Printf("Received req from %s\n", clientIP)
	key := r.Header.Get("key")
	if !auth.IsAuthorized(key) {
		fmt.Printf("Invalid key `%s` received, ignoring request!\n", key)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("Error decoding incoming body: %v\n", err)
	}
}
