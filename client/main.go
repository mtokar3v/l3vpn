package main

import (
	"context"
	"os/signal"
	"syscall"

	"l3vpn-client/internal/vpn"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go vpn.Forward(ctx)
	//go vpn.Listen()

	select {}
}
