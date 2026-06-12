// TODO: add more fields

import React, { useState, useMemo } from 'react';
import { Button } from "@/components/ui/button";
import { AlertCircleIcon, PlusIcon, Trash2Icon } from "lucide-react"; // Visual anchors for scannability
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

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
const SIDES = ['LEFT', 'RIGHT', 'BILATERAL', 'OTHER'];

export default function AddReferralEntryDialog({ isOpen, onClose, onSuccess }: Props) {
  const [lastName, setLastName] = useState('');
  const [firstName, setFirstName] = useState('');
  const [source, setSource] = useState('REGULAR');
  const [urgency, setUrgency] = useState('ELECTIVE');
  const [consultType, setConsultType] = useState('');
  const [complaints, setComplaints] = useState<Complaint[]>([
    //{ bodyPart: 'KNEE', side: 'LEFT', details: '' }
  ]);
  const [loading, setLoading] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const duplicateBodyPart = useMemo(() => {
    const seenBodyParts = new Set<string>();
    for (const c of complaints) {
      const normalizedPart = c.bodyPart.toUpperCase().trim();
      if (!normalizedPart) continue;

      if (seenBodyParts.has(normalizedPart)) {
        return c.bodyPart; // Returns the name of the duplicate part to trigger our flags
      }
      seenBodyParts.add(normalizedPart);
    }
    return null; // All clean
  }, [complaints]);

  // Submit button state flag
  const isSubmitDisabled = loading || !!duplicateBodyPart;

  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isSubmitDisabled) return;
    setLoading(true);
    setBackendError(null);

    try {
      const res = await fetch('/api/v1/referrals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          patientLastName: lastName,
          patientFirstName: firstName,
          patientDob: "1990-01-01",
          source: source,
          urgency: urgency,
          status: "READY_TO_BOOK",
          complaints: complaints,
          consultTYpe: "APP+SX",
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

  const removeComplaint = (index: number) => {
    if (complaints.length === 1) return; // Maintain at least one row
    setComplaints(complaints.filter((_, i) => i !== index));
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

        {/* Backend tracking error if database fails */}
        {backendError && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl text-sm font-medium flex items-center gap-2">
            <AlertCircleIcon size={16} className="text-red-500" />
            <span>{backendError}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          {/* Patient Name Section */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">First Name</label>
              <input
                required
                type="text"
                className="border rounded-md p-2 bg-white text-slate-900"
                value={firstName}
                onChange={e => setFirstName(e.target.value)}
                placeholder="e.g. Jane"
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Last Name</label>
              <input
                required
                type="text"
                className="border rounded-md p-2 bg-white text-slate-900"
                value={lastName}
                onChange={e => setLastName(e.target.value)}
                placeholder="e.g. Doe"
              />
            </div>
          </div>

          {/* Dynamic Complaint List */}
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <label className="text-sm font-medium">Complaints</label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addComplaint}
                className="text-xs flex items-center gap-1"
              >
                <PlusIcon className="h-3 w-3" /> Add Part
              </Button>
            </div>

            {complaints.map((c, index) => (
              <div key={index} className="space-y-2 bg-slate-50 p-3 rounded-md border border-slate-200">
                <div className="flex items-center gap-2">
                  <select
                    className="w-full border rounded-md p-2 bg-white text-sm"
                    value={c.bodyPart}
                    onChange={e => updateComplaint(index, 'bodyPart', e.target.value)}
                  >
                    {BODY_PARTS.map(part => <option key={part} value={part}>{part}</option>)}
                  </select>

                  <select
                    className="w-full border rounded-md p-2 bg-white text-sm"
                    value={c.side}
                    onChange={e => updateComplaint(index, 'side', e.target.value)}
                  >
                    {SIDES.map(side => <option key={side} value={side}>{side}</option>)}
                  </select>

                  {complaints.length > 1 && (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={() => removeComplaint(index)}
                      className="text-slate-500 hover:text-red-600 shrink-0"
                    >
                      <Trash2Icon className="h-4 w-4" />
                    </Button>
                  )}
                </div>

                {c.bodyPart === 'OTHER' && (
                  <input
                    placeholder="Describe part (e.g., Femur)..."
                    className="w-full border rounded-md p-2 bg-white text-slate-900 text-sm"
                    value={c.details}
                    onChange={e => updateComplaint(index, 'details', e.target.value)}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Metadata Section */}
          <div className="grid grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Source</label>
              <select className="border rounded-md p-2 bg-white text-sm" value={source} onChange={e => setSource(e.target.value)}>
                <option value="REGULAR">Regular</option>
                <option value="FRACTURE_CLINIC">Fracture Clinic</option>
                <option value="OTHER">Other</option>
              </select>
            </div>
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Urgency</label>
              <select className="border rounded-md p-2 bg-white text-sm" value={urgency} onChange={e => setUrgency(e.target.value)}>
                <option value="ELECTIVE">Elective</option>
                <option value="URGENT">Urgent</option>
                <option value="ASAP">ASAP</option>
              </select>
            </div>
          </div>

          {/* Real-time small red validation message box at bottom footer zone */}
          {duplicateBodyPart && (
            <div className="text-xs font-semibold text-red-600 flex items-center gap-1.5 pt-2 animate-in fade-in-50 slide-in-from-top-1 duration-100">
              <AlertCircleIcon size={14} className="shrink-0" />
              <span>Cannot save: "{duplicateBodyPart.toLowerCase()}" is selected more than once.</span>
            </div>
          )}

          {/* Actions */}
          <DialogFooter className="pt-4">
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitDisabled}
              className={isSubmitDisabled ? "bg-slate-400/20 text-slate-400 cursor-not-allowed hover:bg-slate-400/20 shadow-none border-transparent" : ""}
            >
              {loading ? "Saving..." : "Save to Referrals"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
