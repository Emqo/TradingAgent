package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emqo/TradingAgent/config"
	"github.com/Emqo/TradingAgent/internal/agent"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse interval
	interval, err := time.ParseDuration(cfg.Agent.Interval)
	if err != nil {
		log.Fatalf("Invalid interval %q: %v", cfg.Agent.Interval, err)
	}

	// Get LLM provider config
	providerCfg, err := cfg.LLM.GetProvider()
	if err != nil {
		log.Fatalf("Failed to get provider config: %v", err)
	}

	// Create LLM provider
	llmProvider := llm.NewClaudeProvider(
		providerCfg.BaseURL,
		providerCfg.APIKey,
		providerCfg.Model,
	)

	// Create Binance exchange
	exchangeProvider := exchange.NewBinanceExchange(
		cfg.Binance.APIKey,
		cfg.Binance.APISecret,
		cfg.Binance.Testnet,
	)

	// Print startup info
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         TradingAgent v0.1.0              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  LLM:      %s (%s)\n", llmProvider.Name(), providerCfg.Model)
	fmt.Printf("  Exchange: %s (testnet: %v)\n", exchangeProvider.Name(), cfg.Binance.Testnet)
	fmt.Printf("  Interval: %s\n", interval)
	fmt.Println()

	// Create agent
	tradingAgent := agent.New(llmProvider, exchangeProvider, agent.Config{
		Interval:    interval,
		MaxTokens:   cfg.Agent.MaxTokens,
		Temperature: cfg.Agent.Temperature,
	})

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Run agent
	log.Println("Starting agent...")
	if err := tradingAgent.Run(ctx); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}
