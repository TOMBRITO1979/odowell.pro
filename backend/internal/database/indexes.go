package database

import (
	"fmt"
	"log"
)

// ApplyTenantIndexes creates performance indexes for a specific tenant schema
func ApplyTenantIndexes(schemaName string) error {
	log.Printf("Applying indexes to schema: %s", schemaName)

	// Set search path to the tenant schema
	if err := DB.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error; err != nil {
		return fmt.Errorf("failed to set search_path: %v", err)
	}

	// List of index creation statements
	indexes := []string{
		// Patients
		"CREATE INDEX IF NOT EXISTS idx_patients_name ON patients(name)",
		"CREATE INDEX IF NOT EXISTS idx_patients_active ON patients(active) WHERE active = true",
		"CREATE INDEX IF NOT EXISTS idx_patients_created_at ON patients(created_at DESC)",

		// Appointments - calendar queries
		"CREATE INDEX IF NOT EXISTS idx_appointments_start_time ON appointments(start_time)",
		"CREATE INDEX IF NOT EXISTS idx_appointments_dentist_date ON appointments(dentist_id, start_time)",
		"CREATE INDEX IF NOT EXISTS idx_appointments_patient_date ON appointments(patient_id, start_time DESC)",
		"CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status)",
		"CREATE INDEX IF NOT EXISTS idx_appointments_status_date ON appointments(status, start_time) WHERE deleted_at IS NULL",

		// Medical Records
		"CREATE INDEX IF NOT EXISTS idx_medical_records_patient_date ON medical_records(patient_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_medical_records_dentist ON medical_records(dentist_id)",
		"CREATE INDEX IF NOT EXISTS idx_medical_records_type ON medical_records(type)",
		"CREATE INDEX IF NOT EXISTS idx_medical_records_signed ON medical_records(is_signed) WHERE is_signed = true",

		// Prescriptions
		"CREATE INDEX IF NOT EXISTS idx_prescriptions_patient_date ON prescriptions(patient_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_prescriptions_dentist ON prescriptions(dentist_id)",
		"CREATE INDEX IF NOT EXISTS idx_prescriptions_status ON prescriptions(status)",
		"CREATE INDEX IF NOT EXISTS idx_prescriptions_signed ON prescriptions(is_signed) WHERE is_signed = true",

		// Budgets
		"CREATE INDEX IF NOT EXISTS idx_budgets_patient ON budgets(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_budgets_status ON budgets(status)",
		"CREATE INDEX IF NOT EXISTS idx_budgets_dentist_status ON budgets(dentist_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_budgets_created_at ON budgets(created_at DESC)",

		// Payments
		"CREATE INDEX IF NOT EXISTS idx_payments_patient ON payments(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)",
		"CREATE INDEX IF NOT EXISTS idx_payments_due_date ON payments(due_date) WHERE status = 'pending'",
		"CREATE INDEX IF NOT EXISTS idx_payments_paid_date ON payments(paid_date) WHERE paid_date IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_payments_type_status ON payments(type, status)",
		"CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC)",

		// Treatments
		"CREATE INDEX IF NOT EXISTS idx_treatments_patient ON treatments(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_treatments_status ON treatments(status)",
		"CREATE INDEX IF NOT EXISTS idx_treatments_dentist ON treatments(dentist_id)",

		// Treatment Payments
		"CREATE INDEX IF NOT EXISTS idx_treatment_payments_treatment ON treatment_payments(treatment_id)",
		"CREATE INDEX IF NOT EXISTS idx_treatment_payments_paid_date ON treatment_payments(paid_date DESC)",

		// Commissions
		"CREATE INDEX IF NOT EXISTS idx_commissions_dentist ON commissions(dentist_id)",
		"CREATE INDEX IF NOT EXISTS idx_commissions_status ON commissions(status)",
		"CREATE INDEX IF NOT EXISTS idx_commissions_paid_date ON commissions(paid_date) WHERE status = 'paid'",

		// Exams
		"CREATE INDEX IF NOT EXISTS idx_exams_patient_date ON exams(patient_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_exams_type ON exams(exam_type)",

		// Attachments
		"CREATE INDEX IF NOT EXISTS idx_attachments_patient ON attachments(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_attachments_category ON attachments(category)",

		// Products
		"CREATE INDEX IF NOT EXISTS idx_products_code ON products(code) WHERE code IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)",
		"CREATE INDEX IF NOT EXISTS idx_products_active ON products(active) WHERE active = true",
		"CREATE INDEX IF NOT EXISTS idx_products_low_stock ON products(quantity, minimum_stock) WHERE quantity <= minimum_stock",

		// Stock Movements
		"CREATE INDEX IF NOT EXISTS idx_stock_movements_product ON stock_movements(product_id)",
		"CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON stock_movements(type)",

		// Suppliers
		"CREATE INDEX IF NOT EXISTS idx_suppliers_cnpj ON suppliers(cnpj) WHERE cnpj IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_suppliers_active ON suppliers(active) WHERE active = true",

		// Tasks
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks(created_by)",

		// Task Assignments
		"CREATE INDEX IF NOT EXISTS idx_task_assignments_task ON task_assignments(task_id)",
		"CREATE INDEX IF NOT EXISTS idx_task_assignments_type_id ON task_assignments(assignable_type, assignable_id)",

		// Task Users
		"CREATE INDEX IF NOT EXISTS idx_task_users_task ON task_users(task_id)",
		"CREATE INDEX IF NOT EXISTS idx_task_users_user ON task_users(user_id)",

		// Waiting List
		"CREATE INDEX IF NOT EXISTS idx_waiting_list_patient ON waiting_lists(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_waiting_list_status ON waiting_lists(status)",
		"CREATE INDEX IF NOT EXISTS idx_waiting_list_priority ON waiting_lists(priority)",
		"CREATE INDEX IF NOT EXISTS idx_waiting_list_dentist ON waiting_lists(dentist_id) WHERE dentist_id IS NOT NULL",

		// Consent Templates
		"CREATE INDEX IF NOT EXISTS idx_consent_templates_type ON consent_templates(type)",
		"CREATE INDEX IF NOT EXISTS idx_consent_templates_active ON consent_templates(active) WHERE active = true",

		// Patient Consents
		"CREATE INDEX IF NOT EXISTS idx_patient_consents_patient ON patient_consents(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_patient_consents_template ON patient_consents(template_id)",
		"CREATE INDEX IF NOT EXISTS idx_patient_consents_status ON patient_consents(status)",
		"CREATE INDEX IF NOT EXISTS idx_patient_consents_signed_at ON patient_consents(signed_at DESC)",

		// Campaigns
		"CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status)",
		"CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled ON campaigns(scheduled_at) WHERE status = 'scheduled'",

		// Campaign Recipients
		"CREATE INDEX IF NOT EXISTS idx_campaign_recipients_campaign ON campaign_recipients(campaign_id)",
		"CREATE INDEX IF NOT EXISTS idx_campaign_recipients_status ON campaign_recipients(status)",

		// Leads
		"CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status)",
		"CREATE INDEX IF NOT EXISTS idx_leads_source ON leads(source)",
		"CREATE INDEX IF NOT EXISTS idx_leads_phone ON leads(phone)",
		"CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads(created_at DESC)",

		// Data Requests (LGPD)
		"CREATE INDEX IF NOT EXISTS idx_data_requests_patient ON data_requests(patient_id)",
		"CREATE INDEX IF NOT EXISTS idx_data_requests_status ON data_requests(status)",
		"CREATE INDEX IF NOT EXISTS idx_data_requests_type ON data_requests(request_type)",
		"CREATE INDEX IF NOT EXISTS idx_data_requests_created_at ON data_requests(created_at DESC)",

		// Treatment Protocols
		"CREATE INDEX IF NOT EXISTS idx_treatment_protocols_category ON treatment_protocols(category)",
		"CREATE INDEX IF NOT EXISTS idx_treatment_protocols_active ON treatment_protocols(active) WHERE active = true",
	}

	// Apply each index
	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			// Log but don't fail - some tables may not exist in older tenants
			log.Printf("Warning: Could not create index in %s: %v", schemaName, err)
		}
	}

	// Reset search path
	DB.Exec("SET search_path TO public")

	log.Printf("Indexes applied to schema: %s", schemaName)
	return nil
}

// ApplyPublicSchemaIndexes creates performance indexes for the public schema
func ApplyPublicSchemaIndexes() error {
	log.Println("Applying indexes to public schema...")

	// Ensure we're in public schema
	if err := DB.Exec("SET search_path TO public").Error; err != nil {
		return fmt.Errorf("failed to set search_path: %v", err)
	}

	indexes := []string{
		// Audit Logs
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource, resource_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON audit_logs(success) WHERE success = false",

		// Users
		"CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, active)",
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON users(tenant_id, role)",

		// Tenants
		"CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(active) WHERE active = true",
		"CREATE INDEX IF NOT EXISTS idx_tenants_subscription_status ON tenants(subscription_status)",
		"CREATE INDEX IF NOT EXISTS idx_tenants_expires_at ON tenants(expires_at) WHERE expires_at IS NOT NULL",

		// Password Resets
		"CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_used ON password_resets(used)",

		// Email Verifications
		"CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_email_verifications_verified ON email_verifications(verified)",

		// User Permissions
		"CREATE INDEX IF NOT EXISTS idx_user_permissions_user_module ON user_permissions(user_id, module_id)",

		// User Certificates
		"CREATE INDEX IF NOT EXISTS idx_user_certificates_user_active ON user_certificates(user_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_user_certificates_expires_at ON user_certificates(expires_at)",
	}

	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			log.Printf("Warning: Could not create index: %v", err)
		}
	}

	log.Println("Public schema indexes applied successfully")
	return nil
}

// ApplyPublicSchemaForeignKeys adds foreign key constraints to the public schema
func ApplyPublicSchemaForeignKeys() error {
	log.Println("Applying foreign key constraints to public schema...")

	if err := DB.Exec("SET search_path TO public").Error; err != nil {
		return fmt.Errorf("failed to set search_path: %v", err)
	}

	// List of foreign key additions (only add if not exists)
	fkStatements := []struct {
		check string
		add   string
	}{
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_tenant'",
			add:   "ALTER TABLE users ADD CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_permissions_user'",
			add:   "ALTER TABLE user_permissions ADD CONSTRAINT fk_user_permissions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_permissions_module'",
			add:   "ALTER TABLE user_permissions ADD CONSTRAINT fk_user_permissions_module FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_permissions_module'",
			add:   "ALTER TABLE permissions ADD CONSTRAINT fk_permissions_module FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_certificates_user'",
			add:   "ALTER TABLE user_certificates ADD CONSTRAINT fk_user_certificates_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_tenant_settings_tenant'",
			add:   "ALTER TABLE tenant_settings ADD CONSTRAINT fk_tenant_settings_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_password_resets_user'",
			add:   "ALTER TABLE password_resets ADD CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		},
		{
			check: "SELECT 1 FROM pg_constraint WHERE conname = 'fk_email_verifications_user'",
			add:   "ALTER TABLE email_verifications ADD CONSTRAINT fk_email_verifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		},
	}

	for _, fk := range fkStatements {
		var exists bool
		DB.Raw(fk.check).Scan(&exists)
		if !exists {
			if err := DB.Exec(fk.add).Error; err != nil {
				log.Printf("Warning: Could not add FK constraint: %v", err)
			}
		}
	}

	log.Println("Public schema foreign keys applied successfully")
	return nil
}

// ApplyAllIndexesAndConstraints applies indexes and FK constraints to all schemas
func ApplyAllIndexesAndConstraints() error {
	log.Println("Starting indexes and constraints migration...")

	// Apply to public schema first
	if err := ApplyPublicSchemaIndexes(); err != nil {
		log.Printf("Warning: Error applying public indexes: %v", err)
	}

	if err := ApplyPublicSchemaForeignKeys(); err != nil {
		log.Printf("Warning: Error applying public FK constraints: %v", err)
	}

	// Get list of all tenant schemas
	var schemas []string
	err := DB.Raw(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`).Scan(&schemas).Error

	if err != nil {
		return fmt.Errorf("failed to list tenant schemas: %v", err)
	}

	// Apply indexes to each tenant
	for _, schema := range schemas {
		if err := ApplyTenantIndexes(schema); err != nil {
			log.Printf("Warning: Error applying indexes to %s: %v", schema, err)
		}
	}

	// Reset search path
	DB.Exec("SET search_path TO public")

	log.Println("Indexes and constraints migration completed")
	return nil
}
