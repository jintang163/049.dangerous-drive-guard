package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"gorm.io/gorm"
)

type ScoreService struct {
	db *gorm.DB
}

func NewScoreService() *ScoreService {
	return &ScoreService{
		db: database.GetDB(),
	}
}

func (s *ScoreService) GetOverview(ctx context.Context) (*model.ScoreOverview, error) {
	var overview model.ScoreOverview

	err := s.db.WithContext(ctx).Table("driving_scores").
		Select("COUNT(DISTINCT driver_id) as total_drivers").
		Scan(&overview).Error
	if err != nil {
		return nil, fmt.Errorf("query overview error: %w", err)
	}

	type LevelCount struct {
		ScoreLevel string `json:"score_level"`
		Count      int    `json:"count"`
	}
	var levelCounts []LevelCount
	err = s.db.WithContext(ctx).Table("driving_scores").
		Select("score_level, COUNT(*) as count").
		Where("trip_date = CURDATE()").
		Group("score_level").
		Scan(&levelCounts).Error
	if err != nil {
		return nil, fmt.Errorf("query level counts error: %w", err)
	}

	for _, lc := range levelCounts {
		switch lc.ScoreLevel {
		case "excellent":
			overview.ExcellentCount = lc.Count
		case "good":
			overview.GoodCount = lc.Count
		case "normal":
			overview.NormalCount = lc.Count
		case "poor":
			overview.PoorCount = lc.Count
		case "danger":
			overview.DangerCount = lc.Count
		}
	}

	var avgResult struct {
		AvgScore float64 `json:"avg_score"`
	}
	err = s.db.WithContext(ctx).Table("driving_scores").
		Select("AVG(total_score) as avg_score").
		Where("trip_date = CURDATE()").
		Scan(&avgResult).Error
	if err == nil {
		overview.AvgScore = math.Round(avgResult.AvgScore*100) / 100
	}

	var todayCount int64
	s.db.WithContext(ctx).Table("driving_scores").
		Where("trip_date = CURDATE()").
		Count(&todayCount)
	overview.TodayCalculated = int(todayCount)

	var retrainingCount int64
	s.db.WithContext(ctx).Table("driver_retraining_tasks").
		Where("status IN ?", []string{"pending", "in_progress"}).
		Count(&retrainingCount)
	overview.RetrainingCount = int(retrainingCount)
	overview.NeedRetrainingCount = int(retrainingCount)

	return &overview, nil
}

func (s *ScoreService) GetLeaderboard(ctx context.Context, top int, orderBy string) ([]model.ScoreLeaderboardItem, error) {
	if top <= 0 {
		top = 20
	}
	if top > 100 {
		top = 100
	}

	if orderBy == "" {
		orderBy = "total_score"
	}

	var items []model.ScoreLeaderboardItem
	err := s.db.WithContext(ctx).Table("driving_scores ds").
		Select(`ds.driver_id, u.real_name as driver_name, v.plate_number,
			ds.total_score, ds.score_level,
			COALESCE(b.total_bonus, 0) as bonus_points,
			o.name as org_name`).
		Joins("LEFT JOIN users u ON u.id = ds.driver_id").
		Joins("LEFT JOIN vehicles v ON v.driver_id = ds.driver_id").
		Joins("LEFT JOIN organizations o ON o.id = u.org_id").
		Joins("LEFT JOIN (SELECT driver_id, SUM(bonus_points) as total_bonus FROM driving_score_bonus WHERE status=1 GROUP BY driver_id) b ON b.driver_id = ds.driver_id").
		Where("ds.trip_date = CURDATE()").
		Order(fmt.Sprintf("ds.%s DESC", orderBy)).
		Limit(top).
		Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("query leaderboard error: %w", err)
	}

	for i := range items {
		items[i].Rank = i + 1

		var avgResult struct {
			AvgScore float64 `json:"avg_score"`
		}
		s.db.WithContext(ctx).Table("driving_scores").
			Select("AVG(total_score) as avg_score").
			Where("driver_id = ? AND trip_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)", items[i].DriverID).
			Scan(&avgResult)
		items[i].AvgScore30d = math.Round(avgResult.AvgScore*100) / 100
	}

	return items, nil
}

func (s *ScoreService) GetDriverScore(ctx context.Context, driverID int64, days int) (*model.DrivingScore, []model.DrivingScore, error) {
	if days <= 0 {
		days = 30
	}

	var latestScore model.DrivingScore
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND trip_date <= CURDATE()", driverID).
		Order("trip_date DESC").
		First(&latestScore).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, fmt.Errorf("query latest score error: %w", err)
	}

	var history []model.DrivingScore
	err = s.db.WithContext(ctx).
		Where("driver_id = ? AND trip_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", driverID, days).
		Order("trip_date ASC").
		Find(&history).Error
	if err != nil {
		return nil, nil, fmt.Errorf("query score history error: %w", err)
	}

	return &latestScore, history, nil
}

func (s *ScoreService) GetDriverBonuses(ctx context.Context, driverID int64) ([]model.DrivingScoreBonus, error) {
	var bonuses []model.DrivingScoreBonus
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND status = 1", driverID).
		Order("created_at DESC").
		Find(&bonuses).Error
	if err != nil {
		return nil, fmt.Errorf("query bonuses error: %w", err)
	}
	return bonuses, nil
}

func (s *ScoreService) CheckAndAwardNoViolationBonus(ctx context.Context, driverID int64) (*model.DrivingScoreBonus, error) {
	var consecutiveCleanDays int
	err := s.db.WithContext(ctx).Table("driving_scores").
		Select("COUNT(*)").
		Where("driver_id = ? AND total_score >= 60 AND trip_date <= CURDATE()", driverID).
		Where("trip_date >= DATE_SUB(CURDATE(), INTERVAL 60 DAY)").
		Order("trip_date DESC").
		Scan(&consecutiveCleanDays).Error
	if err != nil {
		return nil, fmt.Errorf("check clean days error: %w", err)
	}

	var totalDays int
	s.db.WithContext(ctx).Table("driving_scores").
		Where("driver_id = ? AND trip_date <= CURDATE() AND trip_date >= DATE_SUB(CURDATE(), INTERVAL 60 DAY)", driverID).
		Count(&totalDays)

	if totalDays < 30 {
		return nil, nil
	}

	type DayScore struct {
		TripDate   time.Time `json:"trip_date"`
		TotalScore float64   `json:"total_score"`
	}
	var dayScores []DayScore
	s.db.WithContext(ctx).Table("driving_scores").
		Select("trip_date, total_score").
		Where("driver_id = ? AND trip_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY) AND trip_date <= CURDATE()", driverID).
		Order("trip_date DESC").
		Find(&dayScores)

	consecutiveCleanDays = 0
	for _, ds := range dayScores {
		if ds.TotalScore >= 60 {
			consecutiveCleanDays++
		} else {
			break
		}
	}

	if consecutiveCleanDays < 30 {
		return nil, nil
	}

	var existing model.DrivingScoreBonus
	err = s.db.WithContext(ctx).
		Where("driver_id = ? AND bonus_type = ? AND status = 1 AND end_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)",
			driverID, "no_violation_30d").
		First(&existing).Error
	if err == nil {
		return nil, nil
	}

	bonus := &model.DrivingScoreBonus{
		DriverID:    driverID,
		BonusType:   "no_violation_30d",
		BonusPoints: 5.0,
		Reason:      fmt.Sprintf("连续%d天无违规，奖励加分", consecutiveCleanDays),
		StreakDays:  consecutiveCleanDays,
		Status:      1,
	}
	now := time.Now()
	bonus.StartDate.Valid = true
	bonus.StartDate.Time = now.AddDate(0, 0, -consecutiveCleanDays)
	bonus.EndDate.Valid = true
	bonus.EndDate.Time = now
	bonus.AwardedBy.Valid = true
	bonus.AwardedBy.Int64 = 0

	err = s.db.WithContext(ctx).Create(bonus).Error
	if err != nil {
		return nil, fmt.Errorf("create bonus error: %w", err)
	}

	return bonus, nil
}

func (s *ScoreService) GetMonthlyReport(ctx context.Context, driverID int64, month string) (*model.DrivingScoreMonthlyReport, error) {
	var report model.DrivingScoreMonthlyReport
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND report_month = ?", driverID, month).
		First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.generateMonthlyReport(ctx, driverID, month)
		}
		return nil, fmt.Errorf("query monthly report error: %w", err)
	}
	return &report, nil
}

func (s *ScoreService) generateMonthlyReport(ctx context.Context, driverID int64, month string) (*model.DrivingScoreMonthlyReport, error) {
	var scores []model.DrivingScore
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND DATE_FORMAT(trip_date, '%Y-%m') = ?", driverID, month).
		Order("trip_date ASC").
		Find(&scores).Error
	if err != nil {
		return nil, fmt.Errorf("query scores for report error: %w", err)
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no score data for month %s", month)
	}

	var avgScore, minScore, maxScore float64
	var totalFatigue, totalSudden int
	var totalOverspeedDur float64
	var totalDist float64
	var totalDuration int64
	var violationDays, cleanDays int

	minScore = 101
	maxScore = -1
	sum := 0.0

	trend := make([]map[string]interface{}, 0, len(scores))
	for _, s := range scores {
		sum += s.TotalScore
		if s.TotalScore < minScore {
			minScore = s.TotalScore
		}
		if s.TotalScore > maxScore {
			maxScore = s.TotalScore
		}
		totalFatigue += s.FatigueAlarmCount
		totalSudden += s.SuddenBrakeCount + s.SuddenAccelCount + s.SharpTurnCount
		if s.OverspeedDuration.Valid {
			totalOverspeedDur += s.OverspeedDuration.Float64
		}
		if s.TotalDistance.Valid {
			totalDist += s.TotalDistance.Float64
		}
		if s.DrivingDuration.Valid {
			totalDuration += s.DrivingDuration.Int64
		}
		if s.TotalScore >= 60 {
			cleanDays++
		} else {
			violationDays++
		}
		trend = append(trend, map[string]interface{}{
			"date":  s.TripDate.Format("2006-01-02"),
			"score": s.TotalScore,
		})
	}
	avgScore = sum / float64(len(scores))

	trendJSON, _ := json.Marshal(trend)

	var totalBonus float64
	s.db.WithContext(ctx).Table("driving_score_bonus").
		Select("COALESCE(SUM(bonus_points), 0)").
		Where("driver_id = ? AND DATE_FORMAT(created_at, '%Y-%m') = ? AND status = 1", driverID, month).
		Scan(&totalBonus)

	needRetraining := 0
	if avgScore < 60 {
		needRetraining = 1
	}

	report := &model.DrivingScoreMonthlyReport{
		DriverID:               driverID,
		ReportMonth:            month,
		AvgScore:               math.Round(avgScore*100) / 100,
		MinScore:               sql.NullFloat64{Float64: minScore, Valid: true},
		MaxScore:               sql.NullFloat64{Float64: maxScore, Valid: true},
		TotalFatigueAlarms:     totalFatigue,
		TotalSuddenEvents:      totalSudden,
		TotalOverspeedDuration: sql.NullFloat64{Float64: totalOverspeedDur, Valid: true},
		TotalDistance:           sql.NullFloat64{Float64: totalDist, Valid: true},
		TotalDrivingDuration:    sql.NullInt64{Int64: totalDuration, Valid: true},
		TotalBonusPoints:        sql.NullFloat64{Float64: totalBonus, Valid: true},
		ViolationDays:          violationDays,
		CleanDays:              cleanDays,
		ScoreTrend:             model.JSON(trendJSON),
		NeedRetraining:         needRetraining,
		ReportSent:             0,
	}

	err = s.db.WithContext(ctx).Create(report).Error
	if err != nil {
		return nil, fmt.Errorf("create monthly report error: %w", err)
	}

	if needRetraining == 1 {
		s.triggerRetraining(ctx, driverID, avgScore, "low_score", month)
	}

	return report, nil
}

func (s *ScoreService) triggerRetraining(ctx context.Context, driverID int64, score float64, triggerType, month string) error {
	var existing model.DriverRetrainingTask
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND trigger_month = ? AND status IN ?", driverID, month, []string{"pending", "in_progress"}).
		First(&existing).Error
	if err == nil {
		return nil
	}

	task := &model.DriverRetrainingTask{
		DriverID:     driverID,
		TriggerScore: score,
		TriggerType:  triggerType,
		TriggerMonth: sql.NullString{String: month, Valid: true},
		TaskType:     "safety_training",
		Status:       "pending",
		CreatedBy:    sql.NullInt64{Int64: 0, Valid: true},
	}
	return s.db.WithContext(ctx).Create(task).Error
}

func (s *ScoreService) ListRetrainingTasks(ctx context.Context, page, pageSize int, status string) ([]model.DriverRetrainingTask, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.DriverRetrainingTask{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.DriverRetrainingTask
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (s *ScoreService) UpdateRetrainingTask(ctx context.Context, taskID int64, status, resultNote string, resultScore float64) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if resultNote != "" {
		updates["result_note"] = resultNote
	}
	if resultScore > 0 {
		updates["result_score"] = resultScore
	}
	switch status {
	case "in_progress":
		updates["started_at"] = time.Now()
	case "completed":
		updates["completed_at"] = time.Now()
	}
	return s.db.WithContext(ctx).Model(&model.DriverRetrainingTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

func (s *ScoreService) ListMonthlyReports(ctx context.Context, page, pageSize int, month, driverID string) ([]model.DrivingScoreMonthlyReport, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.DrivingScoreMonthlyReport{})

	if month != "" {
		query = query.Where("report_month = ?", month)
	}
	if driverID != "" {
		query = query.Where("driver_id = ?", driverID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reports []model.DrivingScoreMonthlyReport
	offset := (page - 1) * pageSize
	err := query.Order("report_month DESC, avg_score ASC").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

func (s *ScoreService) SendMonthlyReport(ctx context.Context, reportID int64) error {
	return s.db.WithContext(ctx).Model(&model.DrivingScoreMonthlyReport{}).
		Where("id = ?", reportID).
		Updates(map[string]interface{}{
			"report_sent":    1,
			"report_sent_at": time.Now(),
		}).Error
}
