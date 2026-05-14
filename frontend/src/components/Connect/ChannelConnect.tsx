import React, { useState, useEffect } from 'react';
import { pagesAPI } from '../../services/api';

export const ChannelConnect: React.FC = () => {
  const [connectingChannel, setConnectingChannel] = useState<string | null>(null);
  const [connectError, setConnectError] = useState<string>('');
  const [pages, setPages] = useState<any[]>([]);
  const [isLoadingPages, setIsLoadingPages] = useState(false);

  const loadPages = async () => {
    setIsLoadingPages(true);
    try {
      const res = await pagesAPI.list();
      setPages(res?.data?.data || []);
    } catch {
      setPages([]);
    } finally {
      setIsLoadingPages(false);
    }
  };

  useEffect(() => {
    loadPages();
  }, []);

  const handleConnect = async (channelId: string) => {
    setConnectError('');
    setConnectingChannel(channelId);
    try {
      const initRes = await pagesAPI.connect(channelId);
      const payload = initRes?.data?.data || {};
      const authURL = payload.auth_url;

      if (authURL) {
        window.location.href = authURL;
        return;
      }

      if (payload.manual) {
        alert('Tính năng này đang được phát triển hoặc cần cấu hình thêm!');
        return;
      }

      setConnectError('Kênh này chưa hỗ trợ OAuth.');
    } catch (e: any) {
      setConnectError(e?.response?.data?.error || 'Kết nối thất bại, vui lòng thử lại.');
    } finally {
      setConnectingChannel(null);
    }
  };

  const channels = [
    {
      id: 'facebook',
      name: 'Facebook',
      desc: 'Hướng dẫn kết nối Facebook',
      iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/b/b8/2021_Facebook_icon.svg',
    },
    {
      id: 'instagram',
      name: 'Instagram',
      desc: 'Hướng dẫn kết nối Instagram',
      iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/e/e7/Instagram_logo_2016.svg',
    },
    {
      id: 'zalo',
      name: 'Zalo OA (đã xác thực và trả phí)',
      desc: 'Hướng dẫn kết nối Zalo OA',
      iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Icon_of_Zalo.svg',
    },
    {
      id: 'shopee',
      name: 'Shopee',
      desc: 'Hướng dẫn kết nối Shopee cá nhân',
      iconUrl: 'https://upload.wikimedia.org/wikipedia/commons/f/fe/Shopee_logo.svg',
    },
    {
      id: 'tiktok',
      name: 'TikTok for Business',
      desc: 'Hướng dẫn kết nối TikTok Business',
      iconUrl: 'https://upload.wikimedia.org/wikipedia/en/a/a9/TikTok_logo.svg',
    },
  ];

  return (
    <div style={{ flex: 1, backgroundColor: '#f3f4f6', minHeight: '100vh', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '40px 20px', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      
      {/* Header Logo */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '24px' }}>
        <div style={{ background: '#2563eb', color: 'white', width: '32px', height: '32px', borderRadius: '8px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold', fontSize: '14px' }}>
          FB
        </div>
        <span style={{ fontSize: '24px', fontWeight: 'bold', color: '#1e293b' }}>flybox</span>
      </div>

      <h1 style={{ fontSize: '24px', fontWeight: 'bold', color: '#111827', marginBottom: '40px', textAlign: 'center' }}>
        Kết nối kênh chăm sóc & tư vấn bán hàng
      </h1>

      <div style={{ width: '100%', maxWidth: '800px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {channels.map((channel) => (
          <div key={channel.id} style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'space-between',
            backgroundColor: 'white',
            padding: '24px',
            borderRadius: '12px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
            border: '1px solid #e5e7eb'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
              <img src={channel.iconUrl} alt={channel.name} style={{ width: '40px', height: '40px', objectFit: 'contain' }} />
              <div>
                <h3 style={{ margin: '0 0 6px 0', fontSize: '16px', fontWeight: '600', color: '#111827' }}>
                  {channel.name}
                </h3>
                <a href="#" style={{ display: 'flex', alignItems: 'center', gap: '4px', color: '#6b7280', fontSize: '14px', textDecoration: 'none' }}>
                  {channel.desc} <span style={{ color: '#2563eb' }}>tại đây ↗</span>
                </a>
              </div>
            </div>

            <button
              onClick={() => handleConnect(channel.id)}
              disabled={connectingChannel === channel.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                backgroundColor: '#2563eb',
                color: 'white',
                border: 'none',
                padding: '10px 20px',
                borderRadius: '6px',
                fontSize: '14px',
                fontWeight: '600',
                cursor: connectingChannel === channel.id ? 'not-allowed' : 'pointer',
                opacity: connectingChannel === channel.id ? 0.7 : 1,
                transition: 'background-color 0.2s'
              }}
            >
              {connectingChannel === channel.id ? 'Đang tải...' : '+ Thêm kết nối'}
            </button>
          </div>
        ))}
      </div>

      {connectError && (
        <div style={{ marginTop: '24px', padding: '12px 16px', backgroundColor: '#fee2e2', color: '#991b1b', borderRadius: '8px', width: '100%', maxWidth: '800px' }}>
          {connectError}
        </div>
      )}
    </div>
  );
};
