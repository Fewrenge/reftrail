import React, { useState } from 'react';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function AddEntryModal({ isOpen, onClose, onSuccess }: Props) {
  const [name, setName] = useState('');
  const [complaint, setComplaint] = useState('');
  const [urgency, setUrgency] = useState('Elective');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch('/api/v1/waitlist', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        patientName: name,
        complaint: complaint,
        urgency: urgency,
        state: "Ready to book" // Default state
      }),
    });

    if (res.ok) {
      setName(''); setComplaint(''); // Clear form
      onSuccess(); // Refresh the list
      onClose();   // Close popup
    }
  };

  return (
    <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-md p-6 border border-slate-200">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-bold">New Referral</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-2xl">&times;</button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">Patient Name</label>
            <input required type="text" className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" 
              value={name} onChange={e => setName(e.target.value)} />
          </div>
          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">Complaint</label>
            <input required type="text" className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" 
              value={complaint} onChange={e => setComplaint(e.target.value)} />
          </div>
          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">Urgency</label>
            <select className="w-full border border-slate-200 rounded-xl px-4 py-2 mt-1 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all bg-white"
              value={urgency} onChange={e => setUrgency(e.target.value)}>
              <option>Elective</option>
              <option>Urgent</option>
              <option>ASAP</option>
            </select>
          </div>
          <button type="submit" className="w-full bg-blue-600 text-white py-3 rounded-xl font-bold mt-4 shadow-lg shadow-blue-500/20 hover:bg-blue-700 transition-all">
            Save to Waitlist
          </button>
        </form>
      </div>
    </div>
  );
}
