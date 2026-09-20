package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/taigrr/cayswap/util"
)

func TestRootCmdHasSubcommands(t *testing.T) {
	root := rootCmd()
	want := map[string]bool{"serve": false, "swap": false, "version": false}
	for _, c := range root.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestAuthKeyFromFlag(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	root := rootCmd()
	if err := root.PersistentFlags().Set("auth", "flag-key"); err != nil {
		t.Fatalf("set auth flag: %v", err)
	}
	if got := authKey(root); got != "flag-key" {
		t.Errorf("authKey() = %q, want flag-key", got)
	}
}

func TestAuthKeyFallsBackToViper(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("auth", "viper-key")

	root := rootCmd()
	if got := authKey(root); got != "viper-key" {
		t.Errorf("authKey() = %q, want viper-key", got)
	}
}

func TestRequiredFlag(t *testing.T) {
	cmd := swapCmd()
	if _, err := requiredFlag(cmd, "server-endpoint"); err == nil || !strings.Contains(err.Error(), "server-endpoint is required") {
		t.Fatalf("requiredFlag missing value error = %v, want server-endpoint required", err)
	}

	if err := cmd.Flags().Set("server-endpoint", "hub.example.com:5150"); err != nil {
		t.Fatalf("set server-endpoint: %v", err)
	}
	got, err := requiredFlag(cmd, "server-endpoint")
	if err != nil {
		t.Fatalf("requiredFlag set value: %v", err)
	}
	if got != "hub.example.com:5150" {
		t.Errorf("requiredFlag() = %q, want hub.example.com:5150", got)
	}

	if _, err := requiredFlag(cmd, "missing"); err == nil {
		t.Fatal("expected missing flag error")
	}
}

func TestVersionCmdOutput(t *testing.T) {
	cmd := versionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version cmd: %v", err)
	}
	if got := out.String(); got != version+"\n" {
		t.Errorf("version output = %q, want %q", got, version+"\n")
	}
}

func TestPostKeySwapBuildsAuthorizedRequest(t *testing.T) {
	var gotMethod, gotPath, gotKey, gotContentType, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotKey = r.Header.Get("key")
		gotContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpoint := strings.TrimPrefix(server.URL, "http://")
	resp, err := postKeySwap(context.Background(), endpoint, "shared-key", []byte(`{"PubKey":"client"}`))
	if err != nil {
		t.Fatalf("postKeySwap: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/key" {
		t.Errorf("path = %q, want /key", gotPath)
	}
	if gotKey != "shared-key" {
		t.Errorf("key header = %q, want shared-key", gotKey)
	}
	if gotContentType != "application/json; charset=UTF-8" {
		t.Errorf("content type = %q, want application/json; charset=UTF-8", gotContentType)
	}
	if gotBody != `{"PubKey":"client"}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestKeySwapHTTPClientHasTimeout(t *testing.T) {
	if keySwapHTTPClient.Timeout != keySwapHTTPTimeout {
		t.Fatalf("keySwapHTTPClient timeout = %s, want %s", keySwapHTTPClient.Timeout, keySwapHTTPTimeout)
	}
}

func TestRequireRoot(t *testing.T) {
	err := requireRoot()
	if util.IsRoot() {
		if err != nil {
			t.Fatalf("expected requireRoot to succeed as root, got %v", err)
		}
		return
	}
	if err == nil {
		t.Fatal("expected error when not running as root")
	}
}

func TestServeRequiresAuthKey(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	root := rootCmd()
	root.SetArgs([]string{"serve"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected serve to fail without root or auth key")
	}
}
