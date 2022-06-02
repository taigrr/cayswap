package wg

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/taigrr/cayswap/types"
	"github.com/taigrr/cayswap/wg/parser"
	"github.com/taigrr/systemctl"
)

var restart sync.Mutex
var wgInterface string

func ClientExists(ip string) bool {
	return true
}
func ClientAdd(c types.Request) {

}
func ServerAdd(c types.Request) {

}
func getIP() string {
	return ""
}

func getPubKey() string {
	return ""
}

func ReadConfig() (parser.Config, error) {
	restart.Lock()
	defer restart.Unlock()
	return parser.ParseConfig(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface))
}
func WriteConfig(p parser.Config) {
	restart.Lock()
	defer restart.Unlock()
	os.WriteFile(fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface), []byte(p.String()), 0600)
}

func RestartInterface() {
	restart.Lock()
	defer restart.Unlock()
	systemctl.Restart(context.Background(), fmt.Sprintf("wg-quick@%s", wgInterface), systemctl.Options{})
}

func GenerateReq() types.Request {
	var r types.Request
	r.IPAddr = getIP()
	r.PubKey = getPubKey()
	return r
}
func SetWGDevice(d string) {
	wgInterface = d
}
