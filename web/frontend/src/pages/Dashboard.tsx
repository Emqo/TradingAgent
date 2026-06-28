import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  TrendingUp,
  TrendingDown,
  BarChart3,
  Shield,
  Pause,
  Play,
  ArrowUpRight,
  ArrowDownRight,
  Wallet,
  Target,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

interface DashboardStats {
  total_balance: number;
  daily_pnl: number;
  daily_pnl_pct: number;
  total_pnl: number;
  total_pnl_pct: number;
  open_positions: number;
  today_trades: number;
  win_rate: number;
  risk_status: {
    is_paused: boolean;
    pause_reason: string;
    drawdown_pct: number;
    daily_loss_pct: number;
    positions_count: number;
  };
}

// 模拟数据
const chartData = [
  { time: '00:00', value: 10000 },
  { time: '04:00', value: 10200 },
  { time: '08:00', value: 10150 },
  { time: '12:00', value: 10400 },
  { time: '16:00', value: 10350 },
  { time: '20:00', value: 10500 },
  { time: '24:00', value: 10450 },
];

const API_URL = '/api';

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchStats = async () => {
    try {
      const res = await axios.get(`${API_URL}/dashboard/stats`);
      setStats(res.data);
    } catch (err) {
      console.error('获取统计数据失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const toggleTrading = async () => {
    if (!stats) return;
    try {
      if (stats.risk_status.is_paused) {
        await axios.post(`${API_URL}/dashboard/resume`);
      } else {
        await axios.post(`${API_URL}/dashboard/pause`, { reason: '用户手动暂停' });
      }
      fetchStats();
    } catch (err) {
      console.error('切换交易状态失败:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="text-center text-gray-400 py-12">
        加载仪表盘数据失败
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">仪表盘</h1>
          <p className="text-muted-foreground">
            交易代理实时监控
          </p>
        </div>
        <Button
          onClick={toggleTrading}
          variant={stats.risk_status.is_paused ? 'default' : 'destructive'}
          size="lg"
        >
          {stats.risk_status.is_paused ? (
            <>
              <Play className="mr-2 h-4 w-4" />
              恢复交易
            </>
          ) : (
            <>
              <Pause className="mr-2 h-4 w-4" />
              暂停交易
            </>
          )}
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              总资产
            </CardTitle>
            <Wallet className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${stats.total_balance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </div>
            <p className="text-xs text-muted-foreground">
              +2.1% 较昨日
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              今日盈亏
            </CardTitle>
            {stats.daily_pnl >= 0 ? (
              <TrendingUp className="h-4 w-4 text-green-500" />
            ) : (
              <TrendingDown className="h-4 w-4 text-red-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${stats.daily_pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
              {stats.daily_pnl >= 0 ? '+' : ''}${stats.daily_pnl.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </div>
            <p className="text-xs text-muted-foreground">
              {stats.daily_pnl_pct >= 0 ? '+' : ''}{stats.daily_pnl_pct.toFixed(2)}% 今日
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              持仓数量
            </CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.open_positions}</div>
            <p className="text-xs text-muted-foreground">
              活跃交易对
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              胜率
            </CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.win_rate.toFixed(1)}%</div>
            <p className="text-xs text-muted-foreground">
              过去 30 天
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 图表和风险状态 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        {/* 收益曲线 */}
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>收益曲线</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="time"
                  className="text-xs"
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  className="text-xs"
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(value) => `$${value}`}
                />
                <Tooltip
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      return (
                        <div className="rounded-lg border bg-background p-2 shadow-sm">
                          <div className="grid grid-cols-2 gap-2">
                            <div className="flex flex-col">
                              <span className="text-[0.70rem] uppercase text-muted-foreground">
                                时间
                              </span>
                              <span className="font-bold text-muted-foreground">
                                {payload[0].payload.time}
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
                  stroke="#3b82f6"
                  fillOpacity={1}
                  fill="url(#colorValue)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* 风险状态 */}
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              风险状态
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">状态</span>
              <Badge variant={stats.risk_status.is_paused ? 'destructive' : 'default'}>
                {stats.risk_status.is_paused ? '已暂停' : '运行中'}
              </Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">回撤</span>
              <span className={`font-medium ${stats.risk_status.drawdown_pct > 5 ? 'text-red-500' : ''}`}>
                {stats.risk_status.drawdown_pct.toFixed(2)}%
              </span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日亏损</span>
              <span className={`font-medium ${stats.risk_status.daily_loss_pct < -5 ? 'text-red-500' : ''}`}>
                {stats.risk_status.daily_loss_pct.toFixed(2)}%
              </span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">持仓数</span>
              <span className="font-medium">{stats.risk_status.positions_count}</span>
            </div>
            {stats.risk_status.pause_reason && (
              <>
                <Separator />
                <div className="rounded-md bg-muted p-3">
                  <p className="text-sm text-muted-foreground">
                    暂停原因: {stats.risk_status.pause_reason}
                  </p>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 最近交易 */}
      <Card>
        <CardHeader>
          <CardTitle>最近交易</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[
              { time: '14:30', symbol: 'BTCUSDT', side: '买入', price: '$60,500', pnl: '+$125', positive: true },
              { time: '14:35', symbol: 'ETHUSDT', side: '卖出', price: '$1,590', pnl: '+$85', positive: true },
              { time: '15:00', symbol: 'SOLUSDT', side: '买入', price: '$72.50', pnl: '-$20', positive: false },
            ].map((trade, i) => (
              <div key={i} className="flex items-center justify-between p-4 rounded-lg border">
                <div className="flex items-center gap-4">
                  <div className={`p-2 rounded-full ${trade.positive ? 'bg-green-500/10' : 'bg-red-500/10'}`}>
                    {trade.positive ? (
                      <ArrowUpRight className="h-4 w-4 text-green-500" />
                    ) : (
                      <ArrowDownRight className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  <div>
                    <p className="font-medium">{trade.symbol}</p>
                    <p className="text-sm text-muted-foreground">{trade.time}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="font-medium">{trade.price}</p>
                  <p className={`text-sm ${trade.positive ? 'text-green-500' : 'text-red-500'}`}>
                    {trade.pnl}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
