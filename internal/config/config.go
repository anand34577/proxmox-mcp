

package config

import (
	"fmt"
	"strings"
	"time"
	"github.com/spf13/viper"
)


type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Proxmox  ProxmoxConfig  `mapstructure:"proxmox"`
	Log      LogConfig      `mapstructure:"log"`
	Security SecurityConfig `mapstructure:"security"`
}


type ServerConfig struct {
	Transport               string        `mapstructure:"transport"`
	ListenAddr              string        `mapstructure:"listen_addr"`
	Name                    string        `mapstructure:"name"`
	Version                 string        `mapstructure:"version"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}


type ProxmoxConfig struct {
	BaseURL            string        `mapstructure:"base_url"`
	TokenID            string        `mapstructure:"token_id"`
	Secret             string        `mapstructure:"secret"`
	InsecureSkipVerify bool          `mapstructure:"insecure_skip_verify"`
	RequestTimeout     time.Duration `mapstructure:"request_timeout"`
}


type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}


type SecurityConfig struct {
	AllowDestructive bool     `mapstructure:"allow_destructive"`
	AllowedNodes     []string `mapstructure:"allowed_nodes"`
}





var envBindings = map[string]string{
	
	"server.transport":                 "PROXMOX_MCP_SERVER_TRANSPORT",
	"server.listen_addr":               "PROXMOX_MCP_SERVER_LISTEN_ADDR",
	"server.name":                      "PROXMOX_MCP_SERVER_NAME",
	"server.version":                   "PROXMOX_MCP_SERVER_VERSION",
	"server.graceful_shutdown_timeout": "PROXMOX_MCP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT",
	
	"proxmox.base_url":             "PROXMOX_MCP_PROXMOX_BASE_URL",
	"proxmox.token_id":             "PROXMOX_MCP_PROXMOX_TOKEN_ID",
	"proxmox.secret":               "PROXMOX_MCP_PROXMOX_SECRET",
	"proxmox.insecure_skip_verify": "PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY",
	"proxmox.request_timeout":      "PROXMOX_MCP_PROXMOX_REQUEST_TIMEOUT",
	
	"log.level":  "PROXMOX_MCP_LOG_LEVEL",
	"log.format": "PROXMOX_MCP_LOG_FORMAT",
	
	"security.allow_destructive": "PROXMOX_MCP_SECURITY_ALLOW_DESTRUCTIVE",
	"security.allowed_nodes":     "PROXMOX_MCP_SECURITY_ALLOWED_NODES",
}





func Load() (*Config, error) {
	v := viper.New()

	
	v.SetDefault("server.transport", "stdio")
	v.SetDefault("server.listen_addr", ":8080")
	v.SetDefault("server.name", "Proxmox MCP Server")
	v.SetDefault("server.version", "1.0.0")
	v.SetDefault("server.graceful_shutdown_timeout", "30s")

	v.SetDefault("proxmox.base_url", "https://localhost:8006")
	v.SetDefault("proxmox.insecure_skip_verify", false)
	v.SetDefault("proxmox.request_timeout", "30s")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	v.SetDefault("security.allow_destructive", false)

	
	v.SetEnvPrefix("PROXMOX_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	
	
	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("config: binding env %s: %w", env, err)
		}
	}

	
	if cfgFile := v.GetString("config_file"); cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("config: reading file %q: %w", cfgFile, err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshalling: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}


func (c *Config) validate() error {
	if c.Proxmox.BaseURL == "" {
		return fmt.Errorf("config: proxmox.base_url is required")
	}
	if c.Proxmox.TokenID == "" || c.Proxmox.Secret == "" {
		return fmt.Errorf("config: proxmox.token_id and proxmox.secret are required")
	}
	transport := strings.ToLower(c.Server.Transport)
	if transport != "stdio" && transport != "sse" && transport != "http" {
		return fmt.Errorf("config: server.transport must be stdio|sse|http, got %q", c.Server.Transport)
	}
	level := strings.ToLower(c.Log.Level)
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[level] {
		return fmt.Errorf("config: log.level must be debug|info|warn|error, got %q", c.Log.Level)
	}
	return nil
}

