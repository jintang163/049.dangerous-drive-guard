package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"gorm.io/gorm"
)

type WeatherService struct {
	db         *gorm.DB
	amapKey    string
	httpClient *http.Client
}

func NewWeatherService(cfg *config.Config) *WeatherService {
	return &WeatherService{
		db:      database.GetDB(),
		amapKey: cfg.Map.AMap.Key,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type amapWeatherResponse struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Lives    []struct {
		Province      string  `json:"province"`
		City          string  `json:"city"`
		Adcode        string  `json:"adcode"`
		Weather       string  `json:"weather"`
		Temperature   string  `json:"temperature"`
		WindDirection string  `json:"winddirection"`
		WindPower     string  `json:"windpower"`
		Humidity      string  `json:"humidity"`
		ReportTime    string  `json:"reporttime"`
		Visibility    float64 `json:"visibility"`
	} `json:"lives"`
}

func (s *WeatherService) GetCurrentWeather(ctx context.Context, lat, lng float64) (*model.WeatherData, error) {
	if lat == 0 || lng == 0 {
		return nil, fmt.Errorf("invalid coordinates")
	}

	url := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?location=%.6f,%.6f&key=%s&extensions=base",
		lng, lat, s.amapKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Sugar.Errorf("create weather request error: %v", err)
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Sugar.Errorf("call amap weather api error: %v", err)
		return nil, fmt.Errorf("call weather api error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar.Errorf("read weather response error: %v", err)
		return nil, fmt.Errorf("read response error: %w", err)
	}

	var amapResp amapWeatherResponse
	if err := json.Unmarshal(body, &amapResp); err != nil {
		logger.Sugar.Errorf("parse weather response error: %v, body: %s", err, string(body))
		return nil, fmt.Errorf("parse response error: %w", err)
	}

	if amapResp.Status != "1" || len(amapResp.Lives) == 0 {
		logger.Sugar.Warnf("amap weather api returned no data: info=%s", amapResp.Info)
		return &model.WeatherData{
			Temp:       0,
			Humidity:   0,
			WindSpeed:  0,
			Visibility: 0,
			Condition:  "unknown",
		}, nil
	}

	life := amapResp.Lives[0]
	var temp, humidity, windSpeed float64
	fmt.Sscanf(life.Temperature, "%f", &temp)
	fmt.Sscanf(life.Humidity, "%f", &humidity)
	fmt.Sscanf(life.WindPower, "%f", &windSpeed)

	return &model.WeatherData{
		Temp:       temp,
		Humidity:   humidity,
		WindSpeed:  windSpeed,
		Visibility: life.Visibility,
		Condition:  life.Weather,
	}, nil
}

func (s *WeatherService) GetRouteWeather(ctx context.Context, routeID int64) ([]*model.RouteWeatherPoint, error) {
	if routeID <= 0 {
		return nil, fmt.Errorf("invalid route id")
	}

	var routePlan model.RoutePlan
	if err := s.db.WithContext(ctx).Where("id = ?", routeID).First(&routePlan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("route not found")
		}
		return nil, fmt.Errorf("query route error: %w", err)
	}

	if len(routePlan.RoutePath) == 0 {
		return []*model.RouteWeatherPoint{}, nil
	}

	var result []*model.RouteWeatherPoint
	totalDistance := routePlan.TotalDistance
	estimatedDuration := float64(routePlan.EstimatedDuration)

	sampleInterval := 5
	for i := 0; i < len(routePlan.RoutePath); i += sampleInterval {
		point := routePlan.RoutePath[i]
		weather, err := s.GetCurrentWeather(ctx, point.Lat, point.Lng)
		if err != nil {
			logger.Sugar.Warnf("get weather for point (%.6f,%.6f) error: %v", point.Lat, point.Lng, err)
			weather = &model.WeatherData{Condition: "unknown"}
		}

		distanceRatio := float64(i) / float64(len(routePlan.RoutePath))
		distanceFromStart := totalDistance * distanceRatio
		estimatedSeconds := estimatedDuration * distanceRatio
		estimatedTime := time.Now().Add(time.Duration(estimatedSeconds) * time.Second)

		result = append(result, &model.RouteWeatherPoint{
			Lat:               point.Lat,
			Lng:               point.Lng,
			WeatherData:       *weather,
			DistanceFromStart: distanceFromStart,
			EstimatedTime:     estimatedTime,
		})
	}

	return result, nil
}

func (s *WeatherService) ListWarnings(ctx context.Context, page, pageSize int, status string) (*model.WeatherWarningPage, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var warnings []*model.WeatherWarning
	var total int64

	query := s.db.WithContext(ctx).Model(&model.WeatherWarning{})

	if status != "" {
		query = query.Where("processed = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count warnings error: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("publish_time DESC").Offset(offset).Limit(pageSize).Find(&warnings).Error; err != nil {
		return nil, fmt.Errorf("query warnings error: %w", err)
	}

	return &model.WeatherWarningPage{
		List:     warnings,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *WeatherService) GetActiveWarnings(ctx context.Context, province string) ([]*model.WeatherWarning, error) {
	var warnings []*model.WeatherWarning

	query := s.db.WithContext(ctx).Model(&model.WeatherWarning{}).
		Where("processed = 0 AND (end_time IS NULL OR end_time > ?)", time.Now())

	if province != "" {
		query = query.Where("JSON_CONTAINS(affected_provinces, ?)", fmt.Sprintf(`"%s"`, province))
	}

	if err := query.Order("publish_time DESC").Find(&warnings).Error; err != nil {
		return nil, fmt.Errorf("query active warnings error: %w", err)
	}

	return warnings, nil
}

func (s *WeatherService) CheckRouteAffected(ctx context.Context, routeID, warningID int64) (bool, []int64, error) {
	if routeID <= 0 || warningID <= 0 {
		return false, nil, fmt.Errorf("invalid route id or warning id")
	}

	var routePlan model.RoutePlan
	if err := s.db.WithContext(ctx).Where("id = ?", routeID).First(&routePlan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, fmt.Errorf("route not found")
		}
		return false, nil, fmt.Errorf("query route error: %w", err)
	}

	var warning model.WeatherWarning
	if err := s.db.WithContext(ctx).Where("id = ?", warningID).First(&warning).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, fmt.Errorf("warning not found")
		}
		return false, nil, fmt.Errorf("query warning error: %w", err)
	}

	var affectedWaybills []int64
	var affectedCount int64

	s.db.WithContext(ctx).Table("waybills").
		Where("route_plan_id = ? AND status IN (?, ?, ?)", routeID, "assigned", "loading", "in_transit").
		Count(&affectedCount)

	if affectedCount > 0 {
		rows, err := s.db.WithContext(ctx).Table("waybills").
			Select("id").
			Where("route_plan_id = ? AND status IN (?, ?, ?)", routeID, "assigned", "loading", "in_transit").
			Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id int64
				rows.Scan(&id)
				affectedWaybills = append(affectedWaybills, id)
			}
		}
	}

	affected := len(routePlan.RoutePath) > 0 && affectedCount > 0

	return affected, affectedWaybills, nil
}

func (s *WeatherService) DB() *gorm.DB {
	return s.db
}

func (s *WeatherService) GetWarning(ctx context.Context, id int64) (*model.WeatherWarning, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid warning id")
	}

	var warning model.WeatherWarning
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&warning).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("warning not found")
		}
		return nil, fmt.Errorf("query warning error: %w", err)
	}

	return &warning, nil
}

func (s *WeatherService) ReplanAffectedRoutes(ctx context.Context, warningID int64) (int, error) {
	if warningID <= 0 {
		return 0, fmt.Errorf("invalid warning id")
	}

	warning, err := s.GetWarning(ctx, warningID)
	if err != nil {
		return 0, err
	}

	var waybills []struct {
		ID          int64
		RoutePlanID int64
	}

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT w.id, w.route_plan_id
		FROM waybills w
		INNER JOIN route_plans r ON w.route_plan_id = r.id
		WHERE w.status IN (?, ?, ?)
		AND (r.origin_latitude IS NOT NULL AND r.origin_longitude IS NOT NULL)
		AND (r.dest_latitude IS NOT NULL AND r.dest_longitude IS NOT NULL)
	`, "assigned", "loading", "in_transit").Rows()
	if err != nil {
		return 0, fmt.Errorf("query affected waybills error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wb struct {
			ID          int64
			RoutePlanID int64
		}
		rows.Scan(&wb.ID, &wb.RoutePlanID)
		waybills = append(waybills, wb)
	}

	replanCount := 0
	for _, wb := range waybills {
		affected, _, _ := s.CheckRouteAffected(ctx, wb.RoutePlanID, warningID)
		if affected {
			now := time.Now()
			s.db.WithContext(ctx).Exec(`
				UPDATE route_plans 
				SET status = 'deprecated', updated_at = ? 
				WHERE id = ?
			`, now, wb.RoutePlanID)
			replanCount++
		}
	}

	if replanCount > 0 {
		s.db.WithContext(ctx).Model(&warning).Updates(map[string]interface{}{
			"processed":             1,
			"related_waybill_count": replanCount,
			"updated_at":            time.Now(),
		})
	}

	return replanCount, nil
}
