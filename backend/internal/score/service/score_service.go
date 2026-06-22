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
	"github.com/dangerous-drive-guard/backend/pkg/email"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
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

const (
	FatigueDeductionPerAlarm    = 3.0
	OverspeedDeductionPerEvent  = 2.0
	SuddenBrakeDeductionPerEvt  = 2.0
	SuddenAccelDeductionPerEvt  = 2.0
	SharpTurnDeductionPerEvt    = 2.0
	LaneDeviationDeductionPerEvt = 1.0
	PhoneUsageDeductionPerEvt   = 5.0
	SmokingDeductionPerEvt      = 5.0
	SeatbeltDeductionPerEvt     = 5.0
	RouteDeviationDeductionPerEvt = 3.0
	RetrainingThreshold         = 60.0
)

func getScoreLevel(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "normal"
	case score >= 40:
		return "poor"
	default:
		return "danger"
	}
}

type DailyEventCounts struct {
	FatigueAlarmCount      int
	OverspeedCount         int
	SuddenBrakeCount       int
	SuddenAccelCount       int
	SharpTurnCount         int
	LaneDeviationCount     int
	PhoneUsageCount        int
	SmokingCount           int
	SeatbeltViolationCount int
	RouteDeviationCount    int
	OverspeedDurationMin   float64
	TotalDistanceKm        float64
	DrivingDurationMin     int64
	NightDrivingDurationMin int64
}

func (s *ScoreService) CalculateDailyScore(ctx context.Context, driverID int64, tripDate time.Time, events DailyEventCounts) (*model.DrivingScore, error) {
	var existing model.DrivingScore
	err := s.db.WithContext(ctx).
		Where("driver_id = ? AND trip_date = ?", driverID, tripDate.Format("2006-01-02")).
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}

	fatigueDeduction := float64(events.FatigueAlarmCount) * FatigueDeductionPerAlarm
	overspeedDeduction := float64(events.OverspeedCount) * OverspeedDeductionPerEvent
	suddenBrakeDeduction := float64(events.SuddenBrakeCount) * SuddenBrakeDeductionPerEvt
	suddenAccelDeduction := float64(events.SuddenAccelCount) * SuddenAccelDeductionPerEvt
	sharpTurnDeduction := float64(events.SharpTurnCount) * SharpTurnDeductionPerEvt
	laneDeviationDeduction := float64(events.LaneDeviationCount) * LaneDeviationDeductionPerEvt
	phoneUsageDeduction := float64(events.PhoneUsageCount) * PhoneUsageDeductionPerEvt
	smokingDeduction := float64(events.SmokingCount) * SmokingDeductionPerEvt
	seatbeltDeduction := float64(events.SeatbeltViolationCount) * SeatbeltDeductionPerEvt
	routeDeviationDeduction := float64(events.RouteDeviationCount) * RouteDeviationDeductionPerEvt

	totalDeduction := fatigueDeduction + overspeedDeduction + suddenBrakeDeduction +
		suddenAccelDeduction + sharpTurnDeduction + laneDeviationDeduction +
		phoneUsageDeduction + smokingDeduction + seatbeltDeduction + routeDeviationDeduction

	totalScore := 100.0 - totalDeduction
	if totalScore < 0 {
		totalScore = 0
	}
	totalScore = math.Round(totalScore*100) / 100

	scoreLevel := getScoreLevel(totalScore)

	fatigueScore := math.Max(0, 100-fatigueDeduction)
	overspeedScore := math.Max(0, 100-overspeedDeduction)

	score := &model.DrivingScore{
		DriverID:               driverID,
		TripDate:               tripDate,
		TotalScore:             totalScore,
		ScoreLevel:             scoreLevel,
		FatigueScore:           sql.NullFloat64{Float64: math.Round(fatigueScore*100) / 100, Valid: true},
		FatigueDeduction:       math.Round(fatigueDeduction*100) / 100,
		OverspeedScore:         sql.NullFloat64{Float64: math.Round(overspeedScore*100) / 100, Valid: true},
		OverspeedCount:         events.OverspeedCount,
		OverspeedDeduction:     math.Round(overspeedDeduction*100) / 100,
		SuddenBrakeCount:       events.SuddenBrakeCount,
		SuddenBrakeDeduction:   math.Round(suddenBrakeDeduction*100) / 100,
		SuddenAccelCount:       events.SuddenAccelCount,
		SuddenAccelDeduction:   math.Round(suddenAccelDeduction*100) / 100,
		SharpTurnCount:         events.SharpTurnCount,
		SharpTurnDeduction:     math.Round(sharpTurnDeduction*100) / 100,
		LaneDeviationCount:     events.LaneDeviationCount,
		LaneDeviationDeduction: math.Round(laneDeviationDeduction*100) / 100,
		PhoneUsageCount:        events.PhoneUsageCount,
		PhoneUsageDeduction:    math.Round(phoneUsageDeduction*100) / 100,
		SmokingCount:           events.SmokingCount,
		SmokingDeduction:       math.Round(smokingDeduction*100) / 100,
		SeatbeltViolationCount: events.SeatbeltViolationCount,
		SeatbeltDeduction:      math.Round(seatbeltDeduction*100) / 100,
		RouteDeviationCount:    events.RouteDeviationCount,
		RouteDeviationDeduction: math.Round(routeDeviationDeduction*100) / 100,
		FatigueAlarmCount:      events.FatigueAlarmCount,
		TotalDistance:           sql.NullFloat64{Float64: events.TotalDistanceKm, Valid: events.TotalDistanceKm > 0},
		DrivingDuration:        sql.NullInt64{Int64: events.DrivingDurationMin, Valid: events.DrivingDurationMin > 0},
		NightDrivingDuration:   sql.NullInt64{Int64: events.NightDrivingDurationMin, Valid: events.NightDrivingDurationMin > 0},
		OverspeedDuration:      sql.NullFloat64{Float64: events.OverspeedDurationMin, Valid: events.OverspeedDurationMin > 0},
	}

	if err := s.db.WithContext(ctx).Create(score).Error; err != nil {
		return nil, fmt.Errorf("create daily score error: %w", err)
	}

	if totalScore < RetrainingThreshold {
		if err := s.triggerDailyRetraining(ctx, driverID, totalScore, tripDate); err != nil {
			logger.Sugar.Errorf("trigger retraining for driver %d error: %v", driverID, err)
		}
	}

	if err := s.CheckAndAwardNoViolationBonusSilent(ctx, driverID); err != nil {
		logger.Sugar.Errorf("check bonus for driver %d error: %v", driverID, err)
	}

	return score, nil
}

func (s *ScoreService) RunDailyScoreCalculation(ctx context.Context) error {
	yesterday := time.Now().AddDate(0, 0, -1)

	type DriverEventAgg struct {
		DriverID             int64   `json:"driver_id"`
		FatigueAlarmCount    int     `json:"fatigue_alarm_count"`
		OverspeedCount       int     `json:"overspeed_count"`
		SuddenBrakeCount     int     `json:"sudden_brake_count"`
		SuddenAccelCount     int     `json:"sudden_accel_count"`
		SharpTurnCount       int     `json:"sharp_turn_count"`
		LaneDeviationCount   int     `json:"lane_deviation_count"`
		PhoneUsageCount      int     `json:"phone_usage_count"`
		SmokingCount         int     `json:"smoking_count"`
		SeatbeltViolationCount int   `json:"seatbelt_violation_count"`
		RouteDeviationCount  int     `json:"route_deviation_count"`
		OverspeedDurationMin float64 `json:"overspeed_duration_min"`
		TotalDistanceKm      float64 `json:"total_distance_km"`
		DrivingDurationMin   int64   `json:"driving_duration_min"`
		NightDrivingDurationMin int64 `json:"night_driving_duration_min"`
	}

	var drivers []model.User
	if err := s.db.WithContext(ctx).Where("role = ? AND status = 1", model.RoleDriver).Find(&drivers).Error; err != nil {
		return fmt.Errorf("query drivers error: %w", err)
	}

	calculated := 0
	for _, d := range drivers {
		var agg DriverEventAgg
		s.db.WithContext(ctx).Table("fatigue_alarms").
			Select("driver_id, COUNT(*) as fatigue_alarm_count").
			Where("driver_id = ? AND DATE(created_at) = ?", d.ID, yesterday.Format("2006-01-02")).
			Scan(&agg)

		var adasAgg struct {
			LaneDevCount   int `json:"lane_dev_count"`
			OverspeedCount int `json:"overspeed_count"`
		}
		s.db.WithContext(ctx).Table("adas_alerts").
			Select("driver_id, SUM(CASE WHEN alert_type='lane_departure' THEN 1 ELSE 0 END) as lane_dev_count, SUM(CASE WHEN alert_type='forward_collision' THEN 1 ELSE 0 END) as overspeed_count").
			Where("driver_id = ? AND DATE(created_at) = ?", d.ID, yesterday.Format("2006-01-02")).
			Scan(&adasAgg)

		agg.LaneDeviationCount = adasAgg.LaneDevCount
		agg.OverspeedCount = adasAgg.OverspeedCount

		_, err := s.CalculateDailyScore(ctx, d.ID, yesterday, DailyEventCounts{
			FatigueAlarmCount:      agg.FatigueAlarmCount,
			OverspeedCount:         agg.OverspeedCount,
			SuddenBrakeCount:       agg.SuddenBrakeCount,
			SuddenAccelCount:       agg.SuddenAccelCount,
			SharpTurnCount:         agg.SharpTurnCount,
			LaneDeviationCount:     agg.LaneDeviationCount,
			PhoneUsageCount:        agg.PhoneUsageCount,
			SmokingCount:           agg.SmokingCount,
			SeatbeltViolationCount: agg.SeatbeltViolationCount,
			RouteDeviationCount:    agg.RouteDeviationCount,
			OverspeedDurationMin:   agg.OverspeedDurationMin,
			TotalDistanceKm:        agg.TotalDistanceKm,
			DrivingDurationMin:     agg.DrivingDurationMin,
			NightDrivingDurationMin: agg.NightDrivingDurationMin,
		})
		if err != nil {
			logger.Sugar.Errorf("calculate daily score for driver %d error: %v", d.ID, err)
			continue
		}
		calculated++
	}

	logger.Sugar.Infof("Daily score calculation completed: %d drivers calculated for %s", calculated, yesterday.Format("2006-01-02"))
	return nil
}

func (s *ScoreService) triggerDailyRetraining(ctx context.Context, driverID int64, score float64, tripDate time.Time) error {
	month := tripDate.Format("2006-01")
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
		TriggerType:  "low_score",
		TriggerMonth: sql.NullString{String: month, Valid: true},
		TaskType:     "safety_training",
		Status:       "pending",
		CreatedBy:    sql.NullInt64{Int64: 0, Valid: true},
	}
	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return err
	}

	logger.Sugar.Infof("Auto retraining task created: driver_id=%d, score=%.1f, month=%s", driverID, score, month)
	return nil
}

func (s *ScoreService) CheckAndAwardNoViolationBonusSilent(ctx context.Context, driverID int64) error {
	_, err := s.CheckAndAwardNoViolationBonus(ctx, driverID)
	return err
}

func (s *ScoreService) RunMonthlyReportGeneration(ctx context.Context) error {
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")

	var drivers []model.User
	if err := s.db.WithContext(ctx).Where("role = ? AND status = 1", model.RoleDriver).Find(&drivers).Error; err != nil {
		return fmt.Errorf("query drivers error: %w", err)
	}

	generated := 0
	for _, d := range drivers {
		_, err := s.GetMonthlyReport(ctx, d.ID, lastMonth)
		if err != nil {
			logger.Sugar.Errorf("generate monthly report for driver %d (%s) error: %v", d.ID, lastMonth, err)
			continue
		}
		generated++
	}

	logger.Sugar.Infof("Monthly report generation completed: %d reports for %s", generated, lastMonth)
	return nil
}

func (s *ScoreService) RunMonthlyReportPush(ctx context.Context) error {
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")

	var reports []model.DrivingScoreMonthlyReport
	if err := s.db.WithContext(ctx).
		Where("report_month = ? AND report_sent = 0", lastMonth).
		Find(&reports).Error; err != nil {
		return fmt.Errorf("query unsent reports error: %w", err)
	}

	emailSvc := email.GetService()
	sent := 0
	for _, r := range reports {
		if err := s.sendReportEmail(ctx, emailSvc, &r); err != nil {
			logger.Sugar.Errorf("send report email for report %d error: %v", r.ID, err)
			continue
		}
		s.db.WithContext(ctx).Model(&model.DrivingScoreMonthlyReport{}).
			Where("id = ?", r.ID).
			Updates(map[string]interface{}{
				"report_sent":    1,
				"report_sent_at": time.Now(),
			})
		sent++
	}

	logger.Sugar.Infof("Monthly report push completed: %d/%d reports sent for %s", sent, len(reports), lastMonth)
	return nil
}

func (s *ScoreService) sendReportEmail(ctx context.Context, emailSvc *email.Service, report *model.DrivingScoreMonthlyReport) error {
	var driver model.User
	if err := s.db.WithContext(ctx).First(&driver, report.DriverID).Error; err != nil {
		return fmt.Errorf("query driver error: %w", err)
	}

	var trend []email.ScoreTrendPoint
	if report.ScoreTrend != nil {
		var rawTrend []map[string]interface{}
		if err := json.Unmarshal(report.ScoreTrend, &rawTrend); err == nil {
			for _, t := range rawTrend {
				date, _ := t["date"].(string)
				score, _ := t["score"].(float64)
				trend = append(trend, email.ScoreTrendPoint{Date: date, Score: score})
			}
		}
	}

	data := &email.MonthlyReportData{
		DriverName:           driver.RealName,
		ReportMonth:          report.ReportMonth,
		AvgScore:             report.AvgScore,
		MinScore:             report.MinScore.Float64,
		MaxScore:             report.MaxScore.Float64,
		TotalFatigueAlarms:   report.TotalFatigueAlarms,
		TotalSuddenEvents:    report.TotalSuddenEvents,
		TotalOverspeedDur:    report.TotalOverspeedDuration.Float64,
		TotalDistance:        report.TotalDistance.Float64,
		TotalDrivingDuration: report.TotalDrivingDuration.Int64,
		TotalBonusPoints:     report.TotalBonusPoints.Float64,
		ViolationDays:        report.ViolationDays,
		CleanDays:            report.CleanDays,
		NeedRetraining:       report.NeedRetraining == 1,
		ScoreTrend:           trend,
	}

	recipients := []string{}
	if driver.Email != "" {
		recipients = append(recipients, driver.Email)
	}

	var admins []model.User
	s.db.WithContext(ctx).Where("role = ? AND status = 1 AND email != ''", model.RoleAdmin).Find(&admins)
	for _, a := range admins {
		recipients = append(recipients, a.Email)
	}

	if len(recipients) == 0 {
		logger.Sugar.Warnf("No recipients for monthly report: driver_id=%d", report.DriverID)
		return nil
	}

	return emailSvc.SendMonthlyReportEmail(ctx, recipients, data)
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

	if len(dayScores) < 30 {
		return nil, nil
	}

	consecutiveCleanDays := 0
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
	err := s.db.WithContext(ctx).
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

	logger.Sugar.Infof("Bonus awarded: driver_id=%d, type=no_violation_30d, points=5, streak=%d days", driverID, consecutiveCleanDays)
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
	for _, sc := range scores {
		sum += sc.TotalScore
		if sc.TotalScore < minScore {
			minScore = sc.TotalScore
		}
		if sc.TotalScore > maxScore {
			maxScore = sc.TotalScore
		}
		totalFatigue += sc.FatigueAlarmCount
		totalSudden += sc.SuddenBrakeCount + sc.SuddenAccelCount + sc.SharpTurnCount
		if sc.OverspeedDuration.Valid {
			totalOverspeedDur += sc.OverspeedDuration.Float64
		}
		if sc.TotalDistance.Valid {
			totalDist += sc.TotalDistance.Float64
		}
		if sc.DrivingDuration.Valid {
			totalDuration += sc.DrivingDuration.Int64
		}
		if sc.TotalScore >= 60 {
			cleanDays++
		} else {
			violationDays++
		}
		trend = append(trend, map[string]interface{}{
			"date":  sc.TripDate.Format("2006-01-02"),
			"score": sc.TotalScore,
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
	if avgScore < RetrainingThreshold {
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
	var report model.DrivingScoreMonthlyReport
	if err := s.db.WithContext(ctx).First(&report, reportID).Error; err != nil {
		return fmt.Errorf("query report error: %w", err)
	}

	emailSvc := email.GetService()
	if err := s.sendReportEmail(ctx, emailSvc, &report); err != nil {
		return fmt.Errorf("send report email error: %w", err)
	}

	return s.db.WithContext(ctx).Model(&model.DrivingScoreMonthlyReport{}).
		Where("id = ?", reportID).
		Updates(map[string]interface{}{
			"report_sent":    1,
			"report_sent_at": time.Now(),
		}).Error
}

func (s *ScoreService) BatchSendMonthlyReports(ctx context.Context, month string) (int, int, error) {
	var reports []model.DrivingScoreMonthlyReport
	if err := s.db.WithContext(ctx).
		Where("report_month = ? AND report_sent = 0", month).
		Find(&reports).Error; err != nil {
		return 0, 0, fmt.Errorf("query unsent reports error: %w", err)
	}

	emailSvc := email.GetService()
	sent, failed := 0, 0
	for _, r := range reports {
		if err := s.sendReportEmail(ctx, emailSvc, &r); err != nil {
			logger.Sugar.Errorf("batch send report %d error: %v", r.ID, err)
			failed++
			continue
		}
		s.db.WithContext(ctx).Model(&model.DrivingScoreMonthlyReport{}).
			Where("id = ?", r.ID).
			Updates(map[string]interface{}{
				"report_sent":    1,
				"report_sent_at": time.Now(),
			})
		sent++
	}

	return sent, failed, nil
}
