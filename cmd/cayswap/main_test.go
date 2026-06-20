package main

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
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

func TestRequireRoot(t *testing.T) {
	err := requireRoot()
	if isRoot := err == nil; isRoot {
		// Running as root: requireRoot must succeed silently.
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
