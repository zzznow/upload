package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zzznow/common"
)

func main() {
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "prod"
	}
	slog.Info("upload service starting", "mode", mode)

	if err := initConfig(mode); err != nil {
		slog.Error("config init failed", "error", err)
		os.Exit(1)
	}

	backend := DetectBackend()
	if backend == "" {
		slog.Error("no upload backend configured, set COS or OSS credentials")
		os.Exit(1)
	}

	var client UploadClient
	switch backend {
	case "COS":
		client = NewCOSClient()
	case "OSS":
		client = NewOSSClient()
	}
	slog.Info("upload backend ready", "backend", backend)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(common.TraceMiddleware())
	r.Use(func(c *gin.Context) { common.SetServiceName(c, "upload"); c.Next() })

	r.GET("/health", healthHandler)
	r.GET("/healthz", healthHandler)

	api := r.Group("/upload")
	{
		api.POST("/file", uploadHandler(client))
	}

	port := viper.GetString("port")
	if port == "" {
		port = "80"
	}
	addr := ":" + port

	srv := &http.Server{Addr: addr, Handler: r.Handler()}

	go func() {
		slog.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
	}
	slog.Info("shutdown complete")
}
