import React, { useState, useEffect } from 'react';
import Login from './pages/Login';
import WLEntryCard from './components/WLEntryCard';
import type { WLEntry } from './components/WLEntryCard';
import AddEntryModal from './components/AddEntryModal';
import { BrowserRouter, Routes, Route, Link, NavLink } from 'react-router-dom';
import { HospitalIcon, ScrollTextIcon, ChartNoAxesCombinedIcon, SearchIcon } from "lucide-react";
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
            <HospitalIcon size={30} strokeWidth={2.5} />
            <h1 className="font-bold text-lg tracking-tight text-slate-800">Medical Portal</h1>
          </div>

          <nav className="space-y-1 flex-1">
            {/*WAITLIST*/}
            <NavLink
              to="/"
              end
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
                }`
              }
            >
              <ScrollTextIcon size={20} strokeWidth={2.5} />
              <span>Waitlist</span>
            </NavLink>


            {/*ANALYTICS*/}
            <NavLink
              to="/analytics" // Give it a real (even if empty) path
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
                }`
              }
            >
              <ChartNoAxesCombinedIcon size={20} strokeWidth={2.5}/>
              <span>Analytics</span>
            </NavLink>
          </nav>

          <div className="pt-6 border-t border-slate-100 mt-auto">
            <NavLink to="/settings" className="flex items-center gap-3 p-2 rounded-xl hover:bg-slate-100 transition-all group no-underline">
              <div className="w-9 h-9 bg-linear-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center text-white font-bold shadow-md border-2 border-white">
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
            </NavLink>
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
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                    <SearchIcon size={20} strokeWidth={2.5}/>
                  </div>
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

            <Route path="/settings/*" element={<Settings />} />

            <Route path="*" element={
              <div className="flex flex-col items-center justify-center py-20">
                <h2 className="text-2xl font-bold">404 - Not Found</h2>
                <Link to="/" className="text-blue-600 underline mt-2">Go back home</Link>
              </div>
            } />
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