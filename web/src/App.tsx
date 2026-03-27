import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import Login from './pages/Login';
import Settings from './pages/Settings';
import Home from './pages/Home'; // New import
import Sidebar from './components/Sidebar/Sidebar';
import { useAuth } from './contexts/AuthContext';
import { AlertCircleIcon } from "lucide-react";


export default function App() {
  const { user, loading } = useAuth();

  // Auth Guard: Show Login if no user is found
  if (!user && !loading) {
    return <Login onLoginSuccess={() => window.location.reload()} />;
  }

  // Loading Screen: While checking auth
  if (loading) {
    return <div className="min-h-screen flex items-center justify-center bg-slate-50 text-slate-400">Loading Medical Portal...</div>;
  }

  return (
    <BrowserRouter>
      <div className="flex min-h-screen bg-slate-50 text-slate-900">
        
        <Sidebar />

        <main className="flex-1 p-8 max-w-4xl mx-auto">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/settings/*" element={<Settings />} />
            
            {/* 404 CATCH-ALL */}
            <Route path="*" element={
              <div className="flex flex-col items-center justify-center py-20 text-slate-400 text-center">
                <AlertCircleIcon size={48} className="mb-4 text-slate-200" />
                <h2 className="text-2xl font-bold text-slate-800">404 - Page Not Found</h2>
                <p className="mt-1">We couldn't find the chart you were looking for.</p>
                <Link to="/" className="text-blue-600 underline mt-4 font-medium">Return Home</Link>
              </div>
            } />
          </Routes>
        </main>

      </div>
    </BrowserRouter>
  );
}
