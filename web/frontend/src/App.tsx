import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Arbitrage from './pages/Arbitrage';
import ArbitrageBacktest from './pages/ArbitrageBacktest';
import ArbitrageAgent from './pages/ArbitrageAgent';
import Trading from './pages/Trading';
import TradingBacktest from './pages/TradingBacktest';
import Settings from './pages/Settings';
import Notifications from './pages/Notifications';
import Layout from './components/Layout';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            {/* 总览 */}
            <Route index element={<Dashboard />} />

            {/* 套利 Agent */}
            <Route path="arbitrage" element={<Arbitrage />} />
            <Route path="arbitrage/agent" element={<ArbitrageAgent />} />
            <Route path="arbitrage/backtest" element={<ArbitrageBacktest />} />

            {/* 交易 Agent */}
            <Route path="trading" element={<Trading />} />
            <Route path="trading/backtest" element={<TradingBacktest />} />

            {/* 系统 */}
            <Route path="settings" element={<Settings />} />
            <Route path="notifications" element={<Notifications />} />
          </Route>
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
