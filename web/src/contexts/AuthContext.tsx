import React, { createContext, useContext, useState, useEffect } from "react";
import { getUserMe } from "../services/userService";

import type { User } from "../types/users";


interface AuthContextType {
  user: User | null; // Removed any to make it safer
  loading: boolean;
  onLogout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({ user: null, loading: true, onLogout: async () => {} });

// 2. The Provider (The wrapper for your whole app)
export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const onLogout = async () => {
    await fetch('/api/v1/logout', { method: 'POST', credentials: 'same-origin' });
    setUser(null);
    // Optional: window.location.href = "/"; 
  };

  useEffect(() => {
    getUserMe()
      .then((data) => setUser(data))
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, onLogout }}>
      {children}
    </AuthContext.Provider>
  );
};

// 3. The Easy-to-use Hook
export const useAuth = () => useContext(AuthContext);
