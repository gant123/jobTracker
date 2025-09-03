import React from 'react';
import {
  ClipboardDocumentListIcon,
  PaperAirplaneIcon,
  ChatBubbleLeftRightIcon,
  TrophyIcon,
} from '@heroicons/react/24/outline';

const iconMap = {
  clipboard: ClipboardDocumentListIcon,
  'paper-plane': PaperAirplaneIcon,
  comments: ChatBubbleLeftRightIcon,
  trophy: TrophyIcon,
};

const colorMap = {
  purple: 'from-purple-500 to-purple-600',
  blue: 'from-blue-500 to-blue-600',
  yellow: 'from-amber-400 to-amber-500',
  green: 'from-emerald-500 to-emerald-600',
};

const StatsCard = ({ title, value, icon = 'clipboard', color = 'blue' }) => {
  const Icon = iconMap[icon] || ClipboardDocumentListIcon;
  const gradient = colorMap[color] || colorMap.blue;

  return (
    <div className="card p-5">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-500">{title}</p>
          <p className="mt-1 text-3xl font-bold text-gray-900">{value}</p>
        </div>
        <div
          className={`p-3 rounded-lg bg-gradient-to-br ${gradient} text-white`}
        >
          <Icon className="h-7 w-7" />
        </div>
      </div>
    </div>
  );
};

export default StatsCard;
