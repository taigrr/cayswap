package wg

import (
	"errors"
	"fmt"
	"testing"

	"github.com/taigrr/cayswap/types"
)

func TestKeyErrorMatchesSentinelAndKey(t *testing.T) {
	err := keyErr(KeyPrivate, ErrKeyParse, errors.New("bad base64"))

	if !errors.Is(err, ErrKeyParse) {
		t.Errorf("errors.Is(err, ErrKeyParse) = false, want true")
	}

	var ke *KeyError
	if !errors.As(err, &ke) {
		t.Fatalf("errors.As(err, *KeyError) = false, want true")
	}
	if ke.Key != KeyPrivate {
		t.Errorf("Key = %q, want %q", ke.Key, KeyPrivate)
	}
}

func TestKeyErrorNilCause(t *testing.T) {
	err := keyErr(KeyPublic, ErrKeyEmpty, nil)
	if !errors.Is(err, ErrKeyEmpty) {
		t.Errorf("errors.Is(err, ErrKeyEmpty) = false, want true")
	}
	var ke *KeyError
	if !errors.As(err, &ke) || ke.Key != KeyPublic {
		t.Errorf("expected KeyError with Key=public, got %v", err)
	}
}

func TestKeyErrorWrapsCause(t *testing.T) {
	cause := errors.New("underlying")
	err := keyErr(KeyPrivate, ErrKeyGenerate, cause)
	if !errors.Is(err, cause) {
		t.Errorf("errors.Is(err, cause) = false, want true (cause should remain matchable)")
	}
}

func TestPubKeyInvalidReturnsKeyError(t *testing.T) {
	_, err := pubKey("not-a-valid-key")
	if err == nil {
		t.Fatal("expected error for invalid private key")
	}
	if !errors.Is(err, ErrKeyParse) {
		t.Errorf("errors.Is(err, ErrKeyParse) = false, want true")
	}
	var ke *KeyError
	if !errors.As(err, &ke) || ke.Key != KeyPrivate {
		t.Errorf("expected KeyError{Key: private}, got %v", err)
	}
}

func TestNewPrivKeyValid(t *testing.T) {
	priv, err := NewPrivKey()
	if err != nil {
		t.Fatalf("NewPrivKey: %v", err)
	}
	// A freshly generated private key must be derivable to a public key.
	if _, err := pubKey(priv); err != nil {
		t.Errorf("derive from generated key: %v", err)
	}
}

func TestClientAddRejectsEmptyKey(t *testing.T) {
	err := ClientAdd(types.Request{PubKey: ""})
	if !errors.Is(err, ErrKeyEmpty) {
		t.Errorf("ClientAdd empty key: errors.Is(err, ErrKeyEmpty) = false, want true (got %v)", err)
	}
	var ke *KeyError
	if !errors.As(err, &ke) || ke.Key != KeyPublic {
		t.Errorf("expected KeyError{Key: public}, got %v", err)
	}
}

func TestServerAddRejectsEmptyKey(t *testing.T) {
	err := ServerAdd(types.Request{PubKey: ""}, types.ServerOpts{})
	if !errors.Is(err, ErrKeyEmpty) {
		t.Errorf("ServerAdd empty key: errors.Is(err, ErrKeyEmpty) = false, want true (got %v)", err)
	}
}

func TestKeyErrorString(t *testing.T) {
	err := keyErr(KeyPrivate, ErrKeyEmpty, nil)
	want := fmt.Sprintf("%s key: %v", KeyPrivate, ErrKeyEmpty)
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
