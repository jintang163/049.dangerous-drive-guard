package score

import (
	"context"
	"time"

	scoreSvc "github.com/dangerous-drive-guard/backend/internal/score/service"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type CronScheduler struct {
	service *scoreSvc.ScoreService
	stopCh  chan struct{}
}

func NewCronScheduler(svc *scoreSvc.ScoreService) *CronScheduler {
	return &CronScheduler{
		service: svc,
		stopCh:  make(chan struct{}),
	}
}

func (c *CronScheduler) Start() {
	go c.runDailyCalcLoop()
	go c.runMonthlyReportLoop()
	logger.Sugar.Info("Score cron scheduler started")
}

func (c *CronScheduler) Stop() {
	close(c.stopCh)
	logger.Sugar.Info("Score cron scheduler stopped")
}

func (c *CronScheduler) runDailyCalcLoop() {
	cronExpr := "02:00"
	if cfg := config.Global; cfg != nil && cfg.Score.CalcCron != "" {
		cronExpr = cfg.Score.CalcCron
	}

	hour, minute := parseCronTime(cronExpr)
	interval := calcNextInterval(hour, minute)

	for {
		select {
		case <-c.stopCh:
			return
		case <-time.After(interval):
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			if err := c.service.RunDailyScoreCalculation(ctx); err != nil {
				logger.Sugar.Errorf("Daily score calculation cron error: %v", err)
			}
			cancel()
			interval = calcNextInterval(hour, minute)
		}
	}
}

func (c *CronScheduler) runMonthlyReportLoop() {
	for {
		select {
		case <-c.stopCh:
			return
		case <-time.After(time.Hour):
			now := time.Now()
			if now.Day() == 1 && now.Hour() == 3 && now.Minute() < 5 {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
				if err := c.service.RunMonthlyReportGeneration(ctx); err != nil {
					logger.Sugar.Errorf("Monthly report generation cron error: %v", err)
				}
				if err := c.service.RunMonthlyReportPush(ctx); err != nil {
					logger.Sugar.Errorf("Monthly report push cron error: %v", err)
				}
				cancel()
				time.Sleep(10 * time.Minute)
			}
		}
	}
}

func parseCronTime(expr string) (int, int) {
	hour, minute := 2, 0
	for i, ch := range expr {
		if ch == ':' {
			if i > 0 {
				hPart := expr[:i]
				for _, c := range hPart {
					if c >= '0' && c <= '9' {
						hour = hour*10 + int(c-'0')
					}
				}
			}
			if i+1 < len(expr) {
				mPart := expr[i+1:]
				for _, c := range mPart {
					if c >= '0' && c <= '9' {
						minute = minute*10 + int(c-'0')
					}
				}
			}
			break
		}
	}
	return hour, minute
}

func calcNextInterval(hour, minute int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}
