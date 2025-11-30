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
  uploadProfilePicture: (formData) => api.post('/auth/profile/picture', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Patients API
export const patientsAPI = {
  getAll: (params) => api.get('/patients', { params }),
  getOne: (id) => api.get(`/patients/${id}`),
  create: (data) => api.post('/patients', data),
  update: (id, data) => api.put(`/patients/${id}`, data),
  delete: (id) => api.delete(`/patients/${id}`),
  getStats: (id) => api.get(`/patients/${id}/stats`),
  exportCSV: (params) => api.get(`/patients/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/patients/export/pdf?${params}`, { responseType: 'blob' }),
  importCSV: (formData) => api.post('/patients/import/csv', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Appointments API
export const appointmentsAPI = {
  getAll: (params) => api.get('/appointments', { params }),
  getOne: (id) => api.get(`/appointments/${id}`),
  create: (data) => api.post('/appointments', data),
  update: (id, data) => api.put(`/appointments/${id}`, data),
  delete: (id) => api.delete(`/appointments/${id}`),
  updateStatus: (id, status) => api.patch(`/appointments/${id}/status`, { status }),
  exportCSV: (params) => api.get(`/appointments/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/appointments/export/pdf?${params}`, { responseType: 'blob' }),
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
  cancel: (id) => api.post(`/budgets/${id}/cancel`),
  downloadPDF: (id) => api.get(`/budgets/${id}/pdf`, { responseType: 'blob' }),
  downloadPaymentsPDF: (id) => api.get(`/budgets/${id}/payments-pdf`, { responseType: 'blob' }),
  exportCSV: (params) => api.get(`/budgets/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/budgets/export/pdf?${params}`, { responseType: 'blob' }),
  importCSV: (formData) => api.post('/budgets/import/csv', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Payments API
export const paymentsAPI = {
  getAll: (params) => api.get('/payments', { params }),
  getOne: (id) => api.get(`/payments/${id}`),
  create: (data) => api.post('/payments', data),
  update: (id, data) => api.put(`/payments/${id}`, data),
  delete: (id) => api.delete(`/payments/${id}`),
  refund: (id, reason) => api.post(`/payments/${id}/refund`, { reason }),
  getCashFlow: (params) => api.get('/payments/cashflow', { params }),
  downloadPDF: (params) => api.get('/payments/pdf/export', { params, responseType: 'blob' }),
  downloadReceipt: (budgetId, paymentId) => api.get(`/budgets/${budgetId}/payment/${paymentId}/receipt`, { responseType: 'blob' }),
  exportCSV: (params) => api.get(`/payments/export/csv?${params}`, { responseType: 'blob' }),
  importCSV: (formData) => api.post('/payments/import/csv', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Products API
export const productsAPI = {
  getAll: (params) => api.get('/products', { params }),
  getOne: (id) => api.get(`/products/${id}`),
  create: (data) => api.post('/products', data),
  update: (id, data) => api.put(`/products/${id}`, data),
  delete: (id) => api.delete(`/products/${id}`),
  getLowStock: () => api.get('/products/low-stock'),
  exportCSV: (params) => api.get(`/products/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/products/export/pdf?${params}`, { responseType: 'blob' }),
  importCSV: (formData) => api.post('/products/import/csv', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Suppliers API
export const suppliersAPI = {
  getAll: () => api.get('/suppliers'),
  getOne: (id) => api.get(`/suppliers/${id}`),
  create: (data) => api.post('/suppliers', data),
  update: (id, data) => api.put(`/suppliers/${id}`, data),
  delete: (id) => api.delete(`/suppliers/${id}`),
  exportCSV: (params) => api.get(`/suppliers/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/suppliers/export/pdf?${params}`, { responseType: 'blob' }),
  importCSV: (formData) => api.post('/suppliers/import/csv', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
};

// Stock Movements API
export const stockMovementsAPI = {
  getAll: (params) => api.get('/stock-movements', { params }),
  create: (data) => api.post('/stock-movements', data),
  exportCSV: (params) => api.get(`/stock-movements/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/stock-movements/export/pdf?${params}`, { responseType: 'blob' }),
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
  getAdvancedDashboard: (params) => api.get('/reports/dashboard/advanced', { params }),
  downloadDashboardPDF: (params) => api.get('/reports/dashboard/pdf', { params, responseType: 'blob' }),
  getRevenue: (params) => api.get('/reports/revenue', { params }),
  getProcedures: () => api.get('/reports/procedures'),
  getAttendance: (params) => api.get('/reports/attendance', { params }),
  getBudgetConversion: (params) => api.get('/reports/budget-conversion', { params }),
  getOverduePayments: () => api.get('/reports/overdue-payments'),
  downloadRevenuePDF: (params) => api.get('/reports/revenue/pdf', { params, responseType: 'blob' }),
  downloadAttendancePDF: (params) => api.get('/reports/attendance/pdf', { params, responseType: 'blob' }),
  downloadProceduresPDF: () => api.get('/reports/procedures/pdf', { responseType: 'blob' }),
  downloadRevenueExcel: (params) => api.get('/reports/revenue/excel', { params, responseType: 'blob' }),
  downloadAttendanceExcel: (params) => api.get('/reports/attendance/excel', { params, responseType: 'blob' }),
  downloadProceduresExcel: () => api.get('/reports/procedures/excel', { responseType: 'blob' }),
  downloadBudgetConversionPDF: (params) => api.get('/reports/budget-conversion/pdf', { params, responseType: 'blob' }),
  downloadBudgetConversionExcel: (params) => api.get('/reports/budget-conversion/excel', { params, responseType: 'blob' }),
  downloadOverduePaymentsPDF: () => api.get('/reports/overdue-payments/pdf', { responseType: 'blob' }),
  downloadOverduePaymentsExcel: () => api.get('/reports/overdue-payments/excel', { responseType: 'blob' }),
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
  // API Key Management
  getAPIKeyStatus: () => api.get('/settings/api-key/status'),
  generateAPIKey: () => api.post('/settings/api-key/generate'),
  toggleAPIKey: (active) => api.patch('/settings/api-key/toggle', { active }),
  revokeAPIKey: () => api.delete('/settings/api-key'),
  getAPIKeyDocs: () => api.get('/settings/api-key/docs'),
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

// Users API
export const usersAPI = {
  getAll: () => api.get('/users'),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  updateSidebar: (id, hideSidebar) => api.patch(`/users/${id}/sidebar`, { hide_sidebar: hideSidebar }),
};

// Permissions API
export const permissionsAPI = {
  getModules: () => api.get('/permissions/modules'),
  getAllPermissions: () => api.get('/permissions/all'),
  getUserPermissions: (userId) => api.get(`/permissions/users/${userId}`),
  updateUserPermissions: (userId, permissions) => api.put(`/permissions/users/${userId}`, { permissions }),
  bulkUpdateUserPermissions: (userId, permissionIds) => api.post(`/permissions/users/${userId}/bulk`, { permission_ids: permissionIds }),
  getDefaultRolePermissions: (role) => api.get(`/permissions/defaults/${role}`),
};

// Tasks API
export const tasksAPI = {
  getAll: (params) => api.get('/tasks', { params }),
  getOne: (id) => api.get(`/tasks/${id}`),
  create: (data) => api.post('/tasks', data),
  update: (id, data) => api.put(`/tasks/${id}`, data),
  delete: (id) => api.delete(`/tasks/${id}`),
  getPendingCount: () => api.get('/tasks/pending-count'),
};

// Waiting List API
export const waitingListAPI = {
  getAll: (params) => api.get('/waiting-list', { params }),
  getOne: (id) => api.get(`/waiting-list/${id}`),
  create: (data) => api.post('/waiting-list', data),
  update: (id, data) => api.put(`/waiting-list/${id}`, data),
  delete: (id) => api.delete(`/waiting-list/${id}`),
  contact: (id) => api.post(`/waiting-list/${id}/contact`),
  schedule: (id, appointmentId) => api.post(`/waiting-list/${id}/schedule`, { appointment_id: appointmentId }),
  getStats: () => api.get('/waiting-list/stats'),
};

// Consent Templates API
export const consentTemplatesAPI = {
  getAll: (params) => api.get('/consent-templates', { params }),
  getOne: (id) => api.get(`/consent-templates/${id}`),
  create: (data) => api.post('/consent-templates', data),
  update: (id, data) => api.put(`/consent-templates/${id}`, data),
  delete: (id) => api.delete(`/consent-templates/${id}`),
  getTypes: () => api.get('/consent-templates/types'),
  getPDF: (id) => api.get(`/consent-templates/${id}/pdf`, { responseType: 'blob' }),
};

// Patient Consents API
export const consentsAPI = {
  create: (data) => api.post('/consents', data),
  getOne: (id) => api.get(`/consents/${id}`),
  getByPatient: (patientId) => api.get(`/consents/patients/${patientId}`),
  updateStatus: (id, status) => api.patch(`/consents/${id}/status`, { status }),
  delete: (id) => api.delete(`/consents/${id}`),
  getPDF: (id) => api.get(`/consents/${id}/pdf`, { responseType: 'blob' }),
};

// Treatments API (orçamentos aprovados em tratamento)
export const treatmentsAPI = {
  getAll: (params) => api.get('/treatments', { params }),
  getOne: (id) => api.get(`/treatments/${id}`),
  create: (data) => api.post('/treatments', data),
  update: (id, data) => api.put(`/treatments/${id}`, data),
  delete: (id) => api.delete(`/treatments/${id}`),
};

// Treatment Payments API (pagamentos de tratamentos)
export const treatmentPaymentsAPI = {
  getAll: (treatmentId) => api.get(`/treatment-payments/treatment/${treatmentId}`),
  getOne: (id) => api.get(`/treatment-payments/${id}`),
  create: (data) => api.post('/treatment-payments', data),
  update: (id, data) => api.put(`/treatment-payments/${id}`, data),
  delete: (id) => api.delete(`/treatment-payments/${id}`),
  downloadReceipt: (id) => api.get(`/treatment-payments/${id}/receipt`, { responseType: 'blob' }),
};
