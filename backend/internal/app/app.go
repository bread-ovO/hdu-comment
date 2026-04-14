package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/auth"
	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/database"
	"github.com/hdu-dp/backend/internal/handlers"
	adminHandlers "github.com/hdu-dp/backend/internal/handlers/admin"
	"github.com/hdu-dp/backend/internal/logging"
	"github.com/hdu-dp/backend/internal/middleware"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/router"
	backendserver "github.com/hdu-dp/backend/internal/server"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// App is the composed backend application.
type App struct {
	server *backendserver.HTTPServer
}

// New wires the backend dependencies and returns a runnable application.
func New(cfg *config.Config) (*App, error) {
	logger, err := logging.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}
	slog.SetDefault(logger)

	if cfg.Server.Mode != "" {
		gin.SetMode(cfg.Server.Mode)
	}
	gin.EnableJsonDecoderDisallowUnknownFields()

	db, err := database.Init(cfg)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	smsCodeRepo := repository.NewSMSCodeRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	reviewStatsRepo := repository.NewReviewStatsRepository(db)
	reviewReactionRepo := repository.NewReviewReactionRepository(db)
	siteStatsRepo := repository.NewSiteStatsRepository(db)

	emailCfg := config.LoadEmailConfig()
	emailService := services.NewEmailService(emailCfg)
	emailVerificationService := services.NewEmailVerificationService(
		emailVerificationRepo,
		userRepo,
		emailService,
		emailCfg.FrontendBaseURL,
	)

	storageProvider, err := storage.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
	qqOAuthService := services.NewQQOAuthService(
		cfg.Auth.QQ.Enabled,
		cfg.Auth.QQ.AppID,
		cfg.Auth.QQ.AppSecret,
		cfg.Auth.QQ.RedirectURI,
		cfg.Auth.JWTSecret,
	)
	wechatOAuthService := services.NewWeChatOAuthService(
		cfg.Auth.WeChat.Enabled,
		cfg.Auth.WeChat.AppID,
		cfg.Auth.WeChat.Secret,
	)

	authService := services.NewAuthService(
		userRepo,
		jwtManager,
		refreshRepo,
		smsCodeRepo,
		qqOAuthService,
		wechatOAuthService,
		services.AuthServiceOptions{
			RefreshTTL: cfg.Auth.RefreshTokenTTL,
			SMSCodeTTL: cfg.Auth.SMS.CodeTTL,
			SMSEnabled: cfg.Auth.SMS.Enabled,
			SMSDevMode: cfg.Auth.SMS.DevMode,
			AdminEmail: cfg.Admin.Email,
		},
	)
	reviewService := services.NewReviewService(reviewRepo, storageProvider)
	reviewStatsService := services.NewReviewStatsService(reviewStatsRepo, reviewReactionRepo, siteStatsRepo)

	authHandler := handlers.NewAuthHandler(authService, emailVerificationService)
	userHandler := handlers.NewUserHandler(userRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	reviewStatsHandler := handlers.NewReviewStatsHandler(reviewStatsService, reviewService)
	adminReviewHandler := adminHandlers.NewReviewAdminHandler(reviewService)
	adminUserHandler := adminHandlers.NewUserAdminHandler(userRepo)
	emailVerificationHandler := handlers.NewEmailVerificationHandler(emailVerificationService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.StructuredLogger(logger),
		middleware.Recovery(logger),
	)
	engine.Use(cors.New(buildCORSConfig(cfg)))

	staticUploads := cfg.Storage.UploadDir
	if cfg.Storage.Provider != "local" && cfg.Storage.Provider != "" {
		staticUploads = ""
	}

	router.Register(router.Params{
		Engine:                   engine,
		AuthMiddleware:           authMiddleware,
		AuthHandler:              authHandler,
		UserHandler:              userHandler,
		ReviewHandler:            reviewHandler,
		ReviewStatsHandler:       reviewStatsHandler,
		EmailVerificationHandler: emailVerificationHandler,
		AdminHandler:             adminReviewHandler,
		AdminUserHandler:         adminUserHandler,
		StaticUploadDir:          staticUploads,
	})

	return &App{
		server: backendserver.New(cfg, engine, logger),
	}, nil
}

// Run starts the backend server and blocks until shutdown completes.
func (a *App) Run(ctx context.Context) error {
	return a.server.Run(ctx)
}

func buildCORSConfig(cfg *config.Config) cors.Config {
	allowedOrigins := cfg.CORS.AllowOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
			"https://hddp.blueloaf.top",
		}
	}

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
	}
}
