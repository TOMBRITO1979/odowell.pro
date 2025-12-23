package main

import (
	"context"
	"encoding/hex"

	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/handlers"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/metrics"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/scheduler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// serverStartTime tracks when the server started for uptime calculation
var serverStartTime time.Time

func init() {
	serverStartTime = time.Now()
}

func main() {
	// Validate required environment variables
	validateRequiredEnvVars()

	// Initialize timezone - critical for correct date/time handling
	initTimezone()

	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Initialize Redis cache
	if err := cache.Connect(); err != nil {
		log.Printf("WARNING: Redis connection failed: %v (cache disabled)", err)
		// Don't fail startup - app can work without cache
	}

	// Initialize Sentry for error tracking
	if err := helpers.InitSentry(); err != nil {
		log.Printf("WARNING: Sentry initialization failed: %v (error tracking disabled)", err)
	}
	defer helpers.CloseSentry()

	// Run migrations for all existing tenant schemas
	// This ensures new tables are created in all tenants on startup
	if err := database.RunAllMigrations(); err != nil {
		log.Printf("WARNING: Migration errors occurred: %v", err)
		// Don't fail startup, just log the warning
	}

	// Apply performance indexes and FK constraints
	// This is idempotent - safe to run on every startup
	if err := database.ApplyAllIndexesAndConstraints(); err != nil {
		log.Printf("WARNING: Index/constraint errors occurred: %v", err)
	}

	// Start background schedulers (trial expiration, data retention cleanup)
	scheduler.StartScheduler()

	// Create router
	r := gin.Default()

	// CORS - use origins from environment
	allowedOrigins := []string{"http://localhost:3000"}
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

	// Sentry middleware - captures panics and sends to Sentry
	r.Use(middleware.SentryMiddleware())

	// Request ID middleware - adds unique ID to each request for tracing
	r.Use(middleware.RequestIDMiddleware())

	// Security headers middleware - protects against common web vulnerabilities
	r.Use(middleware.SecurityHeadersMiddleware())

	// JSON structured logging middleware - logs all requests in JSON format
	r.Use(middleware.JSONLoggerMiddleware())

	// Prometheus metrics middleware - collects request metrics
	r.Use(metrics.PrometheusMiddleware())

	// Prometheus metrics endpoint
	r.GET("/metrics", metrics.MetricsHandler())

	// Health check with detailed status and metrics
	r.GET("/health", func(c *gin.Context) {
		pgStatus := "healthy"
		redisStatus := "healthy"
		pgError := ""
		redisError := ""
		var pgLatencyMs float64

		// Check PostgreSQL using direct ping with latency measurement
		pgStart := time.Now()
		err := database.Health()
		pgLatencyMs = float64(time.Since(pgStart).Microseconds()) / 1000.0
		if err != nil {
			pgStatus = "unhealthy"
			pgError = err.Error()
			log.Printf("Health check - PostgreSQL error: %v", err)
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

		// Get runtime metrics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Calculate uptime
		uptime := time.Since(serverStartTime)

		// Get database pool stats
		var dbPoolStats gin.H
		if sqlDB, err := database.GetDB().DB(); err == nil {
			stats := sqlDB.Stats()
			dbPoolStats = gin.H{
				"open_connections": stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
				"max_open":         stats.MaxOpenConnections,
				"wait_count":       stats.WaitCount,
				"wait_duration_ms": stats.WaitDuration.Milliseconds(),
			}
		}

		response := gin.H{
			"status": overallStatus,
			"services": gin.H{
				"postgres": pgStatus,
				"redis":    redisStatus,
			},
			"metrics": gin.H{
				"uptime_seconds":   int64(uptime.Seconds()),
				"uptime_human":     uptime.Round(time.Second).String(),
				"goroutines":       runtime.NumGoroutine(),
				"memory_alloc_mb":  float64(memStats.Alloc) / 1024 / 1024,
				"memory_sys_mb":    float64(memStats.Sys) / 1024 / 1024,
				"gc_runs":          memStats.NumGC,
				"pg_latency_ms":    pgLatencyMs,
				"db_pool":          dbPoolStats,
			},
			"version": gin.H{
				"go": runtime.Version(),
			},
		}

		// Add error details when unhealthy
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
		// Tenant registration (very strict rate limit: 3 per hour per IP)
		public.POST("/tenants", middleware.TenantRegistrationRateLimiter.RateLimitMiddleware(), handlers.CreateTenant)
		// Login with rate limiting: 5 attempts per minute, 15 min block (Redis distributed)
		public.POST("/auth/login", middleware.RedisLoginRateLimiter.RateLimitMiddleware(), handlers.Login)
		// Logout (clears httpOnly cookies and invalidates refresh token)
		public.POST("/auth/logout", handlers.Logout)
		// Refresh token (generates new access token using refresh token)
		public.POST("/auth/refresh", handlers.RefreshToken)
		// Email verification
		public.GET("/auth/verify-email", handlers.VerifyEmail)
		public.POST("/auth/resend-verification", handlers.ResendVerificationEmail)
		// Password reset (rate limited to prevent abuse)
		public.POST("/auth/forgot-password", middleware.ForgotPasswordRateLimiter.RateLimitMiddleware(), handlers.ForgotPassword)
		public.POST("/auth/reset-password", middleware.ForgotPasswordRateLimiter.RateLimitMiddleware(), handlers.ResetPassword)
		public.GET("/auth/validate-reset-token", handlers.ValidateResetToken)
		// 2FA verification during login (rate limited to prevent brute force)
		public.POST("/auth/2fa/verify", middleware.TwoFARateLimiter.RateLimitMiddleware(), handlers.Verify2FALogin)
	}

	// Static file serving for uploads
	r.Static("/uploads", "/root/uploads")

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(), middleware.AuditMiddleware())
	{
		protected.GET("/auth/me", handlers.GetMe)
		protected.PUT("/auth/profile", handlers.UpdateProfile)
		protected.PUT("/auth/password", handlers.ChangePassword)
		protected.POST("/auth/profile/picture", handlers.UploadProfilePicture)

		// Two-Factor Authentication (2FA) management
		twoFA := protected.Group("/auth/2fa")
		{
			twoFA.GET("/status", handlers.Get2FAStatus)
			twoFA.POST("/setup", handlers.Setup2FA)
			twoFA.POST("/confirm", handlers.Verify2FA)
			twoFA.POST("/disable", handlers.Disable2FA)
			twoFA.POST("/backup-codes", handlers.RegenerateBackupCodes)
		}

		// Digital Certificates (user-level, not tenant-level)
		certificates := protected.Group("/certificates")
		{
			certificates.POST("", handlers.UploadCertificate)
			certificates.GET("", handlers.GetUserCertificates)
			certificates.POST("/:id/activate", handlers.ActivateCertificate)
			certificates.POST("/:id/validate", handlers.ValidateCertificatePassword)
			certificates.DELETE("/:id", handlers.DeleteCertificate)
		}

		// Document signature verification
		protected.GET("/documents/:type/:id/verify", handlers.VerifyDocumentSignature)
	}

	// Tenant-scoped routes (subscription NOT required - for subscription management)
	tenantedNoSub := r.Group("/api")
	tenantedNoSub.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware(), middleware.AuditMiddleware())
	{
		// Subscription routes - always accessible
		subscription := tenantedNoSub.Group("/subscription")
		{
			subscription.GET("/plans", handlers.GetPlans)
			subscription.GET("/status", handlers.GetSubscriptionStatus)
			subscription.POST("/checkout", handlers.CreateCheckoutSession)
			subscription.POST("/portal", handlers.CreatePortalSession)
			subscription.POST("/cancel", handlers.CancelSubscription)
		}
	}

	// Tenant-scoped routes (subscription REQUIRED)
	tenanted := r.Group("/api")
	tenanted.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware(), middleware.SubscriptionMiddleware(), middleware.AuditMiddleware())
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
			// Digital Signature
			medicalRecords.POST("/:id/sign", middleware.PermissionMiddleware("medical_records", "edit"), handlers.SignMedicalRecord)
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
			// Digital Signature
			prescriptions.POST("/:id/sign", middleware.PermissionMiddleware("prescriptions", "edit"), handlers.SignPrescription)
			prescriptions.GET("/:id/pdf/signed", middleware.PermissionMiddleware("prescriptions", "view"), handlers.GenerateSignedPrescriptionPDF)
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
			payments.GET("/overdue-count", middleware.PermissionMiddleware("payments", "view"), handlers.GetOverduePaymentsCount)
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
			stockMovements.GET("/stats", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GetStockMovementStats)
			stockMovements.GET("/:id", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GetStockMovement)
			stockMovements.PUT("/:id", middleware.PermissionMiddleware("stock_movements", "edit"), handlers.UpdateStockMovement)
			stockMovements.DELETE("/:id", middleware.PermissionMiddleware("stock_movements", "delete"), handlers.DeleteStockMovement)
			// Export
			stockMovements.GET("/export/csv", middleware.PermissionMiddleware("stock_movements", "view"), handlers.ExportStockMovementsCSV)
			stockMovements.GET("/export/pdf", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GenerateStockMovementsListPDF)
			// Sale receipt
			stockMovements.GET("/:id/sale-receipt", middleware.PermissionMiddleware("stock_movements", "view"), handlers.GenerateSaleReceiptPDF)
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

		// Leads CRUD (CRM para WhatsApp e outras fontes)
		leads := tenanted.Group("/leads")
		{
			leads.GET("/check/:phone", middleware.PermissionMiddleware("leads", "view"), handlers.CheckLeadByPhone)
			leads.GET("/stats", middleware.PermissionMiddleware("leads", "view"), handlers.GetLeadStats)
			leads.POST("", middleware.PermissionMiddleware("leads", "create"), handlers.CreateLead)
			leads.GET("", middleware.PermissionMiddleware("leads", "view"), handlers.GetLeads)
			leads.GET("/:id", middleware.PermissionMiddleware("leads", "view"), handlers.GetLead)
			leads.PUT("/:id", middleware.PermissionMiddleware("leads", "edit"), handlers.UpdateLead)
			leads.DELETE("/:id", middleware.PermissionMiddleware("leads", "delete"), handlers.DeleteLead)
			leads.POST("/:id/convert", middleware.PermissionMiddleware("leads", "edit"), handlers.ConvertLeadToPatient)
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
		tenanted.DELETE("/settings/tenant", middleware.RoleMiddleware("admin"), handlers.DeleteOwnTenant) // Admin can delete own company
		tenanted.POST("/settings/smtp/test", middleware.PermissionMiddleware("settings", "edit"), handlers.TestSMTPConnection)

		// API Key Management (for WhatsApp/AI integrations)
		tenanted.POST("/settings/api-key/generate", middleware.PermissionMiddleware("settings", "edit"), handlers.GenerateAPIKey)
		tenanted.GET("/settings/api-key/status", middleware.PermissionMiddleware("settings", "view"), handlers.GetAPIKeyStatus)
		tenanted.PATCH("/settings/api-key/toggle", middleware.PermissionMiddleware("settings", "edit"), handlers.ToggleAPIKey)
		tenanted.DELETE("/settings/api-key", middleware.PermissionMiddleware("settings", "edit"), handlers.RevokeAPIKey)
		tenanted.GET("/settings/api-key/docs", middleware.PermissionMiddleware("settings", "view"), handlers.WhatsAppAPIDocumentation)

		// Embed Token Management (for Chatwell/external panels)
		tenanted.GET("/settings/embed-token", middleware.PermissionMiddleware("settings", "view"), handlers.GetEmbedToken)
		tenanted.POST("/settings/embed-token", middleware.PermissionMiddleware("settings", "edit"), handlers.GenerateEmbedToken)
		tenanted.DELETE("/settings/embed-token", middleware.PermissionMiddleware("settings", "edit"), handlers.RevokeEmbedToken)

		// User Management (admin only)
		users := tenanted.Group("/users")
		{
			users.GET("", middleware.PermissionMiddleware("users", "view"), handlers.GetTenantUsers)
			users.POST("", middleware.PermissionMiddleware("users", "create"), handlers.CreateTenantUser)
			users.PUT("/:id", middleware.PermissionMiddleware("users", "edit"), handlers.UpdateTenantUser)
			users.PATCH("/:id/sidebar", handlers.UpdateUserSidebar) // User can update own sidebar
		}

		// Permission Management (admin only)
		permissions := tenanted.Group("/permissions")
		permissions.Use(middleware.RoleMiddleware("admin"))
		{
			permissions.GET("/modules", handlers.GetModules)
			permissions.GET("/all", handlers.GetAllPermissions)
			permissions.GET("/users/:id", handlers.GetUserPermissions)
			permissions.PUT("/users/:id", handlers.UpdateUserPermissions)
			permissions.POST("/users/:id/bulk", handlers.BulkUpdateUserPermissions)
			permissions.GET("/defaults/:role", handlers.GetDefaultRolePermissions)
		}

		// Audit Logs (admin only) - LGPD compliance
		audit := tenanted.Group("/audit")
		audit.Use(middleware.PermissionMiddleware("audit_logs", "view"))
		{
			audit.GET("/logs", handlers.GetAuditLogs)
			audit.GET("/stats", handlers.GetAuditLogStats)
			audit.GET("/export/csv", handlers.ExportAuditLogsCSV)
		}

		// Data Requests (LGPD - Solicitacoes do Titular)
		dataRequests := tenanted.Group("/data-requests")
		{
			dataRequests.POST("", middleware.PermissionMiddleware("patients", "edit"), handlers.CreateDataRequest)
			dataRequests.GET("", middleware.PermissionMiddleware("patients", "view"), handlers.GetDataRequests)
			dataRequests.GET("/stats", middleware.PermissionMiddleware("patients", "view"), handlers.GetDataRequestStats)
			dataRequests.GET("/:id", middleware.PermissionMiddleware("patients", "view"), handlers.GetDataRequest)
			dataRequests.PATCH("/:id/status", middleware.PermissionMiddleware("patients", "edit"), handlers.UpdateDataRequestStatus)
			dataRequests.GET("/patient/:patient_id", middleware.PermissionMiddleware("patients", "view"), handlers.GetPatientDataRequests)
			// OTP Verification (LGPD identity verification)
			dataRequests.POST("/:id/send-otp", middleware.PermissionMiddleware("data_requests", "edit"), handlers.SendVerificationOTP)
			dataRequests.POST("/:id/verify-otp", middleware.PermissionMiddleware("data_requests", "edit"), handlers.VerifyOTP)
			// Data Export (LGPD portability)
			dataRequests.GET("/:id/export", middleware.PermissionMiddleware("data_requests", "view"), handlers.ExportPatientData)
		}

		// LGPD Data Deletion (admin only)
		lgpd := tenanted.Group("/lgpd")
		lgpd.Use(middleware.PermissionMiddleware("data_requests", "delete"))
		{
			lgpd.GET("/patients/:id/deletion-preview", handlers.GetPatientDeletionPreview)
			lgpd.DELETE("/patients/:id/permanent", handlers.PermanentDeletePatient)
			lgpd.POST("/patients/:id/anonymize", handlers.AnonymizePatient)
		}

		// Data Retention (admin only)
		retention := tenanted.Group("/retention")
		retention.Use(middleware.RoleMiddleware("admin"))
		{
			retention.GET("/stats", handlers.GetRetentionStats)
			retention.GET("/policy", handlers.GetRetentionPolicy)
			retention.POST("/cleanup", handlers.TriggerRetentionCleanup)
		}
	}

	// Super Admin routes (platform-level administration)
	superAdmin := r.Group("/api/admin")
	superAdmin.Use(middleware.AuthMiddleware())
	superAdmin.Use(middleware.SuperAdminMiddleware())
	{
		// Dashboard
		superAdmin.GET("/dashboard", handlers.GetAdminDashboard)

		// Tenant Management
		superAdmin.GET("/tenants", handlers.GetAllTenants)
		superAdmin.GET("/tenants/:id", handlers.GetTenantDetails)
		superAdmin.PATCH("/tenants/:id", handlers.UpdateTenantStatus)
		superAdmin.DELETE("/tenants/:id", handlers.DeleteTenant)
		superAdmin.GET("/tenants/:id/users", handlers.GetAdminTenantUsers)
		superAdmin.PATCH("/tenants/:id/users/:userId", handlers.UpdateTenantUserStatus)

		// Reports
		superAdmin.GET("/tenants/unverified", handlers.GetUnverifiedTenants)
		superAdmin.GET("/tenants/expiring", handlers.GetExpiringTrials)
		superAdmin.GET("/tenants/inactive", handlers.GetInactiveTenants)
	}

	// ==============================================
	// Patient Subscriptions (Planos - for patients)
	// ==============================================
	patientSubs := tenanted.Group("/patient-subscriptions")
	{
		patientSubs.GET("", middleware.PermissionMiddleware("plans", "view"), handlers.ListPatientSubscriptions)
		patientSubs.GET("/:id", middleware.PermissionMiddleware("plans", "view"), handlers.GetPatientSubscription)
		patientSubs.POST("", middleware.PermissionMiddleware("plans", "create"), handlers.CreatePatientSubscription)
		patientSubs.POST("/:id/cancel", middleware.PermissionMiddleware("plans", "edit"), handlers.CancelPatientSubscription)
		patientSubs.DELETE("/:id", middleware.PermissionMiddleware("plans", "delete"), handlers.CancelPatientSubscriptionImmediately)
		patientSubs.POST("/:id/refresh", middleware.PermissionMiddleware("plans", "view"), handlers.RefreshSubscriptionStatus)
		patientSubs.GET("/:id/payments", middleware.PermissionMiddleware("plans", "view"), handlers.GetSubscriptionPayments)
		patientSubs.POST("/:id/resend-link", middleware.PermissionMiddleware("plans", "edit"), handlers.ResendCheckoutLink)
	}

	// Stripe products for patient plans
	tenanted.GET("/stripe/products", middleware.PermissionMiddleware("plans", "view"), handlers.GetStripeProducts)

	// Tenant Stripe settings (for patient subscriptions)
	stripeSettings := tenanted.Group("/stripe-settings")
	{
		stripeSettings.GET("", middleware.PermissionMiddleware("settings", "view"), handlers.GetStripeSettings)
		stripeSettings.PUT("", middleware.PermissionMiddleware("settings", "edit"), handlers.UpdateStripeCredentials)
		stripeSettings.DELETE("", middleware.PermissionMiddleware("settings", "edit"), handlers.DisconnectStripe)
		stripeSettings.GET("/test", middleware.PermissionMiddleware("settings", "view"), handlers.TestStripeConnection)
		stripeSettings.GET("/webhook-url", middleware.PermissionMiddleware("settings", "view"), handlers.GenerateWebhookURL)
	}

	// NOTE: Tenant Subscription routes are defined above in tenantedNoSub group
	// (without SubscriptionMiddleware) so users can always access them to subscribe

	// ==============================================
	// Public Webhook Routes (no auth)
	// ==============================================
	// Webhook for patient subscriptions (per-tenant)
	r.POST("/api/v1/webhooks/stripe/:tenant_id", handlers.HandleStripeWebhook)

	// Webhook for tenant/clinic subscriptions (platform-level)
	r.POST("/api/webhooks/stripe", handlers.StripeWebhook)

	// WhatsApp/External API routes (API key authentication, tenant-isolated)
	// These endpoints are for external integrations like WhatsApp bots and AI assistants
	// Rate limited to 200 requests/minute per API key to prevent abuse (Redis distributed)
	whatsappAPI := r.Group("/api/whatsapp")
	whatsappAPI.Use(middleware.RedisWhatsAppRateLimiter.RateLimitMiddleware())
	whatsappAPI.Use(middleware.APIKeyMiddleware())
	{
		// Patient identity verification
		whatsappAPI.POST("/verify", handlers.WhatsAppVerifyIdentity)

		// Appointments
		whatsappAPI.GET("/appointments", handlers.WhatsAppGetAppointments)
		whatsappAPI.GET("/appointments/history", handlers.WhatsAppGetAppointmentHistory)
		whatsappAPI.GET("/appointments/by-dentist", handlers.WhatsAppGetDentistAppointments)
		whatsappAPI.POST("/appointments", handlers.WhatsAppCreateAppointment)
		whatsappAPI.POST("/appointments/cancel", handlers.WhatsAppCancelAppointment)
		whatsappAPI.POST("/appointments/reschedule", handlers.WhatsAppRescheduleAppointment)

		// Available time slots
		whatsappAPI.GET("/slots", handlers.WhatsAppGetAvailableSlots)

		// Waiting list
		whatsappAPI.POST("/waiting-list", handlers.WhatsAppAddToWaitingList)
		whatsappAPI.GET("/waiting-list", handlers.WhatsAppGetWaitingListStatus)
		whatsappAPI.DELETE("/waiting-list/:id", handlers.WhatsAppRemoveFromWaitingList)

		// Leads (CRM - verificar contato e criar lead)
		whatsappAPI.GET("/leads/check/:phone", handlers.CheckLeadByPhone)
		whatsappAPI.POST("/leads", handlers.WhatsAppCreateLead)
		whatsappAPI.PUT("/leads/:id", handlers.WhatsAppUpdateLead)
		whatsappAPI.POST("/leads/:id/convert", handlers.ConvertLeadToPatient)

		// Patients (criar paciente via API externa)
		whatsappAPI.POST("/patients", handlers.CreatePatient)

		// Reference data
		whatsappAPI.GET("/procedures", handlers.WhatsAppGetProcedures)
		whatsappAPI.GET("/dentists", handlers.WhatsAppGetDentists)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// initTimezone initializes the application timezone
// This ensures all time.Local operations use the correct timezone
func initTimezone() {
	// Get timezone from environment variable, default to America/Sao_Paulo
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "America/Sao_Paulo"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("WARNING: Failed to load timezone '%s': %v. Using UTC.", tz, err)
		return
	}

	time.Local = loc
	log.Printf("Timezone initialized: %s (current time: %s)", tz, time.Now().Format("2006-01-02 15:04:05 MST"))
}

// validateRequiredEnvVars ensures critical environment variables are set
func validateRequiredEnvVars() {
	required := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"ENCRYPTION_KEY",
	}

	missing := []string{}
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("FATAL: Missing required environment variables: %v", missing)
	}

	// Validate JWT_SECRET has minimum length (32 bytes recommended)
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatalf("FATAL: JWT_SECRET must be at least 32 characters long (current: %d)", len(jwtSecret))
	}

	// Validate ENCRYPTION_KEY format (must be 64 hex chars = 32 bytes for AES-256)
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if len(encryptionKey) != 64 {
		log.Fatalf("FATAL: ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes for AES-256), current: %d", len(encryptionKey))
	}
	// Validate it's valid hex
	if _, err := hex.DecodeString(encryptionKey); err != nil {
		log.Fatalf("FATAL: ENCRYPTION_KEY must be valid hexadecimal: %v", err)
	}

	log.Println("Environment variables validated successfully")
}
