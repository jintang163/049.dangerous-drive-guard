import { create } from 'zustand'

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
  score_radar: { safety: number; fatigue: number; speed: number; lane: number; focus: number; compliance: number }
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
  warning_no: string
  warning_type: 'rainstorm' | 'fog' | 'wind' | 'snow' | 'ice' | 'high_temp' | 'low_temp' | 'thunder' | 'hail' | 'typhoon'
  severity_level: 1 | 2 | 3 | 4
  title: string
  content: string
  affected_provinces: string[]
  affected_cities: string[]
  affected_area_polygon?: Array<{ lat: number; lng: number }>
  published_time: string
  expire_time: string
  affected_vehicle_count: number
  status: 'active' | 'expired' | 'cancelled'
  suggestions?: string[]
  affected_waybills?: number[]
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
