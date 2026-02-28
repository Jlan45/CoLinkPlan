import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';

import Home from '@/pages/Home';
import Login from '@/pages/Login';
import Register from '@/pages/Register';
import Dashboard from '@/pages/Dashboard';
import Nodes from '@/pages/Nodes';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { user, loading } = useAuth();
  if (loading) return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="w-5 h-5 border-2 border-white/20 border-t-white/70 rounded-full animate-spin" />
    </div>
  );
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
};

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/nodes" element={<Nodes />} />
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
    </Routes>
  );
}

export default App;
