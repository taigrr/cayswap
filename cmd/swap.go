package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/taigrr/cayswap/auth"
)

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Post the key off as a client.",
	//TODO: better docs here
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		k := cmd.Flag("auth").Value.String()
		if k == "" {
			log.Fatalf("Error: authorization key is empty!\n")
		}
		auth.SetKey(k)
		k = ""
		fmt.Println("swap called")
		fmt.Printf("Called swap with key: %s\n", cmd.Flag("auth").Value.String())
	},
}

func init() {
	swapCmd.Flags().StringP("server-endpoint", "s", "", "Endpoint (such as 127.0.0.1:5150) to send a keyswap event to")
	swapCmd.Flags().StringP("device", "d", "wg0", "Interface to manage in /etc/wireguard/")
	swapCmd.Flags().Bool("restart", true, "Restart the wg-quick@ interface upon update")
	rootCmd.AddCommand(swapCmd)
}
