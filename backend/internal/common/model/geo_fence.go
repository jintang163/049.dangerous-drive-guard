package model

import "time"

type GeoFenceAlertStatus string

const (
	GeoFenceStatusPending    GeoFenceAlertStatus = "pending"
	GeoFenceStatusConfirmed  GeoFenceAlertStatus = "confirmed"
	GeoFenceStatusEscalated  GeoFenceAlertStatus = "escalated"
	GeoFenceStatusResolved   GeoFenceAlertStatus = "resolved"
)

type DeviateReason string

const (
	DeviateReasonDetour  DeviateReason = "detour"
	DeviateReasonDeviate DeviateReason = "deviate"
)

type GeoFenceAlert struct {
	BaseModel
	AlertNo               string              `gorm:"type:varchar(32);uniqueIndex;not null" json:"alert_no"`
	VehicleID             int64               `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber           string              `gorm:"type:varchar(20)" json:"plate_number"`
	DriverID              int64               `gorm:"index" json:"driver_id"`
	DriverName            string              `gorm:"type:varchar(64)" json:"driver_name"`
	EscortID              int64               `gorm:"index" json:"escort_id"`
	EscortName            string              `gorm:"type:varchar(64)" json:"escort_name"`
	WaybillID             int64               `gorm:"index" json:"waybill_id"`
	WaybillNo             string              `gorm:"type:varchar(32)" json:"waybill_no"`
	RoutePlanID           int64               `json:"route_plan_id"`
	Latitude              float64             `gorm:"not null" json:"latitude"`
	Longitude             float64             `gorm:"not null" json:"longitude"`
	Address               string              `gorm:"type:varchar(512)" json:"address"`
	DistanceFromRouteM    int                 `gorm:"column:distance_from_route_meters;not null;default:0" json:"distance_from_route_meters"`
	ThresholdMeters       int                 `gorm:"not null;default:500" json:"threshold_meters"`
	AlertLevel            int                 `gorm:"not null;default:2" json:"alert_level"`
	Status                GeoFenceAlertStatus `gorm:"type:varchar(20);index;not null;default:pending" json:"status"`
	DeviateReason         *DeviateReason      `gorm:"type:varchar(32)" json:"deviate_reason"`
	ConfirmNote           string              `gorm:"type:varchar(1024)" json:"confirm_note"`
	ConfirmedBy           *int64              `gorm:"index" json:"confirmed_by"`
	ConfirmedRole         string              `gorm:"type:varchar(20)" json:"confirmed_role"`
	ConfirmedAt           *time.Time          `json:"confirmed_at"`
	ReportedToDispatch    bool                `gorm:"not null;default:false" json:"reported_to_dispatch"`
	ReportedAt            *time.Time          `json:"reported_at"`
	ResolvedBy            *int64              `json:"resolved_by"`
	ResolvedNote          string              `gorm:"type:varchar(1024)" json:"resolved_note"`
	ResolvedAt            *time.Time          `json:"resolved_at"`
	DailyDeviateCount     int                 `gorm:"not null;default:0" json:"daily_deviate_count"`
	NearestRoutePoint     *JSON               `gorm:"type:json" json:"nearest_route_point"`
	SnapshotURL           string              `gorm:"type:varchar(512)" json:"snapshot_url"`
	PopupDisplayed        bool                `gorm:"not null;default:false" json:"popup_displayed"`
	NotifiedEscort        bool                `gorm:"not null;default:false" json:"notified_escort"`
}

type GeoFenceConfirmLog struct {
	BaseModel
	AlertID        int64         `gorm:"index;not null" json:"alert_id"`
	AlertNo        string        `gorm:"type:varchar(32);not null" json:"alert_no"`
	VehicleID      int64         `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber    string        `gorm:"type:varchar(20)" json:"plate_number"`
	WaybillID      int64         `gorm:"index" json:"waybill_id"`
	WaybillNo      string        `gorm:"type:varchar(32)" json:"waybill_no"`
	ConfirmType    DeviateReason `gorm:"type:varchar(32);not null" json:"confirm_type"`
	ReasonDetail   string        `gorm:"type:varchar(512)" json:"reason_detail"`
	Note           string        `gorm:"type:varchar(1024)" json:"note"`
	ConfirmedBy    int64         `gorm:"index;not null" json:"confirmed_by"`
	ConfirmedName  string        `gorm:"type:varchar(64)" json:"confirmed_name"`
	ConfirmedRole  string        `gorm:"type:varchar(20);not null" json:"confirmed_role"`
	Latitude       *float64      `json:"latitude"`
	Longitude      *float64      `json:"longitude"`
}

type GeoFenceCheckRequest struct {
	VehicleID int64   `json:"vehicle_id" binding:"required"`
	DriverID  int64   `json:"driver_id"`
	WaybillID int64   `json:"waybill_id"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   string  `json:"address"`
	Speed     float64 `json:"speed"`
	Threshold int     `json:"threshold_meters"`
}

type GeoFenceCheckResult struct {
	AlertID             int64              `json:"alert_id"`
	AlertNo             string             `json:"alert_no"`
	IsDeviated          bool               `json:"is_deviated"`
	DistanceFromRouteM  int                `json:"distance_from_route_meters"`
	ThresholdMeters     int                `json:"threshold_meters"`
	AlertLevel          int                `json:"alert_level"`
	DailyDeviateCount   int                `json:"daily_deviate_count"`
	AutoReported        bool               `json:"auto_reported"`
	Status              GeoFenceAlertStatus `json:"status"`
	NearestRoutePoint   *GeoPoint          `json:"nearest_route_point"`
	Message             string             `json:"message"`
}

type GeoFenceConfirmRequest struct {
	AlertID       int64         `json:"alert_id" binding:"required"`
	ConfirmType   DeviateReason `json:"confirm_type" binding:"required"`
	ReasonDetail  string        `json:"reason_detail"`
	Note          string        `json:"note"`
	Latitude      float64       `json:"latitude"`
	Longitude     float64       `json:"longitude"`
}

type GeoFenceResolveRequest struct {
	AlertID      int64  `json:"alert_id" binding:"required"`
	ResolvedNote string `json:"resolved_note" binding:"required"`
}

type GeoFenceListRequest struct {
	VehicleID int64               `json:"vehicle_id" form:"vehicle_id"`
	WaybillID int64               `json:"waybill_id" form:"waybill_id"`
	EscortID  int64               `json:"escort_id" form:"escort_id"`
	Status    GeoFenceAlertStatus `json:"status" form:"status"`
	Page      int                 `json:"page" form:"page"`
	PageSize  int                 `json:"page_size" form:"page_size"`
}

type GeoFenceStats struct {
	TotalAlerts       int `json:"total_alerts"`
	PendingAlerts     int `json:"pending_alerts"`
	TodayAlerts       int `json:"today_alerts"`
	ReportedAlerts    int `json:"reported_alerts"`
	ResolvedAlerts    int `json:"resolved_alerts"`
	TotalConfirmLogs  int `json:"total_confirm_logs"`
	DetourCount       int `json:"detour_count"`
	DeviateCount      int `json:"deviate_count"`
	AutoReportedCount int `json:"auto_reported_count"`
}

func (GeoFenceAlert) TableName() string {
	return "geo_fence_alerts"
}

func (GeoFenceConfirmLog) TableName() string {
	return "geo_fence_confirm_logs"
}
