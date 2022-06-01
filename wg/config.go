package wg

import (
	"context"
	"fmt"
	"sync"

	"github.com/taigrr/cayswap/types"
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
