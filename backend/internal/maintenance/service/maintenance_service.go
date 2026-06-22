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
)

type MaintenanceService struct {
	db *gorm.DB
}

type MaintenancePlan struct {
	ID              int64     `json:"id"`
	PlanNo          string    `json:"plan_no"`
	VehicleID       int64     `json:"vehicle_id"`
	PlanName        string    `json:"plan_name"`
	MaintenanceType string    `json:"maintenance_type"`
	TriggerMode     string    `json:"trigger_mode"`
	TriggerMileage  float64   `json:"trigger_mileage_km"`
	TriggerDays     int       `json:"trigger_days"`
	BaseMileage     float64   `json:"base_mileage_km"`
	BaseDate        string    `json:"base_date"`
	NextMileage     float64   `json:"next_mileage_km"`
	NextDate        string    `json:"next_date"`
	WarnBeforeKm    float64   `json:"warn_before_km"`
	WarnBeforeDays  int       `json:"warn_before_days"`
	Items           string    `json:"items"`
	EstimatedCost   float64   `json:"estimated_cost"`
	Priority        int       `json:"priority"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	LastWorkOrderID int64     `json:"last_work_order_id"`
	CreatedBy       int64     `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	VehiclePlate    string    `json:"vehicle_plate,omitempty"`
}

type WorkOrder struct {
	ID                 int64     `json:"id"`
	WorkOrderNo        string    `json:"work_order_no"`
	VehicleID          int64     `json:"vehicle_id"`
	PlanID             int64     `json:"plan_id"`
	MaintenanceType    string    `json:"maintenance_type"`
	SourceType         string    `json:"source_type"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	TriggerReason      string    `json:"trigger_reason"`
	TriggerDetail      string    `json:"trigger_detail"`
	VehicleMileage     float64   `json:"vehicle_mileage_km"`
	Items              string    `json:"items"`
	PartsUsed          string    `json:"parts_used"`
	EstimatedCost      float64   `json:"estimated_cost"`
	ActualCost         float64   `json:"actual_cost"`
	Workshop           string    `json:"workshop"`
	Mechanic           string    `json:"mechanic"`
	ContactPhone       string    `json:"contact_phone"`
	AppointmentTime    string    `json:"appointment_time"`
	CheckinTime        string    `json:"checkin_time"`
	CheckoutTime       string    `json:"checkout_time"`
	CurrentMileage     float64   `json:"current_mileage_km"`
	QualityCheckDone   int       `json:"quality_check_done"`
	QualityCheckNote   string    `json:"quality_check_note"`
	NextMileageSuggest float64   `json:"next_mileage_km"`
	NextDateSuggest    string    `json:"next_date"`
	Priority           int       `json:"priority"`
	Status             string    `json:"status"`
	AssignedTo         int64     `json:"assigned_to"`
	DispatcherID       int64     `json:"dispatcher_id"`
	DispatchedAt       string    `json:"dispatched_at"`
	CompletedBy        int64     `json:"completed_by"`
	CompletedAt        string    `json:"completed_at"`
	CancelledReason    string    `json:"cancelled_reason"`
	DriverConfirmBfr   int       `json:"driver_confirm_before"`
	DriverConfirmAft   int       `json:"driver_confirm_after"`
	PhotosBefore       string    `json:"photos_before"`
	PhotosAfter        string    `json:"photos_after"`
	Remark             string    `json:"remark"`
	CreatedBy          int64     `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	VehiclePlate       string    `json:"vehicle_plate,omitempty"`
	AssignedName       string    `json:"assigned_name,omitempty"`
}

type WorkOrderLog struct {
	ID           int64     `json:"id"`
	WorkOrderID  int64     `json:"work_order_id"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	ActionType   string    `json:"action_type"`
	ActionNote   string    `json:"action_note"`
	OperatorID   int64     `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	OperatorRole string    `json:"operator_role"`
	CreatedAt    time.Time `json:"created_at"`
}

type WorkOrderCompleteData struct {
	CheckoutTime       string
	ActualCost         float64
	QualityCheckDone   int
	QualityCheckNote   string
	NextMileageSuggest float64
	NextDateSuggest    string
	PartsUsed          string
	Items              string
	PhotosBefore       string
	PhotosAfter        string
	Remark             string
}

type MaintenanceStats struct {
	ActivePlans         int `json:"active_plans"`
	PendingOrders       int `json:"pending_orders"`
	ProcessingOrders    int `json:"processing_orders"`
	CompletedOrders     int `json:"completed_orders"`
	ThisMonthCompleted  int `json:"this_month_completed"`
	ThisMonthCost       float64 `json:"this_month_cost"`
	UrgentPending       int `json:"urgent_pending"`
	OverduePlans        int `json:"overdue_plans"`
	ByType              []struct {
		Type  string  `json:"type"`
		Count int     `json:"count"`
		Cost  float64 `json:"cost"`
	} `json:"by_type"`
	Trend []struct {
		Date      string `json:"date"`
		Created   int    `json:"created"`
		Completed int    `json:"completed"`
	} `json:"trend"`
}

type UpcomingItem struct {
	VehicleID       int64   `json:"vehicle_id"`
	VehiclePlate  string  `json:"vehicle_plate"`
	PlanID        int64   `json:"plan_id"`
	PlanName      string  `json:"plan_name"`
	MaintenanceType string `json:"maintenance_type"`
	NextMileage     float64 `json:"next_mileage_km"`
	CurrentMileage  float64 `json:"current_mileage_km"`
	MileageLeft   float64 `json:"mileage_left_km"`
	NextDate        string  `json:"next_date"`
	DaysLeft       int     `json:"days_left"`
	Priority        int     `json:"priority"`
	Urgent         int     `json:"is_urgent"`
	WarnLevel      string  `json:"warn_level"`
}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{db: database.GetDB()}
}

func (s *MaintenanceService) ListPlans(ctx context.Context, vehicleID int64, status, mType string, page, pageSize int) ([]*MaintenancePlan, int64, error) {
	var plans []*MaintenancePlan
	var total int64

	whereSQL := "WHERE 1=1"
	args := []interface{}{}
	if vehicleID > 0 {
		whereSQL += " AND p.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if status != "" {
		whereSQL += " AND p.status = ?"
		args = append(args, status)
	}
	if mType != "" {
		whereSQL += " AND p.maintenance_type = ?"
		args = append(args, mType)
	}

	countSQL := "SELECT COUNT(*) FROM maintenance_plans p " + whereSQL
	err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listSQL := `
		SELECT p.id, p.plan_no, p.vehicle_id, p.plan_name, p.maintenance_type,
		       p.trigger_mode, p.trigger_mileage_km, p.trigger_days,
		       p.base_mileage_km, p.base_date, p.next_mileage_km, p.next_date,
		       p.warn_before_km, p.warn_before_days, p.items, p.estimated_cost,
		       p.priority, p.description, p.status, p.last_work_order_id,
		       p.created_by, p.created_at, p.updated_at,
		       v.plate_number AS vehicle_plate
		FROM maintenance_plans p
		LEFT JOIN vehicles v ON v.id = p.vehicle_id
		` + whereSQL + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var p MaintenancePlan
		rows.Scan(&p.ID, &p.PlanNo, &p.VehicleID, &p.PlanName, &p.MaintenanceType,
			&p.TriggerMode, &p.TriggerMileage, &p.TriggerDays,
			&p.BaseMileage, &p.BaseDate, &p.NextMileage, &p.NextDate,
			&p.WarnBeforeKm, &p.WarnBeforeDays, &p.Items, &p.EstimatedCost,
			&p.Priority, &p.Description, &p.Status, &p.LastWorkOrderID,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.VehiclePlate)
		plans = append(plans, &p)
	}
	return plans, total, nil
}

func (s *MaintenanceService) GetPlan(ctx context.Context, id int64) (*MaintenancePlan, error) {
	var p MaintenancePlan
	err := s.db.WithContext(ctx).Raw(`
		SELECT p.id, p.plan_no, p.vehicle_id, p.plan_name, p.maintenance_type,
		       p.trigger_mode, p.trigger_mileage_km, p.trigger_days,
		       p.base_mileage_km, p.base_date, p.next_mileage_km, p.next_date,
		       p.warn_before_km, p.warn_before_days, p.items, p.estimated_cost,
		       p.priority, p.description, p.status, p.last_work_order_id,
		       p.created_by, p.created_at, p.updated_at,
		       v.plate_number AS vehicle_plate
		FROM maintenance_plans p
		LEFT JOIN vehicles v ON v.id = p.vehicle_id
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

func (s *MaintenanceService) CreatePlan(ctx context.Context, req *MaintenancePlan) (*MaintenancePlan, error) {
	if req.PlanNo == "" {
		req.PlanNo = fmt.Sprintf("MP%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.TriggerMode == "" {
		req.TriggerMode = "both"
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO maintenance_plans (
			plan_no, vehicle_id, plan_name, maintenance_type, trigger_mode,
			trigger_mileage_km, trigger_days, base_mileage_km, base_date,
			next_mileage_km, next_date, warn_before_km, warn_before_days,
			items, estimated_cost, priority, description, status, created_by,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, req.PlanNo, req.VehicleID, req.PlanName, req.MaintenanceType, req.TriggerMode,
		req.TriggerMileage, req.TriggerDays, req.BaseMileage, req.BaseDate,
		req.NextMileage, req.NextDate, req.WarnBeforeKm, req.WarnBeforeDays,
		req.Items, req.EstimatedCost, req.Priority, req.Description, req.Status, req.CreatedBy)
	if result.Error != nil {
		return nil, result.Error
	}
	var id, _ := result.LastInsertId()
	req.ID = id
	return req, nil
}

func (s *MaintenanceService) UpdatePlan(ctx context.Context, req *MaintenancePlan) (*MaintenancePlan, error) {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_plans SET
			plan_name=?, maintenance_type=?, trigger_mode=?, trigger_mileage_km=?,
			trigger_days=?, base_mileage_km=?, base_date=?, next_mileage_km=?,
			next_date=?, warn_before_km=?, warn_before_days=?, items=?,
			estimated_cost=?, priority=?, description=?, status=?, updated_at=NOW()
		WHERE id=?
	`, req.PlanName, req.MaintenanceType, req.TriggerMode, req.TriggerMileage,
		req.TriggerDays, req.BaseMileage, req.BaseDate, req.NextMileage,
		req.NextDate, req.WarnBeforeKm, req.WarnBeforeDays, req.Items,
		req.EstimatedCost, req.Priority, req.Description, req.Status, req.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return req, nil
}

func (s *MaintenanceService) DeletePlan(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Exec("DELETE FROM maintenance_plans WHERE id=?", id)
	return result.Error
}

func (s *MaintenanceService) SetPlanStatus(ctx context.Context, id int64, status string) error {
	result := s.db.WithContext(ctx).Exec("UPDATE maintenance_plans SET status=?, updated_at=NOW() WHERE id=?", status, id)
	return result.Error
}

func (s *MaintenanceService) CheckAndGenerateWorkOrder(ctx context.Context, planID int64, operatorID int64) (map[string]interface{}, error) {
	plan, err := s.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan.Status != "active" {
		return nil, fmt.Errorf("plan is not active")
	}

	var currentMileage float64
	s.db.WithContext(ctx).Raw("SELECT current_mileage_km FROM vehicles WHERE id=?", plan.VehicleID).Scan(&currentMileage)

	now := time.Now()
	nextDate, _ := time.Parse("2006-01-02", plan.NextDate)
	if plan.NextDate == "" {
		nextDate = now.AddDate(0, 0, 1)
	}

	mileageDiff := plan.NextMileage - currentMileage
	daysDiff := int(nextDate.Sub(now).Hours() / 24)

	needGenerate := false
	var triggerReason := ""
	var triggerDetail := ""

	if plan.TriggerMode == "mileage" || plan.TriggerMode == "both" {
		if mileageDiff <= 0 || currentMileage >= plan.NextMileage-plan.WarnBeforeKm {
			needGenerate = true
			triggerReason = "mileage"
			triggerDetail = fmt.Sprintf("里程触发: 当前%.1fkm, 阈值%.1fkm, 剩余%.1fkm", currentMileage, plan.NextMileage, mileageDiff)
		}
	}
	if plan.TriggerMode == "time" || plan.TriggerMode == "both" {
		if daysDiff <= 0 || daysDiff <= plan.WarnBeforeDays {
			needGenerate = true
			if triggerReason != "" {
				triggerReason += "+time"
			} else {
				triggerReason = "time"
			}
			triggerDetail += fmt.Sprintf(" | 时间触发: 剩余%d天", daysDiff)
		}
	}

	if !needGenerate {
		return map[string]interface{}{
			"generated": false,
			"reason":    "not triggered yet",
			"mileage_left": mileageDiff,
			"days_left":   daysDiff,
		}, nil
	}

	order := &WorkOrder{
		WorkOrderNo:   fmt.Sprintf("WO%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
		VehicleID:     plan.VehicleID,
		PlanID:        plan.ID,
		MaintenanceType: plan.MaintenanceType,
		SourceType:    "auto_trigger",
		Title:         fmt.Sprintf("%s-%s", plan.PlanName, plan.MaintenanceType),
		Description:   plan.Description,
		TriggerReason: triggerReason,
		TriggerDetail: triggerDetail,
		VehicleMileage:  currentMileage,
		Items:         plan.Items,
		EstimatedCost:  plan.EstimatedCost,
		Priority:        plan.Priority,
		Status:          "pending",
		CreatedBy:     operatorID,
	}

	created, err := s.CreateWorkOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	s.db.WithContext(ctx).Exec("UPDATE maintenance_plans SET last_work_order_id=?, updated_at=NOW() WHERE id=?", created.ID, planID)

	return map[string]interface{}{
		"generated":      true,
		"work_order_id": created.ID,
		"work_order_no": created.WorkOrderNo,
		"trigger_reason": triggerReason,
		"trigger_detail": triggerDetail,
	}, nil
}

func (s *MaintenanceService) BatchCheckVehicles(ctx context.Context, orgID int64, operatorID int64) (map[string]interface{}, error) {
	whereClause := "WHERE p.status='active'"
	args := []interface{}{}
	if orgID > 0 {
		whereClause += " AND v.org_id = ?"
		args = append(args, orgID)
	}

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT p.id FROM maintenance_plans p
		LEFT JOIN vehicles v ON v.id = p.vehicle_id
		`+whereClause, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planIDs []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		planIDs = append(planIDs, id)
	}

	generated := 0
	skipped := 0
	failed := 0
	var failedReasons := []string{}

	for _, pid := range planIDs {
		result, err := s.CheckAndGenerateWorkOrder(ctx, pid, operatorID)
		if err != nil {
			failed++
			failedReasons = append(failedReasons, fmt.Sprintf("plan%d: %s", pid, err.Error()))
			continue
		}
		if g, ok := result["generated"].(bool); ok && g {
			generated++
		} else {
			skipped++
		}
	}

	return map[string]interface{}{
		"total":      len(planIDs),
		"generated":  generated,
		"skipped":  skipped,
		"failed":   failed,
		"failed_reasons": failedReasons,
	}, nil
}

func (s *MaintenanceService) ListWorkOrders(ctx context.Context, vehicleID, planID int64, status, mType, source string, page, pageSize int) ([]*WorkOrder, int64, error) {
	var orders []*WorkOrder
	var total int64

	whereSQL := "WHERE 1=1"
	args := []interface{}{}
	if vehicleID > 0 {
		whereSQL += " AND w.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if planID > 0 {
		whereSQL += " AND w.plan_id = ?"
		args = append(args, planID)
	}
	if status != "" {
		whereSQL += " AND w.status = ?"
		args = append(args, status)
	}
	if mType != "" {
		whereSQL += " AND w.maintenance_type = ?"
		args = append(args, mType)
	}
	if source != "" {
		whereSQL += " AND w.source_type = ?"
		args = append(args, source)
	}

	countSQL := "SELECT COUNT(*) FROM maintenance_work_orders w " + whereSQL
	err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	listSQL := `
		SELECT w.id, w.work_order_no, w.vehicle_id, w.plan_id, w.maintenance_type,
		       w.source_type, w.title, w.description, w.trigger_reason, w.trigger_detail,
		       w.vehicle_mileage_km, w.items, w.parts_used, w.estimated_cost, w.actual_cost,
		       w.workshop, w.mechanic, w.contact_phone, w.appointment_time, w.checkin_time,
		       w.checkout_time, w.current_mileage_km, w.quality_check_done,
		       w.quality_check_note, w.next_mileage_km, w.next_date,
		       w.priority, w.status, w.assigned_to, w.dispatcher_id, w.dispatched_at,
		       w.completed_by, w.completed_at, w.cancelled_reason,
		       w.driver_confirm_before, w.driver_confirm_after,
		       w.photos_before, w.photos_after, w.remark, w.created_by,
		       w.created_at, w.updated_at,
		       v.plate_number AS vehicle_plate
		FROM maintenance_work_orders w
		LEFT JOIN vehicles v ON v.id = w.vehicle_id
		` + whereSQL + `
		ORDER BY w.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var w WorkOrder
		rows.Scan(&w.ID, &w.WorkOrderNo, &w.VehicleID, &w.PlanID, &w.MaintenanceType,
			&w.SourceType, &w.Title, &w.Description, &w.TriggerReason, &w.TriggerDetail,
			&w.VehicleMileage, &w.Items, &w.PartsUsed, &w.EstimatedCost, &w.ActualCost,
			&w.Workshop, &w.Mechanic, &w.ContactPhone, &w.AppointmentTime, &w.CheckinTime,
			&w.CheckoutTime, &w.CurrentMileage, &w.QualityCheckDone,
			&w.QualityCheckNote, &w.NextMileageSuggest, &w.NextDateSuggest,
			&w.Priority, &w.Status, &w.AssignedTo, &w.DispatcherID, &w.DispatchedAt,
			&w.CompletedBy, &w.CompletedAt, &w.CancelledReason,
			&w.DriverConfirmBfr, &w.DriverConfirmAft,
			&w.PhotosBefore, &w.PhotosAfter, &w.Remark, &w.CreatedBy,
			&w.CreatedAt, &w.UpdatedAt,
			&w.VehiclePlate)
		orders = append(orders, &w)
	}
	return orders, total, nil
}

func (s *MaintenanceService) GetWorkOrder(ctx context.Context, id int64) (*WorkOrder, error) {
	var w WorkOrder
	err := s.db.WithContext(ctx).Raw(`
		SELECT w.id, w.work_order_no, w.vehicle_id, w.plan_id, w.maintenance_type,
		       w.source_type, w.title, w.description, w.trigger_reason, w.trigger_detail,
		       w.vehicle_mileage_km, w.items, w.parts_used, w.estimated_cost, w.actual_cost,
		       w.workshop, w.mechanic, w.contact_phone, w.appointment_time, w.checkin_time,
		       w.checkout_time, w.current_mileage_km, w.quality_check_done,
		       w.quality_check_note, w.next_mileage_km, w.next_date,
		       w.priority, w.status, w.assigned_to, w.dispatcher_id, w.dispatched_at,
		       w.completed_by, w.completed_at, w.cancelled_reason,
		       w.driver_confirm_before, w.driver_confirm_after,
		       w.photos_before, w.photos_after, w.remark, w.created_by,
		       w.created_at, w.updated_at,
		       v.plate_number AS vehicle_plate
		FROM maintenance_work_orders w
		LEFT JOIN vehicles v ON v.id = w.vehicle_id
		WHERE w.id = ?
	`, id).Scan(&w).Error
	if err != nil {
		return nil, err
	}
	if w.ID == 0 {
		return nil, fmt.Errorf("work order not found")
	}
	return &w, nil
}

func (s *MaintenanceService) CreateWorkOrder(ctx context.Context, req *WorkOrder) (*WorkOrder, error) {
	if req.WorkOrderNo == "" {
		req.WorkOrderNo = fmt.Sprintf("WO%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	}
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.SourceType == "" {
		req.SourceType = "manual"
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO maintenance_work_orders (
			work_order_no, vehicle_id, plan_id, maintenance_type, source_type,
			title, description, trigger_reason, trigger_detail, vehicle_mileage_km,
			items, parts_used, estimated_cost, actual_cost, workshop, mechanic,
			contact_phone, appointment_time, checkin_time, checkout_time,
			current_mileage_km, quality_check_done, quality_check_note,
			next_mileage_km, next_date, priority, status,
			assigned_to, dispatcher_id, dispatched_at, completed_by, completed_at,
			cancelled_reason, driver_confirm_before, driver_confirm_after,
			photos_before, photos_after, remark, created_by,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, req.WorkOrderNo, req.VehicleID, req.PlanID, req.MaintenanceType, req.SourceType,
		req.Title, req.Description, req.TriggerReason, req.TriggerDetail, req.VehicleMileage,
		req.Items, req.PartsUsed, req.EstimatedCost, req.ActualCost, req.Workshop,
		req.Mechanic, req.ContactPhone, req.AppointmentTime, req.CheckinTime,
		req.CheckoutTime, req.CurrentMileage, req.QualityCheckDone, req.QualityCheckNote,
		req.NextMileageSuggest, req.NextDateSuggest, req.Priority, req.Status,
		req.AssignedTo, req.DispatcherID, req.DispatchedAt, req.CompletedBy,
		req.CompletedAt, req.CancelledReason, req.DriverConfirmBfr, req.DriverConfirmAft,
		req.PhotosBefore, req.PhotosAfter, req.Remark, req.CreatedBy)
	if result.Error != nil {
		return nil, result.Error
	}
	id, _ := result.LastInsertId()
	req.ID = id

	s.recordWorkOrderLog(ctx, id, "", req.Status, "create", "创建工单", req.CreatedBy, "")

	sendMQ(req, "maintenance_work_order_created")
	return req, nil
}

func (s *MaintenanceService) UpdateWorkOrder(ctx context.Context, req *WorkOrder) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			title=?, description=?, maintenance_type=?, items=?, parts_used=?,
			estimated_cost=?, actual_cost=?, workshop=?, mechanic=?,
			contact_phone=?, appointment_time=?, priority=?, remark=?, updated_at=NOW()
		WHERE id=?
	`, req.Title, req.Description, req.MaintenanceType, req.Items, req.PartsUsed,
		req.EstimatedCost, req.ActualCost, req.Workshop, req.Mechanic,
		req.ContactPhone, req.AppointmentTime, req.Priority, req.Remark, req.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	if old.Status != req.Status && req.Status != "" {
		s.recordWorkOrderLog(ctx, req.ID, old.Status, req.Status, "update", "更新工单", req.CreatedBy, "")
	}
	return req, nil
}

func (s *MaintenanceService) AssignWorkOrder(ctx context.Context, id, assignedTo, dispatcherID int64, workshop, contactPhone, remark string) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			status='assigned', assigned_to=?, dispatcher_id=?, workshop=?,
			contact_phone=?, remark=?, dispatched_at=?, updated_at=NOW()
		WHERE id=?
	`, assignedTo, dispatcherID, workshop, contactPhone, remark, now, id)
	if result.Error != nil {
		return nil, result.Error
	}
	remarkLog := remark
	if workshop != "" {
		remarkLog += " 车间:" + workshop
	}
	s.recordWorkOrderLog(ctx, id, old.Status, "assigned", "assign", remarkLog, dispatcherID, "")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) SetAppointment(ctx context.Context, id, appointmentTime, workshop, contactPhone, remark string) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	newStatus := old.Status
	if newStatus == "pending" {
		newStatus = "scheduled"
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			appointment_time=?, workshop=?, contact_phone=?, remark=?, status=?, updated_at=NOW()
		WHERE id=?
	`, appointmentTime, workshop, contactPhone, remark, newStatus, id)
	if result.Error != nil {
		return nil, result.Error
	}
	remarkLog := fmt.Sprintf("预约时间:%s", appointmentTime)
	s.recordWorkOrderLog(ctx, id, old.Status, newStatus, "appointment", remarkLog, 0, "")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) CheckinWorkOrder(ctx context.Context, id, checkinTime string, currentMileage float64, mechanic, remark string) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if checkinTime == "" {
		checkinTime = time.Now().Format("2006-01-02 15:04:05")
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			status='in_progress', checkin_time=?, current_mileage_km=?, mechanic=?, remark=?, updated_at=NOW()
		WHERE id=?
	`, checkinTime, currentMileage, mechanic, remark, id)
	if result.Error != nil {
		return nil, result.Error
	}
	remarkLog := fmt.Sprintf("到店,里程:%.1fkm, 技师:%s", currentMileage, mechanic)
	s.recordWorkOrderLog(ctx, id, old.Status, "in_progress", "checkin", remarkLog, 0, "")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) StartWork(ctx context.Context, id int64, items, partsUsed, mechanic, remark string) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	newStatus := "in_progress"
	if old.Status == "pending" || old.Status == "scheduled" || old.Status == "assigned" {
		newStatus = "in_progress"
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			status=?, items=?, parts_used=?, mechanic=?, remark=?, updated_at=NOW()
		WHERE id=?
	`, newStatus, items, partsUsed, mechanic, remark, id)
	if result.Error != nil {
		return nil, result.Error
	}
	remarkLog := "开始施工"
	if mechanic != "" {
		remarkLog += ", 技师:" + mechanic
	}
	s.recordWorkOrderLog(ctx, id, old.Status, newStatus, "start", remarkLog, 0, "")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) CompleteWork(ctx context.Context, id, completedBy int64, data *WorkOrderCompleteData) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if data.CheckoutTime == "" {
		data.CheckoutTime = time.Now().Format("2006-01-02 15:04:05")
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			status='completed', checkout_time=?, actual_cost=?, quality_check_done=?,
			quality_check_note=?, next_mileage_km=?, next_date=?,
			parts_used=?, items=?, photos_before=?, photos_after=?,
			completed_by=?, completed_at=NOW(), remark=?, updated_at=NOW()
		WHERE id=?
	`, data.CheckoutTime, data.ActualCost, data.QualityCheckDone,
		data.QualityCheckNote, data.NextMileageSuggest, data.NextDateSuggest,
		data.PartsUsed, data.Items, data.PhotosBefore, data.PhotosAfter,
		completedBy, data.Remark, id)
	if result.Error != nil {
		return nil, result.Error
	}
	remarkLog := fmt.Sprintf("完工, 实际费用:%.2f", data.ActualCost)
	s.recordWorkOrderLog(ctx, id, old.Status, "completed", "complete", remarkLog, completedBy, "")

	if old.PlanID > 0 {
		var plan MaintenancePlan
		s.db.WithContext(ctx).Raw("SELECT * FROM maintenance_plans WHERE id=?", old.PlanID).Scan(&plan)
		if plan.ID > 0 {
			nextMileage := plan.BaseMileage + plan.TriggerMileage
			baseDate, _ := time.Parse("2006-01-02", plan.BaseDate)
			if baseDate.IsZero() {
				baseDate = time.Now()
			}
			nextDate := baseDate.AddDate(0, 0, plan.TriggerDays).Format("2006-01-02")
			if data.NextMileageSuggest > 0 {
				nextMileage = data.NextMileageSuggest
			}
			if data.NextDateSuggest != "" {
				nextDate = data.NextDateSuggest
			}
			s.db.WithContext(ctx).Exec(`
				UPDATE maintenance_plans SET
					base_mileage_km=?, base_date=?, next_mileage_km=?, next_date=?,
					last_work_order_id=?, updated_at=NOW()
				WHERE id=?
			`, old.CurrentMileage, time.Now().Format("2006-01-02"), nextMileage, nextDate, id, old.PlanID)
		}
	}

	sendMQ(old, "maintenance_work_order_completed")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) CancelWorkOrder(ctx context.Context, id, operatorID int64, reason string) (*WorkOrder, error) {
	old, err := s.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE maintenance_work_orders SET
			status='cancelled', cancelled_reason=?, updated_at=NOW()
		WHERE id=?
	`, reason, id)
	if result.Error != nil {
		return nil, result.Error
	}
	s.recordWorkOrderLog(ctx, id, old.Status, "cancelled", "cancel", reason, operatorID, "")
	return s.GetWorkOrder(ctx, id)
}

func (s *MaintenanceService) GetWorkOrderLogs(ctx context.Context, workOrderID int64) ([]*WorkOrderLog, error) {
	var logs []*WorkOrderLog
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT l.id, l.work_order_id, l.old_status, l.new_status,
		       l.action_type, l.action_note, l.operator_id,
		       l.operator_name, l.operator_role, l.created_at
		FROM maintenance_work_order_logs l
		WHERE l.work_order_id = ?
		ORDER BY l.created_at ASC
	`, workOrderID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var l WorkOrderLog
		rows.Scan(&l.ID, &l.WorkOrderID, &l.OldStatus, &l.NewStatus,
			&l.ActionType, &l.ActionNote, &l.OperatorID,
			&l.OperatorName, &l.OperatorRole, &l.CreatedAt)
		logs = append(logs, &l)
	}
	return logs, nil
}

func (s *MaintenanceService) GetStats(ctx context.Context, orgID int64) (*MaintenanceStats, error) {
	var stats MaintenanceStats
	whereOrg := ""
	args := []interface{}{}
	if orgID > 0 {
		whereOrg = " AND v.org_id = ?"
		args = append(args, orgID)
	}

	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_plans p LEFT JOIN vehicles v ON v.id=p.vehicle_id WHERE p.status='active'`+whereOrg, args...).Scan(&stats.ActivePlans)
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status IN ('pending','scheduled')`+whereOrg, args...).Scan(&stats.PendingOrders)
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status IN ('assigned','in_progress')`+whereOrg, args...).Scan(&stats.ProcessingOrders)
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status='completed'`+whereOrg, args...).Scan(&stats.CompletedOrders)

	thisMonth := time.Now().Format("2006-01")
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status='completed' AND DATE_FORMAT(w.completed_at,'%Y-%m')=?`+whereOrg, append([]interface{}{thisMonth}, args...)...).Scan(&stats.ThisMonthCompleted)
	s.db.WithContext(ctx).Raw(`SELECT IFNULL(SUM(w.actual_cost),0) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status='completed' AND DATE_FORMAT(w.completed_at,'%Y-%m')=?`+whereOrg, append([]interface{}{thisMonth}, args...)...).Scan(&stats.ThisMonthCost)
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id WHERE w.status IN ('pending','scheduled') AND w.priority>=3`+whereOrg, args...).Scan(&stats.UrgentPending)

	now := time.Now()
	s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM maintenance_plans p LEFT JOIN vehicles v ON v.id=p.vehicle_id WHERE p.status='active' AND DATEDIFF(p.next_date, ?)<=0`+whereOrg, append([]interface{}{now.Format("2006-01-02")}, args...)...).Scan(&stats.OverduePlans)

	typeRows, _ := s.db.WithContext(ctx).Raw(`
		SELECT w.maintenance_type, COUNT(*) cnt, IFNULL(SUM(w.actual_cost),0) cost
		FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id
		WHERE 1=1`+whereOrg+` GROUP BY w.maintenance_type
	`, args...).Rows()
	if typeRows != nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var t string
			var cnt int
			var cost float64
			typeRows.Scan(&t, &cnt, &cost)
			stats.ByType = append(stats.ByType, struct {
				Type  string  `json:"type"`
				Count int     `json:"count"`
				Cost  float64 `json:"cost"`
			}{Type: t, Count: cnt, Cost: cost})
		}
	}

	trendRows, _ := s.db.WithContext(ctx).Raw(`
		SELECT DATE_FORMAT(created_at,'%Y-%m-%d') dt,
		       SUM(CASE WHEN status!='deleted' THEN 1 ELSE 0 END) created_cnt,
		       SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) completed_cnt
		FROM maintenance_work_orders w LEFT JOIN vehicles v ON v.id=w.vehicle_id
		WHERE created_at >= DATE_SUB(NOW(),INTERVAL 7 DAY)`+whereOrg+`
		GROUP BY DATE_FORMAT(created_at,'%Y-%m-%d')
		ORDER BY dt ASC
	`, args...).Rows()
	if trendRows != nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var dt string
			var cc, ct int
			trendRows.Scan(&dt, &cc, &ct)
			stats.Trend = append(stats.Trend, struct {
				Date      string `json:"date"`
				Created   int    `json:"created"`
				Completed int    `json:"completed"`
			}{Date: dt, Created: cc, Completed: ct})
		}
	}

	return &stats, nil
}

func (s *MaintenanceService) GetUpcomingMaintenance(ctx context.Context, orgID, vehicleID int64, days int, km float64, limit int) ([]*UpcomingItem, error) {
	var items []*UpcomingItem
	whereSQL := "WHERE p.status='active'"
	args := []interface{}{}
	if orgID > 0 {
		whereSQL += " AND v.org_id = ?"
		args = append(args, orgID)
	}
	if vehicleID > 0 {
		whereSQL += " AND p.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit)

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT p.vehicle_id, v.plate_number, p.id, p.plan_name, p.maintenance_type,
		       p.next_mileage_km, v.current_mileage_km,
		       p.next_date, p.priority,
		       (p.next_mileage_km - v.current_mileage_km) AS mileage_left_km,
		       DATEDIFF(p.next_date, CURDATE()) AS days_left,
		       CASE
		         WHEN p.next_mileage_km <= v.current_mileage_km OR DATEDIFF(p.next_date, CURDATE()) <= 0 THEN 1
		         WHEN (p.next_mileage_km - v.current_mileage_km) <= p.warn_before_km OR DATEDIFF(p.next_date, CURDATE()) <= p.warn_before_days THEN 1
		         ELSE 0
		       END AS is_urgent
		FROM maintenance_plans p
		LEFT JOIN vehicles v ON v.id = p.vehicle_id
		`+whereSQL+`
		HAVING (mileage_left_km <= ? OR days_left <= ?)
		ORDER BY mileage_left_km ASC, days_left ASC
		LIMIT ?
	`, append(args[:len(args)-1], km, days, args[len(args)-1])...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it UpcomingItem
		var daysLeft int
		var mileageLeft float64
		var urgent int
		rows.Scan(&it.VehicleID, &it.VehiclePlate, &it.PlanID, &it.PlanName,
			&it.MaintenanceType, &it.NextMileage, &it.CurrentMileage,
			&it.NextDate, &it.Priority, &mileageLeft, &daysLeft, &urgent)
		it.MileageLeft = mileageLeft
		it.DaysLeft = daysLeft
		it.Urgent = urgent
		if urgent == 1 {
			it.WarnLevel = "urgent"
		} else if mileageLeft <= km || daysLeft <= days {
			it.WarnLevel = "warning"
		} else {
			it.WarnLevel = "normal"
		}
		items = append(items, &it)
	}
	return items, nil
}

func (s *MaintenanceService) GetOverdueMaintenance(ctx context.Context, orgID, vehicleID int64, limit int) ([]*UpcomingItem, error) {
	var items []*UpcomingItem
	whereSQL := "WHERE p.status='active' AND (p.next_mileage_km <= v.current_mileage_km OR DATEDIFF(p.next_date, CURDATE()) <= 0)"
	args := []interface{}{}
	if orgID > 0 {
		whereSQL += " AND v.org_id = ?"
		args = append(args, orgID)
	}
	if vehicleID > 0 {
		whereSQL += " AND p.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit)

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT p.vehicle_id, v.plate_number, p.id, p.plan_name, p.maintenance_type,
		       p.next_mileage_km, v.current_mileage_km,
		       p.next_date, p.priority,
		       (p.next_mileage_km - v.current_mileage_km) mileage_left_km,
		       DATEDIFF(p.next_date, CURDATE()) days_left, 1 urgent
		FROM maintenance_plans p
		LEFT JOIN vehicles v ON v.id = p.vehicle_id
		`+whereSQL+`
		ORDER BY days_left ASC, mileage_left_km ASC
		LIMIT ?
	`, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it UpcomingItem
		rows.Scan(&it.VehicleID, &it.VehiclePlate, &it.PlanID, &it.PlanName,
			&it.MaintenanceType, &it.NextMileage, &it.CurrentMileage,
			&it.NextDate, &it.Priority, &it.MileageLeft, &it.DaysLeft, &it.Urgent)
		it.WarnLevel = "overdue"
		items = append(items, &it)
	}
	return items, nil
}

func (s *MaintenanceService) recordWorkOrderLog(ctx context.Context, workOrderID int64, oldStatus, newStatus, actionType, actionNote string, operatorID int64, operatorName, operatorRole string) {
	if operatorName == "" {
		s.db.WithContext(ctx).Raw("SELECT name FROM users WHERE id=?", operatorID).Scan(&operatorName)
	}
	s.db.WithContext(ctx).Exec(`
		INSERT INTO maintenance_work_order_logs (
			work_order_id, old_status, new_status, action_type, action_note,
			operator_id, operator_name, operator_role, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`, workOrderID, oldStatus, newStatus, actionType, actionNote,
		operatorID, operatorName, operatorRole)
}

func sendMQ(data interface{}, topic string) {
	payload, _ := json.Marshal(data)
	_ = mq.Send(context.Background(), topic, string(payload))
	logger.GetLogger().Infof("[Maintenance] MQ sent topic=%s payload=%s", topic, string(payload))
}
