package main

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/handlers"
	"drcrwell/backend/internal/middleware"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Initialize Redis cache
	if err := cache.Connect(); err != nil {
		log.Printf("WARNING: Redis connection failed: %v (cache disabled)", err)
		// Don't fail startup - app can work without cache
	}

	// Run migrations for all existing tenant schemas
	// This ensures new tables are created in all tenants on startup
	if err := database.RunAllMigrations(); err != nil {
		log.Printf("WARNING: Migration errors occurred: %v", err)
		// Don't fail startup, just log the warning
	}

	// Create router
	r := gin.Default()

	// CORS - use origins from environment
	allowedOrigins := []string{"https://dr.crwell.pro"}
	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		allowedOrigins = strings.Split(corsOrigins, ",")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check with detailed status
	r.GET("/health", func(c *gin.Context) {
		pgStatus := "healthy"
		redisStatus := "healthy"
		pgError := ""
		redisError := ""

		// Check PostgreSQL using direct ping
		err := database.Health()
		if err != nil {
			pgStatus = "unhealthy"
			pgError = err.Error()
			log.Printf("Health check - PostgreSQL error: %v", err)
		} else {
			log.Printf("Health check - PostgreSQL OK")
		}

		// Check Redis
		if err := cache.Health(); err != nil {
			redisStatus = "unhealthy"
			redisError = err.Error()
		} else if cache.GetClient() == nil {
			redisStatus = "unhealthy"
			redisError = "client is nil"
		}

		overallStatus := "ok"
		if pgStatus == "unhealthy" || redisStatus == "unhealthy" {
			overallStatus = "degraded"
		}

		response := gin.H{
			"status": overallStatus,
			"services": gin.H{
				"postgres": pgStatus,
				"redis":    redisStatus,
			},
		}

		// Add error details in debug mode or when unhealthy
		if pgError != "" || redisError != "" {
			response["errors"] = gin.H{}
			if pgError != "" {
				response["errors"].(gin.H)["postgres"] = pgError
			}
			if redisError != "" {
				response["errors"].(gin.H)["redis"] = redisError
			}
		}

		httpStatus := 200
		if overallStatus != "ok" {
			httpStatus = 503
		}
		c.JSON(httpStatus, response)
	})

	// Public routes
	public := r.Group("/api")
	{
		public.POST("/tenants", handlers.CreateTenant)
		// Login with rate limiting: 5 attempts per minute, 15 min block
		public.POST("/auth/login", middleware.LoginRateLimiter.RateLimitMiddleware(), handlers.Login)
		// Email verification
		public.GET("/auth/verify-email", handlers.VerifyEmail)
		public.POST("/auth/resend-verification", handlers.ResendVerificationEmail)
		// Password reset
		public.POST("/auth/forgot-password", handlers.ForgotPassword)
		public.POST("/auth/reset-password", handlers.ResetPassword)
		public.GET("/auth/validate-reset-token", handlers.ValidateResetToken)
	}

	// Static file serving for uploads
	r.Static("/uploads", "/root/uploads")

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/auth/me", handlers.GetMe)
		protected.PUT("/auth/profile", handlers.UpdateProfile)
		protected.PUT("/auth/password", handlers.ChangePassword)
		protected.POST("/auth/profile/picture", handlers.UploadProfilePicture)
	}

	// Tenant-scoped routes
	tenanted := r.Group("/api")
	tenanted.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware())
	{
		// Patients CRUD
		patients := tenanted.Group("/patients")
		{
			patients.POST("", middleware.PermissionMiddleware("patients", "create"), handlers.CreatePatient)
			patients.GET("", middleware.PermissionMiddleware("patients", "view"), handlers.GetPatients)
			patients.GET("/:id", middleware.PermissionMiddleware("patients", "view"), handlers.GetPatient)
			patients.PUT("/:id", middleware.PermissionMiddleware("patients", "edit"), handlers.UpdatePatient)
			patients.DELETE("/:id", middleware.PermissionMiddleware("patients", "delete"), handlers.DeletePatient)
			// Export/Import
			patients.GET("/export/csv", middleware.PermissionMiddleware("patients", "view"), handlers.ExportPatientsCSV)
			patients.POST("/import/csv", middleware.PermissionMiddleware("patients", "create"), handlers.ImportPatientsCSV)
			patients.GET("/export/pdf", middleware.PermissionMiddleware("patients", "view"), handlers.GeneratePatientsListPDF)
		}

		// Appointments CRUD
		appointments := tenanted.Group("/appointments")
		{
			appointments.POST("", middleware.PermissionMiddleware("appointments", "create"), handlers.CreateAppointment)
			appointments.GET("", middleware.PermissionMiddleware("appointments", "view"), handlers.GetAppointments)
			appointments.GET("/:id", middleware.PermissionMiddleware("appointments", "view"), handlers.GetAppointment)
			appointments.PUT("/:id", middleware.PermissionMiddleware("appointments", "edit"), handlers.UpdateAppointment)
			appointments.DELETE("/:id", middleware.PermissionMiddleware("appointments", "delete"), handlers.DeleteAppointment)
			appointments.PATCH("/:id/status", middleware.PermissionMiddleware("appointments", "edit"), handlers.UpdateAppointmentStatus)
			// Export
			appointments.GET("/export/csv", middleware.PermissionMiddleware("appointments", "view"), handlers.ExportAppointmentsCSV)
			appointments.GET("/export/pdf", middleware.PermissionMiddleware("appointments", "view"), handlers.GenerateAppointmentsListPDF)
		}

		// Medical Records CRUD
		medicalRecords := tenanted.Group("/medical-records")
		{
			medicalRecords.POST("", middleware.PermissionMiddleware("medical_records", "create"), handlers.CreateMedicalRecord)
			medicalRecords.GET("", middleware.PermissionMiddleware("medical_records", "view"), handlers.GetMedicalRecords)
			medicalRecords.GET("/:id", middleware.PermissionMiddleware("medical_records", "view"), handlers.GetMedicalRecord)
			medicalRecords.PUT("/:id", middleware.PermissionMiddleware("medical_records", "edit"), handlers.UpdateMedicalRecord)
			medicalRecords.DELETE("/:id", middleware.PermissionMiddleware("medical_records", "delete"), handlers.DeleteMedicalRecord)
			medicalRecords.GET("/:id/pdf", middleware.PermissionMiddleware("medical_records", "view"), handlers.GenerateMedicalRecordPDF)
		}

		// Prescriptions CRUD (Receituário)
		prescriptions := tenanted.Group("/prescriptions")
		{
			prescriptions.POST("", middleware.PermissionMiddleware("prescriptions", "create"), handlers.CreatePrescription)
			prescriptions.GET("", middleware.PermissionMiddleware("prescriptions", "view"), handlers.GetPrescriptions)
			prescriptions.GET("/:id", middleware.PermissionMiddleware("prescriptions", "view"), handlers.GetPrescription)
			prescriptions.PUT("/:id", middleware.PermissionMiddleware("prescriptions", "edit"), handlers.UpdatePrescription)
			prescriptions.DELETE("/:id", middleware.PermissionMiddleware("prescriptions", "delete"), handlers.DeletePrescription)
			prescriptions.POST("/:id/issue", middleware.PermissionMiddleware("prescriptions", "edit"), handlers.IssuePrescription)
			prescriptions.POST("/:id/print", middleware.PermissionMiddleware("prescriptions", "view"), handlers.PrintPrescription)
			prescriptions.GET("/:id/pdf", middleware.PermissionMiddleware("prescriptions", "view"), handlers.GeneratePrescriptionPDF)
		}

		// Budgets CRUD
		budgets := tenanted.Group("/budgets")
		{
			budgets.POST("", middleware.PermissionMiddleware("budgets", "create"), handlers.CreateBudget)
			budgets.GET("", middleware.PermissionMiddleware("budgets", "view"), handlers.GetBudgets)
			budgets.GET("/:id", middleware.PermissionMiddleware("budgets", "view"), handlers.GetBudget)
			budgets.PUT("/:id", middleware.PermissionMiddleware("budgets", "edit"), handlers.UpdateBudget)
			budgets.DELETE("/:id", middleware.PermissionMiddleware("budgets", "delete"), handlers.DeleteBudget)
			budgets.POST("/:id/cancel", middleware.PermissionMiddleware("budgets", "edit"), handlers.CancelBudget)
			budgets.GET("/:id/pdf", middleware.PermissionMiddleware("budgets", "view"), handlers.GenerateBudgetPDF)
			budgets.GET("/:id/payments-pdf", middleware.PermissionMiddleware("budgets", "view"), handlers.GenerateBudgetPaymentsPDF)
			budgets.GET("/:id/payment/:payment_id/receipt", middleware.PermissionMiddleware("budgets", "view"), handlers.GeneratePaymentReceipt)
			// Export/Import
			budgets.GET("/export/csv", middleware.PermissionMiddleware("budgets", "view"), handlers.ExportBudgetsCSV)
			budgets.POST("/import/csv", middleware.PermissionMiddleware("budgets", "create"), handlers.ImportBudgetsCSV)
			budgets.GET("/export/pdf", middleware.PermissionMiddleware("budgets", "view"), handlers.GenerateBudgetsListPDF)
		}

		// Payments CRUD
		payments := tenanted.Group("/payments")
		{
			payments.POST("", middleware.PermissionMiddleware("payments", "create"), handlers.CreatePayment)
			payments.GET("", middleware.PermissionMiddleware("payments", "view"), handlers.GetPayments)
			payments.GET("/:id", middleware.PermissionMiddleware("payments", "view"), handlers.GetPayment)
			payments.PUT("/:id", middleware.PermissionMiddleware("payments", "edit"), handlers.UpdatePayment)
			payments.DELETE("/:id", middleware.PermissionMiddleware("payments", "delete"), handlers.DeletePayment)
			payments.POST("/:id/refund", middleware.PermissionMiddleware("payments", "edit"), handlers.RefundPayment)
			payments.GET("/cashflow", middleware.PermissionMiddleware("payments", "view"), handlers.GetCashFlow)
			payments.GET("/pdf/export", middleware.PermissionMiddleware("payments", "view"), handlers.GeneratePaymentsPDF)
			// Export/Import
			payments.GET("/export/csv", middleware.PermissionMiddleware("payments", "view"), handlers.ExportPaymentsCSV)
			payments.POST("/import/csv", middleware.PermissionMiddleware("payments", "create"), handlers.ImportPaymentsCSV)
		}

		// Products CRUD
		products := tenanted.Group("/products")
		{
			products.POST("", middleware.PermissionMiddleware("products", "create"), handlers.CreateProduct)
			products.GET("", middleware.PermissionMiddleware("products", "view"), handlers.GetProducts)
			products.GET("/:id", middleware.PermissionMiddleware("products", "view"), handlers.GetProduct)
			products.PUT("/:id", middleware.PermissionMiddleware("products", "edit"), handlers.UpdateProduct)
			products.DELETE("/:id", middleware.PermissionMiddleware("products", "delete"), handlers.DeleteProduct)
			products.GET("/low-stock", middleware.PermissionMiddleware("products", "view"), handlers.GetLowStockProducts)
			// Export/Import
			products.GET("/export/csv", middleware.PermissionMiddleware("products", "view"), handlers.ExportProductsCSV)
			products.POST("/import/csv", middleware.PermissionMiddleware("products", "create"), handlers.ImportProductsCSV)
			products.GET("/export/pdf", middleware.PermissionMiddleware("products", "view"), handlers.GenerateProductsListPDF)
		}

		// Suppliers CRUD
		suppliers := tenanted.Group("/suppliers")
		{
			suppliers.POST("", middleware.PermissionMiddleware("suppliers", "create"), handlers.CreateSupplier)
			suppliers.GET("", middleware.PermissionMiddleware("suppliers", "view"), handlers.GetSuppliers)
			suppliers.GET("/:id", middleware.PermissionMiddleware("suppliers", "view"), handlers.GetSupplier)
			suppliers.PUT("/:id", middleware.PermissionMiddleware("suppliers", "edit"), handlers.UpdateSupplier)
			suppliers.DELETE("/:id", middleware.PermissionMiddleware("suppliers", "delete"), handlers.DeleteSupplier)
			// Export/Import
			suppliers.GET("/export/csv", middleware.PermissionMiddleware("suppliers", "view"), handlers.ExportSuppliersCSV)
			suppliers.POST("/import/csv", middleware.PermissionMiddleware("suppliers", "create"), handlers.ImportSuppliersCSV)
			suppliers.GET("/export/pdf", middleware.PermissionMiddleware("suppliers", "view"), handlers.GenerateSuppliersListPDF)
		}

		// Stock Movements
		stockMovements := tenanted.Group("/stock-movements")
		{
			stockMovements.POST("", middleware.PermissionMiddleware("stock_movements", "create"), handlers.CreateStockMovement)
			stockMovements.GET("", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GetStockMovements)
			// Export
			stockMovements.GET("/export/csv", middleware.PermissionMiddleware("stock_movements", "view"), handlers.ExportStockMovementsCSV)
			stockMovements.GET("/export/pdf", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GenerateStockMovementsListPDF)
		}

		// Dashboard and Reports
		reports := tenanted.Group("/reports")
		{
			reports.GET("/dashboard", middleware.PermissionMiddleware("reports", "view"), handlers.GetDashboard)
			reports.GET("/dashboard/advanced", middleware.PermissionMiddleware("reports", "view"), handlers.GetAdvancedDashboard)
			reports.GET("/dashboard/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateDashboardPDF)
			reports.GET("/revenue", middleware.PermissionMiddleware("reports", "view"), handlers.GetRevenueReport)
			reports.GET("/procedures", middleware.PermissionMiddleware("reports", "view"), handlers.GetProceduresReport)
			reports.GET("/attendance", middleware.PermissionMiddleware("reports", "view"), handlers.GetAttendanceReport)
			reports.GET("/budget-conversion", middleware.PermissionMiddleware("reports", "view"), handlers.GetBudgetConversionReport)
			reports.GET("/overdue-payments", middleware.PermissionMiddleware("reports", "view"), handlers.GetOverduePaymentsReport)
			reports.GET("/revenue/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateRevenuePDF)
			reports.GET("/attendance/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateAttendancePDF)
			reports.GET("/procedures/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateProceduresPDF)
			reports.GET("/budget-conversion/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateBudgetConversionPDF)
			reports.GET("/overdue-payments/pdf", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateOverduePaymentsPDF)
			reports.GET("/revenue/excel", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateRevenueExcel)
			reports.GET("/attendance/excel", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateAttendanceExcel)
			reports.GET("/procedures/excel", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateProceduresExcel)
			reports.GET("/budget-conversion/excel", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateBudgetConversionExcel)
			reports.GET("/overdue-payments/excel", middleware.PermissionMiddleware("reports", "view"), handlers.GenerateOverduePaymentsExcel)
		}

		// Campaigns CRUD
		campaigns := tenanted.Group("/campaigns")
		{
			campaigns.POST("", middleware.PermissionMiddleware("campaigns", "create"), handlers.CreateCampaign)
			campaigns.GET("", middleware.PermissionMiddleware("campaigns", "view"), handlers.GetCampaigns)
			campaigns.GET("/:id", middleware.PermissionMiddleware("campaigns", "view"), handlers.GetCampaign)
			campaigns.PUT("/:id", middleware.PermissionMiddleware("campaigns", "edit"), handlers.UpdateCampaign)
			campaigns.DELETE("/:id", middleware.PermissionMiddleware("campaigns", "delete"), handlers.DeleteCampaign)
			campaigns.POST("/:id/send", middleware.PermissionMiddleware("campaigns", "edit"), handlers.SendCampaign)
		}

		// Exams CRUD
		exams := tenanted.Group("/exams")
		{
			exams.POST("", middleware.PermissionMiddleware("exams", "create"), handlers.CreateExam)
			exams.GET("", middleware.PermissionMiddleware("exams", "view"), handlers.GetExams)
			exams.GET("/:id", middleware.PermissionMiddleware("exams", "view"), handlers.GetExam)
			exams.PUT("/:id", middleware.PermissionMiddleware("exams", "edit"), handlers.UpdateExam)
			exams.DELETE("/:id", middleware.PermissionMiddleware("exams", "delete"), handlers.DeleteExam)
			exams.GET("/:id/download", middleware.PermissionMiddleware("exams", "view"), handlers.GetExamDownloadURL)
		}

		// Tasks CRUD
		tasks := tenanted.Group("/tasks")
		{
			tasks.POST("", middleware.PermissionMiddleware("tasks", "create"), handlers.CreateTask)
			tasks.GET("", middleware.PermissionMiddleware("tasks", "view"), handlers.GetTasks)
			tasks.GET("/pending-count", middleware.PermissionMiddleware("tasks", "view"), handlers.GetPendingCount)
			tasks.GET("/:id", middleware.PermissionMiddleware("tasks", "view"), handlers.GetTask)
			tasks.PUT("/:id", middleware.PermissionMiddleware("tasks", "edit"), handlers.UpdateTask)
			tasks.DELETE("/:id", middleware.PermissionMiddleware("tasks", "delete"), handlers.DeleteTask)
		}

		// Waiting List CRUD
		waitingList := tenanted.Group("/waiting-list")
		{
			waitingList.POST("", middleware.PermissionMiddleware("appointments", "create"), handlers.CreateWaitingListEntry)
			waitingList.GET("", middleware.PermissionMiddleware("appointments", "view"), handlers.GetWaitingList)
			waitingList.GET("/stats", middleware.PermissionMiddleware("appointments", "view"), handlers.GetWaitingListStats)
			waitingList.GET("/:id", middleware.PermissionMiddleware("appointments", "view"), handlers.GetWaitingListEntry)
			waitingList.PUT("/:id", middleware.PermissionMiddleware("appointments", "edit"), handlers.UpdateWaitingListEntry)
			waitingList.POST("/:id/contact", middleware.PermissionMiddleware("appointments", "edit"), handlers.ContactWaitingListEntry)
			waitingList.POST("/:id/schedule", middleware.PermissionMiddleware("appointments", "edit"), handlers.ScheduleWaitingListEntry)
			waitingList.DELETE("/:id", middleware.PermissionMiddleware("appointments", "delete"), handlers.DeleteWaitingListEntry)
		}

		// Treatment Protocols CRUD
		protocols := tenanted.Group("/treatment-protocols")
		{
			protocols.POST("", middleware.PermissionMiddleware("clinical_records", "create"), handlers.CreateTreatmentProtocol)
			protocols.GET("", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetTreatmentProtocols)
			protocols.GET("/:id", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetTreatmentProtocol)
			protocols.PUT("/:id", middleware.PermissionMiddleware("clinical_records", "edit"), handlers.UpdateTreatmentProtocol)
			protocols.DELETE("/:id", middleware.PermissionMiddleware("clinical_records", "delete"), handlers.DeleteTreatmentProtocol)
		}

		// Consent Templates CRUD
		consentTemplates := tenanted.Group("/consent-templates")
		{
			consentTemplates.POST("", middleware.PermissionMiddleware("clinical_records", "create"), handlers.CreateConsentTemplate)
			consentTemplates.GET("", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetConsentTemplates)
			consentTemplates.GET("/types", handlers.GetConsentTypes)
			consentTemplates.GET("/:id", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetConsentTemplate)
			consentTemplates.GET("/:id/pdf", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GenerateTemplatePDF)
			consentTemplates.PUT("/:id", middleware.PermissionMiddleware("clinical_records", "edit"), handlers.UpdateConsentTemplate)
			consentTemplates.DELETE("/:id", middleware.PermissionMiddleware("clinical_records", "delete"), handlers.DeleteConsentTemplate)
		}

		// Patient Consents
		consents := tenanted.Group("/consents")
		{
			consents.POST("", middleware.PermissionMiddleware("clinical_records", "create"), handlers.CreatePatientConsent)
			consents.GET("/patients/:patient_id", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetPatientConsents)
			consents.GET("/:id", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GetConsent)
			consents.GET("/:id/pdf", middleware.PermissionMiddleware("clinical_records", "view"), handlers.GenerateConsentPDF)
			consents.PATCH("/:id/status", middleware.PermissionMiddleware("clinical_records", "edit"), handlers.UpdateConsentStatus)
			consents.DELETE("/:id", middleware.PermissionMiddleware("clinical_records", "delete"), handlers.DeleteConsent)
		}

		// Treatments CRUD (orçamentos aprovados em tratamento)
		treatments := tenanted.Group("/treatments")
		{
			treatments.POST("", middleware.PermissionMiddleware("budgets", "create"), handlers.CreateTreatment)
			treatments.GET("", middleware.PermissionMiddleware("budgets", "view"), handlers.GetTreatments)
			treatments.GET("/:id", middleware.PermissionMiddleware("budgets", "view"), handlers.GetTreatment)
			treatments.PUT("/:id", middleware.PermissionMiddleware("budgets", "edit"), handlers.UpdateTreatment)
			treatments.DELETE("/:id", middleware.PermissionMiddleware("budgets", "delete"), handlers.DeleteTreatment)
		}

		// Treatment Payments (pagamentos de tratamentos)
		treatmentPayments := tenanted.Group("/treatment-payments")
		{
			treatmentPayments.POST("", middleware.PermissionMiddleware("payments", "create"), handlers.CreateTreatmentPayment)
			treatmentPayments.GET("/treatment/:treatment_id", middleware.PermissionMiddleware("payments", "view"), handlers.GetTreatmentPayments)
			treatmentPayments.GET("/:id/receipt", middleware.PermissionMiddleware("payments", "view"), handlers.GenerateReceiptPDF)
			treatmentPayments.GET("/:id", middleware.PermissionMiddleware("payments", "view"), handlers.GetTreatmentPayment)
			treatmentPayments.PUT("/:id", middleware.PermissionMiddleware("payments", "edit"), handlers.UpdateTreatmentPayment)
			treatmentPayments.DELETE("/:id", middleware.PermissionMiddleware("payments", "delete"), handlers.DeleteTreatmentPayment)
		}

		// Tenant Settings
		tenanted.GET("/settings", middleware.PermissionMiddleware("settings", "view"), handlers.GetTenantSettings)
		tenanted.PUT("/settings", middleware.PermissionMiddleware("settings", "edit"), handlers.UpdateTenantSettings)

		// API Key Management (for WhatsApp/AI integrations)
		tenanted.POST("/settings/api-key/generate", handlers.GenerateAPIKey)
		tenanted.GET("/settings/api-key/status", handlers.GetAPIKeyStatus)
		tenanted.PATCH("/settings/api-key/toggle", handlers.ToggleAPIKey)
		tenanted.DELETE("/settings/api-key", handlers.RevokeAPIKey)
		tenanted.GET("/settings/api-key/docs", handlers.WhatsAppAPIDocumentation)

		// User Management (admin only)
		users := tenanted.Group("/users")
		{
			users.GET("", handlers.GetTenantUsers)
			users.POST("", handlers.CreateTenantUser)
			users.PUT("/:id", handlers.UpdateTenantUser)
			users.PATCH("/:id/sidebar", handlers.UpdateUserSidebar)
		}

		// Permission Management (admin only)
		permissions := tenanted.Group("/permissions")
		{
			permissions.GET("/modules", handlers.GetModules)
			permissions.GET("/all", handlers.GetAllPermissions)
			permissions.GET("/users/:id", handlers.GetUserPermissions)
			permissions.PUT("/users/:id", handlers.UpdateUserPermissions)
			permissions.POST("/users/:id/bulk", handlers.BulkUpdateUserPermissions)
			permissions.GET("/defaults/:role", handlers.GetDefaultRolePermissions)
		}
	}

	// WhatsApp/External API routes (API key authentication, tenant-isolated)
	// These endpoints are for external integrations like WhatsApp bots and AI assistants
	whatsappAPI := r.Group("/api/whatsapp")
	whatsappAPI.Use(middleware.APIKeyMiddleware())
	{
		// Patient identity verification
		whatsappAPI.POST("/verify", handlers.WhatsAppVerifyIdentity)

		// Appointments
		whatsappAPI.GET("/appointments", handlers.WhatsAppGetAppointments)
		whatsappAPI.GET("/appointments/history", handlers.WhatsAppGetAppointmentHistory)
		whatsappAPI.POST("/appointments/cancel", handlers.WhatsAppCancelAppointment)
		whatsappAPI.POST("/appointments/reschedule", handlers.WhatsAppRescheduleAppointment)

		// Available time slots
		whatsappAPI.GET("/slots", handlers.WhatsAppGetAvailableSlots)

		// Waiting list
		whatsappAPI.POST("/waiting-list", handlers.WhatsAppAddToWaitingList)
		whatsappAPI.GET("/waiting-list", handlers.WhatsAppGetWaitingListStatus)
		whatsappAPI.DELETE("/waiting-list/:id", handlers.WhatsAppRemoveFromWaitingList)

		// Reference data
		whatsappAPI.GET("/procedures", handlers.WhatsAppGetProcedures)
		whatsappAPI.GET("/dentists", handlers.WhatsAppGetDentists)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
