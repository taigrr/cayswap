package wg

import (
	"errors"
	"fmt"
)

// Key identifies which WireGuard key an operation concerns.
type Key string

const (
	KeyPublic  Key = "public"
	KeyPrivate Key = "private"
)

// Sentinel reasons describing why a key operation failed. They are wrapped by
// KeyError; match them with errors.Is.
var (
	// ErrKeyEmpty indicates a required key was empty.
	ErrKeyEmpty = errors.New("key is empty")

	// ErrKeyParse indicates a key could not be parsed.
	ErrKeyParse = errors.New("invalid key")

	// ErrKeyGenerate indicates a new key could not be generated.
	ErrKeyGenerate = errors.New("could not generate key")
)

// Config operation sentinels. Match with errors.Is.
var (
	// ErrReadConfig wraps failures reading the WireGuard configuration.
	ErrReadConfig = errors.New("reading wireguard config")

	// ErrWriteConfig wraps failures writing the WireGuard configuration.
	ErrWriteConfig = errors.New("writing wireguard config")

	// ErrRestart wraps failures restarting the wg-quick interface.
	ErrRestart = errors.New("restarting wg-quick interface")
)

// KeyError records a failure involving a specific WireGuard key. It wraps a
// sentinel reason (ErrKeyEmpty, ErrKeyParse, ...) so callers can match the
// broad reason with errors.Is or recover the key with errors.As:
//
//	var ke *wg.KeyError
//	if errors.As(err, &ke) && ke.Key == wg.KeyPrivate { ... }
type KeyError struct {
	Key Key
	Err error
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("%s key: %v", e.Key, e.Err)
}

func (e *KeyError) Unwrap() error { return e.Err }

// keyErr builds a KeyError, joining a sentinel reason with an optional
// underlying cause so both remain matchable via errors.Is.
func keyErr(key Key, reason, cause error) *KeyError {
	if cause == nil {
		return &KeyError{Key: key, Err: reason}
	}
	return &KeyError{Key: key, Err: fmt.Errorf("%w: %w", reason, cause)}
}
