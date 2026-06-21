package model

import (
	"time"
)

type ShapeType string

const (
	ShapePolygon ShapeType = "polygon"
	ShapeCircle  ShapeType = "circle"
)

type TemplateCategory string

const (
	TemplateHospital   TemplateCategory = "hospital"
	TemplateSchool     TemplateCategory = "school"
	TemplateMall       TemplateCategory = "mall"
	TemplateWaterSource TemplateCategory = "water_source"
	TemplateCustom     TemplateCategory = "custom"
)

type ApprovalStatus int8

const (
	ApprovalPending      ApprovalStatus = 0
	ApprovalFirstLevel   ApprovalStatus = 1
	ApprovalSecondLevel  ApprovalStatus = 2
	ApprovalApproved     ApprovalStatus = 3
	ApprovalRejected     ApprovalStatus = 4
	ApprovalRevoked      ApprovalStatus = 5
)

type TempReason string

const (
	TempReasonAccident    TempReason = "accident"
	TempReasonConstruction TempReason = "construction"
	TempReasonEmergency   TempReason = "emergency"
	TempReasonOther       TempReason = "other"
)

type ApprovalAction string

const (
	ApprovalActionSubmit  ApprovalAction = "submit"
	ApprovalActionApprove ApprovalAction = "approve"
	ApprovalActionReject  ApprovalAction = "reject"
	ApprovalActionRevoke  ApprovalAction = "revoke"
)

type TimeScheduleRule struct {
	Weekdays    []int  `json:"weekdays"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Description string `json:"description,omitempty"`
}

type RestrictedAreaTemplate struct {
	BaseModel
	TemplateName         string           `gorm:"type:varchar(128);not null" json:"template_name"`
	TemplateCategory     TemplateCategory `gorm:"type:varchar(32);not null;index" json:"template_category"`
	AreaType             RestrictedAreaType `gorm:"type:varchar(32);not null" json:"area_type"`
	Level                int              `gorm:"default:2" json:"level"`
	DefaultRadius        float64          `gorm:"type:decimal(10,2);default:500" json:"default_radius"`
	RestrictHazardClasses string          `gorm:"type:varchar(128)" json:"restrict_hazard_classes"`
	RestrictVehicleTypes  string          `gorm:"type:varchar(128)" json:"restrict_vehicle_types"`
	HeightLimit          float64          `gorm:"type:decimal(5,2)" json:"height_limit"`
	WeightLimit          float64          `gorm:"type:decimal(8,2)" json:"weight_limit"`
	TimeRules            JSON             `gorm:"type:json" json:"time_rules"`
	Description          string           `gorm:"type:varchar(512)" json:"description"`
	IsBuiltin            int              `gorm:"default:0" json:"is_builtin"`
	IsEnabled            int              `gorm:"default:1;index" json:"is_enabled"`
	CreatedBy            int64            `json:"created_by"`
}

func (RestrictedAreaTemplate) TableName() string {
	return "restricted_area_templates"
}

type RestrictedAreaExt struct {
	RestrictedArea
	ShapeType           ShapeType       `gorm:"type:varchar(16);default:polygon" json:"shape_type"`
	IsTemporary         int             `gorm:"default:0;index" json:"is_temporary"`
	TempReason          string          `gorm:"type:varchar(512)" json:"temp_reason"`
	TimeSchedule        JSON            `gorm:"type:json" json:"time_schedule"`
	ApprovalStatus      ApprovalStatus  `gorm:"default:1;index" json:"approval_status"`
	FirstApproverID     int64           `json:"first_approver_id"`
	FirstApprovalAt     *time.Time      `json:"first_approval_at"`
	FirstApprovalNote   string          `gorm:"type:varchar(512)" json:"first_approval_note"`
	SecondApproverID    int64           `json:"second_approver_id"`
	SecondApprovalAt    *time.Time      `json:"second_approval_at"`
	SecondApprovalNote  string          `gorm:"type:varchar(512)" json:"second_approval_note"`
	TemplateID          int64           `gorm:"index" json:"template_id"`
	CreatedBy           int64           `json:"created_by"`
	GisImportID         string          `gorm:"type:varchar(64)" json:"gis_import_id"`
}

func (RestrictedAreaExt) TableName() string {
	return "restricted_areas"
}

type RestrictedAreaApproval struct {
	BaseModel
	AreaID         int64          `gorm:"not null;index" json:"area_id"`
	ApprovalLevel  int            `gorm:"not null" json:"approval_level"`
	ApproverID     int64          `gorm:"not null;index" json:"approver_id"`
	ApproverName   string         `gorm:"type:varchar(64)" json:"approver_name"`
	ApprovalAction ApprovalAction `gorm:"type:varchar(16);not null" json:"approval_action"`
	ApprovalNote   string         `gorm:"type:varchar(512)" json:"approval_note"`
	OldStatus      ApprovalStatus `json:"old_status"`
	NewStatus      ApprovalStatus `json:"new_status"`
}

func (RestrictedAreaApproval) TableName() string {
	return "restricted_area_approvals"
}

type GisImportRecord struct {
	BaseModel
	ImportBatchNo string  `gorm:"type:varchar(64);uniqueIndex;not null" json:"import_batch_no"`
	FileName      string  `gorm:"type:varchar(256)" json:"file_name"`
	SourceType    string  `gorm:"type:varchar(32);not null" json:"source_type"`
	TotalCount    int     `gorm:"default:0" json:"total_count"`
	SuccessCount  int     `gorm:"default:0" json:"success_count"`
	FailedCount   int     `gorm:"default:0" json:"failed_count"`
	FailedDetails JSON    `gorm:"type:json" json:"failed_details"`
	ImportStatus  string  `gorm:"type:varchar(16);default:processing;index" json:"import_status"`
	ImportedBy    int64   `json:"imported_by"`
}

func (GisImportRecord) TableName() string {
	return "restricted_area_gis_imports"
}

type RestrictedAreaCreateRequest struct {
	Name                 string              `json:"name" binding:"required"`
	AreaType             RestrictedAreaType  `json:"area_type" binding:"required"`
	ShapeType            ShapeType           `json:"shape_type" binding:"required"`
	Level                int                 `json:"level"`
	Province             string              `json:"province"`
	City                 string              `json:"city"`
	District             string              `json:"district"`
	Address              string              `json:"address"`
	BoundaryPolygon      JSON                `json:"boundary_polygon"`
	CenterLatitude       float64             `json:"center_latitude"`
	CenterLongitude      float64             `json:"center_longitude"`
	Radius               float64             `json:"radius"`
	RestrictHazardClasses string             `json:"restrict_hazard_classes"`
	RestrictVehicleTypes  string             `json:"restrict_vehicle_types"`
	HeightLimit          float64             `json:"height_limit"`
	WeightLimit          float64             `json:"weight_limit"`
	TimeSchedule         []TimeScheduleRule  `json:"time_schedule"`
	EffectiveFrom        *time.Time          `json:"effective_from"`
	EffectiveTo          *time.Time          `json:"effective_to"`
	Source               string              `json:"source"`
	IsTemporary          int                 `json:"is_temporary"`
	TempReason           string              `json:"temp_reason"`
	TemplateID           int64               `json:"template_id"`
	SubmitApproval       bool                `json:"submit_approval"`
}

type RestrictedAreaUpdateRequest struct {
	Name                 string              `json:"name"`
	AreaType             RestrictedAreaType  `json:"area_type"`
	ShapeType            ShapeType           `json:"shape_type"`
	Level                int                 `json:"level"`
	Province             string              `json:"province"`
	City                 string              `json:"city"`
	District             string              `json:"district"`
	Address              string              `json:"address"`
	BoundaryPolygon      JSON                `json:"boundary_polygon"`
	CenterLatitude       float64             `json:"center_latitude"`
	CenterLongitude      float64             `json:"center_longitude"`
	Radius               float64             `json:"radius"`
	RestrictHazardClasses string             `json:"restrict_hazard_classes"`
	RestrictVehicleTypes  string             `json:"restrict_vehicle_types"`
	HeightLimit          float64             `json:"height_limit"`
	WeightLimit          float64             `json:"weight_limit"`
	TimeSchedule         []TimeScheduleRule  `json:"time_schedule"`
	EffectiveFrom        *time.Time          `json:"effective_from"`
	EffectiveTo          *time.Time          `json:"effective_to"`
	IsTemporary          int                 `json:"is_temporary"`
	TempReason           string              `json:"temp_reason"`
	Status               int                 `json:"status"`
}

type ApprovalRequest struct {
	ApprovalNote string `json:"approval_note"`
}

type TemplateCreateRequest struct {
	TemplateName         string           `json:"template_name" binding:"required"`
	TemplateCategory     TemplateCategory `json:"template_category" binding:"required"`
	AreaType             RestrictedAreaType `json:"area_type" binding:"required"`
	Level                int              `json:"level"`
	DefaultRadius        float64          `json:"default_radius"`
	RestrictHazardClasses string          `json:"restrict_hazard_classes"`
	RestrictVehicleTypes  string          `json:"restrict_vehicle_types"`
	HeightLimit          float64          `json:"height_limit"`
	WeightLimit          float64          `json:"weight_limit"`
	TimeRules            []TimeScheduleRule `json:"time_rules"`
	Description          string           `json:"description"`
	IsEnabled            int              `json:"is_enabled"`
}

type GisImportRequest struct {
	SourceType string          `json:"source_type" binding:"required"`
	FileName   string          `json:"file_name"`
	Features   JSON            `json:"features" binding:"required"`
}
