package config

import (
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

const defaultEnvironment = "dev"

var supportedEnvironments = map[string]bool{
	"default": true,
	"dev":     true,
	"test":    true,
	"prod":    true,
}

// Load 根据APP_ENV加载环境配置，未指定时默认使用dev环境。
func Load() (string, error) {
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = defaultEnvironment
	}
	if !supportedEnvironments[environment] {
		return "", fmt.Errorf("unsupported APP_ENV %q", environment)
	}
	fileName := "config/config.yaml"
	if environment != "default" {
		fileName = fmt.Sprintf("config/config.%s.yaml", environment)
	}
	adapter, err := gcfg.NewAdapterFile(fileName)
	if err != nil {
		return "", fmt.Errorf("create config adapter: %w", err)
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		if err = adapter.Set("database.default.link", databaseDSN); err != nil {
			return "", fmt.Errorf("override database DSN: %w", err)
		}
	}
	g.Cfg().SetAdapter(adapter)
	return environment, nil
}
