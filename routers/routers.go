package routes

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/database"
	_ "Backend_Dorm_PTIT/docs" // Import docs to load swagger documentation
	"Backend_Dorm_PTIT/handlers"
	"Backend_Dorm_PTIT/middleware"
	"Backend_Dorm_PTIT/service"

	// "Backend_Dorm_PTIT/middleware"
	"Backend_Dorm_PTIT/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all application routes with dependency injection
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	
    router.Use(middleware.CORS(&cfg.CORS))
	// Health check endpoint
	router.GET("/health", handlers.Health)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	userRepo := repository.NewUserRepository(database.GetDB(), cfg.Database.Schema)
	// Initialize auth handler with config
	authHandler := handlers.NewAuthHandler(cfg, userRepo)

	// User handler
	userHandler := handlers.NewUserHandler(cfg, userRepo)

	WSService := service.NewWSService()

	// Auth routes
	router.POST("/login", authHandler.LoginHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/logout", authHandler.LogoutHandler)
	router.POST("/logout-all", authHandler.LogoutAllSessionsHandler)

	testHandler := handlers.NewTestHandler(cfg, userRepo)
	test := router.Group("/api/test")
	{
		test.Use(middleware.Authentication(cfg.JWT.Secret))
		test.GET("/getprofile", testHandler.GetProfileHandler)
		test.GET("/sendmail", testHandler.SendEmailHandler)

	}

	ws := router.Group("/ws/v1")
	{
		ws.Use(middleware.AuthenticateWS(cfg.JWT.Secret))
		wsHandler := handlers.NewWSHandler(cfg, WSService)
		ws.GET("/admin-connect", wsHandler.HandleWSAdmin)
	}

	v1 := router.Group("/api/v1")
	{
		dormAppRepo := repository.NewDormApplicationRepository(database.GetDB())
		dormAppHandler := handlers.NewDormApplicationHandler(dormAppRepo, cfg)
		mailHandler := handlers.NewMailHandler(cfg, userRepo)
		dormAreaRepo := repository.NewDormAreaRepository(database.GetDB())
		dormAreaHandler := handlers.NewDormAreaHandler(dormAreaRepo)
		registrationPeriodRepo := repository.NewRegistrationPeriodRepository(database.GetDB())
		registrationPeriodHandler := handlers.NewRegistrationPeriodHandler(registrationPeriodRepo)
		contractRepo := repository.NewContractRepository(database.GetDB())
		contractHandler := handlers.NewContractHandler(contractRepo, cfg)
		managerRepo := repository.NewManagerRepository(database.GetDB(), cfg.Database.Schema)
		managerHandler := handlers.NewManagerHandler(cfg, managerRepo, userRepo)

		dutyRepo := repository.NewDutyScheduleRepository(database.GetDB())
		dutyHandler := handlers.NewDutyScheduleHandler(dutyRepo)
		electricBillRepo := repository.NewElectricBillRepository(database.GetDB())
		electricBillHandler := handlers.NewElectricBillHandler(electricBillRepo, cfg)
		electricBillComplaintRepo := repository.NewElectricBillComplaintRepository(database.GetDB())
		electricBillComplaintHandler := handlers.NewElectricBillComplaintHandler(electricBillComplaintRepo, cfg)
		facilityComplaintRepo := repository.NewFacilityComplaintRepository(database.GetDB())
		facilityComplaintHandler := handlers.NewFacilityComplaintHandler(facilityComplaintRepo, cfg)

		backupRepo := repository.NewBackUpRepository(database.GetDB())
		backupHandler := handlers.NewBackupHandler(cfg, backupRepo)
		// Đăng ký ký túc xá
		v1.POST("/dorm-applications", dormAppHandler.CreateDormApplication)
		v1.POST("/send-otp", mailHandler.SendOTPEmailHandler)
		v1.POST("/verify-otp", mailHandler.VerifyOTPHandler)

		v2 := v1.Group("/protected")
		{
			v2.Use(middleware.Authentication(cfg.JWT.Secret))
			// api to backupdata for admin_system
			v2.GET("/backup-data", backupHandler.BackUpData)
			// Đổi avatar và mật khẩu cho user hiện tại
			v2.PATCH("/me/avatar", userHandler.UpdateAvatar)
			v2.PATCH("/me/password", userHandler.UpdatePassword)
			// API: List all users with roles (admin_system only)
			v2.GET("/users", userHandler.ListAllUsers)
			// update profile ( manager and admin_system only)
			v2.PUT("/me/profile", userHandler.UpdateOwnManagerProfile)
			// update status user... (admin_system only)
			v2.PATCH("/users/:id/status", userHandler.UpdateUserStatus)
			v2.GET("/contracts/me", contractHandler.GetMyContract)
			v2.PATCH("/contracts/:id/confirm", contractHandler.ConfirmContract)
			v2.GET("/contracts", contractHandler.GetAllContracts)
			v2.PATCH("/contracts/:id/verify", contractHandler.VerifyContract)
			v2.GET("/dorm-applications", dormAppHandler.GetAllDormApplications)
			v2.PATCH("/dorm-applications/:id/status", dormAppHandler.UpdateDormApplicationStatus)
			v2.POST("/dorm-area", dormAreaHandler.CreateDormArea)
			v2.PATCH("/dorm-area/:id", dormAreaHandler.UpdateDormArea)
			v2.DELETE("/dorm-area/:id", dormAreaHandler.DeleteDormArea)
			v2.GET("/dorm-areas", dormAreaHandler.GetAllDormAreas)
			v2.POST("/registration-periods", registrationPeriodHandler.CreateRegistrationPeriod)
			v2.GET("/registration-periods", registrationPeriodHandler.GetAllRegistrationPeriods)
			v2.PATCH("/registration-periods/:id", registrationPeriodHandler.UpdateRegistrationPeriod)
			v2.DELETE("/registration-periods/:id", registrationPeriodHandler.DeleteRegistrationPeriod)

			v2.POST("/managers", managerHandler.CreateManager)
			v2.PUT("/managers/:id", managerHandler.UpdateManager)
			v2.DELETE("/managers/:id", managerHandler.DeleteManager)
			v2.GET("/managers", managerHandler.ListManagers)
			v2.GET("/managers/:id", managerHandler.GetManagerDetail)

			v2.GET("/duty-schedules", dutyHandler.ListDutySchedules)
			v2.POST("/duty-schedules", dutyHandler.CreateDutySchedule)
			v2.PUT("/duty-schedules/:id", dutyHandler.UpdateDutySchedule)
			v2.DELETE("/duty-schedules/:id", dutyHandler.DeleteDutySchedule)

			// Facility Complaint APIs (protected)
			v2.GET("/facility-complaints", facilityComplaintHandler.List)
			v2.GET("/facility-complaints/:id", facilityComplaintHandler.GetByID)
			v2.POST("/facility-complaints", facilityComplaintHandler.Create)
			v2.PATCH("/facility-complaints/:id", facilityComplaintHandler.Update)
			v2.DELETE("/facility-complaints/:id", facilityComplaintHandler.Delete)

			// Electric Bill APIs (protected)
			v2.GET("/electric-bills", electricBillHandler.List)
			v2.GET("/electric-bills/my-room", electricBillHandler.ListByMyRoom)
			v2.GET("/electric-bills/:id", electricBillHandler.GetByID)
			v2.POST("/electric-bills", electricBillHandler.Create)
			v2.PATCH("/electric-bills/:id", electricBillHandler.Update)
			v2.PATCH("/electric-bills/:id/confirm", electricBillHandler.ConfirmOnlyByStudent)
			v2.PATCH("/electric-bills/:id/payment-proof", electricBillHandler.ConfirmByStudent)
			v2.DELETE("/electric-bills/:id", electricBillHandler.Delete)

			// Electric Bill Complaint APIs (protected)
			v2.GET("/electric-bill-complaints", electricBillComplaintHandler.List)
			v2.GET("/electric-bill-complaints/:id", electricBillComplaintHandler.GetByID)
			v2.POST("/electric-bill-complaints", electricBillComplaintHandler.Create)
			v2.PATCH("/electric-bill-complaints/:id", electricBillComplaintHandler.Update)
			v2.DELETE("/electric-bill-complaints/:id", electricBillComplaintHandler.Delete)

			v2.PATCH("/electric-bills/:id/confirm-only", electricBillHandler.ConfirmOnlyByStudent)
		}
	}
}
