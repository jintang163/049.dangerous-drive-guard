-- ============================================
-- 危险品运输安全监控系统 TiDB 数据库初始化脚本
-- 兼容 MySQL 8.0 / TiDB 7.x
-- ============================================

CREATE DATABASE IF NOT EXISTS ddg_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ddg_db;

-- ============================================
-- 1. 用户与组织管理
-- ============================================

CREATE TABLE IF NOT EXISTS organizations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL COMMENT '企业名称',
    license_no VARCHAR(64) COMMENT '经营许可证号',
    contact_person VARCHAR(64),
    contact_phone VARCHAR(20),
    address VARCHAR(256),
    type TINYINT DEFAULT 1 COMMENT '1-运输企业 2-货主 3-监管机构',
    status TINYINT DEFAULT 1 COMMENT '1-正常 0-停用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='企业组织表';

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL UNIQUE COMMENT '登录账号',
    password VARCHAR(128) NOT NULL COMMENT 'BCrypt加密密码',
    real_name VARCHAR(64) NOT NULL COMMENT '真实姓名',
    phone VARCHAR(20) COMMENT '手机号',
    email VARCHAR(128),
    role VARCHAR(20) NOT NULL COMMENT 'admin/dispatcher/driver/escort/viewer',
    org_id BIGINT COMMENT '所属企业ID',
    avatar_url VARCHAR(256),
    id_card VARCHAR(32) COMMENT '身份证号',
    license_no VARCHAR(64) COMMENT '驾驶证号',
    license_type VARCHAR(20) COMMENT '驾驶证类型 A1/A2/B2等',
    license_issue_date DATE,
    license_expire_date DATE,
    qualification_no VARCHAR(64) COMMENT '从业资格证号',
    status TINYINT DEFAULT 1 COMMENT '1-正常 0-停用',
    last_login_at DATETIME,
    last_login_ip VARCHAR(64),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_org_role (org_id, role),
    INDEX idx_phone (phone),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='用户表';

CREATE TABLE IF NOT EXISTS user_tokens (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    token VARCHAR(512) NOT NULL,
    expires_at DATETIME NOT NULL,
    device_id VARCHAR(128),
    device_type VARCHAR(32),
    ip VARCHAR(64),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_token (token),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB COMMENT='用户Token表';

-- ============================================
-- 2. 车辆管理
-- ============================================

CREATE TABLE IF NOT EXISTS vehicles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    plate_number VARCHAR(20) NOT NULL UNIQUE COMMENT '车牌号',
    vehicle_type VARCHAR(20) NOT NULL COMMENT 'tanker罐车/van厢式/flatbed平板',
    brand VARCHAR(64),
    model VARCHAR(64),
    color VARCHAR(32),
    vin VARCHAR(64) UNIQUE COMMENT '车架号',
    engine_no VARCHAR(64),
    load_weight DECIMAL(10,2) COMMENT '额定载重(吨)',
    load_volume DECIMAL(10,2) COMMENT '额定容积(立方)',
    length DECIMAL(5,2) COMMENT '车长(米)',
    width DECIMAL(5,2) COMMENT '车宽(米)',
    height DECIMAL(5,2) COMMENT '车高(米)',
    max_speed INT COMMENT '最高时速(km/h)',
    fuel_type VARCHAR(20),
    status VARCHAR(20) DEFAULT 'idle' COMMENT 'idle/running/loading/unloading/repair/offline',
    org_id BIGINT,
    driver_id BIGINT,
    escort_id BIGINT,
    device_id VARCHAR(128) COMMENT '车载终端ID',
    insurance_company VARCHAR(128),
    insurance_policy_no VARCHAR(64),
    insurance_expire_date DATE,
    annual_audit_date DATE,
    mileage DECIMAL(12,2) DEFAULT 0 COMMENT '累计里程(km)',
    danger_goods_license_no VARCHAR(64) COMMENT '危货运输证号',
    tank_material VARCHAR(64) COMMENT '罐体材质',
    tank_volume DECIMAL(10,2) COMMENT '罐体容积',
    has_electric_grounding TINYINT DEFAULT 0 COMMENT '是否有静电接地装置',
    has_fire_extinguisher TINYINT DEFAULT 1 COMMENT '是否有灭火器',
    has_emergency_cutoff TINYINT DEFAULT 0 COMMENT '是否有紧急切断阀',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_org_status (org_id, status),
    INDEX idx_driver (driver_id),
    INDEX idx_device (device_id)
) ENGINE=InnoDB COMMENT='车辆表';

CREATE TABLE IF NOT EXISTS vehicle_diagnostics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL,
    obd_data JSON COMMENT 'OBD原始数据',
    engine_rpm INT COMMENT '发动机转速',
    vehicle_speed DECIMAL(8,2) COMMENT '车速km/h',
    coolant_temp DECIMAL(5,2) COMMENT '冷却液温度',
    fuel_level DECIMAL(5,2) COMMENT '油量百分比',
    oil_pressure DECIMAL(8,2) COMMENT '机油压力',
    battery_voltage DECIMAL(5,2) COMMENT '蓄电池电压',
    tire_pressure_fl DECIMAL(5,2) COMMENT '左前胎压',
    tire_pressure_fr DECIMAL(5,2),
    tire_pressure_rl DECIMAL(5,2),
    tire_pressure_rr DECIMAL(5,2),
    brake_pad_wear_fl DECIMAL(5,2) COMMENT '左前刹车片磨损%',
    fault_codes JSON COMMENT '故障码数组',
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    report_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vehicle_time (vehicle_id, report_time DESC),
    INDEX idx_report_time (report_time)
) ENGINE=InnoDB COMMENT='车辆诊断数据表';

CREATE TABLE IF NOT EXISTS vehicle_fault_alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL,
    fault_code VARCHAR(32) NOT NULL,
    fault_level TINYINT NOT NULL COMMENT '1-提示 2-警告 3-严重',
    fault_desc VARCHAR(512),
    fault_suggestion VARCHAR(1024),
    first_report_time DATETIME NOT NULL,
    last_report_time DATETIME NOT NULL,
    report_count INT DEFAULT 1,
    status TINYINT DEFAULT 0 COMMENT '0-未处理 1-处理中 2-已处理',
    handled_by BIGINT,
    handled_at DATETIME,
    handle_note VARCHAR(1024),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vehicle_status (vehicle_id, status),
    INDEX idx_level (fault_level)
) ENGINE=InnoDB COMMENT='车辆故障告警表';

-- ============================================
-- 3. 危险品与电子运单
-- ============================================

CREATE TABLE IF NOT EXISTS dangerous_goods (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    un_code VARCHAR(16) NOT NULL UNIQUE COMMENT 'UN编号',
    cn_name VARCHAR(128) NOT NULL COMMENT '中文名称',
    en_name VARCHAR(256),
    hazard_class VARCHAR(8) NOT NULL COMMENT '危险类别 1爆炸/2气体/3易燃液体等',
    packing_group VARCHAR(4) COMMENT '包装类别 I/II/III',
    flash_point DECIMAL(6,2) COMMENT '闪点(℃)',
    boiling_point DECIMAL(8,2),
    density DECIMAL(8,4),
    solubility VARCHAR(256),
    toxicity VARCHAR(32),
    corrosivity VARCHAR(32),
    storage_condition VARCHAR(512) COMMENT '存储条件',
    transportation_requirement VARCHAR(1024) COMMENT '运输要求',
    emergency_measure VARCHAR(1024) COMMENT '应急措施',
    msds_url VARCHAR(256),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hazard_class (hazard_class)
) ENGINE=InnoDB COMMENT='危险品名录表';

CREATE TABLE IF NOT EXISTS waybills (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    waybill_no VARCHAR(32) NOT NULL UNIQUE COMMENT '运单号',
    order_no VARCHAR(64) COMMENT '订单号',
    shipper_org_id BIGINT NOT NULL COMMENT '托运方企业ID',
    carrier_org_id BIGINT NOT NULL COMMENT '承运方企业ID',
    receiver_org_id BIGINT COMMENT '收货方企业ID',
    vehicle_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    escort_id BIGINT,
    route_plan_id BIGINT COMMENT '路径规划ID',
    goods_id BIGINT COMMENT '危险品ID',
    goods_name VARCHAR(128) NOT NULL,
    goods_un_code VARCHAR(16),
    goods_hazard_class VARCHAR(8),
    goods_weight DECIMAL(12,4) NOT NULL COMMENT '货物重量(吨)',
    goods_volume DECIMAL(12,4) COMMENT '货物体积(立方)',
    package_type VARCHAR(32),
    package_count INT,
    origin_address VARCHAR(512) NOT NULL,
    origin_latitude DECIMAL(10,7) NOT NULL,
    origin_longitude DECIMAL(10,7) NOT NULL,
    dest_address VARCHAR(512) NOT NULL,
    dest_latitude DECIMAL(10,7) NOT NULL,
    dest_longitude DECIMAL(10,7) NOT NULL,
    planned_departure_time DATETIME,
    actual_departure_time DATETIME,
    planned_arrival_time DATETIME,
    actual_arrival_time DATETIME,
    status VARCHAR(20) DEFAULT 'created' COMMENT 'created/assigned/loading/in_transit/unloading/completed/cancelled/exception',
    total_distance DECIMAL(10,2) COMMENT '总里程(km)',
    transport_cost DECIMAL(12,2),
    risk_level TINYINT DEFAULT 2 COMMENT '1-低 2-中 3-高',
    approval_status TINYINT DEFAULT 0 COMMENT '0-待审核 1-已通过 2-已拒绝',
    approved_by BIGINT,
    approved_at DATETIME,
    approval_note VARCHAR(512),
    emergency_contact VARCHAR(64),
    emergency_phone VARCHAR(20),
    remark TEXT,
    blockchain_tx_hash VARCHAR(128) COMMENT '区块链交易哈希',
    blockchain_block_no BIGINT COMMENT '区块链区块号',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_carrier_status (carrier_org_id, status),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_driver (driver_id),
    INDEX idx_status (status),
    INDEX idx_departure (planned_departure_time)
) ENGINE=InnoDB COMMENT='电子运单表';

CREATE TABLE IF NOT EXISTS waybill_status_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    waybill_id BIGINT NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    operator_id BIGINT,
    operator_role VARCHAR(20),
    remark VARCHAR(512),
    extra_data JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_waybill_time (waybill_id, created_at DESC)
) ENGINE=InnoDB COMMENT='运单状态变更日志表';

-- ============================================
-- 4. 路径规划
-- ============================================

CREATE TABLE IF NOT EXISTS restricted_areas (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    area_type VARCHAR(32) NOT NULL COMMENT 'school/hospital/mall/tunnel/bridge/water_protection/height_limit/weight_limit',
    level TINYINT DEFAULT 2 COMMENT '危险等级 1-建议绕行 2-必须避开',
    province VARCHAR(64),
    city VARCHAR(64),
    district VARCHAR(64),
    address VARCHAR(512),
    boundary_polygon JSON NOT NULL COMMENT '区域边界GeoJSON多边形',
    center_latitude DECIMAL(10,7),
    center_longitude DECIMAL(10,7),
    radius DECIMAL(10,2) COMMENT '影响半径(米)',
    restrict_hazard_classes VARCHAR(128) COMMENT '限制的危险品类别,逗号分隔,空=全部限制',
    restrict_vehicle_types VARCHAR(128) COMMENT '限制的车辆类型',
    height_limit DECIMAL(5,2) COMMENT '限高(米)',
    weight_limit DECIMAL(8,2) COMMENT '限重(吨)',
    time_restriction JSON COMMENT '时间限制规则',
    effective_from DATETIME,
    effective_to DATETIME,
    source VARCHAR(32) COMMENT '数据来源 official/manual',
    status TINYINT DEFAULT 1 COMMENT '1-有效 0-无效',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type_level (area_type, level),
    INDEX idx_city (city),
    INDEX idx_status (status),
    SPATIAL INDEX idx_boundary (ST_GeomFromGeoJSON(boundary_polygon))
) ENGINE=InnoDB COMMENT='禁行/限行区域表';

CREATE TABLE IF NOT EXISTS route_plans (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    plan_no VARCHAR(32) NOT NULL UNIQUE,
    waybill_id BIGINT,
    vehicle_id BIGINT,
    driver_id BIGINT,
    strategy VARCHAR(20) NOT NULL COMMENT 'shortest/safest/economic/custom',
    origin_address VARCHAR(512) NOT NULL,
    origin_latitude DECIMAL(10,7) NOT NULL,
    origin_longitude DECIMAL(10,7) NOT NULL,
    dest_address VARCHAR(512) NOT NULL,
    dest_latitude DECIMAL(10,7) NOT NULL,
    dest_longitude DECIMAL(10,7) NOT NULL,
    waypoints JSON COMMENT '途经点 [{lat,lng,name}]',
    route_geometry JSON NOT NULL COMMENT '路径GeoJSON LineString',
    total_distance DECIMAL(12,2) NOT NULL COMMENT '总距离(米)',
    estimated_duration INT NOT NULL COMMENT '预计时长(秒)',
    expected_speed DECIMAL(8,2) COMMENT '预计平均速度',
    toll_fee DECIMAL(10,2) COMMENT '过路费估算',
    fuel_cost DECIMAL(10,2) COMMENT '油费估算',
    avoid_tunnels INT DEFAULT 0 COMMENT '避开隧道数量',
    avoid_bridges INT DEFAULT 0 COMMENT '避开桥梁数量',
    avoid_populated INT DEFAULT 0 COMMENT '避开人口密集区数量',
    avoid_water_protection INT DEFAULT 0,
    restricted_segments JSON COMMENT '禁行路段列表 [{start,end,type,reason,distance}]',
    alternative_routes JSON COMMENT '备选方案数组',
    safety_score DECIMAL(5,2) COMMENT '安全评分(0-100)',
    truck_mode TINYINT DEFAULT 1 COMMENT '1-货车模式 0-普通模式',
    vehicle_height DECIMAL(5,2),
    vehicle_weight DECIMAL(10,2),
    vehicle_width DECIMAL(5,2),
    hazard_class VARCHAR(8),
    weather_condition VARCHAR(32) COMMENT '规划时天气',
    status VARCHAR(20) DEFAULT 'active' COMMENT 'active/used/deprecated',
    created_by BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_waybill (waybill_id),
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_strategy (strategy),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB COMMENT='路径规划方案表';

CREATE TABLE IF NOT EXISTS service_areas (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    highway_name VARCHAR(64) COMMENT '所属高速',
    direction VARCHAR(16) COMMENT '上行/下行',
    province VARCHAR(64),
    city VARCHAR(64),
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    distance_from_start DECIMAL(10,2) COMMENT '距起点公里数',
    has_restaurant TINYINT DEFAULT 1,
    has_hotel TINYINT DEFAULT 0,
    has_fuel_station TINYINT DEFAULT 1,
    has_charging TINYINT DEFAULT 0,
    has_rest_room TINYINT DEFAULT 1,
    has_maintenance TINYINT DEFAULT 0,
    has_danger_goods_parking TINYINT DEFAULT 0 COMMENT '有危化品专用停车位',
    parking_spaces INT,
    danger_parking_spaces INT,
    phone VARCHAR(20),
    rating DECIMAL(3,2),
    status TINYINT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_city (city),
    SPATIAL INDEX idx_location (POINT(latitude, longitude))
) ENGINE=InnoDB COMMENT='高速服务区表';

-- ============================================
-- 5. 疲劳检测与驾驶行为
-- ============================================

CREATE TABLE IF NOT EXISTS fatigue_detection_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    waybill_id BIGINT,
    frame_image_url VARCHAR(512) COMMENT '快照图片URL',
    video_clip_url VARCHAR(512) COMMENT '视频片段URL',
    perclos_value DECIMAL(5,4) COMMENT 'PERCLOS值 眼睑闭合时间占比',
    eye_closed_ratio DECIMAL(5,4) COMMENT '眼睛闭合比例',
    blink_count INT COMMENT '眨眼次数',
    blink_frequency DECIMAL(6,2) COMMENT '眨眼频率(次/分钟)',
    yawn_count INT COMMENT '打哈欠次数',
    yawn_ratio DECIMAL(5,4) COMMENT '嘴巴张合比例',
    head_pitch DECIMAL(6,2) COMMENT '头部俯仰角(度)',
    head_yaw DECIMAL(6,2) COMMENT '头部偏航角(左右偏)',
    head_roll DECIMAL(6,2) COMMENT '头部翻滚角',
    gaze_deviation DECIMAL(5,4) COMMENT '视线偏离程度',
    phone_usage_detected TINYINT DEFAULT 0 COMMENT '是否检测到玩手机',
    smoking_detected TINYINT DEFAULT 0 COMMENT '是否检测到抽烟',
    seatbelt_detected TINYINT DEFAULT 1 COMMENT '是否系安全带',
    fatigue_score DECIMAL(6,2) NOT NULL COMMENT '综合疲劳指数0-100',
    fatigue_level VARCHAR(16) DEFAULT 'normal' COMMENT 'normal/warning/fatigue',
    is_alarm_triggered TINYINT DEFAULT 0 COMMENT '是否触发报警',
    alarm_type VARCHAR(32) COMMENT '报警类型 fatigue/yawn/gaze/phone/smoking/no_seatbelt',
    detection_time DATETIME NOT NULL COMMENT '检测时间',
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    vehicle_speed DECIMAL(8,2),
    edge_computed TINYINT DEFAULT 1 COMMENT '边缘端计算',
    network_status VARCHAR(16) COMMENT 'online/offline/weak',
    camera_position VARCHAR(20) DEFAULT 'center' COMMENT '摄像头位置 left/center/right/multi',
    left_frame_url VARCHAR(256) COMMENT '左摄像头帧URL',
    center_frame_url VARCHAR(256) COMMENT '中摄像头帧URL',
    right_frame_url VARCHAR(256) COMMENT '右摄像头帧URL',
    left_score DECIMAL(5,2) DEFAULT 0 COMMENT '左摄像头疲劳评分',
    center_score DECIMAL(5,2) DEFAULT 0 COMMENT '中摄像头疲劳评分',
    right_score DECIMAL(5,2) DEFAULT 0 COMMENT '右摄像头疲劳评分',
    fusion_method VARCHAR(32) COMMENT '融合方法 single_camera/weighted_fusion/center_fallback',
    fusion_confidence DECIMAL(5,4) DEFAULT 0 COMMENT '融合置信度 0-1',
    occlusion_detected TINYINT DEFAULT 0 COMMENT '是否检测到遮挡',
    backlit_detected TINYINT DEFAULT 0 COMMENT '是否检测到逆光',
    used_cameras VARCHAR(64) COMMENT '实际参与融合的摄像头列表逗号分隔 left,center,right',
    metrics JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vehicle_time (vehicle_id, detection_time DESC),
    INDEX idx_driver (driver_id, detection_time DESC),
    INDEX idx_waybill (waybill_id),
    INDEX idx_fatigue_level (fatigue_level),
    INDEX idx_alarm (is_alarm_triggered, detection_time DESC),
    INDEX idx_camera_position (camera_position),
    INDEX idx_fusion_method (fusion_method)
) ENGINE=InnoDB COMMENT='疲劳检测记录表';

CREATE TABLE IF NOT EXISTS fatigue_alarms (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alarm_no VARCHAR(32) NOT NULL UNIQUE,
    vehicle_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    waybill_id BIGINT,
    detection_record_id BIGINT,
    alarm_type VARCHAR(32) NOT NULL COMMENT 'fatigue_continuous/yawn_excessive/gaze_sustained/phone_usage/smoking/no_seatbelt',
    alarm_level TINYINT NOT NULL COMMENT '1-提醒 2-预警 3-严重',
    fatigue_score DECIMAL(6,2),
    continuous_fatigue_minutes INT COMMENT '连续疲劳分钟数',
    snap_image_url VARCHAR(512),
    video_clip_url VARCHAR(512),
    video_start_time DATETIME COMMENT '视频片段开始时间',
    video_end_time DATETIME,
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    location_address VARCHAR(512),
    vehicle_speed DECIMAL(8,2),
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending/processing/acknowledged/resolved/ignored',
    notify_driver_result VARCHAR(256) COMMENT '通知司机结果',
    dispatcher_id BIGINT COMMENT '处理调度员ID',
    handled_at DATETIME,
    handle_note VARCHAR(1024),
    handle_type VARCHAR(32) COMMENT 'voice_remind/dispatch_rest/intervene_legal/other',
    vehicle_informed TINYINT DEFAULT 0 COMMENT '是否已语音通知司机',
    escalated TINYINT DEFAULT 0 COMMENT '是否已升级上报',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle_status (vehicle_id, status),
    INDEX idx_driver (driver_id),
    INDEX idx_level (alarm_level),
    INDEX idx_status_created (status, created_at DESC)
) ENGINE=InnoDB COMMENT='疲劳报警事件表';

CREATE TABLE IF NOT EXISTS driving_scores (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    driver_id BIGINT NOT NULL,
    waybill_id BIGINT,
    vehicle_id BIGINT,
    trip_date DATE NOT NULL,
    total_score DECIMAL(6,2) NOT NULL DEFAULT 100 COMMENT '总分(满分100)',
    score_level VARCHAR(16) DEFAULT 'excellent' COMMENT 'excellent/good/normal/poor/danger',
    fatigue_score DECIMAL(5,2) COMMENT '疲劳驾驶扣分0-100分',
    fatigue_deduction DECIMAL(5,2) DEFAULT 0,
    overspeed_score DECIMAL(5,2),
    overspeed_count INT DEFAULT 0,
    overspeed_deduction DECIMAL(5,2) DEFAULT 0,
    sudden_brake_count INT DEFAULT 0,
    sudden_brake_deduction DECIMAL(5,2) DEFAULT 0,
    sudden_accel_count INT DEFAULT 0,
    sudden_accel_deduction DECIMAL(5,2) DEFAULT 0,
    sharp_turn_count INT DEFAULT 0,
    sharp_turn_deduction DECIMAL(5,2) DEFAULT 0,
    lane_deviation_count INT DEFAULT 0,
    lane_deviation_deduction DECIMAL(5,2) DEFAULT 0,
    phone_usage_count INT DEFAULT 0,
    phone_usage_deduction DECIMAL(5,2) DEFAULT 0,
    smoking_count INT DEFAULT 0,
    smoking_deduction DECIMAL(5,2) DEFAULT 0,
    seatbelt_violation_count INT DEFAULT 0,
    seatbelt_violation_deduction DECIMAL(5,2) DEFAULT 0,
    route_deviation_count INT DEFAULT 0,
    route_deviation_deduction DECIMAL(5,2) DEFAULT 0,
    fatigue_alarm_count INT DEFAULT 0,
    total_distance DECIMAL(10,2),
    driving_duration INT COMMENT '驾驶时长(分钟)',
    night_driving_duration INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_driver_trip (driver_id, trip_date, waybill_id),
    INDEX idx_driver_date (driver_id, trip_date DESC),
    INDEX idx_score_level (score_level)
) ENGINE=InnoDB COMMENT='驾驶评分表';

-- ============================================
-- 6. 电子押运与实时监控
-- ============================================

CREATE TABLE IF NOT EXISTS vehicle_tracks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id BIGINT NOT NULL,
    waybill_id BIGINT,
    driver_id BIGINT,
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    altitude DECIMAL(8,2) COMMENT '海拔',
    speed DECIMAL(8,2) COMMENT '车速km/h',
    direction INT COMMENT '方向角 0-360',
    satellite_count INT COMMENT '卫星数',
    hdop DECIMAL(5,2) COMMENT '水平精度因子',
    accuracy DECIMAL(8,2) COMMENT '定位精度(米)',
    gps_time DATETIME NOT NULL COMMENT 'GPS时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vehicle_gpstime (vehicle_id, gps_time DESC),
    INDEX idx_waybill (waybill_id),
    SPATIAL INDEX idx_point (POINT(latitude, longitude))
) ENGINE=InnoDB COMMENT='车辆轨迹表';

CREATE TABLE IF NOT EXISTS escort_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    waybill_id BIGINT NOT NULL,
    vehicle_id BIGINT,
    event_type VARCHAR(32) NOT NULL COMMENT 'start_loading/finish_loading/departure/stop_over/rest/resume/arrival_check/customs_check/police_check/accident/leak/other',
    event_level TINYINT DEFAULT 1 COMMENT '1-普通 2-重要 3-紧急',
    reporter_id BIGINT,
    reporter_role VARCHAR(20),
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    address VARCHAR(512),
    description TEXT,
    images JSON COMMENT '图片URL数组',
    video_url VARCHAR(512),
    handled_status TINYINT DEFAULT 0,
    handled_by BIGINT,
    handled_at DATETIME,
    handle_note VARCHAR(1024),
    event_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_waybill_time (waybill_id, event_time DESC),
    INDEX idx_type (event_type),
    INDEX idx_level (event_level)
) ENGINE=InnoDB COMMENT='电子押运事件表';

CREATE TABLE IF NOT EXISTS weather_warnings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    warning_id VARCHAR(64) NOT NULL UNIQUE COMMENT '官方预警ID',
    warning_type VARCHAR(32) NOT NULL COMMENT 'rainstorm/typhoon/snowstorm/fog/haze/thunder/high_temp/low_temp/strong_wind/sandstorm/hail',
    warning_level VARCHAR(8) NOT NULL COMMENT 'blue/yellow/orange/red',
    title VARCHAR(256) NOT NULL,
    content TEXT,
    affected_provinces JSON,
    affected_cities JSON,
    affected_area_polygon JSON COMMENT '影响区域',
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    publish_time DATETIME NOT NULL,
    source VARCHAR(64),
    related_waybill_count INT DEFAULT 0,
    related_vehicle_count INT DEFAULT 0,
    processed TINYINT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_type_level (warning_type, warning_level),
    INDEX idx_time (start_time)
) ENGINE=InnoDB COMMENT='天气预警表';

CREATE TABLE IF NOT EXISTS rescue_requests (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_no VARCHAR(32) NOT NULL UNIQUE,
    waybill_id BIGINT,
    vehicle_id BIGINT NOT NULL,
    driver_id BIGINT,
    sos_type VARCHAR(32) NOT NULL COMMENT 'accident/fire/leak/medical/robbery/breakdown/other',
    sos_level TINYINT NOT NULL COMMENT '1-一般 2-严重 3-紧急',
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    address VARCHAR(512),
    description TEXT,
    images JSON,
    caller_name VARCHAR(64),
    caller_phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending/dispatched/processing/completed/cancelled',
    assigned_resource_id BIGINT COMMENT '分配的救援资源ID',
    dispatcher_id BIGINT,
    dispatched_at DATETIME,
    arrived_at DATETIME,
    completed_at DATETIME,
    result_note VARCHAR(1024),
    injuries INT DEFAULT 0,
    deaths INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle (vehicle_id),
    INDEX idx_status (status),
    INDEX idx_level (sos_level)
) ENGINE=InnoDB COMMENT='紧急救援请求表';

CREATE TABLE IF NOT EXISTS rescue_resources (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    resource_type VARCHAR(32) NOT NULL COMMENT 'fire_truck/ambulance/tow_truck/hazmat_team/police',
    name VARCHAR(128) NOT NULL,
    org_name VARCHAR(128),
    contact_person VARCHAR(64),
    contact_phone VARCHAR(20) NOT NULL,
    province VARCHAR(64),
    city VARCHAR(64),
    district VARCHAR(64),
    address VARCHAR(512),
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    service_radius DECIMAL(10,2) COMMENT '服务半径(km)',
    capabilities JSON COMMENT '能力配置',
    response_time_minutes INT COMMENT '响应时间(分钟)',
    status TINYINT DEFAULT 1 COMMENT '1-可用 0-不可用',
    current_task_count INT DEFAULT 0,
    rating DECIMAL(3,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type_city (resource_type, city),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='救援资源表';

-- ============================================
-- 7. 区块链存证
-- ============================================

CREATE TABLE IF NOT EXISTS blockchain_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tx_hash VARCHAR(128) NOT NULL UNIQUE,
    block_no BIGINT,
    data_type VARCHAR(32) NOT NULL COMMENT 'waybill/route/fatigue_alarm/escort_event/score',
    data_id BIGINT NOT NULL COMMENT '关联业务数据ID',
    data_hash VARCHAR(128) NOT NULL COMMENT '业务数据SHA256哈希',
    previous_hash VARCHAR(128) COMMENT '上一条记录哈希(链式校验)',
    payload JSON COMMENT '存证核心数据',
    submitted_by BIGINT,
    submit_time DATETIME NOT NULL,
    confirmed_time DATETIME,
    chain_status VARCHAR(20) DEFAULT 'submitting' COMMENT 'submitting/confirmed/failed',
    error_msg VARCHAR(512),
    node_info VARCHAR(256),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_type_data (data_type, data_id),
    INDEX idx_status (chain_status),
    INDEX idx_tx (tx_hash)
) ENGINE=InnoDB COMMENT='区块链存证记录表';

-- ============================================
-- 8. 系统审计与操作日志
-- ============================================

CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT,
    username VARCHAR(64),
    module VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL,
    target_type VARCHAR(32),
    target_id BIGINT,
    detail TEXT,
    request_method VARCHAR(16),
    request_url VARCHAR(1024),
    request_params JSON,
    response_code INT,
    ip VARCHAR(64),
    user_agent VARCHAR(512),
    trace_id VARCHAR(64),
    duration_ms INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_time (user_id, created_at DESC),
    INDEX idx_module_action (module, action),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB COMMENT='操作审计日志表';

-- ============================================
-- 初始化基础数据
-- ============================================

INSERT INTO organizations (name, license_no, contact_person, contact_phone, type, status) VALUES
('危险品运输示范企业', 'WH-YUN-2024-0001', '张经理', '13800000001', 1, 1),
('化工原料有限公司', 'WH-HUO-2024-0002', '李主任', '13800000002', 2, 1),
('市交通运输管理局', '', '王科长', '13800000003', 3, 1);

INSERT INTO users (username, password, real_name, phone, role, org_id, id_card, license_no, license_type, qualification_no, status) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '系统管理员', '13900000000', 'admin', 1, '', '', '', '', 1),
('dispatcher01', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '调度员小赵', '13900000001', 'dispatcher', 1, '', '', '', '', 1),
('driver01', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '驾驶员老孙', '13900000101', 'driver', 1, '110101198501011234', '110101198501010001', 'A2', 'WH-JZ-2023-0001', 1),
('driver02', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '驾驶员老李', '13900000102', 'driver', 1, '110101198001011235', '110101198001010002', 'A2', 'WH-JZ-2023-0002', 1),
('escort01', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '押运员小钱', '13900000201', 'escort', 1, '110101198801011236', '', '', 'WH-YY-2023-0001', 1);

INSERT INTO dangerous_goods (un_code, cn_name, en_name, hazard_class, packing_group, flash_point, storage_condition, transportation_requirement, emergency_measure) VALUES
('UN1203', '汽油', 'Gasoline', '3', 'II', -40, '阴凉通风仓库,远离火源', '使用专用罐车,禁止烟火,静电接地', '立即切断火源,用泡沫灭火器,防止流入下水道'),
('UN1072', '压缩氧气', 'Oxygen, compressed', '2', '', NULL, '通风专用仓库,远离油脂', '直立固定,防止撞击,与易燃物隔离', '大量喷水冷却钢瓶,人员撤离到上风处'),
('UN1789', '盐酸', 'Hydrochloric acid', '8', 'II', NULL, '耐酸容器,通风', '耐腐蚀车辆,与碱类隔离,防雨', '用大量清水冲洗皮肤,用石灰中和地面'),
('UN3480', '锂离子电池', 'Lithium-ion batteries', '9', 'II', NULL, '通风干燥,温度15-25℃', '防火防震,与易燃物隔离,破损隔离', '大量水冲洗,隔离火源,收集泄漏物'),
('UN1830', '硫酸', 'Sulphuric acid', '8', 'II', NULL, '耐酸陶瓷坛,通风', '耐腐蚀车辆,禁止与水混装,防雨', '用干沙覆盖,禁止用直流水,防飞溅'),
('UN1993', '乙醇', 'Ethanol', '3', 'II', 12, '阴凉通风,密闭', '专用罐车,防静电,高温时段禁运', '泡沫灭火器,沙土,通风驱散蒸气');

INSERT INTO vehicles (plate_number, vehicle_type, brand, model, color, vin, engine_no, load_weight, load_volume, length, width, height, max_speed, fuel_type, status, org_id, driver_id, escort_id, device_id, mileage, danger_goods_license_no, tank_volume, has_electric_grounding, has_fire_extinguisher, has_emergency_cutoff) VALUES
('京A·危12345', 'tanker', '解放', 'J6P-460', '白色', 'LFWSRXRJ0NAB00001', 'CA6DM246012345', 32.5, 45.0, 12.0, 2.5, 3.9, 80, 'diesel', 'idle', 1, 3, 5, 'DEV-ANDROID-0001', 156800.0, 'WH-GL-2023-0001', 40.0, 1, 1, 1),
('京B·危67890', 'van', '东风', '天龙KL', '银灰色', 'LGAX4C431N1000002', 'DDi1352012345', 18.0, 65.0, 9.6, 2.4, 3.8, 85, 'diesel', 'running', 1, 4, NULL, 'DEV-ANDROID-0002', 89200.0, 'WH-GL-2023-0002', NULL, 1, 1, 0),
('京C·危11111', 'tanker', '重汽', '豪沃TH7', '红色', 'ZZ1327V466HE100003', 'MC13H51123456', 30.0, 42.0, 11.5, 2.5, 4.0, 75, 'diesel', 'idle', 1, NULL, NULL, 'DEV-LINUX-0003', 245600.0, 'WH-GL-2023-0003', 38.0, 1, 1, 1);

INSERT INTO restricted_areas (name, area_type, level, province, city, district, boundary_polygon, center_latitude, center_longitude, radius, restrict_hazard_classes) VALUES
('北京市第一中学', 'school', 2, '北京市', '北京市', '东城区', '{"type":"Polygon","coordinates":[[[116.4074,39.9042],[116.4084,39.9042],[116.4084,39.9052],[116.4074,39.9052],[116.4074,39.9042]]]}', 39.9047, 116.4079, 500, ''),
('协和医院', 'hospital', 2, '北京市', '北京市', '东城区', '{"type":"Polygon","coordinates":[[[116.4150,39.9100],[116.4170,39.9100],[116.4170,39.9120],[116.4150,39.9120],[116.4150,39.9100]]]}', 39.9110, 116.4160, 800, ''),
('王府井商圈', 'mall', 2, '北京市', '北京市', '东城区', '{"type":"Polygon","coordinates":[[[116.4080,39.9130],[116.4160,39.9130],[116.4160,39.9180],[116.4080,39.9180],[116.4080,39.9130]]]}', 39.9155, 116.4120, 1000, '3,8'),
('八达岭隧道', 'tunnel', 2, '北京市', '北京市', '延庆区', '{"type":"Polygon","coordinates":[[[116.0200,40.3600],[116.0300,40.3600],[116.0300,40.3650],[116.0200,40.3650],[116.0200,40.3600]]]}', 40.3625, 116.0250, 500, ''),
('密云水库水源保护区', 'water_protection', 2, '北京市', '北京市', '密云区', '{"type":"Polygon","coordinates":[[[116.8500,40.4500],[117.0500,40.4500],[117.0500,40.5500],[116.8500,40.5500],[116.8500,40.4500]]]}', 40.5000, 116.9500, 5000, '3,6,8'),
('国贸桥限高路段', 'height_limit', 1, '北京市', '北京市', '朝阳区', '{"type":"Polygon","coordinates":[[[116.4580,39.9080],[116.4620,39.9080],[116.4620,39.9110],[116.4580,39.9110],[116.4580,39.9080]]]}', 39.9095, 116.4600, 300, NULL),
('卢沟桥', 'bridge', 1, '北京市', '北京市', '丰台区', '{"type":"Polygon","coordinates":[[[116.2100,39.8500],[116.2200,39.8500],[116.2200,39.8550],[116.2100,39.8550],[116.2100,39.8500]]]}', 39.8525, 116.2150, 400, '1,3');

INSERT INTO service_areas (name, highway_name, direction, province, city, latitude, longitude, distance_from_start, has_danger_goods_parking, danger_parking_spaces, phone, status) VALUES
('马驹桥服务区', 'G45大广高速', '双向', '北京市', '北京市', 39.6950, 116.4650, 35.0, 1, 15, '010-60000001', 1),
('窦店服务区', 'G4京港澳高速', '双向', '北京市', '北京市', 39.6230, 116.1030, 45.0, 1, 20, '010-60000002', 1),
('香河服务区', 'G1京哈高速', '双向', '河北省', '廊坊市', 39.7550, 116.9210, 60.0, 0, 0, '0316-8000001', 1),
('涿州服务区', 'G4京港澳高速', '双向', '河北省', '保定市', 39.4790, 116.0010, 80.0, 1, 12, '0312-6000001', 1);

INSERT INTO rescue_resources (resource_type, name, org_name, contact_person, contact_phone, province, city, address, latitude, longitude, service_radius, status) VALUES
('fire_truck', '朝阳消防救援大队', '朝阳区消防救援支队', '王队长', '119', '北京市', '北京市', '朝阳区建国路88号', 39.9100, 116.4600, 30.0, 1),
('hazmat_team', '市危化品应急处置中心', '市应急管理局', '李主任', '010-12350', '北京市', '北京市', '丰台区丰台路58号', 39.8550, 116.2880, 50.0, 1),
('ambulance', '120急救中心城东分站', '市急救中心', '张站长', '120', '北京市', '北京市', '东城区东四十条27号', 39.9400, 116.4250, 25.0, 1),
('tow_truck', '华泰道路救援公司', '华泰汽车服务', '刘经理', '400-8080-000', '北京市', '北京市', '通州区新华大街100号', 39.9100, 116.6580, 40.0, 1);

-- ============================================================
-- 禁行区扩展表（已合并主迁移）
-- 来源: restricted_area_extension.sql
-- ============================================================
SOURCE restricted_area_extension.sql;

-- ============================================================
-- 实时动态重规划模块扩展表（已合并主迁移）
-- 来源: realtime_replan_extension.sql
-- ============================================================
SOURCE realtime_replan_extension.sql;

-- ============================================================
-- 电子押运模块扩展表（已合并主迁移）
-- 来源: escort_extension.sql
-- ============================================================
SOURCE escort_extension.sql;

-- ============================================================
-- 多摄像头疲劳检测扩展表
-- 来源: multi_camera_fatigue_extension.sql
-- ============================================================
SOURCE multi_camera_fatigue_extension.sql;

-- ============================================================
-- 天气与路面预警模块扩展表
-- 来源: weather_extension.sql
-- ============================================================
SOURCE weather_extension.sql;

-- ============================================================
-- 外部路况接入说明
-- ============================================================
-- Webhook 接口地址: POST /api/v1/traffic/webhook/import
-- 认证:           X-Webhook-Token: <配置文件 traffic.webhook_token>
-- 支持数据源:     高德开放平台交通事件 / 百度地图交通事件 / 交管12123 / 地方交警公开接口
--
-- 请求体示例 (JSON 数组批量 / 单条):
-- [
--   {
--     "event_type": "congestion",           -- congestion/accident/road_closure/construction
--     "event_level": 2,                     -- 1轻微 2中等 3严重
--     "title": "G6京藏高速清河桥段拥堵",
--     "description": "详细描述...",
--     "source": "amap_traffic_v4",          -- 来源标识: amap/baidu/traffic_12123/report
--     "road_name": "G6京藏高速",
--     "start_point": {"type":"Point","coordinates":[116.33,40.05]},
--     "end_point":   {"type":"Point","coordinates":[116.30,40.11]},
--     "affected_geometry": {                 -- GeoJSON Point/LineString/Polygon
--         "type":"LineString",
--         "coordinates":[[116.35,40.05],[116.33,40.08],[116.31,40.11]]
--     },
--     "center_lat": 40.08,  "center_lng": 116.33,
--     "affected_length_km": 6.5,
--     "congestion_level": 3,                 -- 1畅通 2缓行 3拥堵 4严重
--     "avg_speed_kmh": 18.5,
--     "duration_minutes": 25,
--     "started_at": "2026-06-21T08:15:00+08:00",
--     "expected_end_at": "2026-06-21T11:00:00+08:00",
--     "related_official_id": "AMAP-TE-20260621-000123"
--   }
-- ]
--
-- 响应格式:
-- {
--   "code": 0,
--   "message": "success",
--   "data": {
--     "accepted": 1,
--     "ignored": 0,
--     "errors": []
--   }
-- }

-- ============================================================
-- 路况事件常用查询SQL模板
-- ============================================================
-- 1. 获取所有活跃事件 (用于地图叠加 & 扫描器):
-- SELECT id,event_no,event_type,event_level,title,road_name,center_lat,center_lng,
--        congestion_level,avg_speed_kmh,affected_length_km,status
-- FROM traffic_events WHERE status='active'
--   AND (actual_end_at IS NULL OR actual_end_at > NOW());
--
-- 2. 查询某运单近30天重规划历史:
-- SELECT replan_no,waybill_no,vehicle_plate,trigger_type,trigger_reason,
--        distance_delta,duration_delta,status,created_at
-- FROM route_replan_records WHERE waybill_id=? AND created_at>=DATE_SUB(NOW(),INTERVAL 30 DAY)
-- ORDER BY created_at DESC;
--
-- 3. 重规划统计: 按月分组
-- SELECT DATE_FORMAT(created_at,'%Y-%m') month,
--        trigger_type, status, COUNT(*) cnt
-- FROM route_replan_records
-- GROUP BY DATE_FORMAT(created_at,'%Y-%m'), trigger_type, status
-- ORDER BY month DESC;

