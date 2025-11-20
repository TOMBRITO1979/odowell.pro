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

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/auth/me", handlers.GetMe)
		protected.PUT("/auth/profile", handlers.UpdateProfile)
		protected.PUT("/auth/password", handlers.ChangePassword)
	}

	// Tenant-scoped routes
	tenanted := r.Group("/api")
	tenanted.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware())
	{
		// Patients CRUD
		patients := tenanted.Group("/patients")
		{
			patients.POST("", handlers.CreatePatient)
			patients.GET("", handlers.GetPatients)
			patients.GET("/:id", handlers.GetPatient)
			patients.PUT("/:id", handlers.UpdatePatient)
			patients.DELETE("/:id", handlers.DeletePatient)
		}

		// Appointments CRUD
		appointments := tenanted.Group("/appointments")
		{
			appointments.POST("", handlers.CreateAppointment)
			appointments.GET("", handlers.GetAppointments)
			appointments.GET("/:id", handlers.GetAppointment)
			appointments.PUT("/:id", handlers.UpdateAppointment)
			appointments.DELETE("/:id", handlers.DeleteAppointment)
			appointments.PATCH("/:id/status", handlers.UpdateAppointmentStatus)
		}

		// Medical Records CRUD
		medicalRecords := tenanted.Group("/medical-records")
		{
			medicalRecords.POST("", handlers.CreateMedicalRecord)
			medicalRecords.GET("", handlers.GetMedicalRecords)
			medicalRecords.GET("/:id", handlers.GetMedicalRecord)
			medicalRecords.PUT("/:id", handlers.UpdateMedicalRecord)
			medicalRecords.DELETE("/:id", handlers.DeleteMedicalRecord)
			medicalRecords.GET("/:id/pdf", handlers.GenerateMedicalRecordPDF)
		}

		// Prescriptions CRUD (Receituário)
		prescriptions := tenanted.Group("/prescriptions")
		{
			prescriptions.POST("", handlers.CreatePrescription)
			prescriptions.GET("", handlers.GetPrescriptions)
			prescriptions.GET("/:id", handlers.GetPrescription)
			prescriptions.PUT("/:id", handlers.UpdatePrescription)
			prescriptions.DELETE("/:id", handlers.DeletePrescription)
			prescriptions.POST("/:id/issue", handlers.IssuePrescription)
			prescriptions.POST("/:id/print", handlers.PrintPrescription)
			prescriptions.GET("/:id/pdf", handlers.GeneratePrescriptionPDF)
		}

		// Budgets CRUD
		budgets := tenanted.Group("/budgets")
		{
			budgets.POST("", handlers.CreateBudget)
			budgets.GET("", handlers.GetBudgets)
			budgets.GET("/:id", handlers.GetBudget)
			budgets.PUT("/:id", handlers.UpdateBudget)
			budgets.DELETE("/:id", handlers.DeleteBudget)
			budgets.GET("/:id/pdf", handlers.GenerateBudgetPDF)
			budgets.GET("/:id/payment/:payment_id/receipt", handlers.GeneratePaymentReceipt)
		}

		// Payments CRUD
		payments := tenanted.Group("/payments")
		{
			payments.POST("", handlers.CreatePayment)
			payments.GET("", handlers.GetPayments)
			payments.GET("/:id", handlers.GetPayment)
			payments.PUT("/:id", handlers.UpdatePayment)
			payments.DELETE("/:id", handlers.DeletePayment)
			payments.GET("/pdf/export", handlers.GeneratePaymentsPDF)
		}

		// Products CRUD
		products := tenanted.Group("/products")
		{
			products.POST("", handlers.CreateProduct)
			products.GET("", handlers.GetProducts)
			products.GET("/:id", handlers.GetProduct)
			products.PUT("/:id", handlers.UpdateProduct)
			products.DELETE("/:id", handlers.DeleteProduct)
			products.GET("/low-stock", handlers.GetLowStockProducts)
		}

		// Suppliers CRUD
		suppliers := tenanted.Group("/suppliers")
		{
			suppliers.POST("", handlers.CreateSupplier)
			suppliers.GET("", handlers.GetSuppliers)
			suppliers.GET("/:id", handlers.GetSupplier)
			suppliers.PUT("/:id", handlers.UpdateSupplier)
			suppliers.DELETE("/:id", handlers.DeleteSupplier)
		}

		// Stock Movements
		stockMovements := tenanted.Group("/stock-movements")
		{
			stockMovements.POST("", handlers.CreateStockMovement)
			stockMovements.GET("", handlers.GetStockMovements)
		}

		// Dashboard and Reports
		reports := tenanted.Group("/reports")
		{
			reports.GET("/dashboard", handlers.GetDashboard)
			reports.GET("/revenue", handlers.GetRevenueReport)
			reports.GET("/procedures", handlers.GetProceduresReport)
			reports.GET("/attendance", handlers.GetAttendanceReport)
			reports.GET("/revenue/pdf", handlers.GenerateRevenuePDF)
			reports.GET("/attendance/pdf", handlers.GenerateAttendancePDF)
			reports.GET("/procedures/pdf", handlers.GenerateProceduresPDF)
			reports.GET("/revenue/excel", handlers.GenerateRevenueExcel)
			reports.GET("/attendance/excel", handlers.GenerateAttendanceExcel)
			reports.GET("/procedures/excel", handlers.GenerateProceduresExcel)
		}

		// Campaigns CRUD
		campaigns := tenanted.Group("/campaigns")
		{
			campaigns.POST("", handlers.CreateCampaign)
			campaigns.GET("", handlers.GetCampaigns)
			campaigns.GET("/:id", handlers.GetCampaign)
			campaigns.PUT("/:id", handlers.UpdateCampaign)
			campaigns.DELETE("/:id", handlers.DeleteCampaign)
			campaigns.POST("/:id/send", handlers.SendCampaign)
		}

		// Exams CRUD
		exams := tenanted.Group("/exams")
		{
			exams.POST("", handlers.CreateExam)
			exams.GET("", handlers.GetExams)
			exams.GET("/:id", handlers.GetExam)
			exams.PUT("/:id", handlers.UpdateExam)
			exams.DELETE("/:id", handlers.DeleteExam)
			exams.GET("/:id/download", handlers.GetExamDownloadURL)
		}

		// Tenant Settings
		tenanted.GET("/settings", handlers.GetTenantSettings)
		tenanted.PUT("/settings", handlers.UpdateTenantSettings)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
