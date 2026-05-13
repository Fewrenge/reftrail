import { useState, useEffect } from 'react';
import { SearchIcon, Plus } from "lucide-react";
import ReferralEntryCard from '../components/ReferralEntry/ReferralEntryCard';
import AddReferralEntryDialog from '../components/ReferralEntry/AddReferralEntryDialog';
import type { ReferralEntry } from '../components/ReferralEntry/ReferralEntryCard';
import { Button } from "@/components/ui";

export default function Home() {
  const [patients, setPatients] = useState<ReferralEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const refreshData = () => {
    setLoading(true);
    fetch('/api/v1/referrals', { credentials: 'same-origin' })
      .then(res => res.json())
      .then(data => setPatients(Array.isArray(data) ? data : []))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refreshData();
  }, []);

  // Filter patients based on search input
  // TODO: filter through last name and first name
  const filteredPatients = patients.filter(p =>
    p.patientLastName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <>
      <header className="flex justify-between items-center mb-8">
        <h2 className="text-2xl font-bold tracking-tight text-slate-800">Active Referrals</h2>
        <Button onClick={() => setIsModalOpen(true)}>
          <Plus size={18} className="mr-2" />
          Add Referral
        </Button>
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
            <ReferralEntryCard key={p.id} entry={p} onRefresh={refreshData} />
          ))
        ) : (
          <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">
            {searchQuery ? "No patients match your search." : "No entries found."}
          </div>
        )}
      </div>

      <AddReferralEntryDialog
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={refreshData}
      />
    </>
  );
}
