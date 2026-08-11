package auth

import (
	"crypto/sha256"
	"crypto/subtle"
)

var key string

func SetKey(k string) {
	key = k
}

// IsAuthorized checks the given key against the stored key using
// constant-time comparison to prevent timing attacks.
func IsAuthorized(k string) bool {
	if key == "" || k == "" {
		return false
	}
	want := sha256.Sum256([]byte(key))
	got := sha256.Sum256([]byte(k))
	return subtle.ConstantTimeCompare(got[:], want[:]) == 1
}
