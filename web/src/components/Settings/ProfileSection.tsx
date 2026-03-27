import { useState } from "react";
import { User, Edit2, X, Check } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";

const ProfileSection = () => {
  const { user } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({ username: user?.username || "" });

  const toggleEdit = () => {
    setIsEditing(!isEditing);
    if (!isEditing) setFormData({ username: user?.username || "" }); // Reset on cancel
  };

  const handleSave = async () => {
    // ... insert your fetch('/api/v1/users/me', { method: 'PUT' ... }) logic here
    setIsEditing(false);
  };

  return (
    <div className="max-w-md space-y-4">
      <div className="flex items-center justify-between p-4 bg-white border border-slate-200 rounded-2xl shadow-sm">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-blue-50 text-blue-600 rounded-xl">
            <User size={24} />
          </div>
          <div>
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400">Username</p>
            {isEditing ? (
              <input
                type="text"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                className="mt-1 font-semibold text-slate-800 border-b border-blue-500 focus:outline-none"
                autoFocus
              />
            ) : (
              <p className="text-lg font-semibold text-slate-800">{user?.username}</p>
            )}
          </div>
        </div>

        <div className="flex gap-2">
          {isEditing ? (
            <>
              <button onClick={handleSave} className="p-2 text-green-600 hover:bg-green-50 rounded-lg cursor-pointer transition-colors">
                <Check size={20} />
              </button>
              <button onClick={toggleEdit} className="p-2 text-slate-400 hover:bg-slate-50 rounded-lg cursor-pointer transition-colors">
                <X size={20} />
              </button>
            </>
          ) : (
            <button onClick={toggleEdit} className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg cursor-pointer transition-colors">
              <Edit2 size={20} />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default ProfileSection;
