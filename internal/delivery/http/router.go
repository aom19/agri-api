package http

import (
	"agri-api/internal/auth"
	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/delivery/http/middleware"
	"agri-api/internal/logger"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// Permisiunile cerute de rute (tabela permissions); numele sunt cele din migrația RBAC.
const (
	permAuditRead              = "audit:read"
	permCropsRead              = "crops:read"
	permCropsWrite             = "crops:write"
	permDashboardRead          = "dashboard:read"
	permFieldOperationsExecute = "field_operations:execute"
	permFieldOperationsRead    = "field_operations:read"
	permFieldOperationsWrite   = "field_operations:write"
	permFieldsRead             = "fields:read"
	permFieldsWrite            = "fields:write"
	permImplementsRead         = "implements:read"
	permImplementsWrite        = "implements:write"
	permMachinesRead           = "machines:read"
	permMachinesWrite          = "machines:write"
	permNotificationsRead      = "notifications:read"
	permOperationsRead         = "operations:read"
	permOperationsWrite        = "operations:write"
	permOperatorsRead          = "operators:read"
	permOperatorsWrite         = "operators:write"
	permReportsRead            = "reports:read"
	permResourcesRead          = "resources:read"
	permResourcesWrite         = "resources:write"
	permRolesRead              = "roles:read"
	permUsersRead              = "users:read"
	permUsersWrite             = "users:write"
)

type AppDeps struct {
	Log                             *logger.Logger
	MachineService                  *usecase.MachineService
	ResourceService                 *usecase.ResourceService
	ImplementService                *usecase.ImplementService
	OperatorService                 *usecase.OperatorService
	FieldService                    *usecase.FieldService
	OperationService                *usecase.OperationService
	FieldOperationService           *usecase.FieldOperationService
	DashboardService                *usecase.DashboardService
	ReportService                   *usecase.ReportService
	StockMovementService            *usecase.StockMovementService
	CropService                     *usecase.CropService
	ReportEmailService              *usecase.ReportEmailService
	FieldOperationCompletionService *usecase.FieldOperationCompletionService
	WeatherService                  *usecase.WeatherService
	AuditService                    *usecase.AuditService
	NotificationService             *usecase.NotificationService
	AuditRepo                       repository.AuditRepository
	AuthService                     *usecase.AuthService
	ProfileService                  *usecase.ProfileService
	RBACService                     *usecase.RBACService
	UserService                     *usecase.UserService
	PermissionRepo                  repository.PermissionRepository
	UploadDir                       string
	JWTService                      *auth.JWTService
	Blacklist                       *auth.Blacklist
}

func SetupRoutes(r *gin.Engine, deps AppDeps) {
	// helper pentru construirea middleware-ului de permisiuni
	perm := func(name string) gin.HandlerFunc {
		return middleware.RequirePermission(deps.PermissionRepo, name)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/uploads", "uploads")

	authHandler := handlers.NewAuthHandler(deps.AuthService)
	authGroup := r.Group("/api/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/confirm-email", authHandler.ConfirmEmail)
	authGroup.POST("/resend-confirmation", authHandler.ResendConfirmation)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password/:token", authHandler.ResetPassword)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(deps.JWTService, deps.Blacklist), middleware.RequirePasswordChanged())

	// autentificat, ca să nu poată fi folosit de oricine drept proxy către cheia OpenWeather
	weatherHandler := handlers.NewWeatherHandler(deps.WeatherService)
	api.GET("/weather/current", weatherHandler.GetCurrent)

	api.POST("/auth/logout", authHandler.Logout)
	api.POST("/auth/change-password", authHandler.ChangePassword)

	profileHandler := handlers.NewProfileHandler(deps.ProfileService)
	api.GET("/profile", profileHandler.GetProfile)
	api.PATCH("/profile", profileHandler.UpdateProfile)
	api.POST("/profile/photo", profileHandler.UploadPhoto)

	api.GET("/health", func(c *gin.Context) {
		deps.Log.Info("Health check called")
		handlers.HealthCheck(c)
	})

	// ─── Dashboard ────────────────────────────────────────────────────────────
	dashboardHandler := handlers.NewDashboardHandler(deps.DashboardService)
	api.GET("/dashboard/cards", perm(permDashboardRead), dashboardHandler.GetCards)
	api.GET("/dashboard/quick-stats", perm(permDashboardRead), dashboardHandler.GetQuickStats)
	api.GET("/dashboard/activity", perm(permDashboardRead), dashboardHandler.GetActivity)

	// ─── Reports ─────────────────────────────────────────────────────────────
	reportHandler := handlers.NewReportHandler(deps.ReportService, handlers.WithReportEmail(deps.ReportEmailService, deps.UserService))
	api.GET("/reports/summary", perm(permReportsRead), reportHandler.GetSummary)
	api.GET("/reports/operations", perm(permReportsRead), reportHandler.GetOperations)
	// raportul propriu al operatorului: doar lucrările contului curent, fără reports:read
	api.GET("/reports/my-operations", perm(permFieldOperationsRead), reportHandler.GetMyOperations)
	api.GET("/reports/fields", perm(permReportsRead), reportHandler.GetFields)
	api.GET("/reports/fleet", perm(permReportsRead), reportHandler.GetFleet)
	api.GET("/reports/operators", perm(permReportsRead), reportHandler.GetOperators)
	api.GET("/reports/stocks", perm(permReportsRead), reportHandler.GetStocks)
	api.GET("/reports/crops", perm(permReportsRead), reportHandler.GetCrops)
	api.GET("/reports/compare", perm(permReportsRead), reportHandler.Compare)
	api.GET("/reports/weather", perm(permReportsRead), reportHandler.GetWeather)
	api.POST("/reports/summary/email", perm(permReportsRead), reportHandler.SendSummaryEmail)

	// ─── Seasons, crops, field crops ─────────────────────────────────────────
	cropHandler := handlers.NewCropHandler(deps.CropService, deps.AuditService)
	api.GET("/seasons", perm(permCropsRead), cropHandler.GetSeasons)
	api.POST("/seasons", perm(permCropsWrite), cropHandler.CreateSeason)
	api.PATCH("/seasons/:id", perm(permCropsWrite), cropHandler.UpdateSeason)
	api.DELETE("/seasons/:id", perm(permCropsWrite), cropHandler.DeleteSeason)
	api.GET("/crops", perm(permCropsRead), cropHandler.GetCrops)
	api.POST("/crops", perm(permCropsWrite), cropHandler.CreateCrop)
	api.PATCH("/crops/:id", perm(permCropsWrite), cropHandler.UpdateCrop)
	api.DELETE("/crops/:id", perm(permCropsWrite), cropHandler.DeleteCrop)
	api.GET("/field-crops", perm(permCropsRead), cropHandler.ListFieldCrops)
	api.POST("/field-crops", perm(permCropsWrite), cropHandler.CreateFieldCrop)
	api.PATCH("/field-crops/:id", perm(permCropsWrite), cropHandler.UpdateFieldCrop)
	api.DELETE("/field-crops/:id", perm(permCropsWrite), cropHandler.DeleteFieldCrop)
	api.POST("/field-crops/:id/harvest", perm(permCropsWrite), cropHandler.RecordHarvest)
	api.POST("/field-crops/:id/close", perm(permCropsWrite), cropHandler.CloseFieldCrop)
	api.POST("/field-crops/:id/reopen", perm(permCropsWrite), cropHandler.ReopenFieldCrop)

	// ─── Machines ────────────────────────────────────────────────────────────
	machineHandler := handlers.NewMachineHandler(
		deps.MachineService,
		handlers.WithMachineAudit(deps.AuditService),
		handlers.WithMachineNotif(deps.NotificationService),
	)
	api.GET("/machines", perm(permMachinesRead), machineHandler.GetAll)
	api.GET("/machines/:id", perm(permMachinesRead), machineHandler.GetByID)
	api.POST("/machines", perm(permMachinesWrite), machineHandler.Create)
	api.PATCH("/machines/:id", perm(permMachinesWrite), machineHandler.Update)
	api.DELETE("/machines/:id", perm(permMachinesWrite), machineHandler.Delete)

	// ─── Resources ───────────────────────────────────────────────────────────
	resourceHandler := handlers.NewResourceHandler(
		deps.ResourceService,
		handlers.WithResourceAudit(deps.AuditService),
		handlers.WithResourceNotif(deps.NotificationService),
	)
	api.GET("/resource-types", perm(permResourcesRead), resourceHandler.GetAllResourceTypes)
	api.GET("/resource-types/:id", perm(permResourcesRead), resourceHandler.GetResourceTypeByID)
	api.POST("/resource-types", perm(permResourcesWrite), resourceHandler.CreateResourceType)
	api.PATCH("/resource-types/:id", perm(permResourcesWrite), resourceHandler.UpdateResourceType)
	api.DELETE("/resource-types/:id", perm(permResourcesWrite), resourceHandler.DeleteResourceType)

	api.GET("/resources", perm(permResourcesRead), resourceHandler.GetAllResources)
	api.GET("/resources/category-counts", perm(permResourcesRead), resourceHandler.GetResourceCategoryCounts)
	api.GET("/resources/:id", perm(permResourcesRead), resourceHandler.GetResourceByID)
	api.POST("/resources", perm(permResourcesWrite), resourceHandler.CreateResource)
	api.PATCH("/resources/:id", perm(permResourcesWrite), resourceHandler.UpdateResource)
	api.DELETE("/resources/:id", perm(permResourcesWrite), resourceHandler.DeleteResource)

	// Stocul face parte din resursă și se modifică doar prin mișcări.
	stockMovementHandler := handlers.NewStockMovementHandler(deps.StockMovementService, deps.AuditService, deps.NotificationService)
	api.GET("/stock-movements", perm(permResourcesRead), stockMovementHandler.List)
	api.POST("/stock-movements", perm(permResourcesWrite), stockMovementHandler.Create)

	// ─── Implements ──────────────────────────────────────────────────────────
	implementHandler := handlers.NewImplementHandler(
		deps.ImplementService,
		handlers.WithImplementAudit(deps.AuditService),
		handlers.WithImplementNotif(deps.NotificationService),
	)
	api.GET("/implements", perm(permImplementsRead), implementHandler.GetAll)
	api.GET("/implements/:id", perm(permImplementsRead), implementHandler.GetByID)
	api.POST("/implements", perm(permImplementsWrite), implementHandler.Create)
	api.PATCH("/implements/:id", perm(permImplementsWrite), implementHandler.Update)
	api.DELETE("/implements/:id", perm(permImplementsWrite), implementHandler.Delete)
	//activate
	api.PATCH("/implements/:id/activate", perm(permImplementsWrite), implementHandler.Activate)
	api.PATCH("/implements/:id/deactivate", perm(permImplementsWrite), implementHandler.Deactivate)

	// ─── Operators ───────────────────────────────────────────────────────────
	operatorHandler := handlers.NewOperatorHandler(
		deps.OperatorService,
		handlers.WithOperatorAudit(deps.AuditService),
		handlers.WithOperatorNotif(deps.NotificationService),
	)
	api.GET("/operators", perm(permOperatorsRead), operatorHandler.GetAll)
	api.GET("/operators/:id", perm(permOperatorsRead), operatorHandler.GetByID)
	api.POST("/operators", perm(permOperatorsWrite), operatorHandler.Create)
	api.PATCH("/operators/:id", perm(permOperatorsWrite), operatorHandler.Update)
	api.DELETE("/operators/:id", perm(permOperatorsWrite), operatorHandler.Delete)
	api.PATCH("/operators/:id/disable", perm(permOperatorsWrite), operatorHandler.Disable)
	api.PATCH("/operators/:id/enable", perm(permOperatorsWrite), operatorHandler.Enable)

	// ─── Fields ───────────────────────────────────────────────────────────────
	fieldHandler := handlers.NewFieldHandler(
		deps.FieldService,
		handlers.WithFieldAudit(deps.AuditService),
		handlers.WithFieldNotif(deps.NotificationService),
	)
	api.GET("/fields", perm(permFieldsRead), fieldHandler.GetAll)
	api.GET("/fields/:id", perm(permFieldsRead), fieldHandler.GetByID)
	api.POST("/fields", perm(permFieldsWrite), fieldHandler.Create)
	api.PATCH("/fields/:id", perm(permFieldsWrite), fieldHandler.Update)
	api.DELETE("/fields/:id", perm(permFieldsWrite), fieldHandler.Delete)

	// ─── RBAC — Roles & Permissions ──────────────────────────────────────────
	// Rolurile și permisiunile sunt fixe (migrația 000054_fixed_roles_and_permissions), deci doar
	// se citesc. Fiecare modul are `:read` și `:write`; write include ștergerea și dezactivarea.
	rbacHandler := handlers.NewRBACHandler(deps.RBACService)
	api.GET("/roles", perm(permRolesRead), rbacHandler.GetAllRoles)
	api.GET("/roles/:id", perm(permRolesRead), rbacHandler.GetRole)
	api.GET("/roles/:id/permissions", perm(permRolesRead), rbacHandler.GetRolePermissions)
	api.GET("/permissions", perm(permRolesRead), rbacHandler.GetAllPermissions)
	api.GET("/permissions/:id", perm(permRolesRead), rbacHandler.GetPermission)
	api.GET("/auth/me/permissions", rbacHandler.GetMyPermissions)

	// ─── Operations ──────────────────────────────────────────────────────────
	operationHandler := handlers.NewOperationHandler(
		deps.OperationService,
		handlers.WithOperationAudit(deps.AuditService),
		handlers.WithOperationNotif(deps.NotificationService),
	)
	api.GET("/operation-templates", perm(permOperationsRead), operationHandler.GetAllTemplates)
	api.GET("/operation-templates/:id", perm(permOperationsRead), operationHandler.GetTemplateByID)
	api.POST("/operation-templates", perm(permOperationsWrite), operationHandler.CreateTemplate)
	api.PATCH("/operation-templates/:id", perm(permOperationsWrite), operationHandler.UpdateTemplate)
	api.DELETE("/operation-templates/:id", perm(permOperationsWrite), operationHandler.DeleteTemplate)

	// ─── Field Operations ────────────────────────────────────────────────────
	fieldOperationHandler := handlers.NewFieldOperationHandler(
		deps.FieldOperationService,
		handlers.WithFieldOpAudit(deps.AuditService),
		handlers.WithFieldOpNotif(deps.NotificationService),
		handlers.WithFieldOpCompletion(deps.FieldOperationCompletionService),
	)
	api.GET("/field-operations", perm(permFieldOperationsRead), fieldOperationHandler.GetAll)
	api.GET("/field-operations/conflicts", perm(permFieldOperationsRead), fieldOperationHandler.Conflicts)
	api.GET("/field-operations/status-counts", perm(permFieldOperationsRead), fieldOperationHandler.StatusCounts)
	api.GET("/field-operations/:id", perm(permFieldOperationsRead), fieldOperationHandler.GetByID)
	api.POST("/field-operations", perm(permFieldOperationsWrite), fieldOperationHandler.Create)
	api.PATCH("/field-operations/:id", perm(permFieldOperationsWrite), fieldOperationHandler.Update)
	api.PATCH("/field-operations/:id/start", perm(permFieldOperationsExecute), fieldOperationHandler.Start)
	api.GET("/field-operations/:id/consumption-estimate", perm(permFieldOperationsExecute), fieldOperationHandler.ConsumptionEstimate)
	api.PATCH("/field-operations/:id/complete", perm(permFieldOperationsExecute), fieldOperationHandler.Complete)
	api.DELETE("/field-operations/:id", perm(permFieldOperationsWrite), fieldOperationHandler.Delete)

	// ─── Users — assign role ──────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(
		deps.UserService,
		handlers.WithUserAudit(deps.AuditService),
		handlers.WithUserNotif(deps.NotificationService),
	)
	api.GET("/users", perm(permUsersRead), userHandler.GetAll)
	api.GET("/users/:id", perm(permUsersRead), userHandler.GetByID)
	api.POST("/users", perm(permUsersWrite), userHandler.Create)
	api.PATCH("/users/:id", perm(permUsersWrite), userHandler.Update)
	api.DELETE("/users/:id", perm(permUsersWrite), userHandler.Disable)
	api.PATCH("/users/:id/enable", perm(permUsersWrite), userHandler.Enable)
	api.PATCH("/users/:id/role", perm(permUsersWrite), rbacHandler.AssignRoleToUser)

	// ─── Notifications ──────────────────────────────────────────────────────
	notificationHandler := handlers.NewNotificationHandler(deps.NotificationService)
	api.GET("/notifications", perm(permNotificationsRead), notificationHandler.GetAll)
	api.GET("/notifications/count", perm(permNotificationsRead), notificationHandler.CountUnread)
	api.PATCH("/notifications/:id/read", perm(permNotificationsRead), notificationHandler.MarkAsRead)
	api.PATCH("/notifications/read-all", perm(permNotificationsRead), notificationHandler.MarkAllAsRead)

	// ─── WebSocket (notification stream) ─────────────────────────────────────
	wsHandler := handlers.NewWebSocketHandler(deps.NotificationService, deps.JWTService, deps.Blacklist)
	r.GET("/ws/notifications", wsHandler.HandleNotifications)

	// ─── Audit Log ───────────────────────────────────────────────────────
	auditHandler := handlers.NewAuditHandler(deps.AuditRepo)
	api.GET("/audit-log", perm(permAuditRead), auditHandler.GetAll)

}
