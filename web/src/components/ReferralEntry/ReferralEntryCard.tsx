import { useState } from 'react';
import {Trash2Icon} from "lucide-react";


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
  const [showMenu, setShowMenu] = useState(false);

  const urgencyStyles = {
    ASAP: "bg-red-50 text-red-700 border-red-100",
    Urgent: "bg-amber-50 text-amber-700 border-amber-100",
    Elective: "bg-emerald-50 text-emerald-700 border-emerald-100",
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

        {/* This container holds badges AND the dots side-by-side */}
        <div className="flex items-center gap-3">
          <div className="flex gap-1.5">
            <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase border ${urgencyStyles[entry.urgency as keyof typeof urgencyStyles]}`}>
              {entry.urgency}
            </span>
            <span className="px-2 py-0.5 rounded text-[10px] font-black uppercase border bg-blue-50 text-blue-700 border-blue-100">
              {entry.status}
            </span>
          </div>

          {/* THE DOTS MENU */}
          <div className="relative">
            <button 
              onClick={() => setShowMenu(!showMenu)}
              className="p-1 rounded-lg hover:bg-slate-100 text-slate-400 transition-colors cursor-pointer"
            >
              <span className="text-xl leading-none font-bold">⋮</span>
            </button>

            {showMenu && (
              <>
                <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)}></div>
                <div className="absolute right-0 mt-2 w-40 bg-white border border-slate-200 rounded-xl shadow-xl z-20 py-1 overflow-hidden">
                  <button 
                    onClick={() => { handleDelete(); setShowMenu(false); }}
                    className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 font-bold transition-colors flex items-center gap-2"
                  >
                    <Trash2Icon size={15} strokeWidth={2}/>
                  <span>Delete Entry</span>
                  </button>
                </div>
              </>
            )}
          </div>
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
    </div>
  );
}
