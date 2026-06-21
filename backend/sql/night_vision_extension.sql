-- ============================================================
-- 夜视增强与红外补光 数据库扩展
-- 原夜间疲劳识别增加红外补光灯，低光照下自动开启
-- 图像增强算法提升暗部细节，夜间人脸检测准确率
-- ============================================================

-- ------------------------------------------------------------
-- 1. 夜视增强配置表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `night_vision_configs`;
CREATE TABLE `night_vision_configs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `vehicle_id` BIGINT NOT NULL COMMENT '车辆ID',
    `device_id` VARCHAR(64) DEFAULT '' COMMENT '设备编号',

    -- 红外补光配置
    `infrared_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '红外补光功能总开关',
    `infrared_auto_mode` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '自动模式(根据光照自动开关)',
    `infrared_manual_on` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '手动强制开启',
    `infrared_intensity` TINYINT UNSIGNED NOT NULL DEFAULT 50 COMMENT '红外灯强度 0-100',
    `infrared_intensity_auto` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '自动调节强度',
    `low_light_threshold_lux` INT NOT NULL DEFAULT 50 COMMENT '低光照阈值(勒克斯)，低于此值自动开启红外',
    `high_light_threshold_lux` INT NOT NULL DEFAULT 200 COMMENT '高光照阈值(勒克斯)，高于此值自动关闭红外',

    -- 图像增强配置
    `enhancement_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '图像增强总开关',
    `enhance_mode` VARCHAR(32) NOT NULL DEFAULT 'auto' COMMENT '增强模式: auto/night/infrared/manual',
    `gamma_value` DECIMAL(4,3) NOT NULL DEFAULT 1.200 COMMENT '伽马校正值',
    `brightness_boost` TINYINT NOT NULL DEFAULT 30 COMMENT '亮度提升量 -100 ~ 100',
    `contrast_boost` TINYINT NOT NULL DEFAULT 20 COMMENT '对比度提升量 -100 ~ 100',
    `histogram_equalization` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '直方图均衡化',
    `clahe_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'CLAHE 限制对比度自适应直方图均衡化',
    `denoise_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '降噪开关',
    `denoise_strength` TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '降噪强度 1-5',
    `sharpen_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '锐化开关',
    `sharpen_strength` TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '锐化强度 1-5',

    -- 夜间检测相关
    `night_mode_auto` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '自动夜间模式',
    `night_start_hour` TINYINT NOT NULL DEFAULT 19 COMMENT '夜间模式开始时间(小时)',
    `night_end_hour` TINYINT NOT NULL DEFAULT 6 COMMENT '夜间模式结束时间(小时)',
    `low_light_face_detect` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '低光照人脸检测优化',
    `min_face_confidence_night` DECIMAL(5,4) NOT NULL DEFAULT 0.4000 COMMENT '夜间人脸检测最低置信度',

    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_vehicle` (`vehicle_id`),
    KEY `idx_device` (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='夜视增强配置表';

-- 示例配置数据
INSERT INTO `night_vision_configs`
    (vehicle_id, device_id, infrared_enabled, infrared_auto_mode, infrared_intensity,
     low_light_threshold_lux, high_light_threshold_lux,
     enhancement_enabled, enhance_mode, gamma_value, brightness_boost, contrast_boost,
     histogram_equalization, clahe_enabled, denoise_enabled, denoise_strength,
     night_mode_auto, night_start_hour, night_end_hour,
     low_light_face_detect, min_face_confidence_night)
VALUES
    (1, 'DEV-NIGHT-001', 1, 1, 60, 50, 200, 1, 'auto', 1.200, 30, 20, 1, 1, 1, 3, 1, 19, 6, 1, 0.4000),
    (2, 'DEV-NIGHT-002', 1, 1, 55, 60, 180, 1, 'auto', 1.150, 25, 15, 1, 1, 1, 4, 1, 20, 6, 1, 0.4500),
    (3, 'DEV-NIGHT-003', 1, 0, 70, 40, 150, 1, 'night', 1.300, 40, 25, 1, 1, 1, 2, 1, 19, 7, 1, 0.3800);

-- ------------------------------------------------------------
-- 2. 红外补光灯状态日志表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `infrared_light_logs`;
CREATE TABLE `infrared_light_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `vehicle_id` BIGINT NOT NULL COMMENT '车辆ID',
    `driver_id` BIGINT DEFAULT NULL COMMENT '司机ID',
    `device_id` VARCHAR(64) DEFAULT '' COMMENT '设备编号',

    `action` VARCHAR(32) NOT NULL COMMENT '动作类型: turn_on/turn_off/intensity_change/auto_trigger/manual_trigger',
    `trigger_type` VARCHAR(32) NOT NULL DEFAULT 'auto' COMMENT '触发方式: auto/manual/system',
    `light_on` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '补光灯状态',
    `intensity_before` TINYINT UNSIGNED DEFAULT NULL COMMENT '变化前强度',
    `intensity_after` TINYINT UNSIGNED DEFAULT NULL COMMENT '变化后强度',
    `light_level_lux` INT DEFAULT NULL COMMENT '环境光照强度(勒克斯)',

    `reason` VARCHAR(256) DEFAULT '' COMMENT '触发原因',
    `latitude` DECIMAL(10,7) DEFAULT NULL,
    `longitude` DECIMAL(10,7) DEFAULT NULL,
    `timestamp` DATETIME NOT NULL COMMENT '发生时间',

    `face_detected_before` TINYINT(1) DEFAULT NULL COMMENT '开启前是否检测到人脸',
    `face_detected_after` TINYINT(1) DEFAULT NULL COMMENT '开启后是否检测到人脸',
    `confidence_before` DECIMAL(5,4) DEFAULT NULL COMMENT '开启前人脸置信度',
    `confidence_after` DECIMAL(5,4) DEFAULT NULL COMMENT '开启后人脸置信度',

    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    KEY `idx_vehicle_time` (`vehicle_id`, `timestamp`),
    KEY `idx_action` (`action`),
    KEY `idx_trigger` (`trigger_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='红外补光灯状态日志表';

-- 示例数据
INSERT INTO `infrared_light_logs`
    (vehicle_id, driver_id, device_id, action, trigger_type, light_on,
     intensity_before, intensity_after, light_level_lux, reason,
     latitude, longitude, timestamp,
     face_detected_before, face_detected_after, confidence_before, confidence_after)
VALUES
    (1, 1, 'DEV-NIGHT-001', 'turn_on', 'auto', 1, 0, 60, 35, '光照低于阈值50lux，自动开启红外补光',
     31.230416, 121.473701, '2024-06-20 19:23:15', 0, 1, 0.3200, 0.7800),
    (1, 1, 'DEV-NIGHT-001', 'intensity_change', 'auto', 1, 60, 75, 15, '进入隧道，光照进一步降低，自动增强补光',
     31.230416, 121.473701, '2024-06-20 19:45:30', 1, 1, 0.6500, 0.8500),
    (1, 1, 'DEV-NIGHT-001', 'turn_off', 'auto', 0, 75, 0, 280, '驶出隧道，光照恢复，自动关闭红外补光',
     31.230416, 121.473701, '2024-06-20 20:10:42', 1, 1, 0.8200, 0.8800),
    (2, 2, 'DEV-NIGHT-002', 'turn_on', 'auto', 1, 0, 55, 42, '夜间模式开始，自动开启红外补光',
     31.230416, 121.473701, '2024-06-20 20:05:10', 1, 1, 0.4500, 0.8200),
    (3, 3, 'DEV-NIGHT-003', 'turn_on', 'manual', 1, 0, 70, 150, '司机手动开启红外补光（进入地下停车场）',
     31.230416, 121.473701, '2024-06-20 15:30:00', 0, 1, 0.2500, 0.7200);

-- ------------------------------------------------------------
-- 3. 图像增强处理记录表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `image_enhancement_records`;
CREATE TABLE `image_enhancement_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `vehicle_id` BIGINT NOT NULL COMMENT '车辆ID',
    `driver_id` BIGINT DEFAULT NULL COMMENT '司机ID',
    `waybill_id` BIGINT DEFAULT NULL COMMENT '运单ID',
    `device_id` VARCHAR(64) DEFAULT '' COMMENT '设备编号',

    `original_image_url` VARCHAR(256) DEFAULT '' COMMENT '原始图像URL',
    `enhanced_image_url` VARCHAR(256) DEFAULT '' COMMENT '增强后图像URL',
    `original_image_base64` TEXT COMMENT '原始图像Base64(仅边缘端临时)',
    `enhanced_image_base64` TEXT COMMENT '增强后图像Base64(仅边缘端临时)',

    `enhance_mode` VARCHAR(32) NOT NULL DEFAULT 'night' COMMENT '增强模式',
    `gamma_value` DECIMAL(4,3) DEFAULT NULL COMMENT '实际使用伽马值',
    `brightness_delta` INT DEFAULT 0 COMMENT '亮度调整值',
    `contrast_delta` INT DEFAULT 0 COMMENT '对比度调整值',
    `denoise_applied` TINYINT(1) DEFAULT 0 COMMENT '是否应用了降噪',
    `denoise_strength` TINYINT UNSIGNED DEFAULT 0 COMMENT '降噪强度',
    `histogram_eq_applied` TINYINT(1) DEFAULT 0 COMMENT '是否应用了直方图均衡化',
    `sharpen_applied` TINYINT(1) DEFAULT 0 COMMENT '是否应用了锐化',

    `original_brightness_avg` INT DEFAULT NULL COMMENT '原始图像平均亮度(0-255)',
    `enhanced_brightness_avg` INT DEFAULT NULL COMMENT '增强后平均亮度(0-255)',
    `original_contrast` INT DEFAULT NULL COMMENT '原始图像对比度',
    `enhanced_contrast` INT DEFAULT NULL COMMENT '增强后对比度',

    `light_level_lux` INT DEFAULT NULL COMMENT '环境光照强度(勒克斯)',
    `is_night_time` TINYINT(1) DEFAULT 0 COMMENT '是否夜间',

    -- 人脸检测效果对比
    `face_detected_original` TINYINT(1) DEFAULT 0 COMMENT '原始图是否检测到人脸',
    `face_detected_enhanced` TINYINT(1) DEFAULT 1 COMMENT '增强后是否检测到人脸',
    `face_confidence_original` DECIMAL(5,4) DEFAULT 0.0000 COMMENT '原始图人脸置信度',
    `face_confidence_enhanced` DECIMAL(5,4) DEFAULT 0.0000 COMMENT '增强后人脸置信度',
    `landmark_count_original` INT DEFAULT 0 COMMENT '原始图关键点数量',
    `landmark_count_enhanced` INT DEFAULT 0 COMMENT '增强后关键点数量',

    `quality_score_before` DECIMAL(5,4) DEFAULT 0.0000 COMMENT '增强前图像质量评分',
    `quality_score_after` DECIMAL(5,4) DEFAULT 0.0000 COMMENT '增强后图像质量评分',
    `quality_improvement_pct` DECIMAL(5,2) DEFAULT 0.00 COMMENT '质量提升百分比',

    `processing_time_ms` INT DEFAULT 0 COMMENT '处理耗时(毫秒)',
    `process_on_edge` TINYINT(1) DEFAULT 1 COMMENT '是否边缘端处理',

    `timestamp` DATETIME NOT NULL COMMENT '处理时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    KEY `idx_vehicle_time` (`vehicle_id`, `timestamp`),
    KEY `idx_waybill` (`waybill_id`),
    KEY `idx_mode` (`enhance_mode`),
    KEY `idx_night` (`is_night_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图像增强处理记录表';

-- 示例数据
INSERT INTO `image_enhancement_records`
    (vehicle_id, driver_id, waybill_id, device_id,
     original_image_url, enhanced_image_url,
     enhance_mode, gamma_value, brightness_delta, contrast_delta,
     denoise_applied, denoise_strength, histogram_eq_applied, sharpen_applied,
     original_brightness_avg, enhanced_brightness_avg,
     original_contrast, enhanced_contrast,
     light_level_lux, is_night_time,
     face_detected_original, face_detected_enhanced,
     face_confidence_original, face_confidence_enhanced,
     landmark_count_original, landmark_count_enhanced,
     quality_score_before, quality_score_after, quality_improvement_pct,
     processing_time_ms, process_on_edge, timestamp)
VALUES
    (1, 1, 1001, 'DEV-NIGHT-001',
     '/fatigue/night/1/orig_20240620_192316.jpg', '/fatigue/night/1/enh_20240620_192316.jpg',
     'night', 1.200, 30, 20, 1, 3, 1, 0,
     45, 95, 35, 72,
     35, 1,
     0, 1, 0.3200, 0.7800, 0, 468,
     0.2500, 0.7200, 188.00,
     28, 1, '2024-06-20 19:23:16'),

    (1, 1, 1001, 'DEV-NIGHT-001',
     '/fatigue/night/1/orig_20240620_194531.jpg', '/fatigue/night/1/enh_20240620_194531.jpg',
     'infrared', 1.100, 15, 18, 1, 2, 1, 1,
     30, 88, 28, 65,
     15, 1,
     1, 1, 0.6500, 0.8500, 336, 468,
     0.4500, 0.7800, 73.33,
     32, 1, '2024-06-20 19:45:31'),

    (2, 2, 1002, 'DEV-NIGHT-002',
     '/fatigue/night/2/orig_20240620_200511.jpg', '/fatigue/night/2/enh_20240620_200511.jpg',
     'night', 1.150, 25, 15, 1, 4, 1, 0,
     38, 85, 32, 68,
     42, 1,
     1, 1, 0.4500, 0.8200, 156, 468,
     0.3800, 0.7500, 97.37,
     25, 1, '2024-06-20 20:05:11'),

    (3, 3, 1003, 'DEV-NIGHT-003',
     '/fatigue/night/3/orig_20240620_153001.jpg', '/fatigue/night/3/enh_20240620_153001.jpg',
     'low_light', 1.300, 40, 25, 1, 3, 1, 1,
     28, 102, 22, 60,
     150, 0,
     0, 1, 0.2500, 0.7200, 0, 468,
     0.1800, 0.6500, 261.11,
     35, 1, '2024-06-20 15:30:01');

-- ------------------------------------------------------------
-- 4. 夜间检测统计视图（可选，方便查询统计）
-- ------------------------------------------------------------
-- 夜间疲劳检测准确率统计按日期汇总的示例查询结构已通过代码实现
-- 这里补充一些索引优化
ALTER TABLE `fatigue_detection_records` ADD INDEX IF NOT EXISTS `idx_night_time` (`detection_time`, `vehicle_id`);
