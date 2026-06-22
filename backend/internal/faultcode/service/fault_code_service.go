package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type FaultCodeService struct {
	db *gorm.DB
}

type FaultCode struct {
	ID              int64     `json:"id"`
	FaultCode       string    `json:"fault_code"`
	FaultSystem     string    `json:"fault_system"`
	FaultCategory   string    `json:"fault_category"`
	FaultLevel      int       `json:"fault_level"`
	TitleCn         string    `json:"title_cn"`
	TitleEn         string    `json:"title_en"`
	Description     string    `json:"description"`
	PossibleCauses  string    `json:"possible_causes"`
	Symptoms        string    `json:"symptoms"`
	Suggestion      string    `json:"suggestion"`
	EmergencyAction string    `json:"emergency_action"`
	AutoCallRescue  int       `json:"auto_call_rescue"`
	RelatedSystems  string    `json:"related_systems"`
	OemSpec         string    `json:"oem_spec"`
	IsBuiltin       int       `json:"is_builtin"`
	Status          int       `json:"status"`
	CreatedBy       int64     `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type FaultCodeStats struct {
	Total    int `json:"total"`
	Level1   int `json:"level_1"`
	Level2   int `json:"level_2"`
	Level3   int `json:"level_3"`
	Level4   int `json:"level_4"`
	Builtin  int `json:"builtin"`
	Custom   int `json:"custom"`
	BySystem []struct {
		System string `json:"system"`
		Count  int    `json:"count"`
	} `json:"by_system"`
}

func NewFaultCodeService() *FaultCodeService {
	return &FaultCodeService{db: database.GetDB()}
}

func (s *FaultCodeService) ListFaultCodes(ctx context.Context, system, category string, level, status int, keyword string, page, pageSize int) ([]*FaultCode, int64, error) {
	var codes []*FaultCode
	var total int64

	countSQL := `SELECT COUNT(*) FROM fault_code_library WHERE 1=1`
	listSQL := `SELECT id, fault_code, fault_system, fault_category, fault_level, title_cn, title_en,
		description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue,
		related_systems, oem_spec, is_builtin, status, created_by, created_at, updated_at
		FROM fault_code_library WHERE 1=1`

	var args []interface{}

	if system != "" {
		countSQL += ` AND fault_system = ?`
		listSQL += ` AND fault_system = ?`
		args = append(args, system)
	}
	if category != "" {
		countSQL += ` AND fault_category = ?`
		listSQL += ` AND fault_category = ?`
		args = append(args, category)
	}
	if level > 0 {
		countSQL += ` AND fault_level = ?`
		listSQL += ` AND fault_level = ?`
		args = append(args, level)
	}
	if status >= 0 {
		countSQL += ` AND status = ?`
		listSQL += ` AND status = ?`
		args = append(args, status)
	}
	if keyword != "" {
		countSQL += ` AND (fault_code LIKE ? OR title_cn LIKE ?)`
		listSQL += ` AND (fault_code LIKE ? OR title_cn LIKE ?)`
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	s.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total)

	listSQL += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var fc FaultCode
		rows.Scan(&fc.ID, &fc.FaultCode, &fc.FaultSystem, &fc.FaultCategory, &fc.FaultLevel,
			&fc.TitleCn, &fc.TitleEn, &fc.Description, &fc.PossibleCauses, &fc.Symptoms,
			&fc.Suggestion, &fc.EmergencyAction, &fc.AutoCallRescue, &fc.RelatedSystems,
			&fc.OemSpec, &fc.IsBuiltin, &fc.Status, &fc.CreatedBy, &fc.CreatedAt, &fc.UpdatedAt)
		codes = append(codes, &fc)
	}

	return codes, total, nil
}

func (s *FaultCodeService) GetFaultCode(ctx context.Context, id int64) (*FaultCode, error) {
	var fc FaultCode
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, fault_code, fault_system, fault_category, fault_level, title_cn, title_en,
			description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue,
			related_systems, oem_spec, is_builtin, status, created_by, created_at, updated_at
		FROM fault_code_library WHERE id = ?`, id,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, gorm.ErrRecordNotFound
	}

	rows.Scan(&fc.ID, &fc.FaultCode, &fc.FaultSystem, &fc.FaultCategory, &fc.FaultLevel,
		&fc.TitleCn, &fc.TitleEn, &fc.Description, &fc.PossibleCauses, &fc.Symptoms,
		&fc.Suggestion, &fc.EmergencyAction, &fc.AutoCallRescue, &fc.RelatedSystems,
		&fc.OemSpec, &fc.IsBuiltin, &fc.Status, &fc.CreatedBy, &fc.CreatedAt, &fc.UpdatedAt)

	return &fc, nil
}

func (s *FaultCodeService) GetByCode(ctx context.Context, code string) (*FaultCode, error) {
	var fc FaultCode
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, fault_code, fault_system, fault_category, fault_level, title_cn, title_en,
			description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue,
			related_systems, oem_spec, is_builtin, status, created_by, created_at, updated_at
		FROM fault_code_library WHERE fault_code = ?`, code,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, gorm.ErrRecordNotFound
	}

	rows.Scan(&fc.ID, &fc.FaultCode, &fc.FaultSystem, &fc.FaultCategory, &fc.FaultLevel,
		&fc.TitleCn, &fc.TitleEn, &fc.Description, &fc.PossibleCauses, &fc.Symptoms,
		&fc.Suggestion, &fc.EmergencyAction, &fc.AutoCallRescue, &fc.RelatedSystems,
		&fc.OemSpec, &fc.IsBuiltin, &fc.Status, &fc.CreatedBy, &fc.CreatedAt, &fc.UpdatedAt)

	return &fc, nil
}

func (s *FaultCodeService) CreateFaultCode(ctx context.Context, fc *FaultCode) (*FaultCode, error) {
	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO fault_code_library
		(fault_code, fault_system, fault_category, fault_level, title_cn, title_en,
			description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue,
			related_systems, oem_spec, is_builtin, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		fc.FaultCode, fc.FaultSystem, fc.FaultCategory, fc.FaultLevel, fc.TitleCn, fc.TitleEn,
		fc.Description, fc.PossibleCauses, fc.Symptoms, fc.Suggestion, fc.EmergencyAction, fc.AutoCallRescue,
		fc.RelatedSystems, fc.OemSpec, fc.IsBuiltin, fc.Status, fc.CreatedBy,
	)
	if result.Error != nil {
		return nil, result.Error
	}

	var id int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)
	fc.ID = id

	s.db.WithContext(ctx).Raw(`SELECT created_at, updated_at FROM fault_code_library WHERE id = ?`, id).
		Scan(&fc.CreatedAt, &fc.UpdatedAt)

	return fc, nil
}

func (s *FaultCodeService) UpdateFaultCode(ctx context.Context, fc *FaultCode) (*FaultCode, error) {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE fault_code_library SET
			fault_system = ?, fault_category = ?, fault_level = ?, title_cn = ?, title_en = ?,
			description = ?, possible_causes = ?, symptoms = ?, suggestion = ?,
			emergency_action = ?, auto_call_rescue = ?, related_systems = ?, oem_spec = ?,
			status = ?, updated_at = NOW()
		WHERE id = ?`,
		fc.FaultSystem, fc.FaultCategory, fc.FaultLevel, fc.TitleCn, fc.TitleEn,
		fc.Description, fc.PossibleCauses, fc.Symptoms, fc.Suggestion,
		fc.EmergencyAction, fc.AutoCallRescue, fc.RelatedSystems, fc.OemSpec,
		fc.Status, fc.ID,
	)
	if result.Error != nil {
		return nil, result.Error
	}

	s.db.WithContext(ctx).Raw(`SELECT fault_code, is_builtin, created_by, created_at, updated_at
		FROM fault_code_library WHERE id = ?`, fc.ID).
		Scan(&fc.FaultCode, &fc.IsBuiltin, &fc.CreatedBy, &fc.CreatedAt, &fc.UpdatedAt)

	return fc, nil
}

func (s *FaultCodeService) DeleteFaultCode(ctx context.Context, id int64) error {
	var isBuiltin int
	s.db.WithContext(ctx).Raw(`SELECT is_builtin FROM fault_code_library WHERE id = ?`, id).Scan(&isBuiltin)
	if isBuiltin == 1 {
		return errors.New("builtin fault code cannot be deleted")
	}

	result := s.db.WithContext(ctx).Exec(`DELETE FROM fault_code_library WHERE id = ?`, id)
	return result.Error
}

func (s *FaultCodeService) SetStatus(ctx context.Context, id int64, status int) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE fault_code_library SET status = ?, updated_at = NOW() WHERE id = ?`,
		status, id,
	)
	return result.Error
}

func (s *FaultCodeService) BatchImport(ctx context.Context, codes []*FaultCode) (int, int, error) {
	success := 0
	failed := 0

	for _, fc := range codes {
		result := s.db.WithContext(ctx).Exec(`
			INSERT INTO fault_code_library
			(fault_code, fault_system, fault_category, fault_level, title_cn, title_en,
				description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue,
				related_systems, oem_spec, is_builtin, status, created_by, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				fault_system = VALUES(fault_system),
				fault_category = VALUES(fault_category),
				fault_level = VALUES(fault_level),
				title_cn = VALUES(title_cn),
				title_en = VALUES(title_en),
				description = VALUES(description),
				possible_causes = VALUES(possible_causes),
				symptoms = VALUES(symptoms),
				suggestion = VALUES(suggestion),
				emergency_action = VALUES(emergency_action),
				auto_call_rescue = VALUES(auto_call_rescue),
				related_systems = VALUES(related_systems),
				oem_spec = VALUES(oem_spec),
				status = VALUES(status),
				updated_at = NOW()`,
			fc.FaultCode, fc.FaultSystem, fc.FaultCategory, fc.FaultLevel, fc.TitleCn, fc.TitleEn,
			fc.Description, fc.PossibleCauses, fc.Symptoms, fc.Suggestion, fc.EmergencyAction, fc.AutoCallRescue,
			fc.RelatedSystems, fc.OemSpec, fc.IsBuiltin, fc.Status, fc.CreatedBy,
		)
		if result.Error != nil {
			failed++
			logger.Errorf("batch import fault code %s error: %v", fc.FaultCode, result.Error)
		} else {
			success++
		}
	}

	return success, failed, nil
}

func (s *FaultCodeService) GetStats(ctx context.Context) (*FaultCodeStats, error) {
	stats := &FaultCodeStats{}

	row := s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) as total,
		       SUM(CASE WHEN fault_level = 1 THEN 1 ELSE 0 END) as level_1,
		       SUM(CASE WHEN fault_level = 2 THEN 1 ELSE 0 END) as level_2,
		       SUM(CASE WHEN fault_level = 3 THEN 1 ELSE 0 END) as level_3,
		       SUM(CASE WHEN fault_level = 4 THEN 1 ELSE 0 END) as level_4,
		       SUM(CASE WHEN is_builtin = 1 THEN 1 ELSE 0 END) as builtin,
		       SUM(CASE WHEN is_builtin = 0 THEN 1 ELSE 0 END) as custom
		FROM fault_code_library`,
	).Row()
	row.Scan(&stats.Total, &stats.Level1, &stats.Level2, &stats.Level3, &stats.Level4, &stats.Builtin, &stats.Custom)

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT fault_system as system, COUNT(*) as count
		FROM fault_code_library
		WHERE fault_system IS NOT NULL AND fault_system != ''
		GROUP BY fault_system
		ORDER BY count DESC`,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			System string `json:"system"`
			Count  int    `json:"count"`
		}
		rows.Scan(&item.System, &item.Count)
		stats.BySystem = append(stats.BySystem, item)
	}

	return stats, nil
}
