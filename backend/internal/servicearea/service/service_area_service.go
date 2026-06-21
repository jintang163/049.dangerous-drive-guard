package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type ServiceAreaService struct {
	db *database.TIDB
}

func NewServiceAreaService() *ServiceAreaService {
	return &ServiceAreaService{
		db: database.GetDB(),
	}
}

const (
	defaultMaxContinuousDrive = 240
	defaultMinRestRequired    = 20
	earthRadiusKm             = 6371.0
)

func (s *ServiceAreaService) GetRestCountdown(ctx context.Context, driverID, vehicleID int64) (*model.RestCountdownResponse, error) {
	var record model.DrivingRestRecord
	now := time.Now()
	today := now.Format("2006-01-02")

	row := s.db.WithContext(ctx).Raw(`
		SELECT id, driver_id, vehicle_id, waybill_id, record_date,
		       drive_start_time, drive_end_time, continuous_drive_minutes,
		       rest_start_time, rest_end_time, rest_duration_minutes,
		       rest_service_area_id, rest_service_area_name,
		       status, is_overtime, overtime_minutes,
		       check_in_time, check_in_latitude, check_in_longitude,
		       min_rest_required, max_continuous_drive,
		       created_at, updated_at
		FROM driving_rest_records
		WHERE driver_id = ? AND vehicle_id = ? AND record_date = ? AND status IN ('driving', 'resting')
		ORDER BY created_at DESC
		LIMIT 1`,
		driverID, vehicleID, today,
	).Row()

	err := row.Scan(&record.ID, &record.DriverID, &record.VehicleID, &record.WaybillID, &record.RecordDate,
		&record.DriveStartTime, &record.DriveEndTime, &record.ContinuousDriveMinutes,
		&record.RestStartTime, &record.RestEndTime, &record.RestDurationMinutes,
		&record.RestServiceAreaID, &record.RestServiceAreaName,
		&record.Status, &record.IsOvertime, &record.OvertimeMinutes,
		&record.CheckInTime, &record.CheckInLatitude, &record.CheckInLongitude,
		&record.MinRestRequired, &record.MaxContinuousDrive,
		&record.CreatedAt, &record.UpdatedAt)

	if err != nil || record.ID == 0 {
		return &model.RestCountdownResponse{
			DriverID:             driverID,
			VehicleID:            vehicleID,
			Status:               "driving",
			ContinuousDriveMinutes: 0,
			RemainingDriveMinutes: defaultMaxContinuousDrive,
			MaxContinuousDrive:   defaultMaxContinuousDrive,
			IsOvertime:           false,
			OvertimeMinutes:      0,
			MinRestRequired:      defaultMinRestRequired,
			CurrentRestMinutes:   0,
			RestProgressPercent:  0,
			CanContinueDriving:   true,
		}, nil
	}

	response := &model.RestCountdownResponse{
		DriverID:             record.DriverID,
		VehicleID:            record.VehicleID,
		WaybillID:            record.WaybillID,
		Status:               string(record.Status),
		MaxContinuousDrive:   record.MaxContinuousDrive,
		MinRestRequired:      record.MinRestRequired,
	}

	if record.Status == model.DrivingRestStatusDriving {
		continuousMinutes := int(now.Sub(record.DriveStartTime).Minutes())
		remaining := record.MaxContinuousDrive - continuousMinutes
		if remaining < 0 {
			remaining = 0
		}

		response.ContinuousDriveMinutes = continuousMinutes
		response.RemainingDriveMinutes = remaining
		response.IsOvertime = continuousMinutes > record.MaxContinuousDrive
		response.OvertimeMinutes = int(math.Max(0, float64(continuousMinutes-record.MaxContinuousDrive)))
		response.CanContinueDriving = !response.IsOvertime

		if record.WaybillID > 0 {
			response.WaybillID = record.WaybillID
		}

	} else if record.Status == model.DrivingRestStatusResting {
		var restMinutes int
		if record.RestStartTime != nil {
			restMinutes = int(now.Sub(*record.RestStartTime).Minutes())
		}

		progress := 0.0
		if record.MinRestRequired > 0 {
			progress = math.Min(100, float64(restMinutes)/float64(record.MinRestRequired)*100)
		}

		response.CurrentRestMinutes = restMinutes
		response.RestProgressPercent = math.Round(progress*100) / 100
		response.CanContinueDriving = restMinutes >= record.MinRestRequired
		response.CurrentServiceAreaID = record.RestServiceAreaID
		response.CurrentServiceAreaName = record.RestServiceAreaName
	}

	return response, nil
}

func (s *ServiceAreaService) StartDriving(ctx context.Context, driverID, vehicleID, waybillID int64) (*model.DrivingRestRecord, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	var existingID int64
	s.db.WithContext(ctx).Raw(`
		SELECT id FROM driving_rest_records
		WHERE driver_id = ? AND vehicle_id = ? AND status = 'driving'
		ORDER BY created_at DESC LIMIT 1`,
		driverID, vehicleID,
	).Scan(&existingID)

	if existingID > 0 {
		row := s.db.WithContext(ctx).Raw(`
			SELECT id, driver_id, vehicle_id, waybill_id, record_date,
			       drive_start_time, continuous_drive_minutes, status,
			       max_continuous_drive, min_rest_required
			FROM driving_rest_records WHERE id = ?`, existingID).Row()

		var record model.DrivingRestRecord
		row.Scan(&record.ID, &record.DriverID, &record.VehicleID, &record.WaybillID, &record.RecordDate,
			&record.DriveStartTime, &record.ContinuousDriveMinutes, &record.Status,
			&record.MaxContinuousDrive, &record.MinRestRequired)

		continuousMinutes := int(now.Sub(record.DriveStartTime).Minutes())
		record.ContinuousDriveMinutes = continuousMinutes
		record.IsOvertime = continuousMinutes > record.MaxContinuousDrive
		record.OvertimeMinutes = int(math.Max(0, float64(continuousMinutes-record.MaxContinuousDrive)))
		record.RemainingDriveMinutes = int(math.Max(0, float64(record.MaxContinuousDrive-continuousMinutes)))

		return &record, nil
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO driving_rest_records
		(driver_id, vehicle_id, waybill_id, record_date, drive_start_time, status,
		 continuous_drive_minutes, max_continuous_drive, min_rest_required)
		VALUES (?, ?, ?, ?, ?, 'driving', 0, ?, ?)`,
		driverID, vehicleID, waybillID, today, now,
		defaultMaxContinuousDrive, defaultMinRestRequired,
	)

	if result.Error != nil {
		logger.Sugar.Errorf("start driving record error: %v", result.Error)
		return nil, result.Error
	}

	var newID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&newID)

	record := &model.DrivingRestRecord{
		BaseModel: model.BaseModel{
			ID:        newID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		DriverID:            driverID,
		VehicleID:           vehicleID,
		WaybillID:           waybillID,
		RecordDate:          today,
		DriveStartTime:      now,
		Status:              model.DrivingRestStatusDriving,
		MaxContinuousDrive:  defaultMaxContinuousDrive,
		MinRestRequired:     defaultMinRestRequired,
		RemainingDriveMinutes: defaultMaxContinuousDrive,
	}

	return record, nil
}

func (s *ServiceAreaService) CheckInServiceArea(ctx context.Context, req *model.CheckInRequest) (*model.DrivingRestRecord, error) {
	now := time.Now()

	var areaName string
	s.db.WithContext(ctx).Raw("SELECT name FROM service_areas WHERE id = ?", req.ServiceAreaID).Scan(&areaName)

	if areaName == "" {
		return nil, fmt.Errorf("service area not found")
	}

	var recordID int64
	var currentStatus string
	s.db.WithContext(ctx).Raw(`
		SELECT id, status FROM driving_rest_records
		WHERE driver_id = ? AND vehicle_id = ? AND status IN ('driving', 'resting')
		ORDER BY created_at DESC LIMIT 1`,
		req.DriverID, req.VehicleID,
	).Scan(&recordID, &currentStatus)

	if recordID == 0 {
		_, err := s.StartDriving(ctx, req.DriverID, req.VehicleID, req.WaybillID)
		if err != nil {
			return nil, err
		}
		s.db.WithContext(ctx).Raw(`
			SELECT id FROM driving_rest_records
			WHERE driver_id = ? AND vehicle_id = ? AND status = 'driving'
			ORDER BY created_at DESC LIMIT 1`,
			req.DriverID, req.VehicleID,
		).Scan(&recordID)
	}

	result := s.db.WithContext(ctx).Exec(`
		UPDATE driving_rest_records SET
		status = 'resting',
		drive_end_time = ?,
		rest_start_time = ?,
		rest_service_area_id = ?,
		rest_service_area_name = ?,
		check_in_time = ?,
		check_in_latitude = ?,
		check_in_longitude = ?,
		continuous_drive_minutes = TIMESTAMPDIFF(MINUTE, drive_start_time, ?),
		updated_at = ?
		WHERE id = ?`,
		now, now,
		req.ServiceAreaID, areaName,
		now, req.Latitude, req.Longitude,
		now, now,
		recordID,
	)

	if result.Error != nil {
		logger.Sugar.Errorf("check in error: %v", result.Error)
		return nil, result.Error
	}

	var record model.DrivingRestRecord
	row := s.db.WithContext(ctx).Raw(`
		SELECT id, driver_id, vehicle_id, waybill_id, record_date,
		       drive_start_time, drive_end_time, continuous_drive_minutes,
		       rest_start_time, rest_service_area_id, rest_service_area_name,
		       status, check_in_time, check_in_latitude, check_in_longitude,
		       min_rest_required, max_continuous_drive
		FROM driving_rest_records WHERE id = ?`, recordID).Row()

	row.Scan(&record.ID, &record.DriverID, &record.VehicleID, &record.WaybillID, &record.RecordDate,
		&record.DriveStartTime, &record.DriveEndTime, &record.ContinuousDriveMinutes,
		&record.RestStartTime, &record.RestServiceAreaID, &record.RestServiceAreaName,
		&record.Status, &record.CheckInTime, &record.CheckInLatitude, &record.CheckInLongitude,
		&record.MinRestRequired, &record.MaxContinuousDrive)

	restMinutes := int(now.Sub(*record.RestStartTime).Minutes())
	progress := 0.0
	if record.MinRestRequired > 0 {
		progress = math.Min(100, float64(restMinutes)/float64(record.MinRestRequired)*100)
	}
	record.RestProgressPercent = math.Round(progress*100) / 100

	logger.Global.Info("driver checked in service area",
		zap.Int64("driver_id", req.DriverID),
		zap.Int64("vehicle_id", req.VehicleID),
		zap.Int64("service_area_id", req.ServiceAreaID),
		zap.String("service_area_name", areaName),
	)

	return &record, nil
}

func (s *ServiceAreaService) CheckOutServiceArea(ctx context.Context, req *model.CheckOutRequest) (*model.DrivingRestRecord, error) {
	now := time.Now()

	var recordID int64
	var restStartTime *time.Time
	var minRestRequired int
	s.db.WithContext(ctx).Raw(`
		SELECT id, rest_start_time, min_rest_required FROM driving_rest_records
		WHERE driver_id = ? AND vehicle_id = ? AND status = 'resting'
		ORDER BY created_at DESC LIMIT 1`,
		req.DriverID, req.VehicleID,
	).Scan(&recordID, &restStartTime, &minRestRequired)

	if recordID == 0 || restStartTime == nil {
		return nil, fmt.Errorf("no active rest record found")
	}

	restMinutes := int(now.Sub(*restStartTime).Minutes())

	if restMinutes < minRestRequired {
		return nil, fmt.Errorf("休息时间不足，至少需要 %d 分钟，当前已休息 %d 分钟", minRestRequired, restMinutes)
	}

	result := s.db.WithContext(ctx).Exec(`
		UPDATE driving_rest_records SET
		status = 'completed',
		rest_end_time = ?,
		rest_duration_minutes = ?,
		check_out_time = ?,
		check_out_latitude = ?,
		check_out_longitude = ?,
		updated_at = ?
		WHERE id = ?`,
		now, restMinutes,
		now, req.Latitude, req.Longitude,
		now, recordID,
	)

	if result.Error != nil {
		logger.Sugar.Errorf("check out error: %v", result.Error)
		return nil, result.Error
	}

	_, err := s.StartDriving(ctx, req.DriverID, req.VehicleID, 0)
	if err != nil {
		logger.Sugar.Warnf("auto start new driving record failed: %v", err)
	}

	var record model.DrivingRestRecord
	row := s.db.WithContext(ctx).Raw(`
		SELECT id, driver_id, vehicle_id, waybill_id, record_date,
		       drive_start_time, drive_end_time, continuous_drive_minutes,
		       rest_start_time, rest_end_time, rest_duration_minutes,
		       rest_service_area_id, rest_service_area_name,
		       status, min_rest_required, max_continuous_drive
		FROM driving_rest_records WHERE id = ?`, recordID).Row()

	row.Scan(&record.ID, &record.DriverID, &record.VehicleID, &record.WaybillID, &record.RecordDate,
		&record.DriveStartTime, &record.DriveEndTime, &record.ContinuousDriveMinutes,
		&record.RestStartTime, &record.RestEndTime, &record.RestDurationMinutes,
		&record.RestServiceAreaID, &record.RestServiceAreaName,
		&record.Status, &record.MinRestRequired, &record.MaxContinuousDrive)

	logger.Global.Info("driver checked out service area",
		zap.Int64("driver_id", req.DriverID),
		zap.Int64("vehicle_id", req.VehicleID),
		zap.Int("rest_minutes", restMinutes),
	)

	return &record, nil
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func (s *ServiceAreaService) RecommendServiceAreas(ctx context.Context, req *model.RecommendServiceAreaRequest) (*model.ServiceAreaRecommendation, error) {
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 100
	}

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT sa.id, sa.name, sa.highway_name, sa.direction,
		       sa.latitude, sa.longitude, sa.has_danger_goods_parking,
		       sa.danger_parking_spaces, sa.phone, sa.rating,
		       st.available_danger_spaces, st.available_parking_spaces,
		       st.security_level, st.security_patrol_interval,
		       st.restaurant_rating, st.restaurant_wait_minutes,
		       st.has_fuel, st.fuel_price_diesel,
		       st.has_charging, st.charging_piles_available,
		       st.crowd_level, st.update_time
		FROM service_areas sa
		LEFT JOIN service_area_realtime_status st ON st.service_area_id = sa.id
		WHERE sa.status = 1
		HAVING distance_km <= ?
		ORDER BY match_score DESC
		LIMIT 10`,
	).Rows()

	query := `
		SELECT sa.id, sa.name, sa.highway_name, sa.direction,
		       sa.latitude, sa.longitude, sa.has_danger_goods_parking,
		       sa.danger_parking_spaces, sa.phone, sa.rating,
		       st.available_danger_spaces, st.available_parking_spaces,
		       st.security_level, st.security_patrol_interval,
		       st.restaurant_rating, st.restaurant_wait_minutes,
		       st.has_fuel, st.fuel_price_diesel,
		       st.has_charging, st.charging_piles_available,
		       st.crowd_level, st.update_time
		FROM service_areas sa
		LEFT JOIN service_area_realtime_status st ON st.service_area_id = sa.id
		WHERE sa.status = 1
	`

	rows, err = s.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		logger.Sugar.Errorf("recommend service areas query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var candidates []model.RecommendedServiceArea

	for rows.Next() {
		var area model.RecommendedServiceArea
		var lat, lng float64
		var hasDangerParking bool
		var dangerSpaces int
		var availableDangerSpaces *int
		var securityLevel *int
		var restaurantRating *float64
		var hasFuel *bool
		var hasCharging *bool

		err = rows.Scan(
			&area.ServiceAreaID, &area.ServiceAreaName,
			&struct{}{}, &struct{}{},
			&lat, &lng,
			&hasDangerParking, &dangerSpaces,
			&struct{}{}, &struct{}{},
			&availableDangerSpaces, &struct{}{},
			&securityLevel, &struct{}{},
			&restaurantRating, &struct{}{},
			&hasFuel, &struct{}{},
			&hasCharging, &struct{}{},
			&struct{}{}, &struct{}{},
		)
		if err != nil {
			continue
		}

		distance := haversineDistance(req.Latitude, req.Longitude, lat, lng)

		if distance > radiusKm*3 {
			continue
		}

		area.DistanceKm = math.Round(distance*100) / 100

		avgSpeed := 80.0
		area.EstimatedArrivalMinutes = int(math.Round(distance / avgSpeed * 60))

		if availableDangerSpaces != nil {
			area.AvailableDangerSpaces = *availableDangerSpaces
		} else {
			area.AvailableDangerSpaces = dangerSpaces / 2
		}

		if securityLevel != nil {
			area.SecurityLevel = *securityLevel
		} else {
			area.SecurityLevel = 3
		}

		if restaurantRating != nil {
			area.RestaurantRating = *restaurantRating
		}

		if hasFuel != nil {
			area.HasFuel = *hasFuel
		}

		if hasCharging != nil {
			area.HasCharging = *hasCharging
		}

		if req.HazardClass != "" && !hasDangerParking {
			continue
		}

		matchScore := s.calculateMatchScore(&area, req)
		area.MatchScore = matchScore

		candidates = append(candidates, area)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].MatchScore > candidates[j].MatchScore
	})

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable service areas found")
	}

	best := candidates[0]

	reasonParts := []string{}
	if best.AvailableDangerSpaces > 0 {
		reasonParts = append(reasonParts, fmt.Sprintf("危化品车位剩余 %d 个", best.AvailableDangerSpaces))
	}
	if best.SecurityLevel >= 4 {
		reasonParts = append(reasonParts, "安保等级高")
	}
	if best.RestaurantRating >= 4.0 {
		reasonParts = append(reasonParts, "餐饮评价好")
	}
	if best.DistanceKm <= 30 {
		reasonParts = append(reasonParts, "距离最近")
	}

	recommendReason := strings.Join(reasonParts, "，")
	if recommendReason == "" {
		recommendReason = "综合推荐"
	}

	recommendNo := fmt.Sprintf("SA%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))

	var alternatives []model.RecommendedServiceArea
	if len(candidates) > 1 {
		altCount := len(candidates) - 1
		if altCount > 3 {
			altCount = 3
		}
		alternatives = candidates[1 : 1+altCount]
	}

	alternativesJSON, _ := json.Marshal(alternatives)

	continuousDrive := req.FatigueScore
	if continuousDrive == 0 {
		continuousDrive = 120
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO service_area_recommendations
		(recommend_no, driver_id, vehicle_id, waybill_id,
		 current_latitude, current_longitude,
		 continuous_drive_minutes, remaining_drive_minutes,
		 fatigue_score, hazard_class,
		 recommend_reason,
		 recommended_service_area_id, recommended_service_area_name,
		 distance_km, estimated_arrival_minutes,
		 alternatives, status, dispatch_source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 'system')`,
		recommendNo, req.DriverID, req.VehicleID, req.WaybillID,
		req.Latitude, req.Longitude,
		int(continuousDrive), int(math.Max(0, 240-continuousDrive)),
		req.FatigueScore, req.HazardClass,
		recommendReason,
		best.ServiceAreaID, best.ServiceAreaName,
		best.DistanceKm, best.EstimatedArrivalMinutes,
		string(alternativesJSON),
	)

	if result.Error != nil {
		logger.Sugar.Warnf("save recommendation record error: %v", result.Error)
	}

	var recID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&recID)

	recommendation := &model.ServiceAreaRecommendation{
		BaseModel: model.BaseModel{
			ID: recID,
		},
		RecommendNo:            recommendNo,
		DriverID:               req.DriverID,
		VehicleID:              req.VehicleID,
		WaybillID:              req.WaybillID,
		CurrentLatitude:        req.Latitude,
		CurrentLongitude:       req.Longitude,
		ContinuousDriveMinutes: int(continuousDrive),
		RemainingDriveMinutes:  int(math.Max(0, 240-continuousDrive)),
		FatigueScore:           req.FatigueScore,
		HazardClass:            req.HazardClass,
		RecommendReason:        recommendReason,
		RecommendedServiceAreaID:   best.ServiceAreaID,
		RecommendedServiceAreaName: best.ServiceAreaName,
		DistanceKm:             best.DistanceKm,
		EstimatedArrivalMinutes: best.EstimatedArrivalMinutes,
		AlternativesArray:      alternatives,
		Status:                 "pending",
		DispatchSource:         "system",
	}

	logger.Global.Info("service area recommendation generated",
		zap.Int64("driver_id", req.DriverID),
		zap.String("recommend_no", recommendNo),
		zap.String("recommended_area", best.ServiceAreaName),
		zap.Float64("distance_km", best.DistanceKm),
	)

	return recommendation, nil
}

func (s *ServiceAreaService) calculateMatchScore(area *model.RecommendedServiceArea, req *model.RecommendServiceAreaRequest) float64 {
	score := 0.0

	distanceWeight := 30.0
	distanceScore := math.Max(0, 100-area.DistanceKm*2)
	score += distanceScore * distanceWeight / 100

	securityWeight := 25.0
	securityScore := float64(area.SecurityLevel) * 20
	score += securityScore * securityWeight / 100

	parkingWeight := 20.0
	parkingScore := 0.0
	if area.AvailableDangerSpaces > 10 {
		parkingScore = 100
	} else if area.AvailableDangerSpaces > 5 {
		parkingScore = 70
	} else if area.AvailableDangerSpaces > 0 {
		parkingScore = 40
	} else {
		parkingScore = 10
	}
	score += parkingScore * parkingWeight / 100

	foodWeight := 15.0
	foodScore := area.RestaurantRating * 20
	score += foodScore * foodWeight / 100

	facilityWeight := 10.0
	facilityScore := 0.0
	if area.HasFuel {
		facilityScore += 30
	}
	if area.HasCharging {
		facilityScore += 30
	}
	score += facilityScore * facilityWeight / 100

	return math.Round(score*100) / 100
}

func (s *ServiceAreaService) SubmitReview(ctx context.Context, req *model.SubmitReviewRequest) (*model.ServiceAreaReview, error) {
	if req.SecurityScore < 1 || req.SecurityScore > 5 {
		return nil, fmt.Errorf("security score must be between 1 and 5")
	}

	totalScore := float64(req.SecurityScore)
	scoreCount := 1

	if req.EnvironmentScore > 0 {
		totalScore += float64(req.EnvironmentScore)
		scoreCount++
	}
	if req.FoodScore > 0 {
		totalScore += float64(req.FoodScore)
		scoreCount++
	}
	if req.ServiceScore > 0 {
		totalScore += float64(req.ServiceScore)
		scoreCount++
	}

	overallScore := math.Round(totalScore/float64(scoreCount)*100) / 100

	tagsJSON, _ := json.Marshal(req.Tags)
	imagesJSON, _ := json.Marshal(req.Images)

	var driverName string
	s.db.WithContext(ctx).Raw("SELECT real_name FROM users WHERE id = ?", req.DriverID).Scan(&driverName)

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO service_area_reviews
		(service_area_id, driver_id, driver_name, waybill_id, vehicle_id,
		 security_score, environment_score, food_score, service_score,
		 overall_score, comment_text, tags, images, is_anonymous, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		req.ServiceAreaID, req.DriverID, driverName,
		req.WaybillID, req.VehicleID,
		req.SecurityScore, req.EnvironmentScore, req.FoodScore, req.ServiceScore,
		overallScore, req.CommentText,
		string(tagsJSON), string(imagesJSON),
		req.IsAnonymous,
	)

	if result.Error != nil {
		logger.Sugar.Errorf("submit review error: %v", result.Error)
		return nil, result.Error
	}

	var reviewID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&reviewID)

	s.updateServiceAreaRating(ctx, req.ServiceAreaID)

	review := &model.ServiceAreaReview{
		BaseModel: model.BaseModel{
			ID: reviewID,
		},
		ServiceAreaID:  req.ServiceAreaID,
		DriverID:       req.DriverID,
		DriverName:     driverName,
		SecurityScore:  req.SecurityScore,
		EnvironmentScore: req.EnvironmentScore,
		FoodScore:      req.FoodScore,
		ServiceScore:   req.ServiceScore,
		OverallScore:   overallScore,
		CommentText:    req.CommentText,
		TagsArray:      req.Tags,
		ImagesArray:    req.Images,
		IsAnonymous:    req.IsAnonymous,
		Status:         1,
	}

	logger.Global.Info("service area review submitted",
		zap.Int64("service_area_id", req.ServiceAreaID),
		zap.Int64("driver_id", req.DriverID),
		zap.Float64("overall_score", overallScore),
	)

	return review, nil
}

func (s *ServiceAreaService) updateServiceAreaRating(ctx context.Context, serviceAreaID int64) {
	var avgRating float64
	var reviewCount int64

	s.db.WithContext(ctx).Raw(`
		SELECT AVG(overall_score), COUNT(*)
		FROM service_area_reviews
		WHERE service_area_id = ? AND status = 1`,
		serviceAreaID,
	).Scan(&avgRating, &reviewCount)

	if reviewCount > 0 {
		s.db.WithContext(ctx).Exec(`
			UPDATE service_areas SET rating = ? WHERE id = ?`,
			math.Round(avgRating*100)/100, serviceAreaID,
		)

		s.db.WithContext(ctx().Exec(`
			UPDATE service_area_realtime_status SET restaurant_rating = ? WHERE service_area_id = ?`,
			math.Round(avgRating*100)/100, serviceAreaID,
		)
	}
}

func (s *ServiceAreaService) ListServiceAreas(ctx context.Context, page, pageSize int, keyword string, hasDangerParking *bool) ([]*model.ServiceArea, int64, error) {
	var areas []*model.ServiceArea
	var total int64

	query := s.db.WithContext(ctx).Table("service_areas").Where("status = 1")

	if keyword != "" {
		query = query.Where("name LIKE ? OR highway_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if hasDangerParking != nil {
		query = query.Where("has_danger_goods_parking = ?", *hasDangerParking)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var area model.ServiceArea
		rows.Scan(&area.ID, &area.Name, &area.HighwayName, &area.Direction,
			&area.Province, &area.City, &area.Latitude, &area.Longitude,
			&area.DistanceFromStart, &area.HasRestaurant, &area.HasHotel,
			&area.HasFuelStation, &area.HasCharging, &area.HasRestRoom,
			&area.HasMaintenance, &area.HasDangerGoodsParking,
			&area.ParkingSpaces, &area.DangerParkingSpaces,
			&area.Phone, &area.Rating, &area.Status,
			&area.CreatedAt, &area.UpdatedAt)
		areas = append(areas, &area)
	}

	return areas, total, nil
}

func (s *ServiceAreaService) GetServiceAreaDetail(ctx context.Context, id int64) (*model.ServiceArea, *model.ServiceAreaRealtimeStatus, error) {
	var area model.ServiceArea
	row := s.db.WithContext(ctx).Raw(`
		SELECT id, name, highway_name, direction, province, city,
		       latitude, longitude, distance_from_start,
		       has_restaurant, has_hotel, has_fuel_station, has_charging,
		       has_rest_room, has_maintenance, has_danger_goods_parking,
		       parking_spaces, danger_parking_spaces, phone, rating, status
		FROM service_areas WHERE id = ?`, id).Row()

	err := row.Scan(&area.ID, &area.Name, &area.HighwayName, &area.Direction,
		&area.Province, &area.City, &area.Latitude, &area.Longitude,
		&area.DistanceFromStart, &area.HasRestaurant, &area.HasHotel,
		&area.HasFuelStation, &area.HasCharging, &area.HasRestRoom,
		&area.HasMaintenance, &area.HasDangerGoodsParking,
		&area.ParkingSpaces, &area.DangerParkingSpaces,
		&area.Phone, &area.Rating, &area.Status)

	if err != nil || area.ID == 0 {
		return nil, nil, fmt.Errorf("service area not found")
	}

	var status model.ServiceAreaRealtimeStatus
	status.ServiceAreaID = id

	statusRow := s.db.WithContext(ctx).Raw(`
		SELECT id, service_area_id, total_parking_spaces, available_parking_spaces,
		       total_danger_spaces, available_danger_spaces,
		       has_fuel, fuel_price_92, fuel_price_95, fuel_price_diesel,
		       has_charging, charging_piles_total, charging_piles_available,
		       has_restaurant, restaurant_rating, restaurant_wait_minutes,
		       has_hotel, hotel_rating, has_maintenance,
		       security_level, security_patrol_interval,
		       crowd_level, weather_condition, update_time, data_source
		FROM service_area_realtime_status WHERE service_area_id = ?`, id).Row()

	err = statusRow.Scan(&status.ID, &status.ServiceAreaID,
		&status.TotalParkingSpaces, &status.AvailableParkingSpaces,
		&status.TotalDangerSpaces, &status.AvailableDangerSpaces,
		&status.HasFuel, &status.FuelPrice92, &status.FuelPrice95, &status.FuelPriceDiesel,
		&status.HasCharging, &status.ChargingPilesTotal, &status.ChargingPilesAvailable,
		&status.HasRestaurant, &status.RestaurantRating, &status.RestaurantWaitMinutes,
		&status.HasHotel, &status.HotelRating, &status.HasMaintenance,
		&status.SecurityLevel, &status.SecurityPatrolInterval,
		&status.CrowdLevel, &status.WeatherCondition, &status.UpdateTime, &status.DataSource)

	if err != nil || status.ID == 0 {
		status = model.ServiceAreaRealtimeStatus{
			ServiceAreaID:         id,
			TotalParkingSpaces:    area.ParkingSpaces,
			AvailableParkingSpaces: area.ParkingSpaces / 2,
			TotalDangerSpaces:     area.DangerParkingSpaces,
			AvailableDangerSpaces: area.DangerParkingSpaces / 2,
			HasFuel:               area.HasFuelStation,
			HasCharging:           area.HasCharging,
			HasRestaurant:         area.HasRestaurant,
			HasHotel:              area.HasHotel,
			HasMaintenance:        area.HasMaintenance,
			SecurityLevel:         3,
			SecurityPatrolInterval: 30,
			CrowdLevel:            2,
			UpdateTime:            time.Now(),
			DataSource:            "estimated",
		}
	}

	return &area, &status, nil
}

func (s *ServiceAreaService) ListReviews(ctx context.Context, serviceAreaID int64, page, pageSize int) ([]*model.ServiceAreaReview, int64, error) {
	var reviews []*model.ServiceAreaReview
	var total int64

	query := s.db.WithContext(ctx).Table("service_area_reviews").Where("status = 1")
	if serviceAreaID > 0 {
		query = query.Where("service_area_id = ?", serviceAreaID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var review model.ServiceAreaReview
		var tagsJSON, imagesJSON string

		rows.Scan(&review.ID, &review.ServiceAreaID, &review.DriverID, &review.DriverName,
			&review.WaybillID, &review.VehicleID,
			&review.SecurityScore, &review.EnvironmentScore, &review.FoodScore, &review.ServiceScore,
			&review.OverallScore, &review.CommentText, &tagsJSON, &imagesJSON,
			&review.IsAnonymous, &review.Status, &review.CheckInRecordID,
			&review.CreatedAt, &review.UpdatedAt)

		if review.IsAnonymous {
			review.DriverName = "匿名用户"
		}

		json.Unmarshal([]byte(tagsJSON), &review.TagsArray)
		json.Unmarshal([]byte(imagesJSON), &review.ImagesArray)

		reviews = append(reviews, &review)
	}

	return reviews, total, nil
}

func (s *ServiceAreaService) ListDrivingRestRecords(ctx context.Context, driverID int64, page, pageSize int, startDate, endDate string) ([]*model.DrivingRestRecord, int64, error) {
	var records []*model.DrivingRestRecord
	var total int64

	query := s.db.WithContext(ctx).Table("driving_rest_records")
	if driverID > 0 {
		query = query.Where("driver_id = ?", driverID)
	}
	if startDate != "" {
		query = query.Where("record_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("record_date <= ?", endDate)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.DrivingRestRecord
		rows.Scan(&record.ID, &record.DriverID, &record.VehicleID, &record.WaybillID,
			&record.RecordDate, &record.DriveStartTime, &record.DriveEndTime,
			&record.ContinuousDriveMinutes, &record.RestStartTime, &record.RestEndTime,
			&record.RestDurationMinutes, &record.RestServiceAreaID, &record.RestServiceAreaName,
			&record.Status, &record.IsOvertime, &record.OvertimeMinutes,
			&record.CheckInTime, &record.CheckInLatitude, &record.CheckInLongitude,
			&record.CheckOutTime, &record.CheckOutLatitude, &record.CheckOutLongitude,
			&record.MinRestRequired, &record.MaxContinuousDrive,
			&record.CreatedAt, &record.UpdatedAt)
		records = append(records, &record)
	}

	return records, total, nil
}
