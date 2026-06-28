import { useState } from 'react';
import {
  Bot,
  Brain,
  MessageSquare,
  TrendingUp,
  Play,
  Pause,
  RefreshCw,
  Clock,
  Zap,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { formatLocalTimeShort } from '@/lib/utils';

interface Decision {
  time: string;
  action: string;
  reason: string;
  result: string;
  pnl: number;
}

export default function Agent() {
  const [running, setRunning] = useState(true);
  const [decisions] = useState<Decision[]>([
    {
      time: '2026-06-28T14:30:25+08:00',
      action: '买入 BTCUSDT',
      reason: 'BTC 在 $60,500 形成支撑，RSI 超卖，预计反弹',
      result: '成功买入 0.1 BTC @ $60,500',
      pnl: 0,
    },
    {
      time: '2026-06-28T14:25:10+08:00',
      action: '卖出 ETHUSDT',
      reason: 'ETH 触及阻力位 $1,600，获利了结',
      result: '成功卖出 1.5 ETH @ $1,595',
      pnl: 85,
    },
    {
      time: '2026-06-28T14:20:05+08:00',
      action: '持有',
      reason: '市场震荡，等待明确信号',
      result: '无操作',
      pnl: 0,
    },
  ]);

  const [stats] = useState({
    today_decisions: 24,
    today_trades: 8,
    today_pnl: 185,
    win_rate: 62,
    llm_calls: 48,
    tokens_used: 125000,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Agent 监控</h1>
          <p className="text-muted-foreground">
            LLM 驱动的智能交易代理
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={running ? 'default' : 'outline'}
            onClick={() => setRunning(!running)}
          >
            {running ? (
              <>
                <Pause className="mr-2 h-4 w-4" />
                暂停 Agent
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                启动 Agent
              </>
            )}
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日决策</CardTitle>
            <Brain className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_decisions}</div>
            <p className="text-xs text-muted-foreground">次分析决策</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日交易</CardTitle>
            <Zap className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_trades}</div>
            <p className="text-xs text-muted-foreground">笔执行交易</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日收益</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">+${stats.today_pnl}</div>
            <p className="text-xs text-muted-foreground">Agent 交易收益</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">胜率</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.win_rate}%</div>
            <p className="text-xs text-muted-foreground">过去 30 天</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Agent 状态 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bot className="h-5 w-5 text-purple-500" />
              <CardTitle>Agent 状态</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">状态</span>
              <Badge variant={running ? 'default' : 'secondary'}>
                {running ? '运行中' : '已暂停'}
              </Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">LLM 模型</span>
              <span className="font-medium">mimo-v2.5-pro</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">决策周期</span>
              <span className="font-medium">1 分钟</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日 LLM 调用</span>
              <span className="font-medium">{stats.llm_calls} 次</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日 Token</span>
              <span className="font-medium">{(stats.tokens_used / 1000).toFixed(0)}K</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">当前持仓</span>
              <span className="font-medium">3 个</span>
            </div>
          </CardContent>
        </Card>

        {/* 决策日志 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5 text-blue-500" />
                <CardTitle>决策日志</CardTitle>
              </div>
              <Button variant="outline" size="sm">
                <RefreshCw className="mr-2 h-4 w-4" />
                刷新
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {decisions.map((decision, i) => (
                <div key={i} className="p-4 rounded-lg border">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Badge variant={decision.pnl > 0 ? 'default' : decision.pnl < 0 ? 'destructive' : 'secondary'}>
                        {decision.action}
                      </Badge>
                      <span className="text-sm text-muted-foreground">{formatLocalTimeShort(decision.time)}</span>
                    </div>
                    {decision.pnl !== 0 && (
                      <span className={`font-medium ${decision.pnl > 0 ? 'text-green-500' : 'text-red-500'}`}>
                        {decision.pnl > 0 ? '+' : ''}${decision.pnl}
                      </span>
                    )}
                  </div>
                  <p className="text-sm mb-2">{decision.reason}</p>
                  <p className="text-sm text-muted-foreground">{decision.result}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Agent 思考过程 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Brain className="h-5 w-5 text-purple-500" />
            <CardTitle>Agent 思考过程</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Textarea
            readOnly
            value={`当前市场分析：

BTC 在 $60,500 附近形成强支撑，4 小时图显示 RSI 从超卖区域反弹。成交量略有放大，显示买盘介入。

ETH 走势相对强势，ETH/BTC 比率上升，显示资金从 BTC 流向 ETH。

建议操作：
1. BTCUSDT：在 $60,500 附近买入，止损 $59,800，目标 $62,000
2. ETHUSDT：持有现有仓位，等待突破 $1,600

风险提示：
- 美联储议息会议临近，市场波动可能加大
- 注意控制仓位，单笔不超过总资产的 10%

[决策时间: 2026-06-28 14:30:25]`}
            className="min-h-[200px] font-mono text-sm"
          />
        </CardContent>
      </Card>
    </div>
  );
}
