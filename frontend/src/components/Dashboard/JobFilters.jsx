import React from 'react';
import { MagnifyingGlassIcon, FunnelIcon } from '@heroicons/react/24/outline';
import { JOB_STATUSES } from '../../utils/constants';

const JobFilters = ({ filters, setFilters }) => {
  const set = (key) => (e) => setFilters((f) => ({ ...f, [key]: e.target.value }));

  return (
    <div className="card p-4">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="relative">
          <MagnifyingGlassIcon className="h-5 w-5 text-gray-400 absolute left-3 top-3" />
          <input
            type="text"
            placeholder="Search role/title"
            className="input-field pl-10"
            value={filters.search}
            onChange={set('search')}
          />
        </div>
        <div className="relative">
          <FunnelIcon className="h-5 w-5 text-gray-400 absolute left-3 top-3" />
          <select
            className="input-field pl-10"
            value={filters.status}
            onChange={set('status')}
          >
            <option value="">All statuses</option>
            {JOB_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s.charAt(0).toUpperCase() + s.slice(1)}
              </option>
            ))}
          </select>
        </div>
        <input
          className="input-field"
          placeholder="Company"
          value={filters.company}
          onChange={set('company')}
        />
        <input
          className="input-field"
          placeholder="Location"
          value={filters.location}
          onChange={set('location')}
        />
      </div>
    </div>
  );
};

export default JobFilters;
