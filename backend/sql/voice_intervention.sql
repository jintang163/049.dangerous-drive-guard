-- ============================================================
-- 语音疲劳干预模块表结构
-- ============================================================

-- -----------------------------------------------------------
-- 个性化音频库
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `voice_intervention_audios`;
CREATE TABLE `voice_intervention_audios` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `driver_id` bigint DEFAULT 0 COMMENT '司机ID，0为全局',
  `org_id` bigint DEFAULT 0 COMMENT '组织ID，0为全局',
  `name` varchar(128) NOT NULL COMMENT '音频名称，如"家人提醒-妻子"',
  `category` varchar(32) NOT NULL COMMENT '类型:family家人/custom定制/system系统/emergency紧急',
  `audio_url` varchar(512) NOT NULL COMMENT '音频文件URL',
  `audio_format` varchar(16) DEFAULT 'mp3' COMMENT '音频格式 mp3/wav',
  `duration_sec` int DEFAULT 0 COMMENT '时长（秒）',
  `file_size` bigint DEFAULT 0 COMMENT '文件大小（字节）',
  `volume` int DEFAULT 80 COMMENT '默认音量(0-100)',
  `description` varchar(512) DEFAULT '' COMMENT '描述',
  `tags` json DEFAULT NULL COMMENT '标签',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否默认',
  `is_enabled` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `play_count` bigint DEFAULT 0 COMMENT '播放次数',
  `created_by` bigint DEFAULT 0 COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_driver_id` (`driver_id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_category` (`category`),
  KEY `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音干预-个性化音频库';

-- 系统内置音频示例
INSERT INTO `voice_intervention_audios` (`name`, `category`, `audio_url`, `audio_format`, `duration_sec`, `volume`, `description`, `is_default`, `is_enabled`, `is_system`) VALUES
('系统标准提醒', 'system', '/audios/system/standard_alarm.mp3', 'mp3', 5, 70, '系统标准语音提醒：请集中注意力', 1, 1, 1),
('系统高级报警', 'system', '/audios/system/severe_alarm.mp3', 'mp3', 8, 90, '系统高音量报警音：连续刺耳蜂鸣+语音', 0, 1, 1),
('家人温馨提醒-示例', 'family', '/audios/family/sample_wife.mp3', 'mp3', 6, 85, '示例：亲爱的，累了就休息一下，我等你回家', 0, 1, 0);

-- -----------------------------------------------------------
-- 语音干预策略
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `voice_intervention_strategies`;
CREATE TABLE `voice_intervention_strategies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '策略名称',
  `strategy_type` varchar(32) NOT NULL COMMENT '策略类型:normal普通/continuous连续/severe严重/emotional情感',
  `priority` int DEFAULT 1 COMMENT '优先级，数字越小越优先',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否默认策略',
  `is_enabled` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `driver_id` bigint DEFAULT 0 COMMENT '司机ID，0为全局',
  `org_id` bigint DEFAULT 0 COMMENT '组织ID，0为全局',
  `alarm_trigger` json DEFAULT NULL COMMENT '触发条件:alarm_levels/alarm_types/min_continuous_minutes/min_fatigue_score',
  `audio_ids` json DEFAULT NULL COMMENT '要播放的音频ID列表',
  `force_high_volume` tinyint(1) DEFAULT 0 COMMENT '是否强制高音量（覆盖车机设置）',
  `force_volume_percent` int DEFAULT 100 COMMENT '强制音量百分比(0-100)',
  `play_times` int DEFAULT 1 COMMENT '重复播放次数',
  `play_interval_sec` int DEFAULT 5 COMMENT '重复播放间隔秒数',
  `shuffle_audios` tinyint(1) DEFAULT 0 COMMENT '是否随机播放音频',
  `emotional_mode` tinyint(1) DEFAULT 0 COMMENT '情感模式：优先播放家人录音',
  `cooldown_seconds` int DEFAULT 30 COMMENT '冷却时间（秒内不重复触发）',
  `description` varchar(512) DEFAULT '' COMMENT '说明',
  `created_by` bigint DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_strategy_type` (`strategy_type`),
  KEY `idx_priority` (`priority`),
  KEY `idx_driver_id` (`driver_id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音干预策略';

-- 默认策略
INSERT INTO `voice_intervention_strategies` (`name`, `strategy_type`, `priority`, `is_default`, `is_enabled`, `driver_id`, `org_id`, `alarm_trigger`, `audio_ids`, `force_high_volume`, `force_volume_percent`, `play_times`, `play_interval_sec`, `shuffle_audios`, `emotional_mode`, `cooldown_seconds`, `description`) VALUES
('标准疲劳提醒', 'normal', 10, 1, 1, 0, 0,
  '{"alarm_levels":[1,2],"min_fatigue_score":0,"min_continuous_minutes":0}',
  '[1]',
  0, 70, 1, 0, 0, 1, 20, '普通疲劳，播放系统或家人标准提醒'),
('连续疲劳强制报警', 'continuous', 5, 1, 1, 0, 0,
  '{"alarm_levels":[2,3],"min_continuous_minutes":10,"min_fatigue_score":0}',
  '[1,2]',
  1, 100, 3, 3, 0, 1, 60, '连续疲劳超过10分钟，强制高音量反复播放家人+系统报警'),
('严重疲劳高音量', 'severe', 3, 1, 1, 0, 0,
  '{"alarm_levels":[3],"min_fatigue_score":40,"min_continuous_minutes":0}',
  '[2]',
  1, 100, 5, 2, 0, 0, 120, '严重疲劳(评分<40)，强制最高音量重复播放刺耳报警');

-- -----------------------------------------------------------
-- 语音干预日志
-- -----------------------------------------------------------
DROP TABLE IF EXISTS `voice_intervention_logs`;
CREATE TABLE `voice_intervention_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `vehicle_id` bigint NOT NULL COMMENT '车辆ID',
  `driver_id` bigint NOT NULL COMMENT '司机ID',
  `waybill_id` bigint DEFAULT 0,
  `alarm_id` bigint DEFAULT 0 COMMENT '关联的疲劳报警ID',
  `strategy_id` bigint DEFAULT 0 COMMENT '匹配的策略ID',
  `audio_id` bigint DEFAULT 0 COMMENT '播放的音频ID',
  `audio_name` varchar(128) DEFAULT '' COMMENT '音频名称快照',
  `audio_url` varchar(512) DEFAULT '' COMMENT '音频URL快照',
  `audio_format` varchar(16) DEFAULT '' COMMENT '音频格式快照(mp3/wav等)',
  `category` varchar(32) DEFAULT '' COMMENT '音频类型快照',
  `strategy_type` varchar(32) DEFAULT '' COMMENT '策略类型快照',
  `play_status` varchar(32) DEFAULT 'pending' COMMENT '播放状态:pending/sent/playing/completed/failed',
  `is_high_volume` tinyint(1) DEFAULT 0 COMMENT '是否高音量',
  `actual_volume_percent` int DEFAULT 0 COMMENT '实际音量',
  `play_times` int DEFAULT 0 COMMENT '实际播放次数',
  `total_play_duration_ms` bigint DEFAULT 0 COMMENT '总播放时长',
  `alarm_level` int DEFAULT 0,
  `alarm_type` varchar(64) DEFAULT '',
  `fatigue_score` decimal(5,2) DEFAULT 0,
  `continuous_minutes` int DEFAULT 0,
  `driver_ack` tinyint(1) DEFAULT 0 COMMENT '司机是否确认收到',
  `ack_at` datetime DEFAULT NULL,
  `sent_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  `error_msg` varchar(512) DEFAULT '',
  `mq_message_id` varchar(128) DEFAULT '' COMMENT '下发到车端的MQ消息ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_vehicle_id` (`vehicle_id`),
  KEY `idx_driver_id` (`driver_id`),
  KEY `idx_alarm_id` (`alarm_id`),
  KEY `idx_strategy_id` (`strategy_id`),
  KEY `idx_play_status` (`play_status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语音干预日志';
