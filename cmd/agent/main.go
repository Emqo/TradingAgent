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
	"github.com/Emqo/TradingAgent/internal/memory"
	"github.com/Emqo/TradingAgent/internal/risk"
	"github.com/Emqo/TradingAgent/internal/strategy"
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

	// Create Binance exchange (with caching)
	binanceExchange := exchange.NewBinanceExchange(
		cfg.Binance.APIKey,
		cfg.Binance.APISecret,
		cfg.Binance.Testnet,
	)
	exchangeProvider := exchange.NewCachedExchange(binanceExchange)

	// Create memory manager
	memoryManager := memory.NewManager(100, 1000) // 100 short-term, 1000 long-term

	// Create risk manager
	riskManager := risk.NewManager(risk.Config{
		MaxPositionUSDT:       cfg.Risk.MaxPositionUSDT,
		MaxTotalPositionUSDT:  cfg.Risk.MaxPositionUSDT * 5,
		MaxDailyLossUSDT:      cfg.Risk.MaxDailyLossUSDT,
		MaxDailyLossPercent:   10,
		MaxDrawdownPercent:    cfg.Risk.MaxDrawdownPct,
		MaxLeverage:           3.0,
		MaxExposurePerPairUSDT: cfg.Risk.MaxPositionUSDT * 2,
		CooldownAfterLoss:     5 * time.Minute,
	})

	// Add alert listener for logging
	riskManager.AddListener(func(alert risk.Alert) {
		log.Printf("🚨 Risk Alert [%s/%s]: %s", alert.Level, alert.Type, alert.Message)
	})

	// Create strategy engine
	strategyEngine := strategy.NewEngine(llmProvider, memoryManager, strategy.Config{
		StrategyTTL: 1 * time.Hour,
		MaxTokens:   cfg.Agent.MaxTokens,
	})

	// Create tool registry and register tools
	registry := tools.NewRegistry()

	// Market data tools
	registry.Register(tools.NewGetTickerTool(exchangeProvider))
	registry.Register(tools.NewGetOrderBookTool(exchangeProvider))
	registry.Register(tools.NewGetBalanceTool(exchangeProvider))

	// Order tools
	registry.Register(tools.NewPlaceOrderTool(exchangeProvider))
	registry.Register(tools.NewGetOrderStatusTool(exchangeProvider))
	registry.Register(tools.NewCancelOrderTool(exchangeProvider))
	registry.Register(tools.NewGetOpenOrdersTool(exchangeProvider))

	// Arbitrage tools
	registry.Register(tools.NewDetectArbitrageTool(exchangeProvider))

	// Risk tools
	registry.Register(tools.NewCheckRiskTool(riskManager))
	registry.Register(tools.NewGetRiskStatusTool(riskManager))

	// Strategy tools
	registry.Register(tools.NewGenerateStrategyTool(strategyEngine))
	registry.Register(tools.NewGetStrategyStatusTool(strategyEngine))

	// Memory tools
	registry.Register(tools.NewAddMemoryTool(memoryManager))
	registry.Register(tools.NewGetMemoryContextTool(memoryManager))

	// Create arbitrage manager
	arbitrageManager := arbitrage.NewManager(
		exchangeProvider,
		arbitrage.TriangularConfig{
			MinSpreadBps:     15,
			MaxPositionUSDT:  cfg.Risk.MaxPositionUSDT,
			FeeRate:          0.001,
			UseBNBDiscount:   true,
		},
		arbitrage.CashAndCarryConfig{
			MinAnnualizedYield: 10,
			MaxPositionUSDT:    cfg.Risk.MaxPositionUSDT * 10,
			MarginMultiplier:   2.0,
			MaxLeverage:        3.0,
			FeeRate:            0.001,
		},
		arbitrage.ManagerConfig{
			ScanInterval:       1 * time.Minute,
			EnableTriangular:   true,
			EnableCashAndCarry: true,
		},
	)

	// Print startup info
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         TradingAgent v0.5.0              ║")
	fmt.Println("║     Phase 5: Strategy & Memory           ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  LLM:      %s (%s)\n", llmProvider.Name(), providerCfg.Model)
	fmt.Printf("  Exchange: %s (testnet: %v)\n", exchangeProvider.Name(), cfg.Binance.Testnet)
	fmt.Printf("  Interval: %s\n", interval)
	fmt.Printf("  Tools:    %d registered\n", len(registry.List()))
	fmt.Printf("  Memory:   Short-term: 100, Long-term: 1000\n")
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
