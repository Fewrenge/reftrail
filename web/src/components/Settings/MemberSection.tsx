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
    password: string;
    userFirstName: string;
    userLastName: string;
  }>({
    username: "",
    role: ROLES.BOOKING_TEAM,
    password: '',
    userFirstName: '',
    userLastName: '',
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
        setFormData({ username: "", role: ROLES.BOOKING_TEAM, userFirstName: '', userLastName: '', password: '' });
      } else {
        console.error("Failed to create user");
      }
    } catch (error) {
      console.error("Error:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteUser = async (username: string) => {
  const confirmDelete = window.confirm(`Are you sure you want to delete user "${username}"?`);
  if (!confirmDelete) return;

  try {
    const response = await fetch(`/api/v1/users/${username}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        // Add Authorization header here if your Echo admin group uses JWT middleware:
        //"Authorization": `Bearer ${localStorage.getItem("auth_token")}`
      },
    });

    if (response.ok) {
      // Instantly remove the user from your local table state
      setMembers((prev) => prev.filter((member) => member.username !== username));
    } else {
      const errorData = await response.json().catch(() => ({}));
      alert(errorData.message || "Failed to delete user from server.");
    }
  } catch (error) {
    console.error("Error deleting user:", error);
    alert("A network error occurred while trying to delete the user.");
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
      render: (val: string) => <span className="font-medium">{val || "N/A"}</span>,
    },
    {
      key: "userFirstName",
      header: "First Name",
      className: "w-[25%]",
    },
    {
      key: "userLastName",
      header: "Last Name",
      className: "w-[25%]",
    },
    {
      key: "role",
      header: "Role",
      className: "w-[15%]",
      render: (val: any) => {
        const roleStr = typeof val === 'object' ? val?.name : val;
        const cleanRole = roleStr || "USER";
        const isAdmin = cleanRole === ROLES.SYSTEM_ADMIN;
        return (
          <span className={isAdmin ? "text-primary font-bold" : "text-muted-foreground"}>
            {cleanRole.replace(/_/g, ' ')}
          </span>
        );
      },
    },
      {
    key: "actions",
    header: "",
    className: "w-[10%] text-right",
    render: (_: any, row: any) => (
      <div className="flex items-center justify-end gap-1">
        {/* EDIT BUTTON */}
        <Button
          variant="ghost"
          size="sm"
         // onClick={() => handleOpenEditModal(row)} // Handled in your next step
          className="text-slate-400 hover:text-blue-600 rounded-lg h-8 w-8 p-0"
        >
          <span className="text-xs font-bold">Edit</span>
        </Button>

        {/* DELETE BUTTON */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => handleDeleteUser(row.username)} // Triggers our new delete function
          className="text-slate-400 hover:text-red-600 rounded-lg h-8 w-8 p-0"
          // Prevents self-deletion if currentUser state exists later
          disabled={row.username === "admin"} 
        >
          {/* If Trash2Icon isn't imported from lucide-react, you can use text or import it */}
          <span className="text-xs font-bold text-red-500">Delete</span>
        </Button>
      </div>
    ),
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

            {/*TODO: Implement form validation and error handling */}
            {/* Form */}
            <div className="py-4 space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm font-medium">First Name</label>
                  <input
                    className="border rounded-md p-2 bg-white text-slate-900"
                    value={formData.userFirstName || ''} // Fallback to empty string to keep input controlled
                    onChange={(e) => setFormData({ ...formData, userFirstName: e.target.value })}
                    placeholder="e.g. John"
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm font-medium">Last Name</label>
                  <input
                    className="border rounded-md p-2 bg-white text-slate-900"
                    value={formData.userLastName || ''}
                    onChange={(e) => setFormData({ ...formData, userLastName: e.target.value })}
                    placeholder="e.g. Doe"
                  />
                </div>
              </div>

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
                <label className="text-sm font-medium">Temporary Password</label>
                <input
                  type="password"
                  className="border rounded-md p-2 bg-white text-slate-900"
                  value={formData.password || ''}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Enter temp password"
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
          getRowKey={(row) => row.username}
        />
      </div>
    </SettingSection>
  );
};

export default MemberSection;
