package types

import (
	"encoding/json"
	"testing"
)

func TestRequestJSON(t *testing.T) {
	req := Request{
		PubKey:  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
		IPAddr:  "10.0.0.1/24",
		Comment: "test-node",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.PubKey != req.PubKey {
		t.Errorf("PubKey = %q, want %q", decoded.PubKey, req.PubKey)
	}
	if decoded.IPAddr != req.IPAddr {
		t.Errorf("IPAddr = %q, want %q", decoded.IPAddr, req.IPAddr)
	}
	if decoded.Comment != req.Comment {
		t.Errorf("Comment = %q, want %q", decoded.Comment, req.Comment)
	}
}

func TestRequestJSONFieldNames(t *testing.T) {
	data := `{"PubKey":"abc","IPAddr":"10.0.0.1/32","Comment":"node"}`
	var req Request
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if req.PubKey != "abc" {
		t.Errorf("PubKey = %q", req.PubKey)
	}
}
