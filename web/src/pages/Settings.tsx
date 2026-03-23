import { useState, useEffect } from "react";
import { UserIcon, InstanceSection, UsersIcon } from "lucide-react";

// 1. Define sections
type SettingSection = "profile" | "login" | "member";

// 2. Map names to Icons
const SECTION_ICON_MAP: Record<SettingSection, any> = {
  profile: UserIcon,
  login: InstanceSection,
  member: UsersIcon,
};

// 3. Map names to Components
const SECTION_COMPONENT_MAP: Record<SettingSection, React.ComponentType> = {
  profile: () => <div className="p-4">User Profile Form goes here</div>,
  login: () => <div className="p-4">Password & Security settings</div>,
  member: () => <div className="p-4">Member settings</div>,
};


const Setting = () => {

}