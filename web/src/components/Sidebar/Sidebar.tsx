import { Link, NavLink, useNavigate } from 'react-router-dom';
import { 
  HospitalIcon, 
  ScrollTextIcon, 
  ChartNoAxesCombinedIcon, 
  SettingsIcon, 
  LogOutIcon 
} from "lucide-react";
import { useAuth } from '../../contexts/AuthContext';
import { Button } from "@/components/ui/button"; // Ensure this matches your export path
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown";

export default function Sidebar() {
  const { user, onLogout } = useAuth();
  const navigate = useNavigate();

  return (
    <aside className="w-64 bg-white border-r border-slate-200 sticky top-0 h-screen p-6 hidden md:flex flex-col">
      {/* LOGO SECTION */}
      <Link to="/" className="flex items-center gap-3 mb-10 hover:opacity-80 transition-opacity cursor-pointer">
        <HospitalIcon size={30} strokeWidth={2.5} className="text-blue-600" />
        <h1 className="font-bold text-lg tracking-tight text-slate-800">RefTrail</h1>
      </Link>

      {/* NAVIGATION SECTION */}
      <nav className="space-y-1 flex-1">
        <NavLink
          to="/referrals"
          end
          className={({ isActive }) =>
            `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${
              isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
            }`
          }
        >
          <ScrollTextIcon size={20} strokeWidth={2.5} />
          <span>Referrals</span>
        </NavLink>

        <NavLink
          to="/analytics"
          className={({ isActive }) =>
            `flex items-center gap-3 px-3 py-2.5 rounded-xl font-medium transition-all ${
              isActive ? "bg-blue-50 text-blue-700" : "text-slate-500 hover:bg-slate-50"
            }`
          }
        >
          <ChartNoAxesCombinedIcon size={20} strokeWidth={2.5} />
          <span>Analytics</span>
        </NavLink>
      </nav>

      {/* USER MENU SECTION */}
      <div className="pt-6 border-t border-slate-100 mt-auto">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className="w-full justify-start text-left gap-3 h-auto p-2 rounded-xl transition-all data-[state=open]:bg-slate-100"
            >
              <div className="w-9 h-9 bg-linear-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center text-white font-bold shadow-md border-2 border-white shrink-0">
                {user?.username?.charAt(0).toUpperCase() || '?'}
              </div>
              <div className="flex flex-col overflow-hidden">
                <span className="text-sm font-semibold text-slate-700 truncate">
                  {user?.username || 'Guest'}
                </span>
                <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold">
                  {user?.role || 'User'}
                </span>
              </div>
            </Button>
          </DropdownMenuTrigger>

          <DropdownMenuContent
            side="top"
            align="start"
            sideOffset={8}
            className="w-(--radix-dropdown-menu-trigger-width)"
          >
            <DropdownMenuItem onSelect={() => navigate('/settings')}>
              <SettingsIcon size={16} className="mr-2" />
              <span>Settings</span>
            </DropdownMenuItem>


            <DropdownMenuItem
              variant="destructive"
              onSelect={() => onLogout()}
            >
              <LogOutIcon size={16} className="mr-2" />
              <span>Sign Out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </aside>
  );
}
