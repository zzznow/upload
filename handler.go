package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadClient interface {
	Upload(ctx context.Context, objectKey string, data []byte, contentType string) (string, error)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "upload",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func uploadHandler(client UploadClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("read file failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		objectKey := c.PostForm("key")
		if objectKey == "" {
			ext := filepath.Ext(header.Filename)
			objectKey = fmt.Sprintf("uploads/%d%s", time.Now().UnixNano()/1e6, ext)
		}

		url, err := client.Upload(c.Request.Context(), objectKey, data, contentType)
		if err != nil {
			slog.Error("upload failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"url":  url,
				"key":  objectKey,
				"size": len(data),
			},
		})
	}
}

func allowedExt(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".mp4", ".mp3", ".wav", ".pdf", ".txt":
		return true
	}
	return false
}
