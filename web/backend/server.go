package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Emqo/TradingAgent/internal/agent"
	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/risk"
	"github.com/Emqo/TradingAgent/web/backend/handlers"
	"github.com/Emqo/TradingAgent/web/backend/middleware"
	"github.com/Emqo/TradingAgent/web/backend/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server represents the web server.
type Server struct {
	router     *gin.Engine
	db         *sql.DB
	jwtAuth    *middleware.JWTAuth
	userStore  store.UserStoreInterface
	authHandler *handlers.AuthHandler
	dashboardHandler *handlers.DashboardHandler
	arbitrageHandler *handlers.ArbitrageHandler
	agentHandler *handlers.AgentHandler
}

// Config holds server configuration.
type Config struct {
	Port       int
	JWTSecret  string
	JWTExpiry  time.Duration
	AllowedOrigins []string
}

// NewServer creates a new web server.
func NewServer(
	cfg Config,
	db *sql.DB,
	exchange exchange.Exchange,
	riskMgr *risk.Manager,
	arbMgr *arbitrage.Manager,
	agent *agent.Agent,
) (*Server, error) {
	// Create JWT auth
	jwtAuth := middleware.NewJWTAuth(cfg.JWTSecret, cfg.JWTExpiry)

	// Create handlers
	dashboardHandler := handlers.NewDashboardHandler(exchange, riskMgr, arbMgr, agent)
	arbitrageHandler := handlers.NewArbitrageHandler(exchange, arbMgr)
	agentHandler := handlers.NewAgentHandler()

	// Create router
	router := gin.Default()

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		router:          router,
		db:              db,
		jwtAuth:         jwtAuth,
		dashboardHandler: dashboardHandler,
		arbitrageHandler: arbitrageHandler,
		agentHandler:     agentHandler,
	}

	// Initialize user store
	var userStore store.UserStoreInterface
	if db != nil {
		pgStore := store.NewUserStore(db)
		if err := pgStore.Init(); err != nil {
			return nil, fmt.Errorf("init user store: %w", err)
		}
		userStore = pgStore
	} else {
		// Use in-memory store for development
		userStore = store.NewMemoryUserStore()
	}
	s.userStore = userStore
	s.authHandler = handlers.NewAuthHandler(userStore, jwtAuth)

	s.setupRoutes()

	return s, nil
}

// setupRoutes sets up the API routes.
func (s *Server) setupRoutes() {
	api := s.router.Group("/api")

	// Public routes (only if auth handler is available)
	if s.authHandler != nil {
		api.POST("/auth/login", s.authHandler.Login)
		api.POST("/auth/register", s.authHandler.Register)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(s.jwtAuth.Middleware())
	{
		// Auth (only if auth handler is available)
		if s.authHandler != nil {
			protected.GET("/auth/profile", s.authHandler.GetProfile)
			protected.PUT("/auth/password", s.authHandler.ChangePassword)
		}

		// Dashboard
		protected.GET("/dashboard/stats", s.dashboardHandler.GetStats)
		protected.GET("/dashboard/positions", s.dashboardHandler.GetPositions)
		protected.GET("/dashboard/balance", s.dashboardHandler.GetBalance)
		protected.GET("/dashboard/risk", s.dashboardHandler.GetRiskStatus)
		protected.POST("/dashboard/pause", s.dashboardHandler.PauseTrading)
		protected.POST("/dashboard/resume", s.dashboardHandler.ResumeTrading)

		// Arbitrage
		protected.GET("/arbitrage/opportunities", s.arbitrageHandler.GetOpportunities)
		protected.GET("/arbitrage/stats", s.arbitrageHandler.GetStats)

		// Agent
		protected.GET("/agent/decisions", s.agentHandler.GetDecisions)
		protected.GET("/agent/stats", s.agentHandler.GetStats)
	}
}

// Start starts the web server.
func (s *Server) Start(addr string) error {
	log.Printf("🌐 Web server starting on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
