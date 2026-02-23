package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/taigrr/cayswap/api"
	"github.com/taigrr/cayswap/auth"
	"github.com/taigrr/cayswap/wg"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the cayswap server to accept key exchange requests",
	Long: `Run this on the hub of your hub-spoke WireGuard architecture. The server
listens for incoming key swap requests from clients, adds their public keys
to the local WireGuard configuration, and returns its own public key. The
server automatically shuts down after 15 minutes for security.`,
	Run: func(cmd *cobra.Command, args []string) {
		k := cmd.Flag("auth").Value.String()
		if k == "" {
			log.Fatalf("Error: authorization key is empty!\n")
		}
		auth.SetKey(k)
		k = ""
		wg.SetWGDevice(cmd.Flag("device").Value.String())
		fmt.Printf("Starting server...\n")
		router := api.NewRouter()

		server := &http.Server{Addr: cmd.Flag("interface").Value.String(), Handler: router}
		go func() {
			time.Sleep(time.Hour / 4)
			server.Shutdown(context.Background())
		}()
		log.Printf("%v\n", server.ListenAndServe())
		log.Printf("Exit!")
	},
}

func init() {
	serveCmd.Flags().StringP("interface", "i", "0.0.0.0:5150", "Interface to use for the API endpoint")
	serveCmd.Flags().StringP("device", "d", "wg0", "Interface to manage in /etc/wireguard/")
	serveCmd.Flags().Bool("restart", true, "Restart the wg-quick@ interface upon update")
	rootCmd.AddCommand(serveCmd)
}
