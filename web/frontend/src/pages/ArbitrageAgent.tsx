import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  ArrowLeftRight,
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

interface ArbitrageStats {
  today_decisions: number;
  today_opportunities: number;
  today_profit: number;
  success_rate: number;
  llm_calls: number;
  tokens_used: number;
}

const API_URL = '/api';

export default function ArbitrageAgent() {
  const [running, setRunning] = useState(true);
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [stats, setStats] = useState<ArbitrageStats>({
    today_decisions: 0,
    today_opportunities: 0,
    today_profit: 0,
    success_rate: 0,
    llm_calls: 0,
    tokens_used: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchArbitrageData();
    const interval = setInterval(fetchArbitrageData, 10000); // 每 10 秒刷新
    return () => clearInterval(interval);
  }, []);

  const fetchArbitrageData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      // 获取套利决策（过滤 ARBITRAGE 类型）
      const decisionsRes = await axios.get(`${API_URL}/agent/decisions`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const allDecisions = decisionsRes.data.decisions || [];
      // 过滤套利决策
      const arbDecisions = allDecisions.filter((d: Decision) =>
        d.action?.includes('套利') || d.action?.includes('ARBITRAGE')
      );
      setDecisions(arbDecisions);

      // 获取套利统计
      const statsRes = await axios.get(`${API_URL}/arbitrage/stats`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setStats({
        today_decisions: arbDecisions.length,
        today_opportunities: statsRes.data.total_opportunities || 0,
        today_profit: statsRes.data.total_profit || 0,
        success_rate: statsRes.data.success_rate || 0,
        llm_calls: arbDecisions.length,
        tokens_used: 0,
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
            ArbitrageAgent - 套利机会检测与执行
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
                暂停套利
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                启动套利
              </>
            )}
          </Button>
          <Button variant="outline" size="icon" onClick={fetchArbitrageData}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">套利决策</CardTitle>
            <Brain className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_decisions}</div>
            <p className="text-xs text-muted-foreground">次套利分析</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">套利机会</CardTitle>
            <Zap className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_opportunities}</div>
            <p className="text-xs text-muted-foreground">个机会</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">套利收益</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">
              +${stats.today_profit.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">套利收益</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">成功率</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.success_rate}%</div>
            <p className="text-xs text-muted-foreground">套利成功率</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* 套利 Agent 状态 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <div className="flex items-center gap-2">
              <ArrowLeftRight className="h-5 w-5 text-blue-500" />
              <CardTitle>套利 Agent 状态</CardTitle>
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
              <span className="font-medium">30 秒</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日 LLM 调用</span>
              <span className="font-medium">{stats.llm_calls} 次</span>
            </div>
          </CardContent>
        </Card>

        {/* 套利决策日志 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5 text-blue-500" />
                <CardTitle>套利决策日志</CardTitle>
              </div>
              <Button variant="outline" size="sm" onClick={fetchArbitrageData}>
                <RefreshCw className="mr-2 h-4 w-4" />
                刷新
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {decisions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无套利决策记录
              </div>
            ) : (
              <div className="space-y-4">
                {decisions.map((decision, i) => (
                  <div key={i} className="p-4 rounded-lg border">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Badge variant="default">
                          {decision.action}
                        </Badge>
                        <span className="text-sm text-muted-foreground">
                          {formatLocalTimeShort(decision.time)}
                        </span>
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
            )}
          </CardContent>
        </Card>
      </div>

      {/* 套利思考过程 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Brain className="h-5 w-5 text-purple-500" />
            <CardTitle>套利 Agent 思考过程</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Textarea
            readOnly
            value="等待套利 Agent 开始分析..."
            className="min-h-[100px] font-mono text-sm"
          />
        </CardContent>
      </Card>
    </div>
  );
}
