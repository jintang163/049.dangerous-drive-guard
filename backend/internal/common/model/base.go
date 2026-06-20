package model

import (
	"time"
)

type BaseModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"index;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleDispatcher UserRole = "dispatcher"
	RoleDriver     UserRole = "driver"
	RoleEscort     UserRole = "escort"
	RoleViewer     UserRole = "viewer"
)

type User struct {
	BaseModel
	Username    string   `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password    string   `gorm:"type:varchar(128);not null" json:"-"`
	RealName    string   `gorm:"type:varchar(64);not null" json:"real_name"`
	Phone       string   `gorm:"type:varchar(20);index" json:"phone"`
	Email       string   `gorm:"type:varchar(128)" json:"email"`
	Role        UserRole `gorm:"type:varchar(20);index;not null" json:"role"`
	OrgID       int64    `gorm:"index" json:"org_id"`
	AvatarURL   string   `gorm:"type:varchar(256)" json:"avatar_url"`
	IDCard      string   `gorm:"type:varchar(32)" json:"id_card"`
	LicenseNo   string   `gorm:"type:varchar(64)" json:"license_no"`
	LicenseType string   `gorm:"type:varchar(20)" json:"license_type"`
	Status      int      `gorm:"default:1;index" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type UserToken struct {
	BaseModel
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(512);index;not null" json:"token"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	DeviceID  string    `gorm:"type:varchar(128)" json:"device_id"`
	IP        string    `gorm:"type:varchar(64)" json:"ip"`
}

type VehicleType string

const (
	VehicleTanker VehicleType = "tanker"
	VehicleVan    VehicleType = "van"
	VehicleFlatbed VehicleType = "flatbed"
	VehicleOther  VehicleType = "other"
)

type VehicleStatus string

const (
	VehicleStatusIdle     VehicleStatus = "idle"
	VehicleStatusRunning  VehicleStatus = "running"
	VehicleStatusLoading  VehicleStatus = "loading"
	VehicleStatusUnloading VehicleStatus = "unloading"
	VehicleStatusRepair   VehicleStatus = "repair"
	VehicleStatusOffline  VehicleStatus = "offline"
)

type Vehicle struct {
	BaseModel
	PlateNumber    string        `gorm:"type:varchar(20);uniqueIndex;not null" json:"plate_number"`
	VehicleType    VehicleType   `gorm:"type:varchar(20);not null" json:"vehicle_type"`
	Brand          string        `gorm:"type:varchar(64)" json:"brand"`
	Model          string        `gorm:"type:varchar(64)" json:"model"`
	Color          string        `gorm:"type:varchar(32)" json:"color"`
	VIN            string        `gorm:"type:varchar(64);uniqueIndex" json:"vin"`
	EngineNo       string        `gorm:"type:varchar(64)" json:"engine_no"`
	LoadWeight     float64       `gorm:"type:decimal(10,2)" json:"load_weight"`
	LoadVolume     float64       `gorm:"type:decimal(10,2)" json:"load_volume"`
	Length         float64       `gorm:"type:decimal(5,2)" json:"length"`
	Width          float64       `gorm:"type:decimal(5,2)" json:"width"`
	Height         float64       `gorm:"type:decimal(5,2)" json:"height"`
	MaxSpeed       int           `json:"max_speed"`
	FuelType       string        `gorm:"type:varchar(20)" json:"fuel_type"`
	Status         VehicleStatus `gorm:"type:varchar(20);index;default:idle" json:"status"`
	OrgID          int64         `gorm:"index" json:"org_id"`
	DriverID       int64         `gorm:"index" json:"driver_id"`
	EscortID       int64         `gorm:"index" json:"escort_id"`
	DeviceID       string        `gorm:"type:varchar(128);index" json:"device_id"`
	InsuranceInfo  string        `gorm:"type:text" json:"insurance_info"`
	AnnualAuditDate *time.Time   `json:"annual_audit_date"`
	Mileage        float64       `gorm:"type:decimal(12,2);default:0" json:"mileage"`
}
