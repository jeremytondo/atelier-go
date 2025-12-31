// Package main is the client entry point.
package main

import (
	"atelier-go/internal/cli"
	"atelier-go/internal/env"
	"context"
)

func main() {
	env.ConfigurePath()
	cli.Execute(context.Background())
}
