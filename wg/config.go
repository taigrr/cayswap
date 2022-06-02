package wg

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/wg/parser"
	"github.com/taigrr/systemctl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var restart sync.Mutex
var wgInterface string

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

//TODO move mutex to here not inside reader
func ClientAdd(c types.Request) {
	restart.Lock()
	defer restart.Unlock()
	conf, _ := readConfig()
	p := parser.Peer{}
	p.AllowedIPs = append(p.AllowedIPs, parser.Address(c.IPAddr))
	p.Comment = c.Comment
	p.PublicKey = c.PubKey
	conf.Peers = append(conf.Peers, p)
	writeConfig(conf)
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

//TODO return a proper error here instead of the empty string
func GetPubKey() string {
	c, err := ReadConfig()
	if err != nil {
		return ""
	}
	if c.Interface.PrivateKey == "" {
		return ""
	}
	return pubKey(c.Interface.PrivateKey)
}

func pubKey(priv string) string {
	k, err := wgtypes.ParseKey(priv)
	if err != nil {
		return ""
	}
	return k.PublicKey().String()
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
	defer restart.Unlock()
	systemctl.Restart(context.Background(), fmt.Sprintf("wg-quick@%s", wgInterface), systemctl.Options{})
}

func GenerateReq() types.Request {
	var r types.Request
	r.IPAddr = getIP()
	r.PubKey = GetPubKey()
	r.Comment, _ = os.Hostname()
	return r
}
func SetWGDevice(d string) {
	wgInterface = d
}
