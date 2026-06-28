import { useState } from 'react';
import axios from 'axios';
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
import { formatLocalTimeShort } from '@/lib/utils';

const API_URL = '/api';

interface BacktestResult {
  result: {
    total_return: number;
    total_return_pct: number;
    sharpe_ratio: number;
    max_drawdown: number;
    max_drawdown_pct: number;
    win_rate: number;
    total_trades: number;
    winning_trades: number;
    losing_trades: number;
    profit_factor: number;
    avg_trade_pnl: number;
  };
  trades: Array<{
    time: string;
    symbol: string;
    side: string;
    price: number;
    quantity: number;
    pnl: number;
    running_pnl: number;
  }>;
}

export default function ArbitrageBacktest() {
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<BacktestResult | null>(null);
  const [error, setError] = useState('');

  const [config, setConfig] = useState({
    strategy: 'triangular',
    symbol: 'BTCUSDT',
    start_time: '2024-01-01',
    end_time: '2024-01-07',
    initial_usdt: 10000,
  });

  const runBacktest = async () => {
    setRunning(true);
    setError('');
    setResult(null);

    try {
      // Get token
      const token = localStorage.getItem('token');
      if (!token) {
        setError('请先登录');
        return;
      }

      const res = await axios.post(`${API_URL}/backtest/run`, config, {
        headers: { Authorization: `Bearer ${token}` },
      });

      setResult(res.data);
    } catch (err: any) {
      setError(err.response?.data?.error || '回测失败');
    } finally {
      setRunning(false);
    }
  };

  // Prepare chart data
  const chartData = result?.trades?.map(t => ({
    time: t.time,
    value: result.result.total_return > 0
      ? 10000 + t.running_pnl
      : 10000 + t.running_pnl,
  })) || [];

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
              <select
                className="w-full px-3 py-2 bg-background border rounded-md"
                value={config.strategy}
                onChange={(e) => setConfig({ ...config, strategy: e.target.value })}
              >
                <option value="triangular">三角套利</option>
                <option value="cash_and_carry">期现套利</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label>交易对</Label>
              <Input
                value={config.symbol}
                onChange={(e) => setConfig({ ...config, symbol: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>开始日期</Label>
              <Input
                type="date"
                value={config.start_time}
                onChange={(e) => setConfig({ ...config, start_time: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>结束日期</Label>
              <Input
                type="date"
                value={config.end_time}
                onChange={(e) => setConfig({ ...config, end_time: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>初始资金 (USDT)</Label>
              <Input
                type="number"
                value={config.initial_usdt}
                onChange={(e) => setConfig({ ...config, initial_usdt: Number(e.target.value) })}
              />
            </div>

            {error && (
              <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20 text-destructive text-sm">
                {error}
              </div>
            )}

            <Button
              onClick={runBacktest}
              disabled={running}
              className="w-full"
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
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className={`text-2xl font-bold ${result ? (result.result.total_return >= 0 ? 'text-green-500' : 'text-red-500') : 'text-muted-foreground'}`}>
                  {result ? `${result.result.total_return >= 0 ? '+' : ''}${result.result.total_return_pct.toFixed(2)}%` : '--'}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">夏普比率</CardTitle>
                <Target className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-muted-foreground">
                  {result ? result.result.sharpe_ratio.toFixed(2) : '--'}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">最大回撤</CardTitle>
                <TrendingDown className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className={`text-2xl font-bold ${result ? 'text-red-500' : 'text-muted-foreground'}`}>
                  {result ? `-${result.result.max_drawdown_pct.toFixed(2)}%` : '--'}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">胜率</CardTitle>
                <Percent className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-muted-foreground">
                  {result ? `${result.result.win_rate.toFixed(1)}%` : '--'}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* 收益曲线 */}
          <Card>
            <CardHeader>
              <CardTitle>收益曲线</CardTitle>
            </CardHeader>
            <CardContent>
              {chartData.length === 0 ? (
                <div className="h-[300px] flex items-center justify-center text-muted-foreground">
                  运行回测后显示收益曲线
                </div>
              ) : (
                <ResponsiveContainer width="100%" height={300}>
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#22c55e" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis
                      dataKey="time"
                      className="text-xs"
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={(value) => value.split(' ')[1] || value}
                    />
                    <YAxis
                      className="text-xs"
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={(value) => `$${Number(value).toFixed(0)}`}
                    />
                    <Tooltip
                      content={({ active, payload }) => {
                        if (active && payload && payload.length) {
                          const value = Number(payload[0].value) || 0;
                          return (
                            <div className="rounded-lg border bg-background p-2 shadow-sm">
                              <p className="text-xs text-muted-foreground">{payload[0].payload.time}</p>
                              <p className="font-bold">${value.toFixed(2)}</p>
                            </div>
                          );
                        }
                        return null;
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
              )}
            </CardContent>
          </Card>

          {/* 交易记录 */}
          <Card>
            <CardHeader>
              <CardTitle>套利记录</CardTitle>
            </CardHeader>
            <CardContent>
              {!result || result.trades.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  运行回测后显示套利记录
                </div>
              ) : (
                <div className="space-y-2 max-h-[400px] overflow-y-auto">
                  {result.trades.slice(0, 50).map((trade, i) => (
                    <div key={i} className="flex items-center justify-between p-3 rounded-lg border text-sm">
                      <div className="flex items-center gap-3">
                        <Badge variant={trade.pnl >= 0 ? 'default' : 'destructive'}>
                          {trade.side}
                        </Badge>
                        <span className="text-muted-foreground">{formatLocalTimeShort(trade.time)}</span>
                      </div>
                      <div className="flex items-center gap-4">
                        <span>${trade.price.toFixed(2)}</span>
                        <span className={`font-medium ${trade.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                          {trade.pnl >= 0 ? '+' : ''}${trade.pnl.toFixed(2)}
                        </span>
                      </div>
                    </div>
                  ))}
                  {result.trades.length > 50 && (
                    <p className="text-center text-sm text-muted-foreground py-2">
                      显示前 50 条，共 {result.trades.length} 条
                    </p>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
