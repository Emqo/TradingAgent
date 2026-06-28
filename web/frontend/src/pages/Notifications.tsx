import { useState } from 'react';
import { Save, MessageSquare, Mail } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';

export default function Notifications() {
  const [telegramEnabled, setTelegramEnabled] = useState(false);
  const [emailEnabled, setEmailEnabled] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">通知设置</h1>
        <p className="text-muted-foreground">
          配置通知渠道，及时获取交易和风险信息
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Telegram */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5 text-blue-500" />
                <CardTitle>Telegram</CardTitle>
              </div>
              <Switch
                checked={telegramEnabled}
                onCheckedChange={setTelegramEnabled}
              />
            </div>
            <CardDescription>
              通过 Telegram 机器人接收实时通知
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="telegram-token">机器人 Token</Label>
              <Input
                id="telegram-token"
                type="password"
                placeholder="请输入 Telegram 机器人 Token"
                disabled={!telegramEnabled}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="telegram-chat-id">聊天 ID</Label>
              <Input
                id="telegram-chat-id"
                type="text"
                placeholder="请输入聊天 ID"
                disabled={!telegramEnabled}
              />
            </div>
            <Separator />
            <div className="space-y-3">
              <p className="text-sm font-medium">通知类型</p>
              <div className="space-y-2">
                {[
                  { label: '交易执行', description: '当交易被执行时通知' },
                  { label: '风险告警', description: '当触发风控规则时通知' },
                  { label: '套利机会', description: '当发现套利机会时通知' },
                  { label: '每日汇总', description: '每日交易和收益汇总' },
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium">{item.label}</p>
                      <p className="text-xs text-muted-foreground">{item.description}</p>
                    </div>
                    <Switch disabled={!telegramEnabled} defaultChecked={item.label !== '每日汇总'} />
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 邮件 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Mail className="h-5 w-5 text-green-500" />
                <CardTitle>邮件</CardTitle>
              </div>
              <Switch
                checked={emailEnabled}
                onCheckedChange={setEmailEnabled}
              />
            </div>
            <CardDescription>
              通过邮件接收通知和报告
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="smtp-server">SMTP 服务器</Label>
              <Input
                id="smtp-server"
                type="text"
                placeholder="smtp.gmail.com"
                disabled={!emailEnabled}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="from-email">发件人邮箱</Label>
                <Input
                  id="from-email"
                  type="email"
                  placeholder="your@email.com"
                  disabled={!emailEnabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="to-email">收件人邮箱</Label>
                <Input
                  id="to-email"
                  type="email"
                  placeholder="recipient@email.com"
                  disabled={!emailEnabled}
                />
              </div>
            </div>
            <Separator />
            <div className="space-y-3">
              <p className="text-sm font-medium">通知类型</p>
              <div className="space-y-2">
                {[
                  { label: '交易执行', description: '当交易被执行时通知' },
                  { label: '风险告警', description: '当触发风控规则时通知' },
                  { label: '每日汇总', description: '每日交易和收益汇总' },
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium">{item.label}</p>
                      <p className="text-xs text-muted-foreground">{item.description}</p>
                    </div>
                    <Switch disabled={!emailEnabled} defaultChecked={item.label !== '每日汇总'} />
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="flex justify-end">
        <Button size="lg">
          <Save className="mr-2 h-4 w-4" />
          保存通知设置
        </Button>
      </div>
    </div>
  );
}
