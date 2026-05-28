package http

import (
	"agri-api/internal/auth"
	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/delivery/http/middleware"
	"agri-api/internal/logger"
	"agri-api/internal/usecase"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// AppDeps grupează toate dependențele aplicației.
// Adaugă un serviciu nou direct aici, fără să modifici SetupRoutes.
type AppDeps struct {
	Log               *logger.Logger
	MachineService    *usecase.MachineService
	OperatorService   *usecase.OperatorService
	AssignmentService *usecase.AssigmentService
	AuthService       *usecase.AuthService
	JWTService        *auth.JWTService
	Blacklist         *auth.Blacklist
}

// SetupRoutes inregistreaza toate rutele HTTP ale aplicatiei sub prefixul /api
func SetupRoutes(r *gin.Engine, deps AppDeps) {

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authHandler := handlers.NewAuthHandler(deps.AuthService)
	authGroup := r.Group("/api/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password/:token", authHandler.ResetPassword)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(deps.JWTService, deps.Blacklist))

	api.POST("/auth/logout", authHandler.Logout)

	// Rută de verificare a stării serviciului
	api.GET("/health", func(c *gin.Context) {
		deps.Log.Info("Health check called")
		handlers.HealthCheck(c)
	})

	// Rutele pentru resursa "mașini agricole"
	machineHandler := handlers.NewMachineHandler(deps.MachineService)
	api.POST("/machines", machineHandler.Create)       // creare mașină
	api.GET("/machines", machineHandler.GetAll)        // listare toate mașinile
	api.GET("/machines/:id", machineHandler.GetByID)   // obținere mașină după ID
	api.PATCH("/machines/:id", machineHandler.Update)  // actualizare parțială mașină
	api.DELETE("/machines/:id", machineHandler.Delete) // ștergere mașină

	// Rutele pentru resursa "operatori"
	operatorHandler := handlers.NewOperatorHandler(deps.OperatorService)
	api.POST("/operators", operatorHandler.Create)       // creare operator
	api.GET("/operators", operatorHandler.GetAll)        // listare toți operatorii
	api.GET("/operators/:id", operatorHandler.GetByID)   // obținere operator după ID
	api.PATCH("/operators/:id", operatorHandler.Update)  // actualizare parțială operator
	api.DELETE("/operators/:id", operatorHandler.Delete) // ștergere operator

	assignmentHandler := handlers.NewAssigmentHandler(deps.AssignmentService)

	api.POST("/assignments", assignmentHandler.Create)
	api.GET("/assignments", assignmentHandler.GetAll)
	api.GET("/assignments/:id", assignmentHandler.GetByID)
	api.PATCH("/assignments/:id", assignmentHandler.Update)
	api.DELETE("/assignments/:id", assignmentHandler.Delete)
	api.PATCH("/assignments/:id/close", assignmentHandler.Close)
}
