import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeftIcon, PieChartIcon } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';

// Data Transfer Object Interfaces
interface UrgencyMetric {
  urgency: string; // "Elective", "Urgent", "ASAP", "Unassigned"
  count: number;
  percentage: number;
}

interface UrgencyDistributionResponse {
  metrics: UrgencyMetric[];
  totalCount: number;
}

// Color theme variables for the chart slices
const COLORS: Record<string, string> = {
  ASAP: '#ef4444',       // Red-500
  URGENT: '#f59e0b',     // Amber-500
  ELECTIVE: '#3b82f6',   // Blue-500
};

export default function UrgencyDistribution() {
  const navigate = useNavigate();
  
  // 1. Time Window Filter State (Defaults to Year to Date)
  const [filterType, setFilterType] = useState<string>('ytd');
  const [dateRange, setDateRange] = useState({
    from: `${new Date().getFullYear()}-01-01`,
    to: new Date().toISOString().split('T')[0]
  });

  // 2. Request Lifecycle Processing States
  const [data, setData] = useState<UrgencyDistributionResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // 3. Dynamic Side Effect Data Sync Engine
  useEffect(() => {
    const fetchMetrics = async () => {
      setLoading(true);
      setError(null);
      try {
        let url = '/api/v1/analytics/urgency-distribution';
        const params = new URLSearchParams();
        if (dateRange.from) params.append('referralDateFrom', dateRange.from);
        if (dateRange.to) params.append('referralDateTo', dateRange.to);
        
        if (params.toString()) {
          url += `?${params.toString()}`;
        }

        const response = await fetch(url);
        if (!response.ok) {
          throw new Error('Failed to retrieve computed analytics metrics');
        }
        
        const json: UrgencyDistributionResponse = await response.json();
        setData(json);
      } catch (err: any) {
        setError(err.message || 'An error occurred fetching dashboard matrices.');
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
  }, [dateRange]);

  // 4. Quick-Select Date Shortcut Calculator
  const handleFilterChange = (type: string) => {
    setFilterType(type);
    const today = new Date();
    const currentYear = today.getFullYear();
    const currentMonth = today.getMonth();

    let from = '';
    let to = today.toISOString().split('T')[0];

    if (type === 'month') {
      from = `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}-01`;
    } else if (type === 'quarter') {
      const quarterStartMonth = Math.floor(currentMonth / 3) * 3 + 1;
      from = `${currentYear}-${String(quarterStartMonth).padStart(2, '0')}-01`;
    } else if (type === 'ytd') {
      from = `${currentYear}-01-01`;
    }

    setDateRange({ from, to });
  };

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      {/* NAVIGATION AND TIMELINE CHANGER SUB BAR */}
      <div className="flex items-center justify-between">
        <button 
          onClick={() => navigate('/analytics')}
          className="flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-slate-800 transition-colors cursor-pointer group"
        >
          <ArrowLeftIcon size={14} className="transform group-hover:-translate-x-0.5 transition-transform" /> 
          Back to Dashboard
        </button>

        {/* Dynamic Shortcut Selector Pills */}
        <div className="flex items-center bg-white border border-slate-200 rounded-xl p-1 shadow-sm gap-0.5">
          {[
            { id: 'month', label: 'This Month' },
            { id: 'quarter', label: 'This Quarter' },
            { id: 'ytd', label: 'Year to Date' }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => handleFilterChange(tab.id)}
              className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all cursor-pointer ${
                filterType === tab.id
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* COMPONENT PRESENTATION GRAPH SHEET */}
      <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
        <div className="flex justify-between items-start border-b border-slate-100 pb-5 mb-6">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 text-blue-600 rounded-xl">
              <PieChartIcon size={20} />
            </div>
            <div>
              <h2 className="text-lg font-bold text-slate-900">Urgency Volume Distribution</h2>
              <p className="text-xs text-slate-400 mt-0.5">Real-time breakdown of priority triage vectors.</p>
            </div>
          </div>
          {data && (
            <div className="bg-slate-50 px-4 py-2 rounded-xl text-right border border-slate-100">
              <span className="text-[10px] uppercase tracking-wider font-bold text-slate-400">Total Count</span>
              <p className="text-2xl font-black text-slate-800 mt-0.5">{data.totalCount}</p>
            </div>
          )}
        </div>

        {/* GRAPH CORE WRAPPER SCREEN */}
        <div className="h-95 w-full min-h-0 min-w-0 mt-4 relative">
          {loading && (
            <div className="h-full flex items-center justify-center text-slate-400 text-xs">
              Calculating database metrics...
            </div>
          )}
          
          {error && (
            <div className="h-full flex items-center justify-center text-red-500 text-xs font-medium">
              {error}
            </div>
          )}

          {!loading && !error && (!data || data.totalCount === 0) && (
            <div className="h-full flex items-center justify-center text-slate-400 text-xs border border-dashed border-slate-200 rounded-xl">
              No referral entries recorded inside this time scope.
            </div>
          )}

          {!loading && !error && data && data.totalCount > 0 && (
            <ResponsiveContainer width="100%" aspect={2}>
              <PieChart>
                <Pie
                  data={data.metrics}
                  cx="50%"
                  cy="45%"
                  innerRadius={75}
                  outerRadius={105}
                  paddingAngle={4}
                  dataKey="count"
                  nameKey="urgency"
                >
                  {data.metrics.map((entry) => (
                    <Cell 
                      key={`cell-${entry.urgency}`} 
                      fill={COLORS[entry.urgency] || COLORS.Unassigned} 
                    />
                  ))}
                </Pie>
                <Tooltip 
                  formatter={(value: any, name: any, props: any) => {
                    const rawValue = value !== undefined ? Number(value) : 0;
                    const percentage = props?.payload?.percentage !== undefined ? Number(props.payload.percentage) : 0;
                    return [
                      `${rawValue} referrals (${percentage.toFixed(1)}%)`, 
                      String(name || 'Category')
                    ];
                  }}
                  contentStyle={{ background: '#fff', borderRadius: '12px', border: '1px solid #e2e8f0', fontSize: '12px', boxShadow: '0 1px 3px 0 rgb(0 0 0 / 0.1)' }}
                />
                <Legend 
                  verticalAlign="bottom" 
                  iconType="circle"
                  iconSize={8}
                  wrapperStyle={{ fontSize: '12px', color: '#64748b', paddingTop: '10px' }}
                />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  );
}
