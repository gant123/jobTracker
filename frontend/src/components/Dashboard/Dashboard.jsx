import React, { useState, useEffect } from 'react';
import { jobsService } from '../../services/jobs.service';
import StatsCard from './StatsCard';
import JobFilters from './JobFilters';
import JobList from '../Jobs/JobList';
import JobModal from '../Jobs/JobModal';
import { PlusIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

import ImportBar from '../Email/ImportBar';
import ImportModal from '../Email/ImportModal';

const Dashboard = () => {
  const [jobs, setJobs] = useState([]);
  const [stats, setStats] = useState({});
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({
    search: '',
    status: '',
    company: '',
    location: '',
  });
  const [showModal, setShowModal] = useState(false);
  const [editingJob, setEditingJob] = useState(null);

  const [importEvents, setImportEvents] = useState([]);
  // NEW: State to control the visibility of the import modal
  const [showImportModal, setShowImportModal] = useState(false);

  useEffect(() => {
    fetchJobs();
  }, [filters]);

  // NEW: Effect to show the modal when there are events to import
  useEffect(() => {
    if (importEvents.length > 0) {
      setShowImportModal(true);
    }
  }, [importEvents]);

  const fetchJobs = async () => {
    setLoading(true);
    try {
      const response = await jobsService.getJobs(filters);
      setJobs(response.jobs || []);
      setStats(response.stats || {});
    } catch (error) {
      console.error('Error fetching jobs:', error);
      toast.error('Failed to load jobs');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateJob = () => {
    setEditingJob(null);
    setShowModal(true);
  };

  const handleEditJob = (job) => {
    setEditingJob(job);
    setShowModal(true);
  };

  const handleDeleteJob = async (id, silent = false) => {
    if (!silent && !window.confirm('Are you sure you want to delete this job?')) {
      return;
    }

    try {
      await jobsService.deleteJob(id);
      if (!silent) {
        toast.success('Job deleted successfully');
      }
      fetchJobs();
    } catch (error) {
      console.error('Error deleting job:', error);
      throw error; // Re-throw for bulk operations to handle
    }
  };

  const handleSaveJob = async (jobData) => {
    try {
      if (editingJob) {
        await jobsService.updateJob(editingJob.id, jobData);
        toast.success('Job updated successfully');
      } else {
        await jobsService.createJob(jobData);
        toast.success('Job created successfully');
      }
      setShowModal(false);
      fetchJobs();
    } catch (error) {
      console.error('Error saving job:', error);
    }
  };

  const handleScanResult = (events) => {
    setImportEvents(events || []);
  };

  const handleImportJobs = async (jobPayloads) => {
    let successCount = 0;
    let failCount = 0;
    const totalToImport = jobPayloads.length;

    const importPromise = async () => {
      for (const payload of jobPayloads) {
        try {
          // The createJob service will now succeed even for duplicates
          await jobsService.createJob(payload);
          successCount++;
        } catch (e) {
          console.error('Import create failed for job:', payload, e);
          failCount++;
        }
      }
    };

    // Use react-hot-toast to show a loading message
    toast.promise(importPromise(), {
      loading: `Importing ${totalToImport} job(s)...`,
      success: () => {
        // This message will show after the import process completes
        fetchJobs(); // Refresh the job list
        setImportEvents([]); // Clear the modal data
        setShowImportModal(false); // Close the modal

        let message = `Import complete! ${successCount} job(s) processed.`;
        if (failCount > 0) {
          message += ` ${failCount} failed.`;
        }
        return message;
      },
      error: 'An unexpected error occurred during import.',
    });
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
        <h1 className="text-3xl font-bold text-gray-900">Job Applications</h1>

        <div className="flex items-center gap-2">
          <ImportBar onScan={handleScanResult} />

          <button
            onClick={handleCreateJob}
            className="btn-primary flex items-center"
          >
            <PlusIcon className="h-5 w-5 mr-2" />
            Add New Job
          </button>
        </div>
      </div>

      {/* Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <StatsCard
          title="Total Applications"
          value={stats.total || 0}
          icon="clipboard"
          color="purple"
        />
        <StatsCard
          title="Applied"
          value={stats.applied || 0}
          icon="paper-plane"
          color="blue"
        />
        <StatsCard
          title="Interviewing"
          value={stats.interviewing || 0}
          icon="comments"
          color="yellow"
        />
        <StatsCard
          title="Offers"
          value={stats.offer || 0}
          icon="trophy"
          color="green"
        />
      </div>

      {/* Filters */}
      <JobFilters filters={filters} setFilters={setFilters} />

      {/* Jobs List */}
      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <JobList
          jobs={jobs}
          onEdit={handleEditJob}
          onDelete={handleDeleteJob}
        />
      )}

      {/* Job Modal */}
      {showModal && (
        <JobModal
          job={editingJob}
          onSave={handleSaveJob}
          onClose={() => setShowModal(false)}
        />
      )}

      {/* MODIFIED: Import Modal - now controlled by its own state */}
      {showImportModal && (
        <ImportModal
          events={importEvents}
          onClose={() => setShowImportModal(false)} // Just hide the modal
          onImport={handleImportJobs}
        />
      )}
    </div>
  );
};

export default Dashboard;
