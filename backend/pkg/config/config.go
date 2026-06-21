package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	MQ         MQConfig         `mapstructure:"mq"`
	Storage    StorageConfig    `mapstructure:"storage"`
	AI         AIConfig         `mapstructure:"ai"`
	Map        MapConfig        `mapstructure:"map"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
	Route      RouteConfig      `mapstructure:"route"`
	Traffic    TrafficConfig    `mapstructure:"traffic"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
	Name     string `mapstructure:"name"`
	Mode     string `mapstructure:"mode"`
	APIPort  int    `mapstructure:"api_port"`
	GRPCPort int    `mapstructure:"grpc_port"`
}

type DatabaseConfig struct {
	TiDB          TiDBConfig          `mapstructure:"tidb"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

type TiDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
	Timeout  string `mapstructure:"timeout"`
}

func (t TiDBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		t.User, t.Password, t.Host, t.Port, t.DBName, t.Timeout)
}

type RedisConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Password  string   `mapstructure:"password"`
	PoolSize  int      `mapstructure:"pool_size"`
}

type ElasticsearchConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

type MQConfig struct {
	RocketMQ RocketMQConfig `mapstructure:"rocketmq"`
}

type RocketMQConfig struct {
	NameServer    string            `mapstructure:"name_server"`
	ProducerGroup string            `mapstructure:"producer_group"`
	ConsumerGroup string            `mapstructure:"consumer_group"`
	Topics        map[string]string `mapstructure:"topics"`
}

type StorageConfig struct {
	MinIO MinIOConfig `mapstructure:"minio"`
}

type MinIOConfig struct {
	Endpoint  string            `mapstructure:"endpoint"`
	AccessKey string            `mapstructure:"access_key"`
	SecretKey string            `mapstructure:"secret_key"`
	UseSSL    bool              `mapstructure:"use_ssl"`
	Buckets   map[string]string `mapstructure:"buckets"`
}

type AIConfig struct {
	DeepSeek  DeepSeekConfig  `mapstructure:"deepseek"`
	MediaPipe MediaPipeConfig `mapstructure:"mediapipe"`
	Fatigue   FatigueConfig   `mapstructure:"fatigue"`
}

type DeepSeekConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Timeout string `mapstructure:"timeout"`
}

type MediaPipeConfig struct {
	ModelPath string `mapstructure:"model_path"`
}

type FatigueConfig struct {
	PERCLOSThreshold         float64 `mapstructure:"perclos_threshold"`
	YawnThreshold            float64 `mapstructure:"yawn_threshold"`
	HeadPitchThreshold       float64 `mapstructure:"head_pitch_threshold"`
	FatigueScoreThreshold    float64 `mapstructure:"fatigue_score_threshold"`
	ContinuousFatigueMinutes int     `mapstructure:"continuous_fatigue_minutes"`
}

type MapConfig struct {
	AMap  AMapConfig  `mapstructure:"amap"`
	Baidu BaiduConfig `mapstructure:"baidu"`
}

type AMapConfig struct {
	Key      string `mapstructure:"key"`
	TruckKey string `mapstructure:"truck_key"`
	WSURL    string `mapstructure:"ws_url"`
}

type BaiduConfig struct {
	AK string `mapstructure:"ak"`
	SK string `mapstructure:"sk"`
}

type BlockchainConfig struct {
	Fabric FabricConfig `mapstructure:"fabric"`
}

type FabricConfig struct {
	ConfigPath    string `mapstructure:"config_path"`
	ChannelID     string `mapstructure:"channel_id"`
	ChaincodeName string `mapstructure:"chaincode_name"`
}

type RouteConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	DefaultStrategy string `mapstructure:"default_strategy"`
	CacheTTL       int    `mapstructure:"cache_ttl"`
}

type TrafficConfig struct {
	WebhookToken       string `mapstructure:"webhook_token"`
	ScannerIntervalSec int    `mapstructure:"scanner_interval_sec"`
	TriggerWindowMin   int    `mapstructure:"trigger_window_min"`
}

type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	ExpireHours  int    `mapstructure:"expire_hours"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

var Global *Config

func GlobalConfig() *Config {
	if Global == nil {
		return &Config{}
	}
	return Global
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("DDG")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	Global = cfg
	return cfg, nil
}
