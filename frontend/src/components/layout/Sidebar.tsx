import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../store/useAuthStore';
import { AppIcon, IconKey } from '../common/AppIcon';
import { NotificationBell } from '../Notifications/NotificationBell';

export const Sidebar: React.FC = () => {
  const { pathname } = useLocation();
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const navItems: Array<{ id: string; path: string; icon: IconKey; label: string }> = [
    { id: 'inbox', path: '/inbox', icon: 'inbox', label: 'Hộp thư' },
    { id: 'post', path: '/post', icon: 'pencil', label: 'Đăng bài' },
    { id: 'automation', path: '/automation', icon: 'cog', label: 'Tự động hóa' },
    { id: 'customers', path: '/customers', icon: 'user', label: 'Khách hàng' },
    { id: 'settings', path: '/settings', icon: 'cog', label: 'Cài đặt' },
  ];

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  const isActive = (path: string) => {
    if (path === '/' && pathname === '/') return true;
    if (path !== '/' && pathname.startsWith(path)) return true;
    return false;
  };

  return (
    <div className="sidebar">
      <div className="logo-section" style={{ padding: '10px 0', textAlign: 'center', color: 'var(--primary)', fontWeight: 'bold', fontSize: '18px' }}>
        FB
      </div>
      {navItems.map((item) => (
        <div
          key={item.id}
          className={`nav-item ${isActive(item.path) ? 'active' : ''}`}
          onClick={() => navigate(item.path)}
          title={item.label}
        >
          <AppIcon name={item.icon} size={20} />
          {item.id === 'inbox' && <span className="badge" />}
        </div>
      ))}
      <div style={{ marginTop: 'auto' }}>
        <NotificationBell />
      </div>
      <div className="nav-item" onClick={handleLogout} title="Đăng xuất">
        <AppIcon name="logout" size={20} />
      </div>
    </div>
  );
};
