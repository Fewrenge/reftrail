import { ChartNoAxesCombinedIcon, } from 'lucide-react';

export default function AnalyticsPage() {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* HEADER SECTION */}
      <div className="flex items-center justify-between pb-5 border-b border-slate-100">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5 text-blue-700">
            <ChartNoAxesCombinedIcon size={22} strokeWidth={2.5} />
            <h1 className="text-2xl font-bold text-slate-900">Analytics Dashboard</h1>
          </div>
          <p className="text-sm text-slate-500">Monitor your referral performance, traffic, and conversion metrics.</p>
        </div>
      </div>



      
     
    </div>
  );
}