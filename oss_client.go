package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type OSSClient struct {
	accessKeyID     string
	accessKeySecret string
	endpoint        string
	bucket          string
	client          *http.Client
}

func NewOSSClient() *OSSClient {
	return &OSSClient{
		accessKeyID:     viper.GetString("oss.access_key_id"),
		accessKeySecret: viper.GetString("oss.access_key_secret"),
		endpoint:        viper.GetString("oss.endpoint"),
		bucket:          viper.GetString("oss.bucket"),
		client:          &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OSSClient) Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) (string, error) {
	objectURL := fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucket, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = size
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	c.signRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("oss upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("oss upload failed", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("oss: HTTP %d", resp.StatusCode)
	}

	slog.Info("oss upload success", "key", objectKey, "size", size)
	return objectURL, nil
}

func (c *OSSClient) signRequest(req *http.Request) {
	stringToSign := fmt.Sprintf("%s\n\n%s\n%s\n%s",
		req.Method,
		req.Header.Get("Content-Type"),
		req.Header.Get("Date"),
		"/"+c.bucket+req.URL.Path,
	)

	mac := hmac.New(sha1.New, []byte(c.accessKeySecret))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	auth := fmt.Sprintf("OSS %s:%s", c.accessKeyID, signature)
	req.Header.Set("Authorization", auth)
}
