package auth

import "crypto/subtle"

var key string

func SetKey(k string) {
	key = k
}

// IsAuthorized checks the given key against the stored key using
// constant-time comparison to prevent timing attacks.
func IsAuthorized(k string) bool {
	return subtle.ConstantTimeCompare([]byte(k), []byte(key)) == 1
}
