package http

import (
	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/logger"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AppDeps grupează toate dependențele aplicației.
// Adaugă un serviciu nou direct aici, fără să modifici SetupRoutes.
type AppDeps struct {
	Log               *logger.Logger
	MachineService    *usecase.MachineService
	OperatorService   *usecase.OperatorService
	AssignmentService *usecase.AssigmentService
}

// SetupRoutes înregistrează toate rutele HTTP ale aplicației sub prefixul /api
func SetupRoutes(r *gin.Engine, deps AppDeps) {
	api := r.Group("/api")

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
