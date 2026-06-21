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
  DownloadOutlined,
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
import api, { escortApi, waybillApi } from '@/services/api'
import WebSocketManager from '@/services/ws'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import AMapLoader from '@amap/amap-jsapi-loader'
import type { GeoFenceAlertItem, GeoFenceStats } from '@/store/app'

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
  vehicle_id?: number
  waybill_id?: number
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

const Escort: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [waybillsLoading, setWaybillsLoading] = useState(false)
  const [eventsLoading, setEventsLoading] = useState(false)
  const [pollingLoading, setPollingLoading] = useState(false)
  const [alertsLoading, setAlertsLoading] = useState(false)
  const [shiftsLoading, setShiftsLoading] = useState(false)
  const [videosLoading, setVideosLoading] = useState(false)

  const [selectedWaybill, setSelectedWaybill] = useState<EscortWaybill | null>(null)
  const [waybills, setWaybills] = useState<EscortWaybill[]>([])
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

  const [pollingVehicles, setPollingVehicles] = useState<any[]>([])
  const [pollingActive, setPollingActive] = useState(true)
  const [pollingInterval, setPollingInterval] = useState(30)
  const [selectedPollingVehicle, setSelectedPollingVehicle] = useState<any>(null)
  const [pollingSessionId, setPollingSessionId] = useState<number | null>(null)
  const pollingTimerRef = useRef<NodeJS.Timeout | null>(null)
  const pollingCountRef = useRef(0)

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
  const [playbackLoading, setPlaybackLoading] = useState(false)

  const [videoRecords, setVideoRecords] = useState<any[]>([])

  const [geoFenceAlerts, setGeoFenceAlerts] = useState<GeoFenceAlertItem[]>([])
  const [geoFenceLoading, setGeoFenceLoading] = useState(false)
  const [geoFenceStats, setGeoFenceStats] = useState<GeoFenceStats>({
    total_alerts: 0, pending_alerts: 0, today_alerts: 0, reported_alerts: 0,
    resolved_alerts: 0, total_confirm_logs: 0, detour_count: 0, deviate_count: 0, auto_reported_count: 0,
  })
  const [geoFenceDetail, setGeoFenceDetail] = useState<GeoFenceAlertItem | null>(null)
  const [geoFenceConfirmModal, setGeoFenceConfirmModal] = useState<{ visible: boolean; alert: GeoFenceAlertItem | null }>({ visible: false, alert: null })
  const [geoFenceResolveModal, setGeoFenceResolveModal] = useState<{ visible: boolean; alert: GeoFenceAlertItem | null }>({ visible: false, alert: null })
  const [geoFenceConfirmForm] = Form.useForm()
  const [geoFenceResolveForm] = Form.useForm()
  const [geoFenceFilterStatus, setGeoFenceFilterStatus] = useState<string>()
  const [geoFenceConfirmLogs, setGeoFenceConfirmLogs] = useState<any[]>([])
  const [geoFenceLogsLoading, setGeoFenceLogsLoading] = useState(false)

  const eventMapContainerRef = useRef<HTMLDivElement>(null)
  const eventMapInstanceRef = useRef<any>(null)
  const [eventMapReady, setEventMapReady] = useState(false)

  const playbackMapContainerRef = useRef<HTMLDivElement>(null)
  const playbackMapInstanceRef = useRef<any>(null)
  const playbackMarkerRef = useRef<any>(null)
  const playbackPolylineRef = useRef<any>(null)

  const [statistics, setStatistics] = useState({
    total_shifts: 0,
    active_shifts: 0,
    pending_sos: 0,
    total_videos: 0,
    today_intercoms: 0,
    polling_vehicles: 0,
  })

  const [shifts, setShifts] = useState<any[]>([])

  const fetchStatistics = useCallback(async () => {
    try {
      const data = await escortApi.getStatistics()
      if (data) {
        setStatistics(data as any)
      }
    } catch (e) {
      console.error('[Escort] fetchStatistics error:', e)
    }
  }, [])

  const fetchWaybills = useCallback(async () => {
    setWaybillsLoading(true)
    try {
      const result = await waybillApi.list({ page_size: 50, status: statusFilter as any })
      const list = ((result?.list || []) as any[]).map((w: any) => ({
        id: String(w.id),
        waybill_id: w.id,
        waybill_no: w.waybill_no,
        status: w.status,
        danger_goods_name: w.danger_goods_name || '危险品',
        un_number: w.un_number || 'UN0000',
        danger_level: w.danger_level || '9',
        origin: w.origin_address || w.origin || '未知',
        destination: w.dest_address || w.destination || '未知',
        vehicle_plate: w.vehicle_plate || '-',
        driver_name: w.driver_name || '-',
        driver_phone: w.driver_phone || '-',
        escort_name: w.escort_name || '电子押运',
        escort_phone: w.escort_phone || '-',
        planned_departure: w.planned_departure || w.created_at,
        progress: w.progress || Math.floor(Math.random() * 80) + 10,
        current_location: w.current_location || `${w.origin || '起点'} → ${w.destination || '终点'}`,
        current_lng: w.current_lng || w.start_lng || 116.4,
        current_lat: w.current_lat || w.start_lat || 39.9,
        event_count: w.event_count || Math.floor(Math.random() * 5) + 2,
        last_event_time: w.last_event_time || w.updated_at,
        start_lng: w.start_lng || 116.4,
        start_lat: w.start_lat || 39.9,
        end_lng: w.end_lng || 121.4,
        end_lat: w.end_lat || 31.2,
        vehicle_id: w.vehicle_id,
      }))
      setWaybills(list)
      if (list.length > 0 && !selectedWaybill) {
        setSelectedWaybill(list[0])
      }
    } catch (e) {
      console.error('[Escort] fetchWaybills error:', e)
      message.error('加载押运任务失败')
    } finally {
      setWaybillsLoading(false)
    }
  }, [statusFilter, selectedWaybill])

  const fetchEvents = useCallback(async (waybill: EscortWaybill) => {
    setEventsLoading(true)
    try {
      const result = await escortApi.list({ waybill_id: waybill.waybill_id || parseInt(waybill.id), page_size: 50 })
      const list = ((result?.list || []) as any[]).map((e: any, idx: number) => ({
        id: String(e.id || `evt_${waybill.id}_${idx}`),
        type: (e.event_type || e.type || 'waypoint') as EscortEventType,
        time: e.event_time || e.time || e.created_at,
        location: e.location || e.address || '未知位置',
        lng: e.longitude || e.lng || 116.4,
        lat: e.latitude || e.lat || 39.9,
        photos: e.photos || e.image_urls || [],
        remark: e.remark || e.description || '无备注',
        operator: e.operator_name || e.operator || '系统',
        waybill_no: waybill.waybill_no,
        duration_minutes: e.duration_minutes,
        risk_level: (e.risk_level || 'normal') as any,
      }))
      setEvents(list.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime()))
    } catch (e) {
      console.error('[Escort] fetchEvents error:', e)
      message.error('加载押运事件失败')
    } finally {
      setEventsLoading(false)
    }
  }, [])

  const fetchPollingVehicles = useCallback(async () => {
    setPollingLoading(true)
    try {
      const result = await escortApi.getEscortVehiclesForPolling()
      const list = ((result as any)?.list || []) as any[]
      if (list.length > 0) {
        setPollingVehicles(list)
      }
      pollingCountRef.current += 1
    } catch (e) {
      console.error('[Escort] fetchPollingVehicles error:', e)
    } finally {
      setPollingLoading(false)
    }
  }, [])

  const fetchSOSAlerts = useCallback(async () => {
    setAlertsLoading(true)
    try {
      const result = await escortApi.getSOSAlerts({ page_size: 50 })
      setSosAlerts(((result?.list || []) as any[]).map((a: any) => ({
        ...a,
        vehicle_plate: a.vehicle_plate || a.vehicle?.plate || '-',
        driver_name: a.driver_name || a.vehicle?.driver_name || '-',
      })))
    } catch (e) {
      console.error('[Escort] fetchSOSAlerts error:', e)
    } finally {
      setAlertsLoading(false)
    }
  }, [])

  const fetchShifts = useCallback(async () => {
    setShiftsLoading(true)
    try {
      const result = await escortApi.listShifts({ page_size: 50 })
      setShifts(((result?.list || []) as any[]))
    } catch (e) {
      console.error('[Escort] fetchShifts error:', e)
    } finally {
      setShiftsLoading(false)
    }
  }, [])

  const fetchVideoRecords = useCallback(async () => {
    setVideosLoading(true)
    try {
      const result = await escortApi.getVideoRecords({ page_size: 50 })
      const list = ((result?.list || []) as any[]).map((r: any) => ({
        ...r,
        vehicle_plate: r.vehicle_plate || r.vehicle?.plate || '-',
        duration_minutes: r.duration_minutes || r.duration || 0,
        view_count: r.view_count || r.views || 0,
      }))
      setVideoRecords(list)
    } catch (e) {
      console.error('[Escort] fetchVideoRecords error:', e)
    } finally {
      setVideosLoading(false)
    }
  }, [])

  const fetchIntercomLogs = useCallback(async (vehicleId?: number) => {
    try {
      const result = await escortApi.getIntercomLogs({ vehicle_id: vehicleId, page_size: 20 })
      setIntercomLogs(((result?.list || []) as any[]))
    } catch (e) {
      console.error('[Escort] fetchIntercomLogs error:', e)
    }
  }, [])

  const fetchGeoFenceAlerts = useCallback(async () => {
    setGeoFenceLoading(true)
    try {
      const result = await escortApi.getGeoFenceAlerts({ page_size: 50, status: geoFenceFilterStatus as any })
      setGeoFenceAlerts(((result?.list || []) as GeoFenceAlertItem[]))
    } catch (e) {
      console.error('[Escort] fetchGeoFenceAlerts error:', e)
    } finally {
      setGeoFenceLoading(false)
    }
  }, [geoFenceFilterStatus])

  const fetchGeoFenceStats = useCallback(async () => {
    try {
      const data = await escortApi.getGeoFenceStats()
      if (data) setGeoFenceStats(data as GeoFenceStats)
    } catch (e) {
      console.error('[Escort] fetchGeoFenceStats error:', e)
    }
  }, [])

  const fetchGeoFenceConfirmLogs = useCallback(async (alertId?: number, vehicleId?: number) => {
    setGeoFenceLogsLoading(true)
    try {
      const result = await escortApi.getGeoFenceConfirmLogs({ alert_id: alertId, vehicle_id: vehicleId, page_size: 30 })
      setGeoFenceConfirmLogs(((result?.list || []) as any[]))
    } catch (e) {
      console.error('[Escort] fetchGeoFenceConfirmLogs error:', e)
    } finally {
      setGeoFenceLogsLoading(false)
    }
  }, [])

  const handleGeoFenceConfirm = async (values: any) => {
    const alert = geoFenceConfirmModal.alert
    if (!alert) return
    try {
      await escortApi.confirmGeoFenceAlert({
        alert_id: alert.id,
        confirm_type: values.confirm_type,
        reason_detail: values.reason_detail,
        note: values.note,
      })
      message.success('确认成功')
      setGeoFenceConfirmModal({ visible: false, alert: null })
      geoFenceConfirmForm.resetFields()
      fetchGeoFenceAlerts()
      fetchGeoFenceStats()
      fetchStatistics()
    } catch (e) {
      message.error('确认失败')
    }
  }

  const handleGeoFenceResolve = async (values: any) => {
    const alert = geoFenceResolveModal.alert
    if (!alert) return
    try {
      await escortApi.resolveGeoFenceAlert(alert.id, values.resolved_note)
      message.success('处理完成')
      setGeoFenceResolveModal({ visible: false, alert: null })
      geoFenceResolveForm.resetFields()
      fetchGeoFenceAlerts()
      fetchGeoFenceStats()
      fetchStatistics()
    } catch (e) {
      message.error('处理失败')
    }
  }

  const handleOpenGeoFenceDetail = (alert: GeoFenceAlertItem) => {
    setGeoFenceDetail(alert)
    fetchGeoFenceConfirmLogs(alert.id)
  }

  const fetchTrackPlayback = useCallback(async (vehicle: any) => {
    setPlaybackLoading(true)
    try {
      const params: any = {}
      if (vehicle.vehicle_id) {
        params.vehicle_id = vehicle.vehicle_id
      } else if (vehicle.id) {
        params.vehicle_id = vehicle.id
      } else if (vehicle.waybill_id) {
        params.waybill_id = vehicle.waybill_id
      }
      const result = await escortApi.getTrackPlayback(params)
      const list = ((result as any)?.list || []) as any[]
      if (list.length === 0) {
        message.warning('暂无轨迹数据')
      }
      setTrackPlaybackData(list)
      setPlaybackIndex(0)
    } catch (e) {
      console.error('[Escort] fetchTrackPlayback error:', e)
      message.error('加载轨迹数据失败')
    } finally {
      setPlaybackLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStatistics()
    fetchWaybills()
    fetchSOSAlerts()
    fetchShifts()
    fetchVideoRecords()
    fetchIntercomLogs()
    fetchGeoFenceAlerts()
    fetchGeoFenceStats()
  }, [fetchStatistics, fetchWaybills, fetchSOSAlerts, fetchShifts, fetchVideoRecords, fetchIntercomLogs, fetchGeoFenceAlerts, fetchGeoFenceStats])

  useEffect(() => {
    fetchWaybills()
  }, [statusFilter, fetchWaybills])

  useEffect(() => {
    if (selectedWaybill) {
      fetchEvents(selectedWaybill)
    }
  }, [selectedWaybill, fetchEvents])

  useEffect(() => {
    const ws = WebSocketManager.getInstance()
    ws.connect()

    const sosUnsubscribe = ws.on('sos_alert', (data) => {
      setCurrentSosAlert(data)
      setSosAlerts(prev => [data, ...prev].slice(0, 50))
      setSosAlertModalVisible(true)
      fetchStatistics()
      fetchSOSAlerts()
      try {
        const audio = new Audio('/sos-alarm.mp3')
        audio.volume = 0.8
        audio.play().catch(() => {})
      } catch (e) {}
    })

    const pollingUnsubscribe = ws.on('escort_polling', (data) => {
      if (Array.isArray(data)) {
        setPollingVehicles(prev => {
          const updated = [...prev]
          data.forEach((item: any) => {
            const idx = updated.findIndex(v => v.vehicle_id === item.vehicle_id)
            if (idx >= 0) {
              updated[idx] = { ...updated[idx], ...item }
            }
          })
          return updated
        })
      } else if (data.vehicle_id) {
        setPollingVehicles(prev => prev.map(v =>
          v.vehicle_id === data.vehicle_id ? { ...v, ...data } : v
        ))
      }
    })

    const geoFenceUnsubscribe = ws.on('geo_fence_alert', (data) => {
      fetchGeoFenceAlerts()
      fetchGeoFenceStats()
      fetchStatistics()
      if (data && data.alert_level && data.alert_level >= 2) {
        try {
          const audio = new Audio('/alert.mp3')
          audio.volume = 0.7
          audio.play().catch(() => {})
        } catch (e) {}
      }
    })

    return () => {
      sosUnsubscribe()
      pollingUnsubscribe()
      geoFenceUnsubscribe()
    }
  }, [fetchStatistics, fetchSOSAlerts, fetchGeoFenceAlerts, fetchGeoFenceStats])

  useEffect(() => {
    const startPolling = async () => {
      try {
        const result = await escortApi.startPollingSession()
        if (result && (result as any).id) {
          setPollingSessionId((result as any).id)
        }
      } catch (e) {
        console.error('[Escort] startPollingSession error:', e)
      }
      await fetchPollingVehicles()
    }
    startPolling()

    return () => {
      if (pollingSessionId) {
        escortApi.endPollingSession(pollingSessionId, pollingCountRef.current).catch(() => {})
      }
      if (pollingTimerRef.current) {
        clearInterval(pollingTimerRef.current)
      }
    }
  }, [])

  useEffect(() => {
    if (pollingActive) {
      fetchPollingVehicles()
      pollingTimerRef.current = setInterval(() => {
        fetchPollingVehicles()
      }, pollingInterval * 1000)
    }
    return () => {
      if (pollingTimerRef.current) {
        clearInterval(pollingTimerRef.current)
        pollingTimerRef.current = null
      }
    }
  }, [pollingActive, pollingInterval, fetchPollingVehicles])

  useEffect(() => {
    const handleTabChange = (key: string) => {
      switch (key) {
        case 'tasks':
          fetchWaybills()
          break
        case 'polling':
          fetchPollingVehicles()
          break
        case 'alerts':
          fetchSOSAlerts()
          break
        case 'geo-fence':
          fetchGeoFenceAlerts()
          fetchGeoFenceStats()
          break
        case 'shifts':
          fetchShifts()
          break
        case 'videos':
          fetchVideoRecords()
          break
      }
    }
    if (activeTab) {
      handleTabChange(activeTab)
    }
  }, [activeTab, fetchWaybills, fetchPollingVehicles, fetchSOSAlerts, fetchGeoFenceAlerts, fetchGeoFenceStats, fetchShifts, fetchVideoRecords])

  const handleAddEvent = async (values: any) => {
    if (!selectedWaybill) return
    try {
      await escortApi.create({
        waybill_id: selectedWaybill.waybill_id || parseInt(selectedWaybill.id),
        event_type: values.event_type,
        location: values.location,
        remark: values.remark,
        longitude: selectedWaybill.current_lng,
        latitude: selectedWaybill.current_lat,
        risk_level: values.event_type === 'emergency' ? 'warning' : 'normal',
      })
      message.success('押运事件添加成功')
      fetchEvents(selectedWaybill)
      setAddEventModalVisible(false)
      addEventForm.resetFields()
    } catch (e) {
      message.error('添加失败，请重试')
    }
  }

  const handleExportReport = () => {
    if (!selectedWaybill) {
      message.warning('请先选择一个运单')
      return
    }
    escortApi.exportReport(selectedWaybill.waybill_id || parseInt(selectedWaybill.id))
      .then(() => {
        message.success(`正在生成运单 ${selectedWaybill.waybill_no} 的押运报告...`)
      })
      .catch(() => {
        message.success(`正在生成运单 ${selectedWaybill.waybill_no} 的押运报告...`)
      })
  }

  const handleSOSHandle = async (alert: any) => {
    try {
      await escortApi.handleSOS({ alert_id: alert.id, handle_note: '已受理' })
      message.success('已受理 SOS 报警')
      fetchSOSAlerts()
      fetchStatistics()
      setSosAlertModalVisible(false)
    } catch (e) {
      message.error('操作失败')
    }
  }

  const handleSOSResolve = async (alert: any) => {
    try {
      await escortApi.resolveSOS(alert.id, '已解决')
      message.success('SOS 报警已解决')
      fetchSOSAlerts()
      fetchStatistics()
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
      const vehicleId = intercomTarget.vehicle_id || intercomTarget.id
      await escortApi.sendIntercom({
        vehicle_id: vehicleId,
        message: intercomText,
        priority: 'normal',
      })
      message.success('喊话指令已发送')
      fetchIntercomLogs(vehicleId)
      setIntercomText('')
      setIntercomModalVisible(false)
    } catch (e) {
      message.error('发送失败')
    }
  }

  const handleOpenTrackPlayback = async (vehicle: any) => {
    setPlaybackTarget(vehicle)
    setTrackPlaybackVisible(true)
    setPlaybackIndex(0)
    setPlaybackMapReady(false)
    setTrackPlaybackData([])
    await fetchTrackPlayback(vehicle)
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
    if (trackPlaybackVisible && playbackMapContainerRef.current && !playbackMapReady && trackPlaybackData.length > 0) {
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
  }, [trackPlaybackVisible, playbackMapReady, trackPlaybackData.length])

  useEffect(() => {
    if (playbackMarkerRef.current && trackPlaybackData[playbackIndex]) {
      const point = trackPlaybackData[playbackIndex]
      const lng = point.longitude || point.lng
      const lat = point.latitude || point.lat
      playbackMarkerRef.current.setPosition([lng, lat])
    }
  }, [playbackIndex, trackPlaybackData])

  const initPlaybackMap = async () => {
    if (!playbackMapContainerRef.current || !trackPlaybackData.length) return
    try {
      const AMap: any = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: ['AMap.Scale', 'AMap.ToolBar', 'AMap.Polyline'],
      })
      const firstPoint = trackPlaybackData[0]
      const firstLng = firstPoint.longitude || firstPoint.lng
      const firstLat = firstPoint.latitude || firstPoint.lat
      const map = new AMap.Map(playbackMapContainerRef.current, {
        zoom: 14,
        center: [firstLng, firstLat],
        mapStyle: 'amap://styles/light',
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar())

      const path = trackPlaybackData.map(p => [p.longitude || p.lng, p.latitude || p.lat])
      const polyline = new AMap.Polyline({
        path,
        strokeColor: '#1677ff',
        strokeWeight: 4,
        strokeOpacity: 0.8,
      })
      polyline.setMap(map)
      playbackPolylineRef.current = polyline

      const marker = new AMap.Marker({
        position: [firstLng, firstLat],
        label: {
          content: `<div style="padding:4px 10px;background:#1677ff;color:#fff;border-radius:4px;font-size:12px;font-weight:600">
            ${playbackTarget?.vehicle_plate || playbackTarget?.waybill_no || '车辆'}
          </div>`,
          direction: 'top',
        },
      })
      marker.setMap(map)
      playbackMarkerRef.current = marker

      playbackMapInstanceRef.current = map
      setPlaybackMapReady(true)
    } catch (e) {
      console.error('[Escort] initPlaybackMap error:', e)
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
      console.error('[Escort] initEventMap error:', e)
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
  }, [eventDetailDrawer, eventMapReady])

  const filteredWaybills = waybills.filter(w => {
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
    { title: '排班日期', dataIndex: 'shift_date', width: 120, render: (t: any) => t ? dayjs(t).format('YYYY-MM-DD') : '-' },
    { title: '时间段', dataIndex: 'start_time', width: 140, render: (_: any, r: any) => `${r.start_time || '-'} - ${r.end_time || '-'}` },
    { title: '车辆数', dataIndex: 'vehicle_count', width: 80, render: (v: any) => v ?? 0 },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (s: string) => {
        const map: Record<string, { color: string; label: string }> = {
          scheduled: { color: 'blue', label: '已排班' },
          active: { color: 'processing', label: '进行中' },
          completed: { color: 'success', label: '已完成' },
          cancelled: { color: 'default', label: '已取消' },
        }
        return <Tag color={map[s]?.color}>{map[s]?.label || s || '-'}</Tag>
      },
    },
    { title: '调度员', dataIndex: 'dispatcher_name', width: 100 },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t: any) => t ? formatDateTime(t) : '-' },
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
          loading={waybillsLoading}
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
                          <Tag color={statusMap[waybill.status]?.color || 'default'} style={{ margin: 0 }}>
                            {statusMap[waybill.status]?.label || waybill.status}
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
          loading={eventsLoading}
        >
          <div style={{ overflowY: 'auto', flex: 1, padding: '16px 24px' }}>
            {!selectedWaybill ? (
              <Empty
                description={<Space direction="vertical" align="center"><Text>请从左侧选择一个押运任务</Text><Text type="secondary">选择后查看该运单的完整押运事件时间线</Text></Space>}
                style={{ marginTop: 100 }}
              />
            ) : events.length === 0 ? (
              <Empty description="暂无押运事件" style={{ marginTop: 60 }} />
            ) : (
              <Timeline
                mode="left"
                style={{ paddingLeft: 0 }}
                items={events.map((event, idx) => {
                  const et = eventTypeMap[event.type] || eventTypeMap.waypoint
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
                {pollingLoading && <Text type="secondary" style={{ fontSize: 12 }}><LoadingOutlined spin /> 更新中...</Text>}
              </Space>
            </Col>
            <Col xs={24} md={8} style={{ textAlign: 'right' }}>
              <Space>
                <Button icon={<ReloadOutlined />} size="small" onClick={() => fetchPollingVehicles()}>刷新</Button>
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
      {pollingVehicles.length === 0 ? (
        <Col xs={24}>
          <Empty description="暂无需要轮询的车辆" style={{ padding: 60 }} />
        </Col>
      ) : (
        pollingVehicles.map(vehicle => (
          <Col xs={24} sm={12} lg={8} xl={6} key={vehicle.vehicle_id || vehicle.id}>
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
                  {vehicle.cover_url || vehicle.frame_url ? (
                    <Image
                      src={vehicle.cover_url || vehicle.frame_url}
                      alt=""
                      style={{ width: '100%', height: 180, objectFit: 'cover', borderRadius: '12px 12px 0 0' }}
                      preview={{ mask: <EyeOutlined /> }}
                    />
                  ) : (
                    <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                      <CameraOutlined style={{ fontSize: 36, opacity: 0.3 }} />
                      <div style={{ fontSize: 12, marginTop: 8 }}>
                        车载摄像头 - {vehicle.vehicle_plate || vehicle.plate}
                      </div>
                      <div style={{ fontSize: 10, marginTop: 4, color: '#52c41a' }}>
                        ● LIVE · {dayjs(vehicle.last_frame_time || vehicle.last_update || new Date()).format('HH:mm:ss')}
                      </div>
                    </div>
                  )}
                  {vehicle.status === 'resting' && (
                    <Tag color="orange" style={{ position: 'absolute', top: 8, right: 8 }}>休息中</Tag>
                  )}
                  {(vehicle.status === 'driving' || vehicle.status === 'transit') && (
                    <Tag color="green" style={{ position: 'absolute', top: 8, right: 8 }}>行驶中</Tag>
                  )}
                  {vehicle.status === 'stopped' && (
                    <Tag color="default" style={{ position: 'absolute', top: 8, right: 8 }}>已停靠</Tag>
                  )}
                </div>
              }
            >
              <div style={{ padding: 12 }}>
                <Space direction="vertical" size={6} style={{ width: '100%' }}>
                  <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                    <Tag color="blue" style={{ fontSize: 13, margin: 0 }}>{vehicle.vehicle_plate || vehicle.plate}</Tag>
                    <Text strong style={{ fontSize: 16, color: (vehicle.speed || 0) > 70 ? '#ff4d4f' : '#1677ff' }}>
                      {vehicle.speed || 0} <Text type="secondary" style={{ fontSize: 11 }}>km/h</Text>
                    </Text>
                  </Space>
                  <div style={{ fontSize: 12 }}>
                    <UserOutlined style={{ color: '#fa8c16' }} /> 司机: {vehicle.driver_name || '-'}
                  </div>
                  <div style={{ fontSize: 12 }}>
                    <SafetyCertificateOutlined style={{ color: '#52c41a' }} /> 货物: {vehicle.danger_goods || vehicle.goods_name || '-'}
                  </div>
                  <div style={{ fontSize: 12 }}>
                    <span style={{ color: (vehicle.driver_status || '正常') === '正常' ? '#52c41a' : '#faad14' }}>●</span> 驾驶员状态: {vehicle.driver_status || '正常'}
                  </div>
                  {vehicle.location && (
                    <div style={{ fontSize: 12 }}>
                      <EnvironmentOutlined style={{ color: '#1677ff' }} /> {vehicle.location}
                    </div>
                  )}
                  <Divider style={{ margin: '8px 0' }} />
                  <Space size={4}>
                    <Button size="small" icon={<EyeOutlined />} block onClick={(e) => {
                      e.stopPropagation()
                      if (vehicle.video_url || vehicle.live_url) {
                        window.open(vehicle.video_url || vehicle.live_url, '_blank')
                      } else {
                        message.info('查看实时画面')
                      }
                    }}>查看</Button>
                    <Button size="small" type="primary" icon={<PhoneOutlined />} block onClick={(e) => {
                      e.stopPropagation()
                      setIntercomTarget(vehicle)
                      setIntercomModalVisible(true)
                    }}>喊话</Button>
                    <Button size="small" icon={<HistoryOutlined />} block onClick={(e) => {
                      e.stopPropagation()
                      handleOpenTrackPlayback(vehicle)
                    }}>轨迹</Button>
                  </Space>
                </Space>
              </div>
            </Card>
          </Col>
        ))
      )}
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
                    <Text strong>{t || r.vehicle?.plate || '-'}</Text>
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
                  return <Tag color={map[s]?.color}>{map[s]?.label || s || '-'}</Tag>
                },
              },
              { title: '处理人', dataIndex: 'handler_name', width: 100, render: (t) => t || '-' },
              { title: '报警时间', dataIndex: 'created_at', width: 160, render: (t) => t ? formatDateTime(t) : '-' },
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
            dataSource={sosAlerts}
            search={false}
            loading={alertsLoading}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => fetchSOSAlerts()}>刷新</Button>,
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
            dataSource={shifts}
            search={false}
            loading={shiftsLoading}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => fetchShifts()}>刷新</Button>,
              <Button key="add" icon={<PlusOutlined />} type="primary" onClick={() => message.info('请在排班管理模块创建')}>创建排班</Button>,
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
                  return <Tag color={map[t]?.color}>{map[t]?.label || t || '-'}</Tag>
                },
              },
              { title: '开始时间', dataIndex: 'start_time', width: 160, render: (t) => t ? formatDateTime(t) : '-' },
              { title: '结束时间', dataIndex: 'end_time', width: 160, render: (t) => t ? formatDateTime(t) : '-' },
              { title: '时长(分钟)', dataIndex: 'duration_minutes', width: 100 },
              { title: '查看次数', dataIndex: 'view_count', width: 100 },
              { title: '过期时间', dataIndex: 'expire_at', width: 160,
                render: (t) => t ? (
                  <Text type={dayjs(t).isBefore(dayjs().add(7, 'day')) ? 'danger' : 'secondary'}>
                    {formatDateTime(t)}
                    {dayjs(t).isBefore(dayjs().add(7, 'day')) && <Tag color="red" style={{ marginLeft: 8 }}>即将过期</Tag>}
                  </Text>
                ) : '-'
              },
              {
                title: '操作', width: 180,
                render: (_, r) => (
                  <Space size={4}>
                    <Button size="small" type="primary" icon={<EyeOutlined />} onClick={() => {
                      escortApi.viewVideoRecord(r.id).catch(() => {})
                      if (r.file_url || r.video_url) {
                        window.open(r.file_url || r.video_url, '_blank')
                      } else {
                        message.success(`查看录像 ${r.id}`)
                      }
                    }}>查看</Button>
                    <Button size="small" icon={<DownloadOutlined />} onClick={() => {
                      if (r.file_url || r.download_url) {
                        window.open(r.file_url || r.download_url, '_blank')
                      } else {
                        message.success('开始下载')
                      }
                    }}>下载</Button>
                  </Space>
                ),
              },
            ]}
            dataSource={videoRecords}
            search={false}
            loading={videosLoading}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            rowKey="id"
            toolBarRender={() => [
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => fetchVideoRecords()}>刷新</Button>,
              <Alert key="tip" type="info" showIcon message="视频记录云端存储90天，过期自动清理" style={{ marginLeft: 16 }} />,
            ]}
          />
        </ProCard>
      </Col>
    </Row>
  )

  const geoFenceStatusMap: Record<string, { color: string; label: string }> = {
    pending: { color: 'red', label: '待确认' },
    confirmed: { color: 'blue', label: '已确认' },
    escalated: { color: 'orange', label: '已上报' },
    resolved: { color: 'green', label: '已解决' },
  }

  const renderGeoFenceTab = () => (
    <Row gutter={16}>
      <Col xs={24}>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col xs={24} md={6}>
            <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                  <WarningOutlined style={{ color: '#ff4d4f', fontSize: 18 }} />
                  <Text type="secondary">待确认告警</Text>
                </Space>
                <Statistic value={geoFenceStats.pending_alerts} valueStyle={{ color: '#ff4d4f', fontSize: 28 }} />
              </Space>
            </ProCard>
          </Col>
          <Col xs={24} md={6}>
            <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                  <EnvironmentOutlined style={{ color: '#faad14', fontSize: 18 }} />
                  <Text type="secondary">今日偏航</Text>
                </Space>
                <Statistic value={geoFenceStats.today_alerts} valueStyle={{ color: '#faad14', fontSize: 28 }} />
              </Space>
            </ProCard>
          </Col>
          <Col xs={24} md={6}>
            <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                  <DashboardOutlined style={{ color: '#eb2f96', fontSize: 18 }} />
                  <Text type="secondary">自动上报调度</Text>
                </Space>
                <Statistic value={geoFenceStats.auto_reported_count} valueStyle={{ color: '#eb2f96', fontSize: 28 }} />
              </Space>
            </ProCard>
          </Col>
          <Col xs={24} md={6}>
            <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                  <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 18 }} />
                  <Text type="secondary">已解决</Text>
                </Space>
                <Statistic value={geoFenceStats.resolved_alerts} valueStyle={{ color: '#52c41a', fontSize: 28 }} />
              </Space>
            </ProCard>
          </Col>
        </Row>
        <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
          <ProTable<GeoFenceAlertItem>
            columns={[
              { title: '告警编号', dataIndex: 'alert_no', width: 140, render: (t) => <Text copyable style={{ fontSize: 12 }}>{t}</Text> },
              {
                title: '车辆', dataIndex: 'plate_number', width: 110,
                render: (t, r) => (
                  <Space>
                    <CarOutlined style={{ color: '#1677ff' }} />
                    <Text strong>{t}</Text>
                  </Space>
                ),
              },
              { title: '押运员', dataIndex: 'escort_name', width: 90 },
              { title: '运单号', dataIndex: 'waybill_no', width: 130, render: (t) => t ? <Text style={{ fontSize: 12 }}>{t}</Text> : '-' },
              {
                title: '偏离距离', width: 110,
                render: (_, r) => (
                  <Space>
                    <Text strong style={{ color: r.distance_from_route_meters > 1000 ? '#ff4d4f' : r.distance_from_route_meters > 700 ? '#faad14' : '#fa8c16', fontSize: 14 }}>
                      {r.distance_from_route_meters?.toFixed(0)}
                    </Text>
                    <Text type="secondary" style={{ fontSize: 11 }}>米 / {r.threshold_meters}米</Text>
                  </Space>
                ),
              },
              {
                title: '告警级别', dataIndex: 'alert_level', width: 90,
                render: (l) => {
                  const colors = ['', '#faad14', '#fa8c16', '#ff4d4f']
                  const labels = ['', '一般', '关注', '紧急']
                  return <Tag color={colors[l] || 'default'}>{labels[l] || '-'}</Tag>
                },
              },
              {
                title: '当日累计', dataIndex: 'daily_deviate_count', width: 90,
                render: (n) => {
                  const count = n || 0
                  const color = count >= 3 ? 'red' : count >= 2 ? 'orange' : count >= 1 ? 'blue' : 'default'
                  return (
                    <Space>
                      <Tag color={color}>{count} 次</Tag>
                      {count >= 3 && <Tooltip title="当日累计达3次，已自动上报调度"><Badge status="error" /></Tooltip>}
                    </Space>
                  )
                },
              },
              {
                title: '确认原因', dataIndex: 'deviate_reason', width: 90,
                render: (t) => t === 'detour' ? <Tag color="blue">绕路</Tag> : t === 'deviate' ? <Tag color="red">偏航</Tag> : <Tag color="default">未确认</Tag>,
              },
              {
                title: '上报调度', dataIndex: 'reported_to_dispatch', width: 90,
                render: (t, r) => t
                  ? <Space><Badge status="error" text={<Tag color="orange">已上报 {r.auto_reported ? '(自动)' : ''}</Tag>}</Space></Space>
                  : <Text type="secondary">未上报</Text>,
              },
              {
                title: '状态', dataIndex: 'status', width: 90,
                render: (s) => {
                  const m = geoFenceStatusMap[s] || { color: 'default', label: s }
                  return <Tag color={m.color}>{m.label}</Tag>
                },
              },
              { title: '位置', dataIndex: 'address', width: 200, ellipsis: true, render: (t) => t || '-' },
              { title: '发生时间', dataIndex: 'created_at', width: 160, render: (t) => t ? formatDateTime(t) : '-' },
              {
                title: '操作', width: 240,
                render: (_, r) => (
                  <Space size={4} wrap>
                    <Button size="small" icon={<EyeOutlined />} onClick={() => handleOpenGeoFenceDetail(r)}>详情</Button>
                    {(r.status === 'pending') && (
                      <Button size="small" type="primary" onClick={() => setGeoFenceConfirmModal({ visible: true, alert: r })}>
                        确认原因
                      </Button>
                    )}
                    {(r.status === 'escalated' || r.status === 'confirmed') && (
                      <Button size="small" type="primary" danger onClick={() => setGeoFenceResolveModal({ visible: true, alert: r })}>
                        调度处理
                      </Button>
                    )}
                  </Space>
                ),
              },
            ]}
            dataSource={geoFenceAlerts}
            search={false}
            loading={geoFenceLoading}
            pagination={{ pageSize: 10, showSizeChanger: true }}
            rowKey="id"
            toolBarRender={() => [
              <Select
                key="status"
                allowClear
                placeholder="状态筛选"
                style={{ width: 140 }}
                value={geoFenceFilterStatus}
                onChange={(v) => { setGeoFenceFilterStatus(v as any) }}
                size="middle"
              >
                {Object.entries(geoFenceStatusMap).map(([k, v]) => (
                  <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
                ))}
              </Select>,
              <Button key="reload" icon={<ReloadOutlined />} onClick={() => { fetchGeoFenceAlerts(); fetchGeoFenceStats() }}>刷新</Button>,
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
              <Tag color="warning">待确认偏航 {geoFenceStats.pending_alerts}</Tag>
              <Tag color="success">今日排班 {statistics.active_shifts}</Tag>
              <Tag color="blue">视频记录 {statistics.total_videos}</Tag>
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
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button icon={<PlusOutlined />} type="primary" ghost disabled={!selectedWaybill} onClick={() => setAddEventModalVisible(true)}>添加事件</Button>
              <Button icon={<ExportOutlined />} disabled={!selectedWaybill} onClick={handleExportReport}>押运报告</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { fetchWaybills(); fetchStatistics() }}>刷新</Button>
            </Space>
          </Col>
        </Row>
      </ProCard>
      <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          type="card"
          style={{ marginBottom: 0 }}
          items={[
            {
              key: 'tasks',
              label: <Space><DashboardOutlined /> 押运任务</Space>,
              children: renderTasksTab(),
            },
            {
              key: 'polling',
              label: <Space><VideoCameraOutlined /> 视频轮询 <Badge dot status="processing" offset={[4, -2]} /></Space>,
              children: renderPollingTab(),
            },
            {
              key: 'alerts',
              label: <Space><BellOutlined /> SOS 报警 {statistics.pending_sos > 0 && <Badge count={statistics.pending_sos} color="red" style={{ boxShadow: 'none' }} />}</Space>,
              children: renderAlertsTab(),
            },
            {
              key: 'geo-fence',
              label: <Space><EnvironmentOutlined /> 电子围栏 {geoFenceStats.pending_alerts > 0 && <Badge count={geoFenceStats.pending_alerts} color="orange" style={{ boxShadow: 'none' }} />}</Space>,
              children: renderGeoFenceTab(),
            },
            {
              key: 'shifts',
              label: <Space><ScheduleOutlined /> 排班管理</Space>,
              children: renderShiftsTab(),
            },
            {
              key: 'videos',
              label: <Space><HistoryOutlined /> 视频记录</Space>,
              children: renderVideosTab(),
            },
          ]}
        />
      </ProCard>

      <Modal
        title={
          <Space>
            <BellOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
            <Text strong type="danger" style={{ fontSize: 18 }}>⚠️ 紧急 SOS 报警</Text>
          </Space>
        }
        open={sosAlertModalVisible}
        onCancel={() => {}}
        footer={
          <Space>
            <Button danger type="primary" size="large" onClick={() => handleSOSHandle(currentSosAlert)}>
              立即受理
            </Button>
            <Button size="large" onClick={() => handleSOSResolve(currentSosAlert)}>
              标记已解决
            </Button>
          </Space>
        }
        maskClosable={false}
        keyboard={false}
        closable={false}
        style={{ top: 80 }}
      >
        {currentSosAlert && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type="error"
              showIcon
              message="司机已按下驾驶室紧急按钮，请立即采取措施！"
              description={`车辆 ${currentSosAlert.vehicle_plate} 于 ${formatDateTime(currentSosAlert.created_at)} 触发紧急报警。`}
            />
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="报警ID">{currentSosAlert.id}</Descriptions.Item>
              <Descriptions.Item label="报警类型"><Tag color="red">{currentSosAlert.sos_type || '紧急报警'}</Tag></Descriptions.Item>
              <Descriptions.Item label="车牌号">{currentSosAlert.vehicle_plate}</Descriptions.Item>
              <Descriptions.Item label="司机">{currentSosAlert.driver_name}</Descriptions.Item>
              <Descriptions.Item label="当前位置" span={2}>{currentSosAlert.location || '-'}</Descriptions.Item>
              <Descriptions.Item label="报警描述" span={2}>{currentSosAlert.description || '司机触发驾驶室紧急按钮'}</Descriptions.Item>
            </Descriptions>
            <Space style={{ width: '100%', justifyContent: 'center' }}>
              <Button icon={<PhoneOutlined />} type="primary" onClick={() => { setIntercomTarget(currentSosAlert); setIntercomModalVisible(true); setSosAlertModalVisible(false) }}>立即喊话</Button>
              <Button icon={<VideoCameraOutlined />} onClick={() => {
                if (currentSosAlert.live_url || currentSosAlert.video_url) {
                  window.open(currentSosAlert.live_url || currentSosAlert.video_url, '_blank')
                }
              }}>查看实时画面</Button>
            </Space>
          </Space>
        )}
      </Modal>

      <Modal
        title={
          <Space>
            <PhoneOutlined style={{ color: '#1677ff' }} />
            <Text strong>语音喊话指令</Text>
            {intercomTarget && <Tag color="blue">{intercomTarget.vehicle_plate || intercomTarget.plate || '-'}</Tag>}
          </Space>
        }
        open={intercomModalVisible}
        onCancel={() => setIntercomModalVisible(false)}
        footer={
          <Space>
            <Button onClick={() => setIntercomModalVisible(false)}>取消</Button>
            <Button type="primary" icon={<SendOutlined />} onClick={handleSendIntercom}>发送指令</Button>
          </Space>
        }
        width={600}
      >
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Text type="secondary" style={{ fontSize: 12 }}>快捷指令：</Text>
          <Space wrap>
            {[
              '前方检查点请减速',
              '请保持安全车距',
              '驾驶员请勿疲劳驾驶',
              '请遵守限速规定',
              '前方路况复杂请注意',
              '请开启双闪警示灯',
              '服务区请停车休息',
              '已收到报警，正在处理',
            ].map((cmd, idx) => (
              <Tag
                key={idx}
                color="blue"
                style={{ cursor: 'pointer', fontSize: 12, padding: '4px 10px' }}
                onClick={() => setIntercomText(cmd)}
              >
                {cmd}
              </Tag>
            ))}
          </Space>
          <TextArea
            rows={4}
            placeholder="请输入喊话内容..."
            value={intercomText}
            onChange={e => setIntercomText(e.target.value)}
            showCount
            maxLength={100}
          />
          <Divider orientation="left" style={{ margin: '8px 0', fontSize: 12 }}>最近喊话记录</Divider>
          {intercomLogs.length > 0 ? (
            <List
              size="small"
              dataSource={intercomLogs.slice(0, 5)}
              renderItem={(item: any) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={<SoundOutlined />} style={{ background: '#1677ff' }} />}
                    title={
                      <Space>
                        <Text strong style={{ fontSize: 12 }}>{item.message || item.content}</Text>
                        <Tag color={item.priority === 'urgent' ? 'red' : 'default'} style={{ fontSize: 10 }}>
                          {item.priority === 'urgent' ? '紧急' : '普通'}
                        </Tag>
                      </Space>
                    }
                    description={<Text type="secondary" style={{ fontSize: 11 }}>{formatDateTime(item.created_at || item.send_time)}</Text>}
                  />
                </List.Item>
              )}
            />
          ) : (
            <Empty description="暂无喊话记录" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 16 }} />
          )}
        </Space>
      </Modal>

      <Modal
        title={
          <Space>
            <WarningOutlined style={{ color: '#fa8c16', fontSize: 20 }} />
            <Text strong style={{ fontSize: 16 }}>偏航告警确认</Text>
            {geoFenceConfirmModal.alert && <Tag color="orange">{geoFenceConfirmModal.alert.plate_number}</Tag>}
          </Space>
        }
        open={geoFenceConfirmModal.visible}
        onCancel={() => { setGeoFenceConfirmModal({ visible: false, alert: null }); geoFenceConfirmForm.resetFields() }}
        footer={
          <Space>
            <Button onClick={() => { setGeoFenceConfirmModal({ visible: false, alert: null }); geoFenceConfirmForm.resetFields() }}>取消</Button>
            <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => geoFenceConfirmForm.submit()}>确认提交</Button>
          </Space>
        }
        width={560}
      >
        {geoFenceConfirmModal.alert && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type="warning"
              showIcon
              message={`车辆 ${geoFenceConfirmModal.alert.plate_number} 偏离预设路线 ${geoFenceConfirmModal.alert.distance_from_route_meters?.toFixed(0)} 米`}
              description={`告警编号: ${geoFenceConfirmModal.alert.alert_no} · 发生时间: ${formatDateTime(geoFenceConfirmModal.alert.created_at)}`}
            />
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="偏离距离">
                <Text type="danger" strong>{geoFenceConfirmModal.alert.distance_from_route_meters?.toFixed(0)} 米</Text>
                <Text type="secondary"> / {geoFenceConfirmModal.alert.threshold_meters}米阈值</Text>
              </Descriptions.Item>
              <Descriptions.Item label="告警级别">
                <Tag color={geoFenceConfirmModal.alert.alert_level === 3 ? 'red' : geoFenceConfirmModal.alert.alert_level === 2 ? 'orange' : 'blue'}>
                  {geoFenceConfirmModal.alert.alert_level === 3 ? '紧急' : geoFenceConfirmModal.alert.alert_level === 2 ? '关注' : '一般'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="当日累计" span={2}>
                <Tag color={geoFenceConfirmModal.alert.daily_deviate_count >= 3 ? 'red' : geoFenceConfirmModal.alert.daily_deviate_count >= 2 ? 'orange' : 'blue'}>
                  {geoFenceConfirmModal.alert.daily_deviate_count} 次
                </Tag>
                {geoFenceConfirmModal.alert.daily_deviate_count >= 2 && (
                  <Text type="danger" style={{ marginLeft: 8, fontSize: 12 }}>
                    ⚠️ 当日累计偏航已达{geoFenceConfirmModal.alert.daily_deviate_count}次，超过3次将自动上报调度中心
                  </Text>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="位置" span={2}>{geoFenceConfirmModal.alert.address || '-'}</Descriptions.Item>
              <Descriptions.Item label="押运员">{geoFenceConfirmModal.alert.escort_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="司机">{geoFenceConfirmModal.alert.driver_name || '-'}</Descriptions.Item>
            </Descriptions>
            <Form form={geoFenceConfirmForm} layout="vertical" onFinish={handleGeoFenceConfirm}>
              <Form.Item
                label="确认类型"
                name="confirm_type"
                rules={[{ required: true, message: '请选择偏航原因类型' }]}
                style={{ marginBottom: 12 }}
              >
                <Select placeholder="请选择原因类型">
                  <Option value="detour">
                    <Tag color="blue">合理绕路</Tag>
                    <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>因施工、拥堵、检查点等合理原因绕行</Text>
                  </Option>
                  <Option value="deviate">
                    <Tag color="red">异常偏航</Tag>
                    <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>不明原因偏离路线，需进一步核实</Text>
                  </Option>
                </Select>
              </Form.Item>
              <Form.Item
                label="原因详情"
                name="reason_detail"
                rules={[{ required: true, message: '请填写具体原因' }]}
                style={{ marginBottom: 12 }}
              >
                <TextArea rows={3} placeholder="请填写具体的偏航原因，例如：前方高速事故绕行G108国道" showCount maxLength={200} />
              </Form.Item>
              <Form.Item label="备注说明" name="note" style={{ marginBottom: 0 }}>
                <TextArea rows={2} placeholder="可选，其他补充说明" showCount maxLength={100} />
              </Form.Item>
            </Form>
          </Space>
        )}
      </Modal>

      <Modal
        title={
          <Space>
            <DashboardOutlined style={{ color: '#eb2f96', fontSize: 20 }} />
            <Text strong style={{ fontSize: 16 }}>调度处理偏航告警</Text>
            {geoFenceResolveModal.alert && <Tag color="orange">{geoFenceResolveModal.alert.plate_number}</Tag>}
          </Space>
        }
        open={geoFenceResolveModal.visible}
        onCancel={() => { setGeoFenceResolveModal({ visible: false, alert: null }); geoFenceResolveForm.resetFields() }}
        footer={
          <Space>
            <Button onClick={() => { setGeoFenceResolveModal({ visible: false, alert: null }); geoFenceResolveForm.resetFields() }}>取消</Button>
            <Button type="primary" danger icon={<CheckCircleOutlined />} onClick={() => geoFenceResolveForm.submit()}>标记已处理</Button>
          </Space>
        }
        width={560}
      >
        {geoFenceResolveModal.alert && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type={geoFenceResolveModal.alert.reported_to_dispatch ? 'error' : 'warning'}
              showIcon
              message={geoFenceResolveModal.alert.reported_to_dispatch ? '该偏航告警已自动上报至调度中心，请及时处理' : '处理该偏航告警'}
              description={`${geoFenceResolveModal.alert.plate_number} 偏离 ${geoFenceResolveModal.alert.distance_from_route_meters?.toFixed(0)}米 · 当日累计 ${geoFenceResolveModal.alert.daily_deviate_count} 次`}
            />
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="告警编号" span={2}>{geoFenceResolveModal.alert.alert_no}</Descriptions.Item>
              <Descriptions.Item label="偏离距离">
                <Text type="danger" strong>{geoFenceResolveModal.alert.distance_from_route_meters?.toFixed(0)} 米</Text>
              </Descriptions.Item>
              <Descriptions.Item label="确认原因">
                {geoFenceResolveModal.alert.deviate_reason === 'detour'
                  ? <Tag color="blue">绕路</Tag>
                  : geoFenceResolveModal.alert.deviate_reason === 'deviate'
                    ? <Tag color="red">偏航</Tag>
                    : <Tag color="default">未确认</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="确认备注" span={2}>{geoFenceResolveModal.alert.confirm_note || geoFenceResolveModal.alert.resolved_note || '-'}</Descriptions.Item>
              <Descriptions.Item label="上报调度">
                {geoFenceResolveModal.alert.reported_to_dispatch ? <Badge status="error" text="是" /> : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="押运员">{geoFenceResolveModal.alert.escort_name || '-'}</Descriptions.Item>
            </Descriptions>
            <Form form={geoFenceResolveForm} layout="vertical" onFinish={handleGeoFenceResolve}>
              <Form.Item
                label="处理说明"
                name="resolved_note"
                rules={[{ required: true, message: '请填写处理说明' }]}
                style={{ marginBottom: 0 }}
              >
                <TextArea rows={4} placeholder="请填写调度处理说明，例如：已联系押运员核实情况，确认前方封路绕行，通知司机尽快返回规划路线" showCount maxLength={300} />
              </Form.Item>
            </Form>
          </Space>
        )}
      </Modal>

      <Drawer
        title={
          <Space>
            <EnvironmentOutlined style={{ color: '#1677ff' }} />
            <Text strong>偏航告警详情</Text>
            {geoFenceDetail && <Tag color="orange">{geoFenceDetail.alert_no}</Tag>}
          </Space>
        }
        placement="right"
        width={600}
        onClose={() => setGeoFenceDetail(null)}
        open={!!geoFenceDetail}
      >
        {geoFenceDetail && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Card size="small">
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic title="偏离距离" value={geoFenceDetail.distance_from_route_meters?.toFixed(0)} suffix="米" valueStyle={{ color: '#ff4d4f' }} />
                </Col>
                <Col span={12}>
                  <Statistic title="当日累计" value={geoFenceDetail.daily_deviate_count} suffix="次" valueStyle={{ color: geoFenceDetail.daily_deviate_count >= 3 ? '#ff4d4f' : '#fa8c16' }} />
                </Col>
              </Row>
            </Card>
            <Descriptions column={2} bordered size="small" title="基本信息">
              <Descriptions.Item label="告警编号" span={2}>{geoFenceDetail.alert_no}</Descriptions.Item>
              <Descriptions.Item label="状态">
                {(() => {
                  const m = geoFenceStatusMap[geoFenceDetail.status] || { color: 'default', label: geoFenceDetail.status }
                  return <Tag color={m.color}>{m.label}</Tag>
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="告警级别">
                <Tag color={geoFenceDetail.alert_level === 3 ? 'red' : geoFenceDetail.alert_level === 2 ? 'orange' : 'blue'}>
                  {geoFenceDetail.alert_level === 3 ? '紧急' : geoFenceDetail.alert_level === 2 ? '关注' : '一般'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="车牌号">{geoFenceDetail.plate_number}</Descriptions.Item>
              <Descriptions.Item label="运单号">{geoFenceDetail.waybill_no || '-'}</Descriptions.Item>
              <Descriptions.Item label="押运员">{geoFenceDetail.escort_name}</Descriptions.Item>
              <Descriptions.Item label="司机">{geoFenceDetail.driver_name}</Descriptions.Item>
              <Descriptions.Item label="阈值">{geoFenceDetail.threshold_meters} 米</Descriptions.Item>
              <Descriptions.Item label="偏离距离" span={2}>
                <Text type="danger" strong>{geoFenceDetail.distance_from_route_meters?.toFixed(0)} 米</Text>
              </Descriptions.Item>
              <Descriptions.Item label="当前位置" span={2}>{geoFenceDetail.address || '-'}</Descriptions.Item>
              <Descriptions.Item label="发生时间" span={2}>{formatDateTime(geoFenceDetail.created_at)}</Descriptions.Item>
            </Descriptions>
            <Descriptions column={2} bordered size="small" title="确认与处理">
              <Descriptions.Item label="确认原因">
                {geoFenceDetail.deviate_reason === 'detour'
                  ? <Tag color="blue">绕路</Tag>
                  : geoFenceDetail.deviate_reason === 'deviate'
                    ? <Tag color="red">偏航</Tag>
                    : <Tag color="default">未确认</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="上报调度">
                {geoFenceDetail.reported_to_dispatch
                  ? <Space><Badge status="error" text={<Tag color="orange">已上报 {geoFenceDetail.auto_reported ? '(自动)' : ''}</Tag>}</Space>
                  : <Text type="secondary">未上报</Text>}
              </Descriptions.Item>
              <Descriptions.Item label="确认时间">{geoFenceDetail.confirmed_at ? formatDateTime(geoFenceDetail.confirmed_at) : '-'}</Descriptions.Item>
              <Descriptions.Item label="处理时间">{geoFenceDetail.resolved_at ? formatDateTime(geoFenceDetail.resolved_at) : '-'}</Descriptions.Item>
              <Descriptions.Item label="确认备注" span={2}>{geoFenceDetail.confirm_note || '-'}</Descriptions.Item>
              <Descriptions.Item label="处理说明" span={2}>{geoFenceDetail.resolved_note || '-'}</Descriptions.Item>
            </Descriptions>
            <Divider orientation="left" style={{ margin: '8px 0', fontSize: 12 }}>确认处理记录</Divider>
            {geoFenceLogsLoading ? (
              <div style={{ textAlign: 'center', padding: 24 }}><LoadingOutlined spin /> 加载中...</div>
            ) : geoFenceConfirmLogs.length === 0 ? (
              <Empty description="暂无确认记录" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 16 }} />
            ) : (
              <List
                size="small"
                dataSource={geoFenceConfirmLogs}
                renderItem={(item: any) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar
                          icon={item.confirm_type === 'detour' ? <EnvironmentOutlined /> : item.confirm_type === 'deviate' ? <WarningOutlined /> : <CheckCircleOutlined />}
                          style={{
                            background: item.confirm_type === 'detour' ? '#1677ff'
                              : item.confirm_type === 'deviate' ? '#ff4d4f'
                                : item.confirm_type === 'resolve' ? '#52c41a' : '#8c8c8c'
                          }}
                        />
                      }
                      title={
                        <Space>
                          <Text strong style={{ fontSize: 12 }}>
                            {item.confirm_type === 'detour' ? '确认：合理绕路'
                              : item.confirm_type === 'deviate' ? '确认：异常偏航'
                                : item.confirm_type === 'resolve' ? '调度：处理完成'
                                  : (item.confirm_type || '操作记录')}
                          </Text>
                          {item.confirmed_role && <Tag color="default" style={{ fontSize: 10 }}>{item.confirmed_role}</Tag>}
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size={2} style={{ width: '100%' }}>
                          <Text type="secondary" style={{ fontSize: 11 }}>
                            {formatDateTime(item.confirmed_at || item.created_at)} · {item.confirmed_name || item.confirmed_by || '系统'}
                          </Text>
                          {(item.reason_detail || item.note) && <Text style={{ fontSize: 12 }}>{item.reason_detail || item.note}</Text>}
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
            <Space style={{ justifyContent: 'flex-end', width: '100%' }}>
              {geoFenceDetail.status === 'pending' && (
                <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => setGeoFenceConfirmModal({ visible: true, alert: geoFenceDetail })}>
                  确认原因
                </Button>
              )}
              {(geoFenceDetail.status === 'escalated' || geoFenceDetail.status === 'confirmed') && (
                <Button type="primary" danger icon={<DashboardOutlined />} onClick={() => setGeoFenceResolveModal({ visible: true, alert: geoFenceDetail })}>
                  调度处理
                </Button>
              )}
            </Space>
          </Space>
        )}
      </Drawer>

      <Drawer
        title={
          <Space>
            <MapOutlined style={{ color: '#1677ff' }} />
            <Text strong>押运轨迹回放</Text>
            {playbackTarget && <Tag color="blue">{playbackTarget.vehicle_plate || playbackTarget.waybill_no || '-'}</Tag>}
          </Space>
        }
        placement="right"
        width={900}
        onClose={() => { setTrackPlaybackVisible(false); stopPlayback() }}
        open={trackPlaybackVisible}
      >
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          {playbackLoading ? (
            <div style={{ textAlign: 'center', padding: 60 }}><LoadingOutlined spin /> <Text type="secondary">加载轨迹数据中...</Text></div>
          ) : trackPlaybackData.length === 0 ? (
            <Empty description="暂无轨迹数据" style={{ padding: 60 }} />
          ) : (
            <>
              <Card size="small" bodyStyle={{ padding: 12 }}>
                <Row gutter={16} align="middle">
                  <Col flex="none">
                    <Button
                      icon={playbackPlaying ? <PauseCircleOutlined style={{ fontSize: 24 }} /> : <PlayCircleOutlined style={{ fontSize: 24 }} />}
                      type="text"
                      onClick={playbackPlaying ? stopPlayback : startPlayback}
                    />
                  </Col>
                  <Col flex="auto">
                    <Progress
                      percent={Math.round((playbackIndex / Math.max(1, trackPlaybackData.length - 1)) * 100)}
                      showInfo
                      strokeColor={{ '0%': '#1677ff', '100%': '#52c41a' }}
                    />
                  </Col>
                  <Col flex="none" style={{ textAlign: 'right' }}>
                    <Space direction="vertical" size={0} style={{ textAlign: 'center' }}>
                      <Text strong style={{ fontSize: 14 }}>{dayjs(trackPlaybackData[playbackIndex]?.recorded_at || trackPlaybackData[playbackIndex]?.time || new Date()).format('HH:mm:ss')}</Text>
                      <Text type="secondary" style={{ fontSize: 11 }}>{playbackIndex + 1} / {trackPlaybackData.length}</Text>
                    </Space>
                  </Col>
                </Row>
              </Card>
              <Row gutter={16}>
                <Col span={8}>
                  <Card size="small" title="运行统计" bodyStyle={{ padding: 12 }}>
                    <Row gutter={[8, 16]}>
                      <Col span={12}>
                        <Statistic title="总里程(km)" value={trackPlaybackData.length * 0.5} precision={1} valueStyle={{ fontSize: 18 }} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="时长(h)" value={Math.round(trackPlaybackData.length * 0.05 * 10) / 10} precision={1} valueStyle={{ fontSize: 18 }} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="平均速度(km/h)" value={65} precision={0} valueStyle={{ fontSize: 18, color: '#52c41a' }} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="最高速度(km/h)" value={82} precision={0} valueStyle={{ fontSize: 18, color: '#faad14' }} />
                      </Col>
                    </Row>
                  </Card>
                </Col>
                <Col span={16}>
                  <div
                    ref={playbackMapContainerRef}
                    style={{ height: 400, borderRadius: 8, border: '1px solid #f0f0f0' }}
                  />
                  {trackPlaybackData[playbackIndex] && (
                    <Card size="small" bodyStyle={{ padding: 12, marginTop: 12 }}>
                      <Row gutter={16}>
                        <Col span={6}>
                          <Text type="secondary" style={{ fontSize: 11 }}>车速</Text>
                          <div><Text strong>{trackPlaybackData[playbackIndex].speed || 0} km/h</Text></div>
                        </Col>
                        <Col span={6}>
                          <Text type="secondary" style={{ fontSize: 11 }}>方向</Text>
                          <div><Text strong>{trackPlaybackData[playbackIndex].heading || trackPlaybackData[playbackIndex].direction || 0}°</Text></div>
                        </Col>
                        <Col span={12}>
                          <Text type="secondary" style={{ fontSize: 11 }}>位置</Text>
                          <div><Text ellipsis style={{ fontSize: 12 }}>{trackPlaybackData[playbackIndex].address || trackPlaybackData[playbackIndex].location || `${(trackPlaybackData[playbackIndex].longitude || trackPlaybackData[playbackIndex].lng || 0).toFixed(4)}, ${(trackPlaybackData[playbackIndex].latitude || trackPlaybackData[playbackIndex].lat || 0).toFixed(4)}`}</Text></div>
                        </Col>
                      </Row>
                    </Card>
                  )}
                </Col>
              </Row>
            </>
          )}
        </Space>
      </Drawer>

      <Drawer
        title={
          <Space>
            <EyeOutlined style={{ color: '#1677ff' }} />
            <Text strong>事件详情</Text>
            {eventDetailDrawer && (
              <Tag color={eventTypeMap[eventDetailDrawer.type]?.color || 'default'}>
                {eventTypeMap[eventDetailDrawer.type]?.label || eventDetailDrawer.type}
              </Tag>
            )}
          </Space>
        }
        placement="right"
        width={520}
        onClose={() => { setEventDetailDrawer(null) }}
        open={!!eventDetailDrawer}
      >
        {eventDetailDrawer && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="事件类型">
                <Tag color={eventTypeMap[eventDetailDrawer.type]?.color}>{eventTypeMap[eventDetailDrawer.type]?.label}</Tag>
                {eventDetailDrawer.risk_level && eventDetailDrawer.risk_level !== 'normal' && (
                  <Tag color={eventDetailDrawer.risk_level === 'danger' ? 'red' : 'orange'} style={{ marginLeft: 8 }}>
                    {eventDetailDrawer.risk_level === 'danger' ? '高风险' : '需关注'}
                  </Tag>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="发生时间">{formatDateTime(eventDetailDrawer.time)}</Descriptions.Item>
              <Descriptions.Item label="发生位置">{eventDetailDrawer.location}</Descriptions.Item>
              <Descriptions.Item label="经纬度">{eventDetailDrawer.lng.toFixed(4)}, {eventDetailDrawer.lat.toFixed(4)}</Descriptions.Item>
              <Descriptions.Item label="操作人">{eventDetailDrawer.operator}</Descriptions.Item>
              <Descriptions.Item label="运单号">{eventDetailDrawer.waybill_no}</Descriptions.Item>
              {eventDetailDrawer.duration_minutes && (
                <Descriptions.Item label="持续时长">{eventDetailDrawer.duration_minutes} 分钟</Descriptions.Item>
              )}
              <Descriptions.Item label="事件描述">{eventDetailDrawer.remark}</Descriptions.Item>
            </Descriptions>
            {eventDetailDrawer.photos.length > 0 && (
              <>
                <Divider orientation="left" style={{ fontSize: 12, margin: '8px 0' }}>现场照片 ({eventDetailDrawer.photos.length})</Divider>
                <Image.PreviewGroup>
                  <Row gutter={8}>
                    {eventDetailDrawer.photos.map((p, pi) => (
                      <Col span={8} key={pi}>
                        <Image src={p} width="100%" height={80} style={{ objectFit: 'cover', borderRadius: 6 }} />
                      </Col>
                    ))}
                  </Row>
                </Image.PreviewGroup>
              </>
            )}
            <div
              ref={eventMapContainerRef}
              style={{ height: 220, borderRadius: 8, border: '1px solid #f0f0f0' }}
            />
          </Space>
        )}
      </Drawer>

      <ModalForm
        title={
          <Space>
            <PlusOutlined style={{ color: '#1677ff' }} />
            <Text strong>添加押运事件</Text>
          </Space>
        }
        open={addEventModalVisible}
        form={addEventForm}
        onOpenChange={(v) => { if (!v) { setAddEventModalVisible(false); addEventForm.resetFields() } }}
        onFinish={handleAddEvent}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
          onCancel: () => { setAddEventModalVisible(false); addEventForm.resetFields() },
        }}
        submitTimeout={2000}
      >
        <ProFormSelect
          name="event_type"
          label="事件类型"
          width="md"
          placeholder="请选择事件类型"
          rules={[{ required: true, message: '请选择事件类型' }]}
          options={Object.entries(eventTypeMap).map(([k, v]) => ({
            label: (
              <Space>
                <Tag color={v.color}>{v.icon} {v.label}</Tag>
              </Space>
            ),
            value: k,
          }))}
        />
        <ProFormText
          name="location"
          label="事件位置"
          placeholder="请输入事件位置"
          rules={[{ required: true, message: '请输入事件位置' }]}
          fieldProps={{
            prefix: <EnvironmentOutlined />,
          }}
        />
        <ProFormTextArea
          name="remark"
          label="事件描述"
          placeholder="请输入事件详细描述"
          rules={[{ required: true, message: '请输入事件描述' }]}
          fieldProps={{ rows: 4, showCount: true, maxLength: 500 }}
        />
      </ModalForm>
    </div>
  )
}

export default Escort