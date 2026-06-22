import { create } from 'zustand'
import {
  vehicleApi,
  fatigueApi,
  monitorApi,
  waybillApi,
  driverApi,
  rescueApi,
  weatherApi,
  blockchainApi,
  routeApi,
  emergencyApi,
} from '@/services/api'
import type { PageParams } from '@/services/api'

export interface UserInfo {
  id: number
  username: string
  real_name: string
  phone: string
  email: string
  role: string
  org_id: number
  avatar_url: string
  status: number
}

export interface VehicleStatus {
  vehicle_id: number
  plate_number: string
  vehicle_type: string
  status: string
  driver_id: number
  driver_name: string
  waybill_id: number
  waybill_no: string
  latitude: number
  longitude: number
  current_address: string
  speed: number
  direction: number
  remaining_mileage: number
  remaining_time: number
  fatigue_score: number
  fatigue_level: 'normal' | 'warning' | 'fatigue'
  last_update_time: string
  marker_color: string
  alert_count: number
  gps_time: string
}

export interface AlarmItem {
  id: number
  alarm_no: string
  vehicle_id: number
  driver_id: number
  waybill_id: number
  alarm_type: string
  alarm_level: 1 | 2 | 3
  fatigue_score: number
  continuous_fatigue_minutes: number
  snap_image_url: string
  video_clip_url: string
  latitude: number
  longitude: number
  location_address: string
  vehicle_speed: number
  status: string
  vehicle_informed: boolean
  escalated: boolean
  created_at: string
  updated_at: string
  vehicle_plate: string
  driver_name: string
}

export interface FaultAlert {
  id: number
  vehicle_id: number
  plate_number: string
  fault_code: string
  fault_level: 1 | 2 | 3 | 4
  fault_system: string
  fault_desc: string
  fault_suggestion: string
  emergency_action: string
  latitude: number
  longitude: number
  report_time: string
  status: 0 | 1 | 2
}

export interface EmergencyTaskCard {
  id: number
  card_no: string
  plan_id: number
  un_number: string
  danger_class: string
  vehicle_id: number
  driver_id: number
  waybill_id?: number
  card_title: string
  leak_disposal_brief: string
  neutralizer_brief?: string
  protective_equipment_brief: string
  evacuation_distance?: string
  first_aid_brief: string
  special_notes?: string
  push_channel: string
  push_status: 'pending' | 'pushed' | 'acknowledged' | 'expired'
  pushed_at?: string
  acknowledged_at?: string
  source_type: string
  source_id?: number
  status: 'active' | 'completed' | 'cancelled' | 'expired'
  expire_at?: string
  completed_at?: string
  completed_by?: number
  remark?: string
  created_by?: number
  created_at: string
  updated_at: string
  vehicle_plate?: string
  driver_name?: string
}

export interface VehicleItem {
  id: number
  plate_number: string
  vehicle_type: 'tanker' | 'van' | 'flatbed' | 'container'
  load_capacity: number
  danger_level: 1 | 2 | 3 | 4
  current_driver_id: number
  current_driver_name: string
  status: 'online' | 'offline' | 'maintenance' | 'stopped'
  current_address: string
  latitude: number
  longitude: number
  last_report_time: string
  obd_speed: number
  obd_rpm: number
  obd_water_temp: number
  obd_fuel_level: number
  obd_engine_status: string
  fault_codes: Array<{ code: string; desc: string; level: 'low' | 'medium' | 'high'; time: string }>
  maintenance_records: Array<{ id: number; type: string; time: string; content: string; cost: number; operator: string }>
  adas_enabled: boolean
  adas_alert_count_today: number
  created_at: string
}

export interface DriverItem {
  id: number
  name: string
  employee_no: string
  license_type: string
  driving_years: number
  phone: string
  linked_vehicle_id: number
  linked_vehicle_plate: string
  status: 'on_duty' | 'rest' | 'off_duty'
  risk_level: 'low' | 'medium' | 'high'
  fatigue_count_30d: number
  driving_score: number
  id_card_no: string
  license_no: string
  license_expire_date: string
  qualification_cert_no: string
  qualification_cert_expire: string
  score_radar: { safety: number; fatigue: number; speed: number; lane: number; following: number; focus: number; compliance: number }
  fatigue_trend_30d: Array<{ date: string; count: number }>
  waybills: Array<{ id: number; waybill_no: string; from: string; to: string; date: string; status: string }>
  created_at: string
}

export interface WaybillItem {
  id: number
  waybill_no: string
  dangerous_goods_name: string
  un_number: string
  danger_level: 1 | 2 | 3 | 4
  from_address: string
  from_lat: number
  from_lng: number
  to_address: string
  to_lat: number
  to_lng: number
  current_lat?: number
  current_lng?: number
  vehicle_id: number
  vehicle_plate: string
  driver_id: number
  driver_name: string
  escort_name?: string
  plan_departure_time: string
  plan_arrival_time: string
  actual_departure_time?: string
  actual_arrival_time?: string
  status: 'pending' | 'dispatched' | 'in_transit' | 'resting' | 'loading' | 'unloading' | 'arrived' | 'signed' | 'abnormal' | 'cancelled'
  progress: number
  total_mileage_km: number
  traveled_mileage_km: number
  cargo_weight: number
  emergency_contact: string
  emergency_phone: string
  blockchain_hash?: string
  blockchain_height?: number
  signature_url?: string
  created_at: string
  status_logs?: Array<{ status: string; time: string; operator: string; remark?: string }>
}

export interface EscortEvent {
  id: number
  waybill_id: number
  waybill_no: string
  event_type: 'departure_check' | 'waypoint' | 'abnormal_stop' | 'rest' | 'loading' | 'unloading' | 'sign' | 'emergency'
  event_time: string
  latitude: number
  longitude: number
  address: string
  risk_level: 'normal' | 'attention' | 'warning' | 'high'
  escort_id: number
  escort_name: string
  remark?: string
  photos?: string[]
  vehicle_plate?: string
  driver_name?: string
}

export interface GeoFenceAlertItem {
  id: number
  alert_no: string
  vehicle_id: number
  plate_number: string
  driver_id: number
  driver_name: string
  escort_id: number
  escort_name: string
  waybill_id: number
  waybill_no: string
  route_plan_id: number
  latitude: number
  longitude: number
  address: string
  distance_from_route_meters: number
  threshold_meters: number
  alert_level: 1 | 2 | 3
  status: 'pending' | 'confirmed' | 'escalated' | 'resolved'
  deviate_reason?: 'detour' | 'deviate'
  confirm_note?: string
  confirmed_by?: number
  confirmed_role?: string
  confirmed_at?: string
  reported_to_dispatch: boolean
  reported_at?: string
  resolved_by?: number
  resolved_note?: string
  resolved_at?: string
  daily_deviate_count: number
  nearest_route_point?: { lat: number; lng: number }
  snapshot_url?: string
  auto_reported?: boolean
  created_at: string
  updated_at: string
}

export interface GeoFenceStats {
  total_alerts: number
  pending_alerts: number
  today_alerts: number
  reported_alerts: number
  resolved_alerts: number
  total_confirm_logs: number
  detour_count: number
  deviate_count: number
  auto_reported_count: number
}

export interface GeoFenceCheckResult {
  alert_id: number
  alert_no: string
  is_deviated: boolean
  distance_from_route_meters: number
  threshold_meters: number
  alert_level: number
  daily_deviate_count: number
  auto_reported: boolean
  status: string
  nearest_route_point?: { lat: number; lng: number }
  message: string
}

export interface NightVisionConfig {
  id: number
  vehicle_id: number
  device_id: string

  infrared_enabled: boolean
  infrared_auto_mode: boolean
  infrared_manual_on: boolean
  infrared_intensity: number
  infrared_intensity_auto: boolean
  low_light_threshold_lux: number
  high_light_threshold_lux: number

  enhancement_enabled: boolean
  enhance_mode: 'auto' | 'night' | 'infrared' | 'low_light' | 'manual'
  gamma_value: number
  brightness_boost: number
  contrast_boost: number
  histogram_equalization: boolean
  clahe_enabled: boolean
  denoise_enabled: boolean
  denoise_strength: number
  sharpen_enabled: boolean
  sharpen_strength: number

  night_mode_auto: boolean
  night_start_hour: number
  night_end_hour: number
  low_light_face_detect: boolean
  min_face_confidence_night: number

  created_at: string
  updated_at: string
}

export interface InfraredLightLog {
  id: number
  vehicle_id: number
  driver_id: number
  device_id: string
  action: 'turn_on' | 'turn_off' | 'intensity_change' | 'auto_trigger' | 'manual_trigger'
  trigger_type: 'auto' | 'manual' | 'system'
  light_on: boolean
  intensity_before?: number
  intensity_after?: number
  light_level_lux?: number
  reason: string
  latitude: number
  longitude: number
  timestamp: string
  face_detected_before?: boolean
  face_detected_after?: boolean
  confidence_before?: number
  confidence_after?: number
  created_at: string
}

export interface ImageEnhanceRecord {
  id: number
  vehicle_id: number
  driver_id: number
  waybill_id: number
  device_id: string

  original_image_url: string
  enhanced_image_url: string

  enhance_mode: string
  gamma_value?: number
  brightness_delta: number
  contrast_delta: number
  denoise_applied: boolean
  denoise_strength: number
  histogram_eq_applied: boolean
  sharpen_applied: boolean

  original_brightness_avg?: number
  enhanced_brightness_avg?: number
  original_contrast?: number
  enhanced_contrast?: number

  light_level_lux?: number
  is_night_time: boolean

  face_detected_original: boolean
  face_detected_enhanced: boolean
  face_confidence_original: number
  face_confidence_enhanced: number
  landmark_count_original: number
  landmark_count_enhanced: number

  quality_score_before: number
  quality_score_after: number
  quality_improvement_pct: number

  processing_time_ms: number
  process_on_edge: boolean

  timestamp: string
  created_at: string
}

export interface NightVisionStats {
  total_configs: number
  infrared_enabled_count: number
  enhancement_enabled_count: number

  today_infrared_turn_on_count: number
  today_infrared_turn_off_count: number
  today_infrared_duration_minutes: number

  total_enhance_records: number
  today_enhance_records: number
  avg_quality_improvement_pct: number
  avg_processing_time_ms: number

  night_face_detect_rate_before: number
  night_face_detect_rate_after: number
  avg_face_confidence_before: number
  avg_face_confidence_after: number
  confidence_improvement_pct: number

  auto_trigger_count: number
  manual_trigger_count: number
}

export interface RescueRequest {
  id: number
  rescue_no: string
  vehicle_id: number
  vehicle_plate: string
  driver_id: number
  driver_name: string
  driver_phone: string
  waybill_id?: number
  waybill_no?: string
  request_time: string
  rescue_type: 'accident' | 'breakdown' | 'medical' | 'fire' | 'hazard_leak' | 'other'
  severity_level: 'general' | 'serious' | 'critical'
  latitude: number
  longitude: number
  address: string
  injured_count: number
  description: string
  status: 'pending' | 'responding' | 'arrived' | 'processing' | 'resolved' | 'closed'
  eta?: number
  response_time?: string
  arrival_time?: string
  resolved_time?: string
  assigned_resources?: Array<{ type: string; name: string; distance_km: number; eta_min: number }>
  messages?: Array<{ sender: string; role: string; content: string; time: string }>
  timeline?: Array<{ step: string; time: string; operator: string; remark?: string }>
}

export interface WeatherWarning {
  id: number
  warning_id: string
  warning_type: string
  warning_level: 'blue' | 'yellow' | 'orange' | 'red'
  title: string
  content: string
  affected_provinces: string[]
  affected_cities: string[]
  affected_area_polygon?: Array<{ lat: number; lng: number }>
  publish_time: string
  end_time?: string
  related_vehicle_count: number
  related_waybill_count: number
  processed: number
  center_lat?: number
  center_lng?: number
  trigger_operation_stop?: number
  speed_suggestion_kmh?: number
  suggestion?: string
}

export interface BlockchainBlock {
  height: number
  hash: string
  previous_hash: string
  transaction_count: number
  timestamp: string
  nonce: number
  difficulty: string
  miner: string
  size_kb: number
}

export interface EvidenceRecord {
  id: number
  evidence_id: string
  business_type: 'waybill' | 'alarm' | 'score' | 'escort_event' | 'rescue' | 'vehicle_diagnostic'
  business_id: number
  business_no: string
  data_hash: string
  block_height: number
  block_hash: string
  timestamp: string
  operator_id?: number
  operator_name?: string
  verification_status: 'verified' | 'unverified' | 'failed'
  data_snapshot?: Record<string, any>
}

export interface ServiceArea {
  id: number
  name: string
  highway_name: string
  direction: 'east' | 'west' | 'south' | 'north' | 'bidirectional'
  latitude: number
  longitude: number
  distance_from_start_km: number
  has_rest_area: boolean
  has_fuel: boolean
  has_repair: boolean
  has_food: boolean
  has_accommodation: boolean
  has_charging: boolean
  parking_spaces: number
  dangerous_goods_parking: boolean
  contact_phone?: string
}

export interface StatData {
  total_vehicles: number
  running_vehicles: number
  idle_vehicles: number
  total_drivers: number
  total_waybills: number
  in_transit_waybills: number
  pending_alarms: number
  today_alarms: number
  today_mileage_km: number
  today_fatigue_events: number
  alarm_type_distribution: Array<{ alarm_type: string; count: number }>
  daily_trend: Array<{ date: string; alarms: number; events: number }>
  waybill_status_distribution?: Array<{ status: string; count: number }>
  rescue_stats?: { pending: number; processing: number; today: number; total: number }
  weather_alerts?: { active: number; rainstorm: number; fog: number; wind: number; snow: number }
  blockchain_stats?: { block_height: number; total_transactions: number; today_new: number; data_size_mb: number }
}

interface LoadingState {
  vehicles: boolean
  alarms: boolean
  stats: boolean
  waybills: boolean
  drivers: boolean
  vehiclesList: boolean
  rescue: boolean
  weather: boolean
  blockchain: boolean
  serviceAreas: boolean
}

interface AppState {
  user: UserInfo | null
  token: string
  permissions: string[]
  currentPath: string
  sidebarCollapsed: boolean
  vehicles: VehicleStatus[]
  selectedVehicle: VehicleStatus | null
  vehicleList: VehicleItem[]
  driverList: DriverItem[]
  alarms: AlarmItem[]
  unreadAlarmCount: number
  faultAlerts: FaultAlert[]
  unreadFaultAlertCount: number
  emergencyTaskCards: EmergencyTaskCard[]
  unreadEmergencyCount: number
  waybillList: WaybillItem[]
  selectedWaybill: WaybillItem | null
  escortEvents: EscortEvent[]
  rescueRequests: RescueRequest[]
  weatherWarnings: WeatherWarning[]
  blockchainBlocks: BlockchainBlock[]
  evidenceRecords: EvidenceRecord[]
  serviceAreas: ServiceArea[]
  stats: StatData | null
  wsConnected: boolean
  loading: LoadingState
  error: string | null
  setUser: (user: UserInfo | null) => void
  setToken: (token: string) => void
  setPermissions: (perms: string[]) => void
  setCurrentPath: (path: string) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  updateVehicles: (vehicles: VehicleStatus[]) => void
  upsertVehicle: (v: VehicleStatus) => void
  setSelectedVehicle: (v: VehicleStatus | null) => void
  updateVehicleList: (list: VehicleItem[]) => void
  upsertVehicleItem: (v: VehicleItem) => void
  deleteVehicleItem: (id: number) => void
  updateDriverList: (list: DriverItem[]) => void
  upsertDriverItem: (d: DriverItem) => void
  deleteDriverItem: (id: number) => void
  addAlarm: (alarm: AlarmItem) => void
  updateAlarm: (id: number, data: Partial<AlarmItem>) => void
  setUnreadAlarmCount: (n: number) => void
  addFaultAlert: (fault: FaultAlert) => void
  ackFaultAlert: (id: number) => void
  resolveFaultAlert: (id: number) => void
  clearFaultAlerts: () => void
  addEmergencyTaskCard: (card: EmergencyTaskCard) => void
  ackEmergencyTaskCard: (id: number) => void
  updateEmergencyTaskCard: (id: number, data: Partial<EmergencyTaskCard>) => void
  clearEmergencyTaskCards: () => void
  fetchEmergencyTaskCards: (params?: any) => Promise<void>
  updateWaybillList: (list: WaybillItem[]) => void
  upsertWaybillItem: (w: WaybillItem) => void
  deleteWaybillItem: (id: number) => void
  setSelectedWaybill: (w: WaybillItem | null) => void
  updateEscortEvents: (events: EscortEvent[]) => void
  addEscortEvent: (e: EscortEvent) => void
  updateRescueRequests: (reqs: RescueRequest[]) => void
  upsertRescueRequest: (r: RescueRequest) => void
  updateWeatherWarnings: (w: WeatherWarning[]) => void
  updateBlockchainBlocks: (blocks: BlockchainBlock[]) => void
  updateEvidenceRecords: (records: EvidenceRecord[]) => void
  updateServiceAreas: (areas: ServiceArea[]) => void
  setStats: (stats: StatData | null) => void
  setWsConnected: (connected: boolean) => void
  setLoading: (key: keyof LoadingState, value: boolean) => void
  setError: (error: string | null) => void
  fetchVehicles: (params?: { org_id?: number }) => Promise<void>
  fetchAlarms: (params?: PageParams & { status?: string; level?: number; alarm_type?: string }) => Promise<void>
  fetchStats: () => Promise<void>
  fetchWaybills: (params?: PageParams & { status?: string; danger_level?: number }) => Promise<void>
  fetchDrivers: (params?: PageParams) => Promise<void>
  fetchVehiclesList: (params?: PageParams) => Promise<void>
  fetchRescueRequests: (params?: PageParams & { status?: string; type?: string }) => Promise<void>
  fetchWeatherWarnings: (params?: PageParams & { status?: string; type?: string; level?: number }) => Promise<void>
  fetchBlockchainBlocks: (params?: { page?: number; page_size?: number; from_height?: number }) => Promise<void>
  fetchEvidenceRecords: (params?: PageParams & { business_type?: string }) => Promise<void>
  fetchServiceAreas: (params?: { route_id?: number }) => Promise<void>
  logout: () => void
}

const initialStats: StatData = {
  total_vehicles: 0,
  running_vehicles: 0,
  idle_vehicles: 0,
  total_drivers: 0,
  total_waybills: 0,
  in_transit_waybills: 0,
  pending_alarms: 0,
  today_alarms: 0,
  today_mileage_km: 0,
  today_fatigue_events: 0,
  alarm_type_distribution: [],
  daily_trend: [],
}

const initialLoading: LoadingState = {
  vehicles: false,
  alarms: false,
  stats: false,
  waybills: false,
  drivers: false,
  vehiclesList: false,
  rescue: false,
  weather: false,
  blockchain: false,
  serviceAreas: false,
}

export const useAppStore = create<AppState>((set, get) => ({
  user: null,
  token: '',
  permissions: [],
  currentPath: '/',
  sidebarCollapsed: false,
  vehicles: [],
  selectedVehicle: null,
  vehicleList: [],
  driverList: [],
  alarms: [],
  unreadAlarmCount: 0,
  faultAlerts: [],
  unreadFaultAlertCount: 0,
  emergencyTaskCards: [],
  unreadEmergencyCount: 0,
  waybillList: [],
  selectedWaybill: null,
  escortEvents: [],
  rescueRequests: [],
  weatherWarnings: [],
  blockchainBlocks: [],
  evidenceRecords: [],
  serviceAreas: [],
  stats: initialStats,
  wsConnected: false,
  loading: initialLoading,
  error: null,

  setUser: user => set({ user }),
  setToken: token => set({ token }),
  setPermissions: perms => set({ permissions: perms }),
  setCurrentPath: path => set({ currentPath: path }),
  setSidebarCollapsed: collapsed => set({ sidebarCollapsed: collapsed }),

  updateVehicles: vehicles => set({ vehicles }),
  upsertVehicle: v =>
    set(state => {
      const idx = state.vehicles.findIndex(x => x.vehicle_id === v.vehicle_id)
      if (idx >= 0) {
        const updated = [...state.vehicles]
        updated[idx] = { ...updated[idx], ...v }
        return { vehicles: updated }
      }
      return { vehicles: [...state.vehicles, v] }
    }),
  setSelectedVehicle: v => set({ selectedVehicle: v }),

  updateVehicleList: list => set({ vehicleList: list }),
  upsertVehicleItem: v =>
    set(state => {
      const idx = state.vehicleList.findIndex(x => x.id === v.id)
      if (idx >= 0) {
        const updated = [...state.vehicleList]
        updated[idx] = { ...updated[idx], ...v }
        return { vehicleList: updated }
      }
      return { vehicleList: [...state.vehicleList, v] }
    }),
  deleteVehicleItem: id =>
    set(state => ({ vehicleList: state.vehicleList.filter(v => v.id !== id) })),

  updateDriverList: list => set({ driverList: list }),
  upsertDriverItem: d =>
    set(state => {
      const idx = state.driverList.findIndex(x => x.id === d.id)
      if (idx >= 0) {
        const updated = [...state.driverList]
        updated[idx] = { ...updated[idx], ...d }
        return { driverList: updated }
      }
      return { driverList: [...state.driverList, d] }
    }),
  deleteDriverItem: id =>
    set(state => ({ driverList: state.driverList.filter(d => d.id !== id) })),

  addAlarm: alarm =>
    set(state => ({
      alarms: [alarm, ...state.alarms].slice(0, 200),
      unreadAlarmCount: state.unreadAlarmCount + 1,
    })),
  updateAlarm: (id, data) =>
    set(state => ({
      alarms: state.alarms.map(a => (a.id === id ? { ...a, ...data } : a)),
    })),
  setUnreadAlarmCount: n => set({ unreadAlarmCount: n }),

  addFaultAlert: fault =>
    set(state => ({
      faultAlerts: [fault, ...state.faultAlerts].slice(0, 100),
      unreadFaultAlertCount: state.unreadFaultAlertCount + 1,
    })),
  ackFaultAlert: id =>
    set(state => ({
      faultAlerts: state.faultAlerts.map(f => (f.id === id ? { ...f, status: 1 } : f)),
    })),
  resolveFaultAlert: id =>
    set(state => ({
      faultAlerts: state.faultAlerts.map(f => (f.id === id ? { ...f, status: 2 } : f)),
    })),
  clearFaultAlerts: () =>
    set({ faultAlerts: [], unreadFaultAlertCount: 0 }),

  addEmergencyTaskCard: card =>
    set(state => ({
      emergencyTaskCards: [card, ...state.emergencyTaskCards].slice(0, 100),
      unreadEmergencyCount: state.unreadEmergencyCount + 1,
    })),
  ackEmergencyTaskCard: id =>
    set(state => ({
      emergencyTaskCards: state.emergencyTaskCards.map(c =>
        c.id === id ? { ...c, push_status: 'acknowledged' as const } : c
      ),
    })),
  updateEmergencyTaskCard: (id, data) =>
    set(state => ({
      emergencyTaskCards: state.emergencyTaskCards.map(c =>
        c.id === id ? { ...c, ...data } : c
      ),
    })),
  clearEmergencyTaskCards: () =>
    set({ emergencyTaskCards: [], unreadEmergencyCount: 0 }),

  updateWaybillList: list => set({ waybillList: list }),
  upsertWaybillItem: w =>
    set(state => {
      const idx = state.waybillList.findIndex(x => x.id === w.id)
      if (idx >= 0) {
        const updated = [...state.waybillList]
        updated[idx] = { ...updated[idx], ...w }
        return { waybillList: updated }
      }
      return { waybillList: [...state.waybillList, w] }
    }),
  deleteWaybillItem: id =>
    set(state => ({ waybillList: state.waybillList.filter(w => w.id !== id) })),
  setSelectedWaybill: w => set({ selectedWaybill: w }),

  updateEscortEvents: events => set({ escortEvents: events }),
  addEscortEvent: e =>
    set(state => ({ escortEvents: [e, ...state.escortEvents].slice(0, 500) })),

  updateRescueRequests: reqs => set({ rescueRequests: reqs }),
  upsertRescueRequest: r =>
    set(state => {
      const idx = state.rescueRequests.findIndex(x => x.id === r.id)
      if (idx >= 0) {
        const updated = [...state.rescueRequests]
        updated[idx] = { ...updated[idx], ...r }
        return { rescueRequests: updated }
      }
      return { rescueRequests: [r, ...state.rescueRequests] }
    }),

  updateWeatherWarnings: w => set({ weatherWarnings: w }),
  updateBlockchainBlocks: blocks => set({ blockchainBlocks: blocks }),
  updateEvidenceRecords: records => set({ evidenceRecords: records }),
  updateServiceAreas: areas => set({ serviceAreas: areas }),

  setStats: stats => set({ stats }),
  setWsConnected: connected => set({ wsConnected: connected }),
  setLoading: (key, value) => set(state => ({ loading: { ...state.loading, [key]: value } })),
  setError: error => set({ error }),

  fetchVehicles: async (params) => {
    set({ loading: { ...get().loading, vehicles: true }, error: null })
    try {
      const data = await vehicleApi.listRealtimeStatus(params)
      set({ vehicles: data || [], loading: { ...get().loading, vehicles: false } })
    } catch (e: any) {
      set({ error: e.message || '获取车辆状态失败', loading: { ...get().loading, vehicles: false } })
      throw e
    }
  },

  fetchAlarms: async (params) => {
    set({ loading: { ...get().loading, alarms: true }, error: null })
    try {
      const data = await fatigueApi.listAlarms(params)
      set({
        alarms: data?.list || [],
        unreadAlarmCount: data?.list?.filter(a => a.status === 'pending').length || 0,
        loading: { ...get().loading, alarms: false },
      })
    } catch (e: any) {
      set({ error: e.message || '获取报警列表失败', loading: { ...get().loading, alarms: false } })
      throw e
    }
  },

  fetchStats: async () => {
    set({ loading: { ...get().loading, stats: true }, error: null })
    try {
      const data = await monitorApi.getDashboardStats()
      set({ stats: data || initialStats, loading: { ...get().loading, stats: false } })
    } catch (e: any) {
      set({ error: e.message || '获取统计数据失败', loading: { ...get().loading, stats: false } })
      throw e
    }
  },

  fetchWaybills: async (params) => {
    set({ loading: { ...get().loading, waybills: true }, error: null })
    try {
      const data = await waybillApi.list(params)
      set({ waybillList: data?.list || [], loading: { ...get().loading, waybills: false } })
    } catch (e: any) {
      set({ error: e.message || '获取运单列表失败', loading: { ...get().loading, waybills: false } })
      throw e
    }
  },

  fetchDrivers: async (params) => {
    set({ loading: { ...get().loading, drivers: true }, error: null })
    try {
      const data = await driverApi.list(params)
      set({ driverList: data?.list || [], loading: { ...get().loading, drivers: false } })
    } catch (e: any) {
      set({ error: e.message || '获取驾驶员列表失败', loading: { ...get().loading, drivers: false } })
      throw e
    }
  },

  fetchVehiclesList: async (params) => {
    set({ loading: { ...get().loading, vehiclesList: true }, error: null })
    try {
      const data = await vehicleApi.list(params)
      set({ vehicleList: data?.list || [], loading: { ...get().loading, vehiclesList: false } })
    } catch (e: any) {
      set({ error: e.message || '获取车辆列表失败', loading: { ...get().loading, vehiclesList: false } })
      throw e
    }
  },

  fetchRescueRequests: async (params) => {
    set({ loading: { ...get().loading, rescue: true }, error: null })
    try {
      const data = await rescueApi.listRequests(params)
      set({ rescueRequests: data?.list || [], loading: { ...get().loading, rescue: false } })
    } catch (e: any) {
      set({ error: e.message || '获取救援请求失败', loading: { ...get().loading, rescue: false } })
      throw e
    }
  },

  fetchWeatherWarnings: async (params) => {
    set({ loading: { ...get().loading, weather: true }, error: null })
    try {
      const data = await weatherApi.listWarnings(params)
      set({ weatherWarnings: data?.list || [], loading: { ...get().loading, weather: false } })
    } catch (e: any) {
      set({ error: e.message || '获取天气预警失败', loading: { ...get().loading, weather: false } })
      throw e
    }
  },

  fetchBlockchainBlocks: async (params) => {
    set({ loading: { ...get().loading, blockchain: true }, error: null })
    try {
      const data = await blockchainApi.listBlocks(params)
      set({ blockchainBlocks: data?.list || [], loading: { ...get().loading, blockchain: false } })
    } catch (e: any) {
      set({ error: e.message || '获取区块列表失败', loading: { ...get().loading, blockchain: false } })
      throw e
    }
  },

  fetchEvidenceRecords: async (params) => {
    set({ loading: { ...get().loading, blockchain: true }, error: null })
    try {
      const data = await blockchainApi.listEvidence(params)
      set({ evidenceRecords: data?.list || [], loading: { ...get().loading, blockchain: false } })
    } catch (e: any) {
      set({ error: e.message || '获取存证记录失败', loading: { ...get().loading, blockchain: false } })
      throw e
    }
  },

  fetchServiceAreas: async (params) => {
    set({ loading: { ...get().loading, serviceAreas: true }, error: null })
    try {
      const data = await routeApi.getServiceAreas(params)
      set({ serviceAreas: data || [], loading: { ...get().loading, serviceAreas: false } })
    } catch (e: any) {
      set({ error: e.message || '获取服务区列表失败', loading: { ...get().loading, serviceAreas: false } })
      throw e
    }
  },

  fetchEmergencyTaskCards: async (params) => {
    set({ loading: { ...get().loading, rescue: true }, error: null })
    try {
      const data = await emergencyApi.listTaskCards(params)
      set({
        emergencyTaskCards: data?.list || [],
        unreadEmergencyCount: data?.list?.filter((c: EmergencyTaskCard) => c.push_status === 'pending').length || 0,
        loading: { ...get().loading, rescue: false },
      })
    } catch (e: any) {
      set({ error: e.message || '获取应急任务卡失败', loading: { ...get().loading, rescue: false } })
      throw e
    }
  },

  logout: () => {
    localStorage.removeItem('ddg_access_token')
    localStorage.removeItem('ddg_user_info')
    localStorage.removeItem('ddg_token_expire')
    localStorage.removeItem('ddg_permissions')
    set({
      user: null,
      token: '',
      permissions: [],
      vehicles: [],
      selectedVehicle: null,
      vehicleList: [],
      driverList: [],
      alarms: [],
      unreadAlarmCount: 0,
      faultAlerts: [],
      unreadFaultAlertCount: 0,
      emergencyTaskCards: [],
      unreadEmergencyCount: 0,
      waybillList: [],
      selectedWaybill: null,
      escortEvents: [],
      rescueRequests: [],
      weatherWarnings: [],
      blockchainBlocks: [],
      evidenceRecords: [],
      serviceAreas: [],
    })
  },
}))
