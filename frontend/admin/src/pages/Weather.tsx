import React, { useEffect, useState, useRef } from 'react'
import {
  Row, Col, Card, Table, Tag, Button, Space, Typography, Modal,
  Statistic, List, Divider, Descriptions, Empty, Tooltip, Badge,
  message, Select, Input, Form, InputNumber, DatePicker, Alert,
  Radio, Drawer, Tabs,
} from 'antd'
import {
  CloudOutlined, CloudTwoTone, ThunderboltOutlined, ThunderboltTwoTone,
  CloudServerOutlined, CloudServerTwoTone, SnowflakeOutlined, SnowflakeTwoTone,
  EnvironmentOutlined, EyeOutlined, CarOutlined, RouteOutlined,
  ReloadOutlined, FilterOutlined, ExportOutlined, InfoCircleOutlined,
  WarningOutlined, SendOutlined, HistoryOutlined, PauseCircleOutlined,
  PlayCircleOutlined, SafetyOutlined, SearchOutlined, DashboardOutlined,
  BulbOutlined, SyncOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs'
import {
  weatherApi, WeatherPushRecord, HistoricalWeather,
  OperationSuspension, RouteWeatherAnalysis,
} from '@/services/api'
import type { WeatherWarning } from '@/store/app'

const { Text } = Typography
const { Option } = Select

const warningTypeMap: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  rainstorm: { label: '暴雨预警', color: '#1677ff', icon: <CloudOutlined /> },
  typhoon: { label: '台风预警', color: '#722ed1', icon: <ThunderboltOutlined /> },
  fog: { label: '大雾预警', color: '#8c8c8c', icon: <CloudServerOutlined /> },
  haze: { label: '霾预警', color: '#a8071a', icon: <CloudServerOutlined /> },
  wind: { label: '大风预警', color: '#13c2c2', icon: <ThunderboltOutlined /> },
  strong_wind: { label: '大风预警', color: '#13c2c2', icon: <ThunderboltOutlined /> },
  ice: { label: '道路结冰', color: '#69c0ff', icon: <SnowflakeOutlined /> },
  snow: { label: '暴雪预警', color: '#1890ff', icon: <SnowflakeOutlined /> },
  snowstorm: { label: '暴雪预警', color: '#1890ff', icon: <SnowflakeOutlined /> },
  high_temp: { label: '高温预警', color: '#f5222d', icon: <ThunderboltOutlined /> },
  low_temp: { label: '低温预警', color: '#13c2c2', icon: <SnowflakeOutlined /> },
  thunder: { label: '雷电预警', color: '#722ed1', icon: <ThunderboltOutlined /> },
  hail: { label: '冰雹预警', color: '#eb2f96', icon: <ThunderboltOutlined /> },
  sandstorm: { label: '沙尘暴预警', color: '#d4b106', icon: <CloudOutlined /> },
  slippery: { label: '路面湿滑', color: '#1677ff', icon: <CloudOutlined /> },
}

const warningLevelMap: Record<string, { label: string; color: string; bgColor: string; level: number }> = {
  blue: { label: '一般(蓝色)', color: 'blue', bgColor: 'rgba(22,119,255,0.3)', level: 1 },
  yellow: { label: '较重(黄色)', color: 'gold', bgColor: 'rgba(250,173,20,0.3)', level: 2 },
  orange: { label: '严重(橙色)', color: 'orange', bgColor: 'rgba(250,140,22,0.3)', level: 3 },
  red: { label: '特别严重(红色)', color: 'red', bgColor: 'rgba(255,77,79,0.3)', level: 4 },
}

const processedToStatus = (p: number) => (p === 0 ? 'active' : p === 1 ? 'expired' : 'cancelled')

const statusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'red', label: '生效中' },
  expired: { color: 'default', label: '已过期' },
  cancelled: { color: 'blue', label: '已解除' },
}

const phaseMap: Record<string, { label: string; color: string }> = {
  pre_departure: { label: '出发前', color: 'blue' },
  en_route: { label: '行驶中', color: 'orange' },
  emergency: { label: '紧急', color: 'red' },
}

const pushStatusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'default', label: '待发送' },
  sending: { color: 'processing', label: '发送中' },
  sent: { color: 'success', label: '已发送' },
  failed: { color: 'error', label: '发送失败' },
}

const riskLevelMap: Record<string, { label: string; color: string }> = {
  low: { label: '低风险', color: '#52c41a' },
  medium: { label: '中风险', color: '#faad14' },
  high: { label: '高风险', color: '#fa8c16' },
  extreme: { label: '极高风险', color: '#ff4d4f' },
}

const Weather: React.FC = () => {
  const mapRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [warnings, setWarnings] = useState<WeatherWarning[]>([])
  const [detailModal, setDetailModal] = useState<WeatherWarning | null>(null)
  const [levelFilter, setLevelFilter] = useState<number>()
  const [typeFilter, setTypeFilter] = useState<string>()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [syncLoading, setSyncLoading] = useState(false)
  const [pushModalVisible, setPushModalVisible] = useState(false)
  const [pushForm] = Form.useForm()
  const [pushLoading, setPushLoading] = useState(false)
  const [pushRecords, setPushRecords] = useState<WeatherPushRecord[]>([])
  const [pushPage, setPushPage] = useState(1)
  const [pushPageSize, setPushPageSize] = useState(10)
  const [pushTotal, setPushTotal] = useState(0)
  const [historicalDrawerVisible, setHistoricalDrawerVisible] = useState(false)
  const [historicalForm] = Form.useForm()
  const [historicalLoading, setHistoricalLoading] = useState(false)
  const [historicalResult, setHistoricalResult] = useState<HistoricalWeather | null>(null)
  const [suspensionModalVisible, setSuspensionModalVisible] = useState(false)
  const [suspensionForm] = Form.useForm()
  const [suspensionLoading, setSuspensionLoading] = useState(false)
  const [currentSuspension, setCurrentSuspension] = useState<OperationSuspension | null>(null)
  const [suspensions, setSuspensions] = useState<OperationSuspension[]>([])
  const [suspensionPage, setSuspensionPage] = useState(1)
  const [suspensionPageSize, setSuspensionPageSize] = useState(10)
  const [suspensionTotal, setSuspensionTotal] = useState(0)
  const [routeAnalysisDrawerVisible, setRouteAnalysisDrawerVisible] = useState(false)
  const [routeAnalysisForm] = Form.useForm()
  const [routeAnalysisLoading, setRouteAnalysisLoading] = useState(false)
  const [routeAnalysisResult, setRouteAnalysisResult] = useState<RouteWeatherAnalysis | null>(null)
  const [activeTab, setActiveTab] = useState('warnings')

  const stats = {
    rainstorm: warnings.filter((d) => d.warning_type === 'rainstorm' && processedToStatus(d.processed) === 'active').length,
    fog: warnings.filter((d) => (d.warning_type === 'fog' || d.warning_type === 'haze') && processedToStatus(d.processed) === 'active').length,
    wind: warnings.filter((d) => (d.warning_type === 'wind' || d.warning_type === 'strong_wind') && processedToStatus(d.processed) === 'active').length,
    ice: warnings.filter((d) => (d.warning_type === 'ice' || d.warning_type === 'snow' || d.warning_type === 'snowstorm') && processedToStatus(d.processed) === 'active').length,
  }

  const activeWarnings = warnings.filter((d) => processedToStatus(d.processed) === 'active').sort((a, b) => (warningLevelMap[b.warning_level]?.level || 0) - (warningLevelMap[a.warning_level]?.level || 0))

  const fetchWarnings = async () => {
    setLoading(true)
    try {
      const res = await weatherApi.listWarnings({ page, page_size: pageSize, type: typeFilter, level: levelFilter, status: statusFilter })
      setWarnings(res.list || [])
      setTotal(res.total || 0)
    } catch (e: any) {
      message.error(e.message || '获取预警列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchCurrentSuspension = async () => {
    try {
      const res = await weatherApi.getCurrentSuspension()
      setCurrentSuspension(res)
    } catch (e) { /* ignore */ }
  }

  const fetchPushRecords = async () => {
    try {
      const res = await weatherApi.listPushRecords({ page: pushPage, page_size: pushPageSize })
      setPushRecords(res.list || [])
      setPushTotal(res.total || 0)
    } catch (e) { /* ignore */ }
  }

  const fetchSuspensions = async () => {
    try {
      const res = await weatherApi.listSuspensions({ page: suspensionPage, page_size: suspensionPageSize })
      setSuspensions(res.list || [])
      setSuspensionTotal(res.total || 0)
    } catch (e) { /* ignore */ }
  }

  useEffect(() => {
    fetchWarnings()
    fetchCurrentSuspension()
    fetchPushRecords()
    fetchSuspensions()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => { fetchWarnings() }, [page, pageSize, typeFilter, levelFilter, statusFilter])

  const handleSyncWarnings = async () => {
    setSyncLoading(true)
    try {
      const res = await weatherApi.syncWarnings()
      message.success(`同步完成，共处理${res.synced_count}条，新增${res.new_count}条`)
      await fetchWarnings()
    } catch (e: any) {
      message.error(e.message || '同步失败')
    } finally { setSyncLoading(false) }
  }

  const handlePushSubmit = async () => {
    try {
      const values = await pushForm.validateFields()
      setPushLoading(true)
      const res = await weatherApi.pushWeatherWarning({ ...values, target_ids: values.target_ids ? values.target_ids.split(',').map(Number) : undefined })
      message.success(`推送成功，成功${res.sent_count}条，失败${res.failed_count}条`)
      setPushModalVisible(false)
      pushForm.resetFields()
      fetchPushRecords()
    } catch (e: any) {
      if (e.errorFields) return
      message.error(e.message || '推送失败')
    } finally { setPushLoading(false) }
  }

  const handleHistoricalQuery = async () => {
    try {
      const values = await historicalForm.validateFields()
      setHistoricalLoading(true)
      const res = await weatherApi.getHistoricalWeather({
        latitude: values.latitude,
        longitude: values.longitude,
        query_time: values.query_time ? values.query_time.format('YYYY-MM-DD HH:mm:ss') : dayjs().format('YYYY-MM-DD HH:mm:ss'),
        location_name: values.location_name,
        auto_fill: true,
      })
      setHistoricalResult(res)
      message.success('查询成功')
    } catch (e: any) {
      if (e.errorFields) return
      message.error(e.message || '查询失败')
    } finally { setHistoricalLoading(false) }
  }

  const handleSuspendSubmit = async () => {
    try {
      const values = await suspensionForm.validateFields()
      setSuspensionLoading(true)
      const res = await weatherApi.suspendOperation({
        trigger_type: 'manual',
        ...values,
        expires_at: values.expires_at ? values.expires_at.format('YYYY-MM-DD HH:mm:ss') : dayjs().add(4, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      })
      message.success('已发布运营暂停指令')
      setSuspensionModalVisible(false)
      suspensionForm.resetFields()
      setCurrentSuspension(res.suspension)
      fetchSuspensions()
    } catch (e: any) {
      if (e.errorFields) return
      message.error(e.message || '操作失败')
    } finally { setSuspensionLoading(false) }
  }

  const handleResumeOperation = async () => {
    if (!currentSuspension) return
    Modal.confirm({
      title: '确认恢复运营',
      icon: <PlayCircleOutlined style={{ color: '#52c41a' }} />,
      content: '确认解除当前运营暂停状态？所有受影响车辆将恢复正常通行。',
      okText: '确认恢复',
      cancelText: '取消',
      onOk: async () => {
        await weatherApi.resumeOperation({ suspension_id: currentSuspension.id, lift_reason: '天气条件好转，人工解除' })
        message.success('运营已恢复')
        setCurrentSuspension(null)
        fetchSuspensions()
      },
    })
  }

  const handleRouteAnalysis = async () => {
    try {
      const values = await routeAnalysisForm.validateFields()
      setRouteAnalysisLoading(true)
      const res = values.waybill_id ? await weatherApi.analyzeRouteByWaybill(values.waybill_id) : await weatherApi.getRouteAnalysis(values.route_id)
      setRouteAnalysisResult(res)
      message.success('路线分析完成')
    } catch (e: any) {
      if (e.errorFields) return
      message.error(e.message || '分析失败')
    } finally { setRouteAnalysisLoading(false) }
  }

  const columns = [
    {
      title: '预警编号', dataIndex: 'warning_id', width: 160,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '类型', dataIndex: 'warning_type', width: 120,
      render: (v: string) => {
        const t = warningTypeMap[v] || { label: v, color: 'default', icon: null }
        return <Tag color={t.color} icon={t.icon} style={{ fontSize: 12 }}>{t.label}</Tag>
      },
    },
    {
      title: '等级', dataIndex: 'warning_level', width: 120,
      render: (v: string) => {
        const l = warningLevelMap[v] || warningLevelMap.blue
        return <Tag color={l.color}>{l.label}</Tag>
      },
    },
    {
      title: '影响区域', dataIndex: 'affected_cities', ellipsis: true,
      render: (v: string[]) => (
        <Tooltip title={v?.join(', ')}>
          <span><EnvironmentOutlined style={{ color: '#1677ff' }} /> {v?.join(', ') || '-'}</span>
        </Tooltip>
      ),
    },
    { title: '发布时间', dataIndex: 'publish_time', width: 170, render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text> },
    {
      title: '影响车辆数', dataIndex: 'related_vehicle_count', width: 120,
      render: (v: number) => <Space><CarOutlined /><Text strong>{v?.toLocaleString() || 0}</Text></Space>,
    },
    {
      title: '状态', dataIndex: 'processed', width: 100,
      render: (v: number) => {
        const s = statusMap[processedToStatus(v)] || statusMap.active
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    {
      title: '操作', width: 160, fixed: 'right' as const,
      render: (_: any, record: WeatherWarning) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailModal(record)}>详情</Button>
          <Button
            type="link" size="small" icon={<SendOutlined />}
            onClick={() => {
              pushForm.setFieldsValue({ warning_id: record.id, phase: 'en_route', title: record.title, content: record.content, target_type: 'all' })
              setPushModalVisible(true)
            }}
          >推送</Button>
        </Space>
      ),
    },
  ]

  const pushColumns = [
    { title: '推送ID', dataIndex: 'push_id', width: 160, render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text> },
    {
      title: '阶段', dataIndex: 'phase', width: 100,
      render: (v: string) => {
        const p = phaseMap[v] || phaseMap.en_route
        return <Tag color={p.color}>{p.label}</Tag>
      },
    },
    { title: '标题', dataIndex: 'title', ellipsis: true },
    {
      title: '目标类型', dataIndex: 'target_type', width: 100,
      render: (v: string) => ({ vehicle: '车辆', driver: '司机', waybill: '运单', all: '全部' }[v] || v),
    },
    {
      title: '结果', width: 160,
      render: (_: any, r: WeatherPushRecord) => (
        <Space>
          <Text type="success">成功{r.success_count}</Text>
          {r.fail_count > 0 && <Text type="danger">失败{r.fail_count}</Text>}
          <Text type="secondary">已读{r.read_count}</Text>
        </Space>
      ),
    },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v: string) => {
        const s = pushStatusMap[v] || pushStatusMap.pending
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    { title: '推送时间', dataIndex: 'sent_at', width: 170, render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v || '-'}</Text> },
  ]

  const suspensionColumns = [
    { title: '暂停单号', dataIndex: 'suspension_no', width: 160, render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text> },
    {
      title: '触发方式', dataIndex: 'trigger_type', width: 100,
      render: (v: string) => <Tag color={v === 'automatic' ? 'red' : 'blue'}>{v === 'automatic' ? '自动' : '人工'}</Tag>,
    },
    { title: '原因', dataIndex: 'trigger_reason', ellipsis: true },
    { title: '影响区域', dataIndex: 'affected_region', width: 150, ellipsis: true },
    {
      title: '建议车速', dataIndex: 'suggested_speed', width: 100,
      render: (v: number) => v === 0 ? <Text type="danger"><PauseCircleOutlined /> 停运</Text> : <Text strong>{v} km/h</Text>,
    },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v: string) => {
        const m: Record<string, { color: string; label: string }> = { active: { color: 'red', label: '生效中' }, lifted: { color: 'green', label: '已解除' }, expired: { color: 'default', label: '已过期' } }
        const s = m[v] || m.active
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    { title: '创建时间', dataIndex: 'created_at', width: 170, render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text> },
    {
      title: '操作', width: 100,
      render: (_: any, record: OperationSuspension) =>
        record.status === 'active' ? (
          <Button
            type="link" size="small" danger icon={<PlayCircleOutlined />}
            onClick={() => {
              Modal.confirm({
                title: '确认解除暂停',
                content: `确认解除 ${record.suspension_no} 的运营暂停？`,
                onOk: async () => {
                  await weatherApi.resumeOperation({ suspension_id: record.id, lift_reason: '人工解除' })
                  message.success('已解除')
                  fetchCurrentSuspension()
                  fetchSuspensions()
                },
              })
            }}
          >解除</Button>
        ) : null,
    },
  ]

  const trendChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['暴雨', '大雾', '大风', '冰雪'], bottom: 0 },
    grid: { left: 40, right: 20, top: 20, bottom: 40 },
    xAxis: { type: 'category', data: Array.from({ length: 7 }, (_, i) => dayjs().subtract(6 - i, 'day').format('MM-DD')) },
    yAxis: { type: 'value' },
    series: [
      { name: '暴雨', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 6)), itemStyle: { color: '#1677ff' } },
      { name: '大雾', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 4)), itemStyle: { color: '#8c8c8c' } },
      { name: '大风', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 5)), itemStyle: { color: '#13c2c2' } },
      { name: '冰雪', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 3)), itemStyle: { color: '#69c0ff' } },
    ],
  }

  const getPolygonCenter = (polygon: Array<{ lat: number; lng: number }>) => {
    if (!polygon || polygon.length === 0) return { lat: 0, lng: 0 }
    const sum = polygon.reduce((acc, p) => ({ lat: acc.lat + p.lat, lng: acc.lng + p.lng }), { lat: 0, lng: 0 })
    return { lat: sum.lat / polygon.length, lng: sum.lng / polygon.length }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {currentSuspension && (
        <Alert
          type="error" showIcon icon={<PauseCircleOutlined style={{ fontSize: 20 }} />}
          banner
          style={{ borderRadius: 8 }}
          message={
            <Space>
              <Text strong style={{ color: '#fff' }}>当前运营暂停：{currentSuspension.affected_region}</Text>
              <Tag color="red">{currentSuspension.trigger_type === 'automatic' ? '系统自动触发' : '人工触发'}</Tag>
              <Text type="secondary" style={{ color: '#fff' }}>{currentSuspension.trigger_reason}</Text>
              {currentSuspension.visibility !== undefined && <Text type="secondary" style={{ color: '#fff' }}>能见度 {currentSuspension.visibility}m</Text>}
              {currentSuspension.suggested_speed === 0 ? <Tag color="red">建议停运</Tag> : <Tag color="orange">建议车速 {currentSuspension.suggested_speed}km/h</Tag>}
            </Space>
          }
          description={
            <Space>
              <Text style={{ color: 'rgba(255,255,255,0.85)' }}>创建时间：{currentSuspension.created_at}</Text>
              {currentSuspension.expires_at && <Text style={{ color: 'rgba(255,255,255,0.85)' }}>预计到期：{currentSuspension.expires_at}</Text>}
            </Space>
          }
          action={<Button type="primary" ghost icon={<PlayCircleOutlined />} onClick={handleResumeOperation}>恢复运营</Button>}
        />
      )}

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="暴雨预警" value={stats.rainstorm} valueStyle={{ color: '#1677ff' }} prefix={<CloudTwoTone />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="大雾预警" value={stats.fog} valueStyle={{ color: '#8c8c8c' }} prefix={<CloudServerTwoTone />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="大风预警" value={stats.wind} valueStyle={{ color: '#13c2c2' }} prefix={<ThunderboltTwoTone />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="冰雪预警" value={stats.ice} valueStyle={{ color: '#69c0ff' }} prefix={<SnowflakeTwoTone />} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={14}>
          <Card
            bordered={false} style={{ borderRadius: 12 }}
            title={<Space><EnvironmentOutlined style={{ color: '#1677ff' }} /> 预警区域分布</Space>}
            extra={
              <Space size={8}>
                {[
                  { c: 'rgba(255,77,79,0.5)', l: '严重' },
                  { c: 'rgba(250,140,22,0.5)', l: '较重' },
                  { c: 'rgba(250,173,20,0.5)', l: '一般' },
                  { c: 'rgba(22,119,255,0.5)', l: '轻微' },
                ].map((i) => (
                  <div key={i.l} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <span style={{ width: 12, height: 12, background: i.c, borderRadius: 2 }}></span>
                    <Text type="secondary" style={{ fontSize: 12 }}>{i.l}</Text>
                  </div>
                ))}
              </Space>
            }
          >
            <div ref={mapRef} style={{ height: 360, borderRadius: 8, background: 'linear-gradient(180deg, #e6f4ff 0%, #f0f5ff 100%)', position: 'relative', overflow: 'hidden' }}>
              <svg viewBox="100 18 40 25" style={{ width: '100%', height: '100%' }}>
                <path d="M115,35 L125,32 L135,35 L140,30 L150,32 L155,28 L165,30 L170,35 L165,40 L155,38 L150,42 L140,40 L130,42 L120,40 L115,35 Z" fill="#d9d9d9" stroke="#bfbfbf" strokeWidth="0.3" />
                {activeWarnings.map((w) => {
                  if (!w.affected_area_polygon) return null
                  const level = warningLevelMap[w.warning_level]
                  const strokeColor = level.color === 'red' ? '#ff4d4f' : level.color === 'orange' ? '#fa8c16' : level.color === 'gold' ? '#faad14' : '#1677ff'
                  const center = getPolygonCenter(w.affected_area_polygon)
                  const scaledPoints = w.affected_area_polygon.map((p) => `${p.lng},${p.lat}`).join(' ')
                  return (
                    <g key={w.id}>
                      <polygon points={scaledPoints} fill={level.bgColor} stroke={strokeColor} strokeWidth="0.4" style={{ cursor: 'pointer' }} onClick={() => setDetailModal(w)}>
                        <title>{`${w.title} - ${warningLevelMap[w.warning_level].label}`}</title>
                      </polygon>
                      <circle cx={center.lng} cy={center.lat} r="0.6" fill={strokeColor} />
                    </g>
                  )
                })}
              </svg>
              <div style={{ position: 'absolute', bottom: 8, right: 8, background: 'rgba(255,255,255,0.85)', padding: '4px 10px', borderRadius: 4, fontSize: 11, color: '#8c8c8c' }}>
                共 {activeWarnings.length} 个生效预警区域
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card
            bordered={false} style={{ borderRadius: 12, height: '100%' }}
            title={
              <Space>
                <WarningOutlined style={{ color: '#ff4d4f' }} />
                <Text strong style={{ fontSize: 15 }}>实时预警消息</Text>
                <Badge count={activeWarnings.length} showZero style={{ backgroundColor: '#ff4d4f' }} />
              </Space>
            }
            extra={
              <Space>
                <Button size="small" icon={<SendOutlined />} onClick={() => { pushForm.resetFields(); pushForm.setFieldsValue({ phase: 'en_route', target_type: 'all' }); setPushModalVisible(true) }}>
                  发送推送
                </Button>
                <Button size="small" icon={<HistoryOutlined />} onClick={() => setHistoricalDrawerVisible(true)}>历史查询</Button>
              </Space>
            }
            bodyStyle={{ padding: 0, maxHeight: 360, overflow: 'auto' }}
          >
            {activeWarnings.length === 0 ? (
              <Empty description="暂无生效预警" style={{ padding: '40px 0' }} />
            ) : (
              <List
                dataSource={activeWarnings}
                renderItem={(item) => {
                  const t = warningTypeMap[item.warning_type] || { label: item.warning_type, color: 'default', icon: <WarningOutlined /> }
                  const l = warningLevelMap[item.warning_level] || warningLevelMap.blue
                  return (
                    <List.Item key={item.id} style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0', cursor: 'pointer' }} onClick={() => setDetailModal(item)}>
                      <List.Item.Meta
                        avatar={
                          <div style={{
                            width: 40, height: 40, borderRadius: 8,
                            background: l.color === 'red' ? '#fff1f0' : l.color === 'orange' ? '#fff7e6' : l.color === 'gold' ? '#fffbe6' : '#e6f4ff',
                            display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20,
                            color: l.color === 'red' ? '#ff4d4f' : l.color === 'orange' ? '#fa8c16' : l.color === 'gold' ? '#faad14' : '#1677ff',
                          }}>
                            {t.icon}
                          </div>
                        }
                        title={
                          <Space size={4} wrap>
                            <Tag color={l.color}>{l.label}</Tag>
                            <Tag color={t.color}>{t.label}</Tag>
                            <Text strong style={{ fontSize: 13 }}>{item.affected_cities?.join(', ') || item.title}</Text>
                          </Space>
                        }
                        description={
                          <div style={{ fontSize: 12 }}>
                            <div style={{ marginBottom: 4 }}><Text type="secondary"><InfoCircleOutlined /> {item.content}</Text></div>
                            <Space size={16}>
                              <Text type="secondary">{item.publish_time}</Text>
                              <Text type="secondary"><CarOutlined /> 影响{item.related_vehicle_count || 0}车</Text>
                            </Space>
                          </div>
                        }
                      />
                    </List.Item>
                  )
                }}
              />
            )}
          </Card>
        </Col>
      </Row>

      <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><DashboardOutlined style={{ color: '#faad14' }} /><Text strong style={{ fontSize: 15 }}>运营控制台</Text></Space>}>
        <Row gutter={16}>
          <Col xs={24} sm={8}>
            <Card
              size="small" type="inner"
              title={<Space><SafetyOutlined style={{ color: '#52c41a' }} />路线天气分析</Space>}
              extra={<Button type="link" size="small" icon={<SearchOutlined />} onClick={() => setRouteAnalysisDrawerVisible(true)}>分析路线</Button>}
            >
              <Text type="secondary" style={{ fontSize: 12 }}>分析指定路线沿途天气，获取风险等级、分段预警、安全车速建议</Text>
              {routeAnalysisResult && (
                <>
                  <Divider style={{ margin: '12px 0' }} />
                  <Space wrap>
                    <Tag color={riskLevelMap[routeAnalysisResult.overall_risk_level]?.color}>
                      {riskLevelMap[routeAnalysisResult.overall_risk_level]?.label}
                    </Tag>
                    <Tag color="blue">建议车速 {routeAnalysisResult.safe_speed_suggestion} km/h</Tag>
                    {routeAnalysisResult.should_detour && <Tag color="red">建议改道</Tag>}
                  </Space>
                </>
              )}
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" type="inner" title={<Space><BulbOutlined style={{ color: '#fa8c16' }} />路面湿滑提醒</Space>}>
              <Alert
                type="warning" showIcon icon={<WarningOutlined />}
                message={<Space><Text>降雨/高湿天气自动检测</Text><Tag color="orange">自动降速 40%</Tag></Space>}
                description={<Text type="secondary" style={{ fontSize: 12 }}>降雨量≥2.5mm时判定路面湿滑，建议车速降至原限速60%；能见度＜50m或风速＞25m/s自动建议停运</Text>}
                style={{ borderRadius: 6 }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card
              size="small" type="inner"
              title={<Space><PauseCircleOutlined style={{ color: '#ff4d4f' }} />极端天气运营控制</Space>}
              extra={
                <Space>
                  <Button
                    type="primary" danger size="small" icon={<PauseCircleOutlined />}
                    onClick={() => {
                      suspensionForm.resetFields()
                      suspensionForm.setFieldsValue({ trigger_reason: '极端天气预警', weather_type: 'fog', visibility: 50, radius_km: 50 })
                      setSuspensionModalVisible(true)
                    }}
                  >发布暂停</Button>
                  <Button size="small" icon={<ReloadOutlined />} onClick={fetchCurrentSuspension}>刷新</Button>
                </Space>
              }
            >
              {currentSuspension ? (
                <Alert type="error" showIcon message={<Space><PauseCircleOutlined /><Text strong>暂停中</Text><Tag color="red">{currentSuspension.affected_region}</Tag></Space>} description={currentSuspension.trigger_reason} style={{ borderRadius: 6 }} />
              ) : (
                <Alert type="success" showIcon message={<Text type="success">运营状态正常</Text>} description={<Text type="secondary" style={{ fontSize: 12 }}>无生效的极端天气暂停指令</Text>} style={{ borderRadius: 6 }} />
              )}
            </Card>
          </Col>
        </Row>
      </Card>

      <Card
        bordered={false} style={{ borderRadius: 12 }}
        title={
          <Tabs activeKey={activeTab} onChange={setActiveTab}>
            <Tabs.TabPane tab={<Space><CloudOutlined />历史预警记录<Tag color="blue">{total}</Tag></Space>} key="warnings" />
            <Tabs.TabPane tab={<Space><SendOutlined />推送记录<Tag color="blue">{pushTotal}</Tag></Space>} key="pushes" />
            <Tabs.TabPane tab={<Space><PauseCircleOutlined />暂停记录<Tag color="blue">{suspensionTotal}</Tag></Space>} key="suspensions" />
          </Tabs>
        }
        bodyStyle={{ paddingTop: 0 }}
      >
        {activeTab === 'warnings' && (
          <>
            <Space wrap style={{ marginBottom: 12 }}>
              <Select allowClear placeholder="类型筛选" style={{ width: 140 }} value={typeFilter} onChange={setTypeFilter}>
                {Object.entries(warningTypeMap).map(([k, v]) => (<Option key={k} value={k}>{v.label}</Option>))}
              </Select>
              <Select allowClear placeholder="等级筛选" style={{ width: 140 }} value={levelFilter} onChange={setLevelFilter}>
                {Object.entries(warningLevelMap).map(([k, v]) => (<Option key={k} value={Number(k)}>{v.label}</Option>))}
              </Select>
              <Select allowClear placeholder="状态筛选" style={{ width: 140 }} value={statusFilter} onChange={setStatusFilter}>
                {Object.entries(statusMap).map(([k, v]) => (<Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>))}
              </Select>
              <Button icon={<FilterOutlined />} onClick={() => { setTypeFilter(undefined); setLevelFilter(undefined); setStatusFilter(undefined); setPage(1) }}>重置</Button>
              <Button icon={<ReloadOutlined />} onClick={fetchWarnings}>刷新</Button>
              <Button icon={<SyncOutlined />} loading={syncLoading} onClick={handleSyncWarnings} type="primary">同步官方预警</Button>
              <Button icon={<ExportOutlined />}>导出</Button>
            </Space>
            <Table
              rowKey="id" loading={loading} columns={columns as any} dataSource={warnings}
              pagination={{ current: page, pageSize, total, showSizeChanger: true, showQuickJumper: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
              scroll={{ x: 1100 }}
              rowClassName={(r) => r.warning_level === 'red' ? '!bg-red-50' : r.warning_level === 'orange' ? '!bg-orange-50' : ''}
            />
          </>
        )}
        {activeTab === 'pushes' && (
          <Table
            rowKey="id" size="small" columns={pushColumns as any} dataSource={pushRecords}
            pagination={{ current: pushPage, pageSize: pushPageSize, total: pushTotal, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPushPage(p); setPushPageSize(ps); fetchPushRecords() } }}
            scroll={{ x: 1000 }}
          />
        )}
        {activeTab === 'suspensions' && (
          <Table
            rowKey="id" size="small" columns={suspensionColumns as any} dataSource={suspensions}
            pagination={{ current: suspensionPage, pageSize: suspensionPageSize, total: suspensionTotal, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setSuspensionPage(p); setSuspensionPageSize(ps); fetchSuspensions() } }}
            scroll={{ x: 1100 }}
          />
        )}
      </Card>

      <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><CloudOutlined />近7日预警趋势</Space>}>
        <ReactECharts option={trendChart} style={{ height: 220 }} notMerge />
      </Card>

      <Modal
        title={
          <Space>
            <WarningOutlined style={{ color: '#ff4d4f' }} /><Text strong>预警详情</Text>
            {detailModal && (
              <>
                <Tag color={warningLevelMap[detailModal.warning_level]?.color}>{warningLevelMap[detailModal.warning_level]?.label}</Tag>
                <Tag color={warningTypeMap[detailModal.warning_type]?.color}>{warningTypeMap[detailModal.warning_type]?.label}</Tag>
              </>
            )}
          </Space>
        }
        open={!!detailModal} onCancel={() => setDetailModal(null)} width={720}
        footer={
          detailModal && processedToStatus(detailModal.processed) === 'active' ? (
            <Space>
              <Button onClick={() => setDetailModal(null)}>关闭</Button>
              <Button
                icon={<SendOutlined />}
                onClick={() => {
                  pushForm.setFieldsValue({ warning_id: detailModal.id, phase: 'en_route', title: detailModal.title, content: detailModal.content, target_type: 'all' })
                  setDetailModal(null)
                  setPushModalVisible(true)
                }}
              >推送预警</Button>
              <Button type="primary" icon={<RouteOutlined />}>路径重规划建议</Button>
            </Space>
          ) : null
        }
      >
        {detailModal && (
          <div>
            <Alert
              type={(warningLevelMap[detailModal.warning_level]?.level || 0) >= 3 ? 'error' : 'warning'} showIcon icon={<WarningOutlined />}
              message={<Space><Text strong>{detailModal.title}</Text><Text type="secondary">{detailModal.publish_time}</Text></Space>}
              description={detailModal.content} style={{ borderRadius: 8, marginBottom: 16 }}
            />
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="预警编号" span={1}><Text copyable>{detailModal.warning_id}</Text></Descriptions.Item>
              <Descriptions.Item label="状态" span={1}><Tag color={statusMap[processedToStatus(detailModal.processed)]?.color}>{statusMap[processedToStatus(detailModal.processed)]?.label}</Tag></Descriptions.Item>
              <Descriptions.Item label="影响省份" span={1}>{detailModal.affected_provinces?.join(', ') || '-'}</Descriptions.Item>
              <Descriptions.Item label="影响城市" span={1}>{detailModal.affected_cities?.join(', ') || '-'}</Descriptions.Item>
              <Descriptions.Item label="影响车辆数" span={1}><Space><CarOutlined /><Text strong>{detailModal.related_vehicle_count?.toLocaleString() || 0} 辆</Text></Space></Descriptions.Item>
              <Descriptions.Item label="失效时间" span={1}>{detailModal.end_time || '-'}</Descriptions.Item>
            </Descriptions>
            {detailModal.suggestion && (
              <>
                <Divider />
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><InfoCircleOutlined />处置建议</Space>}>
                  <List size="small" dataSource={detailModal.suggestion.split(/[;；\n]/).filter((s) => s.trim())} renderItem={(item) => (<List.Item><BulbOutlined style={{ color: '#faad14', marginRight: 8 }} /><Text>{item}</Text></List.Item>)} />
                </Card>
              </>
            )}
          </div>
        )}
      </Modal>

      <Modal
        title={<Space><SendOutlined style={{ color: '#1677ff' }} /><Text strong>发送天气预警推送</Text></Space>}
        open={pushModalVisible} onCancel={() => setPushModalVisible(false)} onOk={handlePushSubmit}
        confirmLoading={pushLoading} okText="发送推送" width={560}
      >
        <Form form={pushForm} layout="vertical">
          <Form.Item label="推送阶段" name="phase" rules={[{ required: true, message: '请选择推送阶段' }]}>
            <Radio.Group>
              <Radio.Button value="pre_departure">出发前</Radio.Button>
              <Radio.Button value="en_route">行驶中</Radio.Button>
              <Radio.Button value="emergency">紧急</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item label="推送标题" name="title" rules={[{ required: true, message: '请输入推送标题' }]}>
            <Input placeholder="如：广东省深圳市暴雨红色预警" />
          </Form.Item>
          <Form.Item label="推送内容" name="content" rules={[{ required: true, message: '请输入推送内容' }]}>
            <Input.TextArea rows={3} placeholder="详细预警描述及处置建议..." />
          </Form.Item>
          <Form.Item label="推送目标" name="target_type" rules={[{ required: true, message: '请选择推送目标类型' }]}>
            <Select>
              <Option value="all">全部在途车辆/司机</Option>
              <Option value="vehicle">指定车辆</Option>
              <Option value="driver">指定司机</Option>
              <Option value="waybill">指定运单</Option>
            </Select>
          </Form.Item>
          <Form.Item label="目标ID列表(逗号分隔)" name="target_ids">
            <Input placeholder="如: 1,2,3" />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="关联预警ID" name="warning_id"><InputNumber style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={12}><Form.Item label="关联运单ID" name="waybill_id"><InputNumber style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
        </Form>
      </Modal>

      <Drawer
        title={<Space><HistoryOutlined style={{ color: '#1677ff' }} /><Text strong>历史天气回溯查询</Text></Space>}
        width={520} open={historicalDrawerVisible} onClose={() => setHistoricalDrawerVisible(false)}
        extra={<Button type="primary" icon={<SearchOutlined />} loading={historicalLoading} onClick={handleHistoricalQuery}>查询</Button>}
      >
        <Form form={historicalForm} layout="vertical">
          <Form.Item label="地点名称" name="location_name"><Input placeholder="如：广东省深圳市南山区" /></Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="纬度" name="latitude" rules={[{ required: true, message: '请输入纬度' }]}><InputNumber style={{ width: '100%' }} step={0.0001} min={-90} max={90} placeholder="如: 22.5431" /></Form.Item></Col>
            <Col span={12}><Form.Item label="经度" name="longitude" rules={[{ required: true, message: '请输入经度' }]}><InputNumber style={{ width: '100%' }} step={0.0001} min={-180} max={180} placeholder="如: 114.0579" /></Form.Item></Col>
          </Row>
          <Form.Item label="事发时间" name="query_time" rules={[{ required: true, message: '请选择事发时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} format="YYYY-MM-DD HH:mm:ss" />
          </Form.Item>
        </Form>
        {historicalResult && (
          <>
            <Divider orientation="left"><Tag color="blue">{historicalResult.data_source}</Tag>查询结果</Divider>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="天气状况" span={1}><Tag color="blue">{historicalResult.weather_condition}</Tag></Descriptions.Item>
              <Descriptions.Item label="温度" span={1}>{historicalResult.temperature} ℃</Descriptions.Item>
              <Descriptions.Item label="体感温度" span={1}>{historicalResult.feels_like ?? '-'} ℃</Descriptions.Item>
              <Descriptions.Item label="湿度" span={1}>{historicalResult.humidity}%</Descriptions.Item>
              <Descriptions.Item label="风速" span={1}>{historicalResult.wind_speed} m/s</Descriptions.Item>
              <Descriptions.Item label="能见度" span={1}>{historicalResult.visibility} m</Descriptions.Item>
              <Descriptions.Item label="降雨量" span={1}>{historicalResult.precipitation} mm</Descriptions.Item>
              <Descriptions.Item label="气压" span={1}>{historicalResult.pressure ?? '-'} hPa</Descriptions.Item>
              <Descriptions.Item label="路面湿滑" span={1}>
                {historicalResult.road_slippery ? <Tag color="orange">是 - 建议降速40%</Tag> : <Tag color="green">否</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="紫外线指数" span={1}>{historicalResult.uv_index ?? '-'}</Descriptions.Item>
              <Descriptions.Item label="天气预警" span={2}>
                {historicalResult.warnings?.length > 0 ? <Space wrap>{historicalResult.warnings.map((w, i) => <Tag key={i} color="red">{w}</Tag>)}</Space> : <Text type="secondary">无预警</Text>}
              </Descriptions.Item>
              <Descriptions.Item label="事发地点" span={2}>{historicalResult.location_name || `${historicalResult.latitude}, ${historicalResult.longitude}`}</Descriptions.Item>
              <Descriptions.Item label="事发时间" span={2}>{historicalResult.query_time}</Descriptions.Item>
            </Descriptions>
            <Alert type="info" showIcon style={{ marginTop: 12, borderRadius: 8 }} message={<Space><InfoCircleOutlined /><Text>此数据已自动存入本地缓存，可供事故分析报告引用</Text></Space>} />
          </>
        )}
      </Drawer>

      <Modal
        title={<Space><PauseCircleOutlined style={{ color: '#ff4d4f' }} /><Text strong>发布运营暂停指令</Text></Space>}
        open={suspensionModalVisible} onCancel={() => setSuspensionModalVisible(false)} onOk={handleSuspendSubmit}
        confirmLoading={suspensionLoading} okText="确认发布暂停" okButtonProps={{ danger: true }} width={560}
      >
        <Alert type="error" showIcon style={{ borderRadius: 8, marginBottom: 16 }} message="此指令将暂停指定区域内所有危化品车辆运营，请谨慎操作" />
        <Form form={suspensionForm} layout="vertical">
          <Form.Item label="天气类型" name="weather_type" rules={[{ required: true, message: '请选择天气类型' }]}>
            <Select placeholder="选择触发暂停的天气类型">
              {Object.entries(warningTypeMap).map(([k, v]) => (<Option key={k} value={k}>{v.label}</Option>))}
            </Select>
          </Form.Item>
          <Form.Item label="暂停原因" name="trigger_reason" rules={[{ required: true, message: '请输入暂停原因' }]}>
            <Input.TextArea rows={2} placeholder="如：能见度小于50米，系统自动触发停运" />
          </Form.Item>
          <Form.Item label="影响区域描述" name="affected_region" rules={[{ required: true, message: '请输入影响区域' }]}>
            <Input placeholder="如：广东省深圳市及周边高速公路" />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="中心点纬度" name="center_lat" rules={[{ required: true, message: '请输入纬度' }]}><InputNumber style={{ width: '100%' }} step={0.0001} min={-90} max={90} /></Form.Item></Col>
            <Col span={12}><Form.Item label="中心点经度" name="center_lng" rules={[{ required: true, message: '请输入经度' }]}><InputNumber style={{ width: '100%' }} step={0.0001} min={-180} max={180} /></Form.Item></Col>
          </Row>
          <Form.Item label="影响半径（公里）" name="radius_km" rules={[{ required: true, message: '请输入影响半径' }]}>
            <InputNumber style={{ width: '100%' }} min={1} max={1000} />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="能见度（米）" name="visibility"><InputNumber style={{ width: '100%' }} min={0} placeholder="低于50将建议停运" /></Form.Item></Col>
            <Col span={12}><Form.Item label="风速（m/s）" name="wind_speed"><InputNumber style={{ width: '100%' }} min={0} placeholder="高于25将建议停运" /></Form.Item></Col>
          </Row>
          <Form.Item label="建议车速（km/h，0=停运）" name="suggested_speed">
            <InputNumber style={{ width: '100%' }} min={0} max={120} />
          </Form.Item>
          <Form.Item label="到期时间" name="expires_at">
            <DatePicker showTime style={{ width: '100%' }} format="YYYY-MM-DD HH:mm:ss" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={<Space><SafetyOutlined style={{ color: '#52c41a' }} /><Text strong>路线天气分析</Text></Space>}
        width={560} open={routeAnalysisDrawerVisible} onClose={() => setRouteAnalysisDrawerVisible(false)}
        extra={<Button type="primary" icon={<SearchOutlined />} loading={routeAnalysisLoading} onClick={handleRouteAnalysis}>分析</Button>}
      >
        <Form form={routeAnalysisForm} layout="vertical">
          <Form.Item label="查询方式" name="query_type" initialValue="waybill">
            <Radio.Group>
              <Radio.Button value="waybill">运单ID</Radio.Button>
              <Radio.Button value="route">路线ID</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.query_type !== cur.query_type}>
            {({ getFieldValue }) =>
              getFieldValue('query_type') === 'waybill' ? (
                <Form.Item label="运单ID" name="waybill_id" rules={[{ required: true, message: '请输入运单ID' }]}>
                  <InputNumber style={{ width: '100%' }} min={1} />
                </Form.Item>
              ) : (
                <Form.Item label="路线ID" name="route_id" rules={[{ required: true, message: '请输入路线ID' }]}>
                  <InputNumber style={{ width: '100%' }} min={1} />
                </Form.Item>
              )
            }
          </Form.Item>
        </Form>
        {routeAnalysisResult && (
          <>
            <Divider orientation="left">
              <Space>
                <Tag color={riskLevelMap[routeAnalysisResult.overall_risk_level]?.color}>
                  {riskLevelMap[routeAnalysisResult.overall_risk_level]?.label}
                </Tag>
                分析结果
              </Space>
            </Divider>
            <Row gutter={12} style={{ marginBottom: 12 }}>
              <Col span={12}><Card size="small" style={{ borderRadius: 6 }}><Statistic title="建议安全车速" value={routeAnalysisResult.safe_speed_suggestion} suffix="km/h" valueStyle={{ color: '#1677ff' }} /></Card></Col>
              <Col span={12}><Card size="small" style={{ borderRadius: 6 }}><Statistic title="总里程" value={routeAnalysisResult.total_distance_km} suffix="km" valueStyle={{ color: '#52c41a' }} /></Card></Col>
            </Row>
            {routeAnalysisResult.should_detour && (
              <Alert type="error" showIcon style={{ borderRadius: 6, marginBottom: 12 }} message={<Text strong>建议改道</Text>} description={routeAnalysisResult.detour_suggestion || '沿途存在高风险天气区域，建议选择替代路线'} />
            )}
            {routeAnalysisResult.suggestions?.length > 0 && (
              <Card size="small" style={{ borderRadius: 6, marginBottom: 12 }} title={<Space><BulbOutlined style={{ color: '#faad14' }} />处置建议</Space>}>
                <List size="small" dataSource={routeAnalysisResult.suggestions} renderItem={(s) => <List.Item>{s}</List.Item>} />
              </Card>
            )}
            {routeAnalysisResult.segment_warnings?.length > 0 && (
              <Card size="small" style={{ borderRadius: 6, marginBottom: 12 }} title={<Space><WarningOutlined style={{ color: '#fa8c16' }} />分段预警</Space>}>
                <List
                  size="small"
                  dataSource={routeAnalysisResult.segment_warnings}
                  renderItem={(seg) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<Tag color={riskLevelMap[seg.risk_level]?.color || 'default'}>{seg.risk_level}</Tag>}
                        title={<Space><Text strong>第{seg.segment_index + 1}段</Text><Text type="secondary">{seg.distance_km}km</Text></Space>}
                        description={
                          <Space direction="vertical" size={2}>
                            <Text type="secondary">{seg.description}</Text>
                            <Space>
                              {seg.warning_types.map((wt) => {
                                const t = warningTypeMap[wt]
                                return t ? <Tag key={wt} color={t.color}>{t.label}</Tag> : <Tag key={wt}>{wt}</Tag>
                              })}
                              <Tag color="blue">建议 {seg.suggested_speed} km/h</Tag>
                            </Space>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </>
        )}
      </Drawer>
    </div>
  )
}

export default Weather
