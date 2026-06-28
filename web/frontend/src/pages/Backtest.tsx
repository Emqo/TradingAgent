import { useState } from 'react';
import { Play, TrendingUp, TrendingDown, Target, Percent } from 'lucide-react';
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
  { date: '2024-02', value: 10500 },
  { date: '2024-03', value: 10200 },
  { date: '2024-04', value: 10800 },
  { date: '2024-05', value: 11200 },
  { date: '2024-06', value: 11000 },
  { date: '2024-07', value: 11500 },
  { date: '2024-08', value: 12000 },
  { date: '2024-09', value: 11800 },
  { date: '2024-10', value: 12500 },
  { date: '2024-11', value: 13000 },
  { date: '2024-12', value: 12450 },
];

const trades = [
  { time: '2024-01-15 14:30', symbol: 'BTCUSDT', side: '买入', price: '$42,150', quantity: '0.1', pnl: '+$125', positive: true },
  { time: '2024-01-15 14:35', symbol: 'ETHUSDT', side: '卖出', price: '$2,580', quantity: '1.5', pnl: '+$85', positive: true },
  { time: '2024-01-15 15:00', symbol: 'BTCUSDT', side: '卖出', price: '$42,300', quantity: '0.1', pnl: '+$150', positive: true },
  { time: '2024-01-16 09:15', symbol: 'SOLUSDT', side: '买入', price: '$98.50', quantity: '10', pnl: '-$20', positive: false },
  { time: '2024-01-16 10:30', symbol: 'ETHUSDT', side: '买入', price: '$2,550', quantity: '2', pnl: '+$180', positive: true },
];

export default function Backtest() {
  const [running, setRunning] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">回测</h1>
        <p className="text-muted-foreground">
          使用历史数据测试您的策略
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* 配置 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>回测配置</CardTitle>
            <CardDescription>
              设置回测参数
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="strategy">策略</Label>
              <Input id="strategy" type="text" defaultValue="三角套利" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="symbol">交易对</Label>
              <Input id="symbol" type="text" defaultValue="BTCUSDT" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="start-date">开始日期</Label>
              <Input id="start-date" type="date" defaultValue="2024-01-01" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="end-date">结束日期</Label>
              <Input id="end-date" type="date" defaultValue="2024-12-31" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="initial-capital">初始资金 (USDT)</Label>
              <Input id="initial-capital" type="number" defaultValue={10000} />
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
                <div className="text-2xl font-bold text-green-500">+24.5%</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">夏普比率</CardTitle>
                <Target className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">1.85</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">最大回撤</CardTitle>
                <TrendingDown className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-500">-8.2%</div>
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
                      <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#22c55e" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis
                    dataKey="date"
                    className="text-xs"
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    className="text-xs"
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`}
                  />
                  <Tooltip
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        return (
                          <div className="rounded-lg border bg-background p-2 shadow-sm">
                            <div className="grid grid-cols-2 gap-2">
                              <div className="flex flex-col">
                                <span className="text-[0.70rem] uppercase text-muted-foreground">
                                  日期
                                </span>
                                <span className="font-bold text-muted-foreground">
                                  {payload[0].payload.date}
                                </span>
                              </div>
                              <div className="flex flex-col">
                                <span className="text-[0.70rem] uppercase text-muted-foreground">
                                  价值
                                </span>
                                <span className="font-bold">
                                  ${payload[0].value?.toLocaleString()}
                                </span>
                              </div>
                            </div>
                          </div>
                        )
                      }
                      return null
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="value"
                    stroke="#22c55e"
                    fillOpacity={1}
                    fill="url(#colorValue)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* 交易记录 */}
          <Card>
            <CardHeader>
              <CardTitle>交易记录</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {trades.map((trade, i) => (
                  <div key={i} className="flex items-center justify-between p-4 rounded-lg border">
                    <div className="flex items-center gap-4">
                      <Badge variant={trade.positive ? 'default' : 'destructive'}>
                        {trade.side}
                      </Badge>
                      <div>
                        <p className="font-medium">{trade.symbol}</p>
                        <p className="text-sm text-muted-foreground">{trade.time}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-medium">{trade.price}</p>
                      <p className="text-sm text-muted-foreground">数量: {trade.quantity}</p>
                    </div>
                    <div className="text-right">
                      <p className={`font-medium ${trade.positive ? 'text-green-500' : 'text-red-500'}`}>
                        {trade.pnl}
                      </p>
                    </div>
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
