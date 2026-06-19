import { useEffect, useState, useMemo } from "react";
import { ArrowLeftIcon, CalendarIcon, Loader2Icon, RefreshCwIcon, ClockIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Legend,
  LabelList,
  ResponsiveContainer,
} from "recharts";

// TODO: fix date range based filtering
interface WaitingTimeMetric {
  period: string; // "YYYY-MM"
  averageDays: number;
}

interface WaitingTimeResponse {
  data: WaitingTimeMetric[] | null;
}

export default function DirectBookingWaitingTime() {
  const navigate = useNavigate();

  const [dateFrom, setDateFrom] = useState("2026-01-01");
  const [dateTo, setDateTo] = useState("2026-06-30");
  const [chartData, setChartData] = useState<WaitingTimeMetric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchWaitingTimes = async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (dateFrom) params.append("referralDateFrom", dateFrom);
      if (dateTo) params.append("referralDateTo", dateTo);

      // Matches your Go endpoint path precisely
      const response = await fetch(`/api/v1/analytics/direct-booking-waiting-time?${params.toString()}`);
      if (!response.ok) {
        throw new Error(`Server connection failure with status code ${response.status}`);
      }

      const payload: WaitingTimeResponse = await response.json();
      setChartData(payload.data || []);
    } catch (err: any) {
      console.error("Error fetching wait time data:", err);
      setError(err.message || "Failed to load operational speed metrics");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWaitingTimes();
  }, [dateFrom, dateTo]);

  const formatXAxis = (periodStr: string) => {
    if (!periodStr || periodStr === "Unknown") return periodStr;
    try {
      const [year, month] = periodStr.split("-");
      const date = new Date(parseInt(year), parseInt(month) - 1, 1);
      return date.toLocaleDateString("en-CA", { month: "short", year: "numeric" });
    } catch {
      return periodStr;
    }
  };

  // Computes system overall speed average across months for metadata card display
  const averageSystemSpeed = useMemo(() => {
    if (chartData.length === 0) return 0;
    const total = chartData.reduce((sum, item) => sum + item.averageDays, 0);
    return parseFloat((total / chartData.length).toFixed(1));
  }, [chartData]);

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* HEADER ACTION WORKSPACE */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100">
        <div className="space-y-1">
          <button
            onClick={() => navigate("/analytics")}
            className="flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-blue-600 transition-colors group mb-1 cursor-pointer"
          >
            <ArrowLeftIcon size={14} className="transform group-hover:-translate-x-0.5 transition-transform" />
            Back to Dashboard
          </button>
          <div className="flex items-center gap-2 text-emerald-600">
            <ClockIcon size={22} strokeWidth={2.5} />
            <h1 className="text-2xl font-bold text-slate-900">Direct Processing Velocity</h1>
          </div>
          <p className="text-sm text-slate-500">Average days elapsed from file generation straight to booking completion.
            Only calculates referrals that go straight from ready to book to booked.
          </p>
        </div>

        {/* DATE RANGE CONTROLS */}
        <div className="flex flex-wrap items-center gap-3 bg-slate-50 p-2 rounded-xl border border-slate-200/60">
          <div className="flex items-center gap-1.5 text-slate-500 text-xs font-semibold px-1">
            <CalendarIcon size={14} />
          </div>
          <input
            type="date"
            max="9999-12-31" // Restricts input mask to 4-digits natively
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
            onClick={fetchWaitingTimes}
            disabled={loading}
            className="p-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-100 active:scale-95 transition-all cursor-pointer disabled:opacity-50"
          >
            <RefreshCwIcon size={14} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      {/* COMPRESSED VALUE OVERVIEW BLOCK */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white p-5 border border-slate-200/80 rounded-2xl shadow-sm">
          <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Perfect Workflow Average</span>
          <h2 className="text-3xl font-black text-slate-800 mt-1">{averageSystemSpeed} <span className="text-xs font-normal text-slate-400">days</span></h2>
        </div>
      </div>

      {/* RECHARTS PLOT DISPLAY LAYER */}
      <div className="bg-white p-6 border border-slate-200/80 rounded-2xl shadow-sm min-h-100 flex flex-col justify-center">
        {loading ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-400 text-sm">
            <Loader2Icon className="animate-spin text-blue-600" size={32} />
            Compiling velocity historical metrics...
          </div>
        ) : error ? (
          <div className="text-center py-20 text-sm font-medium text-red-600 bg-red-50/50 rounded-xl border border-red-100 max-w-xl mx-auto p-6">
            {error}
          </div>
        ) : chartData.length === 0 ? (
          <div className="text-center py-20 text-slate-400 text-sm">
            No direct bookings exist within your active filter parameters.
          </div>
        ) : (
          <div className="w-full h-87.5 pt-4">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 25, right: 35, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" vertical={false} />
                <XAxis
                  dataKey="period"
                  tickFormatter={formatXAxis}
                  tick={{ fill: '#64748b', fontSize: 11 }}
                  axisLine={{ stroke: '#cbd5e1' }}
                  tickLine={false}
                  padding={{ left: 25, right: 30 }}
                />
                <YAxis
                  allowDecimals={true}
                  tick={{ fill: '#64748b', fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  unit=" d"
                />
                <Legend verticalAlign="top" height={36} iconType="circle" />
                <Line
                  name="Average Execution Duration"
                  type="linear" // Straight point-to-point lines
                  dataKey="averageDays"
                  stroke="#10b981" // Custom Emerald tracking theme line color
                  strokeWidth={3}
                  dot={{ r: 4, stroke: '#10b981', strokeWidth: 2, fill: '#ffffff' }}
                  activeDot={false}
                >
                  {/* Clean text values positioned to the top-right of points, preventing line overlap */}
                  <LabelList
                    dataKey="averageDays"
                    position="top"
                    offset={6}
                    style={{ fill: '#334155', fontSize: 11, fontWeight: 700 }}
                  />
                </Line>
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  );
}
