package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Storage  StorageConfig  `mapstructure:"storage"`
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
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s",
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

var GlobalConfig *Config

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
	return nil
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
}

func Get() *Config {
	return GlobalConfig
}
