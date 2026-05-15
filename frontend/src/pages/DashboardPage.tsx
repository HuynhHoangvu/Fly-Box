import { useState, useEffect } from 'react';
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { pagesAPI, conversationsAPI } from '../services/api';
import './DashboardPage.css';

interface Page {
  id: number;
  platform: string;
  page_name: string;
  connection_status: string;
}

interface Conversation {
  id: number;
  page_id: number;
  customer: {
    name: string;
    platform: string;
  };
  last_message: string;
  unread_count: number;
  updated_at: string;
}

interface Stats {
  totalPages: number;
  totalConversations: number;
  totalUnread: number;
  platformDistribution: { name: string; value: number }[];
  weeklyMessages: { day: string; messages: number }[];
  recentActivity: { time: string; platform: string; action: string }[];
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

export function DashboardPage() {
  const [stats, setStats] = useState<Stats>({
    totalPages: 0,
    totalConversations: 0,
    totalUnread: 0,
    platformDistribution: [],
    weeklyMessages: [],
    recentActivity: [],
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        const [pagesRes, convsRes] = await Promise.all([
          pagesAPI.list(),
          conversationsAPI.list(),
        ]);

        const pages: Page[] = pagesRes.data.data || [];
        const convs: Conversation[] = convsRes.data.data || [];

        // Calculate platform distribution
        const platformCounts: Record<string, number> = {};
        pages.forEach((p) => {
          platformCounts[p.platform] = (platformCounts[p.platform] || 0) + 1;
        });
        const platformDistribution = Object.entries(platformCounts).map(([name, value]) => ({
          name: name.charAt(0).toUpperCase() + name.slice(1),
          value,
        }));

        // Calculate weekly messages (mock data for demo)
        const days = ['T2', 'T3', 'T4', 'T5', 'T6', 'T7', 'CN'];
        const weeklyMessages = days.map((day) => ({
          day,
          messages: Math.floor(Math.random() * 50) + 10,
        }));

        // Recent activity
        const recentActivity = convs.slice(0, 5).map((c) => ({
          time: new Date(c.updated_at).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }),
          platform: c.customer.platform,
          action: c.last_message?.substring(0, 30) + '...' || 'New conversation',
        }));

        setStats({
          totalPages: pages.length,
          totalConversations: convs.length,
          totalUnread: convs.reduce((sum, c) => sum + c.unread_count, 0),
          platformDistribution,
          weeklyMessages,
          recentActivity,
        });
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  if (loading) {
    return <div className="dashboard-loading">Đang tải...</div>;
  }

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <div className="header-info">
          <h1>Tổng quan hệ thống</h1>
          <p>Dữ liệu tổng hợp từ tất cả các kênh kết nối.</p>
        </div>
        <div className="header-actions">
          <button className="btn-secondary">7 ngày qua</button>
          <button className="btn-primary">Xuất báo cáo</button>
        </div>
      </div>

      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-icon blue">📱</div>
          <div className="metric-content">
            <div className="metric-label">Trang đã kết nối</div>
            <div className="metric-value">{stats.totalPages}</div>
            <div className="metric-trend up">↑ 12% so với tháng trước</div>
          </div>
        </div>
        <div className="metric-card">
          <div className="metric-icon green">💬</div>
          <div className="metric-content">
            <div className="metric-label">Tổng hội thoại</div>
            <div className="metric-value">{stats.totalConversations}</div>
            <div className="metric-trend up">↑ 5% so với tháng trước</div>
          </div>
        </div>
        <div className="metric-card highlight">
          <div className="metric-icon yellow">🔔</div>
          <div className="metric-content">
            <div className="metric-label">Tin nhắn chưa đọc</div>
            <div className="metric-value">{stats.totalUnread}</div>
            <div className="metric-trend down">↓ 2% mục tiêu</div>
          </div>
        </div>
        <div className="metric-card">
          <div className="metric-icon purple">⏱️</div>
          <div className="metric-content">
            <div className="metric-label">Phản hồi trung bình</div>
            <div className="metric-value">4.5m</div>
            <div className="metric-trend up">↑ 0.5m nhanh hơn</div>
          </div>
        </div>
      </div>

      <div className="charts-grid">
        <div className="chart-card">
          <div className="chart-header">
            <h3>Lưu lượng tin nhắn</h3>
            <span>Tin nhắn theo ngày</span>
          </div>
          <div className="chart-body">
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={stats.weeklyMessages}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="day" axisLine={false} tickLine={false} tick={{ fill: '#64748b', fontSize: 12 }} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#64748b', fontSize: 12 }} />
                <Tooltip 
                  cursor={{ fill: '#f8fafc' }}
                  contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                />
                <Bar dataKey="messages" fill="var(--primary)" radius={[6, 6, 0, 0]} barSize={32} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="chart-card">
          <div className="chart-header">
            <h3>Nguồn khách hàng</h3>
            <span>Phân bố theo nền tảng</span>
          </div>
          <div className="chart-body">
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={stats.platformDistribution}
                  cx="50%"
                  cy="50%"
                  innerRadius={70}
                  outerRadius={100}
                  paddingAngle={8}
                  dataKey="value"
                >
                  {stats.platformDistribution.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
                <Legend verticalAlign="bottom" height={36} iconType="circle" />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="activity-section card">
        <div className="card-header">
          <h3>Hoạt động gần đây</h3>
          <button className="btn-text">Xem tất cả</button>
        </div>
        <div className="activity-list">
          {stats.recentActivity.length === 0 ? (
            <div className="empty-activity">
              <div className="empty-icon">📂</div>
              <p>Chưa có hoạt động nào được ghi nhận</p>
            </div>
          ) : (
            stats.recentActivity.map((activity, index) => (
              <div key={index} className="activity-item">
                <div className={`platform-tag ${activity.platform.toLowerCase()}`}>
                  {activity.platform}
                </div>
                <div className="activity-details">
                  <span className="activity-action">{activity.action}</span>
                  <span className="activity-time">{activity.time}</span>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}