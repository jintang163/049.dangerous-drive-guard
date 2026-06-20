package routes

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	routeHandler "github.com/dangerous-drive-guard/backend/internal/route/delivery/http"
	"github.com/dangerous-drive-guard/backend/internal/fatigue/delivery/http"
	fatigueWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	monitorHttp "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/http"
	transportHandler "github.com/dangerous-drive-guard/backend/internal/transport/delivery/http"
	vehicleHandler "github.com/dangerous-drive-guard/backend/internal/vehicle/delivery/http"
	userHandler "github.com/dangerous-drive-guard/backend/internal/user/delivery/http"
	authHandler "github.com/dangerous-drive-guard/backend/internal/auth/delivery/http"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
)

func Register(h *server.Hertz) {
	api := h.Group("/api/v1")
	{
		api.Use(middleware.TraceID(), middleware.CORS())

		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", middleware.JWTAuth(), authHandler.Refresh)
			auth.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
		}

		user := api.Group("/users", middleware.JWTAuth())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/:id", middleware.RoleAuth("admin", "dispatcher"), userHandler.GetByID)
			user.GET("", middleware.RoleAuth("admin", "dispatcher"), userHandler.List)
			user.POST("", middleware.RoleAuth("admin"), userHandler.Create)
			user.PUT("/:id", middleware.RoleAuth("admin"), userHandler.Update)
			user.DELETE("/:id", middleware.RoleAuth("admin"), userHandler.Delete)
		}

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
			fatigue.POST("/detect", fatigueHandler.DetectFatigue)
			fatigue.POST("/upload/frame", fatigueHandler.UploadFrame)
			fatigue.GET("/history/:vehicle_id", fatigueHandler.GetHistory)
			fatigue.GET("/alarms", middleware.RoleAuth("admin", "dispatcher"), fatigueHandler.ListAlarms)
			fatigue.POST("/alarms/:id/ack", middleware.RoleAuth("admin", "dispatcher"), fatigueHandler.AckAlarm)
		}

		monitor := api.Group("/monitor", middleware.JWTAuth())
		{
			monitor.GET("/vehicles/realtime", monitorHttp.GetRealtimeVehicles)
			monitor.GET("/vehicle/:id/status", monitorHttp.GetVehicleStatus)
			monitor.GET("/statistics", monitorHttp.GetStatistics)
			monitor.POST("/intercom/:vehicle_id", middleware.RoleAuth("admin", "dispatcher"), monitorHttp.SendVoiceIntercom)
			monitor.POST("/dispatch/service-area", middleware.RoleAuth("admin", "dispatcher"), monitorHttp.DispatchServiceArea)
			}

		api.GET("/ws/monitor", monitorWs.MonitorWebSocket)
		api.GET("/ws/vehicle/:vehicle_id", monitorWs.VehicleWebSocket)

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

		vehicle := api.Group("/vehicles", middleware.JWTAuth())
		{
			vehicle.POST("", middleware.RoleAuth("admin"), vehicleHandler.CreateVehicle)
			vehicle.GET("/:id", vehicleHandler.GetVehicle)
			vehicle.GET("", vehicleHandler.ListVehicles)
			vehicle.PUT("/:id", middleware.RoleAuth("admin"), vehicleHandler.UpdateVehicle)
			vehicle.DELETE("/:id", middleware.RoleAuth("admin"), vehicleHandler.DeleteVehicle)

			diag := vehicle.Group("/diagnostics")
			diag.POST("/upload", vehicleHandler.UploadDiagnostics)
			diag.GET("/:vehicle_id/recent", vehicleHandler.GetRecentDiagnostics)
			diag.GET("/:vehicle_id/faults", vehicleHandler.GetFaultAlerts)

			score := vehicle.Group("/score")
			score.GET("/driver/:driver_id", vehicleHandler.GetDriverScore)
			score.GET("/ranking", vehicleHandler.GetScoreRanking)
		}
	}
}
