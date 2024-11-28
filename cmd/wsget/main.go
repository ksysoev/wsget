package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ksysoev/wsget/pkg/cmd"
)

var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	c := cmd.InitCommands(version)
	if err := c.ExecuteContext(ctx); err != nil {
		cancel()
		os.Exit(1)
	}

	cancel()
}
