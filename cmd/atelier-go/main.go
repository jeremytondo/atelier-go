// Package main is the client entry point.
package main

import (
	"atelier-go/internal/cli"
	"atelier-go/internal/env"
)

func main() {
	env.Bootstrap()
	cli.Execute()
}
