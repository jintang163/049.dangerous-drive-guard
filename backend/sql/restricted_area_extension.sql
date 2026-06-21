-- ============================================
-- 禁行区域动态维护功能扩展
-- 包含: 模板库、临时禁行、二级审批
-- ============================================

USE ddg_db;

-- ============================================
-- 1. 禁行区域模板库
-- ============================================
CREATE TABLE IF NOT EXISTS restricted_area_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    template_name VARCHAR(128) NOT NULL COMMENT '模板名称',
    template_category VARCHAR(32) NOT NULL COMMENT '模板分类: hospital/school/mall/water_source/custom',
    area_type VARCHAR(32) NOT NULL COMMENT '区域类型',
    level TINYINT DEFAULT 2 COMMENT '危险等级 1-建议绕行 2-必须避开',
    default_radius DECIMAL(10,2) DEFAULT 500 COMMENT '默认半径(米)',
    restrict_hazard_classes VARCHAR(128) COMMENT '限制的危险品类别,逗号分隔',
    restrict_vehicle_types VARCHAR(128) COMMENT '限制的车辆类型',
    height_limit DECIMAL(5,2) COMMENT '限高(米)',
    weight_limit DECIMAL(8,2) COMMENT '限重(吨)',
    time_rules JSON COMMENT '默认时间规则',
    description VARCHAR(512) COMMENT '模板说明',
    is_builtin TINYINT DEFAULT 0 COMMENT '是否内置模板 0-否 1-是',
    is_enabled TINYINT DEFAULT 1 COMMENT '是否启用',
    created_by BIGINT COMMENT '创建人',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (template_category),
    INDEX idx_enabled (is_enabled)
) ENGINE=InnoDB COMMENT='禁行区域模板库表';

-- 预置四类模板
INSERT INTO restricted_area_templates (template_name, template_category, area_type, level, default_radius, restrict_hazard_classes, description, is_builtin, is_enabled) VALUES
('医院模板', 'hospital', 'hospital', 2, 800, '', '医院周边禁行模板，适用于综合医院、专科医院等医疗场所，默认半径800米，所有危险品车辆禁止通行', 1, 1),
('学校模板', 'school', 'school', 2, 500, '', '学校周边禁行模板，适用于中小学、幼儿园等教育机构，默认半径500米，建议配置上下学时段限制', 1, 1),
('商圈模板', 'mall', 'mall', 2, 1000, '3,8', '商业中心禁行模板，适用于大型商场、步行街等人员密集区域，默认半径1000米，限制易燃液体和腐蚀品', 1, 1),
('水源地模板', 'water_source', 'water_protection', 2, 5000, '3,6,8', '水源保护区禁行模板，适用于水库、饮用水源地等敏感区域，默认半径5000米，严禁各类危险品泄漏', 1, 1);

-- ============================================
-- 2. 扩展禁行区域表字段
-- ============================================
ALTER TABLE restricted_areas 
    ADD COLUMN IF NOT EXISTS shape_type VARCHAR(16) DEFAULT 'polygon' COMMENT '区域形状: polygon/circle' AFTER area_type,
    ADD COLUMN IF NOT EXISTS is_temporary TINYINT DEFAULT 0 COMMENT '是否临时禁行 0-永久 1-临时' AFTER status,
    ADD COLUMN IF NOT EXISTS temp_reason VARCHAR(512) COMMENT '临时禁行原因: accident/construction/emergency/other' AFTER is_temporary,
    ADD COLUMN IF NOT EXISTS time_schedule JSON COMMENT '生效时间段规则' AFTER time_restriction,
    ADD COLUMN IF NOT EXISTS approval_status TINYINT DEFAULT 1 COMMENT '审批状态 0-待提交 1-一级审批中 2-二级审批中 3-已通过 4-已拒绝 5-已撤销' AFTER status,
    ADD COLUMN IF NOT EXISTS first_approver_id BIGINT COMMENT '一级审批人ID' AFTER approval_status,
    ADD COLUMN IF NOT EXISTS first_approval_at DATETIME COMMENT '一级审批时间' AFTER first_approver_id,
    ADD COLUMN IF NOT EXISTS first_approval_note VARCHAR(512) COMMENT '一级审批意见' AFTER first_approval_at,
    ADD COLUMN IF NOT EXISTS second_approver_id BIGINT COMMENT '二级审批人ID' AFTER first_approval_note,
    ADD COLUMN IF NOT EXISTS second_approval_at DATETIME COMMENT '二级审批时间' AFTER second_approver_id,
    ADD COLUMN IF NOT EXISTS second_approval_note VARCHAR(512) COMMENT '二级审批意见' AFTER second_approval_at,
    ADD COLUMN IF NOT EXISTS template_id BIGINT COMMENT '来源模板ID' AFTER source,
    ADD COLUMN IF NOT EXISTS created_by BIGINT COMMENT '创建人ID' AFTER template_id,
    ADD COLUMN IF NOT EXISTS gis_import_id VARCHAR(64) COMMENT 'GIS导入批次号' AFTER created_by,
    ADD INDEX IF NOT EXISTS idx_approval_status (approval_status),
    ADD INDEX IF NOT EXISTS idx_temporary (is_temporary),
    ADD INDEX IF NOT EXISTS idx_template (template_id);

-- ============================================
-- 3. 禁行区域审批记录表
-- ============================================
CREATE TABLE IF NOT EXISTS restricted_area_approvals (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    area_id BIGINT NOT NULL COMMENT '禁行区域ID',
    approval_level TINYINT NOT NULL COMMENT '审批级别 1-一级审批 2-二级审批',
    approver_id BIGINT NOT NULL COMMENT '审批人ID',
    approver_name VARCHAR(64) COMMENT '审批人姓名',
    approval_action VARCHAR(16) NOT NULL COMMENT '审批动作: submit/approve/reject/revoke',
    approval_note VARCHAR(512) COMMENT '审批意见',
    old_status TINYINT COMMENT '原审批状态',
    new_status TINYINT COMMENT '新审批状态',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_area_id (area_id),
    INDEX idx_approver (approver_id),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB COMMENT='禁行区域审批记录表';

-- ============================================
-- 4. GIS导入记录表
-- ============================================
CREATE TABLE IF NOT EXISTS restricted_area_gis_imports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    import_batch_no VARCHAR(64) NOT NULL UNIQUE COMMENT '导入批次号',
    file_name VARCHAR(256) COMMENT '原始文件名',
    source_type VARCHAR(32) NOT NULL COMMENT '数据来源类型: official_road_network/shp/geojson/other',
    total_count INT DEFAULT 0 COMMENT '总记录数',
    success_count INT DEFAULT 0 COMMENT '成功导入数',
    failed_count INT DEFAULT 0 COMMENT '失败数',
    failed_details JSON COMMENT '失败详情',
    import_status VARCHAR(16) DEFAULT 'processing' COMMENT '导入状态: processing/completed/failed',
    imported_by BIGINT COMMENT '导入人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_batch (import_batch_no),
    INDEX idx_status (import_status)
) ENGINE=InnoDB COMMENT='GIS数据导入记录表';
