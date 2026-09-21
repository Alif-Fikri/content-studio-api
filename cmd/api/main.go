package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alchemist/content-studio-api/config"
	"github.com/alchemist/content-studio-api/internal/ads"
	"github.com/alchemist/content-studio-api/internal/ai"
	"github.com/alchemist/content-studio-api/internal/auth"
	"github.com/alchemist/content-studio-api/internal/content"
	"github.com/alchemist/content-studio-api/internal/db"
	"github.com/alchemist/content-studio-api/internal/render"
	"github.com/alchemist/content-studio-api/internal/storage"
)

const renderWorkerCount = 2
const backgroundAudioPath = "assets/background.mp3"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

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

	contentRepo := content.NewRepo(pool)
	renderRepo := render.NewRepo(pool)
	adsRepo := ads.NewRepo(pool)

	renderPool := render.NewPool(renderRepo, contentRepo, r2, backgroundAudioPath, renderWorkerCount)
	renderPool.Start(ctx)

	cleanup := content.NewCleanup(contentRepo, r2, cfg.RenderedRetentionDays)
	cleanup.Start(ctx)

	contentHandler := content.NewHandler(contentRepo, aiRegistry, r2, renderRepo)
	renderHandler := render.NewHandler(renderRepo)
	adsHandler := ads.NewHandler(adsRepo, metaClient)

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/")
	api.Use(authMiddleware.RequireAuth())
	contentHandler.Register(api)
	renderHandler.Register(api)
	adsHandler.Register(api)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
