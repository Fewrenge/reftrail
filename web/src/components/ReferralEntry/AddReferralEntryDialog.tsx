import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

// Define the new Complaint structure to match your Go backend
interface Complaint {
  bodyPart: string;
  side: string;
  details: string;
}

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

const BODY_PARTS = ['SHOULDER', 'KNEE', 'HIP', 'ELBOW', 'WRIST', 'ANKLE', 'FOOT', 'OTHER'];
const SIDES = ['LEFT', 'RIGHT', 'BILATERAL'];

export default function AddReferralEntryDialog({ isOpen, onClose, onSuccess }: Props) {
  const [lastName, setLastName] = useState('');
  const[firstName, setFirstName]=useState('');
  const [source, setSource] = useState('REGULAR');
  const [urgency, setUrgency] = useState('Elective');
  // Now we manage an array of complaints
  const [complaints, setComplaints] = useState<Complaint[]>([
    { bodyPart: 'KNEE', side: 'LEFT', details: '' }
  ]);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.BaseSyntheticEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const res = await fetch('/api/v1/referrals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          patientLastName: lastName,
          patientFirstName: firstName,
          patientDob: "1990-01-01", // You might want to add a DOB field to your form!
          source: source,
          urgency: urgency,
          status: "READY_TO_BOOK",
          complaints: complaints // Sending the new array
        }),
      });

      if (res.ok) {
        setLastName('');
        setFirstName(''); 
        setComplaints([{ bodyPart: 'KNEE', side: 'LEFT', details: '' }]);
        onSuccess();
        onClose();
      }
    } catch (error) {
      console.error("Failed to save entry:", error);
    } finally {
      setLoading(false);
    }
  };

  const addComplaint = () => {
    setComplaints([...complaints, { bodyPart: 'KNEE', side: 'LEFT', details: '' }]);
  };

  const updateComplaint = (index: number, field: keyof Complaint, value: string) => {
    const newComplaints = [...complaints];
    newComplaints[index] = { ...newComplaints[index], [field]: value };
    setComplaints(newComplaints);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg overflow-y-auto max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold">New Referral</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          {/* Patient Name */}
          <div>
            <label className="text-[10px] font-bold text-slate-400 uppercase">Patient Name</label>
            <input required type="text" className="form-input-style" value={lastName} onChange={e => setLastName(e.target.value)} />
            <input required type="text" className="form-input-style" value={firstName} onChange={e => setFirstName(e.target.value)} />
          </div>

          {/* Dynamic Complaint List */}
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <label className="text-[10px] font-bold text-slate-400 uppercase">Complaints</label>
              <button type="button" onClick={addComplaint} className="text-xs text-blue-600 font-bold hover:underline">+ Add Part</button>
            </div>
            
            {complaints.map((c, index) => (
              <div key={index} className="grid grid-cols-2 gap-2 bg-slate-50 p-3 rounded-xl border border-slate-100">
                <select 
                  className="form-input-style"
                  value={c.bodyPart} 
                  onChange={e => updateComplaint(index, 'bodyPart', e.target.value)}
                >
                  {BODY_PARTS.map(part => <option key={part} value={part}>{part}</option>)}
                </select>
                <select 
                  className="form-input-style"
                  value={c.side} 
                  onChange={e => updateComplaint(index, 'side', e.target.value)}
                >
                  {SIDES.map(side => <option key={side} value={side}>{side}</option>)}
                </select>
                {c.bodyPart === 'OTHER' && (
                  <input 
                    placeholder="Describe part..." 
                    className="col-span-2 form-input-style text-xs" 
                    value={c.details} 
                    onChange={e => updateComplaint(index, 'details', e.target.value)}
                  />
                )}
              </div>
            ))}
          </div>

          <div className="grid grid-cols-2 gap-4">
             <div>
                <label className="text-[10px] font-bold text-slate-400 uppercase">Source</label>
                <select className="form-input-style" value={source} onChange={e => setSource(e.target.value)}>
                    <option value="REGULAR">Regular</option>
                    <option value="FRACTURE_CLINIC">Fracture Clinic</option>
                    <option value="OTHER">Other</option>
                </select>
             </div>
             <div>
                <label className="text-[10px] font-bold text-slate-400 uppercase">Urgency</label>
                <select className="form-input-style" value={urgency} onChange={e => setUrgency(e.target.value)}>
                    <option value="Elective">Elective</option>
                    <option value="Urgent">Urgent</option>
                    <option value="ASAP">ASAP</option>
                </select>
             </div>
          </div>

          <DialogFooter className="pt-4">
            <button type="submit" disabled={loading} className="save-btn-style">
              {loading ? "Saving..." : "Save to Referrals"}
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Note: You can define "form-input-style" and "save-btn-style" in your CSS 
// or keep your existing long Tailwind classes for consistency!
