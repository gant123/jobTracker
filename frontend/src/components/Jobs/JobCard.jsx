import React from 'react';
import {
  PencilIcon,
  TrashIcon,
  CalendarIcon,
  MapPinIcon,
  LinkIcon,
  BuildingOfficeIcon,
  BriefcaseIcon
} from '@heroicons/react/24/outline';
import { format } from 'date-fns';

const JobCard = ({ job, onEdit, onDelete, isSelecting, isSelected, onToggleSelect }) => {
  const statusColors = {
    wishlist: 'bg-gray-100 text-gray-700 border-gray-300',
    applied: 'bg-blue-100 text-blue-700 border-blue-300',
    interviewing: 'bg-yellow-100 text-yellow-700 border-yellow-300',
    offer: 'bg-green-100 text-green-700 border-green-300',
    rejected: 'bg-red-100 text-red-700 border-red-300',
    withdrawn: 'bg-gray-100 text-gray-600 border-gray-300'
  };

  const statusIcons = {
    wishlist: '⭐',
    applied: '📨',
    interviewing: '💬',
    offer: '🎉',
    rejected: '❌',
    withdrawn: '🚫'
  };

  const handleCardClick = (e) => {
    if (isSelecting && !e.target.closest('.action-button')) {
      e.preventDefault();
      onToggleSelect();
    }
  };

  const handleDelete = (e) => {
    e.stopPropagation();
    if (window.confirm(`Delete job at ${job.company}?`)) {
      onDelete(job.id);
    }
  };

  const handleEdit = (e) => {
    e.stopPropagation();
    onEdit(job);
  };

  return (
    <div
      onClick={handleCardClick}
      className={`
        bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-200 
        border-2 overflow-hidden cursor-pointer transform hover:-translate-y-1
        ${isSelected ? 'border-primary-500 bg-primary-50' : 'border-transparent'}
        ${isSelecting ? 'cursor-pointer' : ''}
      `}
    >
      {/* Card Header */}
      <div className="p-5">
        <div className="flex justify-between items-start mb-3">
          <div className="flex-1">
            <h3 className="font-semibold text-lg text-gray-900 mb-1 line-clamp-1">
              {job.position || 'Untitled Position'}
            </h3>
            <div className="flex items-center gap-2 text-gray-600">
              <BuildingOfficeIcon className="w-4 h-4" />
              <span className="text-sm font-medium">{job.company}</span>
            </div>
          </div>

          {/* Selection Checkbox */}
          {isSelecting && (
            <input
              type="checkbox"
              checked={isSelected}
              onChange={onToggleSelect}
              className="w-5 h-5 text-primary-600 rounded focus:ring-primary-500 mt-1"
              onClick={(e) => e.stopPropagation()}
            />
          )}
        </div>

        {/* Status Badge */}
        <div className="mb-3">
          <span className={`
            inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold
            border ${statusColors[job.status]}
          `}>
            <span>{statusIcons[job.status]}</span>
            {job.status.charAt(0).toUpperCase() + job.status.slice(1)}
          </span>
        </div>

        {/* Job Details */}
        <div className="space-y-2 text-sm text-gray-600">
          {job.location && (
            <div className="flex items-center gap-2">
              <MapPinIcon className="w-4 h-4 text-gray-400" />
              <span className="line-clamp-1">{job.location}</span>
            </div>
          )}

          {job.applied_date && (
            <div className="flex items-center gap-2">
              <CalendarIcon className="w-4 h-4 text-gray-400" />
              <span>Applied {format(new Date(job.applied_date), 'MMM d, yyyy')}</span>
            </div>
          )}

          {job.job_type && (
            <div className="flex items-center gap-2">
              <BriefcaseIcon className="w-4 h-4 text-gray-400" />
              <span>{job.job_type}</span>
            </div>
          )}

          {job.url && (
            <div className="flex items-center gap-2">
              <LinkIcon className="w-4 h-4 text-gray-400" />
              <a
                href={job.url}
                target="_blank"
                rel="noreferrer"
                className="text-primary-600 hover:text-primary-700 underline"
                onClick={(e) => e.stopPropagation()}
              >
                View Posting
              </a>
            </div>
          )}
        </div>

        {/* Notes Preview */}
        {job.notes && (
          <div className="mt-3 p-2 bg-gray-50 rounded-lg">
            <p className="text-xs text-gray-600 line-clamp-2">{job.notes}</p>
          </div>
        )}
      </div>

      {/* Card Footer - Action Buttons */}
      {!isSelecting && (
        <div className="px-5 py-3 bg-gray-50 border-t border-gray-100 flex justify-end gap-2">
          <button
            onClick={handleEdit}
            className="action-button p-2 text-gray-600 hover:text-primary-600 hover:bg-white rounded-lg transition"
            title="Edit"
          >
            <PencilIcon className="w-4 h-4" />
          </button>
          <button
            onClick={handleDelete}
            className="action-button p-2 text-gray-600 hover:text-red-600 hover:bg-white rounded-lg transition"
            title="Delete"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      )}
    </div>
  );
};

export default JobCard;
