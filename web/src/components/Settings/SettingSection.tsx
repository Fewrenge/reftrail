import React from "react";
import { cn } from "@/lib/utils";

interface SettingSectionProps {
  title?: React.ReactNode;
  description?: string;
  children: React.ReactNode;
  className?: string;
  actions?: React.ReactNode;
}

const SettingSection: React.FC<SettingSectionProps> = ({ title, description, children, className, actions }) => {
  return (
    <div className={cn("w-full flex flex-col gap-4", className)}>
      {(title || description || actions) && (
        <div className="flex flex-row items-start justify-between gap-2 px-1">
          <div className="flex flex-col gap-0.5">
            {title && (
              <h3 className="text-lg font-semibold text-foreground tracking-tight">
                {title}
              </h3>
            )}
            {description && (
              <p className="text-sm text-muted-foreground leading-relaxed">
                {description}
              </p>
            )}
          </div>
          {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
        </div>
      )}
      <div className="flex flex-col gap-4">{children}</div>
    </div>
  );
};

export default SettingSection;
