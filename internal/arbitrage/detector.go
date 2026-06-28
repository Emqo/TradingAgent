package arbitrage

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// Detector detects arbitrage opportunities in real-time.
type Detector struct {
	exchange    exchange.Exchange
	config      DetectorConfig
	mu          sync.RWMutex
	opportunities []Opportunity
	subscribers []chan []Opportunity
}

// DetectorConfig holds detector configuration.
type DetectorConfig struct {
	// ScanInterval is how often to scan for opportunities.
	ScanInterval time.Duration `yaml:"scan_interval"`
	// MinSpreadBps is the minimum spread in basis points.
	MinSpreadBps float64 `yaml:"min_spread_bps"`
	// EnableTriangular enables triangular arbitrage detection.
	EnableTriangular bool `yaml:"enable_triangular"`
	// EnableCashAndCarry enables cash-and-carry arbitrage detection.
	EnableCashAndCarry bool `yaml:"enable_cash_and_carry"`
}

// NewDetector creates a new arbitrage detector.
func NewDetector(exchange exchange.Exchange, config DetectorConfig) *Detector {
	return &Detector{
		exchange:    exchange,
		config:      config,
		opportunities: make([]Opportunity, 0),
		subscribers: make([]chan []Opportunity, 0),
	}
}

// Start starts the detector.
func (d *Detector) Start(ctx context.Context) {
	log.Println("🔍 Arbitrage detector started")

	ticker := time.NewTicker(d.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🔍 Arbitrage detector stopped")
			return
		case <-ticker.C:
			d.scan(ctx)
		}
	}
}

// scan scans for arbitrage opportunities.
func (d *Detector) scan(ctx context.Context) {
	var opportunities []Opportunity

	// Scan triangular arbitrage
	if d.config.EnableTriangular {
		triangularOpps := d.scanTriangular(ctx)
		opportunities = append(opportunities, triangularOpps...)
	}

	// Scan cash-and-carry arbitrage
	if d.config.EnableCashAndCarry {
		cashAndCarryOpps := d.scanCashAndCarry(ctx)
		opportunities = append(opportunities, cashAndCarryOpps...)
	}

	// Update stored opportunities
	d.mu.Lock()
	d.opportunities = opportunities
	d.mu.Unlock()

	// Notify subscribers
	d.notifySubscribers(opportunities)
}

// scanTriangular scans for triangular arbitrage opportunities.
func (d *Detector) scanTriangular(ctx context.Context) []Opportunity {
	// TODO: Implement real triangular arbitrage detection
	// For now, return empty
	return []Opportunity{}
}

// scanCashAndCarry scans for cash-and-carry arbitrage opportunities.
func (d *Detector) scanCashAndCarry(ctx context.Context) []Opportunity {
	// TODO: Implement real cash-and-carry arbitrage detection
	// For now, return empty
	return []Opportunity{}
}

// GetOpportunities returns current opportunities.
func (d *Detector) GetOpportunities() []Opportunity {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.opportunities
}

// Subscribe subscribes to opportunity updates.
func (d *Detector) Subscribe() chan []Opportunity {
	ch := make(chan []Opportunity, 10)
	d.mu.Lock()
	d.subscribers = append(d.subscribers, ch)
	d.mu.Unlock()
	return ch
}

// notifySubscribers notifies all subscribers.
func (d *Detector) notifySubscribers(opportunities []Opportunity) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, ch := range d.subscribers {
		select {
		case ch <- opportunities:
		default:
			// Skip if channel is full
		}
	}
}
