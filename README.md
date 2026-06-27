# TradingAgent

🤖 **LLM 驱动的加密货币交易智能体**

TradingAgent 是一个使用大语言模型（LLM）作为决策核心的加密货币交易系统。与传统量化策略不同，Agent 通过自然语言推理来分析市场、制定策略、执行交易。

## ✨ 特性

- 🧠 **LLM 决策** - 使用 Claude/GPT 作为策略大脑
- 📊 **实时行情** - Binance API 实时数据
- 🔄 **套利检测** - 三角套利 + 期现套利
- 🛡️ **风险管理** - 硬性风控约束，LLM 无法覆盖
- 💾 **记忆系统** - 短期/长期/工作记忆
- 📈 **策略生成** - LLM 自动生成交易策略
- 🔌 **可扩展** - Tool 系统，易于添加新功能

## 🏗️ 架构

```
┌─────────────────────────────────────────────┐
│              Strategy Layer (LLM)           │
│  分析市场 → 生成策略 → 动态调整参数 → 复盘    │
└─────────────────┬───────────────────────────┘
                  │ 策略配置 (JSON)
                  ▼
┌─────────────────────────────────────────────┐
│           Execution Engine (Go)             │
│  ┌──────────────┐  ┌──────────────┐         │
│  │  Opportunity │  │    Risk      │         │
│  │   Detector   │  │   Manager    │         │
│  │  (高频监控)   │  │  (硬性约束)   │         │
│  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │
│         ▼                 ▼                 │
│  ┌──────────────────────────────────┐       │
│  │         Order Executor           │       │
│  │      (低延迟交易执行)              │       │
│  └──────────────────────────────────┘       │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│           Binance API                       │
└─────────────────────────────────────────────┘
```

## 🚀 快速开始

### 前置要求

- Go 1.24+
- Binance 测试网 API Key
- LLM API Key（Claude 或 OpenAI）

### 安装

```bash
# 克隆仓库
git clone https://github.com/Emqo/TradingAgent.git
cd TradingAgent

# 安装依赖
go mod tidy

# 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml 填入你的 API Key

# 运行
go run cmd/agent/main.go
```

### Docker 部署

```bash
# 复制配置
cp config.example.yaml config.yaml
# 编辑 config.yaml

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f agent

# 停止服务
docker-compose down
```

## ⚙️ 配置

### config.yaml

```yaml
# LLM 配置
llm:
  default: "claude"
  providers:
    claude:
      base_url: "https://api.anthropic.com"
      api_key: "your-api-key"
      model: "claude-sonnet-4-20250514"

# Binance 配置
binance:
  testnet: true
  api_key: "your-binance-api-key"
  api_secret: "your-binance-api-secret"

# Agent 配置
agent:
  interval: "1m"
  max_tokens: 4096
  temperature: 0.7

# 风控配置
risk:
  max_position_usdt: 1000
  max_daily_loss_usdt: 500
  max_drawdown_pct: 10
```

## 🔧 Tool 系统

TradingAgent 通过 Tool 与外部世界交互：

| Tool | 功能 |
|------|------|
| `get_ticker` | 获取交易对价格 |
| `get_orderbook` | 获取订单簿深度 |
| `get_balance` | 查询账户余额 |
| `place_order` | 下单 |
| `detect_arbitrage` | 检测套利机会 |
| `check_risk` | 交易前风险检查 |
| `get_risk_status` | 获取风险状态 |
| `generate_strategy` | 生成交易策略 |
| `get_strategy_status` | 获取策略状态 |
| `add_memory` | 添加记忆 |
| `get_memory_context` | 获取记忆上下文 |

## 📊 风险管理

TradingAgent 实现了多层风险控制：

- **仓位限制** - 单笔最大仓位 $1,000
- **日亏损限制** - 单日最大亏损 $500
- **回撤控制** - 最大回撤 10%
- **杠杆限制** - 最大杠杆 3x
- **冷却期** - 亏损后 5 分钟冷却
- **交易暂停** - 触发风控时自动暂停

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/...

# 查看测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📁 项目结构

```
TradingAgent/
├── cmd/
│   └── agent/
│       └── main.go              # 入口
├── internal/
│   ├── agent/
│   │   └── agent.go             # Agent 决策循环
│   ├── arbitrage/
│   │   ├── triangular.go        # 三角套利
│   │   ├── cashandcarry.go      # 期现套利
│   │   └── manager.go           # 套利管理器
│   ├── exchange/
│   │   ├── exchange.go          # 交易所接口
│   │   └── binance.go           # Binance 实现
│   ├── health/
│   │   └── health.go            # 健康检查
│   ├── llm/
│   │   ├── provider.go          # LLM 接口
│   │   └── claude.go            # Claude 实现
│   ├── memory/
│   │   └── memory.go            # 记忆系统
│   ├── risk/
│   │   └── manager.go           # 风险管理
│   ├── strategy/
│   │   └── engine.go            # 策略引擎
│   └── tools/
│       ├── tool.go              # Tool 接口
│       ├── registry.go          # Tool 注册表
│       ├── market.go            # 行情 Tool
│       ├── portfolio.go         # 余额 Tool
│       ├── order.go             # 交易 Tool
│       ├── risk.go              # 风险 Tool
│       └── strategy.go          # 策略/记忆 Tool
├── config/
│   └── config.go                # 配置加载
├── docs/
│   ├── plans/
│   │   └── implementation-plan.md
│   └── agents/
│       ├── issue-tracker.md
│       ├── triage-labels.md
│       └── domain.md
├── Dockerfile
├── docker-compose.yml
├── config.example.yaml
├── go.mod
└── README.md
```

## ⚠️ 风险提示

**本软件仅供学习和研究目的。**

- 加密货币交易具有高风险，可能导致资金损失
- 本系统不保证盈利，历史表现不代表未来收益
- 使用前请充分了解相关风险
- 建议先在测试网验证，再考虑实盘交易
- 请勿投入超过你能承受损失的资金

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系

如有问题，请在 GitHub 上创建 Issue。
