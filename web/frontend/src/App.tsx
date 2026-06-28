import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Arbitrage from './pages/Arbitrage';
import ArbitrageBacktest from './pages/ArbitrageBacktest';
import Agent from './pages/Agent';
import AgentBacktest from './pages/AgentBacktest';
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

            {/* 套利 */}
            <Route path="arbitrage" element={<Arbitrage />} />
            <Route path="arbitrage/backtest" element={<ArbitrageBacktest />} />

            {/* Agent 交易 */}
            <Route path="agent" element={<Agent />} />
            <Route path="agent/backtest" element={<AgentBacktest />} />

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
