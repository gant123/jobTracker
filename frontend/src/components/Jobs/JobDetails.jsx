
import React, { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { jobsService } from '../../services/jobs.service';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const label = (s) => s ? s.charAt(0).toUpperCase() + s.slice(1) : '';

const JobDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [job, setJob] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const data = await jobsService.getJob(id);
        setJob(data.job || data); // support either {job} or raw job
      } catch (e) {
        toast.error(e.message);
        navigate('/dashboard');
      } finally {
        setLoading(false);
      }
    })();
  }, [id, navigate]);

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (!job) return null;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <button onClick={() => navigate(-1)} className="btn-secondary flex items-center gap-2">
          <ArrowLeftIcon className="h-4 w-4" /> Back
        </button>
        {job.link && (
          <a className="btn-primary" href={job.link} target="_blank" rel="noreferrer">
            View Posting
          </a>
        )}
      </div>

      <div className="card p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold">{job.title}</h1>
            <p className="text-gray-600">{job.company}</p>
          </div>
          <span className="status-badge bg-gray-100 text-gray-800">
            {label(job.status)}
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-6 text-sm">
          {job.location && <p><span className="font-medium">Location:</span> {job.location}</p>}
          {job.salary && <p><span className="font-medium">Salary:</span> {job.salary}</p>}
          {job.appliedDate && (
            <p>
              <span className="font-medium">Applied:</span>{' '}
              {new Date(job.appliedDate).toLocaleDateString()}
            </p>
          )}
        </div>

        {job.notes && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-700">Notes</h3>
            <p className="mt-1 text-gray-700 whitespace-pre-line">{job.notes}</p>
          </div>
        )}
      </div>

      <div className="text-sm text-gray-500">
        <Link to="/dashboard" className="hover:text-primary-600">Go to Dashboard</Link>
      </div>
    </div>
  );
};

export default JobDetails;
