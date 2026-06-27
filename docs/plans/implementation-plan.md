# TradingAgent 实现计划

> 混合架构：LLM 做策略思考，Go 做快速执行
> 目标：全自动加密货币套利交易智能体

---

## 技术选型

| 组件 | 选型 | 理由 |
|------|------|------|
| 语言 | Go 1.24+ | 高性能、并发友好 |
| LLM Provider | `anthropic-sdk-go` + `openai-go` | 官方 SDK，支持 function calling |
| 交易所 | `ccxt/go-binance` | Binance 最成熟的 Go SDK |
| WebSocket | `coder/websocket` | 现代、零依赖、并发安全 |
| 数据库 | `pgx` + `gorm` | 高性能 PG 驱动 + ORM 便利 |
| 技术指标 | `sdcoffey/techan` | 纯 Go，指标齐全 |
| 配置 | Viper | 支持多格式、环境变量 |

---

## Phase 1: 基础骨架 (Week 1)

**目标：** 跑通最小可运行版本，验证 LLM + Binance 连通性

### 1.1 项目初始化
- [ ] `go mod init github.com/Emqo/TradingAgent`
- [ ] 目录结构搭建（cmd/, internal/, config/）
- [ ] 配置管理（Viper + .env）
- [ ] 日志系统（slog 或 zerolog）

### 1.2 LLM Provider 抽象层
- [ ] 定义 `Provider` 接口（Chat, Stream, ToolUse）
- [ ] 实现 Claude Provider（anthropic-sdk-go）
- [ ] 实现 OpenAI Provider（openai-go）
- [ ] Provider 工厂 + 配置切换
- [ ] 单元测试：mock provider 验证接口

### 1.3 Binance 连通性
- [ ] 配置 Binance API Key（测试网）
- [ ] 封装基础 API：获取余额、获取交易对信息
- [ ] WebSocket 连接：实时 Ticker 数据
- [ ] 单元测试：API 调用 + WebSocket 订阅

### 1.4 最小 Agent 循环
- [ ] 实现 Observe → Think → Act 循环骨架
- [ ] LLM 接收市场数据，输出决策文本
- [ ] 验证：Agent 能分析 BTC 价格并给出建议

**交付物：** 运行后能看到 LLM 分析实时行情并输出交易建议

---

## Phase 2: Tool 系统 (Week 2)

**目标：** Agent 通过 Tool 调用执行实际操作

### 2.1 Tool 框架
- [ ] 定义 `Tool` 接口（Name, Description, Parameters, Execute）
- [ ] Tool 注册表 + 发现机制
- [ ] Tool → LLM function calling schema 转换
- [ ] Tool 执行结果格式化（成功/失败/部分成功）

### 2.2 核心 Tool 实现

#### Market Data Tools
- [ ] `get_ticker` — 获取交易对实时价格
- [ ] `get_klines` — 获取 K线数据（多时间周期）
- [ ] `get_orderbook` — 获取订单簿深度
- [ ] `get_funding_rate` — 获取合约资金费率

#### Order Tools
- [ ] `place_order` — 下单（市价/限价）
- [ ] `cancel_order` — 撤单
- [ ] `get_order_status` — 查询订单状态
- [ ] `get_open_orders` — 查询未成交订单

#### Portfolio Tools
- [ ] `get_balance` — 查询账户余额
- [ ] `get_positions` — 查询当前持仓
- [ ] `get_pnl` — 查询盈亏情况

#### Analysis Tools
- [ ] `calculate_indicator` — 计算技术指标（SMA/EMA/RSI/MACD）
- [ ] `detect_arbitrage` — 检测套利机会（价差计算）

### 2.3 Tool 集成测试
- [ ] Agent 调用 Tool 获取真实市场数据
- [ ] Agent 调用 Tool 在测试网下单
- [ ] 验证：Agent 能自主完成"查看行情 → 分析 → 下单"流程

**交付物：** Agent 能通过 Tool 与 Binance 交互，执行完整的交易流程

---

## Phase 3: 套利引擎 (Week 3)

**目标：** 实现三角套利和期现套利的检测与执行

### 3.1 三角套利

#### 机会检测器
- [ ] 构建交易对图（邻接表）
- [ ] 枚举所有三角路径（A→B→C→A）
- [ ] 实时计算价差（含手续费）
- [ ] 过滤：价差 > 阈值（默认 0.3%，BNB 折扣后 0.225%）

#### 执行引擎
- [ ] 三笔订单的原子性执行（尽量同时）
- [ ] 订单簿深度检查（避免滑点过大）
- [ ] 部分成交处理 + 回退策略
- [ ] 执行延迟监控（目标 < 100ms）

#### 收益计算
- [ ] 实际收益 = 理论价差 - 3×手续费 - 滑点
- [ ] 单次收益记录 + 累计统计

### 3.2 期现套利

#### 资金费率监控
- [ ] 定时获取资金费率（每 8 小时）
- [ ] 计算年化收益率
- [ ] 阈值判断：年化 > 10% 才考虑开仓

#### Delta-Neutral 策略
- [ ] 开仓：现货买入 + 合约做空（同时）
- [ ] 持仓监控：保证金率、未实现盈亏
- [ ] 平仓条件：资金费率转负 / 达到目标收益 / 触发止损

#### 风控
- [ ] 保证金率监控（> 2x 安全边际）
- [ ] 强平预警
- [ ] 资金费率趋势监控（连续转负则平仓）

### 3.3 回测框架
- [ ] 历史数据获取（Binance API）
- [ ] 模拟执行引擎（不含真实下单）
- [ ] 收益指标计算（总收益、夏普比率、最大回撤）
- [ ] LLM 成本估算（token 数 × 单价）

**交付物：** 能在回测环境中验证套利策略的可行性

---

## Phase 4: 风险管理 (Week 4)

**目标：** 硬性风控约束，LLM 无法覆盖

### 4.1 Risk Manager 核心
- [ ] 定义 `RiskChecker` 接口
- [ ] 交易前检查：仓位大小、杠杆倍数、单日亏损
- [ ] 交易后更新：持仓统计、盈亏累计
- [ ] 拒绝执行 + 告警通知

### 4.2 风控规则
- [ ] 单笔仓位上限（可配置）
- [ ] 总仓位上限
- [ ] 单日最大亏损（绝对值 + 百分比）
- [ ] 最大回撤阈值
- [ ] 杠杆倍数上限
- [ ] 单个交易对最大敞口

### 4.3 告警系统
- [ ] 日志告警（slog）
- [ ] 可选：Telegram/Slack 通知
- [ ] 风控触发时暂停交易

**交付物：** 风控系统能拦截超限交易，保护资金安全

---

## Phase 5: 策略层与 Memory (Week 5)

**目标：** LLM 策略层与执行层解耦，Agent 有记忆能力

### 5.1 策略引擎
- [ ] LLM 分析市场 → 输出策略配置 JSON
- [ ] 策略配置持久化（PostgreSQL）
- [ ] 策略 TTL 机制（过期自动重新生成）
- [ ] 策略版本管理

### 5.2 Memory 系统
- [ ] Short-term Memory：当前会话上下文（内存）
- [ ] Long-term Memory：交易历史、策略复盘（PostgreSQL）
- [ ] Working Memory：当前分析数据（内存，定期清理）
- [ ] Memory 摘要：超长历史自动压缩

### 5.3 决策循环优化
- [ ] 定时复盘（每小时/每天）
- [ ] 基于历史表现调整策略参数
- [ ] 异常检测：连续亏损时降低仓位或暂停

**交付物：** Agent 能根据市场变化动态调整策略

---

## Phase 6: 生产化 (Week 6)

**目标：** 可部署、可观测、可维护

### 6.1 部署
- [ ] Dockerfile（多阶段构建）
- [ ] docker-compose.yml（Agent + PostgreSQL）
- [ ] 环境变量配置（.env.example）
- [ ] 健康检查端点

### 6.2 可观测性
- [ ] 结构化日志（JSON 格式）
- [ ] 关键指标暴露（Prometheus）
  - 交易次数、成功率、平均收益
  - LLM 调用次数、延迟、token 消耗
  - 风控触发次数
- [ ] Grafana Dashboard 模板

### 6.3 测试
- [ ] 单元测试覆盖率 > 70%
- [ ] 集成测试（Binance 测试网）
- [ ] 压力测试（高并发场景）

### 6.4 文档
- [ ] README.md（项目介绍、快速开始）
- [ ] 配置说明
- [ ] 架构图
- [ ] 风险提示

**交付物：** 可部署到服务器运行的完整系统

---

## 依赖关系

```
Phase 1 (基础骨架)
    │
    ├──► Phase 2 (Tool 系统)
    │        │
    │        ├──► Phase 3 (套利引擎)
    │        │        │
    │        │        └──► Phase 4 (风险管理)
    │        │
    │        └──► Phase 5 (策略层与 Memory)
    │
    └──► Phase 6 (生产化) ◄── 依赖所有 Phase
```

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| LLM 推理延迟过高 | 套利机会丢失 | LLM 只做策略层，不参与实时执行 |
| Binance API 限流 | 无法及时获取数据 | WebSocket 优先，REST 作为降级 |
| 套利价差被手续费吃掉 | 策略不盈利 | 严格阈值过滤，BNB 折扣，VIP 等级提升 |
| LLM 幻觉/错误决策 | 资金损失 | Risk Manager 硬性拦截，不依赖 LLM 做风控 |
| Go LLM SDK 不成熟 | 开发效率低 | 使用官方 SDK，必要时直接调 HTTP API |

---

## 第一步

从 Phase 1.1 开始：项目初始化 + go mod init + 目录结构。
