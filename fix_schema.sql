-- Dr. Crwell - Schema Migration Fix
-- Adding missing columns to tenant_1 schema

-- Fix PATIENTS table
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS rg TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS phone TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS number TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS complement TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS district TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS systemic_diseases TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS blood_type TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS has_insurance BOOLEAN DEFAULT false;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS insurance_name TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS insurance_number TEXT;
ALTER TABLE tenant_1.patients ADD COLUMN IF NOT EXISTS notes TEXT;

-- Fix PRODUCTS table
ALTER TABLE tenant_1.products ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE tenant_1.products ADD COLUMN IF NOT EXISTS barcode TEXT;
ALTER TABLE tenant_1.products ADD COLUMN IF NOT EXISTS expiration_date TIMESTAMP;

-- Fix SUPPLIERS table
ALTER TABLE tenant_1.suppliers ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE tenant_1.suppliers ADD COLUMN IF NOT EXISTS city TEXT;
ALTER TABLE tenant_1.suppliers ADD COLUMN IF NOT EXISTS state TEXT;
ALTER TABLE tenant_1.suppliers ADD COLUMN IF NOT EXISTS zip_code TEXT;
ALTER TABLE tenant_1.suppliers ADD COLUMN IF NOT EXISTS notes TEXT;

-- Fix BUDGETS table
ALTER TABLE tenant_1.budgets ADD COLUMN IF NOT EXISTS items TEXT;
ALTER TABLE tenant_1.budgets ADD COLUMN IF NOT EXISTS notes TEXT;

-- Fix PAYMENTS table
ALTER TABLE tenant_1.payments ADD COLUMN IF NOT EXISTS is_installment BOOLEAN DEFAULT false;
ALTER TABLE tenant_1.payments ADD COLUMN IF NOT EXISTS installment_number INTEGER;
ALTER TABLE tenant_1.payments ADD COLUMN IF NOT EXISTS total_installments INTEGER;
ALTER TABLE tenant_1.payments ADD COLUMN IF NOT EXISTS is_insurance BOOLEAN DEFAULT false;
ALTER TABLE tenant_1.payments ADD COLUMN IF NOT EXISTS insurance_name TEXT;

-- Fix MEDICAL_RECORDS table
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS type TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS odontogram TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS treatment_plan TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS procedure_done TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS materials TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS prescription TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS certificate TEXT;
ALTER TABLE tenant_1.medical_records ADD COLUMN IF NOT EXISTS evolution TEXT;

-- Fix CAMPAIGNS table
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS segment_type TEXT;
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS filters TEXT;
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS total_recipients INTEGER DEFAULT 0;
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS sent_count INTEGER DEFAULT 0;
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMP;
ALTER TABLE tenant_1.campaigns ADD COLUMN IF NOT EXISTS sent_at TIMESTAMP;

-- Fix EXAMS table
ALTER TABLE tenant_1.exams ADD COLUMN IF NOT EXISTS exam_date TIMESTAMP;
ALTER TABLE tenant_1.exams ADD COLUMN IF NOT EXISTS s3_key TEXT;
ALTER TABLE tenant_1.exams ADD COLUMN IF NOT EXISTS file_name TEXT;
ALTER TABLE tenant_1.exams ADD COLUMN IF NOT EXISTS file_size BIGINT;
ALTER TABLE tenant_1.exams ADD COLUMN IF NOT EXISTS uploaded_by_id INTEGER;
