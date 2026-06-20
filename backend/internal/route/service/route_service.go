package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	routecore "github.com/dangerous-drive-guard/backend/internal/route/core"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type RouteService struct {
	db         *database.TIDB
	config     *config.RouteConfig
	graphCache *routecore.Graph
	graphMu    sync.RWMutex
}

func NewRouteService(cfg *config.Config) *RouteService {
	return &RouteService{
		db:         database.GetDB(),
		graphCache: routecore.NewGraph(),
	}
}

func (s *RouteService) buildRoadGraph(ctx context.Context, origin, dest model.GeoPoint, restrictedAreas []*model.RestrictedArea) *routecore.Graph {
	g := routecore.NewGraph()

	originID := g.AddNode(origin, map[string]interface{}{"type": "origin"})
	destID := g.AddNode(dest, map[string]interface{}{"type": "destination"})

	centerLat := (origin.Lat + dest.Lat) / 2
	centerLng := (origin.Lng + dest.Lng) / 2
	latSpan := math.Abs(origin.Lat-dest.Lat) + 0.5
	lngSpan := math.Abs(origin.Lng-dest.Lng) + 0.5

	gridSize := 5
	nodeGrid := make([][]int64, gridSize)
	for i := 0; i < gridSize; i++ {
		nodeGrid[i] = make([]int64, gridSize)
		for j := 0; j < gridSize; j++ {
			lat := centerLat + (float64(i)-float64(gridSize-1)/2) * (latSpan / float64(gridSize-1))
			lng := centerLng + (float64(j)-float64(gridSize-1)/2) * (lngSpan / float64(gridSize-1))

			point := model.GeoPoint{Lat: lat, Lng: lng}
			data := map[string]interface{}{
				"type":  "waypoint",
				"grid_i": i,
				"grid_j": j,
			}
			nodeGrid[i][j] = g.AddNode(point, data)
		}
	}

	roadTypes := []string{"highway", "national", "provincial", "urban"}
	speedLimits := map[string]int{
		"highway": 90,
		"national": 70,
		"provincial": 60,
		"urban": 50,
	}
	roadNames := map[string]string{
		"highway":    "G4京港澳高速",
		"national":   "G107国道",
		"provincial": "S325省道",
		"urban":      "市政道路",
	}

	rand.Seed(time.Now().UnixNano())

	addRoadEdge := func(from, to int64, roadType string, addRestricted bool) {
		fromNode := g.Nodes[from]
		toNode := g.Nodes[to]
		if fromNode == nil || toNode == nil {
			return
		}

		dist := fromNode.Point.DistanceTo(toNode.Point)
		edge := &routecore.GraphEdge{
			Distance:   dist,
			RoadType:   roadType,
			RoadName:   roadNames[roadType],
			NumLanes:   2 + rand.Intn(4),
			SpeedLimit: speedLimits[roadType],
			HasToll:   roadType == "highway",
			TollFee:   0,
		}

		if roadType == "highway" {
			edge.TollFee = dist / 1000 * 0.5
		}

		if addRestricted {
			if rand.Float64() < 0.1 {
				edge.HasTunnel = true
			}
			if rand.Float64() < 0.08 {
				edge.HasBridge = true
			}
			for _, area := range restrictedAreas {
				areaCenter := model.GeoPoint{Lat: area.CenterLatitude, Lng: area.CenterLongitude}
				midPoint := model.GeoPoint{
					Lat: (fromNode.Point.Lat + toNode.Point.Lat) / 2,
					Lng: (fromNode.Point.Lng + toNode.Point.Lng) / 2,
				}
				if midPoint.DistanceTo(areaCenter) < area.Radius*1.5 {
					edge.RestrictedAreaIDs = append(edge.RestrictedAreaIDs, area.ID)
				}
			}
			if roadType != "highway" && rand.Float64() < 0.05 {
				edge.HeightLimit = 3.5 + rand.Float64()*1
				edge.WeightLimit = 20 + rand.Float64()*20
			}
		}

		g.AddEdge(from, to, edge)
	}

	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			current := nodeGrid[i][j]
			if j+1 < gridSize {
				rt := roadTypes[rand.Intn(len(roadTypes))]
				addRoadEdge(current, nodeGrid[i][j+1], rt, true)
				addRoadEdge(nodeGrid[i][j+1], current, rt, true)
			}
			if i+1 < gridSize {
				rt := roadTypes[rand.Intn(len(roadTypes))]
				addRoadEdge(current, nodeGrid[i+1][j], rt, true)
				addRoadEdge(nodeGrid[i+1][j], current, rt, true)
			}
			if i+1 < gridSize && j+1 < gridSize {
				addRoadEdge(current, nodeGrid[i+1][j+1], "provincial", false)
			}
		}
	}

	connectOrigin := func(gridI, gridJ int, roadType string) {
		if gridI >= 0 && gridI < gridSize && gridJ >= 0 && gridJ < gridSize {
			addRoadEdge(originID, nodeGrid[gridI][gridJ], roadType, false)
			addRoadEdge(nodeGrid[gridI][gridJ], originID, roadType, false)
		}
	}
	connectOrigin(0, 0, "national")
	connectOrigin(0, 1, "urban")
	connectOrigin(1, 0, "provincial")

	connectDest := func(gridI, gridJ int, roadType string) {
		if gridI >= 0 && gridI < gridSize && gridJ >= 0 && gridJ < gridSize {
			addRoadEdge(destID, nodeGrid[gridI][gridJ], roadType, false)
			addRoadEdge(nodeGrid[gridI][gridJ], destID, roadType, false)
		}
	}
	connectDest(gridSize-1, gridSize-1, "national")
	connectDest(gridSize-1, gridSize-2, "urban")
	connectDest(gridSize-2, gridSize-1, "provincial")

	return g
}

func (s *RouteService) GetRestrictedAreas(ctx context.Context, hazardClass string, vehicleType model.VehicleType) ([]*model.RestrictedArea, error) {
	var areas []*model.RestrictedArea
	query := s.db.WithContext(ctx).Where("status = ?", 1)
	rows, err := query.Find(&areas).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.RestrictedArea
	for rows.Next() {
		var a model.RestrictedArea
		_ = s.db.ScanRows(rows, &a)
		if a.RestrictHazardClasses != "" && hazardClass != "" {
			matches := false
			for _, c := range strings.Split(a.RestrictHazardClasses, ",") {
				if strings.TrimSpace(c) == hazardClass {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}
		result = append(result, &a)
	}
	return result, nil
}

func (s *RouteService) buildPlanResult(
	ctx context.Context,
	req *model.RoutePlanRequest,
	result *routecore.PlanResult,
	graph *routecore.Graph,
) (*model.RoutePlan, error) {
	path := make([]model.GeoPoint, 0, len(result.Nodes))
	for _, n := range result.Nodes {
		path = append(path, n.Point)
	}

	segments := make([]model.RouteSegment, 0, len(result.Edges))
	totalToll := 0.0
	for idx, e := range result.Edges {
		fromNode := graph.Nodes[e.FromID]
		toNode := graph.Nodes[e.ToID]
		seg := model.RouteSegment{
			Index:         idx,
			Start:         fromNode.Point,
			End:           toNode.Point,
			Distance:      e.Distance,
			Duration:      int(e.Distance / (float64(e.SpeedLimit) / 3.6)),
			RoadName:      e.RoadName,
			RoadType:      e.RoadType,
			HasToll:       e.HasToll,
			TollFee:       e.TollFee,
			SpeedLimit:    e.SpeedLimit,
			RestrictedIDs: e.RestrictedAreaIDs,
		}
		if e.HasToll {
			totalToll += e.TollFee
		}
		if len(e.RestrictedAreaIDs) > 0 {
			seg.Restriction = "restricted_area"
		}
		segments = append(segments, seg)
	}

	geometry := map[string]interface{}{
		"type": "LineString",
		"coordinates": func() [][]float64 {
			coords := make([][]float64, 0, len(path))
			for _, p := range path {
				coords = append(coords, []float64{p.Lng, p.Lat})
			}
			return coords
		}(),
	}
	geomJSON, _ := json.Marshal(geometry)

	restrictedSegs := make([]model.RestrictedSegmentInfo, 0)
	seenAreas := make(map[int64]bool)
	for _, seg := range segments {
		for _, areaID := range seg.RestrictedIDs {
			if seenAreas[areaID] {
				continue
			}
			seenAreas[areaID] = true
			var area *model.RestrictedArea
			_ = s.db.WithContext(ctx).First(&area, areaID).Error
			if area == nil {
				continue
			}
			info := model.RestrictedSegmentInfo{
				AreaID:     areaID,
				AreaName:   area.Name,
				AreaType:   area.AreaType,
				Level:      area.Level,
				EntryPoint: seg.Start,
				ExitPoint:  seg.End,
				Distance:   seg.Distance,
				Reason:     s.buildRestrictionReason(area),
				Suggestion: s.buildRestrictionSuggestion(area),
			}
			restrictedSegs = append(restrictedSegs, info)
		}
	}

	avgSpeed := 0.0
	if result.TotalDuration > 0 {
		avgSpeed = result.TotalDistance / float64(result.TotalDuration) * 3.6
	}

	fuelCost := (result.TotalDistance / 1000) * 0.3 * 7.5

	return &model.RoutePlan{
		PlanNo:               fmt.Sprintf("RP%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:8], "-", ""))),
		WaybillID:            req.WaybillID,
		VehicleID:            req.VehicleID,
		DriverID:             req.DriverID,
		Strategy:             req.Strategy,
		Origin:               req.Origin,
		Destination:          req.Destination,
		Waypoints:            req.Waypoints,
		RouteGeometry:        model.JSON(geomJSON),
		RoutePath:            path,
		Segments:             segments,
		TotalDistance:        result.TotalDistance,
		EstimatedDuration:    result.TotalDuration,
		ExpectedSpeed:        math.Round(avgSpeed*100) / 100,
		TollFee:              math.Round(totalToll*100) / 100,
		FuelCost:             math.Round(fuelCost*100) / 100,
		AvoidTunnels:         result.AvoidedTunnels,
		AvoidBridges:         result.AvoidedBridges,
		AvoidPopulated:       result.AvoidedPopulated,
		AvoidWaterProtection: result.AvoidedWater,
		RestrictedSegments:   restrictedSegs,
		SafetyScore:          math.Round(result.SafetyScore*100) / 100,
		Status:               "active",
		CreatedAt:            time.Now(),
	}, nil
}

func (s *RouteService) buildRestrictionReason(area *model.RestrictedArea) string {
	switch area.AreaType {
	case model.AreaTypeSchool:
		return fmt.Sprintf("%s - 学校区域，学生密集，禁危化品车辆通过", area.Name)
	case model.AreaTypeHospital:
		return fmt.Sprintf("%s - 医院区域，人员密集，禁危化品车辆通过", area.Name)
	case model.AreaTypeMall:
		return fmt.Sprintf("%s - 商业圈，人口密集，禁危化品车辆通过", area.Name)
	case model.AreaTypeTunnel:
		return fmt.Sprintf("%s - 隧道，空间封闭，事故逃生困难", area.Name)
	case model.AreaTypeBridge:
		return fmt.Sprintf("%s - 桥梁，承重有限，坠河风险高", area.Name)
	case model.AreaTypeWaterProtection:
		return fmt.Sprintf("%s - 水源保护区，严禁危险品泄漏", area.Name)
	case model.AreaTypeHeightLimit:
		return fmt.Sprintf("限高路段，限高%.1f米", area.HeightLimit)
	case model.AreaTypeWeightLimit:
		return fmt.Sprintf("限重路段，限重%.0f吨", area.WeightLimit)
	default:
		return fmt.Sprintf("%s - 限行区域", area.Name)
	}
}

func (s *RouteService) buildRestrictionSuggestion(area *model.RestrictedArea) string {
	switch area.AreaType {
	case model.AreaTypeTunnel:
		return "建议选择地面道路绕行，预留15-30分钟时间"
	case model.AreaTypeWaterProtection:
		return "严禁穿越，必须绕道行驶"
	case model.AreaTypeHeightLimit:
		return "请选择无高度限制的道路通行"
	default:
		return "请按照导航路线绕行"
	}
}

func (s *RouteService) PlanRoute(ctx context.Context, req *model.RoutePlanRequest) (*model.RoutePlan, error) {
	logger.Global.Info("planning route",
		zap.String("strategy", string(req.Strategy)),
		zap.Float64("from_lat", req.Origin.Latitude),
		zap.Float64("from_lng", req.Origin.Longitude),
	)

	restrictedAreas, err := s.GetRestrictedAreas(ctx, req.HazardClass, req.VehicleType)
	if err != nil {
		return nil, err
	}

	origin := model.GeoPoint{Lat: req.Origin.Latitude, Lng: req.Origin.Longitude}
	dest := model.GeoPoint{Lat: req.Destination.Latitude, Lng: req.Destination.Longitude}

	graph := s.buildRoadGraph(ctx, origin, dest, restrictedAreas)

	plannerCfg := routecore.PlannerConfig{
		Strategy:             req.Strategy,
		VehicleHeight:        req.VehicleHeight,
		VehicleWeight:        req.VehicleWeight,
		VehicleType:          req.VehicleType,
		HazardClass:          req.HazardClass,
		RestrictedAreas:      restrictedAreas,
		CustomWeights:        req.CustomWeights,
		AvoidTunnels:         true,
		AvoidBridges:         req.Strategy == model.StrategySafest,
		AvoidPopulatedAreas:  req.Strategy == model.StrategySafest,
		AvoidWaterProtection: true,
	}
	planner := routecore.NewAStarPlanner(graph, plannerCfg)

	var startID, endID int64
	for id, node := range graph.Nodes {
		if t, ok := node.Data["type"]; ok && t == "origin" {
			startID = id
		}
		if t, ok := node.Data["type"]; ok && t == "destination" {
			endID = id
		}
	}

	result, err := planner.Plan(startID, endID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("no valid route found, please adjust planning parameters")
	}

	plan, err := s.buildPlanResult(ctx, req, result, graph)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *RouteService) PlanMultiStrategy(ctx context.Context, req *model.RoutePlanRequest) (*model.MultiStrategyPlan, error) {
	result := &model.MultiStrategyPlan{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0, 3)

	strategies := []struct {
		key  model.RouteStrategy
		field **model.RoutePlan
	}{
		{model.StrategyShortest, &result.Shortest},
		{model.StrategySafest, &result.Safest},
		{model.StrategyEconomic, &result.Economic},
	}

	for _, st := range strategies {
		wg.Add(1)
		go func(st model.RouteStrategy, field **model.RoutePlan) {
			defer wg.Done()
			r := *req
			r.Strategy = st
			plan, err := s.PlanRoute(ctx, &r)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			*field = plan
		}(st.key, st.field)
	}
	wg.Wait()

	if result.Shortest == nil && result.Safest == nil && result.Economic == nil {
		if len(errs) > 0 {
			return nil, errs[0]
		}
		return nil, fmt.Errorf("no valid route found")
	}

	return result, nil
}

func (s *RouteService) RecommendServiceAreas(ctx context.Context, routePlanID int64, currentPoint model.GeoPoint, fatigueLevel string) ([]*model.ServiceArea, error) {
	var areas []*model.ServiceArea

	query := s.db.WithContext(ctx).Where("status = ?", 1)
	rows, err := query.Find(&areas).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*model.ServiceArea, 0)
	for rows.Next() {
		var a model.ServiceArea
		_ = s.db.ScanRows(rows, &a)
		dist := currentPoint.DistanceTo(model.GeoPoint{Lat: a.Latitude, Lng: a.Longitude})
		a.DistanceFromCurrent = math.Round(dist/1000*100) / 100

		if dist > 150000 {
			continue
		}
		if fatigueLevel == "fatigue" && !a.HasDangerParking {
			continue
		}
		result = append(result, &a)
	}

	for i := range result {
		speed := 60.0
		eta := int(math.Ceil(result[i].DistanceFromCurrent / speed * 60))
		result[i].EstimatedArrivalTime = eta
		switch fatigueLevel {
		case "fatigue":
			result[i].RestDurationRecommend = 30
		case "warning":
			result[i].RestDurationRecommend = 20
		default:
			result[i].RestDurationRecommend = 15
		}
	}

	return result, nil
}
