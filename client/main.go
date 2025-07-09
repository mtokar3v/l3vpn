package main

import (
	"context"
	"l3vpn/client/vpn"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	vpn.Start(ctx)
}
