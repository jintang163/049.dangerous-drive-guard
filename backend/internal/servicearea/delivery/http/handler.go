package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	serviceareaSvc "github.com/dangerous-drive-guard/backend/internal/servicearea/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var serviceAreaService *serviceareaSvc.ServiceAreaService

func initService() {
	if serviceAreaService == nil {
		serviceAreaService = serviceareaSvc.NewServiceAreaService()
	}
}

func GetRestCountdown(ctx context.Context, c *app.RequestContext) {
	initService()
	driverID, err := strconv.ParseInt(c.Query("driver_id"), 10, 64)
	if err != nil || driverID <= 0 {
		response.BadRequest(c, "invalid driver_id")
		return
	}
	vehicleID, _ := strconv.ParseInt(c.Query("vehicle_id"), 10, 64)

	result, err := serviceAreaService.GetRestCountdown(ctx, driverID, vehicleID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func StartDriving(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		DriverID  int64 `json:"driver_id" binding:"required"`
		VehicleID int64 `json:"vehicle_id" binding:"required"`
		WaybillID int64 `json:"waybill_id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := serviceAreaService.StartDriving(ctx, req.DriverID, req.VehicleID, req.WaybillID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func CheckInServiceArea(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.CheckInRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := serviceAreaService.CheckInServiceArea(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func CheckOutServiceArea(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.CheckOutRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := serviceAreaService.CheckOutServiceArea(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func RecommendServiceAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.RecommendServiceAreaRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := serviceAreaService.RecommendServiceAreas(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func SubmitReview(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.SubmitReviewRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := serviceAreaService.SubmitReview(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func ListServiceAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	hasDangerParking := c.Query("has_danger_parking")
	var hasDangerParkingPtr *bool
	if hasDangerParking != "" {
		b := hasDangerParking == "1" || hasDangerParking == "true"
		hasDangerParkingPtr = &b
	}

	list, total, err := serviceAreaService.ListServiceAreas(ctx, page, pageSize, keyword, hasDangerParkingPtr)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func GetServiceAreaDetail(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	area, status, err := serviceAreaService.GetServiceAreaDetail(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := map[string]interface{}{
		"basic_info":  area,
		"real_status": status,
	}
	response.Success(c, result)
}

func ListReviews(ctx context.Context, c *app.RequestContext) {
	initService()
	serviceAreaID, _ := strconv.ParseInt(c.Query("service_area_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := serviceAreaService.ListReviews(ctx, serviceAreaID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func ListDrivingRestRecords(ctx context.Context, c *app.RequestContext) {
	initService()
	driverID, _ := strconv.ParseInt(c.Query("driver_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	list, total, err := serviceAreaService.ListDrivingRestRecords(ctx, driverID, page, pageSize, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func AcceptRecommendation(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid recommendation id")
		return
	}

	now := time.Now()
	result := serviceAreaService.db.WithContext(ctx).Exec(`
		UPDATE service_area_recommendations SET status = 'accepted', accepted_at = ? WHERE id = ? AND status = 'pending'`,
		now, id,
	)
	if result.Error != nil {
		response.InternalError(c, result.Error.Error())
		return
	}

	response.Success(c, map[string]interface{}{"success": true, "accepted_at": now})
}

func RejectRecommendation(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid recommendation id")
		return
	}

	var reason struct {
		Reason string `json:"reason"`
	}
	c.BindAndValidate(&reason)

	result := serviceAreaService.db.WithContext(ctx).Exec(`
		UPDATE service_area_recommendations SET status = 'rejected' WHERE id = ? AND status = 'pending'`,
		id,
	)
	if result.Error != nil {
		response.InternalError(c, result.Error.Error())
		return
	}

	response.Success(c, map[string]interface{}{"success": true})
}

func UpdateRealtimeStatus(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		ServiceAreaID          int64   `json:"service_area_id" binding:"required"`
		AvailableParkingSpaces int     `json:"available_parking_spaces"`
		AvailableDangerSpaces  int     `json:"available_danger_spaces"`
		SecurityLevel          int     `json:"security_level"`
		RestaurantRating       float64 `json:"restaurant_rating"`
		CrowdLevel             int     `json:"crowd_level"`
		WeatherCondition       string  `json:"weather_condition"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()

	var existsID int64
	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT id FROM service_area_realtime_status WHERE service_area_id = ?",
		req.ServiceAreaID,
	).Scan(&existsID)

	var err error
	if existsID > 0 {
		result := serviceAreaService.db.WithContext(ctx).Exec(`
			UPDATE service_area_realtime_status SET
			available_parking_spaces = ?,
			available_danger_spaces = ?,
			security_level = ?,
			restaurant_rating = ?,
			crowd_level = ?,
			weather_condition = ?,
			update_time = ?,
			data_source = 'manual'
			WHERE service_area_id = ?`,
			req.AvailableParkingSpaces, req.AvailableDangerSpaces,
			req.SecurityLevel, req.RestaurantRating,
			req.CrowdLevel, req.WeatherCondition,
			now, req.ServiceAreaID,
		)
		err = result.Error
	} else {
		result := serviceAreaService.db.WithContext(ctx).Exec(`
			INSERT INTO service_area_realtime_status
			(service_area_id, available_parking_spaces, available_danger_spaces,
			 security_level, restaurant_rating, crowd_level, weather_condition,
			 update_time, data_source)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'manual')`,
			req.ServiceAreaID, req.AvailableParkingSpaces, req.AvailableDangerSpaces,
			req.SecurityLevel, req.RestaurantRating, req.CrowdLevel, req.WeatherCondition,
			now,
		)
		err = result.Error
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{"success": true, "update_time": now})
}

func GetStatistics(ctx context.Context, c *app.RequestContext) {
	initService()
	var totalAreas, dangerAreas int64
	var avgRating float64
	var todayCheckIns, todayReviews int64

	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM service_areas WHERE status = 1",
	).Scan(&totalAreas)

	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM service_areas WHERE status = 1 AND has_danger_goods_parking = 1",
	).Scan(&dangerAreas)

	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT AVG(rating) FROM service_areas WHERE status = 1 AND rating > 0",
	).Scan(&avgRating)

	today := time.Now().Format("2006-01-02")
	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM driving_rest_records WHERE DATE(check_in_time) = ?",
		today,
	).Scan(&todayCheckIns)

	serviceAreaService.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM service_area_reviews WHERE DATE(created_at) = ? AND status = 1",
		today,
	).Scan(&todayReviews)

	result := map[string]interface{}{
		"total_service_areas":    totalAreas,
		"danger_parking_areas":   dangerAreas,
		"average_rating":         avgRating,
		"today_check_ins":        todayCheckIns,
		"today_reviews":          todayReviews,
	}

	response.Success(c, result)
}
