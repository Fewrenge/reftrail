import React from 'react';

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
  onClick?: (physician: ReferralPhysician) => void;
}

export const ReferralPhysicianCard: React.FC<ReferralPhysicianCardProps> = ({ physician, onClick }) => {
  const isClickable = !!onClick;
  const fullName = `${physician.firstName} ${physician.lastName}`;

  return (
    <div className="relative group">
      <div 
        onClick={() => isClickable && onClick(physician)}
        className={`bg-white border border-slate-200 rounded-2xl p-5 shadow-sm relative transition-all ${
          isClickable ? 'hover:border-blue-300 cursor-pointer hover:shadow-md' : ''
        }`}
      >
        {/* Header: Profile Initials and Name */}
        <div className="flex items-center gap-4">
          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-blue-50 text-blue-600 font-semibold text-base shrink-0">
            {physician.firstName[0]}
            {physician.lastName[0]}
          </div>
          <div className="truncate">
            <h3 className="text-slate-900 font-semibold text-lg truncate">
              {fullName}
            </h3>
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
