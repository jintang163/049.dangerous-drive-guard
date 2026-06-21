package routes

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	authHttp "github.com/dangerous-drive-guard/backend/internal/auth/delivery/http"
	authSvc "github.com/dangerous-drive-guard/backend/internal/auth/service"
	blockchainHttp "github.com/dangerous-drive-guard/backend/internal/blockchain/delivery/http"
	blockchainSvc "github.com/dangerous-drive-guard/backend/internal/blockchain/service"
	fatigueHttp "github.com/dangerous-drive-guard/backend/internal/fatigue/delivery/http"
	fatigueSvc "github.com/dangerous-drive-guard/backend/internal/fatigue/service"
	monitorHttp "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/http"
	monitorWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	restrictedHttp "github.com/dangerous-drive-guard/backend/internal/restricted/delivery/http"
	replanHttp "github.com/dangerous-drive-guard/backend/internal/replan/delivery/http"
	routeHandler "github.com/dangerous-drive-guard/backend/internal/route/delivery/http"
	transportHandler "github.com/dangerous-drive-guard/backend/internal/transport/delivery/http"
	userHttp "github.com/dangerous-drive-guard/backend/internal/user/delivery/http"
	userSvc "github.com/dangerous-drive-guard/backend/internal/user/service"
	vehicleHttp "github.com/dangerous-drive-guard/backend/internal/vehicle/delivery/http"
	vehicleSvc "github.com/dangerous-drive-guard/backend/internal/vehicle/service"
	weatherHttp "github.com/dangerous-drive-guard/backend/internal/weather/delivery/http"
	weatherSvc "github.com/dangerous-drive-guard/backend/internal/weather/service"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
)

func Register(h *server.Hertz) {
	authService := authSvc.NewAuthService(config.Global)
	authHandler := authHttp.NewAuthHandler(authService)

	userService := userSvc.NewUserService(config.Global)
	userHandler := userHttp.NewUserHandler(userService)

	vehicleService := vehicleSvc.NewVehicleService()
	vehicleHandler := vehicleHttp.NewVehicleHandler(vehicleService)

	videoService := fatigueSvc.NewVideoService(config.Global)
	videoHandler := fatigueHttp.NewVideoHandler(videoService)

	weatherService := weatherSvc.NewWeatherService(config.Global)
	weatherHandler := weatherHttp.NewWeatherHandler(weatherService)

	blockchainService := blockchainSvc.NewBlockchainService(config.Global)
	blockchainHandler := blockchainHttp.NewBlockchainHandler(blockchainService)

	api := h.Group("/api/v1")
	{
		api.Use(middleware.TraceID())

		authHandler.RegisterRoutes(api)

		userHandler.RegisterRoutes(api, middleware.JWTAuth())

		vehicleHandler.RegisterRoutes(api, middleware.JWTAuth())

		route := api.Group("/routes", middleware.JWTAuth())
		{
			route.POST("/plan", routeHandler.PlanRoute)
			route.POST("/plan/multi", routeHandler.PlanMultiStrategy)
			route.GET("/:id", routeHandler.GetRoute)
			route.GET("", routeHandler.ListRoutes)
			route.POST("/:id/replan", routeHandler.ReplanRoute)
			route.GET("/restricted-areas", routeHandler.ListRestrictedAreas)
		}

		fatigue := api.Group("/fatigue", middleware.JWTAuth())
		{
			fatigue.POST("/detect", fatigueHttp.DetectFatigue)
			fatigue.POST("/upload/frame", fatigueHttp.UploadFrame)
			fatigue.GET("/history/:vehicle_id", fatigueHttp.GetHistory)
			fatigue.GET("/alarms", middleware.RoleAuth("admin", "dispatcher"), fatigueHttp.ListAlarms)
			fatigue.POST("/alarms/:id/ack", middleware.RoleAuth("admin", "dispatcher"), fatigueHttp.AckAlarm)

			videoHandler.RegisterRoutes(fatigue, middleware.JWTAuth())
		}

		monitor := api.Group("/monitor", middleware.JWTAuth())
		{
			monitor.GET("/vehicles/realtime", monitorHttp.GetRealtimeVehicles)
			monitor.GET("/vehicle/:id/status", monitorHttp.GetVehicleStatus)
			monitor.GET("/statistics", monitorHttp.GetStatistics)
			monitor.POST("/intercom/:vehicle_id", middleware.RoleAuth("admin", "dispatcher"), monitorHttp.SendVoiceIntercom)
			monitor.POST("/dispatch/service-area", middleware.RoleAuth("admin", "dispatcher"), monitorHttp.DispatchServiceArea)
		}

		api.GET("/ws/monitor", middleware.JWTAuth(), monitorWs.MonitorWebSocket)
		api.GET("/ws/vehicle/:vehicle_id", middleware.JWTAuth(), monitorWs.VehicleWebSocket)

		transport := api.Group("/transport", middleware.JWTAuth())
		{
			waybill := transport.Group("/waybills")
			waybill.POST("", transportHandler.CreateWaybill)
			waybill.GET("/:id", transportHandler.GetWaybill)
			waybill.GET("", transportHandler.ListWaybills)
			waybill.PUT("/:id/status", transportHandler.UpdateWaybillStatus)
			waybill.POST("/:id/blockchain/save", transportHandler.SaveToBlockchain)
			waybill.GET("/:id/blockchain/verify", transportHandler.VerifyFromBlockchain)

			escort := transport.Group("/escort")
			escort.POST("", transportHandler.StartEscort)
			escort.GET("/:waybill_id", transportHandler.GetEscortInfo)
			escort.POST("/event", transportHandler.ReportEscortEvent)

			transport.GET("/service-areas/recommend", transportHandler.RecommendServiceAreas)
			transport.GET("/weather/warning", transportHandler.GetWeatherWarning)

			rescue := transport.Group("/rescue")
			rescue.POST("/sos", transportHandler.ReportSOS)
			rescue.GET("/resources", transportHandler.ListRescueResources)
			rescue.POST("/dispatch", middleware.RoleAuth("admin", "dispatcher"), transportHandler.DispatchRescue)
		}

		weatherHandler.RegisterRoutes(api, middleware.JWTAuth())

		blockchainHandler.RegisterRoutes(api, middleware.JWTAuth())

		restricted := api.Group("/restricted-areas", middleware.JWTAuth())
		{
			restricted.GET("", restrictedHttp.ListAreas)
			restricted.GET("/:id", restrictedHttp.GetArea)
			restricted.POST("", restrictedHttp.CreateArea)
			restricted.PUT("/:id", restrictedHttp.UpdateArea)
			restricted.DELETE("/:id", middleware.RoleAuth("admin"), restrictedHttp.DeleteArea)

			restricted.POST("/:id/submit", restrictedHttp.SubmitApproval)
			restricted.POST("/:id/approve/first", middleware.RoleAuth("admin", "dispatcher"), restrictedHttp.ApproveFirstLevel)
			restricted.POST("/:id/approve/second", middleware.RoleAuth("admin"), restrictedHttp.ApproveSecondLevel)
			restricted.POST("/:id/reject", middleware.RoleAuth("admin", "dispatcher"), restrictedHttp.RejectApproval)
			restricted.POST("/:id/revoke", middleware.RoleAuth("admin"), restrictedHttp.RevokeApproval)
			restricted.GET("/:id/approvals", restrictedHttp.GetApprovalHistory)
			restricted.GET("/approvals/pending", middleware.RoleAuth("admin", "dispatcher"), restrictedHttp.ListPendingApprovals)

			restricted.GET("/sync/pull", restrictedHttp.PullActiveAreas)

			templates := restricted.Group("/templates")
			{
				templates.GET("", restrictedHttp.ListTemplates)
				templates.GET("/:id", restrictedHttp.GetTemplate)
				templates.POST("", middleware.RoleAuth("admin"), restrictedHttp.CreateTemplate)
				templates.PUT("/:id", middleware.RoleAuth("admin"), restrictedHttp.UpdateTemplate)
				templates.DELETE("/:id", middleware.RoleAuth("admin"), restrictedHttp.DeleteTemplate)
				templates.POST("/:id/apply", restrictedHttp.ApplyTemplate)
			}

			gis := restricted.Group("/gis", middleware.RoleAuth("admin"))
			{
				gis.POST("/import", restrictedHttp.ImportGisFile)
				gis.POST("/import-json", restrictedHttp.ImportGisData)
				gis.GET("/imports", restrictedHttp.ListGisImports)
			}
		}

		// Webhook 免鉴权（仅校验 X-Webhook-Token）
		api.POST("/traffic/webhook/import", replanHttp.WebhookImport)

		traffic := api.Group("/traffic", middleware.JWTAuth())
		{
			traffic.GET("/events", replanHttp.ListTrafficEvents)
			traffic.GET("/events/:id", replanHttp.GetTrafficEvent)
			traffic.POST("/events", replanHttp.CreateTrafficEvent)
			traffic.POST("/events/:id/resolve", middleware.RoleAuth("admin", "dispatcher"), replanHttp.ResolveTrafficEvent)
		}

		replan := api.Group("/replans", middleware.JWTAuth())
		{
			replan.POST("/trigger", middleware.RoleAuth("admin", "dispatcher"), replanHttp.TriggerReplan)
			replan.POST("/:id/confirm", replanHttp.ConfirmReplan)
			replan.GET("", replanHttp.ListReplanRecords)
			replan.GET("/:id", replanHttp.GetReplanRecord)
			replan.GET("/statistics/overview", replanHttp.GetReplanStatistics)
		}
	}
}
