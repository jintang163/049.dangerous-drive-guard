-- ============================================================
-- 驾驶行为评分模块扩展表
-- ============================================================

USE ddg_db;

CREATE TABLE IF NOT EXISTS driving_score_bonus (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    driver_id BIGINT NOT NULL,
    bonus_type VARCHAR(32) NOT NULL COMMENT 'no_violation_30d/safety_champion/continuous_clean',
    bonus_points DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '奖励加分',
    reason VARCHAR(256),
    streak_days INT DEFAULT 0 COMMENT '连续无违规天数',
    start_date DATE,
    end_date DATE,
    awarded_by BIGINT COMMENT '颁发人ID(自动为0)',
    status TINYINT DEFAULT 1 COMMENT '1-有效 0-已撤销',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_driver (driver_id),
    INDEX idx_type (bonus_type),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='驾驶评分加分项表';

CREATE TABLE IF NOT EXISTS driving_score_monthly_report (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    driver_id BIGINT NOT NULL,
    report_month VARCHAR(7) NOT NULL COMMENT '报告月份 YYYY-MM',
    avg_score DECIMAL(6,2) NOT NULL DEFAULT 0 COMMENT '月度平均评分',
    min_score DECIMAL(6,2) COMMENT '月度最低评分',
    max_score DECIMAL(6,2) COMMENT '月度最高评分',
    total_fatigue_alarms INT DEFAULT 0,
    total_sudden_events INT DEFAULT 0 COMMENT '急加速+急刹车+急转弯',
    total_overspeed_duration DECIMAL(10,2) DEFAULT 0 COMMENT '超速时长(分钟)',
    total_distance DECIMAL(10,2) DEFAULT 0 COMMENT '月度总里程(km)',
    total_driving_duration INT DEFAULT 0 COMMENT '月度总驾驶时长(分钟)',
    total_bonus_points DECIMAL(5,2) DEFAULT 0 COMMENT '月度总加分',
    violation_days INT DEFAULT 0 COMMENT '违规天数',
    clean_days INT DEFAULT 0 COMMENT '无违规天数',
    score_trend JSON COMMENT '每日评分趋势 [{date,score}]',
    need_retraining TINYINT DEFAULT 0 COMMENT '是否需要再培训(低于60分)',
    retraining_triggered_at DATETIME,
    report_sent TINYINT DEFAULT 0 COMMENT '报告是否已推送',
    report_sent_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_driver_month (driver_id, report_month),
    INDEX idx_month (report_month),
    INDEX idx_retraining (need_retraining)
) ENGINE=InnoDB COMMENT='驾驶评分月报表';

CREATE TABLE IF NOT EXISTS driver_retraining_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    driver_id BIGINT NOT NULL,
    trigger_score DECIMAL(6,2) NOT NULL COMMENT '触发再培训的评分',
    trigger_type VARCHAR(32) NOT NULL DEFAULT 'low_score' COMMENT 'low_score/serious_violation/repeated_violation',
    trigger_month VARCHAR(7) COMMENT '触发的月份',
    task_type VARCHAR(32) NOT NULL DEFAULT 'safety_training' COMMENT 'safety_training/rule_exam/mentor_drive/observation',
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending/in_progress/completed/cancelled',
    assigned_at DATETIME,
    started_at DATETIME,
    completed_at DATETIME,
    result_score DECIMAL(5,2) COMMENT '培训考核分数',
    result_note VARCHAR(512),
    created_by BIGINT DEFAULT 0 COMMENT '0表示系统自动触发',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_driver (driver_id),
    INDEX idx_status (status),
    INDEX idx_trigger_month (trigger_month)
) ENGINE=InnoDB COMMENT='驾驶员再培训任务表';
