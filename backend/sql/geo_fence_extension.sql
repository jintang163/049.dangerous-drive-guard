-- ================================================
-- 押运员电子围栏告警扩展表
-- 功能：车辆偏离预设路线500米自动告警、押运员确认、3次自动上报
-- ================================================

-- 1. 电子围栏偏航告警表
DROP TABLE IF EXISTS `geo_fence_alerts`;
CREATE TABLE `geo_fence_alerts` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `alert_no` VARCHAR(32) NOT NULL COMMENT '告警编号',
    `vehicle_id` BIGINT NOT NULL COMMENT '车辆ID',
    `plate_number` VARCHAR(20) DEFAULT NULL COMMENT '车牌号',
    `driver_id` BIGINT DEFAULT NULL COMMENT '驾驶员ID',
    `driver_name` VARCHAR(64) DEFAULT NULL COMMENT '驾驶员姓名',
    `escort_id` BIGINT DEFAULT NULL COMMENT '押运员ID',
    `escort_name` VARCHAR(64) DEFAULT NULL COMMENT '押运员姓名',
    `waybill_id` BIGINT DEFAULT NULL COMMENT '运单ID',
    `waybill_no` VARCHAR(32) DEFAULT NULL COMMENT '运单编号',
    `route_plan_id` BIGINT DEFAULT NULL COMMENT '路线规划ID',
    `latitude` DOUBLE NOT NULL COMMENT '偏航位置-纬度',
    `longitude` DOUBLE NOT NULL COMMENT '偏航位置-经度',
    `address` VARCHAR(512) DEFAULT NULL COMMENT '偏航位置-地址',
    `distance_from_route_meters` INT NOT NULL DEFAULT 0 COMMENT '偏离预设路线距离(米)',
    `threshold_meters` INT NOT NULL DEFAULT 500 COMMENT '偏航阈值(米)',
    `alert_level` INT NOT NULL DEFAULT 2 COMMENT '告警级别：1信息 2警告 3危险',
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending待确认/confirmed已确认/escalated已上报/resolved已处理',
    `deviate_reason` VARCHAR(32) DEFAULT NULL COMMENT '偏航原因：detour绕路/deviate偏航',
    `confirm_note` VARCHAR(1024) DEFAULT NULL COMMENT '确认备注',
    `confirmed_by` BIGINT DEFAULT NULL COMMENT '确认人ID',
    `confirmed_role` VARCHAR(20) DEFAULT NULL COMMENT '确认人角色：driver/escort/dispatcher',
    `confirmed_at` DATETIME DEFAULT NULL COMMENT '确认时间',
    `reported_to_dispatch` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已上报调度',
    `reported_at` DATETIME DEFAULT NULL COMMENT '上报时间',
    `resolved_by` BIGINT DEFAULT NULL COMMENT '处理人ID',
    `resolved_note` VARCHAR(1024) DEFAULT NULL COMMENT '处理备注',
    `resolved_at` DATETIME DEFAULT NULL COMMENT '处理时间',
    `daily_deviate_count` INT NOT NULL DEFAULT 0 COMMENT '当日累计偏航次数',
    `nearest_route_point` JSON DEFAULT NULL COMMENT '最近的路线点坐标',
    `snapshot_url` VARCHAR(512) DEFAULT NULL COMMENT '现场快照URL',
    `popup_displayed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '弹窗是否已显示',
    `notified_escort` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已通知押运员',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_alert_no` (`alert_no`),
    KEY `idx_vehicle_status` (`vehicle_id`, `status`),
    KEY `idx_waybill` (`waybill_id`),
    KEY `idx_escort` (`escort_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_vehicle_date` (`vehicle_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子围栏偏航告警表';

-- 2. 电子围栏偏航确认日志表
DROP TABLE IF EXISTS `geo_fence_confirm_logs`;
CREATE TABLE `geo_fence_confirm_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `alert_id` BIGINT NOT NULL COMMENT '偏航告警ID',
    `alert_no` VARCHAR(32) NOT NULL COMMENT '告警编号',
    `vehicle_id` BIGINT NOT NULL COMMENT '车辆ID',
    `plate_number` VARCHAR(20) DEFAULT NULL COMMENT '车牌号',
    `waybill_id` BIGINT DEFAULT NULL COMMENT '运单ID',
    `waybill_no` VARCHAR(32) DEFAULT NULL COMMENT '运单编号',
    `confirm_type` VARCHAR(32) NOT NULL COMMENT '确认类型：detour绕路/deviate偏航',
    `reason_detail` VARCHAR(512) DEFAULT NULL COMMENT '具体原因',
    `note` VARCHAR(1024) DEFAULT NULL COMMENT '补充说明',
    `confirmed_by` BIGINT NOT NULL COMMENT '确认人ID',
    `confirmed_name` VARCHAR(64) DEFAULT NULL COMMENT '确认人姓名',
    `confirmed_role` VARCHAR(20) NOT NULL COMMENT '确认人角色：driver/escort/dispatcher',
    `latitude` DOUBLE DEFAULT NULL COMMENT '确认时位置-纬度',
    `longitude` DOUBLE DEFAULT NULL COMMENT '确认时位置-经度',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_alert_id` (`alert_id`),
    KEY `idx_vehicle_created` (`vehicle_id`, `created_at`),
    KEY `idx_confirmed_by` (`confirmed_by`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子围栏偏航确认日志表';

-- ================================================
-- 初始化示例数据
-- ================================================

-- 示例偏航告警
INSERT INTO `geo_fence_alerts` (
    `alert_no`, `vehicle_id`, `plate_number`, `driver_id`, `driver_name`,
    `escort_id`, `escort_name`, `waybill_id`, `waybill_no`, `route_plan_id`,
    `latitude`, `longitude`, `address`, `distance_from_route_meters`, `threshold_meters`,
    `alert_level`, `status`, `daily_deviate_count`, `created_at`
) VALUES
(
    'GFA202401150001', 1, '京A12345', 1, '张三',
    2, '李四', 1, 'WB202401150001', 1,
    39.9042, 116.4074, '北京市朝阳区建国门外大街1号', 750, 500,
    2, 'pending', 1, '2024-01-15 08:30:00'
),
(
    'GFA202401150002', 1, '京A12345', 1, '张三',
    2, '李四', 1, 'WB202401150001', 1,
    39.9242, 116.4274, '北京市朝阳区望京SOHO附近', 1200, 500,
    3, 'confirmed', 'deviate', 2, 'pending',
    0, NULL, 2, NULL, '2024-01-15 09:45:00'
),
(
    'GFA202401150003', 2, '沪B67890', 3, '王五',
    4, '赵六', 2, 'WB202401150002', 2,
    31.2304, 121.4737, '上海市黄浦区人民广场', 580, 500,
    2, 'resolved', 'detour', '前方事故绕行',
    4, 'escort', '2024-01-15 10:20:00',
    1, '2024-01-15 10:20:00', NULL, NULL, NULL,
    1, NULL, '2024-01-15 11:30:00',
    1, 1, NULL, '2024-01-15 10:00:00'
);

-- 示例确认日志
INSERT INTO `geo_fence_confirm_logs` (
    `alert_id`, `alert_no`, `vehicle_id`, `plate_number`, `waybill_id`, `waybill_no`,
    `confirm_type`, `reason_detail`, `note`, `confirmed_by`, `confirmed_name`, `confirmed_role`,
    `latitude`, `longitude`, `created_at`
) VALUES
(
    3, 'GFA202401150003', 2, '沪B67890', 2, 'WB202401150002',
    'detour', '前方高速公路追尾事故，交警要求绕行',
    '预计多行驶15公里，到达时间延迟30分钟',
    4, '赵六', 'escort',
    31.2304, 121.4737, '2024-01-15 10:20:00'
);
