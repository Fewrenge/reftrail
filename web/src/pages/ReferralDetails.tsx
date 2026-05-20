import { useParams, useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import ReferralEntryCard from "../components/ReferralEntry/ReferralEntryCard";
import type {ReferralEntry} from "../components/ReferralEntry/ReferralEntryCard";

export default function ReferralDetails() {
  const { referralId } = useParams<{ referralId: string }>();
  const navigate = useNavigate();
  
  const [referral, setReferral] = useState<ReferralEntry | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchReferralDetail = async () => {
    try {
      const res = await fetch(`/api/v1/referrals/${referralId}`);
      if (res.ok) {
        const data = await res.json();
        setReferral(data);
      }
    } catch (err) {
      console.error("Failed to sniper target referral:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReferralDetail();
  }, [referralId]);

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

      {/* LOGS PANEL PLACEHOLDER (Future Phase) */}
      <div className="border border-dashed border-slate-200 rounded-2xl p-8 bg-slate-50/50 flex flex-col items-center justify-center text-center">
        <p className="text-sm font-medium text-slate-400">Audit Logs & History</p>
        <p className="text-xs text-slate-400 mt-1">System operational logs will load here once the backend is linked up.</p>
      </div>
    </div>
  );
}
