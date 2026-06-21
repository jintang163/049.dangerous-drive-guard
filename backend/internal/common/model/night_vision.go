package model

import "time"

type EnhanceMode string

const (
	EnhanceModeAuto     EnhanceMode = "auto"
	EnhanceModeNight    EnhanceMode = "night"
	EnhanceModeInfrared EnhanceMode = "infrared"
	EnhanceModeLowLight EnhanceMode = "low_light"
	EnhanceModeManual   EnhanceMode = "manual"
)

type InfraredAction string

const (
	InfraredActionTurnOn       InfraredAction = "turn_on"
	InfraredActionTurnOff      InfraredAction = "turn_off"
	InfraredActionIntensityChg InfraredAction = "intensity_change"
	InfraredActionAutoTrigger  InfraredAction = "auto_trigger"
	InfraredActionManualTrig   InfraredAction = "manual_trigger"
)

type TriggerType string

const (
	TriggerTypeAuto   TriggerType = "auto"
	TriggerTypeManual TriggerType = "manual"
	TriggerTypeSystem TriggerType = "system"
)

type NightVisionConfig struct {
	BaseModel
	VehicleID               int64       `gorm:"index;not null" json:"vehicle_id"`
	DeviceID                string      `gorm:"type:varchar(64);default:''" json:"device_id"`

	InfraredEnabled         bool        `gorm:"default:true" json:"infrared_enabled"`
	InfraredAutoMode        bool        `gorm:"default:true" json:"infrared_auto_mode"`
	InfraredManualOn        bool        `gorm:"default:false" json:"infrared_manual_on"`
	InfraredIntensity       int         `gorm:"type:tinyint unsigned;default:50" json:"infrared_intensity"`
	InfraredIntensityAuto   bool        `gorm:"default:true" json:"infrared_intensity_auto"`
	LowLightThresholdLux    int         `gorm:"default:50" json:"low_light_threshold_lux"`
	HighLightThresholdLux   int         `gorm:"default:200" json:"high_light_threshold_lux"`

	EnhancementEnabled      bool        `gorm:"default:true" json:"enhancement_enabled"`
	EnhanceMode             EnhanceMode `gorm:"type:varchar(32);default:'auto'" json:"enhance_mode"`
	GammaValue              float64     `gorm:"type:decimal(4,3);default:1.200" json:"gamma_value"`
	BrightnessBoost         int         `gorm:"type:tinyint;default:30" json:"brightness_boost"`
	ContrastBoost           int         `gorm:"type:tinyint;default:20" json:"contrast_boost"`
	HistogramEqualization   bool        `gorm:"default:true" json:"histogram_equalization"`
	ClaheEnabled            bool        `gorm:"default:true" json:"clahe_enabled"`
	DenoiseEnabled          bool        `gorm:"default:true" json:"denoise_enabled"`
	DenoiseStrength         int         `gorm:"type:tinyint unsigned;default:3" json:"denoise_strength"`
	SharpenEnabled          bool        `gorm:"default:false" json:"sharpen_enabled"`
	SharpenStrength         int         `gorm:"type:tinyint unsigned;default:2" json:"sharpen_strength"`

	NightModeAuto           bool        `gorm:"default:true" json:"night_mode_auto"`
	NightStartHour          int         `gorm:"type:tinyint;default:19" json:"night_start_hour"`
	NightEndHour            int         `gorm:"type:tinyint;default:6" json:"night_end_hour"`
	LowLightFaceDetect      bool        `gorm:"default:true" json:"low_light_face_detect"`
	MinFaceConfidenceNight  float64     `gorm:"type:decimal(5,4);default:0.4000" json:"min_face_confidence_night"`
}

func (NightVisionConfig) TableName() string {
	return "night_vision_configs"
}

type InfraredLightLog struct {
	BaseModel
	VehicleID            int64          `gorm:"index:idx_vehicle_time;not null" json:"vehicle_id"`
	DriverID             int64          `json:"driver_id"`
	DeviceID             string         `gorm:"type:varchar(64);default:''" json:"device_id"`

	Action               InfraredAction `gorm:"type:varchar(32);index;not null" json:"action"`
	TriggerType          TriggerType    `gorm:"type:varchar(32);index;default:'auto'" json:"trigger_type"`
	LightOn              bool           `gorm:"default:false" json:"light_on"`
	IntensityBefore      *int           `json:"intensity_before"`
	IntensityAfter       *int           `json:"intensity_after"`
	LightLevelLux        *int           `json:"light_level_lux"`

	Reason               string         `gorm:"type:varchar(256);default:''" json:"reason"`
	Latitude             float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude            float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	Timestamp            time.Time      `gorm:"index:idx_vehicle_time;not null" json:"timestamp"`

	FaceDetectedBefore   *bool          `json:"face_detected_before"`
	FaceDetectedAfter    *bool          `json:"face_detected_after"`
	ConfidenceBefore     *float64       `gorm:"type:decimal(5,4)" json:"confidence_before"`
	ConfidenceAfter      *float64       `gorm:"type:decimal(5,4)" json:"confidence_after"`
}

func (InfraredLightLog) TableName() string {
	return "infrared_light_logs"
}

type ImageEnhancementRecord struct {
	BaseModel
	VehicleID               int64       `gorm:"index:idx_vehicle_time;not null" json:"vehicle_id"`
	DriverID                int64       `json:"driver_id"`
	WaybillID               int64       `gorm:"index" json:"waybill_id"`
	DeviceID                string      `gorm:"type:varchar(64);default:''" json:"device_id"`

	OriginalImageURL        string      `gorm:"type:varchar(256);default:''" json:"original_image_url"`
	EnhancedImageURL        string      `gorm:"type:varchar(256);default:''" json:"enhanced_image_url"`

	EnhanceMode             EnhanceMode `gorm:"type:varchar(32);index;default:'night'" json:"enhance_mode"`
	GammaValue              *float64    `gorm:"type:decimal(4,3)" json:"gamma_value"`
	BrightnessDelta         int         `gorm:"default:0" json:"brightness_delta"`
	ContrastDelta           int         `gorm:"default:0" json:"contrast_delta"`
	DenoiseApplied          bool        `gorm:"default:false" json:"denoise_applied"`
	DenoiseStrength         int         `gorm:"type:tinyint unsigned;default:0" json:"denoise_strength"`
	HistogramEqApplied      bool        `gorm:"default:false" json:"histogram_eq_applied"`
	SharpenApplied          bool        `gorm:"default:false" json:"sharpen_applied"`

	OriginalBrightnessAvg   *int        `json:"original_brightness_avg"`
	EnhancedBrightnessAvg   *int        `json:"enhanced_brightness_avg"`
	OriginalContrast        *int        `json:"original_contrast"`
	EnhancedContrast        *int        `json:"enhanced_contrast"`

	LightLevelLux           *int        `json:"light_level_lux"`
	IsNightTime             bool        `gorm:"index;default:false" json:"is_night_time"`

	FaceDetectedOriginal    bool        `gorm:"default:false" json:"face_detected_original"`
	FaceDetectedEnhanced    bool        `gorm:"default:true" json:"face_detected_enhanced"`
	FaceConfidenceOriginal  float64     `gorm:"type:decimal(5,4);default:0" json:"face_confidence_original"`
	FaceConfidenceEnhanced  float64     `gorm:"type:decimal(5,4);default:0" json:"face_confidence_enhanced"`
	LandmarkCountOriginal   int         `gorm:"default:0" json:"landmark_count_original"`
	LandmarkCountEnhanced   int         `gorm:"default:0" json:"landmark_count_enhanced"`

	QualityScoreBefore      float64     `gorm:"type:decimal(5,4);default:0" json:"quality_score_before"`
	QualityScoreAfter       float64     `gorm:"type:decimal(5,4);default:0" json:"quality_score_after"`
	QualityImprovementPct   float64     `gorm:"type:decimal(5,2);default:0" json:"quality_improvement_pct"`

	ProcessingTimeMs        int         `gorm:"default:0" json:"processing_time_ms"`
	ProcessOnEdge           bool        `gorm:"default:true" json:"process_on_edge"`

	Timestamp               time.Time   `gorm:"index:idx_vehicle_time;not null" json:"timestamp"`
}

func (ImageEnhancementRecord) TableName() string {
	return "image_enhancement_records"
}

type NightVisionConfigUpdateRequest struct {
	VehicleID             int64   `json:"vehicle_id" binding:"required"`
	InfraredEnabled       *bool   `json:"infrared_enabled"`
	InfraredAutoMode      *bool   `json:"infrared_auto_mode"`
	InfraredManualOn      *bool   `json:"infrared_manual_on"`
	InfraredIntensity     *int    `json:"infrared_intensity"`
	InfraredIntensityAuto *bool   `json:"infrared_intensity_auto"`
	LowLightThresholdLux  *int    `json:"low_light_threshold_lux"`
	HighLightThresholdLux *int    `json:"high_light_threshold_lux"`

	EnhancementEnabled    *bool   `json:"enhancement_enabled"`
	EnhanceMode           *string `json:"enhance_mode"`
	GammaValue            *float64 `json:"gamma_value"`
	BrightnessBoost       *int    `json:"brightness_boost"`
	ContrastBoost         *int    `json:"contrast_boost"`
	HistogramEqualization *bool   `json:"histogram_equalization"`
	ClaheEnabled          *bool   `json:"clahe_enabled"`
	DenoiseEnabled        *bool   `json:"denoise_enabled"`
	DenoiseStrength       *int    `json:"denoise_strength"`
	SharpenEnabled        *bool   `json:"sharpen_enabled"`
	SharpenStrength       *int    `json:"sharpen_strength"`

	NightModeAuto         *bool   `json:"night_mode_auto"`
	NightStartHour        *int    `json:"night_start_hour"`
	NightEndHour          *int    `json:"night_end_hour"`
	LowLightFaceDetect    *bool   `json:"low_light_face_detect"`
	MinFaceConfidenceNight *float64 `json:"min_face_confidence_night"`
}

type InfraredLightStatus struct {
	VehicleID     int64  `json:"vehicle_id"`
	DeviceID      string `json:"device_id"`
	LightOn       bool   `json:"light_on"`
	Intensity     int    `json:"intensity"`
	IsAutoMode    bool   `json:"is_auto_mode"`
	LightLevelLux int    `json:"light_level_lux"`
	IsNightTime   bool   `json:"is_night_time"`
	Reason        string `json:"reason,omitempty"`
}

type ImageEnhanceRequest struct {
	VehicleID       int64  `json:"vehicle_id" binding:"required"`
	DriverID        int64  `json:"driver_id"`
	WaybillID       int64  `json:"waybill_id"`
	DeviceID        string `json:"device_id"`
	ImageBase64     string `json:"image_base64" binding:"required"`
	ImageURL        string `json:"image_url"`
	EnhanceMode     string `json:"enhance_mode"`
	LightLevelLux   *int   `json:"light_level_lux"`
	IsNightTime     *bool  `json:"is_night_time"`
	ApplyDenoise    *bool  `json:"apply_denoise"`
	ApplySharpen    *bool  `json:"apply_sharpen"`
	Timestamp       string `json:"timestamp"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type ImageEnhanceResult struct {
	EnhancedImageBase64 string  `json:"enhanced_image_base64"`
	EnhancedImageURL    string  `json:"enhanced_image_url"`
	EnhanceMode         string  `json:"enhance_mode"`
	GammaValue          float64 `json:"gamma_value"`
	BrightnessDelta     int     `json:"brightness_delta"`
	ContrastDelta       int     `json:"contrast_delta"`
	DenoiseApplied      bool    `json:"denoise_applied"`
	HistogramEqApplied  bool    `json:"histogram_eq_applied"`
	SharpenApplied      bool    `json:"sharpen_applied"`

	QualityScoreBefore  float64 `json:"quality_score_before"`
	QualityScoreAfter   float64 `json:"quality_score_after"`
	QualityImprovement  float64 `json:"quality_improvement_pct"`

	OriginalBrightness  int     `json:"original_brightness_avg"`
	EnhancedBrightness  int     `json:"enhanced_brightness_avg"`

	ProcessingTimeMs    int     `json:"processing_time_ms"`
	RecordID            int64   `json:"record_id"`
}

type NightVisionStats struct {
	TotalConfigs              int     `json:"total_configs"`
	InfraredEnabledCount      int     `json:"infrared_enabled_count"`
	EnhancementEnabledCount   int     `json:"enhancement_enabled_count"`

	TodayInfraredTurnOnCount  int     `json:"today_infrared_turn_on_count"`
	TodayInfraredTurnOffCount int     `json:"today_infrared_turn_off_count"`
	TodayInfraredDurationMin  float64 `json:"today_infrared_duration_minutes"`

	TotalEnhanceRecords       int64   `json:"total_enhance_records"`
	TodayEnhanceRecords       int     `json:"today_enhance_records"`
	AvgQualityImprovement     float64 `json:"avg_quality_improvement_pct"`
	AvgProcessingTimeMs       int     `json:"avg_processing_time_ms"`

	NightFaceDetectRateBefore float64 `json:"night_face_detect_rate_before"`
	NightFaceDetectRateAfter  float64 `json:"night_face_detect_rate_after"`
	AvgConfidenceBefore       float64 `json:"avg_face_confidence_before"`
	AvgConfidenceAfter        float64 `json:"avg_face_confidence_after"`
	ConfidenceImprovement     float64 `json:"confidence_improvement_pct"`

	AutoTriggerCount          int     `json:"auto_trigger_count"`
	ManualTriggerCount        int     `json:"manual_trigger_count"`
}

type NightVisionListRequest struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
	VehicleID int64 `form:"vehicle_id"`
}
