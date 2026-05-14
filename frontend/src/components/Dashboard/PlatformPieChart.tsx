import React from 'react';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { PlatformDistribution } from '../../types/dashboard';

interface PlatformPieChartProps {
  data: PlatformDistribution[];
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

export const PlatformPieChart: React.FC<PlatformPieChartProps> = ({ data }) => {
  if (data.length === 0) {
    return (
      <div className="chart-empty-state">
        <p>Chưa có dữ liệu nền tảng</p>
      </div>
    );
  }

  return (
    <div className="chart-card">
      <h3>Phân bố nền tảng</h3>
      <ResponsiveContainer width="100%" height={250}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={60}
            outerRadius={85}
            paddingAngle={5}
            dataKey="value"
          >
            {data.map((_, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip 
            formatter={(value: number) => [`${value} trang`, 'Số lượng']}
            contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
          />
          <Legend verticalAlign="bottom" height={36}/>
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
};
