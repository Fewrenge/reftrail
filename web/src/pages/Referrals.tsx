import { useState, useEffect, useRef } from 'react';
import { SearchIcon, PlusIcon, UploadIcon, ChevronLeftIcon, ChevronRightIcon, ChevronDownIcon, FilterIcon } from "lucide-react";
import ReferralEntryCard from '@/components/ReferralEntry/ReferralEntryCard';
import AddReferralEntryDialog from '@/components/Dialog/AddReferralEntryDialog';
import type { ReferralEntry } from '@/components/ReferralEntry/ReferralEntryCard';
import { Button } from "@/components/ui";
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuLabel,
  DropdownMenuSeparator, DropdownMenuCheckboxItem, DropdownMenuItem
} from "@/components/ui/dropdown";
import { useAuth } from '@/contexts/AuthContext';
import { UserRole } from '@/types/users';
import { ALL_STATUSES, ALL_CONSULT_TYPES, ALL_URGENCIES, ALL_SOURCES } from '@/types/referrals';
import type { ReferralStatus, ReferralUrgency, ReferralConsultType, ReferralSource } from '@/types/referrals';

export default function Referrals() {

  const URGENCY_STYLES: Record<ReferralUrgency, string> = {
    ASAP: "bg-red-50 text-red-700 border-red-200 ring-2 ring-red-500/10 font-semibold",
    URGENT: "bg-amber-50 text-amber-700 border-amber-200 ring-2 ring-amber-500/10 font-semibold",
    ELECTIVE: "bg-emerald-50 text-emerald-700 border-emerald-200 ring-2 ring-emerald-500/10 font-semibold",
  };

  const CONSULT_STYLES: Record<ReferralConsultType, string> = {
    "APP+LE": "bg-blue-50 text-blue-700 border-blue-200 ring-2 ring-blue-500/10 font-semibold",
    "APP+UE": "bg-cyan-50 text-cyan-700 border-cyan-200 ring-2 ring-cyan-500/10 font-semibold",
    "APP+SX": "bg-indigo-50 text-indigo-700 border-indigo-200 ring-2 ring-indigo-500/10 font-semibold",
    "SX": "bg-violet-50 text-violet-700 border-violet-200 ring-2 ring-violet-500/10 font-semibold",
    "OTHER": "bg-slate-100 text-slate-700 border-slate-300 ring-2 ring-slate-500/10 font-semibold",
  };

  const SOURCE_STYLES: Record<ReferralSource, string> = {
    REGULAR: "bg-teal-50 text-teal-700 border-teal-200 ring-2 ring-teal-500/10 font-semibold",
    FRACTURE_CLINIC: "bg-orange-50 text-orange-700 border-orange-200 ring-2 ring-orange-500/10 font-semibold",
    OTHER: "bg-slate-100 text-slate-700 border-slate-300 ring-2 ring-slate-500/10 font-semibold",
  };



  const { user: authUser } = useAuth();
  const isAdmin = authUser?.role === UserRole.REFTRAIL_ADMIN;

  const [isCollapsibleSectionOpen, setIsCollapsibleSectionOpen] = useState(true);

  const [patients, setPatients] = useState<ReferralEntry[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 10; // Aligned with Go backend defaults

  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [availableTags, setAvailableTags] = useState<string[]>([]);

  // Date Filter States (Defaults to empty string for no filter active)
  const [referralDateFrom, setReferralDateFrom] = useState<string>(() => {
    return localStorage.getItem("reftrail_date_from") || "";
  });

  const [referralDateTo, setReferralDateTo] = useState<string>(() => {
    return localStorage.getItem("reftrail_date_to") || "";
  });

  // Sync date alterations to browser local disk
  useEffect(() => {
    localStorage.setItem("reftrail_date_from", referralDateFrom);
  }, [referralDateFrom]);

  useEffect(() => {
    localStorage.setItem("reftrail_date_to", referralDateTo);
  }, [referralDateTo]);


  useEffect(() => {
    const fetchTags = async () => {
      try {
        const response = await fetch('/api/v1/tags', {
          method: 'GET',
          credentials: 'same-origin'
        });

        if (!response.ok) throw new Error("Failed to pull tag definitions");
        const data = await response.json();

        if (Array.isArray(data)) {
          const tagNames = data.map((t: any) => typeof t === 'string' ? t : t.name);
          setAvailableTags(tagNames);
        }
      } catch (err) {
        console.error("Failed fetching dynamic tag layout structures:", err);
      }
    };

    fetchTags();
  }, []); // Empty brackets ensure this runs exactly once on initial load



  // Pipeline Queue Statuses (Defaults to READY_TO_BOOK if local storage is blank)
  const [selectedStatuses, setSelectedStatuses] = useState<ReferralStatus[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_statuses");
      return saved ? JSON.parse(saved) : ["READY_TO_BOOK"];
    } catch (err) {
      console.error("Failed to parse statuses from localStorage:", err);
      return ["READY_TO_BOOK"];
    }
  });

  // Urgent / ASAP Priorities (Defaults to an empty array so all display initially)
  const [selectedUrgencies, setSelectedUrgencies] = useState<ReferralUrgency[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_urgencies");
      return saved ? JSON.parse(saved) : [];
    } catch (err) {
      console.error("Failed to parse urgencies from localStorage:", err);
      return [];
    }
  });

  // Clinical Identification Tags (Defaults to an empty array)
  const [selectedTags, setSelectedTags] = useState<string[]>(() => {
    try {
      const saved = localStorage.getItem("reftrail_selected_tags");
      return saved ? JSON.parse(saved) : [];
    } catch (err) {
      console.error("Failed to parse tags from localStorage:", err);
      return [];
    }
  });

  const [selectedConsultTypes, setSelectedConsultTypes] = useState<ReferralConsultType[]>([]);
  const [selectedSources, setSelectedSources] = useState<ReferralSource[]>([]);




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

  useEffect(() => {
    localStorage.setItem("reftrail_selected_sources", JSON.stringify(selectedSources));
  }, [selectedSources]);



  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 600);

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
        selectedTags.forEach(tag => params.append("tagNames", tag));
      }

      if (selectedConsultTypes.length > 0) {
        selectedConsultTypes.forEach(ct => params.append("consultTypes", ct));
      }

      if (selectedSources.length > 0) {
        selectedSources.forEach(source => params.append("sources", source));
      }

      // Append Date Bounds if they have active text values
      if (referralDateFrom !== "") {
        params.append("referralDateFrom", referralDateFrom);
      }

      if (referralDateTo !== "") {
        params.append("referralDateTo", referralDateTo);
      }



      // Passes a single token to the backend
      const cleanSearch = debouncedSearch.trim();
      if (cleanSearch !== "") {
        params.append("generalTerm", cleanSearch);
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
  }, [currentPage, debouncedSearch, selectedStatuses.join(","), selectedUrgencies.join(","), selectedTags.join(","), selectedConsultTypes.join(","), selectedSources.join(","),
    referralDateFrom, referralDateTo]);

  // FIXED: Reset pagination index if search terms or filters shift
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearch, selectedStatuses.join(","), selectedUrgencies.join(","), selectedTags.join(","), selectedConsultTypes.join(","), selectedSources.join(","),
    referralDateFrom, referralDateTo]);

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

          {isAdmin && (
            <>
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
            </>
          )}



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
            placeholder="Search for patient by name or health card number..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-11 bg-white border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2
             focus:ring-blue-500/20 transition-all"
          />
        </div>

        {/* FIXED PIPELINE QUEUE SELECTOR (RADIX PRIMITIVE DROPDOWN WINDOW) */}
        <div className="w-full md:w-64 h-11">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="w-full h-full bg-white border border-slate-200 rounded-xl px-4 flex items-center justify-between text-sm text-slate-700
                 hover:bg-slate-50 hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all cursor-pointer shadow-sm font-medium"
              >
                <div className="flex items-center gap-2 truncate">
                  <FilterIcon size={16} className="text-slate-400 shrink-0" />
                  <span className="truncate">
                    {selectedStatuses.length === 0
                      ? "No Statuses Selected" // Alert user that queries will yield no results
                      : selectedStatuses.length === ALL_STATUSES.length
                        ? "All Statuses Selected"
                        : selectedStatuses.length === 1
                          ? ALL_STATUSES.find((s) => s.id === selectedStatuses[0])?.label || "1 Status Selected"
                          : `Statuses (${selectedStatuses.length} Active)`}
                  </span>

                </div>
                <ChevronRightIcon size={16} className="text-slate-400 rotate-90 shrink-0 transition-transform duration-200" />
              </button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" className="w-64 max-h-95 overflow-y-auto bg-white shadow-xl rounded-xl border border-slate-200/80 p-1.5 z-50">
              <DropdownMenuLabel className="text-xs font-semibold text-slate-400 uppercase tracking-wider px-2.5 py-2">
                Select Statuses
              </DropdownMenuLabel>

              {/* MASTER SELECT ALL / DESELECT ALL SHORTCUT SWITCH */}
              <DropdownMenuItem
                onSelect={(e) => e.preventDefault()} // Prevents dropdown from closing when clicking the shortcut
                onClick={() => {
                  setSelectedStatuses((prev) => {
                    // Everything is considered active if the length matches or if it's completely empty
                    const isEverythingChecked = prev.length === ALL_STATUSES.length || prev.length === 0;

                    if (isEverythingChecked) {
                      // Deselect All: Make the filter completely empty
                      return [];
                    } else {
                      // Select All: Populate the array explicitly with every single ID
                      return ALL_STATUSES.map(s => s.id);
                    }
                  });
                }}

                className="text-xs font-medium text-slate-500 hover:text-slate-800 focus:bg-slate-50 rounded-lg py-1.5 px-2.5 mb-1 cursor-pointer 
                transition-colors flex justify-between items-center"
              >
                <span>Toggle Selection:</span>
                <span className="text-blue-600 font-semibold uppercase tracking-wide text-[10px]">
                  {selectedStatuses.length === 0 || selectedStatuses.length === ALL_STATUSES.length ? "Deselect All" : "Select All"}
                </span>
              </DropdownMenuItem>

              <DropdownMenuSeparator className="bg-slate-100 my-1" />

              {ALL_STATUSES.map((status) => {
                const isChecked = selectedStatuses.includes(status.id);

                return (
                  <DropdownMenuCheckboxItem
                    key={status.id}
                    checked={isChecked}
                    onSelect={(e) => e.preventDefault()}
                    onCheckedChange={() => {
                      setSelectedStatuses((prev) => {
                        if (prev.includes(status.id)) {
                          return prev.filter((id) => id !== status.id);
                        }
                        return [...prev, status.id];
                      });
                    }}
                    className="rounded-lg pr-2.5 py-2 text-sm text-slate-600 focus:bg-slate-50 focus:text-slate-900 data-[state=checked]:text-blue-700
                     data-[state=checked]:bg-blue-50/50 data-[state=checked]:font-semibold transition-all duration-150 cursor-pointer mb-0.5"
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
                    className="text-xs text-center justify-center font-semibold text-blue-600 focus:bg-blue-50/80 focus:text-blue-700 rounded-lg
                     py-1.5 mt-1 cursor-pointer transition-colors"
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
      <div className="w-full bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden my-4">

        {/* The Core Collapse Header Row */}
        <button
          type="button"
          onClick={() => setIsCollapsibleSectionOpen(!isCollapsibleSectionOpen)}
          className="w-full flex justify-between items-center px-4 py-3 bg-slate-50 border-b border-slate-200 hover:bg-slate-100/70 transition-colors cursor-pointer"
        >
          <div className="flex items-center gap-2 text-slate-700 font-semibold text-sm">
            <FilterIcon size={14} className="h-4 w-4 text-slate-500" />
            <span>Active View Parameters</span>
          </div>

          <div className="flex items-center gap-4">
            {/* FORCE RENDER: This check covers every variable hook so it always appears if anything is clicked */}
            {(selectedStatuses?.length > 0 ||
              selectedUrgencies?.length > 0 ||
              selectedTags?.length > 0 ||
              selectedConsultTypes?.length > 0 ||
              selectedSources?.length > 0) && (
                <span
                  onClick={(e) => {
                    e.stopPropagation(); // Stops the button click from toggling the section shut
                    setSelectedStatuses?.([]);
                    setSelectedUrgencies?.([]);
                    setSelectedTags?.([]);
                    setSelectedConsultTypes?.([]);
                    setSelectedSources?.([]);
                  }}
                  className="text-xs font-semibold text-blue-600 hover:text-blue-700 transition-colors cursor-pointer px-2 py-1"
                >
                  Clear Filters
                </span>
              )}

            <div className="text-slate-400 hover:text-slate-600 transition-colors">
              {isCollapsibleSectionOpen ? (
                <ChevronDownIcon className="h-4 w-4 stroke-[2.5]" />
              ) : (
                <ChevronRightIcon className="h-4 w-4 stroke-[2.5]" />
              )}
            </div>
          </div>
        </button>


        {isCollapsibleSectionOpen && (

          <div className="bg-slate-50 border border-slate-100 p-4 space-y-4">

            {/* 1. PRIORITIES ROW (HIGH VISIBILITY SEMANTIC COLOR-CODED PILLS) */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-1 border-slate-200/60">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-20">Urgency:</span>
              <div className="flex flex-wrap gap-2">
                {ALL_URGENCIES.map((urgencyId) => {
                  const isSelected = selectedUrgencies.includes(urgencyId);
                  // Make the label pretty (e.g., 'URGENT' -> 'Urgent', 'ASAP' remains 'ASAP')
                  const label = urgencyId === "ASAP" ? "ASAP" : urgencyId.charAt(0) + urgencyId.slice(1).toLowerCase();

                  return (
                    <button
                      key={urgencyId}
                      type="button"
                      onClick={() => {
                        setSelectedUrgencies(prev =>
                          prev.includes(urgencyId) ? prev.filter(id => id !== urgencyId) : [...prev, urgencyId]
                        );
                      }}
                      className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                        ? URGENCY_STYLES[urgencyId]
                        : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                        }`}
                    >
                      {label}
                    </button>
                  );
                })}
              </div>
            </div>


            {/* 2. TAGS ROW (NEUTRAL INTERACTION PILLS) */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-2 border-t border-slate-200/60">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-20">Tags:</span>
              <div className="flex flex-wrap gap-2">
                {availableTags.length > 0 ? (
                  availableTags.map((tagName) => {
                    const isSelected = selectedTags.includes(tagName);
                    return (
                      <button
                        key={tagName}
                        type="button"
                        onClick={() => {
                          setSelectedTags(prev => {
                            // If the clicked tag is already active, clear the array entirely to deselect it
                            if (prev.includes(tagName)) {
                              return [];
                            }
                            // Otherwise, kick out all other tags and keep only this newly selected tag
                            return [tagName];
                          });
                        }}
                        className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                          ? 'bg-purple-50 text-purple-700 border-purple-200 ring-2 ring-purple-500/10 font-semibold shadow-sm'
                          : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                          }`}
                      >
                        {tagName}
                      </button>
                    );
                  })
                ) : (
                  <span className="text-xs italic text-slate-400">No tag definitions configured.</span>
                )}
              </div>
            </div>

            {/* 3. Consult Types Row */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-2 border-t border-slate-200/60">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-24">Consult Type:</span>
              <div className="flex flex-wrap gap-2">
                {ALL_CONSULT_TYPES.map((consultId) => {
                  const isSelected = selectedConsultTypes.includes(consultId);

                  return (
                    <button
                      key={consultId}
                      type="button"
                      onClick={() => {
                        setSelectedConsultTypes(prev => {
                          if (prev.includes(consultId)) {
                            return [];
                          }
                          return [consultId];
                        });
                      }}
                      className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                        ? CONSULT_STYLES[consultId]
                        : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                        }`}
                    >
                      {consultId}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* 4. Referral Source Row */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 pt-2 border-t border-slate-200/60">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 w-24">Source:</span>
              <div className="flex flex-wrap gap-2">
                {ALL_SOURCES.map((sourceId) => {
                  const isSelected = selectedSources.includes(sourceId);

                  // Clean up the text labels for the UI (e.g., "FRACTURE_CLINIC" -> "Fracture Clinic")
                  const label = sourceId
                    .replace(/_/g, " ")
                    .toLowerCase()
                    .replace(/\b\w/g, (char) => char.toUpperCase());

                  return (
                    <button
                      key={sourceId}
                      type="button"
                      onClick={() => {
                        setSelectedSources(prev => {
                          if (prev.includes(sourceId)) {
                            return [];
                          }
                          return [sourceId];
                        });
                      }}
                      className={`px-3 py-1 text-xs font-medium rounded-full border transition-all duration-150 cursor-pointer active:scale-95 ${isSelected
                        ? SOURCE_STYLES[sourceId]
                        : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100 hover:text-slate-800'
                        }`}
                    >
                      {label}
                    </button>
                  );
                })}
              </div>
            </div>




            {/* DATE FILTERING ROW */}
            <div className="flex items-center gap-4 my-4 p-3 pt-2 bg-gray-50 border-t border-gray-200">
              <div className="flex flex-col gap-1">
                <label className="text-xs font-semibold text-gray-600 uppercase tracking-wider">From Date</label>
                <input
                  type="date"
                  max="9999-12-31"
                  value={referralDateFrom}
                  onChange={(e) => setReferralDateFrom(e.target.value)}
                  className="px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>

              <div className="flex flex-col gap-1">
                <label className="text-xs font-semibold text-gray-600 uppercase tracking-wider">To Date</label>
                <input
                  type="date"
                  max="9999-12-31"
                  value={referralDateTo}
                  onChange={(e) => setReferralDateTo(e.target.value)}
                  className="px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>

              {(referralDateFrom || referralDateTo) && (
                <button
                  onClick={() => { setReferralDateFrom(""); setReferralDateTo(""); }}
                  className="mt-5 px-3 py-1 text-xs font-medium text-red-600 hover:bg-red-50 rounded transition-colors border border-red-200"
                >
                  Clear Dates
                </button>
              )}
            </div>


          </div>)}
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
