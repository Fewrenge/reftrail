import React from 'react';
import { Trash2Icon } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown";

// Interfaces matching your backend's ReferralPhysician type
export interface ReferralPhysician {
  id: string;
  cpsoNumber: string | null;
  firstName: string;
  lastName: string;
  emrPhysicianId: string | null;
}

interface ReferralPhysicianCardProps {
  physician: ReferralPhysician;
  onRefresh?: () => void; // Made optional to prevent TypeErrors
  onClick?: (physician: ReferralPhysician) => void;
}

export const ReferralPhysicianCard: React.FC<ReferralPhysicianCardProps> = ({ physician, onClick, onRefresh }) => {
  const isClickable = !!onClick;
  const fullName = `${physician.firstName} ${physician.lastName}`;

  const handleDelete = async () => {
    if (!window.confirm(`Permanently delete ${physician.lastName}, ${physician.firstName}?`)) return;

    try {
      const res = await fetch(`/api/v1/physicians/${physician.id}`, {
        method: 'DELETE'
      });

      if (res.ok) {
        // If parent passed a function, run it. Otherwise, force a clean browser reload.
        if (typeof onRefresh === 'function') {
          onRefresh();
        } else {
          window.location.reload();
        }
      } else {
        const errorData = await res.text();
        alert(`Delete failed: ${errorData}`);
      }
    } catch (err) {
      console.error("Delete error:", err);
    }
  };

  return (
    <div className="relative group">
      <div
        onClick={() => isClickable && onClick(physician)}
        className={`bg-white border border-slate-200 rounded-2xl p-5 shadow-sm relative transition-all ${
          isClickable ? 'hover:border-blue-300 cursor-pointer hover:shadow-md' : ''
        }`}
      >
        {/* Header: Profile Initials and Name */}
        <div className="flex items-center justify-between gap-4">
          <div className="truncate">
            <h3 className="text-slate-900 font-semibold text-lg truncate">
              {fullName}
            </h3>
          </div>

          {/* THE DOTS MENU */}
          <div onClick={(e) => e.stopPropagation()}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="p-1 rounded-lg hover:bg-slate-100 text-slate-400 transition-colors cursor-pointer outline-none">
                  <span className="text-xl leading-none font-bold">⋮</span>
                </button>
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end" className="w-48 p-1 rounded-xl shadow-xl border-slate-200">
                <DropdownMenuItem
                  onSelect={(e) => {
                    e.preventDefault(); 
                    handleDelete();
                  }}
                  className="text-red-600 hover:bg-red-50 font-bold flex items-center gap-3 px-4 py-3 cursor-pointer rounded-lg transition-colors"
                >
                  <Trash2Icon size={16} strokeWidth={2.5} />
                  <span>Delete Entry</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Technical/Medical Meta Details */}
        <div className="mt-4 pt-4 border-t border-slate-100 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="block text-xs font-medium text-slate-400 uppercase tracking-wider">
              CPSO Number
            </span>
            <span className="font-mono text-slate-700 font-medium mt-0.5 block">
              {physician.cpsoNumber ?? 'N/A'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
