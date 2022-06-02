package parser

import (
	"net"
	"strconv"
	"strings"
)

type Config struct {
	Interface Interface
	Peers     []Peer
}

type Addresses []net.IPNet

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

	b.WriteString("PrivateKey = ")
	b.WriteString(i.PrivateKey)
	b.WriteString("\n")

	if i.Addresses.String() != "" {
		b.WriteString("Addresses = ")
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

	if p.Endpoint != "" {
		b.WriteString("PersistentKeepalive = ")
		b.WriteString(strconv.Itoa(p.PersistentKeepalive))
		b.WriteString("\n")
	}

	if p.Endpoint != "" {
		b.WriteString("PresharedKey = ")
		b.WriteString(p.PresharedKey)
		b.WriteString("\n")
	}

	return b.String()
}

func init() {
	// net.CIDRMask
}
