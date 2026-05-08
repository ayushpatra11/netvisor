package main

import (
	"os"

	"github.com/ayushpatra11/netvisor/internal/cli"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0"
var version = "dev"

func main() {
	os.Exit(cli.Run(version))
}
