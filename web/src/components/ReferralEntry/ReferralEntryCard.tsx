import { useState } from 'react';
import { Trash2Icon, MessageSquareIcon, XIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuLabel
} from "@/components/ui/dropdown";


interface Props {
  entry: ReferralEntry;
  onRefresh: () => void;
}

// This is the "Blueprint" for what data one entry needs
export interface ReferralEntry {
  id: number;
  patientName: string;
  patientDob: string;
  urgency: 'ASAP' | 'Urgent' | 'Elective';
  status: string;
  referringPhysician: string;
  complaint: string;
  triageNote: string;
}

export default function ReferralEntryCard({ entry, onRefresh }: Props) {

  // --- States ---
  const [selectedStatus, setSelectedStatus] = useState<string | null>(null);
  const [note, setNote] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const urgencyStyles = {
    ASAP: "bg-red-50 text-red-700 border-red-100",
    Urgent: "bg-amber-50 text-amber-700 border-amber-100",
    Elective: "bg-emerald-50 text-emerald-700 border-emerald-100",
  };

  // Logic to send the update to your Go backend
  const handleStatusUpdate = async () => {
    setIsLoading(true);

    try {
      const res = await fetch(`/api/v1/referrals/${entry.id}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          newStatus: selectedStatus,
          note: note || "Status updated" // Default note if empty
        })
      });

      if (res.ok) {
        setSelectedStatus(null);
        setNote("");
        onRefresh();
      } else {
        const err = await res.text();
        alert(`Update failed: ${err}`);
      }
    } catch (err) {
      console.error(err);
    }finally{
     setIsLoading(false);
    }
  };


  const handleDelete = async () => {

    if (!window.confirm(`Permanently delete ${entry.patientName}?`)) return;

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
    <div className="bg-white border border-slate-200 rounded-2xl p-5 hover:border-blue-300 transition-all shadow-sm relative group">

      {/* 1. TOP SECTION: Name on left, Badges & Menu on right */}
      <div className="flex justify-between items-start mb-6">
        <div>
          <h3 className="font-bold text-xl text-slate-900">{entry.patientName}</h3>
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
                <Button variant="outline" size="sm" className="h-6 px-2 text-[10px] font-black uppercase bg-blue-50 text-blue-700 border-blue-100 rounded-md">
                  {entry.status.replace(/_/g, ' ')}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuLabel>Transition To...</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {['1ST_CALL_COMPLETE', 'BOOKED', 'UNABLE_TO_CONTACT', 'DECLINED'].map((s) => (
                  <DropdownMenuItem key={s} onSelect={() => setSelectedStatus(s)}>
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
          <p className="text-[9px] text-slate-400 font-bold uppercase tracking-tight mb-1">Complaint</p>
          <p className="text-sm font-medium text-slate-700">{entry.complaint}</p>
        </div>
      </div>

      {/* 3. BOTTOM SECTION: Triage Note */}
      <div className="bg-slate-50 border-l-2 border-blue-400 p-3 rounded-r-lg">
        <p className="text-sm text-slate-600 italic leading-relaxed">
          {entry.triageNote ? `"${entry.triageNote}"` : "No triage notes recorded."}
        </p>
      </div>

      {/* --- QUICK NOTE OVERLAY --- */}
      {/* This only appears AFTER they select a status from the dropdown */}
      {selectedStatus && (
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
                  <p className="text-[10px] font-black uppercase text-slate-400 leading-none">Updating Status To</p>
                  <p className="text-xs font-bold text-slate-700">{selectedStatus.replace(/_/g, ' ')}</p>
                </div>
              </div>
              <button onClick={() => setSelectedStatus(null)} className="text-slate-300 hover:text-slate-600 transition-colors">
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
                  handleStatusUpdate();
                }
              }}
            />

            {/* TODO: Quick note function */}

            <div className="flex gap-2">
              <Button
                variant="primary"
                className="flex-1 shadow-lg shadow-blue-200"
                onClick={handleStatusUpdate}
                disabled={isLoading}
              >
                {isLoading ? "Saving..." : "Confirm & Log"}
              </Button>
            </div>
          </div>
        </>
      )}



    </div>
  );
}
