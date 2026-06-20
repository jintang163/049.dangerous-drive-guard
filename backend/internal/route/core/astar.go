package algorithm

import (
	"container/heap"
	"math"
	"sync"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type Graph struct {
	Nodes    map[int64]*GraphNode
	Edges    map[int64][]*GraphEdge
	mu       sync.RWMutex
	nodeSeq  int64
}

type GraphNode struct {
	ID    int64
	Point model.GeoPoint
	Data  map[string]interface{}
}

type GraphEdge struct {
	FromID    int64
	ToID      int64
	Distance  float64
	BaseCost  float64
	RoadType  string
	RoadName  string
	NumLanes  int
	SpeedLimit int
	HasToll   bool
	TollFee   float64
	HasTunnel bool
	HasBridge bool
	HeightLimit float64
	WeightLimit float64
	RestrictedAreaIDs []int64
	Attributes map[string]interface{}
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:   make(map[int64]*GraphNode),
		Edges:   make(map[int64][]*GraphEdge),
		nodeSeq: 1,
	}
}

func (g *Graph) AddNode(point model.GeoPoint, data map[string]interface{}) int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	id := g.nodeSeq
	g.nodeSeq++
	g.Nodes[id] = &GraphNode{
		ID:    id,
		Point: point,
		Data:  data,
	}
	return id
}

func (g *Graph) AddEdge(from, to int64, edge *GraphEdge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	edge.FromID = from
	edge.ToID = to
	if edge.Distance <= 0 {
		fromNode := g.Nodes[from]
		toNode := g.Nodes[to]
		if fromNode != nil && toNode != nil {
			edge.Distance = fromNode.Point.DistanceTo(toNode.Point)
		}
	}
	if edge.BaseCost <= 0 {
		edge.BaseCost = edge.Distance
	}
	g.Edges[from] = append(g.Edges[from], edge)
}

type PlannerConfig struct {
	Strategy                 model.RouteStrategy
	VehicleHeight            float64
	VehicleWeight            float64
	VehicleType              model.VehicleType
	HazardClass              string
	RestrictedAreas          []*model.RestrictedArea
	CustomWeights            map[string]float64
	AvoidTunnels             bool
	AvoidBridges             bool
	AvoidPopulatedAreas      bool
	AvoidWaterProtection     bool
	MaxAlternativeRoutes     int
}

type pqItem struct {
	nodeID   int64
	cost     float64
	distance float64
	duration int
	index    int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].cost < pq[j].cost
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

type AStarPlanner struct {
	graph  *Graph
	config PlannerConfig
}

func NewAStarPlanner(graph *Graph, config PlannerConfig) *AStarPlanner {
	return &AStarPlanner{
		graph:  graph,
		config: config,
	}
}

func (p *AStarPlanner) getDefaultWeights(strategy model.RouteStrategy) map[string]float64 {
	switch strategy {
	case model.StrategyShortest:
		return map[string]float64{
			"distance":         1.0,
			"time":             0.0,
			"safety":           0.0,
			"toll":             0.0,
			"tunnel_penalty":   10000,
			"bridge_penalty":   500,
			"populated_penalty": 2000,
			"water_penalty":    5000,
			"restricted_level1": 3000,
			"restricted_level2": 100000,
			"high_weight":      0.8,
			"normal_weight":    1.0,
		}
	case model.StrategySafest:
		return map[string]float64{
			"distance":         0.3,
			"time":             0.0,
			"safety":           2.0,
			"toll":             0.0,
			"tunnel_penalty":   50000,
			"bridge_penalty":   30000,
			"populated_penalty": 30000,
			"water_penalty":    100000,
			"restricted_level1": 15000,
			"restricted_level2": 500000,
			"high_weight":      2.0,
			"normal_weight":    1.0,
		}
	case model.StrategyEconomic:
		return map[string]float64{
			"distance":         0.6,
			"time":             0.2,
			"safety":           0.3,
			"toll":             0.5,
			"tunnel_penalty":   5000,
			"bridge_penalty":   300,
			"populated_penalty": 1000,
			"water_penalty":    2000,
			"restricted_level1": 1500,
			"restricted_level2": 50000,
			"high_weight":      0.5,
			"normal_weight":    1.0,
		}
	default:
		return map[string]float64{
			"distance":         1.0,
			"time":             0.5,
			"safety":           1.0,
			"toll":             0.3,
			"tunnel_penalty":   10000,
			"bridge_penalty":   1000,
			"populated_penalty": 5000,
			"water_penalty":    10000,
			"restricted_level1": 3000,
			"restricted_level2": 100000,
			"high_weight":      1.0,
			"normal_weight":    1.0,
		}
	}
}

func (p *AStarPlanner) mergeWeights() map[string]float64 {
	base := p.getDefaultWeights(p.config.Strategy)
	if p.config.CustomWeights != nil {
		for k, v := range p.config.CustomWeights {
			base[k] = v
		}
	}
	return base
}

func (p *AStarPlanner) checkRestricted(edge *GraphEdge, weights map[string]float64) (float64, bool, []string) {
	penalty := 0.0
	blocked := false
	var reasons []string

	if edge.HasTunnel {
		penalty += weights["tunnel_penalty"]
		reasons = append(reasons, "tunnel")
	}
	if edge.HasBridge {
		penalty += weights["bridge_penalty"]
		reasons = append(reasons, "bridge")
	}

	if edge.HeightLimit > 0 && p.config.VehicleHeight > edge.HeightLimit {
		penalty += weights["restricted_level2"]
		blocked = true
		reasons = append(reasons, "height_limit_exceeded")
	}

	if edge.WeightLimit > 0 && p.config.VehicleWeight > edge.WeightLimit {
		penalty += weights["restricted_level2"]
		blocked = true
		reasons = append(reasons, "weight_limit_exceeded")
	}

	for _, areaID := range edge.RestrictedAreaIDs {
		for _, area := range p.config.RestrictedAreas {
			if area.ID != areaID {
				continue
			}

			matchesHazard := true
			if area.RestrictHazardClasses != "" && p.config.HazardClass != "" {
				matchesHazard = false
				classes := splitAndTrim(area.RestrictHazardClasses)
				for _, c := range classes {
					if c == p.config.HazardClass {
						matchesHazard = true
						break
					}
				}
			}
			if !matchesHazard {
				continue
			}

			levelKey := "restricted_level2"
			if area.Level == 1 {
				levelKey = "restricted_level1"
			}
			penalty += weights[levelKey]
			if area.Level >= 2 {
				blocked = true
			}

			switch area.AreaType {
			case model.AreaTypeSchool, model.AreaTypeHospital, model.AreaTypeMall:
				penalty += weights["populated_penalty"]
				reasons = append(reasons, string(area.AreaType))
			case model.AreaTypeWaterProtection:
				penalty += weights["water_penalty"]
				reasons = append(reasons, "water_protection")
			case model.AreaTypeTunnel:
				if p.config.AvoidTunnels {
					penalty += weights["tunnel_penalty"] * 2
					blocked = true
				}
			case model.AreaTypeBridge:
				if p.config.AvoidBridges {
					penalty += weights["bridge_penalty"] * 2
				}
			}
		}
	}

	return penalty, blocked, reasons
}

func (p *AStarPlanner) calcEdgeCost(edge *GraphEdge, weights map[string]float64) (float64, int, []string, bool) {
	distance := edge.Distance
	baseSpeed := float64(edge.SpeedLimit)
	if baseSpeed <= 0 {
		switch edge.RoadType {
		case "highway":
			baseSpeed = 80
		case "national":
			baseSpeed = 60
		case "provincial":
			baseSpeed = 50
		case "urban":
			baseSpeed = 40
		default:
			baseSpeed = 50
		}
	}
	duration := int(math.Ceil(distance / (baseSpeed / 3.6)))

	baseCost := weights["distance"]*distance +
		weights["time"]*float64(duration) +
		weights["toll"]*edge.TollFee*100

	if edge.RoadType == "highway" {
		baseCost *= weights["high_weight"]
	} else {
		baseCost *= weights["normal_weight"]
	}

	penalty, blocked, reasons := p.checkRestricted(edge, weights)
	totalCost := baseCost + penalty

	return totalCost, duration, reasons, blocked
}

func (p *AStarPlanner) heuristic(node, goal *GraphNode) float64 {
	return node.Point.DistanceTo(goal.Point)
}

type PlanResult struct {
	Path            []int64
	Nodes           []*GraphNode
	Edges           []*GraphEdge
	TotalCost       float64
	TotalDistance   float64
	TotalDuration   int
	TotalTollFee    float64
	AvoidedTunnels  int
	AvoidedBridges  int
	AvoidedPopulated int
	AvoidedWater    int
	RestrictedSegs  []model.RestrictedSegmentInfo
	SafetyScore     float64
	SegmentReasons  map[int][]string
}

func (p *AStarPlanner) Plan(startID, endID int64) (*PlanResult, error) {
	p.graph.mu.RLock()
	startNode := p.graph.Nodes[startID]
	endNode := p.graph.Nodes[endID]
	p.graph.mu.RUnlock()

	if startNode == nil || endNode == nil {
		return nil, nil
	}

	weights := p.mergeWeights()
	pq := &priorityQueue{}
	heap.Init(pq)

	cameFrom := make(map[int64]int64)
	edgeFrom := make(map[int64]*GraphEdge)
	gScore := make(map[int64]float64)
	distance := make(map[int64]float64)
	duration := make(map[int64]int)

	gScore[startID] = 0
	distance[startID] = 0
	duration[startID] = 0

	heap.Push(pq, &pqItem{
		nodeID:   startID,
		cost:     0 + p.heuristic(startNode, endNode),
		distance: 0,
		duration: 0,
	})

	segmentReasons := make(map[int][]string)

	maxIterations := len(p.graph.Nodes) * 10
	iterations := 0

	for pq.Len() > 0 && iterations < maxIterations {
		iterations++
		current := heap.Pop(pq).(*pqItem)

		if current.nodeID == endID {
			break
		}

		if currentCost, exists := gScore[current.nodeID]; exists && currentCost < current.cost {
			continue
		}

		p.graph.mu.RLock()
		edges := p.graph.Edges[current.nodeID]
		p.graph.mu.RUnlock()

		for _, edge := range edges {
			edgeCost, edgeDuration, reasons, blocked := p.calcEdgeCost(edge, weights)
			if blocked {
				continue
			}

			tentativeG := gScore[current.nodeID] + edgeCost
			if existing, ok := gScore[edge.ToID]; !ok || tentativeG < existing {
				cameFrom[edge.ToID] = current.nodeID
				edgeFrom[edge.ToID] = edge
				gScore[edge.ToID] = tentativeG
				distance[edge.ToID] = distance[current.nodeID] + edge.Distance
				duration[edge.ToID] = duration[current.nodeID] + edgeDuration

				if len(reasons) > 0 {
					segmentReasons[edge.ToID] = reasons
				}

				p.graph.mu.RLock()
				toNode := p.graph.Nodes[edge.ToID]
				p.graph.mu.RUnlock()

				if toNode != nil {
					f := tentativeG + p.heuristic(toNode, endNode)
					heap.Push(pq, &pqItem{
						nodeID:   edge.ToID,
						cost:     f,
						distance: distance[edge.ToID],
						duration: duration[edge.ToID],
					})
				}
			}
		}
	}

	if _, ok := cameFrom[endID]; endID != startID && !ok {
		return nil, nil
	}

	result := &PlanResult{
		SegmentReasons: segmentReasons,
	}

	var pathIDs []int64
	var edges []*GraphEdge
	current := endID
	for current != startID {
		pathIDs = append([]int64{current}, pathIDs...)
		if e, ok := edgeFrom[current]; ok {
			edges = append([]*GraphEdge{e}, edges...)
		}
		if prev, ok := cameFrom[current]; ok {
			current = prev
		} else {
			break
		}
	}
	pathIDs = append([]int64{startID}, pathIDs...)

	result.Path = pathIDs
	result.Edges = edges
	result.TotalDistance = distance[endID]
	result.TotalDuration = duration[endID]
	result.TotalCost = gScore[endID]

	p.graph.mu.RLock()
	for _, id := range pathIDs {
		if n, ok := p.graph.Nodes[id]; ok {
			result.Nodes = append(result.Nodes, n)
		}
	}
	p.graph.mu.RUnlock()

	for _, e := range edges {
		result.TotalTollFee += e.TollFee
		if e.HasTunnel {
			result.AvoidedTunnels++
		}
		if e.HasBridge {
			result.AvoidedBridges++
		}
		for _, areaID := range e.RestrictedAreaIDs {
			for _, area := range p.config.RestrictedAreas {
				if area.ID == areaID {
					switch area.AreaType {
					case model.AreaTypeSchool, model.AreaTypeHospital, model.AreaTypeMall:
						result.AvoidedPopulated++
					case model.AreaTypeWaterProtection:
						result.AvoidedWater++
					}
				}
			}
		}
	}

	result.SafetyScore = p.calcSafetyScore(result)
	logger.Sugar.Infof("A* plan done: nodes=%d, dist=%.2fm, dur=%ds, safety=%.1f",
		len(pathIDs), result.TotalDistance, result.TotalDuration, result.SafetyScore)

	return result, nil
}

func (p *AStarPlanner) calcSafetyScore(r *PlanResult) float64 {
	score := 100.0
	weights := p.mergeWeights()

	basePenalty := weights["tunnel_penalty"]
	if basePenalty > 0 {
		score -= float64(r.AvoidedTunnels) * 8
		score -= float64(r.AvoidedBridges) * 4
		score -= float64(r.AvoidedPopulated) * 10
		score -= float64(r.AvoidedWater) * 12
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func splitAndTrim(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
