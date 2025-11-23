-- Migration: Create waiting_lists table
-- This table stores patients waiting for appointments

-- Create waiting_lists table in tenant schemas
CREATE TABLE IF NOT EXISTS waiting_lists (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    patient_id BIGINT NOT NULL,
    dentist_id BIGINT,
    procedure TEXT,
    preferred_dates JSONB,
    priority TEXT DEFAULT 'normal',
    status TEXT DEFAULT 'waiting',
    contacted_at TIMESTAMP WITH TIME ZONE,
    contacted_by BIGINT,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    appointment_id BIGINT,
    notes TEXT,
    created_by BIGINT NOT NULL,

    -- Indexes
    CONSTRAINT fk_waiting_lists_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
    CONSTRAINT fk_waiting_lists_appointment FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE SET NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_waiting_lists_patient_id ON waiting_lists(patient_id);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_dentist_id ON waiting_lists(dentist_id);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_status ON waiting_lists(status);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_priority ON waiting_lists(priority);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_deleted_at ON waiting_lists(deleted_at);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_created_at ON waiting_lists(created_at);

-- Add check constraints
ALTER TABLE waiting_lists ADD CONSTRAINT check_priority CHECK (priority IN ('normal', 'urgent'));
ALTER TABLE waiting_lists ADD CONSTRAINT check_status CHECK (status IN ('waiting', 'contacted', 'scheduled', 'cancelled'));

-- Comments
COMMENT ON TABLE waiting_lists IS 'Patients waiting for appointments - supports prioritization and tracking';
COMMENT ON COLUMN waiting_lists.patient_id IS 'Patient waiting for appointment';
COMMENT ON COLUMN waiting_lists.dentist_id IS 'Preferred dentist (NULL = any dentist)';
COMMENT ON COLUMN waiting_lists.procedure IS 'Procedure needed';
COMMENT ON COLUMN waiting_lists.preferred_dates IS 'JSONB array of preferred date ranges';
COMMENT ON COLUMN waiting_lists.priority IS 'Priority: normal or urgent';
COMMENT ON COLUMN waiting_lists.status IS 'Status: waiting, contacted, scheduled, cancelled';
COMMENT ON COLUMN waiting_lists.contacted_at IS 'When patient was contacted';
COMMENT ON COLUMN waiting_lists.contacted_by IS 'User who contacted the patient';
COMMENT ON COLUMN waiting_lists.scheduled_at IS 'When appointment was scheduled';
COMMENT ON COLUMN waiting_lists.appointment_id IS 'Link to scheduled appointment';
COMMENT ON COLUMN waiting_lists.created_by IS 'User who added patient to waiting list';
