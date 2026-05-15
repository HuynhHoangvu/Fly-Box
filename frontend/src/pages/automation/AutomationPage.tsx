import React, { useState } from 'react';
import { Bot, Zap, MessageSquare, Tag, Clock, ChevronRight, Plus, ToggleLeft as Toggle, Play, Pause } from 'lucide-react';
import './AutomationPage.css';

interface Rule {
  id: string;
  name: string;
  description: string;
  trigger: string;
  action: string;
  enabled: boolean;
  type: 'keyword' | 'greeting' | 'follow';
}

export const AutomationPage: React.FC = () => {
  const [rules, setRules] = useState<Rule[]>([
    {
      id: '1',
      name: 'Chào mừng khách hàng mới',
      description: 'Gửi tin nhắn chào mừng khi khách hàng inbox lần đầu',
      trigger: 'Tin nhắn đầu tiên',
      action: 'Gửi tin nhắn mẫu',
      enabled: true,
      type: 'greeting'
    },
    {
      id: '2',
      name: 'Keyword "Báo giá"',
      description: 'Tự động gửi bảng giá khi khách chat từ "báo giá"',
      trigger: 'Keyword: báo giá, giá',
      action: 'Gửi hình ảnh + Text',
      enabled: true,
      type: 'keyword'
    },
    {
      id: '3',
      name: 'Tự động gắn tag',
      description: 'Gắn tag "Khách hàng tiềm năng" khi có số điện thoại',
      trigger: 'Có số điện thoại',
      action: 'Gắn tag: Potential',
      enabled: false,
      type: 'follow'
    }
  ]);

  const toggleRule = (id: string) => {
    setRules(prev => prev.map(r => r.id === id ? { ...r, enabled: !r.enabled } : r));
  };

  return (
    <div className="automation-page">
      <div className="automation-header">
        <div className="header-info">
          <h1>Tự động hóa</h1>
          <p>Thiết lập các kịch bản tự động để phản hồi khách hàng nhanh chóng 24/7.</p>
        </div>
        <button className="btn-primary">
          <Plus size={18} />
          Tạo kịch bản mới
        </button>
      </div>

      <div className="automation-stats">
        <div className="stat-card">
          <Bot className="stat-icon" />
          <div className="stat-meta">
            <span className="stat-value">1,284</span>
            <span className="stat-label">Tin nhắn tự động</span>
          </div>
        </div>
        <div className="stat-card">
          <Zap className="stat-icon yellow" />
          <div className="stat-meta">
            <span className="stat-value">85%</span>
            <span className="stat-label">Tỷ lệ phản hồi bot</span>
          </div>
        </div>
        <div className="stat-card">
          <Clock className="stat-icon green" />
          <div className="stat-meta">
            <span className="stat-value">2s</span>
            <span className="stat-label">Thời gian phản hồi TB</span>
          </div>
        </div>
      </div>

      <div className="rules-section">
        <div className="section-header">
          <h2>Danh sách kịch bản</h2>
          <div className="filters">
            <button className="filter-btn active">Tất cả</button>
            <button className="filter-btn">Đang chạy</button>
            <button className="filter-btn">Bản nháp</button>
          </div>
        </div>

        <div className="rules-grid">
          {rules.map(rule => (
            <div key={rule.id} className={`rule-card ${rule.enabled ? '' : 'disabled'}`}>
              <div className="rule-header">
                <div className={`rule-type-icon ${rule.type}`}>
                  {rule.type === 'greeting' && <MessageSquare size={20} />}
                  {rule.type === 'keyword' && <Zap size={20} />}
                  {rule.type === 'follow' && <Tag size={20} />}
                </div>
                <div className="rule-status-toggle" onClick={() => toggleRule(rule.id)}>
                  {rule.enabled ? <Play size={18} className="icon-play" /> : <Pause size={18} />}
                </div>
              </div>
              
              <div className="rule-body">
                <h3>{rule.name}</h3>
                <p>{rule.description}</p>
                <div className="rule-details">
                  <div className="detail-item">
                    <span className="label">Kích hoạt:</span>
                    <span className="value">{rule.trigger}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Hành động:</span>
                    <span className="value">{rule.action}</span>
                  </div>
                </div>
              </div>

              <div className="rule-footer">
                <button className="btn-edit">Chỉnh sửa</button>
                <ChevronRight size={18} className="chevron" />
              </div>
            </div>
          ))}

          <div className="add-rule-card">
            <Plus size={32} />
            <span>Thêm kịch bản mới</span>
          </div>
        </div>
      </div>
    </div>
  );
};
