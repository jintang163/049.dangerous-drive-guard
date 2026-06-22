package model

import (
	"time"
)

type ADASAlertType string

const (
	ADASAlertCloseFollowing  ADASAlertType = "close_following"
	ADASAlertLaneDeparture   ADASAlertType = "lane_departure"
	ADASAlertForwardCollision ADASAlertType = "forward_collision"
	ADASAlertAutoDecelerate  ADASAlertType = "auto_decelerate"
)

type ADASAlertLevel string

const (
	ADASLevelInfo     ADASAlertLevel = "info"
	ADASLevelWarning  ADASAlertLevel = "warning"
	ADASLevelCritical ADASAlertLevel = "critical"
)

type ADASAlertStatus string

const (
	ADASStatusActive    ADASAlertStatus = "active"
	ADASStatusResolved  ADASAlertStatus = "resolved"
	ADASStatusEscalated ADASAlertStatus = "escalated"
)

type RadarData struct {
	VehicleID           int64   `json:"vehicle_id" binding:"required"`
	DriverID            int64   `json:"driver_id" binding:"required"`
	WaybillID           int64   `json:"waybill_id"`
	Timestamp           int64   `json:"timestamp"`
	FollowingDistance   float64 `json:"following_distance_m"`
	RelativeSpeed       float64 `json:"relative_speed_kmh"`
	LaneOffset          float64 `json:"lane_offset_m"`
	LaneDepartureLeft   bool    `json:"lane_departure_left"`
	LaneDepartureRight  bool    `json:"lane_departure_right"`
	ForwardCollisionTTC float64 `json:"forward_collision_ttc_s"`
	VehicleSpeed        float64 `json:"vehicle_speed_kmh"`
	SteeringAngle       float64 `json:"steering_angle_deg"`
	YawRate             float64 `json:"yaw_rate_deg_s"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	Confidence          float64 `json:"confidence"`
}

type ADASAlert struct {
	BaseModel
	AlertNo           string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"alert_no"`
	VehicleID         int64          `gorm:"index;not null" json:"vehicle_id"`
	DriverID          int64          `gorm:"index;not null" json:"driver_id"`
	WaybillID         int64          `gorm:"index" json:"waybill_id"`
	AlertType         ADASAlertType  `gorm:"type:varchar(32);index;not null" json:"alert_type"`
	AlertLevel        ADASAlertLevel `gorm:"type:varchar(16);index;not null" json:"alert_level"`
	Status            ADASAlertStatus `gorm:"type:varchar(16);index;default:active" json:"status"`
	TriggerValue      float64        `gorm:"type:decimal(10,2)" json:"trigger_value"`
	ThresholdValue    float64        `gorm:"type:decimal(10,2)" json:"threshold_value"`
	FollowingDistance  float64        `gorm:"type:decimal(10,2)" json:"following_distance_m"`
	VehicleSpeed      float64        `gorm:"type:decimal(5,2)" json:"vehicle_speed_kmh"`
	LaneOffset        float64        `gorm:"type:decimal(6,3)" json:"lane_offset_m"`
	DepartureSide     string         `gorm:"type:varchar(8)" json:"departure_side"`
	TTC               float64        `gorm:"type:decimal(6,2)" json:"ttc_s"`
	AlertMessage      string         `gorm:"type:varchar(512)" json:"alert_message"`
	Latitude          float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude         float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	SuggestedAction   string         `gorm:"type:varchar(256)" json:"suggested_action"`
	DecelerateTriggered bool         `gorm:"default:false" json:"decelerate_triggered"`
	DecelerateValue   float64        `gorm:"type:decimal(5,2)" json:"decelerate_value_kmh"`
	ReportedToCenter  bool           `gorm:"default:false" json:"reported_to_center"`
	DriverAcknowledged bool          `gorm:"default:false" json:"driver_acknowledged"`
	AcknowledgedAt    *time.Time     `json:"acknowledged_at"`
	ResolvedAt        *time.Time     `json:"resolved_at"`
	ResolutionNote    string         `gorm:"type:varchar(512)" json:"resolution_note"`
	VehiclePlate      string         `gorm:"-" json:"vehicle_plate,omitempty"`
	DriverName        string         `gorm:"-" json:"driver_name,omitempty"`
}

func (ADASAlert) TableName() string {
	return "adas_alerts"
}

type ADASAlertPage struct {
	List     []*ADASAlert `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type ADASFrequencyTracker struct {
	BaseModel
	VehicleID          int64     `gorm:"index;not null" json:"vehicle_id"`
	DriverID           int64     `gorm:"index;not null" json:"driver_id"`
	WindowStart        time.Time `gorm:"index;not null" json:"window_start"`
	WindowEnd          time.Time `gorm:"not null" json:"window_end"`
	CloseFollowingCount int      `gorm:"default:0" json:"close_following_count"`
	LaneDepartureCount int      `gorm:"default:0" json:"lane_departure_count"`
	TotalAlertCount    int      `gorm:"default:0" json:"total_alert_count"`
	DecelerateTriggered bool     `gorm:"default:false" json:"decelerate_triggered"`
	DecelerateValue    float64   `gorm:"type:decimal(5,2)" json:"decelerate_value_kmh"`
	ReportedToCenter   bool      `gorm:"default:false" json:"reported_to_center"`
}

func (ADASFrequencyTracker) TableName() string {
	return "adas_frequency_trackers"
}

type ADASConfig struct {
	BaseModel
	VehicleID               int64   `gorm:"uniqueIndex;not null" json:"vehicle_id"`
	CloseFollowingMinDist   float64 `gorm:"type:decimal(6,2);default:5.00" json:"close_following_min_dist_m"`
	CloseFollowingWarnDist  float64 `gorm:"type:decimal(6,2);default:10.00" json:"close_following_warn_dist_m"`
	CloseFollowingCritDist  float64 `gorm:"type:decimal(6,2);default:3.00" json:"close_following_crit_dist_m"`
	LaneDepartureThreshold  float64 `gorm:"type:decimal(4,2);default:0.30" json:"lane_departure_threshold_m"`
	ForwardCollisionTTCWarn float64 `gorm:"type:decimal(4,1);default:3.0" json:"forward_collision_ttc_warn_s"`
	ForwardCollisionTTCCrit float64 `gorm:"type:decimal(4,1);default:1.5" json:"forward_collision_ttc_crit_s"`
	FrequencyWindowMinutes  int     `gorm:"default:5" json:"frequency_window_minutes"`
	FrequencyAlertThreshold int     `gorm:"default:6" json:"frequency_alert_threshold"`
	AutoDecelerateSpeed     float64 `gorm:"type:decimal(5,2);default:20.00" json:"auto_decelerate_speed_kmh"`
	EnableCloseFollowing    bool    `gorm:"default:true" json:"enable_close_following"`
	EnableLaneDeparture     bool    `gorm:"default:true" json:"enable_lane_departure"`
	EnableForwardCollision  bool    `gorm:"default:true" json:"enable_forward_collision"`
	EnableAutoDecelerate    bool    `gorm:"default:true" json:"enable_auto_decelerate"`
}

func (ADASConfig) TableName() string {
	return "adas_configs"
}

type ADASAlertAckRequest struct {
	AlertID  int64  `json:"alert_id" binding:"required"`
	AckType  string `json:"ack_type" binding:"required"`
	Note     string `json:"note"`
}

type ADASAlertQuery struct {
	VehicleID  int64  `form:"vehicle_id"`
	DriverID   int64  `form:"driver_id"`
	WaybillID  int64  `form:"waybill_id"`
	AlertType  string `form:"alert_type"`
	AlertLevel string `form:"alert_level"`
	Status     string `form:"status"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type RadarDataResponse struct {
	AlertTriggered    bool         `json:"alert_triggered"`
	Alerts            []*ADASAlert `json:"alerts,omitempty"`
	DecelerateTriggered bool       `json:"decelerate_triggered"`
	DecelerateValue   float64      `json:"decelerate_value_kmh,omitempty"`
	FrequencyAlert    bool         `json:"frequency_alert"`
	CurrentFollowingDistance float64 `json:"current_following_distance_m,omitempty"`
	CurrentLaneOffset float64      `json:"current_lane_offset_m,omitempty"`
}
