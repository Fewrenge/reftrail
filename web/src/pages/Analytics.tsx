import { ChartNoAxesCombinedIcon, ArrowRightIcon} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar } from 'recharts';

// Miniature mock datasets to power the visual thumbnails on the preview cards
const MINI_PIE_DATA = [
  { name: 'ASAP', value: 40 },
  { name: 'Urgent', value: 35 },
  { name: 'Elective', value: 25 }
];
const PIE_COLORS = ['#ef4444', '#f59e0b', '#3b82f6'];

const MINI_BAR_DATA = [
  { name: 'M', v: 20 }, { name: 'T', v: 40 }, { name: 'W', v: 35 },
  { name: 'T', v: 50 }, { name: 'F', v: 30 }, { name: 'S', v: 10 }
];


export default function AnalyticsPage() {
    const navigate = useNavigate();

  // Unified configuration matrix for your dashboard collections
  const reportCards = [
    {
      id: 'urgency-distribution',
      title: 'Urgency Distribution',
      description: 'Breakdown percentages of Elective, Urgent, and ASAP referral entries.',
      path: '/analytics/urgency-distribution',
      status: 'Ready',
      // Inline mini-recharts component for the card preview window
      thumbnail: (
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie data={MINI_PIE_DATA} cx="50%" cy="50%" innerRadius={18} outerRadius={30} paddingAngle={2} dataKey="value">
              {MINI_PIE_DATA.map((entry, idx) => (
                <Cell key={`cell-${idx}`} fill={PIE_COLORS[idx]} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
      )
    },
    {
      id: 'volume-over-time',
      title: 'Referral Volume Trends',
      description: 'Track incoming referral velocity trajectories over daily or monthly timelines.',
      path: '/analytics/volume-over-time',
      status: 'Coming Soon',
      thumbnail: (
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={MINI_BAR_DATA} margin={{ top: 10, bottom: 5, left: 5, right: 5 }}>
            <Bar dataKey="v" fill="#cbd5e1" radius={[2, 2, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      )
    }
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
          const isComingSoon = card.status === 'Coming Soon';

          return (
            <div
              key={card.id}
              onClick={() => !isComingSoon && navigate(card.path)}
              className={`group border bg-white rounded-2xl p-5 flex flex-col justify-between h-56 transition-all duration-200 ${
                isComingSoon
                  ? 'border-slate-100 opacity-60 cursor-not-allowed'
                  : 'border-slate-200 hover:border-blue-500 hover:shadow-md cursor-pointer'
              }`}
            >
              <div className="flex items-start gap-4">
                {/* Visual Thumbnail Window Box */}
                <div className={`w-20 h-20 shrink-0 rounded-xl border flex items-center justify-center overflow-hidden transition-colors ${
                  isComingSoon ? 'bg-slate-50 border-slate-100' : 'bg-slate-50 border-slate-200 group-hover:bg-blue-50/30 group-hover:border-blue-200'
                }`}>
                  {card.thumbnail}
                </div>

                {/* Text Context Content */}
                <div className="space-y-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-bold text-slate-800 group-hover:text-blue-600 transition-colors truncate">
                      {card.title}
                    </h3>
                    <span className={`text-[9px] font-black px-1.5 py-0.5 rounded ${
                      isComingSoon ? 'bg-slate-100 text-slate-400' : 'bg-green-50 text-green-600'
                    }`}>
                      {card.status}
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 leading-normal line-clamp-3">
                    {card.description}
                  </p>
                </div>
              </div>

              {/* Card Footer Action Label */}
              <div className="flex items-center justify-between border-t border-slate-50 pt-3">
                <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                  Report File
                </span>
                {!isComingSoon && (
                  <div className="flex items-center text-xs font-bold text-blue-600 gap-1 opacity-80 group-hover:opacity-100 transform translate-x-0 group-hover:translate-x-1 transition-all">
                    Open Report <ArrowRightIcon size={14} />
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}