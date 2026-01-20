package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	x402relayer "github.com/sei-protocol/x402-relayer"
	"github.com/sei-protocol/x402-relayer/config"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if !cfg.Enabled {
		log.Println("x402-relayer is disabled, exiting")
		return
	}

	// Create server
	server, err := x402relayer.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("received shutdown signal")
		cancel()
	}()

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("starting x402-relayer on port %d", cfg.Port)
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown or error
	select {
	case <-ctx.Done():
		log.Println("shutting down...")
		if err := server.Stop(context.Background()); err != nil {
			log.Printf("error during shutdown: %v", err)
		}
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	}

	log.Println("x402-relayer stopped")
}

func loadConfig(path string) (*config.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	// Set defaults
	v.SetDefault("x402-relayer.enabled", false)
	v.SetDefault("x402-relayer.port", config.DefaultPort)
	v.SetDefault("x402-relayer.usdt_precompile", config.DefaultUSDTPrecompile)
	v.SetDefault("x402-relayer.usdt_denom", config.DefaultUSDTDenom)
	v.SetDefault("x402-relayer.relay_fee_per_tx", config.DefaultRelayFee)
	v.SetDefault("x402-relayer.evm_rpc", config.DefaultEVMRPC)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is okay, use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return config.ReadConfig(v)
}

