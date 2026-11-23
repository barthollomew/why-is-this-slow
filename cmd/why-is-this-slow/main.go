package main

import (
	"context"
	"os"

	"github.com/barthollomew/why-is-this-slow/internal/cli"
)

func main() {
	os.Exit(cli.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
