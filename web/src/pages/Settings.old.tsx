import React, { useState, useEffect } from 'react';

interface User {
  username: string;
  role: string;
  id?: string;
}

interface SettingsProps {
  user: User | null;
  onLogout: () => void;
}



export default function Settings({ user, onLogout }: SettingsProps) {
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [message, setMessage] = useState({ text: '', isError: false });
  const [isUpdating, setIsUpdating] = useState(false);

  if (!user) return <div className="p-8 text-slate-400 italic">Loading 
  ...</div>;

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsUpdating(true);
    setMessage({ text: '', isError: false });
        try {
      const res = await fetch('/api/v1/users/password', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ oldPassword, newPassword }),
      });

      if (res.ok) {
        setMessage({ text: 'Password updated successfully!', isError: false });
        setOldPassword(''); // Clear the form
        setNewPassword('');
      } else {
        const errorText = await res.text();
        setMessage({ text: `Error: ${errorText || 'Failed to update'}`, isError: true });
      }
    } catch (err) {
      setMessage({ text: 'Connection error. Is the server running?', isError: true });
    } finally {
      setIsUpdating(false);
    }
  };


  return (
    <div className="max-w-2xl mx-auto mt-10 space-y-6">
      {/* USER INFO CARD (Existing) */}
      <div className="bg-white rounded-2xl shadow-sm border border-slate-200 p-8">
        <div className="flex items-center gap-6 mb-8">
          <div className="w-20 h-20 bg-blue-600 rounded-full flex items-center justify-center text-3xl font-bold text-white shadow-lg">
            {user.username.charAt(0).toUpperCase()}
          </div>
          <div>
            <h1 className="text-2xl font-bold text-slate-900">{user.username}</h1>
            <p className="text-slate-500 capitalize">{user.role} Account</p>
          </div>
        </div>

        {/* THE PASSWORD CHANGE SECTION */}
        <div className="border-t border-slate-100 pt-8">
          <h3 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-6">Security & Password</h3>
          
          <form onSubmit={handleChangePassword} className="space-y-4 max-w-sm">
            <div>
              <label className="block text-xs font-bold text-slate-500 mb-1">Current Password</label>
              <input 
                required
                type="password" 
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-500 mb-1">New Password</label>
              <input 
                required
                type="password" 
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              />
            </div>

            <button 
              type="submit"
              disabled={isUpdating}
              className={`w-full py-2.5 rounded-xl font-bold text-sm transition-all shadow-sm
                ${isUpdating ? 'bg-slate-100 text-slate-400' : 'bg-slate-900 text-white hover:bg-slate-800'}`}
            >
              {isUpdating ? 'Updating...' : 'Update Password'}
            </button>

            {message.text && (
              <p className={`text-xs font-medium mt-2 ${message.isError ? 'text-red-500' : 'text-emerald-600'}`}>
                {message.text}
              </p>
            )}
          </form>
        </div>

        {/* LOGOUT BUTTON */}
        <div className="mt-12 border-t border-slate-100 pt-6">
          <button 
            onClick={onLogout}
            className="text-sm font-bold text-red-500 hover:text-red-600 transition-colors"
          >
            Sign out of all sessions
          </button>
        </div>
      </div>
    </div>
  );
}
