package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	once      sync.Once
	TIDB      *gorm.DB
	Redis     *redis.ClusterClient
	ES        *elastic.Client
)

func Init(cfg *config.DatabaseConfig) error {
	var initErr error
	once.Do(func() {
		if err := initTiDB(cfg.TiDB); err != nil {
			initErr = fmt.Errorf("init tidb: %w", err)
			return
		}
		if err := initRedis(cfg.Redis); err != nil {
			initErr = fmt.Errorf("init redis: %w", err)
			return
		}
		if err := initES(cfg.Elasticsearch); err != nil {
			initErr = fmt.Errorf("init es: %w", err)
			return
		}
	})
	return initErr
}

func initTiDB(cfg config.TiDBConfig) error {
	gormLogger := logger.Default.LogMode(logger.Warn)
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                                   gormLogger,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	TIDB = db
	logger.Sugar.Infof("TiDB connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.DBName)
	return nil
}

func initRedis(cfg config.RedisConfig) error {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Addresses,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 20,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	Redis = rdb
	logger.Sugar.Infof("Redis Cluster connected: %v", cfg.Addresses)
	return nil
}

func initES(cfg config.ElasticsearchConfig) error {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Addresses...),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10 * time.Second),
		elastic.SetGzip(true),
	}
	if cfg.Username != "" {
		opts = append(opts, elastic.SetBasicAuth(cfg.Username, cfg.Password))
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	info, code, err := client.Ping(cfg.Addresses[0]).Do(ctx)
	if err != nil {
		return err
	}
	logger.Sugar.Infof("Elasticsearch connected: version=%s, code=%d", info.Version.Number, code)
	ES = client
	return nil
}

func GetES() *elastic.Client {
	return ES
}

func GetRedis() *redis.ClusterClient {
	return Redis
}

func GetDB() *gorm.DB {
	return TIDB
}
