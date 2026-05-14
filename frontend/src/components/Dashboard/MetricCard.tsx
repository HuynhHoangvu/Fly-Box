import React from 'react';

interface MetricCardProps {
  icon: string;
  value: number;
  label: string;
  highlight?: boolean;
}

export const MetricCard: React.FC<MetricCardProps> = ({ icon, value, label, highlight }) => {
  return (
    <div className={`metric-card ${highlight ? 'highlight' : ''}`}>
      <div className="metric-icon">{icon}</div>
      <div className="metric-content">
        <div className="metric-value">{value}</div>
        <div className="metric-label">{label}</div>
      </div>
    </div>
  );
};
