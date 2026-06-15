import { ChartNoAxesCombinedIcon, TrendingUpIcon } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

// TODO: prettify this page

export default function AnalyticsPage() {
    const navigate = useNavigate();

  // Unified configuration matrix for your dashboard collections
  const reportCards = [
      {
      id: 'urgency-distribution',
      title: 'Urgency Distribution',
      description: 'Breakdown percentages of Elective, Urgent, and ASAP referral entries.',
      path: '/analytics/urgency-distribution',
    },
     {
      id: 'volume-over-time',
      title: 'Referral Volume Trends',
      description: 'Track incoming referral velocity trajectories over monthly timelines using chronological line charts.',
      path: '/analytics/referral-trend',
      icon: <TrendingUpIcon size={20} className="text-indigo-600" />,
      badge: "Line Chart"
    },
  ];


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

      {/* ALBUM VIEW GRID */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {reportCards.map((card) => {

          return (
            <div
              key={card.id}
              onClick={() => ! navigate(card.path)}
              className={`group border bg-white rounded-2xl p-5 flex flex-col justify-between h-56 transition-all duration-200 ${
                  'border-slate-200 hover:border-blue-500 hover:shadow-md cursor-pointer'
              }`}
            >
              <div className="flex items-start gap-4">

                {/* Text Context */}
                <div className="space-y-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-bold text-slate-800 group-hover:text-blue-600 transition-colors truncate">
                      {card.title}
                    </h3>
                    <span className={`text-[9px] font-black px-1.5 py-0.5 rounded ${
                       'bg-green-50 text-green-600'
                    }`}>
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 leading-normal line-clamp-3">
                    {card.description}
                  </p>
                </div>
              </div>

              {/* Card Footer Action Label */}
              <div className="flex items-center justify-between border-t border-slate-50 pt-3">
                {
                  <div className="flex items-center text-xs font-bold text-blue-600 gap-1 opacity-80 group-hover:opacity-100 transform translate-x-0 group-hover:translate-x-1 transition-all">
                    Open Report
                  </div>
                }
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}