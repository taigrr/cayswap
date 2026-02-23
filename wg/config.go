package wg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/wg/parser"
	"github.com/taigrr/systemctl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var restart sync.Mutex
var wgInterface string
var needsRestart bool

func ClientExists(key string, ip string) bool {
	c, err := ReadConfig()
	if err != nil {
		return true
	}
	for _, a := range c.Peers {
		if a.PublicKey == key {
			return true
		}
		for _, i := range a.AllowedIPs {
			if ip == i.String() {
				return true
			}
		}
	}
	return false
}

func ClientAdd(c types.Request) error {
	if c.PubKey == "" {
		return errors.New("Error: public key is empty!")
	}
	restart.Lock()
	defer restart.Unlock()
	conf, _ := readConfig()
	p := parser.Peer{}
	for _, ip := range strings.Split(c.IPAddr, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			p.AllowedIPs = append(p.AllowedIPs, parser.Address(ip))
		}
	}
	p.Comment = c.Comment
	p.PublicKey = c.PubKey
	conf.Peers = append(conf.Peers, p)
	writeConfig(conf)
	return nil
}
func ServerAdd(c types.Request, opts types.ServerOpts) {
	restart.Lock()
	defer restart.Unlock()
	conf, _ := readConfig()
	p := parser.Peer{}
	p.AllowedIPs = append(p.AllowedIPs, parser.Address(c.IPAddr))
	p.Comment = c.Comment
	p.PublicKey = c.PubKey
	p.Endpoint = opts.Endpoint
	p.PersistentKeepAlive = opts.PersistentKeepAlive
	p.PresharedKey = opts.PresharedKey
	conf.Peers = append(conf.Peers, p)
	writeConfig(conf)

}
func getIP() string {
	c, err := ReadConfig()
	if err != nil {
		return ""
	}
	return c.Interface.Addresses.String()
}

// GetPubKey returns the public key derived from the interface's private key.
func GetPubKey() (string, error) {
	c, err := ReadConfig()
	if err != nil {
		return "", fmt.Errorf("reading config: %w", err)
	}
	if c.Interface.PrivateKey == "" {
		return "", errors.New("private key is empty")
	}
	key, err := pubKey(c.Interface.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("deriving public key: %w", err)
	}
	return key, nil
}

func pubKey(priv string) (string, error) {
	k, err := wgtypes.ParseKey(priv)
	if err != nil {
		return "", fmt.Errorf("parsing private key: %w", err)
	}
	return k.PublicKey().String(), nil
}

func NewPrivKey() string {
	k, _ := wgtypes.GeneratePrivateKey()
	return k.String()
}
func ReadConfig() (parser.Config, error) {
	restart.Lock()
	defer restart.Unlock()
	return readConfig()
}
func readConfig() (parser.Config, error) {
	return parser.ParseConfig(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface))
}
func writeConfig(p parser.Config) {
	os.WriteFile(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface), []byte(p.String()), 0600)
}
func WriteConfig(p parser.Config) {
	restart.Lock()
	defer restart.Unlock()
	writeConfig(p)
}

func RestartInterface() {
	restart.Lock()
	needsRestart = true
	restart.Unlock()
	time.Sleep(time.Second * 30)
	restart.Lock()
	defer restart.Unlock()
	if needsRestart {
		needsRestart = false
		systemctl.Restart(context.Background(), fmt.Sprintf("wg-quick@%s", wgInterface), systemctl.Options{})
	}
}

func GenerateReq() types.Request {
	var r types.Request
	r.IPAddr = getIP()
	r.PubKey, _ = GetPubKey()
	r.Comment, _ = os.Hostname()
	return r
}
func SetWGDevice(d string) {
	wgInterface = d
}
