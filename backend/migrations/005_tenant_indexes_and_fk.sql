-- Migration: 005_tenant_indexes_and_fk.sql
-- Description: Template for tenant schema indexes and foreign keys
-- This is applied per-tenant via Go code
-- Author: Claude Code
-- Date: 2025-12-21

-- ============================================================================
-- TENANT SCHEMA INDEXES (Performance)
-- ============================================================================

-- Patients - search and listing
CREATE INDEX IF NOT EXISTS idx_patients_name ON patients(name);
CREATE INDEX IF NOT EXISTS idx_patients_active ON patients(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_patients_created_at ON patients(created_at DESC);

-- Appointments - calendar queries (most important for performance)
CREATE INDEX IF NOT EXISTS idx_appointments_start_time ON appointments(start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_dentist_date ON appointments(dentist_id, start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_patient_date ON appointments(patient_id, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_status_date ON appointments(status, start_time) WHERE deleted_at IS NULL;

-- Medical Records - patient history
CREATE INDEX IF NOT EXISTS idx_medical_records_patient_date ON medical_records(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_medical_records_dentist ON medical_records(dentist_id);
CREATE INDEX IF NOT EXISTS idx_medical_records_type ON medical_records(type);
CREATE INDEX IF NOT EXISTS idx_medical_records_signed ON medical_records(is_signed) WHERE is_signed = true;

-- Prescriptions - patient lookup
CREATE INDEX IF NOT EXISTS idx_prescriptions_patient_date ON prescriptions(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prescriptions_dentist ON prescriptions(dentist_id);
CREATE INDEX IF NOT EXISTS idx_prescriptions_status ON prescriptions(status);
CREATE INDEX IF NOT EXISTS idx_prescriptions_signed ON prescriptions(is_signed) WHERE is_signed = true;

-- Budgets - financial queries
CREATE INDEX IF NOT EXISTS idx_budgets_patient ON budgets(patient_id);
CREATE INDEX IF NOT EXISTS idx_budgets_status ON budgets(status);
CREATE INDEX IF NOT EXISTS idx_budgets_dentist_status ON budgets(dentist_id, status);
CREATE INDEX IF NOT EXISTS idx_budgets_created_at ON budgets(created_at DESC);

-- Payments - financial reports
CREATE INDEX IF NOT EXISTS idx_payments_patient ON payments(patient_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_due_date ON payments(due_date) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_payments_paid_date ON payments(paid_date) WHERE paid_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payments_type_status ON payments(type, status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);

-- Treatments - treatment tracking
CREATE INDEX IF NOT EXISTS idx_treatments_patient ON treatments(patient_id);
CREATE INDEX IF NOT EXISTS idx_treatments_status ON treatments(status);
CREATE INDEX IF NOT EXISTS idx_treatments_dentist ON treatments(dentist_id);

-- Treatment Payments
CREATE INDEX IF NOT EXISTS idx_treatment_payments_treatment ON treatment_payments(treatment_id);
CREATE INDEX IF NOT EXISTS idx_treatment_payments_paid_date ON treatment_payments(paid_date DESC);

-- Commissions - professional reports
CREATE INDEX IF NOT EXISTS idx_commissions_dentist ON commissions(dentist_id);
CREATE INDEX IF NOT EXISTS idx_commissions_status ON commissions(status);
CREATE INDEX IF NOT EXISTS idx_commissions_paid_date ON commissions(paid_date) WHERE status = 'paid';

-- Exams - patient files
CREATE INDEX IF NOT EXISTS idx_exams_patient_date ON exams(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exams_type ON exams(exam_type);

-- Attachments - file lookup
CREATE INDEX IF NOT EXISTS idx_attachments_patient ON attachments(patient_id);
CREATE INDEX IF NOT EXISTS idx_attachments_category ON attachments(category);

-- Products - inventory
CREATE INDEX IF NOT EXISTS idx_products_code ON products(code) WHERE code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_active ON products(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_products_low_stock ON products(quantity, minimum_stock) WHERE quantity <= minimum_stock;

-- Stock Movements - inventory history
CREATE INDEX IF NOT EXISTS idx_stock_movements_product ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON stock_movements(type);

-- Suppliers
CREATE INDEX IF NOT EXISTS idx_suppliers_cnpj ON suppliers(cnpj) WHERE cnpj IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_suppliers_active ON suppliers(active) WHERE active = true;

-- Tasks - task management
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks(created_by);

-- Task Assignments
CREATE INDEX IF NOT EXISTS idx_task_assignments_task ON task_assignments(task_id);
CREATE INDEX IF NOT EXISTS idx_task_assignments_type_id ON task_assignments(assignable_type, assignable_id);

-- Task Users
CREATE INDEX IF NOT EXISTS idx_task_users_task ON task_users(task_id);
CREATE INDEX IF NOT EXISTS idx_task_users_user ON task_users(user_id);

-- Waiting List
CREATE INDEX IF NOT EXISTS idx_waiting_list_patient ON waiting_lists(patient_id);
CREATE INDEX IF NOT EXISTS idx_waiting_list_status ON waiting_lists(status);
CREATE INDEX IF NOT EXISTS idx_waiting_list_priority ON waiting_lists(priority);
CREATE INDEX IF NOT EXISTS idx_waiting_list_dentist ON waiting_lists(dentist_id) WHERE dentist_id IS NOT NULL;

-- Consent Templates
CREATE INDEX IF NOT EXISTS idx_consent_templates_type ON consent_templates(type);
CREATE INDEX IF NOT EXISTS idx_consent_templates_active ON consent_templates(active) WHERE active = true;

-- Patient Consents
CREATE INDEX IF NOT EXISTS idx_patient_consents_patient ON patient_consents(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_consents_template ON patient_consents(template_id);
CREATE INDEX IF NOT EXISTS idx_patient_consents_status ON patient_consents(status);
CREATE INDEX IF NOT EXISTS idx_patient_consents_signed_at ON patient_consents(signed_at DESC);

-- Campaigns
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled ON campaigns(scheduled_at) WHERE status = 'scheduled';

-- Campaign Recipients
CREATE INDEX IF NOT EXISTS idx_campaign_recipients_campaign ON campaign_recipients(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_recipients_status ON campaign_recipients(status);

-- Leads (CRM)
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
CREATE INDEX IF NOT EXISTS idx_leads_source ON leads(source);
CREATE INDEX IF NOT EXISTS idx_leads_phone ON leads(phone);
CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads(created_at DESC);

-- Data Requests (LGPD)
CREATE INDEX IF NOT EXISTS idx_data_requests_patient ON data_requests(patient_id);
CREATE INDEX IF NOT EXISTS idx_data_requests_status ON data_requests(status);
CREATE INDEX IF NOT EXISTS idx_data_requests_type ON data_requests(type);
CREATE INDEX IF NOT EXISTS idx_data_requests_created_at ON data_requests(created_at DESC);

-- Treatment Protocols
CREATE INDEX IF NOT EXISTS idx_treatment_protocols_category ON treatment_protocols(category);
CREATE INDEX IF NOT EXISTS idx_treatment_protocols_active ON treatment_protocols(active) WHERE active = true;
