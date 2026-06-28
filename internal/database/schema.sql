-- TradingAgent Database Schema

-- 交易记录
CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    type VARCHAR(20) NOT NULL, -- MARKET, LIMIT
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    total DECIMAL(20, 8) NOT NULL,
    fee DECIMAL(20, 8) DEFAULT 0,
    pnl DECIMAL(20, 8) DEFAULT 0,
    strategy VARCHAR(50), -- triangular, cash_and_carry, agent
    status VARCHAR(20) DEFAULT 'filled', -- pending, filled, cancelled, failed
    order_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 决策日志
CREATE TABLE IF NOT EXISTS decisions (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL, -- BUY, SELL, HOLD, ANALYZE
    symbol VARCHAR(20),
    reason TEXT,
    result TEXT,
    pnl DECIMAL(20, 8) DEFAULT 0,
    tokens_used INTEGER DEFAULT 0,
    latency_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 套利机会
CREATE TABLE IF NOT EXISTS arbitrage_opportunities (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- triangular, cash_and_carry
    path VARCHAR(200) NOT NULL,
    spread_bps DECIMAL(10, 2),
    profit_usdt DECIMAL(20, 8),
    executed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 策略配置
CREATE TABLE IF NOT EXISTS strategies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- triangular, cash_and_carry, agent
    config JSONB NOT NULL,
    active BOOLEAN DEFAULT true,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 风控告警
CREATE TABLE IF NOT EXISTS risk_alerts (
    id SERIAL PRIMARY KEY,
    level VARCHAR(20) NOT NULL, -- INFO, WARNING, ERROR, CRITICAL
    type VARCHAR(50) NOT NULL, -- POSITION_LIMIT, DAILY_LOSS, DRAWDOWN, etc.
    message TEXT NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 每日统计
CREATE TABLE IF NOT EXISTS daily_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    total_balance DECIMAL(20, 8),
    daily_pnl DECIMAL(20, 8),
    daily_pnl_pct DECIMAL(10, 4),
    total_trades INTEGER DEFAULT 0,
    winning_trades INTEGER DEFAULT 0,
    losing_trades INTEGER DEFAULT 0,
    win_rate DECIMAL(5, 2),
    max_drawdown DECIMAL(10, 4),
    arbitrage_opportunities INTEGER DEFAULT 0,
    arbitrage_profit DECIMAL(20, 8) DEFAULT 0,
    agent_decisions INTEGER DEFAULT 0,
    agent_pnl DECIMAL(20, 8) DEFAULT 0,
    llm_calls INTEGER DEFAULT 0,
    llm_tokens INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 收益曲线（每小时快照）
CREATE TABLE IF NOT EXISTS equity_snapshots (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    positions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_trades_created_at ON trades(created_at);
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
CREATE INDEX IF NOT EXISTS idx_trades_strategy ON trades(strategy);
CREATE INDEX IF NOT EXISTS idx_decisions_created_at ON decisions(created_at);
CREATE INDEX IF NOT EXISTS idx_arbitrage_created_at ON arbitrage_opportunities(created_at);
CREATE INDEX IF NOT EXISTS idx_risk_alerts_created_at ON risk_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date);
CREATE INDEX IF NOT EXISTS idx_equity_snapshots_timestamp ON equity_snapshots(timestamp);
