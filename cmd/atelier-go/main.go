// Package main is the client entry point.
package main

import (
	"atelier-go/internal/cli"
	"context"
)

func main() {
	cli.Execute(context.Background())
}
