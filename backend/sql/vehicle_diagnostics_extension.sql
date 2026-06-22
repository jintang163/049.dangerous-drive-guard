-- ============================================================
-- 车辆状态远程诊断模块扩展表
-- 包含：故障码库、保养计划、保养工单、诊断表扩展字段
-- ============================================================

-- ============================================================
-- 1. 扩展 vehicle_diagnostics 表：增加刹车片温度、轮胎温度字段
-- ============================================================

ALTER TABLE vehicle_diagnostics
    ADD COLUMN IF NOT EXISTS brake_temp_fl DECIMAL(6,2) COMMENT '左前刹车片温度℃',
    ADD COLUMN IF NOT EXISTS brake_temp_fr DECIMAL(6,2) COMMENT '右前刹车片温度℃',
    ADD COLUMN IF NOT EXISTS brake_temp_rl DECIMAL(6,2) COMMENT '左后刹车片温度℃',
    ADD COLUMN IF NOT EXISTS brake_temp_rr DECIMAL(6,2) COMMENT '右后刹车片温度℃',
    ADD COLUMN IF NOT EXISTS tire_temp_fl   DECIMAL(6,2) COMMENT '左前轮胎温度℃',
    ADD COLUMN IF NOT EXISTS tire_temp_fr   DECIMAL(6,2) COMMENT '右前轮胎温度℃',
    ADD COLUMN IF NOT EXISTS tire_temp_rl   DECIMAL(6,2) COMMENT '左后轮胎温度℃',
    ADD COLUMN IF NOT EXISTS tire_temp_rr   DECIMAL(6,2) COMMENT '右后轮胎温度℃',
    ADD COLUMN IF NOT EXISTS brake_pad_wear_fr DECIMAL(5,2) COMMENT '右前刹车片磨损%',
    ADD COLUMN IF NOT EXISTS brake_pad_wear_rl DECIMAL(5,2) COMMENT '左后刹车片磨损%',
    ADD COLUMN IF NOT EXISTS brake_pad_wear_rr DECIMAL(5,2) COMMENT '右后刹车片磨损%';

-- ============================================================
-- 2. 故障码库表
-- ============================================================

CREATE TABLE IF NOT EXISTS fault_code_library (
    id               BIGINT PRIMARY KEY AUTO_INCREMENT,
    fault_code       VARCHAR(32)  NOT NULL UNIQUE COMMENT '故障码 如P0300',
    fault_system     VARCHAR(32)  NOT NULL COMMENT '系统分类: power动力/chassis底盘/body车身/network网络/brake制动/engine发动机/transmission变速箱/tire胎压/electrical电气',
    fault_category   VARCHAR(32)  COMMENT '故障类别: sensor传感器/actuator执行器/communication通信/mechanical机械/software软件',
    fault_level      TINYINT      NOT NULL DEFAULT 1 COMMENT '严重等级 1-提示 2-警告 3-严重 4-紧急(需立即救援)',
    title_cn         VARCHAR(128) NOT NULL COMMENT '故障名称(中文)',
    title_en         VARCHAR(256) COMMENT '故障名称(英文)',
    description      TEXT         COMMENT '故障详细描述',
    possible_causes  TEXT         COMMENT '可能原因(JSON数组)',
    symptoms         TEXT         COMMENT '表现症状(JSON数组)',
    suggestion       TEXT         NOT NULL COMMENT '处置建议',
    emergency_action TEXT         COMMENT '紧急处置步骤(严重故障)',
    auto_call_rescue TINYINT      DEFAULT 0 COMMENT '是否自动呼叫救援 0否 1是',
    related_systems  VARCHAR(256) COMMENT '关联系统(逗号分隔)',
    oem_spec         VARCHAR(256) COMMENT '厂商说明',
    is_builtin       TINYINT      DEFAULT 1 COMMENT '是否内置 1内置 0自定义',
    status           TINYINT      DEFAULT 1 COMMENT '1启用 0停用',
    created_by       BIGINT,
    created_at       DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (fault_code),
    INDEX idx_system_level (fault_system, fault_level),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='故障码库表';

-- ============================================================
-- 3. 初始化常见OBD故障码数据
-- ============================================================

INSERT INTO fault_code_library
(fault_code, fault_system, fault_category, fault_level, title_cn, title_en, description, possible_causes, symptoms, suggestion, emergency_action, auto_call_rescue, related_systems, is_builtin, status)
VALUES
-- ========== 动力/发动机系统 (P开头) ==========
('P0300', 'engine',     'combustion',  2, '随机/多缸失火检测',       'Random/Multiple Cylinder Misfire Detected',
 '发动机控制单元检测到多个气缸发生随机失火',
 '["火花塞老化","点火线圈故障","燃油喷嘴堵塞","进气泄漏","燃油压力不足"]',
 '["怠速抖动","加速无力","油耗增加","尾气异味"]',
 '建议尽快检查火花塞、点火线圈、燃油系统，必要时更换火花塞', NULL, 0, 'engine,ignition,fuel', 1, 1),

('P0171', 'engine',     'fuel',        2, '燃油系统过稀(第1排)',      'System Too Lean (Bank 1)',
 '空燃比超出稀薄极限，燃油量相对不足',
 '["空气流量计脏污","进气歧管泄漏","燃油泵压力低","氧传感器故障","燃油喷嘴堵塞"]',
 '["怠速不稳","加速迟缓","油耗偏高","故障灯亮"]',
 '清洗空气流量计，检查进气系统密封，检测燃油泵压力和氧传感器', NULL, 0, 'engine,fuel,intake', 1, 1),

('P0128', 'engine',     'sensor',      1, '冷却液温度低于恒温器调节温度', 'Coolant Thermostat (Coolant Temp Below Thermostat Regulating Temperature)',
 '发动机预热时间过长，未达到正常工作温度',
 '["恒温器卡在开启位置","冷却液不足","冷却液温度传感器故障","节温器损坏"]',
 '["暖风效果差","油耗增加","预热时间过长"]',
 '检查冷却液液位，更换恒温器，检测冷却液温度传感器', NULL, 0, 'engine,cooling', 1, 1),

('P0420', 'engine',     'emission',    2, '催化器系统效率低于阈值(第1排)', 'Catalyst System Efficiency Below Threshold (Bank 1)',
 '三元催化器转化效率下降，排放可能超标',
 '["三元催化器老化失效","氧传感器故障","燃油品质差","点火系统问题导致未燃烧燃料"]',
 '["故障灯亮","加速无力","尾气臭","油耗增加"]',
 '检查前后氧传感器，使用优质燃油，严重时更换三元催化器', NULL, 0, 'engine,exhaust,emission', 1, 1),

('P0562', 'electrical', 'power',       2, '系统电压过低',            'System Voltage Low',
 '车辆电气系统供电电压低于正常范围',
 '["蓄电池老化亏电","发电机故障","接线端子松动","用电设备漏电"]',
 '["启动困难","灯光暗淡","仪表闪烁","电子设备异常"]',
 '检查蓄电池电压和接线端子，检测发电机充电电压，必要时更换蓄电池', NULL, 0, 'electrical,battery,charging', 1, 1),

-- ========== 底盘/制动系统 (C开头) ==========
('C0035', 'chassis',    'brake',       3, '左前轮速传感器电路故障',    'Left Front Wheel Speed Sensor Circuit',
 'ABS系统检测到左前轮速度传感器信号异常',
 '["轮速传感器线束断裂","传感器探头脏污","传感器气隙过大","齿圈损坏"]',
 '["ABS故障灯亮","制动时方向跑偏","紧急制动无ABS介入"]',
 '清洁轮速传感器探头，检查线束连接，检查齿圈是否损坏，必要时更换传感器', NULL, 0, 'chassis,brake,abs', 1, 1),

('C0040', 'chassis',    'brake',       3, '右前轮速传感器电路故障',    'Right Front Wheel Speed Sensor Circuit',
 'ABS系统检测到右前轮速度传感器信号异常',
 '["轮速传感器线束断裂","传感器探头脏污","传感器气隙过大","齿圈损坏"]',
 '["ABS故障灯亮","制动时方向跑偏","紧急制动无ABS介入"]',
 '清洁轮速传感器探头，检查线束连接，检查齿圈是否损坏，必要时更换传感器', NULL, 0, 'chassis,brake,abs', 1, 1),

('C1234', 'brake',      'mechanical',  4, '制动主缸压力异常/刹车失效',   'Brake Master Cylinder Pressure Abnormal',
 '制动系统主缸压力严重异常，刹车力不足或完全失效，危及行车安全！',
 '["制动主缸密封圈失效","制动液严重泄漏","制动踏板机构断裂","真空助力器完全失效"]',
 '["刹车踏板踩空无阻力","制动力严重不足","制动距离急剧增加","刹车完全失灵"]',
 '【紧急！立即呼叫救援！】严禁继续行驶！立即打开双闪，缓慢靠边停车，使用手刹辅助制动，人员撤离至安全地带，拨打救援电话',
 '1. 立即松油门开双闪 2. 利用发动机制动(逐级降档) 3. 缓慢拉手刹(不可抱死) 4. 紧急情况可利用路边障碍物减速 5. 人员撤离拨打救援', 1, 'brake,chassis,safety', 1, 1),

('C0000', 'brake',      'mechanical',  4, '制动系统总故障/刹车失效',     'Brake System Total Failure',
 '制动系统综合故障，制动力低于安全阈值，存在刹车失效风险！',
 '["刹车片磨损至极限位","制动管路破裂泄漏","制动总泵/分泵卡死","ABS模块严重故障"]',
 '["刹车行程突然变长","制动无力","制动踏板发软","刹车失灵"]',
 '【紧急！立即呼叫救援！】禁止继续行驶！立即靠边停车，人员撤离安全区，请求专业救援',
 '1. 双闪警示 2. 利用发动机制动 3. 点拉手刹减速 4. 紧急时用路边障碍 5. 人员撤离+呼救', 1, 'brake,safety', 1, 1),

('C0055', 'chassis',    'steering',    3, '转向助力系统故障',          'Steering Assist Control Module',
 '电动/液压转向助力系统发生故障',
 '["转向助力泵损坏","助力油泄漏","转向电机故障","扭矩传感器故障","保险丝熔断"]',
 '["方向盘沉重","转向异响","转向不回位","故障灯亮"]',
 '检查转向助力油液位和管路，检测转向助力泵/电机，低速行驶至维修站', NULL, 0, 'chassis,steering', 1, 1),

-- ========== 胎压系统 ==========
('C0071', 'tire',       'sensor',      3, '胎压监测系统(TPMS)左前轮传感器故障', 'TPMS Left Front Tire Pressure Sensor Fault',
 '胎压监测系统无法接收到左前轮传感器信号',
 '["传感器电池耗尽","传感器损坏","传感器ID未匹配","射频干扰"]',
 '["TPMS故障灯亮","左前胎压显示--/无数据","胎压报警不工作"]',
 '检查TPMS传感器电池状态，重新匹配传感器ID，必要时更换胎压传感器', NULL, 0, 'tire,tpms,sensor', 1, 1),

('TPMS-LOW-FL', 'tire', 'pressure',    2, '左前轮胎压过低',            'Left Front Tire Pressure Too Low',
 '左前轮胎压低于标准值下限，影响行驶安全',
 '["轮胎缓慢漏气","气门嘴漏气","轮胎被扎","自然渗漏"]',
 '["胎压报警灯亮","方向盘可能跑偏","油耗增加","胎肩磨损加剧"]',
 '立即降低车速，尽快充气至标准胎压，检查是否有扎钉/漏气', NULL, 0, 'tire,safety', 1, 1),

('TPMS-HIGH-FL', 'tire','pressure',    1, '左前轮胎压过高',            'Left Front Tire Pressure Too High',
 '左前轮胎压超过标准值上限，影响行驶平稳性',
 '["充气过量","环境温度升高导致膨胀"]',
 '["胎压报警灯亮","乘坐颠簸感增加","胎面中间磨损加剧","爆胎风险增加"]',
 '放气至标准胎压，高温行驶后等待轮胎冷却再检查调整', NULL, 0, 'tire', 1, 1),

('TPMS-TEMP-HIGH', 'tire', 'temperature', 3, '轮胎温度过高警告',        'Tire Temperature Too High Warning',
 '轮胎内部温度超过安全阈值(一般>80℃)，存在爆胎风险',
 '["长时间高速行驶","胎压异常导致滚动阻力大","制动频繁","环境温度过高","轮胎老化"]',
 '["TPMS温度报警","轮胎发热烫手","可能伴随橡胶异味"]',
 '【危险】立即降低车速，进入服务区停车冷却轮胎，检查胎压和轮胎状况，待温度下降后再行驶，必要时更换轮胎', NULL, 0, 'tire,safety', 1, 1),

-- ========== 温度类预警 ==========
('BRAKE-TEMP-HIGH', 'brake', 'temperature', 3, '刹车片温度过高警告',   'Brake Pad Temperature Too High Warning',
 '刹车片/制动盘温度超过安全阈值(一般>300℃)，存在制动热衰退风险！',
 '["长时间连续制动(下长坡)","制动器拖滞","制动片/盘配合间隙过小","频繁急刹车"]',
 '["制动感觉变软","制动距离变长","可能有焦糊味","轮毂发烫"]',
 '【危险！】立即减少制动，利用发动机制动减速，驶入服务区冷却制动系统！严禁连续制动！检查制动片磨损状态',
 '1. 松油门 2. 换低速档利用发动机制动 3. 间歇性点刹而非长踩 4. 尽快靠边停车冷却', 0, 'brake,safety', 1, 1),

-- ========== 网络/通信 (U开头) ==========
('U0100', 'network',    'communication', 2, 'ECM/PCM通信丢失',         'Lost Communication With ECM/PCM',
 '诊断总线无法与发动机控制模块通信',
 '["CAN总线线路故障","ECM电源故障","ECM自身损坏","保险丝/继电器故障"]',
 '["多个故障灯亮","仪表显示异常","发动机无法启动或运行异常"]',
 '检查ECM电源和接地，测量CAN总线高低线电压，检查相关保险丝，必要时检修ECM', NULL, 0, 'network,can,ecm', 1, 1),

('U0121', 'network',    'communication', 3, 'ABS控制模块通信丢失',      'Lost Communication With ABS Control Module',
 '诊断总线无法与ABS防抱死系统控制模块通信',
 '["ABS模块电源故障","CAN总线至ABS模块线路故障","ABS模块损坏","保险丝熔断"]',
 '["ABS故障灯亮","制动时无ABS功能","可能伴随其他模块通信故障"]',
 '检查ABS模块保险丝和电源，检查CAN总线连接，必要时检修ABS控制模块', NULL, 0, 'network,can,abs,brake', 1, 1),

-- ========== 车身系统 (B开头) ==========
('B1000', 'body',       'electrical',  1, 'ECU内部故障',              'ECU Internal Malfunction',
 '车身控制模块内部检测到通用故障',
 '["BCM软件异常","BCM硬件故障","电压波动导致"]',
 '["部分电器功能异常","故障灯偶发"]',
 '断电重启尝试，检查车身电源稳定性，如持续出现需检修/更换BCM', NULL, 0, 'body,bcm,electrical', 1, 1),

-- ========== 安全气囊 ==========
('B0100', 'body',       'safety',      3, '前排驾驶员安全气囊回路故障', 'Frontal Driver Airbag Squib Circuit',
 '安全气囊系统检测到驾驶员侧气囊回路异常，碰撞时可能无法正常展开！',
 '["气囊游丝(时钟弹簧)断裂","气囊插头松动/氧化","气囊气体发生器失效","碰撞传感器故障"]',
 '["SRS/气囊故障灯常亮","喇叭可能不响(游丝)"]',
 '避免激烈驾驶，尽快到专业维修站检测维修安全气囊系统，切勿自行操作气囊部件', NULL, 0, 'body,safety,airbag', 1, 1);

-- ============================================================
-- 4. 车辆保养计划表
-- ============================================================

CREATE TABLE IF NOT EXISTS vehicle_maintenance_plans (
    id                   BIGINT PRIMARY KEY AUTO_INCREMENT,
    plan_no              VARCHAR(32)  NOT NULL UNIQUE COMMENT '保养计划编号',
    vehicle_id           BIGINT       NOT NULL COMMENT '关联车辆ID',
    plan_name            VARCHAR(128) NOT NULL COMMENT '保养计划名称',
    maintenance_type     VARCHAR(32)  NOT NULL COMMENT '保养类型: routine常规/mileage里程/time时间/comprehensive大保/season换季/special专项',
    trigger_mode         VARCHAR(16)  NOT NULL DEFAULT 'both' COMMENT '触发方式: mileage里程/time时间/both双条件任一达到',
    trigger_mileage_km   DECIMAL(10,2) COMMENT '触发里程(km)，相对上次保养',
    trigger_days         INT          COMMENT '触发天数，相对上次保养日期',
    base_mileage_km      DECIMAL(10,2) COMMENT '基准里程(上次保养时里程)',
    base_date            DATE         COMMENT '基准日期(上次保养日期)',
    next_mileage_km      DECIMAL(10,2) COMMENT '下次保养里程',
    next_date            DATE         COMMENT '下次保养日期',
    warn_before_km       DECIMAL(10,2) DEFAULT 500 COMMENT '提前多少公里预警',
    warn_before_days     INT          DEFAULT 7    COMMENT '提前多少天预警',
    items                TEXT         COMMENT '保养项目明细(JSON数组)',
    estimated_cost       DECIMAL(10,2) COMMENT '预估费用',
    priority             TINYINT      DEFAULT 2 COMMENT '优先级 1低 2中 3高',
    description          VARCHAR(512) COMMENT '备注说明',
    status               VARCHAR(16)  DEFAULT 'active' COMMENT 'active有效/triggered已触发待处理/completed已完成/cancelled已取消/paused已暂停',
    last_work_order_id   BIGINT       COMMENT '上次关联工单ID',
    created_by           BIGINT,
    created_at           DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle_status (vehicle_id, status),
    INDEX idx_next_mileage (next_mileage_km),
    INDEX idx_next_date (next_date),
    INDEX idx_type (maintenance_type)
) ENGINE=InnoDB COMMENT='车辆保养计划表';

-- ============================================================
-- 5. 保养工单表
-- ============================================================

CREATE TABLE IF NOT EXISTS maintenance_work_orders (
    id                    BIGINT PRIMARY KEY AUTO_INCREMENT,
    work_order_no         VARCHAR(32)  NOT NULL UNIQUE COMMENT '工单编号',
    vehicle_id            BIGINT       NOT NULL COMMENT '关联车辆ID',
    plan_id               BIGINT       COMMENT '关联保养计划ID',
    maintenance_type      VARCHAR(32)  NOT NULL COMMENT '保养类型: routine常规/comprehensive大保/repair维修/emergency抢修/season换季',
    source_type           VARCHAR(16)  NOT NULL DEFAULT 'auto' COMMENT '来源: auto自动生成/manual手动创建/emergency紧急故障',
    title                 VARCHAR(256) NOT NULL COMMENT '工单标题',
    description           TEXT         COMMENT '详细说明',
    trigger_reason        VARCHAR(64)  COMMENT '触发原因: mileage里程达到/time时间达到/fault_code故障码触发/manual手动/other',
    trigger_detail        VARCHAR(512) COMMENT '触发详情',
    vehicle_mileage_km    DECIMAL(10,2) COMMENT '生成工单时车辆里程',
    items                 TEXT         COMMENT '保养/维修项目明细(JSON数组)',
    parts_used            TEXT         COMMENT '更换配件明细(JSON数组)',
    estimated_cost        DECIMAL(10,2) COMMENT '预估费用',
    actual_cost           DECIMAL(10,2) COMMENT '实际费用',
    workshop              VARCHAR(128) COMMENT '维修厂/保养点名称',
    mechanic              VARCHAR(64)  COMMENT '维修技师',
    contact_phone         VARCHAR(20)  COMMENT '联系电话',
    appointment_time      DATETIME     COMMENT '预约时间',
    checkin_time          DATETIME     COMMENT '进厂时间',
    checkout_time         DATETIME     COMMENT '出厂时间',
    current_mileage_km    DECIMAL(10,2) COMMENT '进厂时实际里程',
    quality_check_done    TINYINT      DEFAULT 0 COMMENT '质检是否完成',
    quality_check_note    VARCHAR(512) COMMENT '质检备注',
    next_mileage_km       DECIMAL(10,2) COMMENT '下次保养里程建议',
    next_date             DATE         COMMENT '下次保养日期建议',
    priority              TINYINT      DEFAULT 2 COMMENT '优先级 1低 2中 3高 4紧急',
    status                VARCHAR(20)  DEFAULT 'pending' COMMENT 'pending待派单/assigned已派单/appointment已预约/checkin已进厂/processing施工中/quality_check质检中/pending_payment待付款/completed已完成/cancelled已取消',
    assigned_to           BIGINT       COMMENT '派单给谁(用户ID)',
    dispatcher_id         BIGINT       COMMENT '派单人ID',
    dispatched_at         DATETIME     COMMENT '派单时间',
    completed_by          BIGINT       COMMENT '完成操作人ID',
    completed_at          DATETIME     COMMENT '完成时间',
    cancelled_reason      VARCHAR(512) COMMENT '取消原因',
    driver_confirm_before TINYINT      DEFAULT 0 COMMENT '司机开工前确认',
    driver_confirm_after  TINYINT      DEFAULT 0 COMMENT '司机完工后确认',
    photos_before         TEXT         COMMENT '施工前照片(JSON数组URL)',
    photos_after          TEXT         COMMENT '施工后照片(JSON数组URL)',
    remark                TEXT         COMMENT '其他备注',
    created_by            BIGINT,
    created_at            DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vehicle_status (vehicle_id, status),
    INDEX idx_plan (plan_id),
    INDEX idx_type_status (maintenance_type, status),
    INDEX idx_created (created_at DESC),
    INDEX idx_priority (priority),
    INDEX idx_checkout (checkout_time)
) ENGINE=InnoDB COMMENT='保养工单表';

-- ============================================================
-- 6. 保养工单操作日志表
-- ============================================================

CREATE TABLE IF NOT EXISTS maintenance_order_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    work_order_id   BIGINT       NOT NULL COMMENT '工单ID',
    old_status      VARCHAR(20)  COMMENT '变更前状态',
    new_status      VARCHAR(20)  NOT NULL COMMENT '变更后状态',
    action_type     VARCHAR(32)  NOT NULL COMMENT '动作: create创建/assign派单/appointment预约/checkin进厂/start施工/pause暂停/resume恢复/quality_check质检/complete完成/cancel取消/payment付款',
    action_note     VARCHAR(1024) COMMENT '动作备注',
    operator_id     BIGINT       COMMENT '操作人ID',
    operator_name   VARCHAR(64)  COMMENT '操作人名称(冗余)',
    operator_role   VARCHAR(20)  COMMENT '操作人角色(冗余)',
    extra_data      JSON         COMMENT '扩展数据',
    created_at      DATETIME     DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order (work_order_id, created_at DESC),
    INDEX idx_action (action_type)
) ENGINE=InnoDB COMMENT='保养工单操作日志表';

-- ============================================================
-- 7. 故障告警处理记录表（vehicle_fault_alerts扩展日志）
-- ============================================================

CREATE TABLE IF NOT EXISTS vehicle_fault_alert_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_id        BIGINT       NOT NULL COMMENT '告警ID',
    action_type     VARCHAR(32)  NOT NULL COMMENT '动作: report上报/notify已通知/dispatch派单/rescue已呼叫救援/ack已确认/resolve已处理/ignore已忽略',
    action_detail   VARCHAR(1024) COMMENT '动作详情',
    operator_id     BIGINT       COMMENT '操作人ID(系统自动则为NULL)',
    operator_name   VARCHAR(64)  COMMENT '操作人名称',
    extra_data      JSON         COMMENT '扩展数据(如救援请求ID、派单ID等)',
    created_at      DATETIME     DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_alert (alert_id, created_at DESC)
) ENGINE=InnoDB COMMENT='故障告警处理日志表';

-- ============================================================
-- 8. 初始化示例保养计划（基于示例车辆）
-- ============================================================

INSERT INTO vehicle_maintenance_plans
(plan_no, vehicle_id, plan_name, maintenance_type, trigger_mode,
 trigger_mileage_km, trigger_days, base_mileage_km, base_date,
 next_mileage_km, next_date, warn_before_km, warn_before_days,
 items, estimated_cost, priority, status)
SELECT
    CONCAT('PM-', v.id, '-', LPAD(@rn:=@rn+1, 4, '0')),
    v.id,
    CONCAT(v.plate_number, ' 常规保养(10000km/6个月)'),
    'routine',
    'both',
    10000, 180,
    v.mileage,
    CURDATE(),
    v.mileage + 10000,
    DATE_ADD(CURDATE(), INTERVAL 180 DAY),
    1000, 15,
    '["更换机油","更换机油滤清器","更换空气滤清器","检查冷却液","检查制动液","检查轮胎胎压和磨损","底盘紧固件检查","灯光系统检查"]',
    1200.00,
    2,
    'active'
FROM vehicles v, (SELECT @rn:=0) t
WHERE v.status IN ('idle','running');

INSERT INTO vehicle_maintenance_plans
(plan_no, vehicle_id, plan_name, maintenance_type, trigger_mode,
 trigger_mileage_km, trigger_days, base_mileage_km, base_date,
 next_mileage_km, next_date, warn_before_km, warn_before_days,
 items, estimated_cost, priority, status)
SELECT
    CONCAT('PM-BIG-', v.id),
    v.id,
    CONCAT(v.plate_number, ' 大保养(40000km/2年)'),
    'comprehensive',
    'both',
    40000, 730,
    v.mileage,
    CURDATE(),
    v.mileage + 40000,
    DATE_ADD(CURDATE(), INTERVAL 730 DAY),
    3000, 30,
    '["更换机油","更换三滤(机滤/空滤/燃油滤)","更换变速箱油","更换刹车油","更换冷却液","更换火花塞(汽油)","检查正时皮带/链条","四轮定位","轮胎动平衡","全车深度检查"]',
    5000.00,
    3,
    'active'
FROM vehicles v
WHERE v.status IN ('idle','running');
