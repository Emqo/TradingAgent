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
	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/logger"
	"github.com/Emqo/TradingAgent/internal/memory"
	"github.com/Emqo/TradingAgent/internal/metrics"
	"github.com/Emqo/TradingAgent/internal/risk"
	"github.com/Emqo/TradingAgent/internal/strategy"
	"github.com/Emqo/TradingAgent/internal/tools"
	"github.com/Emqo/TradingAgent/web/backend"
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

	// Create Binance exchange (with WebSocket for real-time data)
	wsExchange := exchange.NewWebSocketExchange(
		cfg.Binance.APIKey,
		cfg.Binance.APISecret,
		cfg.Binance.Testnet,
	)

	// Connect WebSocket
	if err := wsExchange.Connect(); err != nil {
		logger.Warnf("WebSocket connection failed, using REST only: %v", err)
	} else {
		logger.Info("WebSocket connected for real-time data")
	}
	exchangeProvider := wsExchange

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

	// Create trading agent tool registry
	tradingRegistry := tools.NewRegistry()

	// Market data tools
	tradingRegistry.Register(tools.NewGetTickerTool(exchangeProvider))
	tradingRegistry.Register(tools.NewGetOrderBookTool(exchangeProvider))
	tradingRegistry.Register(tools.NewGetBalanceTool(exchangeProvider))

	// Order tools
	tradingRegistry.Register(tools.NewPlaceOrderTool(exchangeProvider))
	tradingRegistry.Register(tools.NewGetOrderStatusTool(exchangeProvider))
	tradingRegistry.Register(tools.NewCancelOrderTool(exchangeProvider))
	tradingRegistry.Register(tools.NewGetOpenOrdersTool(exchangeProvider))

	// Risk tools
	tradingRegistry.Register(tools.NewCheckRiskTool(riskManager))
	tradingRegistry.Register(tools.NewGetRiskStatusTool(riskManager))

	// Strategy tools
	tradingRegistry.Register(tools.NewGenerateStrategyTool(strategyEngine))
	tradingRegistry.Register(tools.NewGetStrategyStatusTool(strategyEngine))

	// Memory tools
	tradingRegistry.Register(tools.NewAddMemoryTool(memoryManager))
	tradingRegistry.Register(tools.NewGetMemoryContextTool(memoryManager))

	// Create arbitrage agent tool registry
	arbitrageRegistry := tools.NewRegistry()

	// Market data tools
	arbitrageRegistry.Register(tools.NewGetTickerTool(exchangeProvider))
	arbitrageRegistry.Register(tools.NewGetOrderBookTool(exchangeProvider))
	arbitrageRegistry.Register(tools.NewGetBalanceTool(exchangeProvider))

	// Risk tools
	arbitrageRegistry.Register(tools.NewCheckRiskTool(riskManager))
	arbitrageRegistry.Register(tools.NewGetRiskStatusTool(riskManager))

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

	// Create arbitrage detector (fast, background scanning)
	arbitrageDetector := arbitrage.NewDetector(
		exchangeProvider,
		arbitrage.DetectorConfig{
			ScanInterval:       10 * time.Second,
			MinSpreadBps:       15,
			EnableTriangular:   true,
			EnableCashAndCarry: true,
		},
	)

	// Start arbitrage detector in background
	go arbitrageDetector.Start(context.Background())

	// Register arbitrage tools to arbitrage registry
	arbitrageRegistry.Register(tools.NewGetArbitrageOpportunitiesTool(arbitrageManager))
	arbitrageRegistry.Register(tools.NewExecuteArbitrageTool(exchangeProvider, arbitrageManager))

	// Create logger
	logLevel := logger.LevelInfo
	log := logger.New(logLevel, os.Stdout)
	logger.SetDefault(log)

	// Create metrics
	metricsInstance := metrics.NewMetrics()

	// Start metrics server in background
	go func() {
		log.Info("Starting metrics server on :9090")
		if err := metrics.StartMetricsServer(":9090"); err != nil {
			log.Errorf("Metrics server error: %v", err)
		}
	}()

	// Start web server (optional - requires database)
	webServerPort := 8080
	if port := os.Getenv("WEB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &webServerPort)
	}

	// Connect to PostgreSQL
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := 5432
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &dbPort)
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "trading_agent"
	}

	db, err := database.New(database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
	})
	if err != nil {
		log.Warnf("Failed to connect to database: %v", err)
		log.Warnf("Running without database (data will not be persisted)")
		db = nil
	} else {
		log.Info("Connected to PostgreSQL")
		defer db.Close()
	}

	// Print startup info
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         TradingAgent v1.0.0              ║")
	fmt.Println("║     Dual Agent Architecture              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  LLM:      %s (%s)\n", llmProvider.Name(), providerCfg.Model)
	fmt.Printf("  Exchange: %s (testnet: %v)\n", exchangeProvider.Name(), cfg.Binance.Testnet)
	fmt.Printf("  Interval: %s\n", interval)
	fmt.Printf("  Tools:    %d registered (trading), %d registered (arbitrage)\n", len(tradingRegistry.List()), len(arbitrageRegistry.List()))
	fmt.Printf("  Memory:   Short-term: 100, Long-term: 1000\n")
	fmt.Printf("  Logging:  JSON structured\n")
	fmt.Printf("  Database: %s\n", map[bool]string{true: "Connected", false: "Not connected"}[db != nil])
	fmt.Printf("  Agents:   Trading + Arbitrage\n")
	fmt.Printf("  Web UI:   http://localhost:%d\n", webServerPort)
	fmt.Println()

	// Create trading agent
	tradingAgent := agent.NewTradingAgent(
		llmProvider,
		exchangeProvider,
		tradingRegistry,
		arbitrageManager,
		metricsInstance,
		log,
		db,
		agent.TradingAgentConfig{
			Interval:    interval,
			MaxTokens:   cfg.Agent.MaxTokens,
			Temperature: cfg.Agent.Temperature,
		},
	)

	// Create arbitrage agent (faster interval)
	arbitrageAgent := agent.NewArbitrageAgent(
		llmProvider,
		exchangeProvider,
		arbitrageRegistry,
		arbitrageManager,
		metricsInstance,
		log,
		db,
		agent.ArbitrageAgentConfig{
			Interval:    30 * time.Second, // Faster than trading agent
			MaxTokens:   cfg.Agent.MaxTokens,
			Temperature: cfg.Agent.Temperature,
		},
	)

	// Start web server
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "trading-agent-secret-key-change-in-production"
	}

	// Start web server
	webServer, err := backend.NewServer(
		backend.Config{
			Port:      webServerPort,
			JWTSecret: jwtSecret,
			JWTExpiry: 24 * time.Hour,
			AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		db,
		exchangeProvider,
		riskManager,
		arbitrageManager,
		tradingAgent,
	)
	if err != nil {
		log.Warnf("Failed to create web server: %v", err)
	} else {
		go func() {
			if err := webServer.Start(fmt.Sprintf(":%d", webServerPort)); err != nil {
				log.Errorf("Web server error: %v", err)
			}
		}()
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.WithField("signal", sig).Info("Received signal, shutting down")
		cancel()
	}()

	// Run both agents concurrently
	log.Info("Starting agents")

	// Start trading agent
	go func() {
		log.Info("Starting Trading Agent")
		if err := tradingAgent.Run(ctx); err != nil {
			log.Errorf("Trading Agent error: %v", err)
		}
	}()

	// Start arbitrage agent
	go func() {
		log.Info("Starting Arbitrage Agent")
		if err := arbitrageAgent.Run(ctx); err != nil {
			log.Errorf("Arbitrage Agent error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Info("All agents stopped")
}
