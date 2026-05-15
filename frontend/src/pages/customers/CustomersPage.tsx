import React, { useState } from 'react';
import { Search, Filter, MoreVertical, MessageSquare, Phone, Mail, ShieldCheck, UserPlus } from 'lucide-react';
import './CustomersPage.css';

interface Customer {
  id: string;
  name: string;
  avatar?: string;
  platform: 'facebook' | 'instagram' | 'zalo' | 'shopee' | 'tiktok';
  lastInteraction: string;
  tags: string[];
  status: 'active' | 'potential' | 'blocked';
  email?: string;
  phone?: string;
}

export const CustomersPage: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [customers] = useState<Customer[]>([
    {
      id: '1',
      name: 'Nguyễn Văn A',
      platform: 'facebook',
      lastInteraction: '2 giờ trước',
      tags: ['VIP', 'Thanh toán'],
      status: 'active',
      phone: '0901234567',
      email: 'vana@gmail.com'
    },
    {
      id: '2',
      name: 'Trần Thị B',
      platform: 'instagram',
      lastInteraction: '1 ngày trước',
      tags: ['Tiềm năng'],
      status: 'potential',
      phone: '0987654321'
    },
    {
      id: '3',
      name: 'Lê Văn C',
      platform: 'facebook',
      lastInteraction: '5 phút trước',
      tags: ['Mới'],
      status: 'active',
      email: 'vanc@outlook.com'
    },
    {
      id: '4',
      name: 'Phạm Minh D',
      platform: 'tiktok',
      lastInteraction: '3 ngày trước',
      tags: ['Đã mua'],
      status: 'active'
    }
  ]);

  return (
    <div className="customers-page">
      <div className="customers-header">
        <div className="header-left">
          <h1>Khách hàng</h1>
          <span className="customer-count">1,248 khách hàng</span>
        </div>
        <div className="header-actions">
          <button className="btn-secondary">
            <Filter size={18} />
            Bộ lọc
          </button>
          <button className="btn-primary">
            <UserPlus size={18} />
            Thêm khách hàng
          </button>
        </div>
      </div>

      <div className="search-bar-container">
        <div className="search-input-wrapper">
          <Search size={18} className="search-icon" />
          <input 
            type="text" 
            placeholder="Tìm kiếm theo tên, số điện thoại, email..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div className="customers-table-container">
        <table className="customers-table">
          <thead>
            <tr>
              <th><input type="checkbox" /></th>
              <th>Khách hàng</th>
              <th>Nền tảng</th>
              <th>Thông tin liên hệ</th>
              <th>Nhãn (Tags)</th>
              <th>Tương tác cuối</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {customers.map(customer => (
              <tr key={customer.id}>
                <td><input type="checkbox" /></td>
                <td>
                  <div className="customer-cell">
                    <div className="customer-avatar">
                      {customer.avatar ? <img src={customer.avatar} alt="" /> : customer.name.charAt(0)}
                    </div>
                    <div className="customer-info">
                      <span className="name">{customer.name}</span>
                      <span className={`status-badge ${customer.status}`}>
                        {customer.status === 'active' && 'Đang hoạt động'}
                        {customer.status === 'potential' && 'Tiềm năng'}
                        {customer.status === 'blocked' && 'Đã chặn'}
                      </span>
                    </div>
                  </div>
                </td>
                <td>
                  <div className="platform-cell">
                    {customer.platform === 'facebook' && <span className="platform-icon">📘</span>}
                    {customer.platform === 'instagram' && <span className="platform-icon">📷</span>}
                    {customer.platform === 'zalo' && <span className="platform-icon">💬</span>}
                    {customer.platform === 'shopee' && <span className="platform-icon">🛒</span>}
                    {customer.platform === 'tiktok' && <span className="platform-icon">🎵</span>}
                    <span className="platform-name">{customer.platform}</span>
                  </div>
                </td>
                <td>
                  <div className="contact-cell">
                    {customer.phone && (
                      <div className="contact-item" title={customer.phone}>
                        <Phone size={14} /> {customer.phone}
                      </div>
                    )}
                    {customer.email && (
                      <div className="contact-item" title={customer.email}>
                        <Mail size={14} /> {customer.email}
                      </div>
                    )}
                  </div>
                </td>
                <td>
                  <div className="tags-cell">
                    {customer.tags.map((tag, i) => (
                      <span key={i} className="tag">{tag}</span>
                    ))}
                  </div>
                </td>
                <td>{customer.lastInteraction}</td>
                <td>
                  <div className="row-actions">
                    <button className="action-btn"><MessageSquare size={16} /></button>
                    <button className="action-btn"><MoreVertical size={16} /></button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="pagination">
        <span className="pagination-info">Hiển thị 1-10 của 1,248</span>
        <div className="pagination-controls">
          <button disabled>Trước</button>
          <button className="active">1</button>
          <button>2</button>
          <button>3</button>
          <span>...</span>
          <button>Tiếp</button>
        </div>
      </div>
    </div>
  );
};
