import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { AlertCircleIcon } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";

interface AddReferralPhysicianDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

export const AddReferralPhysicianDialog: React.FC<AddReferralPhysicianDialogProps> = ({
    isOpen,
    onClose,
    onSuccess,
}) => {
    // Form Field States
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [cpsoNumber, setCPSOnumber] = useState("");
    const [emrPhysicianId, setEmrPhysicianId] = useState("");

    // UI State Managers
    const [submitting, setSubmitting] = useState(false);
    const [errorMsg, setErrorMsg] = useState<string | null>(null);

    // Reset form contents completely upon exit
    const handleResetAndClose = () => {
        setFirstName("");
        setLastName("");
        setCPSOnumber("");
        setEmrPhysicianId("");
        setErrorMsg(null);
        onClose();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        // Front-end presence validation checking
        if (!firstName.trim() || !lastName.trim()) {
            setErrorMsg("First name and Last name are strictly required fields.");
            return;
        }

        setSubmitting(true);
        setErrorMsg(null);

        try {
            // Structure fields matching your Go ReferralPhysician json payload requirements
            const payload = {
                firstName: firstName.trim(),
                lastName: lastName.trim(),
                cpsoNumber: cpsoNumber.trim() || null, // Convert blank space variants to null pointers
                emrPhysicianId: emrPhysicianId.trim() || null,
            };

            const response = await fetch("/api/v1/physicians", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(payload),
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || "Failed to establish database profile record.");
            }

            // Refresh directory and clear overlay modal container states on absolute success
            onSuccess();
            handleResetAndClose();
        } catch (err: any) {
            console.error("Creation endpoint pipeline exception:", err);
            setErrorMsg(err.message || "An unexpected error occurred while saving the profile.");
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && handleResetAndClose()}>
            <DialogContent className="sm:max-w-120 bg-white rounded-2xl p-6 border border-slate-100 shadow-xl">
                <DialogHeader>
                    <DialogTitle className="text-xl font-bold text-slate-900 tracking-tight">
                        Add New Referral Physician
                    </DialogTitle>
                </DialogHeader>

                <form onSubmit={handleSubmit} className="space-y-5 mt-4">
                    {/* Error Banner Notification Alert */}
                    {errorMsg && (
                        <div className="flex items-start gap-3 bg-red-50 text-red-700 text-sm p-3.5 rounded-xl border border-red-100">
                            <AlertCircleIcon size={18} className="shrink-0 mt-0.5" />
                            <p className="font-medium leading-relaxed">{errorMsg}</p>
                        </div>
                    )}

                    {/* First & Last Name Grid */}
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-1.5">
                            <label className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                                First Name <span className="text-red-500">*</span>
                            </label>
                            <input
                                type="text"
                                required
                                disabled={submitting}
                                placeholder="e.g. John"
                                value={firstName}
                                onChange={(e) => setFirstName(e.target.value)}
                                className="w-full h-11 bg-slate-50 border border-slate-200 rounded-xl px-3.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20
                 focus:border-blue-500 transition-all text-slate-900 placeholder:text-slate-400 disabled:opacity-50"
                            />
                        </div>

                        <div className="space-y-1.5">
                            <label className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                                Last Name <span className="text-red-500">*</span>
                            </label>
                            <input
                                type="text"
                                required
                                disabled={submitting}
                                placeholder="e.g. Smith"
                                value={lastName}
                                onChange={(e) => setLastName(e.target.value)}
                                className="w-full h-11 bg-slate-50 border border-slate-200 rounded-xl px-3.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20
                 focus:border-blue-500 transition-all text-slate-900 placeholder:text-slate-400 disabled:opacity-50"
                            />
                        </div>
                    </div>

                    {/* CPSO Number Input */}
                    <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                            CPSO Number
                        </label>
                        <input
                            type="text"
                            disabled={submitting}
                            placeholder="e.g. 12345 (Optional)"
                            value={cpsoNumber}
                            onChange={(e) => setCPSOnumber(e.target.value)}
                            className="w-full h-11 bg-slate-50 border border-slate-200 rounded-xl px-3.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20
               focus:border-blue-500 transition-all text-slate-900 placeholder:text-slate-400 disabled:opacity-50"
                        />
                    </div>

                    {/* EMR Physician ID Input */}
                    <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                            EMR Physician ID
                        </label>
                        <input
                            type="text"
                            disabled={submitting}
                            placeholder="e.g. EMR-9982 (Optional)"
                            value={emrPhysicianId}
                            onChange={(e) => setEmrPhysicianId(e.target.value)}
                            className="w-full h-11 bg-slate-50 border border-slate-200 rounded-xl px-3.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20
               focus:border-blue-500 transition-all text-slate-900 placeholder:text-slate-400 disabled:opacity-50"
                        />
                    </div>

                    {/* Action Execution Footer Layout Panel */}
                    <DialogFooter className="pt-4 border-t border-slate-100 flex items-center justify-end gap-3 sm:space-x-0">
                        <Button
                            type="button"
                            variant="ghost"
                            disabled={submitting}
                            onClick={handleResetAndClose}
                            className="h-11 rounded-xl px-5 text-slate-500 font-medium hover:bg-slate-50 disabled:opacity-50"
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            disabled={submitting}
                            className="h-11 bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-6 text-sm font-semibold 
              shadow-sm transition-colors disabled:opacity-50 flex items-center justify-center min-w-30"
                        >
                            {submitting ? "Saving..." : "Save Profile"}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
};

export default AddReferralPhysicianDialog;