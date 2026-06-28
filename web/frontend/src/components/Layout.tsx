import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import {
  LayoutDashboard,
  Settings,
  History,
  Bell,
  LogOut,
  Activity,
  ChevronLeft,
  ChevronRight,
  ArrowLeftRight,
  Bot,
  BarChart3,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { useState, useRef, useEffect } from 'react';

const navigation = [
  {
    name: '总览',
    items: [
      { name: '仪表盘', href: '/', icon: LayoutDashboard },
    ],
  },
  {
    name: '套利 Agent',
    items: [
      { name: '套利监控', href: '/arbitrage', icon: ArrowLeftRight },
      { name: '套利回测', href: '/arbitrage/backtest', icon: History },
    ],
  },
  {
    name: '交易 Agent',
    items: [
      { name: '交易监控', href: '/trading', icon: Bot },
      { name: '交易回测', href: '/trading/backtest', icon: BarChart3 },
    ],
  },
  {
    name: '系统',
    items: [
      { name: '设置', href: '/settings', icon: Settings },
      { name: '通知', href: '/notifications', icon: Bell },
    ],
  },
];

export default function Layout() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭菜单
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="flex h-screen bg-background">
      {/* 侧边栏 */}
      <div className={`flex flex-col border-r bg-card transition-all duration-300 ${collapsed ? 'w-16' : 'w-64'}`}>
        {/* Logo */}
        <div className="flex h-16 items-center justify-between px-4 border-b">
          {!collapsed && (
            <div className="flex items-center gap-2">
              <Activity className="h-6 w-6 text-primary" />
              <span className="text-lg font-bold">TradingAgent</span>
            </div>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setCollapsed(!collapsed)}
            className="h-8 w-8"
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>

        {/* 导航 */}
        <nav className="flex-1 overflow-y-auto p-2">
          {navigation.map((group) => (
            <div key={group.name} className="mb-4">
              {!collapsed && (
                <p className="px-3 mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  {group.name}
                </p>
              )}
              <div className="space-y-1">
                {group.items.map((item) => {
                  const isActive = location.pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      to={item.href}
                      className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                        isActive
                          ? 'bg-primary text-primary-foreground'
                          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                      }`}
                      title={collapsed ? item.name : undefined}
                    >
                      <item.icon className="h-5 w-5 flex-shrink-0" />
                      {!collapsed && <span>{item.name}</span>}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        {/* 用户信息 */}
        <div className="border-t p-4 relative" ref={menuRef}>
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
          >
            <Avatar className="h-8 w-8">
              <AvatarFallback className="bg-primary text-primary-foreground">
                {user?.username?.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            {!collapsed && (
              <div className="flex flex-col items-start">
                <span className="text-sm font-medium">{user?.username}</span>
                <span className="text-xs text-muted-foreground">{user?.email}</span>
              </div>
            )}
          </button>

          {/* 下拉菜单 */}
          {menuOpen && (
            <div className="absolute bottom-full left-4 right-4 mb-2 bg-card border rounded-lg shadow-lg overflow-hidden">
              <div className="px-4 py-3 border-b">
                <p className="text-sm font-medium">{user?.username}</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
              <div className="py-1">
                <Link
                  to="/settings"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-3 px-4 py-2 text-sm hover:bg-accent transition-colors"
                >
                  <Settings className="h-4 w-4" />
                  <span>设置</span>
                </Link>
                <button
                  onClick={() => {
                    setMenuOpen(false);
                    logout();
                  }}
                  className="w-full flex items-center gap-3 px-4 py-2 text-sm hover:bg-accent transition-colors text-red-500"
                >
                  <LogOut className="h-4 w-4" />
                  <span>退出登录</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* 主内容 */}
      <div className="flex-1 overflow-auto">
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
