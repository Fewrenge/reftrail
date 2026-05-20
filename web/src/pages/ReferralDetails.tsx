import { useParams, useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Button } from "@/components/ui/button";
import { ArrowLeft, ClockIcon, MessageSquareIcon, User } from "lucide-react";
import ReferralEntryCard from "../components/ReferralEntry/ReferralEntryCard";
import type { ReferralEntry } from "../components/ReferralEntry/ReferralEntryCard";

interface ReferralLog {
  id: string;
  entryId: string;
  userId: number;
  oldStatus: string;
  newStatus: string;
  note: string;
  createdTs: string;
}

export default function ReferralDetails() {
  const { referralId } = useParams<{ referralId: string }>();
  const navigate = useNavigate();

  const [referral, setReferral] = useState<ReferralEntry | null>(null);
  const [logs, setLogs] = useState<ReferralLog[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchReferralDetail = async () => {
    try {
      // Parallel fetch for both referral state and historical logs
      const [refRes, logsRes] = await Promise.all([
        fetch(`/api/v1/referrals/${referralId}`),
        fetch(`/api/v1/referrals/${referralId}/logs`)
      ]);

      if (refRes.ok) {
        const refData = await refRes.json();
        setReferral(refData);
      }

      if (logsRes.ok) {
        const logsData = await logsRes.json();
        setLogs(logsData);
      }
    } catch (err) {
      console.error("Failed to fetch timeline data:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReferralDetail();
  }, [referralId]);

    const formatTime = (tsString: string) => {
    try {
      const date = new Date(tsString);
      return date.toLocaleDateString(undefined, { 
        month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' 
      });
    } catch {
      return tsString;
    }
  };

  if (loading) return <div className="p-6 text-sm text-slate-400">Targeting record entry...</div>;
  if (!referral) return <div className="p-6 text-sm text-red-500">Referral records missing or deleted.</div>;

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Control Navigation Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate('/referrals')}
          className="gap-2 text-slate-600"
        >
          <ArrowLeft className="h-4 w-4" /> Back to Referrals
        </Button>
        <span className="text-xs font-mono text-slate-400 bg-slate-100 px-2 py-1 rounded">
          ID: {referralId}
        </span>
      </div>

      {/* SNIPER CARD VIEW */}
      <div className="[&_>_div]:shadow-none [&_>_div]:bg-transparent [&_>_div]:border-none [&_>_div]:p-0">
        <ReferralEntryCard
          entry={referral}
          onRefresh={fetchReferralDetail}
          isClickable={false} // Disable navigation on the card when in detail view
        />
      </div>

      {/* LOGS PANEL (Future Phase) */}
       <div className="space-y-4">
        <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
          <ClockIcon size={14} /> Audit Trail History
        </h4>

        {logs.length === 0 ? (
          <div className="border border-dashed border-slate-200 rounded-2xl p-6 text-center text-xs text-slate-400 bg-slate-50/50">
            No historical logs or automated transitions tracked for this entry.
          </div>
        ) : (
          <div className="relative border-l border-slate-200 pl-4 ml-2 space-y-6">
            {logs.map((log) => {
              const isStatusChange = log.oldStatus !== log.newStatus;
              
              return (
                <div key={log.id} className="relative group">
                  {/* Timeline Bullet Anchor */}
                  <div className={`absolute -left-5.25 top-1 w-3 h-3 rounded-full border-2 bg-white ${
                    isStatusChange ? 'border-blue-500' : 'border-slate-300'
                  }`} />
                  
                  <div className="space-y-1">
                    {/* Log Header Metadata */}
                    <div className="flex flex-wrap items-center gap-x-2 text-xs text-slate-500">
                      <span className="font-semibold text-slate-700 flex items-center gap-0.5">
                        <User size={12} className="text-slate-400" /> User #{log.userId}
                      </span>
                      <span className="text-slate-300">•</span>
                      <span>{formatTime(log.createdTs)}</span>
                    </div>

                    {/* Operational Event Text */}
                    {isStatusChange ? (
                      <p className="text-sm text-slate-800 font-medium">
                        Changed status from{' '}
                        <span className="px-1.5 py-0.5 bg-slate-100 rounded text-xs text-slate-600 font-mono">
                          {log.oldStatus.replace(/_/g, ' ')}
                        </span>{' '}
                        →{' '}
                        <span className="px-1.5 py-0.5 bg-blue-50 border border-blue-100 rounded text-xs text-blue-700 font-bold font-mono">
                          {log.newStatus.replace(/_/g, ' ')}
                        </span>
                      </p>
                    ) : (
                      <p className="text-sm text-slate-500 font-medium italic flex items-center gap-1">
                        <MessageSquareIcon size={12} className="text-slate-400" /> Added internal case note
                      </p>
                    )}

                    {/* Associated Note Box */}
                    {log.note && (
                      <div className="bg-slate-50 border border-slate-100 p-2.5 rounded-xl text-xs text-slate-600 max-w-xl mt-1">
                        "{log.note}"
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>



    </div>
  );
}
