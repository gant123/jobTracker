import React, { useState } from 'react';
import { gmailService } from '../../services/email.service';

const GmailDebug = () => {
  const [results, setResults] = useState({});
  const [loading, setLoading] = useState(false);

  const testEndpoint = async (name, fn) => {
    setLoading(true);
    try {
      const result = await fn();
      setResults(prev => ({
        ...prev,
        [name]: { success: true, data: result }
      }));
    } catch (error) {
      setResults(prev => ({
        ...prev,
        [name]: {
          success: false,
          error: error.message,
          status: error.response?.status,
          data: error.response?.data
        }
      }));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-4 bg-gray-100 rounded-lg">
      <h3 className="font-bold mb-4">Gmail API Debug Panel</h3>

      <div className="space-y-2">
        <button
          onClick={() => testEndpoint('status', gmailService.getStatus)}
          disabled={loading}
          className="btn-secondary text-sm"
        >
          Test Status Endpoint
        </button>

        <button
          onClick={() => testEndpoint('authUrl', gmailService.getAuthUrl)}
          disabled={loading}
          className="btn-secondary text-sm"
        >
          Test Auth URL Endpoint
        </button>

        <button
          onClick={() => testEndpoint('scan', () => gmailService.scanEmails(null, 10))}
          disabled={loading}
          className="btn-secondary text-sm"
        >
          Test Scan Endpoint (10 emails)
        </button>
      </div>

      {Object.keys(results).length > 0 && (
        <div className="mt-4">
          <h4 className="font-semibold mb-2">Results:</h4>
          <pre className="bg-white p-2 rounded text-xs overflow-auto max-h-96">
            {JSON.stringify(results, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
};

export default GmailDebug;
