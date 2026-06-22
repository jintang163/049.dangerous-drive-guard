package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
	monitorWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
)

type EmergencyPlan struct {
	ID                  int64     `json:"id"`
	PlanNo              string    `json:"plan_no"`
	UNNumber            string    `json:"un_number"`
	ProperShippingName  string    `json:"proper_shipping_name"`
	ProperShippingNameEn string   `json:"proper_shipping_name_en"`
	DangerClass         string    `json:"danger_class"`
	SecondaryDanger     string    `json:"secondary_danger"`
	PackingGroup        string    `json:"packing_group"`
	HazardSummary       string    `json:"hazard_summary"`
	LeakDisposal        string    `json:"leak_disposal"`
	Neutralizer         string    `json:"neutralizer"`
	NeutralizerUsage    string    `json:"neutralizer_usage"`
	ProtectiveEquipment string    `json:"protective_equipment"`
	EvacuationDistance  string    `json:"evacuation_distance"`
	IsolationDistance   string    `json:"isolation_distance"`
	FireFighting        string    `json:"fire_fighting"`
	FirstAid            string    `json:"first_aid"`
	EnvironmentalProtection string `json:"environmental_protection"`
	SpecialPrecautions  string    `json:"special_precautions"`
	EmergencyContacts   string    `json:"emergency_contacts"`
	ReferenceStandard   string    `json:"reference_standard"`
	IsBuiltin           int       `json:"is_builtin"`
	Status              string    `json:"status"`
	CreatedBy           int64     `json:"created_by"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type EmergencyTaskCard struct {
	ID                    int64     `json:"id"`
	CardNo                string    `json:"card_no"`
	PlanID                int64     `json:"plan_id"`
	UNNumber              string    `json:"un_number"`
	DangerClass           string    `json:"danger_class"`
	VehicleID             int64     `json:"vehicle_id"`
	DriverID              int64     `json:"driver_id"`
	WaybillID             int64     `json:"waybill_id"`
	CardTitle             string    `json:"card_title"`
	LeakDisposalBrief     string    `json:"leak_disposal_brief"`
	NeutralizerBrief      string    `json:"neutralizer_brief"`
	ProtectiveEquipmentBrief string `json:"protective_equipment_brief"`
	EvacuationDistance    string    `json:"evacuation_distance"`
	FirstAidBrief         string    `json:"first_aid_brief"`
	SpecialNotes          string    `json:"special_notes"`
	PushChannel           string    `json:"push_channel"`
	PushStatus            string    `json:"push_status"`
	PushedAt              *time.Time `json:"pushed_at"`
	AcknowledgedAt        *time.Time `json:"acknowledged_at"`
	SourceType            string    `json:"source_type"`
	SourceID              int64     `json:"source_id"`
	Status                string    `json:"status"`
	ExpireAt              *time.Time `json:"expire_at"`
	CompletedAt           *time.Time `json:"completed_at"`
	CompletedBy           int64     `json:"completed_by"`
	Remark                string    `json:"remark"`
	CreatedBy             int64     `json:"created_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	VehiclePlate          string    `json:"vehicle_plate,omitempty"`
	DriverName            string    `json:"driver_name,omitempty"`
}

type EmergencyStats struct {
	TotalPlans      int `json:"total_plans"`
	BuiltinPlans    int `json:"builtin_plans"`
	CustomPlans     int `json:"custom_plans"`
	ActiveCards     int `json:"active_cards"`
	PushedCards     int `json:"pushed_cards"`
	AcknowledgedCards int `json:"acknowledged_cards"`
	ByDangerClass   []struct {
		Class string `json:"class"`
		Count int    `json:"count"`
	} `json:"by_danger_class"`
}

type TaskCardGenerateData struct {
	PlanID      int64  `json:"plan_id"`
	VehicleID   int64  `json:"vehicle_id"`
	DriverID    int64  `json:"driver_id"`
	WaybillID   int64  `json:"waybill_id"`
	SourceType  string `json:"source_type"`
	SourceID    int64  `json:"source_id"`
	PushChannel string `json:"push_channel"`
}

type EmergencyService struct {
	db *gorm.DB
}

func NewEmergencyService() *EmergencyService {
	return &EmergencyService{db: database.GetDB()}
}

func (s *EmergencyService) ListPlans(ctx context.Context, unNumber, dangerClass, keyword string, page, pageSize int) ([]*EmergencyPlan, int64, error) {
	var plans []*EmergencyPlan
	var total int64

	whereSQL := "WHERE 1=1"
	args := []interface{}{}
	if unNumber != "" {
		whereSQL += " AND p.un_number LIKE ?"
		args = append(args, unNumber+"%")
	}
	if dangerClass != "" {
		whereSQL += " AND p.danger_class = ?"
		args = append(args, dangerClass)
	}
	if keyword != "" {
		whereSQL += " AND p.proper_shipping_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	countSQL := "SELECT COUNT(*) FROM emergency_plans p " + whereSQL
	err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listSQL := `
		SELECT p.id, p.plan_no, p.un_number, p.proper_shipping_name, p.proper_shipping_name_en,
		       p.danger_class, p.secondary_danger, p.packing_group, p.hazard_summary,
		       p.leak_disposal, p.neutralizer, p.neutralizer_usage, p.protective_equipment,
		       p.evacuation_distance, p.isolation_distance, p.fire_fighting, p.first_aid,
		       p.environmental_protection, p.special_precautions, p.emergency_contacts,
		       p.reference_standard, p.is_builtin, p.status, p.created_by,
		       p.created_at, p.updated_at
		FROM emergency_plans p
		` + whereSQL + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var p EmergencyPlan
		rows.Scan(&p.ID, &p.PlanNo, &p.UNNumber, &p.ProperShippingName, &p.ProperShippingNameEn,
			&p.DangerClass, &p.SecondaryDanger, &p.PackingGroup, &p.HazardSummary,
			&p.LeakDisposal, &p.Neutralizer, &p.NeutralizerUsage, &p.ProtectiveEquipment,
			&p.EvacuationDistance, &p.IsolationDistance, &p.FireFighting, &p.FirstAid,
			&p.EnvironmentalProtection, &p.SpecialPrecautions, &p.EmergencyContacts,
			&p.ReferenceStandard, &p.IsBuiltin, &p.Status, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt)
		plans = append(plans, &p)
	}
	return plans, total, nil
}

func (s *EmergencyService) GetPlan(ctx context.Context, id int64) (*EmergencyPlan, error) {
	var p EmergencyPlan
	err := s.db.WithContext(ctx).Raw(`
		SELECT p.id, p.plan_no, p.un_number, p.proper_shipping_name, p.proper_shipping_name_en,
		       p.danger_class, p.secondary_danger, p.packing_group, p.hazard_summary,
		       p.leak_disposal, p.neutralizer, p.neutralizer_usage, p.protective_equipment,
		       p.evacuation_distance, p.isolation_distance, p.fire_fighting, p.first_aid,
		       p.environmental_protection, p.special_precautions, p.emergency_contacts,
		       p.reference_standard, p.is_builtin, p.status, p.created_by,
		       p.created_at, p.updated_at
		FROM emergency_plans p
		WHERE p.id = ?
	`, id).Scan(&p).Error
	if err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("plan not found")
	}
	return &p, nil
}

func (s *EmergencyService) GetPlanByUNNumber(ctx context.Context, unNumber string) (*EmergencyPlan, error) {
	var p EmergencyPlan
	err := s.db.WithContext(ctx).Raw(`
		SELECT p.id, p.plan_no, p.un_number, p.proper_shipping_name, p.proper_shipping_name_en,
		       p.danger_class, p.secondary_danger, p.packing_group, p.hazard_summary,
		       p.leak_disposal, p.neutralizer, p.neutralizer_usage, p.protective_equipment,
		       p.evacuation_distance, p.isolation_distance, p.fire_fighting, p.first_aid,
		       p.environmental_protection, p.special_precautions, p.emergency_contacts,
		       p.reference_standard, p.is_builtin, p.status, p.created_by,
		       p.created_at, p.updated_at
		FROM emergency_plans p
		WHERE p.un_number = ?
	`, unNumber).Scan(&p).Error
	if err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("plan not found")
	}
	return &p, nil
}

func (s *EmergencyService) CreatePlan(ctx context.Context, req *EmergencyPlan) (*EmergencyPlan, error) {
	if req.PlanNo == "" {
		req.PlanNo = fmt.Sprintf("EP%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	}
	if req.Status == "" {
		req.Status = "active"
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO emergency_plans (
			plan_no, un_number, proper_shipping_name, proper_shipping_name_en,
			danger_class, secondary_danger, packing_group, hazard_summary,
			leak_disposal, neutralizer, neutralizer_usage, protective_equipment,
			evacuation_distance, isolation_distance, fire_fighting, first_aid,
			environmental_protection, special_precautions, emergency_contacts,
			reference_standard, is_builtin, status, created_by,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, req.PlanNo, req.UNNumber, req.ProperShippingName, req.ProperShippingNameEn,
		req.DangerClass, req.SecondaryDanger, req.PackingGroup, req.HazardSummary,
		req.LeakDisposal, req.Neutralizer, req.NeutralizerUsage, req.ProtectiveEquipment,
		req.EvacuationDistance, req.IsolationDistance, req.FireFighting, req.FirstAid,
		req.EnvironmentalProtection, req.SpecialPrecautions, req.EmergencyContacts,
		req.ReferenceStandard, req.IsBuiltin, req.Status, req.CreatedBy)
	if result.Error != nil {
		return nil, result.Error
	}
	id, _ := result.LastInsertId()
	req.ID = id
	return req, nil
}

func (s *EmergencyService) UpdatePlan(ctx context.Context, req *EmergencyPlan) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE emergency_plans SET
			un_number=?, proper_shipping_name=?, proper_shipping_name_en=?,
			danger_class=?, secondary_danger=?, packing_group=?, hazard_summary=?,
			leak_disposal=?, neutralizer=?, neutralizer_usage=?, protective_equipment=?,
			evacuation_distance=?, isolation_distance=?, fire_fighting=?, first_aid=?,
			environmental_protection=?, special_precautions=?, emergency_contacts=?,
			reference_standard=?, status=?, updated_at=NOW()
		WHERE id=?
	`, req.UNNumber, req.ProperShippingName, req.ProperShippingNameEn,
		req.DangerClass, req.SecondaryDanger, req.PackingGroup, req.HazardSummary,
		req.LeakDisposal, req.Neutralizer, req.NeutralizerUsage, req.ProtectiveEquipment,
		req.EvacuationDistance, req.IsolationDistance, req.FireFighting, req.FirstAid,
		req.EnvironmentalProtection, req.SpecialPrecautions, req.EmergencyContacts,
		req.ReferenceStandard, req.Status, req.ID)
	return result.Error
}

func (s *EmergencyService) DeletePlan(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Exec("DELETE FROM emergency_plans WHERE id=?", id)
	return result.Error
}

func (s *EmergencyService) SearchByUNNumber(ctx context.Context, unNumber string) ([]*EmergencyPlan, error) {
	var plans []*EmergencyPlan
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT p.id, p.plan_no, p.un_number, p.proper_shipping_name, p.proper_shipping_name_en,
		       p.danger_class, p.secondary_danger, p.packing_group, p.hazard_summary,
		       p.leak_disposal, p.neutralizer, p.neutralizer_usage, p.protective_equipment,
		       p.evacuation_distance, p.isolation_distance, p.fire_fighting, p.first_aid,
		       p.environmental_protection, p.special_precautions, p.emergency_contacts,
		       p.reference_standard, p.is_builtin, p.status, p.created_by,
		       p.created_at, p.updated_at
		FROM emergency_plans p
		WHERE p.un_number = ? OR p.un_number LIKE ?
		ORDER BY p.un_number
	`, unNumber, unNumber+"%").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p EmergencyPlan
		rows.Scan(&p.ID, &p.PlanNo, &p.UNNumber, &p.ProperShippingName, &p.ProperShippingNameEn,
			&p.DangerClass, &p.SecondaryDanger, &p.PackingGroup, &p.HazardSummary,
			&p.LeakDisposal, &p.Neutralizer, &p.NeutralizerUsage, &p.ProtectiveEquipment,
			&p.EvacuationDistance, &p.IsolationDistance, &p.FireFighting, &p.FirstAid,
			&p.EnvironmentalProtection, &p.SpecialPrecautions, &p.EmergencyContacts,
			&p.ReferenceStandard, &p.IsBuiltin, &p.Status, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt)
		plans = append(plans, &p)
	}
	return plans, nil
}

func (s *EmergencyService) GenerateTaskCard(ctx context.Context, req *TaskCardGenerateData, operatorID int64) (*EmergencyTaskCard, error) {
	plan, err := s.GetPlan(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	var vehiclePlate string
	s.db.WithContext(ctx).Raw("SELECT plate_number FROM vehicles WHERE id = ?", req.VehicleID).Scan(&vehiclePlate)

	var driverName string
	s.db.WithContext(ctx).Raw("SELECT name FROM drivers WHERE id = ?", req.DriverID).Scan(&driverName)

	card := &EmergencyTaskCard{
		CardNo:                fmt.Sprintf("EC%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
		PlanID:                plan.ID,
		UNNumber:              plan.UNNumber,
		DangerClass:           plan.DangerClass,
		VehicleID:             req.VehicleID,
		DriverID:              req.DriverID,
		WaybillID:             req.WaybillID,
		CardTitle:             fmt.Sprintf("%s(%s)应急处置卡", plan.ProperShippingName, plan.UNNumber),
		LeakDisposalBrief:     truncateStr(plan.LeakDisposal, 200),
		NeutralizerBrief:      truncateStr(plan.Neutralizer, 200),
		ProtectiveEquipmentBrief: truncateStr(plan.ProtectiveEquipment, 200),
		EvacuationDistance:     plan.EvacuationDistance,
		FirstAidBrief:         truncateStr(plan.FirstAid, 200),
		SpecialNotes:          truncateStr(plan.SpecialPrecautions, 200),
		PushChannel:           req.PushChannel,
		PushStatus:            "pending",
		SourceType:            req.SourceType,
		SourceID:              req.SourceID,
		Status:                "active",
		CreatedBy:             operatorID,
		VehiclePlate:          vehiclePlate,
		DriverName:            driverName,
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO emergency_task_cards (
			card_no, plan_id, un_number, danger_class, vehicle_id, driver_id, waybill_id,
			card_title, leak_disposal_brief, neutralizer_brief, protective_equipment_brief,
			evacuation_distance, first_aid_brief, special_notes,
			push_channel, push_status, source_type, source_id,
			status, created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, card.CardNo, card.PlanID, card.UNNumber, card.DangerClass, card.VehicleID, card.DriverID, card.WaybillID,
		card.CardTitle, card.LeakDisposalBrief, card.NeutralizerBrief, card.ProtectiveEquipmentBrief,
		card.EvacuationDistance, card.FirstAidBrief, card.SpecialNotes,
		card.PushChannel, card.PushStatus, card.SourceType, card.SourceID,
		card.Status, card.CreatedBy)
	if result.Error != nil {
		return nil, result.Error
	}
	id, _ := result.LastInsertId()
	card.ID = id

	emergencySendMQ(card, "emergency_task_card_created")

	hub := monitorWs.GetHub()
	hub.BroadcastToMonitor(&monitorWs.WSMessage{
		Type:      "emergency_task_card",
		Timestamp: time.Now().Unix(),
		Data:      card,
	}, "admin", "dispatcher")

	return card, nil
}

func (s *EmergencyService) ListTaskCards(ctx context.Context, vehicleID, driverID int64, status, unNumber string, page, pageSize int) ([]*EmergencyTaskCard, int64, error) {
	var cards []*EmergencyTaskCard
	var total int64

	whereSQL := "WHERE 1=1"
	args := []interface{}{}
	if vehicleID > 0 {
		whereSQL += " AND t.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if driverID > 0 {
		whereSQL += " AND t.driver_id = ?"
		args = append(args, driverID)
	}
	if status != "" {
		whereSQL += " AND t.status = ?"
		args = append(args, status)
	}
	if unNumber != "" {
		whereSQL += " AND t.un_number LIKE ?"
		args = append(args, unNumber+"%")
	}

	countSQL := "SELECT COUNT(*) FROM emergency_task_cards t " + whereSQL
	err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listSQL := `
		SELECT t.id, t.card_no, t.plan_id, t.un_number, t.danger_class,
		       t.vehicle_id, t.driver_id, t.waybill_id, t.card_title,
		       t.leak_disposal_brief, t.neutralizer_brief, t.protective_equipment_brief,
		       t.evacuation_distance, t.first_aid_brief, t.special_notes,
		       t.push_channel, t.push_status, t.pushed_at, t.acknowledged_at,
		       t.source_type, t.source_id, t.status, t.expire_at,
		       t.completed_at, t.completed_by, t.remark,
		       t.created_by, t.created_at, t.updated_at,
		       v.plate_number AS vehicle_plate,
		       d.name AS driver_name
		FROM emergency_task_cards t
		LEFT JOIN vehicles v ON v.id = t.vehicle_id
		LEFT JOIN drivers d ON d.id = t.driver_id
		` + whereSQL + `
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var c EmergencyTaskCard
		rows.Scan(&c.ID, &c.CardNo, &c.PlanID, &c.UNNumber, &c.DangerClass,
			&c.VehicleID, &c.DriverID, &c.WaybillID, &c.CardTitle,
			&c.LeakDisposalBrief, &c.NeutralizerBrief, &c.ProtectiveEquipmentBrief,
			&c.EvacuationDistance, &c.FirstAidBrief, &c.SpecialNotes,
			&c.PushChannel, &c.PushStatus, &c.PushedAt, &c.AcknowledgedAt,
			&c.SourceType, &c.SourceID, &c.Status, &c.ExpireAt,
			&c.CompletedAt, &c.CompletedBy, &c.Remark,
			&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
			&c.VehiclePlate, &c.DriverName)
		cards = append(cards, &c)
	}
	return cards, total, nil
}

func (s *EmergencyService) GetTaskCard(ctx context.Context, id int64) (*EmergencyTaskCard, error) {
	var c EmergencyTaskCard
	err := s.db.WithContext(ctx).Raw(`
		SELECT t.id, t.card_no, t.plan_id, t.un_number, t.danger_class,
		       t.vehicle_id, t.driver_id, t.waybill_id, t.card_title,
		       t.leak_disposal_brief, t.neutralizer_brief, t.protective_equipment_brief,
		       t.evacuation_distance, t.first_aid_brief, t.special_notes,
		       t.push_channel, t.push_status, t.pushed_at, t.acknowledged_at,
		       t.source_type, t.source_id, t.status, t.expire_at,
		       t.completed_at, t.completed_by, t.remark,
		       t.created_by, t.created_at, t.updated_at,
		       v.plate_number AS vehicle_plate,
		       d.name AS driver_name
		FROM emergency_task_cards t
		LEFT JOIN vehicles v ON v.id = t.vehicle_id
		LEFT JOIN drivers d ON d.id = t.driver_id
		WHERE t.id = ?
	`, id).Scan(&c).Error
	if err != nil {
		return nil, err
	}
	if c.ID == 0 {
		return nil, fmt.Errorf("task card not found")
	}
	return &c, nil
}

func (s *EmergencyService) AckTaskCard(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE emergency_task_cards SET
			push_status='acknowledged', acknowledged_at=NOW(), updated_at=NOW()
		WHERE id=?
	`, id)
	return result.Error
}

func (s *EmergencyService) CompleteTaskCard(ctx context.Context, id int64, operatorID int64) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE emergency_task_cards SET
			status='completed', completed_at=NOW(), completed_by=?, updated_at=NOW()
		WHERE id=?
	`, operatorID, id)
	return result.Error
}

func (s *EmergencyService) CancelTaskCard(ctx context.Context, id int64, reason string) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE emergency_task_cards SET
			status='cancelled', remark=?, updated_at=NOW()
		WHERE id=?
	`, reason, id)
	return result.Error
}

func (s *EmergencyService) GetStats(ctx context.Context, orgID int64) (*EmergencyStats, error) {
	var stats EmergencyStats

	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_plans WHERE status='active'").Scan(&stats.TotalPlans)
	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_plans WHERE is_builtin=1 AND status='active'").Scan(&stats.BuiltinPlans)
	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_plans WHERE is_builtin=0 AND status='active'").Scan(&stats.CustomPlans)

	taskWhere := ""
	taskArgs := []interface{}{}
	if orgID > 0 {
		taskWhere = " AND v.org_id = ?"
		taskArgs = append(taskArgs, orgID)
	}
	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_task_cards t LEFT JOIN vehicles v ON v.id=t.vehicle_id WHERE t.status='active'"+taskWhere, taskArgs...).Scan(&stats.ActiveCards)
	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_task_cards t LEFT JOIN vehicles v ON v.id=t.vehicle_id WHERE t.push_status='pushed'"+taskWhere, taskArgs...).Scan(&stats.PushedCards)
	s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM emergency_task_cards t LEFT JOIN vehicles v ON v.id=t.vehicle_id WHERE t.push_status='acknowledged'"+taskWhere, taskArgs...).Scan(&stats.AcknowledgedCards)

	classRows, err := s.db.WithContext(ctx).Raw(`
		SELECT p.danger_class, COUNT(*) cnt
		FROM emergency_plans p
		WHERE p.status='active'
		GROUP BY p.danger_class
		ORDER BY p.danger_class
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer classRows.Close()
	for classRows.Next() {
		var cls string
		var cnt int
		classRows.Scan(&cls, &cnt)
		stats.ByDangerClass = append(stats.ByDangerClass, struct {
			Class string `json:"class"`
			Count int    `json:"count"`
		}{Class: cls, Count: cnt})
	}

	return &stats, nil
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func emergencySendMQ(data interface{}, topic string) {
	payload, _ := json.Marshal(data)
	_ = mq.Send(context.Background(), topic, string(payload))
	logger.GetLogger().Infof("[Emergency] MQ sent topic=%s payload=%s", topic, string(payload))
}
