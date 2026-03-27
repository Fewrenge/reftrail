import { useState, useMemo, useEffect } from "react";
import { UserIcon, LogInIcon, UsersIcon } from "lucide-react";
import { useLocation } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { ROLES } from "../helpers/constants";
import ProfileSection from "../components/Settings/ProfileSection";
import MemberSection from "../components/Settings/MemberSection";

// Define sections
type SettingSection = "profile" | "login" | "member";

const BASIC_SECTIONS: SettingSection[]=["profile", "login"];
const ADMIN_SECTIONS: SettingSection[]=["member"];

// Map names to Icons
const SECTION_ICON_MAP: Record<SettingSection, any> = {
  profile: UserIcon,
  login: LogInIcon,
  member: UsersIcon,
};

// Sub-pages that swap out in the middle of the screen
const SECTION_COMPONENT_MAP: Record<SettingSection, React.ComponentType> = {
  profile: ProfileSection, // to be changed to ProfileSection.tsx, same with the following
  login: () => <div className="p-4">Password & Security settings</div>,
  member: MemberSection,
};


const Setting = () => {

  // Check what's following the hashtag in the URL
  const location = useLocation();
  const { user, loading } = useAuth();
  const [selectedSection, setSelectedSection] = useState<SettingSection>("profile");

  // Figure out if the user is an Admin
  const isHost = user?.role === ROLES.SYSTEM_ADMIN;

  // Create the list of menu items the user is allowed to see
  const settingsSectionList = useMemo(() => {
    return isHost ? [...BASIC_SECTIONS, ...ADMIN_SECTIONS] : BASIC_SECTIONS;
  }, [isHost]);

  

  // SYNC: If the URL changes (e.g. someone clicks a link to #login), update the state
  useEffect(() => {
    const hash = location.hash.slice(1) as SettingSection; // slice(1) removes the '#'
    // Fall-back for an invalid hash location. Kicks the user back to #profile page otherwise
    const nextSection = settingsSectionList.includes(hash) ? hash : "profile";
    setSelectedSection(nextSection);
  }, [location.hash, settingsSectionList]);

  // If the app is still talking to the Go backend, show a simple loading message
  if (loading) return <div className="p-10">Loading settings...</div>;

  const ActiveSection = SECTION_COMPONENT_MAP[selectedSection];

    return (
    <div className="flex flex-row w-full max-w-5xl mx-auto p-6 gap-6">
      {/* --- SIDEBAR --- */}
      <aside className="w-48 flex flex-col gap-2">
        <h2 className="text-xl font-bold mb-4">Settings</h2>
        {settingsSectionList.map((sectionId) => {
          const Icon = SECTION_ICON_MAP[sectionId];
          const isActive = selectedSection === sectionId;
          
          return (
            <button
              key={sectionId}
              onClick={() => window.location.hash = sectionId} // Update the URL hash
              className={`flex items-center gap-2 p-2 rounded-lg cursor-pointer transition-colors ${
                isActive ? "bg-blue-100 text-blue-700" : "hover:bg-gray-100"
              }`}
            >
              <Icon size={18} />
              <span className="capitalize">{sectionId}</span>
            </button>
          );
        })}
      </aside>

      {/* --- MAIN CONTENT --- */}
      <main className="flex-1 bg-white shadow rounded-xl p-6">
        <h3 className="text-lg font-medium mb-4 capitalize">{selectedSection} Settings</h3>
        <ActiveSection />
      </main>
    </div>
  );

};

export default Setting;