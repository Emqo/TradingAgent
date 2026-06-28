import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  TrendingUp,
  TrendingDown,
  ArrowLeftRight,
  Bot,
  Wallet,
  Shield,
  BarChart3,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
// Charts will be added when historical data is available

interface DashboardStats {
  total_balance: number;
  daily_pnl: number;
  daily_pnl_pct: number;
  risk_status: {
    is_paused: boolean;
    drawdown_pct: number;
  };
}

interface ArbitrageStats {
  total_opportunities: number;
  total_profit: number;
}

interface AgentStats {
  today_decisions: number;
  today_pnl: number;
}

const API_URL = '/api';

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [arbStats, setArbStats] = useState<ArbitrageStats | null>(null);
  const [agentStats, setAgentStats] = useState<AgentStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchStats = async () => {
    try {
      // 获取 Dashboard 统计
      const dashRes = await axios.get(`${API_URL}/dashboard/stats`);
      setStats(dashRes.data);

      // 获取套利统计
      try {
        const arbRes = await axios.get(`${API_URL}/arbitrage/stats`);
        setArbStats(arbRes.data);
      } catch {
        setArbStats({ total_opportunities: 0, total_profit: 0 });
      }

      // 获取 Agent 统计
      try {
        const agentRes = await axios.get(`${API_URL}/agent/stats`);
        setAgentStats({
          today_decisions: agentRes.data.today_decisions || 0,
          today_pnl: agentRes.data.today_pnl || 0,
        });
      } catch {
        setAgentStats({ today_decisions: 0, today_pnl: 0 });
      }
    } catch (err) {
      console.error('获取统计数据失败:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">总览</h1>
        <p className="text-muted-foreground">
          TradingAgent 运行状态概览
        </p>
      </div>

      {/* 总体统计 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总资产</CardTitle>
            <Wallet className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${stats?.total_balance?.toLocaleString(undefined, { minimumFractionDigits: 2 }) || '0.00'}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日盈亏</CardTitle>
            {(stats?.daily_pnl ?? 0) >= 0 ? (
              <TrendingUp className="h-4 w-4 text-green-500" />
            ) : (
              <TrendingDown className="h-4 w-4 text-red-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${(stats?.daily_pnl ?? 0) >= 0 ? 'text-green-500' : 'text-red-500'}`}>
              {(stats?.daily_pnl ?? 0) >= 0 ? '+' : ''}${stats?.daily_pnl?.toLocaleString(undefined, { minimumFractionDigits: 2 }) || '0.00'}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">系统状态</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Badge variant={stats?.risk_status?.is_paused ? 'destructive' : 'default'}>
              {stats?.risk_status?.is_paused ? '已暂停' : '运行中'}
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">回撤</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${(stats?.risk_status?.drawdown_pct ?? 0) > 5 ? 'text-red-500' : ''}`}>
              {stats?.risk_status?.drawdown_pct?.toFixed(2) || '0.00'}%
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
          <div className="h-[300px] flex items-center justify-center text-muted-foreground">
            暂无历史数据
          </div>
        </CardContent>
      </Card>

      {/* 策略概览 */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <ArrowLeftRight className="h-5 w-5 text-blue-500" />
              <CardTitle>套利策略</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">今日机会</span>
                <span className="font-medium">{arbStats?.total_opportunities || 0} 个</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">今日收益</span>
                <span className="font-medium text-green-500">
                  +${(arbStats?.total_profit || 0).toFixed(2)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bot className="h-5 w-5 text-purple-500" />
              <CardTitle>Agent 交易</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">LLM 模型</span>
                <span className="font-medium">mimo-v2.5-pro</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">今日决策</span>
                <span className="font-medium">{agentStats?.today_decisions || 0} 次</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">今日收益</span>
                <span className="font-medium text-green-500">
                  +${(agentStats?.today_pnl || 0).toFixed(2)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
