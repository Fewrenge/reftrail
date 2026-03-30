import { useState, useEffect } from "react";
import { PlusIcon, Loader2Icon } from "lucide-react";
import { ROLES } from "@/helpers/constants";
import { Button } from "@/components/ui/button";
import SettingSection from "./SettingSection";
import SettingTable from "./SettingTable";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogDescription,
} from "@/components/ui/dialog";

interface Member {
  id: number;
  username: string;
  role: string;
  nickname?: string;
  email?: string;
}

const MemberSection = () => {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [formData, setFormData] = useState<{
    username: string;
    role: string;
  }>({
    username: "",
    role: ROLES.SYSTEM_ADMIN,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);


  const handleCreateUser = async () => {
    setIsSubmitting(true);
    try {
      const response = await fetch("/api/v1/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        const newUser = await response.json();
        // Update the local list so the new user appears immediately
        setMembers((prev) => [...prev, newUser]);
        // Close the dialog and reset form
        setIsCreateDialogOpen(false);
        setFormData({ username: "", role: ROLES.SYSTEM_ADMIN });
      } else {
        console.error("Failed to create user");
      }
    } catch (error) {
      console.error("Error:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    fetch("/api/v1/users")
      .then((res) => res.json())
      .then((data) => setMembers(Array.isArray(data) ? data : []))
      .finally(() => setLoading(false));
  }, []);

  const columns = [
    {
      key: "username",
      header: "Username",
      className: "w-[25%]",
      render: (val: string) => <span className="font-medium">{val}</span>,
    },
    {
      key: "role",
      header: "Role",
      className: "w-[15%]",
      render: (val: string) => {
        const isAdmin = val === ROLES.SYSTEM_ADMIN;

        return (
          <span className={isAdmin ? "text-primary font-bold" : "text-muted-foreground"}>
            {val}
          </span>
        );
      },
    },
  ];

  if (loading) return <div className="p-10 flex justify-center"><Loader2Icon className="animate-spin opacity-20" /></div>;

  return (
    <SettingSection
      title="Member list"
      className="p-1"
      actions={
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-primary text-primary-foreground rounded-lg px-3 py-2 flex items-center gap-1 shadow-none border-none">
              <PlusIcon className="w-4 h-4" />
              <span>Create</span>
            </Button>
          </DialogTrigger>

          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add New Member</DialogTitle>
              <DialogDescription>
                Fill in the details below to add a new member to your team.
              </DialogDescription>
            </DialogHeader>

            {/* Form */}
            <div className="py-4 space-y-4">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium">Username</label>
                <input
                  className="border rounded-md p-2 bg-white text-slate-900"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="Enter username"
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium">Role</label>
                <select
                  className="border rounded-md p-2 bg-white"
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                >
                  <option value={ROLES.BOOKING_TEAM}>{ROLES.BOOKING_TEAM}</option>
                  <option value={ROLES.SYSTEM_ADMIN}>{ROLES.SYSTEM_ADMIN}</option>
                </select>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreateUser}
                disabled={isSubmitting || !formData.username.trim()}
              >
                {isSubmitting ? "Creating..." : "Confirm"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      }
    >
      <div className="mt-4">
        <SettingTable
          columns={columns}
          data={members}
          className="border-border bg-transparent" // Use the beige border
          getRowKey={(row) => row.id.toString()}
        />
      </div>
    </SettingSection>
  );
};

export default MemberSection;
