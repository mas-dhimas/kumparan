package config_test

import (
	"kumparan-test/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServiceConfig_EnvFallback(t *testing.T) {
	os.Setenv("SERVICE_DATA_PORT", "8080")
	os.Setenv("SERVICE_DATA_LOG_LEVEL", "debug")
	os.Setenv("SERVICE_DATA_RATE_LIMIT", "100")

	os.Setenv("SOURCE_DATA_POSTGRESDB_SERVER", "localhost")
	os.Setenv("SOURCE_DATA_POSTGRESDB_PORT", "5432")
	os.Setenv("SOURCE_DATA_POSTGRESDB_NAME", "mydb")
	os.Setenv("SOURCE_DATA_POSTGRESDB_USERNAME", "user")
	os.Setenv("SOURCE_DATA_POSTGRESDB_PASSWORD", "pass")
	os.Setenv("SOURCE_DATA_POSTGRESDB_TIMEOUT", "5")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONNS", "10")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MIN_CONNS", "1")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONN_LIFETIME", "60")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONN_IDLE_TIME", "30")
	os.Setenv("SOURCE_DATA_ELASTICSEARCH_URL", "http://localhost:9200")

	loader := config.NewConfig("nonexistent.yaml")
	cfg, err := loader.GetServiceConfig()

	expectedDSN := "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable connect_timeout=5"
	assert.Equal(t, expectedDSN, cfg.SourceData.PostgresDSN())

	assert.NoError(t, err)
	assert.Equal(t, "8080", cfg.ServiceData.Address)
	assert.Equal(t, "localhost", cfg.SourceData.PostgresDBServer)
}

func TestGetServiceConfig_EnvFallbackErr(t *testing.T) {
	os.Setenv("SERVICE_DATA_PORT", "8080")
	os.Setenv("SERVICE_DATA_LOG_LEVEL", "debug")
	os.Setenv("SERVICE_DATA_RATE_LIMIT", "100")

	os.Setenv("SOURCE_DATA_POSTGRESDB_SERVER", "localhost")
	os.Setenv("SOURCE_DATA_POSTGRESDB_PORT", "5432")
	os.Setenv("SOURCE_DATA_POSTGRESDB_NAME", "mydb")
	os.Setenv("SOURCE_DATA_POSTGRESDB_USERNAME", "user")
	os.Setenv("SOURCE_DATA_POSTGRESDB_PASSWORD", "pass")
	os.Setenv("SOURCE_DATA_POSTGRESDB_TIMEOUT", "5")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONNS", "10")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MIN_CONNS", "1")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONN_LIFETIME", "null")
	os.Setenv("SOURCE_DATA_POSTGRESDB_MAX_CONN_IDLE_TIME", "30")
	os.Setenv("SOURCE_DATA_ELASTICSEARCH_URL", "http://localhost:9200")

	loader := config.NewConfig("nonexistent.yaml")
	_, err := loader.GetServiceConfig()
	assert.Error(t, err)
}

func TestGetServiceConfig_InvalidFile(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "bad*.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("not: valid: yaml: - structure")
	tmpFile.Close()

	loader := config.NewConfig(tmpFile.Name())
	_, err := loader.GetServiceConfig()
	assert.Error(t, err)
}
