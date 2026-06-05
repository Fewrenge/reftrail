import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import type { UserRole } from "../types/users";

interface ProtectedRouteProps {
  allowedRoles: UserRole[];
}

// Bounces users back to referrals if they try to access a route they don't have permissions for, or to login if they're not authenticated at all
export default function ProtectedRoute({ allowedRoles }: ProtectedRouteProps) {
  const { user, isAuthenticating } = useAuth();

  if (isAuthenticating) {
    return <div className="p-10 text-muted-foreground">Loading...</div>;
  }

  // Redirect to login if user isn't logged in, or to referrals if they don't have permission
  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (!allowedRoles.includes(user.role)) {
    return <Navigate to="/referrals" replace />;
  }

  // Outlet renders the children components defined inside this route path
  return <Outlet />;
}
