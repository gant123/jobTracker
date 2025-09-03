import React, { useState, useEffect } from 'react';
import { gmailService } from '../../services/email.service';
import { EnvelopeIcon, ArrowPathIcon, LinkIcon, CheckCircleIcon, CalendarIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import { jobsService } from '../../services/jobs.service';

const ImportBar = ({ onScan }) => {
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [error, setError] = useState("");
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [dateRange, setDateRange] = useState({
    from: new Date(new Date().setMonth(new Date().getMonth() - 1)).toISOString().split('T')[0],
    to: new Date().toISOString().split('T')[0]
  });
  useEffect(() => {
    checkStatus();

    const handleMessage = (event) => {
      if (event.data?.type === 'gmail-auth-complete') {
        setTimeout(() => {
          checkStatus();
          if (event.data.success) {
            toast.success('Gmail connected successfully!');
          }
        }, 1000);
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, []);

  const checkStatus = async () => {
    try {
      const response = await gmailService.getStatus();
      setStatus(response);
    } catch (error) {
      setStatus({ connected: false });
    }
  };

  const handleConnect = async () => {
    try {
      setLoading(true);
      const response = await gmailService.getAuthUrl();

      if (response.url) {
        const width = 600;
        const height = 700;
        const left = window.screen.width / 2 - width / 2;
        const top = window.screen.height / 2 - height / 2;

        const authWindow = window.open(
          response.url,
          'gmail-auth',
          `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no`
        );

        const checkInterval = setInterval(async () => {
          try {
            if (authWindow && authWindow.closed) {
              clearInterval(checkInterval);
              setLoading(false);
              await checkStatus();
            }
          } catch (e) {
            clearInterval(checkInterval);
            setLoading(false);
          }
        }, 1000);

        setTimeout(() => {
          clearInterval(checkInterval);
          setLoading(false);
        }, 300000);
      }
    } catch (error) {
      console.error('Gmail connect error:', error);
      toast.error('Failed to connect Gmail');
      setLoading(false);
    }
  };

  const handleScan = async () => {
    setScanning(true);
    setError(null);

    try {
      const statusCheck = await gmailService.getStatus();
      if (!statusCheck.connected) {
        throw new Error('Gmail not connected. Please reconnect.');
      }

      console.log(`Scanning emails from ${dateRange.from} to ${dateRange.to}`);

      // Scan emails using both 'from' and 'to' dates
      const events = await gmailService.scanEmails(dateRange.from, dateRange.to, 500);

      // Get existing jobs to check for duplicates
      console.log('Checking for duplicates...');
      const existingJobsResponse = await jobsService.getJobs();
      const existingJobs = existingJobsResponse.jobs || [];

      const existingGmailIds = new Set(
        existingJobs
          .filter(job => job.gmail_message_id)
          .map(job => job.gmail_message_id)
      );

      const newEvents = events.filter(e => !existingGmailIds.has(e.messageId));
      const duplicateCount = events.length - newEvents.length;

      console.log(`Found ${events.length} total, ${newEvents.length} new, ${duplicateCount} already imported`);

      if (newEvents.length > 0) {
        toast.success(
          duplicateCount > 0
            ? `Found ${newEvents.length} new emails (${duplicateCount} already imported)`
            : `Found ${newEvents.length} job-related emails`
        );
        onScan(newEvents);
      } else if (duplicateCount > 0) {
        toast.info(`All ${events.length} emails were already imported`);
      } else {
        toast.info('No job-related emails found in this date range');
      }

      setShowDatePicker(false);
    } catch (error) {
      console.error('Scan error:', error);
      let errorMessage = 'Failed to scan emails';

      if (error.response?.status === 401 || error.response?.status === 403) {
        errorMessage = 'Gmail authorization expired. Please reconnect.';
        setStatus({ connected: false });
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      } else if (error.message) {
        errorMessage = error.message;
      }

      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setScanning(false);
    }
  };

  const handleDisconnect = async () => {
    if (!window.confirm('Disconnect Gmail? You can reconnect anytime.')) {
      return;
    }

    try {
      await gmailService.disconnect();
      setStatus({ connected: false });
      toast.success('Gmail disconnected');
    } catch (error) {
      console.error('Disconnect error:', error);
      setStatus({ connected: false });
    }
  };

  if (!status?.connected) {
    return (
      <button
        onClick={handleConnect}
        disabled={loading}
        className="btn-secondary flex items-center gap-2"
      >
        {loading ? (
          <>
            <ArrowPathIcon className="h-4 w-4 animate-spin" />
            <span>Connecting...</span>
          </>
        ) : (
          <>
            <LinkIcon className="h-4 w-4" />
            <span>Connect Gmail</span>
          </>
        )}
      </button>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <div className="flex items-center gap-1 text-sm text-green-600 bg-green-50 px-2 py-1 rounded">
        <CheckCircleIcon className="h-4 w-4" />
        <span>Gmail</span>
        {status.email && (
          <span className="text-xs text-gray-600 ml-1">({status.email})</span>
        )}
      </div>

      {showDatePicker ? (
        <div className="flex items-center gap-2 bg-white border rounded-lg px-3 py-1">
          <CalendarIcon className="h-4 w-4 text-gray-500" />
          <input
            type="date"
            value={dateRange.from}
            max={new Date().toISOString().split('T')[0]}
            onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
            className="text-sm border-0 focus:outline-none"
          />
          <span className="text-gray-500">to</span>
          <input
            type="date"
            value={dateRange.to}
            max={new Date().toISOString().split('T')[0]}
            onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
            className="text-sm border-0 focus:outline-none"
          />
          <button
            onClick={handleScan}
            disabled={scanning}
            className="btn-primary text-sm px-3 py-1"
          >
            {scanning ? 'Scanning...' : 'Scan'}
          </button>
          <button
            onClick={() => setShowDatePicker(false)}
            className="text-gray-500 hover:text-gray-700"
          >
            Cancel
          </button>
        </div>
      ) : (
        <>
          <button
            onClick={() => setShowDatePicker(true)}
            disabled={scanning}
            className="btn-secondary flex items-center gap-2"
          >
            <EnvelopeIcon className="h-4 w-4" />
            <span>Import from Email</span>
          </button>

          <button
            onClick={handleDisconnect}
            className="text-gray-500 hover:text-gray-700 p-1"
            title="Disconnect Gmail"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </>
      )}
    </div>
  );
};

export default ImportBar;
