import { useState, useEffect } from "react";
import { PlusIcon, Loader2Icon, ArchiveIcon, Trash2Icon, EditIcon } from "lucide-react";
import { UserRole } from "@/types/users";
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

import UpdateUserInfoDialog from "@/components/Dialog/UpdateUserInfoDialog";

interface Member {
  username: string;
  role: string;
  nickname?: string;
  email?: string;
}

const MemberSection = () => {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateUserDialogOpen, setIsCreateUserDialogOpen] = useState(false);
  const [formData, setFormData] = useState<{
    username: string;
    role: string;
    password: string;
    userFirstName: string;
    userLastName: string;
  }>({
    username: "",
    role: UserRole.BOOKING_TEAM,
    password: '',
    userFirstName: '',
    userLastName: '',
  });
  const [submittingCreateUser, setSubmittingCreateUser] = useState(false);
  const [isUpdateUserInfoDialogOpen, setIsUpdateUserInfoDialogOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<Member | null>(null);

  const handleOpenEditModal = (user: Member) => {
    setEditingUser(user);
    setIsUpdateUserInfoDialogOpen(true);
  };


  const handleCreateUser = async () => {
    setSubmittingCreateUser(true);
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
        setIsCreateUserDialogOpen(false);
        setFormData({ username: "", role: UserRole.BOOKING_TEAM, userFirstName: '', userLastName: '', password: '' });
      } else {
        console.error("Failed to create user");
      }
    } catch (error) {
      console.error("Error:", error);
    } finally {
      setSubmittingCreateUser(false);
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

  const handleArchiveUser = async (username: string) => {
    const confirmArchive = window.confirm(
      `Are you sure you want to archive user "${username}"? They will lose all system login access.`
    );
    if (!confirmArchive) return;

    try {
      const response = await fetch(`/api/v1/users/${username}/archive`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
      });

      if (response.ok) {
        // Instantly filter out the archived user from your UI grid view array
        setMembers((prev) => prev.filter((member) => member.username !== username));
      } else {
        const errorData = await response.json().catch(() => ({}));
        alert(errorData.error || "Server failed to archive user.");
      }
    } catch (error) {
      console.error("Archive transaction error:", error);
      alert("A network error occurred while processing the request.");
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
        const isAdmin = cleanRole === UserRole.REFTRAIL_ADMIN;
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
          <button
            type="button"
            onClick={() => handleOpenEditModal(row)}
            className="flex items-center justify-center h-8 w-8 text-slate-400 hover:text-blue-600
             hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
            title="Edit"
          >
            <EditIcon size={18} />
          </button>

          {/* ARCHIVE BUTTON */}
          <button
            type="button"
            onClick={() => handleArchiveUser(row.username)}
            disabled={row.username === "admin"}
            className="flex items-center justify-center h-8 w-8 text-amber-600 hover:text-amber-700
             hover:bg-amber-50 disabled:opacity-40 disabled:hover:bg-transparent disabled:cursor-not-allowed rounded-lg transition-colors cursor-pointer"
            title="Archive"
          >
            <ArchiveIcon size={18} />
          </button>

          {/* DELETE BUTTON */}
          <button
            type="button"
            onClick={() => handleDeleteUser(row.username)}
            disabled={row.username === "admin"}
            className="flex items-center justify-center h-8 w-8 text-red-500 hover:text-red-600
             hover:bg-red-50 disabled:opacity-40 disabled:hover:bg-transparent disabled:cursor-not-allowed rounded-lg transition-colors cursor-pointer"
            title="Delete"
          >
            <Trash2Icon size={18} />
          </button>
        </div>
      )

    },

  ];

  if (loading) return <div className="p-10 flex justify-center"><Loader2Icon className="animate-spin opacity-20" /></div>;

  return (
    <>
      <SettingSection
        title="Member List"
        className="p-1"
        actions={
          <Dialog open={isCreateUserDialogOpen} onOpenChange={setIsCreateUserDialogOpen}>
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
                    <option value={UserRole.BOOKING_TEAM}>{UserRole.BOOKING_TEAM}</option>
                    <option value={UserRole.REFTRAIL_ADMIN}>{UserRole.REFTRAIL_ADMIN}</option>
                  </select>
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateUserDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateUser}
                  disabled={submittingCreateUser || !formData.username.trim()}
                >
                  {submittingCreateUser ? "Creating..." : "Confirm"}
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

      <UpdateUserInfoDialog
        open={isUpdateUserInfoDialogOpen}
        onOpenChange={setIsUpdateUserInfoDialogOpen}
        user={editingUser}
        onSuccess={(updatedUser) => {
          // Automatically injects backend updates directly back into your screen grid row state
          setMembers((prev) =>
            prev.map((m) => (m.username === updatedUser.username ? { ...m, ...updatedUser } : m))
          );
        }}
      />
    </>

  );
};

export default MemberSection;
