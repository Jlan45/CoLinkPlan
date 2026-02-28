package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"CoLinkPlan/internal/config"
	"CoLinkPlan/internal/db"
	"CoLinkPlan/internal/limiter"
	"CoLinkPlan/internal/server"
	"CoLinkPlan/pkg/logger"
	"CoLinkPlan/web"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadServerConfig()

	logger.Log.Info("Starting Co-Link Server", "port", cfg.Port)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Log.Error("Failed to connect to postgres database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.InitializeSchema(); err != nil {
		logger.Log.Error("Failed to initialize database schema", "err", err)
	}

	rl, err := limiter.NewRateLimiter(cfg.RedisURL)
	if err != nil {
		logger.Log.Error("Failed to connect to redis rate limiter", "err", err)
		os.Exit(1)
	}

	hub := server.NewHub()
	go hub.Run()

	gw := server.NewGateway(hub, database, rl)

	router := gin.Default()

	// Add very permissive CORS for local dev testing with vite
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Client-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API routes
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", gw.ChatCompletionsHandler)
		v1.GET("/models", gw.ModelsHandler)
		v1.GET("/models/:model", gw.ModelsHandler)
	}

	// Internal Dashboard APIs
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", server.RegisterHandler(database))
			auth.POST("/login", server.LoginHandler(database))
		}

		// Public API: nodes are visible without auth
		api.GET("/nodes", server.NodesHandler(hub))

		protected := api.Group("/")
		protected.Use(server.AuthMiddleware())
		{
			protected.GET("/user/me", server.MeHandler(database))
		}
	}

	// WS endpoint for Clients
	router.GET("/ws", gw.WsHandler)

	// Serve Embedded React Frontend
	uiFS, uiFSErr := web.GetUIFS()
	if uiFSErr != nil {
		logger.Log.Error("Failed to load embedded UI filesystem", "err", uiFSErr)
	}

	// serveFile reads a file from the embedded FS and writes it to the response.
	// We avoid http.FileServer / FileFromFS entirely to prevent the 301 redirect
	// that net/http emits when URL path != file name (e.g. "/" != "index.html").
	serveFile := func(c *gin.Context, name string) bool {
		if uiFS == nil {
			return false
		}
		f, err := uiFS.Open(name)
		if err != nil {
			return false
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			return false
		}
		data, err := io.ReadAll(f)
		if err != nil {
			return false
		}
		ct := mime.TypeByExtension(filepath.Ext(name))
		if ct == "" {
			ct = "application/octet-stream"
		}
		c.Data(http.StatusOK, ct, data)
		return true
	}

	router.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if strings.HasPrefix(urlPath, "/api") || strings.HasPrefix(urlPath, "/v1") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// Try the exact path first (handles /assets/... and other static files)
		cleanPath := strings.TrimPrefix(urlPath, "/")
		if cleanPath != "" && serveFile(c, cleanPath) {
			return
		}

		// Fallback: send index.html for all React Router routes (and root "/")
		if !serveFile(c, "index.html") {
			c.String(http.StatusNotFound, "Frontend unavailable")
		}
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
