import { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";
import { UserIcon, UsersIcon, TagIcon } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useAuth } from "../contexts/AuthContext";
import { UserRole } from "../types/users";


import ProfileSection from "../components/Settings/ProfileSection";
import MemberSection from "../components/Settings/MemberSection";
import TagSection from "@/components/Settings/TagSection";

import SectionMenuItem from "@/components/Settings/SectionMenuItem";

type SettingSection = "profile" | "member" | "tag";

const BASIC_SECTIONS: SettingSection[] = ["profile",];
const ADMIN_SECTIONS: SettingSection[] = ["member", "tag", ];

const SECTION_ICON_MAP: Record<SettingSection, LucideIcon> = {
  profile: UserIcon,
  // login: LogInIcon,
  member: UsersIcon,
  tag: TagIcon,
};

const SECTION_COMPONENT_MAP: Record<SettingSection, React.ComponentType> = {
  profile: ProfileSection,
  // login: () => <div className="p-4 text-muted-foreground text-sm">Security settings coming soon...</div>,
  member: MemberSection,
  tag: TagSection,
};

const Setting = () => {
  const location = useLocation();
  const { user, isAuthenticating } = useAuth();
  const [selectedSection, setSelectedSection] = useState<SettingSection>("profile");

  const isAdmin = user?.role === UserRole.REFTRAIL_ADMIN;

  const settingsSectionList = useMemo(() => {
    return isAdmin ? [...BASIC_SECTIONS, ...ADMIN_SECTIONS] : BASIC_SECTIONS;
  }, [isAdmin]);

  useEffect(() => {
    const hash = location.hash.slice(1) as SettingSection;
    const nextSection = settingsSectionList.includes(hash) ? hash : "profile";
    setSelectedSection(nextSection);
  }, [location.hash, settingsSectionList]);

  if (isAuthenticating) return <div className="p-10 text-muted-foreground">Loading...</div>;

  const ActiveSection = SECTION_COMPONENT_MAP[selectedSection];

  return (
    <section className="@container w-full max-w-5xl min-h-full flex flex-col justify-start items-start sm:pt-3 md:pt-6 pb-8 mx-auto">
      <div className="w-full px-4 sm:px-6">
        {/* This wrapper creates the "Single Card" look with a sidebar inside */}
        <div className="w-full border border-border flex flex-row justify-start items-start px-4 py-3 rounded-xl bg-background">
          
          {/* SIDEBAR (Visible on sm screens and up) */}
          <div className="hidden sm:flex flex-col justify-start items-start w-44 h-auto shrink-0 py-2 border-r border-border pr-4 mr-4">
            <span className="text-xs mb-2 pl-3 font-mono font-bold uppercase tracking-widest text-muted-foreground opacity-60">
              Settings
            </span>
            
            <div className="w-full flex flex-col justify-start items-start gap-1">
              {settingsSectionList.map((id) => (
                <SectionMenuItem
                  key={id}
                  text={id.charAt(0).toUpperCase() + id.slice(1)} // Basic capitalization
                  icon={SECTION_ICON_MAP[id]}
                  isSelected={selectedSection === id}
                  onClick={() => window.location.hash = id}
                />
              ))}
            </div>
          </div>

          {/* MAIN CONTENT AREA */}
          <main className="flex-1 w-full min-w-0 py-2">
            <ActiveSection />
          </main>
        </div>
      </div>
    </section>
  );
};

export default Setting;
