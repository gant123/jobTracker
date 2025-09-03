import api from './api';

export const jobsService = {
  async getJobs(filters = {}) {
    const params = new URLSearchParams();
    Object.keys(filters).forEach(key => {
      if (filters[key]) params.append(key, filters[key]);
    });
    const response = await api.get(`/jobs?${params}`);
    return response.data;
  },

  async getJob(id) {
    const response = await api.get(`/jobs/${id}`);
    return response.data;
  },

  async createJob(job) {
    const response = await api.post('/jobs', job);
    return response.data;
  },

  async updateJob(id, job) {
    const response = await api.put(`/jobs/${id}`, job);
    return response.data;
  },

  async deleteJob(id) {
    const response = await api.delete(`/jobs/${id}`);
    return response.data;
  }
};
