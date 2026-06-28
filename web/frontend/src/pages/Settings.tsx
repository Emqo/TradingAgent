import { useState } from 'react';
import { Save } from 'lucide-react';

export default function Settings() {
  const [activeTab, setActiveTab] = useState('risk');

  const tabs = [
    { id: 'risk', name: '风险管理' },
    { id: 'arbitrage', name: '套利设置' },
    { id: 'llm', name: 'LLM 配置' },
    { id: 'exchange', name: '交易所' },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">设置</h1>
        <p className="text-gray-400">配置您的交易代理</p>
      </div>

      <div className="bg-gray-800 rounded-xl border border-gray-700">
        {/* 选项卡 */}
        <div className="border-b border-gray-700">
          <nav className="flex space-x-8 px-6">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-500'
                    : 'border-transparent text-gray-400 hover:text-gray-300'
                }`}
              >
                {tab.name}
              </button>
            ))}
          </nav>
        </div>

        {/* 内容 */}
        <div className="p-6">
          {activeTab === 'risk' && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-white">风险管理设置</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400">最大仓位 (USDT)</label>
                  <input
                    type="number"
                    defaultValue={1000}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最大日亏损 (USDT)</label>
                  <input
                    type="number"
                    defaultValue={500}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最大回撤 (%)</label>
                  <input
                    type="number"
                    defaultValue={10}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最大杠杆</label>
                  <input
                    type="number"
                    defaultValue={3}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">亏损后冷却时间（分钟）</label>
                  <input
                    type="number"
                    defaultValue={5}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'arbitrage' && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-white">套利设置</h3>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-white">三角套利</p>
                    <p className="text-sm text-gray-400">启用三角套利检测</p>
                  </div>
                  <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-blue-600">
                    <span className="inline-block h-4 w-4 transform rounded-full bg-white translate-x-6" />
                  </button>
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-white">期现套利</p>
                    <p className="text-sm text-gray-400">启用期现套利（资金费率套利）</p>
                  </div>
                  <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-blue-600">
                    <span className="inline-block h-4 w-4 transform rounded-full bg-white translate-x-6" />
                  </button>
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最小价差 (bps)</label>
                  <input
                    type="number"
                    defaultValue={15}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最大仓位 (USDT)</label>
                  <input
                    type="number"
                    defaultValue={1000}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'llm' && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-white">LLM 配置</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400">提供商</label>
                  <select className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white">
                    <option value="claude">Claude</option>
                    <option value="openai">OpenAI</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-gray-400">模型</label>
                  <input
                    type="text"
                    defaultValue="mimo-v2.5-pro"
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm text-gray-400">Base URL</label>
                  <input
                    type="text"
                    defaultValue="https://token-plan-cn.xiaomimimo.com/anthropic"
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm text-gray-400">API Key</label>
                  <input
                    type="password"
                    placeholder="请输入 API Key"
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">最大 Token 数</label>
                  <input
                    type="number"
                    defaultValue={4096}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">温度</label>
                  <input
                    type="number"
                    defaultValue={0.7}
                    step={0.1}
                    min={0}
                    max={2}
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'exchange' && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-white">交易所设置</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400">交易所</label>
                  <select className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white">
                    <option>Binance</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-gray-400">模式</label>
                  <select className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white">
                    <option>测试网</option>
                    <option>主网</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-gray-400">API Key</label>
                  <input
                    type="password"
                    placeholder="请输入 API Key"
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400">API Secret</label>
                  <input
                    type="password"
                    placeholder="请输入 API Secret"
                    className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                  />
                </div>
              </div>
            </div>
          )}

          <div className="mt-6 pt-6 border-t border-gray-700">
            <button className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium">
              <Save className="h-5 w-5 mr-2" />
              保存设置
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
