package model

import (
	"time"
)

type ServiceArea struct {
	BaseModel
	Name              string  `gorm:"type:varchar(128);not null" json:"name"`
	HighwayName       string  `gorm:"type:varchar(64)" json:"highway_name"`
	Direction         string  `gorm:"type:varchar(16)" json:"direction"`
	Province          string  `gorm:"type:varchar(64)" json:"province"`
	City              string  `gorm:"type:varchar(64)" json:"city"`
	Latitude          float64 `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude         float64 `gorm:"type:decimal(10,7);not null" json:"longitude"`
	DistanceFromStart float64 `gorm:"type:decimal(10,2)" json:"distance_from_start_km"`
	HasRestaurant     bool    `gorm:"default:true" json:"has_restaurant"`
	HasHotel          bool    `gorm:"default:false" json:"has_hotel"`
	HasFuelStation    bool    `gorm:"default:true" json:"has_fuel_station"`
	HasCharging       bool    `gorm:"default:false" json:"has_charging"`
	HasRestRoom       bool    `gorm:"default:true" json:"has_rest_room"`
	HasMaintenance    bool    `gorm:"default:false" json:"has_maintenance"`
	HasDangerGoodsParking bool `gorm:"default:false" json:"has_danger_goods_parking"`
	ParkingSpaces     int     `json:"parking_spaces"`
	DangerParkingSpaces int   `json:"danger_parking_spaces"`
	Phone             string  `gorm:"type:varchar(20)" json:"phone"`
	Rating            float64 `gorm:"type:decimal(3,2)" json:"rating"`
	Status            int     `gorm:"default:1;index" json:"status"`
}

func (ServiceArea) TableName() string {
	return "service_areas"
}

type DrivingRestRecordStatus string

const (
	DrivingRestStatusDriving   DrivingRestRecordStatus = "driving"
	DrivingRestStatusResting   DrivingRestRecordStatus = "resting"
	DrivingRestStatusCompleted DrivingRestRecordStatus = "completed"
)

type DrivingRestRecord struct {
	BaseModel
	DriverID             int64                    `gorm:"index;not null" json:"driver_id"`
	VehicleID            int64                    `gorm:"index;not null" json:"vehicle_id"`
	WaybillID            int64                    `gorm:"index" json:"waybill_id"`
	RecordDate           string                   `gorm:"type:date;index" json:"record_date"`
	DriveStartTime       time.Time                `gorm:"not null" json:"drive_start_time"`
	DriveEndTime         *time.Time               `json:"drive_end_time"`
	ContinuousDriveMinutes int                   `gorm:"default:0" json:"continuous_drive_minutes"`
	RestStartTime        *time.Time               `json:"rest_start_time"`
	RestEndTime          *time.Time               `json:"rest_end_time"`
	RestDurationMinutes  int                      `gorm:"default:0" json:"rest_duration_minutes"`
	RestServiceAreaID    int64                    `json:"rest_service_area_id"`
	RestServiceAreaName  string                   `gorm:"type:varchar(128)" json:"rest_service_area_name"`
	Status               DrivingRestRecordStatus  `gorm:"type:varchar(20);index;default:driving" json:"status"`
	IsOvertime           bool                     `gorm:"default:false;index" json:"is_overtime"`
	OvertimeMinutes      int                      `gorm:"default:0" json:"overtime_minutes"`
	CheckInTime          *time.Time               `json:"check_in_time"`
	CheckInLatitude      float64                  `gorm:"type:decimal(10,7)" json:"check_in_latitude"`
	CheckInLongitude     float64                  `gorm:"type:decimal(10,7)" json:"check_in_longitude"`
	CheckOutTime         *time.Time               `json:"check_out_time"`
	CheckOutLatitude     float64                  `gorm:"type:decimal(10,7)" json:"check_out_latitude"`
	CheckOutLongitude    float64                  `gorm:"type:decimal(10,7)" json:"check_out_longitude"`
	MinRestRequired      int                      `gorm:"default:20" json:"min_rest_required"`
	MaxContinuousDrive   int                      `gorm:"default:240" json:"max_continuous_drive"`

	DriverName           string  `gorm:"-" json:"driver_name,omitempty"`
	VehiclePlate         string  `gorm:"-" json:"vehicle_plate,omitempty"`
	RemainingDriveMinutes int    `gorm:"-" json:"remaining_drive_minutes,omitempty"`
	RestProgressPercent  float64 `gorm:"-" json:"rest_progress_percent,omitempty"`
}

func (DrivingRestRecord) TableName() string {
	return "driving_rest_records"
}

type ServiceAreaRealtimeStatus struct {
	BaseModel
	ServiceAreaID          int64     `gorm:"uniqueIndex;not null" json:"service_area_id"`
	TotalParkingSpaces     int       `gorm:"default:0" json:"total_parking_spaces"`
	AvailableParkingSpaces int       `gorm:"default:0" json:"available_parking_spaces"`
	TotalDangerSpaces      int       `gorm:"default:0" json:"total_danger_spaces"`
	AvailableDangerSpaces  int       `gorm:"default:0" json:"available_danger_spaces"`
	HasFuel                bool      `gorm:"default:true" json:"has_fuel"`
	FuelPrice92            float64   `gorm:"type:decimal(5,2)" json:"fuel_price_92"`
	FuelPrice95            float64   `gorm:"type:decimal(5,2)" json:"fuel_price_95"`
	FuelPriceDiesel        float64   `gorm:"type:decimal(5,2)" json:"fuel_price_diesel"`
	HasCharging            bool      `gorm:"default:false" json:"has_charging"`
	ChargingPilesTotal     int       `gorm:"default:0" json:"charging_piles_total"`
	ChargingPilesAvailable int       `gorm:"default:0" json:"charging_piles_available"`
	HasRestaurant          bool      `gorm:"default:true" json:"has_restaurant"`
	RestaurantRating       float64   `gorm:"type:decimal(3,2)" json:"restaurant_rating"`
	RestaurantWaitMinutes  int       `gorm:"default:0" json:"restaurant_wait_minutes"`
	HasHotel               bool      `gorm:"default:false" json:"has_hotel"`
	HotelRating            float64   `gorm:"type:decimal(3,2)" json:"hotel_rating"`
	HasMaintenance         bool      `gorm:"default:false" json:"has_maintenance"`
	SecurityLevel          int       `gorm:"default:3;index" json:"security_level"`
	SecurityPatrolInterval int       `gorm:"default:30" json:"security_patrol_interval"`
	CrowdLevel             int       `gorm:"default:2" json:"crowd_level"`
	WeatherCondition       string    `gorm:"type:varchar(32)" json:"weather_condition"`
	UpdateTime             time.Time `gorm:"not null;index" json:"update_time"`
	DataSource             string    `gorm:"type:varchar(32);default:manual" json:"data_source"`
}

func (ServiceAreaRealtimeStatus) TableName() string {
	return "service_area_realtime_status"
}

type ServiceAreaReview struct {
	BaseModel
	ServiceAreaID  int64    `gorm:"index;not null" json:"service_area_id"`
	DriverID       int64    `gorm:"index;not null" json:"driver_id"`
	DriverName     string   `gorm:"type:varchar(64)" json:"driver_name"`
	WaybillID      int64    `json:"waybill_id"`
	VehicleID      int64    `json:"vehicle_id"`
	SecurityScore  int      `gorm:"not null;index" json:"security_score"`
	EnvironmentScore int    `json:"environment_score"`
	FoodScore      int      `json:"food_score"`
	ServiceScore   int      `json:"service_score"`
	OverallScore   float64  `gorm:"type:decimal(3,2);not null" json:"overall_score"`
	CommentText    string   `gorm:"type:text" json:"comment_text"`
	Tags           string   `gorm:"type:json" json:"tags"`
	Images         string   `gorm:"type:json" json:"images"`
	IsAnonymous    bool     `gorm:"default:false" json:"is_anonymous"`
	Status         int      `gorm:"default:1;index" json:"status"`
	CheckInRecordID int64   `json:"check_in_record_id"`

	ServiceAreaName string `gorm:"-" json:"service_area_name,omitempty"`
	TagsArray       []string `gorm:"-" json:"tags_array,omitempty"`
	ImagesArray     []string `gorm:"-" json:"images_array,omitempty"`
}

func (ServiceAreaReview) TableName() string {
	return "service_area_reviews"
}

type ServiceAreaRecommendation struct {
	BaseModel
	RecommendNo            string      `gorm:"type:varchar(32);uniqueIndex;not null" json:"recommend_no"`
	DriverID               int64       `gorm:"index;not null" json:"driver_id"`
	VehicleID              int64       `gorm:"index;not null" json:"vehicle_id"`
	WaybillID              int64       `gorm:"index" json:"waybill_id"`
	CurrentLatitude        float64     `gorm:"type:decimal(10,7);not null" json:"current_latitude"`
	CurrentLongitude       float64     `gorm:"type:decimal(10,7);not null" json:"current_longitude"`
	CurrentAddress         string      `gorm:"type:varchar(512)" json:"current_address"`
	ContinuousDriveMinutes int         `json:"continuous_drive_minutes"`
	RemainingDriveMinutes  int         `json:"remaining_drive_minutes"`
	FatigueScore           float64     `gorm:"type:decimal(5,2)" json:"fatigue_score"`
	HazardClass            string      `gorm:"type:varchar(8)" json:"hazard_class"`
	RecommendReason        string      `gorm:"type:varchar(512)" json:"recommend_reason"`
	RecommendedServiceAreaID int64    `gorm:"index" json:"recommended_service_area_id"`
	RecommendedServiceAreaName string  `gorm:"type:varchar(128)" json:"recommended_service_area_name"`
	DistanceKm             float64     `gorm:"type:decimal(8,2)" json:"distance_km"`
	EstimatedArrivalMinutes int        `json:"estimated_arrival_minutes"`
	Alternatives           string      `gorm:"type:json" json:"alternatives"`
	Status                 string      `gorm:"type:varchar(20);index;default:pending" json:"status"`
	AcceptedAt             *time.Time  `json:"accepted_at"`
	ArrivedAt              *time.Time  `json:"arrived_at"`
	DispatchSource         string      `gorm:"type:varchar(32);default:system" json:"dispatch_source"`
	DispatcherID           int64       `json:"dispatcher_id"`

	AlternativesArray []RecommendedServiceArea `gorm:"-" json:"alternatives_array,omitempty"`
}

type RecommendedServiceArea struct {
	ServiceAreaID   int64   `json:"service_area_id"`
	ServiceAreaName string  `json:"service_area_name"`
	DistanceKm      float64 `json:"distance_km"`
	EstimatedArrivalMinutes int `json:"estimated_arrival_minutes"`
	AvailableDangerSpaces int `json:"available_danger_spaces"`
	SecurityLevel   int     `json:"security_level"`
	RestaurantRating float64 `json:"restaurant_rating"`
	HasFuel         bool    `json:"has_fuel"`
	HasCharging     bool    `json:"has_charging"`
	RecommendReason string  `json:"recommend_reason"`
	MatchScore      float64 `json:"match_score"`
}

func (ServiceAreaRecommendation) TableName() string {
	return "service_area_recommendations"
}

type CheckInRequest struct {
	DriverID      int64   `json:"driver_id" binding:"required"`
	VehicleID     int64   `json:"vehicle_id" binding:"required"`
	ServiceAreaID int64   `json:"service_area_id" binding:"required"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	WaybillID     int64   `json:"waybill_id"`
}

type CheckOutRequest struct {
	DriverID  int64   `json:"driver_id" binding:"required"`
	VehicleID int64   `json:"vehicle_id" binding:"required"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type SubmitReviewRequest struct {
	ServiceAreaID  int64    `json:"service_area_id" binding:"required"`
	DriverID       int64    `json:"driver_id" binding:"required"`
	SecurityScore  int      `json:"security_score" binding:"required,min=1,max=5"`
	EnvironmentScore int    `json:"environment_score"`
	FoodScore      int      `json:"food_score"`
	ServiceScore   int      `json:"service_score"`
	CommentText    string   `json:"comment_text"`
	Tags           []string `json:"tags"`
	Images         []string `json:"images"`
	IsAnonymous    bool     `json:"is_anonymous"`
	WaybillID      int64    `json:"waybill_id"`
	VehicleID      int64    `json:"vehicle_id"`
}

type RecommendServiceAreaRequest struct {
	DriverID    int64   `json:"driver_id" binding:"required"`
	VehicleID   int64   `json:"vehicle_id" binding:"required"`
	WaybillID   int64   `json:"waybill_id"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	HazardClass string  `json:"hazard_class"`
	FatigueScore float64 `json:"fatigue_score"`
	RadiusKm    float64 `json:"radius_km"`
}

type RestCountdownResponse struct {
	DriverID             int64  `json:"driver_id"`
	VehicleID            int64  `json:"vehicle_id"`
	WaybillID            int64  `json:"waybill_id"`
	Status               string `json:"status"`
	ContinuousDriveMinutes int   `json:"continuous_drive_minutes"`
	RemainingDriveMinutes int   `json:"remaining_drive_minutes"`
	MaxContinuousDrive   int    `json:"max_continuous_drive"`
	IsOvertime           bool   `json:"is_overtime"`
	OvertimeMinutes      int    `json:"overtime_minutes"`
	MinRestRequired      int    `json:"min_rest_required"`
	CurrentRestMinutes   int    `json:"current_rest_minutes"`
	RestProgressPercent  float64 `json:"rest_progress_percent"`
	CanContinueDriving   bool   `json:"can_continue_driving"`
	CurrentServiceAreaID int64  `json:"current_service_area_id,omitempty"`
	CurrentServiceAreaName string `json:"current_service_area_name,omitempty"`
	NextRecommendationID int64  `json:"next_recommendation_id,omitempty"`
}
