import { useEffect, useState,  } from "react";
import { ArrowLeftIcon, CalendarIcon, Loader2Icon, RefreshCwIcon, TrendingUpIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

// 1. Define matching TypeScript interface for Go TrendMetric struct
interface TrendMetric {
  period: string; // "YYYY-MM"
  count: number;
}

interface ReferralTrendResponse {
  data: TrendMetric[] | null;
  totalCount: number;
}

export default function ReferralTrend() {
  const navigate = useNavigate();
  
  // Date filter states (Defaults to trailing 6 months from current 2026 date context)
  const [dateFrom, setDateFrom] = useState("2026-01-01");
  const [dateTo, setDateTo] = useState("2026-06-30");
  
  const [trendData, setTrendData] = useState<TrendMetric[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 2. Data Fetching Execution
  const fetchTrends = async () => {
    setLoading(true);
    setError(null);
    try {
      // Constructs query string arguments exactly matching echo.Bind rules
      const params = new URLSearchParams();
      if (dateFrom) params.append("referralDateFrom", dateFrom);
      if (dateTo) params.append("referralDateTo", dateTo);

      const response = await fetch(`/api/v1/analytics/referral-trend?${params.toString()}`);
      if (!response.ok) {
        throw new Error(`Server responded with status ${response.status}`);
      }
      
      const payload: ReferralTrendResponse = await response.json();
      setTrendData(payload.data || []);
      setTotalCount(payload.totalCount);
    } catch (err: any) {
      console.error("Failed fetching trend metrics:", err);
      setError(err.message || "An error occurred while compiling charts");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTrends();
  }, [dateFrom, dateTo]);

  // 3. Custom Date Format Helper for Recharts X-Axis Labels (e.g. "2026-06" -> "Jun 26")
  const formatXAxis = (periodStr: string) => {
    if (!periodStr || periodStr === "Unknown") return periodStr;
    try {
      const [year, month] = periodStr.split("-");
      const date = new Date(parseInt(year), parseInt(month) - 1, 1);
      return date.toLocaleDateString("en-US", { month: "short", year: "2-digit" });
    } catch {
      return periodStr;
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* ACTION HEADER */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100">
        <div className="space-y-1">
          <button
            onClick={() => navigate("/analytics")}
            className="flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-blue-600 transition-colors group mb-1 cursor-pointer"
          >
            <ArrowLeftIcon size={14} className="transform group-hover:-translate-x-0.5 transition-transform" />
            Back to Dashboard
          </button>
          <div className="flex items-center gap-2 text-indigo-600">
            <TrendingUpIcon size={22} strokeWidth={2.5} />
            <h1 className="text-2xl font-bold text-slate-900">Referral Volume Trends</h1>
          </div>
          <p className="text-sm text-slate-500">Chronological pipeline volume and distribution changes over time.</p>
        </div>

        {/* FILTERS TOOLBAR */}
        <div className="flex flex-wrap items-center gap-3 bg-slate-50 p-2 rounded-xl border border-slate-200/60">
          <div className="flex items-center gap-1.5 text-slate-500 text-xs font-semibold px-1">
            <CalendarIcon size={14} />
            Range:
          </div>
          <input
            type="date"
            max="9999-12-31" // Implements your 4-digit input enforcement rule!
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="p-1.5 border bg-white rounded-lg text-xs font-medium text-slate-700 border-slate-200 focus:outline-blue-500"
          />
          <span className="text-slate-400 text-xs">to</span>
          <input
            type="date"
            max="9999-12-31"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="p-1.5 border bg-white rounded-lg text-xs font-medium text-slate-700 border-slate-200 focus:outline-blue-500"
          />
          <button
            onClick={fetchTrends}
            disabled={loading}
            className="p-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-100 active:scale-95 transition-all cursor-pointer disabled:opacity-50"
            title="Refresh Report"
          >
            <RefreshCwIcon size={14} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      {/* OVERVIEW STAT CARD */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white p-5 border border-slate-200/80 rounded-2xl shadow-sm">
          <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Evaluated Traffic</span>
          <h2 className="text-3xl font-black text-slate-800 mt-1">{totalCount} <span className="text-xs font-normal text-slate-400">referrals</span></h2>
        </div>
      </div>

      {/* CHART CONTENT LAYER */}
      <div className="bg-white p-6 border border-slate-200/80 rounded-2xl shadow-sm min-h-100 flex flex-col justify-center">
        {loading ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-400 text-sm">
            <Loader2Icon className="animate-spin text-blue-600" size={32} />
            Calculating database timeline aggregates...
          </div>
        ) : error ? (
          <div className="text-center py-20 text-sm font-medium text-red-600 bg-red-50/50 rounded-xl border border-red-100 max-w-xl mx-auto p-6">
            {error}
          </div>
        ) : trendData.length === 0 ? (
          <div className="text-center py-20 text-slate-400 text-sm">
            No referral data entries exist within the selected date window boundary parameters.
          </div>
        ) : (
          <div className="w-full h-87.5 pt-4">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trendData} margin={{ top: 10, right: 30, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" vertical={false} />
                <XAxis 
                  dataKey="period" 
                  tickFormatter={formatXAxis}
                  tick={{ fill: '#64748b', fontSize: 11 }}
                  axisLine={{ stroke: '#cbd5e1' }}
                  tickLine={false}
                />
                <YAxis 
                  allowDecimals={false}
                  tick={{ fill: '#64748b', fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip 
                  labelFormatter={(label) => `Timeline Block: ${formatXAxis(label)}`}
                  contentStyle={{ backgroundColor: '#ffffff', borderRadius: '12px', border: '1px solid #e2e8f0', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                />
                <Legend verticalAlign="top" height={36} iconType="circle" />
                <Line
                  name="Incoming Referrals Volume"
                  type="monotone"
                  dataKey="count"
                  stroke="#4f46e5" // Darker indigo color line
                  strokeWidth={3}
                  dot={{ r: 4, stroke: '#4f46e5', strokeWidth: 2, fill: '#ffffff' }}
                  activeDot={{ r: 7, strokeWidth: 0 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  );
}
