package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"gorm.io/gorm"
)

type WeatherService struct {
	db         *gorm.DB
	cfg        *config.Config
	amapKey    string
	provider   string
	qweatherKey string
	qweatherURL string
	caiyunKey  string
	caiyunURL  string
	httpClient *http.Client
}

func NewWeatherService(cfg *config.Config) *WeatherService {
	return &WeatherService{
		db:          database.GetDB(),
		cfg:         cfg,
		amapKey:     cfg.Map.AMap.Key,
		provider:    cfg.Weather.Provider,
		qweatherKey: cfg.Weather.QWeather.Key,
		qweatherURL: cfg.Weather.QWeather.BaseURL,
		caiyunKey:   cfg.Weather.CaiYun.Key,
		caiyunURL:   cfg.Weather.CaiYun.BaseURL,
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

type qWeatherNowResponse struct {
	Code    string `json:"code"`
	UpdateTime string `json:"updateTime"`
	Now     struct {
		Temp        string  `json:"temp"`
		FeelsLike   string  `json:"feelsLike"`
		Icon        string  `json:"icon"`
		Text        string  `json:"text"`
		Wind360     string  `json:"wind360"`
		WindDir     string  `json:"windDir"`
		WindScale   string  `json:"windScale"`
		WindSpeed   string  `json:"windSpeed"`
		Humidity    string  `json:"humidity"`
		Precip      string  `json:"precip"`
		Pressure    string  `json:"pressure"`
		Vis         string  `json:"vis"`
		Cloud       string  `json:"cloud"`
		Dew         string  `json:"dew"`
	} `json:"now"`
}

type qWeatherWarningResponse struct {
	Code   string `json:"code"`
	UpdateTime string `json:"updateTime"`
	Warning []struct {
		ID           string `json:"id"`
		Sender       string `json:"sender"`
		PubTime      string `json:"pubTime"`
		Title        string `json:"title"`
		StartTime    string `json:"startTime"`
		EndTime      string `json:"endTime"`
		Status       string `json:"status"`
		Level        string `json:"level"`
		LevelType    string `json:"levelType"`
		Type         string `json:"type"`
		TypeName     string `json:"typeName"`
		Text         string `json:"text"`
		Related      string `json:"related"`
	} `json:"warning"`
}

type qWeatherHistoricalResponse struct {
	Code string `json:"code"`
	WeatherDaily []struct {
		Date       string `json:"fxDate"`
		TempMax    string `json:"tempMax"`
		TempMin    string `json:"tempMin"`
		Humidity   string `json:"humidity"`
		Precip     string `json:"precip"`
		WindDir    string `json:"windDir"`
		WindScale  string `json:"windScale"`
		WindSpeed  string `json:"windSpeed"`
	} `json:"weatherDaily"`
}

type caiyunRealtimeResponse struct {
	Status      string `json:"status"`
	Lang        string `json:"lang"`
	ServerTime  int64  `json:"server_time"`
	TZShift     int    `json:"tzshift"`
	Location    []float64 `json:"location"`
	Unit        string `json:"unit"`
	Result      struct {
		Realtime struct {
			Temperature       float64 `json:"temperature"`
			Humidity          float64 `json:"humidity"`
			WindDirection     float64 `json:"wind_direction"`
			WindSpeed         float64 `json:"wind_speed"`
			Precipitation     struct {
				Datasource string  `json:"datasource"`
				Intensity  float64 `json:"intensity"`
			} `json:"precipitation"`
			Visibility        float64 `json:"visibility"`
			Skycon            string  `json:"skycon"`
			Skycon08H20       string  `json:"skycon_08h_20h"`
			Skycon20H32       string  `json:"skycon_20h_32h"`
			LifeIndex         struct {
				Ultraviolet struct {
					Index float64 `json:"index"`
					Desc  string  `json:"desc"`
				} `json:"ultraviolet"`
			} `json:"life_index"`
			Pressure          float64 `json:"pressure"`
			ApparentTemperature float64 `json:"apparent_temperature"`
		} `json:"realtime"`
		Alert struct {
			Status string        `json:"status"`
			Alerts []interface{} `json:"alerts"`
		} `json:"alert"`
	} `json:"result"`
}

func (s *WeatherService) GetCurrentWeather(ctx context.Context, lat, lng float64) (*model.WeatherData, error) {
	if lat == 0 || lng == 0 {
		return nil, fmt.Errorf("invalid coordinates")
	}

	if s.provider == "qweather" && s.qweatherKey != "" && s.qweatherKey != "your-qweather-api-key" {
		return s.getQWeatherCurrent(ctx, lat, lng)
	}
	if s.provider == "caiyun" && s.caiyunKey != "" && s.caiyunKey != "your-caiyun-api-key" {
		return s.getCaiyunCurrent(ctx, lat, lng)
	}
	return s.getAMapWeather(ctx, lat, lng)
}

func (s *WeatherService) getQWeatherCurrent(ctx context.Context, lat, lng float64) (*model.WeatherData, error) {
	location := fmt.Sprintf("%.4f,%.4f", lng, lat)
	url := fmt.Sprintf("%s/weather/now?location=%s&key=%s", s.qweatherURL, location, s.qweatherKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Sugar.Warnf("create qweather request error: %v, fallback to amap", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Sugar.Warnf("call qweather api error: %v, fallback to amap", err)
		return s.getAMapWeather(ctx, lat, lng)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar.Warnf("read qweather response error: %v", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	var qResp qWeatherNowResponse
	if err := json.Unmarshal(body, &qResp); err != nil {
		logger.Sugar.Warnf("parse qweather response error: %v", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	if qResp.Code != "200" {
		logger.Sugar.Warnf("qweather api returned error code: %s", qResp.Code)
		return s.getAMapWeather(ctx, lat, lng)
	}

	now := qResp.Now
	temp, _ := strconv.ParseFloat(now.Temp, 64)
	feelsLike, _ := strconv.ParseFloat(now.FeelsLike, 64)
	humidity, _ := strconv.ParseFloat(now.Humidity, 64)
	windSpeed, _ := strconv.ParseFloat(now.WindSpeed, 64)
	visibility, _ := strconv.ParseFloat(now.Vis, 64)
	precip, _ := strconv.ParseFloat(now.Precip, 64)
	pressure, _ := strconv.ParseFloat(now.Pressure, 64)

	result := &model.WeatherData{
		Temp:          temp,
		FeelsLike:     feelsLike,
		Humidity:      humidity,
		WindSpeed:     windSpeed,
		WindDirection: now.WindDir,
		Visibility:    visibility * 1000,
		Condition:     now.Text,
		Precipitation: precip,
		Pressure:      pressure,
		RoadSlippery:  precip >= s.cfg.Weather.SlipperyRainMm,
	}

	if humidity > 90 && (strings.Contains(now.Text, "雾") || strings.Contains(now.Text, "霾")) {
		result.RoadSlippery = true
	}
	if temp <= 0 && precip > 0 {
		result.RoadSlippery = true
	}

	return result, nil
}

func (s *WeatherService) getCaiyunCurrent(ctx context.Context, lat, lng float64) (*model.WeatherData, error) {
	url := fmt.Sprintf("%s/%s/%.4f,%.4f/realtime?alert=true", s.caiyunURL, s.caiyunKey, lng, lat)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Sugar.Warnf("create caiyun request error: %v, fallback to amap", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Sugar.Warnf("call caiyun api error: %v, fallback to amap", err)
		return s.getAMapWeather(ctx, lat, lng)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar.Warnf("read caiyun response error: %v", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	var cResp caiyunRealtimeResponse
	if err := json.Unmarshal(body, &cResp); err != nil {
		logger.Sugar.Warnf("parse caiyun response error: %v", err)
		return s.getAMapWeather(ctx, lat, lng)
	}

	if cResp.Status != "ok" {
		logger.Sugar.Warnf("caiyun api returned not ok")
		return s.getAMapWeather(ctx, lat, lng)
	}

	rt := cResp.Result.Realtime
	result := &model.WeatherData{
		Temp:          rt.Temperature,
		FeelsLike:     rt.ApparentTemperature,
		Humidity:      rt.Humidity * 100,
		WindSpeed:     rt.WindSpeed,
		WindDirection: fmt.Sprintf("%.0f度", rt.WindDirection),
		Visibility:    rt.Visibility * 1000,
		Condition:     mapSkycon(rt.Skycon),
		Precipitation: rt.Precipitation.Intensity,
		Pressure:      rt.Pressure,
		UvIndex:       int(rt.LifeIndex.Ultraviolet.Index),
		RoadSlippery:  rt.Precipitation.Intensity >= s.cfg.Weather.SlipperyRainMm,
	}

	if rt.Humidity > 0.9 && (rt.Skycon == "FOG" || rt.Skycon == "HAZE") {
		result.RoadSlippery = true
	}
	if rt.Temperature <= 0 && rt.Precipitation.Intensity > 0 {
		result.RoadSlippery = true
	}

	return result, nil
}

func (s *WeatherService) getAMapWeather(ctx context.Context, lat, lng float64) (*model.WeatherData, error) {
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

	result := &model.WeatherData{
		Temp:          temp,
		Humidity:      humidity,
		WindSpeed:     windSpeed,
		WindDirection: life.WindDirection,
		Visibility:    life.Visibility,
		Condition:     life.Weather,
		RoadSlippery:  strings.Contains(life.Weather, "雨") || strings.Contains(life.Weather, "雪") || strings.Contains(life.Weather, "雾"),
	}

	return result, nil
}

func mapSkycon(skycon string) string {
	mapping := map[string]string{
		"CLEAR_DAY": "晴",
		"CLEAR_NIGHT": "晴",
		"PARTLY_CLOUDY_DAY": "多云",
		"PARTLY_CLOUDY_NIGHT": "多云",
		"CLOUDY": "阴",
		"LIGHT_HAZE": "轻度雾霾",
		"MODERATE_HAZE": "中度雾霾",
		"HEAVY_HAZE": "重度雾霾",
		"HAZE": "霾",
		"FOG": "雾",
		"LIGHT_RAIN": "小雨",
		"MODERATE_RAIN": "中雨",
		"HEAVY_RAIN": "大雨",
		"STORM_RAIN": "暴雨",
		"RAIN": "雨",
		"LIGHT_SNOW": "小雪",
		"MODERATE_SNOW": "中雪",
		"HEAVY_SNOW": "大雪",
		"STORM_SNOW": "暴雪",
		"SNOW": "雪",
		"WIND": "大风",
		"HAIL": "冰雹",
		"THUNDER_SHOWER": "雷阵雨",
	}
	if v, ok := mapping[skycon]; ok {
		return v
	}
	return skycon
}

func mapQWeatherWarningType(typeName string) model.WeatherWarningType {
	mapping := map[string]model.WeatherWarningType{
		"暴雨": model.WarningTypeRainstorm,
		"台风": model.WarningTypeTyphoon,
		"暴雪": model.WarningTypeSnowstorm,
		"大雾": model.WarningTypeFog,
		"霾":  model.WarningTypeHaze,
		"雷电": model.WarningTypeThunder,
		"高温": model.WarningTypeHighTemp,
		"寒潮": model.WarningTypeLowTemp,
		"低温": model.WarningTypeLowTemp,
		"大风": model.WarningTypeStrongWind,
		"沙尘暴": model.WarningTypeSandstorm,
		"冰雹": model.WarningTypeHail,
		"道路结冰": model.WarningTypeIcyRoad,
		"冰冻": model.WarningTypeIcyRoad,
	}
	for k, v := range mapping {
		if strings.Contains(typeName, k) {
			return v
		}
	}
	return model.WarningTypeRainstorm
}

func mapQWeatherLevel(level, levelType string) model.WarningLevel {
	level = strings.TrimSpace(level)
	levelType = strings.TrimSpace(levelType)

	if strings.Contains(level, "红") || strings.Contains(levelType, "红") {
		return model.WarningLevelRed
	}
	if strings.Contains(level, "橙") || strings.Contains(levelType, "橙") {
		return model.WarningLevelOrange
	}
	if strings.Contains(level, "黄") || strings.Contains(levelType, "黄") {
		return model.WarningLevelYellow
	}
	if strings.Contains(level, "蓝") || strings.Contains(levelType, "蓝") {
		return model.WarningLevelBlue
	}
	return model.WarningLevelBlue
}

func (s *WeatherService) SyncWarningsFromAPI(ctx context.Context, province string) (int, error) {
	if s.provider != "qweather" || s.qweatherKey == "" || s.qweatherKey == "your-qweather-api-key" {
		return 0, nil
	}

	url := fmt.Sprintf("%s/warning/now?range=cn&key=%s", s.qweatherURL, s.qweatherKey)
	if province != "" {
		url = fmt.Sprintf("%s/warning/now?range=%s&key=%s", s.qweatherURL, province, s.qweatherKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request error: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("call api error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response error: %w", err)
	}

	var qResp qWeatherWarningResponse
	if err := json.Unmarshal(body, &qResp); err != nil {
		return 0, fmt.Errorf("parse response error: %w", err)
	}

	if qResp.Code != "200" {
		return 0, fmt.Errorf("api error code: %s", qResp.Code)
	}

	syncedCount := 0
	now := time.Now()

	for _, w := range qResp.Warning {
		startTime, _ := time.Parse(time.RFC3339, w.StartTime)
		endTime, _ := time.Parse(time.RFC3339, w.EndTime)
		pubTime, _ := time.Parse(time.RFC3339, w.PubTime)

		if startTime.IsZero() {
			startTime = pubTime
		}
		if pubTime.IsZero() {
			pubTime = now
		}

		warningType := mapQWeatherWarningType(w.TypeName)
		warningLevel := mapQWeatherLevel(w.Level, w.LevelType)

		var existing model.WeatherWarning
		err := s.db.WithContext(ctx).Where("warning_id = ?", w.ID).First(&existing).Error

		suggestion := s.generateWarningSuggestion(warningType, warningLevel, w.Text)
		speedSuggestion := s.calculateSuggestedSpeed(warningType, warningLevel)
		shouldStopOperation := s.shouldTriggerOperationStop(warningLevel, 0, 0)

		warningModel := model.WeatherWarning{
			WarningID:           w.ID,
			WarningType:         warningType,
			WarningLevel:        warningLevel,
			Title:               w.Title,
			Content:             w.Text,
			StartTime:           startTime,
			PublishTime:         pubTime,
			Source:              "qweather",
			Processed:           0,
			TriggerOperationStop: boolToInt(shouldStopOperation),
			SpeedSuggestion:     speedSuggestion,
			Suggestion:          suggestion,
		}
		if !endTime.IsZero() {
			warningModel.EndTime = &endTime
		}

		if err == gorm.ErrRecordNotFound {
			if err := s.db.WithContext(ctx).Create(&warningModel).Error; err != nil {
				logger.Sugar.Warnf("create warning error: %v, id=%s", err, w.ID)
				continue
			}
			syncedCount++
		} else if err == nil {
			updates := map[string]interface{}{
				"warning_type":          warningType,
				"warning_level":         warningLevel,
				"title":                 w.Title,
				"content":               w.Text,
				"start_time":            startTime,
				"publish_time":          pubTime,
				"processed":             0,
				"trigger_operation_stop": boolToInt(shouldStopOperation),
				"speed_suggestion_kmh":  speedSuggestion,
				"suggestion":            suggestion,
				"updated_at":            now,
			}
			if !endTime.IsZero() {
				updates["end_time"] = &endTime
			}
			s.db.WithContext(ctx).Model(&existing).Updates(updates)
		}
	}

	if syncedCount > 0 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Sugar.Errorf("auto suspend after sync panic: %v", r)
				}
			}()
			_, _, _ = s.CheckAndAutoSuspend(context.Background())
		}()
	}

	return syncedCount, nil
}

func (s *WeatherService) generateWarningSuggestion(wType model.WeatherWarningType, level model.WarningLevel, content string) string {
	var suggestion strings.Builder

	if level == model.WarningLevelRed || level == model.WarningLevelOrange {
		suggestion.WriteString("【紧急】")
	}

	switch wType {
	case model.WarningTypeRainstorm:
		suggestion.WriteString("强降雨天气，")
		if level == model.WarningLevelRed {
			suggestion.WriteString("建议立即停止危化品运输作业，就近服务区避险。")
		} else if level == model.WarningLevelOrange {
			suggestion.WriteString("建议暂停新任务派发，在途车辆减速慢行，注意路面积水。")
		} else {
			suggestion.WriteString("驾驶员注意开启雨刷，保持安全车距，减速行驶。")
		}
	case model.WarningTypeFog, model.WarningTypeHaze:
		suggestion.WriteString("低能见度天气，")
		if level == model.WarningLevelRed {
			suggestion.WriteString("能见度极低，必须立即停止运营，车辆就近停靠安全区域。")
		} else if level == model.WarningLevelOrange {
			suggestion.WriteString("能见度不足，建议开启雾灯和双闪，限速40km/h以下。")
		} else {
			suggestion.WriteString("注意开启雾灯，谨慎驾驶，与前车保持更大距离。")
		}
	case model.WarningTypeStrongWind:
		suggestion.WriteString("大风天气，")
		if level == model.WarningLevelRed || level == model.WarningLevelOrange {
			suggestion.WriteString("空车及罐装车注意横风，避开桥梁、高架等路段。")
		} else {
			suggestion.WriteString("注意横风影响，握稳方向盘，适当降速。")
		}
	case model.WarningTypeIcyRoad, model.WarningTypeSlippery:
		suggestion.WriteString("路面易滑，")
		suggestion.WriteString("请减速慢行，避免紧急制动和急打方向，建议安装防滑链。")
	case model.WarningTypeSnowstorm:
		suggestion.WriteString("暴雪天气，")
		if level == model.WarningLevelRed {
			suggestion.WriteString("建议暂停运营，路面结冰风险极大。")
		} else {
			suggestion.WriteString("注意路面积雪结冰，安装防滑链，低速行驶。")
		}
	case model.WarningTypeTyphoon:
		suggestion.WriteString("台风天气，")
		suggestion.WriteString("建议所有危化品车辆立即停运，寻找坚固停车场避险。")
	default:
		suggestion.WriteString("请密切关注天气变化，谨慎驾驶。")
	}

	if content != "" {
		content = strings.TrimSpace(content)
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		suggestion.WriteString(" 详情: ")
		suggestion.WriteString(content)
	}

	return suggestion.String()
}

func (s *WeatherService) calculateSuggestedSpeed(wType model.WeatherWarningType, level model.WarningLevel) int {
	baseSpeed := 80

	switch level {
	case model.WarningLevelRed:
		baseSpeed = 0
	case model.WarningLevelOrange:
		baseSpeed = 40
	case model.WarningLevelYellow:
		baseSpeed = 60
	case model.WarningLevelBlue:
		baseSpeed = 70
	}

	switch wType {
	case model.WarningTypeFog, model.WarningTypeHaze:
		baseSpeed = int(float64(baseSpeed) * 0.6)
	case model.WarningTypeIcyRoad, model.WarningTypeSlippery, model.WarningTypeSnowstorm:
		baseSpeed = int(float64(baseSpeed) * 0.5)
	case model.WarningTypeTyphoon, model.WarningTypeStrongWind:
		baseSpeed = int(float64(baseSpeed) * 0.7)
	}

	if baseSpeed < 0 {
		baseSpeed = 0
	}
	return baseSpeed
}

func (s *WeatherService) shouldTriggerOperationStop(level model.WarningLevel, visibilityM float64, windSpeedMs float64) bool {
	if level == model.WarningLevelRed {
		return true
	}
	if visibilityM > 0 && visibilityM <= s.cfg.Weather.ExtremeVisibility {
		return true
	}
	if windSpeedMs > 0 && windSpeedMs >= s.cfg.Weather.ExtremeWindSpeed {
		return true
	}
	return false
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
	if len(routePlan.RoutePath) > 50 {
		sampleInterval = len(routePlan.RoutePath) / 10
	}

	activeWarnings, _ := s.GetActiveWarnings(ctx, "")

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

		speedSuggestion := s.calculateRoadSpeedSuggestion(weather)
		hasWarning, wType, wLevel := s.checkPointWarning(point.Lat, point.Lng, activeWarnings)

		result = append(result, &model.RouteWeatherPoint{
			Lat:               point.Lat,
			Lng:               point.Lng,
			WeatherData:       *weather,
			DistanceFromStart: distanceFromStart,
			EstimatedTime:     estimatedTime,
			SpeedSuggestion:   speedSuggestion,
			HasWarning:        hasWarning,
			WarningType:       wType,
			WarningLevel:      wLevel,
		})
	}

	return result, nil
}

func (s *WeatherService) calculateRoadSpeedSuggestion(weather *model.WeatherData) int {
	suggestedSpeed := 80

	if weather.RoadSlippery {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.6)
	}

	if weather.Precipitation > 5 {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.5)
	} else if weather.Precipitation > 2.5 {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.7)
	}

	if weather.Visibility > 0 {
		if weather.Visibility < 50 {
			suggestedSpeed = 0
		} else if weather.Visibility < 100 {
			suggestedSpeed = 20
		} else if weather.Visibility < 200 {
			suggestedSpeed = 40
		} else if weather.Visibility < 500 {
			suggestedSpeed = 60
		}
	}

	if weather.WindSpeed > 25 {
		suggestedSpeed = 0
	} else if weather.WindSpeed > 17 {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.7)
	} else if weather.WindSpeed > 10 {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.85)
	}

	if weather.Temp <= 0 && weather.Precipitation > 0 {
		suggestedSpeed = int(float64(suggestedSpeed) * 0.5)
	}

	if suggestedSpeed < 0 {
		suggestedSpeed = 0
	}
	return suggestedSpeed
}

func (s *WeatherService) checkPointWarning(lat, lng float64, warnings []*model.WeatherWarning) (bool, string, string) {
	for _, w := range warnings {
		if w.CenterLat == 0 && w.CenterLng == 0 {
			continue
		}
		distanceKm := haversine(lat, lng, w.CenterLat, w.CenterLng) / 1000
		if distanceKm < 100 {
			return true, string(w.WarningType), string(w.WarningLevel)
		}
	}
	return false, "", ""
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *WeatherService) AnalyzeRouteWeather(ctx context.Context, routeID int64) (*model.RouteWeatherAnalysis, error) {
	if routeID <= 0 {
		return nil, fmt.Errorf("invalid route id")
	}

	weatherPoints, err := s.GetRouteWeather(ctx, routeID)
	if err != nil {
		return nil, err
	}

	if len(weatherPoints) == 0 {
		return &model.RouteWeatherAnalysis{
			RouteID:           routeID,
			OverallRiskLevel:  "low",
			SafeSpeed:         80,
			WeatherPoints:     []*model.RouteWeatherPoint{},
			SegmentWarnings:   []*model.SegmentWarning{},
			WarningsOnRoute:   []*model.WeatherWarning{},
			Suggestions:       []string{"路线天气状况良好，可正常行驶。"},
			ShouldDetour:      false,
			AnalyzedAt:        time.Now(),
		}, nil
	}

	var routePlan model.RoutePlan
	s.db.WithContext(ctx).Where("id = ?", routeID).First(&routePlan)

	analysis := &model.RouteWeatherAnalysis{
		RouteID:           routeID,
		WaybillID:         routePlan.WaybillID,
		TotalDistance:     routePlan.TotalDistance / 1000,
		WeatherPoints:     weatherPoints,
		SegmentWarnings:   []*model.SegmentWarning{},
		WarningsOnRoute:   []*model.WeatherWarning{},
		Suggestions:       []string{},
		AnalyzedAt:        time.Now(),
	}

	var totalVisibility float64
	var visibilityCount int
	minSpeed := 120
	hasExtreme := false
	hasHighRisk := false
	seenWarnings := make(map[string]bool)

	for i, point := range weatherPoints {
		if point.WeatherData.Visibility > 0 {
			totalVisibility += point.WeatherData.Visibility
			visibilityCount++
		}
		if point.SpeedSuggestion < minSpeed {
			minSpeed = point.SpeedSuggestion
		}
		if point.SpeedSuggestion == 0 {
			hasExtreme = true
		}

		sw := &model.SegmentWarning{
			SegmentIndex:    i,
			StartLat:        point.Lat,
			StartLng:        point.Lng,
			EndLat:          point.Lat,
			EndLng:          point.Lng,
			Distance:        point.DistanceFromStart / 1000,
			SpeedSuggestion: point.SpeedSuggestion,
			EstimatedETA:    point.EstimatedTime,
		}

		condition := point.WeatherData.Condition
		if point.HasWarning {
			sw.WarningType = point.WarningType
			sw.WarningLevel = point.WarningLevel
			sw.Description = fmt.Sprintf("路段天气预警: %s-%s", point.WarningType, point.WarningLevel)
			sw.DetourSuggested = point.WarningLevel == "red" || point.WarningLevel == "orange"
			analysis.SegmentWarnings = append(analysis.SegmentWarnings, sw)
			if sw.DetourSuggested {
				hasHighRisk = true
			}
			warnKey := fmt.Sprintf("%s-%s", point.WarningType, point.WarningLevel)
			if !seenWarnings[warnKey] {
				seenWarnings[warnKey] = true
				analysis.WarningsOnRoute = append(analysis.WarningsOnRoute, &model.WeatherWarning{
					WarningType:  model.WarningType(point.WarningType),
					WarningLevel: model.WarningLevel(point.WarningLevel),
					Title:        fmt.Sprintf("%s预警", point.WarningType),
				})
			}
		}

		if point.WeatherData.Precipitation > 0 || strings.Contains(condition, "雨") {
			if sw.WarningType == "" {
				sw.WarningType = "rain"
				sw.Description = fmt.Sprintf("降雨路段，降雨量%.1fmm", point.WeatherData.Precipitation)
				analysis.SegmentWarnings = append(analysis.SegmentWarnings, sw)
			}
		}

		if point.WeatherData.RoadSlippery {
			if sw.WarningType == "" {
				sw.WarningType = "slippery"
				sw.Description = "路面湿滑，请减速慢行"
				analysis.SegmentWarnings = append(analysis.SegmentWarnings, sw)
			}
		}

		if strings.Contains(condition, "雾") || strings.Contains(condition, "霾") ||
			(point.WeatherData.Visibility > 0 && point.WeatherData.Visibility < 1000) {
			if sw.WarningType == "" {
				sw.WarningType = "fog"
				sw.Description = fmt.Sprintf("低能见度路段，能见度%.0fm", point.WeatherData.Visibility)
				if point.WeatherData.Visibility < 200 {
					sw.DetourSuggested = true
					hasHighRisk = true
				}
				analysis.SegmentWarnings = append(analysis.SegmentWarnings, sw)
			}
		}

		if point.WeatherData.WindSpeed > 10 {
			if sw.WarningType == "" {
				sw.WarningType = "wind"
				sw.Description = fmt.Sprintf("大风路段，风速%.1fm/s", point.WeatherData.WindSpeed)
				if point.WeatherData.WindSpeed > 17 {
					sw.DetourSuggested = true
					hasHighRisk = true
				}
				analysis.SegmentWarnings = append(analysis.SegmentWarnings, sw)
			}
		}
	}

	analysis.SafeSpeed = minSpeed
	analysis.HasExtremeWeather = hasExtreme
	analysis.ShouldDetour = hasExtreme || hasHighRisk

	switch {
	case hasExtreme:
		analysis.OverallRiskLevel = "extreme"
	case len(analysis.SegmentWarnings) > 3:
		analysis.OverallRiskLevel = "high"
	case len(analysis.SegmentWarnings) > 0:
		analysis.OverallRiskLevel = "medium"
	default:
		analysis.OverallRiskLevel = "low"
	}

	switch analysis.OverallRiskLevel {
	case "extreme":
		analysis.Suggestions = []string{
			"路线存在极端天气，能见度极低或风力过大，强烈建议暂停运营。",
			"车辆立即就近停靠安全区域，等待天气好转。",
			"通知调度中心，评估是否需要改派或取消运输任务。",
		}
		analysis.DetourSuggestion = "建议立即暂停所有相关运输任务，车辆就近停靠安全区域。"
	case "high":
		analysis.Suggestions = []string{
			"路线存在多处恶劣天气路段，高风险运营。",
			"建议重新规划路线或推迟出发。",
			"提醒驾驶员谨慎驾驶，保持安全车距。",
		}
		analysis.DetourSuggestion = "建议触发路径重规划，避开高风险天气区域。"
	case "medium":
		analysis.Suggestions = []string{
			"路线存在部分恶劣天气路段，中风险运营。",
			"需提醒驾驶员谨慎驾驶，注意路况变化。",
			"保持安全车速，随时关注天气变化。",
		}
	default:
		analysis.Suggestions = []string{
			"路线天气状况良好，可正常行驶。",
			"注意常规安全驾驶，遵守交通规则。",
		}
	}

	return analysis, nil
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

func (s *WeatherService) CheckAndAutoSuspend(ctx context.Context) (bool, *model.OperationSuspension, error) {
	currentSuspension, err := s.GetCurrentSuspension(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("check current suspension error: %w", err)
	}
	if currentSuspension != nil {
		return false, currentSuspension, nil
	}

	activeWarnings, err := s.GetActiveWarnings(ctx, "")
	if err != nil {
		return false, nil, fmt.Errorf("get active warnings error: %w", err)
	}

	var targetWarning *model.WeatherWarning
	var minVisibility float64
	var maxWindSpeed float64

	for _, w := range activeWarnings {
		triggerStop := false

		if w.WarningLevel == model.WarningLevelRed {
			triggerStop = true
		}

		if w.WarningType == model.WarningTypeFog && w.SpeedSuggestion == 0 {
			triggerStop = true
		}

		if triggerStop {
			targetWarning = w
			break
		}
	}

	if targetWarning == nil {
		return false, nil, nil
	}

	extremeVisibility := s.cfg.Weather.ExtremeVisibility
	if extremeVisibility <= 0 {
		extremeVisibility = 50
	}

	weatherType := string(targetWarning.WarningType)
	visibility := 0.0
	windSpeed := 0.0

	switch targetWarning.WarningType {
	case model.WarningTypeFog:
		visibility = extremeVisibility / 2
	case model.WarningTypeRainstorm:
		visibility = 500
	case model.WarningTypeStrongWind:
		windSpeed = s.cfg.Weather.ExtremeWindSpeed
		if windSpeed <= 0 {
			windSpeed = 25
		}
	}

	affectedRegion := targetWarning.Title
	if len(targetWarning.AffectedCities) > 0 {
		affectedRegion = fmt.Sprintf("%s及周边地区", targetWarning.AffectedCities[0])
	}

	radiusKm := 50.0
	if targetWarning.CenterLat != 0 {
		radiusKm = 100
	}

	centerLat := targetWarning.CenterLat
	centerLng := targetWarning.CenterLng
	if centerLat == 0 && centerLng == 0 {
		centerLat = 30.0
		centerLng = 114.0
	}

	triggerWarningID := targetWarning.ID

	suspendReq := &model.OperationSuspendRequest{
		TriggerType:     "automatic",
		TriggerReason:   fmt.Sprintf("系统检测到%s-%s预警，自动触发运营暂停", targetWarning.WarningType, targetWarning.WarningLevel),
		TriggerWarningID: &triggerWarningID,
		WeatherType:     weatherType,
		Visibility:      visibility,
		WindSpeed:       windSpeed,
		AffectedRegion:  affectedRegion,
		CenterLat:       centerLat,
		CenterLng:       centerLng,
		RadiusKm:        radiusKm,
		SuggestedSpeed:  0,
		AutoTriggered:   true,
		Remark:          "系统自动触发：极端天气条件下自动暂停运营",
	}

	suspension, err := s.TriggerOperationSuspend(ctx, suspendReq, 0, "system")
	if err != nil {
		return false, nil, fmt.Errorf("auto suspend error: %w", err)
	}

	go s.sendSuspensionNotifications(context.Background(), suspension)

	logger.Sugar.Warnf("auto operation suspended: no=%s, warning_type=%s, visibility=%.1fm, wind_speed=%.1fm/s",
		suspension.SuspensionNo, weatherType, visibility, windSpeed)

	return true, suspension, nil
}

func (s *WeatherService) sendSuspensionNotifications(ctx context.Context, suspension *model.OperationSuspension) {
	defer func() {
		if r := recover(); r != nil {
			logger.Sugar.Errorf("send suspension notifications panic: %v", r)
		}
	}()

	pushReq := &model.WeatherPushRequest{
		Phase:      string(model.PushPhaseEmergency),
		TargetType: "all",
		Title:      fmt.Sprintf("【紧急停运通知】%s", suspension.TriggerReason),
		Content: fmt.Sprintf("因%s，现暂停%s范围内所有危化品车辆运营。\n"+
			"建议措施：\n"+
			"1. 在途车辆请立即就近停靠安全区域\n"+
			"2. 开启雾灯和双闪警示灯\n"+
			"3. 等待天气好转后恢复运营\n"+
			"4. 如有紧急情况请联系调度中心",
			suspension.TriggerReason, suspension.AffectedRegion),
	}

	_, err := s.PushWeatherWarning(ctx, pushReq)
	if err != nil {
		logger.Sugar.Errorf("push suspension warning error: %v", err)
	}

	logger.Sugar.Infof("suspension notification sent: suspension_no=%s", suspension.SuspensionNo)
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

func (s *WeatherService) GetHistoricalWeather(ctx context.Context, query *model.HistoricalWeatherQuery) ([]*model.HistoricalWeather, error) {
	var startTime, endTime time.Time
	var err error

	if query.StartTime != "" && query.EndTime != "" {
		startTime, err = time.Parse(time.RFC3339, query.StartTime)
		if err != nil {
			startTime, err = time.Parse("2006-01-02 15:04:05", query.StartTime)
			if err != nil {
				return nil, fmt.Errorf("invalid start time format: %w", err)
			}
		}
		endTime, err = time.Parse(time.RFC3339, query.EndTime)
		if err != nil {
			endTime, err = time.Parse("2006-01-02 15:04:05", query.EndTime)
			if err != nil {
				return nil, fmt.Errorf("invalid end time format: %w", err)
			}
		}
	} else if query.QueryTime != "" {
		queryTime, err := time.Parse(time.RFC3339, query.QueryTime)
		if err != nil {
			queryTime, err = time.Parse("2006-01-02 15:04:05", query.QueryTime)
			if err != nil {
				return nil, fmt.Errorf("invalid query time format: %w", err)
			}
		}
		startTime = queryTime.Add(-1 * time.Hour)
		endTime = queryTime.Add(1 * time.Hour)
	} else {
		return nil, fmt.Errorf("query_time or start_time/end_time is required")
	}

	dbQuery := s.db.WithContext(ctx).Model(&model.HistoricalWeather{}).
		Where("query_time >= ? AND query_time < ?", startTime, endTime)

	if query.LocationName != "" {
		dbQuery = dbQuery.Where("location_name LIKE ?", "%"+query.LocationName+"%")
	}
	if query.Latitude != 0 && query.Longitude != 0 {
		dbQuery = dbQuery.Where(
			"ABS(latitude - ?) < 0.5 AND ABS(longitude - ?) < 0.5",
			query.Latitude, query.Longitude,
		)
	}

	var results []*model.HistoricalWeather
	if err := dbQuery.Order("query_time ASC").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("query historical weather error: %w", err)
	}

	if len(results) > 0 {
		return results, nil
	}

	return s.fetchAndStoreHistorical(ctx, query, startTime, endTime)
}

func (s *WeatherService) fetchAndStoreHistorical(ctx context.Context, query *model.HistoricalWeatherQuery, startTime, endTime time.Time) ([]*model.HistoricalWeather, error) {
	if s.provider != "qweather" || s.qweatherKey == "" || s.qweatherKey == "your-qweather-api-key" {
		return s.generateMockHistorical(query, startTime, endTime), nil
	}

	var results []*model.HistoricalWeather
	location := fmt.Sprintf("%.4f,%.4f", query.Longitude, query.Latitude)

	step := time.Hour
	for t := startTime; t.Before(endTime); t = t.Add(step) {
		dateStr := t.Format("20060102")
		url := fmt.Sprintf("%s/historical/weather?location=%s&date=%s&key=%s",
			s.qweatherURL, location, dateStr, s.qweatherKey)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var hResp qWeatherHistoricalResponse
		if err := json.Unmarshal(body, &hResp); err != nil || hResp.Code != "200" {
			continue
		}

		for _, day := range hResp.WeatherDaily {
			tempMax, _ := strconv.ParseFloat(day.TempMax, 64)
			tempMin, _ := strconv.ParseFloat(day.TempMin, 64)
			humidity, _ := strconv.ParseFloat(day.Humidity, 64)
			precip, _ := strconv.ParseFloat(day.Precip, 64)
			windSpeed, _ := strconv.ParseFloat(day.WindSpeed, 64)
			windDir, _ := strconv.Atoi(day.WindDir)
			visibility := 10000.0
			if day.Visibility != "" {
				visibility, _ = strconv.ParseFloat(day.Visibility, 64)
				visibility *= 1000
			}

			avgTemp := (tempMax + tempMin) / 2
			weatherCondition := day.TextDay
			if weatherCondition == "" {
				weatherCondition = "晴"
			}

			hw := &model.HistoricalWeather{
				Latitude:         query.Latitude,
				Longitude:        query.Longitude,
				LocationName:     query.LocationName,
				QueryTime:        t,
				WeatherCondition: weatherCondition,
				Temperature:      avgTemp,
				FeelsLike:        avgTemp,
				Humidity:         humidity,
				WindSpeed:        windSpeed,
				WindDirection:    windDir,
				Visibility:       visibility,
				Precipitation:    precip,
				PrecipType:       "rain",
				RoadSlippery:     precip >= s.cfg.Weather.SlipperyRainMm || tempMin <= 0,
				Warnings:         model.JSON("[]"),
				DataSource:       "qweather",
			}
			if precip > 0 {
				hw.PrecipType = "rain"
				hw.RoadCondition = "wet"
			} else if tempMin <= 0 {
				hw.RoadCondition = "icy"
			} else {
				hw.RoadCondition = "dry"
			}

			s.db.WithContext(ctx).Create(hw)
			results = append(results, hw)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if len(results) == 0 {
		return s.generateMockHistorical(query, startTime, endTime), nil
	}

	return results, nil
}

func (s *WeatherService) generateMockHistorical(query *model.HistoricalWeatherQuery, startTime, endTime time.Time) []*model.HistoricalWeather {
	var results []*model.HistoricalWeather
	step := time.Hour * 3
	for t := startTime; t.Before(endTime); t = t.Add(step) {
		seed := float64(t.Unix() / 3600)
		temp := 15 + math.Sin(seed/24)*10
		precip := math.Max(0, math.Sin(seed/7)*5)
		humidity := 50 + math.Sin(seed/5)*30
		windSpeed := 2 + math.Abs(math.Sin(seed/11))*8
		windDir := int(math.Mod(seed, 360))
		visibility := 8000 + math.Sin(seed/3)*3000
		feelsLike := temp + math.Sin(seed/4)*2

		weatherCondition := "晴"
		roadCondition := "dry"
		roadSlippery := false

		switch {
		case precip > 10:
			weatherCondition = "暴雨"
			roadCondition = "wet"
			roadSlippery = true
		case precip > 5:
			weatherCondition = "中雨"
			roadCondition = "wet"
			roadSlippery = true
		case precip > 0:
			weatherCondition = "小雨"
			roadCondition = "moist"
		case visibility < 1000:
			weatherCondition = "雾"
		case temp < 0:
			weatherCondition = "晴"
			roadCondition = "icy"
		}

		var warnings []string
		if visibility < 200 {
			warnings = append(warnings, "fog_orange")
		} else if visibility < 500 {
			warnings = append(warnings, "fog_yellow")
		}
		if precip > 8 {
			warnings = append(warnings, "rainstorm_orange")
		} else if precip > 3 {
			warnings = append(warnings, "rainstorm_yellow")
		}
		if windSpeed > 17 {
			warnings = append(warnings, "strong_wind_blue")
		}

		warningsJSON, _ := json.Marshal(warnings)

		hw := &model.HistoricalWeather{
			Latitude:         query.Latitude,
			Longitude:        query.Longitude,
			LocationName:     query.LocationName,
			QueryTime:        t,
			WeatherCondition: weatherCondition,
			Temperature:      temp,
			FeelsLike:        feelsLike,
			Humidity:         humidity,
			WindSpeed:        windSpeed,
			WindDirection:    windDir,
			Visibility:       visibility,
			Precipitation:    precip,
			PrecipType:       "rain",
			RoadCondition:    roadCondition,
			RoadSlippery:     roadSlippery,
			UvIndex:          int(math.Max(0, math.Sin(seed/12)*5)),
			Warnings:         model.JSON(warningsJSON),
			DataSource:       "mock",
		}
		if len(warnings) > 0 {
			hw.WarningType = strings.Split(warnings[0], "_")[0]
			hw.WarningLevel = strings.Split(warnings[0], "_")[1]
		}

		results = append(results, hw)
	}
	return results
}

func (s *WeatherService) PushWeatherWarning(ctx context.Context, req *model.WeatherPushRequest) (*model.WeatherPushRecord, error) {
	var waybillID int64
	var waybillNo, plateNumber, driverName string
	var vehicleID, driverID int64

	if req.WaybillID != nil && *req.WaybillID > 0 {
		waybillID = *req.WaybillID
		var waybill model.Waybill
		if err := s.db.WithContext(ctx).Where("id = ?", waybillID).First(&waybill).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("waybill not found")
			}
			return nil, fmt.Errorf("query waybill error: %w", err)
		}
		waybillNo = waybill.WaybillNo
		vehicleID = waybill.VehicleID
		driverID = waybill.DriverID

		var vehicle model.Vehicle
		s.db.WithContext(ctx).Where("id = ?", vehicleID).First(&vehicle)
		plateNumber = vehicle.PlateNumber

		var driver model.User
		s.db.WithContext(ctx).Where("id = ?", driverID).First(&driver)
		driverName = driver.RealName
	}

	title := req.Title
	content := req.Content
	pushPhase := model.PushPhase(req.Phase)
	if pushPhase == "" {
		pushPhase = model.PushPhaseEnRoute
	}

	var warningID int64
	var warningNo, wType, wLevel string
	var speedSuggestion int

	if req.WarningID != nil && *req.WarningID > 0 {
		warning, err := s.GetWarning(ctx, *req.WarningID)
		if err == nil {
			warningID = warning.ID
			warningNo = warning.WarningID
			wType = string(warning.WarningType)
			wLevel = string(warning.WarningLevel)
			speedSuggestion = warning.SpeedSuggestion
			if title == "" {
				title = fmt.Sprintf("【天气预警】%s %s", warning.WarningType, warning.Title)
			}
			if content == "" {
				content = warning.Suggestion
			}
		}
	}

	if title == "" {
		title = "天气提醒"
		switch pushPhase {
		case model.PushPhasePreDeparture:
			title = "【出发前】路线天气预警提醒"
		case model.PushPhaseEmergency:
			title = "【紧急】极端天气预警"
		default:
			title = "【行驶中】前方路段天气预警"
		}
	}
	if content == "" {
		content = "前方路段有恶劣天气，请谨慎驾驶，注意安全。"
	}

	now := time.Now()
	pushID := fmt.Sprintf("WX-PUSH-%s-%04d", now.Format("20060102-150405"), time.Now().UnixNano()%10000)

	targetType := req.TargetType
	if targetType == "" {
		targetType = "waybill"
	}

	var targetIDs model.JSON
	if len(req.TargetIDs) > 0 {
		data, _ := json.Marshal(req.TargetIDs)
		targetIDs = model.JSON(data)
	}

	record := &model.WeatherPushRecord{
		PushID:          pushID,
		PushPhase:       pushPhase,
		WarningID:       warningID,
		WarningNo:       warningNo,
		WarningType:     wType,
		WarningLevel:    wLevel,
		Title:           title,
		Content:         content,
		TargetType:      targetType,
		TargetIDs:       targetIDs,
		WaybillID:       waybillID,
		WaybillNo:       waybillNo,
		VehicleID:       vehicleID,
		PlateNumber:     plateNumber,
		DriverID:        driverID,
		DriverName:      driverName,
		PushChannels:    model.JSON(`["app","push"]`),
		Status:          "sent",
		SuccessCount:    1,
		FailCount:       0,
		ReadCount:       0,
		ReadStatus:      0,
		SpeedSuggestion: speedSuggestion,
		SegmentStartLat: req.SegmentStartLat,
		SegmentStartLng: req.SegmentStartLng,
		SegmentEndLat:   req.SegmentEndLat,
		SegmentEndLng:   req.SegmentEndLng,
		SentAt:          &now,
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("create push record error: %w", err)
	}

	logger.Sugar.Infof("weather push sent: push_id=%s, phase=%s, target_type=%s, title=%s", pushID, pushPhase, targetType, title)

	return record, nil
}

func (s *WeatherService) ListPushRecords(ctx context.Context, page, pageSize int, waybillID int64, driverID int64, phase string) (*model.WeatherPushPage, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var records []*model.WeatherPushRecord
	var total int64

	query := s.db.WithContext(ctx).Model(&model.WeatherPushRecord{})
	if waybillID > 0 {
		query = query.Where("waybill_id = ?", waybillID)
	}
	if driverID > 0 {
		query = query.Where("driver_id = ?", driverID)
	}
	if phase != "" {
		query = query.Where("push_phase = ?", phase)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count push records error: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("query push records error: %w", err)
	}

	return &model.WeatherPushPage{
		List:     records,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *WeatherService) GetDriverUnreadCount(ctx context.Context, driverID int64) (int64, error) {
	if driverID <= 0 {
		return 0, fmt.Errorf("invalid driver id")
	}

	var count int64
	err := s.db.WithContext(ctx).Model(&model.WeatherPushRecord{}).
		Where("driver_id = ? AND read_status = 0 AND status = ?", driverID, "sent").
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count unread error: %w", err)
	}
	return count, nil
}

func (s *WeatherService) MarkPushRecordRead(ctx context.Context, pushID string, driverID int64) error {
	if pushID == "" {
		return fmt.Errorf("push id is required")
	}

	updates := map[string]interface{}{
		"read_status": 1,
		"read_time":   time.Now(),
	}

	query := s.db.WithContext(ctx).Model(&model.WeatherPushRecord{}).
		Where("push_id = ?", pushID)
	if driverID > 0 {
		query = query.Where("driver_id = ?", driverID)
	}

	result := query.Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("mark read error: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		s.db.WithContext(ctx).Model(&model.WeatherPushRecord{}).
			Where("push_id = ?", pushID).
			UpdateColumn("read_count", gorm.Expr("read_count + 1"))
	}
	return nil
}

func (s *WeatherService) RespondToPushRecord(ctx context.Context, pushID string, driverID int64, response, note string) error {
	if pushID == "" {
		return fmt.Errorf("push id is required")
	}
	if response == "" {
		return fmt.Errorf("response is required")
	}

	updates := map[string]interface{}{
		"driver_response": response,
		"response_time":   time.Now(),
		"read_status":     1,
	}
	if note != "" {
		updates["response_note"] = note
	}

	query := s.db.WithContext(ctx).Model(&model.WeatherPushRecord{}).
		Where("push_id = ?", pushID)
	if driverID > 0 {
		query = query.Where("driver_id = ?", driverID)
	}

	if err := query.Updates(updates).Error; err != nil {
		return fmt.Errorf("response error: %w", err)
	}

	logger.Sugar.Infof("driver responded to push: push_id=%s, driver_id=%d, response=%s", pushID, driverID, response)
	return nil
}

func (s *WeatherService) PreDepartureWarning(ctx context.Context, waybillID int64) (*model.WeatherPushRecord, error) {
	if waybillID <= 0 {
		return nil, fmt.Errorf("invalid waybill id")
	}

	var waybill model.Waybill
	if err := s.db.WithContext(ctx).Where("id = ?", waybillID).First(&waybill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("waybill not found")
		}
		return nil, fmt.Errorf("query waybill error: %w", err)
	}

	if waybill.RoutePlanID <= 0 {
		return nil, fmt.Errorf("waybill has no route plan")
	}

	analysis, err := s.AnalyzeRouteWeather(ctx, waybill.RoutePlanID)
	if err != nil {
		return nil, fmt.Errorf("analyze route weather error: %w", err)
	}

	if analysis.OverallRiskLevel == "low" {
		return nil, nil
	}

	var warningID int64
	var warningType, warningLevel string
	if len(analysis.WarningsOnRoute) > 0 {
		warningID = analysis.WarningsOnRoute[0].ID
		warningType = string(analysis.WarningsOnRoute[0].WarningType)
		warningLevel = string(analysis.WarningsOnRoute[0].WarningLevel)
	}

	riskLevelText := map[string]string{
		"low":     "低风险",
		"medium":  "中风险",
		"high":    "高风险",
		"extreme": "极高风险",
	}[analysis.OverallRiskLevel]

	title := fmt.Sprintf("【出发前提醒】路线天气%s预警", riskLevelText)

	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("您的运单 %s 出发路线天气评估：%s\n\n", waybill.WaybillNo, riskLevelText))
	contentBuilder.WriteString(fmt.Sprintf("· 总距离：%.1f公里\n", analysis.TotalDistance))
	contentBuilder.WriteString(fmt.Sprintf("· 建议车速：%d km/h\n", analysis.SafeSpeed))
	contentBuilder.WriteString(fmt.Sprintf("· 风险等级：%s\n\n", riskLevelText))

	contentBuilder.WriteString("【预警路段】\n")
	for i, sw := range analysis.SegmentWarnings {
		if i >= 5 {
			contentBuilder.WriteString(fmt.Sprintf("... 共%d段预警路段\n", len(analysis.SegmentWarnings)))
			break
		}
		contentBuilder.WriteString(fmt.Sprintf("%d. 第%.1f公里处：%s%s，建议车速%dkm/h\n",
			i+1, sw.Distance, sw.WarningType, sw.WarningLevel, sw.SpeedSuggestion))
	}

	contentBuilder.WriteString("\n【安全建议】\n")
	for i, sug := range analysis.Suggestions {
		if i >= 5 {
			break
		}
		contentBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, sug))
	}

	if analysis.ShouldDetour && analysis.DetourSuggestion != "" {
		contentBuilder.WriteString(fmt.Sprintf("\n【绕行建议】%s\n", analysis.DetourSuggestion))
	}

	pushReq := &model.WeatherPushRequest{
		Phase:      string(model.PushPhasePreDeparture),
		WarningID:  &warningID,
		TargetType: "waybill",
		WaybillID:  &waybillID,
		Title:      title,
		Content:    contentBuilder.String(),
	}

	record, err := s.PushWeatherWarning(ctx, pushReq)
	if err != nil {
		return nil, fmt.Errorf("push pre-departure warning error: %w", err)
	}

	logger.Sugar.Infof("pre-departure warning sent: waybill=%s, risk_level=%s, push_id=%s",
		waybill.WaybillNo, analysis.OverallRiskLevel, record.PushID)

	return record, nil
}

func (s *WeatherService) EnRouteWarning(ctx context.Context, waybillID int64, currentLat, currentLng float64) (*model.WeatherPushRecord, error) {
	if waybillID <= 0 {
		return nil, fmt.Errorf("invalid waybill id")
	}

	var waybill model.Waybill
	if err := s.db.WithContext(ctx).Where("id = ?", waybillID).First(&waybill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("waybill not found")
		}
		return nil, fmt.Errorf("query waybill error: %w", err)
	}

	if waybill.RoutePlanID <= 0 {
		return nil, fmt.Errorf("waybill has no route plan")
	}

	weatherPoints, err := s.GetRouteWeather(ctx, waybill.RoutePlanID)
	if err != nil {
		return nil, fmt.Errorf("get route weather error: %w", err)
	}

	var nearestPoint *model.RouteWeatherPoint
	var minDistance float64 = -1
	for _, p := range weatherPoints {
		d := haversine(currentLat, currentLng, p.Lat, p.Lng)
		if minDistance < 0 || d < minDistance {
			minDistance = d
			nearestPoint = p
		}
	}

	if nearestPoint == nil {
		return nil, nil
	}

	if !nearestPoint.HasWarning && nearestPoint.SpeedSuggestion >= 80 {
		return nil, nil
	}

	lookAheadDistance := 10.0
	var upcomingWarnings []*model.SegmentWarning
	for i, p := range weatherPoints {
		distFromCurrent := math.Abs(p.DistanceFromStart/1000 - nearestPoint.DistanceFromStart/1000)
		if distFromCurrent <= lookAheadDistance && (p.HasWarning || p.SpeedSuggestion < 80) {
			upcomingWarnings = append(upcomingWarnings, &model.SegmentWarning{
				SegmentIndex:    i,
				StartLat:        p.Lat,
				StartLng:        p.Lng,
				EndLat:          p.Lat,
				EndLng:          p.Lng,
				Distance:        p.DistanceFromStart / 1000,
				WarningType:     p.WarningType,
				WarningLevel:    p.WarningLevel,
				SpeedSuggestion: p.SpeedSuggestion,
				Description:     fmt.Sprintf("前方%.1f公里处有%s天气", distFromCurrent, p.WarningType),
			})
		}
	}

	if len(upcomingWarnings) == 0 {
		return nil, nil
	}

	title := fmt.Sprintf("【行驶中提醒】前方%d公里处有天气预警", int(nearestPoint.DistanceFromStart/1000))

	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("运单 %s 前方路段天气预警：\n\n", waybill.WaybillNo))
	contentBuilder.WriteString(fmt.Sprintf("· 当前车速建议：%d km/h\n", nearestPoint.SpeedSuggestion))
	contentBuilder.WriteString(fmt.Sprintf("· 预警类型：%s\n\n", nearestPoint.WarningType))

	contentBuilder.WriteString("【前方预警路段】\n")
	for i, w := range upcomingWarnings {
		if i >= 5 {
			break
		}
		contentBuilder.WriteString(fmt.Sprintf("%d. 距离%.1f公里：%s%s，建议车速%dkm/h\n",
			i+1, w.Distance, w.WarningType, w.WarningLevel, w.SpeedSuggestion))
	}

	contentBuilder.WriteString("\n【安全提示】\n")
	contentBuilder.WriteString("1. 请减速慢行，保持安全车距\n")
	contentBuilder.WriteString("2. 注意路况变化，谨慎驾驶\n")
	contentBuilder.WriteString("3. 如遇极端天气，请就近服务区停靠\n")
	contentBuilder.WriteString("4. 紧急情况请联系调度中心\n")

	firstWarning := upcomingWarnings[0]

	pushReq := &model.WeatherPushRequest{
		Phase:           string(model.PushPhaseEnRoute),
		TargetType:      "waybill",
		WaybillID:       &waybillID,
		Title:           title,
		Content:         contentBuilder.String(),
		SegmentStartLat: nearestPoint.Lat,
		SegmentStartLng: nearestPoint.Lng,
		SegmentEndLat:   firstWarning.StartLat,
		SegmentEndLng:   firstWarning.StartLng,
	}

	record, err := s.PushWeatherWarning(ctx, pushReq)
	if err != nil {
		return nil, fmt.Errorf("push en-route warning error: %w", err)
	}

	logger.Sugar.Infof("en-route warning sent: waybill=%s, warning_type=%s, push_id=%s",
		waybill.WaybillNo, nearestPoint.WarningType, record.PushID)

	return record, nil
}

func (s *WeatherService) TriggerOperationSuspend(ctx context.Context, req *model.OperationSuspendRequest, operatorID int64, operatorName string) (*model.OperationSuspension, error) {
	if req.TriggerReason == "" {
		return nil, fmt.Errorf("trigger reason is required")
	}
	if req.CenterLat == 0 || req.CenterLng == 0 {
		return nil, fmt.Errorf("center lat/lng is required")
	}
	if req.RadiusKm <= 0 {
		return nil, fmt.Errorf("radius km is required")
	}

	now := time.Now()
	suspensionNo := fmt.Sprintf("OPS-SUS-%s-%04d", now.Format("20060102-150405"), time.Now().UnixNano()%10000)

	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = "manual"
	}
	autoTriggered := 0
	if triggerType == "automatic" {
		autoTriggered = 1
	}

	suspension := &model.OperationSuspension{
		SuspensionNo:       suspensionNo,
		TriggerType:        triggerType,
		TriggerReason:      req.TriggerReason,
		WeatherType:        req.WeatherType,
		Visibility:         req.Visibility,
		WindSpeed:          req.WindSpeed,
		AffectedRegion:     req.AffectedRegion,
		CenterLat:          req.CenterLat,
		CenterLng:          req.CenterLng,
		RadiusKm:           req.RadiusKm,
		Status:             "active",
		SuggestedSpeed:     req.SuggestedSpeed,
		SuspendTime:        &now,
		OperatorID:         operatorID,
		OperatorName:       operatorName,
		CreatedBy:          operatorID,
		AutoTriggered:      autoTriggered,
		Remark:             req.Remark,
	}

	if req.ExpiresAt != "" {
		expireTime, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			suspension.ExpiresAt = &expireTime
		}
	} else {
		expireTime := now.Add(24 * time.Hour)
		suspension.ExpiresAt = &expireTime
	}

	if req.TriggerWarningID != nil && *req.TriggerWarningID > 0 {
		warning, err := s.GetWarning(ctx, *req.TriggerWarningID)
		if err == nil {
			suspension.TriggerWarningID = warning.ID
			if suspension.CenterLat == 0 {
				suspension.CenterLat = warning.CenterLat
			}
			if suspension.CenterLng == 0 {
				suspension.CenterLng = warning.CenterLng
			}
		}
	}

	if len(req.TargetVehicleIDs) > 0 {
		vehicleIDsJSON, _ := json.Marshal(req.TargetVehicleIDs)
		suspension.AffectedVehicleIDs = model.JSON(vehicleIDsJSON)
		suspension.SuspendedVehicleCount = len(req.TargetVehicleIDs)
	}
	if len(req.TargetWaybillIDs) > 0 {
		waybillIDsJSON, _ := json.Marshal(req.TargetWaybillIDs)
		suspension.AffectedWaybillIDs = model.JSON(waybillIDsJSON)
		suspension.SuspendedWaybillCount = len(req.TargetWaybillIDs)
	}

	if len(req.TargetVehicleIDs) == 0 && len(req.TargetWaybillIDs) == 0 {
		var vehicleCount int64
		s.db.WithContext(ctx).Model(&model.Vehicle{}).Where("status = ?", "running").Count(&vehicleCount)
		suspension.SuspendedVehicleCount = int(vehicleCount)

		var waybillCount int64
		s.db.WithContext(ctx).Model(&model.Waybill{}).
			Where("status IN (?, ?, ?)", "assigned", "loading", "in_transit").
			Count(&waybillCount)
		suspension.SuspendedWaybillCount = int(waybillCount)
	}

	if err := s.db.WithContext(ctx).Create(suspension).Error; err != nil {
		return nil, fmt.Errorf("create suspension error: %w", err)
	}

	if len(req.TargetVehicleIDs) == 0 {
		s.db.WithContext(ctx).Model(&model.Vehicle{}).
			Where("status = ?", "running").
			Update("status", "offline")
	} else {
		s.db.WithContext(ctx).Model(&model.Vehicle{}).
			Where("id IN ?", req.TargetVehicleIDs).
			Where("status = ?", "running").
			Update("status", "offline")
	}

	logger.Sugar.Warnf("operation suspended: no=%s, type=%s, reason=%s, vehicles=%d, waybills=%d, auto=%v",
		suspensionNo, triggerType, req.TriggerReason,
		suspension.SuspendedVehicleCount, suspension.SuspendedWaybillCount, autoTriggered == 1)

	return suspension, nil
}

func (s *WeatherService) ResumeOperation(ctx context.Context, req *model.OperationResumeRequest, operatorID int64, operatorName string) (*model.OperationSuspension, error) {
	if req.SuspensionID <= 0 {
		return nil, fmt.Errorf("suspension id is required")
	}

	var suspension model.OperationSuspension
	if err := s.db.WithContext(ctx).Where("id = ?", req.SuspensionID).First(&suspension).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("suspension not found")
		}
		return nil, fmt.Errorf("query suspension error: %w", err)
	}

	if suspension.Status != "active" {
		return nil, fmt.Errorf("suspension is not active")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "lifted",
		"lifted_at":    &now,
		"lifted_by":    operatorID,
		"lift_reason":  req.LiftReason,
		"resume_time":  &now,
		"updated_at":   now,
	}

	if err := s.db.WithContext(ctx).Model(&suspension).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update suspension error: %w", err)
	}

	var vehicleIDs []int64
	if len(suspension.AffectedVehicleIDs) > 0 {
		json.Unmarshal(suspension.AffectedVehicleIDs, &vehicleIDs)
	}

	if len(vehicleIDs) > 0 {
		s.db.WithContext(ctx).Model(&model.Vehicle{}).
			Where("id IN ?", vehicleIDs).
			Where("status = ?", "offline").
			Update("status", "idle")
	} else {
		s.db.WithContext(ctx).Model(&model.Vehicle{}).
			Where("status = ?", "offline").
			Update("status", "idle")
	}

	logger.Sugar.Infof("operation resumed: no=%s, reason=%s, operator=%s",
		suspension.SuspensionNo, req.LiftReason, operatorName)

	suspension.Status = "lifted"
	suspension.LiftedAt = &now
	suspension.LiftedBy = operatorID
	suspension.LiftReason = req.LiftReason
	suspension.ResumeTime = &now

	return &suspension, nil
}

func (s *WeatherService) ListSuspensions(ctx context.Context, page, pageSize int, status string) (*model.OperationSuspensionPage, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var suspensions []*model.OperationSuspension
	var total int64

	query := s.db.WithContext(ctx).Model(&model.OperationSuspension{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count suspensions error: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&suspensions).Error; err != nil {
		return nil, fmt.Errorf("query suspensions error: %w", err)
	}

	return &model.OperationSuspensionPage{
		List:     suspensions,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *WeatherService) GetCurrentSuspension(ctx context.Context) (*model.OperationSuspension, error) {
	var suspension model.OperationSuspension
	if err := s.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("created_at DESC").
		First(&suspension).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &suspension, nil
