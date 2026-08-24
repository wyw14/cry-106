package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wyw14/cry-106/internal/api"
	"github.com/wyw14/cry-106/internal/app"
)

func main() {
	config := app.DefaultConfig()
	flag.StringVar(&config.Address, "addr", config.Address, "HTTP listen address")
	flag.StringVar(&config.DataDir, "data", config.DataDir, "persistent event directory")
	flag.Parse()
	if err := config.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	system, err := app.Open(config.DataDir)
	if err != nil {
		log.Fatalf("open MineAir: %v", err)
	}
	defer system.Close()
	server := api.NewServer(system)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("MineAir listening on %s", config.Address)
	if err := server.ListenAndServe(ctx, config.Address); err != nil {
		log.Fatalf("serve MineAir: %v", err)
	}
}
