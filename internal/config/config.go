package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/go-pkgs/observability"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Config is the root application configuration.
type Config struct {
	Server  ServerConfig         `mapstructure:"server"`
	Logging logging.Config       `mapstructure:"logging"`
	Opentel observability.Config `mapstructure:"opentel"`
	Mongo   MongoConfig          `mapstructure:"mongo"`
	Swagger SwaggerConfig        `mapstructure:"swagger"`
}

// ServerConfig holds HTTP listen and CORS settings.
type ServerConfig struct {
	HTTP HTTPConfig  `mapstructure:"http"`
	CORS CORSConfig  `mapstructure:"cors"`
	Env  Environment `mapstructure:"environment"`
}

// Environment is the deployment environment name.
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) IsDevelopment() bool {
	return e == EnvironmentDevelopment
}

// HTTPConfig holds the REST API listen settings.
type HTTPConfig struct {
	Port         int           `mapstructure:"port" validate:"required"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout" validate:"required"`
}

// CORSConfig holds CORS settings for the Fiber HTTP server.
type CORSConfig struct {
	AllowOrigins        string `mapstructure:"allowOrigins"`
	AllowOriginSuffixes string `mapstructure:"allowOriginSuffixes"`
	AllowMethods        string `mapstructure:"allowMethods"`
	AllowHeaders        string `mapstructure:"allowHeaders"`
	ExposeHeaders       string `mapstructure:"exposeHeaders"`
	AllowCredentials    bool   `mapstructure:"allowCredentials"`
}

// MongoConfig backs persistence.
type MongoConfig struct {
	URI      string `mapstructure:"uri" validate:"required"`
	Database string `mapstructure:"database" validate:"required"`
}

// SwaggerConfig gates OpenAPI UI exposure.
type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LoadConfig loads base YAML by basename, optional "<base>.<env>.yaml" overlay,
// expands ${VAR} placeholders, then unmarshals and validates.
func LoadConfig(logger logging.Logger, configPath string, env string) (*Config, error) {
	const envFile = ".env"
	if _, err := os.Stat(envFile); err == nil {
		if err := gotenv.Load(envFile); err == nil {
			logger.Info("loaded environment file", logging.Fields{"path": envFile})
		}
	}

	baseConfig := viper.New()
	baseConfig.SetConfigType("yaml")
	baseConfig.AutomaticEnv()
	baseConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	baseConfig.AddConfigPath(".")
	baseConfig.AddConfigPath("..")
	baseConfig.AddConfigPath("../..")

	configName := filepath.Base(configPath)
	configExt := filepath.Ext(configName)
	configNameWithoutExt := configName[:len(configName)-len(configExt)]
	baseConfig.SetConfigName(configNameWithoutExt)

	setDefaults(baseConfig)

	if err := baseConfig.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if strings.TrimSpace(env) != "" {
		envConfig := viper.New()
		envConfig.SetConfigType("yaml")
		envConfig.AddConfigPath(".")
		envConfig.AddConfigPath("..")
		envConfig.AddConfigPath("../..")
		envConfigName := fmt.Sprintf("%s.%s", configNameWithoutExt, env)
		envConfig.SetConfigName(envConfigName)
		if err := envConfig.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("environment config file %q not found or unreadable: %w", envConfigName, err)
		}
		if err := baseConfig.MergeConfigMap(envConfig.AllSettings()); err != nil {
			return nil, fmt.Errorf("merge environment config: %w", err)
		}
		logger.Info("merged environment config", logging.Fields{"name": envConfigName})
	}

	settings := baseConfig.AllSettings()
	expandEnvInMap(settings)
	if err := baseConfig.MergeConfigMap(settings); err != nil {
		return nil, fmt.Errorf("failed to apply expanded env settings: %w", err)
	}

	var cfg Config
	if err := baseConfig.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.http.port", 8080)
	v.SetDefault("server.http.readTimeout", "30s")
	v.SetDefault("server.http.writeTimeout", "30s")
	v.SetDefault("server.http.idleTimeout", "120s")
	v.SetDefault("server.environment", "development")
	v.SetDefault("server.cors.allowOrigins", "*")
	v.SetDefault("server.cors.allowMethods", "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS")
	v.SetDefault("server.cors.allowHeaders", "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Correlation-ID")
	v.SetDefault("logging.service_name", "golang-service-template")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("opentel.enabled", false)
	v.SetDefault("opentel.service_name", "golang-service-template")
	v.SetDefault("opentel.insecure", true)
	v.SetDefault("opentel.sampling.ratio", 1.0)
	v.SetDefault("mongo.uri", "mongodb://admin:password@127.0.0.1:27017/?replicaSet=rs0&authSource=admin")
	v.SetDefault("mongo.database", "golang-service-template")
	v.SetDefault("swagger.enabled", false)
	v.SetDefault("swagger.username", "admin")
	v.SetDefault("swagger.password", "changeme")
}

func expandEnvInMap(m map[string]interface{}) {
	for k, v := range m {
		switch typed := v.(type) {
		case string:
			m[k] = os.ExpandEnv(typed)
		case map[string]interface{}:
			expandEnvInMap(typed)
		case map[interface{}]interface{}:
			norm := make(map[string]interface{}, len(typed))
			for nk, nv := range typed {
				keyStr, ok := nk.(string)
				if !ok {
					continue
				}
				norm[keyStr] = nv
			}
			expandEnvInMap(norm)
			m[k] = norm
		}
	}
}

func validateConfig(cfg *Config) error {
	v := validator.New()
	return v.Struct(cfg)
}
