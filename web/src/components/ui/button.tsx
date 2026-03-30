import React from "react"

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'outline' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
}

export const Button = ({ 
  variant = 'primary', 
  size = 'md', 
  className = "", 
  ...props 
}: ButtonProps) => {
  
  const baseStyles = "inline-flex items-center rounded-xl font-medium transition-all cursor-pointer disabled:opacity-50";
  
  const variants = {
    primary: "bg-blue-600 text-white hover:bg-blue-700 shadow-sm",
    outline: "border border-slate-200 bg-white text-slate-700 hover:bg-slate-50",
    ghost: "text-slate-600 hover:bg-slate-100",
    danger: "bg-red-50 text-red-600 hover:bg-red-100"
  };

  const sizes = {
    sm: "px-3 py-1.5 text-xs",
    md: "px-4 py-2 text-sm",
    lg: "px-6 py-3 text-base"
  };

  const alignment = className.includes("justify-") ? "" : "justify-center";

  return (
    <button 
      className={`${baseStyles} ${alignment} ${variants[variant]} ${sizes[size]} ${className}`} 
      {...props} 
    />
  );
};
