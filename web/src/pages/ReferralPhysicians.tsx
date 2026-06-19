import React, { useState, useEffect } from 'react';
import { ReferralPhysicianCard } from '@/components/ReferralPhysician/ReferralPhysicianCard';
import type { ReferralPhysician } from '@/components/ReferralPhysician/ReferralPhysicianCard';

import AddReferralPhysicianDialog from '@/components/Dialog/AddReferralPhysicianDialog';
import { SearchIcon, PlusIcon } from 'lucide-react'
import { Button } from "@/components/ui";
import { useAuth } from '../contexts/AuthContext';

export const ReferralPhysicians: React.FC = () => {
  const { user: authUser } = useAuth();
  const isAdmin = authUser?.role === "REFTRAIL_ADMIN";

  const [physiciansList, setPhysiciansList] = useState<ReferralPhysician[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 10;

  const [isModalOpen, setIsModalOpen] = useState(false);

  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");

  // Debounce logic sequence exactly from Referrals.tsx
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 600);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const refreshData = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();

      // Append Pagination Indices matching FindReferralPhysician struct
      const offset = (currentPage - 1) * pageSize;
      params.append("limit", pageSize.toString());
      params.append("offset", offset.toString());

      // Passes a single search token to backend FindReferralPhysician generalTerm binding
      const cleanSearch = debouncedSearch.trim();
      if (cleanSearch !== "") {
        params.append("generalTerm", cleanSearch);
      }

      const response = await fetch(`/api/v1/physicians?${params.toString()}`, {
        method: 'GET',
        credentials: 'same-origin'
      });

      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Failed to fetch entries");

      // Set data and total count variables from backend payload format
      setPhysiciansList(result.referralPhysicians || []);
      setTotalCount(result.totalCount || 0);
    } catch (err: any) {
      console.error("Physician list refresh error:", err);
    } finally {
      setLoading(false);
    }
  };

  // Trigger page re-render sequences whenever pages or filters change
  useEffect(() => {
    refreshData();
  }, [currentPage, debouncedSearch]);

  // Reset pagination index if search terms shift
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearch]);

  const handleSelect = (physician: ReferralPhysician) => {
    console.log('Selected physician:', physician.id);
  };

  // Basic calculation for rendering the pagination page count
  const totalPages = Math.ceil(totalCount / pageSize) || 1;

  return (

    <div className="p-6 bg-slate-50 min-h-screen">
      <div className="max-w-7xl mx-auto">

        {/* HEADER SECTION */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl font-bold text-slate-900">Physicians Directory</h1>
            <p className="text-sm text-slate-500 mt-1">
              Showing {physiciansList.length} of {totalCount} total physicians
            </p>
          </div>

          {/* ADMIN ACTION BUTTON */}
          {isAdmin && (
            <div className="shrink-0">
              <Button variant="outline" onClick={() => setIsModalOpen(true)}>
                <PlusIcon size={18} className="mr-2" />
                Add Physician
              </Button>
            </div>
          )}

        </div>


        {/* SEARCH BAR CONTAINER */}
        <div className="flex gap-4 mb-6">
          <div className="relative flex-1 group">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
              <SearchIcon size={20} strokeWidth={2} />
            </div>
            <input
              type="text"
              placeholder="Search for physician by name or CPSO number..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full h-11 bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all text-slate-900 placeholder:text-slate-400 group-hover:border-slate-300 focus:border-blue-500"
            />
          </div>
        </div>

        {/* PHYSICIANS LIST */}
        <div className="space-y-4">
          {loading ? (
            <p className="text-center text-slate-400 animate-pulse py-10">Syncing database...</p>
          ) : physiciansList.length > 0 ? (
            physiciansList.map((physician) => (
              <ReferralPhysicianCard
                key={physician.id}
                physician={physician}
                onClick={handleSelect}
              />
            ))
          ) : (
            <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">
              {searchQuery ? "No physicians match your search filters." : "No physicians found."}
            </div>
          )}
        </div>


        {/* PAGINATION PANEL FOOTER */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between border-t border-slate-200 mt-8 pt-4">
            <button
              disabled={currentPage === 1 || loading}
              onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
              className="px-4 py-2 border border-slate-200 rounded-xl text-sm font-medium bg-white text-slate-700 hover:bg-slate-50 disabled:opacity-50 disabled:pointer-events-none transition-colors"
            >
              Previous
            </button>
            <span className="text-sm font-medium text-slate-600">
              Page {currentPage} of {totalPages}
            </span>
            {/*TODO: fix the pagination bug*/}
            <button
              disabled={currentPage === totalPages || loading}
              onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
              className="px-4 py-2 border border-slate-200 rounded-xl text-sm font-medium bg-white text-slate-700 hover:bg-slate-50 disabled:opacity-50 disabled:pointer-events-none transition-colors"
            >
              Next 
            </button>
          </div>
        )}

        <AddReferralPhysicianDialog
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onSuccess={refreshData} // Automatically re-polls your Go backend listing on success!
        />

      </div>
    </div>


  );
};
