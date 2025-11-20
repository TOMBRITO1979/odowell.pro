import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Handle responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  register: (data) => api.post('/auth/register', data),
  getMe: () => api.get('/auth/me'),
  createTenant: (data) => api.post('/tenants', data),
  updateProfile: (data) => api.put('/auth/profile', data),
  changePassword: (data) => api.put('/auth/password', data),
};

// Patients API
export const patientsAPI = {
  getAll: (params) => api.get('/patients', { params }),
  getOne: (id) => api.get(`/patients/${id}`),
  create: (data) => api.post('/patients', data),
  update: (id, data) => api.put(`/patients/${id}`, data),
  delete: (id) => api.delete(`/patients/${id}`),
  getStats: (id) => api.get(`/patients/${id}/stats`),
};

// Appointments API
export const appointmentsAPI = {
  getAll: (params) => api.get('/appointments', { params }),
  getOne: (id) => api.get(`/appointments/${id}`),
  create: (data) => api.post('/appointments', data),
  update: (id, data) => api.put(`/appointments/${id}`, data),
  delete: (id) => api.delete(`/appointments/${id}`),
  updateStatus: (id, status) => api.patch(`/appointments/${id}/status`, { status }),
};

// Medical Records API
export const medicalRecordsAPI = {
  getAll: (params) => api.get('/medical-records', { params }),
  getOne: (id) => api.get(`/medical-records/${id}`),
  create: (data) => api.post('/medical-records', data),
  update: (id, data) => api.put(`/medical-records/${id}`, data),
  delete: (id) => api.delete(`/medical-records/${id}`),
  downloadPDF: (id) => api.get(`/medical-records/${id}/pdf`, { responseType: 'blob' }),
};

// Budgets API
export const budgetsAPI = {
  getAll: (params) => api.get('/budgets', { params }),
  getOne: (id) => api.get(`/budgets/${id}`),
  create: (data) => api.post('/budgets', data),
  update: (id, data) => api.put(`/budgets/${id}`, data),
  delete: (id) => api.delete(`/budgets/${id}`),
  downloadPDF: (id) => api.get(`/budgets/${id}/pdf`, { responseType: 'blob' }),
};

// Payments API
export const paymentsAPI = {
  getAll: (params) => api.get('/payments', { params }),
  getOne: (id) => api.get(`/payments/${id}`),
  create: (data) => api.post('/payments', data),
  update: (id, data) => api.put(`/payments/${id}`, data),
  delete: (id) => api.delete(`/payments/${id}`),
  getCashFlow: (params) => api.get('/payments/cashflow', { params }),
  downloadPDF: (params) => api.get('/payments/pdf/export', { params, responseType: 'blob' }),
  downloadReceipt: (budgetId, paymentId) => api.get(`/budgets/${budgetId}/payment/${paymentId}/receipt`, { responseType: 'blob' }),
};

// Products API
export const productsAPI = {
  getAll: (params) => api.get('/products', { params }),
  getOne: (id) => api.get(`/products/${id}`),
  create: (data) => api.post('/products', data),
  update: (id, data) => api.put(`/products/${id}`, data),
  delete: (id) => api.delete(`/products/${id}`),
  getLowStock: () => api.get('/products/low-stock'),
};

// Suppliers API
export const suppliersAPI = {
  getAll: () => api.get('/suppliers'),
  getOne: (id) => api.get(`/suppliers/${id}`),
  create: (data) => api.post('/suppliers', data),
  update: (id, data) => api.put(`/suppliers/${id}`, data),
  delete: (id) => api.delete(`/suppliers/${id}`),
};

// Stock Movements API
export const stockMovementsAPI = {
  getAll: (params) => api.get('/stock-movements', { params }),
  create: (data) => api.post('/stock-movements', data),
};

// Campaigns API
export const campaignsAPI = {
  getAll: (params) => api.get('/campaigns', { params }),
  getOne: (id) => api.get(`/campaigns/${id}`),
  create: (data) => api.post('/campaigns', data),
  update: (id, data) => api.put(`/campaigns/${id}`, data),
  delete: (id) => api.delete(`/campaigns/${id}`),
  send: (id) => api.post(`/campaigns/${id}/send`),
};

// Reports API
export const reportsAPI = {
  getDashboard: () => api.get('/reports/dashboard'),
  getRevenue: (params) => api.get('/reports/revenue', { params }),
  getProcedures: () => api.get('/reports/procedures'),
  getAttendance: (params) => api.get('/reports/attendance', { params }),
  downloadRevenuePDF: (params) => api.get('/reports/revenue/pdf', { params, responseType: 'blob' }),
  downloadAttendancePDF: (params) => api.get('/reports/attendance/pdf', { params, responseType: 'blob' }),
  downloadProceduresPDF: () => api.get('/reports/procedures/pdf', { responseType: 'blob' }),
  downloadRevenueExcel: (params) => api.get('/reports/revenue/excel', { params, responseType: 'blob' }),
  downloadAttendanceExcel: (params) => api.get('/reports/attendance/excel', { params, responseType: 'blob' }),
  downloadProceduresExcel: () => api.get('/reports/procedures/excel', { responseType: 'blob' }),
};

// Attachments API
export const attachmentsAPI = {
  upload: (formData) => api.post('/attachments', formData, {
    headers: { 'Content-Type': undefined }, // Let axios set the correct multipart boundary
  }),
  getOne: (id) => api.get(`/attachments/${id}`),
  download: (id) => api.get(`/attachments/${id}?download=true`, {
    responseType: 'blob',
  }),
  delete: (id) => api.delete(`/attachments/${id}`),
};

// Settings API
export const settingsAPI = {
  get: () => api.get('/settings'),
  update: (data) => api.put('/settings', data),
};

// Exams API
export const examsAPI = {
  getAll: (params) => api.get('/exams', { params }),
  getOne: (id) => api.get(`/exams/${id}`),
  create: (formData) => api.post('/exams', formData, {
    headers: { 'Content-Type': undefined }, // Let axios set the correct multipart boundary
  }),
  update: (id, data) => api.put(`/exams/${id}`, data),
  delete: (id) => api.delete(`/exams/${id}`),
  getDownloadURL: (id) => api.get(`/exams/${id}/download`),
};

// Prescriptions API
export const prescriptionsAPI = {
  getAll: (params) => api.get('/prescriptions', { params }),
  getOne: (id) => api.get(`/prescriptions/${id}`),
  create: (data) => api.post('/prescriptions', data),
  update: (id, data) => api.put(`/prescriptions/${id}`, data),
  delete: (id) => api.delete(`/prescriptions/${id}`),
  issue: (id) => api.post(`/prescriptions/${id}/issue`),
  print: (id) => api.post(`/prescriptions/${id}/print`),
  downloadPDF: (id) => api.get(`/prescriptions/${id}/pdf`, { responseType: 'blob' }),
};
