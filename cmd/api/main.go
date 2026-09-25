package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/alchemist/content-studio-api/config"
	"github.com/alchemist/content-studio-api/internal/ads"
	"github.com/alchemist/content-studio-api/internal/ai"
	"github.com/alchemist/content-studio-api/internal/apps"
	"github.com/alchemist/content-studio-api/internal/auth"
	"github.com/alchemist/content-studio-api/internal/content"
	"github.com/alchemist/content-studio-api/internal/db"
	"github.com/alchemist/content-studio-api/internal/playstore"
	"github.com/alchemist/content-studio-api/internal/render"
	"github.com/alchemist/content-studio-api/internal/storage"
)

const renderWorkerCount = 2
const backgroundAudioPath = "assets/background.mp3"
const shutdownTimeout = 15 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r2, err := storage.NewR2(ctx, cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Bucket)
	if err != nil {
		log.Fatal(err)
	}

	authMiddleware, err := auth.NewMiddleware(cfg.SupabaseJWKSURL)
	if err != nil {
		log.Fatal(err)
	}

	aiRegistry := ai.NewRegistry(cfg.DefaultAIProvider)
	if cfg.AnthropicAPIKey != "" {
		aiRegistry.Register("claude", ai.NewClaudeProvider(cfg.AnthropicAPIKey))
	}
	if cfg.OpenAIAPIKey != "" {
		aiRegistry.Register("openai", ai.NewOpenAIProvider(cfg.OpenAIAPIKey))
	}
	if cfg.GeminiAPIKey != "" {
		aiRegistry.Register("gemini", ai.NewGeminiProvider(cfg.GeminiAPIKey))
	}

	metaClient := ads.NewMetaClient(cfg.MetaSystemUserToken, cfg.MetaAdAccountID)

	var playstoreClient *playstore.Client
	if cfg.GooglePlayServiceAccountJSONPath != "" {
		serviceAccountJSON, err := os.ReadFile(cfg.GooglePlayServiceAccountJSONPath)
		if err != nil {
			log.Fatal(err)
		}
		playstoreClient, err = playstore.NewClient(ctx, serviceAccountJSON)
		if err != nil {
			log.Fatal(err)
		}
	}

	contentRepo := content.NewRepo(pool)
	renderRepo := render.NewRepo(pool)
	adsRepo := ads.NewRepo(pool)
	appsRepo := apps.NewRepo(pool)

	renderPool := render.NewPool(renderRepo, contentRepo, r2, backgroundAudioPath, renderWorkerCount)
	renderPool.Start(ctx)

	cleanup := content.NewCleanup(contentRepo, r2, cfg.RenderedRetentionDays)
	cleanup.Start(ctx)

	contentHandler := content.NewHandler(contentRepo, aiRegistry, r2, renderRepo)
	renderHandler := render.NewHandler(renderRepo)
	adsHandler := ads.NewHandler(adsRepo, metaClient)

	var appsPublisher apps.Publisher
	if playstoreClient != nil {
		appsPublisher = playstoreClient
	}
	appsHandler := apps.NewHandler(appsRepo, r2, appsPublisher)

	router := gin.Default()

	if len(cfg.CORSAllowedOrigins) > 0 {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSAllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	router.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": "database unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/")
	api.Use(authMiddleware.RequireAuth())
	contentHandler.Register(api)
	renderHandler.Register(api)
	adsHandler.Register(api)
	appsHandler.Register(api)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
