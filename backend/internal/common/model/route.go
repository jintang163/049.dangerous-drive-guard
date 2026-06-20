package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

type RestrictedAreaType string

const (
	AreaTypeSchool           RestrictedAreaType = "school"
	AreaTypeHospital         RestrictedAreaType = "hospital"
	AreaTypeMall             RestrictedAreaType = "mall"
	AreaTypeTunnel           RestrictedAreaType = "tunnel"
	AreaTypeBridge           RestrictedAreaType = "bridge"
	AreaTypeWaterProtection  RestrictedAreaType = "water_protection"
	AreaTypeHeightLimit      RestrictedAreaType = "height_limit"
	AreaTypeWeightLimit      RestrictedAreaType = "weight_limit"
)

type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal JSON value:", value))
	}
	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	Name      string  `json:"name,omitempty"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (p GeoPoint) DistanceTo(other GeoPoint) float64 {
	const R = 6371000.0
	lat1 := p.Lat * math.Pi / 180.0
	lat2 := other.Lat * math.Pi / 180.0
	deltaLat := (other.Lat - p.Lat) * math.Pi / 180.0
	deltaLng := (other.Lng - p.Lng) * math.Pi / 180.0
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

type RestrictedArea struct {
	ID                   int64                `json:"id"`
	Name                 string               `json:"name"`
	AreaType             RestrictedAreaType   `json:"area_type"`
	Level                int                  `json:"level"`
	Province             string               `json:"province"`
	City                 string               `json:"city"`
	District             string               `json:"district"`
	Address              string               `json:"address"`
	BoundaryPolygon      JSON                 `json:"boundary_polygon"`
	CenterLatitude       float64              `json:"center_latitude"`
	CenterLongitude      float64              `json:"center_longitude"`
	Radius               float64              `json:"radius"`
	RestrictHazardClasses string              `json:"restrict_hazard_classes"`
	RestrictVehicleTypes  string              `json:"restrict_vehicle_types"`
	HeightLimit          float64              `json:"height_limit"`
	WeightLimit          float64              `json:"weight_limit"`
	EffectiveFrom        *time.Time           `json:"effective_from"`
	EffectiveTo          *time.Time           `json:"effective_to"`
	Source               string               `json:"source"`
	Status               int                  `json:"status"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type RouteStrategy string

const (
	StrategyShortest RouteStrategy = "shortest"
	StrategySafest   RouteStrategy = "safest"
	StrategyEconomic RouteStrategy = "economic"
	StrategyCustom   RouteStrategy = "custom"
)

type RoutePlanRequest struct {
	Origin          Coordinate        `json:"origin" binding:"required"`
	Destination     Coordinate        `json:"destination" binding:"required"`
	Waypoints       []Coordinate      `json:"waypoints"`
	Strategy        RouteStrategy     `json:"strategy" binding:"required"`
	VehicleID       int64             `json:"vehicle_id"`
	VehicleType     VehicleType       `json:"vehicle_type"`
	VehicleHeight   float64           `json:"vehicle_height"`
	VehicleWeight   float64           `json:"vehicle_weight"`
	VehicleWidth    float64           `json:"vehicle_width"`
	VehicleLength   float64           `json:"vehicle_length"`
	HazardClass     string            `json:"hazard_class"`
	WaybillID       int64             `json:"waybill_id"`
	DriverID        int64             `json:"driver_id"`
	CustomWeights   map[string]float64 `json:"custom_weights"`
}

type RouteSegment struct {
	Index         int         `json:"index"`
	Start         GeoPoint    `json:"start"`
	End           GeoPoint    `json:"end"`
	Distance      float64     `json:"distance"`
	Duration      int         `json:"duration"`
	RoadName      string      `json:"road_name"`
	RoadType      string      `json:"road_type"`
	HasToll       bool        `json:"has_toll"`
	TollFee       float64     `json:"toll_fee"`
	SpeedLimit    int         `json:"speed_limit"`
	Restriction   string      `json:"restriction,omitempty"`
	RestrictedIDs []int64     `json:"restricted_ids,omitempty"`
	Instructions  string      `json:"instructions"`
}

type RestrictedSegmentInfo struct {
	AreaID      int64              `json:"area_id"`
	AreaName    string             `json:"area_name"`
	AreaType    RestrictedAreaType `json:"area_type"`
	Level       int                `json:"level"`
	EntryPoint  GeoPoint           `json:"entry_point"`
	ExitPoint   GeoPoint           `json:"exit_point"`
	Distance    float64            `json:"distance"`
	Reason      string             `json:"reason"`
	Suggestion  string             `json:"suggestion"`
}

type RoutePlan struct {
	ID                    int64                   `json:"id"`
	PlanNo                string                  `json:"plan_no"`
	WaybillID             int64                   `json:"waybill_id"`
	VehicleID             int64                   `json:"vehicle_id"`
	DriverID              int64                   `json:"driver_id"`
	Strategy              RouteStrategy           `json:"strategy"`
	Origin                Coordinate              `json:"origin"`
	Destination           Coordinate              `json:"destination"`
	Waypoints             []Coordinate            `json:"waypoints"`
	RouteGeometry         JSON                    `json:"route_geometry"`
	RoutePath             []GeoPoint              `json:"route_path"`
	Segments              []RouteSegment          `json:"segments"`
	TotalDistance         float64                 `json:"total_distance"`
	EstimatedDuration     int                     `json:"estimated_duration"`
	ExpectedSpeed         float64                 `json:"expected_speed"`
	TollFee               float64                 `json:"toll_fee"`
	FuelCost              float64                 `json:"fuel_cost"`
	AvoidTunnels          int                     `json:"avoid_tunnels"`
	AvoidBridges          int                     `json:"avoid_bridges"`
	AvoidPopulated        int                     `json:"avoid_populated"`
	AvoidWaterProtection  int                     `json:"avoid_water_protection"`
	RestrictedSegments    []RestrictedSegmentInfo `json:"restricted_segments"`
	SafetyScore           float64                 `json:"safety_score"`
	AlternativeRoutes     []*RoutePlan            `json:"alternative_routes,omitempty"`
	Status                string                  `json:"status"`
	CreatedAt             time.Time               `json:"created_at"`
}

type MultiStrategyPlan struct {
	Shortest *RoutePlan `json:"shortest"`
	Safest   *RoutePlan `json:"safest"`
	Economic *RoutePlan `json:"economic"`
}

type ServiceArea struct {
	ID                  int64     `json:"id"`
	Name                string    `json:"name"`
	HighwayName         string    `json:"highway_name"`
	Direction           string    `json:"direction"`
	Province            string    `json:"province"`
	City                string    `json:"city"`
	Latitude            float64   `json:"latitude"`
	Longitude           float64   `json:"longitude"`
	DistanceFromStart   float64   `json:"distance_from_start"`
	DistanceFromCurrent float64   `json:"distance_from_current,omitempty"`
	HasRestaurant       bool      `json:"has_restaurant"`
	HasHotel            bool      `json:"has_hotel"`
	HasFuelStation      bool      `json:"has_fuel_station"`
	HasCharging         bool      `json:"has_charging"`
	HasMaintenance      bool      `json:"has_maintenance"`
	HasDangerParking    bool      `json:"has_danger_goods_parking"`
	ParkingSpaces       int       `json:"parking_spaces"`
	DangerParkingSpaces int       `json:"danger_parking_spaces"`
	Phone               string    `json:"phone"`
	Rating              float64   `json:"rating"`
	EstimatedArrivalTime int      `json:"estimated_arrival_time,omitempty"`
	RestDurationRecommend int     `json:"rest_duration_recommend,omitempty"`
	Status              int       `json:"status"`
}

type WaybillStatus string

const (
	WaybillCreated    WaybillStatus = "created"
	WaybillAssigned   WaybillStatus = "assigned"
	WaybillLoading    WaybillStatus = "loading"
	WaybillInTransit  WaybillStatus = "in_transit"
	WaybillUnloading  WaybillStatus = "unloading"
	WaybillCompleted  WaybillStatus = "completed"
	WaybillCancelled  WaybillStatus = "cancelled"
	WaybillException  WaybillStatus = "exception"
)

type Waybill struct {
	BaseModel
	WaybillNo            string        `json:"waybill_no"`
	OrderNo              string        `json:"order_no"`
	ShipperOrgID         int64         `json:"shipper_org_id"`
	CarrierOrgID         int64         `json:"carrier_org_id"`
	ReceiverOrgID        int64         `json:"receiver_org_id"`
	VehicleID            int64         `json:"vehicle_id"`
	DriverID             int64         `json:"driver_id"`
	EscortID             int64         `json:"escort_id"`
	RoutePlanID          int64         `json:"route_plan_id"`
	GoodsID              int64         `json:"goods_id"`
	GoodsName            string        `json:"goods_name"`
	GoodsUNCode          string        `json:"goods_un_code"`
	GoodsHazardClass     string        `json:"goods_hazard_class"`
	GoodsWeight          float64       `json:"goods_weight"`
	GoodsVolume          float64       `json:"goods_volume"`
	PackageType          string        `json:"package_type"`
	PackageCount         int           `json:"package_count"`
	OriginAddress        string        `json:"origin_address"`
	OriginLatitude       float64       `json:"origin_latitude"`
	OriginLongitude      float64       `json:"origin_longitude"`
	DestAddress          string        `json:"dest_address"`
	DestLatitude         float64       `json:"dest_latitude"`
	DestLongitude        float64       `json:"dest_longitude"`
	PlannedDepartureTime *time.Time    `json:"planned_departure_time"`
	ActualDepartureTime  *time.Time    `json:"actual_departure_time"`
	PlannedArrivalTime   *time.Time    `json:"planned_arrival_time"`
	ActualArrivalTime    *time.Time    `json:"actual_arrival_time"`
	Status               WaybillStatus `json:"status"`
	TotalDistance        float64       `json:"total_distance"`
	TransportCost        float64       `json:"transport_cost"`
	RiskLevel            int           `json:"risk_level"`
	ApprovalStatus       int           `json:"approval_status"`
	ApprovedBy           int64         `json:"approved_by"`
	ApprovedAt           *time.Time    `json:"approved_at"`
	EmergencyContact     string        `json:"emergency_contact"`
	EmergencyPhone       string        `json:"emergency_phone"`
	Remark               string        `json:"remark"`
	BlockchainTxHash     string        `json:"blockchain_tx_hash"`
	BlockchainBlockNo    int64         `json:"blockchain_block_no"`
	Driver               *User         `json:"driver,omitempty" gorm:"-"`
	Vehicle              *Vehicle      `json:"vehicle,omitempty" gorm:"-"`
	RoutePlan            *RoutePlan    `json:"route_plan,omitempty" gorm:"-"`
}

type WaybillCreateRequest struct {
	OrderNo              string     `json:"order_no"`
	ShipperOrgID         int64      `json:"shipper_org_id" binding:"required"`
	CarrierOrgID         int64      `json:"carrier_org_id" binding:"required"`
	ReceiverOrgID        int64      `json:"receiver_org_id"`
	VehicleID            int64      `json:"vehicle_id" binding:"required"`
	DriverID             int64      `json:"driver_id" binding:"required"`
	EscortID             int64      `json:"escort_id"`
	GoodsID              int64      `json:"goods_id" binding:"required"`
	GoodsWeight          float64    `json:"goods_weight" binding:"required,gt=0"`
	GoodsVolume          float64    `json:"goods_volume"`
	PackageType          string     `json:"package_type"`
	PackageCount         int        `json:"package_count"`
	OriginAddress        string     `json:"origin_address" binding:"required"`
	OriginLatitude       float64    `json:"origin_latitude" binding:"required"`
	OriginLongitude      float64    `json:"origin_longitude" binding:"required"`
	DestAddress          string     `json:"dest_address" binding:"required"`
	DestLatitude         float64    `json:"dest_latitude" binding:"required"`
	DestLongitude        float64    `json:"dest_longitude" binding:"required"`
	PlannedDepartureTime *time.Time `json:"planned_departure_time"`
	TransportCost        float64    `json:"transport_cost"`
	RiskLevel            int        `json:"risk_level"`
	EmergencyContact     string     `json:"emergency_contact"`
	EmergencyPhone       string     `json:"emergency_phone"`
	Remark               string     `json:"remark"`
}
