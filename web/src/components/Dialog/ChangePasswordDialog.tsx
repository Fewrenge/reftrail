import { useState } from "react";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";

const ChangePasswordDialog = ({ open, onOpenChange, user, onSuccess }: any) => {
    const [showOldPassword, setShowOldPassword] = useState(false);
    const [showNewPasswords, setShowNewPasswords] = useState(false);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        oldPassword: "",
        newPassword: "",
        confirmPassword: "",
    });

    const handleUpdate = async () => {
        if (!user) return;
        setLoading(true);

        try {
            const response = await fetch("/api/v1/users/password", {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    oldPassword: formData.oldPassword,
                    newPassword: formData.newPassword,
                }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || "Failed to update password");
            }

            alert(data.message);

            onSuccess?.(); // Trigger the success alert/toast
            onOpenChange(false);

            // Clear the form for security
            setFormData({ oldPassword: "", newPassword: "", confirmPassword: "" });
        } catch (error: any) {
            alert(error.message); // Replace with a toast later
        } finally {
            setLoading(false);
        }
    };

    if (!user) return null;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-106.25">
                <DialogHeader>
                    <DialogTitle>Update Password</DialogTitle>
                    <DialogDescription>
                        Change password for <span className="font-bold text-slate-900">{user.username}</span>
                    </DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-4 py-4">
                    {/* FIELD 1: CURRENT PASSWORD */}
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Current Password</p>
                        <div className="relative">
                            <input
                                type={showOldPassword ? "text" : "password"}
                                value={formData.oldPassword}
                                onChange={(e) => setFormData(p => ({ ...p, oldPassword: e.target.value }))}
                                className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                            />
                            <button type="button" onClick={() => setShowOldPassword(!showOldPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                                {showOldPassword ? <EyeOffIcon size={16} /> : <EyeIcon size={16} />}
                            </button>
                        </div>
                    </div>

                    {/* FIELD 2: NEW PASSWORD */}
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">New Password</p>
                        <div className="relative">
                            <input
                                type={showNewPasswords ? "text" : "password"}
                                value={formData.newPassword}
                                onChange={(e) => setFormData(p => ({ ...p, newPassword: e.target.value }))}
                                className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                            />
                            <button
                                type="button"
                                onClick={() => setShowNewPasswords(!showNewPasswords)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                            >
                                {showNewPasswords ? <EyeOffIcon size={16} /> : <EyeIcon size={16} />}
                            </button>
                        </div>
                    </div>

                    {/* FIELD 3: REPEAT PASSWORD */}
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Confirm New Password</p>
                        <div className="relative">
                            <input
                                type={showNewPasswords ? "text" : "password"}
                                value={formData.confirmPassword}
                                onChange={(e) => setFormData(p => ({ ...p, confirmPassword: e.target.value }))}
                                className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                            />
                        </div>
                    </div>

                    {formData.confirmPassword && formData.newPassword !== formData.confirmPassword && (
                        <p className="text-[10px] text-destructive font-medium -mt-2">Passwords do not match</p>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button
                        onClick={handleUpdate}
                        disabled={loading || !formData.oldPassword || !formData.newPassword || formData.newPassword !== formData.confirmPassword}
                    >
                        {loading ? "Updating..." : "Update Password"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default ChangePasswordDialog;
