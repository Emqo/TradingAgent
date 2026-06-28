import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  ArrowLeftRight,
  TrendingUp,
  RefreshCw,
  Play,
  Pause,
  Clock,
  Zap,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { formatLocalTimeShort } from '@/lib/utils';

interface ArbitrageOpportunity {
  type: string;
  path: string;
  spread_bps: number;
  profit_usdt: number;
  timestamp: string;
}

interface ArbitrageStats {
  total_opportunities: number;
  total_profit: number;
  avg_spread: number;
  success_rate: number;
}

const API_URL = '/api';

export default function Arbitrage() {
  const [scanning, setScanning] = useState(true);
  const [opportunities, setOpportunities] = useState<ArbitrageOpportunity[]>([]);
  const [stats, setStats] = useState<ArbitrageStats>({
    total_opportunities: 0,
    total_profit: 0,
    avg_spread: 0,
    success_rate: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchArbitrageData();
    const interval = setInterval(fetchArbitrageData, 10000); // 每 10 秒刷新
    return () => clearInterval(interval);
  }, []);

  const fetchArbitrageData = async () => {
    try {
      // 获取套利机会
      const oppRes = await axios.get(`${API_URL}/arbitrage/opportunities`);
      setOpportunities(oppRes.data.opportunities || []);

      // 获取套利统计
      const statsRes = await axios.get(`${API_URL}/arbitrage/stats`);
      setStats({
        total_opportunities: statsRes.data.total_opportunities || 0,
        total_profit: statsRes.data.total_profit || 0,
        avg_spread: statsRes.data.avg_spread || 0,
        success_rate: statsRes.data.success_rate || 0,
      });
    } catch (err) {
      console.error('获取套利数据失败:', err);
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">套利监控</h1>
          <p className="text-muted-foreground">
            三角套利和期现套利实时监控
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={scanning ? 'default' : 'outline'}
            onClick={() => setScanning(!scanning)}
          >
            {scanning ? (
              <>
                <Pause className="mr-2 h-4 w-4" />
                暂停扫描
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                开始扫描
              </>
            )}
          </Button>
          <Button variant="outline" size="icon" onClick={fetchArbitrageData}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日机会</CardTitle>
            <Zap className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_opportunities}</div>
            <p className="text-xs text-muted-foreground">个套利机会</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日收益</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">+${stats.total_profit.toFixed(2)}</div>
            <p className="text-xs text-muted-foreground">套利收益</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均价差</CardTitle>
            <ArrowLeftRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.avg_spread.toFixed(1)} bps</div>
            <p className="text-xs text-muted-foreground">基点</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">成功率</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.success_rate}%</div>
            <p className="text-xs text-muted-foreground">执行成功率</p>
          </CardContent>
        </Card>
      </div>

      {/* 套利机会列表 */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* 三角套利 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ArrowLeftRight className="h-5 w-5 text-blue-500" />
                <CardTitle>三角套利</CardTitle>
              </div>
              <Badge variant={scanning ? 'default' : 'secondary'}>
                {scanning ? '扫描中' : '已暂停'}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            {opportunities.filter(o => o.type === '三角套利').length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无套利机会
              </div>
            ) : (
              <div className="space-y-4">
                {opportunities
                  .filter(o => o.type === '三角套利')
                  .map((opp, i) => (
                    <div key={i} className="p-4 rounded-lg border bg-muted/50">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium">{opp.path}</span>
                        <span className="text-sm text-muted-foreground">
                          {formatLocalTimeShort(opp.timestamp)}
                        </span>
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div>
                            <p className="text-xs text-muted-foreground">价差</p>
                            <p className="font-medium">{opp.spread_bps} bps</p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">预计收益</p>
                            <p className="font-medium text-green-500">+${opp.profit_usdt}</p>
                          </div>
                        </div>
                        <Button size="sm" variant="outline">
                          执行
                        </Button>
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* 期现套利 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-green-500" />
                <CardTitle>期现套利</CardTitle>
              </div>
              <Badge variant={scanning ? 'default' : 'secondary'}>
                {scanning ? '监控中' : '已暂停'}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            {opportunities.filter(o => o.type === '期现套利').length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无套利机会
              </div>
            ) : (
              <div className="space-y-4">
                {opportunities
                  .filter(o => o.type === '期现套利')
                  .map((opp, i) => (
                    <div key={i} className="p-4 rounded-lg border bg-muted/50">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium">{opp.path}</span>
                        <span className="text-sm text-muted-foreground">
                          {formatLocalTimeShort(opp.timestamp)}
                        </span>
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div>
                            <p className="text-xs text-muted-foreground">资金费率</p>
                            <p className="font-medium">0.01%</p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">年化收益</p>
                            <p className="font-medium text-green-500">10.95%</p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">今日收益</p>
                            <p className="font-medium text-green-500">+${opp.profit_usdt}</p>
                          </div>
                        </div>
                        <Button size="sm" variant="outline">
                          详情
                        </Button>
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
