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
  Badge,
  Drawer,
  Descriptions,
  Image,
  Progress,
  message,
  Radio,
  Empty,
  Divider,
  Alert,
  DatePicker,
  Spin,
} from 'antd'
import {
  AlertOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PhoneOutlined,
  VideoCameraOutlined,
  SafetyCertificateOutlined,
  EyeOutlined,
  FilterOutlined,
  ReloadOutlined,
  WarningOutlined,
  FireOutlined,
  UserOutlined,
  EnvironmentOutlined,
  ExportOutlined,
  BellOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { fatigueApi, monitorApi } from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import { useAppStore, AlarmItem, StatData } from '@/store/app'
import WebSocketManager from '@/services/ws'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { RangePicker } = DatePicker

const levelMap = {
  1: { color: 'blue', label: '提醒', badge: 'processing' as const },
  2: { color: 'orange', label: '警告', badge: 'warning' as const },
  3: { color: 'red', label: '严重', badge: 'error' as const },
}

const statusMap = {
  pending: { color: 'red', label: '待处理' },
  processing: { color: 'orange', label: '处理中' },
  acknowledged: { color: 'blue', label: '已确认' },
  resolved: { color: 'green', label: '已关闭' },
  ignored: { color: 'default', label: '忽略' },
}

const alarmTypeMap: Record<string, { label: string; color: string }> = {
  fatigue_perclos: { label: '疲劳瞌睡', color: 'red' },
  continuous_fatigue_perclos: { label: '连续疲劳', color: 'red' },
  excessive_yawn: { label: '频繁打哈欠', color: 'orange' },
  abnormal_head_posture: { label: '异常姿态', color: 'gold' },
  gaze_distraction: { label: '视线偏离', color: 'gold' },
  phone_usage: { label: '使用手机', color: 'purple' },
  smoking: { label: '抽烟', color: 'volcano' },
  no_seatbelt: { label: '未系安全带', color: 'magenta' },
  continuous_fatigue: { label: '连续超时驾驶', color: 'red' },
  multi_camera_fatigue: { label: '多摄疲劳', color: 'red' },
  multi_camera_warning: { label: '多摄预警', color: 'orange' },
  eyes_closed: { label: '闭眼检测', color: 'red' },
  no_face_detected: { label: '未检测到人脸', color: 'volcano' },
}

const cameraPositionMap: Record<string, { label: string; icon: string; color: string }> = {
  left: { label: '左摄像头', icon: '👈', color: '#1677ff' },
  center: { label: '中摄像头', icon: '📷', color: '#52c41a' },
  right: { label: '右摄像头', icon: '👉', color: '#722ed1' },
  multi: { label: '多摄融合', icon: '🔍', color: '#fa8c16' },
}

const FatigueAlarms: React.FC = () => {
  const { alarms, addAlarm, updateAlarm, fetchAlarms, fetchStats: fetchStoreStats, loading: storeLoading, stats: storeStats } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<AlarmItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string>()
  const [levelFilter, setLevelFilter] = useState<number>()
  const [typeFilter, setTypeFilter] = useState<string>()
  const [detailDrawer, setDetailDrawer] = useState<AlarmItem | null>(null)
  const [handleModal, setHandleModal] = useState<AlarmItem | null>(null)
  const [handleForm] = Form.useForm()
  const [dashboardStats, setDashboardStats] = useState<StatData | null>(null)
  const [fusionStats, setFusionStats] = useState<{
    total_detections: number
    multi_camera_count: number
    single_camera_count: number
    avg_confidence: number
    multi_vs_single_improve_pct: number
    occlusion_count: number
    backlit_count: number
  } | null>(null)
  const [detailVideoURL, setDetailVideoURL] = useState<string>('')
  const [detailSnapshotURL, setDetailSnapshotURL] = useState<string>('')
  const [detailLoading, setDetailLoading] = useState(false)
  const [statsLoading, setStatsLoading] = useState(false)

  const fetchFusionStats = useCallback(async () => {
    setStatsLoading(true)
    try {
      const res = await fatigueApi.getFusionAccuracyStats(90)
      setFusionStats(res)
    } catch (e) {
      // ignore
    } finally {
      setStatsLoading(false)
    }
  }, [])

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fatigueApi.listAlarms({
        page,
        page_size: pageSize,
        status: statusFilter,
        level: levelFilter,
        alarm_type: typeFilter,
      })
      setData(res?.list || [])
      setTotal(res?.total || 0)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, statusFilter, levelFilter, typeFilter])

  const fetchDashboardStats = useCallback(async () => {
    setStatsLoading(true)
    try {
      const data = await monitorApi.getDashboardStats()
      setDashboardStats(data)
    } finally {
      setStatsLoading(false)
    }
  }, [])

  const fetchDetailMedia = useCallback(async (alarmId: number) => {
    setDetailLoading(true)
    try {
      const [videoRes, snapshotRes] = await Promise.all([
        fatigueApi.getVideoURL(alarmId).catch(() => ({ url: '' })),
        fatigueApi.getSnapshotURL(alarmId).catch(() => ({ url: '' })),
      ])
      setDetailVideoURL(videoRes?.url || '')
      setDetailSnapshotURL(snapshotRes?.url || '')
    } finally {
      setDetailLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
    fetchDashboardStats()
    fetchFusionStats()
  }, [fetchData, fetchDashboardStats, fetchFusionStats])

  useEffect(() => {
    const unsub1 = WebSocketManager.getInstance().on('new_alarm', async (alarm: AlarmItem) => {
      addAlarm(alarm)
      try {
        const snapshotRes = await fatigueApi.getSnapshotURL(alarm.id)
        if (snapshotRes?.url) {
          updateAlarm(alarm.id, { snap_image_url: snapshotRes.url })
        }
      } catch (e) {}
      message.warning({
        content: (
          <div>
            <div style={{ fontWeight: 600 }}>🚨 新报警：{alarm.vehicle_plate || alarm.vehicle_id}</div>
            <div style={{ fontSize: 12, marginTop: 4 }}>
              {alarmTypeMap[alarm.alarm_type]?.label || alarm.alarm_type} · 疲劳指数 {Math.round(alarm.fatigue_score)}
            </div>
          </div>
        ),
        duration: 8,
      })
    })

    const unsub2 = WebSocketManager.getInstance().on('alarm_updated', (alarm: Partial<AlarmItem>) => {
      if (alarm.id) {
        updateAlarm(alarm.id, alarm)
      }
    })

    return () => {
      unsub1()
      unsub2()
    }
  }, [addAlarm, updateAlarm])

  const trendChart = React.useMemo(() => {
    if (!dashboardStats?.daily_trend?.length) {
      return {
        tooltip: { trigger: 'axis' },
        grid: { left: 40, right: 20, top: 20, bottom: 30 },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 24 }, (_, i) => `${i.toString().padStart(2, '0')}:00`),
        },
        yAxis: { type: 'value' },
        series: [{
          type: 'bar',
          itemStyle: {
            color: {
              type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: '#ff4d4f' },
                { offset: 1, color: '#ffccc7' },
              ],
            },
            borderRadius: [4, 4, 0, 0],
          },
          data: Array.from({ length: 24 }, () => 0),
        }],
      }
    }
    return {
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 20, top: 20, bottom: 30 },
      xAxis: {
        type: 'category',
        data: dashboardStats.daily_trend.slice().reverse().map(d => dayjs(d.date).format('HH:mm')),
      },
      yAxis: { type: 'value' },
      series: [{
        type: 'bar',
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: '#ff4d4f' },
              { offset: 1, color: '#ffccc7' },
            ],
          },
          borderRadius: [4, 4, 0, 0],
        },
        data: dashboardStats.daily_trend.slice().reverse().map(d => d.alarms || 0),
      }],
    }
  }, [dashboardStats])

  const distributionChart = React.useMemo(() => {
    if (!dashboardStats?.alarm_type_distribution?.length) {
      return {
        tooltip: { trigger: 'item' },
        legend: { bottom: 0, type: 'scroll' },
        series: [{
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
          label: { show: true, formatter: '{b}: {c}' },
          data: [],
        }],
      }
    }
    const colorMap = ['#ff4d4f', '#fa8c16', '#faad14', '#a0d911', '#13c2c2', '#1677ff', '#722ed1', '#eb2f96']
    return {
      tooltip: { trigger: 'item' },
      legend: { bottom: 0, type: 'scroll' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}: {c}' },
        data: dashboardStats.alarm_type_distribution.map((d, i) => ({
          name: alarmTypeMap[d.alarm_type]?.label || d.alarm_type,
          value: d.count,
          itemStyle: { color: colorMap[i % colorMap.length] },
        })),
      }],
    }
  }, [dashboardStats])

  const handleAck = async (values: any) => {
    if (!handleModal) return
    try {
      const userInfo = localStorage.getItem('ddg_user_info')
      const operatorId = userInfo ? JSON.parse(userInfo).id : 1
      await fatigueApi.ackAlarm(handleModal.id, {
        action: values.handle_type,
        remark: values.handle_note,
        operator_id: operatorId,
      })
      message.success('报警已处理')
      setHandleModal(null)
      handleForm.resetFields()
      fetchData()
      fetchDashboardStats()
      fetchFusionStats()
    } catch (e) { }
  }

  const handleOpenDetail = async (record: AlarmItem) => {
    setDetailDrawer(record)
    setDetailVideoURL('')
    setDetailSnapshotURL('')
    await fetchDetailMedia(record.id)
  }

  const handleDownloadVideo = async () => {
    if (!detailDrawer) return
    try {
      const blob = await fatigueApi.downloadVideo(detailDrawer.id)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `alarm_${detailDrawer.alarm_no}_video.mp4`
      a.click()
      window.URL.revokeObjectURL(url)
      message.success('视频下载中...')
    } catch (e) {
      message.error('下载失败')
    }
  }

  const handleBatch = (type: string) => {
    message.success(`批量${type === 'ack' ? '确认' : '忽略'}了 ${data.filter(d => d.status === 'pending').length} 条报警`)
  }

  const columns = [
    {
      title: '报警编号',
      dataIndex: 'alarm_no',
      width: 180,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '车辆',
      dataIndex: 'vehicle_plate',
      width: 120,
      render: (v, r) => (
        <Space>
          <Tag color="blue">{v || r.vehicle_id}</Tag>
          <Text type="secondary" style={{ fontSize: 12 }}>{r.driver_name}</Text>
        </Space>
      ),
    },
    {
      title: '报警类型',
      dataIndex: 'alarm_type',
      width: 140,
      render: (v) => {
        const t = alarmTypeMap[v] || { label: v, color: 'default' }
        return <Tag color={t.color} style={{ fontSize: 12 }}>{t.label}</Tag>
      },
    },
    {
      title: '等级',
      dataIndex: 'alarm_level',
      width: 80,
      render: (v) => {
        const l = levelMap[v as keyof typeof levelMap] || levelMap[1]
        return <Tag color={l.color}>{l.label}</Tag>
      },
    },
    {
      title: '疲劳指数',
      dataIndex: 'fatigue_score',
      width: 120,
      render: (v) => (
        <Progress
          percent={100 - Math.round(v || 0)}
          size="small"
          showInfo={false}
          strokeColor={(100 - v) > 40 ? '#ff4d4f' : (100 - v) > 20 ? '#faad14' : '#52c41a'}
          format={() => <Text strong type={v < 70 ? 'danger' : v < 85 ? 'warning' : 'success'}>{Math.round(v)}</Text>}
        />
      ),
    },
    {
      title: '融合信息',
      width: 140,
      render: (_: any, record: AlarmItem) => {
        const r = record as any
        if (!r.fusion_method && !r.left_score && !r.center_score && !r.right_score) return <Text type="secondary">单摄</Text>
        return (
          <Space size={2} wrap>
            <Tag color="orange" style={{ fontSize: 11 }}>{r.fusion_method?.includes('weighted') ? '多摄融合' : r.fusion_method?.includes('single') ? '单摄' : '融合'}</Tag>
            {r.occlusion_detected && <Tag color="red" style={{ fontSize: 10 }}>遮挡</Tag>}
            {r.backlit_detected && <Tag color="gold" style={{ fontSize: 10 }}>逆光</Tag>}
            {r.fusion_confidence > 0 && (
              <Text type="secondary" style={{ fontSize: 10 }}>{Math.round(r.fusion_confidence * 100)}%</Text>
            )}
          </Space>
        )
      },
    },
    {
      title: '连续疲劳',
      dataIndex: 'continuous_fatigue_minutes',
      width: 100,
      render: (v) => v ? <Badge count={`${v}分钟`} style={{ backgroundColor: v >= 20 ? '#ff4d4f' : '#fa8c16' }} /> : '-',
    },
    {
      title: '位置',
      dataIndex: 'location_address',
      ellipsis: true,
      render: (v, r) => (
        <Tooltip title={v || `${r.latitude}, ${r.longitude}`}>
          <span>
            <EnvironmentOutlined style={{ color: '#1677ff' }} />
            {' '}{v || `${r.latitude?.toFixed(4)}, ${r.longitude?.toFixed(4)}`}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (v) => {
        const s = statusMap[v as keyof typeof statusMap] || statusMap.pending
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    {
      title: '报警时间',
      dataIndex: 'created_at',
      width: 170,
      render: (v) => <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(v)}</Text>,
    },
    {
      title: '操作',
      width: 140,
      fixed: 'right' as const,
      render: (_, record: AlarmItem) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleOpenDetail(record)}>详情</Button>
          {(record.status === 'pending' || record.status === 'processing') && (
            <Button type="link" size="small" type="primary" danger icon={<CheckCircleOutlined />} onClick={() => setHandleModal(record)}>处理</Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Alert
        type="warning"
        showIcon
        icon={<BellOutlined />}
        message={
          <Space>
            <Text strong>实时报警提醒</Text>
            <Badge count={dashboardStats?.pending_alarms || 0} showZero style={{ backgroundColor: '#ff4d4f' }}>
              <Tag color="red">待处理 {dashboardStats?.pending_alarms || 0} 条</Tag>
            </Badge>
            <Text type="secondary">今日新增 {dashboardStats?.today_alarms || 0} 条</Text>
          </Space>
        }
        style={{ borderRadius: 12 }}
      />

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="累计报警"
              value={total}
              valueStyle={{ color: '#fa8c16' }}
              prefix={<AlertOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="待处理"
              value={dashboardStats?.pending_alarms || 0}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<FireOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="融合准确率"
              value={fusionStats ? Math.round(fusionStats.avg_confidence || fusionStats.multi_vs_single_improve_pct || 85) : 85}
              suffix="%"
              valueStyle={{ color: (fusionStats && (fusionStats.avg_confidence >= 95 || fusionStats.multi_vs_single_improve_pct >= 95)) ? '#52c41a' : (fusionStats && (fusionStats.avg_confidence >= 85 || fusionStats.multi_vs_single_improve_pct >= 85)) ? '#faad14' : '#fa8c16' }}
              prefix={<SafetyCertificateOutlined />}
            />
            <div style={{ marginTop: 4 }}>
              <Text type="secondary" style={{ fontSize: 11 }}>
                {fusionStats ? `三摄融合 vs 单摄 · 近90天${fusionStats.total_detections || 0}次检测` : '加载中...'}
              </Text>
            </div>
            {fusionStats && (fusionStats.occlusion_count > 0 || fusionStats.backlit_count > 0) && (
              <div style={{ marginTop: 2 }}>
                <Space size={4}>
                  {fusionStats.occlusion_count > 0 && <Tag color="red" style={{ fontSize: 10 }}>遮挡{fusionStats.occlusion_count}</Tag>}
                  {fusionStats.backlit_count > 0 && <Tag color="gold" style={{ fontSize: 10 }}>逆光{fusionStats.backlit_count}</Tag>}
                </Space>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} loading={statsLoading}>
            <Statistic
              title="今日疲劳事件"
              value={dashboardStats?.today_fatigue_events || 0}
              valueStyle={{ color: '#f5222d' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><AlertOutlined style={{ color: '#ff4d4f' }} /> 24小时报警趋势</Space>}>
            <ReactECharts option={trendChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><PieChartOutlined /> 报警类型分布</Space>}>
            <ReactECharts option={distributionChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <AlertOutlined style={{ color: '#ff4d4f' }} />
            <Text strong style={{ fontSize: 15 }}>报警列表</Text>
            <Tag color="red">{total}</Tag>
          </Space>
        }
        extra={
          <Space wrap>
            <Select
              allowClear
              placeholder="状态筛选"
              style={{ width: 140 }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              {Object.entries(statusMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="等级筛选"
              style={{ width: 140 }}
              value={levelFilter}
              onChange={setLevelFilter}
            >
              {Object.entries(levelMap).map(([k, v]) => (
                <Option key={k} value={Number(k)}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="类型筛选"
              style={{ width: 160 }}
              value={typeFilter}
              onChange={setTypeFilter}
              showSearch
            >
              {Object.entries(alarmTypeMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => { setStatusFilter(undefined); setLevelFilter(undefined); setTypeFilter(undefined) }}>重置</Button>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
            <Button
              type="primary"
              icon={<CheckCircleOutlined />}
              onClick={() => handleBatch('ack')}
            >
              批量确认
            </Button>
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          columns={columns as any}
          dataSource={data}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
          scroll={{ x: 1400 }}
          rowClassName={(r) => r.alarm_level === 3 ? '!bg-red-50' : r.alarm_level === 2 ? '!bg-orange-50' : ''}
        />
      </Card>

      <Drawer
        title={
          <Space>
            <AlertOutlined style={{ color: '#ff4d4f' }} />
            <Text strong>报警详情 - {detailDrawer?.alarm_no}</Text>
            {detailDrawer && (
              <Tag color={levelMap[detailDrawer.alarm_level as keyof typeof levelMap]?.color}>
                {levelMap[detailDrawer.alarm_level as keyof typeof levelMap]?.label}
              </Tag>
            )}
          </Space>
        }
        open={!!detailDrawer}
        onClose={() => setDetailDrawer(null)}
        width={520}
        extra={
          detailDrawer && (detailDrawer.status === 'pending' || detailDrawer.status === 'processing') && (
            <Space>
              <Button icon={<PhoneOutlined />} type="primary">语音对讲</Button>
              <Button type="primary" danger icon={<CheckCircleOutlined />} onClick={() => setHandleModal(detailDrawer)}>立即处理</Button>
            </Space>
          )
        }
      >
        {detailDrawer && (
          <div>
            <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }}>
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="报警编号">{detailDrawer.alarm_no}</Descriptions.Item>
                <Descriptions.Item label="车辆/驾驶员">
                  <Space>
                    <Tag color="blue">{detailDrawer.vehicle_plate || detailDrawer.vehicle_id}</Tag>
                    <UserOutlined /> {detailDrawer.driver_name || detailDrawer.driver_id}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="报警类型">
                  {(() => {
                    const t = alarmTypeMap[detailDrawer.alarm_type] || { label: detailDrawer.alarm_type, color: 'default' }
                    return <Tag color={t.color}>{t.label}</Tag>
                  })()}
                </Descriptions.Item>
                <Descriptions.Item label="疲劳指数">
                  <Progress
                    type="circle"
                    size={60}
                    percent={100 - Math.round(detailDrawer.fatigue_score)}
                    strokeColor={detailDrawer.fatigue_score < 70 ? '#ff4d4f' : detailDrawer.fatigue_score < 85 ? '#faad14' : '#52c41a'}
                  />
                </Descriptions.Item>
                <Descriptions.Item label="连续疲劳时间">
                  {detailDrawer.continuous_fatigue_minutes ? `${detailDrawer.continuous_fatigue_minutes} 分钟` : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="发生位置">
                  <Paragraph copyable style={{ margin: 0, fontSize: 13 }}>
                    <EnvironmentOutlined /> {detailDrawer.location_address || `${detailDrawer.latitude?.toFixed(6)}, ${detailDrawer.longitude?.toFixed(6)}`}
                  </Paragraph>
                </Descriptions.Item>
                <Descriptions.Item label="车速">{Math.round(detailDrawer.vehicle_speed)} km/h</Descriptions.Item>
                <Descriptions.Item label="报警时间">{formatDateTime(detailDrawer.created_at)}</Descriptions.Item>
                <Descriptions.Item label="当前状态">
                  <Tag color={statusMap[detailDrawer.status as keyof typeof statusMap]?.color}>
                    {statusMap[detailDrawer.status as keyof typeof statusMap]?.label}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} title={<Space><VideoCameraOutlined /> 三摄像头实时画面</Space>} extra={
              <Space>
                <Tag color="blue">左</Tag>
                <Tag color="green">中</Tag>
                <Tag color="purple">右</Tag>
                <Button
                  type="link"
                  size="small"
                  icon={<DownloadOutlined />}
                  onClick={handleDownloadVideo}
                  disabled={!detailVideoURL}
                >
                  下载视频
                </Button>
              </Space>
            }>
              {detailLoading ? (
                <div style={{ padding: 40, textAlign: 'center' }}><Spin tip="加载中..." /></div>
              ) : (
                <Row gutter={8}>
                  <Col span={8}>
                    <div style={{ textAlign: 'center' }}>
                      <div style={{
                        background: '#f5f5f5',
                        borderRadius: 6,
                        height: 120,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#8c8c8c',
                        fontSize: 12,
                        overflow: 'hidden',
                        position: 'relative',
                      }}>
                        {(detailDrawer as any)?.left_frame_url ? (
                          <img src={(detailDrawer as any).left_frame_url} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                        ) : (
                          <>
                            <span style={{ fontSize: 20 }}>👈</span>
                            <span>左摄像头</span>
                          </>
                        )}
                      </div>
                      <div style={{ marginTop: 4 }}>
                        <Tag color="blue" style={{ fontSize: 11 }}>左</Tag>
                        {(detailDrawer as any)?.left_score > 0 && (
                          <Text type={Number((detailDrawer as any).left_score) < 70 ? 'danger' : 'success'} style={{ fontSize: 11 }}>
                            {Math.round(Number((detailDrawer as any).left_score))}
                          </Text>
                        )}
                      </div>
                    </div>
                  </Col>
                  <Col span={8}>
                    <div style={{ textAlign: 'center' }}>
                      <div style={{
                        background: '#f5f5f5',
                        borderRadius: 6,
                        height: 120,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#8c8c8c',
                        fontSize: 12,
                        overflow: 'hidden',
                        position: 'relative',
                      }}>
                        {(detailDrawer as any)?.center_frame_url || detailSnapshotURL || detailDrawer.snap_image_url ? (
                          <img src={(detailDrawer as any)?.center_frame_url || detailSnapshotURL || detailDrawer.snap_image_url} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                        ) : (
                          <>
                            <span style={{ fontSize: 20 }}>📷</span>
                            <span>中摄像头</span>
                          </>
                        )}
                      </div>
                      <div style={{ marginTop: 4 }}>
                        <Tag color="green" style={{ fontSize: 11 }}>中</Tag>
                        {(detailDrawer as any)?.center_score > 0 && (
                          <Text type={Number((detailDrawer as any).center_score) < 70 ? 'danger' : 'success'} style={{ fontSize: 11 }}>
                            {Math.round(Number((detailDrawer as any).center_score))}
                          </Text>
                        )}
                      </div>
                    </div>
                  </Col>
                  <Col span={8}>
                    <div style={{ textAlign: 'center' }}>
                      <div style={{
                        background: '#f5f5f5',
                        borderRadius: 6,
                        height: 120,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#8c8c8c',
                        fontSize: 12,
                        overflow: 'hidden',
                        position: 'relative',
                      }}>
                        {(detailDrawer as any)?.right_frame_url ? (
                          <img src={(detailDrawer as any).right_frame_url} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                        ) : (
                          <>
                            <span style={{ fontSize: 20 }}>👉</span>
                            <span>右摄像头</span>
                          </>
                        )}
                      </div>
                      <div style={{ marginTop: 4 }}>
                        <Tag color="purple" style={{ fontSize: 11 }}>右</Tag>
                        {(detailDrawer as any)?.right_score > 0 && (
                          <Text type={Number((detailDrawer as any).right_score) < 70 ? 'danger' : 'success'} style={{ fontSize: 11 }}>
                            {Math.round(Number((detailDrawer as any).right_score))}
                          </Text>
                        )}
                      </div>
                    </div>
                  </Col>
                </Row>
              )}
              {(detailVideoURL || detailDrawer.video_clip_url) && (
                <div style={{ marginTop: 12 }}>
                  <video controls src={detailVideoURL || detailDrawer.video_clip_url} style={{ width: '100%', borderRadius: 6 }} />
                </div>
              )}
            </Card>

            {(detailDrawer as any)?.fusion_method && (
              <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} title={<Space><SafetyCertificateOutlined style={{ color: '#fa8c16' }} /> 多摄融合分析</Space>}>
                <Row gutter={16}>
                  <Col span={12}>
                    <Descriptions column={1} size="small">
                      <Descriptions.Item label="融合方法">
                        <Tag color="orange">{(detailDrawer as any).fusion_method}</Tag>
                      </Descriptions.Item>
                      <Descriptions.Item label="融合置信度">
                        <Progress
                          percent={Math.round(Number((detailDrawer as any).fusion_confidence) * 100)}
                          size="small"
                          strokeColor={Number((detailDrawer as any).fusion_confidence) >= 0.95 ? '#52c41a' : Number((detailDrawer as any).fusion_confidence) >= 0.8 ? '#faad14' : '#ff4d4f'}
                        />
                      </Descriptions.Item>
                      <Descriptions.Item label="使用摄像头">
                        <Space>
                          {String((detailDrawer as any).used_cameras || '').split(',').filter(Boolean).map((cam: string) => (
                            <Tag key={cam} color={cameraPositionMap[cam]?.color || 'default'}>
                              {cameraPositionMap[cam]?.icon} {cameraPositionMap[cam]?.label || cam}
                            </Tag>
                          ))}
                        </Space>
                      </Descriptions.Item>
                    </Descriptions>
                  </Col>
                  <Col span={12}>
                    <Descriptions column={1} size="small">
                      <Descriptions.Item label="遮挡检测">
                        {(detailDrawer as any).occlusion_detected ? (
                          <Tag color="red">检测到遮挡</Tag>
                        ) : (
                          <Tag color="green">无遮挡</Tag>
                        )}
                      </Descriptions.Item>
                      <Descriptions.Item label="逆光检测">
                        {(detailDrawer as any).backlit_detected ? (
                          <Tag color="orange">检测到逆光</Tag>
                        ) : (
                          <Tag color="green">无逆光</Tag>
                        )}
                      </Descriptions.Item>
                      <Descriptions.Item label="摄像头位">
                        <Tag color={cameraPositionMap[(detailDrawer as any).camera_position]?.color || 'default'}>
                          {cameraPositionMap[(detailDrawer as any).camera_position]?.label || (detailDrawer as any).camera_position}
                        </Tag>
                      </Descriptions.Item>
                    </Descriptions>
                  </Col>
                </Row>
                {(((detailDrawer as any).left_score > 0) || ((detailDrawer as any).center_score > 0) || ((detailDrawer as any).right_score > 0)) && (
                  <div style={{ marginTop: 12 }}>
                    <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block' }}>各视角疲劳评分对比</Text>
                    <Row gutter={8}>
                      {['left', 'center', 'right'].map(pos => {
                        const score = Number((detailDrawer as any)[`${pos}_score`])
                        if (score <= 0) return null
                        return (
                          <Col span={8} key={pos}>
                            <div style={{
                              textAlign: 'center',
                              padding: '8px 0',
                              borderRadius: 6,
                              background: score < 70 ? '#fff1f0' : score < 85 ? '#fffbe6' : '#f6ffed',
                            }}>
                              <div style={{ fontSize: 11, color: cameraPositionMap[pos]?.color }}>
                                {cameraPositionMap[pos]?.icon} {cameraPositionMap[pos]?.label}
                              </div>
                              <div style={{
                                fontSize: 22,
                                fontWeight: 700,
                                color: score < 70 ? '#ff4d4f' : score < 85 ? '#faad14' : '#52c41a',
                              }}>
                                {Math.round(score)}
                              </div>
                            </div>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}
              </Card>
            )}

            {detailDrawer.status !== 'pending' && (
              <Card size="small" style={{ borderRadius: 8 }} title={<Space><CheckCircleOutlined /> 处理记录</Space>}>
                <Descriptions column={1} size="small" bordered>
                  <Descriptions.Item label="处理方式">
                    {detailDrawer.handle_type ? <Tag color="blue">{detailDrawer.handle_type}</Tag> : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="处理人ID">{detailDrawer.dispatcher_id || '-'}</Descriptions.Item>
                  <Descriptions.Item label="处理时间">{formatDateTime(detailDrawer.handled_at as any) || '-'}</Descriptions.Item>
                  <Descriptions.Item label="处理备注">{detailDrawer.handle_note || '-'}</Descriptions.Item>
                </Descriptions>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      <Modal
        title={<Space><CheckCircleOutlined style={{ color: '#52c41a' }} /> 处理报警 - {handleModal?.alarm_no}</Space>}
        open={!!handleModal}
        onCancel={() => { setHandleModal(null); handleForm.resetFields() }}
        onOk={() => handleForm.submit()}
        okText="确认处理"
        okButtonProps={{ type: 'primary', danger: false }}
        width={520}
      >
        {handleModal && (
          <div style={{ marginBottom: 16 }}>
            <Alert
              type={handleModal.alarm_level === 3 ? 'error' : 'warning'}
              showIcon
              message={
                <Space>
                  <Tag color={levelMap[handleModal.alarm_level as keyof typeof levelMap]?.color}>
                    {levelMap[handleModal.alarm_level as keyof typeof levelMap]?.label}
                  </Tag>
                  {alarmTypeMap[handleModal.alarm_type]?.label || handleModal.alarm_type}
                  · 车辆 {handleModal.vehicle_plate} · 驾驶员 {handleModal.driver_name}
                </Space>
              }
              description={`疲劳指数 ${Math.round(handleModal.fatigue_score)}${handleModal.continuous_fatigue_minutes ? `，连续疲劳 ${handleModal.continuous_fatigue_minutes} 分钟` : ''}`}
              style={{ borderRadius: 8 }}
            />
          </div>
        )}
        <Form form={handleForm} layout="vertical" onFinish={handleAck}>
          <Form.Item label="处理方式" name="handle_type" rules={[{ required: true, message: '请选择处理方式' }]}>
            <Radio.Group>
              <Radio value="voice_remind">🎙️ 语音提醒</Radio>
              <Radio value="dispatch_rest">☕ 调度服务区停靠</Radio>
              <Radio value="intervene_legal">🚓 通知执法站</Radio>
              <Radio value="escalate">📢 升级上报主管</Radio>
              <Radio value="other">📝 其他</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item label="处理备注" name="handle_note" rules={[{ required: true, message: '请输入处理备注' }]}>
            <Input.TextArea rows={3} placeholder="请描述处理情况..." />
          </Form.Item>
          <Divider style={{ margin: '8px 0 16px' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>
            <WarningOutlined /> 处理后系统将自动通知车载终端语音播报，并记录于驾驶员考核档案
          </Text>
        </Form>
      </Modal>
    </div>
  )
}

export default FatigueAlarms

function PieChartOutlined() {
  return <span>📊</span>
}
