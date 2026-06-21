package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"

	"github.com/dangerous-drive-guard/backend/api-gateway/routes"
	escortSvc "github.com/dangerous-drive-guard/backend/internal/escort/service"
	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
	"github.com/dangerous-drive-guard/backend/pkg/response"
	"github.com/dangerous-drive-guard/backend/pkg/storage"
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

	hlog.SetLogger(logger.Global)

	logger.Sugar.Infof("Starting %s API Gateway...", cfg.Server.Name)

	if err := database.Init(&cfg.Database); err != nil {
		logger.Sugar.Fatalf("init database failed: %v", err)
	}

	db := database.GetDB()
	if err := db.AutoMigrate(
		&model.DrivingRestRecord{},
		&model.ServiceAreaRealtimeStatus{},
		&model.ServiceAreaReview{},
		&model.ServiceAreaRecommendation{},
	); err != nil {
		logger.Sugar.Warnf("auto migrate failed: %v", err)
	}

	if _, err := storage.InitMinIO(&cfg.Storage); err != nil {
		logger.Sugar.Fatalf("init minio failed: %v", err)
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
		server.WithMaxRequestBodySize(500*1024*1024),
		server.WithTransport(standard.NewTransporter),
		server.WithServerBasicInfo(&app.BasicInfo{
			ServiceName: cfg.Server.Name,
			Version:     "v1.0.0",
		}),
	)

	h.Use(Recovery())
	h.Use(CORS())

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": cfg.Server.Name,
			"time":    time.Now().Unix(),
		})
	})

	routes.Register(h)

	escortService := escortSvc.NewEscortService()
	escortService.StartCleanupTask(context.Background())

	logger.Sugar.Infof("API Gateway started on :%d", cfg.Server.APIPort)
	h.Spin()
}

func Recovery() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.Global.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
				)
				response.InternalError(ctx, "internal server error")
				ctx.Abort()
			}
		}()
		ctx.Next(c)
	}
}

func CORS() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-ID")
		ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, X-Trace-ID")
		ctx.Header("Access-Control-Max-Age", "86400")
		if string(ctx.Method()) == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next(c)
	}
}
