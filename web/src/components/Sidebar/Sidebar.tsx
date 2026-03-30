import { useState, useRef, useEffect } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { HospitalIcon, ScrollTextIcon, ChartNoAxesCombinedIcon, SettingsIcon, LogOutIcon } from "lucide-react";
import { useAuth } from '../../contexts/AuthContext';
import { Button } from "@/components/ui";

export default function Sidebar() {
  const { user, onLogout } = useAuth();
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <aside className="w-64 bg-white border-r border-slate-200 sticky top-0 h-screen p-6 hidden md:flex flex-col">
      <div className="flex items-center gap-3 mb-10">
        <HospitalIcon size={30} strokeWidth={2.5} className="text-blue-600" />
        <h1 className="font-bold text-lg tracking-tight text-slate-800">Medical Portal</h1>
      </div>

      <nav className="space-y-1 flex-1">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
            }`
          }
        >
          <ScrollTextIcon size={20} strokeWidth={2.5} />
          <span>Waitlist</span>
        </NavLink>

        <NavLink
          to="/analytics"
          className={({ isActive }) =>
            `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
            }`
          }
        >
          <ChartNoAxesCombinedIcon size={20} strokeWidth={2.5} />
          <span>Analytics</span>
        </NavLink>
      </nav>

      {/* POPUP MENU SECTION */}
      <div className="pt-6 border-t border-slate-100 mt-auto relative" ref={menuRef}>
        {isMenuOpen && (
          <div className="absolute bottom-full left-0 w-full mb-2 bg-white border border-slate-200 rounded-xl shadow-xl py-2 z-50 animate-in fade-in slide-in-from-bottom-2">
            <Button
              variant="ghost"
              className="w-full justify-start gap-3 p-2 h-auto"
              onClick={() => {
                navigate('/settings'); // 1. Go to settings
                setIsMenuOpen(false);  // 2. Close the menu
              }}
            >
              <SettingsIcon size={16} /> Settings
            </Button>

            <Button
              variant="ghost"
              className="w-full justify-start gap-3 p-2 h-auto"
              onClick={() => {
                onLogout();           // 1. Call the Go logout API
                setIsMenuOpen(false); // 2. Close the menu
              }}
            >
              <LogOutIcon size={16} /> Sign Out
            </Button>
          </div>
        )}

        <Button
          variant="ghost"
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className={`w-full justify-start text-left gap-3 h-auto p-2 rounded-xl transition-all ${isMenuOpen ? "bg-slate-100" : ""}`}
        >
          <div className="w-9 h-9 bg-linear-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center text-white font-bold shadow-md border-2 border-white shrink-0">
            {user?.username?.charAt(0).toUpperCase() || '?'}
          </div>
          <div className="flex flex-col overflow-hidden">
            <span className="text-sm font-semibold text-slate-700 truncate">{user?.username || 'Guest'}</span>
            <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold">{user?.role || 'User'}</span>
          </div>
        </Button>
      </div>
    </aside>
  );
}
