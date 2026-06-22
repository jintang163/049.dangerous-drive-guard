package model

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type DrivingScore struct {
	BaseModel
	DriverID               int64          `gorm:"not null;index:idx_driver_date" json:"driver_id"`
	WaybillID              sql.NullInt64  `gorm:"index:idx_driver_date" json:"waybill_id"`
	VehicleID              sql.NullInt64  `json:"vehicle_id"`
	TripDate               time.Time      `gorm:"type:date;not null;index:idx_driver_date" json:"trip_date"`
	TotalScore             float64        `gorm:"type:decimal(6,2);not null;default:100" json:"total_score"`
	ScoreLevel             string         `gorm:"type:varchar(16);default:excellent" json:"score_level"`
	FatigueScore           sql.NullFloat64 `gorm:"type:decimal(5,2)" json:"fatigue_score"`
	FatigueDeduction       float64        `gorm:"type:decimal(5,2);default:0" json:"fatigue_deduction"`
	OverspeedScore         sql.NullFloat64 `gorm:"type:decimal(5,2)" json:"overspeed_score"`
	OverspeedCount         int            `gorm:"default:0" json:"overspeed_count"`
	OverspeedDeduction     float64        `gorm:"type:decimal(5,2);default:0" json:"overspeed_deduction"`
	SuddenBrakeCount       int            `gorm:"default:0" json:"sudden_brake_count"`
	SuddenBrakeDeduction   float64        `gorm:"type:decimal(5,2);default:0" json:"sudden_brake_deduction"`
	SuddenAccelCount       int            `gorm:"default:0" json:"sudden_accel_count"`
	SuddenAccelDeduction   float64        `gorm:"type:decimal(5,2);default:0" json:"sudden_accel_deduction"`
	SharpTurnCount         int            `gorm:"default:0" json:"sharp_turn_count"`
	SharpTurnDeduction     float64        `gorm:"type:decimal(5,2);default:0" json:"sharp_turn_deduction"`
	LaneDeviationCount     int            `gorm:"default:0" json:"lane_deviation_count"`
	LaneDeviationDeduction float64        `gorm:"type:decimal(5,2);default:0" json:"lane_deviation_deduction"`
	PhoneUsageCount        int            `gorm:"default:0" json:"phone_usage_count"`
	PhoneUsageDeduction    float64        `gorm:"type:decimal(5,2);default:0" json:"phone_usage_deduction"`
	SmokingCount           int            `gorm:"default:0" json:"smoking_count"`
	SmokingDeduction       float64        `gorm:"type:decimal(5,2);default:0" json:"smoking_deduction"`
	SeatbeltViolationCount int            `gorm:"default:0" json:"seatbelt_violation_count"`
	SeatbeltDeduction      float64        `gorm:"type:decimal(5,2);default:0" json:"seatbelt_violation_deduction"`
	RouteDeviationCount    int            `gorm:"default:0" json:"route_deviation_count"`
	RouteDeviationDeduction float64       `gorm:"type:decimal(5,2);default:0" json:"route_deviation_deduction"`
	FatigueAlarmCount      int            `gorm:"default:0" json:"fatigue_alarm_count"`
	TotalDistance           sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"total_distance"`
	DrivingDuration        sql.NullInt64  `json:"driving_duration"`
	NightDrivingDuration   sql.NullInt64  `json:"night_driving_duration"`
	OverspeedDuration      sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"overspeed_duration"`
}

func (DrivingScore) TableName() string {
	return "driving_scores"
}

type DrivingScoreBonus struct {
	BaseModel
	DriverID   int64   `gorm:"not null;index" json:"driver_id"`
	BonusType  string  `gorm:"type:varchar(32);not null" json:"bonus_type"`
	BonusPoints float64 `gorm:"type:decimal(5,2);not null;default:0" json:"bonus_points"`
	Reason     string  `gorm:"type:varchar(256)" json:"reason"`
	StreakDays int     `gorm:"default:0" json:"streak_days"`
	StartDate  sql.NullTime `json:"start_date"`
	EndDate    sql.NullTime `json:"end_date"`
	AwardedBy  sql.NullInt64 `json:"awarded_by"`
	Status     int     `gorm:"default:1" json:"status"`
}

func (DrivingScoreBonus) TableName() string {
	return "driving_score_bonus"
}

type DrivingScoreMonthlyReport struct {
	BaseModel
	DriverID              int64          `gorm:"not null;uniqueIndex:uk_driver_month" json:"driver_id"`
	ReportMonth           string         `gorm:"type:varchar(7);not null;uniqueIndex:uk_driver_month" json:"report_month"`
	AvgScore              float64        `gorm:"type:decimal(6,2);not null;default:0" json:"avg_score"`
	MinScore              sql.NullFloat64 `gorm:"type:decimal(6,2)" json:"min_score"`
	MaxScore              sql.NullFloat64 `gorm:"type:decimal(6,2)" json:"max_score"`
	TotalFatigueAlarms    int            `gorm:"default:0" json:"total_fatigue_alarms"`
	TotalSuddenEvents     int            `gorm:"default:0" json:"total_sudden_events"`
	TotalOverspeedDuration sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"total_overspeed_duration"`
	TotalDistance         sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"total_distance"`
	TotalDrivingDuration  sql.NullInt64  `json:"total_driving_duration"`
	TotalBonusPoints      sql.NullFloat64 `gorm:"type:decimal(5,2)" json:"total_bonus_points"`
	ViolationDays         int            `gorm:"default:0" json:"violation_days"`
	CleanDays             int            `gorm:"default:0" json:"clean_days"`
	ScoreTrend            JSON           `gorm:"type:json" json:"score_trend"`
	NeedRetraining        int            `gorm:"default:0" json:"need_retraining"`
	RetrainingTriggeredAt sql.NullTime   `json:"retraining_triggered_at"`
	ReportSent            int            `gorm:"default:0" json:"report_sent"`
	ReportSentAt          sql.NullTime   `json:"report_sent_at"`
}

func (DrivingScoreMonthlyReport) TableName() string {
	return "driving_score_monthly_report"
}

type DriverRetrainingTask struct {
	BaseModel
	DriverID      int64          `gorm:"not null;index" json:"driver_id"`
	TriggerScore  float64        `gorm:"type:decimal(6,2);not null" json:"trigger_score"`
	TriggerType   string         `gorm:"type:varchar(32);not null;default:low_score" json:"trigger_type"`
	TriggerMonth  sql.NullString `gorm:"type:varchar(7)" json:"trigger_month"`
	TaskType      string         `gorm:"type:varchar(32);not null;default:safety_training" json:"task_type"`
	Status        string         `gorm:"type:varchar(20);default:pending" json:"status"`
	AssignedAt    sql.NullTime   `json:"assigned_at"`
	StartedAt     sql.NullTime   `json:"started_at"`
	CompletedAt   sql.NullTime   `json:"completed_at"`
	ResultScore   sql.NullFloat64 `gorm:"type:decimal(5,2)" json:"result_score"`
	ResultNote    string         `gorm:"type:varchar(512)" json:"result_note"`
	CreatedBy     sql.NullInt64  `json:"created_by"`
}

func (DriverRetrainingTask) TableName() string {
	return "driver_retraining_tasks"
}

type ScoreLeaderboardItem struct {
	DriverID    int64   `json:"driver_id"`
	DriverName  string  `json:"driver_name"`
	PlateNumber string  `json:"plate_number"`
	TotalScore  float64 `json:"total_score"`
	ScoreLevel  string  `json:"score_level"`
	Rank        int     `json:"rank"`
	AvgScore30d float64 `json:"avg_score_30d"`
	BonusPoints float64 `json:"bonus_points"`
	OrgName     string  `json:"org_name"`
}

type ScoreOverview struct {
	TotalDrivers       int     `json:"total_drivers"`
	ExcellentCount     int     `json:"excellent_count"`
	GoodCount          int     `json:"good_count"`
	NormalCount        int     `json:"normal_count"`
	PoorCount          int     `json:"poor_count"`
	DangerCount        int     `json:"danger_count"`
	AvgScore           float64 `json:"avg_score"`
	RetrainingCount    int     `json:"retraining_count"`
	TodayCalculated    int     `json:"today_calculated"`
	NeedRetrainingCount int    `json:"need_retraining_count"`
}

type JSON []byte

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}
