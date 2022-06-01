package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Post the key off as a client.",
	//TODO: better docs here
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
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
