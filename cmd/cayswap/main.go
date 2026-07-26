// cayswap automates the exchange of WireGuard public keys between
// authenticated nodes in a hub-and-spoke topology.
//
// The server (hub) listens for short-lived key-swap requests; clients
// (spokes) post their public key and receive the hub's in return. Both
// sides update their wg-quick configuration and reload the interface.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/taigrr/cayswap/api"
	"github.com/taigrr/cayswap/auth"
	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/util"
	"github.com/taigrr/cayswap/wg"
	"github.com/taigrr/cayswap/wg/parser"
)

// version defaults to the module version embedded by the Go toolchain and can
// be overridden at build time via -ldflags "-X main.version=...".
var version = "devel"

func init() {
	// Only derive from build info when not overridden via ldflags, so
	// release builds that stamp -X main.version win.
	if version != "devel" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			version = v
		}
	}
}

// serverLifetime bounds how long the hub stays up before shutting down,
// keeping the key-exchange window small.
const serverLifetime = 15 * time.Minute

func main() {
	if err := fang.Execute(context.Background(), rootCmd()); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var cfgFile string
	cmd := &cobra.Command{
		Use:   "cayswap",
		Short: "Exchange WireGuard keys automatically and painlessly",
		Long: `cayswap automates swapping WireGuard public keys between authenticated
nodes using a shared key that is only valid for a few minutes at a time.

Run "cayswap serve" on the hub and "cayswap swap" on each spoke. Designed
to be driven from provisioning tooling such as Terraform.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringP("auth", "k", "", "shared authentication key")
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/cayswap/cayswap.yaml)")

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	cmd.AddCommand(serveCmd(), swapCmd(), versionCmd())
	return cmd
}

func initConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/cayswap/")
		viper.SetConfigType("yaml")
		viper.SetConfigName("cayswap")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// authKey resolves the shared key from the --auth flag, falling back to the
// config file / environment via viper.
func authKey(cmd *cobra.Command) string {
	if k := cmd.Flag("auth").Value.String(); k != "" {
		return k
	}
	return viper.GetString("auth")
}

func requiredFlag(cmd *cobra.Command, name string) (string, error) {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func requireRoot() error {
	if !util.IsRoot() {
		return errors.New("cayswap must be run as root to manage WireGuard")
	}
	return nil
}

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the cayswap server to accept key-exchange requests",
		Long: `Run this on the hub of your hub-and-spoke WireGuard topology. The server
listens for incoming key-swap requests from spokes, adds their public keys
to the local WireGuard configuration, and returns its own public key. The
server automatically shuts down after 15 minutes for security.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireRoot(); err != nil {
				return err
			}
			key := authKey(cmd)
			if key == "" {
				return errors.New("authentication key is empty (set --auth or config)")
			}
			auth.SetKey(key)
			wg.SetWGDevice(cmd.Flag("device").Value.String())
			restart, err := cmd.Flags().GetBool("restart")
			if err != nil {
				return err
			}
			api.SetRestartEnabled(restart)

			addr := cmd.Flag("interface").Value.String()
			server := &http.Server{Addr: addr, Handler: api.NewRouter()}

			ctx, cancel := context.WithTimeout(cmd.Context(), serverLifetime)
			defer cancel()
			go func() {
				<-ctx.Done()
				shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
				defer stop()
				if err := server.Shutdown(shutdownCtx); err != nil {
					log.Printf("graceful shutdown failed: %v", err)
				}
			}()

			log.Printf("cayswap listening on %s (auto-stops in %s)", addr, serverLifetime)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("serving: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringP("interface", "i", "0.0.0.0:5150", "address to listen on for the API endpoint")
	cmd.Flags().StringP("device", "d", "wg0", "WireGuard interface to manage in /etc/wireguard/")
	cmd.Flags().Bool("restart", true, "restart the wg-quick@ interface upon update")
	return cmd
}

func swapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap",
		Short: "Exchange WireGuard keys with a cayswap server",
		Long: `Send this node's WireGuard public key to a cayswap server and receive the
server's public key in return. Both nodes update their WireGuard
configuration and restart the interface to establish the tunnel.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireRoot(); err != nil {
				return err
			}
			key := authKey(cmd)
			if key == "" {
				return errors.New("authentication key is empty (set --auth or config)")
			}
			auth.SetKey(key)
			wg.SetWGDevice(cmd.Flag("device").Value.String())

			serverEndpoint, err := requiredFlag(cmd, "server-endpoint")
			if err != nil {
				return err
			}
			wireguardEndpoint, err := requiredFlag(cmd, "wireguard-endpoint")
			if err != nil {
				return err
			}
			req, err := wg.GenerateReq()
			if err != nil {
				return fmt.Errorf("reading local WireGuard config: %w", err)
			}
			// Advertise this spoke as a single host (/32 or /128).
			req.IPAddr = parser.HostIP(req.IPAddr)

			payload, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("encoding request: %w", err)
			}
			url := fmt.Sprintf("http://%s/key", serverEndpoint)
			httpReq, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, url, bytes.NewReader(payload))
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
			httpReq.Header.Set("key", key)

			log.Printf("connecting to server %s", serverEndpoint)
			resp, err := http.DefaultClient.Do(httpReq)
			if err != nil {
				return fmt.Errorf("contacting server: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server returned %s", resp.Status)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}
			if err := json.Unmarshal(body, &req); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
			if err := wg.ServerAdd(req, types.ServerOpts{
				PersistentKeepAlive: 25,
				Endpoint:            wireguardEndpoint,
			}); err != nil {
				return fmt.Errorf("adding server peer: %w", err)
			}
			restart, err := cmd.Flags().GetBool("restart")
			if err != nil {
				return err
			}
			if restart {
				if err := wg.RestartInterface(); err != nil {
					return fmt.Errorf("restarting interface: %w", err)
				}
			}
			fmt.Println("Interface swapped!")
			return nil
		},
	}
	cmd.Flags().StringP("wireguard-endpoint", "w", "", "endpoint (e.g. 127.0.0.1:41574) for WireGuard to connect to")
	cmd.Flags().StringP("server-endpoint", "s", "", "endpoint (e.g. 127.0.0.1:5150) to send the key-swap request to")
	cmd.Flags().StringP("device", "d", "wg0", "WireGuard interface to manage in /etc/wireguard/")
	cmd.Flags().Bool("restart", true, "restart the wg-quick@ interface upon update")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the cayswap version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}
