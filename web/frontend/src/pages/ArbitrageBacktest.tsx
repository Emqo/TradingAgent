import { useState } from 'react';
import {
  Play,
  TrendingUp,
  TrendingDown,
  Target,
  Percent,
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
  { date: '2024-02', value: 10150 },
  { date: '2024-03', value: 10280 },
  { date: '2024-04', value: 10420 },
  { date: '2024-05', value: 10550 },
  { date: '2024-06', value: 10680 },
  { date: '2024-07', value: 10820 },
  { date: '2024-08', value: 10950 },
  { date: '2024-09', value: 11080 },
  { date: '2024-10', value: 11220 },
  { date: '2024-11', value: 11350 },
  { date: '2024-12', value: 11485 },
];

const trades = [
  { time: '2024-01-15T14:30:00+08:00', type: '三角套利', path: 'USDT→BTC→ETH→USDT', spread: '18.5 bps', pnl: '+$12.50', positive: true },
  { time: '2024-01-15T15:00:00+08:00', type: '三角套利', path: 'USDT→ETH→SOL→USDT', spread: '15.2 bps', pnl: '+$8.30', positive: true },
  { time: '2024-01-16T09:00:00+08:00', type: '期现套利', path: 'BTC 永续合约', spread: '0.01%', pnl: '+$45.00', positive: true },
  { time: '2024-01-16T10:30:00+08:00', type: '三角套利', path: 'USDT→BTC→BNB→USDT', spread: '12.8 bps', pnl: '-$3.20', positive: false },
  { time: '2024-01-16T14:00:00+08:00', type: '期现套利', path: 'ETH 永续合约', spread: '0.01%', pnl: '+$32.00', positive: true },
];

export default function ArbitrageBacktest() {
  const [running, setRunning] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">套利回测</h1>
        <p className="text-muted-foreground">
          测试套利策略的历史表现
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* 配置 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>回测配置</CardTitle>
            <CardDescription>设置套利回测参数</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>套利类型</Label>
              <Input defaultValue="三角套利 + 期现套利" />
            </div>
            <div className="space-y-2">
              <Label>交易对</Label>
              <Input defaultValue="BTC, ETH, SOL, BNB" />
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
              <Label>最小价差 (bps)</Label>
              <Input type="number" defaultValue={15} />
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
                <div className="text-2xl font-bold text-green-500">+14.85%</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">夏普比率</CardTitle>
                <Target className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">2.15</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">最大回撤</CardTitle>
                <TrendingDown className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-500">-2.8%</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">成功率</CardTitle>
                <Percent className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">85%</div>
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
                  <Area type="monotone" dataKey="value" stroke="#22c55e" fillOpacity={1} fill="url(#colorValue)" />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* 交易记录 */}
          <Card>
            <CardHeader>
              <CardTitle>套利记录</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {trades.map((trade, i) => (
                  <div key={i} className="flex items-center justify-between p-4 rounded-lg border">
                    <div className="flex items-center gap-4">
                      <Badge variant={trade.type === '三角套利' ? 'default' : 'secondary'}>
                        {trade.type}
                      </Badge>
                      <div>
                        <p className="font-medium">{trade.path}</p>
                        <p className="text-sm text-muted-foreground">{trade.time}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-muted-foreground">价差: {trade.spread}</p>
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
