import React, { useState, useEffect } from 'react';
import Login from './pages/Login';
import WLEntryCard from './components/WLEntryCard';
import type { WLEntry } from './components/WLEntryCard';
import AddEntryModal from './components/AddEntryModal';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import Settings from './pages/Settings'

export default function App() {
  // 1. Hooks (Memory)
  const [user, setUser] = useState<{ username: string, role: string } | null>(null);
  const [patients, setPatients] = useState<WLEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // 2. Helper Functions (Actions)
  const handleLogout = async () => {
    await fetch('/api/v1/logout', { method: 'POST', credentials: 'same-origin' });
    setUser(null);
    setPatients([]);
  };

  const refreshData = () => {
    setLoading(true);
    fetch('/api/v1/waitlist', {
      credentials: 'same-origin'
    })
      .then(res => res.json())
      .then(data => setPatients(Array.isArray(data) ? data : []))
      .finally(() => setLoading(false));
  };

  // 3. Side Effects (Data Fetching)
  useEffect(() => {

    setLoading(true);

    // ELEGANT FIX: Fetch both at once
    Promise.all([
      fetch('/api/v1/users/me', {
        credentials: 'same-origin'
      }),
      fetch('/api/v1/waitlist', {
        credentials: 'same-origin'
      })
    ])
      .then(async ([userRes, wlRes]) => {
        // Only logout if BOTH explicitly say 401
        if (userRes.status === 401 && wlRes.status === 401) {
          handleLogout();
          return;
        }

        if (userRes.ok) {
          const userData = await userRes.json();
          setUser(userData);
        }

        if (wlRes.ok) {
          const wlData = await wlRes.json();
          setPatients(Array.isArray(wlData) ? wlData : []);
        }
      })
      .catch((err) => console.error("Connection glitch:", err))
      .finally(() => setLoading(false));
  }, []);

  // 4. Conditional Rendering (Auth Guard)
  if (!user && !loading) {
    return <Login onLoginSuccess={() => window.location.reload()} />;
  }

  // 5. Main UI
  return (
    <BrowserRouter>
      <div className="flex min-h-screen bg-slate-50 text-slate-900">

        <aside className="w-64 bg-white border-r border-slate-200 sticky top-0 h-screen p-6 hidden md:flex flex-col">
          <div className="flex items-center gap-3 mb-10">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold shadow-sm">M</div>
            <h1 className="font-bold text-lg tracking-tight text-slate-800">Medical Portal</h1>
          </div>

          <nav className="space-y-1 flex-1">
            <Link to="/" className="flex items-center gap-3 bg-blue-50 text-blue-700 px-3 py-2.5 rounded-xl font-medium transition-all">
              <span className="text-lg">📋</span> Waitlist
            </Link>
            <div className="flex items-center gap-3 text-slate-500 px-3 py-2.5 hover:bg-slate-50 rounded-xl transition-colors cursor-pointer group">
              <span className="text-lg">📊</span> Analytics
            </div>
          </nav>

          <div className="pt-6 border-t border-slate-100 mt-auto">
            <Link to="/settings" className="flex items-center gap-3 p-2 rounded-xl hover:bg-slate-100 transition-all group no-underline">
              <div className="w-9 h-9 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center text-white font-bold shadow-md border-2 border-white">
                {user ? user.username.charAt(0).toUpperCase() : '?'}
              </div>
              <div className="flex flex-col overflow-hidden">
                <span className="text-sm font-semibold text-slate-700 truncate">
                  {loading ? 'Syncing...' : (user ? user.username : 'Guest')}
                </span>
                <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold">
                  {user ? user.role : 'Authorized'}
                </span>
              </div>
            </Link>
          </div>
        </aside>

        <main className="flex-1 p-8 max-w-4xl mx-auto">
          <Routes>
            <Route path='/' element={
              <>
                <header className="flex justify-between items-center mb-8">
                  <h2 className="text-2xl font-bold tracking-tight">Active Waitlist</h2>
                  <button
                    onClick={() => setIsModalOpen(true)}
                    className="bg-blue-600 text-white px-4 py-2 rounded-xl font-medium shadow-sm hover:bg-blue-700 transition-all">
                    + Add Waitlist Entry
                  </button>
                </header>

                <div className="relative mb-6 group">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">🔍</div>
                  <input
                    type="text"
                    placeholder="Search by name..."
                    className="w-full bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20"
                  />
                </div>

                <div className="space-y-4">
                  {loading ? (
                    <p className="text-center text-slate-400 animate-pulse py-10">Syncing database...</p>
                  ) : patients.length > 0 ? (
                    patients.map((p) => (
                      <WLEntryCard key={p.id} entry={p} onRefresh={refreshData} />
                    ))
                  ) : (
                    <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">No entries found.</div>
                  )}
                </div>
              </>
            } />

              <Route
                path="/settings"
                element={<Settings user={user} onLogout={handleLogout} />}
              />

            </Routes>
        </main>

        <AddEntryModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onSuccess={refreshData}
        />
      </div>
    </BrowserRouter>
  );
}