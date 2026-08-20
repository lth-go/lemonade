package main

import (
	"fmt"
	"os"

	"github.com/lemonade-command/lemonade/cmd"
	"github.com/lemonade-command/lemonade/lemon"
)

func main() {
	cfg := &lemon.Config{}
	root := cmd.NewRootCmd(cfg)
	err := root.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
