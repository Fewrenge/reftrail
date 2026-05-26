import { useState, useMemo } from 'react';
import { ROLES, ALL_STATUSES, STATUS_RULES } from "@/helpers/constants";
import { useAuth } from "@/contexts/AuthContext";
import { Trash2Icon, MessageSquareIcon, XIcon, LogsIcon, PlusIcon, XCircleIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuLabel
} from "@/components/ui/dropdown";
import { useNavigate } from 'react-router-dom';

interface Props {
  entry: ReferralEntry;
  onRefresh: () => void;
  isClickable?: boolean; // Optional prop to control if the card is clickable
}

export interface Complaint {
  referralId: string;
  bodyPart: string;
  side: string;
  details: string;
}

// This is the "Blueprint" for what data one entry needs
export interface ReferralEntry {
  id: string; // UUID
  patientLastName: string;
  patientFirstName: string;
  patientDob: string;
  urgency: 'ASAP' | 'Urgent' | 'Elective';
  status: string;
  referringPhysician: string;
  complaints: Complaint[];
  triageNote: string;
  tags: string[];
}

export default function ReferralEntryCard({ entry, onRefresh, isClickable }: Props) {

  // --- States ---
  const [selectedStatus, setSelectedStatus] = useState<string | null>(null);
  const [isLogMode, setIsLogMode] = useState(false);
  const [note, setNote] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [allGlobalTags, setAllGlobalTags] = useState<any[]>([]);
  const { user } = useAuth();
  const isAdmin = user?.role === ROLES.SYSTEM_ADMIN;

  const navigate = useNavigate();

  const allowedStatuses = useMemo(() => {
    if (isAdmin) {
      // Admins can move to any status except the one they are currently in
      return ALL_STATUSES.filter(s => s !== entry.status);
    }
    // Booking team follows the matrix
    return STATUS_RULES[entry.status] || [];
  }, [isAdmin, entry.status]);

  const urgencyStyles = {
    ASAP: "bg-red-50 text-red-700 border-red-100",
    Urgent: "bg-amber-50 text-amber-700 border-amber-100",
    Elective: "bg-emerald-50 text-emerald-700 border-emerald-100",
  };

  // Logic to send the update to your Go backend
  const handleSavePad = async () => {
    setIsLoading(true);

    try {
      // 1. Determine target endpoint based on current active mode
      const url = isLogMode
        ? `/api/v1/referrals/${entry.id}/logs`
        : `/api/v1/referrals/${entry.id}/status`;

      const method = isLogMode ? 'POST' : 'PATCH';

      // 2. Format body data dynamically
      const bodyData = isLogMode
        ? { note: note || "Manual log entry" }
        : { newStatus: selectedStatus, note: note || "Status updated" };

      const res = await fetch(url, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(bodyData)
      });

      if (res.ok) {
        // Clear out state variables on success
        setSelectedStatus(null);
        setIsLogMode(false);
        setNote("");
        onRefresh();
      } else {
        const err = await res.text();
        alert(`Save failed: ${err}`);
      }
    } catch (err) {
      console.error("Network submission error:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchGlobalTags = async () => {
    try {
      const res = await fetch('/api/v1/tags'); // Replace with your exact route
      if (res.ok) {
        const data = await res.json();
        setAllGlobalTags(data);
      }
    } catch (err) {
      console.error("Failed to load tag definitions:", err);
    }
  };


  const handleAssignTag = async (tagId: number) => {
    setIsLoading(true);
    try {
      const res = await fetch(`/api/v1/referrals/${entry.id}/tags/${tagId}`, { method: 'POST' });
      if (res.ok) onRefresh();
      else alert("Failed to append tag");
    } catch (err) {
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  // Native Single-Tag Delete Link Request
  const handleRemoveTag = async (tagName: string) => {
    setIsLoading(true);
    try {
      const res = await fetch(`/api/v1/referrals/${entry.id}/tags/${tagName}`, { method: 'DELETE' });
      if (res.ok) onRefresh();
      else alert("Failed to remove tag");
    } catch (err) {
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {

    if (!window.confirm(`Permanently delete ${entry.patientLastName}, ${entry.patientFirstName}?`)) return;

    try {
      const res = await fetch(`/api/v1/referrals/${entry.id}`, {
        method: 'DELETE'
      });

      if (res.ok) {
        onRefresh(); // This MUST be called to update the UI
      } else {
        const errorData = await res.text();
        alert(`Delete failed: ${errorData}`);
      }
    } catch (err) {
      console.error("Delete error:", err);
    }
  };


  return (
    <div
      className={`relative group`}
    >
      <div className={`bg-white border border-slate-200 rounded-2xl p-5 shadow-sm relative group transition-all ${isClickable ? 'hover:border-blue-300' : ''}`}>

        {/* 1. TOP SECTION: Name on left, Badges & Menu on right */}
        <div className="flex justify-between items-start mb-6">
          <div>
            {/* Wrap the patient name text in a clickable button element */}
            {isClickable ? (
              <button
                onClick={() => navigate(`/referrals/${entry.id}`)}
                className="text-left font-bold text-xl text-slate-900 hover:text-blue-600 cursor-pointer transition-colors focus:outline-none"
              >
                {entry.patientLastName}{", "}{entry.patientFirstName}
              </button>
            ) : (
              <h3 className="font-bold text-xl text-slate-900">
                {entry.patientLastName}{", "}{entry.patientFirstName}
              </h3>
            )}

            <p className="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-1">
              DOB: {entry.patientDob || 'N/A'}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex gap-1.5 items-center">
              {/* URGENCY BADGE */}
              <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase border ${urgencyStyles[entry.urgency as keyof typeof urgencyStyles]}`}>
                {entry.urgency}
              </span>

              {/* STATUS DROPDOWN */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="outline"
                    size="sm"
                    // Disable the button if the user has no allowed transitions
                    disabled={allowedStatuses.length === 0}
                    className="h-6 px-2 text-[10px] font-black uppercase bg-blue-50 text-blue-700 border-blue-100 rounded-md disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {entry.status.replace(/_/g, ' ')}
                  </Button>
                </DropdownMenuTrigger>

                <DropdownMenuContent align="end" className="w-48">
                  <DropdownMenuLabel>
                    {isAdmin ? "Admin: Change Status" : "Transition To..."}
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {allowedStatuses.map((s) => (
                    <DropdownMenuItem
                      key={s}
                      onSelect={() => setSelectedStatus(s)}
                      className="text-[11px] font-medium"
                    >
                      {s.replace(/_/g, ' ')}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

            </div>



            {/* THE DOTS MENU */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="p-1 rounded-lg hover:bg-slate-100 text-slate-400 transition-colors cursor-pointer outline-none">
                  <span className="text-xl leading-none font-bold">⋮</span>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48 p-1 rounded-xl shadow-xl border-slate-200">
                <DropdownMenuItem
                  onSelect={() => {
                    setIsLogMode(true); // Open the overlay window in log mode
                  }}
                  className="hover:bg-slate-50 font-bold flex items-center gap-3 px-4 py-3 cursor-pointer rounded-lg transition-colors"
                >
                  <LogsIcon size={16} strokeWidth={2.5} />
                  <span>Add a Note</span>
                </DropdownMenuItem>


                <DropdownMenuItem
                  onSelect={() => { handleDelete(); }}
                  className="text-red-600 hover:bg-red-50 font-bold flex items-center gap-3 px-4 py-3 cursor-pointer rounded-lg transition-colors"
                >
                  <Trash2Icon size={16} strokeWidth={2.5} />
                  <span>Delete Entry</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>




          </div>
        </div>

        {/* 2. MIDDLE SECTION: Details Grid */}
        <div className="grid grid-cols-2 gap-8 mb-4">
          <div>
            <p className="text-[9px] text-slate-400 font-bold uppercase tracking-tight mb-1">Physician</p>
            <p className="text-sm font-medium text-slate-700">{entry.referringPhysician || 'Unassigned'}</p>
          </div>
          <div>
            <p className="text-[9px] text-slate-400 font-bold uppercase tracking-tight mb-1">Complaints</p>
            <div className="space-y-1">
              {entry.complaints && entry.complaints.length > 0 ? (
                entry.complaints.map((c, index) => {
                  // Generate composite key string (including index to guard against duplicates)
                  const compositeKey = `${c.side}-${c.bodyPart}-${index}`;

                  return (
                    <p key={compositeKey} className="text-sm font-medium text-slate-700 capitalize">
                      {`${c.side?.toLowerCase() || ''} ${c.bodyPart?.toLowerCase() || ''}`}
                      {c.details && <span className="text-xs text-slate-400 block font-normal">{c.details}</span>}
                    </p>
                  );
                })
              ) : (
                <p className="text-sm font-medium text-slate-400 italic">None reported</p>
              )}
            </div>
          </div>
        </div>



        {/* TAG SECTION: Tags Row */}
        <div className="flex flex-wrap items-center gap-1.5 mb-5 mt-2">
          {entry.tags && entry.tags.map((tag: any, index: number) => {

            // 1. Safely extract the string text regardless of how the backend shapes it
            const tagNameStr = typeof tag === 'object' && tag !== null
              ? (tag.name || tag.tagName || "")
              : String(tag);

            // 2. Ignore empty items safely
            if (!tagNameStr) return null;

            return (
              <span
                // 3. FIX: Combine the string text with the array index to guarantee 100% uniqueness
                key={`tag-${tagNameStr}-${index}`}
                className="inline-flex items-center gap-1 pl-2.5 pr-1.5 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600 border border-slate-200/60 uppercase tracking-wider shadow-2xs hover:bg-slate-200 transition-colors"
              >
                <span>{tagNameStr}</span>

                {/* Little Cross to remove tags (Only renders for Admins) */}
                {isAdmin && (
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(tagNameStr)}
                    className="text-slate-400 hover:text-red-500 transition-colors cursor-pointer rounded-full outline-none"
                    disabled={isLoading}
                  >
                    <XCircleIcon size={12} fill="currentColor" className="text-white" />
                  </button>
                )}
              </span>
            );
          })}

          {/* Administrative Dropdown Selection Tool to assign fresh tags */}
          {isAdmin && (
            <DropdownMenu onOpenChange={(open) => open && fetchGlobalTags()}>
              <DropdownMenuTrigger asChild>
                <button
                  className="inline-flex items-center justify-center w-5 h-5 rounded-full border border-dashed border-slate-300 bg-slate-50 text-slate-500 hover:bg-blue-50 hover:text-blue-600 hover:border-blue-300 transition-colors cursor-pointer outline-none"
                  disabled={isLoading}
                >
                  <PlusIcon size={12} strokeWidth={3} />
                </button>
              </DropdownMenuTrigger>

              <DropdownMenuContent align="start" className="w-48 p-1 rounded-xl shadow-xl border-slate-200">
                <DropdownMenuLabel className="text-[10px] font-bold uppercase text-slate-400 px-2 py-1.5">
                  Available Tags
                </DropdownMenuLabel>
                <DropdownMenuSeparator />

                {allGlobalTags.length === 0 ? (
                  <div className="text-xs text-slate-400 p-2 text-center">No tags left</div>
                ) : (
                  allGlobalTags
                    // Safely filter out tags already applied to this specific card
                    .filter(gt => !entry.tags?.some((t: any) => {
                      const appliedName = typeof t === 'object' && t !== null ? (t.name || t.tagName) : String(t);
                      return appliedName === gt.name;
                    }))
                    .map((globalTag, dropdownIndex) => (
                      <DropdownMenuItem
                        // FIX: Guarantee uniqueness for the dropdown list elements too
                        key={`available-tag-${globalTag.name}-${dropdownIndex}`}
                        onSelect={() => handleAssignTag(globalTag.name)}
                        className="text-xs font-semibold flex items-center px-3 py-2 cursor-pointer rounded-lg transition-colors"
                      >
                        {globalTag.name}
                      </DropdownMenuItem>
                    ))
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>



        {/* BOTTOM SECTION: Triage Note */}
        <div className="bg-slate-50 border-l-2 border-blue-400 p-3 rounded-r-lg">
          <p className="text-sm text-slate-600 italic leading-relaxed">
            {entry.triageNote ? `"${entry.triageNote}"` : "No triage notes recorded."}
          </p>
        </div>





        {/* --- QUICK NOTE OVERLAY --- */}
        {/* This only appears AFTER they select a status from the dropdown */}
        {(selectedStatus || isLogMode) && (
          <>
            {/* 1. Backdrop: Clicking anywhere else closes the window */}
            <div
              className="fixed inset-0 z-40 bg-slate-900/5 backdrop-blur-[1px]"
              onClick={() => { setSelectedStatus(null); setNote(""); }}
            />

            {/* 2. The Square Pad: Positioned relative to the card, but z-50 to stay on top */}
            <div className="absolute top-2 right-2 w-80 h-80 bg-white border border-blue-200 shadow-2xl z-50 rounded-2xl flex flex-col p-5 animate-in zoom-in-95 duration-150">

              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-blue-50 text-blue-600 rounded-lg">
                    <MessageSquareIcon size={16} />
                  </div>
                  <div>
                    <p className="text-[10px] font-black uppercase text-slate-400 leading-none">
                      {isLogMode ? "Adding Audit Log To" : "Updating Status To"}
                    </p>
                    <p className="text-xs font-bold text-slate-700">
                      {isLogMode ? `${entry.patientLastName}, ${entry.patientFirstName}` : selectedStatus?.replace(/_/g, ' ')}
                    </p>
                  </div>
                </div>
                <button onClick={() => { setSelectedStatus(null); setIsLogMode(false); }} className="text-slate-300 hover:text-slate-600 transition-colors">
                  <XIcon size={18} />
                </button>
              </div>

              <textarea
                className="flex-1 w-full bg-slate-50 border border-slate-100 rounded-xl p-4 text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none mb-4 font-medium text-slate-700"
                placeholder="Write a note about this status update."
                value={note}
                onChange={(e) => setNote(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleSavePad();
                  }
                }}
              />

              {/* TODO: Quick note function */}

              <div className="flex gap-2">
                <Button
                  variant="primary"
                  className="flex-1 shadow-lg shadow-blue-200"
                  onClick={handleSavePad}
                  disabled={isLoading}
                >
                  {isLoading ? "Saving..." : "Confirm & Log"}
                </Button>
              </div>
            </div>
          </>
        )}



      </div>
    </div>
  );
}