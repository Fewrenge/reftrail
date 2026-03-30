import { useState, useEffect } from "react";
import { PlusIcon, MoreVerticalIcon, Loader2Icon } from "lucide-react";
import { ROLES } from "@/helpers/constants";
import { Button } from "@/components/ui/button";
import SettingSection from "./SettingSection";
import SettingTable from "./SettingTable";

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
        <Button className="bg-primary text-primary-foreground hover:opacity-90 rounded-lg px-4 py-2 flex items-center gap-1 shadow-none border-none">
          <PlusIcon className="w-4 h-4" />
          <span>Create</span>
        </Button>
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
