package config

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/types"
	"github.com/spf13/viper"
)

type ConfigOidcProvider struct {
	Name         string `mapstructure:"name"`
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	IssuerUrl    string `mapstructure:"issuer_url"`
	RedirectUrl  string `mapstructure:"redirect_url"`
}

type Config struct {
	RunMigrations    bool   `mapstructure:"run_migrations"`
	ListenAddr       string `mapstructure:"listen_addr"`
	DataDir          string `mapstructure:"data_dir"`
	JwtSecret        string `mapstructure:"jwt_secret"`

	OidcProviders map[string]ConfigOidcProvider `mapstructure:"oidc_providers"`
}

func (c *Config) WorkDir() types.WorkDir {
	return types.WorkDir(c.DataDir)
}

func setDefaults() {
	viper.SetDefault("run_migrations", "true")
	viper.SetDefault("listen_addr", ":3000")
	viper.BindEnv("data_dir")
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

	override := viper.GetString("config_override")
	if override != "" {
		viper.SetConfigFile(override)
		viper.MergeInConfig()
	}

	err = viper.Unmarshal(&LoadedConfig)
	if err != nil {
		slog.Error("Failed to unmarshal config: ", err)
		os.Exit(-1)
	}

	configCopy := LoadedConfig
	// configCopy.OidcClientId = "***"
	// configCopy.OidcClientSecret = "***"
	// configCopy.JwtSecret = "***"

	slog.Info("Current Config", "config", configCopy)

	validateConfig(&LoadedConfig)
}
