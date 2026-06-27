# CLAUDE.md

## Agent skills

### Issue tracker

GitHub Issues，外部 PR 不纳入 triage。See `docs/agents/issue-tracker.md`.

### Triage labels

使用五个标准标签：needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix。See `docs/agents/triage-labels.md`.

### Domain docs

单上下文 — 根目录一个 CONTEXT.md + docs/adr/。See `docs/agents/domain.md`.

---

## Skill 调用指南

**核心原则：在开始任何实质性工作之前，先检查是否有现成的 skill 适用。** 使用 `/using-superpowers` 来强化这一习惯。

### 可用 Skills 完整列表

#### 元技能

| Skill | 用途 |
|-------|------|
| `/using-superpowers` | 强化"先找 skill"的意识，适用于所有场景 |
| `/setup-matt-pocock-skills` | 初始化项目配置（已执行） |

#### 规划与设计

| Skill | 用途 |
|-------|------|
| `/brainstorming` | 模糊需求 → 发散探索 → 收敛方案 |
| `/writing-plans` | 制定实现计划，先想清楚再动手 |
| `/backend-engineering` | 设计 Agent 架构、Tool 接口、API 等后端系统 |
| `/frontend-design` | 设计前端界面（如有 Web UI 需求） |

#### 实现

| Skill | 用途 |
|-------|------|
| `/test-driven-development` | TDD：先写测试再实现，保证代码质量 |
| `/executing-early` | 最小可运行版本先跑通，快速验证可行性 |
| `/subagent-driven-development` | 复杂子系统交给子代理独立开发 |
| `/dispatching-parallel-agents` | 并行处理多个独立任务 |

#### 测试与验证

| Skill | 用途 |
|-------|------|
| `/test-automation` | 编写和运行测试 |
| `/verification-before-completion` | 完成任务前必须验证，通过才算完成 |

#### 调试

| Skill | 用途 |
|-------|------|
| `/systematic-debugging` | 系统性排查 bug，避免瞎猜 |

#### 代码审查

| Skill | 用途 |
|-------|------|
| `/requesting-code-review` | 完成功能后主动请求审查 |
| `/receiving-code-review` | 高效处理审查反馈 |
| `/code-review-agent` | 子代理自动代码审查 |

#### 分支与提交

| Skill | 用途 |
|-------|------|
| `/finishing-a-development-branch` | 功能完成，准备合并 |
| `/using-git-worktrees` | 并行开发多个功能，互不干扰 |

#### 文档与知识

| Skill | 用途 |
|-------|------|
| `/writing-playbooks` | 记录操作流程和最佳实践 |
| `/writing-skills` | 创建新的 skill |

#### 子代理

| Skill | 用途 |
|-------|------|
| `Explore` | 代码库探索和研究（子代理类型） |
| `Plan` | 复杂实现的战略规划（子代理类型） |

---

## 项目上下文

TradingAgent 是 LLM 驱动的加密货币交易智能体（详见 CONTEXT.md）。典型开发任务：

- **Agent 核心**：决策循环（Observe → Think → Act → Remember）、LLM 集成、Prompt 工程
- **Tool 系统**：市场数据、订单执行、风控检查等 Tool 的定义与实现
- **Memory 系统**：短期/长期/工作记忆的存储与检索
- **市场数据**：交易所 API 对接、K线/订单簿数据处理
- **风险管理**：风控约束作为 Tool 前置检查
- **回测系统**：模拟 Agent 决策循环、评估指标计算

## Skill 组合示例

| 任务 | 推荐 Skill 组合 |
|------|-----------------|
| 新建一个 Tool | `/test-driven-development` → `/executing-early` → `/verification-before-completion` |
| 设计 Agent 架构 | `/brainstorming` → `/writing-plans` → `/backend-engineering` |
| 修 bug | `/systematic-debugging` → `/verification-before-completion` |
| 完成功能 | `/verification-before-completion` → `/requesting-code-review` → `/finishing-a-development-branch` |
| 大规模重构 | `/writing-plans` → `/dispatching-parallel-agents` → `/verification-before-completion` |
| 探索新技术方案 | `Explore` 子代理 → `/brainstorming` → `/writing-plans` |
