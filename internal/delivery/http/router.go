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
	Log               *logger.Logger
	MachineService    *usecase.MachineService
	ResourceService   *usecase.ResourceService
	StockService      *usecase.StockService
	ImplementService  *usecase.ImplementService
	OperatorService   *usecase.OperatorService
	FieldService      *usecase.FieldService
	AssignmentService *usecase.AssigmentService
	AuthService       *usecase.AuthService
	ProfileService    *usecase.ProfileService
	RBACService       *usecase.RBACService
	UserService       *usecase.UserService
	PermissionRepo    repository.PermissionRepository
	UploadDir         string
	JWTService        *auth.JWTService
	Blacklist         *auth.Blacklist
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

	// ─── Machines ────────────────────────────────────────────────────────────
	machineHandler := handlers.NewMachineHandler(deps.MachineService)
	api.GET("/machines", perm("machines:read"), machineHandler.GetAll)
	api.GET("/machines/:id", perm("machines:read"), machineHandler.GetByID)
	api.POST("/machines", perm("machines:write"), machineHandler.Create)
	api.PATCH("/machines/:id", perm("machines:write"), machineHandler.Update)
	api.DELETE("/machines/:id", perm("machines:delete"), machineHandler.Delete)

	// ─── Resources ───────────────────────────────────────────────────────────
	resourceHandler := handlers.NewResourceHandler(deps.ResourceService)
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
	stockHandler := handlers.NewStockHandler(deps.StockService)
	api.GET("/stocks", perm("stock.view"), stockHandler.GetAll)
	api.GET("/stocks/:id", perm("stock.view"), stockHandler.GetByID)
	api.POST("/stocks", perm("stock.create"), stockHandler.Create)
	api.PATCH("/stocks/:id", perm("stock.update"), stockHandler.Update)
	api.DELETE("/stocks/:id", perm("stock.delete"), stockHandler.Delete)

	// ─── Implements ──────────────────────────────────────────────────────────
	implementHandler := handlers.NewImplementHandler(deps.ImplementService)
	api.GET("/implements", perm("implements:read"), implementHandler.GetAll)
	api.GET("/implements/:id", perm("implements:read"), implementHandler.GetByID)
	api.POST("/implements", perm("implements:write"), implementHandler.Create)
	api.PATCH("/implements/:id", perm("implements:write"), implementHandler.Update)
	api.DELETE("/implements/:id", perm("implements:delete"), implementHandler.Delete)
	//activate
	api.PATCH("/implements/:id/activate", perm("implements:write"), implementHandler.Activate)
	api.PATCH("/implements/:id/deactivate", perm("implements:write"), implementHandler.Deactivate)

	// ─── Operators ───────────────────────────────────────────────────────────
	operatorHandler := handlers.NewOperatorHandler(deps.OperatorService)
	api.GET("/operators", perm("operators:read"), operatorHandler.GetAll)
	api.GET("/operators/:id", perm("operators:read"), operatorHandler.GetByID)
	api.POST("/operators", perm("operators:write"), operatorHandler.Create)
	api.PATCH("/operators/:id", perm("operators:write"), operatorHandler.Update)
	api.DELETE("/operators/:id", perm("operators:delete"), operatorHandler.Delete)
	api.PATCH("/operators/:id/disable", perm("operators:disable"), operatorHandler.Disable)
	api.PATCH("/operators/:id/enable", perm("operators:disable"), operatorHandler.Enable)

	// ─── Fields ───────────────────────────────────────────────────────────────
	fieldHandler := handlers.NewFieldHandler(deps.FieldService)
	api.GET("/fields", perm("fields:read"), fieldHandler.GetAll)
	api.GET("/fields/:id", perm("fields:read"), fieldHandler.GetByID)
	api.POST("/fields", perm("fields:write"), fieldHandler.Create)
	api.PATCH("/fields/:id", perm("fields:write"), fieldHandler.Update)
	api.DELETE("/fields/:id", perm("fields:delete"), fieldHandler.Delete)

	// ─── Assignments ──────────────────────────────────────────────────────────
	assignmentHandler := handlers.NewAssigmentHandler(deps.AssignmentService)
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

	// ─── Users — assign role ──────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(deps.UserService)
	api.GET("/users", perm("users:read"), userHandler.GetAll)
	api.GET("/users/:id", perm("users:read"), userHandler.GetByID)
	api.POST("/users", perm("users:write"), userHandler.Create)
	api.PATCH("/users/:id", perm("users:write"), userHandler.Update)
	api.DELETE("/users/:id", perm("users:disable"), userHandler.Disable)
	api.PATCH("/users/:id/enable", perm("users:enable"), userHandler.Enable)
	api.PATCH("/users/:id/role", perm("users:write"), rbacHandler.AssignRoleToUser)

}
