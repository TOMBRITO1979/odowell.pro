-- Migration: Add WhatsApp Business API columns to tenant_settings
-- This migration adds columns required for Meta WhatsApp Business API integration

-- Add WhatsApp Business API columns
ALTER TABLE public.tenant_settings
ADD COLUMN IF NOT EXISTS whatsapp_phone_number_id VARCHAR(50),
ADD COLUMN IF NOT EXISTS whatsapp_access_token TEXT,
ADD COLUMN IF NOT EXISTS whatsapp_business_account_id VARCHAR(50),
ADD COLUMN IF NOT EXISTS whatsapp_webhook_verify_token VARCHAR(100),
ADD COLUMN IF NOT EXISTS whatsapp_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS whatsapp_template_confirmation VARCHAR(100),
ADD COLUMN IF NOT EXISTS whatsapp_template_reminder VARCHAR(100),
ADD COLUMN IF NOT EXISTS whatsapp_template_reminder_hours INTEGER DEFAULT 24;

-- Add comment for documentation
COMMENT ON COLUMN public.tenant_settings.whatsapp_phone_number_id IS 'Meta WhatsApp Business Phone Number ID';
COMMENT ON COLUMN public.tenant_settings.whatsapp_access_token IS 'Meta WhatsApp Business API Access Token (encrypted)';
COMMENT ON COLUMN public.tenant_settings.whatsapp_business_account_id IS 'Meta WhatsApp Business Account ID';
COMMENT ON COLUMN public.tenant_settings.whatsapp_webhook_verify_token IS 'Token for verifying Meta webhook callbacks';
COMMENT ON COLUMN public.tenant_settings.whatsapp_enabled IS 'Whether WhatsApp Business integration is enabled';
COMMENT ON COLUMN public.tenant_settings.whatsapp_template_confirmation IS 'Template name for appointment confirmations';
COMMENT ON COLUMN public.tenant_settings.whatsapp_template_reminder IS 'Template name for appointment reminders';
COMMENT ON COLUMN public.tenant_settings.whatsapp_template_reminder_hours IS 'Hours before appointment to send reminder';
