import React from 'react';

interface User {
  username: string;
  role: string;
  id?: string;
}

interface ProfileProps {
  user: User | null;
  onLogout: () => void;
}

export default function Profile({ user, onLogout }: ProfileProps) {
  if (!user) return <div className="p-8 text-slate-400 italic">Loading profile...</div>;

  return (
    <div className="max-w-2xl mx-auto mt-10">
      <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
        {/* Header/Cover */}
        <div className="h-24 bg-gradient-to-r from-blue-500 to-blue-600"></div>
        
        <div className="px-8 pb-8">
          {/* Avatar overlap */}
          <div className="-mt-12 mb-6">
            <div className="w-24 h-24 bg-white rounded-full p-1 shadow-md">
              <div className="w-full h-full bg-slate-100 rounded-full flex items-center justify-center text-3xl font-bold text-blue-600">
                {user.username.charAt(0).toUpperCase()}
              </div>
            </div>
          </div>

          <h1 className="text-2xl font-bold text-slate-900">{user.username}</h1>
          <p className="text-slate-500 mb-6 capitalize">{user.role} Account</p>

          <div className="space-y-4 border-t border-slate-100 pt-6">
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Status</span>
              <span className="text-green-600 font-medium">● Active</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">System Permissions</span>
              <span className="text-slate-700 font-mono text-xs">{user.role === 'admin' ? 'Full Access' : 'Read/Write'}</span>
            </div>
          </div>

          <button 
            onClick={onLogout}
            className="w-full mt-10 bg-red-50 text-red-600 py-3 rounded-xl font-semibold hover:bg-red-100 transition-colors"
          >
            Sign Out of Portal
          </button>
        </div>
      </div>
    </div>
  );
}
