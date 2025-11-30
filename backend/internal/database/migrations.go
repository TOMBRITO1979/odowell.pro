package database

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MigrateAllTenants runs migrations for all existing tenant schemas
// This ensures that new tables are created in all tenants when the application starts
func MigrateAllTenants() error {
	log.Println("Starting migrations for all tenants...")

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

	if len(schemas) == 0 {
		log.Println("No tenant schemas found, skipping migrations")
		return nil
	}

	log.Printf("Found %d tenant schemas to migrate", len(schemas))

	// Migrate each tenant schema
	successCount := 0
	for _, schema := range schemas {
		log.Printf("Migrating schema: %s", schema)

		// Run migrations for new tables only (without foreign key constraints)
		if err := migrateNewTablesOnly(schema); err != nil {
			log.Printf("ERROR: Failed to migrate %s: %v", schema, err)
			continue
		}

		successCount++
		log.Printf("Successfully migrated schema: %s", schema)
	}

	// Reset search path to public
	DB.Exec("SET search_path TO public")

	log.Printf("Migration completed: %d/%d schemas migrated successfully", successCount, len(schemas))

	return nil
}

// migrateNewTablesOnly creates only new tables without touching existing ones
// This avoids foreign key constraint issues with existing data
func migrateNewTablesOnly(schema string) error {
	// Get raw SQL connection to build a new GORM instance with FK constraints disabled
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Create a new GORM connection with FK constraints disabled for migration
	migrationDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true, // This is the key!
	})
	if err != nil {
		return fmt.Errorf("failed to create migration DB: %v", err)
	}

	// Set search path for this schema
	if err := migrationDB.Exec(fmt.Sprintf("SET search_path TO %s", schema)).Error; err != nil {
		return fmt.Errorf("failed to set search_path: %v", err)
	}

	// Only migrate tables that might be new or have new columns
	// These are the tables added recently that may not exist in older tenants
	err = migrationDB.AutoMigrate(
		&models.Treatment{},
		&models.TreatmentPayment{},
		&models.ConsentTemplate{},
		&models.PatientConsent{},
		&models.Prescription{}, // Added for new signer fields
		&models.Appointment{},  // Added for new room field
	)

	return err
}

// autoMigrateTenantSchema creates/updates all tables in a tenant schema
// Used for NEW tenants only (not existing ones)
func autoMigrateTenantSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		// Core tables
		&models.Patient{},
		&models.Appointment{},
		&models.MedicalRecord{},

		// Financial tables
		&models.Budget{},
		&models.Payment{},
		&models.Commission{},
		&models.Treatment{},
		&models.TreatmentPayment{},

		// Inventory tables
		&models.Product{},
		&models.Supplier{},
		&models.StockMovement{},

		// Marketing tables
		&models.Campaign{},
		&models.CampaignRecipient{},

		// Document tables
		&models.Attachment{},
		&models.Exam{},
		&models.Prescription{},

		// Clinical tables
		&models.TreatmentProtocol{},
		&models.ConsentTemplate{},
		&models.PatientConsent{},

		// Utility tables
		&models.Task{},
		&models.TenantSettings{},
		&models.WaitingList{},
	)
}

// MigratePublicSchema runs migrations for the public schema (tenants, users, permissions)
func MigratePublicSchema() error {
	log.Println("Migrating public schema...")

	// Get raw SQL connection to build a new GORM instance with FK constraints disabled
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Create a new GORM connection with FK constraints disabled for migration
	migrationDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration DB: %v", err)
	}

	// Set search path to public
	if err := migrationDB.Exec("SET search_path TO public").Error; err != nil {
		return fmt.Errorf("failed to set search_path to public: %v", err)
	}

	// Run migrations for public tables
	err = migrationDB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Module{},
		&models.Permission{},
		&models.UserPermission{},
		&models.AuditLog{},
		&models.EmailVerification{},
		&models.PasswordReset{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate public schema: %v", err)
	}

	log.Println("Public schema migrated successfully")
	return nil
}

// RunAllMigrations runs migrations for public schema and all tenant schemas
func RunAllMigrations() error {
	// First migrate public schema
	if err := MigratePublicSchema(); err != nil {
		return err
	}

	// Then migrate all tenant schemas
	if err := MigrateAllTenants(); err != nil {
		return err
	}

	return nil
}
