package util

import (
	"log"
	"os/user"
)

func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get current user: %s", err)
	}
	return currentUser.Uid == "0"
}
