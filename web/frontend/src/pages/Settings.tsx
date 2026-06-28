import { useState } from 'react';
import { Save } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';

const tabs = [
  { id: 'risk', name: '风险管理' },
  { id: 'arbitrage', name: '套利设置' },
  { id: 'llm', name: 'LLM 配置' },
  { id: 'exchange', name: '交易所' },
];

export default function Settings() {
  const [activeTab, setActiveTab] = useState('risk');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">设置</h1>
        <p className="text-muted-foreground">
          配置您的交易代理参数
        </p>
      </div>

      {/* 选项卡 */}
      <div className="border-b">
        <nav className="flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === tab.id
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground'
              }`}
            >
              {tab.name}
            </button>
          ))}
        </nav>
      </div>

      {/* 内容 */}
      {activeTab === 'risk' && (
        <Card>
          <CardHeader>
            <CardTitle>风险管理</CardTitle>
            <CardDescription>
              配置风险控制参数，保护您的资金安全
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="max-position">最大仓位 (USDT)</Label>
                <Input id="max-position" type="number" defaultValue={1000} />
                <p className="text-xs text-muted-foreground">
                  单笔交易的最大仓位
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="max-daily-loss">最大日亏损 (USDT)</Label>
                <Input id="max-daily-loss" type="number" defaultValue={500} />
                <p className="text-xs text-muted-foreground">
                  单日最大允许亏损金额
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="max-drawdown">最大回撤 (%)</Label>
                <Input id="max-drawdown" type="number" defaultValue={10} />
                <p className="text-xs text-muted-foreground">
                  从最高点到最低点的最大跌幅
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="max-leverage">最大杠杆</Label>
                <Input id="max-leverage" type="number" defaultValue={3} />
                <p className="text-xs text-muted-foreground">
                  允许的最大杠杆倍数
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="cooldown">亏损后冷却时间（分钟）</Label>
                <Input id="cooldown" type="number" defaultValue={5} />
                <p className="text-xs text-muted-foreground">
                  亏损后暂停交易的时间
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {activeTab === 'arbitrage' && (
        <Card>
          <CardHeader>
            <CardTitle>套利设置</CardTitle>
            <CardDescription>
              配置三角套利和期现套利参数
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="min-spread">最小价差 (bps)</Label>
                <Input id="min-spread" type="number" defaultValue={15} />
                <p className="text-xs text-muted-foreground">
                  触发套利的最小价差（基点）
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="arb-position">最大仓位 (USDT)</Label>
                <Input id="arb-position" type="number" defaultValue={1000} />
                <p className="text-xs text-muted-foreground">
                  套利交易的最大仓位
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {activeTab === 'llm' && (
        <Card>
          <CardHeader>
            <CardTitle>LLM 配置</CardTitle>
            <CardDescription>
              配置大语言模型参数
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="provider">提供商</Label>
                <Input id="provider" type="text" defaultValue="Claude" disabled />
              </div>
              <div className="space-y-2">
                <Label htmlFor="model">模型</Label>
                <Input id="model" type="text" defaultValue="mimo-v2.5-pro" />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="base-url">Base URL</Label>
                <Input id="base-url" type="text" defaultValue="https://token-plan-cn.xiaomimimo.com/anthropic" />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="api-key">API Key</Label>
                <Input id="api-key" type="password" placeholder="请输入 API Key" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="max-tokens">最大 Token 数</Label>
                <Input id="max-tokens" type="number" defaultValue={4096} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="temperature">温度</Label>
                <Input id="temperature" type="number" defaultValue={0.7} step={0.1} />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {activeTab === 'exchange' && (
        <Card>
          <CardHeader>
            <CardTitle>交易所设置</CardTitle>
            <CardDescription>
              配置交易所连接参数
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="exchange">交易所</Label>
                <Input id="exchange" type="text" defaultValue="Binance" disabled />
              </div>
              <div className="space-y-2">
                <Label htmlFor="mode">模式</Label>
                <Input id="mode" type="text" defaultValue="测试网" disabled />
              </div>
              <div className="space-y-2">
                <Label htmlFor="exchange-api-key">API Key</Label>
                <Input id="exchange-api-key" type="password" placeholder="请输入 API Key" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="exchange-api-secret">API Secret</Label>
                <Input id="exchange-api-secret" type="password" placeholder="请输入 API Secret" />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="flex justify-end">
        <Button size="lg">
          <Save className="mr-2 h-4 w-4" />
          保存设置
        </Button>
      </div>
    </div>
  );
}
