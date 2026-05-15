import React, { Suspense, lazy } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { MainAppShell } from './components/layout/MainAppShell';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { useAuthStore } from './store/useAuthStore';

// Lazy load pages for better performance
const HubChannelPage = lazy(() => import('./pages/hub/HubChannelPage').then(m => ({ default: m.HubChannelPage })));
const LoginPage = lazy(() => import('./pages/auth/LoginPage').then(m => ({ default: m.LoginPage })));
const FacebookConnectCallbackPage = lazy(() => import('./pages/auth/FacebookConnectCallbackPage').then(m => ({ default: m.FacebookConnectCallbackPage })));
const InboxPage = lazy(() => import('./pages/InboxPage').then(m => ({ default: m.InboxPage })));
const DashboardPage = lazy(() => import('./pages/DashboardPage').then(m => ({ default: m.DashboardPage })));
const NotificationCenter = lazy(() => import('./components/Notifications/NotificationCenter').then(m => ({ default: m.NotificationCenter })));


const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID || '';

// Loading Fallback
const PageLoader = () => (
  <div className="page-loader" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', flexDirection: 'column', gap: '20px' }}>
    <div className="spinner" style={{ width: '40px', height: '40px', border: '4px solid #f3f3f3', borderTop: '4px solid var(--primary)', borderRadius: '50%', animation: 'spin 1s linear infinite' }} />
    <style>{`@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }`}</style>
  </div>
);

const AppContent: React.FC = () => {
  const user = useAuthStore(s => s.user);
  const userId = user ? Number(user.id) : undefined;

  return (
    <WebSocketProvider userId={userId}>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Auth Routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/connect/callback" element={<FacebookConnectCallbackPage />} />
          
          {/* Protected Routes */}
          <Route path="/" element={<MainAppShell />}>
            <Route index element={<Navigate to="/hub/channels" replace />} />
            <Route path="hub/channels" element={<HubChannelPage />} />
            <Route path="inbox" element={<InboxPage />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="notifications" element={<NotificationCenter />} />
            <Route path="post" element={<div className="workspace"><h1>Post</h1></div>} />
            <Route path="automation" element={<div className="workspace"><h1>Automation</h1></div>} />
            <Route path="customers" element={<div className="workspace"><h1>Customers</h1></div>} />
            <Route path="settings" element={<div className="workspace"><h1>Settings</h1></div>} />
          </Route>

          <Route path="*" element={<Navigate to="/hub/channels" replace />} />

        </Routes>
      </Suspense>
    </WebSocketProvider>
  );
};

const App: React.FC = () => {
  return (
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </GoogleOAuthProvider>
  );
};

export default App;
