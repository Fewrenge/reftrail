import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Merges Tailwind classes and handles conflicts.
 * Example: cn("p-4 bg-red-500", true && "p-8", className)
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}