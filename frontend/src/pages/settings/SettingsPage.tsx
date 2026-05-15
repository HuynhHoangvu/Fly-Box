import React, { useState } from 'react';
import { useAuthStore } from '../../store/useAuthStore';
import { Bell, Globe, Shield, Users, Save, LogOut, User } from 'lucide-react';
import './SettingsPage.css';

export const SettingsPage: React.FC = () => {
  const user = useAuthStore(s => s.user);
  const logout = useAuthStore(s => s.logout);
  const [activeTab, setActiveTab] = useState('profile');
  const [saved, setSaved] = useState(false);

  const [profileForm, setProfileForm] = useState({
    name: user?.email?.split('@')[0] || '',
    email: user?.email || '',
    bio: '',
  });

  const [notifSettings, setNotifSettings] = useState({
    newMessage: true,
    newComment: true,
    newOrder: true,
    newFollower: false,
    sound: true,
    browser: false,
  });

  const tabs = [
    { id: 'profile', label: 'Hồ sơ', icon: User },
    { id: 'notifications', label: 'Thông báo', icon: Bell },
    { id: 'team', label: 'Thành viên', icon: Users },
    { id: 'security', label: 'Bảo mật', icon: Shield },
    { id: 'general', label: 'Cài đặt chung', icon: Globe },
  ];

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="settings-page">
      <div className="settings-header">
        <h1>Cài đặt</h1>
        <p className="settings-subtitle">Quản lý tài khoản và tùy chỉnh trải nghiệm của bạn</p>
      </div>

      <div className="settings-body">
        {/* Tab Navigation - Horizontal */}
        <div className="settings-tabs">
          {tabs.map(tab => (
            <button
              key={tab.id}
              className={`settings-tab ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              <tab.icon size={16} />
              <span>{tab.label}</span>
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="settings-panel">
          {activeTab === 'profile' && (
            <div className="settings-section">
              <h3>Thông tin hồ sơ</h3>
              <p className="section-desc">Cập nhật ảnh đại diện và thông tin cá nhân</p>

              <div className="profile-avatar-section">
                <div className="avatar-circle">
                  {profileForm.name ? profileForm.name[0].toUpperCase() : 'U'}
                </div>
                <div className="avatar-actions">
                  <button className="btn-outline">Thay đổi ảnh</button>
                  <button className="btn-ghost">Gỡ bỏ</button>
                </div>
              </div>

              <div className="form-grid">
                <div className="form-group">
                  <label>Họ và tên</label>
                  <input
                    type="text"
                    value={profileForm.name}
                    onChange={e => setProfileForm({ ...profileForm, name: e.target.value })}
                    placeholder="Nhập họ và tên"
                  />
                </div>
                <div className="form-group">
                  <label>Email</label>
                  <input type="email" value={profileForm.email} disabled />
                </div>
              </div>

              <div className="form-group full-width">
                <label>Giới thiệu</label>
                <textarea
                  value={profileForm.bio}
                  onChange={e => setProfileForm({ ...profileForm, bio: e.target.value })}
                  placeholder="Giới thiệu ngắn về bạn..."
                  rows={3}
                />
              </div>

              <div className="form-actions">
                <button className="btn-ghost">Hủy</button>
                <button className="btn-primary" onClick={handleSave}>
                  <Save size={16} />
                  {saved ? 'Đã lưu!' : 'Lưu thay đổi'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="settings-section">
              <h3>Cài đặt thông báo</h3>
              <p className="section-desc">Chọn loại thông báo bạn muốn nhận</p>

              <div className="toggle-list">
                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Tin nhắn mới</span>
                    <span className="toggle-desc">Nhận thông báo khi có tin nhắn từ khách hàng</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.newMessage}
                      onChange={e => setNotifSettings({ ...notifSettings, newMessage: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>

                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Bình luận mới</span>
                    <span className="toggle-desc">Nhận thông báo khi có bình luận trên bài viết</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.newComment}
                      onChange={e => setNotifSettings({ ...notifSettings, newComment: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>

                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Đơn hàng mới</span>
                    <span className="toggle-desc">Nhận thông báo khi có đơn hàng từ Shopee, TikTok Shop</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.newOrder}
                      onChange={e => setNotifSettings({ ...notifSettings, newOrder: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>

                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Người theo dõi mới</span>
                    <span className="toggle-desc">Nhận thông báo khi có người theo dõi mới</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.newFollower}
                      onChange={e => setNotifSettings({ ...notifSettings, newFollower: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>

                <div className="toggle-divider" />

                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Âm thanh thông báo</span>
                    <span className="toggle-desc">Phát âm thanh khi có thông báo mới</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.sound}
                      onChange={e => setNotifSettings({ ...notifSettings, sound: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>

                <div className="toggle-item">
                  <div className="toggle-info">
                    <span className="toggle-label">Thông báo trình duyệt</span>
                    <span className="toggle-desc">Hiển thị thông báo trên trình duyệt</span>
                  </div>
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={notifSettings.browser}
                      onChange={e => setNotifSettings({ ...notifSettings, browser: e.target.checked })}
                    />
                    <span className="toggle-slider" />
                  </label>
                </div>
              </div>

              <div className="form-actions">
                <button className="btn-primary" onClick={handleSave}>
                  <Save size={16} />
                  {saved ? 'Đã lưu!' : 'Lưu thay đổi'}
                </button>
              </div>
            </div>
          )}

          {activeTab === 'team' && (
            <div className="settings-section">
              <h3>Quản lý thành viên</h3>
              <p className="section-desc">Thêm hoặc quản lý thành viên trong đội ngũ</p>
              <div className="empty-state">
                <Users size={48} />
                <h4>Tính năng đang phát triển</h4>
                <p>Quản lý thành viên sẽ sớm được ra mắt</p>
              </div>
            </div>
          )}

          {activeTab === 'security' && (
            <div className="settings-section">
              <h3>Bảo mật tài khoản</h3>
              <p className="section-desc">Quản lý mật khẩu và bảo mật tài khoản</p>

              <div className="form-grid">
                <div className="form-group">
                  <label>Mật khẩu hiện tại</label>
                  <input type="password" placeholder="Nhập mật khẩu hiện tại" />
                </div>
                <div className="form-group">
                  <label>Mật khẩu mới</label>
                  <input type="password" placeholder="Nhập mật khẩu mới" />
                </div>
              </div>
              <div className="form-group" style={{ maxWidth: '50%' }}>
                <label>Xác nhận mật khẩu</label>
                <input type="password" placeholder="Nhập lại mật khẩu mới" />
              </div>

              <div className="form-actions">
                <button className="btn-primary" onClick={handleSave}>
                  <Shield size={16} />
                  Cập nhật mật khẩu
                </button>
              </div>

              <div className="danger-zone">
                <h4>Vùng nguy hiểm</h4>
                <p>Xóa tài khoản vĩnh viễn. Hành động này không thể hoàn tác.</p>
                <button className="btn-danger">Xóa tài khoản</button>
              </div>
            </div>
          )}

          {activeTab === 'general' && (
            <div className="settings-section">
              <h3>Cài đặt chung</h3>
              <p className="section-desc">Tùy chỉnh ngôn ngữ và giao diện</p>

              <div className="form-grid">
                <div className="form-group">
                  <label>Ngôn ngữ</label>
                  <select defaultValue="vi">
                    <option value="vi">Tiếng Việt</option>
                    <option value="en">English</option>
                  </select>
                </div>
                <div className="form-group">
                  <label>Múi giờ</label>
                  <select defaultValue="Asia/Ho_Chi_Minh">
                    <option value="Asia/Ho_Chi_Minh">GMT+7 (Hà Nội)</option>
                    <option value="Asia/Bangkok">GMT+7 (Bangkok)</option>
                    <option value="Asia/Singapore">GMT+8 (Singapore)</option>
                  </select>
                </div>
              </div>

              <div className="form-actions">
                <button className="btn-primary" onClick={handleSave}>
                  <Save size={16} />
                  {saved ? 'Đã lưu!' : 'Lưu thay đổi'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
