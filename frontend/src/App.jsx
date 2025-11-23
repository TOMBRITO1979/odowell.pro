import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, Spin } from 'antd';
import { useAuth } from './contexts/AuthContext';
import ptBR from 'antd/locale/pt_BR';
import InstallPWA from './components/InstallPWA';

// Auth pages
import Login from './pages/auth/Login';
import Register from './pages/auth/Register';
import CreateTenant from './pages/auth/CreateTenant';

// Protected pages
import DashboardLayout from './components/layouts/DashboardLayout';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import Settings from './pages/Settings';
import Patients from './pages/patients/Patients';
import PatientForm from './pages/patients/PatientForm';
import PatientDetails from './pages/patients/PatientDetails';
import Appointments from './pages/appointments/Appointments';
import AppointmentForm from './pages/appointments/AppointmentForm';
import MedicalRecords from './pages/medical-records/MedicalRecords';
import MedicalRecordForm from './pages/medical-records/MedicalRecordForm';
import MedicalRecordDetails from './pages/medical-records/MedicalRecordDetails';
import AppointmentDetails from './pages/appointments/AppointmentDetails';
import Budgets from './pages/financial/Budgets';
import BudgetForm from './pages/financial/BudgetForm';
import BudgetView from './pages/financial/BudgetView';
import Payments from './pages/financial/Payments';
import PaymentForm from './pages/financial/PaymentForm';
import Products from './pages/inventory/Products';
import ProductForm from './pages/inventory/ProductForm';
import Suppliers from './pages/inventory/Suppliers';
import StockMovements from './pages/inventory/StockMovements';
import Campaigns from './pages/campaigns/Campaigns';
import CampaignForm from './pages/campaigns/CampaignForm';
import Exams from './pages/exams/Exams';
import ExamForm from './pages/exams/ExamForm';
import ExamDetails from './pages/exams/ExamDetails';
import Prescriptions from './pages/prescriptions/Prescriptions';
import PrescriptionForm from './pages/prescriptions/PrescriptionForm';
import PrescriptionDetails from './pages/prescriptions/PrescriptionDetails';
import Reports from './pages/reports/Reports';
import Users from './pages/users/Users';
import Tasks from './pages/tasks/Tasks';
import TaskForm from './pages/tasks/TaskForm';
import TaskDetails from './pages/tasks/TaskDetails';
import WaitingList from './pages/waiting-list/WaitingList';
import WaitingListForm from './pages/waiting-list/WaitingListForm';

const PrivateRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return isAuthenticated ? children : <Navigate to="/login" />;
};

function App() {
  return (
    <>
      <InstallPWA />
      <ConfigProvider
        locale={ptBR}
        theme={{
        token: {
          colorPrimary: '#16a34a',
          colorSuccess: '#52c41a',
          colorWarning: '#faad14',
          colorError: '#ff4d4f',
          colorInfo: '#1890ff',
          colorTextBase: '#262626',
          colorBgBase: '#ffffff',
          fontFamily: "'Inter', 'Poppins', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif",
          fontSize: 14,
          borderRadius: 6,
          controlHeight: 36,
        },
        components: {
          Button: {
            controlHeight: 36,
            controlHeightLG: 40,
            controlHeightSM: 32,
            borderRadius: 6,
            fontWeight: 500,
          },
          Input: {
            controlHeight: 36,
            borderRadius: 6,
          },
          Select: {
            controlHeight: 36,
            borderRadius: 6,
          },
          Card: {
            borderRadiusLG: 8,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
          },
          Table: {
            borderRadiusLG: 8,
          },
          Modal: {
            borderRadiusLG: 8,
          },
        },
      }}
    >
      <Routes>
      {/* Public routes */}
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/create-tenant" element={<CreateTenant />} />

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <PrivateRoute>
            <DashboardLayout />
          </PrivateRoute>
        }
      >
        <Route index element={<Dashboard />} />

        {/* User Profile & Settings */}
        <Route path="profile" element={<Profile />} />
        <Route path="settings" element={<Settings />} />

        {/* Patients */}
        <Route path="patients" element={<Patients />} />
        <Route path="patients/new" element={<PatientForm />} />
        <Route path="patients/:id" element={<PatientDetails />} />
        <Route path="patients/:id/edit" element={<PatientForm />} />

        {/* Appointments */}
        <Route path="appointments" element={<Appointments />} />
        <Route path="appointments/new" element={<AppointmentForm />} />
        <Route path="appointments/:id" element={<AppointmentDetails />} />
        <Route path="appointments/:id/edit" element={<AppointmentForm />} />

        {/* Medical Records */}
        <Route path="medical-records" element={<MedicalRecords />} />
        <Route path="medical-records/new" element={<MedicalRecordForm />} />
        <Route path="medical-records/:id" element={<MedicalRecordDetails />} />
        <Route path="medical-records/:id/view" element={<MedicalRecordDetails />} />
        <Route path="medical-records/:id/edit" element={<MedicalRecordForm />} />

        {/* Financial */}
        <Route path="budgets" element={<Budgets />} />
        <Route path="budgets/new" element={<BudgetForm />} />
        <Route path="budgets/:id/view" element={<BudgetView />} />
        <Route path="budgets/:id/edit" element={<BudgetForm />} />
        <Route path="payments" element={<Payments />} />
        <Route path="payments/new" element={<PaymentForm />} />
        <Route path="payments/:id/edit" element={<PaymentForm />} />

        {/* Inventory */}
        <Route path="products" element={<Products />} />
        <Route path="products/new" element={<ProductForm />} />
        <Route path="products/:id/edit" element={<ProductForm />} />
        <Route path="suppliers" element={<Suppliers />} />
        <Route path="stock-movements" element={<StockMovements />} />

        {/* Campaigns */}
        <Route path="campaigns" element={<Campaigns />} />
        <Route path="campaigns/new" element={<CampaignForm />} />
        <Route path="campaigns/:id/edit" element={<CampaignForm />} />

        {/* Exams */}
        <Route path="exams" element={<Exams />} />
        <Route path="exams/:id" element={<ExamDetails />} />
        <Route path="exams/:id/edit" element={<ExamForm />} />

        {/* Prescriptions */}
        <Route path="prescriptions" element={<Prescriptions />} />
        <Route path="prescriptions/new" element={<PrescriptionForm />} />
        <Route path="prescriptions/:id" element={<PrescriptionDetails />} />
        <Route path="prescriptions/:id/edit" element={<PrescriptionForm />} />

        {/* Tasks */}
        <Route path="tasks" element={<Tasks />} />
        <Route path="tasks/new" element={<TaskForm />} />
        <Route path="tasks/:id" element={<TaskDetails />} />
        <Route path="tasks/:id/edit" element={<TaskForm />} />

        {/* Waiting List */}
        <Route path="waiting-list" element={<WaitingList />} />
        <Route path="waiting-list/new" element={<WaitingListForm />} />
        <Route path="waiting-list/:id/edit" element={<WaitingListForm />} />

        {/* Reports */}
        <Route path="reports" element={<Reports />} />

        {/* User Management (Admin only) */}
        <Route path="users" element={<Users />} />
      </Route>

      {/* Catch all */}
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
      </ConfigProvider>
    </>
  );
}

export default App;
