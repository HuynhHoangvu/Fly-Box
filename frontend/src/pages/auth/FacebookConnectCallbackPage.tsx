import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { pagesAPI } from '../../services/api';

export const FacebookConnectCallbackPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('Đang hoàn tất kết nối Facebook...');

  useEffect(() => {
    const run = async () => {
      const error = searchParams.get('error');
      const errorDescription = searchParams.get('error_description');
      const code = searchParams.get('code');
      const state = searchParams.get('state');

      if (error) {
        setStatus('error');
        setMessage(errorDescription || 'Facebook từ chối cấp quyền.');
        return;
      }

      if (!code || !state) {
        setStatus('error');
        setMessage('Thiếu code/state từ callback Facebook.');
        return;
      }

      try {
        await pagesAPI.completeConnect({
          platform: 'facebook',
          code,
          state,
          page_id: '',
        });
        setStatus('success');
        setMessage('Kết nối Facebook thành công!');
        navigate('/inbox', { replace: true });
      } catch (e: any) {
        setStatus('error');
        setMessage(e?.response?.data?.error || 'Hoàn tất kết nối thất bại.');
      }
    };

    run();
  }, [navigate, searchParams]);

  return (
    <div style={{ padding: 24, textAlign: 'center' }}>
      <div className="callback-container card" style={{ maxWidth: 400, margin: '40px auto' }}>
        <h3>{status === 'error' ? 'Kết nối thất bại' : 'Kết nối Facebook'}</h3>
        <p style={{ margin: '20px 0', color: status === 'error' ? 'var(--error)' : 'var(--text-secondary)' }}>
          {message}
        </p>
        {status === 'error' && (
          <button className="btn-primary" onClick={() => navigate('/inbox', { replace: true })}>
            Quay về trang chính
          </button>
        )}
      </div>
    </div>
  );
};
