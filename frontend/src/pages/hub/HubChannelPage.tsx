import React, { useEffect, useState, useCallback } from 'react';
import { Plus, Settings, ExternalLink, RefreshCw, Trash2, Check, AlertCircle, MessageSquare } from 'lucide-react';
import { BsShop } from 'react-icons/bs';
import { pagesAPI } from '../../services/api';
import { SocialPage } from '../../types';
import './HubChannelPage.css';

const PLATFORMS = [
  {
    id: 'facebook',
    name: 'Facebook',
    desc: 'Kết nối Fanpage để quản lý tin nhắn và bình luận',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/b/b8/2021_Facebook_icon.svg',
    color: '#1877F2',
    category: 'social'
  },
  {
    id: 'instagram',
    name: 'Instagram',
    desc: 'Quản lý Direct Message từ Instagram Business',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/e/e7/Instagram_logo_2016.svg',
    color: '#E4405F',
    category: 'social'
  },
  {
    id: 'zalo',
    name: 'Zalo OA',
    desc: 'Chăm sóc khách hàng qua Zalo Official Account',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Icon_of_Zalo.svg',
    color: '#0068FF',
    category: 'social'
  },
  {
    id: 'shopee',
    name: 'Shopee',
    desc: 'Đồng bộ đơn hàng và tin nhắn từ Shopee',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/f/fe/Shopee_logo.svg',
    color: '#EE4D2D',
    category: 'ecommerce'
  },
  {
    id: 'tiktok',
    name: 'TikTok Shop',
    desc: 'Đồng bộ đơn hàng và tin nhắn từ TikTok Shop',
    iconUrl: 'https://upload.wikimedia.org/wikipedia/en/a/a9/TikTok_logo.svg',
    color: '#000000',
    category: 'ecommerce'
  }
];

export const HubChannelPage: React.FC = () => {
  const [pages, setPages] = useState<SocialPage[]>([]);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState<string | null>(null);
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message });
    setTimeout(() => setToast(null), 4000);
  };

  const fetchConnectedPages = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await pagesAPI.list();
      setPages(data.data || []);
    } catch (error) {
      console.error('Failed to fetch pages:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConnectedPages();

    // Check for OAuth callback params
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');
    const status = urlParams.get('status');
    const errorMsg = urlParams.get('message');

    if (code && state) {
      // Clean URL
      window.history.replaceState({}, '', window.location.pathname);
      
      // Complete connection
      const platform = state.split('-')[0];
      completeConnection(platform, code, state);
    } else if (status === 'error' && errorMsg) {
      showToast('error', decodeURIComponent(errorMsg));
      window.history.replaceState({}, '', window.location.pathname);
    }
  }, [fetchConnectedPages]);

  const completeConnection = async (platform: string, code: string, state: string) => {
    setConnecting(platform);
    try {
      await pagesAPI.completeConnect({
        platform,
        code,
        state,
      });
      showToast('success', `Kết nối ${platform.charAt(0).toUpperCase() + platform.slice(1)} thành công!`);
      fetchConnectedPages();
    } catch (error: any) {
      showToast('error', error?.response?.data?.error || `Kết nối ${platform} thất bại`);
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
      
      if (data.data?.auth_url) {
        // Redirect to OAuth
        window.location.href = data.data.auth_url;
      } else if (data.data?.manual) {
        showToast('error', data.data.message || 'Tính năng đang phát triển');
        setConnecting(null);
      } else {
        showToast('error', 'Không thể lấy đường dẫn kết nối');
        setConnecting(null);
      }
    } catch (error: any) {
      showToast('error', error?.response?.data?.error || 'Lỗi kết nối');
      setConnecting(null);
    }
  };

  const handleDisconnect = async (pageId: number) => {
    if (!confirm('Bạn có chắc muốn ngắt kết nối kênh này?')) return;
    // Note: Backend needs disconnect endpoint
    showToast('error', 'Tính năng ngắt kết nối đang phát triển');
  };

  const socialPlatforms = PLATFORMS.filter(p => p.category === 'social');
  const ecommercePlatforms = PLATFORMS.filter(p => p.category === 'ecommerce');

  const renderPlatformCard = (platform: typeof PLATFORMS[0]) => {
    const connections = getPlatformConnections(platform.id);
    const isConnected = connections.length > 0;
    const isConnecting = connecting === platform.id;

    return (
      <div 
        key={platform.id} 
        className={`channel-card ${isConnected ? 'connected' : ''}`}
        style={{ '--platform-color': platform.color } as React.CSSProperties}
      >
        <div className="channel-header">
          <div className="channel-icon-wrap">
            <img src={platform.iconUrl} alt={platform.name} className="channel-icon" />
          </div>
          <div className="channel-info">
            <h3>{platform.name}</h3>
            <p>{platform.desc}</p>
          </div>
        </div>

        <div className="channel-body">
          {isConnected ? (
            <div className="connected-pages">
              {connections.map(conn => (
                <div key={conn.id} className="connected-page-item">
                  <div className="page-info">
                    <span className="page-name">{conn.page_name}</span>
                    <span className={`page-status ${conn.status}`}>
                      {conn.status === 'active' ? 'Hoạt động' : 'Không hoạt động'}
                    </span>
                  </div>
                  {conn.warning_message && (
                    <div className="page-warning">
                      <AlertCircle size={14} />
                      <span>{conn.warning_message}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : null}
        </div>

        <div className="channel-actions">
          {isConnected ? (
            <div className="connected-actions">
              <span className="status-badge">
                <Check size={14} />
                {connections.length} đã kết nối
              </span>
              <button 
                className="btn-add-more"
                onClick={() => handleConnect(platform.id)}
                disabled={isConnecting}
              >
                <Plus size={16} /> Thêm
              </button>
            </div>
          ) : (
            <button 
              className="btn-connect"
              onClick={() => handleConnect(platform.id)}
              disabled={isConnecting}
            >
              {isConnecting ? (
                <>
                  <RefreshCw size={18} className="spin" /> Đang kết nối...
                </>
              ) : (
                <>
                  <Plus size={18} /> Kết nối
                </>
              )}
            </button>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="hub-page">
      {toast && (
        <div className={`toast ${toast.type}`}>
          {toast.type === 'success' ? <Check size={18} /> : <AlertCircle size={18} />}
          <span>{toast.message}</span>
          <button onClick={() => setToast(null)}>×</button>
        </div>
      )}

      <div className="hub-container">
        <div className="hub-header">
          <div className="header-badge">Omnichannel Messaging Hub</div>
          <h1>Kết nối đa kênh</h1>
          <p>Tăng trưởng doanh thu bằng cách kết nối và quản lý tất cả khách hàng từ các nền tảng phổ biến nhất trên một giao diện duy nhất.</p>
        </div>

        {loading ? (
          <div className="hub-loading">
            <div className="loading-spinner">
              <RefreshCw size={32} className="spin" />
            </div>
            <span>Đang đồng bộ dữ liệu...</span>
          </div>
        ) : (
          <div className="hub-content">
            <div className="channel-section">
              <div className="section-header">
                <div className="section-title-wrap">
                  <MessageSquare size={24} className="section-icon blue" />
                  <h2>Mạng xã hội & Nhắn tin</h2>
                </div>
                <span className="section-count">{socialPlatforms.length} nền tảng</span>
              </div>
              <div className="channel-grid">
                {socialPlatforms.map(renderPlatformCard)}
              </div>
            </div>
            
            <div className="channel-section">
              <div className="section-header">
                <div className="section-title-wrap">
                  <BsShop size={24} className="section-icon orange" />
                  <h2>Thương mại điện tử</h2>
                </div>
                <span className="section-count">{ecommercePlatforms.length} nền tảng</span>
              </div>
              <div className="channel-grid">
                {ecommercePlatforms.map(renderPlatformCard)}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
