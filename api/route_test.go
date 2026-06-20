package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taigrr/cayswap/types"
)

func resetAPIHooks() {
	isAuthorized = func(string) bool { return true }
	clientExists = func(string, string) bool { return false }
	clientAdd = func(types.Request) error { return nil }
	restartInterface = func() error { return nil }
	generateReq = func() (types.Request, error) {
		return types.Request{IPAddr: "10.0.0.1/24", PubKey: "server-pub", Comment: "server"}, nil
	}
	reduceIP = func(input string) string { return input }
	marshalJSON = defaultMarshalJSON
}

func TestReceiveKeyRejectsUnauthorizedRequests(t *testing.T) {
	resetAPIHooks()
	isAuthorized = func(string) bool { return false }
	defer resetAPIHooks()

	req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString(`{"PubKey":"client","IPAddr":"10.0.0.2/32"}`))
	req.Header.Set("key", "wrong")
	rec := httptest.NewRecorder()

	ReceiveKey(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestReceiveKeyRejectsInvalidJSON(t *testing.T) {
	resetAPIHooks()
	defer resetAPIHooks()

	called := false
	clientExists = func(string, string) bool {
		called = true
		return false
	}

	req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString(`{"PubKey":`))
	req.Header.Set("key", "ok")
	rec := httptest.NewRecorder()

	ReceiveKey(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Fatal("expected handler to stop before checking existing clients")
	}
}

func TestReceiveKeyAddsClientAndReturnsServerDetails(t *testing.T) {
	resetAPIHooks()
	defer resetAPIHooks()

	added := false
	restarted := make(chan struct{}, 1)
	clientAdd = func(req types.Request) error {
		added = true
		if req.PubKey != "client-pub" || req.IPAddr != "10.0.0.2/32" || req.Comment != "laptop" {
			t.Fatalf("unexpected request payload: %#v", req)
		}
		return nil
	}
	restartInterface = func() error { restarted <- struct{}{}; return nil }
	reduceIP = func(input string) string {
		if input != "10.0.0.1/24" {
			t.Fatalf("unexpected ip to reduce: %q", input)
		}
		return "10.0.0.1/24"
	}

	body := `{"PubKey":"client-pub","IPAddr":"10.0.0.2/32","Comment":"laptop"}`
	req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString(body))
	req.Header.Set("key", "ok")
	rec := httptest.NewRecorder()

	ReceiveKey(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !added {
		t.Fatal("expected client add to be called")
	}
	select {
	case <-restarted:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected interface restart to be triggered")
	}
	const want = `{"PubKey":"server-pub","IPAddr":"10.0.0.1/24","Comment":"server"}`
	if got := rec.Body.String(); got != want {
		t.Fatalf("expected body %s, got %s", want, got)
	}
}

func TestReceiveKeyHandlesClientAddErrors(t *testing.T) {
	resetAPIHooks()
	defer resetAPIHooks()

	clientAdd = func(types.Request) error { return errors.New("boom") }
	req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString(`{"PubKey":"client-pub","IPAddr":"10.0.0.2/32"}`))
	req.Header.Set("key", "ok")
	rec := httptest.NewRecorder()

	ReceiveKey(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestReceiveKeyHandlesMarshalErrors(t *testing.T) {
	resetAPIHooks()
	defer resetAPIHooks()

	marshalJSON = func(any) ([]byte, error) { return nil, errors.New("boom") }
	// clientAdd succeeds, so the restart goroutine is spawned; synchronize on
	// it so the deferred resetAPIHooks doesn't race the hook read.
	restarted := make(chan struct{}, 1)
	restartInterface = func() error { restarted <- struct{}{}; return nil }
	req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString(`{"PubKey":"client-pub","IPAddr":"10.0.0.2/32"}`))
	req.Header.Set("key", "ok")
	rec := httptest.NewRecorder()

	ReceiveKey(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	select {
	case <-restarted:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected interface restart to be triggered")
	}
}
