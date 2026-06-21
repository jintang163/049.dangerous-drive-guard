import React, { useEffect, useState, useRef, useCallback } from 'react'
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
  Tabs,
  Table,
  DatePicker,
  Progress,
  Statistic,
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
  BellOutlined,
  PhoneOutlined,
  VideoCameraOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ScheduleOutlined,
  SoundOutlined,
  DashboardOutlined,
  HistoryOutlined,
} from '@ant-design/icons'
import {
  ProCard,
  ProFormText,
  ProFormSelect,
  ProFormTextArea,
  ModalForm,
  ProTable,
} from '@ant-design/pro-components'
import type { ProColumns } from '@ant-design/pro-components'
import api, { escortApi } from '@/services/api'
import WebSocketManager from '@/services/ws'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import AMapLoader from '@amap/amap-jsapi-loader'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TextArea } = Input
const { Dragger } = Upload
const { RangePicker } = DatePicker

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
  departure_check: { label: '出发检查', color: 'blue', icon: <CheckCircleOutlined />, dot: 'blue' },
  waypoint: { label: '途经点', color: 'cyan', icon: <FlagOutlined />, dot: 'cyan' },
  abnormal_stop: { label: '异常停靠', color: 'orange', icon: <WarningOutlined />, dot: 'orange' },
  rest: { label: '休息', color: 'green', icon: <CoffeeOutlined />, dot: 'green' },
  loading_unloading: { label: '装卸货', color: 'purple', icon: <LoadingOutlined />, dot: 'purple' },
  sign_receipt: { label: '签收', color: 'green', icon: <FileTextOutlined />, dot: 'green' },
  emergency: { label: '突发事件', color: 'red', icon: <ThunderboltOutlined />, dot: 'red' },
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
    'departure_check', 'waypoint', 'rest', 'waypoint',
    'loading_unloading', 'abnormal_stop', 'waypoint', 'sign_receipt',
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
      photos: Math.random() > 0.3 ? [`https://picsum.photos/seed/escort${i}${waybill.id}/400/300`] : [],
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

const mockPollingVehicles = [
  {
    vehicle_id: 1,
    vehicle_plate: '沪A12345',
    driver_name: '张建国',
    danger_goods: '汽油',
    speed: 62,
    status: 'driving',
    last_frame_time: dayjs().toISOString(),
    driver_status: '正常',
    latitude: 31.2304,
    longitude: 121.4737,
  },
  {
    vehicle_id: 2,
    vehicle_plate: '京B67890',
    driver_name: '李明辉',
    danger_goods: '液化石油气',
    speed: 0,
    status: 'resting',
    last_frame_time: dayjs().subtract(10, 'second').toISOString(),
    driver_status: '休息',
    latitude: 39.9042,
    longitude: 116.4074,
  },
  {
    vehicle_id: 3,
    vehicle_plate: '粤C11111',
    driver_name: '王志强',
    danger_goods: '硫酸',
    speed: 58,
    status: 'driving',
    last_frame_time: dayjs().toISOString(),
    driver_status: '正常',
    latitude: 23.1291,
    longitude: 113.3245,
  },
  {
    vehicle_id: 4,
    vehicle_plate: '粤D22222',
    driver_name: '刘文华',
    danger_goods: '烟花爆竹',
    speed: 45,
    status: 'driving',
    last_frame_time: dayjs().subtract(5, 'second').toISOString(),
    driver_status: '正常',
    latitude: 22.5333,
    longitude: 113.9304,
  },
]

const Escort: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [selectedWaybill, setSelectedWaybill] = useState<EscortWaybill | null>(mockEscortWaybills[0])
  const [events, setEvents] = useState<EscortEvent[]>([])
  const [eventDetailDrawer, setEventDetailDrawer] = useState<EscortEvent | null>(null)
  const [addEventModalVisible, setAddEventModalVisible] = useState(false)
  const [addEventForm] = Form.useForm()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [searchKeyword, setSearchKeyword] = useState('')

  const [activeTab, setActiveTab] = useState('tasks')

  const [sosAlerts, setSosAlerts] = useState<any[]>([])
  const [sosAlertModalVisible, setSosAlertModalVisible] = useState(false)
  const [currentSosAlert, setCurrentSosAlert] = useState<any>(null)

  const [pollingVehicles, setPollingVehicles] = useState<any[]>(mockPollingVehicles)
  const [pollingActive, setPollingActive] = useState(true)
  const [pollingInterval, setPollingInterval] = useState(30)
  const [selectedPollingVehicle, setSelectedPollingVehicle] = useState<any>(null)
  const pollingTimerRef = useRef<NodeJS.Timeout | null>(null)

  const [intercomTarget, setIntercomTarget] = useState<any>(null)
  const [intercomModalVisible, setIntercomModalVisible] = useState(false)
  const [intercomText, setIntercomText] = useState('')
  const [intercomLogs, setIntercomLogs] = useState<any[]>([])

  const [trackPlaybackVisible, setTrackPlaybackVisible] = useState(false)
  const [trackPlaybackData, setTrackPlaybackData] = useState<any[]>([])
  const [playbackTarget, setPlaybackTarget] = useState<any>(null)
  const [playbackPlaying, setPlaybackPlaying] = useState(false)
  const [playbackIndex, setPlaybackIndex] = useState(0)
  const playbackTimerRef = useRef<NodeJS.Timeout | null>(null)
  const [playbackMapReady, setPlaybackMapReady] = useState(false)

  const [videoRecords, setVideoRecords] = useState<any[]>([])
  const [videoRecordLoading, setVideoRecordLoading] = useState(false)

  const eventMapContainerRef = useRef<HTMLDivElement>(null)
  const eventMapInstanceRef = useRef<any>(null)
  const [eventMapReady, setEventMapReady] = useState(false)

  const playbackMapContainerRef = useRef<HTMLDivElement>(null)
  const playbackMapInstanceRef = useRef<any>(null)
  const playbackMarkerRef = useRef<any>(null)
  const playbackPolylineRef = useRef<any>(null)

  const [statistics, setStatistics] = useState({
    total_shifts: 12,
    active_shifts: 5,
    pending_sos: 2,
    total_videos: 156,
    today_intercoms: 23,
    polling_vehicles: 8,
  })

  useEffect(() => {
    if (selectedWaybill) {
      setLoading(true)
      setTimeout(() => {
        setEvents(generateMockEvents(selectedWaybill))
        setLoading(false)
      }, 300)
    }
  }, [selectedWaybill])

  useEffect(() => {
    const ws = WebSocketManager.getInstance()
    ws.connect()

    const sosUnsubscribe = ws.on('sos_alert', (data) => {
      setCurrentSosAlert(data)
      setSosAlerts(prev => [data, ...prev].slice(0, 50))
      setSosAlertModalVisible(true)
      try {
        const audio = new Audio('/sos-alarm.mp3')
        audio.volume = 0.8
        audio.play().catch(() => {})
      } catch (e) {}
    })

    const pollingUnsubscribe = ws.on('escort_polling', (data) => {
      setPollingVehicles(prev => prev.map(v =>
        v.vehicle_id === data.vehicle_id ? { ...v, ...data } : v
      ))
    })

    return () => {
      sosUnsubscribe()
      pollingUnsubscribe()
    }
  }, [])

  useEffect(() => {
    if (pollingActive) {
      pollingTimerRef.current = setInterval(() => {
        setPollingVehicles(prev => prev.map(v => ({
          ...v,
          last_frame_time: new Date().toISOString(),
          speed: v.status === 'driving' ? Math.floor(Math.random() * 30) + 40 : 0,
        })))
      }, pollingInterval * 1000)
    }
    return () => {
      if (pollingTimerRef.current) {
        clearInterval(pollingTimerRef.current)
      }
    }
  }, [pollingActive, pollingInterval])

  useEffect(() => {
    const logs = [
      { id: 1, time: dayjs().subtract(5, 'minute').toISOString(), sender: '李安全', target: '沪A12345', message: '前方注意，前方5公里有服务区', priority: 'normal' },
      { id: 2, time: dayjs().subtract(12, 'minute').toISOString(), sender: '王押运', target: '京B67890', message: '前方检查点请减速慢行', priority: 'high' },
      { id: 3, time: dayjs().subtract(25, 'minute').toISOString(), sender: '调度中心', target: '粤C11111', message: '前方道路拥堵，请提前变道', priority: 'normal' },
    ]
    setIntercomLogs(logs)
  }, [])

  useEffect(() => {
    const records = [
      { id: 1, vehicle_plate: '沪A12345', record_type: 'scheduled', start_time: dayjs().subtract(2, 'hour').toISOString(), end_time: dayjs().subtract(1, 'hour').toISOString(), duration_minutes: 60, view_count: 5, file_url: '' },
      { id: 2, vehicle_plate: '京B67890', record_type: 'alarm', start_time: dayjs().subtract(4, 'hour').toISOString(), end_time: dayjs().subtract(3.5, 'hour').toISOString(), duration_minutes: 30, view_count: 12, file_url: '' },
      { id: 3, vehicle_plate: '粤C11111', record_type: 'manual', start_time: dayjs().subtract(6, 'hour').toISOString(), end_time: dayjs().subtract(5.5, 'hour').toISOString(), duration_minutes: 30, view_count: 2, file_url: '' },
      { id: 4, vehicle_plate: '粤D22222', record_type: 'scheduled', start_time: dayjs().subtract(1, 'day').toISOString(), end_time: dayjs().subtract(1, 'day').add(1, 'hour').toISOString(), duration_minutes: 60, view_count: 3, file_url: '' },
    ]
    setVideoRecords(records)
  }, [])

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

  const handleSOSHandle = async (alert: any) => {
    try {
      message.success('已受理 SOS 报警')
      setSosAlertModalVisible(false)
    } catch (e) {
      message.error('操作失败')
    }
  }

  const handleSOSResolve = async (alert: any) => {
    try {
      message.success('SOS 报警已解决')
      setSosAlertModalVisible(false)
    } catch (e) {
      message.error('操作失败')
    }
  }

  const handleSendIntercom = async () => {
    if (!intercomTarget || !intercomText.trim()) {
      message.warning('请输入喊话内容')
      return
    }
    try {
      const newLog = {
        id: Date.now(),
        time: new Date().toISOString(),
        sender: '当前用户',
        target: intercomTarget.vehicle_plate,
        message: intercomText,
        priority: 'normal',
      }
      setIntercomLogs(prev => [newLog, ...prev])
      message.success('喊话指令已发送')
      setIntercomText('')
      setIntercomModalVisible(false)
    } catch (e) {
      message.error('发送失败')
    }
  }

  const handleOpenTrackPlayback = async (vehicle: any) => {
    setPlaybackTarget(vehicle)
    setTrackPlaybackVisible(true)
    const mockTracks: any[] = Array.from({ length: 20 }, (_, i) => ({
      track_id: i + 1,
      vehicle_id: vehicle.vehicle_id || 1,
      latitude: (vehicle.latitude || 31.23) + (Math.random() - 0.5) * 0.1 + i * 0.002,
      longitude: (vehicle.longitude || 121.47) + (Math.random() - 0.5) * 0.1 + i * 0.003,
      speed: Math.floor(Math.random() * 40) + 40,
      timestamp: dayjs().subtract(20 - i, 'minute').toISOString(),
    }))
    setTrackPlaybackData(mockTracks)
    setPlaybackIndex(0)
  }

  const startPlayback = () => {
    if (!trackPlaybackData.length) return
    setPlaybackPlaying(true)
    playbackTimerRef.current = setInterval(() => {
      setPlaybackIndex(prev => {
        if (prev >= trackPlaybackData.length - 1) {
          stopPlayback()
          return prev
        }
        return prev + 1
      })
    }, 500)
  }

  const stopPlayback = () => {
    setPlaybackPlaying(false)
    if (playbackTimerRef.current) {
      clearInterval(playbackTimerRef.current)
      playbackTimerRef.current = null
    }
  }

  useEffect(() => {
    if (trackPlaybackVisible && playbackMapContainerRef.current && !playbackMapReady) {
      const timer = setTimeout(() => initPlaybackMap(), 200)
      return () => clearTimeout(timer)
    }
    return () => {
      if (playbackMapInstanceRef.current) {
        playbackMapInstanceRef.current.destroy()
        playbackMapInstanceRef.current = null
        playbackMarkerRef.current = null
        playbackPolylineRef.current = null
        setPlaybackMapReady(false)
      }
      stopPlayback()
    }
  }, [trackPlaybackVisible])

  useEffect(() => {
    if (playbackMarkerRef.current && trackPlaybackData[playbackIndex]) {
      const point = trackPlaybackData[playbackIndex]
      playbackMarkerRef.current.setPosition([point.longitude, point.latitude])
    }
  }, [playbackIndex])

  const initPlaybackMap = async () => {
    if (!playbackMapContainerRef.current || !trackPlaybackData.length) return
    try {
      const AMap: any = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: ['AMap.Scale', 'AMap.ToolBar', 'AMap.Polyline'],
      })
      const firstPoint = trackPlaybackData[0]
      const map = new AMap.Map(playbackMapContainerRef.current, {
        zoom: 14,
        center: [firstPoint.longitude, firstPoint.latitude],
        mapStyle: 'amap://styles/light',
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar())

      const path = trackPlaybackData.map(p => [p.longitude, p.latitude])
      const polyline = new AMap.Polyline({
        path,
        strokeColor: '#1677ff',
        strokeWeight: 4,
        strokeOpacity: 0.8,
      })
      polyline.setMap(map)
      playbackPolylineRef.current = polyline

      const marker = new AMap.Marker({
        position: [firstPoint.longitude, firstPoint.latitude],
        label: {
          content: `<div style="padding:4px 10px;background:#1677ff;color:#fff;border-radius:4px;font-size:12px;font-weight:600">
            ${playbackTarget?.vehicle_plate || '车辆'}
          </div>`,
          direction: 'top',
        },
      })
      marker.setMap(map)
      playbackMarkerRef.current = marker

      playbackMapInstanceRef.current = map
      setPlaybackMapReady(true)
    } catch (e) {
    }
  }

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

  const shiftColumns: ProColumns<any>[] = [
    { title: '排班ID', dataIndex: 'id', width: 80 },
    { title: '押运员', dataIndex: 'escort_name', width: 100 },
    { title: '排班日期', dataIndex: 'shift_date', width: 120, render: (t: any) => dayjs(t).format('YYYY-MM-DD') },
    { title: '时间段', dataIndex: 'start_time', width: 140, render: (_: any, r: any) => `${r.start_time} - ${r.end_time}` },
    { title: '车辆数', dataIndex: 'vehicle_count', width: 80 },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (s: string) => {
        const map: Record<string, { color: string; label: string }> = {
          scheduled: { color: 'blue', label: '已排班' },
          active: { color: 'processing', label: '进行中' },
          completed: { color: 'success', label: '已完成' },
          cancelled: { color: 'default', label: '已取消' },
        }
        return <Tag color={map[s]?.color}>{map[s]?.label || s}</Tag>
      },
    },
    { title: '调度员', dataIndex: 'dispatcher_name', width: 100 },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t: any) => formatDateTime(t) },
  ]

  const mockShifts = [
    { id: 1, escort_id: 1, escort_name: '李安全', shift_date: dayjs().format('YYYY-MM-DD'), start_time: '08:00', end_time: '20:00', vehicle_count: 3, status: 'active', dispatcher_name: '王调度', created_at: dayjs().subtract(1, 'day').toISOString() },
    { id: 2, escort_id: 2, escort_name: '王押运', shift_date: dayjs().format('YYYY-MM-DD'), start_time: '08:00', end_time: '20:00', vehicle_count: 2, status: 'active', dispatcher_name: '王调度', created_at: dayjs().subtract(1, 'day').toISOString() },
    { id: 3, escort_id: 3, escort_name: '张督察', shift_date: dayjs().add(1, 'day').format('YYYY-MM-DD'), start_time: '20:00', end_time: '08:00', vehicle_count: 4, status: 'scheduled', dispatcher_name: '李调度', created_at: dayjs().subtract(2, 'hour').toISOString() },
    { id: 4, escort_id: 1, escort_name: '李安全', shift_date: dayjs().subtract(1, 'day').format('YYYY-MM-DD'), start_time: '08:00', end_time: '20:00', vehicle_count: 3, status: 'completed', dispatcher_name: '王调度', created_at: dayjs().subtract(2, 'day').toISOString() },
  ]

  const renderTasksTab = () => (
    <Row gutter={16}>
      <Col xs={24} lg={9} xl={8}>
        <ProCard
          bordered={false}
          style={{ borderRadius: 12, height: 'calc(100vh - 320px)', minHeight: 550 }}
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
                        <div style={{ height: 6, background: '#f0f0f0', borderRadius: 3, overflow: 'hidden' }}>
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
          style={{ borderRadius: 12, height: 'calc(100vh - 320px)', minHeight: 550 }}
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
                <Button type="link" icon={<HistoryOutlined />} size="small" onClick={() => handleOpenTrackPlayback(selectedWaybill)}>轨迹回放</Button>
                <Button type="link" icon={<PhoneOutlined />} size="small" onClick={() => { setIntercomTarget(selectedWaybill); setIntercomModalVisible(true) }}>喊话</Button>
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
                        position: 'relative', left: 0,
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
                          marginBottom: 16, borderRadius: 10,
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
                            <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setEventDetailDrawer(event)}>
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
                          <Paragraph style={{ fontSize: 12, marginBottom: 0 }} ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}>
                            {event.remark}
                          </Paragraph>
                          {event.photos.length > 0 && (
                            <Space size={8} wrap>
                              {event.photos.map((p, pi) => (
                                <Image key={pi} src={p} alt="" width={100} height={75}
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
  )

  const renderPollingTab = () => (
    <Row gutter={16}>
      <Col xs={24}>
        <ProCard bordered={false} style={{ borderRadius: 12, marginBottom: 16 }} bodyStyle={{ padding: 16 }}>
          <Row gutter={16} align="middle">
            <Col xs={24} md={6}>
              <Space>
                <Text strong><VideoCameraOutlined style={{ color: '#1677ff' }} /> 视频轮询监控</Text>
                <Badge status={pollingActive ? 'processing' : 'default'} text={pollingActive ? '轮询中' : '已暂停'} />
              </Space>
            </Col>
            <Col xs={24} md={10}>
              <Space>
                <Text type="secondary" style={{ fontSize: 12 }}>轮询间隔:</Text>
                <Select value={pollingInterval} onChange={setPollingInterval} style={{ width: 120 }} size="small">
                  <Option value={10}>10 秒</Option>
                  <Option value={20}>20 秒</Option>
                  <Option value={30}>30 秒（推荐）</Option>
                  <Option value={60}>60 秒</Option>
                </Select>
                <Text type="secondary" style={{ fontSize: 12 }}>监控车辆: {pollingVehicles.length} 辆</Text>
              </Space>
            </Col>
            <Col xs={24} md={8} style={{ textAlign: 'right' }}>
              <Space>
                <Button icon={<ReloadOutlined />} size="small" onClick={() => {
                  message.success('已刷新轮询画面')
                  setPollingVehicles(prev => [...prev])
                }}>刷新</Button>
                <Button
                  icon={pollingActive ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                  type={pollingActive ? 'default' : 'primary'}
                  size="small"
                  onClick={() => setPollingActive(!pollingActive)}
                >
                  {pollingActive ? '暂停轮询' : '开始轮询'}
                </Button>
              </Space>
            </Col>
          </Row>
        </ProCard>
      </Col>
      {pollingVehicles.map(vehicle => (
        <Col xs={24} sm={12} lg={8} xl={6} key={vehicle.vehicle_id}>
          <Card
            hoverable
            style={{ borderRadius: 12, marginBottom: 16 }}
            bodyStyle={{ padding: 0 }}
            onClick={() => setSelectedPollingVehicle(vehicle)}
            cover={
              <div
                style={{
                  height: 180,
                  background: `linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  position: 'relative',
                  borderRadius: '12px 12px 0 0',
                }}
              >
                <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                  <CameraOutlined style={{ fontSize: 36, opacity: 0.3 }} />
                  <div style={{ fontSize: 12, marginTop: 8 }}>
                    车载摄像头 - {vehicle.vehicle_plate}
                  </div>
                  <div style={{ fontSize: 10, marginTop: 4, color: '#52c41a' }}>
                    ● LIVE · {dayjs(vehicle.last_frame_time).format('HH:mm:ss')}
                  </div>
                </div>
                {vehicle.status === 'resting' && (
                  <Tag color="orange" style={{ position: 'absolute', top: 8, right: 8 }}>休息中</Tag>
                )}
                {vehicle.status === 'driving' && (
                  <Tag color="green" style={{ position: 'absolute', top: 8, right: 8 }}>行驶中</Tag>
                )}
              </div>
            }
          >
            <div style={{ padding: 12 }}>
              <Space direction="vertical" size={6} style={{ width: '100%' }}>
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <Tag color="blue" style={{ fontSize: 13, margin: 0 }}>{vehicle.vehicle_plate}</Tag>
                  <Text strong style={{ fontSize: 16, color: vehicle.speed > 70 ? '#ff4d4f' : '#1677ff' }}>
                    {vehicle.speed} <Text type="secondary" style={{ fontSize: 11 }}>km/h</Text>
                  </Text>
                </Space>
                <div style={{ fontSize: 12 }}>
                  <UserOutlined style={{ color: '#fa8c16' }} /> 司机: {vehicle.driver_name}
                </div>
                <div style={{ fontSize: 12 }}>
                  <SafetyCertificateOutlined style={{ color: '#52c41a' }} /> 货物: {vehicle.danger_goods}
                </div>
                <div style={{ fontSize: 12 }}>
                  <span style={{ color: vehicle.driver_status === '正常' ? '#52c41a' : '#faad14' }}>●</span> 驾驶员状态: {vehicle.driver_status}
                </div>
                <Divider style={{ margin: '8px 0' }} />
                <Space size={4}>
                  <Button size="small" icon={<EyeOutlined />} block onClick={(e) => { e.stopPropagation(); message.success('查看实时画面') }}>查看</Button>
                  <Button size="small" type="primary" icon={<PhoneOutlined />} block onClick={(e) => { e.stopPropagation(); setIntercomTarget(vehicle); setIntercomModalVisible(true) }}>喊话</Button>
                  <Button size="small" icon={<HistoryOutlined />} block onClick={(e) => { e.stopPropagation(); handleOpenTrackPlayback(vehicle) }}>轨迹</Button>
                </Space>
              </Space>
            </div>
          </Card>
        </Col>
      ))}
    </Row>
  )

  const renderAlertsTab = () => (
    <Row gutter={16}>
      <Col xs={24}>
        <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
          <ProTable<any>
            columns={[
              { title: '报警ID', dataIndex: 'id', width: 80 },
              {
                title: '车辆', dataIndex: 'vehicle_plate', width: 120,
                render: (t, r) => (
                  <Space>
                    <CarOutlined style={{ color: '#1677ff' }} />
                    <Text strong>{t}</Text>
                  </Space>
                ),
              },
              { title: '司机', dataIndex: 'driver_name', width: 100 },
              {
                title: '报警类型', dataIndex: 'sos_type', width: 120,
                render: (t) => <Tag color="red">{t || '紧急报警'}</Tag>,
              },
              { title: '位置', dataIndex: 'location', width: 200, ellipsis: true },
              { title: '描述', dataIndex: 'description', width: 200, ellipsis: true },
              {
                title: '状态', dataIndex: 'status', width: 100,
                render: (s) => {
                  const map: Record<string, { color: string; label: string }> = {
                    pending: { color: 'red', label: '待处理' },
                    processing: { color: 'orange', label: '处理中' },
                    resolved: { color: 'green', label: '已解决' },
                    ignored: { color: 'default', label: '已忽略' },
                  }
                  return <Tag color={map[s]?.color}>{map[s]?.label || s}</Tag>
                },
              },
              { title: '处理人', dataIndex: 'handler_name', width: 100 },
              { title: '报警时间', dataIndex: 'created_at', width: 160, render: (t) => formatDateTime(t) },
              {
                title: '操作', width: 180,
                render: (_, r) => (
                  <Space size={4}>
                    <Button size="small" type="primary" danger disabled={r.status !== 'pending'} onClick={() => handleSOSHandle(r)}>
                      受理
                    </Button>
                    <Button size="small" disabled={r.status === 'resolved'} onClick={() => handleSOSResolve(r)}>
                      解决
                    </Button>
                  </Space>
                ),
              },
            ]}
            dataSource={[
              { id: 1, vehicle_plate: '沪A12345', driver_name: '张建国', sos_type: '紧急求救', location: '上海市浦东新区G1503高速', description: '车辆故障，请求援助', status: 'pending', handler_name: null, created_at: dayjs().subtract(5, 'minute').toISOString() },
              { id: 2, vehicle_plate: '粤D22222', driver_name: '刘文华', sos_type: '交通事故', location: '深圳市南山区北环大道', description: '轻微追尾，无人员伤亡', status: 'processing', handler_name: '李安全', created_at: dayjs().subtract(25, 'minute').toISOString() },
              { id: 3, vehicle_plate: '京B67890', driver_name: '李明辉', sos_type: '货物异常', location: '北京市朝阳区六环', description: '闻到轻微异味，已停车检查', status: 'resolved', handler_name: '王押运', created_at: dayjs().subtract(2, 'hour').toISOString() },
            ]}
            search={false}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => message.success('已刷新')}>刷新</Button>,
            ]}
          />
        </ProCard>
      </Col>
    </Row>
  )

  const renderShiftsTab = () => (
    <Row gutter={16}>
      <Col xs={24}>
        <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
          <ProTable<any>
            columns={shiftColumns}
            dataSource={mockShifts}
            search={false}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => message.success('已刷新')}>刷新</Button>,
              <Button key="add" icon={<PlusOutlined />} type="primary" onClick={() => message.success('创建排班功能')}>创建排班</Button>,
            ]}
          />
        </ProCard>
      </Col>
    </Row>
  )

  const renderVideosTab = () => (
    <Row gutter={16}>
      <Col xs={24}>
        <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
          <ProTable<any>
            columns={[
              { title: '录像ID', dataIndex: 'id', width: 80 },
              { title: '车牌号', dataIndex: 'vehicle_plate', width: 120 },
              {
                title: '录像类型', dataIndex: 'record_type', width: 100,
                render: (t) => {
                  const map: Record<string, { color: string; label: string }> = {
                    scheduled: { color: 'blue', label: '定时轮询' },
                    alarm: { color: 'red', label: '报警触发' },
                    manual: { color: 'purple', label: '人工录制' },
                  }
                  return <Tag color={map[t]?.color}>{map[t]?.label || t}</Tag>
                },
              },
              { title: '开始时间', dataIndex: 'start_time', width: 160, render: (t) => formatDateTime(t) },
              { title: '结束时间', dataIndex: 'end_time', width: 160, render: (t) => formatDateTime(t) },
              { title: '时长(分钟)', dataIndex: 'duration_minutes', width: 100 },
              { title: '查看次数', dataIndex: 'view_count', width: 100 },
              {
                title: '操作', width: 150,
                render: (_, r) => (
                  <Space size={4}>
                    <Button size="small" type="primary" icon={<EyeOutlined />} onClick={() => message.success(`查看录像 ${r.id}`)}>查看</Button>
                    <Button size="small" icon={<Download />} onClick={() => message.success('开始下载')}>下载</Button>
                  </Space>
                ),
              },
            ]}
            dataSource={videoRecords}
            search={false}
            loading={videoRecordLoading}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => message.success('已刷新')}>刷新</Button>,
              <Alert key="tip" type="info" showIcon message="视频记录云端存储90天，过期自动清理" style={{ marginLeft: 16 }} />,
            ]}
          />
        </ProCard>
      </Col>
    </Row>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 12 }}>
        <Row gutter={16} align="middle">
          <Col xs={24} md={8}>
            <Space wrap>
              <Text strong style={{ fontSize: 15 }}>
                <SafetyCertificateOutlined style={{ color: '#1677ff' }} /> 危险品电子押运
              </Text>
              <Tag color="processing">轮询中 {statistics.polling_vehicles} 车</Tag>
              <Tag color="red">待处理 SOS {statistics.pending_sos}</Tag>
              <Tag color="success">今日排班 {statistics.active_shifts}</Tag>
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
          <Col xs={24} md={8}>
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

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'tasks',
            label: (
              <Space>
                <CarOutlined />
                押运任务
                <Badge count={filteredWaybills.length} showZero style={{ backgroundColor: '#1677ff' }} />
              </Space>
            ),
            children: renderTasksTab(),
          },
          {
            key: 'polling',
            label: (
              <Space>
                <VideoCameraOutlined />
                视频轮询
                <Badge status={pollingActive ? 'processing' : 'default'} />
              </Space>
            ),
            children: renderPollingTab(),
          },
          {
            key: 'alerts',
            label: (
              <Space>
                <BellOutlined />
                SOS 报警
                {sosAlerts.length > 0 && <Badge count={sosAlerts.length} style={{ backgroundColor: '#ff4d4f' }} />}
              </Space>
            ),
            children: renderAlertsTab(),
          },
          {
            key: 'shifts',
            label: (
              <Space>
                <ScheduleOutlined />
                排班管理
              </Space>
            ),
            children: renderShiftsTab(),
          },
          {
            key: 'videos',
            label: (
              <Space>
                <CameraOutlined />
                视频记录
              </Space>
            ),
            children: renderVideosTab(),
          },
        ]}
      />

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
              <Alert type="error" showIcon icon={<WarningOutlined />} message="高风险事件提醒" description="该事件为高风险押运事件，建议立即核查并采取处置措施" style={{ borderRadius: 8 }} />
            )}
            {eventDetailDrawer.risk_level === 'warning' && (
              <Alert type="warning" showIcon message="需关注事件" description="该事件存在一定风险，请持续关注后续进展" style={{ borderRadius: 8 }} />
            )}
            <Card size="small" style={{ borderRadius: 8 }}>
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="事件编号"><Text copyable>{eventDetailDrawer.id}</Text></Descriptions.Item>
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
              <div ref={eventMapContainerRef} style={{ height: 240, borderRadius: 6, background: '#f0f5ff', display: eventMapReady ? 'block' : 'flex', alignItems: 'center', justifyContent: 'center' }}>
                {!eventMapReady && (
                  <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                    <MapOutlined style={{ fontSize: 36, opacity: 0.4 }} />
                    <div style={{ marginTop: 8, fontSize: 12 }}>{eventDetailDrawer.location}</div>
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
                      <Image src={p} alt="" style={{ width: '100%', height: 140, objectFit: 'cover', borderRadius: 6 }} preview={{ mask: <EyeOutlined /> }} />
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

      <Modal
        title={
          <Space>
            <WarningOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
            <Text strong type="danger" style={{ fontSize: 16 }}>🚨 紧急 SOS 报警</Text>
          </Space>
        }
        open={sosAlertModalVisible}
        onCancel={() => setSosAlertModalVisible(false)}
        width={640}
        maskClosable={false}
        footer={[
          <Button key="ignore" onClick={() => setSosAlertModalVisible(false)}>忽略</Button>,
          <Button key="intercom" icon={<PhoneOutlined />} onClick={() => { setIntercomTarget(currentSosAlert); setSosAlertModalVisible(false); setIntercomModalVisible(true) }}>
            立即喊话
          </Button>,
          <Button key="handle" type="primary" danger icon={<BellOutlined />} onClick={() => handleSOSHandle(currentSosAlert)}>
            受理报警
          </Button>,
        ]}
      >
        {currentSosAlert && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Alert type="error" showIcon message="司机按下紧急按钮，请立即处理！" description={`报警时间: ${formatDateTime(currentSosAlert.created_at || new Date().toISOString())}`} />
            <Descriptions column={2} size="small" bordered>
              <Descriptions.Item label="车牌号">{currentSosAlert.vehicle_plate || '-'}</Descriptions.Item>
              <Descriptions.Item label="司机">{currentSosAlert.driver_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="报警类型" span={2}>
                <Tag color="red">{currentSosAlert.sos_type || '紧急求救'}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="报警位置" span={2}>{currentSosAlert.location || currentSosAlert.description || '-'}</Descriptions.Item>
              <Descriptions.Item label="详细描述" span={2}>{currentSosAlert.description || '无详细描述'}</Descriptions.Item>
            </Descriptions>
            <Card size="small" title={<Space><PhoneOutlined />对讲记录</Space>}>
              <List
                size="small"
                dataSource={intercomLogs.slice(0, 3)}
                renderItem={(item: any) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<SoundOutlined />} style={{ background: '#52c41a' }} />}
                      title={<Space><Text strong>{item.sender}</Text><Tag color={item.priority === 'high' ? 'red' : 'blue'}>{item.priority === 'high' ? '高优先级' : '普通'}</Tag></Space>}
                      description={<Space><ClockCircleOutlined /><Text type="secondary">{formatDateTime(item.time)}</Text></Space>}
                    />
                    <Text>{item.message}</Text>
                  </List.Item>
                )}
              />
            </Card>
          </div>
        )}
      </Modal>

      <Modal
        title={
          <Space>
            <PhoneOutlined style={{ color: '#1677ff' }} />
            <Text strong>语音喊话指令</Text>
            {intercomTarget && <Tag color="blue">{intercomTarget.vehicle_plate || intercomTarget.waybill_no}</Tag>}
          </Space>
        }
        open={intercomModalVisible}
        onCancel={() => { setIntercomModalVisible(false); setIntercomText('') }}
        footer={[
          <Button key="cancel" onClick={() => { setIntercomModalVisible(false); setIntercomText('') }}>取消</Button>,
          <Button key="send" type="primary" icon={<SendOutlined />} onClick={handleSendIntercom}>
            发送喊话
          </Button>,
        ]}
        width={560}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Alert type="info" showIcon message="快捷指令" description="点击下方快捷按钮快速发送常用喊话指令" />
          <Space wrap>
            {['前方检查点请减速', '前方限速请慢行', '前方服务区可休息', '请保持安全车距', '即将到达目的地', '请勿疲劳驾驶'].map(msg => (
              <Tag key={msg} color="blue" style={{ cursor: 'pointer', padding: '6px 12px', fontSize: 13 }} onClick={() => setIntercomText(msg)}>
                {msg}
              </Tag>
            ))}
          </Space>
          <Form layout="vertical">
            <Form.Item label="喊话内容">
              <TextArea
                value={intercomText}
                onChange={e => setIntercomText(e.target.value)}
                rows={4}
                placeholder="请输入要发送的语音喊话内容，系统将自动转换为语音播放给司机..."
                maxLength={200}
                showCount
              />
            </Form.Item>
            <Form.Item label="优先级">
              <Select defaultValue="normal">
                <Option value="normal">普通 - 正常播报</Option>
                <Option value="high">高优先级 - 强提醒</Option>
              </Select>
            </Form.Item>
          </Form>
          <Divider style={{ margin: '4px 0 8px' }} />
          <Card size="small" title={<Space><HistoryOutlined />历史对讲记录</Space>} bodyStyle={{ padding: 0 }}>
            <List
              size="small"
              dataSource={intercomLogs.slice(0, 5)}
              renderItem={(item: any) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar size={24} icon={<SoundOutlined />} style={{ background: '#52c41a' }} />}
                    title={<Space><Text strong>{item.sender}</Text><Text type="secondary">→</Text><Tag>{item.target}</Tag></Space>}
                    description={<Space><ClockCircleOutlined /><Text type="secondary" style={{ fontSize: 11 }}>{formatDateTime(item.time)}</Text></Space>}
                  />
                  <Text style={{ fontSize: 12 }}>{item.message}</Text>
                </List.Item>
              )}
            />
          </Card>
        </div>
      </Modal>

      <Modal
        title={
          <Space>
            <MapOutlined style={{ color: '#1677ff' }} />
            <Text strong>押运轨迹回放</Text>
            {playbackTarget && <Tag color="blue">{playbackTarget.vehicle_plate || playbackTarget.waybill_no}</Tag>}
          </Space>
        }
        open={trackPlaybackVisible}
        onCancel={() => {
          setTrackPlaybackVisible(false)
          stopPlayback()
          setPlaybackIndex(0)
          if (playbackMapInstanceRef.current) {
            playbackMapInstanceRef.current.destroy()
            playbackMapInstanceRef.current = null
            playbackMarkerRef.current = null
            playbackPolylineRef.current = null
            setPlaybackMapReady(false)
          }
        }}
        width={900}
        footer={[
          <Button key="close" onClick={() => {
            setTrackPlaybackVisible(false)
            stopPlayback()
          }}>关闭</Button>,
        ]}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Row gutter={16}>
            <Col span={24}>
              <div
                ref={playbackMapContainerRef}
                style={{
                  height: 400,
                  borderRadius: 8,
                  background: '#f0f5ff',
                  display: playbackMapReady ? 'block' : 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {!playbackMapReady && (
                  <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                    <MapOutlined style={{ fontSize: 48, opacity: 0.4 }} />
                    <div style={{ marginTop: 12 }}>正在加载轨迹地图...</div>
                  </div>
                )}
              </div>
            </Col>
          </Row>
          {trackPlaybackData.length > 0 && (
            <Card size="small" title={<Space><PlayCircleOutlined />播放控制</Space>}>
              <Space direction="vertical" style={{ width: '100%' }} size={12}>
                <Space style={{ width: '100%', justifyContent: 'center' }}>
                  <Button
                    icon={playbackPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                    type="primary"
                    size="large"
                    onClick={() => playbackPlaying ? stopPlayback() : startPlayback()}
                  >
                    {playbackPlaying ? '暂停' : '播放'}
                  </Button>
                  <Button onClick={() => { stopPlayback(); setPlaybackIndex(0) }}>重置</Button>
                </Space>
                <Progress
                  percent={Math.floor((playbackIndex / (trackPlaybackData.length - 1)) * 100)}
                  showInfo
                  format={() => `${playbackIndex + 1} / ${trackPlaybackData.length}`}
                />
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic title="当前速度" value={trackPlaybackData[playbackIndex]?.speed || 0} suffix="km/h" />
                  </Col>
                  <Col span={8}>
                    <Statistic title="当前时间" value={trackPlaybackData[playbackIndex]?.timestamp ? formatDateTime(trackPlaybackData[playbackIndex].timestamp, 'HH:mm:ss') : '-'} />
                  </Col>
                  <Col span={8}>
                    <Statistic title="总轨迹点" value={trackPlaybackData.length} />
                  </Col>
                </Row>
              </Space>
            </Card>
          )}
        </div>
      </Modal>

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
        modalProps={{ destroyOnClose: true, maskClosable: false }}
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
        <Alert type="info" showIcon message="正在为选中的运单添加押运事件" description={selectedWaybill ? `${selectedWaybill.waybill_no} · ${selectedWaybill.vehicle_plate} · ${selectedWaybill.driver_name}` : ''} style={{ marginBottom: 16, borderRadius: 8 }} />
        <ProFormSelect
          label="事件类型"
          name="event_type"
          width="md"
          rules={[{ required: true, message: '请选择事件类型' }]}
          placeholder="请选择押运事件类型"
          options={Object.entries(eventTypeMap).map(([k, v]) => ({ label: v.label, value: k }))}
        />
        <ProFormText
          label="发生位置"
          name="location"
          width="md"
          placeholder="请输入事件发生位置，留空则使用车辆当前位置"
          fieldProps={{ prefix: <EnvironmentOutlined /> }}
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
          <Dragger multiple listType="picture" beforeUpload={() => false} maxCount={9}>
            <p className="ant-upload-drag-icon"><CameraOutlined style={{ fontSize: 32, color: '#1677ff' }} /></p>
            <p className="ant-upload-text">点击或拖拽照片到此处上传</p>
            <p className="ant-upload-hint" style={{ fontSize: 12 }}>支持JPG/PNG格式，最多9张</p>
          </Dragger>
        </Form.Item>
        <Divider style={{ margin: '4px 0 12px' }} />
        <Alert type="warning" showIcon message="添加说明" description="押运事件将同步记录于运单档案，作为运输合规性凭证。添加后将自动通知相关人员。" style={{ borderRadius: 6 }} />
      </ModalForm>
    </div>
  )
}

export default Escort
