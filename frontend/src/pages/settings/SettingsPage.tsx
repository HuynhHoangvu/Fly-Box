import React, { useState } from 'react';
import { User, Bell, Globe, Shield, CreditCard, Users, Trash2, Save, LogOut } from 'lucide-react';
import './SettingsPage.css';

export const SettingsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('profile');

  const tabs = [
    { id: 'profile', label: 'Hồ sơ', icon: User },
    { id: 'notifications', label: 'Thông báo', icon: Bell },
    { id: 'team', label: 'Thành viên', icon: Users },
    { id: 'billing', label: 'Thanh toán', icon: CreditCard },
    { id: 'security', label: 'Bảo mật', icon: Shield },
    { id: 'general', label: 'Cài đặt chung', icon: Globe },
  ];

  return (
    <div className="settings-page">
      <div className="settings-container">
        <aside className="settings-sidebar">
          <h2>Cài đặt</h2>
          <nav>
            {tabs.map(tab => (
              <button 
                key={tab.id}
                className={`tab-btn ${activeTab === tab.id ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <tab.icon size={18} />
                {tab.label}
              </button>
            ))}
          </nav>
          <div className="sidebar-footer">
            <button className="logout-btn">
              <LogOut size={18} />
              Đăng xuất
            </button>
          </div>
        </aside>

        <main className="settings-content">
          {activeTab === 'profile' && (
            <div className="settings-section">
              <h3>Thông tin hồ sơ</h3>
              <p className="section-desc">Cập nhật ảnh đại diện và thông tin cá nhân của bạn.</p>

              <div className="profile-upload">
                <div className="avatar-preview" />
                <div className="upload-actions">
                  <button className="btn-primary">Thay đổi ảnh</button>
                  <button className="btn-secondary">Gỡ bỏ</button>
                </div>
              </div>

              <form className="settings-form">
                <div className="form-row">
                  <div className="form-group">
                    <label>Họ và tên</label>
                    <input type="text" defaultValue="Admin User" />
                  </div>
                  <div className="form-group">
                    <label>Email</label>
                    <input type="email" defaultValue="admin@flybox.com" disabled />
                  </div>
                </div>
                <div className="form-group">
                  <label>Bio</label>
                  <textarea placeholder="Giới thiệu ngắn về bạn..." rows={4} />
                </div>
                
                <div className="form-actions">
                  <button type="button" className="btn-secondary">Hủy</button>
                  <button type="submit" className="btn-primary">
                    <Save size={18} />
                    Lưu thay đổi
                  </button>
                </div>
              </form>
            </div>
          )}

          {activeTab !== 'profile' && (
            <div className="empty-settings-state">
              <div className="icon-wrap">
                {React.createElement(tabs.find(t => t.id === activeTab)?.icon || User, { size: 48 })}
              </div>
              <h4>Tính năng đang phát triển</h4>
              <p>Trang {tabs.find(t => t.id === activeTab)?.label} sẽ sớm được ra mắt trong bản cập nhật tới.</p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
};
