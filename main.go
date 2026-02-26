package main

import "github.com/danielmrdev/dtasks/cmd"

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	cmd.Execute(version)
}
