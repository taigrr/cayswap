package util

import "os/user"

// IsRoot reports whether the current process is running as the root user.
// If the current user cannot be determined, it returns false.
func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}
