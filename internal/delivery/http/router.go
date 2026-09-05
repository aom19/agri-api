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

type AppDeps struct {
	Log                        *logger.Logger
	MachineService             *usecase.MachineService
	ResourceService            *usecase.ResourceService
	StockService               *usecase.StockService
	ImplementService           *usecase.ImplementService
	OperatorService            *usecase.OperatorService
	FieldService               *usecase.FieldService
	AssignmentService          *usecase.AssigmentService
	OperationService           *usecase.OperationService
	FieldOperationService      *usecase.FieldOperationService
	DashboardService           *usecase.DashboardService
	WeatherService             *usecase.WeatherService
	AuditService               *usecase.AuditService
	NotificationService        *usecase.NotificationService
	AuditRepo                  repository.AuditRepository
	ImplementCompatibilityRepo repository.ImplementCompatibilityRepository
	AuthService                *usecase.AuthService
	ProfileService             *usecase.ProfileService
	RBACService                *usecase.RBACService
	UserService                *usecase.UserService
	PermissionRepo             repository.PermissionRepository
	UploadDir                  string
	JWTService                 *auth.JWTService
	Blacklist                  *auth.Blacklist
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

	weatherHandler := handlers.NewWeatherHandler(deps.WeatherService)
	r.GET("/api/weather/current", weatherHandler.GetCurrent)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(deps.JWTService, deps.Blacklist))

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
	api.GET("/dashboard/cards", perm("dashboard:read"), dashboardHandler.GetCards)
	api.GET("/dashboard/quick-stats", perm("dashboard:read"), dashboardHandler.GetQuickStats)
	api.GET("/dashboard/activity", perm("dashboard:read"), dashboardHandler.GetActivity)

	// ─── Machines ────────────────────────────────────────────────────────────
	machineHandler := handlers.NewMachineHandler(
		deps.MachineService,
		handlers.WithMachineAudit(deps.AuditService),
		handlers.WithMachineNotif(deps.NotificationService),
	)
	api.GET("/machines", perm("machines:read"), machineHandler.GetAll)
	api.GET("/machines/:id", perm("machines:read"), machineHandler.GetByID)
	api.POST("/machines", perm("machines:write"), machineHandler.Create)
	api.PATCH("/machines/:id", perm("machines:write"), machineHandler.Update)
	api.DELETE("/machines/:id", perm("machines:delete"), machineHandler.Delete)

	// ─── Resources ───────────────────────────────────────────────────────────
	resourceHandler := handlers.NewResourceHandler(
		deps.ResourceService,
		handlers.WithResourceAudit(deps.AuditService),
		handlers.WithResourceNotif(deps.NotificationService),
	)
	api.GET("/resource-types", perm("resources:read"), resourceHandler.GetAllResourceTypes)
	api.GET("/resource-types/:id", perm("resources:read"), resourceHandler.GetResourceTypeByID)
	api.POST("/resource-types", perm("resources:write"), resourceHandler.CreateResourceType)
	api.PATCH("/resource-types/:id", perm("resources:write"), resourceHandler.UpdateResourceType)
	api.DELETE("/resource-types/:id", perm("resources:delete"), resourceHandler.DeleteResourceType)

	api.GET("/resources", perm("resources:read"), resourceHandler.GetAllResources)
	api.GET("/resources/:id", perm("resources:read"), resourceHandler.GetResourceByID)
	api.POST("/resources", perm("resources:write"), resourceHandler.CreateResource)
	api.PATCH("/resources/:id", perm("resources:write"), resourceHandler.UpdateResource)
	api.DELETE("/resources/:id", perm("resources:delete"), resourceHandler.DeleteResource)

	// ─── Stocks ──────────────────────────────────────────────────────────────
	stockHandler := handlers.NewStockHandler(
		deps.StockService,
		handlers.WithStockAudit(deps.AuditService),
		handlers.WithStockNotif(deps.NotificationService),
	)
	api.GET("/stocks", perm("stock.view"), stockHandler.GetAll)
	api.GET("/stocks/:id", perm("stock.view"), stockHandler.GetByID)
	api.POST("/stocks", perm("stock.create"), stockHandler.Create)
	api.PATCH("/stocks/:id", perm("stock.update"), stockHandler.Update)
	api.DELETE("/stocks/:id", perm("stock.delete"), stockHandler.Delete)

	// ─── Implements ──────────────────────────────────────────────────────────
	implementHandler := handlers.NewImplementHandler(
		deps.ImplementService,
		handlers.WithImplementAudit(deps.AuditService),
		handlers.WithImplementNotif(deps.NotificationService),
	)
	api.GET("/implements", perm("implements:read"), implementHandler.GetAll)
	api.GET("/implements/:id", perm("implements:read"), implementHandler.GetByID)
	api.POST("/implements", perm("implements:write"), implementHandler.Create)
	api.PATCH("/implements/:id", perm("implements:write"), implementHandler.Update)
	api.DELETE("/implements/:id", perm("implements:delete"), implementHandler.Delete)
	//activate
	api.PATCH("/implements/:id/activate", perm("implements:write"), implementHandler.Activate)
	api.PATCH("/implements/:id/deactivate", perm("implements:write"), implementHandler.Deactivate)

	// ─── Operators ───────────────────────────────────────────────────────────
	operatorHandler := handlers.NewOperatorHandler(
		deps.OperatorService,
		handlers.WithOperatorAudit(deps.AuditService),
		handlers.WithOperatorNotif(deps.NotificationService),
	)
	api.GET("/operators", perm("operators:read"), operatorHandler.GetAll)
	api.GET("/operators/:id", perm("operators:read"), operatorHandler.GetByID)
	api.POST("/operators", perm("operators:write"), operatorHandler.Create)
	api.PATCH("/operators/:id", perm("operators:write"), operatorHandler.Update)
	api.DELETE("/operators/:id", perm("operators:delete"), operatorHandler.Delete)
	api.PATCH("/operators/:id/disable", perm("operators:disable"), operatorHandler.Disable)
	api.PATCH("/operators/:id/enable", perm("operators:disable"), operatorHandler.Enable)

	// ─── Fields ───────────────────────────────────────────────────────────────
	fieldHandler := handlers.NewFieldHandler(
		deps.FieldService,
		handlers.WithFieldAudit(deps.AuditService),
		handlers.WithFieldNotif(deps.NotificationService),
	)
	api.GET("/fields", perm("fields:read"), fieldHandler.GetAll)
	api.GET("/fields/:id", perm("fields:read"), fieldHandler.GetByID)
	api.POST("/fields", perm("fields:write"), fieldHandler.Create)
	api.PATCH("/fields/:id", perm("fields:write"), fieldHandler.Update)
	api.DELETE("/fields/:id", perm("fields:delete"), fieldHandler.Delete)

	// ─── Assignments ──────────────────────────────────────────────────────────
	assignmentHandler := handlers.NewAssigmentHandler(
		deps.AssignmentService,
		handlers.WithAssigmentAudit(deps.AuditService),
		handlers.WithAssigmentNotif(deps.NotificationService),
	)
	api.GET("/assignments", perm("assignments:read"), assignmentHandler.GetAll)
	api.GET("/assignments/:id", perm("assignments:read"), assignmentHandler.GetByID)
	api.POST("/assignments", perm("assignments:write"), assignmentHandler.Create)
	api.PATCH("/assignments/:id", perm("assignments:write"), assignmentHandler.Update)
	api.DELETE("/assignments/:id", perm("assignments:delete"), assignmentHandler.Delete)
	api.PATCH("/assignments/:id/close", perm("assignments:write"), assignmentHandler.Close)

	// ─── RBAC — Roles & Permissions ──────────────────────────────────────────
	rbacHandler := handlers.NewRBACHandler(deps.RBACService)
	api.GET("/roles", perm("roles:read"), rbacHandler.GetAllRoles)
	api.GET("/roles/:id", perm("roles:read"), rbacHandler.GetRole)
	api.POST("/roles", perm("roles:write"), rbacHandler.CreateRole)
	api.PATCH("/roles/:id", perm("roles:write"), rbacHandler.UpdateRole)
	api.DELETE("/roles/:id", perm("roles:delete"), rbacHandler.DeleteRole)
	api.GET("/roles/:id/permissions", perm("roles:read"), rbacHandler.GetRolePermissions)
	api.PUT("/roles/:id/permissions", perm("roles:write"), rbacHandler.SetRolePermissions)
	api.GET("/permissions", perm("roles:read"), rbacHandler.GetAllPermissions)
	api.GET("/permissions/:id", perm("roles:read"), rbacHandler.GetPermission)
	api.GET("/auth/me/permissions", rbacHandler.GetMyPermissions)

	// ─── Operations ──────────────────────────────────────────────────────────
	operationHandler := handlers.NewOperationHandler(
		deps.OperationService,
		handlers.WithOperationAudit(deps.AuditService),
		handlers.WithOperationNotif(deps.NotificationService),
	)
	api.GET("/operation-types", perm("operations:read"), operationHandler.GetAllTypes)
	api.GET("/operation-types/:id", perm("operations:read"), operationHandler.GetTypeByID)
	api.POST("/operation-types", perm("operations:write"), operationHandler.CreateType)
	api.PATCH("/operation-types/:id", perm("operations:write"), operationHandler.UpdateType)
	api.DELETE("/operation-types/:id", perm("operations:delete"), operationHandler.DeleteType)

	api.GET("/operation-templates", perm("operations:read"), operationHandler.GetAllTemplates)
	api.GET("/operation-templates/:id", perm("operations:read"), operationHandler.GetTemplateByID)
	api.GET("/operation-types/:id/templates", perm("operations:read"), operationHandler.GetTemplatesByType)
	api.POST("/operation-templates", perm("operations:write"), operationHandler.CreateTemplate)
	api.PATCH("/operation-templates/:id", perm("operations:write"), operationHandler.UpdateTemplate)
	api.DELETE("/operation-templates/:id", perm("operations:delete"), operationHandler.DeleteTemplate)

	// ─── Implement Compatibilities ───────────────────────────────────────────
	compatibilityHandler := handlers.NewImplementCompatibilityHandler(deps.ImplementCompatibilityRepo)
	api.GET("/implement-compatibilities", perm("operations:read"), compatibilityHandler.GetAll)

	// ─── Field Operations ────────────────────────────────────────────────────
	fieldOperationHandler := handlers.NewFieldOperationHandler(
		deps.FieldOperationService,
		handlers.WithFieldOpAudit(deps.AuditService),
		handlers.WithFieldOpNotif(deps.NotificationService),
	)
	api.GET("/field-operations", perm("field_operations:read"), fieldOperationHandler.GetAll)
	api.GET("/field-operations/:id", perm("field_operations:read"), fieldOperationHandler.GetByID)
	api.POST("/field-operations", perm("field_operations:write"), fieldOperationHandler.Create)
	api.PATCH("/field-operations/:id", perm("field_operations:write"), fieldOperationHandler.Update)
	api.PATCH("/field-operations/:id/checklist", perm("field_operations:checklist"), fieldOperationHandler.UpdateChecklist)
	api.PATCH("/field-operations/:id/start", perm("field_operations:start"), fieldOperationHandler.Start)
	api.DELETE("/field-operations/:id", perm("field_operations:delete"), fieldOperationHandler.Delete)

	// ─── Users — assign role ──────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(
		deps.UserService,
		handlers.WithUserAudit(deps.AuditService),
		handlers.WithUserNotif(deps.NotificationService),
	)
	api.GET("/users", perm("users:read"), userHandler.GetAll)
	api.GET("/users/:id", perm("users:read"), userHandler.GetByID)
	api.POST("/users", perm("users:write"), userHandler.Create)
	api.PATCH("/users/:id", perm("users:write"), userHandler.Update)
	api.DELETE("/users/:id", perm("users:disable"), userHandler.Disable)
	api.PATCH("/users/:id/enable", perm("users:enable"), userHandler.Enable)
	api.PATCH("/users/:id/role", perm("users:write"), rbacHandler.AssignRoleToUser)

	// ─── Notifications ──────────────────────────────────────────────────────
	notificationHandler := handlers.NewNotificationHandler(deps.NotificationService)
	api.GET("/notifications", perm("notifications:read"), notificationHandler.GetAll)
	api.GET("/notifications/count", perm("notifications:read"), notificationHandler.CountUnread)
	api.PATCH("/notifications/:id/read", perm("notifications:read"), notificationHandler.MarkAsRead)
	api.PATCH("/notifications/read-all", perm("notifications:read"), notificationHandler.MarkAllAsRead)

	// ─── WebSocket (notification stream) ─────────────────────────────────────
	wsHandler := handlers.NewWebSocketHandler(deps.NotificationService, deps.JWTService, deps.Blacklist)
	r.GET("/ws/notifications", wsHandler.HandleNotifications)

	// ─── Audit Log ───────────────────────────────────────────────────────
	auditHandler := handlers.NewAuditHandler(deps.AuditRepo)
	api.GET("/audit-log", perm("audit:read"), auditHandler.GetAll)

}
