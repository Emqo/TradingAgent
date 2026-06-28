import { useState } from 'react';
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

export default function Arbitrage() {
  const [scanning, setScanning] = useState(true);
  const [opportunities] = useState<ArbitrageOpportunity[]>([
    { type: '三角套利', path: 'USDT→BTC→ETH→USDT', spread_bps: 18.5, profit_usdt: 12.50, timestamp: '2026-06-28T14:30:25+08:00' },
    { type: '三角套利', path: 'USDT→ETH→SOL→USDT', spread_bps: 15.2, profit_usdt: 8.30, timestamp: '2026-06-28T14:28:10+08:00' },
    { type: '期现套利', path: 'BTC 永续合约', spread_bps: 0, profit_usdt: 45.00, timestamp: '2026-06-28T14:00:00+08:00' },
    { type: '期现套利', path: 'ETH 永续合约', spread_bps: 0, profit_usdt: 32.00, timestamp: '2026-06-28T14:00:00+08:00' },
  ]);

  const [stats] = useState({
    total_opportunities: 4,
    total_profit: 97.80,
    avg_spread: 16.85,
    success_rate: 85,
  });

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
          <Button variant="outline" size="icon">
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
            <div className="text-2xl font-bold text-green-500">+${stats.total_profit}</div>
            <p className="text-xs text-muted-foreground">套利收益</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均价差</CardTitle>
            <ArrowLeftRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.avg_spread} bps</div>
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
            <div className="space-y-4">
              {opportunities
                .filter(o => o.type === '三角套利')
                .map((opp, i) => (
                  <div key={i} className="p-4 rounded-lg border bg-muted/50">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">{opp.path}</span>
                      <span className="text-sm text-muted-foreground">{formatLocalTimeShort(opp.timestamp)}</span>
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
            <div className="space-y-4">
              {opportunities
                .filter(o => o.type === '期现套利')
                .map((opp, i) => (
                  <div key={i} className="p-4 rounded-lg border bg-muted/50">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">{opp.path}</span>
                      <span className="text-sm text-muted-foreground">{formatLocalTimeShort(opp.timestamp)}</span>
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
          </CardContent>
        </Card>
      </div>

      {/* 配置 */}
      <Card>
        <CardHeader>
          <CardTitle>套利配置</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">最小价差</p>
              <p className="font-medium">15 bps</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">最大仓位</p>
              <p className="font-medium">$1,000</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">手续费率</p>
              <p className="font-medium">0.1%</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">BNB 折扣</p>
              <p className="font-medium">启用 (25%)</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
