// src/types/user.ts

// 1. Define the possible roles as an Enum (prevents typos like "admin" vs "ADMIN")
export const UserRole = {
  WL_SYSTEM_ADMIN: "WL_SYSTEM_ADMIN",
  BOOKING_TEAM: "BOOKING_TEAM",
} as const;

export type UserRole = (typeof UserRole)[keyof typeof UserRole];

// 2. Define the User interface to match your Go Struct
export interface User {
  id: number;          // Go: int
  username: string;    // Go: string
  email: string;       // Go: string
  role: UserRole;      // Go: string (mapped to our Enum)
  createdAt?: string;  // The '?' means it's optional
}