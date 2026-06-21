-- ============================================================
-- 多摄像头疲劳检测扩展迁移
-- 表: fatigue_detection_records, fatigue_alarms
-- 创建: 2026-06-21
-- ============================================================

USE ddg_db;

-- ============================================================
-- 1. fatigue_detection_records 增加多摄字段
-- ============================================================

-- 摄像头位置 (left/center/right/multi)
ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS camera_position VARCHAR(20) DEFAULT 'center' COMMENT '摄像头位置 left/center/right/multi'
    AFTER network_status;

-- 三摄帧图像URL
ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS left_frame_url VARCHAR(256) COMMENT '左摄像头帧URL'
    AFTER camera_position;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS center_frame_url VARCHAR(256) COMMENT '中摄像头帧URL'
    AFTER left_frame_url;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS right_frame_url VARCHAR(256) COMMENT '右摄像头帧URL'
    AFTER center_frame_url;

-- 三摄独立疲劳评分
ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS left_score DECIMAL(5,2) DEFAULT 0 COMMENT '左摄像头疲劳评分'
    AFTER right_frame_url;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS center_score DECIMAL(5,2) DEFAULT 0 COMMENT '中摄像头疲劳评分'
    AFTER left_score;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS right_score DECIMAL(5,2) DEFAULT 0 COMMENT '右摄像头疲劳评分'
    AFTER center_score;

-- 多视角融合信息
ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS fusion_method VARCHAR(32) COMMENT '融合方法 single_camera/weighted_fusion/center_fallback'
    AFTER right_score;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS fusion_confidence DECIMAL(5,4) DEFAULT 0 COMMENT '融合置信度 0-1'
    AFTER fusion_method;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS occlusion_detected TINYINT DEFAULT 0 COMMENT '是否检测到遮挡'
    AFTER fusion_confidence;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS backlit_detected TINYINT DEFAULT 0 COMMENT '是否检测到逆光'
    AFTER occlusion_detected;

ALTER TABLE fatigue_detection_records
    ADD COLUMN IF NOT EXISTS used_cameras VARCHAR(64) COMMENT '实际参与融合的摄像头列表逗号分隔 left,center,right'
    AFTER backlit_detected;

-- ============================================================
-- 2. fatigue_alarms 表增加多摄相关字段
-- ============================================================

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS camera_position VARCHAR(20) DEFAULT 'center' COMMENT '摄像头位置'
    AFTER location_address;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS fusion_method VARCHAR(32) COMMENT '融合方法'
    AFTER camera_position;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS fusion_confidence DECIMAL(5,4) DEFAULT 0 COMMENT '融合置信度'
    AFTER fusion_method;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS occlusion_detected TINYINT DEFAULT 0 COMMENT '遮挡检测'
    AFTER fusion_confidence;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS backlit_detected TINYINT DEFAULT 0 COMMENT '逆光检测'
    AFTER occlusion_detected;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS left_snap_url VARCHAR(256) COMMENT '左摄像头报警快照'
    AFTER snap_image_url;

ALTER TABLE fatigue_alarms
    ADD COLUMN IF NOT EXISTS right_snap_url VARCHAR(256) COMMENT '右摄像头报警快照'
    AFTER left_snap_url;

-- ============================================================
-- 3. 索引优化
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_fatigue_camera_position
    ON fatigue_detection_records (camera_position);

CREATE INDEX IF NOT EXISTS idx_fatigue_fusion_method
    ON fatigue_detection_records (fusion_method);

CREATE INDEX IF NOT EXISTS idx_fatigue_occlusion
    ON fatigue_detection_records (occlusion_detected)
    WHERE occlusion_detected = 1;

CREATE INDEX IF NOT EXISTS idx_fatigue_backlit
    ON fatigue_detection_records (backlit_detected)
    WHERE backlit_detected = 1;

-- ============================================================
-- 4. 融合准确率统计视图
-- ============================================================

CREATE OR REPLACE VIEW v_fusion_accuracy_stats AS
SELECT
    DATE(detection_time) AS stat_date,
    camera_position,
    fusion_method,
    COUNT(*) AS total_detections,
    SUM(CASE WHEN is_alarm_triggered = 1 THEN 1 ELSE 0 END) AS alarm_count,
    AVG(fatigue_score) AS avg_score,
    AVG(fusion_confidence) AS avg_confidence,
    SUM(CASE WHEN occlusion_detected = 1 THEN 1 ELSE 0 END) AS occlusion_count,
    SUM(CASE WHEN backlit_detected = 1 THEN 1 ELSE 0 END) AS backlit_count,
    SUM(CASE WHEN camera_position = 'multi' THEN 1 ELSE 0 END) AS multi_camera_count,
    SUM(CASE WHEN camera_position != 'multi' AND camera_position IN ('left','center','right') THEN 1 ELSE 0 END) AS single_camera_count
FROM fatigue_detection_records
WHERE detection_time >= DATE_SUB(CURDATE(), INTERVAL 90 DAY)
GROUP BY DATE(detection_time), camera_position, fusion_method
WITH CASCADED CHECK OPTION;

-- ============================================================
-- 迁移说明
-- ============================================================
-- 兼容策略:
--   - 所有新增字段均有默认值，对旧数据完全透明
--   - 旧接口 /fatigue/detect 和 /fatigue/upload/frame 写入时 camera_position=center
--   - 新接口 /fatigue/upload/multi-camera 写入时 camera_position=multi，并填写全部三摄字段
--   - 索引使用 IF NOT EXISTS 保证幂等
