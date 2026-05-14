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
      <h1>📊 Dashboard</h1>

      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-icon">📱</div>
          <div className="metric-content">
            <div className="metric-value">{stats.totalPages}</div>
            <div className="metric-label">Trang đã kết nối</div>
          </div>
        </div>
        <div className="metric-card">
          <div className="metric-icon">💬</div>
          <div className="metric-content">
            <div className="metric-value">{stats.totalConversations}</div>
            <div className="metric-label">Hội thoại</div>
          </div>
        </div>
        <div className="metric-card highlight">
          <div className="metric-icon">🔔</div>
          <div className="metric-content">
            <div className="metric-value">{stats.totalUnread}</div>
            <div className="metric-label">Tin nhắn chưa đọc</div>
          </div>
        </div>
      </div>

      <div className="charts-grid">
        <div className="chart-card">
          <h3>Tin nhắn theo ngày</h3>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={stats.weeklyMessages}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="messages" fill="#3b82f6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-card">
          <h3>Phân bố nền tảng</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={stats.platformDistribution}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={90}
                paddingAngle={2}
                dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              >
                {stats.platformDistribution.map((_, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="activity-section">
        <h3>Hoạt động gần đây</h3>
        <div className="activity-list">
          {stats.recentActivity.length === 0 ? (
            <div className="empty-activity">Chưa có hoạt động nào</div>
          ) : (
            stats.recentActivity.map((activity, index) => (
              <div key={index} className="activity-item">
                <span className="activity-platform">{activity.platform}</span>
                <span className="activity-action">{activity.action}</span>
                <span className="activity-time">{activity.time}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}