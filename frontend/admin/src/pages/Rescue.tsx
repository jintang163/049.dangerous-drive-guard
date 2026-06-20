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
  message,
  Timeline,
  Divider,
  Badge,
  Tooltip,
  Empty,
  Tabs,
  Table,
  Avatar,
  Alert,
  List,
  Statistic,
  Progress,
  Steps,
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
  PhoneOutlined,
  EditOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  MedicineBoxOutlined,
  FireOutlined,
  ToolOutlined,
  SendOutlined,
  ThunderboltOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
  HeartOutlined,
  SearchOutlined,
  HomeOutlined,
  TeamOutlined,
  ClockOutlined,
  RiseOutlined,
  StopOutlined,
  CaretRightOutlined,
} from '@ant-design/icons'
import {
  ProCard,
  ProFormText,
  ProFormSelect,
  ProFormTextArea,
  ActionType,
  ProTable,
  RequestData,
} from '@ant-design/pro-components'
import type { ProColumns } from '@ant-design/pro-components'
import api from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import AMapLoader from '@amap/amap-jsapi-loader'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs
const { TextArea } = Input

type RescueType = 'accident' | 'breakdown' | 'medical' | 'other'
type RescueStatus = 'pending' | 'responding' | 'arrived' | 'processing' | 'resolved' | 'closed' | 'cancelled'
type ResourceType = 'hospital' | 'fire' | 'repair'

interface RescueRequest {
  id: string
  request_no: string
  type: RescueType
  status: RescueStatus
  vehicle_plate: string
  driver_name: string
  driver_phone: string
  escort_name?: string
  danger_goods?: string
  danger_level?: string
  location: string
  lng: number
  lat: number
  request_time: string
  description: string
  severity: 'minor' | 'moderate' | 'severe' | 'critical'
  injuries?: number
  waybill_no?: string
  assigned_resources?: string[]
  response_time?: string
  arrival_time?: string
  resolved_time?: string
  eta_minutes?: number
  contact_name?: string
  contact_phone?: string
  timeline: { time: string; status: string; operator: string; note?: string }[]
  chat_messages: {
    id: string
    sender: string
    role: 'dispatcher' | 'driver' | 'rescue' | 'supervisor'
    time: string
    content: string
  }[]
}

interface RescueResource {
  id: string
  type: ResourceType
  name: string
  address: string
  lng: number
  lat: number
  phone: string
  contact: string
  capacity: string
  response_time_avg: number
  distance_km?: number
  available: boolean
  rating?: number
  specialties?: string[]
  status?: 'idle' | 'busy' | 'offline'
}

const rescueTypeMap: Record<RescueType, { label: string; color: string; icon: React.ReactNode }> = {
  accident: { label: '交通事故', color: 'red', icon: <CarOutlined /> },
  breakdown: { label: '车辆故障', color: 'orange', icon: <ToolOutlined /> },
  medical: { label: '医疗急救', color: 'volcano', icon: <MedicineBoxOutlined /> },
  other: { label: '其他事件', color: 'default', icon: <ExclamationCircleOutlined /> },
}

const rescueStatusMap: Record<RescueStatus, { label: string; color: string; step: number }> = {
  pending: { label: '待响应', color: 'red', step: 0 },
  responding: { label: '已派单', color: 'processing', step: 1 },
  arrived: { label: '已到达', color: 'orange', step: 2 },
  processing: { label: '处置中', color: 'warning', step: 3 },
  resolved: { label: '已解决', color: 'success', step: 4 },
  closed: { label: '已结案', color: 'default', step: 5 },
  cancelled: { label: '已取消', color: 'default', step: 5 },
}

const severityMap: Record<string, { label: string; color: string }> = {
  minor: { label: '轻微', color: 'green' },
  moderate: { label: '一般', color: 'blue' },
  severe: { label: '严重', color: 'orange' },
  critical: { label: '危重', color: 'red' },
}

const resourceTypeMap: Record<ResourceType, {
  label: string
  color: string
  icon: React.ReactNode
  markerColor: string
}> = {
  hospital: {
    label: '医院',
    color: 'volcano',
    icon: <MedicineBoxOutlined />,
    markerColor: '#f5222d',
  },
  fire: {
    label: '消防站',
    color: 'red',
    icon: <FireOutlined />,
    markerColor: '#ff4d4f',
  },
  repair: {
    label: '维修站',
    color: 'orange',
    icon: <ToolOutlined />,
    markerColor: '#fa8c16',
  },
}

const generatePoint = (baseLng: number, baseLat: number, radiusKm: number) => {
  const r = radiusKm / 111
  const theta = Math.random() * Math.PI * 2
  return {
    lng: baseLng + r * Math.cos(theta),
    lat: baseLat + r * Math.sin(theta),
  }
}

const calcDistance = (lng1: number, lat1: number, lng2: number, lat2: number) => {
  const R = 6371
  const dLat = (lat2 - lat1) * Math.PI / 180
  const dLng = (lng2 - lng1) * Math.PI / 180
  const a = Math.sin(dLat / 2) ** 2 +
    Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
    Math.sin(dLng / 2) ** 2
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
  return R * c
}

const baseLng = 121.4737
const baseLat = 31.2304

const mockRescueRequests: RescueRequest[] = Array.from({ length: 15 }, (_, i) => {
  const types: RescueType[] = ['accident', 'breakdown', 'medical', 'other', 'accident', 'breakdown']
  const statuses: RescueStatus[] = ['pending', 'responding', 'arrived', 'processing', 'resolved', 'closed', 'pending', 'responding']
  const severities: Array<RescueRequest['severity']> = ['minor', 'moderate', 'severe', 'critical', 'moderate']
  const type = types[i % types.length]
  const status = statuses[i % statuses.length]
  const severity = severities[i % severities.length]
  const point = generatePoint(baseLng, baseLat, 25)
  const plates = ['沪A12345', '京B67890', '粤C11111', '浙E33333', '川F44444', '苏H66666']
  const drivers = ['张建国', '李明辉', '王志强', '刘文华', '陈晓峰', '赵大海']
  const requestTime = dayjs().subtract(Math.random() * 180, 'minute')

  return {
    id: `rescue_${10000 + i}`,
    request_no: `SOS${dayjs().format('YYYYMMDD')}${String(100 + i).padStart(4, '0')}`,
    type,
    status,
    vehicle_plate: plates[i % plates.length],
    driver_name: drivers[i % drivers.length],
    driver_phone: `138${String(10000000 + i * 357).slice(0, 8)}`,
    escort_name: i % 3 === 0 ? `押运员${i + 1}` : undefined,
    danger_goods: i % 2 === 0 ? ['汽油', '液化石油气', '硫酸', '液氯'][i % 4] : undefined,
    danger_level: i % 2 === 0 ? ['3', '2.1', '8', '2.3'][i % 4] : undefined,
    location: `${['浦东新区', '徐汇区', '长宁区', '静安区', '普陀区', '虹口区'][i % 6]}${['陆家嘴', '张江', '金桥', '漕河泾', '虹桥', '北外滩'][i % 6]}附近`,
    lng: point.lng,
    lat: point.lat,
    request_time: requestTime.toISOString(),
    description: type === 'accident'
      ? '车辆与前车追尾，车头受损，驾驶员受轻微擦伤，货物未发现泄漏。'
      : type === 'breakdown'
        ? '车辆轮胎爆胎，已停靠应急车道，需要更换备胎及检修。'
        : type === 'medical'
          ? '驾驶员突发胸闷头晕，已靠边停车，需要医疗支援。'
          : '车辆无法启动，怀疑电路故障，请求救援。',
    severity,
    injuries: type === 'accident' ? (severity === 'critical' ? 3 : severity === 'severe' ? 2 : severity === 'moderate' ? 1 : 0) : 0,
    waybill_no: i % 2 === 0 ? `DDG${dayjs().format('YYYYMM')}${String(3000 + i).padStart(4, '0')}` : undefined,
    assigned_resources: status !== 'pending' ? [`${['医院A', '消防B', '维修C'][i % 3]}`, `救援车辆${i + 1}号`] : undefined,
    response_time: status !== 'pending' ? requestTime.add(2 + Math.random() * 5, 'minute').toISOString() : undefined,
    arrival_time: ['arrived', 'processing', 'resolved', 'closed'].includes(status)
      ? requestTime.add(10 + Math.random() * 20, 'minute').toISOString()
      : undefined,
    resolved_time: ['resolved', 'closed'].includes(status)
      ? requestTime.add(40 + Math.random() * 80, 'minute').toISOString()
      : undefined,
    eta_minutes: status === 'responding' ? Math.floor(Math.random() * 15) + 5 : undefined,
    contact_name: drivers[(i + 1) % drivers.length],
    contact_phone: `139${String(10000000 + i * 159).slice(0, 8)}`,
    timeline: [
      { time: requestTime.toISOString(), status: 'SOS请求', operator: '车载终端自动', note: '触发紧急救援SOS按钮' },
      ...(status !== 'pending' ? [{
        time: requestTime.add(2 + Math.random() * 5, 'minute').toISOString(),
        status: '调度响应', operator: '调度员小王', note: '已受理，正在协调最近救援资源',
      }] : []),
      ...(status !== 'pending' && status !== 'responding' ? [{
        time: requestTime.add(10 + Math.random() * 20, 'minute').toISOString(),
        status: '救援到达', operator: '消防/医疗/维修', note: `救援人员已到达现场，开始处置`,
      }] : []),
      ...(['processing', 'resolved', 'closed'].includes(status) ? [{
        time: requestTime.add(25 + Math.random() * 30, 'minute').toISOString(),
        status: '处置进行中', operator: '现场指挥', note: '伤员已送医/车辆正在抢修/火情已控制',
      }] : []),
      ...(['resolved', 'closed'].includes(status) ? [{
        time: requestTime.add(40 + Math.random() * 80, 'minute').toISOString(),
        status: '处置完成', operator: '现场指挥', note: '事件妥善处置，人员安全，货物未泄漏',
      }] : []),
      ...(status === 'closed' ? [{
        time: requestTime.add(60 + Math.random() * 120, 'minute').toISOString(),
        status: '结案', operator: '调度主管', note: '已完成事件报告归档',
      }] : []),
    ],
    chat_messages: [
      {
        id: `m1_${i}`,
        sender: '车载终端',
        role: 'driver',
        time: requestTime.toISOString(),
        content: '🚨 SOS！紧急救援请求！',
      },
      {
        id: `m2_${i}`,
        sender: '调度中心',
        role: 'dispatcher',
        time: requestTime.add(30, 'second').toISOString(),
        content: `已收到SOS，正在定位车辆位置...`,
      },
      {
        id: `m3_${i}`,
        sender: '调度中心',
        role: 'dispatcher',
        time: requestTime.add(1, 'minute').toISOString(),
        content: `📍 已定位：车牌号 ${plates[i % plates.length]}，位置已下发，请保持电话畅通。最近救援资源正在赶往现场。`,
      },
      ...(status !== 'pending' ? [{
        id: `m4_${i}`,
        sender: '救援队',
        role: 'rescue' as const,
        time: requestTime.add(3, 'minute').toISOString(),
        content: '🚑 救援队已出发，前往事故现场，请驾驶员在安全位置等待。',
      }] : []),
      ...(['arrived', 'processing', 'resolved', 'closed'].includes(status) ? [{
        id: `m5_${i}`,
        sender: '救援队',
        role: 'rescue' as const,
        time: requestTime.add(15, 'minute').toISOString(),
        content: '✅ 已到达现场，正在评估情况并开展救援工作。',
      }] : []),
      ...(['resolved', 'closed'].includes(status) ? [{
        id: `m6_${i}`,
        sender: '救援队',
        role: 'rescue' as const,
        time: requestTime.add(60, 'minute').toISOString(),
        content: '✔️ 事件已处置完毕，驾驶员安全，现场清理完成。',
      }] : []),
    ],
  }
})

const hospitalNames = [
  '上海浦东新区人民医院', '上海交通大学医学院附属仁济医院',
  '上海市第一人民医院', '复旦大学附属华山医院',
  '上海长征医院', '上海长海医院',
  '上海市第六人民医院', '上海市第十人民医院',
]
const fireStationNames = [
  '浦东新区消防救援支队陆家嘴中队', '徐汇区消防救援支队漕河泾中队',
  '长宁区消防救援支队虹桥中队', '静安区消防救援支队北站中队',
  '上海市消防救援总队特勤支队', '黄浦区消防救援支队外滩中队',
]
const repairNames = [
  '重汽维修服务站(浦东店)', '一汽解放维修中心',
  '东风商用车特约维修站', '危险品运输车专业维修厂',
  '重型卡车4S维修中心', '汽车应急救援维修服务',
]

const generateResources = (type: ResourceType, names: string[]): RescueResource[] => {
  return names.map((name, i) => {
    const p = generatePoint(baseLng, baseLat, 30)
    return {
      id: `res_${type}_${i}`,
      type,
      name,
      address: `${['浦东新区', '徐汇区', '长宁区', '静安区', '普陀区', '闵行区'][i % 6]}${['XX路', 'YY大道', 'ZZ街', 'AA路'][i % 4]}${100 + i * 37}号`,
      lng: p.lng,
      lat: p.lat,
      phone: `021-${String(50000000 + i * 12345).slice(0, 8)}`,
      contact: ['张主任', '李队长', '王站长', '赵经理', '陈主管', '刘工'][i % 6],
      capacity: type === 'hospital' ? `${100 + i * 50} 张床位 · 24h急诊`
        : type === 'fire' ? `${8 + i * 2} 辆消防车 · ${20 + i * 5}人`
          : `${5 + i} 个工位 · 专业维修`,
      response_time_avg: 5 + i * 3,
      available: i % 4 !== 3,
      rating: 4 + Math.random(),
      specialties: type === 'hospital' ? ['急诊', '烧伤科', '中毒救治', 'ICU']
        : type === 'fire' ? ['危化品处置', '交通事故救援', '水域救援']
          : ['重型车维修', '危化品车检修', '轮胎更换', '拖车服务'],
      status: i % 5 === 0 ? 'busy' : i % 7 === 0 ? 'offline' : 'idle',
    }
  })
}

const mockResources: Record<ResourceType, RescueResource[]> = {
  hospital: generateResources('hospital', hospitalNames),
  fire: generateResources('fire', fireStationNames),
  repair: generateResources('repair', repairNames),
}

const Rescue: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [selectedRequest, setSelectedRequest] = useState<RescueRequest | null>(null)
  const [resourceTab, setResourceTab] = useState<ResourceType>('hospital')
  const [statusFilter, setStatusFilter] = useState<string>()
  const [typeFilter, setTypeFilter] = useState<string>()
  const [severityFilter, setSeverityFilter] = useState<string>()
  const [searchKeyword, setSearchKeyword] = useState('')
  const [chatInput, setChatInput] = useState('')
  const [chatMessages, setChatMessages] = useState<RescueRequest['chat_messages']>([])

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const mapInstanceRef = useRef<any>(null)
  const [mapReady, setMapReady] = useState(false)
  const chatEndRef = useRef<HTMLDivElement>(null)

  const [stats, setStats] = useState({
    pending: 0,
    responding: 0,
    today: 0,
    total: mockRescueRequests.length,
  })

  useEffect(() => {
    setStats({
      pending: mockRescueRequests.filter(r => r.status === 'pending').length,
      responding: mockRescueRequests.filter(r => ['responding', 'arrived', 'processing'].includes(r.status)).length,
      today: mockRescueRequests.filter(r => dayjs(r.request_time).isSame(dayjs(), 'day')).length || 8,
      total: mockRescueRequests.length,
    })
  }, [])

  const initMap = async () => {
    if (!mapContainerRef.current) return
    try {
      const AMap: any = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: ['AMap.Scale', 'AMap.ToolBar', 'AMap.MarkerCluster'],
      })
      const map = new AMap.Map(mapContainerRef.current, {
        zoom: 10,
        center: [baseLng, baseLat],
        mapStyle: 'amap://styles/light',
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar())

      mockRescueRequests.forEach(req => {
        const isPending = req.status === 'pending'
        const isEmergency = req.severity === 'critical' || req.severity === 'severe'
        new AMap.Marker({
          position: [req.lng, req.lat],
          label: {
            content: `<div style="padding:4px 10px;background:${isPending ? '#ff4d4f' : isEmergency ? '#ff7a45' : '#faad14'};color:#fff;border-radius:16px;font-size:11px;font-weight:600;white-space:nowrap;box-shadow:0 2px 8px rgba(0,0,0,0.15)">
              🚨 ${rescueTypeMap[req.type].label}
            </div>`,
            direction: 'top',
            offset: new AMap.Pixel(0, -10),
          },
          offset: new AMap.Pixel(-16, -32),
          animation: isPending ? 'AMAP_ANIMATION_BOUNCE' : undefined,
          map,
        })
      })

      Object.values(mockResources).flat().forEach(res => {
        const rt = resourceTypeMap[res.type]
        new AMap.Marker({
          position: [res.lng, res.lat],
          label: {
            content: `<div style="padding:2px 8px;background:${res.available ? rt.markerColor : '#bfbfbf'};color:#fff;border-radius:4px;font-size:10px;white-space:nowrap;opacity:${res.available ? 1 : 0.6}">
              ${rt.icon} ${res.name.slice(0, 6)}
            </div>`,
            direction: 'top',
          },
          offset: new AMap.Pixel(-12, -24),
          map,
        })
      })

      mapInstanceRef.current = map
      setMapReady(true)
    } catch (e) {
    }
  }

  useEffect(() => {
    const timer = setTimeout(initMap, 300)
    return () => {
      clearTimeout(timer)
      if (mapInstanceRef.current) {
        mapInstanceRef.current.destroy()
        mapInstanceRef.current = null
      }
    }
  }, [])

  useEffect(() => {
    if (selectedRequest) {
      setChatMessages([...selectedRequest.chat_messages])
    }
  }, [selectedRequest])

  useEffect(() => {
    if (chatEndRef.current) {
      chatEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [chatMessages, selectedRequest])

  const handleSendChat = () => {
    if (!chatInput.trim() || !selectedRequest) return
    const newMsg = {
      id: `msg_${Date.now()}`,
      sender: '我(调度员)',
      role: 'dispatcher' as const,
      time: new Date().toISOString(),
      content: chatInput,
    }
    setChatMessages(prev => [...prev, newMsg])
    setChatInput('')
    message.success('消息已发送')
  }

  const filteredRequests = mockRescueRequests.filter(r => {
    if (statusFilter && r.status !== statusFilter) return false
    if (typeFilter && r.type !== typeFilter) return false
    if (severityFilter && r.severity !== severityFilter) return false
    if (searchKeyword && !(
      r.request_no.includes(searchKeyword) ||
      r.vehicle_plate.includes(searchKeyword) ||
      r.driver_name.includes(searchKeyword) ||
      r.location.includes(searchKeyword)
    )) return false
    return true
  }).sort((a, b) => {
    const order: Record<RescueStatus, number> = { pending: 0, responding: 1, arrived: 2, processing: 3, resolved: 4, closed: 5, cancelled: 6 }
    return order[a.status] - order[b.status]
  })

  const getResourcesWithDistance = (type: ResourceType, targetLng?: number, targetLat?: number) => {
    const resources = mockResources[type]
    if (!targetLng || !targetLat) {
      return resources.map(r => ({ ...r, distance_km: 3 + Math.random() * 25 }))
    }
    return resources
      .map(r => ({ ...r, distance_km: calcDistance(r.lng, r.lat, targetLng, targetLat) }))
      .sort((a, b) => (a.distance_km || 0) - (b.distance_km || 0))
  }

  const requestColumns: ProColumns<RescueRequest>[] = [
    {
      title: 'SOS编号',
      dataIndex: 'request_no',
      width: 150,
      render: (v: string, r) => (
        <Space>
          <Badge dot={r.status === 'pending'} color="#ff4d4f" offset={[-4, 4]}>
            <Text copyable strong style={{ fontSize: 12 }}>{v}</Text>
          </Badge>
        </Space>
      ),
    },
    {
      title: '车辆位置/司机',
      width: 220,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Space size={4}>
            <CarOutlined style={{ color: '#1677ff' }} />
            <Tag color="blue" style={{ margin: 0, fontSize: 12 }}>{r.vehicle_plate}</Tag>
            <Space size={2}>
              <UserOutlined style={{ color: '#fa8c16', fontSize: 11 }} />
              <Text style={{ fontSize: 12 }}>{r.driver_name}</Text>
            </Space>
          </Space>
          <Space size={4} style={{ width: '100%' }}>
            <EnvironmentOutlined style={{ color: '#ff4d4f', fontSize: 11 }} />
            <Text ellipsis style={{ fontSize: 11, maxWidth: 160 }} title={r.location}>
              {r.location}
            </Text>
          </Space>
        </Space>
      ),
    },
    {
      title: '类型/级别',
      width: 160,
      render: (_, r) => {
        const t = rescueTypeMap[r.type]
        const s = severityMap[r.severity]
        return (
          <Space direction="vertical" size={2}>
            <Space size={4}>
              <Tag color={t.color} style={{ margin: 0, fontSize: 12 }}>{t.icon} {t.label}</Tag>
            </Space>
            <Tag color={s.color} style={{ fontSize: 11, padding: '0 6px' }}>
              {s.label}
            </Tag>
          </Space>
        )
      },
    },
    {
      title: '请求时间',
      dataIndex: 'request_time',
      width: 140,
      render: (v: string) => (
        <Space direction="vertical" size={0}>
          <Text style={{ fontSize: 12 }}>{formatDateTime(v, 'MM-DD HH:mm:ss')}</Text>
          <Text type="secondary" style={{ fontSize: 10 }}>
            {dayjs().to(dayjs(v))}
          </Text>
        </Space>
      ),
    },
    {
      title: '处理状态',
      width: 120,
      render: (_, r) => {
        const st = rescueStatusMap[r.status]
        return (
          <Space direction="vertical" size={4}>
            <Tag color={st.color} style={{ fontSize: 12 }}>{st.label}</Tag>
            {r.eta_minutes && (
              <Text type="secondary" style={{ fontSize: 11 }}>
                🕐 预计 {r.eta_minutes} 分钟到达
              </Text>
            )}
          </Space>
        )
      },
    },
    {
      title: '操作',
      width: 160,
      fixed: 'right',
      render: (_, record) => (
        <Space size={4}>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => setSelectedRequest(record)}
          >
            详情
          </Button>
          {(record.status === 'pending' || record.status === 'responding') && (
            <Button
              type="link"
              size="small"
              type="primary"
              danger
              icon={<SendOutlined />}
            >
              派单
            </Button>
          )}
          {['processing', 'arrived'].includes(record.status) && (
            <Button
              type="link"
              size="small"
              icon={<CheckCircleOutlined />}
            >
              结案
            </Button>
          )}
          {record.status === 'pending' && (
            <Button
              type="link"
              size="small"
              icon={<PhoneOutlined />}
            >
              拨号
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const resourceColumns: ProColumns<RescueResource>[] = [
    {
      title: '资源名称',
      dataIndex: 'name',
      width: 260,
      render: (v: string, r) => {
        const rt = resourceTypeMap[r.type]
        return (
          <Space>
            <div style={{
              width: 32, height: 32, borderRadius: 8,
              background: r.available ? `${rt.markerColor}15` : '#f5f5f5',
              color: r.available ? rt.markerColor : '#bfbfbf',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 16,
            }}>
              {rt.icon}
            </div>
            <div>
              <Text strong style={{ fontSize: 13 }}>{v}</Text>
              <div>
                {r.status === 'busy' && <Tag color="orange" style={{ fontSize: 11 }}>忙碌中</Tag>}
                {r.status === 'idle' && <Tag color="green" style={{ fontSize: 11 }}>待命</Tag>}
                {r.status === 'offline' && <Tag color="default" style={{ fontSize: 11 }}>离线</Tag>}
                <Text type="secondary" style={{ fontSize: 11 }}> 联系: {r.contact}</Text>
              </div>
            </div>
          </Space>
        )
      },
    },
    {
      title: '距离',
      dataIndex: 'distance_km',
      width: 100,
      sorter: (a, b) => (a.distance_km || 0) - (b.distance_km || 0),
      render: (v?: number) => (
        <Space direction="vertical" size={0}>
          <Text strong style={{
            color: (v || 0) < 5 ? '#52c41a' : (v || 0) < 15 ? '#1677ff' : (v || 0) < 30 ? '#faad14' : '#ff4d4f',
            fontSize: 14,
          }}>
            {(v || 0).toFixed(1)} km
          </Text>
          <Text type="secondary" style={{ fontSize: 10 }}>
            车程约 {Math.max(3, Math.round((v || 0) * 2))} 分钟
          </Text>
        </Space>
      ),
    },
    {
      title: '地址/联系方式',
      width: 280,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Space size={4}>
            <HomeOutlined style={{ color: '#8c8c8c', fontSize: 11 }} />
            <Text style={{ fontSize: 12 }}>{r.address}</Text>
          </Space>
          <Space size={4}>
            <PhoneOutlined style={{ color: '#1677ff', fontSize: 11 }} />
            <Text type="secondary" style={{ fontSize: 11 }} copyable>{r.phone}</Text>
          </Space>
        </Space>
      ),
    },
    {
      title: '响应能力',
      width: 220,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Text style={{ fontSize: 12 }}>{r.capacity}</Text>
          <Space size={4} wrap>
            <Text type="secondary" style={{ fontSize: 11 }}>
              <ClockOutlined /> 平均响应 {r.response_time_avg} 分钟
            </Text>
            {r.rating && (
              <Text type="secondary" style={{ fontSize: 11 }}>
                ⭐ {r.rating.toFixed(1)}
              </Text>
            )}
          </Space>
          {r.specialties && (
            <Space size={2} wrap>
              {r.specialties.slice(0, 3).map(s => (
                <Tag key={s} color="blue" style={{ fontSize: 10, padding: '0 4px', margin: 0 }}>{s}</Tag>
              ))}
            </Space>
          )}
        </Space>
      ),
    },
    {
      title: '操作',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />}>查看</Button>
          {record.available && (
            <Button type="link" size="small" type="primary" icon={<SendOutlined />}>调度</Button>
          )}
          <Button type="link" size="small" icon={<PhoneOutlined />}></Button>
        </Space>
      ),
    },
  ]

  const roleAvatarMap = {
    dispatcher: { bg: '#1677ff', label: '调度' },
    driver: { bg: '#fa8c16', label: '司机' },
    rescue: { bg: '#52c41a', label: '救援' },
    supervisor: { bg: '#722ed1', label: '主管' },
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Alert
        type="warning"
        showIcon
        icon={<ThunderboltOutlined />}
        message={
          <Space wrap>
            <Space>
              <Badge count={stats.pending} showZero style={{ backgroundColor: '#ff4d4f', boxShadow: '0 0 12px rgba(255,77,79,0.5)' }}>
                <Tag color="red" style={{ fontSize: 13, padding: '4px 12px', margin: 0 }}>
                  <strong>🔴 待响应 {stats.pending} 起</strong>
                </Tag>
              </Badge>
              <Tag color="processing" style={{ fontSize: 13, padding: '4px 12px' }}>处置中 {stats.responding} 起</Tag>
            </Space>
            <Text type="secondary">今日救援 {stats.today} 起 · 累计 {stats.total} 起</Text>
            <Text type="warning" strong style={{ fontSize: 12 }}>
              <ClockCircleOutlined /> 请及时处理待响应的SOS请求！
            </Text>
          </Space>
        }
        style={{ borderRadius: 12 }}
        banner
      />

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
            <Statistic
              title="待响应"
              value={stats.pending}
              valueStyle={{ color: '#ff4d4f', fontSize: 30 }}
              prefix={<StopOutlined style={{ fontSize: 22 }} />}
              suffix={<Text type="secondary" style={{ fontSize: 13 }}>起</Text>}
            />
            <div style={{ marginTop: 4 }}>
              <Progress percent={Math.min(100, stats.pending * 20)} size="small" strokeColor="#ff4d4f" showInfo={false} />
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
            <Statistic
              title="处置中"
              value={stats.responding}
              valueStyle={{ color: '#fa8c16', fontSize: 30 }}
              prefix={<CaretRightOutlined style={{ fontSize: 22 }} />}
              suffix={<Text type="secondary" style={{ fontSize: 13 }}>起</Text>}
            />
            <div style={{ marginTop: 4 }}>
              <Progress percent={Math.min(100, stats.responding * 15)} size="small" strokeColor="#fa8c16" showInfo={false} />
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
            <Statistic
              title="今日救援"
              value={stats.today}
              valueStyle={{ color: '#1677ff', fontSize: 30 }}
              prefix={<RiseOutlined style={{ fontSize: 22 }} />}
              suffix={<Text type="secondary" style={{ fontSize: 13 }}>起</Text>}
            />
            <div style={{ marginTop: 4 }}>
              <Progress percent={Math.min(100, stats.today * 8)} size="small" strokeColor="#1677ff" showInfo={false} />
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
            <Statistic
              title="累计救援"
              value={stats.total}
              valueStyle={{ color: '#52c41a', fontSize: 30 }}
              prefix={<HeartOutlined style={{ fontSize: 22 }} />}
              suffix={<Text type="secondary" style={{ fontSize: 13 }}>起</Text>}
            />
            <div style={{ marginTop: 4 }}>
              <Progress percent={100} size="small" strokeColor="#52c41a" showInfo={false} />
            </div>
          </Card>
        </Col>
      </Row>

      <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
        <div style={{ padding: '12px 16px 0' }}>
          <Row gutter={12} align="middle">
            <Col xs={24} md={14}>
              <Space>
                <MapOutlined style={{ color: '#ff4d4f', fontSize: 18 }} />
                <Text strong style={{ fontSize: 15 }}>救援态势地图</Text>
                <Badge status="processing" text="实时刷新" />
              </Space>
            </Col>
            <Col xs={24} md={10}>
              <Space wrap style={{ justifyContent: 'flex-end', width: '100%' }}>
                <Space size={4}>
                  <span style={{ display: 'inline-block', width: 14, height: 14, borderRadius: '50%', background: '#ff4d4f' }} />
                  <Text type="secondary" style={{ fontSize: 12 }}>SOS请求</Text>
                </Space>
                <Space size={4}>
                  <span style={{ display: 'inline-block', width: 14, height: 14, borderRadius: '50%', background: '#f5222d' }} />
                  <Text type="secondary" style={{ fontSize: 12 }}>医院</Text>
                </Space>
                <Space size={4}>
                  <span style={{ display: 'inline-block', width: 14, height: 14, borderRadius: '50%', background: '#ff4d4f' }} />
                  <Text type="secondary" style={{ fontSize: 12 }}>消防站</Text>
                </Space>
                <Space size={4}>
                  <span style={{ display: 'inline-block', width: 14, height: 14, borderRadius: '50%', background: '#fa8c16' }} />
                  <Text type="secondary" style={{ fontSize: 12 }}>维修站</Text>
                </Space>
                <Button icon={<ReloadOutlined />} size="small" onClick={() => {
                  setMapReady(false)
                  setTimeout(() => {
                    if (mapInstanceRef.current) { mapInstanceRef.current.destroy() }
                    initMap()
                  }, 200)
                }}>刷新</Button>
              </Space>
            </Col>
          </Row>
        </div>
        <div
          ref={mapContainerRef}
          style={{
            height: 360,
            background: '#f0f5ff',
            display: mapReady ? 'block' : 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginTop: 12,
          }}
        >
          {!mapReady && (
            <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
              <MapOutlined style={{ fontSize: 56, opacity: 0.3 }} />
              <div style={{ marginTop: 12 }}>
                <Text type="secondary">救援态势地图加载中...</Text>
                <div style={{ marginTop: 8, fontSize: 12 }}>
                  <Space direction="vertical" size={4}>
                    <Text type="secondary">当前区域: 上海市 · {filteredRequests.length} 起SOS请求</Text>
                    <Text type="secondary">
                      周边资源: {Object.values(mockResources).flat().length} 个
                      (医院{mockResources.hospital.length} · 消防{mockResources.fire.length} · 维修{mockResources.repair.length})
                    </Text>
                  </Space>
                </div>
              </div>
            </div>
          )}
        </div>
      </ProCard>

      <Tabs
        defaultActiveKey="requests"
        size="large"
        items={[
          {
            key: 'requests',
            label: (
              <Space>
                <ThunderboltOutlined style={{ color: '#ff4d4f' }} />
                <Text strong>SOS救援请求</Text>
                <Badge count={filteredRequests.length} showZero style={{ backgroundColor: '#ff4d4f' }} />
              </Space>
            ),
            children: (
              <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
                <ProTable<RescueRequest>
                  rowKey="id"
                  loading={loading}
                  columns={requestColumns}
                  search={{
                    labelWidth: 'auto',
                    defaultCollapsed: false,
                    span: { xs: 24, sm: 12, md: 8, lg: 6 },
                  }}
                  pagination={{
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: t => `共 ${t} 条SOS请求`,
                    defaultPageSize: 10,
                  }}
                  scroll={{ x: 1100 }}
                  headerTitle={
                    <Space>
                      <ThunderboltOutlined style={{ color: '#ff4d4f' }} />
                      <Text strong style={{ fontSize: 15 }}>SOS请求列表</Text>
                    </Space>
                  }
                  extra={
                    <Space wrap>
                      <Input
                        allowClear
                        prefix={<SearchOutlined />}
                        placeholder="搜索编号/车牌/司机/位置"
                        style={{ width: 220 }}
                        value={searchKeyword}
                        onChange={e => setSearchKeyword(e.target.value)}
                      />
                      <Select
                        allowClear
                        placeholder="类型"
                        style={{ width: 130 }}
                        value={typeFilter}
                        onChange={setTypeFilter}
                      >
                        {Object.entries(rescueTypeMap).map(([k, v]) => (
                          <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
                        ))}
                      </Select>
                      <Select
                        allowClear
                        placeholder="状态"
                        style={{ width: 130 }}
                        value={statusFilter}
                        onChange={setStatusFilter}
                      >
                        {Object.entries(rescueStatusMap).map(([k, v]) => (
                          <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
                        ))}
                      </Select>
                      <Select
                        allowClear
                        placeholder="级别"
                        style={{ width: 120 }}
                        value={severityFilter}
                        onChange={setSeverityFilter}
                      >
                        {Object.entries(severityMap).map(([k, v]) => (
                          <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
                        ))}
                      </Select>
                      <Button icon={<ReloadOutlined />}>刷新</Button>
                      <Button icon={<ExportOutlined />}>导出</Button>
                    </Space>
                  }
                  toolBarRender={() => []}
                  columnsState={{ persistenceKey: 'rescue-requests', persistenceType: 'localStorage' }}
                  request={async (params): Promise<RequestData<RescueRequest>> => {
                    await new Promise(r => setTimeout(r, 300))
                    const { current = 1, pageSize = 10 } = params
                    const total = filteredRequests.length
                    const pageList = filteredRequests.slice((current - 1) * pageSize, current * pageSize)
                    return { data: pageList, success: true, total }
                  }}
                  rowClassName={r => r.status === 'pending' ? '!bg-red-50/60' : r.severity === 'critical' ? '!bg-orange-50/60' : ''}
                />
              </ProCard>
            ),
          },
          {
            key: 'resources',
            label: (
              <Space>
                <TeamOutlined style={{ color: '#1677ff' }} />
                <Text strong>救援资源管理</Text>
              </Space>
            ),
            children: (
              <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
                <Tabs
                  activeKey={resourceTab}
                  onChange={k => setResourceTab(k as ResourceType)}
                  items={Object.entries(resourceTypeMap).map(([k, v]) => ({
                    key: k,
                    label: (
                      <Space>
                        <span style={{
                          width: 20, height: 20, borderRadius: 6,
                          background: `${v.markerColor}15`, color: v.markerColor,
                          display: 'inline-flex', alignItems: 'center', justifyContent: 'center', fontSize: 11,
                        }}>
                          {v.icon}
                        </span>
                        {v.label}
                        <Tag color={v.color} style={{ margin: 0 }}>
                          {mockResources[k as ResourceType].filter(r => r.available).length}/
                          {mockResources[k as ResourceType].length}
                        </Tag>
                      </Space>
                    ),
                  }))}
                />

                <div style={{ marginTop: 12 }}>
                  {selectedRequest && (
                    <Alert
                      type="info"
                      showIcon
                      message={
                        <Space>
                          <EnvironmentOutlined />
                          <Text>基于选中的SOS请求 <Text strong>{selectedRequest.request_no}</Text> ({selectedRequest.location}) 按距离排序显示</Text>
                        </Space>
                      }
                      style={{ marginBottom: 12, borderRadius: 8 }}
                      closable
                      onClose={() => setSelectedRequest(null)}
                    />
                  )}

                  <ProTable<RescueResource>
                    rowKey="id"
                    columns={resourceColumns}
                    search={false}
                    toolBarRender={() => [
                      <Button key="add" icon={<PlusOutlined />} type="primary">添加资源</Button>,
                      <Button key="export" icon={<ExportOutlined />}>导出</Button>,
                    ]}
                    pagination={{ pageSize: 6, showSizeChanger: false, showTotal: t => `共 ${t} 个${resourceTypeMap[resourceTab].label}资源` }}
                    scroll={{ x: 1100 }}
                    request={async (params): Promise<RequestData<RescueResource>> => {
                      await new Promise(r => setTimeout(r, 200))
                      const list = getResourcesWithDistance(
                        resourceTab,
                        selectedRequest?.lng,
                        selectedRequest?.lat
                      )
                      const { current = 1, pageSize = 6 } = params
                      const total = list.length
                      const pageList = list.slice((current - 1) * pageSize, current * pageSize)
                      return { data: pageList, success: true, total }
                    }}
                    rowClassName={r => !r.available ? '!bg-gray-50' : ''}
                  />
                </div>
              </ProCard>
            ),
          },
        ]}
      />

      <Drawer
        title={
          selectedRequest ? (
            <Space>
              <div style={{
                width: 32, height: 32, borderRadius: '50%',
                background: severityMap[selectedRequest.severity].color === 'red' ? '#fff1f0'
                  : severityMap[selectedRequest.severity].color === 'orange' ? '#fff7e6'
                    : severityMap[selectedRequest.severity].color === 'blue' ? '#e6f4ff'
                      : '#f6ffed',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: severityMap[selectedRequest.severity].color === 'red' ? '#ff4d4f'
                  : severityMap[selectedRequest.severity].color === 'orange' ? '#fa8c16'
                    : severityMap[selectedRequest.severity].color === 'blue' ? '#1677ff'
                      : '#52c41a',
                fontSize: 16,
              }}>
                {rescueTypeMap[selectedRequest.type].icon}
              </div>
              <div>
                <div>
                  <Text strong style={{ fontSize: 16 }}>SOS救援详情 - {selectedRequest.request_no}</Text>
                  <Tag color={rescueStatusMap[selectedRequest.status].color} style={{ marginLeft: 8 }}>
                    {rescueStatusMap[selectedRequest.status].label}
                  </Tag>
                  <Tag color={severityMap[selectedRequest.severity].color} style={{ marginLeft: 4 }}>
                    {severityMap[selectedRequest.severity].label}
                  </Tag>
                </div>
                <div style={{ marginTop: 2 }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {rescueTypeMap[selectedRequest.type].label} · 发起于 {dayjs(selectedRequest.request_time).fromNow()}
                  </Text>
                </div>
              </div>
            </Space>
          ) : null
        }
        open={!!selectedRequest}
        onClose={() => setSelectedRequest(null)}
        width={680}
        extra={
          selectedRequest && (
            <Space wrap>
              <Button icon={<PhoneOutlined />} type="primary" danger>
                联系司机 {selectedRequest.driver_phone}
              </Button>
              {(selectedRequest.status === 'pending' || selectedRequest.status === 'responding') && (
                <Button type="primary" icon={<SendOutlined />}>一键派单</Button>
              )}
              {['processing', 'arrived'].includes(selectedRequest.status) && (
                <Button type="primary" icon={<CheckCircleOutlined />}>完成处置</Button>
              )}
              <Button icon={<ExportOutlined />}>导出报告</Button>
            </Space>
          )
        }
      >
        {selectedRequest && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {selectedRequest.severity === 'critical' && (
              <Alert
                type="error"
                showIcon
                icon={<WarningOutlined />}
                message="🚨 紧急高危事件"
                description={
                  <Space direction="vertical" size={0} style={{ width: '100%' }}>
                    <Text type="danger" strong>{selectedRequest.description}</Text>
                    {selectedRequest.injuries ? <Text type="danger">受伤人数: {selectedRequest.injuries} 人</Text> : null}
                    {selectedRequest.danger_goods ? <Text type="danger">⚠️ 涉及危险货物: {selectedRequest.danger_goods} (类{selectedRequest.danger_level})</Text> : null}
                  </Space>
                }
                style={{ borderRadius: 8 }}
                closable
              />
            )}
            {(selectedRequest.severity === 'severe' || selectedRequest.status === 'pending') && (
              <Alert
                type="warning"
                showIcon
                message={selectedRequest.status === 'pending' ? '⏰ 事件尚未响应，请立即处理' : '⚠️ 严重等级事件，请持续关注'}
                style={{ borderRadius: 8 }}
                closable
              />
            )}

            <Tabs defaultActiveKey="detail" size="small">
              <TabPane tab={<Space><FileTextOutlined />请求详情</Space>} key="detail">
                <Card size="small" style={{ borderRadius: 8 }}>
                  <Descriptions column={2} size="small" bordered>
                    <Descriptions.Item label="SOS编号" span={2}>
                      <Text copyable strong>{selectedRequest.request_no}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="事件类型">
                      <Tag color={rescueTypeMap[selectedRequest.type].color}>
                        {rescueTypeMap[selectedRequest.type].icon} {rescueTypeMap[selectedRequest.type].label}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="严重等级">
                      <Tag color={severityMap[selectedRequest.severity].color}>
                        {severityMap[selectedRequest.severity].label}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="车辆信息">
                      <Space direction="vertical" size={0}>
                        <Space>
                          <Tag color="blue">{selectedRequest.vehicle_plate}</Tag>
                          {selectedRequest.danger_goods && (
                            <Tag color="red">危化品: {selectedRequest.danger_goods}</Tag>
                          )}
                        </Space>
                        {selectedRequest.waybill_no && (
                          <Text type="secondary" style={{ fontSize: 11 }}>
                            关联运单: {selectedRequest.waybill_no}
                          </Text>
                        )}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="司机/押运">
                      <Space direction="vertical" size={0}>
                        <Text><UserOutlined /> {selectedRequest.driver_name}</Text>
                        <Text type="secondary" style={{ fontSize: 11 }} copyable>
                          <PhoneOutlined /> {selectedRequest.driver_phone}
                        </Text>
                        {selectedRequest.escort_name && (
                          <Text type="secondary" style={{ fontSize: 11 }}>押运: {selectedRequest.escort_name}</Text>
                        )}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="发生位置" span={2}>
                      <Paragraph copyable style={{ margin: 0 }}>
                        <EnvironmentOutlined style={{ color: '#ff4d4f' }} /> {selectedRequest.location}
                        <Text type="secondary"> ({selectedRequest.lng.toFixed(5)}, {selectedRequest.lat.toFixed(5)})</Text>
                      </Paragraph>
                    </Descriptions.Item>
                    <Descriptions.Item label="请求时间">{formatDateTime(selectedRequest.request_time)}</Descriptions.Item>
                    <Descriptions.Item label="受伤人数">
                      <Text type={selectedRequest.injuries ? 'danger' : undefined}>
                        {selectedRequest.injuries ? `${selectedRequest.injuries} 人` : '无'}
                      </Text>
                    </Descriptions.Item>
                    {selectedRequest.response_time && (
                      <Descriptions.Item label="响应时间">{formatDateTime(selectedRequest.response_time)}</Descriptions.Item>
                    )}
                    {selectedRequest.arrival_time && (
                      <Descriptions.Item label="到达时间">{formatDateTime(selectedRequest.arrival_time)}</Descriptions.Item>
                    )}
                    {selectedRequest.resolved_time && (
                      <Descriptions.Item label="解决时间">{formatDateTime(selectedRequest.resolved_time)}</Descriptions.Item>
                    )}
                    {selectedRequest.assigned_resources && (
                      <Descriptions.Item label="已派资源" span={2}>
                        <Space wrap>
                          {selectedRequest.assigned_resources.map(r => (
                            <Tag key={r} color="blue">{r}</Tag>
                          ))}
                        </Space>
                      </Descriptions.Item>
                    )}
                    <Descriptions.Item label="紧急联系人" span={2}>
                      <Space>
                        <UserOutlined /> {selectedRequest.contact_name}
                        <PhoneOutlined style={{ color: '#1677ff' }} />
                        <Text copyable>{selectedRequest.contact_phone}</Text>
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="事件描述" span={2}>
                      <Paragraph style={{ margin: 0, lineHeight: 1.8 }}>
                        {selectedRequest.description}
                      </Paragraph>
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </TabPane>

              <TabPane tab={<Space><EditOutlined />处置时间轴</Space>} key="timeline">
                <Card size="small" style={{ borderRadius: 8 }}>
                  <div style={{ padding: '4px 0' }}>
                    <Steps
                      size="small"
                      current={rescueStatusMap[selectedRequest.status].step}
                      status={selectedRequest.status === 'cancelled' ? 'error' : undefined}
                      items={[
                        { title: 'SOS发起', description: dayjs(selectedRequest.request_time).format('HH:mm:ss') },
                        { title: '调度响应', description: selectedRequest.response_time ? dayjs(selectedRequest.response_time).format('HH:mm:ss') : '-' },
                        { title: '救援到达', description: selectedRequest.arrival_time ? dayjs(selectedRequest.arrival_time).format('HH:mm:ss') : '-' },
                        { title: '现场处置', description: '-' },
                        { title: '处置完成', description: selectedRequest.resolved_time ? dayjs(selectedRequest.resolved_time).format('HH:mm:ss') : '-' },
                      ]}
                      style={{ marginBottom: 20 }}
                    />
                    <Divider style={{ margin: '4px 0 16px' }} />
                    <Timeline
                      mode="left"
                      items={selectedRequest.timeline.map((item, idx) => ({
                        color: idx === selectedRequest.timeline.length - 1
                          ? (selectedRequest.status === 'pending' ? 'red' : 'blue')
                          : 'gray',
                        label: <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(item.time, 'MM-DD HH:mm:ss')}</Text>,
                        children: (
                          <div>
                            <Space>
                              <Text strong style={{ fontSize: 13 }}>{item.status}</Text>
                              <Tag color="default" style={{ fontSize: 11, padding: '0 6px' }}>{item.operator}</Tag>
                            </Space>
                            {item.note && (
                              <Paragraph style={{ fontSize: 12, marginTop: 4, marginBottom: 0 }} type="secondary">
                                {item.note}
                              </Paragraph>
                            )}
                          </div>
                        ),
                      }))}
                    />
                  </div>
                </Card>
              </TabPane>

              <TabPane tab={<Space><PhoneOutlined />对话记录 ({chatMessages.length})</Space>} key="chat">
                <Card
                  size="small"
                  style={{ borderRadius: 8 }}
                  styles={{ body: { padding: 0 } }}
                  title={<Space>实时通讯</Space>}
                  extra={
                    <Space>
                      <Badge status="success" text="在线" />
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        已通知: 司机、押运员、附近救援资源
                      </Text>
                    </Space>
                  }
                >
                  <div
                    style={{
                      height: 380,
                      overflowY: 'auto',
                      padding: 16,
                      background: '#fafafa',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 12,
                    }}
                  >
                    {chatMessages.map(item => {
                      const isMe = item.role === 'dispatcher' && item.sender.includes('我')
                      const ra = roleAvatarMap[item.role] || { bg: '#8c8c8c', label: '未知' }
                      return (
                        <div
                          key={item.id}
                          style={{
                            display: 'flex',
                            flexDirection: isMe ? 'row-reverse' : 'row',
                            gap: 8,
                            maxWidth: '100%',
                          }}
                        >
                          <Avatar
                            size={32}
                            style={{
                              background: isMe ? '#1677ff' : ra.bg,
                              flexShrink: 0,
                              fontSize: 11,
                            }}
                          >
                            {isMe ? '我' : ra.label}
                          </Avatar>
                          <div style={{
                            maxWidth: '75%',
                            display: 'flex',
                            flexDirection: isMe ? 'row-reverse' : 'row',
                            alignItems: 'flex-start',
                            gap: 8,
                          }}>
                            <div style={{ minWidth: 0 }}>
                              <div style={{
                                fontSize: 11,
                                color: '#8c8c8c',
                                marginBottom: 4,
                                textAlign: isMe ? 'right' : 'left',
                              }}>
                                <Space>
                                  <Text>{item.sender}</Text>
                                  <Text>{dayjs(item.time).format('HH:mm:ss')}</Text>
                                </Space>
                              </div>
                              <div
                                style={{
                                  padding: '10px 14px',
                                  borderRadius: isMe ? '14px 14px 4px 14px' : '14px 14px 14px 4px',
                                  background: isMe ? '#e6f4ff' : '#fff',
                                  color: isMe ? '#0958d9' : '#262626',
                                  border: isMe ? 'none' : '1px solid #f0f0f0',
                                  fontSize: 13,
                                  lineHeight: 1.6,
                                  wordBreak: 'break-word',
                                  boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                                }}
                              >
                                {item.content}
                              </div>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                    <div ref={chatEndRef} />
                  </div>
                  <Divider style={{ margin: 0 }} />
                  <div style={{ padding: 12, display: 'flex', gap: 8 }}>
                    <Input
                      value={chatInput}
                      onChange={e => setChatInput(e.target.value)}
                      onKeyDown={e => e.key === 'Enter' && handleSendChat()}
                      placeholder="输入消息，回车发送... (通知: 司机/救援队/主管)"
                      suffix={
                        <Space>
                          <Tooltip title="发送图片"><Button type="text" icon={<PlusOutlined />} size="small" /></Tooltip>
                          <Tooltip title="紧急广播"><Button type="text" danger icon={<ThunderboltOutlined />} size="small" /></Tooltip>
                        </Space>
                      }
                    />
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      onClick={handleSendChat}
                      disabled={!chatInput.trim()}
                    >
                      发送
                    </Button>
                  </div>
                </Card>
              </TabPane>
            </Tabs>
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default Rescue
