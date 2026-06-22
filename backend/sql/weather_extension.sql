-- ============================================================
-- 天气与路面预警模块数据库扩展脚本
-- 适用: TiDB 7.x / MySQL 8.0+
-- 来源: weather_extension.sql
-- ============================================================

USE ddg_db;

-- ============================================================
-- 1. weather_warnings 表扩展字段
-- ============================================================

ALTER TABLE weather_warnings
    ADD COLUMN IF NOT EXISTS center_lat DECIMAL(10,7) COMMENT '预警影响区域中心点纬度' AFTER affected_area_polygon,
    ADD COLUMN IF NOT EXISTS center_lng DECIMAL(10,7) COMMENT '预警影响区域中心点经度' AFTER center_lat,
    ADD COLUMN IF NOT EXISTS trigger_operation_stop TINYINT DEFAULT 0 COMMENT '是否触发运营暂停 0-否 1-是' AFTER center_lng,
    ADD COLUMN IF NOT EXISTS speed_suggestion_kmh INT DEFAULT 0 COMMENT '建议车速(km/h),0表示停运' AFTER trigger_operation_stop,
    ADD COLUMN IF NOT EXISTS suggestion TEXT COMMENT '预警处置建议文案' AFTER speed_suggestion_kmh,
    ADD COLUMN IF NOT EXISTS updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at;

-- ============================================================
-- 2. 天气预警推送记录表
-- ============================================================

DROP TABLE IF EXISTS weather_push_records;

CREATE TABLE IF NOT EXISTS weather_push_records (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    push_id             VARCHAR(64) NOT NULL UNIQUE COMMENT '推送记录唯一ID',
    push_phase          VARCHAR(20) NOT NULL COMMENT '推送阶段 pre_departure出发前/en_route行驶中/emergency紧急',
    warning_id          BIGINT COMMENT '关联预警ID',
    warning_no          VARCHAR(64) COMMENT '预警编号',
    warning_type        VARCHAR(32) COMMENT '预警类型',
    warning_level       VARCHAR(8) COMMENT '预警等级 blue/yellow/orange/red',
    title               VARCHAR(256) NOT NULL COMMENT '推送标题',
    content             TEXT NOT NULL COMMENT '推送内容',
    target_type         VARCHAR(20) NOT NULL COMMENT '推送目标 vehicle车辆/driver司机/waybill运单/all全部',
    target_ids          JSON COMMENT '目标ID数组',
    waybill_id          BIGINT COMMENT '关联运单ID',
    waybill_no          VARCHAR(32) COMMENT '运单号',
    vehicle_id          BIGINT COMMENT '车辆ID',
    plate_number        VARCHAR(20) COMMENT '车牌号',
    driver_id           BIGINT COMMENT '司机ID',
    driver_name         VARCHAR(64) COMMENT '司机姓名',
    push_channels       JSON COMMENT '推送渠道数组 ["app","sms","push"]',
    status              VARCHAR(20) DEFAULT 'pending' COMMENT '状态 pending待发送/sending发送中/sent已发送/failed发送失败',
    success_count       INT DEFAULT 0 COMMENT '发送成功数量',
    fail_count          INT DEFAULT 0 COMMENT '发送失败数量',
    read_count          INT DEFAULT 0 COMMENT '已读数量',
    read_status         TINYINT DEFAULT 0 COMMENT '阅读状态 0未读 1已读',
    read_time           DATETIME COMMENT '阅读时间',
    driver_response     VARCHAR(32) COMMENT '司机响应 ack/ignore/call_support',
    response_time       DATETIME COMMENT '响应时间',
    response_note       TEXT COMMENT '响应备注',
    speed_suggestion_kmh INT COMMENT '建议车速(km/h)',
    segment_start_lat   DECIMAL(10,7) COMMENT '预警路段起点纬度',
    segment_start_lng   DECIMAL(10,7) COMMENT '预警路段起点经度',
    segment_end_lat     DECIMAL(10,7) COMMENT '预警路段终点纬度',
    segment_end_lng     DECIMAL(10,7) COMMENT '预警路段终点经度',
    segment_distance_km DECIMAL(8,2) COMMENT '预警路段距离(km)',
    operator_id         BIGINT COMMENT '操作人用户ID',
    operator_name       VARCHAR(64) COMMENT '操作人姓名',
    sent_at             DATETIME COMMENT '发送时间',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_push_phase (push_phase),
    INDEX idx_status (status),
    INDEX idx_target_type (target_type),
    INDEX idx_warning (warning_id),
    INDEX idx_waybill (waybill_id),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_driver (driver_id),
    INDEX idx_read_status (read_status),
    INDEX idx_created (created_at DESC),
    INDEX idx_operator (operator_id)
) ENGINE=InnoDB COMMENT='天气预警推送记录表';

-- ============================================================
-- 3. 历史天气数据表（用于事故天气回溯）
-- ============================================================

DROP TABLE IF EXISTS historical_weather;

CREATE TABLE IF NOT EXISTS historical_weather (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    latitude            DECIMAL(10,7) NOT NULL COMMENT '纬度',
    longitude           DECIMAL(10,7) NOT NULL COMMENT '经度',
    location_name       VARCHAR(256) COMMENT '地点名称',
    query_time          DATETIME NOT NULL COMMENT '查询的历史时间点',
    weather_condition   VARCHAR(64) COMMENT '天气状况(晴/多云/阴/小雨/中雨/大雨/暴雨/小雪/中雪/大雪/雾/霾/雷暴等)',
    temperature         DECIMAL(5,2) COMMENT '气温(摄氏度)',
    feels_like          DECIMAL(5,2) COMMENT '体感温度(摄氏度)',
    humidity            DECIMAL(5,2) COMMENT '相对湿度(%)',
    wind_speed          DECIMAL(8,2) COMMENT '风速(m/s)',
    wind_direction      INT COMMENT '风向角度 0-360度',
    visibility          DECIMAL(10,2) COMMENT '能见度(米)',
    pressure            DECIMAL(8,2) COMMENT '气压(hPa)',
    precipitation       DECIMAL(8,2) DEFAULT 0 COMMENT '降水量(mm)',
    precip_type         VARCHAR(16) COMMENT '降水类型 rain/snow/sleet',
    road_condition      VARCHAR(32) COMMENT '路面状况 dry/wet/icy/snowy',
    road_slippery       TINYINT DEFAULT 0 COMMENT '路面是否湿滑 0-否 1-是',
    uv_index            INT COMMENT '紫外线指数',
    warnings            JSON COMMENT '当时生效的预警列表',
    warning_type        VARCHAR(32) COMMENT '主要预警类型',
    warning_level       VARCHAR(8) COMMENT '主要预警等级',
    data_source         VARCHAR(32) COMMENT '数据来源 qweather/caiyun/amap/mock/cache',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_location_time (latitude, longitude, query_time),
    INDEX idx_query_time (query_time DESC),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB COMMENT='历史天气数据表';

-- ============================================================
-- 4. 极端天气运营暂停表
-- ============================================================

DROP TABLE IF EXISTS operation_suspensions;

CREATE TABLE IF NOT EXISTS operation_suspensions (
    id                     BIGINT PRIMARY KEY AUTO_INCREMENT,
    suspension_no         VARCHAR(64) NOT NULL UNIQUE COMMENT '暂停单号',
    trigger_type          VARCHAR(20) NOT NULL COMMENT '触发方式 automatic系统自动/manual人工',
    trigger_reason        VARCHAR(512) NOT NULL COMMENT '触发原因',
    trigger_warning_id    BIGINT COMMENT '触发的预警ID',
    weather_type          VARCHAR(32) COMMENT '触发天气类型 fog/rainstorm/typhoon/strong_wind/ice/snowstorm',
    visibility            DECIMAL(10,2) COMMENT '能见度(米)',
    wind_speed            DECIMAL(8,2) COMMENT '风速(m/s)',
    precipitation         DECIMAL(8,2) COMMENT '降水量(mm)',
    affected_region       VARCHAR(512) COMMENT '影响区域描述',
    center_lat            DECIMAL(10,7) NOT NULL COMMENT '中心点纬度',
    center_lng            DECIMAL(10,7) NOT NULL COMMENT '中心点经度',
    radius_km             DECIMAL(10,2) NOT NULL COMMENT '影响半径(公里)',
    affected_provinces    JSON COMMENT '影响省份数组',
    affected_cities       JSON COMMENT '影响城市数组',
    affected_polygon      JSON COMMENT '影响区域多边形坐标',
    affected_vehicle_ids  JSON COMMENT '受影响车辆ID数组',
    affected_waybill_ids  JSON COMMENT '受影响运单ID数组',
    status                VARCHAR(20) DEFAULT 'active' COMMENT '状态 active生效中/lifted已解除/expired已过期',
    suggested_speed       INT DEFAULT 0 COMMENT '建议车速 km/h,0表示停运',
    suspend_time          DATETIME COMMENT '暂停开始时间',
    resume_time           DATETIME COMMENT '恢复时间',
    expires_at            DATETIME COMMENT '预计到期时间',
    lift_reason           VARCHAR(512) COMMENT '解除原因',
    lifted_by             BIGINT COMMENT '解除操作人ID',
    lifted_at             DATETIME COMMENT '解除时间',
    suspended_waybill_count INT DEFAULT 0 COMMENT '受影响运单数',
    suspended_vehicle_count INT DEFAULT 0 COMMENT '受影响车辆数',
    operator_id           BIGINT COMMENT '操作人ID',
    operator_name         VARCHAR(64) COMMENT '操作人姓名',
    created_by            BIGINT COMMENT '创建人ID',
    auto_triggered        TINYINT DEFAULT 0 COMMENT '是否自动触发 0否 1是',
    remark                TEXT COMMENT '备注',
    created_at            DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_trigger_type (trigger_type),
    INDEX idx_weather_type (weather_type),
    INDEX idx_created (created_at DESC),
    INDEX idx_center (center_lat, center_lng),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB COMMENT='极端天气运营暂停记录表';

-- ============================================================
-- 5. 初始化基础示例数据 (可选，用于演示)
-- ============================================================

INSERT INTO weather_push_records (push_id, push_phase, warning_id, warning_type, warning_level, title, content, target_type, waybill_id, waybill_no, vehicle_id, plate_number, driver_id, driver_name, status, success_count, fail_count, read_count, operator_name, sent_at) VALUES
('WX-PUSH-20260620-001', 'pre_departure', 1, 'rainstorm', 'orange', '出发前天气预警 - 广东地区暴雨预警', '广东省部分地区有暴雨，建议车辆检查雨刮器，注意积水路段，建议降速30%行驶', 'all', NULL, NULL, NULL, NULL, NULL, NULL, 'sent', 156, 2, 89, '系统', '2026-06-20 07:30:00'),
('WX-PUSH-20260620-002', 'en_route', 2, 'fog', 'red', '紧急天气预警 - 南京大雾红色预警', '南京市江宁区能见度不足50米，请立即就近服务区停靠，开启雾灯双闪，等待天气好转', 'all', NULL, NULL, NULL, NULL, NULL, NULL, 'sent', 89, 1, 45, '调度员小赵', '2026-06-20 06:00:00');

INSERT INTO operation_suspensions (suspension_no, trigger_type, trigger_reason, trigger_warning_id, weather_type, visibility, wind_speed, affected_region, center_lat, center_lng, radius_km, affected_provinces, affected_cities, status, suggested_speed, created_by, operator_id, operator_name, auto_triggered, expires_at, suspended_vehicle_count, suspended_waybill_count) VALUES
('OPS-SUS-20260620-001', 'automatic', '能见度低于50米，系统自动触发停运', 2, 'fog', 35.5, 2.1, '江苏省南京市及周边高速公路网', 31.86, 118.76, 80, '["江苏省"]', '["南京市","镇江市","扬州市"]', 'lifted', 0, NULL, NULL, '系统', 1, '2026-06-20 12:00:00', 12, 8);

INSERT INTO historical_weather (latitude, longitude, location_name, query_time, weather_condition, temperature, feels_like, humidity, wind_speed, wind_direction, visibility, pressure, precipitation, precip_type, road_slippery, uv_index, warnings, data_source) VALUES
(22.5431, 114.0579, '广东省深圳市南山区', '2026-06-18 14:30:00', '暴雨', 25.6, 27.2, 92, 8.5, 180, 800, 1008.5, 18.5, 'rain', 1, 0, '["rainstorm_orange"]', 'mock');
