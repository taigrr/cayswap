package parser

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testConfig = `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.0.0.1/24
ListenPort = 51820
DNS = 1.1.1.1, 8.8.8.8
MTU = 1420
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT

[Peer] # server-node
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.0.0.2/32
Endpoint = 203.0.113.1:51820
PersistentKeepalive = 25

[Peer] # client-node
PublicKey = TrMvSoP4jYQlY6RIzBgbssQqY3vxI2piVFBs2LMkNwQ=
AllowedIPs = 10.0.0.3/32, 10.0.0.4/32
`

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "wg0.conf")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}

func TestParseConfigInterface(t *testing.T) {
	path := writeTestConfig(t, testConfig)
	cfg, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}

	iface := cfg.Interface
	if iface.PrivateKey != "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=" {
		t.Errorf("PrivateKey = %q", iface.PrivateKey)
	}
	if iface.Addresses.String() != "10.0.0.1/24" {
		t.Errorf("Addresses = %q", iface.Addresses.String())
	}
	if iface.ListenPort != 51820 {
		t.Errorf("ListenPort = %d, want 51820", iface.ListenPort)
	}
	if iface.MTU != 1420 {
		t.Errorf("MTU = %d, want 1420", iface.MTU)
	}
	if iface.PostUp != "iptables -A FORWARD -i wg0 -j ACCEPT" {
		t.Errorf("PostUp = %q", iface.PostUp)
	}
	if iface.PostDown != "iptables -D FORWARD -i wg0 -j ACCEPT" {
		t.Errorf("PostDown = %q", iface.PostDown)
	}
}

func TestParseConfigDNS(t *testing.T) {
	path := writeTestConfig(t, testConfig)
	cfg, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}

	dns := cfg.Interface.DNS
	if len(dns) != 2 {
		t.Fatalf("DNS count = %d, want 2", len(dns))
	}
	if dns[0].String() != "1.1.1.1" {
		t.Errorf("DNS[0] = %q, want 1.1.1.1", dns[0])
	}
	if dns[1].String() != "8.8.8.8" {
		t.Errorf("DNS[1] = %q, want 8.8.8.8", dns[1])
	}
}

func TestParseConfigPeers(t *testing.T) {
	path := writeTestConfig(t, testConfig)
	cfg, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}

	if len(cfg.Peers) != 2 {
		t.Fatalf("peer count = %d, want 2", len(cfg.Peers))
	}

	p0 := cfg.Peers[0]
	if p0.Comment != "server-node" {
		t.Errorf("Peer[0].Comment = %q, want server-node", p0.Comment)
	}
	if p0.PublicKey != "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=" {
		t.Errorf("Peer[0].PublicKey = %q", p0.PublicKey)
	}
	if p0.AllowedIPs.String() != "10.0.0.2/32" {
		t.Errorf("Peer[0].AllowedIPs = %q", p0.AllowedIPs.String())
	}
	if p0.Endpoint != "203.0.113.1:51820" {
		t.Errorf("Peer[0].Endpoint = %q", p0.Endpoint)
	}
	if p0.PersistentKeepAlive != 25 {
		t.Errorf("Peer[0].PersistentKeepAlive = %d, want 25", p0.PersistentKeepAlive)
	}

	p1 := cfg.Peers[1]
	if p1.Comment != "client-node" {
		t.Errorf("Peer[1].Comment = %q", p1.Comment)
	}
	if len(p1.AllowedIPs) != 2 {
		t.Fatalf("Peer[1].AllowedIPs count = %d, want 2", len(p1.AllowedIPs))
	}
}

func TestParseConfigRoundTrip(t *testing.T) {
	path := writeTestConfig(t, testConfig)
	cfg, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}

	output := cfg.String()

	// Write the output and re-parse to check round-trip stability
	path2 := writeTestConfig(t, output)
	cfg2, err := ParseConfig(path2)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	output2 := cfg2.String()
	if output != output2 {
		t.Errorf("round-trip mismatch:\n--- first ---\n%s\n--- second ---\n%s", output, output2)
	}
}

func TestParseConfigMissingFile(t *testing.T) {
	_, err := ParseConfig("/nonexistent/wg0.conf")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseConfigScannerError(t *testing.T) {
	path := writeTestConfig(t, "[Interface]\n# "+strings.Repeat("x", bufio.MaxScanTokenSize+1))
	_, err := ParseConfig(path)
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected scanner token too long error, got %v", err)
	}
}

func TestParseConfigEmpty(t *testing.T) {
	path := writeTestConfig(t, "")
	cfg, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig empty: %v", err)
	}
	if len(cfg.Peers) != 0 {
		t.Errorf("expected no peers, got %d", len(cfg.Peers))
	}
	if cfg.Interface.ListenPort != -1 {
		t.Errorf("default ListenPort = %d, want -1", cfg.Interface.ListenPort)
	}
	if cfg.Interface.MTU != -1 {
		t.Errorf("default MTU = %d, want -1", cfg.Interface.MTU)
	}
}

func TestNewDefaults(t *testing.T) {
	cfg := New()
	if cfg.Interface.ListenPort != -1 {
		t.Errorf("New().Interface.ListenPort = %d, want -1", cfg.Interface.ListenPort)
	}
	if cfg.Interface.MTU != -1 {
		t.Errorf("New().Interface.MTU = %d, want -1", cfg.Interface.MTU)
	}
}

func TestNewPeerDefaults(t *testing.T) {
	p := NewPeer()
	if p.PersistentKeepAlive != -1 {
		t.Errorf("NewPeer().PersistentKeepAlive = %d, want -1", p.PersistentKeepAlive)
	}
}

func TestPeerStringOmitsPersistentKeepaliveWhenMinusOne(t *testing.T) {
	p := NewPeer()
	p.PublicKey = "testkey"
	p.AllowedIPs = Addresses{Address("10.0.0.1/32")}
	output := p.String()
	if strings.Contains(output, "PersistentKeepalive") {
		t.Errorf("NewPeer().String() should not contain PersistentKeepalive, got:\n%s", output)
	}
}

func TestReduceIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"10.0.0.5/24", "10.0.0.0/24"},
		{"192.168.1.100/32", "192.168.1.100/32"},
		{"10.0.0.1/8", "10.0.0.0/8"},
		{"not-cidr", "not-cidr"},
		{"", ""},
	}
	for _, tc := range tests {
		got := ReduceIP(tc.input)
		if got != tc.want {
			t.Errorf("ReduceIP(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestHostIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"10.0.0.5/24", "10.0.0.5/32"},
		{"10.0.0.5/16", "10.0.0.5/32"},
		{"192.168.1.100/32", "192.168.1.100/32"},
		{"fd00::1/64", "fd00::1/128"},
		{"not-cidr", "not-cidr"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := HostIP(tc.input); got != tc.want {
			t.Errorf("HostIP(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseAddress(t *testing.T) {
	// CIDR notation should work
	addr, err := parseAddress("10.0.0.1/32")
	if err != nil {
		t.Errorf("parseAddress(CIDR) error: %v", err)
	}
	if addr.String() != "10.0.0.1/32" {
		t.Errorf("parseAddress(CIDR) = %q", addr)
	}

	// Plain IP (for DNS) should also work
	addr, err = parseAddress("1.1.1.1")
	if err != nil {
		t.Errorf("parseAddress(plain IP) should succeed: %v", err)
	}
	if addr.String() != "1.1.1.1" {
		t.Errorf("parseAddress(plain IP) = %q", addr)
	}
}

func TestInterfaceStringSaveConfig(t *testing.T) {
	cfg := New()
	cfg.Interface.PrivateKey = "testkey"
	cfg.Interface.SaveConfig = true
	output := cfg.Interface.String()
	if !strings.Contains(output, "SaveConfig = true") {
		t.Error("expected SaveConfig = true in output")
	}
}

func TestInterfaceStringTable(t *testing.T) {
	cfg := New()
	cfg.Interface.PrivateKey = "testkey"
	cfg.Interface.Table = "auto"
	output := cfg.Interface.String()
	if !strings.Contains(output, "Table = auto") {
		t.Error("expected Table = auto in output")
	}
}

func TestPeerStringWithEndpoint(t *testing.T) {
	p := NewPeer()
	p.PublicKey = "testkey"
	p.AllowedIPs = Addresses{Address("10.0.0.1/32")}
	p.Endpoint = "example.com:51820"
	output := p.String()
	if !strings.Contains(output, "Endpoint = example.com:51820") {
		t.Errorf("expected Endpoint in output, got:\n%s", output)
	}
}

func TestPeerStringWithPresharedKey(t *testing.T) {
	p := NewPeer()
	p.PublicKey = "testkey"
	p.AllowedIPs = Addresses{Address("10.0.0.1/32")}
	p.PresharedKey = "pskvalue"
	output := p.String()
	if !strings.Contains(output, "PresharedKey = pskvalue") {
		t.Errorf("expected PresharedKey in output, got:\n%s", output)
	}
}

func TestConfigStringMultiplePeers(t *testing.T) {
	cfg := New()
	cfg.Interface.PrivateKey = "privkey"
	cfg.Interface.Addresses = Addresses{Address("10.0.0.1/24")}

	p1 := NewPeer()
	p1.PublicKey = "pubkey1"
	p1.AllowedIPs = Addresses{Address("10.0.0.2/32")}
	p1.Comment = "peer-one"

	p2 := NewPeer()
	p2.PublicKey = "pubkey2"
	p2.AllowedIPs = Addresses{Address("10.0.0.3/32")}
	p2.PersistentKeepAlive = 25

	cfg.Peers = []Peer{p1, p2}
	output := cfg.String()

	if !strings.Contains(output, "[Peer] # peer-one") {
		t.Error("expected peer-one comment")
	}
	if !strings.Contains(output, "PersistentKeepalive = 25") {
		t.Error("expected PersistentKeepalive = 25")
	}
	if strings.Count(output, "[Peer]") != 2 {
		t.Errorf("expected 2 [Peer] sections, got %d", strings.Count(output, "[Peer]"))
	}
}
