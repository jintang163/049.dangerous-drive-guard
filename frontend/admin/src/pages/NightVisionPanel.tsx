import React, { useEffect, useState, useCallback } from 'react'
import {
  Row,
  Col,
  Card,
  Table,
  Tag,
  Button,
  Space,
  Typography,
  Modal,
  Form,
  Select,
  Input,
  Statistic,
  Drawer,
  Descriptions,
  Image,
  Progress,
  message,
  Switch,
  Slider,
  InputNumber,
  Divider,
  Tabs,
  List,
  Tooltip,
  Empty,
} from 'antd'
import {
  BulbOutlined,
  BulbFilled,
  EyeOutlined,
  SettingOutlined,
  HistoryOutlined,
  ThunderboltOutlined,
  BarChartOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  RiseOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExperimentOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { fatigueApi } from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import { NightVisionConfig, InfraredLightLog, ImageEnhanceRecord, NightVisionStats } from '@/store/app'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs

const enhanceModeMap: Record<string, { label: string; color: string }> = {
  auto: { label: '自动模式', color: 'blue' },
  night: { label: '夜间模式', color: 'purple' },
  infrared: { label: '红外模式', color: 'red' },
  low_light: { label: '低光模式', color: 'orange' },
  manual: { label: '手动模式', color: 'default' },
}

const triggerTypeMap: Record<string, { label: string; color: string }> = {
  auto: { label: '自动触发', color: 'green' },
  manual: { label: '手动触发', color: 'blue' },
  system: { label: '系统触发', color: 'purple' },
}

const actionTypeMap: Record<string, { label: string; color: string }> = {
  turn_on: { label: '开启', color: 'green' },
  turn_off: { label: '关闭', color: 'red' },
  intensity_change: { label: '强度调整', color: 'blue' },
  auto_trigger: { label: '自动触发', color: 'purple' },
  manual_trigger: { label: '手动触发', color: 'orange' },
}

const NightVisionPanel: React.FC = () => {
  const [statsLoading, setStatsLoading] = useState(false)
  const [configLoading, setConfigLoading] = useState(false)
  const [infraredLogsLoading, setInfraredLogsLoading] = useState(false)
  const [enhanceRecordsLoading, setEnhanceRecordsLoading] = useState(false)

  const [stats, setStats] = useState<NightVisionStats | null>(null)
  const [configs, setConfigs] = useState<NightVisionConfig[]>([])
  const [configsTotal, setConfigsTotal] = useState(0)
  const [configsPage, setConfigsPage] = useState(1)
  const [configsPageSize, setConfigsPageSize] = useState(10)

  const [infraredLogs, setInfraredLogs] = useState<InfraredLightLog[]>([])
  const [infraredLogsTotal, setInfraredLogsTotal] = useState(0)
  const [infraredLogsPage, setInfraredLogsPage] = useState(1)
  const [infraredLogsPageSize, setInfraredLogsPageSize] = useState(10)

  const [enhanceRecords, setEnhanceRecords] = useState<ImageEnhanceRecord[]>([])
  const [enhanceRecordsTotal, setEnhanceRecordsTotal] = useState(0)
  const [enhanceRecordsPage, setEnhanceRecordsPage] = useState(1)
  const [enhanceRecordsPageSize, setEnhanceRecordsPageSize] = useState(10)

  const [configEditModal, setConfigEditModal] = useState<NightVisionConfig | null>(null)
  const [configEditForm] = Form.useForm()
  const [configDetailDrawer, setConfigDetailDrawer] = useState<NightVisionConfig | null>(null)
  const [enhanceDetailDrawer, setEnhanceDetailDrawer] = useState<ImageEnhanceRecord | null>(null)

  const [activeTab, setActiveTab] = useState('config')

  const fetchStats = useCallback(async () => {
    setStatsLoading(true)
    try {
      const res = await fatigueApi.getNightVisionStats()
      setStats(res)
    } catch (e) {
      // ignore
    } finally {
      setStatsLoading(false)
    }
  }, [])

  const fetchConfigs = useCallback(async () => {
    setConfigLoading(true)
    try {
      const res = await fatigueApi.listNightVisionConfigs({
        page: configsPage,
        page_size: configsPageSize,
      })
      setConfigs(res?.list || [])
      setConfigsTotal(res?.total || 0)
    } finally {
      setConfigLoading(false)
    }
  }, [configsPage, configsPageSize])

  const fetchInfraredLogs = useCallback(async () => {
    setInfraredLogsLoading(true)
    try {
      const res = await fatigueApi.listInfraredLogs({
        page: infraredLogsPage,
        page_size: infraredLogsPageSize,
      })
      setInfraredLogs(res?.list || [])
      setInfraredLogsTotal(res?.total || 0)
    } finally {
      setInfraredLogsLoading(false)
    }
  }, [infraredLogsPage, infraredLogsPageSize])

  const fetchEnhanceRecords = useCallback(async () => {
    setEnhanceRecordsLoading(true)
    try {
      const res = await fatigueApi.listEnhanceRecords({
        page: enhanceRecordsPage,
        page_size: enhanceRecordsPageSize,
      })
      setEnhanceRecords(res?.list || [])
      setEnhanceRecordsTotal(res?.total || 0)
    } finally {
      setEnhanceRecordsLoading(false)
    }
  }, [enhanceRecordsPage, enhanceRecordsPageSize])

  useEffect(() => {
    fetchStats()
    fetchConfigs()
    fetchInfraredLogs()
    fetchEnhanceRecords()
  }, [fetchStats, fetchConfigs, fetchInfraredLogs, fetchEnhanceRecords])

  const handleUpdateConfig = async (values: any) => {
    if (!configEditModal) return
    try {
      const updateData: Partial<NightVisionConfig> & { vehicle_id: number } = {
        vehicle_id: configEditModal.vehicle_id,
        ...values,
      }
      await fatigueApi.updateNightVisionConfig(updateData)
      message.success('配置更新成功')
      setConfigEditModal(null)
      configEditForm.resetFields()
      fetchConfigs()
      fetchStats()
    } catch (e) {
      message.error('配置更新失败')
    }
  }

  const handleResetConfig = async (vehicleID: number) => {
    try {
      await fatigueApi.resetNightVisionConfig(vehicleID)
      message.success('配置已重置为默认')
      fetchConfigs()
      fetchStats()
    } catch (e) {
      message.error('重置失败')
    }
  }

  const qualityTrendChart = React.useMemo(() => {
    return {
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 20, top: 20, bottom: 30 },
      xAxis: {
        type: 'category',
        data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'],
      },
      yAxis: { type: 'value', max: 100, name: '质量评分' },
      legend: { data: ['增强前', '增强后'], bottom: 0 },
      series: [
        {
          name: '增强前',
          type: 'line',
          smooth: true,
          itemStyle: { color: '#fa8c16' },
          areaStyle: { color: 'rgba(250, 140, 22, 0.1)' },
          data: [45, 42, 55, 78, 82, 65, 48],
        },
        {
          name: '增强后',
          type: 'line',
          smooth: true,
          itemStyle: { color: '#52c41a' },
          areaStyle: { color: 'rgba(82, 196, 26, 0.15)' },
          data: [72, 70, 82, 90, 92, 85, 75],
        },
      ],
    }
  }, [])

  const infraredTriggerChart = React.useMemo(() => {
    if (!stats) return { series: [{ data: [] }] }
    return {
      tooltip: { trigger: 'item' },
      legend: { bottom: 0, type: 'scroll' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}: {c}' },
        data: [
          { name: '自动触发', value: stats.auto_trigger_count || 0, itemStyle: { color: '#52c41a' } },
          { name: '手动触发', value: stats.manual_trigger_count || 0, itemStyle: { color: '#1677ff' } },
        ],
      }],
    }
  }, [stats])

  const configColumns = [
    {
      title: '车辆ID',
      dataIndex: 'vehicle_id',
      key: 'vehicle_id',
      width: 80,
    },
    {
      title: '设备ID',
      dataIndex: 'device_id',
      key: 'device_id',
      width: 120,
      ellipsis: true,
    },
    {
      title: '红外补光',
      dataIndex: 'infrared_enabled',
      key: 'infrared_enabled',
      width: 100,
      render: (v: boolean) => (
        <Tag color={v ? 'green' : 'default'}>
          {v ? <BulbFilled style={{ color: '#52c41a' }} /> : <BulbOutlined />}
          {' '}{v ? '已开启' : '已关闭'}
        </Tag>
      ),
    },
    {
      title: '红外模式',
      dataIndex: 'infrared_auto_mode',
      key: 'infrared_auto_mode',
      width: 100,
      render: (v: boolean) => (
        <Tag color={v ? 'blue' : 'orange'}>
          {v ? '自动' : '手动'}
        </Tag>
      ),
    },
    {
      title: '图像增强',
      dataIndex: 'enhancement_enabled',
      key: 'enhancement_enabled',
      width: 100,
      render: (v: boolean) => (
        <Tag color={v ? 'purple' : 'default'}>
          {v ? '已开启' : '已关闭'}
        </Tag>
      ),
    },
    {
      title: '增强模式',
      dataIndex: 'enhance_mode',
      key: 'enhance_mode',
      width: 100,
      render: (v: string) => (
        <Tag color={enhanceModeMap[v]?.color || 'default'}>
          {enhanceModeMap[v]?.label || v}
        </Tag>
      ),
    },
    {
      title: '低光阈值(lux)',
      dataIndex: 'low_light_threshold_lux',
      key: 'low_light_threshold_lux',
      width: 110,
    },
    {
      title: '夜间时段',
      key: 'night_hours',
      width: 120,
      render: (_: any, record: NightVisionConfig) => (
        <Text type="secondary">
          {record.night_start_hour.toString().padStart(2, '0')}:00 -
          {record.night_end_hour.toString().padStart(2, '0')}:00
        </Text>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right' as const,
      render: (_: any, record: NightVisionConfig) => (
        <Space size="small">
          <Button size="small" type="link" icon={<EyeOutlined />} onClick={() => setConfigDetailDrawer(record)}>
            详情
          </Button>
          <Button size="small" type="link" icon={<SettingOutlined />} onClick={() => {
            setConfigEditModal(record)
            configEditForm.setFieldsValue(record)
          }}>
            编辑
          </Button>
        </Space>
      ),
    },
  ]

  const infraredLogColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '车辆ID',
      dataIndex: 'vehicle_id',
      key: 'vehicle_id',
      width: 80,
    },
    {
      title: '动作类型',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (v: string) => (
        <Tag color={actionTypeMap[v]?.color || 'default'}>
          {actionTypeMap[v]?.label || v}
        </Tag>
      ),
    },
    {
      title: '触发方式',
      dataIndex: 'trigger_type',
      key: 'trigger_type',
      width: 100,
      render: (v: string) => (
        <Tag color={triggerTypeMap[v]?.color || 'default'}>
          {triggerTypeMap[v]?.label || v}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'light_on',
      key: 'light_on',
      width: 80,
      render: (v: boolean) => (
        <Tag color={v ? 'green' : 'default'}>
          {v ? '开启' : '关闭'}
        </Tag>
      ),
    },
    {
      title: '强度变化',
      key: 'intensity',
      width: 120,
      render: (_: any, record: InfraredLightLog) => {
        if (record.action !== 'intensity_change') return '-'
        return (
          <Space size="small">
            <Text type="secondary">{record.intensity_before}%</Text>
            <RiseOutlined style={{ color: '#52c41a' }} />
            <Text strong style={{ color: '#52c41a' }}>{record.intensity_after}%</Text>
          </Space>
        )
      },
    },
    {
      title: '光照(lux)',
      dataIndex: 'light_level_lux',
      key: 'light_level_lux',
      width: 90,
      render: (v?: number) => v ?? '-',
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
      ellipsis: true,
    },
    {
      title: '人脸检测提升',
      key: 'face_improve',
      width: 130,
      render: (_: any, record: InfraredLightLog) => {
        if (record.face_detected_before === undefined) return '-'
        return (
          <Space size="small">
            <Tag color={record.face_detected_before ? 'green' : 'red'}>
              {record.face_detected_before ? '检测到' : '未检测'}
            </Tag>
            <Text type="secondary">→</Text>
            <Tag color={record.face_detected_after ? 'green' : 'red'}>
              {record.face_detected_after ? '检测到' : '未检测'}
            </Tag>
          </Space>
        )
      },
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 160,
      render: (v: string) => formatDateTime(v),
    },
  ]

  const enhanceRecordColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '车辆ID',
      dataIndex: 'vehicle_id',
      key: 'vehicle_id',
      width: 80,
    },
    {
      title: '增强模式',
      dataIndex: 'enhance_mode',
      key: 'enhance_mode',
      width: 100,
      render: (v: string) => (
        <Tag color={enhanceModeMap[v]?.color || 'default'}>
          {enhanceModeMap[v]?.label || v}
        </Tag>
      ),
    },
    {
      title: '是否夜间',
      dataIndex: 'is_night_time',
      key: 'is_night_time',
      width: 80,
      render: (v: boolean) => (
        <Tag color={v ? 'purple' : 'default'}>
          {v ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '光照(lux)',
      dataIndex: 'light_level_lux',
      key: 'light_level_lux',
      width: 90,
      render: (v?: number) => v ?? '-',
    },
    {
      title: '质量提升',
      dataIndex: 'quality_improvement_pct',
      key: 'quality_improvement_pct',
      width: 100,
      render: (v: number) => (
        <Tag color={v > 20 ? 'green' : v > 10 ? 'blue' : 'orange'}>
          +{v.toFixed(1)}%
        </Tag>
      ),
    },
    {
      title: '人脸检测',
      key: 'face_detect',
      width: 130,
      render: (_: any, record: ImageEnhanceRecord) => (
        <Space size="small">
          <Tag color={record.face_detected_original ? 'green' : 'red'}>
            原图
          </Tag>
          <Text type="secondary">→</Text>
          <Tag color={record.face_detected_enhanced ? 'green' : 'red'}>
            增强
          </Tag>
        </Space>
      ),
    },
    {
      title: '置信度提升',
      key: 'confidence',
      width: 100,
      render: (_: any, record: ImageEnhanceRecord) => {
        const diff = record.face_confidence_enhanced - record.face_confidence_original
        return (
          <Text style={{ color: diff > 0 ? '#52c41a' : '#ff4d4f' }}>
            {diff > 0 ? '+' : ''}{(diff * 100).toFixed(1)}%
          </Text>
        )
      },
    },
    {
      title: '处理耗时',
      dataIndex: 'processing_time_ms',
      key: 'processing_time_ms',
      width: 100,
      render: (v: number) => `${v}ms`,
    },
    {
      title: '处理位置',
      dataIndex: 'process_on_edge',
      key: 'process_on_edge',
      width: 90,
      render: (v: boolean) => (
        <Tag color={v ? 'blue' : 'purple'}>
          {v ? '边缘端' : '云端'}
        </Tag>
      ),
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 160,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right' as const,
      render: (_: any, record: ImageEnhanceRecord) => (
        <Button size="small" type="link" icon={<EyeOutlined />} onClick={() => setEnhanceDetailDrawer(record)}>
          详情
        </Button>
      ),
    },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="夜视增强配置数"
              value={stats?.total_configs || 0}
              valueStyle={{ color: '#1677ff' }}
              prefix={<SettingOutlined />}
            />
            <div style={{ marginTop: 8 }}>
              <Space size="small">
                <Tag color="green">红外开启 {stats?.infrared_enabled_count || 0}</Tag>
                <Tag color="purple">增强开启 {stats?.enhancement_enabled_count || 0}</Tag>
              </Space>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="今日红外开启次数"
              value={stats?.today_infrared_turn_on_count || 0}
              valueStyle={{ color: '#52c41a' }}
              prefix={<BulbFilled />}
            />
            <div style={{ marginTop: 8 }}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                今日累计开启时长 {stats?.today_infrared_duration_minutes || 0} 分钟
              </Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="图像质量提升"
              value={stats?.avg_quality_improvement_pct || 0}
              suffix="%"
              valueStyle={{ color: '#722ed1' }}
              prefix={<ThunderboltOutlined />}
            />
            <div style={{ marginTop: 8 }}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                平均处理耗时 {stats?.avg_processing_time_ms || 0}ms
              </Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="夜间人脸检出率提升"
              value={stats ? Math.round((stats.night_face_detect_rate_after - stats.night_face_detect_rate_before) * 100) : 0}
              suffix="%"
              valueStyle={{ color: '#fa8c16' }}
              prefix={<SafetyCertificateOutlined />}
            />
            <div style={{ marginTop: 8 }}>
              <Space size="small">
                <Text type="secondary" style={{ fontSize: 11 }}>
                  前: {Math.round((stats?.night_face_detect_rate_before || 0) * 100)}%
                </Text>
                <Text type="secondary" style={{ fontSize: 11 }}>
                  → 后: {Math.round((stats?.night_face_detect_rate_after || 0) * 100)}%
                </Text>
              </Space>
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} md={16}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={
              <Space>
                <BarChartOutlined style={{ color: '#722ed1' }} />
                <Text strong>图像质量增强效果趋势</Text>
              </Space>
            }
            extra={
              <Button size="small" icon={<ReloadOutlined />} onClick={fetchStats}>
                刷新
              </Button>
            }
          >
            <ReactECharts option={qualityTrendChart} style={{ height: 280 }} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={
              <Space>
                <ThunderboltOutlined style={{ color: '#fa8c16' }} />
                <Text strong>红外补光触发分布</Text>
              </Space>
            }
          >
            <ReactECharts option={infraredTriggerChart} style={{ height: 280 }} />
          </Card>
        </Col>
      </Row>

      <Card bordered={false} style={{ borderRadius: 12 }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab={<span><SettingOutlined /> 配置管理</span>} key="config">
            <div style={{ marginBottom: 16 }}>
              <Space>
                <Button type="primary" icon={<ReloadOutlined />} onClick={fetchConfigs}>
                  刷新
                </Button>
              </Space>
            </div>
            <Table
              rowKey="id"
              columns={configColumns}
              dataSource={configs}
              loading={configLoading}
              pagination={{
                current: configsPage,
                pageSize: configsPageSize,
                total: configsTotal,
                onChange: (page, pageSize) => {
                  setConfigsPage(page)
                  setConfigsPageSize(pageSize)
                },
                showSizeChanger: true,
                showQuickJumper: true,
              }}
              scroll={{ x: 1100 }}
            />
          </TabPane>

          <TabPane tab={<span><HistoryOutlined /> 红外补光日志</span>} key="infrared">
            <Table
              rowKey="id"
              columns={infraredLogColumns}
              dataSource={infraredLogs}
              loading={infraredLogsLoading}
              pagination={{
                current: infraredLogsPage,
                pageSize: infraredLogsPageSize,
                total: infraredLogsTotal,
                onChange: (page, pageSize) => {
                  setInfraredLogsPage(page)
                  setInfraredLogsPageSize(pageSize)
                },
                showSizeChanger: true,
                showQuickJumper: true,
              }}
              scroll={{ x: 1200 }}
            />
          </TabPane>

          <TabPane tab={<span><ExperimentOutlined /> 图像增强记录</span>} key="enhance">
            <Table
              rowKey="id"
              columns={enhanceRecordColumns}
              dataSource={enhanceRecords}
              loading={enhanceRecordsLoading}
              pagination={{
                current: enhanceRecordsPage,
                pageSize: enhanceRecordsPageSize,
                total: enhanceRecordsTotal,
                onChange: (page, pageSize) => {
                  setEnhanceRecordsPage(page)
                  setEnhanceRecordsPageSize(pageSize)
                },
                showSizeChanger: true,
                showQuickJumper: true,
              }}
              scroll={{ x: 1300 }}
            />
          </TabPane>
        </Tabs>
      </Card>

      <Modal
        title={<Space><SettingOutlined /> 编辑夜视配置</Space>}
        open={!!configEditModal}
        onCancel={() => { setConfigEditModal(null); configEditForm.resetFields() }}
        onOk={() => configEditForm.submit()}
        okText="保存"
        width={720}
      >
        {configEditModal && (
          <Form form={configEditForm} layout="vertical" onFinish={handleUpdateConfig}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="红外补光开关" name="infrared_enabled" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="自动模式" name="infrared_auto_mode" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="手动红外开启" name="infrared_manual_on" valuePropName="checked">
                  <Switch disabled={configEditForm.getFieldValue('infrared_auto_mode')} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="红外强度(%)" name="infrared_intensity">
                  <Slider min={0} max={100} />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="低光阈值(lux)" name="low_light_threshold_lux">
                  <InputNumber min={1} max={1000} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="高光阈值(lux)" name="high_light_threshold_lux">
                  <InputNumber min={10} max={5000} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>

            <Divider />

            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="图像增强开关" name="enhancement_enabled" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="增强模式" name="enhance_mode">
                  <Select>
                    <Option value="auto">自动模式</Option>
                    <Option value="night">夜间模式</Option>
                    <Option value="infrared">红外模式</Option>
                    <Option value="low_light">低光模式</Option>
                    <Option value="manual">手动模式</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label="Gamma值" name="gamma_value">
                  <InputNumber min={0.1} max={3} step={0.1} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="亮度提升" name="brightness_boost">
                  <InputNumber min={0} max={100} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="对比度提升" name="contrast_boost">
                  <InputNumber min={0} max={100} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label="直方图均衡" name="histogram_equalization" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="CLAHE" name="clahe_enabled" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="降噪" name="denoise_enabled" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="降噪强度" name="denoise_strength">
                  <Slider min={1} max={10} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="锐化" name="sharpen_enabled" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
            </Row>

            <Divider />

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label="夜间自动模式" name="night_mode_auto" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="夜间开始(时)" name="night_start_hour">
                  <InputNumber min={0} max={23} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="夜间结束(时)" name="night_end_hour">
                  <InputNumber min={0} max={23} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="低光人脸检测优化" name="low_light_face_detect" valuePropName="checked">
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="夜间最低人脸置信度" name="min_face_confidence_night">
                  <Slider min={0.3} max={0.9} step={0.05} />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        )}
      </Modal>

      <Drawer
        title="配置详情"
        placement="right"
        width={520}
        open={!!configDetailDrawer}
        onClose={() => setConfigDetailDrawer(null)}
      >
        {configDetailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Descriptions title="基础信息" bordered column={1} size="small">
              <Descriptions.Item label="车辆ID">{configDetailDrawer.vehicle_id}</Descriptions.Item>
              <Descriptions.Item label="设备ID">{configDetailDrawer.device_id}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(configDetailDrawer.updated_at)}</Descriptions.Item>
            </Descriptions>

            <Descriptions title="红外补光配置" bordered column={1} size="small">
              <Descriptions.Item label="红外补光">
                <Tag color={configDetailDrawer.infrared_enabled ? 'green' : 'default'}>
                  {configDetailDrawer.infrared_enabled ? '已开启' : '已关闭'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="控制模式">
                <Tag color={configDetailDrawer.infrared_auto_mode ? 'blue' : 'orange'}>
                  {configDetailDrawer.infrared_auto_mode ? '自动' : '手动'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="红外强度">
                <Progress percent={configDetailDrawer.infrared_intensity} size="small" />
              </Descriptions.Item>
              <Descriptions.Item label="低光阈值">{configDetailDrawer.low_light_threshold_lux} lux</Descriptions.Item>
              <Descriptions.Item label="高光阈值">{configDetailDrawer.high_light_threshold_lux} lux</Descriptions.Item>
            </Descriptions>

            <Descriptions title="图像增强配置" bordered column={1} size="small">
              <Descriptions.Item label="图像增强">
                <Tag color={configDetailDrawer.enhancement_enabled ? 'purple' : 'default'}>
                  {configDetailDrawer.enhancement_enabled ? '已开启' : '已关闭'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="增强模式">
                <Tag color={enhanceModeMap[configDetailDrawer.enhance_mode]?.color || 'default'}>
                  {enhanceModeMap[configDetailDrawer.enhance_mode]?.label || configDetailDrawer.enhance_mode}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Gamma值">{configDetailDrawer.gamma_value}</Descriptions.Item>
              <Descriptions.Item label="亮度提升">{configDetailDrawer.brightness_boost}</Descriptions.Item>
              <Descriptions.Item label="对比度提升">{configDetailDrawer.contrast_boost}</Descriptions.Item>
              <Descriptions.Item label="直方图均衡">{configDetailDrawer.histogram_equalization ? '是' : '否'}</Descriptions.Item>
              <Descriptions.Item label="CLAHE">{configDetailDrawer.clahe_enabled ? '是' : '否'}</Descriptions.Item>
              <Descriptions.Item label="降噪">
                {configDetailDrawer.denoise_enabled ? `是 (强度${configDetailDrawer.denoise_strength})` : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="锐化">{configDetailDrawer.sharpen_enabled ? '是' : '否'}</Descriptions.Item>
            </Descriptions>

            <Descriptions title="夜间模式配置" bordered column={1} size="small">
              <Descriptions.Item label="自动夜间模式">{configDetailDrawer.night_mode_auto ? '是' : '否'}</Descriptions.Item>
              <Descriptions.Item label="夜间时段">
                {configDetailDrawer.night_start_hour.toString().padStart(2, '0')}:00 -
                {configDetailDrawer.night_end_hour.toString().padStart(2, '0')}:00
              </Descriptions.Item>
              <Descriptions.Item label="低光人脸检测优化">
                {configDetailDrawer.low_light_face_detect ? '是' : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="夜间最低置信度">
                {(configDetailDrawer.min_face_confidence_night * 100).toFixed(0)}%
              </Descriptions.Item>
            </Descriptions>

            <Space style={{ marginTop: 8 }}>
              <Button type="primary" icon={<SettingOutlined />} onClick={() => {
                setConfigEditModal(configDetailDrawer)
                configEditForm.setFieldsValue(configDetailDrawer)
                setConfigDetailDrawer(null)
              }}>
                编辑配置
              </Button>
              <Button danger onClick={() => {
                Modal.confirm({
                  title: '确认重置配置',
                  content: '确定要将配置重置为默认值吗？',
                  onOk: () => handleResetConfig(configDetailDrawer.vehicle_id),
                })
              }}>
                重置默认
              </Button>
            </Space>
          </div>
        )}
      </Drawer>

      <Drawer
        title="图像增强详情"
        placement="right"
        width={560}
        open={!!enhanceDetailDrawer}
        onClose={() => setEnhanceDetailDrawer(null)}
      >
        {enhanceDetailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Row gutter={12}>
              <Col span={12}>
                <Card size="small" title="原图" style={{ borderRadius: 8 }}>
                  <Image
                    width="100%"
                    src={enhanceDetailDrawer.original_image_url}
                    fallback="https://via.placeholder.com/240x180?text=No+Image"
                  />
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small" title="增强后" style={{ borderRadius: 8 }}>
                  <Image
                    width="100%"
                    src={enhanceDetailDrawer.enhanced_image_url}
                    fallback="https://via.placeholder.com/240x180?text=No+Image"
                  />
                </Card>
              </Col>
            </Row>

            <Descriptions title="增强参数" bordered column={2} size="small">
              <Descriptions.Item label="增强模式">
                <Tag color={enhanceModeMap[enhanceDetailDrawer.enhance_mode]?.color || 'default'}>
                  {enhanceModeMap[enhanceDetailDrawer.enhance_mode]?.label || enhanceDetailDrawer.enhance_mode}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="是否夜间">
                <Tag color={enhanceDetailDrawer.is_night_time ? 'purple' : 'default'}>
                  {enhanceDetailDrawer.is_night_time ? '是' : '否'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="光照强度">
                {enhanceDetailDrawer.light_level_lux ? `${enhanceDetailDrawer.light_level_lux} lux` : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Gamma值">{enhanceDetailDrawer.gamma_value || '-'}</Descriptions.Item>
              <Descriptions.Item label="亮度调整">+{enhanceDetailDrawer.brightness_delta}</Descriptions.Item>
              <Descriptions.Item label="对比度调整">+{enhanceDetailDrawer.contrast_delta}</Descriptions.Item>
              <Descriptions.Item label="直方图均衡">
                {enhanceDetailDrawer.histogram_eq_applied ? '是' : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="降噪">
                {enhanceDetailDrawer.denoise_applied ? `是 (强度${enhanceDetailDrawer.denoise_strength})` : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="锐化">{enhanceDetailDrawer.sharpen_applied ? '是' : '否'}</Descriptions.Item>
              <Descriptions.Item label="处理耗时">{enhanceDetailDrawer.processing_time_ms}ms</Descriptions.Item>
            </Descriptions>

            <Descriptions title="质量评估" bordered column={2} size="small">
              <Descriptions.Item label="原图亮度">
                {enhanceDetailDrawer.original_brightness_avg?.toFixed(1) || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="增强后亮度">
                {enhanceDetailDrawer.enhanced_brightness_avg?.toFixed(1) || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="原图对比度">
                {enhanceDetailDrawer.original_contrast?.toFixed(1) || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="增强后对比度">
                {enhanceDetailDrawer.enhanced_contrast?.toFixed(1) || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="原图质量评分">
                <Progress percent={Math.round(enhanceDetailDrawer.quality_score_before)} size="small" />
              </Descriptions.Item>
              <Descriptions.Item label="增强后质量评分">
                <Progress percent={Math.round(enhanceDetailDrawer.quality_score_after)} size="small" strokeColor="#52c41a" />
              </Descriptions.Item>
              <Descriptions.Item label="质量提升">
                <Tag color="green">+{enhanceDetailDrawer.quality_improvement_pct.toFixed(1)}%</Tag>
              </Descriptions.Item>
            </Descriptions>

            <Descriptions title="人脸检测效果" bordered column={2} size="small">
              <Descriptions.Item label="原脸检测">
                <Tag color={enhanceDetailDrawer.face_detected_original ? 'green' : 'red'}>
                  {enhanceDetailDrawer.face_detected_original ? '检测到' : '未检测'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="增强后检测">
                <Tag color={enhanceDetailDrawer.face_detected_enhanced ? 'green' : 'red'}>
                  {enhanceDetailDrawer.face_detected_enhanced ? '检测到' : '未检测'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="原图置信度">
                {(enhanceDetailDrawer.face_confidence_original * 100).toFixed(1)}%
              </Descriptions.Item>
              <Descriptions.Item label="增强后置信度">
                <Text style={{ color: '#52c41a' }}>
                  {(enhanceDetailDrawer.face_confidence_enhanced * 100).toFixed(1)}%
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="原图标点数">{enhanceDetailDrawer.landmark_count_original}</Descriptions.Item>
              <Descriptions.Item label="增强后标点数">{enhanceDetailDrawer.landmark_count_enhanced}</Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default NightVisionPanel
