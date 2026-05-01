import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import SignIn from "./pages/AuthPages/SignIn";
import NotFound from "./pages/OtherPage/NotFound";
import AppLayout from "./layout/AppLayout";
import { ScrollToTop } from "./components/common/ScrollToTop";
import Home from "./pages/Dashboard/Home";
import SIList from "./pages/SI/SIList";
import SIDetail from "./pages/SI/SIDetail";
import Schools from "./pages/Schools/Schools";
import AuditLogs from "./pages/AuditLogs/AuditLogs";
import { isAuthenticated } from "./api";

// Auth guard component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!isAuthenticated()) {
    return <Navigate to="/signin" replace />;
  }
  return <>{children}</>;
}

export default function App() {
  return (
    <>
      <Router>
        <ScrollToTop />
        <Routes>
          {/* Protected Dashboard Layout */}
          <Route
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            <Route index path="/" element={<Home />} />
            <Route path="/si" element={<SIList />} />
            <Route path="/si/:id" element={<SIDetail />} />
            <Route path="/schools" element={<Schools />} />
            <Route path="/audit-logs" element={<AuditLogs />} />
          </Route>

          {/* Auth */}
          <Route path="/signin" element={<SignIn />} />

          {/* Fallback */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Router>
    </>
  );
}
