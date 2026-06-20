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
)

type WarningLevel string

const (
	WarningLevelBlue   WarningLevel = "blue"
	WarningLevelYellow WarningLevel = "yellow"
	WarningLevelOrange WarningLevel = "orange"
	WarningLevelRed    WarningLevel = "red"
)

type WeatherData struct {
	Temp       float64 `json:"temperature"`
	Humidity   float64 `json:"humidity"`
	WindSpeed  float64 `json:"wind_speed"`
	Visibility float64 `json:"visibility"`
	Condition  string  `json:"condition"`
}

type RouteWeatherPoint struct {
	Lat               float64     `json:"latitude"`
	Lng               float64     `json:"longitude"`
	WeatherData       WeatherData `json:"weather"`
	DistanceFromStart float64     `json:"distance_from_start"`
	EstimatedTime     time.Time   `json:"estimated_time"`
}

type WeatherWarning struct {
	BaseModel
	WarningID          string             `gorm:"type:varchar(64);uniqueIndex;not null" json:"warning_id"`
	WarningType        WeatherWarningType `gorm:"type:varchar(32);index;not null" json:"warning_type"`
	WarningLevel       WarningLevel       `gorm:"type:varchar(8);index;not null" json:"warning_level"`
	Title              string             `gorm:"type:varchar(256);not null" json:"title"`
	Content            string             `gorm:"type:text" json:"content"`
	AffectedProvinces  JSON               `gorm:"type:json" json:"affected_provinces"`
	AffectedCities     JSON               `gorm:"type:json" json:"affected_cities"`
	AffectedAreaPolygon JSON              `gorm:"type:json" json:"affected_area_polygon"`
	StartTime          time.Time          `gorm:"index;not null" json:"start_time"`
	EndTime            *time.Time         `json:"end_time"`
	PublishTime        time.Time          `gorm:"not null" json:"publish_time"`
	Source             string             `gorm:"type:varchar(64)" json:"source"`
	RelatedWaybillCount int                `gorm:"default:0" json:"related_waybill_count"`
	RelatedVehicleCount int                `gorm:"default:0" json:"related_vehicle_count"`
	Processed          int                `gorm:"default:0;index" json:"processed"`
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
