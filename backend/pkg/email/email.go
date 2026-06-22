package email

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type Service struct {
	cfg *config.SMTPConfig
}

var (
	instance *Service
)

func Init(cfg *config.SMTPConfig) {
	instance = &Service{cfg: cfg}
	logger.Sugar.Infof("Email service initialized: %s:%d", cfg.Host, cfg.Port)
}

func GetService() *Service {
	if instance == nil {
		return &Service{cfg: &config.SMTPConfig{}}
	}
	return instance
}

func (s *Service) Send(ctx context.Context, to []string, subject, body string) error {
	if s.cfg.Host == "" {
		logger.Sugar.Warn("SMTP not configured, skipping email send")
		return nil
	}

	from := s.cfg.From
	if from == "" {
		from = "ddg-system@dangerous-drive-guard.com"
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	err := smtp.SendMail(addr, auth, from, to, msg.Bytes())
	if err != nil {
		logger.Sugar.Errorf("Send email error: to=%v, subject=%s, err=%v", to, subject, err)
		return fmt.Errorf("send email error: %w", err)
	}

	logger.Sugar.Infof("Email sent: to=%v, subject=%s", to, subject)
	return nil
}

type MonthlyReportData struct {
	DriverName           string
	ReportMonth          string
	AvgScore             float64
	MinScore             float64
	MaxScore             float64
	TotalFatigueAlarms   int
	TotalSuddenEvents    int
	TotalOverspeedDur    float64
	TotalDistance        float64
	TotalDrivingDuration int64
	TotalBonusPoints     float64
	ViolationDays        int
	CleanDays            int
	NeedRetraining       bool
	ScoreTrend           []ScoreTrendPoint
}

type ScoreTrendPoint struct {
	Date  string
	Score float64
}

const monthlyReportTpl = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><style>
body { font-family: 'Microsoft YaHei', sans-serif; color: #333; max-width: 800px; margin: 0 auto; padding: 20px; }
.header { background: linear-gradient(135deg, #1677ff, #0958d9); color: white; padding: 30px; border-radius: 12px 12px 0 0; text-align: center; }
.header h1 { margin: 0; font-size: 24px; }
.header p { margin: 8px 0 0; opacity: 0.9; }
.content { background: #fff; border: 1px solid #e8e8e8; border-top: none; padding: 24px; border-radius: 0 0 12px 12px; }
.score-circle { width: 120px; height: 120px; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 20px auto; }
.score-value { font-size: 36px; font-weight: 700; }
.section { margin: 20px 0; }
.section h3 { color: #1677ff; border-bottom: 2px solid #1677ff; padding-bottom: 8px; }
table { width: 100%; border-collapse: collapse; margin: 12px 0; }
table th, table td { padding: 10px 12px; text-align: left; border-bottom: 1px solid #f0f0f0; }
table th { background: #fafafa; font-weight: 600; color: #666; }
.danger { color: #ff4d4f; font-weight: 700; }
.warning { color: #fa8c16; font-weight: 700; }
.success { color: #52c41a; font-weight: 700; }
.badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; color: #fff; }
.badge-danger { background: #ff4d4f; }
.badge-success { background: #52c41a; }
.badge-warning { background: #fa8c16; }
.footer { text-align: center; color: #999; font-size: 12px; margin-top: 24px; padding-top: 16px; border-top: 1px solid #f0f0f0; }
</style></head>
<body>
<div class="header">
  <h1>🚛 驾驶行为评分月报</h1>
  <p>{{.DriverName}} · {{.ReportMonth}}</p>
</div>
<div class="content">
  <div style="text-align:center;">
    <div class="score-circle" style="border: 4px solid {{if ge .AvgScore 90.0}}#52c41a{{else if ge .AvgScore 75.0}}#1677ff{{else if ge .AvgScore 60.0}}#faad14{{else}}#ff4d4f{{end}};">
      <span class="score-value" style="color: {{if ge .AvgScore 90.0}}#52c41a{{else if ge .AvgScore 75.0}}#1677ff{{else if ge .AvgScore 60.0}}#faad14{{else}}#ff4d4f{{end}};">{{printf "%.1f" .AvgScore}}</span>
    </div>
    <p>月度平均安全评分</p>
  </div>

  {{if .NeedRetraining}}
  <div style="background:#fff2f0; border:1px solid #ffccc7; border-radius:8px; padding:16px; margin:16px 0; text-align:center;">
    <span class="danger">⚠️ 您的月均评分低于60分，已自动触发再培训任务，请尽快完成安全培训。</span>
  </div>
  {{end}}

  <div class="section">
    <h3>📊 评分概览</h3>
    <table>
      <tr><th>指标</th><th>数值</th></tr>
      <tr><td>最高分</td><td class="success">{{printf "%.1f" .MaxScore}}</td></tr>
      <tr><td>最低分</td><td class="{{if lt .MinScore 60.0}}danger{{else}}success{{end}}">{{printf "%.1f" .MinScore}}</td></tr>
      <tr><td>无违规天数</td><td>{{.CleanDays}} 天</td></tr>
      <tr><td>违规天数</td><td class="{{if gt .ViolationDays 0}}danger{{end}}">{{.ViolationDays}} 天</td></tr>
      <tr><td>加分项</td><td class="success">+{{printf "%.1f" .TotalBonusPoints}}</td></tr>
    </table>
  </div>

  <div class="section">
    <h3>⚠️ 违规统计</h3>
    <table>
      <tr><th>违规类型</th><th>次数/时长</th></tr>
      <tr><td>疲劳报警</td><td class="{{if gt .TotalFatigueAlarms 0}}danger{{end}}">{{.TotalFatigueAlarms}} 次</td></tr>
      <tr><td>急加速/急刹车/急转弯</td><td class="{{if gt .TotalSuddenEvents 0}}warning{{end}}">{{.TotalSuddenEvents}} 次</td></tr>
      <tr><td>超速时长</td><td class="{{if gt .TotalOverspeedDur 0}}warning{{end}}">{{printf "%.1f" .TotalOverspeedDur}} 分钟</td></tr>
    </table>
  </div>

  <div class="section">
    <h3>🚗 行驶统计</h3>
    <table>
      <tr><th>指标</th><th>数值</th></tr>
      <tr><td>总里程</td><td>{{printf "%.1f" .TotalDistance}} km</td></tr>
      <tr><td>总驾驶时长</td><td>{{printf "%.0f" (fdiv .TotalDrivingDuration 60.0)}} 小时</td></tr>
    </table>
  </div>

  <div class="footer">
    <p>本报告由危险品运输安全监控平台自动生成 · {{.ReportMonth}}</p>
    <p>如有疑问请联系管理员</p>
  </div>
</div>
</body>
</html>`

var reportTemplate *template.Template

func init() {
	funcMap := template.FuncMap{
		"ge":   func(a, b float64) bool { return a >= b },
		"lt":   func(a, b float64) bool { return a < b },
		"gt":   func(a, b int) bool { return a > b },
		"fdiv": func(a int64, b float64) float64 { return float64(a) / b },
	}
	reportTemplate = template.Must(template.New("monthly_report").Funcs(funcMap).Parse(monthlyReportTpl))
}

func (s *Service) RenderMonthlyReport(data *MonthlyReportData) (string, error) {
	var buf bytes.Buffer
	if err := reportTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render monthly report template error: %w", err)
	}
	return buf.String(), nil
}

func (s *Service) SendMonthlyReportEmail(ctx context.Context, to []string, data *MonthlyReportData) error {
	subject := fmt.Sprintf("【危运安全监控】%s 驾驶行为评分月报 - %s", data.DriverName, data.ReportMonth)
	body, err := s.RenderMonthlyReport(data)
	if err != nil {
		return err
	}
	return s.Send(ctx, to, subject, body)
}
