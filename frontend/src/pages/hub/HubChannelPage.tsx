import React, { useEffect, useState } from 'react';
import { Plus, Settings, ExternalLink, RefreshCw } from 'lucide-react';
import { pagesAPI } from '../../services/api';
import { SocialPage } from '../../types/dashboard';
import './HubChannelPage.css';

const PLATFORMS = [
  {
    id: 'facebook',
    name: 'Facebook',
    desc: 'Kết nối Fanpage để quản lý tin nhắn và bình luận',
    link: '#',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/b/b8/2021_Facebook_icon.svg',
    category: 'social'
  },
  {
    id: 'instagram',
    name: 'Instagram',
    desc: 'Quản lý Direct Message từ Instagram',
    link: '#',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/e/e7/Instagram_logo_2016.svg',
    category: 'social'
  },
  {
    id: 'zalo',
    name: 'Zalo OA',
    desc: 'Chăm sóc khách hàng qua Zalo Official Account',
    link: '#',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Icon_of_Zalo.svg',
    category: 'social'
  },
  {
    id: 'shopee',
    name: 'Shopee',
    desc: 'Đồng bộ đơn hàng và tin nhắn từ Shopee',
    link: '#',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/f/fe/Shopee_logo.svg',
    category: 'ecommerce'
  },
  {
    id: 'tiktok',
    name: 'TikTok Shop',
    desc: 'Đồng bộ đơn hàng và tin nhắn từ TikTok Shop',
    link: '#',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/en/a/a9/TikTok_logo.svg',
    category: 'ecommerce'
  }
];

export const HubChannelPage: React.FC = () => {
  const [pages, setPages] = useState<SocialPage[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchConnectedPages = async () => {
    setLoading(true);
    try {
      const { data } = await pagesAPI.list();
      setPages(data.data || []);
    } catch (error) {
      console.error('Failed to fetch pages:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConnectedPages();
  }, []);

  const getPlatformConnections = (platformId: string) => {
    return pages.filter(p => p.platform === platformId);
  };

  const handleConnect = async (platformId: string) => {
    try {
      setLoading(true);
      const { data } = await pagesAPI.connect(platformId);
      
      if (data.data && data.data.auth_url) {
        window.location.href = data.data.auth_url;
      } else if (data.data && data.data.manual) {
        alert(data.data.message || 'Tính năng này đang được phát triển hoặc cần cấu hình thêm!');
        setLoading(false);
      } else {
        alert('Không thể lấy được đường dẫn kết nối!');
        setLoading(false);
      }
    } catch (error) {
      console.error('Lỗi khi kết nối kênh:', error);
      alert('Đã xảy ra lỗi khi yêu cầu kết nối kênh.');
      setLoading(false);
    }
  };

  const socialPlatforms = PLATFORMS.filter(p => p.category === 'social');
  const ecommercePlatforms = PLATFORMS.filter(p => p.category === 'ecommerce');

  const renderPlatformCard = (platform: typeof PLATFORMS[0]) => {
    const connections = getPlatformConnections(platform.id);
    const isConnected = connections.length > 0;

    return (
      <div key={platform.id} className="channel-card" data-platform={platform.id}>
        <div className="channel-info">
          <div className="channel-icon-wrap">
            <img src={platform.iconUrl} alt={platform.name} className="channel-icon" />
          </div>
          <div className="channel-text">
            <h3>{platform.name}</h3>
            <a href={platform.link} target="_blank" rel="noreferrer" className="help-link">
              {platform.desc} <ExternalLink size={14} />
            </a>
          </div>
        </div>

        <div className="channel-actions">
          {isConnected ? (
            <div className="connected-status">
              <span className="status-badge">
                <span className="status-dot"></span>
                Đã kết nối {connections.length}
              </span>
              <div className="connected-actions">
                <button 
                  className="btn-settings"
                  title="Cài đặt kênh"
                >
                  <Settings size={18} />
                </button>
                <button 
                  className="btn-add-more"
                  onClick={() => handleConnect(platform.id)}
                >
                  <Plus size={16} /> Thêm
                </button>
              </div>
            </div>
          ) : (
            <button 
              className={`btn-connect ${['facebook', 'zalo'].includes(platform.id) ? 'primary-action' : ''}`}
              onClick={() => handleConnect(platform.id)}
            >
              <Plus size={18} /> Kết nối {platform.name}
            </button>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="hub-page-wrapper">
      <div className="hub-container">
        <div className="hub-header">
          <div className="hub-logo">
            <div className="logo-icon">FB</div>
            <span>Flybox Hub</span>
          </div>
          <h1>Kết nối đa kênh</h1>
          <p>Quản lý tập trung tin nhắn, bình luận và đơn hàng từ nhiều nền tảng khác nhau trên một giao diện duy nhất.</p>
        </div>

        {loading ? (
          <div className="hub-loading">
            <div className="spinner"></div>
            <span>Đang tải cấu hình kênh...</span>
          </div>
        ) : (
          <>
            <div className="channel-section">
              <h2 className="section-title">Mạng xã hội & Nhắn tin</h2>
              <div className="channel-list">
                {socialPlatforms.map(renderPlatformCard)}
              </div>
            </div>
            
            <div className="channel-section">
              <h2 className="section-title">Sàn thương mại điện tử</h2>
              <div className="channel-list">
                {ecommercePlatforms.map(renderPlatformCard)}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
