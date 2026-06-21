package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/cloudwego/hertz/pkg/app"
	hzws "github.com/hertz-contrib/websocket"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

var upgrader = hzws.NewHertzUpgrader(func(c *hzws.Upgrader) {
	c.Upgrader.CheckOrigin = func(r *app.RequestContext) bool {
		return true
	}
	c.ReadBufferSize = 1024
	c.WriteBufferSize = 1024
})

type MessageType string

const (
	MsgVehicleStatus        MessageType = "vehicle_status"
	MsgAlarmNotify          MessageType = "alarm_notify"
	MsgVehicleTrack         MessageType = "vehicle_track"
	MsgDriverFatigue        MessageType = "driver_fatigue"
	MsgIntercomMessage      MessageType = "intercom_message"
	MsgDispatchCommand      MessageType = "dispatch_command"
	MsgWaybillUpdate        MessageType = "waybill_update"
	MsgWeatherAlert         MessageType = "weather_alert"
	MsgHeartbeat            MessageType = "heartbeat"
	MsgRestrictedAreaUpdate MessageType = "restricted_area_update"
	MsgTrafficEvent         MessageType = "traffic_event"
	MsgRouteReplanSuggest   MessageType = "route_replan_suggest"
	MsgRouteReplanConfirmed MessageType = "route_replan_confirmed"
	MsgRouteApplied         MessageType = "route_applied"
	MsgSOSAlert             MessageType = "sos_alert"
	MsgEscortPolling        MessageType = "escort_polling"
	MsgSubscribe            MessageType = "subscribe"
	MsgError                MessageType = "error"
)

type WSMessage struct {
	Type      MessageType    `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Data      interface{}    `json:"data"`
	TraceID   string         `json:"trace_id,omitempty"`
}

type Client struct {
	ID         string
	UserID     int64
	Role       string
	OrgID      int64
	Conn       *hzws.Conn
	Send       chan []byte
	Subscribed map[string]bool
	mu         sync.Mutex
}

type Hub struct {
	clients        map[*Client]bool
	monitorClients map[int64][]*Client
	vehicleClients map[int64][]*Client
	broadcast      chan *WSMessage
	register       chan *Client
	unregister     chan *Client
	db             *database.TIDB
	mu             sync.RWMutex
}

var GlobalHub *Hub
var hubOnce sync.Once

func GetHub() *Hub {
	hubOnce.Do(func() {
		GlobalHub = &Hub{
			clients:        make(map[*Client]bool),
			monitorClients: make(map[int64][]*Client),
			vehicleClients: make(map[int64][]*Client),
			broadcast:      make(chan *WSMessage, 1024),
			register:       make(chan *Client, 256),
			unregister:     make(chan *Client, 256),
			db:             database.GetDB(),
		}
		go GlobalHub.run()
	})
	return GlobalHub
}

func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logger.Sugar.Infof("WS client registered: id=%s, user=%d", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				for orgID, clients := range h.monitorClients {
					for i, c := range clients {
						if c == client {
							h.monitorClients[orgID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
				}
				for vid, clients := range h.vehicleClients {
					for i, c := range clients {
						if c == client {
							h.vehicleClients[vid] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()
			logger.Sugar.Infof("WS client unregistered: id=%s", client.ID)

		case msg := <-h.broadcast:
			h.broadcastMessage(msg)

		case <-ticker.C:
			h.sendHeartbeats()
		}
	}
}

func (h *Hub) broadcastMessage(msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, _ := json.Marshal(msg)
	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) sendToOrg(orgID int64, msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, _ := json.Marshal(msg)
	if clients, ok := h.monitorClients[orgID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) sendToVehicle(vehicleID int64, msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, _ := json.Marshal(msg)
	if clients, ok := h.vehicleClients[vehicleID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) sendHeartbeats() {
	heartbeat := &WSMessage{
		Type:      MsgHeartbeat,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"server_time": time.Now().Format(time.RFC3339),
		},
	}
	h.broadcastMessage(heartbeat)
}

func (h *Hub) BroadcastVehicleStatus(ctx context.Context, status *model.RealtimeVehicleStatus) {
	msg := &WSMessage{
		Type:      MsgVehicleStatus,
		Timestamp: time.Now().Unix(),
		Data:      status,
	}
	if status != nil {
		h.sendToVehicle(status.VehicleID, msg)
		h.sendToOrg(1, msg)
	}
}

func (h *Hub) BroadcastAlarm(ctx context.Context, alarm *model.FatigueAlarm) {
	msg := &WSMessage{
		Type:      MsgAlarmNotify,
		Timestamp: time.Now().Unix(),
		Data:      alarm,
	}
	h.broadcast <- msg
}

func (h *Hub) BroadcastTrack(ctx context.Context, track *model.VehicleTrack) {
	msg := &WSMessage{
		Type:      MsgVehicleTrack,
		Timestamp: time.Now().Unix(),
		Data:      track,
	}
	if track != nil {
		h.sendToVehicle(track.VehicleID, msg)
	}
}

func (h *Hub) SendIntercomToVehicle(vehicleID int64, message string, priority int) {
	msg := &WSMessage{
		Type:      MsgIntercomMessage,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"vehicle_id": vehicleID,
			"message":    message,
			"priority":   priority,
			"timestamp":  time.Now().Format("15:04:05"),
		},
	}
	h.sendToVehicle(vehicleID, msg)
}

func (h *Hub) DispatchCommand(vehicleID int64, cmdType string, payload map[string]interface{}) {
	msg := &WSMessage{
		Type:      MsgDispatchCommand,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"vehicle_id": vehicleID,
			"command":    cmdType,
			"payload":    payload,
		},
	}
	h.sendToVehicle(vehicleID, msg)
}

type RestrictedAreaSyncPayload struct {
	Event     string      `json:"event"`
	AreaID    int64       `json:"area_id"`
	AreaName  string      `json:"area_name"`
	AreaType  string      `json:"area_type"`
	Level     int         `json:"level"`
	ShapeType string      `json:"shape_type"`
	Geometry  interface{} `json:"geometry,omitempty"`
	Schedule  interface{} `json:"schedule,omitempty"`
	Version   int64       `json:"version"`
}

func (h *Hub) BroadcastRestrictedAreaUpdate(event string, areaID int64, areaName, areaType string, level int, shapeType string, geometry interface{}, schedule interface{}, version int64) {
	payload := RestrictedAreaSyncPayload{
		Event:     event,
		AreaID:    areaID,
		AreaName:  areaName,
		AreaType:  areaType,
		Level:     level,
		ShapeType: shapeType,
		Geometry:  geometry,
		Schedule:  schedule,
		Version:   version,
	}
	msg := &WSMessage{
		Type:      MsgRestrictedAreaUpdate,
		Timestamp: time.Now().Unix(),
		Data:      payload,
	}
	h.broadcast <- msg
	h.sendToAllVehicles(msg)
}

func (h *Hub) sendToAllVehicles(msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, _ := json.Marshal(msg)
	for vid, clients := range h.vehicleClients {
		_ = vid
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		if c.Conn != nil {
			c.Conn.Close()
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if c.Conn == nil {
				return
			}
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			if c.Conn == nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		if c.Conn != nil {
			c.Conn.Close()
		}
	}()

	for {
		if c.Conn == nil {
			return
		}
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}
		c.handleMessage(hub, &wsMsg)
	}
}

func (c *Client) handleMessage(hub *Hub, msg *WSMessage) {
	switch msg.Type {
	case MsgSubscribe:
		dataMap, ok := msg.Data.(map[string]interface{})
		if !ok {
			return
		}
		if orgID, ok := dataMap["org_id"].(float64); ok {
			hub.mu.Lock()
			hub.monitorClients[int64(orgID)] = append(hub.monitorClients[int64(orgID)], c)
			hub.mu.Unlock()
		}
		if vid, ok := dataMap["vehicle_id"].(float64); ok {
			hub.mu.Lock()
			hub.vehicleClients[int64(vid)] = append(hub.vehicleClients[int64(vid)], c)
			hub.mu.Unlock()
		}
	case MsgHeartbeat:
		resp := &WSMessage{
			Type:      MsgHeartbeat,
			Timestamp: time.Now().Unix(),
			Data: map[string]interface{}{
				"pong": true,
			},
		}
		data, _ := json.Marshal(resp)
		c.mu.Lock()
		select {
		case c.Send <- data:
		default:
		}
		c.mu.Unlock()
	}
}

func MonitorWebSocket(ctx context.Context, c *app.RequestContext) {
	hub := GetHub()
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	orgID, _ := c.Get("org_id")

	clientID := "monitor-" + generateClientID()

	err := upgrader.Upgrade(c, func(conn *hzws.Conn) {
		client := &Client{
			ID:         clientID,
			UserID:     toInt64(userID),
			Role:       toString(role),
			OrgID:      toInt64(orgID),
			Conn:       conn,
			Send:       make(chan []byte, 256),
			Subscribed: make(map[string]bool),
		}
		hub.register <- client

		if client.OrgID > 0 {
			hub.mu.Lock()
			hub.monitorClients[client.OrgID] = append(hub.monitorClients[client.OrgID], client)
			hub.mu.Unlock()
		}

		go client.writePump()
		client.readPump(hub)
	})
	if err != nil {
		logger.Sugar.Errorf("ws upgrade error: %v", err)
	}
}

func VehicleWebSocket(ctx context.Context, c *app.RequestContext) {
	hub := GetHub()
	vehicleIDStr := c.Param("vehicle_id")
	vehicleID := parseID(vehicleIDStr)

	clientID := "vehicle-" + vehicleIDStr + "-" + generateClientID()

	err := upgrader.Upgrade(c, func(conn *hzws.Conn) {
		client := &Client{
			ID:         clientID,
			UserID:     vehicleID,
			Role:       "vehicle_device",
			OrgID:      1,
			Conn:       conn,
			Send:       make(chan []byte, 256),
			Subscribed: make(map[string]bool),
		}
		hub.register <- client

		if vehicleID > 0 {
			hub.mu.Lock()
			hub.vehicleClients[vehicleID] = append(hub.vehicleClients[vehicleID], client)
			hub.mu.Unlock()
		}

		go client.writePump()
		client.readPump(hub)
	})
	if err != nil {
		logger.Sugar.Errorf("vehicle ws upgrade error: %v", err)
	}
}

func generateClientID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func toInt64(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	default:
		return 0
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func parseID(s string) int64 {
	var id int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int64(c-'0')
		}
	}
	return id
}

func (h *Hub) BroadcastReplanSuggestion(vehicleID, driverID int64, payload map[string]interface{}) {
	msg := &WSMessage{
		Type:      MsgRouteReplanSuggest,
		Timestamp: time.Now().Unix(),
		Data:      payload,
	}
	data, _ := json.Marshal(msg)

	if vehicleID > 0 {
		h.sendToVehicle(vehicleID, msg)
	}
	h.mu.RLock()
	for _, clients := range h.monitorClients {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) NotifyRouteApplied(vehicleID, driverID int64, payload map[string]interface{}) {
	msg := &WSMessage{
		Type:      MsgRouteApplied,
		Timestamp: time.Now().Unix(),
		Data:      payload,
	}
	if vehicleID > 0 {
		h.sendToVehicle(vehicleID, msg)
	}
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	for _, clients := range h.monitorClients {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastTrafficEvent(evt *model.TrafficEvent) {
	msg := &WSMessage{
		Type:      MsgTrafficEvent,
		Timestamp: time.Now().Unix(),
		Data:      evt,
	}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
	h.mu.RLock()
	for _, clients := range h.monitorClients {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastSOS(ctx context.Context, alert interface{}) {
	msg := &WSMessage{
		Type:      MsgSOSAlert,
		Timestamp: time.Now().Unix(),
		Data:      alert,
	}
	data, _ := json.Marshal(msg)
	h.broadcast <- msg

	h.mu.RLock()
	for _, clients := range h.monitorClients {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastEscortPolling(escortID int64, payload interface{}) {
	msg := &WSMessage{
		Type:      MsgEscortPolling,
		Timestamp: time.Now().Unix(),
		Data:      payload,
	}
	msgData, _ := json.Marshal(msg)

	h.mu.RLock()
	for _, clients := range h.monitorClients {
		for _, client := range clients {
			if client.UserID == escortID || client.Role == "admin" || client.Role == "dispatcher" {
				select {
				case client.Send <- msgData:
				default:
				}
			}
		}
	}
	h.mu.RUnlock()
}
