import { useState, useEffect, useRef } from 'react';
import { SearchIcon, PlusIcon, UploadIcon, ChevronLeftIcon, ChevronRightIcon, FilterIcon } from "lucide-react";
import ReferralEntryCard from '../components/ReferralEntry/ReferralEntryCard';
import AddReferralEntryDialog from '../components/ReferralEntry/AddReferralEntryDialog';
import type { ReferralEntry } from '../components/ReferralEntry/ReferralEntryCard';
import { Button } from "@/components/ui";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuCheckboxItem, DropdownMenuItem } from "@/components/ui/dropdown";

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

  // 1. Pipeline Queue Statuses (Defaults to READY_TO_BOOK if local storage is blank)
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_statuses");
      return saved ? JSON.parse(saved) : ["READY_TO_BOOK"];
    } catch (err) {
      console.error("Failed to parse statuses from localStorage:", err);
      return ["READY_TO_BOOK"];
    }
  });

  // 2. Urgent / ASAP Priorities (Defaults to an empty array so all display initially)
  const [selectedUrgencies, setSelectedUrgencies] = useState<string[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_urgencies");
      return saved ? JSON.parse(saved) : [];
    } catch (err) {
      console.error("Failed to parse urgencies from localStorage:", err);
      return [];
    }
  });

  // 3. Clinical Identification Tags (Defaults to an empty array)
  const [selectedTags, setSelectedTags] = useState<string[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_tags");
      return saved ? JSON.parse(saved) : [];
    } catch (err) {
      console.error("Failed to parse tags from localStorage:", err);
      return [];
    }
  });

  // Sync status queue alterations to browser local disk
  useEffect(() => {
    localStorage.setItem("reftrail_selected_statuses", JSON.stringify(selectedStatuses));
  }, [selectedStatuses]);

  // Sync priority checkbox toggles to browser local disk
  useEffect(() => {
    localStorage.setItem("reftrail_selected_urgencies", JSON.stringify(selectedUrgencies));
  }, [selectedUrgencies]);

  // Sync tag pill selections to browser local disk
  useEffect(() => {
    localStorage.setItem("reftrail_selected_tags", JSON.stringify(selectedTags));
  }, [selectedTags]);



  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const refreshData = async () => {
    setLoading(true);
    try {
      // 1. Instantiate a clean URL constructor
      const params = new URLSearchParams();

      // 2. Append Pagination Indices
      const limit = 10; // Change this to match your backend page sizes
      const offset = (currentPage - 1) * limit;
      params.append("limit", limit.toString());
      params.append("offset", offset.toString());

      // 1. Clear status array parameters mapped repetitively for Echo bindings
      if (selectedStatuses.length > 0) {
        selectedStatuses.forEach(status => params.append("statuses", status));
      }

      // Clear urgency priority parameters mapped repetitively
      if (selectedUrgencies.length > 0) {
        selectedUrgencies.forEach(urgency => params.append("urgencies", urgency));
      }

      // Clear tag lookup string parameters mapped repetitively 
      if (selectedTags.length > 0) {
        selectedTags.forEach(tag => params.append("tag_names", tag));
      }


      // Passes a single token to the backend
      const cleanSearch = debouncedSearch.trim();
      if (cleanSearch !== "") {
        params.append("patient_name_search", cleanSearch);
      }


      // Fire off the synchronized fetch call
      const response = await fetch(`/api/v1/referrals?${params.toString()}`, {
        method: 'GET',
        credentials: 'same-origin'
      });

      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Failed to fetch entries");

      // Map your response parameters to state variables
      setPatients(result.referralEntries || []);
      setTotalCount(result.totalCount || 0);

    } catch (err: any) {
      console.error("Pipeline refresh error:", err);
    } finally {
      setLoading(false);
    }
  };


  // Trigger page re-render sequences whenever pages or filters change
  useEffect(() => {
    refreshData();
  }, [currentPage, debouncedSearch, selectedStatuses.join(","), selectedUrgencies.join(","), selectedTags.join(",")]);

  // FIXED: Reset pagination index if search terms or filters shift
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearch, selectedStatuses.join(","), selectedUrgencies.join(","), selectedTags.join(",")]);

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

      {/* WORKFLOW CONTROLS SECTION (SEARCH BAR & STATUS DROPDOWN MENU) */}
      <div className="flex flex-col md:flex-row gap-4 mb-6">
        {/* SEARCH BAR CONTAINER */}
        <div className="relative flex-1 group">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
            <SearchIcon size={20} strokeWidth={2} />
          </div>
          <input
            type="text"
            placeholder="Search by name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-11 bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
          />
        </div>

        {/* FIXED PIPELINE QUEUE SELECTOR (RADIX PRIMITIVE DROPDOWN WINDOW) */}
        <div className="w-full md:w-64 h-11">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="w-full h-full bg-white border border-slate-200 rounded-xl px-4 flex items-center justify-between text-sm text-slate-700 hover:bg-slate-50 hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all cursor-pointer shadow-sm font-medium"
              >
                <div className="flex items-center gap-2 truncate">
                  <FilterIcon size={16} className="text-slate-400 shrink-0" />
                  <span className="truncate">
                    {selectedStatuses.length === 0
                      ? "All Queues Active"
                      : selectedStatuses.length === 1
                        ? AVAILABLE_STATUSES.find((s) => s.id === selectedStatuses[0])?.label || "1 Queue Selected"
                        : `Queues (${selectedStatuses.length} Active)`}
                  </span>
                </div>
                <ChevronRightIcon size={16} className="text-slate-400 rotate-90 shrink-0 transition-transform duration-200" />
              </button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" className="w-64 max-h-95 overflow-y-auto bg-white shadow-xl rounded-xl border border-slate-200/80 p-1.5 z-50">
              <DropdownMenuLabel className="text-xs font-semibold text-slate-400 uppercase tracking-wider px-2.5 py-2">
                Select Workflow Queues
              </DropdownMenuLabel>
              <DropdownMenuSeparator className="bg-slate-100 my-1" />

              {AVAILABLE_STATUSES.map((status) => {
                const isChecked = selectedStatuses.includes(status.id);
                return (
                  <DropdownMenuCheckboxItem
                    key={status.id}
                    checked={isChecked}
                    onCheckedChange={() => {
                      setSelectedStatuses((prev) =>
                        prev.includes(status.id)
                          ? prev.filter((id) => id !== status.id)
                          : [...prev, status.id]
                      );
                    }}
                    className="rounded-lg px-2.5 py-2 text-sm text-slate-600 focus:bg-slate-50 focus:text-slate-900 data-[state=checked]:text-blue-700 data-[state=checked]:bg-blue-50/50 data-[state=checked]:font-medium transition-all duration-150 cursor-pointer mb-0.5"
                  >
                    {status.label}
                  </DropdownMenuCheckboxItem>
                );
              })}

              {selectedStatuses.length > 0 && (
                <>
                  <DropdownMenuSeparator className="bg-slate-100 my-1" />
                  <DropdownMenuItem
                    onClick={(e) => {
                      e.preventDefault(); // Prevents close lifecycle events during rapid workflow toggling
                      setSelectedStatuses([]);
                    }}
                    className="text-xs text-center justify-center font-semibold text-blue-600 focus:bg-blue-50/80 focus:text-blue-700 rounded-lg py-1.5 mt-1 cursor-pointer transition-colors"
                  >
                    RESET TO ALL QUEUES
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* METRIC MODULAR FILTERS PANEL (URGENCIES & TAGS) */}
      <div className="bg-slate-50 border border-slate-100 rounded-xl p-4 mb-6 space-y-4">
        {/* Header with Global Reset Trigger */}
        <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider">
          <div className="flex items-center gap-2">
            <FilterIcon size={14} />
            <span>Active View Parameters</span>
          </div>
          {(selectedStatuses.length > 0 || selectedUrgencies.length > 0 || selectedTags.length > 0) && (
            <button
              onClick={() => {
                setSelectedStatuses(["READY_TO_BOOK"]); // Resets back to team default baseline queue
                setSelectedUrgencies([]);
                setSelectedTags([]);
              }}
              className="text-blue-600 hover:text-blue-700 font-normal normal-case transition-colors cursor-pointer"
            >
              Clear Filters
            </button>
          )}
        </div>

        {/* 1. PRIORITIES ROW (HIGH VISIBILITY SEMANTIC COLOR-CODED PILLS) */}
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-1 border-t border-slate-200/60">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-20">Urgency:</span>
          <div className="flex flex-wrap gap-2">
            {[
              { id: "ASAP", label: "ASAP", activeStyle: "bg-red-50 text-red-700 border-red-200 ring-2 ring-red-500/10 font-semibold" },
              { id: "URGENT", label: "Urgent", activeStyle: "bg-amber-50 text-amber-700 border-amber-200 ring-2 ring-amber-500/10 font-semibold" },
              { id: "ELECTIVE", label: "Elective", activeStyle: "bg-emerald-50 text-emerald-700 border-emerald-200 ring-2 ring-emerald-500/10 font-semibold" }
            ].map((urgency) => {
              const isSelected = selectedUrgencies.includes(urgency.id);
              return (
                <button
                  key={urgency.id}
                  type="button"
                  onClick={() => {
                    setSelectedUrgencies(prev =>
                      prev.includes(urgency.id) ? prev.filter(id => id !== urgency.id) : [...prev, urgency.id]
                    );
                  }}
                  className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                    ? urgency.activeStyle
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                    }`}
                >
                  {urgency.label}
                </button>
              );
            })}
          </div>
        </div>

        {/* 2. TAGS ROW (NEUTRAL INTERACTION PILLS) */}
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-1 border-t border-slate-200/60">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-20">Tags:</span>
          <div className="flex flex-wrap gap-2">
            {["SAN", "DAN", "ORANGE", "BANANA"].map((tagName) => {
              const isSelected = selectedTags.includes(tagName);
              return (
                <button
                  key={tagName}
                  type="button"
                  onClick={() => {
                    setSelectedTags(prev =>
                      prev.includes(tagName) ? prev.filter(t => t !== tagName) : [...prev, tagName]
                    );
                  }}
                  className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                    ? 'bg-purple-50 text-purple-700 border-purple-200 ring-2 ring-purple-500/10 font-semibold shadow-sm'
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                    }`}
                >
                  {tagName}
                </button>
              );
            })}
          </div>
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
      {totalCount > 0 && (
        <div className="flex justify-between items-center mt-8 pt-4 border-t border-slate-100">
          {/* This summary metric panel will now remain visible for all filter sets! */}
          <p className="text-sm text-slate-500">
            Showing Page <span className="font-semibold text-slate-700">{currentPage}</span> of{" "}
            <span className="font-semibold text-slate-700">{Math.max(totalPages, 1)}</span> ({totalCount} total records)
          </p>

          {/* Only show the page toggle buttons if there is more than 1 page to browse */}
          {totalPages > 1 && (
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
          )}
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
