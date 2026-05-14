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
  const [connecting, setConnecting] = useState<string | null>(null);
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

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

    // Listen for OAuth callback from popup
    const handleOAuthCallback = async (event: MessageEvent) => {
      if (event.data?.type === 'FB_OAUTH_SUCCESS') {
        const { code, state } = event.data;
        await completeFacebookConnect(code, state);
      }
    };

    // Check for stored OAuth data
    const storedCode = sessionStorage.getItem('fb_oauth_code');
    const storedState = sessionStorage.getItem('fb_oauth_state');
    if (storedCode && storedState) {
      sessionStorage.removeItem('fb_oauth_code');
      sessionStorage.removeItem('fb_oauth_state');
      completeFacebookConnect(storedCode, storedState);
    }

    window.addEventListener('message', handleOAuthCallback);
    return () => window.removeEventListener('message', handleOAuthCallback);
  }, []);

  const completeFacebookConnect = async (code: string, state: string) => {
    setConnecting('facebook');
    try {
      await pagesAPI.completeConnect({
        platform: 'facebook',
        code,
        state,
        page_id: '',
      });
      setNotification({ type: 'success', message: 'Kết nối Facebook thành công!' });
      fetchConnectedPages();
    } catch (error: any) {
      setNotification({ type: 'error', message: error?.response?.data?.error || 'Kết nối Facebook thất bại' });
    } finally {
      setConnecting(null);
    }
  };

  const getPlatformConnections = (platformId: string) => {
    return pages.filter(p => p.platform === platformId);
  };

  const handleConnect = async (platformId: string) => {
    try {
      setConnecting(platformId);
      const { data } = await pagesAPI.connect(platformId);
      
      if (data.data && data.data.auth_url) {
        // Open OAuth in popup window
        const width = 600;
        const height = 700;
        const left = window.screenX + (window.outerWidth - width) / 2;
        const top = window.screenY + (window.outerHeight - height) / 2;
        
        const popup = window.open(
          data.data.auth_url,
          'oauth_popup',
          `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no`
        );

        // Listen for the callback in the popup
        const checkPopup = setInterval(() => {
          if (!popup || popup.closed) {
            clearInterval(checkPopup);
            setConnecting(null);
          }
        }, 1000);
      } else if (data.data && data.data.manual) {
        alert(data.data.message || 'Tính năng này đang được phát triển hoặc cần cấu hình thêm!');
        setConnecting(null);
      } else {
        alert('Không thể lấy được đường dẫn kết nối!');
        setConnecting(null);
      }
    } catch (error) {
      console.error('Lỗi khi kết nối kênh:', error);
      setConnecting(null);
    }
  };

  const socialPlatforms = PLATFORMS.filter(p => p.category === 'social');
  const ecommercePlatforms = PLATFORMS.filter(p => p.category === 'ecommerce');

  const renderPlatformCard = (platform: typeof PLATFORMS[0]) => {
    const connections = getPlatformConnections(platform.id);
    const isConnected = connections.length > 0;
    const isConnecting = connecting === platform.id;

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
                  disabled={isConnecting}
                >
                  <Plus size={16} /> Thêm
                </button>
              </div>
            </div>
          ) : (
            <button 
              className={`btn-connect ${['facebook', 'zalo'].includes(platform.id) ? 'primary-action' : ''}`}
              onClick={() => handleConnect(platform.id)}
              disabled={isConnecting}
            >
              {isConnecting ? (
                <>
                  <RefreshCw size={18} className="spin" /> Đang kết nối...
                </>
              ) : (
                <>
                  <Plus size={18} /> Kết nối {platform.name}
                </>
              )}
            </button>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="hub-page-wrapper">
      {notification && (
        <div className={`notification ${notification.type}`}>
          {notification.message}
          <button onClick={() => setNotification(null)}>×</button>
        </div>
      )}
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
