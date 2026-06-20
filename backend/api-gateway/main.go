package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/dangerous-drive-guard/backend/api-gateway/routes"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

func main() {
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})
	defer logger.Global.Sync()

	logger.Sugar.Infof("Starting %s API Gateway...", cfg.Server.Name)

	if err := database.Init(&cfg.Database); err != nil {
		logger.Sugar.Fatalf("init database failed: %v", err)
	}

	if err := mq.Init(&cfg.MQ.RocketMQ); err != nil {
		logger.Sugar.Warnf("init mq failed: %v", err)
	}
	defer mq.Shutdown()

	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", cfg.Server.APIPort)),
		server.WithReadTimeout(30*time.Second),
		server.WithWriteTimeout(30*time.Second),
		server.WithIdleTimeout(120*time.Second),
		server.WithMaxRequestBodySize(50*1024*1024),
		server.WithTransport(standard.NewTransporter),
		server.WithServerBasicInfo(&app.BasicInfo{
			ServiceName: cfg.Server.Name,
			Version:     "v1.0.0",
		}),
	)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": cfg.Server.Name,
			"time":    time.Now().Unix(),
		})
	})

	routes.Register(h)

	logger.Sugar.Infof("API Gateway started on :%d", cfg.Server.APIPort)
	h.Spin()
}
