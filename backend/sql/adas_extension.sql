-- ============================================
-- ADAS 辅助驾驶提醒 - 数据库扩展
-- ============================================

CREATE TABLE IF NOT EXISTS adas_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL,
    close_following_min_dist_m DECIMAL(6,2) DEFAULT 5.00 COMMENT '跟车过近最小安全距离(米)',
    close_following_warn_dist_m DECIMAL(6,2) DEFAULT 10.00 COMMENT '跟车过近预警距离(米)',
    close_following_crit_dist_m DECIMAL(6,2) DEFAULT 3.00 COMMENT '跟车过近严重距离(米)',
    lane_departure_threshold_m DECIMAL(4,2) DEFAULT 0.30 COMMENT '车道偏离阈值(米)',
    forward_collision_ttc_warn_s DECIMAL(4,1) DEFAULT 3.0 COMMENT '前碰撞预警TTC(秒)',
    forward_collision_ttc_crit_s DECIMAL(4,1) DEFAULT 1.5 COMMENT '前碰撞严重TTC(秒)',
    frequency_window_minutes INT DEFAULT 5 COMMENT '频率统计窗口(分钟)',
    frequency_alert_threshold INT DEFAULT 6 COMMENT '频率告警阈值(次/窗口)',
    auto_decelerate_speed_kmh DECIMAL(5,2) DEFAULT 20.00 COMMENT '自动降速目标车速(km/h)',
    enable_close_following BOOLEAN DEFAULT TRUE COMMENT '是否启用跟车过近检测',
    enable_lane_departure BOOLEAN DEFAULT TRUE COMMENT '是否启用车道偏离检测',
    enable_forward_collision BOOLEAN DEFAULT TRUE COMMENT '是否启用前碰撞检测',
    enable_auto_decelerate BOOLEAN DEFAULT TRUE COMMENT '是否启用自动降速',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_vehicle (vehicle_id)
) ENGINE=InnoDB COMMENT='ADAS辅助驾驶配置表';

CREATE TABLE IF NOT EXISTS adas_alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_no VARCHAR(64) NOT NULL COMMENT '预警编号',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    driver_id BIGINT NOT NULL COMMENT '司机ID',
    waybill_id BIGINT COMMENT '运单ID',
    alert_type VARCHAR(32) NOT NULL COMMENT '预警类型: close_following/lane_departure/forward_collision/auto_decelerate',
    alert_level VARCHAR(16) NOT NULL COMMENT '预警等级: info/warning/critical',
    status VARCHAR(16) DEFAULT 'active' COMMENT '状态: active/resolved/escalated',
    trigger_value DECIMAL(10,2) COMMENT '触发值',
    threshold_value DECIMAL(10,2) COMMENT '阈值',
    following_distance_m DECIMAL(10,2) COMMENT '跟车距离(米)',
    vehicle_speed_kmh DECIMAL(5,2) COMMENT '车速(km/h)',
    lane_offset_m DECIMAL(6,3) COMMENT '车道偏移量(米)',
    departure_side VARCHAR(8) COMMENT '偏离方向: left/right',
    ttc_s DECIMAL(6,2) COMMENT '碰撞时间TTC(秒)',
    alert_message VARCHAR(512) COMMENT '预警消息',
    latitude DECIMAL(10,7) COMMENT '纬度',
    longitude DECIMAL(10,7) COMMENT '经度',
    suggested_action VARCHAR(256) COMMENT '建议操作',
    decelerate_triggered BOOLEAN DEFAULT FALSE COMMENT '是否触发自动降速',
    decelerate_value_kmh DECIMAL(5,2) COMMENT '降速目标值(km/h)',
    reported_to_center BOOLEAN DEFAULT FALSE COMMENT '是否已上报调度中心',
    driver_acknowledged BOOLEAN DEFAULT FALSE COMMENT '司机是否已确认',
    acknowledged_at DATETIME COMMENT '确认时间',
    resolved_at DATETIME COMMENT '解决时间',
    resolution_note VARCHAR(512) COMMENT '解决备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_alert_no (alert_no),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_driver (driver_id),
    INDEX idx_waybill (waybill_id),
    INDEX idx_type_level (alert_type, alert_level),
    INDEX idx_status (status),
    INDEX idx_vehicle_time (vehicle_id, created_at DESC)
) ENGINE=InnoDB COMMENT='ADAS辅助驾驶预警记录表';

CREATE TABLE IF NOT EXISTS adas_frequency_trackers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    driver_id BIGINT NOT NULL COMMENT '司机ID',
    window_start DATETIME NOT NULL COMMENT '统计窗口开始时间',
    window_end DATETIME NOT NULL COMMENT '统计窗口结束时间',
    close_following_count INT DEFAULT 0 COMMENT '跟车过近次数',
    lane_departure_count INT DEFAULT 0 COMMENT '车道偏离次数',
    total_alert_count INT DEFAULT 0 COMMENT '总预警次数',
    decelerate_triggered BOOLEAN DEFAULT FALSE COMMENT '是否触发降速',
    decelerate_value_kmh DECIMAL(5,2) COMMENT '降速目标值(km/h)',
    reported_to_center BOOLEAN DEFAULT FALSE COMMENT '是否已上报',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_driver (driver_id),
    INDEX idx_window (vehicle_id, window_start, window_end)
) ENGINE=InnoDB COMMENT='ADAS频率跟踪统计表';

-- DrivingScore 表增加跟车过近相关字段
ALTER TABLE driving_scores
    ADD COLUMN close_following_count INT DEFAULT 0 COMMENT '跟车过近次数',
    ADD COLUMN close_following_deduction DECIMAL(5,2) DEFAULT 0 COMMENT '跟车过近扣分';

-- 插入默认配置
INSERT INTO adas_configs (vehicle_id, close_following_min_dist_m, close_following_warn_dist_m, close_following_crit_dist_m,
    lane_departure_threshold_m, forward_collision_ttc_warn_s, forward_collision_ttc_crit_s,
    frequency_window_minutes, frequency_alert_threshold, auto_decelerate_speed_kmh)
VALUES (0, 5.00, 10.00, 3.00, 0.30, 3.0, 1.5, 5, 6, 20.00)
ON DUPLICATE KEY UPDATE vehicle_id = vehicle_id;
