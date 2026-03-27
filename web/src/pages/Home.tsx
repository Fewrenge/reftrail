import { useState, useEffect } from 'react';
import { SearchIcon, Plus } from "lucide-react";
import WLEntryCard from '../components/WLEntry/WLEntryCard';
import AddEntryModal from '../components/WLEntry/AddEntryModal';
import type { WLEntry } from '../components/WLEntry/WLEntryCard';

export default function Home() {
  const [patients, setPatients] = useState<WLEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const refreshData = () => {
    setLoading(true);
    fetch('/api/v1/waitlist', { credentials: 'same-origin' })
      .then(res => res.json())
      .then(data => setPatients(Array.isArray(data) ? data : []))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refreshData();
  }, []);

  // Filter patients based on search input
  const filteredPatients = patients.filter(p => 
    p.patientName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <>
      <header className="flex justify-between items-center mb-8">
        <h2 className="text-2xl font-bold tracking-tight text-slate-800">Active Waitlist</h2>
        <button
          onClick={() => setIsModalOpen(true)}
          className="bg-blue-600 text-white px-4 py-2 rounded-xl font-medium shadow-sm hover:bg-blue-700 transition-all cursor-pointer flex items-center gap-2"
        >
          <Plus size={18} />
          Add Waitlist Entry
        </button>
      </header>

      {/* SEARCH BAR */}
      <div className="relative mb-6 group">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
          <SearchIcon size={20} strokeWidth={2} />
        </div>
        <input
          type="text"
          placeholder="Search by name..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
        />
      </div>

      {/* PATIENT LIST */}
      <div className="space-y-4">
        {loading ? (
          <p className="text-center text-slate-400 animate-pulse py-10">Syncing database...</p>
        ) : filteredPatients.length > 0 ? (
          filteredPatients.map((p) => (
            <WLEntryCard key={p.id} entry={p} onRefresh={refreshData} />
          ))
        ) : (
          <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">
            {searchQuery ? "No patients match your search." : "No entries found."}
          </div>
        )}
      </div>

      <AddEntryModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={refreshData}
      />
    </>
  );
}
