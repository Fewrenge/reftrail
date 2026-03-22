import React, { createContext, useContext, useState, useEffect } from "react";
import { getUserMe } from "../services/userService";

// 1. Define what's in the "Backpack"
interface AuthContextType {
  user: any | null;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType>({ user: null, loading: true });

// 2. The Provider (The wrapper for your whole app)
export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getUserMe()
      .then((data) => setUser(data))
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading }}>
      {children}
    </AuthContext.Provider>
  );
};

// 3. The Easy-to-use Hook
export const useAuth = () => useContext(AuthContext);
