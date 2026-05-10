package main

import (
	"fmt"
	"os"

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
	BucketURL string
	Region    string
}

type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
}

func initConfig(env string) error {
	viper.SetConfigName("application-prod")
	viper.SetConfigType("yml")

	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "."
	}
	viper.AddConfigPath(configDir)

	viper.SetEnvPrefix("UPLOAD")
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
