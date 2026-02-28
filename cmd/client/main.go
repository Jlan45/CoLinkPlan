package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"CoLinkPlan/internal/client"
	"CoLinkPlan/internal/config"
	"CoLinkPlan/pkg/logger"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&configPath, "c", "config.yaml", "Path to config file (shorthand)")
	flag.Parse()

	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		logger.Log.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	logger.Log.Info("Starting Co-Link Client", "config", configPath, "server", cfg.ServerURL, "max_parallel", cfg.MaxParallel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr := client.NewManager(cfg)
	go mgr.Start(ctx)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down client...")
	cancel()
}
