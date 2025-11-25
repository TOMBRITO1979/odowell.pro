package main

import (
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

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	public := r.Group("/api")
	{
		public.POST("/tenants", handlers.CreateTenant)
		public.POST("/auth/login", handlers.Login)
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

		// Tenant Settings
		tenanted.GET("/settings", middleware.PermissionMiddleware("settings", "view"), handlers.GetTenantSettings)
		tenanted.PUT("/settings", middleware.PermissionMiddleware("settings", "edit"), handlers.UpdateTenantSettings)

		// User Management (admin only)
		users := tenanted.Group("/users")
		{
			users.GET("", handlers.GetTenantUsers)
			users.POST("", handlers.CreateTenantUser)
			users.PUT("/:id", handlers.UpdateTenantUser)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
