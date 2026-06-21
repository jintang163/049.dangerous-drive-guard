import React, { useState, useCallback, useEffect } from 'react'
import {
  Row,
  Col,
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  Typography,
  Tag,
  Divider,
  Statistic,
  Progress,
  Tabs,
  List,
  Descriptions,
  Tooltip,
  message,
  Radio,
  Empty,
  Spin,
  Alert,
  Modal,
  Drawer,
  Badge,
  Timeline,
  Table,
  DatePicker,
} from 'antd'
import {
  EnvironmentOutlined,
  CarOutlined,
  RouteOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
  BankOutlined,
  ExperimentOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  ZoomInOutlined,
  WarningTwoTone,
  HistoryOutlined,
  AlertOutlined,
  EyeOutlined,
  InfoCircleOutlined,
  BellOutlined,
} from '@ant-design/icons'
import AMap from '@/components/AMap'
import { routeApi, trafficApi, replanApi } from '@/services/api'
import { formatDistance, formatDuration } from '@/utils/auth'
import WebSocketManager from '@/services/ws'

const { RangePicker } = DatePicker

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs

interface RestrictedArea {
  id: number
  name: string
  area_type: string
  level: number
  center_latitude: number
  center_longitude: number
  radius: number
}

interface RouteResult {
  id?: number
  plan_no: string
  strategy: string
  origin: { latitude: number; longitude: number; address: string }
  destination: { latitude: number; longitude: number; address: string }
  route_path: Array<{ lat: number; lng: number }>
  total_distance: number
  estimated_duration: number
  expected_speed: number
  toll_fee: number
  fuel_cost: number
  avoid_tunnels: number
  avoid_bridges: number
  avoid_populated: number
  avoid_water_protection: number
  restricted_segments: Array<{
    area_id: number
    area_name: string
    area_type: string
    level: number
    distance: number
    reason: string
    suggestion: string
  }>
  safety_score: number
}

type StrategyType = 'shortest' | 'safest' | 'economic'

interface DeviationEvent {
  waybill_id: number
  vehicle_plate: string
  current_lat: number
  current_lng: number
  planned_route_id: number
  deviation_distance: number
}

const RoutePlan: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [routes, setRoutes] = useState<Record<StrategyType, RouteResult | null>>({
    shortest: null,
    safest: null,
    economic: null,
  })
  const [activeStrategy, setActiveStrategy] = useState<StrategyType>('safest')
  const [restrictedAreas, setRestrictedAreas] = useState<RestrictedArea[]>([])
  const [originPicked, setOriginPicked] = useState<[number, number] | null>(null)
  const [destPicked, setDestPicked] = useState<[number, number] | null>(null)
  const [pickingMode, setPickingMode] = useState<'origin' | 'dest' | null>(null)
  const [planId, setPlanId] = useState<number | null>(null)
  const [deviationModal, setDeviationModal] = useState(false)
  const [currentDeviation, setCurrentDeviation] = useState<DeviationEvent | null>(null)
  const [replanLoading, setReplanLoading] = useState(false)

  const [trafficEvents, setTrafficEvents] = useState<any[]>([])
  const [trafficPage, setTrafficPage] = useState(1)
  const [trafficTotal, setTrafficTotal] = useState(0)
  const [trafficLoading, setTrafficLoading] = useState(false)
  const [eventDetail, setEventDetail] = useState<any | null>(null)

  const [replanPage, setReplanPage] = useState(1)
  const [replanPageSize] = useState(10)
  const [replanTotal, setReplanTotal] = useState(0)
  const [replanList, setReplanList] = useState<any[]>([])
  const [replanListLoading, setReplanListLoading] = useState(false)
  const [replanDetail, setReplanDetail] = useState<any | null>(null)

  const [replanSuggestModal, setReplanSuggestModal] = useState(false)
  const [currentSuggestion, setCurrentSuggestion] = useState<any | null>(null)
  const [confirmLoading, setConfirmLoading] = useState(false)

  const [activeTab, setActiveTab] = useState<'result' | 'traffic' | 'history'>('result')
  const [replanFilterStatus, setReplanFilterStatus] = useState<string>('')
  const [replanFilterTrigger, setReplanFilterTrigger] = useState<string>('')
  const [replanDateRange, setReplanDateRange] = useState<any>(null)

  const fetchRestricted = useCallback(async () => {
    try {
      const data = await routeApi.listRestrictedAreas()
      setRestrictedAreas(data || [])
    } catch (e) { }
  }, [])

  const fetchTrafficEvents = useCallback(async () => {
    setTrafficLoading(true)
    try {
      const res: any = await trafficApi.listEvents({ status: 'active', page: trafficPage, page_size: 20 })
      setTrafficEvents(res?.list || res?.items || [])
      setTrafficTotal(res?.total || 0)
    } catch (e) { } finally {
      setTrafficLoading(false)
    }
  }, [trafficPage])

  const fetchReplanHistory = useCallback(async () => {
    setReplanListLoading(true)
    try {
      const params: any = { page: replanPage, page_size: replanPageSize }
      if (replanFilterStatus) params.status = replanFilterStatus
      if (replanFilterTrigger) params.trigger_type = replanFilterTrigger
      if (replanDateRange && replanDateRange.length === 2) {
        params.start_date = replanDateRange[0].format('YYYY-MM-DD')
        params.end_date = replanDateRange[1].format('YYYY-MM-DD')
      }
      const res: any = await replanApi.list(params)
      setReplanList(res?.list || res?.items || [])
      setReplanTotal(res?.total || 0)
    } catch (e) { } finally {
      setReplanListLoading(false)
    }
  }, [replanPage, replanPageSize, replanFilterStatus, replanFilterTrigger, replanDateRange])

  useEffect(() => {
    fetchRestricted()
    fetchTrafficEvents()

    const unsub1 = WebSocketManager.getInstance().on('route_deviation', (deviation: DeviationEvent) => {
      setCurrentDeviation(deviation)
      setDeviationModal(true)
      message.warning({
        content: (
          <div>
            <div style={{ fontWeight: 600 }}>
              <WarningTwoTone /> 车辆偏航：{deviation.vehicle_plate}
            </div>
            <div style={{ fontSize: 12, marginTop: 4 }}>
              偏离路线 {deviation.deviation_distance.toFixed(1)} 米
            </div>
          </div>
        ),
        duration: 10,
      })
    })

    const unsub2 = WebSocketManager.getInstance().on('route_replan_suggest', (data: any) => {
      setCurrentSuggestion(data)
      setReplanSuggestModal(true)
      message.info({
        content: (
          <div>
            <div style={{ fontWeight: 600 }}>
              <BellOutlined /> 重规划建议：{data?.vehicle_plate || '车辆'}
            </div>
            <div style={{ fontSize: 12, marginTop: 4 }}>
              {data?.trigger_reason}，预计新增 {Math.max(0, data?.duration_delta || 0)} 分钟
            </div>
          </div>
        ),
        duration: 15,
      })
      fetchReplanHistory()
    })

    const unsub3 = WebSocketManager.getInstance().on('traffic_event', () => {
      fetchTrafficEvents()
    })

    const unsub4 = WebSocketManager.getInstance().on('route_applied', () => {
      fetchReplanHistory()
    })

    return () => {
      unsub1(); unsub2(); unsub3(); unsub4()
    }
  }, [fetchRestricted, fetchTrafficEvents, fetchReplanHistory])

  useEffect(() => {
    fetchTrafficEvents()
  }, [trafficPage, fetchTrafficEvents])

  useEffect(() => {
    fetchReplanHistory()
  }, [replanPage, replanFilterStatus, replanFilterTrigger, replanDateRange, fetchReplanHistory])

  const handleSubmit = async (values: any) => {
    setLoading(true)
    try {
      const payload = {
        origin: {
          address: values.origin_address,
          latitude: values.origin_lat,
          longitude: values.origin_lng,
        },
        destination: {
          address: values.dest_address,
          latitude: values.dest_lat,
          longitude: values.dest_lng,
        },
        vehicle_type: values.vehicle_type,
        vehicle_height: values.vehicle_height,
        vehicle_weight: values.vehicle_weight,
        hazard_class: values.hazard_class,
      }

      const result = await routeApi.planMultiStrategy(payload)
      if (result && typeof result === 'object') {
        const routesData = result as any
        setRoutes({
          shortest: routesData.shortest || null,
          safest: routesData.safest || null,
          economic: routesData.economic || null,
        })
        setPlanId(routesData.plan_id || null)
      } else {
        const strategies: StrategyType[] = ['shortest', 'safest', 'economic']
        const results = await Promise.all(
          strategies.map(s =>
            routeApi.plan({
              start: { lat: payload.origin.latitude, lng: payload.origin.longitude, address: payload.origin.address },
              end: { lat: payload.destination.latitude, lng: payload.destination.longitude, address: payload.destination.address },
              vehicle_type: payload.vehicle_type,
              danger_level: Number(payload.hazard_class),
              strategy: s,
            }).catch(() => null)
          )
        )
        setRoutes({
          shortest: results[0] as unknown as RouteResult,
          safest: results[1] as unknown as RouteResult,
          economic: results[2] as unknown as RouteResult,
        })
      }
      message.success('路径规划完成')
    } catch (e: any) {
      message.error(e.message || '规划失败')
    } finally {
      setLoading(false)
    }
  }

  const handleReplan = async () => {
    if (!currentDeviation) return
    setReplanLoading(true)
    try {
      const result = await routeApi.replan({
        waybill_id: currentDeviation.waybill_id,
        current_latitude: currentDeviation.current_lat,
        current_longitude: currentDeviation.current_lng,
        route_id: currentDeviation.planned_route_id,
      })
      if (result) {
        message.success('偏航重规划完成')
        setDeviationModal(false)
      }
    } catch (e) {
      message.error('重规划失败')
    } finally {
      setReplanLoading(false)
    }
  }

  const handleConfirmReplan = async (action: 'confirm' | 'reject') => {
    if (!currentSuggestion) return
    setConfirmLoading(true)
    try {
      await replanApi.confirm(currentSuggestion.replan_id, { action })
      message.success(action === 'confirm' ? '已确认，新路线已应用' : '已拒绝重规划建议')
      setReplanSuggestModal(false)
      fetchReplanHistory()
    } catch (e: any) {
      message.error(e.message || '操作失败')
    } finally {
      setConfirmLoading(false)
    }
  }

  const handleViewReplanDetail = async (id: number) => {
    try {
      const detail = await replanApi.get(id)
      setReplanDetail(detail)
    } catch (e: any) {
      message.error(e.message || '获取详情失败')
    }
  }

  const handleResolveEvent = async (id: number) => {
    try {
      await trafficApi.resolveEvent(id)
      message.success('事件已标记为解决')
      fetchTrafficEvents()
    } catch (e: any) {
      message.error(e.message || '操作失败')
    }
  }

  const triggerTypeLabel: Record<string, { label: string; color: string; icon: any }> = {
    traffic: { label: '路况事件', color: 'red', icon: <AlertOutlined /> },
    deviation: { label: '车辆偏航', color: 'orange', icon: <WarningOutlined /> },
    restricted: { label: '禁行区变更', color: 'purple', icon: <SafetyCertificateOutlined /> },
    manual: { label: '手动触发', color: 'blue', icon: <RouteOutlined /> },
  }

  const statusLabel: Record<string, { label: string; color: string }> = {
    pending: { label: '待确认', color: 'orange' },
    confirmed: { label: '已确认', color: 'green' },
    rejected: { label: '已拒绝', color: 'red' },
    auto_applied: { label: '系统自动应用', color: 'purple' },
    cancelled: { label: '已取消', color: 'default' },
  }

  const handleMapClick = (lng: number, lat: number) => {
    if (pickingMode === 'origin') {
      form.setFieldsValue({
        origin_lat: Number(lat.toFixed(6)),
        origin_lng: Number(lng.toFixed(6)),
        origin_address: `坐标 (${lat.toFixed(4)}, ${lng.toFixed(4)})`,
      })
      setOriginPicked([lng, lat])
      setPickingMode(null)
    } else if (pickingMode === 'dest') {
      form.setFieldsValue({
        dest_lat: Number(lat.toFixed(6)),
        dest_lng: Number(lng.toFixed(6)),
        dest_address: `坐标 (${lat.toFixed(4)}, ${lng.toFixed(4)})`,
      })
      setDestPicked([lng, lat])
      setPickingMode(null)
    }
  }

  const currentRoute = routes[activeStrategy]

  const renderStrategyResult = (s: typeof strategyOptions[number]) => {
    const r = routes[s.key]
    if (!r) {
      return (
        <div style={{ padding: 40, textAlign: 'center' }}>
          <Empty description="请先点击「开始三策略规划」" />
        </div>
      )
    }
    return (
      <div style={{ padding: '0 20px 20px', overflow: 'auto', height: '100%' }}>
        <Row gutter={12} style={{ marginBottom: 16 }}>
          <Col span={12}>
            <Card size="small" style={{ borderRadius: 8 }}>
              <Statistic
                title="总里程"
                value={formatDistance(r.total_distance)}
                valueStyle={{ fontSize: 18, color: s.color }}
              />
            </Card>
          </Col>
          <Col span={12}>
            <Card size="small" style={{ borderRadius: 8 }}>
              <Statistic
                title="预计时长"
                value={formatDuration(r.estimated_duration)}
                valueStyle={{ fontSize: 18, color: s.color }}
              />
            </Card>
          </Col>
        </Row>
        <Row gutter={12} style={{ marginBottom: 16 }}>
          <Col span={12}>
            <Card size="small" style={{ borderRadius: 8 }}>
              <Statistic
                title="过路费"
                value={r.toll_fee || 0}
                suffix="元"
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
          <Col span={12}>
            <Card size="small" style={{ borderRadius: 8 }}>
              <Statistic
                title="油费估算"
                value={r.fuel_cost || 0}
                suffix="元"
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
        </Row>
        <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} title="安全评分">
          <Progress
            percent={r.safety_score}
            showInfo
            strokeColor={r.safety_score >= 80 ? '#52c41a' : r.safety_score >= 60 ? '#faad14' : '#ff4d4f'}
            trailColor="#f5f5f5"
          />
        </Card>
        <Divider orientation="left" style={{ margin: '12px 0' }}>绕行统计</Divider>
        <Row gutter={8} style={{ marginBottom: 12 }}>
          <Col span={12}>
            <Text type="secondary" style={{ fontSize: 12 }}>🏥 避开人口密集</Text>
            <div style={{ fontWeight: 600, marginTop: 2 }}>{r.avoid_populated || 0} 处</div>
          </Col>
          <Col span={12}>
            <Text type="secondary" style={{ fontSize: 12 }}>🌊 避开水源保护</Text>
            <div style={{ fontWeight: 600, marginTop: 2 }}>{r.avoid_water_protection || 0} 处</div>
          </Col>
        </Row>
        <Row gutter={8} style={{ marginBottom: 12 }}>
          <Col span={12}>
            <Text type="secondary" style={{ fontSize: 12 }}>🚇 避开隧道</Text>
            <div style={{ fontWeight: 600, marginTop: 2 }}>{r.avoid_tunnels || 0} 处</div>
          </Col>
          <Col span={12}>
            <Text type="secondary" style={{ fontSize: 12 }}>🌉 避开桥梁</Text>
            <div style={{ fontWeight: 600, marginTop: 2 }}>{r.avoid_bridges || 0} 处</div>
          </Col>
        </Row>
        <Divider orientation="left" style={{ margin: '12px 0' }}>
          <Space>
            <WarningOutlined style={{ color: '#fa8c16' }} />
            禁行路段提示
            <Tag color="red">{r.restricted_segments?.length || 0}</Tag>
          </Space>
        </Divider>
        {r.restricted_segments && r.restricted_segments.length > 0 ? (
          <List
            size="small"
            dataSource={r.restricted_segments}
            renderItem={(seg: any) => (
              <List.Item style={{ padding: '8px 0' }}>
                <List.Item.Meta
                  avatar={<Badge status={seg.level === 2 ? 'error' : 'warning'} />}
                  title={
                    <Space>
                      <Text strong>{seg.area_name}</Text>
                      <Tag color={seg.level === 2 ? 'red' : 'orange'}>
                        {seg.level === 2 ? '严重' : '注意'}
                      </Tag>
                    </Space>
                  }
                  description={
                    <div>
                      <div style={{ fontSize: 12, color: '#595959', marginBottom: 4 }}>
                        {seg.reason} · {seg.distance?.toFixed(1)} km
                      </div>
                      <Text type="success" style={{ fontSize: 12 }}>💡 {seg.suggestion}</Text>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        ) : (
          <Empty description="无禁行路段，路径非常安全！" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        )}
      </div>
    )
  }

  const renderTrafficPanel = () => (
    <div style={{ padding: 16, overflow: 'auto', height: '100%' }}>
      <Space style={{ marginBottom: 12, width: '100%' }}>
        <Button size="small" icon={<ReloadOutlined />} onClick={fetchTrafficEvents}>刷新</Button>
        <Tag color="red">严重 {trafficEvents.filter(e => e.event_level >= 3).length}</Tag>
        <Tag color="orange">中等 {trafficEvents.filter(e => e.event_level === 2).length}</Tag>
      </Space>
      {trafficEvents.length === 0 ? (
        <Empty description="暂无活跃路况事件" />
      ) : (
        <List
          size="small"
          dataSource={trafficEvents}
          renderItem={(evt: any) => (
            <List.Item style={{ padding: '10px 0', borderBottom: '1px solid #f0f0f0' }}>
              <List.Item.Meta
                avatar={
                  <AlertOutlined style={{
                    fontSize: 20,
                    color: evt.event_level >= 3 ? '#ff4d4f' : evt.event_level === 2 ? '#fa8c16' : '#8c8c8c',
                  }} />
                }
                title={
                  <Space>
                    <Text strong>{evt.title}</Text>
                    {evt.event_level >= 3 && <Tag color="red">严重</Tag>}
                    {evt.event_level === 2 && <Tag color="orange">中等</Tag>}
                  </Space>
                }
                description={
                  <div style={{ fontSize: 12, color: '#595959' }}>
                    <div>🛣️ {evt.road_name || '未知路段'}</div>
                    <div>⏱️ 预计延误 {evt.duration_minutes || 0} 分钟 · 平均车速 {evt.avg_speed_kmh || 0} km/h</div>
                    <div style={{ marginTop: 4 }}>
                      <Button size="small" type="link" style={{ padding: 0 }} onClick={() => setEventDetail(evt)}>
                        详情
                      </Button>
                      <Button size="small" type="link" danger style={{ paddingLeft: 12 }} onClick={() => handleResolveEvent(evt.id)}>
                        标记解决
                      </Button>
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}
    </div>
  )

  const renderReplanHistory = () => (
    <div style={{ padding: 16, overflow: 'auto', height: '100%' }}>
      <Space style={{ marginBottom: 12 }} wrap>
        <Select
          size="small"
          placeholder="状态"
          allowClear
          style={{ width: 110 }}
          value={replanFilterStatus || undefined}
          onChange={setReplanFilterStatus}
        >
          <Option value="pending">待确认</Option>
          <Option value="confirmed">已确认</Option>
          <Option value="rejected">已拒绝</Option>
        </Select>
        <Select
          size="small"
          placeholder="触发类型"
          allowClear
          style={{ width: 110 }}
          value={replanFilterTrigger || undefined}
          onChange={setReplanFilterTrigger}
        >
          <Option value="traffic">路况</Option>
          <Option value="deviation">偏航</Option>
          <Option value="restricted">禁行</Option>
          <Option value="manual">手动</Option>
        </Select>
      </Space>
      {replanList.length === 0 ? (
        <Empty description="暂无重规划记录" />
      ) : (
        <Timeline
          mode="left"
          items={replanList.map((r: any) => {
            const trigger = triggerTypeLabel[r.trigger_type] || { label: r.trigger_type, color: 'default', icon: <InfoCircleOutlined /> }
            const status = statusLabel[r.status] || { label: r.status, color: 'default' }
            return {
              color: trigger.color,
              dot: trigger.icon,
              children: (
                <Card size="small" style={{ borderRadius: 8 }}>
                  <Space direction="vertical" size={2} style={{ width: '100%' }}>
                    <Space>
                      <Tag color={trigger.color}>{trigger.label}</Tag>
                      <Tag color={status.color}>{status.label}</Tag>
                      <Text code style={{ fontSize: 11 }}>{r.replan_no}</Text>
                    </Space>
                    <div style={{ fontSize: 13, color: '#1f1f1f' }}>{r.trigger_reason}</div>
                    <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                      {r.vehicle_plate} · {r.waybill_no}
                      {r.duration_delta !== undefined && r.duration_delta !== 0 && (
                        <span style={{ marginLeft: 8, color: r.duration_delta >= 0 ? '#ff4d4f' : '#52c41a' }}>
                          · {r.duration_delta >= 0 ? '+' : ''}{r.duration_delta} 分钟
                        </span>
                      )}
                    </div>
                    <div style={{ fontSize: 11, color: '#bfbfbf' }}>
                      {new Date(r.created_at).toLocaleString()}
                    </div>
                    <Button type="link" size="small" icon={<EyeOutlined />} style={{ padding: 0 }} onClick={() => handleViewReplanDetail(r.id)}>
                      查看详情
                    </Button>
                  </Space>
                </Card>
              ),
            }
          })}
        />
      )}
    </div>
  )

  const routeMarkers = [
    ...(originPicked ? [{
      position: originPicked as [number, number],
      title: '起点',
      color: '#52c41a',
    }] : []),
    ...(destPicked ? [{
      position: destPicked as [number, number],
      title: '终点',
      color: '#1677ff',
    }] : []),
    ...(currentRoute ? currentRoute.restricted_segments.map((seg, i) => {
      const area = restrictedAreas.find(a => a.id === seg.area_id)
      return {
        position: [
          area?.center_longitude || 116.4 + i * 0.05,
          area?.center_latitude || 39.9 + i * 0.03,
        ] as [number, number],
        title: seg.area_name,
        color: seg.level === 2 ? '#ff4d4f' : '#fa8c16',
        info: seg,
      }
    }) : []),
    ...trafficEvents.filter(e => e.center_lat && e.center_lng).map(evt => ({
      position: [evt.center_lng, evt.center_lat] as [number, number],
      title: evt.title,
      color: evt.event_level >= 3 ? '#ff4d4f' : evt.event_level === 2 ? '#fa8c16' : '#8c8c8c',
      info: evt,
    })),
  ]

  const routePolylines = currentRoute?.route_path ? [{
    path: currentRoute.route_path.map(p => [p.lng, p.lat]) as [number, number][],
    color: activeStrategy === 'shortest' ? '#1677ff' :
      activeStrategy === 'safest' ? '#52c41a' : '#722ed1',
    weight: 7,
  }] : []

  const restrictedPolygons = restrictedAreas.slice(0, 50).map(area => {
    let path: [number, number][] = []
    try {
      const boundary = (area as any).boundary_polygon
      if (boundary?.coordinates?.[0]) {
        path = boundary.coordinates[0].map((p: number[]) => [p[0], p[1]] as [number, number])
      } else if (area.center_latitude && area.center_longitude) {
        const r = (area.radius || 500) / 111000
        path = Array.from({ length: 24 }, (_, i) => {
          const a = (i / 24) * Math.PI * 2
          return [area.center_longitude + r * Math.cos(a), area.center_latitude + r * Math.sin(a)] as [number, number]
        })
      }
    } catch (e) {
      const r = ((area as any).radius || 500) / 111000
      path = Array.from({ length: 24 }, (_, i) => {
        const a = (i / 24) * Math.PI * 2
        return [area.center_longitude + r * Math.cos(a), area.center_latitude + r * Math.sin(a)] as [number, number]
      })
    }
    return {
      path,
      fillColor: area.level === 2 ? '#ff4d4f' : '#fa8c16',
      strokeColor: area.level === 2 ? '#ff4d4f' : '#fa8c16',
      fillOpacity: 0.08,
      strokeWeight: 1,
    }
  })

  const strategyOptions: Array<{ key: StrategyType; label: string; color: string; desc: string; icon: any }> = [
    { key: 'shortest', label: '最短路径', color: '#1677ff', desc: '优先距离最短', icon: <ThunderboltOutlined /> },
    { key: 'safest', label: '最安全路径', color: '#52c41a', desc: '最大限度绕行，避开所有风险', icon: <SafetyCertificateOutlined /> },
    { key: 'economic', label: '经济路径', color: '#722ed1', desc: '高速优先，省时省油费', icon: <BankOutlined /> },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <RouteOutlined style={{ color: '#1677ff', fontSize: 18 }} />
            <Text strong style={{ fontSize: 16 }}>危险品运输路径规划</Text>
            <Tag color="blue">A* 算法 + 路权权重</Tag>
          </Space>
        }
      >
        <Alert
          type="info"
          showIcon
          message="系统自动避开：人口密集区（学校/医院/商圈）、隧道、桥梁、水源保护区、限高限重路段"
          style={{ marginBottom: 24, borderRadius: 8 }}
        />
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            vehicle_type: 'tanker',
            hazard_class: '3',
            origin_address: '北京市朝阳区化工路',
            origin_lat: 39.8726,
            origin_lng: 116.4888,
            dest_address: '天津市滨海新区化工园区',
            dest_lat: 39.0275,
            dest_lng: 117.6394,
          }}
        >
          <Row gutter={24}>
            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #e6f4ff' }} title={<Space><EnvironmentOutlined style={{ color: '#52c41a' }} /> 起点</Card>}>
                <Form.Item label="起点地址" name="origin_address" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input placeholder="输入起点地址或点击地图选点" />
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="纬度" name="origin_lat" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="经度" name="origin_lng" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                </Row>
                <Button
                  type={pickingMode === 'origin' ? 'primary' : 'dashed'}
                  size="small"
                  block
                  style={{ marginTop: 12 }}
                  icon={<ZoomInOutlined />}
                  onClick={() => setPickingMode(pickingMode === 'origin' ? null : 'origin')}
                >
                  {pickingMode === 'origin' ? '正在地图上选点...' : '在地图上选择起点'}
                </Button>
              </Card>
            </Col>

            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #ffccc7' }} title={<Space><EnvironmentOutlined style={{ color: '#ff4d4f' }} /> 终点</Card>}>
                <Form.Item label="终点地址" name="dest_address" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input placeholder="输入终点地址或点击地图选点" />
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="纬度" name="dest_lat" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="经度" name="dest_lng" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                </Row>
                <Button
                  type={pickingMode === 'dest' ? 'primary' : 'dashed'}
                  size="small"
                  block
                  style={{ marginTop: 12 }}
                  icon={<ZoomInOutlined />}
                  onClick={() => setPickingMode(pickingMode === 'dest' ? null : 'dest')}
                >
                  {pickingMode === 'dest' ? '正在地图上选点...' : '在地图上选择终点'}
                </Button>
              </Card>
            </Col>

            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #ffe7ba' }} title={<Space><CarOutlined style={{ color: '#fa8c16' }} /> 车辆 & 危险品</Card>}>
                <Form.Item label="车辆类型" name="vehicle_type" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Select>
                    <Option value="tanker">🏺 罐车</Option>
                    <Option value="van">📦 厢式货车</Option>
                    <Option value="flatbed">🪵 平板车</Option>
                    <Option value="other">🚛 其他</Option>
                  </Select>
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="车高(m)" name="vehicle_height" initialValue={3.9} style={{ marginBottom: 0 }}>
                      <Input type="number" step={0.1} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="总重(吨)" name="vehicle_weight" initialValue={32.5} style={{ marginBottom: 0 }}>
                      <Input type="number" step={0.5} />
                    </Form.Item>
                  </Col>
                </Row>
                <Form.Item label="危险品类别" name="hazard_class" rules={[{ required: true }]} style={{ marginTop: 8, marginBottom: 0 }}>
                  <Select>
                    <Option value="1">1类 - 爆炸品</Option>
                    <Option value="2">2类 - 压缩气体</Option>
                    <Option value="3">3类 - 易燃液体 ⭐</Option>
                    <Option value="4">4类 - 易燃固体</Option>
                    <Option value="5">5类 - 氧化剂</Option>
                    <Option value="6">6类 - 毒害品</Option>
                    <Option value="8">8类 - 腐蚀品</Option>
                    <Option value="9">9类 - 杂项</Option>
                  </Select>
                </Form.Item>
              </Card>
            </Col>
          </Row>

          <Divider style={{ margin: '20px 0 16px' }} />

          <Row gutter={16} style={{ marginBottom: 16 }}>
            {strategyOptions.map(s => (
              <Col xs={24} md={8} key={s.key}>
                <div
                  onClick={() => setActiveStrategy(s.key)}
                  style={{
                    padding: 16,
                    borderRadius: 10,
                    border: `2px solid ${activeStrategy === s.key ? s.color : '#f0f0f0'}`,
                    background: activeStrategy === s.key ? `${s.color}08` : '#fafafa',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                  }}
                >
                  <Space style={{ fontSize: 18, color: s.color }}>
                    {s.icon}
                    <Text strong style={{ fontSize: 16, color: '#1f1f1f' }}>{s.label}</Text>
                  </Space>
                  <div style={{ marginTop: 4, fontSize: 12, color: '#8c8c8c' }}>{s.desc}</div>
                </div>
              </Col>
            ))}
          </Row>

          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading} size="large" icon={<RouteOutlined />}>
                {loading ? '规划中...' : '开始三策略规划'}
              </Button>
              <Button
                size="large"
                icon={<ReloadOutlined />}
                onClick={() => {
                  form.resetFields()
                  setRoutes({ shortest: null, safest: null, economic: null })
                  setOriginPicked(null)
                  setDestPicked(null)
                }}
              >
                重置
              </Button>
              {pickingMode && (
                <Tag color="red">
                  📍 点击地图选择{pickingMode === 'origin' ? '起点' : '终点'}位置
                </Tag>
              )}
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            bodyStyle={{ padding: 0 }}
            title={
              <Space>
                <ExperimentOutlined style={{ color: '#1677ff' }} />
                <Text strong style={{ fontSize: 15 }}>路径规划地图</Text>
                <Tag color="geekblue">高德高精地图 · 货车模式</Tag>
                <Tag color="orange">实时路况：{trafficEvents.length} 起</Tag>
              </Space>
            }
          >
            <AMap
              style={{ height: 560, borderRadius: '0 0 12px 12px' }}
              markers={routeMarkers}
              polylines={routePolylines}
              polygons={restrictedPolygons}
              center={[116.8, 39.5]}
              zoom={9}
              onMapClick={handleMapClick}
              showTraffic
            />
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            bordered={false}
            style={{ borderRadius: 12, height: 560, display: 'flex', flexDirection: 'column' }}
            bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}
          >
            <Tabs
              activeKey={activeTab}
              onChange={k => setActiveTab(k as any)}
              size="large"
              style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
              items={[
                {
                  key: 'result',
                  label: <Space><RouteOutlined /> 策略结果</Space>,
                  children: (
                    <Tabs
                      activeKey={activeStrategy}
                      onChange={k => setActiveStrategy(k as StrategyType)}
                      size="small"
                      style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
                      items={strategyOptions.map(s => ({
                        key: s.key,
                        label: (
                          <span style={{ color: activeStrategy === s.key ? s.color : undefined }}>
                            {s.icon} {s.label}
                          </span>
                        ),
                        children: renderStrategyResult(s),
                      }))}
                    />
                  ),
                },
                {
                  key: 'traffic',
                  label: <Space><Badge count={trafficEvents.length} size="small"><AlertOutlined /></Badge> 实时路况</Space>,
                  children: renderTrafficPanel(),
                },
                {
                  key: 'history',
                  label: <Space><HistoryOutlined /> 重规划历史</Space>,
                  children: renderReplanHistory(),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><HistoryOutlined /> 重规划历史记录（完整列表）</Space>}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="触发类型"
            allowClear
            value={replanFilterTrigger || undefined}
            onChange={setReplanFilterTrigger}
            style={{ width: 160 }}
          >
            <Option value="traffic">路况事件</Option>
            <Option value="deviation">车辆偏航</Option>
            <Option value="restricted">禁行区变更</Option>
            <Option value="manual">手动触发</Option>
          </Select>
          <Select
            placeholder="处理状态"
            allowClear
            value={replanFilterStatus || undefined}
            onChange={setReplanFilterStatus}
            style={{ width: 160 }}
          >
            <Option value="pending">待确认</Option>
            <Option value="confirmed">已确认</Option>
            <Option value="rejected">已拒绝</Option>
            <Option value="auto_applied">系统自动应用</Option>
          </Select>
          <RangePicker value={replanDateRange} onChange={setReplanDateRange} />
          <Button icon={<ReloadOutlined />} onClick={fetchReplanHistory}>刷新</Button>
        </Space>
        <Table
          dataSource={replanList}
          loading={replanListLoading}
          rowKey="id"
          pagination={{
            current: replanPage,
            pageSize: replanPageSize,
            total: replanTotal,
            onChange: setReplanPage,
            showTotal: t => `共 ${t} 条重规划记录`,
          }}
          columns={[
            {
              title: '编号',
              dataIndex: 'replan_no',
              width: 180,
              render: (v: string) => <Text code>{v}</Text>,
            },
            {
              title: '运单/车牌',
              width: 200,
              render: (_: any, r: any) => (
                <div>
                  <div>{r.waybill_no}</div>
                  <div style={{ fontSize: 12, color: '#8c8c8c' }}>{r.vehicle_plate} · {r.driver_name}</div>
                </div>
              ),
            },
            {
              title: '触发类型',
              dataIndex: 'trigger_type',
              width: 120,
              render: (v: string) => {
                const meta = triggerTypeLabel[v] || { label: v, color: 'default', icon: <InfoCircleOutlined /> }
                return <Tag color={meta.color} icon={meta.icon}>{meta.label}</Tag>
              },
            },
            {
              title: '触发原因',
              dataIndex: 'trigger_reason',
              ellipsis: true,
              width: 260,
              render: (v: string) => <Tooltip title={v}>{v}</Tooltip>,
            },
            {
              title: '对比(新/原)',
              width: 200,
              render: (_: any, r: any) => (
                <Space direction="vertical" size={2}>
                  <Text type={r.distance_delta >= 0 ? 'danger' : 'success'}>
                    {r.distance_delta >= 0 ? '+' : ''}{(r.distance_delta || 0).toFixed(1)} km
                  </Text>
                  <Text type={r.duration_delta >= 0 ? 'danger' : 'success'}>
                    {r.duration_delta >= 0 ? '+' : ''}{r.duration_delta || 0} 分钟
                  </Text>
                </Space>
              ),
            },
            {
              title: '状态',
              dataIndex: 'status',
              width: 120,
              render: (v: string) => {
                const meta = statusLabel[v] || { label: v, color: 'default' }
                return <Tag color={meta.color}>{meta.label}</Tag>
              },
            },
            {
              title: '触发时间',
              dataIndex: 'created_at',
              width: 170,
              render: (v: string) => new Date(v).toLocaleString(),
            },
            {
              title: '操作',
              width: 90,
              render: (_: any, r: any) => (
                <Button
                  type="link"
                  icon={<EyeOutlined />}
                  size="small"
                  onClick={() => handleViewReplanDetail(r.id)}
                >
                  详情
                </Button>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={<Space><WarningOutlined style={{ color: '#faad14' }} /> 车辆偏航提醒</Space>}
        open={deviationModal}
        onCancel={() => setDeviationModal(false)}
        footer={
          <Space>
            <Button onClick={() => setDeviationModal(false)}>忽略</Button>
            <Button type="primary" loading={replanLoading} onClick={handleReplan}>
              立即重规划
            </Button>
          </Space>
        }
        width={520}
      >
        {currentDeviation && (
          <div>
            <Alert
              type="warning"
              showIcon
              message={`车辆 ${currentDeviation.vehicle_plate} 已偏离规划路线`}
              description={`偏离距离 ${currentDeviation.deviation_distance.toFixed(1)} 米，请确认是否需要重新规划路线`}
              style={{ marginBottom: 16, borderRadius: 8 }}
            />
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="运单ID">{currentDeviation.waybill_id}</Descriptions.Item>
              <Descriptions.Item label="当前位置">
                {currentDeviation.current_lat.toFixed(6)}, {currentDeviation.current_lng.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="原路线ID">{currentDeviation.planned_route_id}</Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      <Modal
        title={
          <Space>
            <BellOutlined style={{ color: '#1677ff' }} />
            路线重规划建议
            {currentSuggestion?.status === 'pending' && <Tag color="orange">待司机确认</Tag>}
          </Space>
        }
        open={replanSuggestModal}
        onCancel={() => setReplanSuggestModal(false)}
        footer={
          <Space>
            <Button
              danger
              loading={confirmLoading}
              onClick={() => handleConfirmReplan('reject')}
            >
              拒绝建议
            </Button>
            <Button
              type="primary"
              loading={confirmLoading}
              onClick={() => handleConfirmReplan('confirm')}
            >
              <CheckCircleOutlined /> 确认并应用新路线
            </Button>
          </Space>
        }
        width={620}
        destroyOnClose
      >
        {currentSuggestion && (
          <Spin spinning={confirmLoading}>
            <Alert
              type="info"
              showIcon
              message={`${currentSuggestion.vehicle_plate || '车辆'} 前方路况变化`}
              description={currentSuggestion.trigger_reason}
              style={{ marginBottom: 16, borderRadius: 8 }}
            />

            {currentSuggestion.traffic_event && (
              <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} type="inner" title="关联路况事件">
                <Space direction="vertical" size={2}>
                  <Space>
                    <Tag color="red">{currentSuggestion.traffic_event.type}</Tag>
                    <Tag color="orange">等级 {currentSuggestion.traffic_event.level}</Tag>
                    <Text strong>{currentSuggestion.traffic_event.title}</Text>
                  </Space>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    🛣️ {currentSuggestion.traffic_event.road_name || '未知路段'}
                  </Text>
                </Space>
              </Card>
            )}

            <Descriptions column={2} size="small" bordered style={{ marginBottom: 16 }}>
              <Descriptions.Item label="运单号">{currentSuggestion.waybill_no}</Descriptions.Item>
              <Descriptions.Item label="驾驶员">{currentSuggestion.driver_name}</Descriptions.Item>
              <Descriptions.Item label="原剩余里程">
                {(currentSuggestion.original_distance_remaining || 0).toFixed(1)} km
              </Descriptions.Item>
              <Descriptions.Item label="原剩余时长">
                {currentSuggestion.original_duration_remaining || 0} 分钟
              </Descriptions.Item>
              <Descriptions.Item label="新剩余里程">
                <Text type={currentSuggestion.distance_delta >= 0 ? 'danger' : 'success'}>
                  {(currentSuggestion.new_distance_remaining || 0).toFixed(1)} km
                  ({currentSuggestion.distance_delta >= 0 ? '+' : ''}
                  {(currentSuggestion.distance_delta || 0).toFixed(1)})
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="新剩余时长">
                <Text type={currentSuggestion.duration_delta >= 0 ? 'danger' : 'success'}>
                  {currentSuggestion.new_duration_remaining || 0} 分钟
                  ({currentSuggestion.duration_delta >= 0 ? '+' : ''}
                  {currentSuggestion.duration_delta || 0})
                </Text>
              </Descriptions.Item>
            </Descriptions>

            <Divider orientation="left" style={{ margin: '8px 0 12px' }}>候选路线</Divider>
            <List
              size="small"
              dataSource={currentSuggestion.candidates || []}
              renderItem={(c: any, idx: number) => (
                <List.Item
                  style={c.is_recommended ? {
                    background: '#f6ffed',
                    borderRadius: 8,
                    padding: '10px 12px',
                    marginBottom: 8,
                    border: '1px solid #b7eb8f',
                  } : { padding: '10px 12px', borderRadius: 8, marginBottom: 8, border: '1px solid #f0f0f0' }}
                >
                  <Space direction="vertical" size={2} style={{ width: '100%' }}>
                    <Space>
                      <Tag color={idx === 0 ? 'blue' : 'default'}>方案 {idx + 1}</Tag>
                      {c.is_recommended && <Tag color="green">推荐</Tag>}
                      <Text strong style={{ textTransform: 'uppercase' }}>{c.strategy}</Text>
                    </Space>
                    <Space size={16}>
                      <Text>📏 {(c.total_distance || 0).toFixed(1)} km</Text>
                      <Text>⏱️ {c.estimated_duration || 0} 分钟</Text>
                      <Text>🛡️ 安全 {c.safety_score || 0}</Text>
                      {c.toll_fee > 0 && <Text>💰 过路费 {c.toll_fee} 元</Text>}
                    </Space>
                  </Space>
                </List.Item>
              )}
            />
            {(!currentSuggestion.candidates || currentSuggestion.candidates.length === 0) && (
              <Empty description="无候选方案" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Spin>
        )}
      </Modal>

      <Drawer
        title={
          <Space>
            <AlertOutlined /> 路况事件详情
            {eventDetail?.status === 'active' && <Badge status="processing" text="活跃中" />}
          </Space>
        }
        placement="right"
        width={480}
        open={!!eventDetail}
        onClose={() => setEventDetail(null)}
      >
        {eventDetail && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Tag color={eventDetail.event_level >= 3 ? 'red' : 'orange'}>
              {eventDetail.event_level >= 3 ? '严重事件' : '中等事件'}
            </Tag>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="事件编号">{eventDetail.event_no}</Descriptions.Item>
              <Descriptions.Item label="事件类型">{eventDetail.event_type}</Descriptions.Item>
              <Descriptions.Item label="事件标题">{eventDetail.title}</Descriptions.Item>
              <Descriptions.Item label="路段名称">{eventDetail.road_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="事件描述">{eventDetail.description || '-'}</Descriptions.Item>
              <Descriptions.Item label="中心坐标">
                {eventDetail.center_lat}, {eventDetail.center_lng}
              </Descriptions.Item>
              <Descriptions.Item label="影响长度">{(eventDetail.affected_length_km || 0).toFixed(2)} km</Descriptions.Item>
              <Descriptions.Item label="拥堵等级">
                {['', '畅通', '缓行', '拥堵', '严重拥堵'][eventDetail.congestion_level || 0]}
              </Descriptions.Item>
              <Descriptions.Item label="平均车速">{eventDetail.avg_speed_kmh || 0} km/h</Descriptions.Item>
              <Descriptions.Item label="预计延误">{eventDetail.duration_minutes || 0} 分钟</Descriptions.Item>
              <Descriptions.Item label="开始时间">{new Date(eventDetail.started_at).toLocaleString()}</Descriptions.Item>
              <Descriptions.Item label="预计结束">
                {eventDetail.expected_end_at ? new Date(eventDetail.expected_end_at).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="数据来源">{eventDetail.source || 'system'}</Descriptions.Item>
            </Descriptions>
            <Space>
              <Button type="primary" danger onClick={() => handleResolveEvent(eventDetail.id)}>
                <CheckCircleOutlined /> 标记已解决
              </Button>
            </Space>
          </Space>
        )}
      </Drawer>

      <Drawer
        title={
          <Space>
            <HistoryOutlined /> 重规划详情
            {replanDetail && <Tag color={statusLabel[replanDetail.status]?.color || 'default'}>
              {statusLabel[replanDetail.status]?.label || replanDetail.status}
            </Tag>}
          </Space>
        }
        placement="right"
        width={620}
        open={!!replanDetail}
        onClose={() => setReplanDetail(null)}
      >
        {replanDetail && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Tag color={triggerTypeLabel[replanDetail.trigger_type]?.color || 'default'}>
              {triggerTypeLabel[replanDetail.trigger_type]?.icon}
              {' '}{triggerTypeLabel[replanDetail.trigger_type]?.label || replanDetail.trigger_type}
            </Tag>

            <Descriptions column={2} size="small" bordered title="基础信息">
              <Descriptions.Item label="重规划编号">{replanDetail.replan_no}</Descriptions.Item>
              <Descriptions.Item label="触发时间">
                {new Date(replanDetail.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="运单号">{replanDetail.waybill_no}</Descriptions.Item>
              <Descriptions.Item label="车牌号">{replanDetail.vehicle_plate}</Descriptions.Item>
              <Descriptions.Item label="驾驶员">{replanDetail.driver_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="调度员">{replanDetail.operator_name || '系统自动'}</Descriptions.Item>
              <Descriptions.Item label="触发原因" span={2}>{replanDetail.trigger_reason}</Descriptions.Item>
            </Descriptions>

            <Descriptions column={2} size="small" bordered title="路线对比">
              <Descriptions.Item label="原剩余里程">
                {(replanDetail.original_distance_remaining || 0).toFixed(2)} km
              </Descriptions.Item>
              <Descriptions.Item label="新剩余里程">
                <Text type={replanDetail.distance_delta >= 0 ? 'danger' : 'success'}>
                  {(replanDetail.new_distance_remaining || 0).toFixed(2)} km
                  ({replanDetail.distance_delta >= 0 ? '+' : ''}{(replanDetail.distance_delta || 0).toFixed(2)})
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="原剩余时长">
                {replanDetail.original_duration_remaining || 0} 分钟
              </Descriptions.Item>
              <Descriptions.Item label="新剩余时长">
                <Text type={replanDetail.duration_delta >= 0 ? 'danger' : 'success'}>
                  {replanDetail.new_duration_remaining || 0} 分钟
                  ({replanDetail.duration_delta >= 0 ? '+' : ''}{replanDetail.duration_delta || 0})
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="原路线ID">{replanDetail.original_route_plan_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="新路线ID">{replanDetail.new_route_plan_id || '-'}</Descriptions.Item>
            </Descriptions>

            <Card size="small" title="处理时间线" style={{ borderRadius: 8 }}>
              <Timeline
                items={[
                  {
                    color: 'blue',
                    children: (
                      <span>
                        触发重规划：{replanDetail.trigger_reason}
                        <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                          {new Date(replanDetail.created_at).toLocaleString()}
                        </div>
                      </span>
                    ),
                  },
                  replanDetail.notified_at && {
                    color: 'cyan',
                    children: (
                      <span>
                        通知已推送至司机端
                        <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                          {new Date(replanDetail.notified_at).toLocaleString()}
                        </div>
                      </span>
                    ),
                  },
                  replanDetail.driver_confirm_at && {
                    color: replanDetail.status === 'confirmed' || replanDetail.status === 'auto_applied' ? 'green' : 'red',
                    children: (
                      <span>
                        司机已{replanDetail.status === 'confirmed' || replanDetail.status === 'auto_applied' ? '确认' : '拒绝'}
                        {replanDetail.confirm_note && <div>备注：{replanDetail.confirm_note}</div>}
                        <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                          {new Date(replanDetail.driver_confirm_at).toLocaleString()}
                        </div>
                      </span>
                    ),
                  },
                  replanDetail.applied_at && {
                    color: 'green',
                    dot: <CheckCircleOutlined />,
                    children: (
                      <span>
                        新路线已应用
                        <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                          {new Date(replanDetail.applied_at).toLocaleString()}
                        </div>
                      </span>
                    ),
                  },
                ].filter(Boolean) as any}
              />
            </Card>

            {replanDetail.candidate_routes?.length > 0 && (
              <Card size="small" title="候选路线" style={{ borderRadius: 8 }}>
                <List
                  size="small"
                  dataSource={replanDetail.candidate_routes}
                  renderItem={(c: any) => (
                    <List.Item style={{
                      padding: '10px 12px',
                      borderRadius: 8,
                      marginBottom: 8,
                      background: c.is_recommended ? '#f6ffed' : '#fafafa',
                      border: `1px solid ${c.is_recommended ? '#b7eb8f' : '#f0f0f0'}`,
                    }}>
                      <Space direction="vertical" size={2} style={{ width: '100%' }}>
                        <Space>
                          <Tag color="blue">{c.strategy}</Tag>
                          {c.is_recommended && <Tag color="green">✨ 应用路线</Tag>}
                        </Space>
                        <Space size={16}>
                          <span>📏 {(c.total_distance || 0).toFixed(1)} km</span>
                          <span>⏱️ {c.estimated_duration || 0} 分钟</span>
                          <span>🛡️ {c.safety_score || 0} 分</span>
                        </Space>
                      </Space>
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </Space>
        )}
      </Drawer>
    </div>
  )
}

export default RoutePlan
