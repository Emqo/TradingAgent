import { useState } from 'react';
import {
  Bot,
  Play,
  TrendingUp,
  TrendingDown,
  Target,
  Percent,
  Brain,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

const performanceData = [
  { date: '2024-01', value: 10000 },
  { date: '2024-02', value: 10800 },
  { date: '2024-03', value: 10500 },
  { date: '2024-04', value: 11200 },
  { date: '2024-05', value: 11800 },
  { date: '2024-06', value: 11500 },
  { date: '2024-07', value: 12200 },
  { date: '2024-08', value: 12800 },
  { date: '2024-09', value: 12500 },
  { date: '2024-10', value: 13200 },
  { date: '2024-11', value: 13800 },
  { date: '2024-12', value: 13450 },
];

const decisions = [
  { time: '2024-01-15 14:30', action: '买入 BTCUSDT', reason: 'RSI 超卖，支撑位反弹', pnl: '+$250', positive: true },
  { time: '2024-01-16 10:00', action: '卖出 ETHUSDT', reason: '触及阻力位，获利了结', pnl: '+$180', positive: true },
  { time: '2024-01-17 09:30', action: '买入 SOLUSDT', reason: '突破下降趋势线', pnl: '+$120', positive: true },
  { time: '2024-01-18 14:00', action: '卖出 BTCUSDT', reason: '市场转弱，止损出场', pnl: '-$80', positive: false },
  { time: '2024-01-19 11:00', action: '买入 ETHUSDT', reason: '资金费率转正，做多', pnl: '+$150', positive: true },
];

export default function AgentBacktest() {
  const [running, setRunning] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Agent 回测</h1>
        <p className="text-muted-foreground">
          测试 Agent 交易策略的历史表现
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* 配置 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>回测配置</CardTitle>
            <CardDescription>设置 Agent 回测参数</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>LLM 模型</Label>
              <Input defaultValue="mimo-v2.5-pro" />
            </div>
            <div className="space-y-2">
              <Label>交易对</Label>
              <Input defaultValue="BTCUSDT, ETHUSDT, SOLUSDT" />
            </div>
            <div className="space-y-2">
              <Label>开始日期</Label>
              <Input type="date" defaultValue="2024-01-01" />
            </div>
            <div className="space-y-2">
              <Label>结束日期</Label>
              <Input type="date" defaultValue="2024-12-31" />
            </div>
            <div className="space-y-2">
              <Label>初始资金 (USDT)</Label>
              <Input type="number" defaultValue={10000} />
            </div>
            <div className="space-y-2">
              <Label>决策周期</Label>
              <Input defaultValue="1 分钟" />
            </div>
            <div className="space-y-2">
              <Label>最大 Token 数</Label>
              <Input type="number" defaultValue={4096} />
            </div>
            <Button
              onClick={() => setRunning(!running)}
              className="w-full"
              variant={running ? 'destructive' : 'default'}
            >
              {running ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                  运行中...
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  开始回测
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        {/* 结果 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 统计卡片 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">总收益</CardTitle>
                <TrendingUp className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-500">+34.5%</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">夏普比率</CardTitle>
                <Target className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">1.92</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">最大回撤</CardTitle>
                <TrendingDown className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-500">-8.5%</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">胜率</CardTitle>
                <Percent className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">62%</div>
              </CardContent>
            </Card>
          </div>

          {/* LLM 统计 */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Brain className="h-5 w-5 text-purple-500" />
                <CardTitle>LLM 统计</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">总决策次数</p>
                  <p className="text-xl font-bold">1,245</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">总交易次数</p>
                  <p className="text-xl font-bold">156</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">总 Token 消耗</p>
                  <p className="text-xl font-bold">2.5M</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">LLM 成本</p>
                  <p className="text-xl font-bold">$12.50</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 收益曲线 */}
          <Card>
            <CardHeader>
              <CardTitle>收益曲线</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <AreaChart data={performanceData}>
                  <defs>
                    <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="date" className="text-xs" tickLine={false} axisLine={false} />
                  <YAxis className="text-xs" tickLine={false} axisLine={false} tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`} />
                  <Tooltip
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        return (
                          <div className="rounded-lg border bg-background p-2 shadow-sm">
                            <p className="font-bold">${payload[0].value?.toLocaleString()}</p>
                          </div>
                        );
                      }
                      return null;
                    }}
                  />
                  <Area type="monotone" dataKey="value" stroke="#8b5cf6" fillOpacity={1} fill="url(#colorValue)" />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* 决策记录 */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Bot className="h-5 w-5 text-purple-500" />
                <CardTitle>Agent 决策记录</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {decisions.map((decision, i) => (
                  <div key={i} className="p-4 rounded-lg border">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Badge variant={decision.positive ? 'default' : 'destructive'}>
                          {decision.action}
                        </Badge>
                        <span className="text-sm text-muted-foreground">{decision.time}</span>
                      </div>
                      <span className={`font-medium ${decision.positive ? 'text-green-500' : 'text-red-500'}`}>
                        {decision.pnl}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">{decision.reason}</p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
