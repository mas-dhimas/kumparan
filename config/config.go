package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigLoader struct {
	file string
}

// NewConfig create new ConfigLoader to GetServiceConfig()
func NewConfig(file string) *ConfigLoader {
	return &ConfigLoader{
		file: file,
	}
}

func (c *ConfigLoader) GetServiceConfig() (*ServiceConfig, error) {
	config := &ServiceConfig{}

	err := cleanenv.ReadConfig(c.file, config)
	if err != nil {
		// Return early if the error is not 'file not found'
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		// If the error is 'file not found', try reading from environment variables
		err = cleanenv.ReadEnv(config)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

// ServiceConfig stores the whole configuration for service.
type ServiceConfig struct {
	ServiceData ServiceDataConfig `yaml:"service_data"`
	SourceData  SourceDataConfig  `yaml:"source_data"`
}

// ServiceDataConfig contains the service data configuration.
type ServiceDataConfig struct {
	Address   string `yaml:"address" env:"SERVICE_DATA_PORT"`
	LogLevel  string `yaml:"log_level" env:"SERVICE_DATA_LOG_LEVEL"`
	RateLimit int    `yaml:"rate_limit" env:"SERVICE_DATA_RATE_LIMIT"`
}

// SourceDataConfig contains the source data configuration.
type SourceDataConfig struct {
	PostgresDBServer          string `yaml:"postgresdb_server" env:"SOURCE_DATA_POSTGRESDB_SERVER"`
	PostgresDBPort            int    `yaml:"postgresdb_port" env:"SOURCE_DATA_POSTGRESDB_PORT"`
	PostgresDBName            string `yaml:"postgresdb_name" env:"SOURCE_DATA_POSTGRESDB_NAME"`
	PostgresDBUsername        string `yaml:"postgresdb_username" env:"SOURCE_DATA_POSTGRESDB_USERNAME"`
	PostgresDBPassword        string `yaml:"postgresdb_password" env:"SOURCE_DATA_POSTGRESDB_PASSWORD"`
	PostgresDBTimeout         int    `yaml:"postgresdb_timeout" env:"SOURCE_DATA_POSTGRESDB_TIMEOUT"`
	PostgresDBMaxConns        int    `yaml:"postgresdb_max_conns" env:"SOURCE_DATA_POSTGRESDB_MAX_CONNS"`
	PostgresDBMinConns        int    `yaml:"postgresdb_min_conns" env:"SOURCE_DATA_POSTGRESDB_MIN_CONNS"`
	PostgresDBMaxConnLifetime int    `yaml:"postgresdb_max_conn_lifetime" env:"SOURCE_DATA_POSTGRESDB_MAX_CONN_LIFETIME"`
	PostgresDBMaxConnIdleTime int    `yaml:"postgresdb_max_conn_idle_time" env:"SOURCE_DATA_POSTGRESDB_MAX_CONN_IDLE_TIME"`
	ElasticURL                string `yaml:"elasticsearch_url" env:"SOURCE_DATA_ELASTICSEARCH_URL"`
}

func (sdc *SourceDataConfig) PostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		sdc.PostgresDBServer,
		sdc.PostgresDBPort,
		sdc.PostgresDBUsername,
		sdc.PostgresDBPassword,
		sdc.PostgresDBName,
		sdc.PostgresDBTimeout,
	)
}
