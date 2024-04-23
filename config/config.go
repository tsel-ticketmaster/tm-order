package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var (
	c              *Config
	configSyncOnce sync.Once
)

type Config struct {
	Application struct {
		Name        string
		Port        int
		Environment string
		Debug       bool
		Timeout     time.Duration
		Timezone    *time.Location
		TMUser      struct {
			BaseURL string
		}
		TMOrder struct {
			BaseURL string
		}
	}
	Crypto struct {
		Secret string
	}
	OpenTelemetry struct {
		Collector struct {
			Endpoint string
		}
	}
	CORS struct {
		AllowedOrigins   []string
		AllowedMethods   []string
		AllowedHeaders   []string
		ExposedHeaders   []string
		MaxAge           int
		AllowCredentials bool
	}
	JWT struct {
		PrivateKey []byte
		PublicKey  []byte
	}
	Postgresql struct {
		Host         string
		Port         int
		User         string
		Password     string
		DBName       string
		SSLMode      string
		MaxOpenConns int
		MaxIdleConns int
	}
	Redis struct {
		Addrs    []string
		Username string
		Password string
		DB       int
	}
	Kafka struct {
		Hosts            string
		SecurityProtocol string
		SASLMechanisms   string
		SASLUsername     string
		SASLPassword     string
		SessionTimeout   int
	}
	GCP struct {
		ProjectID      string
		ServiceAccount []byte
	}
	Order struct {
		Expiration              time.Duration
		TaxChargePercentage     float64
		ServiceChargePercentage float64
	}
	Midtrans struct {
		BaseURL      string
		BasicAuthKey string
	}
}

func (cfg *Config) application() {
	cfg.Application.Name = os.Getenv("APP_NAME")
	cfg.Application.Port, _ = strconv.Atoi(os.Getenv("APP_PORT"))
	cfg.Application.Environment = os.Getenv("APP_ENVIRONMENT")
	cfg.Application.Debug, _ = strconv.ParseBool(os.Getenv("APP_DEBUG"))

	timeoutInSec, _ := strconv.Atoi(os.Getenv("APP_TIMEOUT"))
	cfg.Application.Timeout = time.Duration(timeoutInSec) * time.Second

	cfg.Application.Timezone, _ = time.LoadLocation(os.Getenv("APP_TIMEZONE"))

	cfg.Application.TMUser.BaseURL = os.Getenv("APP_TMUSER_BASE_URL")
	cfg.Application.TMOrder.BaseURL = os.Getenv("APP_TMORDER_BASE_URL")
}

func (cfg *Config) midtrans() {
	cfg.Midtrans.BaseURL = os.Getenv("MIDTRANS_BASE_URL")
	cfg.Midtrans.BasicAuthKey = os.Getenv("MIDTRANS_BASIC_AUTH_KEY")
}

func (cfg *Config) order() {
	expiration, _ := strconv.Atoi(os.Getenv("ORDER_EXPIRATION"))
	cfg.Order.Expiration = time.Duration(expiration) * time.Minute

	cfg.Order.TaxChargePercentage, _ = strconv.ParseFloat(os.Getenv("ORDER_TAX_CHARGE"), 64)
	cfg.Order.ServiceChargePercentage, _ = strconv.ParseFloat(os.Getenv("ORDER_SERVICE_CHARGE"), 64)
}

func (cfg *Config) crypto() {
	cfg.Crypto.Secret = os.Getenv("CRYPTO_SECRET")
}

func (cfg *Config) openTelemetry() {
	collectorEndpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")
	cfg.OpenTelemetry.Collector.Endpoint = collectorEndpoint
}

func (cfg *Config) cors() {
	cfg.CORS.AllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	cfg.CORS.AllowedMethods = strings.Split(os.Getenv("CORS_ALLOWED_METHODS"), ",")
	cfg.CORS.AllowedHeaders = strings.Split(os.Getenv("CORS_ALLOWED_HEADERS"), ",")
	cfg.CORS.ExposedHeaders = strings.Split(os.Getenv("CORS_EXPOSED_HEADERS"), ",")
	cfg.CORS.AllowCredentials, _ = strconv.ParseBool(os.Getenv("CORS_ALLOW_CREDENTIALS"))
	cfg.CORS.MaxAge, _ = strconv.Atoi(os.Getenv("CORS_MAX_AGE"))
}

func (c *Config) jwt() {
	jwtRsaPlain := os.Getenv("JWT_RSA")
	var jwtRsa = struct {
		PrivateKey string `json:"private"`
		PublicKey  string `json:"public"`
	}{}

	json.Unmarshal([]byte(jwtRsaPlain), &jwtRsa)

	c.JWT.PrivateKey = []byte(jwtRsa.PrivateKey)
	c.JWT.PublicKey = []byte(jwtRsa.PublicKey)
}

func (c *Config) postgresql() {
	c.Postgresql.Host = os.Getenv("POSTGRESQL_HOST")
	c.Postgresql.Port, _ = strconv.Atoi(os.Getenv("POSTGRESQL_PORT"))
	c.Postgresql.User = os.Getenv("POSTGRESQL_USER")
	c.Postgresql.Password = os.Getenv("POSTGRESQL_PASSWORD")
	c.Postgresql.DBName = os.Getenv("POSTGRESQL_DBNAME")
	c.Postgresql.SSLMode = os.Getenv("POSTGRESQL_SSLMODE")
	c.Postgresql.MaxOpenConns, _ = strconv.Atoi(os.Getenv("POSTGRESQL_MAX_OPEN_CONNS"))
	c.Postgresql.MaxIdleConns, _ = strconv.Atoi(os.Getenv("POSTGRESQL_MAX_IDLE_CONNS"))
}

func (cfg *Config) redis() {
	cfg.Redis.Addrs = strings.Split(os.Getenv("REDIS_HOSTS"), ",")
	cfg.Redis.Username = os.Getenv("REDIS_USERNAME")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DB, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
}

func (cfg *Config) kafka() {
	cfg.Kafka.Hosts = os.Getenv("KAFKA_HOSTS")
	cfg.Kafka.SecurityProtocol = os.Getenv("KAFKA_SECURITY_PROTOCOL")
	cfg.Kafka.SASLMechanisms = os.Getenv("KAFKA_SASL_MECHANISMS")
	cfg.Kafka.SASLUsername = os.Getenv("KAFKA_SASL_USERNAME")
	cfg.Kafka.SASLPassword = os.Getenv("KAFKA_SASL_PASSWORD")
	cfg.Kafka.SessionTimeout, _ = strconv.Atoi(os.Getenv("KAFKA_SESSION_TIMEOUT_MS"))
}

func (cfg *Config) gcp() {
	cfg.GCP.ServiceAccount = []byte(os.Getenv("GCP_SERVICE_ACCOUNT"))
	cfg.GCP.ProjectID = os.Getenv("GCP_PROJECT_ID")
}

func load() *Config {
	cfg := new(Config)
	cfg.application()
	cfg.order()
	cfg.crypto()
	cfg.openTelemetry()
	cfg.jwt()
	cfg.postgresql()
	cfg.cors()
	cfg.redis()
	cfg.kafka()
	cfg.gcp()
	cfg.midtrans()
	return cfg
}

func Get() *Config {
	configSyncOnce.Do(func() {
		c = load()
	})

	return c
}
