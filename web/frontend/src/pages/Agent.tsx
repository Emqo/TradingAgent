import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  Bot,
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

interface AgentStats {
  today_decisions: number;
  today_trades: number;
  today_pnl: number;
  win_rate: number;
  llm_calls: number;
  tokens_used: number;
}

const API_URL = '/api';

export default function Agent() {
  const [running, setRunning] = useState(true);
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [stats, setStats] = useState<AgentStats>({
    today_decisions: 0,
    today_trades: 0,
    today_pnl: 0,
    win_rate: 0,
    llm_calls: 0,
    tokens_used: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAgentData();
    const interval = setInterval(fetchAgentData, 10000); // 每 10 秒刷新
    return () => clearInterval(interval);
  }, []);

  const fetchAgentData = async () => {
    try {
      // 获取决策日志
      const decisionsRes = await axios.get(`${API_URL}/agent/decisions`);
      setDecisions(decisionsRes.data.decisions || []);

      // 获取 Agent 统计
      const statsRes = await axios.get(`${API_URL}/agent/stats`);
      setStats({
        today_decisions: statsRes.data.today_decisions || 0,
        today_trades: statsRes.data.today_trades || 0,
        today_pnl: statsRes.data.today_pnl || 0,
        win_rate: statsRes.data.win_rate || 0,
        llm_calls: statsRes.data.llm_calls || 0,
        tokens_used: statsRes.data.tokens_used || 0,
      });
    } catch (err) {
      console.error('获取 Agent 数据失败:', err);
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
          <h1 className="text-3xl font-bold tracking-tight">Agent 监控</h1>
          <p className="text-muted-foreground">
            LLM 驱动的智能交易代理
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
                暂停 Agent
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                启动 Agent
              </>
            )}
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日决策</CardTitle>
            <Brain className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_decisions}</div>
            <p className="text-xs text-muted-foreground">次分析决策</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日交易</CardTitle>
            <Zap className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today_trades}</div>
            <p className="text-xs text-muted-foreground">笔执行交易</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日收益</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${stats.today_pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
              {stats.today_pnl >= 0 ? '+' : ''}${stats.today_pnl.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">Agent 交易收益</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">胜率</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.win_rate}%</div>
            <p className="text-xs text-muted-foreground">过去 30 天</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Agent 状态 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bot className="h-5 w-5 text-purple-500" />
              <CardTitle>Agent 状态</CardTitle>
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
              <span className="font-medium">1 分钟</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日 LLM 调用</span>
              <span className="font-medium">{stats.llm_calls} 次</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">今日 Token</span>
              <span className="font-medium">{(stats.tokens_used / 1000).toFixed(0)}K</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">当前持仓</span>
              <span className="font-medium">-</span>
            </div>
          </CardContent>
        </Card>

        {/* 决策日志 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5 text-blue-500" />
                <CardTitle>决策日志</CardTitle>
              </div>
              <Button variant="outline" size="sm" onClick={fetchAgentData}>
                <RefreshCw className="mr-2 h-4 w-4" />
                刷新
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {decisions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无决策记录
              </div>
            ) : (
              <div className="space-y-4">
                {decisions.map((decision, i) => (
                  <div key={i} className="p-4 rounded-lg border">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Badge variant={decision.pnl > 0 ? 'default' : decision.pnl < 0 ? 'destructive' : 'secondary'}>
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

      {/* Agent 思考过程 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Brain className="h-5 w-5 text-purple-500" />
            <CardTitle>Agent 思考过程</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Textarea
            readOnly
            value="等待 Agent 开始决策..."
            className="min-h-[100px] font-mono text-sm"
          />
        </CardContent>
      </Card>
    </div>
  );
}
