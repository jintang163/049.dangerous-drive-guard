-- ============================================
-- 实时动态重规划模块扩展脚本
-- 新增表：实时路况事件表、重规划记录表、重规划候选路线表
-- ============================================

USE ddg_db;

-- 1. 实时路况事件表
CREATE TABLE IF NOT EXISTS traffic_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_no VARCHAR(64) NOT NULL UNIQUE COMMENT '事件编号 TE+时间戳',
    event_type VARCHAR(32) NOT NULL COMMENT '路况类型: congestion拥堵/accident事故/road_closure封路/construction施工',
    event_level TINYINT DEFAULT 1 COMMENT '事件等级: 1-轻微 2-中等 3-严重',
    title VARCHAR(256) NOT NULL COMMENT '事件标题',
    description VARCHAR(1024) COMMENT '事件描述',
    source VARCHAR(32) DEFAULT 'system' COMMENT '事件来源: system系统/official官方/report人工上报',
    road_name VARCHAR(128) COMMENT '受影响路段名称',
    start_point JSON COMMENT '起点坐标 GeoJSON Point',
    end_point JSON COMMENT '终点坐标 GeoJSON Point',
    affected_geometry JSON COMMENT '受影响区域 GeoJSON LineString/Polygon',
    center_lat DECIMAL(10,6) COMMENT '中心纬度',
    center_lng DECIMAL(10,6) COMMENT '中心经度',
    affected_length_km DECIMAL(8,2) DEFAULT 0 COMMENT '受影响路段长度(km)',
    congestion_level TINYINT COMMENT '拥堵等级: 1-畅通 2-缓行 3-拥堵 4-严重拥堵',
    avg_speed_kmh DECIMAL(6,2) COMMENT '平均车速 km/h',
    duration_minutes INT COMMENT '预计通行时长增加(分钟)',
    started_at DATETIME NOT NULL COMMENT '事件开始时间',
    expected_end_at DATETIME COMMENT '预计结束时间',
    actual_end_at DATETIME COMMENT '实际结束时间',
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态: active生效/expired已过期/resolved已解决',
    related_official_id VARCHAR(128) COMMENT '关联官方事件ID',
    extra_info JSON COMMENT '扩展信息(JSON)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_type (event_type),
    INDEX idx_level (event_level),
    INDEX idx_active_time (status, started_at, expected_end_at),
    INDEX idx_center (center_lat, center_lng)
) ENGINE=InnoDB COMMENT='实时路况事件表';

-- 2. 路线重规划记录表
CREATE TABLE IF NOT EXISTS route_replan_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    replan_no VARCHAR(64) NOT NULL UNIQUE COMMENT '重规划编号 RP+时间戳',
    waybill_id BIGINT NOT NULL COMMENT '运单ID',
    waybill_no VARCHAR(64) COMMENT '运单号',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    vehicle_plate VARCHAR(32) COMMENT '车牌号',
    driver_id BIGINT COMMENT '驾驶员ID',
    driver_name VARCHAR(64) COMMENT '驾驶员姓名',
    original_route_plan_id BIGINT COMMENT '原始路线规划ID',
    new_route_plan_id BIGINT COMMENT '新路线规划ID',
    trigger_type VARCHAR(32) NOT NULL COMMENT '触发类型: traffic路况事件/deviation偏航/restricted禁行区变更/manual手动',
    trigger_source_id BIGINT COMMENT '触发源ID(路况事件ID/偏航事件ID)',
    trigger_reason VARCHAR(256) NOT NULL COMMENT '触发原因描述',
    event_type VARCHAR(32) COMMENT '关联路况事件类型',
    current_lat DECIMAL(10,6) NOT NULL COMMENT '触发时车辆纬度',
    current_lng DECIMAL(10,6) NOT NULL COMMENT '触发时车辆经度',
    original_distance_remaining DECIMAL(8,2) COMMENT '原路线剩余距离(km)',
    original_duration_remaining INT COMMENT '原路线剩余时间(分钟)',
    new_distance_remaining DECIMAL(8,2) COMMENT '新路线剩余距离(km)',
    new_duration_remaining INT COMMENT '新路线剩余时间(分钟)',
    distance_delta DECIMAL(8,2) COMMENT '距离变化(km):正=增加,负=减少',
    duration_delta INT COMMENT '时间变化(分钟):正=增加,负=减少',
    avoided_traffic_ids JSON COMMENT '规避的路况事件ID列表',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending待确认/confirmed司机已确认/rejected司机已拒绝/auto_applied系统自动应用/cancelled取消',
    notified_at DATETIME COMMENT '推送通知时间',
    driver_confirm_at DATETIME COMMENT '司机确认时间',
    applied_at DATETIME COMMENT '应用新路线时间',
    confirm_note VARCHAR(256) COMMENT '司机确认备注',
    operator_id BIGINT COMMENT '调度员ID(手动触发时)',
    operator_name VARCHAR(64) COMMENT '调度员姓名',
    extra_info JSON COMMENT '扩展信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_waybill (waybill_id),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_status (status),
    INDEX idx_trigger (trigger_type, status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB COMMENT='路线重规划记录表';

-- 3. 重规划候选路线表
CREATE TABLE IF NOT EXISTS replan_candidate_routes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    replan_record_id BIGINT NOT NULL COMMENT '重规划记录ID',
    route_geometry JSON COMMENT '候选路线几何',
    route_path JSON COMMENT '路径点列表',
    strategy VARCHAR(20) NOT NULL COMMENT '策略: shortest/safest/economic/fastest',
    total_distance DECIMAL(8,2) COMMENT '总距离(km)',
    estimated_duration INT COMMENT '预估时长(分钟)',
    estimated_delay INT COMMENT '相较原路线延迟(分钟)',
    toll_fee DECIMAL(10,2) COMMENT '过路费',
    fuel_cost DECIMAL(10,2) COMMENT '油费估算',
    safety_score DECIMAL(5,2) COMMENT '安全评分',
    pass_traffic_events JSON COMMENT '途经路况事件ID列表',
    restricted_segments JSON COMMENT '禁行路段信息',
    is_recommended TINYINT DEFAULT 0 COMMENT '是否为推荐路线: 0否 1是',
    rank_order TINYINT DEFAULT 0 COMMENT '推荐排序',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_replan (replan_record_id)
) ENGINE=InnoDB COMMENT='重规划候选路线表';

-- ============================================
-- 预置测试路况事件数据
-- ============================================
INSERT IGNORE INTO traffic_events
(event_no, event_type, event_level, title, description, source, road_name,
 affected_geometry, center_lat, center_lng, affected_length_km, congestion_level,
 avg_speed_kmh, duration_minutes, started_at, expected_end_at, status)
VALUES
('TE20260621001', 'congestion', 2, '京藏高速清河桥段拥堵',
 '京藏高速出京方向清河桥路段车流密度大，行驶缓慢', 'system', 'G6京藏高速',
 JSON_OBJECT('type','LineString','coordinates',JSON_ARRAY(
   JSON_ARRAY(116.35, 40.05), JSON_ARRAY(116.33, 40.08), JSON_ARRAY(116.31, 40.11))),
 40.08, 116.33, 6.50, 3, 18.50, 25, NOW(), DATE_ADD(NOW(), INTERVAL 2 HOUR), 'active'),

('TE20260621002', 'accident', 3, '京港澳高速追尾事故',
 '京港澳高速进京方向K35+200处发生三车追尾事故，占用应急车道', 'official', 'G4京港澳高速',
 JSON_OBJECT('type','Point','coordinates',JSON_ARRAY(116.20, 39.85)),
 39.85, 116.20, 2.00, 4, 5.00, 60, DATE_SUB(NOW(), INTERVAL 30 MINUTE),
 DATE_ADD(NOW(), INTERVAL 3 HOUR), 'active'),

('TE20260621003', 'road_closure', 3, '北五环封闭施工',
 '北五环东段顾家庄桥至五元桥夜间封闭施工，禁止车辆通行', 'official', '北五环',
 JSON_OBJECT('type','LineString','coordinates',JSON_ARRAY(
   JSON_ARRAY(116.40, 40.01), JSON_ARRAY(116.45, 40.02), JSON_ARRAY(116.48, 40.01))),
 40.015, 116.44, 5.20, 4, 0.00, 999, NOW(), DATE_ADD(NOW(), INTERVAL 5 HOUR), 'active'),

('TE20260621004', 'construction', 2, '西三环道路维护施工',
 '西三环南向北新兴桥至航天桥路面维护，占用一条车道', 'system', '西三环',
 JSON_OBJECT('type','LineString','coordinates',JSON_ARRAY(
   JSON_ARRAY(116.30, 39.90), JSON_ARRAY(116.30, 39.92), JSON_ARRAY(116.31, 39.94))),
 39.92, 116.305, 4.50, 2, 28.00, 20, DATE_SUB(NOW(), INTERVAL 1 HOUR),
 DATE_ADD(NOW(), INTERVAL 12 HOUR), 'active');
