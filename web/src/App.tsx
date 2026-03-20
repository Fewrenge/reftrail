import React, { useState, useEffect } from 'react';
import Login from './pages/Login';
import WLEntryCard from './components/WLEntryCard';
import type { WLEntry } from './components/WLEntryCard';
import AddEntryModal from './components/AddEntryModal';

export default function App() {

  // All hooks must be at the top
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [patients, setPatients] = useState<WLEntry[]>([]);// This starts as an empty list []
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Effect to fetch patients but with the Badge (JWT)
  useEffect(() => {
    if (!token) {
      setPatients([]); // Clear list on logout
      return;
    }

    setLoading(true);

    fetch('/api/v1/waitlist', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.ok ? res.json() : Promise.reject())
      .then(data => setPatients(Array.isArray(data) ? data : []))
      .catch(() => {
        // If token is expired/invalid, force logout
        handleLogout();
      })
      .finally(() => setLoading(false));
  }, [token]);

  // --- HELPER FUNCTIONS ---
  const handleLogout = () => {
    // 1. Clear the "Cookie Jar"
    localStorage.removeItem('token');
    // window.location.href = '/'; 
  };
  // 3. CONDITIONAL RENDERING (Must be AFTER all Hooks)
  if (!token) {
    return <Login onLoginSuccess={(newToken) => setToken(newToken)} />;
  }

  const refreshData = () => {
    setLoading(true);
    fetch('/api/v1/waitlist', { headers: { 'Authorization': `Bearer ${token}` } })
      .then(res => res.json())
      .then(data => setPatients(Array.isArray(data) ? data : []))
      .finally(() => setLoading(false));
  };

  return (

    // 'flex' puts items side-by-side. 'bg-slate-50' gives that clean Memos grey background.
    <div className="flex min-h-screen bg-slate-50 text-slate-900">

      {/* LEFT SIDEBAR: Fixed width (64), white background, border on the right */}
      <aside className="w-64 bg-white border-r border-slate-200 sticky top-0 h-screen p-6 hidden md:block">
        <div className="flex items-center gap-3 mb-10">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">M</div>
          <h1 className="font-bold text-lg">Medical Portal</h1>
        </div>

        <nav className="space-y-2">
          <div className="bg-blue-50 text-blue-700 px-3 py-2 rounded-lg font-medium cursor-pointer">
            📋 Waitlist
          </div>
          <div className="text-slate-500 px-3 py-2 hover:bg-slate-50 rounded-lg transition-colors cursor-pointer">
            📊 Analytics
          </div>
          <div className="text-slate-500 px-3 py-2 hover:bg-slate-50 rounded-lg transition-colors cursor-pointer">
            ⚙️ Administration
          </div>
        </nav>
      </aside>

      {/* MAIN CONTENT AREA: This will hold your feed */}
      <main className="flex-1 p-8 max-w-4xl mx-auto">
        <header className="flex justify-between items-center mb-8">
          <h2 className="text-2xl font-bold tracking-tight">Active Waitlist</h2>
          <button
            onClick={() => setIsModalOpen(true)} // Open the modal
            className="bg-blue-600 text-white px-4 py-2 rounded-xl font-medium shadow-sm hover:bg-blue-700 transition-all">
            + Add Waitlist Entry
          </button>
        </header>

        {/* SEARCH BAR SECTION */}
        <div className="relative mb-6 group">
          {/* The Search Icon */}
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
            🔍
          </div>

          <input
            type="text"
            placeholder="Search by name, complaint, or physician..."
            className="w-full bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all placeholder:text-slate-400"
          />
        </div>


        {/* Where the WLEntry cards are */}
        <div className="space-y-4">
          {loading ? (
            <p className="text-center text-slate-400 animate-pulse py-10">Syncing...</p>
          ) : patients.length > 0 ? (
            patients.map((p) => (
              <WLEntryCard
                key={p.id}
                entry={p}
                token={token || ''}      // Provide the badge
                onRefresh={refreshData}  // Provide the reloader
              />
            ))
          ) : (
            <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">
              No entries found.
            </div>
          )}
        </div>

        <button
          onClick={() => { localStorage.removeItem('token'); setToken(null); }}
          className="text-xs text-slate-400 mt-10 hover:text-red-500"
        >
          Logout
        </button>
      </main>

      <AddEntryModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={refreshData}
        token={token || ''}
      />

    </div>
  );
}