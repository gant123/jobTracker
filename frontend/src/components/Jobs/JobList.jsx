import React, { useState } from 'react';
import JobCard from './JobCard';
import { TrashIcon, CheckIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const JobList = ({ jobs, onEdit, onDelete }) => {
  const [selectedJobs, setSelectedJobs] = useState(new Set());
  const [isSelecting, setIsSelecting] = useState(false);

  const toggleSelection = (jobId) => {
    setSelectedJobs(prev => {
      const newSet = new Set(prev);
      if (newSet.has(jobId)) {
        newSet.delete(jobId);
      } else {
        newSet.add(jobId);
      }
      return newSet;
    });
  };

  const selectAll = () => {
    if (selectedJobs.size === jobs.length) {
      setSelectedJobs(new Set());
    } else {
      setSelectedJobs(new Set(jobs.map(job => job.id)));
    }
  };

  const handleBulkDelete = async () => {
    if (selectedJobs.size === 0) {
      toast.error('No jobs selected');
      return;
    }

    const confirmMessage = selectedJobs.size === 1
      ? 'Delete 1 selected job?'
      : `Delete ${selectedJobs.size} selected jobs?`;

    if (!window.confirm(confirmMessage)) return;

    let successCount = 0;
    let failCount = 0;

    for (const jobId of selectedJobs) {
      try {
        await onDelete(jobId, true); // true = silent mode (no individual toasts)
        successCount++;
      } catch (error) {
        failCount++;
      }
    }

    if (successCount > 0) {
      toast.success(`Deleted ${successCount} job${successCount > 1 ? 's' : ''}`);
    }
    if (failCount > 0) {
      toast.error(`Failed to delete ${failCount} job${failCount > 1 ? 's' : ''}`);
    }

    setSelectedJobs(new Set());
    setIsSelecting(false);
  };

  const clearSelection = () => {
    setSelectedJobs(new Set());
    setIsSelecting(false);
  };

  if (jobs.length === 0) {
    return (
      <div className="bg-white rounded-xl shadow-sm p-12 text-center">
        <div className="mx-auto w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mb-4">
          <svg className="w-12 h-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
        </div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">No jobs yet</h3>
        <p className="text-gray-500">Start by adding a job or importing from Gmail</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Bulk Actions Bar */}
      {(isSelecting || selectedJobs.size > 0) && (
        <div className="bg-white rounded-lg shadow-sm p-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={selectAll}
              className="flex items-center gap-2 px-3 py-1.5 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition"
            >
              <CheckIcon className="w-4 h-4" />
              {selectedJobs.size === jobs.length ? 'Deselect All' : 'Select All'}
            </button>
            <span className="text-sm text-gray-600">
              {selectedJobs.size} of {jobs.length} selected
            </span>
          </div>

          <div className="flex items-center gap-2">
            {selectedJobs.size > 0 && (
              <button
                onClick={handleBulkDelete}
                className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 transition"
              >
                <TrashIcon className="w-4 h-4" />
                Delete Selected
              </button>
            )}
            <button
              onClick={clearSelection}
              className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Toggle Selection Mode */}
      {!isSelecting && selectedJobs.size === 0 && (
        <div className="flex justify-end">
          <button
            onClick={() => setIsSelecting(true)}
            className="text-sm text-gray-600 hover:text-gray-800 transition"
          >
            Select items
          </button>
        </div>
      )}

      {/* Jobs Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {jobs.map(job => (
          <JobCard
            key={job.id}
            job={job}
            onEdit={onEdit}
            onDelete={onDelete}
            isSelecting={isSelecting || selectedJobs.size > 0}
            isSelected={selectedJobs.has(job.id)}
            onToggleSelect={() => toggleSelection(job.id)}
          />
        ))}
      </div>
    </div>
  );
};

export default JobList;
