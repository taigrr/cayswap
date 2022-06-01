/*
Copyright © 2022 Tai Groot

*/
package main

import (
	"fmt"
	"os"

	"github.com/taigrr/cayswap/cmd"
	"github.com/taigrr/cayswap/util"
)

func init() {
	if !util.IsRoot() {
		fmt.Println("Error: this program must be run as root to function correctly!")
		os.Exit(1)
	}

}

func main() {
	cmd.Execute()
}
