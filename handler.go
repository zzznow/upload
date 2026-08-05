package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzznow/common"
)

type UploadClient interface {
	Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) (string, error)
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
			common.ErrorMsg(c, http.StatusBadRequest, "upload: missing file")
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		objectKey := c.PostForm("key")
		if objectKey == "" {
			ext := filepath.Ext(header.Filename)
			objectKey = fmt.Sprintf("%s/%d%s", time.Now().Format("2006/01/02"), time.Now().UnixNano()/1e6, ext)
		}

		slog.Info("upload starting", "filename", header.Filename, "size", header.Size, "contentType", contentType, "key", objectKey)

		url, err := client.Upload(c.Request.Context(), objectKey, file, header.Size, contentType)
		if err != nil {
			slog.Warn("stream upload failed, falling back to buffered", "error", err)
			if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
				slog.Error("seek file failed", "error", seekErr)
				common.ErrorMsg(c, http.StatusInternalServerError, "upload: read file failed")
				return
			}
			data, readErr := io.ReadAll(file)
			if readErr != nil {
				slog.Error("read file failed on fallback", "error", readErr)
				common.ErrorMsg(c, http.StatusInternalServerError, "upload: read file failed")
				return
			}
			url, err = client.Upload(c.Request.Context(), objectKey, bytes.NewReader(data), int64(len(data)), contentType)
			if err != nil {
				slog.Error("upload failed", "filename", header.Filename, "size", len(data), "key", objectKey, "error", err.Error())
				common.ErrorMsg(c, http.StatusInternalServerError, fmt.Sprintf("upload: %s", err.Error()))
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"url":  url,
					"key":  objectKey,
					"size": len(data),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"url":  url,
				"key":  objectKey,
				"size": header.Size,
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

