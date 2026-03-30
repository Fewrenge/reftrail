import { useState } from "react";
import { UserIcon, MoreVerticalIcon, PenLineIcon, CheckIcon, XIcon } from "lucide-react";
import { Button } from "@/components/ui/button"; // Use your UI button or a plain <button>
import { useAuth } from "@/contexts/AuthContext";

const ProfileSection = () => {
  const { user } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({ username: user?.username || "" });

  const toggleEdit = () => {
    setIsEditing(!isEditing);
    if (!isEditing) setFormData({ username: user?.username || "" });
  };

  const handleSave = async () => {
    // Backend logic goes here later
    setIsEditing(false);
  };

  return (
    <section className="w-full space-y-4">
      <div className="w-full flex flex-row justify-start items-center gap-4 p-4 border rounded-xl bg-card text-card-foreground">
        
        <div className="shrink-0 w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center">
          <UserIcon size={24} />
        </div>

        <div className="flex-1 min-w-0 flex flex-col justify-center items-start">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-tight">Username</p>
          {isEditing ? (
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              className="w-full text-lg font-semibold bg-transparent border-b border-primary focus:outline-none"
              autoFocus
            />
          ) : (
            <div className="flex items-center gap-2">
              <span className="text-lg font-semibold truncate">{user?.username || "Guest"}</span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {isEditing ? (
            <>
              <Button variant="ghost" size="sm" className="text-green-600 hover:bg-green-50" onClick={handleSave}>
                <CheckIcon className="w-4 h-4 mr-1" /> Save
              </Button>
              <Button variant="ghost" size="sm" onClick={toggleEdit}>
                <XIcon className="w-4 h-4" />
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" size="sm" onClick={toggleEdit}>
                <PenLineIcon className="w-4 h-4 mr-1.5" />
                Edit
              </Button>
              {/* Optional: Add a simple button for the "More" menu later */}
              <Button variant="ghost" size="sm">
                <MoreVerticalIcon className="w-4 h-4 text-muted-foreground" />
              </Button>
            </>
          )}
        </div>
      </div>
    </section>
  );
};

export default ProfileSection;
