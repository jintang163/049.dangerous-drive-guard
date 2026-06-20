import React, { useEffect, useState, useCallback } from 'react'
import {
  Row,
  Col,
  Card,
  Statistic,
  Tag,
  Progress,
  Empty,
  Spin,
  Space,
  Typography,
  Tooltip,
  Avatar,
  List,
  Badge,
  Descriptions,
} from 'antd'
import {
  TruckOutlined,
  AlertOutlined,
  UserOutlined,
  FileTextOutlined,
  FieldTimeOutlined,
  WarningOutlined,
  FireOutlined,
  RiseOutlined,
  EnvironmentOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import AMap from '@/components/AMap'
import { monitorApi, vehicleApi, fatigueApi } from '@/services/api'
import { useAppStore, StatData, AlarmItem, VehicleStatus } from '@/store/app'
import { formatDateTime, formatDistance } from '@/utils/auth'
import dayjs from 'dayjs'
import WebSocketManager from '@/services/ws'

const { Text, Title } = Typography

const Dashboard: React.FC = () => {
  const {
    stats,
    setStats,
    vehicles,
    updateVehicles,
    alarms,
    addAlarm,
    updateAlarm,
    fetchStats: fetchStoreStats,
    loading: storeLoading,
  } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [realtimeStats, setRealtimeStats] = useState<StatData | null>(null)

  const fetchDashboardData = useCallback(async () => {
    setLoading(true)
    try {
      const [statsData, vehiclesData] = await Promise.all([
        monitorApi.getDashboardStats(),
        vehicleApi.listRealtimeStatus(),
      ])
      setRealtimeStats(statsData)
      setStats(statsData)
      updateVehicles(vehiclesData || [])
    } finally {
      setLoading(false)
    }
  }, [setStats, updateVehicles])

  useEffect(() => {
    fetchDashboardData()
    const timer = setInterval(fetchDashboardData, 60000)

    const unsub1 = WebSocketManager.getInstance().on('new_alarm', async (alarm: AlarmItem) => {
      addAlarm(alarm)
      try {
        const snapshotRes = await fatigueApi.getSnapshotURL(alarm.id)
        if (snapshotRes?.url) {
          updateAlarm(alarm.id, { snap_image_url: snapshotRes.url })
        }
      } catch (e) {}
    })

    const unsub2 = WebSocketManager.getInstance().on('vehicle_status', (vehicle: VehicleStatus) => {
      updateVehicles(
        vehicles.map(v => (v.vehicle_id === vehicle.vehicle_id ? { ...v, ...vehicle } : v))
      )
    })

    const unsub3 = WebSocketManager.getInstance().on('alarm_updated', (alarm: Partial<AlarmItem>) => {
      if (alarm.id) {
        updateAlarm(alarm.id, alarm)
      }
    })

    return () => {
      clearInterval(timer)
      unsub1()
      unsub2()
      unsub3()
    }
  }, [fetchDashboardData, addAlarm, updateAlarm, updateVehicles, vehicles])

  const displayStats = realtimeStats || stats

  const trendChart = React.useMemo(() => {
    if (!displayStats?.daily_trend?.length) {
      return {
        tooltip: { trigger: 'axis' },
        legend: { data: ['报警数', '预警事件'], top: 0 },
        grid: { left: 40, right: 20, top: 40, bottom: 30 },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 7 }, (_, i) => dayjs().subtract(6 - i, 'day').format('MM-DD')),
          axisLine: { lineStyle: { color: '#e5e7eb' } },
        },
        yAxis: { type: 'value', axisLine: { show: false }, splitLine: { lineStyle: { color: '#f0f0f0' } } },
        series: [
          {
            name: '报警数',
            type: 'bar',
            data: Array.from({ length: 7 }, () => 0),
            itemStyle: { color: '#1677ff', borderRadius: [4, 4, 0, 0] },
            barWidth: 14,
          },
          {
            name: '预警事件',
            type: 'line',
            smooth: true,
            data: Array.from({ length: 7 }, () => 0),
            itemStyle: { color: '#fa8c16' },
            lineStyle: { width: 3 },
            areaStyle: { color: 'rgba(250,140,22,0.1)' },
          },
        ],
      }
    }
    return {
      tooltip: { trigger: 'axis' },
      legend: { data: ['报警数', '预警事件'], top: 0 },
      grid: { left: 40, right: 20, top: 40, bottom: 30 },
      xAxis: {
        type: 'category',
        data: displayStats.daily_trend.slice().reverse().map(d => dayjs(d.date).format('MM-DD')),
        axisLine: { lineStyle: { color: '#e5e7eb' } },
      },
      yAxis: { type: 'value', axisLine: { show: false }, splitLine: { lineStyle: { color: '#f0f0f0' } } },
      series: [
        {
          name: '报警数',
          type: 'bar',
          data: displayStats.daily_trend.slice().reverse().map(d => d.alarms),
          itemStyle: { color: '#1677ff', borderRadius: [4, 4, 0, 0] },
          barWidth: 14,
        },
        {
          name: '预警事件',
          type: 'line',
          smooth: true,
          data: displayStats.daily_trend.slice().reverse().map(d => d.events),
          itemStyle: { color: '#fa8c16' },
          lineStyle: { width: 3 },
          areaStyle: { color: 'rgba(250,140,22,0.1)' },
        },
      ],
    }
  }, [displayStats])

  const alarmTypeChart = React.useMemo(() => {
    const typeMap: Record<string, string> = {
      fatigue_perclos: '疲劳瞌睡',
      excessive_yawn: '频繁打哈欠',
      abnormal_head_posture: '异常姿态',
      gaze_distraction: '视线偏离',
      phone_usage: '使用手机',
      smoking: '抽烟',
      no_seatbelt: '未系安全带',
      continuous_fatigue: '连续疲劳',
    }
    if (!displayStats?.alarm_type_distribution?.length) {
      return {
        tooltip: { trigger: 'item' },
        legend: { type: 'scroll', orient: 'vertical', right: 10, top: 'center' },
        series: [
          {
            type: 'pie',
            radius: ['45%', '72%'],
            center: ['35%', '50%'],
            avoidLabelOverlap: false,
            itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
            label: { show: false },
            data: [],
          },
        ],
        color: ['#ff4d4f', '#fa8c16', '#faad14', '#a0d911', '#13c2c2', '#1677ff', '#722ed1', '#eb2f96'],
      }
    }
    return {
      tooltip: { trigger: 'item' },
      legend: { type: 'scroll', orient: 'vertical', right: 10, top: 'center' },
      series: [
        {
          type: 'pie',
          radius: ['45%', '72%'],
          center: ['35%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
          label: { show: false },
          data: displayStats.alarm_type_distribution.map(d => ({
            name: typeMap[d.alarm_type] || d.alarm_type,
            value: d.count,
          })),
        },
      ],
      color: ['#ff4d4f', '#fa8c16', '#faad14', '#a0d911', '#13c2c2', '#1677ff', '#722ed1', '#eb2f96'],
    }
  }, [displayStats])

  const vehicleMapData = React.useMemo(() => {
    return vehicles.map(v => ({
      position: [v.longitude, v.latitude],
      title: v.plate_number,
      color: v.marker_color || '#1677ff',
      status: v.status,
      info: {
        ...v,
        fatigueLevel: v.fatigue_level,
      },
    })).filter(v => v.position[0] && v.position[1])
  }, [vehicles])

  const topVehicles = React.useMemo(() => {
    return [...vehicles]
      .sort((a, b) => (b.fatigue_score || 100) - (a.fatigue_score || 100))
      .slice(0, 5)
  }, [vehicles])

  const recentAlarms = React.useMemo(() => alarms.slice(0, 5), [alarms])

  if (loading) {
    return (
      <div style={{ padding: 40, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    )
  }

  const StatCard = ({ icon, title, value, suffix, color, extra, trend, loading }: any) => (
    <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 20 }} loading={loading}>
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
        <div>
          <Text type="secondary" style={{ fontSize: 13 }}>{title}</Text>
          <div style={{ marginTop: 12, display: 'flex', alignItems: 'baseline', gap: 6 }}>
            <span style={{ fontSize: 30, fontWeight: 700, color }}>{value}</span>
            {suffix && <span style={{ fontSize: 14, color: '#8c8c8c' }}>{suffix}</span>}
          </div>
          {trend && (
            <div style={{ marginTop: 8 }}>
              <Tag color={trend > 0 ? 'green' : 'red'} style={{ fontSize: 12 }}>
                <RiseOutlined /> {Math.abs(trend)}% 较昨日
              </Tag>
            </div>
          )}
        </div>
        <div
          style={{
            width: 48,
            height: 48,
            borderRadius: 12,
            background: `${color}15`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 24,
            color,
          }}
        >
          {icon}
        </div>
      </div>
      {extra}
    </Card>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            icon={<TruckOutlined />}
            title="在线车辆"
            value={displayStats?.running_vehicles || 0}
            suffix={`/ ${displayStats?.total_vehicles || 0} 辆`}
            color="#1677ff"
            loading={loading}
            extra={
              <Progress
                percent={Math.round(((displayStats?.running_vehicles || 0) / Math.max(displayStats?.total_vehicles || 1, 1)) * 100)}
                showInfo={false}
                size="small"
                strokeColor="#1677ff"
                style={{ marginTop: 16 }}
              />
            }
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            icon={<UserOutlined />}
            title="在岗驾驶员"
            value={displayStats?.total_drivers || 0}
            suffix="人"
            color="#52c41a"
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            icon={<FileTextOutlined />}
            title="在途运单"
            value={displayStats?.in_transit_waybills || 0}
            suffix={`/ ${displayStats?.total_waybills || 0}`}
            color="#722ed1"
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            icon={<AlertOutlined />}
            title="待处理报警"
            value={displayStats?.pending_alarms || 0}
            suffix={`今日 ${displayStats?.today_alarms || 0}`}
            color="#ff4d4f"
            loading={loading}
          />
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} sm={12} md={3}>
          <StatCard
            icon={<FieldTimeOutlined />}
            title="今日里程"
            value={displayStats?.today_mileage_km?.toFixed?.(1) || '0'}
            suffix="km"
            color="#13c2c2"
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12} md={3}>
          <StatCard
            icon={<WarningOutlined />}
            title="今日预警事件"
            value={displayStats?.today_fatigue_events || 0}
            suffix="起"
            color="#fa8c16"
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12} md={3}>
          <StatCard
            icon={<SafetyOutlined />}
            title="待处理救援"
            value={displayStats?.rescue_stats?.pending || 0}
            suffix={`/ ${displayStats?.rescue_stats?.total || 0}`}
            color="#eb2f96"
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12} md={3}>
          <StatCard
            icon={<ThunderboltOutlined />}
            title="活跃天气预警"
            value={displayStats?.weather_alerts?.active || 0}
            suffix="条"
            color="#faad14"
            loading={loading}
          />
        </Col>
        <Col xs={24} md={12}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
              <Space>
                <FireOutlined style={{ color: '#1677ff', fontSize: 18 }} />
                <Text strong style={{ fontSize: 15 }}>安全态势 - 近7日趋势</Text>
              </Space>
              <Tag color="blue" style={{ margin: 0 }}>实时刷新</Tag>
            </div>
            {displayStats?.daily_trend?.length ? (
              <ReactECharts option={trendChart} style={{ height: 180 }} notMerge lazyUpdate />
            ) : (
              <Empty description="暂无数据" style={{ padding: 20 }} />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card bordered={false} style={{ borderRadius: 12, height: 500 }} bodyStyle={{ padding: 0, height: '100%' }}>
            <div style={{ padding: '16px 20px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid #f0f0f0' }}>
              <Space>
                <EnvironmentOutlined style={{ color: '#1677ff', fontSize: 18 }} />
                <Text strong style={{ fontSize: 15 }}>车辆实时分布</Text>
              </Space>
              <Space size={12}>
                <Space><span style={{ width: 10, height: 10, borderRadius: '50%', background: '#52c41a', display: 'inline-block' }} /> 正常</Space>
                <Space><span style={{ width: 10, height: 10, borderRadius: '50%', background: '#faad14', display: 'inline-block' }} /> 预警</Space>
                <Space><span style={{ width: 10, height: 10, borderRadius: '50%', background: '#ff4d4f', display: 'inline-block' }} /> 疲劳</Space>
                <Space><span style={{ width: 10, height: 10, borderRadius: '50%', background: '#8c8c8c', display: 'inline-block' }} /> 离线</Space>
              </Space>
            </div>
            <AMap
              style={{ height: 'calc(100% - 54px)', borderRadius: '0 0 12px 12px' }}
              markers={vehicleMapData}
              center={[116.4074, 39.9042]}
              zoom={11}
            />
          </Card>
        </Col>
        <Col xs={24} lg={8} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Card bordered={false} style={{ borderRadius: 12, flex: 1 }} title={<Space><AlertOutlined style={{ color: '#ff4d4f' }} /> 报警类型分布</Card>}>
            {displayStats?.alarm_type_distribution?.length ? (
              <ReactECharts option={alarmTypeChart} style={{ height: 220 }} notMerge />
            ) : (
              <Empty description="暂无报警" style={{ padding: 40 }} />
            )}
          </Card>
          <Card
            bordered={false}
            style={{ borderRadius: 12, flex: 1 }}
            title={<Space><UserOutlined style={{ color: '#fa8c16' }} /> 高风险驾驶员TOP5</Card>}
          >
            {topVehicles.length ? (
              <List
                size="small"
                dataSource={topVehicles}
                renderItem={(v, idx) => (
                  <List.Item style={{ padding: '8px 0' }}>
                    <div style={{ width: '100%' }}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
                        <Space>
                          <Tag color={idx < 2 ? 'red' : idx < 4 ? 'orange' : 'blue'}>#{idx + 1}</Tag>
                          <Text strong>{v.driver_name}</Text>
                        </Space>
                        <Badge
                          status={v.fatigue_level === 'fatigue' ? 'error' : v.fatigue_level === 'warning' ? 'warning' : 'success'}
                          text={v.plate_number}
                        />
                      </div>
                      <Progress
                        percent={100 - Math.round(v.fatigue_score || 0)}
                        showInfo
                        formatText={() => (
                          <Text type={v.fatigue_score < 70 ? 'danger' : v.fatigue_score < 85 ? 'warning' : 'success'} strong>
                            风险 {(100 - Math.round(v.fatigue_score || 0))}%
                          </Text>
                        )}
                        strokeColor={v.fatigue_score < 70 ? '#ff4d4f' : v.fatigue_score < 85 ? '#fa8c16' : '#52c41a'}
                        size="small"
                      />
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无数据" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={
              <Space>
                <AlertOutlined style={{ color: '#ff4d4f' }} />
                <Text strong style={{ fontSize: 15 }}>最新报警</Text>
                <Badge count={alarms.length} showZero style={{ backgroundColor: '#ff4d4f' }} />
              </Space>
            }
          >
            {recentAlarms.length ? (
              <List
                dataSource={recentAlarms}
                renderItem={(alarm) => (
                  <List.Item
                    style={{ padding: '12px 0', borderBottom: '1px dashed #f0f0f0' }}
                    actions={[
                      <Tag key="lvl" color={alarm.alarm_level === 3 ? 'red' : alarm.alarm_level === 2 ? 'orange' : 'blue'}>
                        {alarm.alarm_level === 3 ? '严重' : alarm.alarm_level === 2 ? '警告' : '提醒'}
                      </Tag>,
                      <Tag key="st" color={
                        alarm.status === 'pending' ? 'red' :
                          alarm.status === 'processing' ? 'orange' :
                            alarm.status === 'acknowledged' ? 'blue' : 'green'
                      }>
                        {alarm.status === 'pending' ? '待处理' :
                          alarm.status === 'processing' ? '处理中' :
                            alarm.status === 'acknowledged' ? '已确认' : '已关闭'}
                      </Tag>,
                    ]}
                  >
                    <List.Item.Meta
                      avatar={
                        <Avatar
                          style={{ backgroundColor: alarm.alarm_level === 3 ? '#ff4d4f' : '#fa8c16' }}
                          icon={<WarningOutlined />}
                        />
                      }
                      title={
                        <Space>
                          <Text strong>{alarm.vehicle_plate}</Text>
                          <Text type="secondary">{alarm.driver_name}</Text>
                        </Space>
                      }
                      description={
                        <Space size={24} wrap>
                          <span><EnvironmentOutlined /> {alarm.location_address || `${alarm.latitude?.toFixed(4)}, ${alarm.longitude?.toFixed(4)}`}</span>
                          <span>疲劳指数: <Text strong type={alarm.fatigue_score < 70 ? 'danger' : 'warning'}>{Math.round(alarm.fatigue_score)}</Text></span>
                          <span style={{ color: '#8c8c8c', fontSize: 12 }}>{formatDateTime(alarm.created_at)}</span>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无报警" />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card
            bordered={false}
            style={{ borderRadius: 12, height: '100%' }}
            title={
              <Space>
                <TruckOutlined style={{ color: '#1677ff' }} />
                <Text strong style={{ fontSize: 15 }}>在途车辆列表</Text>
              </Space>
            }
          >
            {vehicles.length ? (
              <List
                size="small"
                dataSource={vehicles.filter(v => v.status === 'running').slice(0, 8)}
                renderItem={(v) => (
                  <List.Item style={{ padding: '10px 0' }}>
                    <div style={{ width: '100%' }}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <Space>
                          <Badge
                            color={v.marker_color || '#52c41a'}
                            text={<Text strong>{v.plate_number}</Text>}
                          />
                        </Space>
                        <Space>
                          <Text type="secondary" style={{ fontSize: 12 }}>{Math.round(v.speed)} km/h</Text>
                        </Space>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>{v.driver_name}</Text>
                        <Text
                          type={v.fatigue_level === 'fatigue' ? 'danger' : v.fatigue_level === 'warning' ? 'warning' : 'success'}
                          style={{ fontSize: 12 }}
                          strong
                        >
                          状态: {v.fatigue_level === 'fatigue' ? '疲劳' : v.fatigue_level === 'warning' ? '预警' : '正常'} {Math.round(v.fatigue_score || 0)}
                        </Text>
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无在途车辆" />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
