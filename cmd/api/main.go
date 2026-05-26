package main

import (
	"fmt"

	"agri-api/internal/config"
	"agri-api/internal/db"
	httpdelivery "agri-api/internal/delivery/http"
	"agri-api/internal/logger"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/store"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {

	// Încarcă configurația din variabilele de mediu (.env)
	cfg := config.LoadConfig()

	// Inițializează logger-ul în funcție de mediul aplicației (dev/prod)
	log := logger.NewLogger(cfg.AppEnv)
	log.Info("Starting application...")

	// Conectare la baza de date PostgreSQL
	pg, err := db.NewPostgres(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	if err != nil {
		log.Fatal("DB connection failed: ", err)
	}

	log.Info("DB connected successfully")
	sqlDB := pg.DB

	// Inițializează repository-urile și serviciile (dependency injection manual)
	//1. Creează repository-ul pentru mașini și serviciul aferent
	machineRepo := postgres.NewMachineRepo(sqlDB)
	machineService := usecase.NewMachineService(machineRepo)

	// 2. Creează repository-ul pentru operatori și serviciul aferent
	operatorRepo := postgres.NewOperatorRepo(sqlDB)
	operatorService := usecase.NewOperatorService(operatorRepo)

	// 3. Inițializează store-ul cu toate repository-urile și serviciul de asignări
	appStore := store.NewInitialiedStore(sqlDB)
	assigmentService := usecase.NewAssigmentService(appStore)

	// Configurează serverul HTTP Gin fără middleware implicit
	server := gin.New()
	// Adaugă middleware pentru logarea request-urilor și recuperare din panic
	server.Use(gin.Logger(), gin.Recovery())

	// Dezactivează avertismentele legate de proxy-uri de încredere
	if err := server.SetTrustedProxies(nil); err != nil {
		log.Fatalf("SetTrustedProxies: %v", err)
	}

	// Înregistrează toate rutele API
	httpdelivery.SetupRoutes(server, httpdelivery.AppDeps{
		Log:               log,
		MachineService:    machineService,
		OperatorService:   operatorService,
		AssignmentService: assigmentService,
	})

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)

	log.Infof("Server running on %s", addr)

	// Pornește serverul HTTP
	if err := server.Run(addr); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
}
