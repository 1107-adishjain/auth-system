package server

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"

    "github.com/1107-adishjain/auth-system/config"
    "github.com/1107-adishjain/auth-system/internal/auth"
    "github.com/1107-adishjain/auth-system/internal/middleware"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Server struct {
    router *gin.Engine
    cfg    *config.Config
    db     *gorm.DB
    redis  *redis.Client
}

func NewServer(cfg *config.Config) *Server {
    router := gin.Default()
    return &Server{
        router: router,
        cfg:    cfg,
    }
}

func (s *Server) Start(ctx context.Context) error {
    // Initialize Database Connection with GORM
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
        s.cfg.DBHost, s.cfg.DBUser, s.cfg.DBPassword, s.cfg.DBName, s.cfg.DBPort)
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    s.db = db

    // Initialize Redis Connection
    rdb := redis.NewClient(&redis.Options{
        Addr: s.cfg.RedisAddr,
    })
    if _, err := rdb.Ping(ctx).Result(); err != nil {
        return fmt.Errorf("unable to connect to redis: %w", err)
    }
    s.redis = rdb

    // Setup Repositories, Services, and Handlers
    authRepo := auth.NewRepository(s.db, s.redis)
    authService := auth.NewService(authRepo, s.cfg.JWTSecret, s.cfg.AccessTokenTTL, s.cfg.RefreshTokenTTL)
    authHandler := auth.NewAuthHandler(authService, s.cfg.AccessTokenTTL, s.cfg.RefreshTokenTTL)

    // Setup routes
    api := s.router.Group("/api/v1")
    auth.RegisterRoutes(api, authHandler)

    // Example of a protected route
    api.GET("/protected", middleware.AuthMiddleware(s.cfg.JWTSecret), func(c *gin.Context) {
        userID, _ := c.Get("userID")
        c.JSON(http.StatusOK, gin.H{"message": "This is a protected route", "user_id": userID})
    })

    // Health check endpoint
    s.router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    addr := fmt.Sprintf(":%s", s.cfg.ServerPort)
    log.Printf("Server starting on %s", addr)
    return s.router.Run(addr)
}
