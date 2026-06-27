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
	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/tools"
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

	// Create tool registry and register tools
	registry := tools.NewRegistry()
	registry.Register(tools.NewGetTickerTool(exchangeProvider))
	registry.Register(tools.NewGetOrderBookTool(exchangeProvider))
	registry.Register(tools.NewGetBalanceTool(exchangeProvider))
	registry.Register(tools.NewPlaceOrderTool(exchangeProvider))
	registry.Register(tools.NewGetOrderStatusTool(exchangeProvider))
	registry.Register(tools.NewDetectArbitrageTool(exchangeProvider))

	// Create arbitrage manager
	arbitrageManager := arbitrage.NewManager(
		exchangeProvider,
		arbitrage.TriangularConfig{
			MinSpreadBps:     15,     // 0.15% minimum spread
			MaxPositionUSDT:  1000,   // $1000 max position
			FeeRate:          0.001,  // 0.1% fee
			UseBNBDiscount:   true,   // Use BNB for fee discount
		},
		arbitrage.CashAndCarryConfig{
			MinAnnualizedYield: 10,    // 10% minimum annualized yield
			MaxPositionUSDT:    10000, // $10000 max position
			MarginMultiplier:   2.0,   // 2x safety margin
			MaxLeverage:        3.0,   // 3x max leverage
			FeeRate:            0.001, // 0.1% fee
		},
		arbitrage.ManagerConfig{
			ScanInterval:       1 * time.Minute,
			EnableTriangular:   true,
			EnableCashAndCarry: true,
		},
	)

	// Print startup info
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         TradingAgent v0.3.0              ║")
	fmt.Println("║         Phase 3: Arbitrage Engine        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  LLM:      %s (%s)\n", llmProvider.Name(), providerCfg.Model)
	fmt.Printf("  Exchange: %s (testnet: %v)\n", exchangeProvider.Name(), cfg.Binance.Testnet)
	fmt.Printf("  Interval: %s\n", interval)
	fmt.Printf("  Tools:    %d registered\n", len(registry.List()))
	fmt.Printf("  Arbitrage: Triangular + Cash-and-Carry\n")
	fmt.Println()

	// Create agent
	tradingAgent := agent.New(
		llmProvider,
		exchangeProvider,
		registry,
		arbitrageManager,
		agent.Config{
			Interval:    interval,
			MaxTokens:   cfg.Agent.MaxTokens,
			Temperature: cfg.Agent.Temperature,
		},
	)

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
