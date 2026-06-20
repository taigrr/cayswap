package parser

import (
	"bufio"
	"fmt"
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
	PersistentKeepAlive int
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

// writeField writes "key = val\n" to b when val is non-empty.
func writeField(b *strings.Builder, key, val string) {
	if val == "" {
		return
	}
	b.WriteString(key)
	b.WriteString(" = ")
	b.WriteString(val)
	b.WriteString("\n")
}

func (i Interface) String() string {
	var b strings.Builder
	b.WriteString("[Interface]\n")

	writeField(&b, "PrivateKey", i.PrivateKey)
	writeField(&b, "Address", i.Addresses.String())
	writeField(&b, "DNS", i.DNS.String())
	if i.ListenPort != -1 {
		writeField(&b, "ListenPort", strconv.Itoa(i.ListenPort))
	}
	if i.SaveConfig {
		b.WriteString("SaveConfig = true\n")
	}
	writeField(&b, "Table", i.Table)
	if i.MTU != -1 {
		writeField(&b, "MTU", strconv.Itoa(i.MTU))
	}
	writeField(&b, "PreUp", i.PreUp)
	writeField(&b, "PreDown", i.PreDown)
	writeField(&b, "PostUp", i.PostUp)
	writeField(&b, "PostDown", i.PostDown)

	return b.String()
}

func (p Peer) String() string {
	var b strings.Builder
	b.WriteString("[Peer]")
	if p.Comment != "" {
		b.WriteString(" # ")
		b.WriteString(p.Comment)
	}
	b.WriteString("\n")

	writeField(&b, "PublicKey", p.PublicKey)
	writeField(&b, "AllowedIPs", p.AllowedIPs.String())
	writeField(&b, "Endpoint", p.Endpoint)
	if p.PersistentKeepAlive != -1 {
		writeField(&b, "PersistentKeepalive", strconv.Itoa(p.PersistentKeepAlive))
	}
	writeField(&b, "PresharedKey", p.PresharedKey)

	return b.String()
}

func New() Config {
	var c Config
	c.Interface.ListenPort = -1
	c.Interface.MTU = -1
	return c
}

// NewPeer returns a Peer with sentinel defaults matching the
// serialization logic (PersistentKeepAlive = -1 means omit).
func NewPeer() Peer {
	return Peer{PersistentKeepAlive: -1}
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
		p.AllowedIPs = append(p.AllowedIPs, parseAddresses(val)...)
	case "endpoint":
		p.Endpoint = val
	case "persistentkeepalive":
		parseInt(val, &p.PersistentKeepAlive)
	case "presharedkey":
		p.PresharedKey = val
	}

	return p
}

// ReduceIP returns the network address for the given CIDR (e.g. 10.0.0.5/24
// becomes 10.0.0.0/24). If the input is not valid CIDR it is returned
// unchanged. This is used to advertise the hub's whole subnet to spokes.
func ReduceIP(i string) string {
	_, a, err := net.ParseCIDR(i)
	if err != nil {
		return i
	}
	return a.String()
}

// HostIP returns the address with a single-host mask (/32 for IPv4, /128 for
// IPv6), suitable for an AllowedIPs entry that pins exactly one peer. If the
// input is not valid CIDR it is returned unchanged.
func HostIP(i string) string {
	ip, _, err := net.ParseCIDR(i)
	if err != nil {
		return i
	}
	bits := 32
	if ip.To4() == nil {
		bits = 128
	}
	return fmt.Sprintf("%s/%d", ip.String(), bits)
}

func parseAddress(a string) (Address, error) {
	a = strings.TrimSpace(a)
	// Try CIDR first (e.g. 10.0.0.1/24)
	if _, _, err := net.ParseCIDR(a); err == nil {
		return Address(a), nil
	}
	// Fall back to plain IP (e.g. DNS entries like 1.1.1.1)
	if ip := net.ParseIP(a); ip != nil {
		return Address(a), nil
	}
	return Address(a), &net.ParseError{Type: "address", Text: a}
}

// parseAddresses parses a comma-separated list of addresses, returning only
// the entries that parse successfully.
func parseAddresses(val string) Addresses {
	var out Addresses
	for a := range strings.SplitSeq(val, ",") {
		if address, err := parseAddress(a); err == nil {
			out = append(out, address)
		}
	}
	return out
}

// parseInt sets *dst to the parsed value of val when it is a valid integer,
// leaving *dst untouched otherwise.
func parseInt(val string, dst *int) {
	if n, err := strconv.Atoi(val); err == nil {
		*dst = n
	}
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
		parseInt(val, &i.ListenPort)
	case "address":
		i.Addresses = append(i.Addresses, parseAddresses(val)...)
	case "mtu":
		parseInt(val, &i.MTU)
	case "dns":
		i.DNS = append(i.DNS, parseAddresses(val)...)
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
		i.SaveConfig = strings.EqualFold(val, "true")
	}

	return i
}

func ParseConfig(file string) (Config, error) {
	c := New()
	readFile, err := os.Open(file)
	if err != nil {
		return c, err
	}
	defer func() { _ = readFile.Close() }()
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
			c.Peers = append(c.Peers, NewPeer())
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

	if err := scanner.Err(); err != nil {
		return c, err
	}

	return c, nil
}
