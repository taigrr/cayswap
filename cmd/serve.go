package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/taigrr/cayswap/auth"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "begin listening for keyswap requests",
	//TODO: Add better docs here
	Long: `Run this on the hub of your hub-spoke architecture.`,
	Run: func(cmd *cobra.Command, args []string) {
		// spin up webserver here
		k := cmd.Flag("auth").Value.String()
		if k == "" {
			log.Fatalf("Error: authorization key is empty!\n")
		}
		auth.SetKey(k)
		k = ""
		fmt.Println("serve called")
	},
}

func init() {
	serveCmd.Flags().StringP("interface", "i", "0.0.0.0:5150", "Interface to use for the API endpoint")
	serveCmd.Flags().StringP("device", "d", "wg0", "Interface to manage in /etc/wireguard/")
	serveCmd.Flags().Bool("restart", true, "Restart the wg-quick@ interface upon update")
	rootCmd.AddCommand(serveCmd)
}
