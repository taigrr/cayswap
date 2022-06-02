package parser

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Interface Interface
	Peers     []Peer
}

type Addresses []Address
type Address string

type Interface struct {
	PrivateKey string    // string privatekey
	Addresses  Addresses // may be specified multiple times, comma-separated
	DNS        Addresses // may be specified multiple times, comma-separated
	ListenPort int
	SaveConfig bool
	Table      string
	MTU        int
	PostUp     string
	PostDown   string
	PreUp      string
	PreDown    string
}

type Peer struct {
	Comment             string
	PublicKey           string
	AllowedIPs          Addresses // may be specified multiple times, comma-separated
	Endpoint            string
	PersistentKeepalive int
	PresharedKey        string
}

func (c Config) String() string {
	var b strings.Builder
	b.WriteString(c.Interface.String())
	for _, a := range c.Peers {
		b.WriteString(a.String())
	}
	return b.String()
}

func (a Address) String() string {
	return string(a)
}

func (a Addresses) String() string {
	var addr []string
	for _, x := range a {
		addr = append(addr, x.String())
	}
	return strings.Join(addr, ",")
}

func (i Interface) String() string {
	var b strings.Builder
	b.WriteString("[Interface]\n")

	if i.PrivateKey != "" {
		b.WriteString("PrivateKey = ")
		b.WriteString(i.PrivateKey)
		b.WriteString("\n")
	}

	if i.Addresses.String() != "" {
		b.WriteString("Address = ")
		b.WriteString(i.Addresses.String())
		b.WriteString("\n")
	}

	if i.DNS.String() != "" {
		b.WriteString("DNS = ")
		b.WriteString(i.DNS.String())
		b.WriteString("\n")
	}
	if i.ListenPort != -1 {
		b.WriteString("ListenPort = ")
		b.WriteString(strconv.Itoa(i.ListenPort))
		b.WriteString("\n")
	}
	if i.SaveConfig {
		b.WriteString("SaveConfig = true\n")
	}
	if i.Table != "" {
		b.WriteString("Table = ")
		b.WriteString(i.Table)
		b.WriteString("\n")
	}
	if i.MTU != -1 {
		b.WriteString("MTU = ")
		b.WriteString(strconv.Itoa(i.MTU))
		b.WriteString("\n")
	}
	if i.PreUp != "" {
		b.WriteString("PreUp = ")
		b.WriteString(i.PreUp)
		b.WriteString("\n")
	}
	if i.PreDown != "" {
		b.WriteString("PreDown = ")
		b.WriteString(i.PreDown)
		b.WriteString("\n")
	}
	if i.PostUp != "" {
		b.WriteString("PostUp = ")
		b.WriteString(i.PostUp)
		b.WriteString("\n")
	}
	if i.PostDown != "" {
		b.WriteString("PostDown = ")
		b.WriteString(i.PostDown)
		b.WriteString("\n")
	}

	return b.String()
}

func (p Peer) String() string {
	var b strings.Builder
	b.WriteString("[Peer]")
	if p.Comment != "" {
		b.WriteString(" # " + p.Comment)
	}
	b.WriteString("\n")

	b.WriteString("PublicKey = ")
	b.WriteString(p.PublicKey)
	b.WriteString("\n")

	b.WriteString("AllowedIPs = ")
	b.WriteString(p.AllowedIPs.String())
	b.WriteString("\n")

	if p.Endpoint != "" {
		b.WriteString("Endpoint = ")
		b.WriteString(p.Endpoint)
		b.WriteString("\n")
	}

	if p.PersistentKeepalive != -1 {
		b.WriteString("PersistentKeepalive = ")
		b.WriteString(strconv.Itoa(p.PersistentKeepalive))
		b.WriteString("\n")
	}

	if p.PresharedKey != "" {
		b.WriteString("PresharedKey = ")
		b.WriteString(p.PresharedKey)
		b.WriteString("\n")
	}

	return b.String()
}

func init() {
	// net.CIDRMask
}

func New() Config {
	var c Config
	c.Interface.ListenPort = -1
	c.Interface.MTU = -1
	return c
}

const (
	modeNil = iota
	modeInterface
	modePeer
)

func (p Peer) addLine(line string) Peer {
	splits := strings.SplitN(line, "=", 2)
	key := strings.ToLower(strings.TrimSpace(splits[0]))
	if len(splits) != 2 {
		return p
	}
	val := strings.TrimSpace(splits[1])

	switch key {
	case "publickey":
		p.PublicKey = val
	case "allowedips":
		for _, a := range strings.Split(val, ",") {
			address, err := parseAddress(a)
			if err == nil {
				p.AllowedIPs = append(p.AllowedIPs, address)
			}
		}
	case "endpoint":
		p.Endpoint = val
	case "persistentkeepalive":
		pka, err := strconv.Atoi(val)
		if err == nil {
			p.PersistentKeepalive = pka
		}
	case "presharedkey":
		p.PresharedKey = val
	}

	return p
}

func parseAddress(a string) (Address, error) {
	a = strings.TrimSpace(a)
	_, _, err := net.ParseCIDR(a)
	return Address(a), err
}

func (i Interface) addLine(line string) Interface {
	splits := strings.SplitN(line, "=", 2)
	key := strings.ToLower(strings.TrimSpace(splits[0]))
	if len(splits) != 2 {
		return i
	}
	val := strings.TrimSpace(splits[1])

	switch key {
	case "privatekey":
		i.PrivateKey = val
	case "listenport":
		port, err := strconv.Atoi(val)
		if err == nil {
			i.ListenPort = port
		}
	case "address":
		for _, a := range strings.Split(val, ",") {
			address, err := parseAddress(a)
			if err == nil {
				i.Addresses = append(i.Addresses, address)
			}
		}
	case "mtu":
		mtu, err := strconv.Atoi(val)
		if err != nil {
			i.MTU = mtu
		}
	case "dns":
		for _, a := range strings.Split(val, ",") {
			address, err := parseAddress(a)
			if err == nil {
				i.DNS = append(i.Addresses, address)
			}
		}
	case "table":
		i.Table = val
	case "preup":
		i.PreUp = val
	case "predown":
		i.PreDown = val
	case "postup":
		i.PostUp = val
	case "postdown":
		i.PostDown = val
	case "saveconfig":
		if strings.ToLower(val) == "true" {
			i.SaveConfig = true
		}

	}

	return i
}

func ParseConfig(file string) (Config, error) {
	c := New()
	readFile, err := os.Open(file)
	if err != nil {
		return c, err
	}
	defer readFile.Close()
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)

	mode := modeNil
	peer := -1

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		// Determine what kind of block we're in
		if strings.HasPrefix(line, "[Interface]") {
			mode = modeInterface
		}
		if strings.HasPrefix(line, "[Peer]") {
			mode = modePeer
			peer++
			c.Peers = append(c.Peers, Peer{})
			splits := strings.SplitN(line, "#", 2)
			if len(splits) == 2 {
				c.Peers[peer].Comment = strings.TrimSpace(splits[1])
			}
		}
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch mode {
		case modeNil:
			// normally, we could error out here but let's go for
			// 'happy path'
			continue
		case modeInterface:
			c.Interface = c.Interface.addLine(line)
		case modePeer:
			c.Peers[peer] = c.Peers[peer].addLine(line)
		}
	}

	return c, nil
}
