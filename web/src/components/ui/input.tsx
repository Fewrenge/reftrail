import React from "react";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  icon?: React.ReactNode;
}

export const Input = ({ label, icon, className = "", ...props }: InputProps) => {
  return (
    <div className="w-full space-y-1">
      {label && <label className="block text-sm font-medium text-slate-700">{label}</label>}
      <div className="relative">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
            {icon}
          </div>
        )}
        <input
          className={`w-full ${icon ? 'pl-10' : 'px-4'} py-2 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500/20 outline-none transition-all ${className}`}
          {...props}
        />
      </div>
    </div>
  );
};
