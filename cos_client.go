package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type COSClient struct {
	secretID  string
	secretKey string
	endpoint  string
	cdnURL    string
	region    string
	client    *http.Client
}

func NewCOSClient() *COSClient {
	return &COSClient{
		secretID:  viper.GetString("cos.secret_id"),
		secretKey: viper.GetString("cos.secret_key"),
		endpoint:  viper.GetString("cos.endpoint"),
		cdnURL:    viper.GetString("cos.bucket_url"),
		region:    viper.GetString("cos.region"),
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *COSClient) Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) (string, error) {
	if c.endpoint == "" {
		c.endpoint = c.cdnURL
	}
	objectURL := fmt.Sprintf("%s/%s", c.endpoint, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = size

	c.signRequest(req, objectKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cos upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("cos upload failed", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("cos: HTTP %d", resp.StatusCode)
	}

	slog.Info("cos upload success", "key", objectKey, "size", size)
	publicURL := objectURL
	if c.cdnURL != "" && c.cdnURL != c.endpoint {
		publicURL = fmt.Sprintf("%s/%s", c.cdnURL, objectKey)
	}
	return publicURL, nil
}

func (c *COSClient) signRequest(req *http.Request, objectKey string) {
	t := time.Now().UTC()
	signTime := fmt.Sprintf("%d;%d", t.Unix()-300, t.Unix()+7200)

	lowerMethod := strings.ToLower(req.Method)
	host := req.URL.Host

	httpString := fmt.Sprintf("%s\n/%s\n\nhost=%s\n", lowerMethod, objectKey, host)
	sha1Hash := sha1Sum([]byte(httpString))

	strToSign := fmt.Sprintf("sha1\n%s\n%s\n", signTime, sha1Hash)

	mac := hmac.New(sha1.New, []byte(c.secretKey))
	mac.Write([]byte(signTime))

	signKey := hex.EncodeToString(mac.Sum(nil))

	mac2 := hmac.New(sha1.New, []byte(signKey))
	mac2.Write([]byte(strToSign))
	signature := hex.EncodeToString(mac2.Sum(nil))

	auth := fmt.Sprintf(
		"q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=host&q-url-param-list=&q-signature=%s",
		c.secretID, signTime, signTime, signature,
	)
	req.Header.Set("Authorization", auth)
}

func sha1Sum(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
