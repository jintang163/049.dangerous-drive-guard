package model

import (
	"time"
)

type EscortShiftStatus string

const (
	EscortShiftScheduled EscortShiftStatus = "scheduled"
	EscortShiftActive    EscortShiftStatus = "active"
	EscortShiftCompleted EscortShiftStatus = "completed"
	EscortShiftCancelled EscortShiftStatus = "cancelled"
)

type EscortSOSStatus string

const (
	EscortSOSPending    EscortSOSStatus = "pending"
	EscortSOSProcessing EscortSOSStatus = "processing"
	EscortSOSResolved   EscortSOSStatus = "resolved"
	EscortSOSIgnored    EscortSOSStatus = "ignored"
)

type EscortVideoRecordType string

const (
	VideoTypeScheduled EscortVideoRecordType = "scheduled"
	VideoTypeAlarm     EscortVideoRecordType = "alarm"
	VideoTypeManual    EscortVideoRecordType = "manual"
)

type EscortShift struct {
	BaseModel
	ShiftNo          string            `gorm:"type:varchar(32);uniqueIndex;not null" json:"shift_no"`
	EscortID         int64             `gorm:"index;not null" json:"escort_id"`
	EscortName       string            `gorm:"type:varchar(64)" json:"escort_name"`
	DispatcherID     int64             `gorm:"index;not null" json:"dispatcher_id"`
	DispatcherName   string            `gorm:"type:varchar(64)" json:"dispatcher_name"`
	VehicleIDs       string            `gorm:"type:varchar(512)" json:"vehicle_ids"`
	WaybillIDs       string            `gorm:"type:varchar(512)" json:"waybill_ids"`
	ScheduledStart   *time.Time        `json:"scheduled_start"`
	ScheduledEnd     *time.Time        `json:"scheduled_end"`
	ActualStart      *time.Time        `json:"actual_start"`
	ActualEnd        *time.Time        `json:"actual_end"`
	Status           EscortShiftStatus `gorm:"type:varchar(20);index;default:scheduled" json:"status"`
	Remark           string            `gorm:"type:varchar(512)" json:"remark"`
	MaxConcurrent    int               `gorm:"default:5" json:"max_concurrent"`
	PollingInterval  int               `gorm:"default:30" json:"polling_interval"`
}

type EscortVehicleAssignment struct {
	BaseModel
	ShiftID      int64  `gorm:"index;not null" json:"shift_id"`
	VehicleID    int64  `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber  string `gorm:"type:varchar(20)" json:"plate_number"`
	WaybillID    int64  `gorm:"index" json:"waybill_id"`
	WaybillNo    string `gorm:"type:varchar(32)" json:"waybill_no"`
	Priority     int    `gorm:"default:1" json:"priority"`
	AssignedBy   int64  `json:"assigned_by"`
	AssignedAt   *time.Time `json:"assigned_at"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
}

type EscortSOSAlert struct {
	BaseModel
	AlertNo        string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"alert_no"`
	VehicleID      int64          `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber    string         `gorm:"type:varchar(20)" json:"plate_number"`
	DriverID       int64          `gorm:"index" json:"driver_id"`
	DriverName     string         `gorm:"type:varchar(64)" json:"driver_name"`
	WaybillID      int64          `gorm:"index" json:"waybill_id"`
	WaybillNo      string         `gorm:"type:varchar(32)" json:"waybill_no"`
	AlertType      string         `gorm:"type:varchar(32);not null" json:"alert_type"`
	AlertLevel     int            `gorm:"default:3" json:"alert_level"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	Address        string         `gorm:"type:varchar(512)" json:"address"`
	Description    string         `gorm:"type:text" json:"description"`
	SnapshotURL    string         `gorm:"type:varchar(512)" json:"snapshot_url"`
	VideoClipURL   string         `gorm:"type:varchar(512)" json:"video_clip_url"`
	Status         EscortSOSStatus `gorm:"type:varchar(20);index;default:pending" json:"status"`
	HandledBy      int64          `gorm:"index" json:"handled_by"`
	HandlerName    string         `gorm:"type:varchar(64)" json:"handler_name"`
	HandledAt      *time.Time     `json:"handled_at"`
	HandleNote     string         `gorm:"type:varchar(1024)" json:"handle_note"`
	HandleType     string         `gorm:"type:varchar(32)" json:"handle_type"`
	Notified       bool           `gorm:"default:false" json:"notified"`
	PopupDisplayed bool           `gorm:"default:false" json:"popup_displayed"`
	AckedAt        *time.Time     `json:"acked_at"`
	EscortID       int64          `gorm:"index" json:"escort_id"`
}

type EscortVideoRecord struct {
	BaseModel
	RecordNo      string              `gorm:"type:varchar(32);uniqueIndex;not null" json:"record_no"`
	VehicleID     int64               `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber   string              `gorm:"type:varchar(20)" json:"plate_number"`
	WaybillID     int64               `gorm:"index" json:"waybill_id"`
	WaybillNo     string              `gorm:"type:varchar(32)" json:"waybill_no"`
	RecordType    EscortVideoRecordType `gorm:"type:varchar(20);index" json:"record_type"`
	VideoURL      string              `gorm:"type:varchar(512);not null" json:"video_url"`
	SnapshotURL   string              `gorm:"type:varchar(512)" json:"snapshot_url"`
	StartTime     *time.Time          `json:"start_time"`
	EndTime       *time.Time          `json:"end_time"`
	Duration      int                 `json:"duration"`
	Latitude      float64             `json:"latitude"`
	Longitude     float64             `json:"longitude"`
	TriggerReason string              `gorm:"type:varchar(256)" json:"trigger_reason"`
	AlertID       int64               `gorm:"index" json:"alert_id"`
	ViewedCount   int                 `gorm:"default:0" json:"viewed_count"`
	ExpireAt      *time.Time          `gorm:"index" json:"expire_at"`
}

type EscortIntercomLog struct {
	BaseModel
	VehicleID    int64      `gorm:"index;not null" json:"vehicle_id"`
	PlateNumber  string     `gorm:"type:varchar(20)" json:"plate_number"`
	SenderID     int64      `gorm:"index;not null" json:"sender_id"`
	SenderName   string     `gorm:"type:varchar(64)" json:"sender_name"`
	SenderRole   string     `gorm:"type:varchar(20)" json:"sender_role"`
	MessageType  string     `gorm:"type:varchar(20);default:text" json:"message_type"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	AudioURL     string     `gorm:"type:varchar(512)" json:"audio_url"`
	Priority     int        `gorm:"default:1" json:"priority"`
	Delivered    bool       `gorm:"default:false" json:"delivered"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	Acked        bool       `gorm:"default:false" json:"acked"`
	AckedAt      *time.Time `json:"acked_at"`
}

type EscortPollingSession struct {
	BaseModel
	SessionNo    string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"session_no"`
	EscortID     int64      `gorm:"index;not null" json:"escort_id"`
	EscortName   string     `gorm:"type:varchar(64)" json:"escort_name"`
	ShiftID      int64      `gorm:"index" json:"shift_id"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Vehicles     string     `gorm:"type:text" json:"vehicles"`
	PollingCount int        `gorm:"default:0" json:"polling_count"`
	Status       string     `gorm:"type:varchar(20);default:active" json:"status"`
}

type EscortShiftCreateRequest struct {
	EscortID        int64      `json:"escort_id" binding:"required"`
	DispatcherID    int64      `json:"dispatcher_id"`
	VehicleIDs      []int64    `json:"vehicle_ids"`
	WaybillIDs      []int64    `json:"waybill_ids"`
	ScheduledStart  *time.Time `json:"scheduled_start"`
	ScheduledEnd    *time.Time `json:"scheduled_end"`
	Remark          string     `json:"remark"`
	MaxConcurrent   int        `json:"max_concurrent"`
	PollingInterval int        `json:"polling_interval"`
}

type EscortSOSReportRequest struct {
	VehicleID   int64   `json:"vehicle_id" binding:"required"`
	DriverID    int64   `json:"driver_id"`
	WaybillID   int64   `json:"waybill_id"`
	AlertType   string  `json:"alert_type" binding:"required"`
	AlertLevel  int     `json:"alert_level"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	Address     string  `json:"address"`
	Description string  `json:"description"`
	SnapshotURL string  `json:"snapshot_url"`
}

type EscortSOSHandleRequest struct {
	AlertID    int64  `json:"alert_id" binding:"required"`
	HandleNote string `json:"handle_note" binding:"required"`
	HandleType string `json:"handle_type"`
}

type EscortIntercomRequest struct {
	VehicleID   int64  `json:"vehicle_id" binding:"required"`
	MessageType string `json:"message_type"`
	Content     string `json:"content" binding:"required"`
	AudioURL    string `json:"audio_url"`
	Priority    int    `json:"priority"`
}

type EscortTrackPlaybackRequest struct {
	WaybillID   int64      `json:"waybill_id" binding:"required"`
	VehicleID   int64      `json:"vehicle_id"`
	StartTime   *time.Time `json:"start_time" binding:"required"`
	EndTime     *time.Time `json:"end_time" binding:"required"`
}

type EscortVideoListRequest struct {
	VehicleID  int64  `json:"vehicle_id"`
	WaybillID  int64  `json:"waybill_id"`
	RecordType string `json:"record_type"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

type VoiceIntercomRequest struct {
	VehicleID int64  `json:"vehicle_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Priority  int    `json:"priority"`
}

type DispatchServiceAreaRequest struct {
	VehicleID     int64  `json:"vehicle_id" binding:"required"`
	ServiceAreaID int64  `json:"service_area_id" binding:"required"`
	Reason        string `json:"reason"`
	RestDuration  int    `json:"rest_duration"`
}

type RealtimeVehicleStatus struct {
	VehicleID      int64         `json:"vehicle_id"`
	PlateNumber    string        `json:"plate_number"`
	VehicleType    VehicleType   `json:"vehicle_type"`
	Status         VehicleStatus `json:"status"`
	DriverID       int64         `json:"driver_id"`
	DriverName     string        `json:"driver_name"`
	WaybillID      int64         `json:"waybill_id"`
	WaybillNo      string        `json:"waybill_no"`
	Latitude       float64       `json:"latitude"`
	Longitude      float64       `json:"longitude"`
	CurrentAddress string        `json:"current_address"`
	Speed          float64       `json:"speed"`
	FatigueScore   float64       `json:"fatigue_score"`
	FatigueLevel   string        `json:"fatigue_level"`
	AlertCount     int           `json:"alert_count"`
	MarkerColor    string        `json:"marker_color"`
	LastUpdateTime time.Time     `json:"last_update_time"`
	GPSTime        time.Time     `json:"gps_time"`
	GoodsName      string        `json:"goods_name"`
	DangerGoods    string        `json:"danger_goods"`
	DriverStatus   string        `json:"driver_status"`
	CoverURL       string        `json:"cover_url"`
	FrameURL       string        `json:"frame_url"`
	VideoURL       string        `json:"video_url"`
	LiveURL        string        `json:"live_url"`
	Location       string        `json:"location"`
}

type EscortEvent struct {
	BaseModel
	WaybillID      int64  `json:"waybill_id"`
	VehicleID      int64  `json:"vehicle_id"`
	EventType      string `json:"event_type"`
	EventLevel     int    `json:"event_level"`
	ReporterID     int64  `json:"reporter_id"`
	ReporterRole   string `json:"reporter_role"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Address        string `json:"address"`
	Description    string `json:"description"`
	EventTime      time.Time `json:"event_time"`
	HandledStatus  int    `json:"handled_status"`
	HandledBy      int64  `json:"handled_by"`
	HandledAt      *time.Time `json:"handled_at"`
	HandleNote     string `json:"handle_note"`
}

type VehicleTrack struct {
	BaseModel
	VehicleID      int64     `json:"vehicle_id"`
	WaybillID      int64     `json:"waybill_id"`
	DriverID       int64     `json:"driver_id"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Altitude       float64   `json:"altitude"`
	Speed          float64   `json:"speed"`
	Direction      int       `json:"direction"`
	SatelliteCount int       `json:"satellite_count"`
	Hdop           float64   `json:"hdop"`
	Accuracy       float64   `json:"accuracy"`
	GPSTime        time.Time `json:"gps_time"`
}

type FatigueAlarm struct {
	BaseModel
	AlarmNo           string  `json:"alarm_no"`
	VehicleID         int64   `json:"vehicle_id"`
	DriverID          int64   `json:"driver_id"`
	WaybillID         int64   `json:"waybill_id"`
	DetectionRecordID int64   `json:"detection_record_id"`
	AlarmType         string  `json:"alarm_type"`
	AlarmLevel        int     `json:"alarm_level"`
	FatigueScore      float64 `json:"fatigue_score"`
	ContinuousFatigueMinutes int `json:"continuous_fatigue_minutes"`
	SnapImageURL      string  `json:"snap_image_url"`
	VideoClipURL      string  `json:"video_clip_url"`
	VideoStartTime    *time.Time `json:"video_start_time"`
	VideoEndTime      *time.Time `json:"video_end_time"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	LocationAddress   string  `json:"location_address"`
	VehicleSpeed      float64 `json:"vehicle_speed"`
	Status            string  `json:"status"`
	NotifyDriverResult string `json:"notify_driver_result"`
	DispatcherID      int64   `json:"dispatcher_id"`
	HandledAt         *time.Time `json:"handled_at"`
	HandleNote        string  `json:"handle_note"`
	HandleType        string  `json:"handle_type"`
	VehicleInformed   bool    `json:"vehicle_informed"`
	Escalated         bool    `json:"escalated"`
}

type AlarmLevel int

const (
	AlarmLevelInfo    AlarmLevel = 1
	AlarmLevelWarning AlarmLevel = 2
	AlarmLevelDanger  AlarmLevel = 3
)

type RescueRequest struct {
	BaseModel
	RequestNo        string  `json:"request_no"`
	WaybillID        int64   `json:"waybill_id"`
	VehicleID        int64   `json:"vehicle_id"`
	DriverID         int64   `json:"driver_id"`
	SOSType          string  `json:"sos_type"`
	SOSLevel         AlarmLevel `json:"sos_level"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Address          string  `json:"address"`
	Description      string  `json:"description"`
	CallerName       string  `json:"caller_name"`
	CallerPhone      string  `json:"caller_phone"`
	Status           string  `json:"status"`
	AssignedResourceID int64 `json:"assigned_resource_id"`
	DispatcherID     int64   `json:"dispatcher_id"`
	DispatchedAt     *time.Time `json:"dispatched_at"`
	ArrivedAt        *time.Time `json:"arrived_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	ResultNote       string  `json:"result_note"`
	Injuries         int     `json:"injuries"`
	Deaths           int     `json:"deaths"`
}

type RescueResource struct {
	BaseModel
	ResourceType   string  `json:"resource_type"`
	Name           string  `json:"name"`
	OrgName        string  `json:"org_name"`
	ContactPerson  string  `json:"contact_person"`
	ContactPhone   string  `json:"contact_phone"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	District       string  `json:"district"`
	Address        string  `json:"address"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	ServiceRadius  float64 `json:"service_radius"`
	ResponseTimeMinutes int `json:"response_time_minutes"`
	Status         int     `json:"status"`
	CurrentTaskCount int   `json:"current_task_count"`
	Rating         float64 `json:"rating"`
}

func (EscortShift) TableName() string {
	return "escort_shifts"
}

func (EscortVehicleAssignment) TableName() string {
	return "escort_vehicle_assignments"
}

func (EscortSOSAlert) TableName() string {
	return "escort_sos_alerts"
}

func (EscortVideoRecord) TableName() string {
	return "escort_video_records"
}

func (EscortIntercomLog) TableName() string {
	return "escort_intercom_logs"
}

func (EscortPollingSession) TableName() string {
	return "escort_polling_sessions"
}
