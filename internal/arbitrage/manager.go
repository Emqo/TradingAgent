package arbitrage

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// Manager coordinates all arbitrage engines.
type Manager struct {
	triangular   *TriangularEngine
	cashAndCarry *CashAndCarryEngine
	exchange     exchange.Exchange
	config       ManagerConfig
	mu           sync.RWMutex
}

// ManagerConfig holds configuration for the arbitrage manager.
type ManagerConfig struct {
	// ScanInterval is how often to scan for opportunities.
	ScanInterval time.Duration `yaml:"scan_interval"`
	// EnableTriangular enables triangular arbitrage scanning.
	EnableTriangular bool `yaml:"enable_triangular"`
	// EnableCashAndCarry enables cash-and-carry arbitrage scanning.
	EnableCashAndCarry bool `yaml:"enable_cash_and_carry"`
}

// ScanResult contains all detected arbitrage opportunities.
type ScanResult struct {
	TriangularOpportunities   []Opportunity             `json:"triangular_opportunities"`
	CashAndCarryOpportunities []CashAndCarryOpportunity `json:"cash_and_carry_opportunities"`
	Timestamp                 time.Time                 `json:"timestamp"`
}

// NewManager creates a new arbitrage manager.
func NewManager(
	exchange exchange.Exchange,
	triangularConfig TriangularConfig,
	cashAndCarryConfig CashAndCarryConfig,
	managerConfig ManagerConfig,
) *Manager {
	return &Manager{
		triangular:   NewTriangularEngine(exchange, triangularConfig),
		cashAndCarry: NewCashAndCarryEngine(exchange, cashAndCarryConfig),
		exchange:     exchange,
		config:       managerConfig,
	}
}

// Scan scans all arbitrage engines for opportunities.
func (m *Manager) Scan(ctx context.Context) (*ScanResult, error) {
	result := &ScanResult{
		Timestamp: time.Now(),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Scan triangular arbitrage
	if m.config.EnableTriangular {
		wg.Add(1)
		go func() {
			defer wg.Done()

			opportunities, err := m.triangular.Scan(ctx)
			if err != nil {
				log.Printf("⚠️ Triangular scan error: %v", err)
				return
			}

			mu.Lock()
			result.TriangularOpportunities = opportunities
			mu.Unlock()
		}()
	}

	// Scan cash-and-carry arbitrage
	if m.config.EnableCashAndCarry {
		wg.Add(1)
		go func() {
			defer wg.Done()

			opportunities, err := m.cashAndCarry.Scan(ctx)
			if err != nil {
				log.Printf("⚠️ Cash-and-carry scan error: %v", err)
				return
			}

			mu.Lock()
			result.CashAndCarryOpportunities = opportunities
			mu.Unlock()
		}()
	}

	wg.Wait()

	return result, nil
}

// GetTriangularEngine returns the triangular arbitrage engine.
func (m *Manager) GetTriangularEngine() *TriangularEngine {
	return m.triangular
}

// GetCashAndCarryEngine returns the cash-and-carry arbitrage engine.
func (m *Manager) GetCashAndCarryEngine() *CashAndCarryEngine {
	return m.cashAndCarry
}

// FormatScanResult formats a scan result for display.
func FormatScanResult(result *ScanResult) string {
	output := "📊 Arbitrage Scan Results\n"
	output += "========================\n\n"

	if len(result.TriangularOpportunities) > 0 {
		output += "🔺 Triangular Arbitrage Opportunities:\n"
		for _, opp := range result.TriangularOpportunities {
			output += "  • " + opp.Path.Name + "\n"
			output += "    Spread: " + formatBps(opp.Spread) + " bps\n"
			output += "    Profit: $" + formatFloat(opp.Profit, 2) + "\n"
		}
		output += "\n"
	} else {
		output += "🔺 No triangular arbitrage opportunities found\n\n"
	}

	if len(result.CashAndCarryOpportunities) > 0 {
		output += "💰 Cash-and-Carry Opportunities:\n"
		for _, opp := range result.CashAndCarryOpportunities {
			output += "  • " + opp.Symbol + "\n"
			output += "    Annualized Yield: " + formatFloat(opp.AnnualizedYield, 2) + "%\n"
			output += "    Basis: " + formatFloat(opp.BasisPercent, 4) + "%\n"
			output += "    Required Hold: " + opp.RequiredHold.String() + "\n"
		}
	} else {
		output += "💰 No cash-and-carry opportunities found\n"
	}

	return output
}

func formatBps(bps float64) string {
	if bps >= 0 {
		return "+" + formatFloat(bps, 2)
	}
	return formatFloat(bps, 2)
}

func formatFloat(f float64, decimals int) string {
	return strconv.FormatFloat(f, 'f', decimals, 64)
}
