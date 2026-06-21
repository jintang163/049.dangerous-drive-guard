-- ============================================================
-- 电子押运模块扩展表
-- 危险品电子押运 - 无需随车押运员，AI远程电子押运
-- ============================================================

USE ddg_db;

-- ============================================================
-- 1. 押运任务排班表
-- 调度员分配押运员负责多辆车
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_shifts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    shift_no VARCHAR(32) NOT NULL UNIQUE COMMENT '排班编号',
    escort_id BIGINT NOT NULL COMMENT '押运员ID',
    escort_name VARCHAR(64) COMMENT '押运员姓名',
    dispatcher_id BIGINT NOT NULL COMMENT '调度员ID',
    dispatcher_name VARCHAR(64) COMMENT '调度员姓名',
    vehicle_ids VARCHAR(512) COMMENT '负责的车辆ID列表，逗号分隔',
    waybill_ids VARCHAR(512) COMMENT '负责的运单ID列表，逗号分隔',
    scheduled_start DATETIME COMMENT '计划开始时间',
    scheduled_end DATETIME COMMENT '计划结束时间',
    actual_start DATETIME COMMENT '实际开始时间',
    actual_end DATETIME COMMENT '实际结束时间',
    status VARCHAR(20) DEFAULT 'scheduled' COMMENT 'scheduled/active/completed/cancelled',
    remark VARCHAR(512) COMMENT '备注',
    max_concurrent INT DEFAULT 5 COMMENT '最大同时监控车辆数',
    polling_interval INT DEFAULT 30 COMMENT '轮询间隔秒数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_escort_status (escort_id, status),
    INDEX idx_dispatcher (dispatcher_id),
    INDEX idx_scheduled (scheduled_start, scheduled_end)
) ENGINE=InnoDB COMMENT='电子押运排班表';

-- ============================================================
-- 2. 押运车辆分配明细表
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_vehicle_assignments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    shift_id BIGINT NOT NULL COMMENT '排班ID',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    plate_number VARCHAR(20) COMMENT '车牌号',
    waybill_id BIGINT COMMENT '运单ID',
    waybill_no VARCHAR(32) COMMENT '运单号',
    priority INT DEFAULT 1 COMMENT '优先级 1-普通 2-重要 3-紧急',
    assigned_by BIGINT COMMENT '分配人ID',
    assigned_at DATETIME COMMENT '分配时间',
    is_active TINYINT DEFAULT 1 COMMENT '是否生效',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shift (shift_id),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_waybill (waybill_id),
    INDEX idx_active (is_active)
) ENGINE=InnoDB COMMENT='押运车辆分配明细表';

-- ============================================================
-- 3. 紧急报警表（一键报警弹窗）
-- 司机按驾驶室紧急按钮，管理端强制弹窗
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_sos_alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_no VARCHAR(32) NOT NULL UNIQUE COMMENT '报警编号',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    plate_number VARCHAR(20) COMMENT '车牌号',
    driver_id BIGINT COMMENT '司机ID',
    driver_name VARCHAR(64) COMMENT '司机姓名',
    waybill_id BIGINT COMMENT '运单ID',
    waybill_no VARCHAR(32) COMMENT '运单号',
    alert_type VARCHAR(32) NOT NULL COMMENT '报警类型 emergency_button/leak/fire/accident/other',
    alert_level TINYINT DEFAULT 3 COMMENT '1-提示 2-预警 3-严重',
    latitude DECIMAL(10,7) COMMENT '纬度',
    longitude DECIMAL(10,7) COMMENT '经度',
    address VARCHAR(512) COMMENT '地址',
    description TEXT COMMENT '描述',
    snapshot_url VARCHAR(512) COMMENT '现场快照URL',
    video_clip_url VARCHAR(512) COMMENT '视频片段URL',
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending/processing/resolved/ignored',
    handled_by BIGINT COMMENT '处理人ID',
    handler_name VARCHAR(64) COMMENT '处理人姓名',
    handled_at DATETIME COMMENT '处理时间',
    handle_note VARCHAR(1024) COMMENT '处理备注',
    handle_type VARCHAR(32) COMMENT '处理类型 voice_intercom/dispatch_rescue/escalate/other',
    notified TINYINT DEFAULT 0 COMMENT '是否已通知相关人员',
    popup_displayed TINYINT DEFAULT 0 COMMENT '管理端是否已弹窗显示',
    acked_at DATETIME COMMENT '确认时间',
    escort_id BIGINT COMMENT '负责押运员ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle_status (vehicle_id, status),
    INDEX idx_escort (escort_id),
    INDEX idx_level (alert_level),
    INDEX idx_status_created (status, created_at DESC),
    INDEX idx_waybill (waybill_id)
) ENGINE=InnoDB COMMENT='押运紧急报警表';

-- ============================================================
-- 4. 视频录像记录表
-- 押运记录云端存储90天
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_video_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    record_no VARCHAR(32) NOT NULL UNIQUE COMMENT '录像编号',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    plate_number VARCHAR(20) COMMENT '车牌号',
    waybill_id BIGINT COMMENT '运单ID',
    waybill_no VARCHAR(32) COMMENT '运单号',
    record_type VARCHAR(20) COMMENT 'scheduled定时轮询/alarm报警触发/manual人工录制',
    video_url VARCHAR(512) NOT NULL COMMENT '视频存储URL',
    snapshot_url VARCHAR(512) COMMENT '封面快照URL',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    duration INT COMMENT '时长(秒)',
    latitude DECIMAL(10,7) COMMENT '录制时纬度',
    longitude DECIMAL(10,7) COMMENT '录制时经度',
    trigger_reason VARCHAR(256) COMMENT '触发原因',
    alert_id BIGINT COMMENT '关联报警ID',
    viewed_count INT DEFAULT 0 COMMENT '查看次数',
    expire_at DATETIME COMMENT '过期时间(云端存储90天)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle_time (vehicle_id, start_time DESC),
    INDEX idx_waybill (waybill_id),
    INDEX idx_record_type (record_type),
    INDEX idx_expire (expire_at),
    INDEX idx_alert (alert_id)
) ENGINE=InnoDB COMMENT='押运视频录像记录表';

-- ============================================================
-- 5. 对讲/喊话指令日志表
-- 支持喊话指令（"前方检查点请减速"）
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_intercom_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    plate_number VARCHAR(20) COMMENT '车牌号',
    sender_id BIGINT NOT NULL COMMENT '发送人ID',
    sender_name VARCHAR(64) COMMENT '发送人姓名',
    sender_role VARCHAR(20) COMMENT '发送人角色 admin/dispatcher/escort',
    message_type VARCHAR(20) DEFAULT 'text' COMMENT 'text/voice',
    content TEXT NOT NULL COMMENT '消息内容/语音转文本',
    audio_url VARCHAR(512) COMMENT '语音文件URL',
    priority INT DEFAULT 1 COMMENT '优先级 1-普通 2-重要 3-紧急',
    delivered TINYINT DEFAULT 0 COMMENT '是否已送达',
    delivered_at DATETIME COMMENT '送达时间',
    acked TINYINT DEFAULT 0 COMMENT '司机是否已确认',
    acked_at DATETIME COMMENT '确认时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vehicle_time (vehicle_id, created_at DESC),
    INDEX idx_sender (sender_id),
    INDEX idx_priority (priority)
) ENGINE=InnoDB COMMENT='押运对讲喊话日志表';

-- ============================================================
-- 6. 视频轮询会话表
-- 管理端可轮询查看车内画面（每车30秒）
-- ============================================================
CREATE TABLE IF NOT EXISTS escort_polling_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_no VARCHAR(64) NOT NULL UNIQUE COMMENT '轮询会话编号',
    escort_id BIGINT NOT NULL COMMENT '押运员ID',
    escort_name VARCHAR(64) COMMENT '押运员姓名',
    shift_id BIGINT COMMENT '关联排班ID',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    vehicles TEXT COMMENT '轮询的车辆列表JSON',
    polling_count INT DEFAULT 0 COMMENT '轮询次数',
    status VARCHAR(20) DEFAULT 'active' COMMENT 'active/completed/interrupted',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_escort_status (escort_id, status),
    INDEX idx_shift (shift_id)
) ENGINE=InnoDB COMMENT='视频轮询会话表';

-- ============================================================
-- 初始化测试数据
-- ============================================================

INSERT INTO escort_shifts (shift_no, escort_id, escort_name, dispatcher_id, dispatcher_name, vehicle_ids, waybill_ids, scheduled_start, scheduled_end, status, remark, max_concurrent, polling_interval) VALUES
('ES20260621001', 5, '小钱', 2, '调度员小赵', '1,2', '', '2026-06-21 08:00:00', '2026-06-21 18:00:00', 'active', '白班押运，负责京A和京B两辆车', 5, 30),
('ES20260621002', 5, '小钱', 2, '调度员小赵', '3', '', '2026-06-21 14:00:00', '2026-06-21 22:00:00', 'scheduled', '中班押运', 3, 30);

INSERT INTO escort_vehicle_assignments (shift_id, vehicle_id, plate_number, priority, assigned_by, assigned_at, is_active) VALUES
(1, 1, '京A·危12345', 1, 2, NOW(), 1),
(1, 2, '京B·危67890', 2, 2, NOW(), 1),
(2, 3, '京C·危11111', 1, 2, NOW(), 1);

INSERT INTO escort_sos_alerts (alert_no, vehicle_id, plate_number, driver_id, driver_name, alert_type, alert_level, latitude, longitude, address, description, status, escort_id) VALUES
('SOS20260621001', 1, '京A·危12345', 3, '驾驶员老孙', 'emergency_button', 3, 39.9042, 116.4074, '北京市东城区长安街附近', '司机按下驾驶室紧急按钮，疑似遇到突发状况，请立即核查！', 'pending', 5),
('SOS20260621002', 2, '京B·危67890', 4, '驾驶员老李', 'leak_suspected', 2, 39.9142, 116.4174, '北京市朝阳区建国路附近', '押运员AI检测到罐体周围有可疑液体痕迹，疑似轻微泄漏', 'processing', 5);

INSERT INTO escort_video_records (record_no, vehicle_id, plate_number, record_type, video_url, snapshot_url, start_time, end_time, duration, latitude, longitude, trigger_reason, viewed_count, expire_at) VALUES
('VR20260621001', 1, '京A·危12345', 'scheduled', '/videos/escort/VR20260621001.mp4', '/videos/escort/VR20260621001.jpg', '2026-06-21 09:00:00', '2026-06-21 09:00:30', 30, 39.9042, 116.4074, '定时轮询录制', 5, DATE_ADD(NOW(), INTERVAL 90 DAY)),
('VR20260621002', 1, '京A·危12345', 'alarm', '/videos/escort/VR20260621002.mp4', '/videos/escort/VR20260621002.jpg', '2026-06-21 10:15:00', '2026-06-21 10:16:00', 60, 39.9142, 116.4174, 'SOS报警触发自动录制', 12, DATE_ADD(NOW(), INTERVAL 90 DAY)),
('VR20260621003', 2, '京B·危67890', 'scheduled', '/videos/escort/VR20260621003.mp4', '/videos/escort/VR20260621003.jpg', '2026-06-21 08:30:00', '2026-06-21 08:30:30', 30, 39.8550, 116.2880, '定时轮询录制', 3, DATE_ADD(NOW(), INTERVAL 90 DAY)),
('VR20260621004', 2, '京B·危67890', 'manual', '/videos/escort/VR20260621004.mp4', '/videos/escort/VR20260621004.jpg', '2026-06-21 09:45:00', '2026-06-21 09:46:30', 90, 39.8600, 116.2900, '押运员手动录制检查', 8, DATE_ADD(NOW(), INTERVAL 90 DAY));

INSERT INTO escort_intercom_logs (vehicle_id, plate_number, sender_id, sender_name, sender_role, message_type, content, priority, delivered, delivered_at, acked) VALUES
(1, '京A·危12345', 2, '调度员小赵', 'dispatcher', 'text', '前方检查点请减速，注意安全驾驶', 2, 1, '2026-06-21 09:15:00', 1),
(1, '京A·危12345', 5, '小钱', 'escort', 'text', '请保持当前车道行驶，即将进入隧道', 1, 1, '2026-06-21 09:30:00', 1),
(2, '京B·危67890', 5, '小钱', 'escort', 'text', '检测到罐体附近有可疑情况，请停车检查！', 3, 1, '2026-06-21 10:20:00', 0);
