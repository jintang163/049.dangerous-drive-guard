-- ============================================
-- 服务区智能停靠推荐模块扩展表
-- ============================================

-- ============================================
-- 1. 驾驶休息记录（连续驾驶时长追踪）
-- ============================================
CREATE TABLE IF NOT EXISTS driving_rest_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    driver_id BIGINT NOT NULL COMMENT '驾驶员ID',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    waybill_id BIGINT COMMENT '运单ID',
    record_date DATE NOT NULL COMMENT '记录日期',
    drive_start_time DATETIME NOT NULL COMMENT '本次驾驶开始时间',
    drive_end_time DATETIME COMMENT '本次驾驶结束时间',
    continuous_drive_minutes INT DEFAULT 0 COMMENT '连续驾驶时长(分钟)',
    rest_start_time DATETIME COMMENT '休息开始时间',
    rest_end_time DATETIME COMMENT '休息结束时间',
    rest_duration_minutes INT DEFAULT 0 COMMENT '休息时长(分钟)',
    rest_service_area_id BIGINT COMMENT '休息服务区ID',
    rest_service_area_name VARCHAR(128) COMMENT '休息服务区名称',
    status VARCHAR(20) DEFAULT 'driving' COMMENT 'driving/resting/completed',
    is_overtime TINYINT DEFAULT 0 COMMENT '是否超时驾驶',
    overtime_minutes INT DEFAULT 0 COMMENT '超时分钟数',
    check_in_time DATETIME COMMENT '签到时间',
    check_in_latitude DECIMAL(10,7) COMMENT '签到纬度',
    check_in_longitude DECIMAL(10,7) COMMENT '签到经度',
    check_out_time DATETIME COMMENT '签退时间',
    check_out_latitude DECIMAL(10,7) COMMENT '签退纬度',
    check_out_longitude DECIMAL(10,7) COMMENT '签退经度',
    min_rest_required INT DEFAULT 20 COMMENT '法定最低休息时长(分钟)',
    max_continuous_drive INT DEFAULT 240 COMMENT '法定最长连续驾驶(分钟)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_driver_date (driver_id, record_date DESC),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_waybill (waybill_id),
    INDEX idx_status (status),
    INDEX idx_overtime (is_overtime)
) ENGINE=InnoDB COMMENT='驾驶休息记录表';

-- ============================================
-- 2. 服务区实时状态（车位余量等）
-- ============================================
CREATE TABLE IF NOT EXISTS service_area_realtime_status (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    service_area_id BIGINT NOT NULL COMMENT '服务区ID',
    total_parking_spaces INT DEFAULT 0 COMMENT '总车位',
    available_parking_spaces INT DEFAULT 0 COMMENT '可用普通车位',
    total_danger_spaces INT DEFAULT 0 COMMENT '危化品总车位',
    available_danger_spaces INT DEFAULT 0 COMMENT '可用危化品车位',
    has_fuel TINYINT DEFAULT 1 COMMENT '是否有燃油',
    fuel_price_92 DECIMAL(5,2) COMMENT '92号油价',
    fuel_price_95 DECIMAL(5,2) COMMENT '95号油价',
    fuel_price_diesel DECIMAL(5,2) COMMENT '柴油价',
    has_charging TINYINT DEFAULT 0 COMMENT '是否有充电桩',
    charging_piles_total INT DEFAULT 0 COMMENT '充电桩总数',
    charging_piles_available INT DEFAULT 0 COMMENT '可用充电桩数',
    has_restaurant TINYINT DEFAULT 1 COMMENT '是否有餐厅',
    restaurant_rating DECIMAL(3,2) COMMENT '餐饮评分',
    restaurant_wait_minutes INT DEFAULT 0 COMMENT '餐饮等位时间(分钟)',
    has_hotel TINYINT DEFAULT 0 COMMENT '是否有住宿',
    hotel_rating DECIMAL(3,2) COMMENT '住宿评分',
    has_maintenance TINYINT DEFAULT 0 COMMENT '是否有维修',
    security_level TINYINT DEFAULT 3 COMMENT '安保等级 1-低 2-中 3-高 4-很高 5-极高',
    security_patrol_interval INT DEFAULT 30 COMMENT '安保巡逻间隔(分钟)',
    crowd_level TINYINT DEFAULT 2 COMMENT '人流量等级 1-很少 2-较少 3-一般 4-较多 5-很多',
    weather_condition VARCHAR(32) COMMENT '天气状况',
    update_time DATETIME NOT NULL COMMENT '数据更新时间',
    data_source VARCHAR(32) DEFAULT 'manual' COMMENT '数据来源 manual/api/sync',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_service_area (service_area_id),
    INDEX idx_security (security_level),
    INDEX idx_update_time (update_time)
) ENGINE=InnoDB COMMENT='服务区实时状态表';

-- ============================================
-- 3. 服务区评价
-- ============================================
CREATE TABLE IF NOT EXISTS service_area_reviews (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    service_area_id BIGINT NOT NULL COMMENT '服务区ID',
    driver_id BIGINT NOT NULL COMMENT '评价司机ID',
    driver_name VARCHAR(64) COMMENT '司机姓名',
    waybill_id BIGINT COMMENT '关联运单ID',
    vehicle_id BIGINT COMMENT '关联车辆ID',
    security_score TINYINT NOT NULL COMMENT '安全性评分 1-5',
    environment_score TINYINT COMMENT '环境评分 1-5',
    food_score TINYINT COMMENT '餐饮评分 1-5',
    service_score TINYINT COMMENT '服务评分 1-5',
    overall_score DECIMAL(3,2) NOT NULL COMMENT '综合评分',
    comment_text TEXT COMMENT '评价内容',
    tags JSON COMMENT '评价标签 ["安保好","餐饮棒","车位充足"]',
    images JSON COMMENT '评价图片URL数组',
    is_anonymous TINYINT DEFAULT 0 COMMENT '是否匿名',
    status TINYINT DEFAULT 1 COMMENT '1-正常 0-删除',
    check_in_record_id BIGINT COMMENT '关联签到记录ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_service_area (service_area_id, created_at DESC),
    INDEX idx_driver (driver_id),
    INDEX idx_security_score (security_score),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='服务区评价表';

-- ============================================
-- 4. 服务区推荐记录
-- ============================================
CREATE TABLE IF NOT EXISTS service_area_recommendations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    recommend_no VARCHAR(32) NOT NULL UNIQUE COMMENT '推荐编号',
    driver_id BIGINT NOT NULL COMMENT '司机ID',
    vehicle_id BIGINT NOT NULL COMMENT '车辆ID',
    waybill_id BIGINT COMMENT '运单ID',
    current_latitude DECIMAL(10,7) NOT NULL COMMENT '当前位置纬度',
    current_longitude DECIMAL(10,7) NOT NULL COMMENT '当前位置经度',
    current_address VARCHAR(512) COMMENT '当前位置地址',
    continuous_drive_minutes INT COMMENT '已连续驾驶分钟数',
    remaining_drive_minutes INT COMMENT '剩余可驾驶分钟数',
    fatigue_score DECIMAL(5,2) COMMENT '当前疲劳指数',
    hazard_class VARCHAR(8) COMMENT '运输危险品类别',
    recommend_reason VARCHAR(512) COMMENT '推荐理由摘要',
    recommended_service_area_id BIGINT COMMENT '推荐服务区ID',
    recommended_service_area_name VARCHAR(128) COMMENT '推荐服务区名称',
    distance_km DECIMAL(8,2) COMMENT '距离(公里)',
    estimated_arrival_minutes INT COMMENT '预计到达时间(分钟)',
    alternatives JSON COMMENT '备选服务区列表',
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending/accepted/rejected/arrived',
    accepted_at DATETIME COMMENT '接受推荐时间',
    arrived_at DATETIME COMMENT '到达服务区时间',
    dispatch_source VARCHAR(32) DEFAULT 'system' COMMENT '推荐来源 system/dispatcher/driver',
    dispatcher_id BIGINT COMMENT '调度员ID(如果是调度推荐)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_driver (driver_id, created_at DESC),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_waybill (waybill_id),
    INDEX idx_status (status),
    INDEX idx_recommend_area (recommended_service_area_id)
) ENGINE=InnoDB COMMENT='服务区推荐记录表';

-- ============================================
-- 初始化示例数据
-- ============================================

-- 服务区实时状态数据
INSERT INTO service_area_realtime_status 
(service_area_id, total_parking_spaces, available_parking_spaces, total_danger_spaces, available_danger_spaces,
 has_fuel, fuel_price_92, fuel_price_95, fuel_price_diesel, has_charging, charging_piles_total, charging_piles_available,
 has_restaurant, restaurant_rating, restaurant_wait_minutes, has_hotel, hotel_rating, has_maintenance,
 security_level, security_patrol_interval, crowd_level, weather_condition, update_time, data_source)
VALUES
(1, 200, 85, 15, 5, 1, 7.85, 8.32, 7.52, 1, 8, 3, 1, 4.2, 15, 1, 3.8, 1, 4, 25, 3, '晴', NOW(), 'sync'),
(2, 180, 62, 20, 8, 1, 7.82, 8.30, 7.50, 0, 0, 0, 1, 4.0, 10, 0, 0.0, 1, 5, 20, 2, '多云', NOW(), 'sync'),
(3, 150, 30, 0, 0, 1, 7.78, 8.25, 7.45, 1, 6, 2, 1, 3.5, 20, 0, 0.0, 0, 3, 30, 4, '阴', NOW(), 'manual'),
(4, 220, 120, 12, 4, 1, 7.80, 8.28, 7.48, 1, 10, 5, 1, 4.5, 8, 1, 4.2, 1, 4, 25, 2, '晴', NOW(), 'sync');

-- 驾驶休息记录示例
INSERT INTO driving_rest_records 
(driver_id, vehicle_id, waybill_id, record_date, drive_start_time, continuous_drive_minutes, 
 status, min_rest_required, max_continuous_drive)
VALUES
(3, 1, 1, CURDATE(), DATE_SUB(NOW(), INTERVAL 125 MINUTE), 125, 'driving', 20, 240),
(4, 2, 2, CURDATE(), DATE_SUB(NOW(), INTERVAL 200 MINUTE), 200, 'driving', 20, 240);

-- 服务区评价示例
INSERT INTO service_area_reviews 
(service_area_id, driver_id, driver_name, security_score, environment_score, food_score, service_score, 
 overall_score, comment_text, tags, is_anonymous, status)
VALUES
(1, 3, '驾驶员老孙', 5, 4, 4, 4, 4.25, '安保很到位，巡逻频繁，危化品专用区域管理规范，让人放心。餐厅菜品一般但价格公道。', 
 '["安保好","车位充足","餐饮一般"]', 0, 1),
(1, 4, '驾驶员老李', 4, 4, 5, 4, 4.25, '服务区很大，停车位充足，特别是危化品区域有专门的安保人员值守。餐厅味道不错。', 
 '["安保好","餐饮棒","环境整洁"]', 0, 1),
(2, 3, '驾驶员老孙', 5, 5, 3, 4, 4.25, '安保等级很高，24小时巡逻。就是餐饮选择少了点，整体还是很不错的。', 
 '["安保好","餐饮一般","服务好"]', 0, 1),
(4, 4, '驾驶员老李', 4, 5, 5, 5, 4.75, '非常好的服务区，设施新，安保到位，餐饮选择多。强烈推荐！', 
 '["安保好","餐饮棒","环境好","服务好"]', 0, 1);
