package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Log           LogConfig           `mapstructure:"log"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Storage       StorageConfig       `mapstructure:"storage"`
	AIProcurement AIProcurementConfig `mapstructure:"ai_procurement"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s&tls=false",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	Expire        int    `mapstructure:"expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
	Issuer        string `mapstructure:"issuer"`
}

func (c *JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(c.Expire) * time.Second
}

func (c *JWTConfig) RefreshExpireDuration() time.Duration {
	return time.Duration(c.RefreshExpire) * time.Second
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
}

type StorageConfig struct {
	Provider       string `mapstructure:"provider"`
	Region         string `mapstructure:"region"`
	SecretID       string `mapstructure:"secret_id"`
	SecretKey      string `mapstructure:"secret_key"`
	PublicBucket   string `mapstructure:"public_bucket"`
	PrivateBucket  string `mapstructure:"private_bucket"`
	AdminBucket    string `mapstructure:"admin_bucket"`
	PublicBaseURL  string `mapstructure:"public_base_url"`
	PrivateBaseURL string `mapstructure:"private_base_url"`
	AdminBaseURL   string `mapstructure:"admin_base_url"`
	MaxFileSizeMB  int64  `mapstructure:"max_file_size_mb"`
}

func (c StorageConfig) MaxFileSizeBytes() int64 {
	if c.MaxFileSizeMB <= 0 {
		return 20 << 20
	}
	return c.MaxFileSizeMB << 20
}

type AIProcurementConfig struct {
	ModelParsingEnabled bool               `mapstructure:"model_parsing_enabled"`
	Provider            string             `mapstructure:"provider"`
	BaseURL             string             `mapstructure:"base_url"`
	APIKeyEnv           string             `mapstructure:"api_key_env"`
	Model               string             `mapstructure:"model"`
	Temperature         float64            `mapstructure:"temperature"`
	TopP                float64            `mapstructure:"top_p"`
	MaxTokens           int                `mapstructure:"max_tokens"`
	TimeoutSec          int                `mapstructure:"timeout_sec"`
	ConfidenceThreshold float64            `mapstructure:"confidence_threshold"`
	VectorSearch        VectorSearchConfig `mapstructure:"vector_search"`
	RuntimeConfig       RuntimeConfig      `mapstructure:"runtime_config"`
}

func (c AIProcurementConfig) APIKey() string {
	if c.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(c.APIKeyEnv)
}

type VectorSearchConfig struct {
	Enabled                  bool    `mapstructure:"enabled"`
	Provider                 string  `mapstructure:"provider"`
	Endpoint                 string  `mapstructure:"endpoint"`
	Database                 string  `mapstructure:"database"`
	Username                 string  `mapstructure:"username"`
	PasswordEnv              string  `mapstructure:"password_env"`
	ConnectionEnv            string  `mapstructure:"connection_env"`
	ProductCollection        string  `mapstructure:"product_collection"`
	SupplierCollection       string  `mapstructure:"supplier_collection"`
	FinanceProductCollection string  `mapstructure:"finance_product_collection"`
	EmbeddingProvider        string  `mapstructure:"embedding_provider"`
	EmbeddingBaseURL         string  `mapstructure:"embedding_base_url"`
	EmbeddingAPIKeyEnv       string  `mapstructure:"embedding_api_key_env"`
	EmbeddingModel           string  `mapstructure:"embedding_model"`
	Dimension                int     `mapstructure:"dimension"`
	TopK                     int     `mapstructure:"top_k"`
	MinScore                 float64 `mapstructure:"min_score"`
	HybridKeywordWeight      float64 `mapstructure:"hybrid_keyword_weight"`
	HybridVectorWeight       float64 `mapstructure:"hybrid_vector_weight"`
}

type RuntimeConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	RedisKey    string `mapstructure:"redis_key"`
	CacheTTLSec int    `mapstructure:"cache_ttl_sec"`
}

type aiProcurementAdminConfig struct {
	Enabled bool `json:"enabled"`
	Model   struct {
		ParsingEnabled      bool    `json:"parsingEnabled"`
		Provider            string  `json:"provider"`
		BaseURL             string  `json:"baseUrl"`
		APIKeyEnv           string  `json:"apiKeyEnv"`
		Model               string  `json:"model"`
		Temperature         float64 `json:"temperature"`
		TopP                float64 `json:"topP"`
		MaxTokens           int     `json:"maxTokens"`
		TimeoutSec          int     `json:"timeoutSec"`
		ConfidenceThreshold float64 `json:"confidenceThreshold"`
	} `json:"model"`
	Matching struct {
		VectorEnabled            bool    `json:"vectorEnabled"`
		VectorProvider           string  `json:"vectorProvider"`
		VectorEndpoint           string  `json:"vectorEndpoint"`
		VectorDatabase           string  `json:"vectorDatabase"`
		VectorUsername           string  `json:"vectorUsername"`
		VectorPasswordEnv        string  `json:"vectorPasswordEnv"`
		VectorConnectionEnv      string  `json:"vectorConnectionEnv"`
		ProductCollection        string  `json:"productCollection"`
		SupplierCollection       string  `json:"supplierCollection"`
		FinanceProductCollection string  `json:"financeProductCollection"`
		EmbeddingProvider        string  `json:"embeddingProvider"`
		EmbeddingBaseURL         string  `json:"embeddingBaseUrl"`
		EmbeddingAPIKeyEnv       string  `json:"embeddingApiKeyEnv"`
		EmbeddingModel           string  `json:"embeddingModel"`
		EmbeddingDimension       int     `json:"embeddingDimension"`
		VectorTopK               int     `json:"vectorTopK"`
		VectorMinScore           float64 `json:"vectorMinScore"`
		HybridKeywordWeight      float64 `json:"hybridKeywordWeight"`
		HybridVectorWeight       float64 `json:"hybridVectorWeight"`
	} `json:"matching"`
}

func (c VectorSearchConfig) Password() string {
	if c.PasswordEnv == "" {
		return ""
	}
	return os.Getenv(c.PasswordEnv)
}

func (c VectorSearchConfig) ConnectionString() string {
	if c.ConnectionEnv != "" {
		if value := os.Getenv(c.ConnectionEnv); value != "" {
			return value
		}
	}
	if strings.HasPrefix(c.Endpoint, "postgres://") || strings.HasPrefix(c.Endpoint, "postgresql://") {
		return c.Endpoint
	}
	return ""
}

func (c VectorSearchConfig) EmbeddingAPIKey() string {
	if c.EmbeddingAPIKeyEnv == "" {
		return ""
	}
	return os.Getenv(c.EmbeddingAPIKeyEnv)
}

var GlobalConfig *Config

var aiProcurementRuntimeCache struct {
	mu            sync.Mutex
	lastAttemptAt time.Time
}

var aiProcurementConfigMu sync.RWMutex

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Support environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyEnvOverrides(GlobalConfig)
	applyAIProcurementDefaults(&GlobalConfig.AIProcurement)

	if GlobalConfig.JWT.Secret == "" || GlobalConfig.JWT.Secret == "change-me" {
		env := os.Getenv("APP_ENV")
		if env == "production" || env == "prod" {
			return fmt.Errorf("JWT_SECRET must be set in production environment")
		}
		log.Println("WARNING: JWT secret is using insecure default. Set JWT_SECRET env var for production!")
	}

	return nil
}

func applyAIProcurementDefaults(c *AIProcurementConfig) {
	if c.Provider == "" {
		c.Provider = "rule_based"
	}
	if c.BaseURL == "" {
		c.BaseURL = "https://api.minimax.io/v1/chat/completions"
	}
	if c.APIKeyEnv == "" {
		c.APIKeyEnv = "MINIMAX_API_KEY"
	}
	if c.Model == "" {
		c.Model = "MiniMax-M2.7"
	}
	if c.Temperature <= 0 || c.Temperature > 1 {
		c.Temperature = 0.2
	}
	if c.TopP <= 0 || c.TopP > 1 {
		c.TopP = 0.9
	}
	if c.MaxTokens <= 0 {
		c.MaxTokens = 1200
	}
	if c.TimeoutSec <= 0 {
		c.TimeoutSec = 20
	}
	if c.ConfidenceThreshold <= 0 || c.ConfidenceThreshold > 1 {
		c.ConfidenceThreshold = 0.78
	}
	applyVectorSearchDefaults(&c.VectorSearch)
	applyRuntimeConfigDefaults(&c.RuntimeConfig)
}

func applyRuntimeConfigDefaults(c *RuntimeConfig) {
	if c.RedisKey == "" {
		c.RedisKey = "ai_procurement_config"
	}
	if c.CacheTTLSec <= 0 {
		c.CacheTTLSec = 30
	}
}

func applyVectorSearchDefaults(c *VectorSearchConfig) {
	if c.Provider == "" {
		c.Provider = "pgvector"
	}
	if c.Database == "" {
		c.Database = "steellink_vectors"
	}
	if c.ConnectionEnv == "" {
		c.ConnectionEnv = "PGVECTOR_DSN"
	}
	if c.PasswordEnv == "" {
		c.PasswordEnv = "PGVECTOR_PASSWORD"
	}
	if c.ProductCollection == "" {
		c.ProductCollection = "ai_product_embeddings"
	}
	if c.SupplierCollection == "" {
		c.SupplierCollection = "ai_supplier_embeddings"
	}
	if c.FinanceProductCollection == "" {
		c.FinanceProductCollection = "ai_finance_product_embeddings"
	}
	if c.EmbeddingProvider == "" {
		c.EmbeddingProvider = "minimax"
	}
	if c.EmbeddingBaseURL == "" {
		c.EmbeddingBaseURL = "https://api.minimax.chat/v1/embeddings"
	}
	if c.EmbeddingAPIKeyEnv == "" {
		c.EmbeddingAPIKeyEnv = "MINIMAX_API_KEY"
	}
	if c.EmbeddingModel == "" {
		c.EmbeddingModel = "embo-01"
	}
	if c.Dimension <= 0 {
		c.Dimension = 1536
	}
	if c.TopK <= 0 {
		c.TopK = 10
	}
	if c.MinScore <= 0 || c.MinScore > 1 {
		c.MinScore = 0.65
	}
	if c.HybridKeywordWeight <= 0 {
		c.HybridKeywordWeight = 0.35
	}
	if c.HybridVectorWeight <= 0 {
		c.HybridVectorWeight = 0.65
	}
}

// applyEnvOverrides reads os environment (including values from a loaded .env file)
// so secrets and per-environment hosts are not committed in config.yaml.
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Server.Port = p
		}
	}
	if v := os.Getenv("SERVER_MODE"); v != "" {
		c.Server.Mode = v
	}

	if v := os.Getenv("DATABASE_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Database.Port = p
		}
	}
	if v := os.Getenv("DATABASE_USERNAME"); v != "" {
		c.Database.Username = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		c.Database.Database = v
	}
	if v := os.Getenv("DATABASE_CHARSET"); v != "" {
		c.Database.Charset = v
	}
	if v := os.Getenv("DATABASE_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxIdleConns = n
		}
	}
	if v := os.Getenv("DATABASE_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxOpenConns = n
		}
	}
	if v := os.Getenv("DATABASE_CONN_MAX_LIFETIME"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.ConnMaxLifetime = n
		}
	}

	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = p
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Redis.DB = n
		}
	}
	if v := os.Getenv("REDIS_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Redis.PoolSize = n
		}
	}

	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}
	if v := os.Getenv("JWT_EXPIRE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.JWT.Expire = n
		}
	}
	if v := os.Getenv("JWT_REFRESH_EXPIRE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.JWT.RefreshExpire = n
		}
	}
	if v := os.Getenv("JWT_ISSUER"); v != "" {
		c.JWT.Issuer = v
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.Log.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		c.Log.Format = v
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		c.Log.Output = v
	}

	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			c.Kafka.Brokers = out
		}
	}
	if v := os.Getenv("KAFKA_GROUP_ID"); v != "" {
		c.Kafka.GroupID = v
	}

	if v := os.Getenv("STORAGE_PROVIDER"); v != "" {
		c.Storage.Provider = v
	}
	if v := os.Getenv("COS_REGION"); v != "" {
		c.Storage.Region = v
	}
	if v := os.Getenv("COS_SECRET_ID"); v != "" {
		c.Storage.SecretID = v
	}
	if v := os.Getenv("COS_SECRET_KEY"); v != "" {
		c.Storage.SecretKey = v
	}
	if v := os.Getenv("COS_PUBLIC_BUCKET"); v != "" {
		c.Storage.PublicBucket = v
	}
	if v := os.Getenv("COS_PRIVATE_BUCKET"); v != "" {
		c.Storage.PrivateBucket = v
	}
	if v := os.Getenv("COS_ADMIN_BUCKET"); v != "" {
		c.Storage.AdminBucket = v
	}
	if v := os.Getenv("COS_PUBLIC_BASE_URL"); v != "" {
		c.Storage.PublicBaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("COS_PRIVATE_BASE_URL"); v != "" {
		c.Storage.PrivateBaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("COS_ADMIN_BASE_URL"); v != "" {
		c.Storage.AdminBaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("STORAGE_MAX_FILE_SIZE_MB"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.Storage.MaxFileSizeMB = n
		}
	}

	if v := os.Getenv("AI_PROCUREMENT_MODEL_PARSING_ENABLED"); v != "" {
		c.AIProcurement.ModelParsingEnabled = parseBool(v)
	}
	if v := os.Getenv("AI_PROCUREMENT_PROVIDER"); v != "" {
		c.AIProcurement.Provider = v
	}
	if v := os.Getenv("AI_PROCUREMENT_BASE_URL"); v != "" {
		c.AIProcurement.BaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("AI_PROCUREMENT_API_KEY_ENV"); v != "" {
		c.AIProcurement.APIKeyEnv = v
	}
	if v := os.Getenv("AI_PROCUREMENT_MODEL"); v != "" {
		c.AIProcurement.Model = v
	}
	if v := os.Getenv("AI_PROCUREMENT_TEMPERATURE"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.Temperature = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_TOP_P"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.TopP = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.AIProcurement.MaxTokens = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.AIProcurement.TimeoutSec = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_CONFIDENCE_THRESHOLD"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.ConfidenceThreshold = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_VECTOR_SEARCH_ENABLED"); v != "" {
		c.AIProcurement.VectorSearch.Enabled = parseBool(v)
	}
	if v := os.Getenv("AI_PROCUREMENT_VECTOR_PROVIDER"); v != "" {
		c.AIProcurement.VectorSearch.Provider = v
	}
	if v := os.Getenv("PGVECTOR_ENDPOINT"); v != "" {
		c.AIProcurement.VectorSearch.Endpoint = v
	}
	if v := os.Getenv("PGVECTOR_DATABASE"); v != "" {
		c.AIProcurement.VectorSearch.Database = v
	}
	if v := os.Getenv("PGVECTOR_USERNAME"); v != "" {
		c.AIProcurement.VectorSearch.Username = v
	}
	if v := os.Getenv("PGVECTOR_PASSWORD_ENV"); v != "" {
		c.AIProcurement.VectorSearch.PasswordEnv = v
	}
	if v := os.Getenv("PGVECTOR_DSN_ENV"); v != "" {
		c.AIProcurement.VectorSearch.ConnectionEnv = v
	}
	if v := os.Getenv("AI_PRODUCT_VECTOR_COLLECTION"); v != "" {
		c.AIProcurement.VectorSearch.ProductCollection = v
	}
	if v := os.Getenv("AI_SUPPLIER_VECTOR_COLLECTION"); v != "" {
		c.AIProcurement.VectorSearch.SupplierCollection = v
	}
	if v := os.Getenv("AI_FINANCE_PRODUCT_VECTOR_COLLECTION"); v != "" {
		c.AIProcurement.VectorSearch.FinanceProductCollection = v
	}
	if v := os.Getenv("AI_EMBEDDING_PROVIDER"); v != "" {
		c.AIProcurement.VectorSearch.EmbeddingProvider = v
	}
	if v := os.Getenv("AI_EMBEDDING_BASE_URL"); v != "" {
		c.AIProcurement.VectorSearch.EmbeddingBaseURL = v
	}
	if v := os.Getenv("AI_EMBEDDING_API_KEY_ENV"); v != "" {
		c.AIProcurement.VectorSearch.EmbeddingAPIKeyEnv = v
	}
	if v := os.Getenv("AI_EMBEDDING_MODEL"); v != "" {
		c.AIProcurement.VectorSearch.EmbeddingModel = v
	}
	if v := os.Getenv("AI_EMBEDDING_DIMENSION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.AIProcurement.VectorSearch.Dimension = n
		}
	}
	if v := os.Getenv("AI_VECTOR_TOP_K"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.AIProcurement.VectorSearch.TopK = n
		}
	}
	if v := os.Getenv("AI_VECTOR_MIN_SCORE"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.VectorSearch.MinScore = n
		}
	}
	if v := os.Getenv("AI_HYBRID_KEYWORD_WEIGHT"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.VectorSearch.HybridKeywordWeight = n
		}
	}
	if v := os.Getenv("AI_HYBRID_VECTOR_WEIGHT"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			c.AIProcurement.VectorSearch.HybridVectorWeight = n
		}
	}
	if v := os.Getenv("AI_PROCUREMENT_RUNTIME_CONFIG_ENABLED"); v != "" {
		c.AIProcurement.RuntimeConfig.Enabled = parseBool(v)
	}
	if v := os.Getenv("AI_PROCUREMENT_RUNTIME_CONFIG_REDIS_KEY"); v != "" {
		c.AIProcurement.RuntimeConfig.RedisKey = v
	}
	if v := os.Getenv("AI_PROCUREMENT_RUNTIME_CONFIG_CACHE_TTL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.AIProcurement.RuntimeConfig.CacheTTLSec = n
		}
	}
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	default:
		return false
	}
}

func Get() *Config {
	return GlobalConfig
}

func RefreshAIProcurementRuntimeConfig(ctx context.Context) {
	if GlobalConfig == nil {
		return
	}
	runtimeCfg := GlobalConfig.AIProcurement.RuntimeConfig
	if runtimeCfg.RedisKey == "" {
		runtimeCfg.RedisKey = "ai_procurement_config"
	}
	ttl := time.Duration(runtimeCfg.CacheTTLSec) * time.Second
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	aiProcurementRuntimeCache.mu.Lock()
	if time.Since(aiProcurementRuntimeCache.lastAttemptAt) < ttl {
		aiProcurementRuntimeCache.mu.Unlock()
		return
	}
	aiProcurementRuntimeCache.lastAttemptAt = time.Now()
	aiProcurementRuntimeCache.mu.Unlock()

	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     GlobalConfig.Redis.Addr(),
		Password: GlobalConfig.Redis.Password,
		DB:       GlobalConfig.Redis.DB,
		PoolSize: GlobalConfig.Redis.PoolSize,
	})
	defer client.Close()

	raw, err := client.Get(requestCtx, runtimeCfg.RedisKey).Result()
	aiProcurementConfigMu.Lock()
	defer aiProcurementConfigMu.Unlock()
	if err != nil {
		applyHardcodedAIProcurementFallback(&GlobalConfig.AIProcurement)
		return
	}
	var payload aiProcurementAdminConfig
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		applyHardcodedAIProcurementFallback(&GlobalConfig.AIProcurement)
		return
	}
	applyAdminAIProcurementConfig(&GlobalConfig.AIProcurement, payload)
}

func GetAIProcurement() AIProcurementConfig {
	aiProcurementConfigMu.RLock()
	defer aiProcurementConfigMu.RUnlock()
	if GlobalConfig == nil {
		return AIProcurementConfig{}
	}
	return GlobalConfig.AIProcurement
}

func applyHardcodedAIProcurementFallback(target *AIProcurementConfig) {
	runtimeCfg := target.RuntimeConfig
	*target = hardcodedAIProcurementFallback()
	target.RuntimeConfig = runtimeCfg
}

func hardcodedAIProcurementFallback() AIProcurementConfig {
	return AIProcurementConfig{
		ModelParsingEnabled: false,
		Provider:            "rule_based",
		BaseURL:             "https://api.minimax.io/v1/chat/completions",
		APIKeyEnv:           "MINIMAX_API_KEY",
		Model:               "MiniMax-M2.7",
		Temperature:         0.2,
		TopP:                0.9,
		MaxTokens:           1200,
		TimeoutSec:          20,
		ConfidenceThreshold: 0.78,
		VectorSearch: VectorSearchConfig{
			Enabled:                  false,
			Provider:                 "pgvector",
			Database:                 "steellink_vectors",
			PasswordEnv:              "PGVECTOR_PASSWORD",
			ConnectionEnv:            "PGVECTOR_DSN",
			ProductCollection:        "ai_product_embeddings",
			SupplierCollection:       "ai_supplier_embeddings",
			FinanceProductCollection: "ai_finance_product_embeddings",
			EmbeddingProvider:        "minimax",
			EmbeddingBaseURL:         "https://api.minimax.chat/v1/embeddings",
			EmbeddingAPIKeyEnv:       "MINIMAX_API_KEY",
			EmbeddingModel:           "embo-01",
			Dimension:                1536,
			TopK:                     10,
			MinScore:                 0.65,
			HybridKeywordWeight:      0.35,
			HybridVectorWeight:       0.65,
		},
		RuntimeConfig: RuntimeConfig{
			Enabled:     true,
			RedisKey:    "ai_procurement_config",
			CacheTTLSec: 30,
		},
	}
}

func applyAdminAIProcurementConfig(target *AIProcurementConfig, source aiProcurementAdminConfig) {
	applyHardcodedAIProcurementFallback(target)
	target.ModelParsingEnabled = source.Enabled && source.Model.ParsingEnabled
	if source.Model.Provider != "" {
		target.Provider = source.Model.Provider
	}
	if source.Model.BaseURL != "" {
		target.BaseURL = strings.TrimRight(source.Model.BaseURL, "/")
	}
	if source.Model.APIKeyEnv != "" {
		target.APIKeyEnv = source.Model.APIKeyEnv
	}
	if source.Model.Model != "" {
		target.Model = source.Model.Model
	}
	if source.Model.Temperature > 0 && source.Model.Temperature <= 1 {
		target.Temperature = source.Model.Temperature
	}
	if source.Model.TopP > 0 && source.Model.TopP <= 1 {
		target.TopP = source.Model.TopP
	}
	if source.Model.MaxTokens > 0 {
		target.MaxTokens = source.Model.MaxTokens
	}
	if source.Model.TimeoutSec > 0 {
		target.TimeoutSec = source.Model.TimeoutSec
	}
	if source.Model.ConfidenceThreshold > 0 {
		target.ConfidenceThreshold = source.Model.ConfidenceThreshold
	}

	target.VectorSearch.Enabled = source.Enabled && source.Matching.VectorEnabled
	if source.Matching.VectorProvider != "" {
		target.VectorSearch.Provider = source.Matching.VectorProvider
	}
	if source.Matching.VectorEndpoint != "" {
		target.VectorSearch.Endpoint = source.Matching.VectorEndpoint
	}
	if source.Matching.VectorDatabase != "" {
		target.VectorSearch.Database = source.Matching.VectorDatabase
	}
	if source.Matching.VectorUsername != "" {
		target.VectorSearch.Username = source.Matching.VectorUsername
	}
	if source.Matching.VectorPasswordEnv != "" {
		target.VectorSearch.PasswordEnv = source.Matching.VectorPasswordEnv
	}
	if source.Matching.VectorConnectionEnv != "" {
		target.VectorSearch.ConnectionEnv = source.Matching.VectorConnectionEnv
	}
	if source.Matching.ProductCollection != "" {
		target.VectorSearch.ProductCollection = source.Matching.ProductCollection
	}
	if source.Matching.SupplierCollection != "" {
		target.VectorSearch.SupplierCollection = source.Matching.SupplierCollection
	}
	if source.Matching.FinanceProductCollection != "" {
		target.VectorSearch.FinanceProductCollection = source.Matching.FinanceProductCollection
	}
	if source.Matching.EmbeddingProvider != "" {
		target.VectorSearch.EmbeddingProvider = source.Matching.EmbeddingProvider
	}
	if source.Matching.EmbeddingBaseURL != "" {
		target.VectorSearch.EmbeddingBaseURL = strings.TrimRight(source.Matching.EmbeddingBaseURL, "/")
	}
	if source.Matching.EmbeddingAPIKeyEnv != "" {
		target.VectorSearch.EmbeddingAPIKeyEnv = source.Matching.EmbeddingAPIKeyEnv
	}
	if source.Matching.EmbeddingModel != "" {
		target.VectorSearch.EmbeddingModel = source.Matching.EmbeddingModel
	}
	if source.Matching.EmbeddingDimension > 0 {
		target.VectorSearch.Dimension = source.Matching.EmbeddingDimension
	}
	if source.Matching.VectorTopK > 0 {
		target.VectorSearch.TopK = source.Matching.VectorTopK
	}
	if source.Matching.VectorMinScore > 0 {
		target.VectorSearch.MinScore = source.Matching.VectorMinScore
	}
	if source.Matching.HybridKeywordWeight > 0 {
		target.VectorSearch.HybridKeywordWeight = source.Matching.HybridKeywordWeight
	}
	if source.Matching.HybridVectorWeight > 0 {
		target.VectorSearch.HybridVectorWeight = source.Matching.HybridVectorWeight
	}
}
