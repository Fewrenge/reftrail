import { useState, useEffect } from "react";
import { UsersIcon, UserPlusIcon, Trash2Icon, ShieldCheckIcon, Loader2Icon } from "lucide-react";
import { ROLES } from "@/helpers/constants";
import { Button } from "@/components/ui";

type UserRole = typeof ROLES[keyof typeof ROLES];

interface Member {
  id: number;
  username: string;
  role: UserRole;
}





const MemberSection = () => {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAdding, setIsAdding] = useState(false); // Controls the "Add Member" popup
  const [deleteTarget, setDeleteTarget] = useState<Member | null>(null); // Stores WHO we are deleting


  // 1. Fetch all members from Go backend
  useEffect(() => {
    fetch("/api/v1/users", { credentials: "same-origin" })
      .then((res) => res.json())
      .then((data) => setMembers(Array.isArray(data) ? data : []))
      .catch((err) => console.error("Failed to fetch members:", err))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex justify-center p-10"><Loader2Icon className="animate-spin text-blue-600" /></div>;

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-slate-800 flex items-center gap-2">
          <UsersIcon size={20} className="text-blue-600" />
          System Members ({members.length})
        </h3>
        <Button onClick={() => setIsAdding(true)}>
          <UserPlusIcon size={16} className="mr-2" />
          Add Member
        </Button>
      </div>

      <div className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
        <table className="w-full text-left border-collapse">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              <th className="px-6 py-3 text-xs font-bold uppercase tracking-wider text-slate-500">Member</th>
              <th className="px-6 py-3 text-xs font-bold uppercase tracking-wider text-slate-500">Role</th>
              <th className="px-6 py-3 text-right text-xs font-bold uppercase tracking-wider text-slate-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {members.map((member) => (
              <tr key={member.id} className="hover:bg-slate-50 transition-colors">
                <td className="px-6 py-4 flex items-center gap-3">
                  <div className="w-8 h-8 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center font-bold text-xs">
                    {member.username.charAt(0).toUpperCase()}
                  </div>
                  <span className="font-medium text-slate-700">{member.username}</span>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-tighter ${member.role === ROLES.SYSTEM_ADMIN ? "bg-purple-100 text-purple-700" : "bg-slate-100 text-slate-600"
                    }`}>
                    {member.role === ROLES.SYSTEM_ADMIN && <ShieldCheckIcon size={10} />}
                    {member.role}
                  </span>
                </td>
                <td className="px-6 py-4 text-right">
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => setDeleteTarget(member)}
                  >
                    <Trash2Icon size={16} />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default MemberSection;
