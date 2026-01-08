package config

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/types"
	"github.com/spf13/viper"
)

type Config struct {
	RunMigrations    bool   `mapstructure:"run_migrations"`
	ListenAddr       string `mapstructure:"listen_addr"`
	DataDir          string `mapstructure:"data_dir"`
	OidcClientId     string `mapstructure:"oidc_client_id"`
	OidcClientSecret string `mapstructure:"oidc_client_secret"`
	OidcIssuerUrl    string `mapstructure:"oidc_issuer_url"`
	OidcRedirectUrl  string `mapstructure:"oidc_redirect_url"`
	JwtSecret        string `mapstructure:"jwt_secret"`
}

func (c *Config) WorkDir() types.WorkDir {
	return types.WorkDir(c.DataDir)
}

func setDefaults() {
	viper.SetDefault("run_migrations", "true")
	viper.SetDefault("listen_addr", ":3000")
	viper.BindEnv("data_dir")
	viper.BindEnv("oidc_client_id")
	viper.BindEnv("oidc_client_secret")
	viper.BindEnv("oidc_issuer_url")
	viper.BindEnv("oidc_redirect_url")
	viper.BindEnv("jwt_secret")
}

func validateConfig(config *Config) {
	hasError := false

	validate := func(expr bool, msg string) {
		if expr {
			slog.Error("Config Validation", "err", msg)
			hasError = true
		}
	}

	// NOTE(patrik): Has default value, here for completeness
	// validate(config.RunMigrations == "", "run_migrations needs to be set")
	validate(config.ListenAddr == "", "listen_addr needs to be set")
	validate(config.DataDir == "", "data_dir needs to be set")
	validate(config.OidcClientId == "", "oidc_client_id needs to be set")
	validate(config.OidcClientSecret == "", "oidc_client_secret needs to be set")
	validate(config.OidcIssuerUrl == "", "oidc_issuer_url needs to be set")
	validate(config.OidcRedirectUrl == "", "oidc_redirect_url needs to be set")
	validate(config.JwtSecret == "", "jwt_secret needs to be set")

	if hasError {
		slog.Error("Config not valid")
		os.Exit(-1)
	}
}

var ConfigFile string
var LoadedConfig Config

func InitConfig() {
	setDefaults()

	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix(authlab.AppName)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("Failed to load config", "err", err)
	}

	err = viper.Unmarshal(&LoadedConfig)
	if err != nil {
		slog.Error("Failed to unmarshal config: ", err)
		os.Exit(-1)
	}

	configCopy := LoadedConfig
	configCopy.OidcClientId = "***"
	configCopy.OidcClientSecret = "***"
	configCopy.JwtSecret = "***"

	slog.Info("Current Config", "config", configCopy)

	validateConfig(&LoadedConfig)
}
