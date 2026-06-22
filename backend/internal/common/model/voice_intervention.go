package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type AudioCategory string

const (
	AudioCategoryFamily     AudioCategory = "family"
	AudioCategoryCustom     AudioCategory = "custom"
	AudioCategorySystem     AudioCategory = "system"
	AudioCategoryEmergency  AudioCategory = "emergency"
)

type VoiceInterventionAudio struct {
	BaseModel
	DriverID     int64         `gorm:"index" json:"driver_id"`
	OrgID        int64         `gorm:"index" json:"org_id"`
	Name         string        `gorm:"type:varchar(128);not null" json:"name"`
	Category     AudioCategory `gorm:"type:varchar(32);index;not null" json:"category"`
	AudioURL     string        `gorm:"type:varchar(512);not null" json:"audio_url"`
	AudioFormat  string        `gorm:"type:varchar(16);default:mp3" json:"audio_format"`
	DurationSec  int           `gorm:"default:0" json:"duration_sec"`
	FileSize     int64         `gorm:"default:0" json:"file_size"`
	Volume       int           `gorm:"default:80" json:"volume"`
	Description  string        `gorm:"type:varchar(512)" json:"description"`
	Tags         JSON          `gorm:"type:json" json:"tags"`
	IsDefault    bool          `gorm:"default:false" json:"is_default"`
	IsEnabled    bool          `gorm:"default:true" json:"is_enabled"`
	PlayCount    int64         `gorm:"default:0" json:"play_count"`
	CreatedBy    int64         `gorm:"index" json:"created_by"`
}

func (VoiceInterventionAudio) TableName() string {
	return "voice_intervention_audios"
}

type InterventionStrategyType string

const (
	StrategyTypeNormal     InterventionStrategyType = "normal"
	StrategyTypeContinuous InterventionStrategyType = "continuous"
	StrategyTypeSevere     InterventionStrategyType = "severe"
	StrategyTypeEmotional  InterventionStrategyType = "emotional"
)

type InterventionAlarmTrigger struct {
	AlarmLevels        []int  `json:"alarm_levels,omitempty"`
	AlarmTypes         []string `json:"alarm_types,omitempty"`
	MinContinuousMinutes int  `json:"min_continuous_minutes,omitempty"`
	MinFatigueScore    float64 `json:"min_fatigue_score,omitempty"`
}

type VoiceInterventionStrategy struct {
	BaseModel
	Name               string                    `gorm:"type:varchar(128);not null" json:"name"`
	StrategyType       InterventionStrategyType  `gorm:"type:varchar(32);index;not null" json:"strategy_type"`
	Priority           int                       `gorm:"default:1;index" json:"priority"`
	IsDefault          bool                      `gorm:"default:false" json:"is_default"`
	IsEnabled          bool                      `gorm:"default:true" json:"is_enabled"`
	DriverID           int64                     `gorm:"index;default:0" json:"driver_id"`
	OrgID              int64                     `gorm:"index;default:0" json:"org_id"`
	AlarmTrigger       InterventionAlarmTrigger  `gorm:"type:json" json:"alarm_trigger"`
	AudioIDs           JSON                      `gorm:"type:json" json:"audio_ids"`
	ForceHighVolume    bool                      `gorm:"default:false" json:"force_high_volume"`
	ForceVolumePercent int                       `gorm:"default:100" json:"force_volume_percent"`
	PlayTimes          int                       `gorm:"default:1" json:"play_times"`
	PlayIntervalSec    int                       `gorm:"default:5" json:"play_interval_sec"`
	ShuffleAudios      bool                      `gorm:"default:false" json:"shuffle_audios"`
	EmotionalMode      bool                      `gorm:"default:false" json:"emotional_mode"`
	CooldownSeconds    int                       `gorm:"default:30" json:"cooldown_seconds"`
	Description        string                    `gorm:"type:varchar(512)" json:"description"`
	CreatedBy          int64                     `json:"created_by"`
}

func (VoiceInterventionStrategy) TableName() string {
	return "voice_intervention_strategies"
}

type InterventionPlayStatus string

const (
	InterventionStatusPending   InterventionPlayStatus = "pending"
	InterventionStatusSent      InterventionPlayStatus = "sent"
	InterventionStatusPlaying   InterventionPlayStatus = "playing"
	InterventionStatusCompleted InterventionPlayStatus = "completed"
	InterventionStatusFailed    InterventionPlayStatus = "failed"
)

type VoiceInterventionLog struct {
	BaseModel
	VehicleID            int64                  `gorm:"index;not null" json:"vehicle_id"`
	DriverID             int64                  `gorm:"index;not null" json:"driver_id"`
	WaybillID            int64                  `gorm:"index" json:"waybill_id"`
	AlarmID              int64                  `gorm:"index" json:"alarm_id"`
	StrategyID           int64                  `gorm:"index" json:"strategy_id"`
	AudioID              int64                  `gorm:"index" json:"audio_id"`
	AudioName            string                 `gorm:"type:varchar(128)" json:"audio_name"`
	AudioURL             string                 `gorm:"type:varchar(512)" json:"audio_url"`
	AudioFormat          string                 `gorm:"type:varchar(16)" json:"audio_format"`
	Category             AudioCategory          `gorm:"type:varchar(32)" json:"category"`
	StrategyType         InterventionStrategyType `gorm:"type:varchar(32)" json:"strategy_type"`
	PlayStatus           InterventionPlayStatus `gorm:"type:varchar(32);index" json:"play_status"`
	IsHighVolume         bool                   `gorm:"default:false" json:"is_high_volume"`
	ActualVolumePercent  int                    `gorm:"default:0" json:"actual_volume_percent"`
	PlayTimes            int                    `gorm:"default:0" json:"play_times"`
	TotalPlayDurationMs  int64                  `gorm:"default:0" json:"total_play_duration_ms"`
	AlarmLevel           int                    `gorm:"default:0" json:"alarm_level"`
	AlarmType            string                 `gorm:"type:varchar(64)" json:"alarm_type"`
	FatigueScore         float64                `gorm:"type:decimal(5,2)" json:"fatigue_score"`
	ContinuousMinutes    int                    `gorm:"default:0" json:"continuous_minutes"`
	DriverAck            bool                   `gorm:"default:false" json:"driver_ack"`
	AckAt                *time.Time             `json:"ack_at"`
	SentAt               *time.Time             `json:"sent_at"`
	CompletedAt          *time.Time             `json:"completed_at"`
	ErrorMsg             string                 `gorm:"type:varchar(512)" json:"error_msg"`
	MQMessageID          string                 `gorm:"type:varchar(128)" json:"mq_message_id"`
	VehiclePlate         string                 `gorm:"-" json:"vehicle_plate,omitempty"`
	DriverName           string                 `gorm:"-" json:"driver_name,omitempty"`
}

func (VoiceInterventionLog) TableName() string {
	return "voice_intervention_logs"
}

type AudioTestPlayRequest struct {
	VehicleID int64 `json:"vehicle_id" binding:"required"`
	AudioID   int64 `json:"audio_id" binding:"required"`
	Volume    int   `json:"volume"`
}

type InterventionMatchedResult struct {
	Matched        bool                      `json:"matched"`
	Strategy       *VoiceInterventionStrategy `json:"strategy,omitempty"`
	SelectedAudio  *VoiceInterventionAudio   `json:"selected_audio,omitempty"`
	IsHighVolume   bool                      `json:"is_high_volume"`
	VolumePercent  int                       `json:"volume_percent"`
	Reason         string                    `json:"reason,omitempty"`
}

type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON(), nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	*j = JSON(data)
	return nil
}
