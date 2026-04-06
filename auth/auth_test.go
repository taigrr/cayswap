package auth

import "testing"

func TestIsAuthorized(t *testing.T) {
	SetKey("secret-key-123")

	if !IsAuthorized("secret-key-123") {
		t.Error("expected correct key to be authorized")
	}
	if IsAuthorized("wrong-key") {
		t.Error("expected wrong key to be rejected")
	}
	if IsAuthorized("") {
		t.Error("expected empty key to be rejected")
	}
	if IsAuthorized("secret-key-1234") {
		t.Error("expected near-miss key to be rejected")
	}
}

func TestIsAuthorizedEmptyStored(t *testing.T) {
	SetKey("")
	if IsAuthorized("anything") {
		t.Error("expected rejection when stored key is empty")
	}
	// Empty vs empty: technically matches, but both empty
	if !IsAuthorized("") {
		t.Error("expected empty==empty to match")
	}
}
