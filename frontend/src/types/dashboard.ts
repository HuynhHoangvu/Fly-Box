export interface MetricData {
  totalPages: number;
  totalConversations: number;
  totalUnread: number;
}

export interface PlatformDistribution {
  name: string;
  value: number;
}

export interface WeeklyMessageData {
  day: string;
  messages: number;
}

export interface RecentActivityData {
  id: string | number;
  time: string;
  platform: string;
  action: string;
}

export interface DashboardStats extends MetricData {
  platformDistribution: PlatformDistribution[];
  weeklyMessages: WeeklyMessageData[];
  recentActivity: RecentActivityData[];
}
