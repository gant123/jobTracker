import api from './api';

export const gmailService = {
  async getAuthUrl() {
    console.log('[Gmail Service] Getting auth URL...');
    try {
      const response = await api.get('/google/auth-url');
      console.log('[Gmail Service] Auth URL response:', response.data);
      return response.data;
    } catch (error) {
      console.error('[Gmail Service] Auth URL error:', error);
      throw error;
    }
  },

  async getStatus() {
    console.log('[Gmail Service] Getting status...');
    try {
      const response = await api.get('/google/status');
      console.log('[Gmail Service] Status response:', response.data);
      return response.data;
    } catch (error) {
      console.error('[Gmail Service] Status error:', error);
      throw error;
    }
  },

  async disconnect() {
    console.log('[Gmail Service] Disconnecting...');
    try {
      const response = await api.post('/google/disconnect');
      console.log('[Gmail Service] Disconnect response:', response.data);
      return response.data;
    } catch (error) {
      console.error('[Gmail Service] Disconnect error:', error);
      throw error;
    }
  },

  async scanEmails(since = null, until = null, max = 500) {
    console.log('[Gmail Service] Scanning emails...', { since, until, max });

    const params = new URLSearchParams();
    if (since) {
      const sinceDate = new Date(since);
      if (!isNaN(sinceDate.getTime())) {
        params.append('since', sinceDate.toISOString().split('T')[0]);
      }
    }

    if (until) {
      const untilDate = new Date(until);
      if (!isNaN(untilDate.getTime())) {
        params.append('until', untilDate.toISOString().split('T')[0]);
      }
    }
    params.append('limit', max.toString());

    const url = `/google/scan?${params.toString()}`;
    console.log('[Gmail Service] Scan URL:', url);

    try {
      const response = await api.get(url);
      console.log('[Gmail Service] Scan response:', response.data);

      if (response.data && Array.isArray(response.data.events)) {
        return response.data.events;
      }

      console.warn('[Gmail Service] Unexpected response format:', response.data);
      return [];

    } catch (error) {
      console.error('[Gmail Service] Scan error:', error);
      throw error;
    }
  }
};
