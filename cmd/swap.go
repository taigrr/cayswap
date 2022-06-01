package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/taigrr/cayswap/auth"
	"github.com/taigrr/cayswap/wg"
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
		wg.SetWGDevice(cmd.Flag("device").Value.String())
		fmt.Printf("Connecting to Server...\n")
		req := wg.GenerateReq()
		jr, _ := json.Marshal(req)
		url := fmt.Sprintf("http://%s/key", cmd.Flag("server-endpoint").Value.String())
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jr))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		request.Header.Set("key", cmd.Flag("auth").Value.String())
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			log.Printf("Error communicating with the server: %d", response.StatusCode)
			return
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		json.Unmarshal(body, &req)
		wg.ServerAdd(req)
		wg.RestartInterface()
		fmt.Println("Interface swapped!")
	},
}

func init() {
	swapCmd.Flags().StringP("server-endpoint", "s", "", "Endpoint (such as 127.0.0.1:5150) to send a keyswap event to")
	swapCmd.Flags().StringP("device", "d", "wg0", "Interface to manage in /etc/wireguard/")
	swapCmd.Flags().Bool("restart", true, "Restart the wg-quick@ interface upon update")
	rootCmd.AddCommand(swapCmd)
}
