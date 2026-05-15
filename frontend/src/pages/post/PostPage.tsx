import React, { useState } from 'react';
import { Image, Video, Send, Calendar, Monitor, Smartphone, Trash2, Plus } from 'lucide-react';
import './PostPage.css';

export const PostPage: React.FC = () => {
  const [content, setContent] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<string[]>(['facebook']);
  const [previewDevice, setPreviewDevice] = useState<'desktop' | 'mobile'>('mobile');
  const [media, setMedia] = useState<{ type: 'image' | 'video'; url: string }[]>([]);

  const togglePlatform = (id: string) => {
    setSelectedPlatforms(prev => 
      prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]
    );
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    // Mock upload
    const file = e.target.files?.[0];
    if (file) {
      const url = URL.createObjectURL(file);
      setMedia(prev => [...prev, { type: file.type.startsWith('video') ? 'video' : 'image', url }]);
    }
  };

  return (
    <div className="post-page">
      <div className="post-container">
        {/* Editor Section */}
        <div className="post-editor-section">
          <div className="card editor-card">
            <div className="card-header">
              <h2>Tạo bài viết mới</h2>
              <div className="platform-selector">
                <button 
                  className={`platform-btn ${selectedPlatforms.includes('facebook') ? 'active' : ''}`}
                  onClick={() => togglePlatform('facebook')}
                >
                  <span className="platform-icon">📘</span>
                </button>
                <button 
                  className={`platform-btn ${selectedPlatforms.includes('instagram') ? 'active' : ''}`}
                  onClick={() => togglePlatform('instagram')}
                >
                  <span className="platform-icon">📷</span>
                </button>
              </div>
            </div>

            <div className="card-body">
              <textarea 
                placeholder="Bạn đang nghĩ gì?..."
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className="post-textarea"
              />

              <div className="media-preview-grid">
                {media.map((m, i) => (
                  <div key={i} className="media-item">
                    {m.type === 'image' ? <img src={m.url} alt="" /> : <video src={m.url} />}
                    <button className="remove-media" onClick={() => setMedia(prev => prev.filter((_, idx) => idx !== i))}>
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
                <label className="add-media-btn">
                  <input type="file" hidden onChange={handleFileUpload} accept="image/*,video/*" />
                  <Plus size={24} />
                  <span>Thêm ảnh/video</span>
                </label>
              </div>
            </div>

            <div className="card-footer">
              <div className="post-actions">
                <button className="btn-secondary">
                  <Calendar size={18} />
                  Lên lịch
                </button>
                <button className="btn-primary" disabled={!content && media.length === 0}>
                  <Send size={18} />
                  Đăng ngay
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Preview Section */}
        <div className="post-preview-section">
          <div className="preview-header">
            <h3>Xem trước bài viết</h3>
            <div className="device-toggle">
              <button 
                className={previewDevice === 'mobile' ? 'active' : ''} 
                onClick={() => setPreviewDevice('mobile')}
              >
                <Smartphone size={18} />
              </button>
              <button 
                className={previewDevice === 'desktop' ? 'active' : ''} 
                onClick={() => setPreviewDevice('desktop')}
              >
                <Monitor size={18} />
              </button>
            </div>
          </div>

          <div className={`preview-container ${previewDevice}`}>
            <div className="fb-post-mockup">
              <div className="post-user">
                <div className="user-avatar" />
                <div className="user-meta">
                  <div className="user-name">Fly-Box Store</div>
                  <div className="post-time">Vừa xong • <Facebook size={10} inline /></div>
                </div>
              </div>
              <div className="post-content">
                {content || <span className="placeholder-text">Nội dung bài viết sẽ hiển thị ở đây...</span>}
              </div>
              {media.length > 0 && (
                <div className="post-media">
                  {media[0].type === 'image' ? <img src={media[0].url} alt="" /> : <video src={media[0].url} controls />}
                </div>
              )}
              <div className="post-footer">
                <div className="footer-item">Thích</div>
                <div className="footer-item">Bình luận</div>
                <div className="footer-item">Chia sẻ</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
