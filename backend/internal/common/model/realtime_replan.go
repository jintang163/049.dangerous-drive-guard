package model

import (
	"time"
)

// ============================================
// 实时路况事件模型
// ============================================

type TrafficEventType string

const (
	TrafficEventCongestion  TrafficEventType = "congestion"
	TrafficEventAccident    TrafficEventType = "accident"
	TrafficEventRoadClosure TrafficEventType = "road_closure"
	TrafficEventConstruction TrafficEventType = "construction"
)

type TrafficEventStatus string

const (
	TrafficEventActive    TrafficEventStatus = "active"
	TrafficEventExpired   TrafficEventStatus = "expired"
	TrafficEventResolved  TrafficEventStatus = "resolved"
)

type TrafficEvent struct {
	ID              int64                `gorm:"primaryKey;autoIncrement" json:"id"`
	EventNo         string               `gorm:"size:64;uniqueIndex;not null" json:"event_no"`
	EventType       TrafficEventType     `gorm:"size:32;not null;index" json:"event_type"`
	EventLevel      int                  `gorm:"default:1;index" json:"event_level"`
	Title           string               `gorm:"size:256;not null" json:"title"`
	Description     string               `gorm:"size:1024" json:"description"`
	Source          string               `gorm:"size:32;default:system" json:"source"`
	RoadName        string               `gorm:"size:128" json:"road_name"`
	StartPoint      JSON                 `json:"start_point"`
	EndPoint        JSON                 `json:"end_point"`
	AffectedGeometry JSON                `json:"affected_geometry"`
	CenterLat       float64              `gorm:"index" json:"center_lat"`
	CenterLng       float64              `gorm:"index" json:"center_lng"`
	AffectedLengthKm float64             `gorm:"type:decimal(8,2);default:0" json:"affected_length_km"`
	CongestionLevel int                  `json:"congestion_level"`
	AvgSpeedKmh     float64              `gorm:"type:decimal(6,2)" json:"avg_speed_kmh"`
	DurationMinutes int                  `json:"duration_minutes"`
	StartedAt       time.Time            `gorm:"not null" json:"started_at"`
	ExpectedEndAt   *time.Time           `json:"expected_end_at"`
	ActualEndAt     *time.Time           `json:"actual_end_at"`
	Status          TrafficEventStatus   `gorm:"size:20;default:active;index" json:"status"`
	RelatedOfficialID string             `gorm:"size:128" json:"related_official_id"`
	ExtraInfo       JSON                 `json:"extra_info"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

func (TrafficEvent) TableName() string { return "traffic_events" }

// ============================================
// 重规划记录模型
// ============================================

type ReplanTriggerType string

const (
	ReplanTriggerTraffic    ReplanTriggerType = "traffic"
	ReplanTriggerDeviation  ReplanTriggerType = "deviation"
	ReplanTriggerRestricted ReplanTriggerType = "restricted"
	ReplanTriggerManual     ReplanTriggerType = "manual"
)

type ReplanRecordStatus string

const (
	ReplanStatusPending      ReplanRecordStatus = "pending"
	ReplanStatusConfirmed    ReplanRecordStatus = "confirmed"
	ReplanStatusRejected     ReplanRecordStatus = "rejected"
	ReplanStatusAutoApplied  ReplanRecordStatus = "auto_applied"
	ReplanStatusCancelled    ReplanRecordStatus = "cancelled"
)

type RouteReplanRecord struct {
	ID                       int64              `gorm:"primaryKey;autoIncrement" json:"id"`
	ReplanNo                 string             `gorm:"size:64;uniqueIndex;not null" json:"replan_no"`
	WaybillID                int64              `gorm:"not null;index" json:"waybill_id"`
	WaybillNo                string             `gorm:"size:64" json:"waybill_no"`
	VehicleID                int64              `gorm:"not null;index" json:"vehicle_id"`
	VehiclePlate             string             `gorm:"size:32" json:"vehicle_plate"`
	DriverID                 int64              `json:"driver_id"`
	DriverName               string             `gorm:"size:64" json:"driver_name"`
	OriginalRoutePlanID      int64              `json:"original_route_plan_id"`
	NewRoutePlanID           int64              `json:"new_route_plan_id"`
	TriggerType              ReplanTriggerType  `gorm:"size:32;not null;index" json:"trigger_type"`
	TriggerSourceID          int64              `json:"trigger_source_id"`
	TriggerReason            string             `gorm:"size:256;not null" json:"trigger_reason"`
	EventType                string             `gorm:"size:32" json:"event_type"`
	CurrentLat               float64            `gorm:"type:decimal(10,6);not null" json:"current_lat"`
	CurrentLng               float64            `gorm:"type:decimal(10,6);not null" json:"current_lng"`
	OriginalDistanceRemaining float64           `gorm:"type:decimal(8,2)" json:"original_distance_remaining"`
	OriginalDurationRemaining int               `json:"original_duration_remaining"`
	NewDistanceRemaining     float64            `gorm:"type:decimal(8,2)" json:"new_distance_remaining"`
	NewDurationRemaining     int                `json:"new_duration_remaining"`
	DistanceDelta            float64            `gorm:"type:decimal(8,2)" json:"distance_delta"`
	DurationDelta            int                `json:"duration_delta"`
	AvoidedTrafficIDs        JSON               `json:"avoided_traffic_ids"`
	Status                   ReplanRecordStatus `gorm:"size:20;default:pending;index" json:"status"`
	NotifiedAt               *time.Time         `json:"notified_at"`
	DriverConfirmAt          *time.Time         `json:"driver_confirm_at"`
	AppliedAt                *time.Time         `json:"applied_at"`
	ConfirmNote              string             `gorm:"size:256" json:"confirm_note"`
	OperatorID               int64              `json:"operator_id"`
	OperatorName             string             `gorm:"size:64" json:"operator_name"`
	ExtraInfo                JSON               `json:"extra_info"`
	CreatedAt                time.Time          `gorm:"index" json:"created_at"`
	UpdatedAt                time.Time          `json:"updated_at"`

	OriginalRoutePlan *RoutePlan              `gorm:"-" json:"original_route_plan,omitempty"`
	NewRoutePlan      *RoutePlan              `gorm:"-" json:"new_route_plan,omitempty"`
	CandidateRoutes   []*ReplanCandidateRoute `gorm:"-" json:"candidate_routes,omitempty"`
	TrafficEvents     []*TrafficEvent         `gorm:"-" json:"traffic_events,omitempty"`
}

func (RouteReplanRecord) TableName() string { return "route_replan_records" }

// ============================================
// 重规划候选路线模型
// ============================================

type ReplanCandidateRoute struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ReplanRecordID    int64     `gorm:"not null;index" json:"replan_record_id"`
	RouteGeometry     JSON      `json:"route_geometry"`
	RoutePath         JSON      `json:"route_path"`
	Strategy          string    `gorm:"size:20;not null" json:"strategy"`
	TotalDistance     float64   `gorm:"type:decimal(8,2)" json:"total_distance"`
	EstimatedDuration int       `json:"estimated_duration"`
	EstimatedDelay    int       `json:"estimated_delay"`
	TollFee           float64   `gorm:"type:decimal(10,2)" json:"toll_fee"`
	FuelCost          float64   `gorm:"type:decimal(10,2)" json:"fuel_cost"`
	SafetyScore       float64   `gorm:"type:decimal(5,2)" json:"safety_score"`
	PassTrafficEvents JSON      `json:"pass_traffic_events"`
	RestrictedSegments JSON     `json:"restricted_segments"`
	IsRecommended     int       `gorm:"default:0" json:"is_recommended"`
	RankOrder         int       `gorm:"default:0" json:"rank_order"`
	CreatedAt         time.Time `json:"created_at"`
}

func (ReplanCandidateRoute) TableName() string { return "replan_candidate_routes" }

// ============================================
// 请求/响应结构体
// ============================================

type TrafficEventCreateRequest struct {
	EventType       TrafficEventType `json:"event_type" binding:"required"`
	EventLevel      int              `json:"event_level"`
	Title           string           `json:"title" binding:"required"`
	Description     string           `json:"description"`
	Source          string           `json:"source"`
	RoadName        string           `json:"road_name"`
	CenterLat       float64          `json:"center_lat"`
	CenterLng       float64          `json:"center_lng"`
	AffectedGeometry JSON            `json:"affected_geometry"`
	AffectedLengthKm float64         `json:"affected_length_km"`
	CongestionLevel int              `json:"congestion_level"`
	AvgSpeedKmh     float64          `json:"avg_speed_kmh"`
	DurationMinutes int              `json:"duration_minutes"`
	StartedAt       *time.Time       `json:"started_at"`
	ExpectedEndAt   *time.Time       `json:"expected_end_at"`
}

type WebhookTrafficEvent struct {
	EventType          string   `json:"event_type" binding:"required"`
	EventLevel         int      `json:"event_level"`
	Title              string   `json:"title" binding:"required"`
	Description        string   `json:"description"`
	Source             string   `json:"source"`
	RoadName           string   `json:"road_name"`
	StartPoint         *JSON    `json:"start_point"`
	EndPoint           *JSON    `json:"end_point"`
	AffectedGeometry   *JSON    `json:"affected_geometry"`
	CenterLat          *float64 `json:"center_lat"`
	CenterLng          *float64 `json:"center_lng"`
	AffectedLengthKm   *float64 `json:"affected_length_km"`
	CongestionLevel    *int     `json:"congestion_level"`
	AvgSpeedKmh        *float64 `json:"avg_speed_kmh"`
	DurationMinutes    *int     `json:"duration_minutes"`
	StartedAt          string   `json:"started_at"`
	ExpectedEndAt      string   `json:"expected_end_at"`
	RelatedOfficialID  string   `json:"related_official_id"`
	ExtraInfo          *JSON    `json:"extra_info"`
}

type WebhookImportResponse struct {
	Accepted int               `json:"accepted"`
	Ignored  int               `json:"ignored"`
	Errors   map[string]string `json:"errors"`
}

type ReplanTriggerRequest struct {
	WaybillID       int64              `json:"waybill_id" binding:"required"`
	TriggerType     ReplanTriggerType  `json:"trigger_type" binding:"required"`
	TriggerSourceID int64              `json:"trigger_source_id"`
	TriggerReason   string             `json:"trigger_reason"`
	CurrentLat      float64            `json:"current_lat"`
	CurrentLng      float64            `json:"current_lng"`
	OperatorNote    string             `json:"operator_note"`
}

type ReplanConfirmRequest struct {
	Action      string `json:"action" binding:"required"`
	ConfirmNote string `json:"confirm_note"`
}

type ReplanQueryParams struct {
	WaybillID   int64              `json:"-" form:"waybill_id"`
	VehicleID   int64              `json:"-" form:"vehicle_id"`
	DriverID    int64              `json:"-" form:"driver_id"`
	TriggerType ReplanTriggerType  `json:"-" form:"trigger_type"`
	Status      ReplanRecordStatus `json:"-" form:"status"`
	Keyword     string             `json:"-" form:"keyword"`
	StartDate   string             `json:"-" form:"start_date"`
	EndDate     string             `json:"-" form:"end_date"`
	Page        int                `json:"-" form:"page,default=1"`
	PageSize    int                `json:"-" form:"page_size,default=20"`
}
