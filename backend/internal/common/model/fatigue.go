package model

import (
	"time"
)

type FaceLandmark struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z,omitempty"`
}

type FaceLandmarks struct {
	FaceDetected bool           `json:"face_detected"`
	LeftEye      []FaceLandmark `json:"left_eye"`
	RightEye     []FaceLandmark `json:"right_eye"`
	Mouth        []FaceLandmark `json:"mouth"`
	FaceOutline  []FaceLandmark `json:"face_outline"`
	Nose         []FaceLandmark `json:"nose"`
	LeftIris     FaceLandmark   `json:"left_iris"`
	RightIris    FaceLandmark   `json:"right_iris"`
}

type FatigueLevel string

const (
	FatigueNormal  FatigueLevel = "normal"
	FatigueWarning FatigueLevel = "warning"
	FatigueFatigue FatigueLevel = "fatigue"
)

type FatigueMetrics struct {
	PERCLOS            float64 `json:"perclos"`
	EyeClosedRatio     float64 `json:"eye_closed_ratio"`
	BlinkCount         int     `json:"blink_count"`
	BlinkFrequency     float64 `json:"blink_frequency"`
	YawnCount          int     `json:"yawn_count"`
	MouthOpenRatio     float64 `json:"mouth_open_ratio"`
	HeadPitch          float64 `json:"head_pitch"`
	HeadYaw            float64 `json:"head_yaw"`
	HeadRoll           float64 `json:"head_roll"`
	GazeDeviation      float64 `json:"gaze_deviation"`
	PhoneUsageDetected bool    `json:"phone_usage_detected"`
	SmokingDetected    bool    `json:"smoking_detected"`
	SeatbeltOn         bool    `json:"seatbelt_on"`
}

type CameraPosition string

const (
	CameraLeft   CameraPosition = "left"
	CameraCenter CameraPosition = "center"
	CameraRight  CameraPosition = "right"
)

type MultiCameraFrame struct {
	Position     CameraPosition `json:"position"`
	ImageURL     string         `json:"image_url"`
	ImageBase64  string         `json:"image_base64"`
	Landmarks    FaceLandmarks  `json:"landmarks"`
	Metrics      FatigueMetrics `json:"metrics"`
	FaceDetected bool           `json:"face_detected"`
	Confidence   float64        `json:"confidence"`
	Quality      float64        `json:"quality"`
	Occluded     bool           `json:"occluded"`
	Backlit      bool           `json:"backlit"`
}

type FusionResult struct {
	FatigueScore      float64      `json:"fatigue_score"`
	FatigueLevel      FatigueLevel `json:"fatigue_level"`
	FusionMethod      string       `json:"fusion_method"`
	UsedCameras       []string     `json:"used_cameras"`
	PrimaryCamera     string       `json:"primary_camera"`
	FusionConfidence  float64      `json:"fusion_confidence"`
	LeftScore         float64      `json:"left_score"`
	CenterScore       float64      `json:"center_score"`
	RightScore        float64      `json:"right_score"`
	OcclusionDetected bool         `json:"occlusion_detected"`
	BacklitDetected   bool         `json:"backlit_detected"`
}

type FatigueDetectRequest struct {
	VehicleID     int64          `json:"vehicle_id" binding:"required"`
	DriverID      int64          `json:"driver_id" binding:"required"`
	WaybillID     int64          `json:"waybill_id"`
	ImageURL      string         `json:"image_url"`
	ImageBase64   string         `json:"image_base64"`
	Landmarks     FaceLandmarks  `json:"landmarks"`
	Metrics       FatigueMetrics `json:"metrics"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	VehicleSpeed  float64        `json:"vehicle_speed"`
	DetectionTime time.Time      `json:"detection_time"`
	EdgeComputed  bool           `json:"edge_computed"`
	NetworkStatus string         `json:"network_status"`
	CameraPosition CameraPosition `json:"camera_position"`
	Frames        []MultiCameraFrame `json:"frames"`
	EnableFusion  bool           `json:"enable_fusion"`
}

type FatigueDetectResponse struct {
	FatigueScore  float64      `json:"fatigue_score"`
	FatigueLevel  FatigueLevel `json:"fatigue_level"`
	NeedAlarm     bool         `json:"need_alarm"`
	AlarmType     string       `json:"alarm_type,omitempty"`
	AlarmMessage  string       `json:"alarm_message,omitempty"`
	SeatbeltAlert bool         `json:"seatbelt_alert,omitempty"`
	PhoneAlert    bool         `json:"phone_alert,omitempty"`
	SmokingAlert  bool         `json:"smoking_alert,omitempty"`
	RecommendRest int          `json:"recommend_rest_minutes,omitempty"`
	FusionResult  *FusionResult `json:"fusion_result,omitempty"`
	CameraFrames  map[string]*MultiCameraFrame `json:"camera_frames,omitempty"`
}

type FatigueDetectionRecord struct {
	BaseModel
	VehicleID           int64          `gorm:"index;not null" json:"vehicle_id"`
	DriverID            int64          `gorm:"index;not null" json:"driver_id"`
	WaybillID           int64          `gorm:"index" json:"waybill_id"`
	FrameImageURL       string         `gorm:"type:varchar(256)" json:"frame_image_url"`
	VideoClipURL        string         `gorm:"type:varchar(256)" json:"video_clip_url"`
	Metrics             FatigueMetrics `gorm:"type:json" json:"metrics"`
	FatigueScore        float64        `gorm:"type:decimal(5,2)" json:"fatigue_score"`
	FatigueLevel        FatigueLevel   `gorm:"type:varchar(20)" json:"fatigue_level"`
	IsAlarmTriggered    bool           `gorm:"default:false" json:"is_alarm_triggered"`
	AlarmType           string         `gorm:"type:varchar(64)" json:"alarm_type"`
	DetectionTime       time.Time      `gorm:"index;not null" json:"detection_time"`
	Latitude            float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude           float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	VehicleSpeed        float64        `gorm:"type:decimal(5,2)" json:"vehicle_speed"`
	EdgeComputed        bool           `gorm:"default:false" json:"edge_computed"`
	NetworkStatus       string         `gorm:"type:varchar(20)" json:"network_status"`
	CameraPosition      string         `gorm:"type:varchar(20);default:center" json:"camera_position"`
	LeftFrameURL        string         `gorm:"type:varchar(256)" json:"left_frame_url"`
	CenterFrameURL      string         `gorm:"type:varchar(256)" json:"center_frame_url"`
	RightFrameURL       string         `gorm:"type:varchar(256)" json:"right_frame_url"`
	LeftScore           float64        `gorm:"type:decimal(5,2);default:0" json:"left_score"`
	CenterScore         float64        `gorm:"type:decimal(5,2);default:0" json:"center_score"`
	RightScore          float64        `gorm:"type:decimal(5,2);default:0" json:"right_score"`
	FusionMethod        string         `gorm:"type:varchar(32)" json:"fusion_method"`
	FusionConfidence    float64        `gorm:"type:decimal(5,4);default:0" json:"fusion_confidence"`
	OcclusionDetected   bool           `gorm:"default:false" json:"occlusion_detected"`
	BacklitDetected     bool           `gorm:"default:false" json:"backlit_detected"`
	UsedCameras         string         `gorm:"type:varchar(64)" json:"used_cameras"`
}

func (FatigueDetectionRecord) TableName() string {
	return "fatigue_detection_records"
}

type AlarmLevel int

const (
	AlarmLevelRemind AlarmLevel = 1
	AlarmLevelWarn   AlarmLevel = 2
	AlarmLevelSevere AlarmLevel = 3
)

type AlarmStatus string

const (
	AlarmStatusPending   AlarmStatus = "pending"
	AlarmStatusProcessing AlarmStatus = "processing"
	AlarmStatusAck       AlarmStatus = "acknowledged"
	AlarmStatusResolved  AlarmStatus = "resolved"
	AlarmStatusIgnored   AlarmStatus = "ignored"
)

type FatigueAlarm struct {
	BaseModel
	AlarmNo                 string      `gorm:"type:varchar(64);uniqueIndex;not null" json:"alarm_no"`
	VehicleID               int64       `gorm:"index;not null" json:"vehicle_id"`
	DriverID                int64       `gorm:"index;not null" json:"driver_id"`
	WaybillID               int64       `gorm:"index" json:"waybill_id"`
	DetectionRecordID       int64       `gorm:"index" json:"detection_record_id"`
	AlarmType               string      `gorm:"type:varchar(64)" json:"alarm_type"`
	AlarmLevel              AlarmLevel  `gorm:"index;not null" json:"alarm_level"`
	FatigueScore            float64     `gorm:"type:decimal(5,2)" json:"fatigue_score"`
	ContinuousFatigueMinutes int         `gorm:"default:0" json:"continuous_fatigue_minutes"`
	SnapImageURL            string      `gorm:"type:varchar(256)" json:"snap_image_url"`
	VideoClipURL            string      `gorm:"type:varchar(256)" json:"video_clip_url"`
	VideoStartTime          *time.Time  `json:"video_start_time"`
	VideoEndTime            *time.Time  `json:"video_end_time"`
	Latitude                float64     `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude               float64     `gorm:"type:decimal(10,7)" json:"longitude"`
	LocationAddress         string      `gorm:"type:varchar(256)" json:"location_address"`
	VehicleSpeed            float64     `gorm:"type:decimal(5,2)" json:"vehicle_speed"`
	Status                  AlarmStatus `gorm:"type:varchar(20);index;not null" json:"status"`
	NotifyDriverResult      string      `gorm:"type:varchar(256)" json:"notify_driver_result"`
	DispatcherID            int64       `gorm:"index" json:"dispatcher_id"`
	HandledAt               *time.Time  `json:"handled_at"`
	HandleNote              string      `gorm:"type:text" json:"handle_note"`
	HandleType              string      `gorm:"type:varchar(64)" json:"handle_type"`
	VehicleInformed         bool        `gorm:"default:false" json:"vehicle_informed"`
	Escalated               bool        `gorm:"default:false" json:"escalated"`
	VehiclePlate            string      `gorm:"-" json:"vehicle_plate,omitempty"`
	DriverName              string      `gorm:"-" json:"driver_name,omitempty"`
}

func (FatigueAlarm) TableName() string {
	return "fatigue_alarms"
}

type DrivingScore struct {
	BaseModel
	DriverID                   int64   `gorm:"index;not null" json:"driver_id"`
	WaybillID                  int64   `gorm:"index" json:"waybill_id"`
	VehicleID                  int64   `gorm:"index" json:"vehicle_id"`
	TripDate                   string  `gorm:"type:date;index" json:"trip_date"`
	TotalScore                 float64 `gorm:"type:decimal(5,2);not null" json:"total_score"`
	ScoreLevel                 string  `gorm:"type:varchar(20)" json:"score_level"`
	FatigueDeduction           float64 `gorm:"type:decimal(5,2);default:0" json:"fatigue_deduction"`
	OverspeedCount             int     `gorm:"default:0" json:"overspeed_count"`
	OverspeedDeduction         float64 `gorm:"type:decimal(5,2);default:0" json:"overspeed_deduction"`
	SuddenBrakeCount           int     `gorm:"default:0" json:"sudden_brake_count"`
	SuddenBrakeDeduction       float64 `gorm:"type:decimal(5,2);default:0" json:"sudden_brake_deduction"`
	SuddenAccelCount           int     `gorm:"default:0" json:"sudden_accel_count"`
	SuddenAccelDeduction       float64 `gorm:"type:decimal(5,2);default:0" json:"sudden_accel_deduction"`
	SharpTurnCount             int     `gorm:"default:0" json:"sharp_turn_count"`
	SharpTurnDeduction         float64 `gorm:"type:decimal(5,2);default:0" json:"sharp_turn_deduction"`
	LaneDeviationCount         int     `gorm:"default:0" json:"lane_deviation_count"`
	LaneDeviationDeduction     float64 `gorm:"type:decimal(5,2);default:0" json:"lane_deviation_deduction"`
	PhoneUsageCount            int     `gorm:"default:0" json:"phone_usage_count"`
	PhoneUsageDeduction        float64 `gorm:"type:decimal(5,2);default:0" json:"phone_usage_deduction"`
	SmokingCount               int     `gorm:"default:0" json:"smoking_count"`
	SmokingDeduction           float64 `gorm:"type:decimal(5,2);default:0" json:"smoking_deduction"`
	SeatbeltViolationCount     int     `gorm:"default:0" json:"seatbelt_violation_count"`
	SeatbeltViolationDeduction float64 `gorm:"type:decimal(5,2);default:0" json:"seatbelt_violation_deduction"`
	RouteDeviationCount        int     `gorm:"default:0" json:"route_deviation_count"`
	RouteDeviationDeduction    float64 `gorm:"type:decimal(5,2);default:0" json:"route_deviation_deduction"`
	CloseFollowingCount        int     `gorm:"default:0" json:"close_following_count"`
	CloseFollowingDeduction    float64 `gorm:"type:decimal(5,2);default:0" json:"close_following_deduction"`
	FatigueAlarmCount          int     `gorm:"default:0" json:"fatigue_alarm_count"`
	TotalDistance              float64 `gorm:"type:decimal(10,2);default:0" json:"total_distance"`
	DrivingDuration            int     `gorm:"default:0" json:"driving_duration"`
	NightDrivingDuration       int     `gorm:"default:0" json:"night_driving_duration"`
}

func (DrivingScore) TableName() string {
	return "driving_scores"
}

type VehicleTrack struct {
	BaseModel
	VehicleID      int64     `gorm:"index;not null" json:"vehicle_id"`
	WaybillID      int64     `gorm:"index" json:"waybill_id"`
	DriverID       int64     `gorm:"index" json:"driver_id"`
	Latitude       float64   `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude      float64   `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Altitude       float64   `gorm:"type:decimal(8,2)" json:"altitude"`
	Speed          float64   `gorm:"type:decimal(5,2)" json:"speed"`
	Direction      int       `json:"direction"`
	SatelliteCount int       `json:"satellite_count"`
	Accuracy       float64   `gorm:"type:decimal(5,2)" json:"accuracy"`
	GPSTime        time.Time `gorm:"index;not null" json:"gps_time"`
}

func (VehicleTrack) TableName() string {
	return "vehicle_tracks"
}

type RealtimeVehicleStatus struct {
	VehicleID        int64        `json:"vehicle_id"`
	PlateNumber      string       `json:"plate_number"`
	VehicleType      VehicleType  `json:"vehicle_type"`
	Status           VehicleStatus `json:"status"`
	DriverID         int64        `json:"driver_id"`
	DriverName       string       `json:"driver_name"`
	WaybillID        int64        `json:"waybill_id"`
	WaybillNo        string       `json:"waybill_no"`
	Latitude         float64      `json:"latitude"`
	Longitude        float64      `json:"longitude"`
	CurrentAddress   string       `json:"current_address"`
	Speed            float64      `json:"speed"`
	Direction        int          `json:"direction"`
	RemainingMileage float64      `json:"remaining_mileage"`
	RemainingTime    int          `json:"remaining_time"`
	FatigueScore     float64      `json:"fatigue_score"`
	FatigueLevel     FatigueLevel `json:"fatigue_level"`
	LastUpdateTime   time.Time    `json:"last_update_time"`
	MarkerColor      string       `json:"marker_color"`
	AlertCount       int          `json:"alert_count"`
	EngineStatus     string       `json:"engine_status"`
	FuelLevel        float64      `json:"fuel_level"`
	TirePressureOK   bool         `json:"tire_pressure_ok"`
	GPSTime          time.Time    `json:"gps_time"`
}

type AlarmAckRequest struct {
	AlarmID    int64  `json:"alarm_id" binding:"required"`
	HandleType string `json:"handle_type" binding:"required"`
	HandleNote string `json:"handle_note"`
	DispatchServiceAreaID int64 `json:"dispatch_service_area_id"`
	NotifyLegalStation bool  `json:"notify_legal_station"`
}

type VoiceIntercomRequest struct {
	VehicleID    int64  `json:"vehicle_id" binding:"required"`
	Message      string `json:"message" binding:"required"`
	MessageType  string `json:"message_type"`
	Priority     int    `json:"priority"`
}

type DispatchServiceAreaRequest struct {
	VehicleID     int64   `json:"vehicle_id" binding:"required"`
	ServiceAreaID int64   `json:"service_area_id" binding:"required"`
	Reason        string  `json:"reason" binding:"required"`
	RestDuration  int     `json:"rest_duration"`
}
