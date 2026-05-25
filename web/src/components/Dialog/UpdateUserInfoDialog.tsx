import { useState, useEffect } from "react";
import { Loader2Icon } from "lucide-react";
//import { ROLES } from "@/helpers/constants";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";

//type RoleType = typeof ROLES[keyof typeof ROLES];

interface UpdateUserInfoDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    user: any;
    onSuccess: (updatedUser: any) => void;
}

const UpdateUserInfoDialog = ({ open, onOpenChange, user, onSuccess }: UpdateUserInfoDialogProps) => {
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<{
        userFirstName: string;
        userLastName: string;
        //role: RoleType; // <-- This lets it accept only valid role values
    }>({
        userFirstName: "",
        userLastName: "",
        //role: ROLES.BOOKING_TEAM,
    });

    // Sync form data whenever a new user row is selected
    useEffect(() => {
        if (user) {
            setFormData({
                userFirstName: user.userFirstName || "",
                userLastName: user.userLastName || "",
                //role: typeof user.role === "object" ? user.role?.name : user.role || ROLES.BOOKING_TEAM,
            });
        }
    }, [user]);

    const handleUpdate = async () => {
        if (!user) return;
        setLoading(true);

        try {
            // Clean REST architecture targets the specific user via path parameters
            const response = await fetch(`/api/v1/users/${user.username}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(formData),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || "Failed to update user profile");
            }

            alert(data.message || "User updated successfully");

            // Pass the updated user back to parent to refresh the grid layout dynamically
            onSuccess?.(data);
            onOpenChange(false);
        } catch (error: any) {
            alert(error.message || "An error occurred");
        } finally {
            setLoading(false);
        }
    };

    if (!user) return null;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-106.25">
                <DialogHeader>
                    <DialogTitle>Edit Member Profile</DialogTitle>
                    <DialogDescription>
                        Modify profile information for account: <span className="font-bold text-slate-900">@{user.username}</span>
                    </DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-4 py-4">
                    {/* FIELD 1: FIRST NAME */}
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">First Name</p>
                        <input
                            type="text"
                            value={formData.userFirstName}
                            onChange={(e) => setFormData((p) => ({ ...p, userFirstName: e.target.value }))}
                            className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                        />
                    </div>

                    {/* FIELD 2: LAST NAME */}
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Last Name</p>
                        <input
                            type="text"
                            value={formData.userLastName}
                            onChange={(e) => setFormData((p) => ({ ...p, userLastName: e.target.value }))}
                            className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                        />
                    </div>


                


                </div>

                <DialogFooter>
                    <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleUpdate} disabled={loading}>
                        {loading && <Loader2Icon className="w-4 h-4 mr-2 animate-spin" />}
                        Save Changes
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default UpdateUserInfoDialog;
