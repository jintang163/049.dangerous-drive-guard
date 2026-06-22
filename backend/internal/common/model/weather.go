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
	WaybillID       int64     `gorm:"index" json:"waybill_id"`
	WaybillNo       string    `gorm:"type:varchar(32);index" json:"waybill_no"`
	VehicleID       int64     `gorm:"index" json:"vehicle_id"`
	PlateNumber     string    `gorm:"type:varchar(20)" json:"plate_number"`
	DriverID        int64     `json:"driver_id"`
	DriverName      string    `gorm:"type:varchar(64)" json:"driver_name"`
	WarningID       int64     `gorm:"index" json:"warning_id"`
	WarningNo       string    `gorm:"type:varchar(64)" json:"warning_no"`
	WarningType     string    `gorm:"type:varchar(32)" json:"warning_type"`
	WarningLevel    string    `gorm:"type:varchar(8)" json:"warning_level"`
	PushPhase       PushPhase `gorm:"type:varchar(20);index" json:"push_phase"`
	MessageTitle    string    `gorm:"type:varchar(256)" json:"message_title"`
	MessageContent  string    `gorm:"type:text" json:"message_content"`
	PushChannels    JSON      `gorm:"type:json" json:"push_channels"`
	SegmentStartLat float64   `json:"segment_start_lat"`
	SegmentStartLng float64   `json:"segment_start_lng"`
	SegmentEndLat   float64   `json:"segment_end_lat"`
	SegmentEndLng   float64   `json:"segment_end_lng"`
	SegmentDistance float64   `json:"segment_distance_km"`
	SpeedSuggestion int       `json:"speed_suggestion_kmh"`
	ReadStatus      int       `gorm:"default:0;index" json:"read_status"`
	ReadTime        *time.Time `json:"read_time"`
	DriverResponse  string    `gorm:"type:varchar(32)" json:"driver_response"`
	ResponseTime    *time.Time `json:"response_time"`
	ResponseNote    string    `gorm:"type:text" json:"response_note"`
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

type HistoricalWeather struct {
	BaseModel
	Latitude       float64   `gorm:"index" json:"latitude"`
	Longitude      float64   `gorm:"index" json:"longitude"`
	Province       string    `gorm:"type:varchar(64)" json:"province"`
	City           string    `gorm:"type:varchar(64)" json:"city"`
	District       string    `gorm:"type:varchar(64)" json:"district"`
	RecordDate     time.Time `gorm:"index" json:"record_date"`
	TempMax        float64   `json:"temp_max"`
	TempMin        float64   `json:"temp_min"`
	TempAvg        float64   `json:"temp_avg"`
	Humidity       float64   `json:"humidity"`
	WindSpeed      float64   `json:"wind_speed"`
	WindDirection  string    `gorm:"type:varchar(16)" json:"wind_direction"`
	Visibility     float64   `json:"visibility"`
	Precipitation  float64   `json:"precipitation"`
	PrecipType     string    `gorm:"type:varchar(16)" json:"precip_type"`
	Condition      string    `gorm:"type:varchar(64)" json:"condition"`
	Pressure       float64   `json:"pressure"`
	UvIndex        int       `json:"uv_index"`
	RoadCondition  string    `gorm:"type:varchar(32)" json:"road_condition"`
	RoadSlippery   bool      `json:"road_slippery"`
	WarningType    string    `gorm:"type:varchar(32)" json:"warning_type,omitempty"`
	WarningLevel   string    `gorm:"type:varchar(8)" json:"warning_level,omitempty"`
	DataSource     string    `gorm:"type:varchar(32)" json:"data_source"`
}

func (HistoricalWeather) TableName() string {
	return "historical_weather"
}

type OperationSuspension struct {
	BaseModel
	SuspensionNo    string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"suspension_no"`
	TriggerType     string     `gorm:"type:varchar(32);index" json:"trigger_type"`
	TriggerReason   string     `gorm:"type:varchar(256)" json:"trigger_reason"`
	TriggerWarningID int64     `gorm:"index" json:"trigger_warning_id"`
	TriggerLat      float64    `json:"trigger_lat"`
	TriggerLng      float64    `json:"trigger_lng"`
	TriggerProvince string     `gorm:"type:varchar(64)" json:"trigger_province"`
	TriggerCity     string     `gorm:"type:varchar(64)" json:"trigger_city"`
	Visibility      float64    `json:"visibility_m"`
	WindSpeed       float64    `json:"wind_speed_ms"`
	Precipitation   float64    `json:"precipitation_mm"`
	AreaScope       string     `gorm:"type:varchar(32)" json:"area_scope"`
	AffectedProvinces JSON     `gorm:"type:json" json:"affected_provinces"`
	AffectedCities  JSON       `gorm:"type:json" json:"affected_cities"`
	AffectedPolygon JSON       `gorm:"type:json" json:"affected_polygon"`
	Status          string     `gorm:"type:varchar(20);index" json:"status"`
	SuspendTime     *time.Time `json:"suspend_time"`
	ResumeTime      *time.Time `json:"resume_time"`
	SuspendedWaybillCount int   `gorm:"default:0" json:"suspended_waybill_count"`
	SuspendedVehicleCount int   `gorm:"default:0" json:"suspended_vehicle_count"`
	OperatorID      int64      `json:"operator_id"`
	OperatorName    string     `gorm:"type:varchar(64)" json:"operator_name"`
	AutoTriggered   int        `gorm:"default:0" json:"auto_triggered"`
	ResumeReason    string     `gorm:"type:varchar(256)" json:"resume_reason"`
	ResumeOperatorID int64     `json:"resume_operator_id"`
	ResumeOperatorName string   `gorm:"type:varchar(64)" json:"resume_operator_name"`
	Remark          string     `gorm:"type:text" json:"remark"`
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
	RouteID              int64            `json:"route_id"`
	TotalDistance        float64          `json:"total_distance_km"`
	EstimatedDuration    int              `json:"estimated_duration_min"`
	OverallRiskLevel     string           `json:"overall_risk_level"`
	AverageVisibility    float64          `json:"average_visibility_km"`
	SafeSpeed            int              `json:"safe_speed_kmh"`
	HasExtremeWeather    bool             `json:"has_extreme_weather"`
	OperationSuggested   bool             `json:"operation_suggested"`
	SegmentWarnings      []*SegmentWarning `json:"segment_warnings"`
	RainSegments         []*SegmentWarning `json:"rain_segments"`
	SlipperySegments     []*SegmentWarning `json:"slippery_segments"`
	FogSegments          []*SegmentWarning `json:"fog_segments"`
	WindSegments         []*SegmentWarning `json:"wind_segments"`
	WeatherPoints        []*RouteWeatherPoint `json:"weather_points"`
	SuggestionSummary    string           `json:"suggestion_summary"`
	AlternativeRouteHint string           `json:"alternative_route_hint,omitempty"`
}

type WeatherPushRequest struct {
	WaybillID    int64   `json:"waybill_id" binding:"required"`
	PushPhase    string  `json:"push_phase"`
	WarningID    *int64  `json:"warning_id"`
	CustomTitle  string  `json:"custom_title"`
	CustomContent string `json:"custom_content"`
}

type OperationSuspendRequest struct {
	TriggerWarningID *int64  `json:"trigger_warning_id"`
	TriggerReason    string  `json:"trigger_reason" binding:"required"`
	AreaScope        string  `json:"area_scope"`
	Provinces        []string `json:"provinces"`
	Cities           []string `json:"cities"`
	Remark           string  `json:"remark"`
	AutoTriggered    bool    `json:"auto_triggered"`
}

type OperationResumeRequest struct {
	SuspensionID int64  `json:"suspension_id" binding:"required"`
	ResumeReason string `json:"resume_reason" binding:"required"`
}

type HistoricalWeatherQuery struct {
	Latitude   float64 `json:"latitude" form:"lat"`
	Longitude  float64 `json:"longitude" form:"lng"`
	Province   string  `json:"province" form:"province"`
	City       string  `json:"city" form:"city"`
	District   string  `json:"district" form:"district"`
	StartDate  string  `json:"start_date" form:"start_date" binding:"required"`
	EndDate    string  `json:"end_date" form:"end_date" binding:"required"`
}
