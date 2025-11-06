package utils

import (
	"os"
	"strings"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	SMTP      SMTPConfig      `mapstructure:"smtp"`
	AI        AIConfig        `mapstructure:"ai"`
	Email     EmailConfig     `mapstructure:"email"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Log       LogConfig       `mapstructure:"log"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Webhook   WebhookConfig   `mapstructure:"webhook"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	FromName string `mapstructure:"from_name"`
}

type AIConfig struct {
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	CustomAPI CustomAPIConfig `mapstructure:"custom_api"`
}

type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

type CustomAPIConfig struct {
	URL     string            `mapstructure:"url"`
	Headers map[string]string `mapstructure:"headers"`
}

type EmailConfig struct {
	BatchSize         int `mapstructure:"batch_size"`
	SendInterval      int `mapstructure:"send_interval"`
	RetryTimes        int `mapstructure:"retry_times"`
	TemplateSizeLimit int `mapstructure:"template_size_limit"`
}

type SchedulerConfig struct {
	CheckInterval int `mapstructure:"check_interval"`
	MaxWorkers    int `mapstructure:"max_workers"`
}

type LogConfig struct {
	Level         string `mapstructure:"level"`
	RetentionDays int    `mapstructure:"retention_days"`
	FilePath      string `mapstructure:"file_path"`
}

type AuthConfig struct {
	Token string `mapstructure:"token"`
}

type WebhookConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	
	// 支持环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// 设置环境变量前缀
	viper.SetEnvPrefix("GME")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	
	// 环境变量覆盖敏感配置
	if token := os.Getenv("GME_AUTH_TOKEN"); token != "" {
		config.Auth.Token = token
	}
	if dbPassword := os.Getenv("GME_DATABASE_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if redisPassword := os.Getenv("GME_REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if smtpPassword := os.Getenv("GME_SMTP_PASSWORD"); smtpPassword != "" {
		config.SMTP.Password = smtpPassword
	}
	if apiKey := os.Getenv("GME_AI_OPENAI_API_KEY"); apiKey != "" {
		config.AI.OpenAI.APIKey = apiKey
	}
	
	return &config, nil
}