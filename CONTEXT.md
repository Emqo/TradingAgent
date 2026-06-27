# CONTEXT.md — TradingAgent 领域语言

本文件定义 TradingAgent 项目的核心领域概念和术语。所有代码、文档和 issue 中涉及领域概念时，应使用此处定义的术语。

## 项目定位

TradingAgent 是一个 **LLM 驱动的加密货币交易智能体**。与传统量化策略（基于规则/代码）不同，Agent 使用大语言模型作为决策核心，通过自然语言推理来分析市场、制定策略、执行交易。

## Agent 架构

### Agent（智能体）

项目的核心 — 一个 LLM 实例，负责理解市场状况并做出交易决策。Agent 不是执行预定义规则，而是**推理**后行动。

Agent 的生命周期：
1. **Observe** — 通过 Tool 获取市场数据
2. **Think** — 分析数据，形成判断
3. **Act** — 调用 Tool 执行交易或调整仓位
4. **Remember** — 将决策和结果存入 Memory

### Tool（工具）

Agent 与外部世界交互的接口。每个 Tool 封装一个原子操作：

- **Market Data Tool** — 获取 K线、订单簿、Ticker 等行情数据
- **Order Tool** — 下单、撤单、查询订单状态
- **Portfolio Tool** — 查询仓位、余额、盈亏
- **Analysis Tool** — 计算技术指标、生成图表
- **Risk Tool** — 检查风控约束、计算仓位大小

Agent 通过 Tool 调用来"感知"和"行动"，而非直接访问 API。

### Memory（记忆）

Agent 的上下文存储，决定它能"看到"什么历史信息：

- **Short-term Memory** — 当前会话的对话历史（上下文窗口内的内容）
- **Long-term Memory** — 跨会话持久化的决策记录、市场观察、策略复盘
- **Working Memory** — 当前分析所需的临时数据（如最近 100 根 K线）

Memory 管理是关键挑战 — 上下文窗口有限，需要决定保留什么、丢弃什么。

### Context Window（上下文窗口）

LLM 单次能处理的最大 token 数量。决定了 Agent 一次能"看到"多少市场数据和历史记忆。

约束：
- 市场数据占用大量 token（K线数据、订单簿深度）
- 需要在数据丰富度和上下文空间之间权衡
- 可能需要数据压缩/摘要策略

## 决策循环

### Observation（观察）

Agent 获取当前市场状态的过程。一次 Observation 可能包含：
- 多个交易对的 K线数据
- 当前持仓和余额
- 最近的交易记录
- 新闻/社交媒体情绪（可选）

### Reasoning（推理）

Agent 基于 Observation 进行思考的过程。输出是一个自然语言的分析，包含：
- 市场趋势判断
- 潜在交易机会
- 风险评估
- 行动计划

### Action（行动）

Agent 的最终输出 — 调用一个或多个 Tool。Action 必须是结构化的：
- 调用哪个 Tool
- 传入什么参数
- 是否需要确认（高风险操作）

### Decision Loop（决策循环）

Agent 的主循环：

```
while running:
    observation = observe(market_data, portfolio, memory)
    reasoning = llm.think(observation, system_prompt)
    action = llm.decide(reasoning)
    result = execute(action)
    memory.store(observation, reasoning, action, result)
    sleep(interval)
```

## 交易领域概念

### Asset（资产）

可交易的加密货币资产，如 BTC、ETH、SOL。格式为 `BASE/QUOTE`（如 `BTC/USDT`）。

### Exchange（交易所）

提供流动性和订单执行的平台（如 Binance、OKX、Bybit）。

### Order（订单）

一次买卖请求：
- **side**: `buy` / `sell`
- **type**: `market` / `limit`
- **quantity**: 数量
- **price**: 价格（限价单）
- **status**: `pending` → `filled` / `cancelled` / `rejected`

### Position（仓位）

当前持有的资产：
- **long** / **short**: 方向
- **size**: 仓位大小
- **entry_price**: 开仓均价
- **unrealized_pnl**: 未实现盈亏

### Portfolio（投资组合）

所有 positions 和 cash balance 的总和。

## 市场数据

### Candle（K线）

OHLCV 数据：
- **open / high / low / close / volume**
- **timeframe**: 1m, 5m, 15m, 1h, 4h, 1d

### OrderBook（订单簿）

实时买卖挂单：bids, asks, spread。

### Ticker（行情快照）

实时摘要：last price, 24h volume, 24h change。

## 风险管理

### Risk Constraint（风控约束）

Agent 必须遵守的硬性规则，作为 System Prompt 的一部分或 Tool 的前置检查：
- **position_size_limit**: 单笔最大仓位
- **max_drawdown**: 最大回撤阈值
- **stop_loss**: 止损价格
- **take_profit**: 止盈价格
- **daily_loss_limit**: 单日最大亏损

Agent 可以"建议"突破风控，但系统应拒绝执行。

## 回测

### Backtest（回测）

在历史数据上模拟 Agent 的决策循环，评估表现。

与传统回测不同：Agent 的每次决策都是一次 LLM 调用，成本高、速度慢。需要考虑：
- 模拟 LLM 响应 vs 真实调用
- 决策一致性（相同输入是否产生相同输出）
- 成本控制

### BacktestResult（回测结果）

- **total_return**: 总收益率
- **sharpe_ratio**: 夏普比率
- **max_drawdown**: 最大回撤
- **win_rate**: 胜率
- **total_trades**: 总交易次数
- **llm_cost**: LLM 调用成本

## 术语约定

| 中文 | 英文 | 说明 |
|------|------|------|
| 智能体 | Agent | LLM 驱动的决策核心 |
| 工具 | Tool | Agent 与外部交互的接口 |
| 记忆 | Memory | Agent 的上下文存储 |
| 观察 | Observation | 获取市场状态 |
| 推理 | Reasoning | Agent 的思考过程 |
| 行动 | Action | Agent 的决策输出 |
| 交易对 | Trading Pair | 如 BTC/USDT |
| K线 | Candle | OHLCV 数据 |
| 开仓 | Open Position | 建立新仓位 |
| 平仓 | Close Position | 结束仓位 |
| 止损 | Stop Loss | 价格触及阈值自动卖出 |
| 止盈 | Take Profit | 价格触及阈值自动卖出 |
| 回撤 | Drawdown | 从峰值到谷值的下降 |
