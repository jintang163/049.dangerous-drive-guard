import React, { useEffect, useState, useCallback } from 'react'
import {
  Row,
  Col,
  Card,
  List,
  Tag,
  Statistic,
  Drawer,
  Descriptions,
  Badge,
  Button,
  Space,
  Tooltip,
  Typography,
  Empty,
  Spin,
  Progress,
  Input,
  Select,
  Divider,
  message,
  Modal,
  Form,
  Image,
} from 'antd'
import {
  TruckOutlined,
  EnvironmentOutlined,
  AlertOutlined,
  PhoneOutlined,
  CoffeeOutlined,
  SafetyCertificateOutlined,
  SearchOutlined,
  UserOutlined,
  AudioOutlined,
  ThunderboltOutlined,
  ReloadOutlined,
  DashboardOutlined,
  VideoCameraOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import AMap from '@/components/AMap'
import { useAppStore, VehicleStatus, AlarmItem, StatData, ServiceArea } from '@/store/app'
import { vehicleApi, monitorApi, routeApi, fatigueApi } from '@/services/api'
import type { PageParams } from '@/services/api'
import { formatDateTime, formatDistance, formatDuration } from '@/utils/auth'
import WebSocketManager from '@/services/ws'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select

const Monitor: React.FC = () => {
  const {
    vehicles,
    alarms,
    updateVehicles,
    upsertVehicle,
    addAlarm,
    updateAlarm,
    selectedVehicle,
    setSelectedVehicle,
    setUnreadAlarmCount,
    fetchVehicles: fetchStoreVehicles,
    loading: storeLoading,
  } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [mediaLoading, setMediaLoading] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [newAlarms, setNewAlarms] = useState<AlarmItem[]>([])
  const [intercomModal, setIntercomModal] = useState(false)
  const [dispatchModal, setDispatchModal] = useState(false)
  const [vehicleDetail, setVehicleDetail] = useState<any>(null)
  const [historyRecords, setHistoryRecords] = useState<any[]>([])
  const [restrictedAreas, setRestrictedAreas] = useState<any[]>([])
  const [form] = Form.useForm()
  const [serviceAreas, setServiceAreas] = useState<ServiceArea[]>([])
  const [statistics, setStatistics] = useState<StatData | null>(null)
  const [detailVideoURL, setDetailVideoURL] = useState<string>('')
  const [detailSnapshotURL, setDetailSnapshotURL] = useState<string>('')
  const [recommendedAreas, setRecommendedAreas] = useState<any[]>([])

  const fetchVehicles = useCallback(async () => {
    try {
      const data = await vehicleApi.listRealtimeStatus()
      updateVehicles(data || [])
    } finally {
      setLoading(false)
    }
  }, [updateVehicles])

  const fetchStatistics = useCallback(async () => {
    try {
      const data = await monitorApi.getStatistics()
      setStatistics(data)
    } catch (e) {}
  }, [])

  const fetchRestrictedAreas = useCallback(async () => {
    try {
      const data = await routeApi.listRestrictedAreas()
      setRestrictedAreas(data || [])
    } catch (e) {}
  }, [])

  const fetchVehicleMedia = useCallback(async (alarmId: number) => {
    setMediaLoading(true)
    try {
      const [videoRes, snapshotRes] = await Promise.all([
        fatigueApi.getVideoURL(alarmId).catch(() => ({ url: '' })),
        fatigueApi.getSnapshotURL(alarmId).catch(() => ({ url: '' })),
      ])
      setDetailVideoURL(videoRes?.url || '')
      setDetailSnapshotURL(snapshotRes?.url || '')
    } finally {
      setMediaLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchVehicles()
    fetchStatistics()
    fetchRestrictedAreas()
    const timer = setInterval(fetchVehicles, 30000)
    const statsTimer = setInterval(fetchStatistics, 60000)

    const unsub1 = WebSocketManager.getInstance().on('vehicle_status', (vehicle: VehicleStatus) => {
      upsertVehicle(vehicle)
    })

    const unsub2 = WebSocketManager.getInstance().on('new_alarm', async (alarm: AlarmItem) => {
      addAlarm(alarm)
      setNewAlarms(prev => [alarm, ...prev].slice(0, 10))
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
              {alarm.alarm_type} · 疲劳指数 {Math.round(alarm.fatigue_score)}
            </div>
          </div>
        ),
        duration: 8,
      })
    })

    const unsub3 = WebSocketManager.getInstance().on('alarm_updated', (alarm: Partial<AlarmItem>) => {
      if (alarm.id) {
        updateAlarm(alarm.id, alarm)
      }
    })

    return () => {
      clearInterval(timer)
      clearInterval(statsTimer)
      unsub1()
      unsub2()
      unsub3()
    }
  }, [fetchVehicles, fetchStatistics, fetchRestrictedAreas, upsertVehicle, addAlarm, updateAlarm])

  useEffect(() => {
    setUnreadAlarmCount(0)
  }, [alarms, setUnreadAlarmCount])

  const filteredVehicles = vehicles.filter(v => {
    if (searchText) {
      const kw = searchText.toLowerCase()
      if (
        !v.plate_number?.toLowerCase().includes(kw) &&
        !v.driver_name?.toLowerCase().includes(kw) &&
        !String(v.vehicle_id).includes(kw)
      ) {
        return false
      }
    }
    if (statusFilter) {
      if (statusFilter === 'fatigue' && v.fatigue_level !== 'fatigue') return false
      if (statusFilter === 'warning' && v.fatigue_level !== 'warning') return false
      if (statusFilter === 'normal' && v.fatigue_level !== 'normal') return false
      if (statusFilter === 'offline' && v.status !== 'offline') return false
      if (statusFilter === 'running' && v.status !== 'running') return false
    }
    return true
  })

  const mapMarkers = filteredVehicles.map(v => ({
    position: [v.longitude || 116.4074 + (Math.random() - 0.5) * 0.2, v.latitude || 39.9042 + (Math.random() - 0.5) * 0.2],
    title: v.plate_number || `车辆${v.vehicle_id}`,
    color: v.marker_color || (
      v.fatigue_level === 'fatigue' ? '#ff4d4f' :
        v.fatigue_level === 'warning' ? '#faad14' :
          v.status === 'offline' ? '#8c8c8c' : '#52c41a'
    ),
    info: v,
  }))

  const mapPolygons = restrictedAreas.slice(0, 20).map(area => {
    let path: [number, number][] = []
    try {
      const poly = area.boundary_polygon
      if (poly?.coordinates?.[0]) {
        path = poly.coordinates[0].map((p: number[]) => [p[0], p[1]])
      } else if (area.center_latitude && area.center_longitude) {
        const r = (area.radius || 500) / 111000
        path = Array.from({ length: 32 }, (_, i) => {
          const a = (i / 32) * Math.PI * 2
          return [area.center_longitude + r * Math.cos(a), area.center_latitude + r * Math.sin(a)]
        })
      }
    } catch (e) { }
    return {
      path,
      fillColor: area.level === 2 ? '#ff4d4f' : '#fa8c16',
      strokeColor: area.level === 2 ? '#ff7875' : '#ffa940',
      fillOpacity: 0.12,
    }
  })

  const handleVehicleClick = async (v: VehicleStatus) => {
    setSelectedVehicle(v)
    setDetailLoading(true)
    setDetailVideoURL('')
    setDetailSnapshotURL('')
    try {
      const detail = await vehicleApi.getRealtimeStatus(v.vehicle_id)
      setVehicleDetail(detail)
      const recs = await fatigueApi.history({ vehicle_id: v.vehicle_id, page: 1, page_size: 20 })
      setHistoryRecords(recs?.list || [])

      const latestAlarm = alarms.find(a => a.vehicle_id === v.vehicle_id)
      if (latestAlarm) {
        await fetchVehicleMedia(latestAlarm.id)
      }
    } finally {
      setDetailLoading(false)
    }
  }

  const handleSendIntercom = async (values: any) => {
    if (!selectedVehicle) return
    try {
      const userInfo = localStorage.getItem('ddg_user_info')
      const operatorId = userInfo ? JSON.parse(userInfo).id : 1
      await monitorApi.voiceIntercom(selectedVehicle.vehicle_id, {
        action: 'start',
        operator_id: operatorId,
        message: values.message,
        priority: values.priority || 1,
      })
      message.success('语音对讲指令已下发')
      setIntercomModal(false)
      form.resetFields()
    } catch (e) { }
  }

  const handleDispatchService = async () => {
    try {
      const areas = await routeApi.recommendServiceArea({
        waybill_id: selectedVehicle?.waybill_id || 0,
        fatigue_score: selectedVehicle?.fatigue_score,
      })
      setRecommendedAreas(areas || [])
      setDispatchModal(true)
    } catch (e) { }
  }

  const handleConfirmDispatch = async (areaId: number, restDuration: number) => {
    if (!selectedVehicle) return
    try {
      await monitorApi.dispatchServiceArea(selectedVehicle.vehicle_id, {
        service_area_id: areaId,
        reason: `调度停靠 - ${selectedVehicle?.fatigue_level === 'fatigue' ? '严重疲劳' : '预警状态'}`,
        rest_duration: restDuration,
      })
      message.success('停靠指令已下发')
      setDispatchModal(false)
    } catch (e) { }
  }

  const getStatusBadge = (v: VehicleStatus) => {
    if (v.status === 'offline') return <Tag color="default">离线</Tag>
    if (v.fatigue_level === 'fatigue') return <Tag color="red">严重疲劳</Tag>
    if (v.fatigue_level === 'warning') return <Tag color="orange">疲劳预警</Tag>
    if (v.status === 'running') return <Tag color="green">行驶中</Tag>
    return <Tag color="blue">待命</Tag>
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="在线车辆"
              value={filteredVehicles.filter(v => v.status !== 'offline').length}
              suffix={`/ ${vehicles.length}`}
              prefix={<TruckOutlined style={{ color: '#1677ff' }} />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="行驶中"
              value={vehicles.filter(v => v.status === 'running').length}
              prefix={<DashboardOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="疲劳预警"
              value={vehicles.filter(v => v.fatigue_level === 'warning').length}
              prefix={<AlertOutlined style={{ color: '#faad14' }} />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="严重疲劳"
              value={vehicles.filter(v => v.fatigue_level === 'fatigue').length}
              prefix={<ThunderboltOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={17}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            bodyStyle={{ padding: 0 }}
            title={
              <Space>
                <EnvironmentOutlined style={{ color: '#1677ff' }} />
                <Text strong style={{ fontSize: 15 }}>实时监控地图</Text>
                {newAlarms.length > 0 && (
                  <Badge count={newAlarms.length} offset={[4, -2]}>
                    <Tag color="red">新报警</Tag>
                  </Badge>
                )}
              </Space>
            }
            extra={
              <Space>
                <Tooltip title="刷新车辆位置">
                  <Button type="text" icon={<ReloadOutlined />} onClick={fetchVehicles} />
                </Tooltip>
              </Space>
            }
          >
            <div style={{ position: 'relative' }}>
              <AMap
                style={{ height: 620, borderRadius: '0 0 12px 12px' }}
                markers={mapMarkers}
                polygons={mapPolygons}
                onMarkerClick={(m) => handleVehicleClick(m.info as VehicleStatus)}
                center={mapMarkers[0]?.position || [116.4074, 39.9042]}
                zoom={10}
                showTraffic
              />
              <div
                style={{
                  position: 'absolute',
                  top: 16,
                  left: 16,
                  background: 'rgba(255,255,255,0.96)',
                  padding: 12,
                  borderRadius: 8,
                  boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
                  minWidth: 200,
                }}
              >
                <Text type="secondary" style={{ fontSize: 12 }}>图例</Text>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 8 }}>
                  {[
                    { color: '#52c41a', label: '正常' },
                    { color: '#faad14', label: '预警' },
                    { color: '#ff4d4f', label: '疲劳' },
                    { color: '#8c8c8c', label: '离线' },
                  ].map(item => (
                    <div key={item.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span style={{ width: 10, height: 10, borderRadius: 2, background: item.color, display: 'inline-block' }} />
                      <Text style={{ fontSize: 12 }}>{item.label}</Text>
                    </div>
                  ))}
                  <Divider style={{ margin: '8px 0' }} />
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{
                      width: 18,
                      height: 18,
                      border: '2px dashed #ff4d4f',
                      background: 'rgba(255,77,79,0.12)',
                      borderRadius: 4,
                      display: 'inline-block',
                    }} />
                    <Text style={{ fontSize: 12 }}>禁行区</Text>
                  </div>
                </div>
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={7}>
          <Card
            bordered={false}
            style={{ borderRadius: 12, height: 620, display: 'flex', flexDirection: 'column' }}
            bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}
            title={
              <Space>
                <TruckOutlined style={{ color: '#1677ff' }} />
                <Text strong style={{ fontSize: 15 }}>车辆列表</Text>
              </Space>
            }
            extra={
              <Text type="secondary" style={{ fontSize: 12 }}>
                共 {filteredVehicles.length} 辆
              </Text>
            }
          >
            <div style={{ padding: '0 16px 12px', display: 'flex', gap: 8, flexDirection: 'column' }}>
              <Input
                allowClear
                placeholder="搜索车牌/驾驶员"
                prefix={<SearchOutlined />}
                value={searchText}
                onChange={e => setSearchText(e.target.value)}
              />
              <Select
                allowClear
                placeholder="状态筛选"
                value={statusFilter}
                onChange={setStatusFilter}
                style={{ width: '100%' }}
              >
                <Option value="running">行驶中</Option>
                <Option value="normal">状态正常</Option>
                <Option value="warning">疲劳预警</Option>
                <Option value="fatigue">严重疲劳</Option>
                <Option value="offline">离线</Option>
              </Select>
            </div>
            <div style={{ flex: 1, overflow: 'auto', padding: '0 4px 16px 16px' }}>
              {loading ? (
                <div style={{ padding: 40, textAlign: 'center' }}>
                  <Spin />
                </div>
              ) : filteredVehicles.length ? (
                <List
                  size="small"
                  dataSource={filteredVehicles}
                  renderItem={(v) => (
                    <List.Item
                      onClick={() => handleVehicleClick(v)}
                      style={{
                        padding: '12px 12px',
                        marginBottom: 8,
                        borderRadius: 8,
                        cursor: 'pointer',
                        background: selectedVehicle?.vehicle_id === v.vehicle_id
                          ? '#e6f4ff'
                          : '#fafafa',
                        border: selectedVehicle?.vehicle_id === v.vehicle_id
                          ? '1px solid #1677ff'
                          : '1px solid transparent',
                        transition: 'all 0.2s',
                      }}
                    >
                      <div style={{ width: '100%' }}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                          <Space>
                            <Text strong>{v.plate_number || `车辆${v.vehicle_id}`}</Text>
                            {getStatusBadge(v)}
                          </Space>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {Math.round(v.speed || 0)} km/h
                          </Text>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8 }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            <UserOutlined /> {v.driver_name || '未分配'}
                          </Text>
                          <Text
                            type={v.fatigue_level === 'fatigue' ? 'danger' : v.fatigue_level === 'warning' ? 'warning' : 'success'}
                            strong
                            style={{ fontSize: 12 }}
                          >
                            {Math.round(v.fatigue_score || 0)} 分
                          </Text>
                        </div>
                        <Progress
                          percent={100 - Math.round(v.fatigue_score || 0)}
                          showInfo={false}
                          size="small"
                          style={{ marginTop: 6 }}
                          strokeColor={
                            v.fatigue_level === 'fatigue' ? '#ff4d4f' :
                              v.fatigue_level === 'warning' ? '#faad14' : '#52c41a'
                          }
                        />
                      </div>
                    </List.Item>
                  )}
                />
              ) : (
                <Empty description="暂无车辆数据" style={{ padding: 40 }} />
              )}
            </div>
          </Card>
        </Col>
      </Row>

      <Drawer
        title={
          <Space>
            <TruckOutlined />
            <Text strong style={{ fontSize: 16 }}>{selectedVehicle?.plate_number || '车辆详情'}</Text>
            {selectedVehicle && getStatusBadge(selectedVehicle)}
          </Space>
        }
        open={!!selectedVehicle}
        onClose={() => { setSelectedVehicle(null); setVehicleDetail(null) }}
        width={480}
        extra={
          selectedVehicle && (
            <Space>
              <Tooltip title="语音对讲">
                <Button icon={<AudioOutlined />} type="primary" onClick={() => setIntercomModal(true)}>
                  语音对讲
                </Button>
              </Tooltip>
              <Tooltip title="调度停靠">
                <Button icon={<CoffeeOutlined />} onClick={handleDispatchService}>
                  调度停靠
                </Button>
              </Tooltip>
              <Tooltip title="电子押运">
                <Button icon={<SafetyCertificateOutlined />} type="dashed">
                  押运
                </Button>
              </Tooltip>
            </Space>
          )
        }
      >
        {detailLoading ? (
          <div style={{ padding: 40, textAlign: 'center' }}><Spin /></div>
        ) : selectedVehicle ? (
          <div>
            <Descriptions column={1} size="small" bordered style={{ marginBottom: 16 }}>
              <Descriptions.Item label="车牌">{selectedVehicle.plate_number}</Descriptions.Item>
              <Descriptions.Item label="驾驶员">{selectedVehicle.driver_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="运单号">{selectedVehicle.waybill_no || '无'}</Descriptions.Item>
              <Descriptions.Item label="当前速度">{Math.round(selectedVehicle.speed || 0)} km/h</Descriptions.Item>
              <Descriptions.Item label="方向">
                {['北', '东北', '东', '东南', '南', '西南', '西', '西北'][Math.round((selectedVehicle.direction || 0) / 45) % 8]} ({selectedVehicle.direction}°)
              </Descriptions.Item>
              <Descriptions.Item label="当前位置">
                <Paragraph copyable style={{ margin: 0, fontSize: 12 }}>
                  {selectedVehicle.current_address || `${selectedVehicle.latitude?.toFixed(6)}, ${selectedVehicle.longitude?.toFixed(6)}`}
                </Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label="剩余里程">{formatDistance(selectedVehicle.remaining_mileage * 1000)}</Descriptions.Item>
              <Descriptions.Item label="预计到达">{formatDuration(selectedVehicle.remaining_time)}</Descriptions.Item>
              <Descriptions.Item label="疲劳指数">
                <Space>
                  <Progress
                    type="circle"
                    size={80}
                    percent={Math.round(selectedVehicle.fatigue_score || 0)}
                    strokeColor={
                      (selectedVehicle.fatigue_score || 0) < 70 ? '#ff4d4f' :
                        (selectedVehicle.fatigue_score || 0) < 85 ? '#faad14' : '#52c41a'
                    }
                  />
                  <Tag
                    color={selectedVehicle.fatigue_level === 'fatigue' ? 'red' : selectedVehicle.fatigue_level === 'warning' ? 'orange' : 'green'}
                    style={{ fontSize: 14, padding: '4px 12px' }}
                  >
                    {selectedVehicle.fatigue_level === 'fatigue' ? '严重疲劳' : selectedVehicle.fatigue_level === 'warning' ? '疲劳预警' : '状态正常'}
                  </Tag>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="报警数">
                <Badge count={selectedVehicle.alert_count || 0} showZero status="error" />
              </Descriptions.Item>
              <Descriptions.Item label="最后更新">
                {formatDateTime(selectedVehicle.last_update_time || selectedVehicle.gps_time)}
              </Descriptions.Item>
            </Descriptions>

            <Card size="small" title="近期疲劳记录" style={{ borderRadius: 8, marginBottom: 16 }}>
              {historyRecords.length ? (
                <List
                  size="small"
                  dataSource={historyRecords.slice(0, 6)}
                  renderItem={(r: any) => (
                    <List.Item>
                      <Space size="large">
                        <Text style={{ fontSize: 12, minWidth: 130 }}>{formatDateTime(r.detection_time)}</Text>
                        <Tag
                          color={r.fatigue_level === 'fatigue' ? 'red' : r.fatigue_level === 'warning' ? 'orange' : 'green'}
                          style={{ margin: 0 }}
                        >
                          {Math.round(r.fatigue_score)} 分
                        </Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          车速 {Math.round(r.vehicle_speed)}km/h
                        </Text>
                        {r.is_alarm_triggered && <AlertOutlined style={{ color: '#ff4d4f' }} />
                      </Space>
                    </List.Item>
                  )}
                />
              ) : (
                <Empty description="暂无记录" />
              )}
            </Card>

            <Card size="small" title={<Space><VideoCameraOutlined /> 疲劳快照和视频</Space>} style={{ borderRadius: 8 }}>
              {mediaLoading ? (
                <div style={{ padding: 20, textAlign: 'center' }}><Spin tip="加载中..." /></div>
              ) : (
                <Row gutter={8}>
                  <Col span={12}>
                    {detailSnapshotURL ? (
                      <Image src={detailSnapshotURL} />
                    ) : (
                      <div style={{
                        background: '#f5f5f5',
                        borderRadius: 6,
                        height: 120,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#8c8c8c',
                        fontSize: 12,
                      }}>
                        <EyeOutlined /> 无快照图片
                      </div>
                    )}
                  </Col>
                  <Col span={12}>
                    {detailVideoURL ? (
                      <video controls src={detailVideoURL} style={{ width: '100%', borderRadius: 6 }} />
                    ) : (
                      <div style={{
                        background: '#f5f5f5',
                        borderRadius: 6,
                        height: 120,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#8c8c8c',
                        fontSize: 12,
                      }}>
                        <VideoCameraOutlined /> 无视频片段
                      </div>
                    )}
                  </Col>
                </Row>
              )}
            </Card>
          </div>
        ) : null}
      </Drawer>

      <Modal
        title={<Space><PhoneOutlined /> 语音对讲 - {selectedVehicle?.plate_number}</Space>}
        open={intercomModal}
        onCancel={() => setIntercomModal(false)}
        onOk={() => form.submit()}
        okText="发送"
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="priority"
            label="优先级"
            initialValue={1}
          >
            <Select>
              <Option value={1}>普通提醒</Option>
              <Option value={2}>重要提醒</Option>
              <Option value={3}>紧急通知</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="message"
            label="对讲内容"
            rules={[{ required: true, message: '请输入对讲内容' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入要发送给驾驶员的语音内容，系统将通过TTS播报..." />
          </Form.Item>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {[
              '请注意驾驶安全',
              '前方有服务区，建议休息',
              '检测到疲劳，请立即停车休息',
              '请勿使用手机',
              '请系好安全带',
            ].map(text => (
              <Tag
                key={text}
                color="blue"
                style={{ cursor: 'pointer', padding: '4px 10px' }}
                onClick={() => form.setFieldsValue({ message: text })}
              >
                {text}
              </Tag>
            ))}
          </div>
        </Form>
      </Modal>

      <Modal
        title={<Space><CoffeeOutlined /> 推荐服务区停靠</Space>}
        open={dispatchModal}
        onCancel={() => setDispatchModal(false)}
        footer={null}
        width={600}
      >
        {recommendedAreas.length ? (
          <List
            dataSource={recommendedAreas.slice(0, 5)}
            renderItem={(area: any) => (
              <Card
                size="small"
                style={{ marginBottom: 12, borderRadius: 8 }}
                title={
                  <Space>
                    <Tag color={area.has_danger_goods_parking ? 'green' : 'orange'}>
                      {area.has_danger_goods_parking ? '危化品停车位' : '普通停车位'}
                    </Tag>
                    <Text strong>{area.name}</Text>
                    <Text type="secondary">{area.highway_name}</Text>
                  </Space>
                }
                extra={
                  <Space>
                    <Text type="secondary">{area.distance_from_current?.toFixed(1)} km</Text>
                    <Text type="warning">约 {area.estimated_arrival_time} 分钟</Text>
                    <Button
                      type="primary"
                      size="small"
                      onClick={() => handleConfirmDispatch(area.id, area.rest_duration_recommend || area.recommended_rest_duration || 20)}
                    >
                      调度停靠 {area.rest_duration_recommend || area.recommended_rest_duration || 20}分钟
                    </Button>
                  </Space>
                }
              >
                <Space size={16} wrap>
                  {area.has_restaurant && <Tag>🍴 餐厅</Tag>}
                  {area.has_hotel && <Tag>🛏️ 住宿</Tag>}
                  {area.has_fuel_station && <Tag>⛽ 加油站</Tag>}
                  {area.has_charging && <Tag>🔋 充电桩</Tag>}
                  {area.has_maintenance && <Tag>🔧 维修</Tag>}
                  {area.phone && <Tag><PhoneOutlined /> {area.phone}</Tag>}
                  {area.danger_parking_spaces > 0 && (
                    <Tag color="green">危化品车位 {area.danger_parking_spaces} 个</Tag>
                  )}
                </Space>
              </Card>
            )}
          />
        ) : (
          <Empty description="附近暂无推荐服务区" />
        )}
      </Modal>
    </div>
  )
}

export default Monitor
