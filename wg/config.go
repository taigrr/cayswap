package wg

import (
	"context"
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
		return keyErr(KeyPublic, ErrKeyEmpty, nil)
	}
	restart.Lock()
	defer restart.Unlock()
	conf, err := readConfig()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrReadConfig, err)
	}
	p := parser.NewPeer()
	for ip := range strings.SplitSeq(c.IPAddr, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			p.AllowedIPs = append(p.AllowedIPs, parser.Address(ip))
		}
	}
	p.Comment = c.Comment
	p.PublicKey = c.PubKey
	conf.Peers = append(conf.Peers, p)
	if err := writeConfig(conf); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteConfig, err)
	}
	return nil
}

func ServerAdd(c types.Request, opts types.ServerOpts) error {
	if c.PubKey == "" {
		return keyErr(KeyPublic, ErrKeyEmpty, nil)
	}
	restart.Lock()
	defer restart.Unlock()
	conf, err := readConfig()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrReadConfig, err)
	}
	p := parser.NewPeer()
	p.AllowedIPs = append(p.AllowedIPs, parser.Address(c.IPAddr))
	p.Comment = c.Comment
	p.PublicKey = c.PubKey
	p.Endpoint = opts.Endpoint
	p.PersistentKeepAlive = opts.PersistentKeepAlive
	p.PresharedKey = opts.PresharedKey
	conf.Peers = append(conf.Peers, p)
	if err := writeConfig(conf); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteConfig, err)
	}
	return nil
}
func getIP() (string, error) {
	c, err := ReadConfig()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadConfig, err)
	}
	return c.Interface.Addresses.String(), nil
}

// GetPubKey returns the public key derived from the interface's private key.
func GetPubKey() (string, error) {
	c, err := ReadConfig()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadConfig, err)
	}
	if c.Interface.PrivateKey == "" {
		return "", keyErr(KeyPrivate, ErrKeyEmpty, nil)
	}
	key, err := pubKey(c.Interface.PrivateKey)
	if err != nil {
		return "", err
	}
	return key, nil
}

func pubKey(priv string) (string, error) {
	k, err := wgtypes.ParseKey(priv)
	if err != nil {
		return "", keyErr(KeyPrivate, ErrKeyParse, err)
	}
	return k.PublicKey().String(), nil
}

func NewPrivKey() (string, error) {
	k, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", keyErr(KeyPrivate, ErrKeyGenerate, err)
	}
	return k.String(), nil
}
func ReadConfig() (parser.Config, error) {
	restart.Lock()
	defer restart.Unlock()
	return readConfig()
}
func readConfig() (parser.Config, error) {
	return parser.ParseConfig(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface))
}
func writeConfig(p parser.Config) error {
	return os.WriteFile(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface), []byte(p.String()), 0600)
}

func WriteConfig(p parser.Config) error {
	restart.Lock()
	defer restart.Unlock()
	return writeConfig(p)
}

// RestartInterface debounces interface restarts: concurrent callers within a
// 30s window collapse into a single wg-quick restart. It returns any error
// from the restart it actually performs (callers that lose the debounce race
// return nil).
func RestartInterface() error {
	restart.Lock()
	needsRestart = true
	restart.Unlock()
	time.Sleep(time.Second * 30)
	restart.Lock()
	defer restart.Unlock()
	if needsRestart {
		needsRestart = false
		if err := systemctl.Restart(context.Background(), fmt.Sprintf("wg-quick@%s", wgInterface), systemctl.Options{}); err != nil {
			return fmt.Errorf("%w: %w", ErrRestart, err)
		}
	}
	return nil
}

// GenerateReq builds a key-exchange request describing this node: its
// interface address, derived public key, and hostname (best-effort, used only
// as a peer comment).
func GenerateReq() (types.Request, error) {
	var r types.Request
	ip, err := getIP()
	if err != nil {
		return r, err
	}
	r.IPAddr = ip
	pub, err := GetPubKey()
	if err != nil {
		return r, err
	}
	r.PubKey = pub
	r.Comment, _ = os.Hostname()
	return r, nil
}
func SetWGDevice(d string) {
	wgInterface = d
}
