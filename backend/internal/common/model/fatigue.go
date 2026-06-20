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
}

type FatigueDetectionRecord struct {
	ID                  int64         `json:"id"`
	VehicleID           int64         `json:"vehicle_id"`
	DriverID            int64         `json:"driver_id"`
	WaybillID           int64         `json:"waybill_id"`
	FrameImageURL       string        `json:"frame_image_url"`
	VideoClipURL        string        `json:"video_clip_url"`
	Metrics             FatigueMetrics `json:"metrics"`
	FatigueScore        float64       `json:"fatigue_score"`
	FatigueLevel        FatigueLevel  `json:"fatigue_level"`
	IsAlarmTriggered    bool          `json:"is_alarm_triggered"`
	AlarmType           string        `json:"alarm_type"`
	DetectionTime       time.Time     `json:"detection_time"`
	Latitude            float64       `json:"latitude"`
	Longitude           float64       `json:"longitude"`
	VehicleSpeed        float64       `json:"vehicle_speed"`
	EdgeComputed        bool          `json:"edge_computed"`
	NetworkStatus       string        `json:"network_status"`
	CreatedAt           time.Time     `json:"created_at"`
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
	ID                      int64       `json:"id"`
	AlarmNo                 string      `json:"alarm_no"`
	VehicleID               int64       `json:"vehicle_id"`
	DriverID                int64       `json:"driver_id"`
	WaybillID               int64       `json:"waybill_id"`
	DetectionRecordID       int64       `json:"detection_record_id"`
	AlarmType               string      `json:"alarm_type"`
	AlarmLevel              AlarmLevel  `json:"alarm_level"`
	FatigueScore            float64     `json:"fatigue_score"`
	ContinuousFatigueMinutes int         `json:"continuous_fatigue_minutes"`
	SnapImageURL            string      `json:"snap_image_url"`
	VideoClipURL            string      `json:"video_clip_url"`
	VideoStartTime          *time.Time  `json:"video_start_time"`
	VideoEndTime            *time.Time  `json:"video_end_time"`
	Latitude                float64     `json:"latitude"`
	Longitude               float64     `json:"longitude"`
	LocationAddress         string      `json:"location_address"`
	VehicleSpeed            float64     `json:"vehicle_speed"`
	Status                  AlarmStatus `json:"status"`
	NotifyDriverResult      string      `json:"notify_driver_result"`
	DispatcherID            int64       `json:"dispatcher_id"`
	HandledAt               *time.Time  `json:"handled_at"`
	HandleNote              string      `json:"handle_note"`
	HandleType              string      `json:"handle_type"`
	VehicleInformed         bool        `json:"vehicle_informed"`
	Escalated               bool        `json:"escalated"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	VehiclePlate            string      `json:"vehicle_plate,omitempty"`
	DriverName              string      `json:"driver_name,omitempty"`
}

type DrivingScore struct {
	ID                         int64   `json:"id"`
	DriverID                   int64   `json:"driver_id"`
	WaybillID                  int64   `json:"waybill_id"`
	VehicleID                  int64   `json:"vehicle_id"`
	TotalScore                 float64 `json:"total_score"`
	ScoreLevel                 string  `json:"score_level"`
	FatigueDeduction           float64 `json:"fatigue_deduction"`
	OverspeedCount             int     `json:"overspeed_count"`
	OverspeedDeduction         float64 `json:"overspeed_deduction"`
	SuddenBrakeCount           int     `json:"sudden_brake_count"`
	SuddenBrakeDeduction       float64 `json:"sudden_brake_deduction"`
	SuddenAccelCount           int     `json:"sudden_accel_count"`
	SuddenAccelDeduction       float64 `json:"sudden_accel_deduction"`
	SharpTurnCount             int     `json:"sharp_turn_count"`
	SharpTurnDeduction         float64 `json:"sharp_turn_deduction"`
	LaneDeviationCount         int     `json:"lane_deviation_count"`
	LaneDeviationDeduction     float64 `json:"lane_deviation_deduction"`
	PhoneUsageCount            int     `json:"phone_usage_count"`
	PhoneUsageDeduction        float64 `json:"phone_usage_deduction"`
	SmokingCount               int     `json:"smoking_count"`
	SmokingDeduction           float64 `json:"smoking_deduction"`
	SeatbeltViolationCount     int     `json:"seatbelt_violation_count"`
	SeatbeltViolationDeduction float64 `json:"seatbelt_violation_deduction"`
	RouteDeviationCount        int     `json:"route_deviation_count"`
	RouteDeviationDeduction    float64 `json:"route_deviation_deduction"`
	FatigueAlarmCount          int     `json:"fatigue_alarm_count"`
	TotalDistance              float64 `json:"total_distance"`
	DrivingDuration            int     `json:"driving_duration"`
	NightDrivingDuration       int     `json:"night_driving_duration"`
}

type VehicleTrack struct {
	ID            int64     `json:"id"`
	VehicleID     int64     `json:"vehicle_id"`
	WaybillID     int64     `json:"waybill_id"`
	DriverID      int64     `json:"driver_id"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Altitude      float64   `json:"altitude"`
	Speed         float64   `json:"speed"`
	Direction     int       `json:"direction"`
	SatelliteCount int       `json:"satellite_count"`
	Accuracy      float64   `json:"accuracy"`
	GPSTime       time.Time `json:"gps_time"`
	CreatedAt     time.Time `json:"created_at"`
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
