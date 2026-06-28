import { useState } from 'react';
import { Save, MessageSquare, Mail } from 'lucide-react';

export default function Notifications() {
  const [telegramEnabled, setTelegramEnabled] = useState(false);
  const [emailEnabled, setEmailEnabled] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">通知设置</h1>
        <p className="text-gray-400">配置通知渠道</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Telegram */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center">
              <MessageSquare className="h-6 w-6 text-blue-500 mr-3" />
              <div>
                <h3 className="text-lg font-semibold text-white">Telegram</h3>
                <p className="text-sm text-gray-400">通过 Telegram 机器人接收通知</p>
              </div>
            </div>
            <button
              onClick={() => setTelegramEnabled(!telegramEnabled)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full ${
                telegramEnabled ? 'bg-blue-600' : 'bg-gray-600'
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition ${
                  telegramEnabled ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-400">机器人 Token</label>
              <input
                type="password"
                placeholder="请输入 Telegram 机器人 Token"
                disabled={!telegramEnabled}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white disabled:opacity-50"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400">聊天 ID</label>
              <input
                type="text"
                placeholder="请输入聊天 ID"
                disabled={!telegramEnabled}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white disabled:opacity-50"
              />
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">通知类型：</p>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    defaultChecked
                    disabled={!telegramEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">交易执行</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    defaultChecked
                    disabled={!telegramEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">风险告警</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    defaultChecked
                    disabled={!telegramEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">套利机会</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    disabled={!telegramEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">每日汇总</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* 邮件 */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center">
              <Mail className="h-6 w-6 text-blue-500 mr-3" />
              <div>
                <h3 className="text-lg font-semibold text-white">邮件</h3>
                <p className="text-sm text-gray-400">通过邮件接收通知</p>
              </div>
            </div>
            <button
              onClick={() => setEmailEnabled(!emailEnabled)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full ${
                emailEnabled ? 'bg-blue-600' : 'bg-gray-600'
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition ${
                  emailEnabled ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-400">SMTP 服务器</label>
              <input
                type="text"
                placeholder="smtp.gmail.com"
                disabled={!emailEnabled}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white disabled:opacity-50"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400">发件人邮箱</label>
              <input
                type="email"
                placeholder="your@email.com"
                disabled={!emailEnabled}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white disabled:opacity-50"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400">收件人邮箱</label>
              <input
                type="email"
                placeholder="recipient@email.com"
                disabled={!emailEnabled}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white disabled:opacity-50"
              />
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">通知类型：</p>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    defaultChecked
                    disabled={!emailEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">交易执行</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    defaultChecked
                    disabled={!emailEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">风险告警</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    disabled={!emailEnabled}
                    className="mr-2 rounded"
                  />
                  <span className="text-white text-sm">每日汇总</span>
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <button className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium">
          <Save className="h-5 w-5 mr-2" />
          保存通知设置
        </button>
      </div>
    </div>
  );
}
