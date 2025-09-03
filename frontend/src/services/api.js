
import axios from 'axios';
import toast from 'react-hot-toast';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8232/api';

const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: attach JWT token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: handle errors globally
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;
    const url = error?.config?.url || '';

    // Special handling for Gmail endpoints
    if (url.includes('/google/') || url.includes('/gmail/')) {
      if (status === 401 || status === 403) {
        // Don't logout the entire app for Gmail auth issues
        // Just return the error for the component to handle
        return Promise.reject(error);
      }
    }

    // Regular 401 handling for non-Gmail endpoints
    if (status === 401 && !url.includes('/google/') && !url.includes('/gmail/')) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      toast.error('Session expired. Please login again.');
      window.location.href = '/login';
      return Promise.reject(error);
    }

    // Show error messages
    if (error.response?.data?.error) {
      toast.error(error.response.data.error);
    } else if (error.response?.data?.message) {
      toast.error(error.response.data.message);
    }

    return Promise.reject(error);
  }
);

export default api;

