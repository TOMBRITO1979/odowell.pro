import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL ;

const api = axios.create({
  baseURL: `${API_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
  // Enable cookies for cross-origin requests (httpOnly auth cookies)
  withCredentials: true,
});

// Add token to requests (fallback for localStorage during migration)
api.interceptors.request.use(
  (config) => {
    // Only add Authorization header if token exists in localStorage
    // This is for backward compatibility - cookies are now the primary auth method
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

// Flag to prevent multiple refresh attempts
let isRefreshing = false;
let refreshSubscribers = [];

const subscribeTokenRefresh = (callback) => {
  refreshSubscribers.push(callback);
};

const onTokenRefreshed = (token) => {
  refreshSubscribers.forEach((callback) => callback(token));
  refreshSubscribers = [];
};

const onTokenRefreshFailed = () => {
  refreshSubscribers.forEach((callback) => callback(null));
  refreshSubscribers = [];
};

// Handle responses
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Handle 401 Unauthorized - try to refresh token first
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // Wait for the refresh to complete
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh((token) => {
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              resolve(api(originalRequest));
            } else {
              reject(error);
            }
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Try to refresh the token
        const response = await api.post('/auth/refresh');
        const newToken = response.data.token;

        // Update localStorage if token was stored there
        if (localStorage.getItem('token')) {
          localStorage.setItem('token', newToken);
        }

        isRefreshing = false;
        onTokenRefreshed(newToken);

        // Retry the original request
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return api(originalRequest);
      } catch (refreshError) {
        isRefreshing = false;
        onTokenRefreshFailed();

        // Refresh failed - logout
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('tenant');

        // Redirect to appropriate login page based on subdomain
        const hostname = window.location.hostname;
        const isClinicPortal = hostname.split('.').length >= 3 &&
          hostname.includes('odowell.pro') &&
          !['app', 'api', 'www'].includes(hostname.split('.')[0]);

        window.location.href = isClinicPortal ? '/portal-login' : '/login';
        return Promise.reject(error);
      }
    }

    // Handle 402 Payment Required - subscription expired
    if (error.response?.status === 402) {
      const currentPath = window.location.pathname;
      // Don't redirect if already on subscription page
      if (!currentPath.startsWith('/subscription')) {
        // Store the subscription error info
        localStorage.setItem('subscription_expired', JSON.stringify({
          message: error.response?.data?.message || 'Sua assinatura expirou',
          status: error.response?.data?.subscription_status,
          days_expired: error.response?.data?.days_expired
        }));
        window.location.href = '/subscription';
      }
    }
    // Handle 403 Forbidden - insufficient permissions
    if (error.response?.status === 403) {
      const message = error.response?.data?.error || 'Você não tem permissão para acessar este recurso';
      // Import message from antd dynamically to show notification
      import('antd').then(({ message: antdMessage }) => {
        antdMessage.error(message);
      });
    }
    return Promise.reject(error);
  }
);

export default api;

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  logout: () => api.post('/auth/logout'),
  register: (data) => api.post('/auth/register', data),
  getMe: () => api.get('/auth/me'),
  createTenant: (data) => api.post('/tenants', data),
  updateProfile: (data) => api.put('/auth/profile', data),
  changePassword: (data) => api.put('/auth/password', data),
  uploadProfilePicture: (formData) => api.post('/auth/profile/picture', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  verifyEmail: (token) => api.get(`/auth/verify-email?token=${token}`),
  resendVerification: (email) => api.post('/auth/resend-verification', { email }),
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
  getOverdueCount: () => api.get('/payments/overdue-count'),
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
  getById: (id) => api.get(`/stock-movements/${id}`),
  create: (data) => api.post('/stock-movements', data),
  update: (id, data) => api.put(`/stock-movements/${id}`, data),
  delete: (id) => api.delete(`/stock-movements/${id}`),
  getStats: (params) => api.get('/stock-movements/stats', { params }),
  exportCSV: (params) => api.get(`/stock-movements/export/csv?${params}`, { responseType: 'blob' }),
  exportPDF: (params) => api.get(`/stock-movements/export/pdf?${params}`, { responseType: 'blob' }),
  downloadSaleReceipt: (id) => api.get(`/stock-movements/${id}/sale-receipt`, { responseType: 'blob' }),
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
  getProcedures: (params) => api.get('/reports/procedures', { params }),
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
  // SMTP Test
  testSMTP: () => api.post('/settings/smtp/test'),
  // API Key Management
  getAPIKeyStatus: () => api.get('/settings/api-key/status'),
  generateAPIKey: () => api.post('/settings/api-key/generate'),
  toggleAPIKey: (active) => api.patch('/settings/api-key/toggle', { active }),
  revokeAPIKey: () => api.delete('/settings/api-key'),
  getAPIKeyDocs: () => api.get('/settings/api-key/docs'),
  // Embed Token Management (Chatwell)
  getEmbedToken: () => api.get('/settings/embed-token'),
  generateEmbedToken: () => api.post('/settings/embed-token'),
  revokeEmbedToken: () => api.delete('/settings/embed-token'),
  // Danger Zone - Delete Tenant
  deleteTenant: () => api.delete('/settings/tenant'),
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

// Leads API (CRM para WhatsApp)
export const leadsAPI = {
  getAll: (params) => api.get('/leads', { params }),
  getOne: (id) => api.get(`/leads/${id}`),
  create: (data) => api.post('/leads', data),
  update: (id, data) => api.put(`/leads/${id}`, data),
  delete: (id) => api.delete(`/leads/${id}`),
  checkByPhone: (phone) => api.get(`/leads/check/${encodeURIComponent(phone)}`),
  convert: (id, additionalData) => api.post(`/leads/${id}/convert`, additionalData),
  getStats: () => api.get('/leads/stats'),
};

// WhatsApp Business API (Meta WABA Integration)
export const whatsappBusinessAPI = {
  // Get approved templates from Meta
  getTemplates: () => api.get('/settings/whatsapp/templates'),
  // Test WhatsApp connection
  testConnection: () => api.post('/settings/whatsapp/test'),
  // Send a message using a template
  sendMessage: (data) => api.post('/settings/whatsapp/send', data),
  // Send appointment confirmation
  sendConfirmation: (appointmentId) => api.post('/settings/whatsapp/send-confirmation', { appointment_id: parseInt(appointmentId, 10) }),
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
  getByBudgetId: (budgetId) => api.get('/treatments', { params: { budget_id: budgetId } }),
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

// Patient Subscriptions API (Planos - for patients)
export const patientSubscriptionsAPI = {
  getAll: (params) => api.get('/patient-subscriptions', { params }),
  getOne: (id) => api.get(`/patient-subscriptions/${id}`),
  create: (data) => api.post('/patient-subscriptions', data),
  cancel: (id) => api.post(`/patient-subscriptions/${id}/cancel`),
  cancelImmediately: (id) => api.delete(`/patient-subscriptions/${id}`),
  refresh: (id) => api.post(`/patient-subscriptions/${id}/refresh`),
  getPayments: (id, params) => api.get(`/patient-subscriptions/${id}/payments`, { params }),
  resendLink: (id) => api.post(`/patient-subscriptions/${id}/resend-link`),
  getStripeProducts: () => api.get('/stripe/products'),
};

// Stripe Settings API (for patient subscriptions)
export const stripeSettingsAPI = {
  get: () => api.get('/stripe-settings'),
  update: (data) => api.put('/stripe-settings', data),
  disconnect: () => api.delete('/stripe-settings'),
  test: () => api.get('/stripe-settings/test'),
  getWebhookURL: () => api.get('/stripe-settings/webhook-url'),
};

// Tenant Subscription API (Assinatura - for clinics)
export const subscriptionAPI = {
  getPlans: () => api.get('/subscription/plans'),
  getStatus: () => api.get('/subscription/status'),
  createCheckout: (data) => api.post('/subscription/checkout', data),
  createPortal: () => api.post('/subscription/portal'),
  cancel: () => api.post('/subscription/cancel'),
};

// Super Admin API (platform administration)
export const adminAPI = {
  // Dashboard
  getDashboard: () => api.get('/admin/dashboard'),

  // Tenant Management
  getAllTenants: () => api.get('/admin/tenants'),
  getTenantDetails: (id) => api.get(`/admin/tenants/${id}`),
  updateTenantStatus: (id, data) => api.patch(`/admin/tenants/${id}`, data),
  deleteTenant: (id) => api.delete(`/admin/tenants/${id}`),
  getTenantUsers: (id) => api.get(`/admin/tenants/${id}/users`),
  updateTenantUserStatus: (tenantId, userId, data) => api.patch(`/admin/tenants/${tenantId}/users/${userId}`, data),

  // Reports
  getUnverifiedTenants: () => api.get('/admin/tenants/unverified'),
  getExpiringTrials: () => api.get('/admin/tenants/expiring'),
  getInactiveTenants: () => api.get('/admin/tenants/inactive'),
};

// Audit Logs API (LGPD Compliance)
export const auditAPI = {
  getLogs: (params) => api.get('/audit/logs', { params }),
  getStats: () => api.get('/audit/stats'),
  exportCSV: (params) => api.get('/audit/export/csv', { params, responseType: 'blob' }),
};

// Portal Notifications API (Patient Portal Activities)
export const portalNotificationsAPI = {
  getAll: (params) => api.get('/portal-notifications', { params }),
};

// Data Requests API (LGPD - Solicitacoes do Titular)
export const dataRequestAPI = {
  getAll: (params) => api.get('/data-requests', { params }),
  getOne: (id) => api.get(`/data-requests/${id}`),
  create: (data) => api.post('/data-requests', data),
  updateStatus: (id, data) => api.patch(`/data-requests/${id}/status`, data),
  getStats: () => api.get('/data-requests/stats'),
  getByPatient: (patientId) => api.get(`/data-requests/patient/${patientId}`),
  // OTP Verification (LGPD identity verification)
  sendOTP: (id) => api.post(`/data-requests/${id}/send-otp`),
  verifyOTP: (id, code) => api.post(`/data-requests/${id}/verify-otp`, { code }),
  // Data Export (LGPD portability)
  exportData: (id, format = 'json') => api.get(`/data-requests/${id}/export?format=${format}`, { responseType: 'blob' }),
};

// LGPD Data Deletion API
export const lgpdAPI = {
  getDeletionPreview: (patientId) => api.get(`/lgpd/patients/${patientId}/deletion-preview`),
  permanentDelete: (patientId, data) => api.delete(`/lgpd/patients/${patientId}/permanent`, { data }),
  anonymize: (patientId, data) => api.post(`/lgpd/patients/${patientId}/anonymize`, data),
};

// Digital Certificates API (ICP-Brasil A1)
export const certificatesAPI = {
  getAll: () => api.get('/certificates'),
  upload: (formData) => api.post('/certificates', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  activate: (id) => api.post(`/certificates/${id}/activate`),
  validatePassword: (id, password) => api.post(`/certificates/${id}/validate`, { password }),
  delete: (id) => api.delete(`/certificates/${id}`),
};

// Document Signing API
export const signingAPI = {
  signPrescription: (id, password) => api.post(`/prescriptions/${id}/sign`, { password }),
  signMedicalRecord: (id, password) => api.post(`/medical-records/${id}/sign`, { password }),
  verifySignature: (type, id) => api.get(`/documents/${type}/${id}/verify`),
  downloadSignedPrescriptionPDF: (id) => api.get(`/prescriptions/${id}/pdf/signed`, { responseType: 'blob' }),
};

// Patient Portal API (for patients)
export const patientPortalAPI = {
  // Patient's own data
  getProfile: () => api.get('/patient/me'),
  getClinic: () => api.get('/patient/clinic'),

  // Appointments
  getAppointments: (status) => api.get('/patient/appointments', { params: { status } }),
  createAppointment: (data) => api.post('/patient/appointments', data),
  cancelAppointment: (id) => api.delete(`/patient/appointments/${id}`),

  // Available slots
  getAvailableSlots: (dentistId, date) => api.get('/patient/available-slots', { params: { dentist_id: dentistId, date } }),

  // Medical Records (prontuarios)
  getMedicalRecords: (type) => api.get('/patient/medical-records', { params: { type: type || undefined } }),
  getMedicalRecordDetail: (id) => api.get(`/patient/medical-records/${id}`),
};

// Patient Portal Admin API (for staff to manage patient access)
export const patientPortalAdminAPI = {
  getAccess: (patientId) => api.get(`/patient-portal/${patientId}`),
  createAccess: (data) => api.post('/patient-portal', data),
  updatePassword: (patientId, password) => api.put(`/patient-portal/${patientId}/password`, { password }),
  deleteAccess: (patientId) => api.delete(`/patient-portal/${patientId}`),
};

// Patient Portal Public API (no auth required - for login page)
export const patientPortalPublicAPI = {
  getClinicInfo: (slug) => api.get('/portal/clinic-info', { params: { slug } }),
  login: (slug, email, password) => api.post('/portal/login', { slug, email, password }),
};
