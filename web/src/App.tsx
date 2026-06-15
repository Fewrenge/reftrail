import ProtectedRoute from "@/components/ProtectedRoutes"; // Adjust path based on your choice
import { UserRole } from "@/types/users";
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import Login from './pages/Login';
import Settings from './pages/Settings';
import Sidebar from './components/Sidebar/Sidebar';
import { useAuth } from './contexts/AuthContext';
import { AlertCircleIcon, LayoutDashboardIcon } from "lucide-react";
import Referrals from './pages/Referrals';
import ReferralDetails from './pages/ReferralDetails';
import Analytics from './pages/Analytics';
import UrgencyDistribution from './pages/Analytics/UrgencyDistribution';
import ReferralVolume from './pages/Analytics/ReferralVolume'
import DirectBookingWaitingTime from "./pages/Analytics/DirectBookingWaitingTime";

export default function App() {
  const { user, isAuthenticating } = useAuth();

  // Auth Guard: Show Login if no user is found
  if (!user && !isAuthenticating) {
    return <Login onLoginSuccess={() => window.location.reload()} />;
  }

  // Loading Screen: While checking auth
  if (isAuthenticating) {
    return <div className="min-h-screen flex items-center justify-center bg-slate-50 text-slate-400">Loading RefTrail...</div>;
  }

  return (
    <BrowserRouter>
      <div className="flex min-h-screen bg-slate-50 text-slate-900">

        <Sidebar />

        <main className="flex-1 p-8 max-w-4xl mx-auto">
          <Routes>
            {/*HOME PATH PLACEHOLDER*/}
            <Route path="/" element={
              <div className="flex flex-col items-center justify-center py-20 text-slate-400 text-center border border-dashed border-slate-200 bg-white/50 rounded-2xl p-8">
                <LayoutDashboardIcon size={48} className="mb-4 text-slate-300" />
                <h2 className="text-2xl font-bold text-slate-800">Welcome to RefTrail</h2>
                <p className="mt-1 text-sm text-slate-500 max-w-sm">
                  Select a section from the sidebar or jump straight into tracking your patient list below.
                </p>
                <Link to="/referrals" className="mt-6 inline-flex items-center justify-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition-colors">
                  Go to Referrals
                </Link>
              </div>
            } />



            <Route path="/referrals" element={<Referrals />} />
            <Route element={<ProtectedRoute allowedRoles={[UserRole.REFTRAIL_ADMIN]} />}>
              <Route path="/analytics" element={<Analytics />} />
              <Route path="/analytics/urgency-distribution" element={<UrgencyDistribution />} />
              <Route path="/analytics/referral-trend" element={<ReferralVolume />} />
              <Route path="/analytics/direct-booking-waiting-time" element={<DirectBookingWaitingTime />}/>
            </Route>

            <Route path="/referrals/:referralId" element={<ReferralDetails />} />
            <Route path="/settings/*" element={<Settings />} />

            {/* 404 CATCH-ALL */}
            <Route path="*" element={
              <div className="flex flex-col items-center justify-center py-20 text-slate-400 text-center">
                <AlertCircleIcon size={48} className="mb-4 text-slate-200" />
                <h2 className="text-2xl font-bold text-slate-800">404 - Page Not Found</h2>
                <p className="mt-1">We couldn't find the chart you were looking for.</p>
                <Link to="/" className="text-blue-600 underline mt-4 font-medium">Return Home</Link>
              </div>
            } />
          </Routes>
        </main>

      </div>
    </BrowserRouter>
  );
}
