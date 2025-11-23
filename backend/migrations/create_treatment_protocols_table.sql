-- Migration: Create treatment_protocols table
-- This table stores treatment protocol templates

CREATE TABLE IF NOT EXISTS treatment_protocols (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name TEXT NOT NULL,
    description TEXT,
    procedures JSONB, -- Array of procedure objects [{name, description, order}]
    duration INTEGER DEFAULT 0, -- Estimated duration in minutes
    cost DECIMAL(10,2) DEFAULT 0.00, -- Estimated cost
    active BOOLEAN DEFAULT true,
    created_by BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_treatment_protocols_deleted_at ON treatment_protocols(deleted_at);
CREATE INDEX IF NOT EXISTS idx_treatment_protocols_active ON treatment_protocols(active);
CREATE INDEX IF NOT EXISTS idx_treatment_protocols_name ON treatment_protocols(name);
