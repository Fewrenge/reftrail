import { useState, useEffect, useRef } from 'react';
import { SearchIcon, PlusIcon, UploadIcon, ChevronLeftIcon, ChevronRightIcon, FilterIcon } from "lucide-react";
import ReferralEntryCard from '../components/ReferralEntry/ReferralEntryCard';
import AddReferralEntryDialog from '../components/ReferralEntry/AddReferralEntryDialog';
import type { ReferralEntry } from '../components/ReferralEntry/ReferralEntryCard';
import { Button } from "@/components/ui";

const AVAILABLE_STATUSES = [
  { id: 'READY_TO_BOOK', label: 'Ready to Book' },
  { id: '1ST_CALL_COMPLETE', label: '1st Call Complete' },
  { id: '2ND_CALL_COMPLETE', label: '2nd Call Complete' },
  { id: '3RD_CALL_COMPLETE', label: '3rd Call Complete' },
  { id: 'BOOKED', label: 'Booked' },
  { id: 'UNABLE_TO_CONTACT', label: 'Unable to Contact' },
  { id: 'PATIENT_TO_CALL_BACK', label: 'Patient to Call Back' },
  { id: 'DECLINED', label: 'Declined' },
  { id: 'SUSPENDED', label: 'Suspended' },
  { id: 'CLOSED', label: 'Closed' }
];

export default function Referrals() {
  const [patients, setPatients] = useState<ReferralEntry[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 25; // Aligned with Go backend defaults


  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);

  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const handler = setTimeout(() => setDebouncedSearch(searchQuery), 1000);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  const handleStatusToggle = (statusId: string) => {
    setSelectedStatuses(prev => 
      prev.includes(statusId) 
        ? prev.filter(id => id !== statusId) 
        : [...prev, statusId]
    );
  };

  const refreshData = () => {
    setLoading(true);

    const params = new URLSearchParams({
      limit: String(pageSize),
      offset: String((currentPage - 1) * pageSize),
    });

    if (debouncedSearch.trim() !== "") {
      // Direct pass to your Echo c.Bind fuzzy filter variables
      params.append("patientFirstName", debouncedSearch);
      params.append("patientLastName", debouncedSearch);
    }

     selectedStatuses.forEach(status => {
      params.append("statuses", status);
    });

    fetch(`/api/v1/referrals?${params.toString()}`, { credentials: 'same-origin' })
      .then(res => res.json())
      .then(data => {
        // FIXED: Safely dissect the envelop struct contract keys
        if (data && Array.isArray(data.referralEntries)) {
          setPatients(data.referralEntries);
          setTotalCount(data.totalCount || 0);
        } else {
          setPatients([]);
          setTotalCount(0);
        }
      })
      .catch(err => console.error("Data fetching error:", err))
      .finally(() => setLoading(false));
  };

  // Trigger page re-render sequences whenever pages or filters change
  useEffect(() => {
    refreshData();
  }, [currentPage, debouncedSearch, selectedStatuses]);

  // Reset pagination index if search terms shift
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearch, selectedStatuses]);

  // Handler to pipe the binary file stream to your new backend handler
  const handleBatchImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setUploading(true);
    const formData = new FormData();
    formData.append('file', file); // 'file' matches c.FormFile("file") exactly

    try {
      const response = await fetch('/api/v1/referrals/batch', {
        method: 'POST',
        body: formData, // Browser automatically sets 'multipart/form-data' boundary header
        credentials: 'same-origin'
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to import batch file.');
      }

      alert('Batch file import successful!');
      refreshData(); // Re-fetch list to show the freshly uploaded entries
    } catch (err: any) {
      alert(`Import Error: ${err.message}`);
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = ""; // Reset input
    }
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  return (

    <>
      <header className="flex justify-between items-center mb-8">
        <h2 className="text-2xl font-bold tracking-tight text-slate-800">Referrals</h2>
        <div className="flex gap-3">


          <Button variant="outline" onClick={() => setIsModalOpen(true)}>
            <PlusIcon size={18} className="mr-2" />
            Add Referral
          </Button>

          {/* Hidden binary file input wrapper */}
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleBatchImport}
            accept=".tsv,.csv"
            className="hidden"
          />
          <Button
            // variant="outline" 
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
          >
            <UploadIcon size={18} className="mr-2" />
            {uploading ? "Importing..." : "Batch Import"}
          </Button>
        </div>
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

      {/* DYNAMIC MULTI-SELECT STATUS CHECKBOX BAR */}
      <div className="bg-slate-50 border border-slate-100 rounded-xl p-4 mb-6">
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3">
          <FilterIcon size={14} />
          <span>Filter Statuses</span>
          {selectedStatuses.length > 0 && (
            <button 
              onClick={() => setSelectedStatuses([])} 
              className="ml-auto text-blue-600 hover:text-blue-700 lowercase font-normal"
            >
              Clear filters
            </button>
          )}
        </div>
        <div className="flex flex-wrap gap-x-5 gap-y-2">
          {AVAILABLE_STATUSES.map((status) => (
            <label 
              key={status.id} 
              className="flex items-center gap-2 text-sm text-slate-600 cursor-pointer select-none hover:text-slate-800 transition-colors"
            >
              <input
                type="checkbox"
                checked={selectedStatuses.includes(status.id)}
                onChange={() => handleStatusToggle(status.id)}
                className="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500/20"
              />
              <span>{status.label}</span>
            </label>
          ))}
        </div>
      </div>

      {/* PATIENT LIST */}
      <div className="space-y-4">
        {loading ? (
          <p className="text-center text-slate-400 animate-pulse py-10">Syncing database...</p>
        ) : patients.length > 0 ? (
          patients.map((p) => (
            <ReferralEntryCard key={p.id} entry={p} onRefresh={refreshData} isClickable={true} />
          ))
        ) : (
          <div className="py-20 text-center border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 italic">
            {searchQuery || selectedStatuses.length > 0 ? "No patients match your search filters." : "No entries found."}
          </div>
        )}
      </div>

      {/* PAGINATION CONTROLS FOOTER PANEL BAR */}
      {totalPages > 1 && (
        <div className="flex justify-between items-center mt-8 pt-4 border-t border-slate-100">
          <p className="text-sm text-slate-500">
            Showing Page <span className="font-semibold text-slate-700">{currentPage}</span> of{" "}
            <span className="font-semibold text-slate-700">{totalPages}</span> ({totalCount} total records)
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
              disabled={currentPage === 1}
            >
              <ChevronLeftIcon size={16} className="mr-1" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
              disabled={currentPage === totalPages}
            >
              Next
              <ChevronRightIcon size={16} className="ml-1" />
            </Button>
          </div>
        </div>
      )}

      <AddReferralEntryDialog
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={refreshData}
      />
    </>
  );
}
