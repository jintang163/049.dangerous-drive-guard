package model

import (
	"time"
)

type WeatherWarningType string

const (
	WarningTypeRainstorm  WeatherWarningType = "rainstorm"
	WarningTypeTyphoon    WeatherWarningType = "typhoon"
	WarningTypeSnowstorm  WeatherWarningType = "snowstorm"
	WarningTypeFog        WeatherWarningType = "fog"
	WarningTypeHaze       WeatherWarningType = "haze"
	WarningTypeThunder    WeatherWarningType = "thunder"
	WarningTypeHighTemp   WeatherWarningType = "high_temp"
	WarningTypeLowTemp    WeatherWarningType = "low_temp"
	WarningTypeStrongWind WeatherWarningType = "strong_wind"
	WarningTypeSandstorm  WeatherWarningType = "sandstorm"
	WarningTypeHail       WeatherWarningType = "hail"
	WarningTypeIcyRoad    WeatherWarningType = "icy_road"
	WarningTypeSlippery   WeatherWarningType = "slippery"
)

type WarningLevel string

const (
	WarningLevelBlue   WarningLevel = "blue"
	WarningLevelYellow WarningLevel = "yellow"
	WarningLevelOrange WarningLevel = "orange"
	WarningLevelRed    WarningLevel = "red"
)

type PushPhase string

const (
	PushPhasePreDeparture PushPhase = "pre_departure"
	PushPhaseEnRoute      PushPhase = "en_route"
	PushPhaseEmergency    PushPhase = "emergency"
)

type WeatherData struct {
	Temp         float64 `json:"temperature"`
	Humidity     float64 `json:"humidity"`
	WindSpeed    float64 `json:"wind_speed"`
	WindDirection string `json:"wind_direction"`
	Visibility   float64 `json:"visibility"`
	Condition    string  `json:"condition"`
	Precipitation float64 `json:"precipitation"`
	RoadSlippery bool    `json:"road_slippery"`
	FeelsLike    float64 `json:"feels_like"`
	Pressure     float64 `json:"pressure"`
	UvIndex      int     `json:"uv_index"`
}

type RouteWeatherPoint struct {
	Lat               float64     `json:"latitude"`
	Lng               float64     `json:"longitude"`
	WeatherData       WeatherData `json:"weather"`
	DistanceFromStart float64     `json:"distance_from_start"`
	EstimatedTime     time.Time   `json:"estimated_time"`
	SpeedSuggestion   int         `json:"speed_suggestion_kmh"`
	HasWarning        bool        `json:"has_warning"`
	WarningType       string      `json:"warning_type,omitempty"`
	WarningLevel      string      `json:"warning_level,omitempty"`
}

type WeatherWarning struct {
	BaseModel
	WarningID           string             `gorm:"type:varchar(64);uniqueIndex;not null" json:"warning_id"`
	WarningType         WeatherWarningType `gorm:"type:varchar(32);index;not null" json:"warning_type"`
	WarningLevel        WarningLevel       `gorm:"type:varchar(8);index;not null" json:"warning_level"`
	Title               string             `gorm:"type:varchar(256);not null" json:"title"`
	Content             string             `gorm:"type:text" json:"content"`
	AffectedProvinces   JSON               `gorm:"type:json" json:"affected_provinces"`
	AffectedCities      JSON               `gorm:"type:json" json:"affected_cities"`
	AffectedAreaPolygon JSON               `gorm:"type:json" json:"affected_area_polygon"`
	CenterLat           float64            `json:"center_lat"`
	CenterLng           float64            `json:"center_lng"`
	StartTime           time.Time          `gorm:"index;not null" json:"start_time"`
	EndTime             *time.Time         `json:"end_time"`
	PublishTime         time.Time          `gorm:"not null" json:"publish_time"`
	Source              string             `gorm:"type:varchar(64)" json:"source"`
	RelatedWaybillCount int                `gorm:"default:0" json:"related_waybill_count"`
	RelatedVehicleCount int                `gorm:"default:0" json:"related_vehicle_count"`
	Processed           int                `gorm:"default:0;index" json:"processed"`
	TriggerOperationStop int               `gorm:"default:0;index" json:"trigger_operation_stop"`
	SpeedSuggestion     int                `gorm:"default:0" json:"speed_suggestion_kmh"`
	Suggestion          string             `gorm:"type:text" json:"suggestion"`
}

func (WeatherWarning) TableName() string {
	return "weather_warnings"
}

type WeatherWarningPage struct {
	List     []*WeatherWarning `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type WeatherPushRecord struct {
	BaseModel
	PushID            string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"push_id"`
	PushPhase         PushPhase `gorm:"type:varchar(20);index" json:"push_phase"`
	WarningID         int64     `gorm:"index" json:"warning_id"`
	WarningNo         string    `gorm:"type:varchar(64)" json:"warning_no"`
	WarningType       string    `gorm:"type:varchar(32)" json:"warning_type"`
	WarningLevel      string    `gorm:"type:varchar(8)" json:"warning_level"`
	Title             string    `gorm:"type:varchar(256);not null" json:"title"`
	Content           string    `gorm:"type:text;not null" json:"content"`
	TargetType        string    `gorm:"type:varchar(20);index" json:"target_type"`
	TargetIDs         JSON      `gorm:"type:json" json:"target_ids"`
	WaybillID         int64     `gorm:"index" json:"waybill_id"`
	WaybillNo         string    `gorm:"type:varchar(32)" json:"waybill_no"`
	VehicleID         int64     `gorm:"index" json:"vehicle_id"`
	PlateNumber       string    `gorm:"type:varchar(20)" json:"plate_number"`
	DriverID          int64     `json:"driver_id"`
	DriverName        string    `gorm:"type:varchar(64)" json:"driver_name"`
	PushChannels      JSON      `gorm:"type:json" json:"push_channels"`
	Status            string    `gorm:"type:varchar(20);index;default:pending" json:"status"`
	SuccessCount      int       `gorm:"default:0" json:"success_count"`
	FailCount         int       `gorm:"default:0" json:"fail_count"`
	ReadCount         int       `gorm:"default:0" json:"read_count"`
	ReadStatus        int       `gorm:"default:0;index" json:"read_status"`
	ReadTime          *time.Time `json:"read_time"`
	DriverResponse    string    `gorm:"type:varchar(32)" json:"driver_response"`
	ResponseTime      *time.Time `json:"response_time"`
	ResponseNote      string    `gorm:"type:text" json:"response_note"`
	SpeedSuggestion   int       `json:"speed_suggestion_kmh"`
	SegmentStartLat   float64   `json:"segment_start_lat"`
	SegmentStartLng   float64   `json:"segment_start_lng"`
	SegmentEndLat     float64   `json:"segment_end_lat"`
	SegmentEndLng     float64   `json:"segment_end_lng"`
	SegmentDistance   float64   `json:"segment_distance_km"`
	OperatorID        int64     `json:"operator_id"`
	OperatorName      string    `gorm:"type:varchar(64)" json:"operator_name"`
	SentAt            *time.Time `json:"sent_at"`
}

func (WeatherPushRecord) TableName() string {
	return "weather_push_records"
}

type WeatherPushPage struct {
	List     []*WeatherPushRecord `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type WeatherPushRequest struct {
	Phase         string  `json:"phase" binding:"required"`
	WarningID     *int64  `json:"warning_id"`
	TargetType    string  `json:"target_type" binding:"required"`
	TargetIDs     []int64 `json:"target_ids"`
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	WaybillID     *int64  `json:"waybill_id"`
	SegmentStartLat float64 `json:"segment_start_lat"`
	SegmentStartLng float64 `json:"segment_start_lng"`
	SegmentEndLat   float64 `json:"segment_end_lat"`
	SegmentEndLng   float64 `json:"segment_end_lng"`
}

type HistoricalWeather struct {
	BaseModel
	Latitude         float64   `gorm:"index" json:"latitude"`
	Longitude        float64   `gorm:"index" json:"longitude"`
	LocationName     string    `gorm:"type:varchar(256)" json:"location_name"`
	QueryTime        time.Time `gorm:"index" json:"query_time"`
	WeatherCondition string    `gorm:"type:varchar(64)" json:"weather_condition"`
	Temperature      float64   `json:"temperature"`
	FeelsLike        float64   `json:"feels_like"`
	Humidity         float64   `json:"humidity"`
	WindSpeed        float64   `json:"wind_speed"`
	WindDirection    int       `json:"wind_direction"`
	Visibility       float64   `json:"visibility"`
	Pressure         float64   `json:"pressure"`
	Precipitation    float64   `gorm:"default:0" json:"precipitation"`
	PrecipType       string    `gorm:"type:varchar(16)" json:"precip_type"`
	RoadSlippery     bool      `gorm:"default:false" json:"road_slippery"`
	RoadCondition    string    `gorm:"type:varchar(32)" json:"road_condition"`
	UvIndex          int       `json:"uv_index"`
	Warnings         JSON      `gorm:"type:json" json:"warnings"`
	WarningType      string    `gorm:"type:varchar(32)" json:"warning_type,omitempty"`
	WarningLevel     string    `gorm:"type:varchar(8)" json:"warning_level,omitempty"`
	DataSource       string    `gorm:"type:varchar(32)" json:"data_source"`
}

func (HistoricalWeather) TableName() string {
	return "historical_weather"
}

type HistoricalWeatherQuery struct {
	Latitude     float64 `json:"latitude" form:"lat"`
	Longitude    float64 `json:"longitude" form:"lng"`
	LocationName string  `json:"location_name" form:"location_name"`
	QueryTime    string  `json:"query_time" form:"query_time"`
	StartTime    string  `json:"start_time" form:"start_time"`
	EndTime      string  `json:"end_time" form:"end_time"`
}

type OperationSuspension struct {
	BaseModel
	SuspensionNo         string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"suspension_no"`
	TriggerType          string     `gorm:"type:varchar(20);index" json:"trigger_type"`
	TriggerReason        string     `gorm:"type:varchar(512)" json:"trigger_reason"`
	TriggerWarningID     int64      `gorm:"index" json:"trigger_warning_id"`
	WeatherType          string     `gorm:"type:varchar(32)" json:"weather_type"`
	Visibility           float64    `json:"visibility"`
	WindSpeed            float64    `json:"wind_speed"`
	Precipitation        float64    `json:"precipitation"`
	AffectedRegion       string     `gorm:"type:varchar(512)" json:"affected_region"`
	CenterLat            float64    `json:"center_lat"`
	CenterLng            float64    `json:"center_lng"`
	RadiusKm             float64    `json:"radius_km"`
	AffectedProvinces    JSON       `gorm:"type:json" json:"affected_provinces"`
	AffectedCities       JSON       `gorm:"type:json" json:"affected_cities"`
	AffectedPolygon      JSON       `gorm:"type:json" json:"affected_polygon"`
	AffectedVehicleIDs   JSON       `gorm:"type:json" json:"affected_vehicle_ids"`
	AffectedWaybillIDs   JSON       `gorm:"type:json" json:"affected_waybill_ids"`
	Status               string     `gorm:"type:varchar(20);index;default:active" json:"status"`
	SuggestedSpeed       int        `gorm:"default:0" json:"suggested_speed"`
	SuspendTime          *time.Time `json:"suspend_time"`
	ResumeTime           *time.Time `json:"resume_time"`
	ExpiresAt            *time.Time `json:"expires_at"`
	LiftReason           string     `gorm:"type:varchar(512)" json:"lift_reason"`
	LiftedBy             int64      `json:"lifted_by"`
	LiftedAt             *time.Time `json:"lifted_at"`
	SuspendedWaybillCount int       `gorm:"default:0" json:"suspended_waybill_count"`
	SuspendedVehicleCount int       `gorm:"default:0" json:"suspended_vehicle_count"`
	OperatorID           int64      `json:"operator_id"`
	OperatorName         string     `gorm:"type:varchar(64)" json:"operator_name"`
	CreatedBy            int64      `json:"created_by"`
	AutoTriggered        int        `gorm:"default:0" json:"auto_triggered"`
	Remark               string     `gorm:"type:text" json:"remark"`
}

func (OperationSuspension) TableName() string {
	return "operation_suspensions"
}

type OperationSuspensionPage struct {
	List     []*OperationSuspension `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type OperationSuspendRequest struct {
	TriggerType       string   `json:"trigger_type" binding:"required"`
	TriggerReason     string   `json:"trigger_reason" binding:"required"`
	WeatherType       string   `json:"weather_type"`
	Visibility        float64  `json:"visibility"`
	WindSpeed         float64  `json:"wind_speed"`
	AffectedRegion    string   `json:"affected_region" binding:"required"`
	CenterLat         float64  `json:"center_lat" binding:"required"`
	CenterLng         float64  `json:"center_lng" binding:"required"`
	RadiusKm          float64  `json:"radius_km" binding:"required"`
	SuggestedSpeed    int      `json:"suggested_speed"`
	ExpiresAt         string   `json:"expires_at"`
	TargetVehicleIDs  []int64  `json:"affected_vehicle_ids"`
	TargetWaybillIDs  []int64  `json:"affected_waybill_ids"`
	TriggerWarningID  *int64   `json:"trigger_warning_id"`
	AutoTriggered     bool     `json:"auto_triggered"`
	Remark            string   `json:"remark"`
}

type OperationResumeRequest struct {
	SuspensionID int64  `json:"suspension_id" binding:"required"`
	LiftReason   string `json:"lift_reason" binding:"required"`
	OperatorID   int64  `json:"operator_id"`
	OperatorName string `json:"operator_name"`
}

type SegmentWarning struct {
	SegmentIndex    int       `json:"segment_index"`
	StartLat        float64   `json:"start_lat"`
	StartLng        float64   `json:"start_lng"`
	EndLat          float64   `json:"end_lat"`
	EndLng          float64   `json:"end_lng"`
	Distance        float64   `json:"distance_km"`
	WarningType     string    `json:"warning_type"`
	WarningLevel    string    `json:"warning_level"`
	Description     string    `json:"description"`
	SpeedSuggestion int       `json:"speed_suggestion_kmh"`
	DetourSuggested bool      `json:"detour_suggested"`
	EstimatedETA    time.Time `json:"estimated_eta"`
}

type RouteWeatherAnalysis struct {
	RouteID           int64              `json:"route_id"`
	WaybillID         int64              `json:"waybill_id,omitempty"`
	TotalDistance     float64            `json:"total_distance_km"`
	EstimatedDuration int                `json:"estimated_duration_min"`
	OverallRiskLevel  string             `json:"overall_risk_level"`
	AverageVisibility float64            `json:"average_visibility_km"`
	SafeSpeed         int                `json:"safe_speed_suggestion"`
	HasExtremeWeather bool               `json:"has_extreme_weather"`
	OperationSuggested bool              `json:"operation_suggested"`
	ShouldDetour      bool               `json:"should_detour"`
	WarningsOnRoute   []*WeatherWarning  `json:"warnings_on_route"`
	SegmentWarnings   []*SegmentWarning  `json:"segment_warnings"`
	RainSegments      []*SegmentWarning  `json:"rain_segments"`
	SlipperySegments  []*SegmentWarning  `json:"slippery_segments"`
	FogSegments       []*SegmentWarning  `json:"fog_segments"`
	WindSegments      []*SegmentWarning  `json:"wind_segments"`
	WeatherPoints     []*RouteWeatherPoint `json:"weather_points"`
	Suggestions       []string           `json:"suggestions"`
	DetourSuggestion  string             `json:"detour_suggestion,omitempty"`
	AnalyzedAt        time.Time          `json:"analyzed_at"`
}
