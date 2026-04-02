import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function AddWLEntryDialog({ isOpen, onClose, onSuccess }: Props) {
  const [name, setName] = useState('');
  const [complaint, setComplaint] = useState('');
  const [urgency, setUrgency] = useState('Elective');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.BaseSyntheticEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const res = await fetch('/api/v1/waitlist', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          patientName: name,
          complaint: complaint,
          urgency: urgency,
          state: "Ready to book"
        }),
      });

      if (res.ok) {
        setName(''); 
        setComplaint('');
        onSuccess();
        onClose();
      }
    } catch (error) {
      console.error("Failed to save entry:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold">New Referral</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">
              Patient Name
            </label>
            <input 
              required 
              type="text" 
              className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all bg-white text-slate-900" 
              value={name} 
              onChange={e => setName(e.target.value)} 
            />
          </div>

          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">
              Complaint
            </label>
            <input 
              required 
              type="text" 
              className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all bg-white text-slate-900" 
              value={complaint} 
              onChange={e => setComplaint(e.target.value)} 
            />
          </div>

          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">
              Urgency
            </label>
            <select 
              className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all bg-white text-slate-900"
              value={urgency} 
              onChange={e => setUrgency(e.target.value)}
            >
              <option value="Elective">Elective</option>
              <option value="Urgent">Urgent</option>
              <option value="ASAP">ASAP</option>
            </select>
          </div>

          <DialogFooter className="pt-4">
            <button 
              type="submit" 
              disabled={loading}
              className="w-full bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg shadow-blue-500/20 hover:bg-blue-700 transition-all disabled:opacity-50"
            >
              {loading ? "Saving..." : "Save to Waitlist"}
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
