package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/auth"
	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/database"
	"github.com/hdu-dp/backend/internal/handlers"
	adminHandlers "github.com/hdu-dp/backend/internal/handlers/admin"
	"github.com/hdu-dp/backend/internal/middleware"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/router"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// @title           杭电点评 API
// @version         1.0
// @description     这是一个简单的校园点评网站的后端 API 文档。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apiKey  ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.Server.Mode != "" {
		gin.SetMode(cfg.Server.Mode)
	}

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)

	storageProvider, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)

	authService := services.NewAuthService(userRepo, jwtManager, refreshRepo, cfg.Auth.RefreshTokenTTL)
	reviewService := services.NewReviewService(reviewRepo, storageProvider)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	adminReviewHandler := adminHandlers.NewReviewAdminHandler(reviewService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	staticUploads := cfg.Storage.UploadDir
	if cfg.Storage.Provider != "local" && cfg.Storage.Provider != "" {
		staticUploads = ""
	}

	router.Register(router.Params{
		Engine:          engine,
		AuthMiddleware:  authMiddleware,
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		ReviewHandler:   reviewHandler,
		AdminHandler:    adminReviewHandler,
		StaticUploadDir: staticUploads,
	})

	if err := engine.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
