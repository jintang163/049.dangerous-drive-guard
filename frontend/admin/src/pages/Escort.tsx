import React, { useEffect, useState, useRef } from 'react'
import {
  Row,
  Col,
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Form,
  Select,
  Input,
  Drawer,
  Descriptions,
  Image,
  message,
  Timeline,
  Divider,
  Badge,
  Tooltip,
  Empty,
  Modal,
  List,
  Avatar,
  Upload,
  Alert,
} from 'antd'
import {
  EnvironmentOutlined,
  CarOutlined,
  UserOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
  ExportOutlined,
  SafetyCertificateOutlined,
  MapOutlined,
  ClockCircleOutlined,
  CameraOutlined,
  EditOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CoffeeOutlined,
  LoadingOutlined,
  SendOutlined,
  FlagOutlined,
  ThunderboltOutlined,
  FileTextOutlined,
  PaperClipOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import {
  ProCard,
  ProFormText,
  ProFormSelect,
  ProFormTextArea,
  ModalForm,
  ActionType,
} from '@ant-design/pro-components'
import type { ProColumns } from '@ant-design/pro-components'
import api from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import AMapLoader from '@amap/amap-jsapi-loader'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TextArea } = Input
const { Dragger } = Upload

type EscortEventType =
  | 'departure_check'
  | 'waypoint'
  | 'abnormal_stop'
  | 'rest'
  | 'loading_unloading'
  | 'sign_receipt'
  | 'emergency'

type WaybillStatus = 'pending' | 'transit' | 'signed' | 'abnormal' | 'cancelled'

interface EscortEvent {
  id: string
  type: EscortEventType
  time: string
  location: string
  lng: number
  lat: number
  photos: string[]
  remark: string
  operator: string
  waybill_no: string
  duration_minutes?: number
  risk_level?: 'normal' | 'attention' | 'warning' | 'danger'
}

interface EscortWaybill {
  id: string
  waybill_no: string
  status: WaybillStatus
  danger_goods_name: string
  un_number: string
  danger_level: string
  origin: string
  destination: string
  vehicle_plate: string
  driver_name: string
  driver_phone: string
  escort_name: string
  escort_phone: string
  planned_departure: string
  progress: number
  current_location?: string
  current_lng?: number
  current_lat?: number
  event_count: number
  last_event_time?: string
  start_lng: number
  start_lat: number
  end_lng: number
  end_lat: number
}

const eventTypeMap: Record<EscortEventType, {
  label: string
  color: string
  icon: React.ReactNode
  dot: 'blue' | 'green' | 'orange' | 'red' | 'purple' | 'cyan' | 'magenta'
}> = {
  departure_check: {
    label: '出发检查',
    color: 'blue',
    icon: <CheckCircleOutlined />,
    dot: 'blue',
  },
  waypoint: {
    label: '途经点',
    color: 'cyan',
    icon: <FlagOutlined />,
    dot: 'cyan',
  },
  abnormal_stop: {
    label: '异常停靠',
    color: 'orange',
    icon: <WarningOutlined />,
    dot: 'orange',
  },
  rest: {
    label: '休息',
    color: 'green',
    icon: <CoffeeOutlined />,
    dot: 'green',
  },
  loading_unloading: {
    label: '装卸货',
    color: 'purple',
    icon: <LoadingOutlined />,
    dot: 'purple',
  },
  sign_receipt: {
    label: '签收',
    color: 'green',
    icon: <FileTextOutlined />,
    dot: 'green',
  },
  emergency: {
    label: '突发事件',
    color: 'red',
    icon: <ThunderboltOutlined />,
    dot: 'red',
  },
}

const statusMap: Record<WaybillStatus, { label: string; color: string }> = {
  pending: { label: '待调度', color: 'default' },
  transit: { label: '运输中', color: 'processing' },
  signed: { label: '已签收', color: 'success' },
  abnormal: { label: '异常', color: 'error' },
  cancelled: { label: '取消', color: 'warning' },
}

const dangerLevelColorMap: Record<string, string> = {
  '1.1': 'red', '1.2': 'red', '1.3': 'red', '1.4': 'red',
  '2.1': 'volcano', '2.2': 'orange', '2.3': 'red',
  '3': 'red',
  '4.1': 'orange', '4.2': 'gold', '4.3': 'volcano',
  '5.1': 'purple', '5.2': 'magenta',
  '6.1': 'red', '6.2': 'magenta',
  '7': 'geekblue',
  '8': 'purple',
  '9': 'default',
}

const cities = [
  { name: '上海市浦东新区', lng: 121.4737, lat: 31.2304 },
  { name: '北京市朝阳区', lng: 116.4074, lat: 39.9042 },
  { name: '广州市天河区', lng: 113.3245, lat: 23.1291 },
  { name: '深圳市南山区', lng: 113.9304, lat: 22.5333 },
  { name: '杭州市西湖区', lng: 120.1551, lat: 30.2741 },
  { name: '成都市武侯区', lng: 104.0668, lat: 30.5728 },
  { name: '武汉市江汉区', lng: 114.3055, lat: 30.5931 },
  { name: '南京市鼓楼区', lng: 118.778, lat: 32.0617 },
  { name: '苏州市工业园区', lng: 120.5853, lat: 31.2990 },
  { name: '青岛市市南区', lng: 120.3826, lat: 36.0671 },
]

const mockEscortWaybills: EscortWaybill[] = Array.from({ length: 12 }, (_, i) => {
  const statuses: WaybillStatus[] = ['transit', 'transit', 'transit', 'transit', 'signed', 'transit']
  const status = statuses[i % statuses.length]
  const goodsList = [
    { name: '汽油', un: 'UN1203', level: '3' },
    { name: '液化石油气', un: 'UN1075', level: '2.1' },
    { name: '硫酸', un: 'UN2790', level: '8' },
    { name: '液氯', un: 'UN1017', level: '2.3' },
    { name: '烟花爆竹', un: 'UN0336', level: '1.4' },
  ]
  const goods = goodsList[i % goodsList.length]
  const origin = cities[i % cities.length]
  const dest = cities[(i + 4) % cities.length]
  const plates = ['沪A12345', '京B67890', '粤C11111', '粤D22222', '浙E33333', '川F44444']
  const drivers = ['张建国', '李明辉', '王志强', '刘文华', '陈晓峰', '赵大海']
  const escorts = ['李安全', '王押运', '张督察', '刘监管', '陈守卫', '孙警戒']
  const planned = dayjs().add(i * 5 - 25, 'hour')
  const progress = status === 'signed' ? 100 : Math.floor(Math.random() * 75) + 15
  const eventCount = Math.floor(Math.random() * 6) + 3

  return {
    id: `ewb_${10000 + i}`,
    waybill_no: `DDG${dayjs().format('YYYYMM')}${String(2000 + i).padStart(4, '0')}`,
    status,
    danger_goods_name: goods.name,
    un_number: goods.un,
    danger_level: goods.level,
    origin: origin.name,
    destination: dest.name,
    vehicle_plate: plates[i % plates.length],
    driver_name: drivers[i % drivers.length],
    driver_phone: `138${String(10000000 + i * 246).slice(0, 8)}`,
    escort_name: escorts[i % escorts.length],
    escort_phone: `139${String(10000000 + i * 357).slice(0, 8)}`,
    planned_departure: planned.toISOString(),
    progress,
    current_location: `${origin.name} → ${dest.name} 途中约 ${progress}%`,
    current_lng: (origin.lng + dest.lng) / 2 + (Math.random() - 0.5) * 1,
    current_lat: (origin.lat + dest.lat) / 2 + (Math.random() - 0.5) * 0.8,
    event_count: eventCount,
    last_event_time: planned.add(Math.floor(progress / 10), 'hour').toISOString(),
    start_lng: origin.lng,
    start_lat: origin.lat,
    end_lng: dest.lng,
    end_lat: dest.lat,
  }
})

const generateMockEvents = (waybill: EscortWaybill): EscortEvent[] => {
  const events: EscortEvent[] = []
  const types: EscortEventType[] = [
    'departure_check',
    'waypoint',
    'rest',
    'waypoint',
    'loading_unloading',
    'abnormal_stop',
    'waypoint',
    'sign_receipt',
  ]
  const count = waybill.event_count
  const baseTime = dayjs(waybill.planned_departure)

  for (let i = 0; i < count; i++) {
    const typeIdx = Math.min(i, types.length - 1)
    const type = i === count - 1 && waybill.status === 'signed' ? 'sign_receipt' : types[typeIdx]
    const progressRatio = (i + 1) / count
    const curLng = waybill.start_lng + (waybill.end_lng - waybill.start_lng) * progressRatio + (Math.random() - 0.5) * 0.3
    const curLat = waybill.start_lat + (waybill.end_lat - waybill.start_lat) * progressRatio + (Math.random() - 0.5) * 0.2
    const nearestCity = cities.reduce((prev, curr) => {
      const pd = Math.hypot(prev.lng - curLng, prev.lat - curLat)
      const cd = Math.hypot(curr.lng - curLng, curr.lat - curLat)
      return cd < pd ? curr : prev
    })
    const riskRoll = Math.random()
    const riskLevel: EscortEvent['risk_level'] = type === 'emergency' || type === 'abnormal_stop'
      ? (riskRoll > 0.5 ? 'warning' : 'danger')
      : (riskRoll > 0.8 ? 'attention' : 'normal')

    events.push({
      id: `evt_${waybill.id}_${i}`,
      type,
      time: baseTime.add(i * (2 + Math.random() * 2), 'hour').add(Math.random() * 30, 'minute').toISOString(),
      location: `${nearestCity.name}附近 (${curLng.toFixed(4)}, ${curLat.toFixed(4)})`,
      lng: curLng,
      lat: curLat,
      photos: Math.random() > 0.3 ? [
        `https://picsum.photos/seed/escort${i}${waybill.id}/400/300`,
      ] : [],
      remark: type === 'departure_check'
        ? '车辆安全检查通过：刹车、轮胎、灯光、灭火器、应急出口均正常。货物固定完好。驾驶员酒精检测0mg/100ml。'
        : type === 'waypoint'
          ? `途经检查点${i}，车速68km/h，驾驶员状态良好，未偏离规划路线。`
          : type === 'rest'
            ? '进入服务区休息25分钟，驾驶员已签到，车辆熄火锁闭，周边巡视正常。'
            : type === 'loading_unloading'
              ? '到达临时装卸点，已完成接地装置连接，操作人员均佩戴防护装备。'
              : type === 'abnormal_stop'
                ? '因前方交通事故临时停靠路边，已开启双闪，设置警示标志，已报告调度中心。'
                : type === 'sign_receipt'
                  ? '货物完好无损到达目的地，收货方核对数量无误，完成电子签收。'
                  : '⚠️ 发生突发事件：车辆轻微颠簸，已检查货物未发现异常，已减速慢行继续观察。',
      operator: waybill.escort_name,
      waybill_no: waybill.waybill_no,
      duration_minutes: type === 'rest' ? 25 : type === 'loading_unloading' ? 80 : undefined,
      risk_level: riskLevel,
    })
  }
  return events.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
}

const Escort: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [selectedWaybill, setSelectedWaybill] = useState<EscortWaybill | null>(mockEscortWaybills[0])
  const [events, setEvents] = useState<EscortEvent[]>([])
  const [eventDetailDrawer, setEventDetailDrawer] = useState<EscortEvent | null>(null)
  const [addEventModalVisible, setAddEventModalVisible] = useState(false)
  const [addEventForm] = Form.useForm()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [searchKeyword, setSearchKeyword] = useState('')

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const eventMapContainerRef = useRef<HTMLDivElement>(null)
  const mapInstanceRef = useRef<any>(null)
  const eventMapInstanceRef = useRef<any>(null)
  const [eventMapReady, setEventMapReady] = useState(false)

  useEffect(() => {
    if (selectedWaybill) {
      setLoading(true)
      setTimeout(() => {
        setEvents(generateMockEvents(selectedWaybill))
        setLoading(false)
      }, 300)
    }
  }, [selectedWaybill])

  const initEventMap = async (event: EscortEvent) => {
    if (!eventMapContainerRef.current) return
    try {
      const AMap: any = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: ['AMap.Scale', 'AMap.ToolBar'],
      })
      const map = new AMap.Map(eventMapContainerRef.current, {
        zoom: 13,
        center: [event.lng, event.lat],
        mapStyle: 'amap://styles/light',
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar())
      new AMap.Marker({
        position: [event.lng, event.lat],
        label: {
          content: `<div style="padding:4px 10px;background:${event.risk_level === 'danger' ? '#ff4d4f' : event.risk_level === 'warning' ? '#faad14' : '#1677ff'};color:#fff;border-radius:4px;font-size:12px;font-weight:600">
            ${eventTypeMap[event.type].label}
          </div>`,
          direction: 'top',
        },
        map,
      })
      eventMapInstanceRef.current = map
      setEventMapReady(true)
    } catch (e) {
    }
  }

  useEffect(() => {
    if (eventDetailDrawer && !eventMapReady) {
      const timer = setTimeout(() => initEventMap(eventDetailDrawer), 200)
      return () => {
        clearTimeout(timer)
        if (eventMapInstanceRef.current) {
          eventMapInstanceRef.current.destroy()
          eventMapInstanceRef.current = null
          setEventMapReady(false)
        }
      }
    }
    return () => {
      if (eventMapInstanceRef.current && !eventDetailDrawer) {
        eventMapInstanceRef.current.destroy()
        eventMapInstanceRef.current = null
        setEventMapReady(false)
      }
    }
  }, [eventDetailDrawer])

  const filteredWaybills = mockEscortWaybills.filter(w => {
    if (statusFilter && w.status !== statusFilter) return false
    if (searchKeyword && !(
      w.waybill_no.includes(searchKeyword) ||
      w.vehicle_plate.includes(searchKeyword) ||
      w.driver_name.includes(searchKeyword) ||
      w.danger_goods_name.includes(searchKeyword)
    )) return false
    return true
  })

  const handleAddEvent = async (values: any) => {
    if (!selectedWaybill) return
    const newEvent: EscortEvent = {
      id: `evt_manual_${Date.now()}`,
      type: values.event_type,
      time: new Date().toISOString(),
      location: values.location || `${selectedWaybill.current_location}`,
      lng: selectedWaybill.current_lng || 116.4,
      lat: selectedWaybill.current_lat || 39.9,
      photos: [],
      remark: values.remark,
      operator: '人工添加-调度员',
      waybill_no: selectedWaybill.waybill_no,
      risk_level: values.event_type === 'emergency' ? 'warning' : 'normal',
    }
    setEvents(prev => [newEvent, ...prev])
    message.success('押运事件添加成功')
    setAddEventModalVisible(false)
    addEventForm.resetFields()
  }

  const handleExportReport = () => {
    if (!selectedWaybill) {
      message.warning('请先选择一个运单')
      return
    }
    message.success(`正在生成运单 ${selectedWaybill.waybill_no} 的押运报告...`)
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 12 }}>
        <Row gutter={16} align="middle">
          <Col xs={24} md={10}>
            <Space wrap>
              <Text strong style={{ fontSize: 15 }}>
                <SafetyCertificateOutlined style={{ color: '#1677ff' }} /> 电子押运管理
              </Text>
              <Tag color="blue">运输途中 {mockEscortWaybills.filter(w => w.status === 'transit').length}</Tag>
              <Tag color="success">已完成 {mockEscortWaybills.filter(w => w.status === 'signed').length}</Tag>
            </Space>
          </Col>
          <Col xs={24} md={8}>
            <Space.Compact style={{ width: '100%' }}>
              <Input
                allowClear
                prefix={<SearchOutlined />}
                placeholder="搜索运单号/车牌/司机/货物"
                value={searchKeyword}
                onChange={e => setSearchKeyword(e.target.value)}
              />
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
            </Space.Compact>
          </Col>
          <Col xs={24} md={6}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }} wrap>
              <Button icon={<ReloadOutlined />}>刷新</Button>
              <Button
                icon={<PlusOutlined />}
                type="primary"
                onClick={() => setAddEventModalVisible(true)}
                disabled={!selectedWaybill || selectedWaybill.status !== 'transit'}
              >
                人工添加事件
              </Button>
              <Button icon={<ExportOutlined />} onClick={handleExportReport}>
                导出押运报告
              </Button>
            </Space>
          </Col>
        </Row>
      </ProCard>

      <Row gutter={16}>
        <Col xs={24} lg={9} xl={8}>
          <ProCard
            bordered={false}
            style={{ borderRadius: 12, height: 'calc(100vh - 240px)', minHeight: 600 }}
            bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column' }}
            title={
              <Space>
                <CarOutlined style={{ color: '#1677ff' }} />
                <Text strong>押运任务列表</Text>
                <Badge count={filteredWaybills.length} showZero style={{ backgroundColor: '#1677ff' }} />
              </Space>
            }
          >
            <div style={{ overflowY: 'auto', flex: 1, padding: 12 }}>
              {filteredWaybills.length === 0 ? (
                <Empty description="暂无匹配的押运任务" style={{ marginTop: 60 }} />
              ) : (
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  {filteredWaybills.map(waybill => (
                    <Card
                      key={waybill.id}
                      size="small"
                      hoverable
                      onClick={() => setSelectedWaybill(waybill)}
                      style={{
                        borderRadius: 10,
                        cursor: 'pointer',
                        border: selectedWaybill?.id === waybill.id ? '2px solid #1677ff' : '1px solid #f0f0f0',
                        background: selectedWaybill?.id === waybill.id ? '#f0f7ff' : '#fff',
                        boxShadow: selectedWaybill?.id === waybill.id ? '0 4px 12px rgba(22,119,255,0.15)' : 'none',
                        transition: 'all 0.2s',
                      }}
                      bodyStyle={{ padding: 12 }}
                    >
                      <Space direction="vertical" size={8} style={{ width: '100%' }}>
                        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                          <Space>
                            <Text copyable strong style={{ fontSize: 13 }}>{waybill.waybill_no}</Text>
                            <Tag color={statusMap[waybill.status].color} style={{ margin: 0 }}>
                              {statusMap[waybill.status].label}
                            </Tag>
                          </Space>
                          <Tag color={dangerLevelColorMap[waybill.danger_level] || 'default'} style={{ fontSize: 11 }}>
                            类{waybill.danger_level} · {waybill.danger_goods_name}
                          </Tag>
                        </Space>

                        <Row gutter={8}>
                          <Col span={12}>
                            <Space direction="vertical" size={2}>
                              <Space size={4}>
                                <CarOutlined style={{ color: '#1677ff', fontSize: 11 }} />
                                <Tag color="blue" style={{ fontSize: 11, padding: '0 6px', margin: 0 }}>
                                  {waybill.vehicle_plate}
                                </Tag>
                              </Space>
                              <Space size={4}>
                                <UserOutlined style={{ color: '#fa8c16', fontSize: 11 }} />
                                <Text style={{ fontSize: 11 }}>{waybill.driver_name}</Text>
                              </Space>
                              <Space size={4}>
                                <SafetyCertificateOutlined style={{ color: '#52c41a', fontSize: 11 }} />
                                <Text style={{ fontSize: 11 }}>{waybill.escort_name}</Text>
                              </Space>
                            </Space>
                          </Col>
                          <Col span={12}>
                            <Space direction="vertical" size={2} style={{ width: '100%' }}>
                              <div>
                                <EnvironmentOutlined style={{ color: '#52c41a', fontSize: 11 }} />
                                <Text ellipsis style={{ fontSize: 11, maxWidth: 100 }} title={waybill.origin}>
                                  {' '}{waybill.origin.split('区')[0]}区
                                </Text>
                              </div>
                              <div style={{ paddingLeft: 14, color: '#d9d9d9', fontSize: 8 }}>↓</div>
                              <div>
                                <EnvironmentOutlined style={{ color: '#ff4d4f', fontSize: 11 }} />
                                <Text ellipsis style={{ fontSize: 11, maxWidth: 100 }} title={waybill.destination}>
                                  {' '}{waybill.destination.split('区')[0]}区
                                </Text>
                              </div>
                            </Space>
                          </Col>
                        </Row>

                        <Divider style={{ margin: '4px 0' }} />

                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <Text type="secondary" style={{ fontSize: 11 }}>运输进度</Text>
                            <Text strong style={{ fontSize: 11, color: waybill.progress > 80 ? '#52c41a' : '#1677ff' }}>
                              {waybill.progress}%
                            </Text>
                          </div>
                          <div
                            style={{
                              height: 6,
                              background: '#f0f0f0',
                              borderRadius: 3,
                              overflow: 'hidden',
                            }}
                          >
                            <div
                              style={{
                                width: `${waybill.progress}%`,
                                height: '100%',
                                background: `linear-gradient(90deg, #1677ff 0%, ${waybill.progress > 80 ? '#52c41a' : '#13c2c2'} 100%)`,
                                borderRadius: 3,
                                transition: 'width 0.3s',
                              }}
                            />
                          </div>
                          <Space style={{ width: '100%', justifyContent: 'space-between' }} size={4}>
                            <Space size={4}>
                              <ClockCircleOutlined style={{ color: '#8c8c8c', fontSize: 11 }} />
                              <Text type="secondary" style={{ fontSize: 11 }}>
                                {waybill.last_event_time ? dayjs(waybill.last_event_time).format('MM-DD HH:mm') : '-'}
                              </Text>
                            </Space>
                            <Tag color="purple" style={{ fontSize: 11, padding: '0 6px', margin: 0 }}>
                              {waybill.event_count} 条事件
                            </Tag>
                          </Space>
                        </Space>
                      </Space>
                    </Card>
                  ))}
                </Space>
              )}
            </div>
          </ProCard>
        </Col>

        <Col xs={24} lg={15} xl={16}>
          <ProCard
            bordered={false}
            style={{ borderRadius: 12, height: 'calc(100vh - 240px)', minHeight: 600 }}
            bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column' }}
            title={
              selectedWaybill ? (
                <Space wrap>
                  <Space>
                    <SafetyCertificateOutlined style={{ color: '#52c41a' }} />
                    <Text strong>押运事件流</Text>
                  </Space>
                  <Text copyable style={{ fontSize: 13 }}>{selectedWaybill.waybill_no}</Text>
                  <Tag color="blue">{selectedWaybill.vehicle_plate}</Tag>
                  <Tag color={dangerLevelColorMap[selectedWaybill.danger_level]}>
                    {selectedWaybill.danger_goods_name} ({selectedWaybill.un_number})
                  </Tag>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    <UserOutlined /> 司机: {selectedWaybill.driver_name} · 押运: {selectedWaybill.escort_name}
                  </Text>
                </Space>
              ) : (
                <Space>
                  <SafetyCertificateOutlined style={{ color: '#52c41a' }} />
                  <Text strong>押运事件流</Text>
                </Space>
              )
            }
            extra={
              selectedWaybill && (
                <Space size={4}>
                  <Tag color="green">{events.filter(e => e.risk_level === 'normal').length} 正常</Tag>
                  <Tag color="gold">{events.filter(e => e.risk_level === 'attention').length} 关注</Tag>
                  <Tag color="orange">{events.filter(e => e.risk_level === 'warning').length} 预警</Tag>
                  <Tag color="red">{events.filter(e => e.risk_level === 'danger').length} 危险</Tag>
                </Space>
              )
            }
          >
            <div style={{ overflowY: 'auto', flex: 1, padding: '16px 24px' }}>
              {!selectedWaybill ? (
                <Empty
                  description={<Space direction="vertical" align="center"><Text>请从左侧选择一个押运任务</Text><Text type="secondary">选择后查看该运单的完整押运事件时间线</Text></Space>}
                  style={{ marginTop: 100 }}
                />
              ) : loading ? (
                <div style={{ textAlign: 'center', padding: 60 }}>
                  <Space direction="vertical" align="center">
                    <LoadingOutlined style={{ fontSize: 32, color: '#1677ff' }} />
                    <Text type="secondary">加载押运事件中...</Text>
                  </Space>
                </div>
              ) : events.length === 0 ? (
                <Empty description="暂无押运事件" style={{ marginTop: 60 }} />
              ) : (
                <Timeline
                  mode="left"
                  style={{ paddingLeft: 0 }}
                  items={events.map((event, idx) => {
                    const et = eventTypeMap[event.type]
                    return {
                      color: et.dot,
                      dot: (
                        <div style={{
                          width: 32, height: 32, borderRadius: '50%',
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          background: event.risk_level === 'danger' ? '#fff1f0'
                            : event.risk_level === 'warning' ? '#fff7e6'
                              : event.risk_level === 'attention' ? '#fffbe6'
                                : '#f0f7ff',
                          border: `2px solid ${event.risk_level === 'danger' ? '#ff4d4f'
                            : event.risk_level === 'warning' ? '#faad14'
                              : event.risk_level === 'attention' ? '#faad14'
                                : '#1677ff'}`,
                          color: event.risk_level === 'danger' ? '#ff4d4f'
                            : event.risk_level === 'warning' ? '#faad14'
                              : event.risk_level === 'attention' ? '#faad14'
                                : '#1677ff',
                          fontSize: 14,
                          position: 'relative',
                          left: 0,
                        }}>
                          {et.icon}
                        </div>
                      ),
                      label: (
                        <div style={{ paddingTop: 6 }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {formatDateTime(event.time, 'MM-DD HH:mm:ss')}
                          </Text>
                          {event.duration_minutes && (
                            <div>
                              <Tag color="default" style={{ fontSize: 11, marginTop: 4 }}>
                                持续 {event.duration_minutes} 分钟
                              </Tag>
                            </div>
                          )}
                        </div>
                      ),
                      children: (
                        <Card
                          size="small"
                          style={{
                            marginBottom: 16,
                            borderRadius: 10,
                            borderLeft: `4px solid ${event.risk_level === 'danger' ? '#ff4d4f'
                              : event.risk_level === 'warning' ? '#faad14'
                                : event.risk_level === 'attention' ? '#faad14'
                                  : '#1677ff'}`,
                            background: event.risk_level === 'danger' ? '#fffbfb'
                              : event.risk_level === 'warning' ? '#fffbf5'
                                : '#fff',
                          }}
                          bodyStyle={{ padding: 12 }}
                        >
                          <Space direction="vertical" size={8} style={{ width: '100%' }}>
                            <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                              <Space>
                                <Tag color={et.color} style={{ margin: 0, fontSize: 12 }}>
                                  <strong>{et.label}</strong>
                                </Tag>
                                {event.risk_level && event.risk_level !== 'normal' && (
                                  <Badge
                                    status={event.risk_level === 'danger' ? 'error' : event.risk_level === 'warning' ? 'warning' : 'processing'}
                                    text={
                                      <Text style={{ fontSize: 11 }} type={event.risk_level === 'danger' ? 'danger' : 'warning'}>
                                        {event.risk_level === 'danger' ? '高风险' : event.risk_level === 'warning' ? '需关注' : '一般'}
                                      </Text>
                                    }
                                  />
                                )}
                              </Space>
                              <Button
                                type="link"
                                size="small"
                                icon={<EyeOutlined />}
                                onClick={() => setEventDetailDrawer(event)}
                              >
                                查看详情
                              </Button>
                            </Space>

                            <Space size={6} wrap>
                              <EnvironmentOutlined style={{ color: '#1677ff' }} />
                              <Tooltip title={event.location}>
                                <Text style={{ fontSize: 12 }} ellipsis style={{ maxWidth: 320 }}>
                                  {event.location}
                                </Text>
                              </Tooltip>
                            </Space>

                            <Paragraph
                              style={{ fontSize: 12, marginBottom: 0 }}
                              ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
                            >
                              {event.remark}
                            </Paragraph>

                            {event.photos.length > 0 && (
                              <Space size={8} wrap>
                                {event.photos.map((p, pi) => (
                                  <Image
                                    key={pi}
                                    src={p}
                                    alt=""
                                    width={100}
                                    height={75}
                                    style={{ borderRadius: 6, objectFit: 'cover', cursor: 'pointer' }}
                                    preview={{ mask: <EyeOutlined /> }}
                                  />
                                ))}
                              </Space>
                            )}

                            <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                              <Space>
                                <Avatar size={20} icon={<UserOutlined />} style={{ background: '#1677ff' }} />
                                <Text type="secondary" style={{ fontSize: 11 }}>{event.operator}</Text>
                              </Space>
                              <Space size={4}>
                                <PaperClipOutlined style={{ color: '#8c8c8c', fontSize: 11 }} />
                                <Text type="secondary" style={{ fontSize: 11 }}>
                                  {event.photos.length} 张照片
                                </Text>
                              </Space>
                            </Space>
                          </Space>
                        </Card>
                      ),
                    }
                  })}
                />
              )}
            </div>
          </ProCard>
        </Col>
      </Row>

      <Drawer
        title={
          eventDetailDrawer ? (
            <Space>
              <span style={{
                display: 'inline-flex', width: 28, height: 28, borderRadius: '50%',
                alignItems: 'center', justifyContent: 'center',
                background: eventTypeMap[eventDetailDrawer.type].color === 'red' ? '#fff1f0'
                  : eventTypeMap[eventDetailDrawer.type].color === 'orange' ? '#fff7e6'
                    : eventTypeMap[eventDetailDrawer.type].color === 'green' ? '#f6ffed'
                      : '#e6f4ff',
                color: eventTypeMap[eventDetailDrawer.type].color === 'red' ? '#ff4d4f'
                  : eventTypeMap[eventDetailDrawer.type].color === 'orange' ? '#faad14'
                    : eventTypeMap[eventDetailDrawer.type].color === 'green' ? '#52c41a'
                      : '#1677ff',
                fontSize: 14,
              }}>
                {eventTypeMap[eventDetailDrawer.type].icon}
              </span>
              <Text strong>{eventTypeMap[eventDetailDrawer.type].label} - 事件详情</Text>
              <Tag color={eventTypeMap[eventDetailDrawer.type].color}>
                {eventDetailDrawer.waybill_no}
              </Tag>
              {eventDetailDrawer.risk_level && eventDetailDrawer.risk_level !== 'normal' && (
                <Tag color={eventDetailDrawer.risk_level === 'danger' ? 'red' : 'orange'}>
                  {eventDetailDrawer.risk_level === 'danger' ? '⚠️ 高风险' : '⚡ 需关注'}
                </Tag>
              )}
            </Space>
          ) : null
        }
        open={!!eventDetailDrawer}
        onClose={() => {
          setEventDetailDrawer(null)
          setEventMapReady(false)
          if (eventMapInstanceRef.current) {
            eventMapInstanceRef.current.destroy()
            eventMapInstanceRef.current = null
          }
        }}
        width={560}
        extra={
          eventDetailDrawer && (
            <Space>
              <Button icon={<ExportOutlined />}>导出详情</Button>
              {(eventDetailDrawer.risk_level === 'warning' || eventDetailDrawer.risk_level === 'danger') && (
                <Button type="primary" danger icon={<WarningOutlined />}>上报处置</Button>
              )}
            </Space>
          )
        }
      >
        {eventDetailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {eventDetailDrawer.risk_level === 'danger' && (
              <Alert
                type="error"
                showIcon
                icon={<WarningOutlined />}
                message="高风险事件提醒"
                description="该事件为高风险押运事件，建议立即核查并采取处置措施"
                style={{ borderRadius: 8 }}
              />
            )}
            {eventDetailDrawer.risk_level === 'warning' && (
              <Alert
                type="warning"
                showIcon
                message="需关注事件"
                description="该事件存在一定风险，请持续关注后续进展"
                style={{ borderRadius: 8 }}
              />
            )}

            <Card size="small" style={{ borderRadius: 8 }}>
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="事件编号">
                  <Text copyable>{eventDetailDrawer.id}</Text>
                </Descriptions.Item>
                <Descriptions.Item label="事件类型">
                  <Tag color={eventTypeMap[eventDetailDrawer.type].color} style={{ fontSize: 13 }}>
                    {eventTypeMap[eventDetailDrawer.type].label}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="关联运单">{eventDetailDrawer.waybill_no}</Descriptions.Item>
                <Descriptions.Item label="发生时间">{formatDateTime(eventDetailDrawer.time)}</Descriptions.Item>
                {eventDetailDrawer.duration_minutes && (
                  <Descriptions.Item label="持续时长">{eventDetailDrawer.duration_minutes} 分钟</Descriptions.Item>
                )}
                <Descriptions.Item label="发生位置">
                  <Paragraph copyable style={{ margin: 0, fontSize: 13 }}>
                    <EnvironmentOutlined /> {eventDetailDrawer.location}
                  </Paragraph>
                </Descriptions.Item>
                <Descriptions.Item label="经纬度">
                  <Text type="secondary">({eventDetailDrawer.lng.toFixed(6)}, {eventDetailDrawer.lat.toFixed(6)})</Text>
                </Descriptions.Item>
                <Descriptions.Item label="记录人员">
                  <Space>
                    <Avatar size={20} icon={<UserOutlined />} style={{ background: '#1677ff' }} />
                    {eventDetailDrawer.operator}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="风险等级">
                  <Tag color={
                    eventDetailDrawer.risk_level === 'danger' ? 'red'
                      : eventDetailDrawer.risk_level === 'warning' ? 'orange'
                        : eventDetailDrawer.risk_level === 'attention' ? 'gold'
                          : 'green'
                  }>
                    {eventDetailDrawer.risk_level === 'danger' ? '高风险'
                      : eventDetailDrawer.risk_level === 'warning' ? '预警'
                        : eventDetailDrawer.risk_level === 'attention' ? '关注'
                          : '正常'}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><MapOutlined />位置地图</Space>}>
              <div
                ref={eventMapContainerRef}
                style={{
                  height: 240,
                  borderRadius: 6,
                  background: '#f0f5ff',
                  display: eventMapReady ? 'block' : 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {!eventMapReady && (
                  <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                    <MapOutlined style={{ fontSize: 36, opacity: 0.4 }} />
                    <div style={{ marginTop: 8, fontSize: 12 }}>
                      {eventDetailDrawer.location}
                    </div>
                  </div>
                )}
              </div>
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><EditOutlined />押运员备注</Space>}>
              <Paragraph style={{ fontSize: 13, margin: 0, lineHeight: 1.8 }}>
                {eventDetailDrawer.remark}
              </Paragraph>
            </Card>

            {eventDetailDrawer.photos.length > 0 ? (
              <Card size="small" style={{ borderRadius: 8 }} title={<Space><CameraOutlined />现场照片 ({eventDetailDrawer.photos.length})</Space>}>
                <Row gutter={8}>
                  {eventDetailDrawer.photos.map((p, i) => (
                    <Col span={12} key={i} style={{ marginBottom: 8 }}>
                      <Image
                        src={p}
                        alt=""
                        style={{
                          width: '100%',
                          height: 140,
                          objectFit: 'cover',
                          borderRadius: 6,
                        }}
                        preview={{ mask: <EyeOutlined /> }}
                      />
                    </Col>
                  ))}
                </Row>
              </Card>
            ) : (
              <Card size="small" style={{ borderRadius: 8 }} title={<Space><CameraOutlined />现场照片</Space>}>
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该事件未上传照片" style={{ padding: '20px 0' }} />
              </Card>
            )}
          </div>
        )}
      </Drawer>

      <ModalForm
        title={
          <Space>
            <PlusOutlined style={{ color: '#1677ff' }} />
            <Text strong>人工添加押运事件</Text>
            {selectedWaybill && <Tag color="blue">{selectedWaybill.waybill_no}</Tag>}
          </Space>
        }
        open={addEventModalVisible}
        onOpenChange={setAddEventModalVisible}
        form={addEventForm}
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        onFinish={handleAddEvent}
        layout="vertical"
        submitTimeout={2000}
        submitter={{
          searchConfig: { submitText: '确认添加', resetText: '取消' },
          resetButtonProps: { onClick: () => setAddEventModalVisible(false) },
          submitButtonProps: { type: 'primary' },
        }}
        width={520}
      >
        <Alert
          type="info"
          showIcon
          message="正在为选中的运单添加押运事件"
          description={selectedWaybill ? `${selectedWaybill.waybill_no} · ${selectedWaybill.vehicle_plate} · ${selectedWaybill.driver_name}` : ''}
          style={{ marginBottom: 16, borderRadius: 8 }}
        />
        <ProFormSelect
          label="事件类型"
          name="event_type"
          width="md"
          rules={[{ required: true, message: '请选择事件类型' }]}
          placeholder="请选择押运事件类型"
          options={Object.entries(eventTypeMap).map(([k, v]) => ({
            label: v.label,
            value: k,
          }))}
        />
        <ProFormText
          label="发生位置"
          name="location"
          width="md"
          placeholder="请输入事件发生位置，留空则使用车辆当前位置"
          fieldProps={{
            prefix: <EnvironmentOutlined />,
          }}
        />
        <ProFormTextArea
          label="事件详情描述"
          name="remark"
          width="md"
          rules={[{ required: true, message: '请输入事件描述' }]}
          placeholder="请详细描述押运事件情况，包括处置措施等"
          fieldProps={{ rows: 4, showCount: true, maxLength: 500 }}
        />
        <Form.Item label="现场照片" valuePropName="fileList">
          <Dragger
            multiple
            listType="picture"
            beforeUpload={() => false}
            maxCount={9}
          >
            <p className="ant-upload-drag-icon"><CameraOutlined style={{ fontSize: 32, color: '#1677ff' }} /></p>
            <p className="ant-upload-text">点击或拖拽照片到此处上传</p>
            <p className="ant-upload-hint" style={{ fontSize: 12 }}>支持JPG/PNG格式，最多9张</p>
          </Dragger>
        </Form.Item>
        <Divider style={{ margin: '4px 0 12px' }} />
        <Alert
          type="warning"
          showIcon
          message="添加说明"
          description="押运事件将同步记录于运单档案，作为运输合规性凭证。添加后将自动通知相关人员。"
          style={{ borderRadius: 6 }}
        />
      </ModalForm>
    </div>
  )
}

export default Escort
