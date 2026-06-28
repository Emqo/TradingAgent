import { useState } from 'react';
import { History, Play, BarChart3 } from 'lucide-react';

export default function Backtest() {
  const [running, setRunning] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">回测</h1>
        <p className="text-gray-400">使用历史数据测试您的策略</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 配置 */}
        <div className="lg:col-span-1 bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">配置</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-400">策略</label>
              <select className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white">
                <option>三角套利</option>
                <option>期现套利</option>
                <option>自定义策略</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-gray-400">交易对</label>
              <select className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white">
                <option>BTCUSDT</option>
                <option>ETHUSDT</option>
                <option>SOLUSDT</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-gray-400">开始日期</label>
              <input
                type="date"
                defaultValue="2024-01-01"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400">结束日期</label>
              <input
                type="date"
                defaultValue="2024-12-31"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400">初始资金 (USDT)</label>
              <input
                type="number"
                defaultValue={10000}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
              />
            </div>
            <button
              onClick={() => setRunning(!running)}
              className={`w-full flex items-center justify-center px-4 py-2 rounded-lg font-medium ${
                running
                  ? 'bg-red-600 hover:bg-red-700 text-white'
                  : 'bg-blue-600 hover:bg-blue-700 text-white'
              }`}
            >
              {running ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2" />
                  运行中...
                </>
              ) : (
                <>
                  <Play className="h-5 w-5 mr-2" />
                  开始回测
                </>
              )}
            </button>
          </div>
        </div>

        {/* 结果 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 统计 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <p className="text-sm text-gray-400">总收益</p>
              <p className="text-xl font-bold text-green-500">+24.5%</p>
            </div>
            <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <p className="text-sm text-gray-400">夏普比率</p>
              <p className="text-xl font-bold text-white">1.85</p>
            </div>
            <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <p className="text-sm text-gray-400">最大回撤</p>
              <p className="text-xl font-bold text-red-500">-8.2%</p>
            </div>
            <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <p className="text-sm text-gray-400">胜率</p>
              <p className="text-xl font-bold text-white">62%</p>
            </div>
          </div>

          {/* 图表占位 */}
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <div className="flex items-center mb-4">
              <BarChart3 className="h-6 w-6 text-blue-500 mr-3" />
              <h3 className="text-lg font-semibold text-white">收益曲线</h3>
            </div>
            <div className="h-64 flex items-center justify-center border border-gray-600 rounded-lg">
              <p className="text-gray-400">图表将在此处显示</p>
            </div>
          </div>

          {/* 交易记录 */}
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <div className="flex items-center mb-4">
              <History className="h-6 w-6 text-blue-500 mr-3" />
              <h3 className="text-lg font-semibold text-white">交易记录</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-gray-400 text-sm">
                    <th className="pb-3">时间</th>
                    <th className="pb-3">交易对</th>
                    <th className="pb-3">方向</th>
                    <th className="pb-3">价格</th>
                    <th className="pb-3">数量</th>
                    <th className="pb-3">盈亏</th>
                  </tr>
                </thead>
                <tbody className="text-white text-sm">
                  <tr className="border-t border-gray-700">
                    <td className="py-3">2024-01-15 14:30</td>
                    <td>BTCUSDT</td>
                    <td className="text-green-500">买入</td>
                    <td>$42,150</td>
                    <td>0.1</td>
                    <td className="text-green-500">+$125</td>
                  </tr>
                  <tr className="border-t border-gray-700">
                    <td className="py-3">2024-01-15 14:35</td>
                    <td>ETHUSDT</td>
                    <td className="text-red-500">卖出</td>
                    <td>$2,580</td>
                    <td>1.5</td>
                    <td className="text-green-500">+$85</td>
                  </tr>
                  <tr className="border-t border-gray-700">
                    <td className="py-3">2024-01-15 15:00</td>
                    <td>BTCUSDT</td>
                    <td className="text-red-500">卖出</td>
                    <td>$42,300</td>
                    <td>0.1</td>
                    <td className="text-green-500">+$150</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
