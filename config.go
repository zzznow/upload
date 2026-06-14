package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port string
	COS  COSConfig
	OSS  OSSConfig
}

type COSConfig struct {
	SecretID  string
	SecretKey string
	Endpoint  string
	BucketURL string
	Region    string
}

type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
}

func initConfig(mode string) error {
	viper.SetConfigName("application-" + mode)
	viper.SetConfigType("yml")

	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "."
	}
	viper.AddConfigPath(configDir)

	viper.SetEnvPrefix("UPLOAD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("config file not found, using env vars only")
		} else {
			return fmt.Errorf("read config: %w", err)
		}
	}

	return nil
}

func DetectBackend() string {
	if viper.GetString("cos.secret_id") != "" && viper.GetString("cos.secret_key") != "" {
		return "COS"
	}
	if viper.GetString("oss.access_key_id") != "" && viper.GetString("oss.access_key_secret") != "" {
		return "OSS"
	}
	return ""
}
