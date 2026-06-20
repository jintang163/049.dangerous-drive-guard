package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"go.uber.org/zap"
)

type AMapTruckRouteClient struct {
	key        string
	httpClient *http.Client
}

type AMapTruckRouteRequest struct {
	Origin          string
	Destination     string
	Waypoints       string
	Size            int
	Height          float64
	Width           float64
	Length          float64
	Weight          float64
	AxleWeight      float64
	AxleCount       int
	TrailerType     int
	GoodsType       int
	IsHazardous     int
	HazardousType   int
	Strategy        int
}

type AMapRouteResponse struct {
	Status   int               `json:"status"`
	Info     string            `json:"info"`
	Infocode string            `json:"infocode"`
	Route    *AMapTruckRoute   `json:"route"`
}

type AMapTruckRoute struct {
	Origin      string                `json:"origin"`
	Destination string                `json:"destination"`
	Paths       []*AMapTruckPath      `json:"paths"`
}

type AMapTruckPath struct {
	Distance     string               `json:"distance"`
	Duration     string               `json:"duration"`
	TollDistance string               `json:"toll_distance"`
	Tolls        string               `json:"tolls"`
	TrafficLights string              `json:"traffic_lights"`
	Steps        []*AMapTruckStep     `json:"steps"`
	Restriction  []*AMapRestriction   `json:"restriction"`
}

type AMapTruckStep struct {
	Instruction string        `json:"instruction"`
	Direction   string        `json:"direction"`
	Distance    string        `json:"distance"`
	Duration    string        `json:"duration"`
	Polyline    string        `json:"polyline"`
	Action      string        `json:"action"`
	AssistantAction string     `json:"assistant_action"`
	Road        string        `json:"road"`
	Waypoints   []string      `json:"waypoints"`
}

type AMapRestriction struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Polyline    string `json:"polyline"`
}

func NewAMapTruckRouteClient(cfg *config.AMapConfig) *AMapTruckRouteClient {
	return &AMapTruckRouteClient{
		key: cfg.TruckKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

func (c *AMapTruckRouteClient) PlanRoute(
	ctx context.Context,
	req *AMapTruckRouteRequest,
) (*AMapRouteResponse, error) {
	if c.key == "" || c.key == "your-amap-truck-key" {
		logger.Global.Warn("AMap truck key not configured, will use fallback A* algorithm")
		return nil, fmt.Errorf("amap key not configured")
	}

	params := url.Values{}
	params.Set("key", c.key)
	params.Set("origin", req.Origin)
	params.Set("destination", req.Destination)
	params.Set("size", strconv.Itoa(req.Size))

	if req.Waypoints != "" {
		params.Set("waypoints", req.Waypoints)
	}
	if req.Height > 0 {
		params.Set("height", fmt.Sprintf("%.2f", req.Height))
	}
	if req.Width > 0 {
		params.Set("width", fmt.Sprintf("%.2f", req.Width))
	}
	if req.Length > 0 {
		params.Set("length", fmt.Sprintf("%.2f", req.Length))
	}
	if req.Weight > 0 {
		params.Set("weight", fmt.Sprintf("%.2f", req.Weight))
	}
	if req.AxleWeight > 0 {
		params.Set("axle_weight", fmt.Sprintf("%.2f", req.AxleWeight))
	}
	if req.AxleCount > 0 {
		params.Set("axle_count", strconv.Itoa(req.AxleCount))
	}
	if req.TrailerType > 0 {
		params.Set("trailer_type", strconv.Itoa(req.TrailerType))
	}
	if req.GoodsType > 0 {
		params.Set("goods_type", strconv.Itoa(req.GoodsType))
	}
	if req.IsHazardous > 0 {
		params.Set("is_hazardous", strconv.Itoa(req.IsHazardous))
	}
	if req.HazardousType > 0 {
		params.Set("hazardous_type", strconv.Itoa(req.HazardousType))
	}
	if req.Strategy > 0 {
		params.Set("strategy", strconv.Itoa(req.Strategy))
	}

	apiURL := fmt.Sprintf("https://restapi.amap.com/v3/direction/truckdriving?%s", params.Encode())
	logger.Global.Debug("AMap truck route request", zap.String("url", apiURL))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request amap: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	logger.Global.Debug("AMap truck route response", zap.String("body", string(body)))

	var result AMapRouteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Status != 1 {
		return &result, fmt.Errorf("amap error: %s (code: %s)", result.Info, result.Infocode)
	}

	return &result, nil
}

func (c *AMapTruckRouteClient) ConvertToRoutePlan(
	amapResp *AMapRouteResponse,
	req *model.RoutePlanRequest,
	strategy model.RouteStrategy,
) (*model.RoutePlan, error) {
	if amapResp == nil || amapResp.Route == nil || len(amapResp.Route.Paths) == 0 {
		return nil, fmt.Errorf("no route data from amap")
	}

	path := amapResp.Route.Paths[0]
	totalDistance, _ := strconv.ParseFloat(path.Distance, 64)
	totalDuration, _ := strconv.Atoi(path.Duration)
	totalTolls, _ := strconv.ParseFloat(path.Tolls, 64)

	var routePath []model.GeoPoint
	var segments []model.RouteSegment
	var restrictedSegs []model.RestrictedSegmentInfo

	for idx, step := range path.Steps {
		dist, _ := strconv.ParseFloat(step.Distance, 64)
		dur, _ := strconv.Atoi(step.Duration)
		speedLimit := 60

		start, end := parseStepPolyline(step.Polyline)
		seg := model.RouteSegment{
			Index:        idx,
			Start:        start,
			End:          end,
			Distance:     dist,
			Duration:     dur,
			RoadName:     step.Road,
			RoadType:     detectRoadType(step.Road),
			Instructions: step.Instruction,
			SpeedLimit:   speedLimit,
		}
		segments = append(segments, seg)

		points := parsePolylineToGeoPoints(step.Polyline)
		routePath = append(routePath, points...)
	}

	for _, restr := range path.Restriction {
		info := model.RestrictedSegmentInfo{
			AreaName:   restr.Description,
			AreaType:   convertRestrictionType(restr.Type),
			Reason:     restr.Description,
			Suggestion: "请按照导航提示绕行",
		}
		restrictedSegs = append(restrictedSegs, info)
	}

	geometry := map[string]interface{}{
		"type": "LineString",
		"coordinates": func() [][]float64 {
			coords := make([][]float64, 0, len(routePath))
			for _, p := range routePath {
				coords = append(coords, []float64{p.Lng, p.Lat})
			}
			return coords
		}(),
	}
	geomJSON, _ := json.Marshal(geometry)

	avoidTunnels := countRestrictionsByType(path.Restriction, "tunnel")
	avoidBridges := countRestrictionsByType(path.Restriction, "bridge")
	avoidPopulated := countRestrictionsByType(path.Restriction, "populated")
	avoidWater := countRestrictionsByType(path.Restriction, "water")

	safetyScore := calculateSafetyScore(totalDistance, avoidTunnels, avoidBridges, avoidPopulated, avoidWater)

	fuelCost := (totalDistance / 1000) * 0.3 * 7.5
	avgSpeed := 0.0
	if totalDuration > 0 {
		avgSpeed = totalDistance / float64(totalDuration) * 3.6
	}

	return &model.RoutePlan{
		PlanNo:               fmt.Sprintf("RP%s", time.Now().Format("20060102150405")),
		Strategy:             strategy,
		Origin:               req.Origin,
		Destination:          req.Destination,
		Waypoints:            req.Waypoints,
		RouteGeometry:        model.JSON(geomJSON),
		RoutePath:            routePath,
		Segments:             segments,
		TotalDistance:        totalDistance,
		EstimatedDuration:    totalDuration,
		ExpectedSpeed:        avgSpeed,
		TollFee:              totalTolls,
		FuelCost:             fuelCost,
		AvoidTunnels:         avoidTunnels,
		AvoidBridges:         avoidBridges,
		AvoidPopulated:       avoidPopulated,
		AvoidWaterProtection: avoidWater,
		RestrictedSegments:   restrictedSegs,
		SafetyScore:          safetyScore,
		Status:               "active",
		CreatedAt:            time.Now(),
	}, nil
}

func parseStepPolyline(polyline string) (model.GeoPoint, model.GeoPoint) {
	points := parsePolylineToGeoPoints(polyline)
	if len(points) >= 2 {
		return points[0], points[len(points)-1]
	}
	return model.GeoPoint{}, model.GeoPoint{}
}

func parsePolylineToGeoPoints(polyline string) []model.GeoPoint {
	var result []model.GeoPoint
	coords := strings.Split(polyline, ";")
	for _, coord := range coords {
		parts := strings.Split(coord, ",")
		if len(parts) == 2 {
			lng, _ := strconv.ParseFloat(parts[0], 64)
			lat, _ := strconv.ParseFloat(parts[1], 64)
			result = append(result, model.GeoPoint{Lat: lat, Lng: lng})
		}
	}
	return result
}

func detectRoadType(roadName string) string {
	switch {
	case strings.Contains(roadName, "高速") || strings.HasPrefix(roadName, "G") && len(roadName) <= 4:
		return "highway"
	case strings.Contains(roadName, "国道") || (strings.HasPrefix(roadName, "G") && len(roadName) == 5):
		return "national"
	case strings.Contains(roadName, "省道") || strings.HasPrefix(roadName, "S"):
		return "provincial"
	default:
		return "urban"
	}
}

func convertRestrictionType(amapType string) model.RestrictedAreaType {
	switch amapType {
	case "tunnel":
		return model.AreaTypeTunnel
	case "bridge":
		return model.AreaTypeBridge
	case "school":
		return model.AreaTypeSchool
	case "hospital":
		return model.AreaTypeHospital
	case "water":
		return model.AreaTypeWaterProtection
	case "height_limit":
		return model.AreaTypeHeightLimit
	case "weight_limit":
		return model.AreaTypeWeightLimit
	default:
		return model.AreaTypeMall
	}
}

func countRestrictionsByType(restrs []*AMapRestriction, typ string) int {
	count := 0
	for _, r := range restrs {
		if r.Type == typ {
			count++
		}
	}
	return count
}

func calculateSafetyScore(distance float64, tunnels, bridges, populated, water int) float64 {
	score := 100.0
	score -= float64(tunnels) * 8
	score -= float64(bridges) * 4
	score -= float64(populated) * 10
	score -= float64(water) * 12

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func ConvertStrategyToAMap(strategy model.RouteStrategy) int {
	switch strategy {
	case model.StrategyShortest:
		return 2
	case model.StrategySafest:
		return 4
	case model.StrategyEconomic:
		return 3
	default:
		return 1
	}
}

func ConvertHazardClassToAMap(hazardClass string) (int, int) {
	isHazardous := 0
	hazardousType := 0

	if hazardClass != "" {
		isHazardous = 1
		switch hazardClass {
		case "explosive", "1":
			hazardousType = 1
		case "gas", "2":
			hazardousType = 2
		case "flammable_liquid", "3":
			hazardousType = 3
		case "flammable_solid", "4":
			hazardousType = 4
		case "oxidizer", "5":
			hazardousType = 5
		case "toxic", "6":
			hazardousType = 6
		case "radioactive", "7":
			hazardousType = 7
		case "corrosive", "8":
			hazardousType = 8
		default:
			hazardousType = 9
		}
	}
	return isHazardous, hazardousType
}
