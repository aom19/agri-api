package main

import (
	"context"
	"fmt"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/config"
	"agri-api/internal/db"
	httpdelivery "agri-api/internal/delivery/http"
	"agri-api/internal/email"
	"agri-api/internal/logger"
	redisclient "agri-api/internal/redis"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/store"
	"agri-api/internal/usecase"

	_ "agri-api/docs"

	"github.com/gin-gonic/gin"
)

// @title           Agri API
// @version         1.0
// @description     API REST pentru managementul mașinilor agricole, operatorilor și asignărilor.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Introdu token-ul cu prefixul "Bearer ": **Bearer &lt;token&gt;**

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

	// 2.1 Creează repository-ul pentru utilaje agricole și serviciul aferent
	implementRepo := postgres.NewImplementRepo(sqlDB)
	implementService := usecase.NewImplementService(implementRepo)

	// 2.2 Creează repository-urile pentru resurse și serviciul aferent
	resourceTypeRepo := postgres.NewResourceTypeRepo(sqlDB)
	resourceRepo := postgres.NewResourceRepo(sqlDB)
	resourceService := usecase.NewResourceService(resourceTypeRepo, resourceRepo)
	stockRepo := postgres.NewStockRepo(sqlDB)
	stockService := usecase.NewStockService(stockRepo, resourceRepo)

	// 2.3 Creează repository-ul pentru terenuri și serviciul aferent
	fieldRepo := postgres.NewFieldRepo(sqlDB)
	fieldService := usecase.NewFieldService(fieldRepo)

	// 2.4 Creează repository-urile pentru operațiuni și serviciul aferent
	operationTypeRepo := postgres.NewOperationTypeRepo(sqlDB)
	operationTemplateRepo := postgres.NewOperationTemplateRepo(sqlDB)
	operationService := usecase.NewOperationService(operationTypeRepo, operationTemplateRepo)

	// 2.4.1 Repository și serviciul pentru operațiuni pe teren
	fieldOperationRepo := postgres.NewFieldOperationRepo(sqlDB)
	fieldOperationService := usecase.NewFieldOperationService(fieldOperationRepo)

	// 2.6 Audit & Notifications
	auditRepo := postgres.NewAuditRepo(sqlDB)
	auditService := usecase.NewAuditService(auditRepo)
	notificationRepo := postgres.NewNotificationRepo(sqlDB)
	notificationService := usecase.NewNotificationService(notificationRepo)

	// 2.6.1 Monitor: notifică admin/manager și operatorul când o operațiune în lucru
	// depășește timpul estimat de lucru (sfârșitul planificat)
	overdueInterval, err := time.ParseDuration(cfg.FieldOperationOverdueCheckInterval)
	if err != nil || overdueInterval <= 0 {
		overdueInterval = time.Minute
	}
	overdueMonitor := usecase.NewFieldOperationOverdueMonitor(
		fieldOperationRepo,
		notificationService,
		auditService,
		log,
		overdueInterval,
	)
	go overdueMonitor.Start(context.Background())

	// 2.5 Repository pentru compatibilitățile utilaj ↔ echipament
	implementCompatibilityRepo := postgres.NewImplementCompatibilityRepo(sqlDB)
	dashboardRepo := postgres.NewDashboardRepo(sqlDB)
	dashboardService := usecase.NewDashboardService(dashboardRepo, auditRepo)
	weatherService := usecase.NewWeatherService(cfg.OpenWeatherAPIKey)

	// 3. Inițializează store-ul cu toate repository-urile și serviciul de asignări
	appStore := store.NewInitialiedStore(sqlDB)
	assigmentService := usecase.NewAssigmentService(appStore)

	// 4. Inițializează serviciile de autentificare
	accessTTL, err := time.ParseDuration(cfg.AccessTokenTTL)
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(cfg.RefreshTokenTTL)
	if err != nil {
		refreshTTL = 7 * 24 * time.Hour
	}
	jwtService := auth.NewJWTService(cfg.JWTSecret, accessTTL, refreshTTL)
	refreshRepo := auth.NewRepo(sqlDB)

	// 5. Inițializează Redis și blacklist-ul pentru access tokens
	rdb, err := redisclient.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Info("Redis connected successfully")
	blacklist := auth.NewBlacklist(rdb)

	emailService := email.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
		log,
	)

	authService := usecase.NewAuthService(appStore, jwtService, refreshRepo, blacklist, emailService, cfg.ClientOrigin)
	rbacService := usecase.NewRBACService(appStore, jwtService, refreshRepo, blacklist)
	userService := usecase.NewUserService(appStore)

	// 6. Profile service
	uploadDir := "uploads/avatars"
	profileService := usecase.NewProfileService(appStore, uploadDir, cfg.PublicURL)

	// Configurează serverul HTTP Gin fără middleware implicit
	server := gin.New()
	// Adaugă middleware pentru logarea request-urilor și recuperare din panic
	server.Use(gin.Logger(), gin.Recovery())

	// CORS middleware — TREBUIE să fie înainte de rute
	server.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.ClientOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Dezactivează avertismentele legate de proxy-uri de încredere
	if err := server.SetTrustedProxies(nil); err != nil {
		log.Fatalf("SetTrustedProxies: %v", err)
	}

	// Înregistrează toate rutele API
	httpdelivery.SetupRoutes(server, httpdelivery.AppDeps{
		Log:                        log,
		MachineService:             machineService,
		ResourceService:            resourceService,
		StockService:               stockService,
		ImplementService:           implementService,
		OperatorService:            operatorService,
		FieldService:               fieldService,
		AssignmentService:          assigmentService,
		OperationService:           operationService,
		FieldOperationService:      fieldOperationService,
		DashboardService:           dashboardService,
		WeatherService:             weatherService,
		AuditService:               auditService,
		NotificationService:        notificationService,
		AuditRepo:                  auditRepo,
		ImplementCompatibilityRepo: implementCompatibilityRepo,
		AuthService:                authService,
		ProfileService:             profileService,
		RBACService:                rbacService,
		UserService:                userService,
		PermissionRepo:             appStore.PermissionRepo,
		UploadDir:                  uploadDir,
		JWTService:                 jwtService,
		Blacklist:                  blacklist,
	})

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)

	log.Infof("Server running on %s", addr)

	// Pornește serverul HTTP
	if err := server.Run(addr); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
}
