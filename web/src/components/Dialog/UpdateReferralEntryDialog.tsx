"use client"

import { useEffect, useState } from "react"
import { PlusIcon, Trash2Icon, ShieldAlertIcon } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog"

import type {
  ReferralStatus,
  ReferralUrgency,
  ReferralConsultType,
  ReferralSource
} from "@/types/referrals"


export interface FrontEndComplaint {
  bodyPart: string
  side: string
  details: string
}

export interface AdminUpdatePayload {
  id: string
  status?: ReferralStatus
  urgency?: ReferralUrgency
  source?: ReferralSource
  triageNote?: string
  referringPhysician?: string
  consultType?: ReferralConsultType
  referralDate?: string
  emrPatientId?: string
  emrReferralDocID?: string
  complaints?: FrontEndComplaint[]
  //note?: string
  force: boolean
}

interface UpdateReferralEntryDialogProps {
  isOpen: boolean
  onClose: () => void
  referralId: string
  // Pass down the current record data to pre-populate form states
  initialData?: {
    status: ReferralStatus
    urgency: ReferralUrgency
    source: ReferralSource
    triageNote: string
    referringPhysician: string
    consultType: ReferralConsultType
    referralDate: string
    emrPatientId: string
    emrReferralDocID: string
    emrApptId: string
    complaints: FrontEndComplaint[]
  }
  onSave: (payload: AdminUpdatePayload) => Promise<void>
}

// Core static data pools matching database constraints
const BODY_PARTS = ['SHOULDER', 'KNEE', 'HIP', 'ELBOW', 'WRIST', 'ANKLE', 'FOOT', 'OTHER']
const SIDES = ['LEFT', 'RIGHT', 'BILATERAL', 'OTHER']
const STATUS_OPTIONS: ReferralStatus[] = ['READY_TO_BOOK', '1ST_CALL_COMPLETE', '2ND_CALL_COMPLETE', '3RD_CALL_COMPLETE', 'BOOKED', 'UNABLE_TO_CONTACT', 'PATIENT_TO_CALL_BACK', 'DECLINED', 'SUSPENDED', 'CLOSED']
const URGENCY_OPTIONS: ReferralUrgency[] = ['ELECTIVE', 'URGENT', 'ASAP']
const CONSULT_OPTIONS: ReferralConsultType[] = ['APP+LE', 'APP+UE', 'APP+SX', 'SX', 'OTHER']
const SOURCE_OPTIONS: ReferralSource[] = ['REGULAR', 'FRACTURE_CLINIC', 'OTHER']

// TODO: physician search
export function UpdateReferralEntryDialog({ isOpen, onClose, referralId, initialData, onSave }: UpdateReferralEntryDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  // --- Form States ---
  const [status, setStatus] = useState<ReferralStatus>('READY_TO_BOOK')
  const [urgency, setUrgency] = useState<ReferralUrgency>('ELECTIVE')
  const [source, setSource] = useState<ReferralSource>('REGULAR')
  const [consultType, setConsultType] = useState<ReferralConsultType>('OTHER')
  const [triageNote, setTriageNote] = useState("")
  const [referringPhysician, setReferringPhysician] = useState("")
  const [referralDate, setReferralDate] = useState("")
  const [emrPatientId, setEmrPatientId] = useState("")
  const [emrReferralDocID, setEmrReferralDocID] = useState("")
  const [adminReasonNote, setAdminReasonNote] = useState("")
  const [force, setForce] = useState(false)

  // Reused dynamic complaint array hook management
  const [complaints, setComplaints] = useState<FrontEndComplaint[]>([
    { bodyPart: 'KNEE', side: 'LEFT', details: '' }
  ])

  // Reset form with historical row metrics when dialog opens
  useEffect(() => {
    if (isOpen && initialData) {
      setStatus(initialData.status)
      setUrgency(initialData.urgency)
      setSource(initialData.source)
      setConsultType(initialData.consultType)
      setTriageNote(initialData.triageNote || "")
      setReferringPhysician(initialData.referringPhysician || "")
      setReferralDate(initialData.referralDate || "")
      setEmrPatientId(initialData.emrPatientId || "")
      setEmrReferralDocID(initialData.emrReferralDocID || "")
      setAdminReasonNote("")
      setForce(false)
      if (initialData.complaints && initialData.complaints.length > 0) {
        setComplaints(initialData.complaints)
      }
    }
  }, [isOpen, initialData])

  // --- Reused Dynamic Complaints Functions ---
  const addComplaint = () => {
    setComplaints([...complaints, { bodyPart: 'KNEE', side: 'LEFT', details: '' }])
  }

  const removeComplaint = (index: number) => {
    if (complaints.length > 1) {
      setComplaints(complaints.filter((_, idx) => idx !== index))
    }
  }

  const updateComplaint = (index: number, field: keyof FrontEndComplaint, value: string) => {
    const updated = complaints.map((item, idx) => {
      if (idx === index) {
        return { ...item, [field]: value }
      }
      return item
    })
    setComplaints(updated)
  }

  // --- Form Submission Handler ---
  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      const payload: AdminUpdatePayload = {
        id: referralId,
        status,
        urgency,
        source,
        consultType,
        triageNote,
        referringPhysician,
        referralDate,
        emrPatientId,
        emrReferralDocID,
        complaints,
        //note: adminReasonNote,
        force
      }
      await onSave(payload)
      onClose()
    } catch (err) {
      console.error("Submission step failed", err)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto border-red-200 bg-white text-slate-900 shadow-2xl">
        <DialogHeader className="border-b pb-3 border-slate-100">
          <DialogTitle className="text-xl font-bold text-red-700 flex items-center gap-2">
            <ShieldAlertIcon className="h-5 w-5 text-red-600" />
            Administrative Update
          </DialogTitle>
          <DialogDescription className="text-xs text-slate-500 font-normal mt-1">
            Force modifications directly into Referral tracking reference: <span className="font-mono font-bold bg-slate-100 p-0.5 rounded text-red-700">{referralId}</span>.
            This action commits directly to core database structures.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-5 my-2">

          {/* Section 1: Core EMR Structural IDs */}
          <div className="bg-slate-50 p-4 rounded-lg border border-slate-200/80 space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500">EMR Core Identifiers</h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">EMR Patient ID</label>
                <input
                  type="text"
                  className="w-full text-sm border rounded-md p-2 bg-white"
                  value={emrPatientId}
                  onChange={e => setEmrPatientId(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">EMR Referral Doc ID</label>
                <input
                  type="text"
                  className="w-full text-sm border rounded-md p-2 bg-white"
                  value={emrReferralDocID}
                  onChange={e => setEmrReferralDocID(e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Section 2: Clinical Metrics */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Referring Physician</label>
              <input
                type="text"
                className="w-full text-sm border rounded-md p-2 bg-white"
                value={referringPhysician}
                onChange={e => setReferringPhysician(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Consult Type</label>
              <select
                className="w-full text-sm border rounded-md p-2 bg-white"
                value={consultType}
                onChange={e => setConsultType(e.target.value as ReferralConsultType)}
              >
                {CONSULT_OPTIONS.map(opt => <option key={opt} value={opt}>{opt}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Referral Date</label>
              <input
                type="date"
                max="9999-12-31"
                className="w-full text-sm border rounded-md p-2 bg-white"
                value={referralDate}
                onChange={e => setReferralDate(e.target.value)}
              />
            </div>
          </div>

          {/* Section 3: Workflow Pipelines */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Pipeline Status</label>
              <select
                className="w-full text-sm border rounded-md p-2 bg-white text-slate-900"
                value={status}
                onChange={e => setStatus(e.target.value as ReferralStatus)}
              >
                {STATUS_OPTIONS.map(opt => <option key={opt} value={opt}>{opt.replace(/_/g, ' ')}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Urgency</label>
              <select
                className="w-full text-sm border rounded-md p-2 bg-white text-slate-900"
                value={urgency}
                onChange={e => setUrgency(e.target.value as ReferralUrgency)}
              >
                {URGENCY_OPTIONS.map(opt => <option key={opt} value={opt}>{opt}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-700 mb-1">Source</label>
              <select
                className="w-full text-sm border rounded-md p-2 bg-white text-slate-900"
                value={source}
                onChange={e => setSource(e.target.value as ReferralSource)}
              >
                {SOURCE_OPTIONS.map(opt => <option key={opt} value={opt}>{opt.replace(/_/g, ' ')}</option>)}
              </select>
            </div>
          </div>

          {/* Section 4: Dynamic Reused Complaint List Block */}
          <div className="space-y-3 border-t pt-4 border-slate-100">
            <div className="flex justify-between items-center">
              <label className="text-xs font-bold uppercase tracking-wider text-slate-500">Complaints Mapping</label>
              <button
                type="button"
                onClick={addComplaint}
                className="text-xs flex items-center gap-1 bg-slate-100 border border-slate-200 hover:bg-slate-200
                font-medium px-2 py-1 rounded-md text-slate-700 transition cursor-pointer"
              >
                <PlusIcon className="h-3 w-3" /> Add Part
              </button>
            </div>

            {complaints.map((c, index) => (
              <div key={index} className="space-y-2 bg-slate-50 p-3 rounded-md border border-slate-200">
                <div className="flex items-center gap-2">
                  <select
                    className="w-full border rounded-md p-2 bg-white text-slate-900 text-sm"
                    value={c.bodyPart}
                    onChange={e => updateComplaint(index, 'bodyPart', e.target.value)}
                  >
                    {BODY_PARTS.map(part => <option key={part} value={part}>{part}</option>)}
                  </select>

                  <select
                    className="w-full border rounded-md p-2 bg-white text-slate-900 text-sm"
                    value={c.side}
                    onChange={e => updateComplaint(index, 'side', e.target.value)}
                  >
                    {SIDES.map(side => <option key={side} value={side}>{side}</option>)}
                  </select>

                  {complaints.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeComplaint(index)}
                      className="p-2 text-slate-400 hover:text-red-600 rounded-md hover:bg-red-50 shrink-0 transition cursor-pointer"
                    >
                      <Trash2Icon className="h-4 w-4" />
                    </button>
                  )}
                </div>

                {c.bodyPart === 'OTHER' && (
                  <input
                    placeholder="Describe part (e.g., Femur)..."
                    className="w-full border rounded-md p-2 bg-white text-slate-900 text-sm focus:outline-none focus:ring-1 focus:ring-slate-400"
                    value={c.details}
                    onChange={e => updateComplaint(index, 'details', e.target.value)}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Section 5: Clinical Notes Override */}
          <div className="space-y-1">
            <label className="block text-xs font-medium text-slate-700">Triage Note</label>
            <textarea
              rows={2}
              className="w-full text-sm border rounded-md p-2 bg-white text-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-400"
              value={triageNote}
              onChange={e => setTriageNote(e.target.value)}
            />
          </div>

          {/*TODO: change this*/}
          {/* Section 6: Mandatory Audit Tracking & Force Checks */}
          <div className="bg-red-50/50 p-4 rounded-lg border border-red-200/60 space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-red-800">Administrative Log Accountability</h3>
            <div>
              <label className="block text-xs font-semibold text-red-900 mb-1">
                Reason for Update
              </label>
              <input
                type="text"
                // required
                placeholder="e.g., Corrected spelling error in referring doctor / sync mismatch"
                className="w-full text-sm border border-red-200 focus:border-red-500 rounded-md p-2 bg-white text-slate-900 focus:outline-none"
                value={adminReasonNote}
                onChange={e => setAdminReasonNote(e.target.value)}
              />
            </div>
            {/*
            <div className="flex items-center gap-2 pt-1">
              <input
                type="checkbox"
                id="force-override-chk"
                className="h-4 w-4 rounded border-slate-300 text-red-600 focus:ring-red-500 cursor-pointer"
                checked={force}
                onChange={e => setForce(e.target.checked)}
              />
              
              <label htmlFor="force-override-chk" className="text-xs font-medium text-slate-700 select-none cursor-pointer">
                Bypass pipeline guards and business rules entirely (`force = true`)
              </label>
              
            </div>
            */}
          </div>

          {/* Dialog Footer Operations */}
          <DialogFooter className="border-t pt-3 border-slate-100 mt-6">
            <button
              type="button"
              disabled={isSubmitting}
              onClick={onClose}
              className="px-4 py-2 border border-slate-200 rounded-md text-sm font-medium hover:bg-slate-50 transition disabled:opacity-50 text-slate-700 bg-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}//|| !adminReasonNote.trim()}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium cursor-pointer text-sm rounded-md shadow transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSubmitting ? "Committing Changes..." : "Apply Updates"}
            </button>
          </DialogFooter>

        </form>
      </DialogContent>
    </Dialog>
  )
}
