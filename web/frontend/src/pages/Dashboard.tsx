import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Activity,
  BarChart3,
  Shield,
  Pause,
  Play,
} from 'lucide-react';

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

const API_URL = '/api';

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 10000); // 每 10 秒刷新
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

  const statCards = [
    {
      name: '总资产',
      value: `$${stats.total_balance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      icon: DollarSign,
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
    },
    {
      name: '今日盈亏',
      value: `${stats.daily_pnl >= 0 ? '+' : ''}$${stats.daily_pnl.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      subValue: `${stats.daily_pnl_pct >= 0 ? '+' : ''}${stats.daily_pnl_pct.toFixed(2)}%`,
      icon: stats.daily_pnl >= 0 ? TrendingUp : TrendingDown,
      color: stats.daily_pnl >= 0 ? 'text-green-500' : 'text-red-500',
      bgColor: stats.daily_pnl >= 0 ? 'bg-green-500/10' : 'bg-red-500/10',
    },
    {
      name: '持仓数量',
      value: stats.open_positions.toString(),
      icon: BarChart3,
      color: 'text-purple-500',
      bgColor: 'bg-purple-500/10',
    },
    {
      name: '胜率',
      value: `${stats.win_rate.toFixed(1)}%`,
      icon: Activity,
      color: 'text-yellow-500',
      bgColor: 'bg-yellow-500/10',
    },
  ];

  return (
    <div className="space-y-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">仪表盘</h1>
          <p className="text-gray-400">交易代理概览</p>
        </div>
        <button
          onClick={toggleTrading}
          className={`flex items-center px-4 py-2 rounded-lg font-medium transition-colors ${
            stats.risk_status.is_paused
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : 'bg-red-600 hover:bg-red-700 text-white'
          }`}
        >
          {stats.risk_status.is_paused ? (
            <>
              <Play className="h-5 w-5 mr-2" />
              恢复交易
            </>
          ) : (
            <>
              <Pause className="h-5 w-5 mr-2" />
              暂停交易
            </>
          )}
        </button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => (
          <div
            key={stat.name}
            className="bg-gray-800 rounded-xl p-6 border border-gray-700"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-400">{stat.name}</p>
                <p className="text-2xl font-bold text-white mt-1">{stat.value}</p>
                {stat.subValue && (
                  <p className={`text-sm mt-1 ${stat.color}`}>{stat.subValue}</p>
                )}
              </div>
              <div className={`${stat.bgColor} ${stat.color} p-3 rounded-lg`}>
                <stat.icon className="h-6 w-6" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 风险状态 */}
      <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <div className="flex items-center mb-4">
          <Shield className="h-6 w-6 text-blue-500 mr-3" />
          <h2 className="text-lg font-semibold text-white">风险状态</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-gray-700/50 rounded-lg p-4">
            <p className="text-sm text-gray-400">状态</p>
            <p className={`text-lg font-semibold mt-1 ${stats.risk_status.is_paused ? 'text-red-500' : 'text-green-500'}`}>
              {stats.risk_status.is_paused ? '已暂停' : '运行中'}
            </p>
            {stats.risk_status.pause_reason && (
              <p className="text-xs text-gray-500 mt-1">{stats.risk_status.pause_reason}</p>
            )}
          </div>
          <div className="bg-gray-700/50 rounded-lg p-4">
            <p className="text-sm text-gray-400">回撤</p>
            <p className={`text-lg font-semibold mt-1 ${stats.risk_status.drawdown_pct > 5 ? 'text-red-500' : 'text-white'}`}>
              {stats.risk_status.drawdown_pct.toFixed(2)}%
            </p>
          </div>
          <div className="bg-gray-700/50 rounded-lg p-4">
            <p className="text-sm text-gray-400">今日亏损</p>
            <p className={`text-lg font-semibold mt-1 ${stats.risk_status.daily_loss_pct < -5 ? 'text-red-500' : 'text-white'}`}>
              {stats.risk_status.daily_loss_pct.toFixed(2)}%
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
