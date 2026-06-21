import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { message, Modal } from 'antd'
import { getToken, clearToken, getTokenExpire } from '@/utils/auth'
import type {
  UserInfo,
  VehicleStatus,
  AlarmItem,
  VehicleItem,
  DriverItem,
  WaybillItem,
  EscortEvent,
  RescueRequest,
  WeatherWarning,
  BlockchainBlock,
  EvidenceRecord,
  ServiceArea,
  StatData,
  GeoFenceAlertItem,
  GeoFenceStats,
  GeoFenceCheckResult,
} from '@/store/app'

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  trace_id?: string
  total?: number
  page?: number
  page_size?: number
  list?: T extends any[] ? T : any[]
}

export interface PageParams {
  page?: number
  page_size?: number
  [key: string]: any
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

class ApiService {
  private instance: AxiosInstance
  private pendingRequests: Map<string, AbortController> = new Map()

  constructor() {
    this.instance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.instance.interceptors.request.use(this.handleRequest.bind(this))
    this.instance.interceptors.response.use(this.handleResponse.bind(this), this.handleError.bind(this))
  }

  private handleRequest(config: AxiosRequestConfig) {
    const token = getToken()
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    const traceId = `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
    if (config.headers) {
      config.headers['X-Trace-ID'] = traceId
    }

    const controller = new AbortController()
    const key = `${config.method}-${config.url}-${JSON.stringify(config.params || config.data)}`
    this.pendingRequests.set(key, controller)
    config.signal = controller.signal

    return config
  }

  private handleResponse(response: AxiosResponse) {
    const res = response.data as ApiResponse

    if (res.code === 0) {
      return res.data !== undefined ? res : res
    }

    if (res.code === 40100) {
      clearToken()
      Modal.error({
        title: '登录已过期',
        content: res.message || '请重新登录',
        onOk: () => {
          window.location.href = '/login'
        },
      })
      return Promise.reject(res)
    }

    if (res.code === 40300) {
      message.error('无权限访问')
      return Promise.reject(res)
    }

    message.error(res.message || `请求失败 (${res.code})`)
    return Promise.reject(res)
  }

  private handleError(error: any) {
    if (axios.isCancel(error)) {
      return Promise.reject({ message: '请求已取消', canceled: true })
    }

    if (error.code === 'ECONNABORTED') {
      message.error('请求超时，请稍后重试')
      return Promise.reject(error)
    }

    const status = error.response?.status
    const msg = error.response?.data?.message || error.message

    if (status === 401) {
      clearToken()
      window.location.href = '/login'
    } else if (status === 500 || !status) {
      message.error(msg || '服务器错误，请稍后重试')
    } else {
      message.error(msg || `请求错误 (${status})`)
    }

    return Promise.reject(error)
  }

  cancelPendingRequests() {
    this.pendingRequests.forEach(ctrl => ctrl.abort())
    this.pendingRequests.clear()
  }

  async get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
    const res = await this.instance.get(url, { ...config, params })
    return this.extractData<T>(res)
  }

  async getPage<T = any>(url: string, params?: PageParams): Promise<PageResult<T>> {
    const res = await this.instance.get(url, { params })
    const data = res.data as ApiResponse
    if (data?.data && 'list' in (data.data as any)) {
      return data.data as PageResult<T>
    }
    return {
      list: (data as any)?.list || [],
      total: (data as any)?.total || 0,
      page: params?.page || 1,
      page_size: params?.page_size || 20,
    }
  }

  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const res = await this.instance.post(url, data, config)
    return this.extractData<T>(res)
  }

  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const res = await this.instance.put(url, data, config)
    return this.extractData<T>(res)
  }

  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await this.instance.delete(url, config)
    return this.extractData<T>(res)
  }

  private extractData<T>(res: AxiosResponse): T {
    const payload = res.data as ApiResponse
    if (payload?.data !== undefined) {
      return payload.data as T
    }
    return payload as unknown as T
  }
}

export const api = new ApiService()

export const authApi = {
  login: (data: { username: string; password: string }) =>
    api.post<{ access_token: string; user: any; permissions: string[]; expires_in: number }>('/auth/login', data),
  refreshToken: () => api.post<{ access_token: string; expires_in: number }>('/auth/refresh'),
  logout: () => api.post('/auth/logout'),
  getCurrentUser: () => api.get<UserInfo>('/auth/me'),
  changePassword: (data: { old_password: string; new_password: string }) =>
    api.put('/auth/password', data),
}

export const userApi = {
  list: (params?: PageParams) => api.getPage<UserInfo>('/users', params),
  get: (id: number) => api.get<UserInfo>(`/users/${id}`),
  create: (data: Partial<UserInfo> & { password: string }) => api.post<UserInfo>('/users', data),
  update: (id: number, data: Partial<UserInfo>) => api.put<UserInfo>(`/users/${id}`, data),
  delete: (id: number) => api.delete(`/users/${id}`),
  resetPassword: (id: number) => api.post<{ password: string }>(`/users/${id}/reset-password`),
}

export const vehicleApi = {
  list: (params?: PageParams) => api.getPage<VehicleItem>('/vehicles', params),
  get: (id: number) => api.get<VehicleItem>(`/vehicles/${id}`),
  create: (data: Partial<VehicleItem>) => api.post<VehicleItem>('/vehicles', data),
  update: (id: number, data: Partial<VehicleItem>) => api.put<VehicleItem>(`/vehicles/${id}`, data),
  delete: (id: number) => api.delete(`/vehicles/${id}`),
  diagnostics: (id: number) => api.get<Array<{ code: string; desc: string; level: string; time: string }>>(`/vehicles/${id}/diagnostics`),
  uploadDiagnostic: (id: number, data: { code: string; desc: string; level: string }) => api.post(`/vehicles/${id}/diagnostics`, data),
  faults: (id: number) => api.get<Array<{ code: string; desc: string; level: 'low' | 'medium' | 'high'; time: string }>>(`/vehicles/${id}/faults`),
  getRealtimeStatus: (id: number) => api.get<VehicleStatus>(`/vehicles/${id}/status`),
  listRealtimeStatus: (params?: { org_id?: number }) => api.get<VehicleStatus[]>('/vehicles/status', params),
}

export const driverApi = {
  list: (params?: PageParams) => api.getPage<DriverItem>('/drivers', params),
  get: (id: number) => api.get<DriverItem>(`/drivers/${id}`),
  create: (data: Partial<DriverItem> & { password: string }) => api.post<DriverItem>('/drivers', data),
  update: (id: number, data: Partial<DriverItem>) => api.put<DriverItem>(`/drivers/${id}`, data),
  delete: (id: number) => api.delete(`/drivers/${id}`),
  getScore: (id: number, params?: { days?: number }) =>
    api.get<{ driving_score: number; safety_score: number; fatigue_score: number; compliance_score: number }>(`/drivers/${id}/score`, params),
  getScoreRank: (params?: { top?: number }) => api.get<Array<{ id: number; name: string; score: number; rank: number }>>('/drivers/scores/rank', params),
}

export const routeApi = {
  plan: (data: {
    start: { lat: number; lng: number; address?: string }
    end: { lat: number; lng: number; address?: string }
    vehicle_type: string
    danger_level: number
    strategy?: 'shortest' | 'safest' | 'economic'
  }) => api.post<any>('/routes/plan', data),
  planMultiStrategy: (data: any) => api.post<any>('/routes/plan/multi', data),
  replan: (data: any) => api.post<any>('/routes/replan', data),
  listRestrictedAreas: (params?: { type?: string; bounds?: string }) =>
    api.get<any[]>('/routes/restricted-areas', params),
  getServiceAreas: (params?: { route_id?: number }) => api.get<any[]>('/routes/service-areas', params),
  recommendServiceArea: (data: { waybill_id: number; fatigue_score?: number }) =>
    api.post<any>('/routes/service-areas/recommend', data),
}

export const trafficApi = {
  listEvents: (params?: { status?: string; event_type?: string; keyword?: string } & PageParams) =>
    api.getPage<any>('/traffic/events', params),
  getEvent: (id: number) => api.get<any>(`/traffic/events/${id}`),
  createEvent: (data: any) => api.post<any>('/traffic/events', data),
  resolveEvent: (id: number) => api.post<any>(`/traffic/events/${id}/resolve`),
}

export const replanApi = {
  trigger: (data: any) => api.post<any>('/replans/trigger', data),
  confirm: (id: number, data: { action: 'confirm' | 'reject'; confirm_note?: string }) =>
    api.post<any>(`/replans/${id}/confirm`, data),
  list: (params?: {
    waybill_id?: number
    vehicle_id?: number
    trigger_type?: string
    status?: string
    keyword?: string
    start_date?: string
    end_date?: string
  } & PageParams) => api.getPage<any>('/replans', params),
  get: (id: number) => api.get<any>(`/replans/${id}`),
  getStatistics: (params?: { days?: number; org_id?: number }) =>
    api.get<any>('/replans/statistics/overview', params),
}

export const fatigueApi = {
  detect: (data: { image_base64: string; vehicle_id: number; driver_id: number }) =>
    api.post<{ fatigue_score: number; fatigue_level: string; alarm_triggered: boolean; landmarks: number[] }>('/fatigue/detect', data),
  uploadFrame: (data: FormData) =>
    api.post<{ frame_id: number; fatigue_score: number; processed: boolean }>('/fatigue/frames', data, { headers: { 'Content-Type': 'multipart/form-data' } }),
  uploadMultiCamera: (data: {
    vehicle_id: number;
    driver_id: number;
    waybill_id?: number;
    frames: Array<{
      position: string;
      image_url?: string;
      image_base64?: string;
      face_detected: boolean;
      confidence: number;
      quality: number;
      occluded: boolean;
      backlit: boolean;
      metrics: Record<string, unknown>;
      landmarks: Record<string, unknown>;
    }>;
    latitude?: number;
    longitude?: number;
    vehicle_speed?: number;
    edge_computed?: boolean;
    network_status?: string;
  }) =>
    api.post<{
      fatigue_score: number;
      fatigue_level: string;
      need_alarm: boolean;
      alarm_type?: string;
      alarm_message?: string;
      fusion_result?: {
        fatigue_score: number;
        fatigue_level: string;
        fusion_method: string;
        used_cameras: string[];
        primary_camera: string;
        fusion_confidence: number;
        left_score: number;
        center_score: number;
        right_score: number;
        occlusion_detected: boolean;
        backlit_detected: boolean;
      };
      camera_frames?: Record<string, {
        position: string;
        image_url: string;
        face_detected: boolean;
        confidence: number;
        quality: number;
      }>;
    }>('/fatigue/upload/multi-camera', data),
  getMultiCameraHistory: (vehicleID: number, params?: { camera_position?: string; start_time?: string; end_time?: string; page?: number; page_size?: number }) =>
    api.getPage<{
      id: number; vehicle_id: number; driver_id: number; fatigue_score: number; fatigue_level: string;
      detection_time: string; camera_position: string; left_score: number; center_score: number; right_score: number;
      fusion_method: string; fusion_confidence: number; occlusion_detected: boolean; backlit_detected: boolean;
      left_frame_url: string; center_frame_url: string; right_frame_url: string; used_cameras: string;
    }>(`/fatigue/history/${vehicleID}/multi-camera`, params),
  getFusionAccuracyStats: (days = 90) =>
    api.get<{
      total_detections: number;
      multi_camera_count: number;
      single_camera_count: number;
      alarm_count: number;
      avg_score: number;
      avg_confidence: number;
      occlusion_count: number;
      backlit_count: number;
      multi_vs_single_improve_pct: number;
    }>(`/fatigue/fusion/stats`, { days }),
  history: (params?: PageParams & { vehicle_id?: number; driver_id?: number }) =>
    api.getPage<{ id: number; vehicle_id: number; driver_id: number; fatigue_score: number; fatigue_level: string; detection_time: string; vehicle_speed: number; is_alarm_triggered: boolean }>('/fatigue/records', params),
  listAlarms: (params?: PageParams & { status?: string; level?: number; alarm_type?: string }) =>
    api.getPage<AlarmItem>('/fatigue/alarms', params),
  getAlarm: (id: number) => api.get<AlarmItem>(`/fatigue/alarms/${id}`),
  ackAlarm: (id: number, data: { action: string; remark?: string; operator_id: number }) =>
    api.post<{ success: boolean; handled_at: string }>(`/fatigue/alarms/${id}/ack`, data),
  getScore: (driver_id: number, params?: { from?: string; to?: string }) =>
    api.get<{ average_score: number; alarm_count: number; trend: Array<{ date: string; score: number }> }>('/fatigue/scores', params),
  uploadVideo: (vehicleID: number, alarmID: number, file: File, onProgress?: (p: number) => void) => {
    const formData = new FormData()
    formData.append('vehicle_id', String(vehicleID))
    formData.append('alarm_id', String(alarmID))
    formData.append('video', file)
    return api.post<{ video_id: number; url: string }>('/fatigue/video/upload', formData, {
      onUploadProgress: (e) => onProgress?.(e.total ? Math.round((e.loaded / e.total) * 100) : 0),
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  uploadSnapshot: (vehicleID: number, alarmID: number, file: File) => {
    const formData = new FormData()
    formData.append('vehicle_id', String(vehicleID))
    formData.append('alarm_id', String(alarmID))
    formData.append('snapshot', file)
    return api.post<{ snapshot_id: number; url: string }>('/fatigue/video/snapshot/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  getVideoURL: (alarmID: number) => api.get<{ url: string }>(`/fatigue/video/${alarmID}/video`),
  getSnapshotURL: (alarmID: number) => api.get<{ url: string }>(`/fatigue/video/${alarmID}/snapshot`),
  downloadVideo: (alarmID: number) => api.get<Blob>(`/fatigue/video/${alarmID}/download`, undefined, { responseType: 'blob' }),
}

export const waybillApi = {
  list: (params?: PageParams & { status?: string; danger_level?: number }) =>
    api.getPage<WaybillItem>('/waybills', params),
  get: (id: number) => api.get<WaybillItem>(`/waybills/${id}`),
  create: (data: Partial<WaybillItem>) => api.post<WaybillItem>('/waybills', data),
  update: (id: number, data: Partial<WaybillItem>) => api.put<WaybillItem>(`/waybills/${id}`, data),
  delete: (id: number) => api.delete(`/waybills/${id}`),
  updateStatus: (id: number, data: { status: string; remark?: string }) =>
    api.post<{ success: boolean; updated_at: string }>(`/waybills/${id}/status`, data),
  getStatusLogs: (id: number) => api.get<Array<{ status: string; time: string; operator: string; remark?: string }>>(`/waybills/${id}/status-logs`),
  sign: (id: number, data: { signature: string; receiver_name: string }) =>
    api.post<{ success: boolean; signed_at: string; signature_url: string }>(`/waybills/${id}/sign`, data),
  getBlockchainEvidence: (id: number) => api.get<{ blockchain_hash: string; blockchain_height: number; timestamp: string; evidence: Record<string, any> }>(`/waybills/${id}/blockchain`),
}

export const escortApi = {
  list: (params?: PageParams & { waybill_id?: number }) =>
    api.getPage<EscortEvent>('/escort/events', params),
  create: (data: Partial<EscortEvent>) => api.post<EscortEvent>('/escort/events', data),
  get: (id: number) => api.get<EscortEvent>(`/escort/events/${id}`),
  getByWaybill: (waybill_id: number) => api.get<EscortEvent[]>(`/escort/events/waybill/${waybill_id}`),
  exportReport: (waybill_id: number) =>
    api.get<Blob>(`/escort/report/${waybill_id}`, undefined, { responseType: 'blob' }),
  getStatistics: (params?: { org_id?: number }) =>
    api.get<any>('/escort/statistics', params),

  createShift: (data: {
    escort_id: number
    escort_name: string
    shift_date: string
    start_time: string
    end_time: string
    polling_interval?: number
    description?: string
  }) => api.post<any>('/escort/shifts', data),
  getShift: (id: number) => api.get<any>(`/escort/shifts/${id}`),
  listShifts: (params?: PageParams & { escort_id?: number; dispatcher_id?: number; status?: string }) =>
    api.getPage<any>('/escort/shifts', params),
  updateShiftStatus: (id: number, status: string) =>
    api.put<any>(`/escort/shifts/${id}/status`, { status }),
  assignVehicles: (id: number, vehicle_ids: number[]) =>
    api.post<any>(`/escort/shifts/${id}/assign-vehicles`, { vehicle_ids }),
  getShiftAssignments: (id: number) =>
    api.get<{ list: any[]; total: number }>(`/escort/shifts/${id}/assignments`),

  reportSOS: (data: {
    vehicle_id: number
    latitude?: number
    longitude?: string
    sos_type?: string
    description?: string
  }) => api.post<any>('/escort/sos/report', data),
  getSOSAlerts: (params?: PageParams & { vehicle_id?: number; escort_id?: number; status?: string }) =>
    api.getPage<any>('/escort/sos', params),
  handleSOS: (data: { alert_id: number; handle_note?: string }) =>
    api.post<any>('/escort/sos/handle', data),
  resolveSOS: (id: number, note?: string) =>
    api.post<any>(`/escort/sos/${id}/resolve`, { note }),

  getTrackPlayback: (params: {
    waybill_id?: number
    vehicle_id?: number
    start_time?: string
    end_time?: string
  }) => api.get<{ list: any[]; total: number }>('/escort/track/playback', params),

  getVideoRecords: (params?: PageParams & {
    vehicle_id?: number
    waybill_id?: number
    record_type?: string
    start_time?: string
    end_time?: string
  }) => api.getPage<any>('/escort/videos', params),
  viewVideoRecord: (id: number) => api.post<any>(`/escort/videos/${id}/view`),

  sendIntercom: (data: {
    vehicle_id: number
    message: string
    priority?: string
  }) => api.post<any>('/escort/intercom', data),
  getIntercomLogs: (params?: PageParams & { vehicle_id?: number }) =>
    api.getPage<any>('/escort/intercom/logs', params),

  startPollingSession: (params?: { shift_id?: number }) =>
    api.post<any>('/escort/polling/start', undefined, { params }),
  endPollingSession: (id: number, polling_count?: number) =>
    api.post<any>(`/escort/polling/${id}/end`, { polling_count }),
  getEscortVehiclesForPolling: (params?: { escort_id?: number }) =>
    api.get<{ list: any[]; total: number }>('/escort/polling/vehicles', params),

  getGeoFenceStats: (params?: { org_id?: number }) =>
    api.get<GeoFenceStats>('/escort/geo-fence/statistics', params),
  checkGeoFence: (data: {
    vehicle_id: number
    driver_id?: number
    waybill_id?: number
    latitude: number
    longitude: number
    address?: string
    threshold_meters?: number
  }) => api.post<GeoFenceCheckResult>('/escort/geo-fence/check', data),
  getGeoFenceAlerts: (params?: PageParams & {
    vehicle_id?: number
    waybill_id?: number
    escort_id?: number
    status?: string
  }) => api.getPage<GeoFenceAlertItem>('/escort/geo-fence/alerts', params),
  confirmGeoFenceAlert: (data: {
    alert_id: number
    confirm_type: 'detour' | 'deviate'
    reason_detail?: string
    note?: string
    latitude?: number
    longitude?: number
  }) => api.post<any>('/escort/geo-fence/alerts/confirm', data),
  resolveGeoFenceAlert: (id: number, resolved_note: string) =>
    api.post<any>(`/escort/geo-fence/alerts/${id}/resolve`, { resolved_note }),
  getGeoFenceConfirmLogs: (params?: PageParams & { alert_id?: number; vehicle_id?: number }) =>
    api.getPage<any>('/escort/geo-fence/confirm-logs', params),
}

export const rescueApi = {
  listRequests: (params?: PageParams & { status?: string; type?: string }) =>
    api.getPage<RescueRequest>('/rescue/requests', params),
  getRequest: (id: number) => api.get<RescueRequest>(`/rescue/requests/${id}`),
  createSOS: (data: { vehicle_id: number; latitude: number; longitude: number; rescue_type: string; description: string }) =>
    api.post<RescueRequest>('/rescue/sos', data),
  updateRequest: (id: number, data: Partial<RescueRequest>) => api.put<RescueRequest>(`/rescue/requests/${id}`, data),
  assignResource: (id: number, data: { resource_type: string; resource_id: number; distance_km?: number }) =>
    api.post<{ success: boolean; assigned_at: string }>(`/rescue/requests/${id}/assign`, data),
  sendMessage: (id: number, data: { sender: string; role: string; content: string }) =>
    api.post<{ id: number; time: string }>(`/rescue/requests/${id}/messages`, data),
  getResources: (params?: { type?: string; lat?: number; lng?: number; radius_km?: number }) =>
    api.get<Array<{ id: number; name: string; type: string; latitude: number; longitude: number; distance_km?: number; available: boolean }>>('/rescue/resources', params),
  listResources: (params?: PageParams & { type?: string }) =>
    api.getPage<{ id: number; name: string; type: string; status: string; contact?: string }>('/rescue/resources', params),
}

export const weatherApi = {
  listWarnings: (params?: PageParams & { status?: string; type?: string; level?: number }) =>
    api.getPage<WeatherWarning>('/weather/warnings', params),
  getActiveWarnings: (params?: { province?: string }) => api.get<WeatherWarning[]>('/weather/warnings/active', params),
  getWarning: (id: number) => api.get<WeatherWarning>(`/weather/warnings/${id}`),
  getRouteAffectedRoutes: (warning_id: number) => api.get<Array<{ waybill_id: number; waybill_no: string; vehicle_plate: string; affected: boolean }>>(`/weather/warnings/${warning_id}/routes`),
  replanAffectedRoutes: (warning_id: number) => api.post<{ success: boolean; replanned_count: number }>(`/weather/warnings/${warning_id}/replan`),
}

export const blockchainApi = {
  listBlocks: (params?: { page?: number; page_size?: number; from_height?: number }) =>
    api.getPage<BlockchainBlock>('/blockchain/blocks', params),
  getBlock: (height: number) => api.get<BlockchainBlock>(`/blockchain/blocks/${height}`),
  listEvidence: (params?: PageParams & { business_type?: string }) =>
    api.getPage<EvidenceRecord>('/blockchain/evidence', params),
  getEvidence: (id: number) => api.get<EvidenceRecord>(`/blockchain/evidence/${id}`),
  createEvidence: (data: { business_type: string; business_id: number; business_no: string; data: Record<string, any> }) =>
    api.post<EvidenceRecord>('/blockchain/evidence', data),
  verifyEvidence: (data: { hash: string } | { file: File }) =>
    api.post<{ valid: boolean; evidence?: EvidenceRecord; message: string }>('/blockchain/verify', data),
  getStats: () => api.get<{ block_height: number; total_transactions: number; today_new: number; data_size_mb: number }>('/blockchain/stats'),
}

export const monitorApi = {
  getDashboardStats: () => api.get<StatData>('/monitor/stats/dashboard'),
  getStatistics: () => api.get<StatData>('/monitor/statistics'),
  getBigScreenData: () => api.get<StatData>('/monitor/stats/big-screen'),
  listAlarms: (params?: PageParams & { status?: string }) => api.getPage<AlarmItem>('/monitor/alarms', params),
  voiceIntercom: (vehicle_id: number, data: { action: 'start' | 'stop'; operator_id: number; message?: string; priority?: number }) =>
    api.post<{ success: boolean; session_id?: string }>(`/monitor/vehicles/${vehicle_id}/intercom`, data),
  dispatchServiceArea: (vehicle_id: number, data: { service_area_id: number; reason?: string; rest_duration?: number }) =>
    api.post<{ success: boolean; dispatched_at: string }>(`/monitor/vehicles/${vehicle_id}/dispatch-service-area`, data),
  notifyLawEnforcement: (vehicle_id: number, data: { station_id: number; reason: string }) =>
    api.post<{ success: boolean; notified_at: string }>(`/monitor/vehicles/${vehicle_id}/notify-law-enforcement`, data),
  exportReport: (params?: { type: string; from: string; to: string }) =>
    api.get<Blob>('/monitor/export/report', params, { responseType: 'blob' }),
}

export interface TimeScheduleRule {
  weekdays: number[]
  start_time: string
  end_time: string
  description?: string
}

export interface RestrictedAreaItem {
  id: number
  name: string
  area_type: string
  shape_type: 'polygon' | 'circle'
  level: number
  province?: string
  city?: string
  district?: string
  address?: string
  boundary_polygon: any
  center_latitude: number
  center_longitude: number
  radius: number
  restrict_hazard_classes?: string
  restrict_vehicle_types?: string
  height_limit?: number
  weight_limit?: number
  time_schedule?: TimeScheduleRule[]
  effective_from?: string
  effective_to?: string
  source?: string
  is_temporary: number
  temp_reason?: string
  template_id?: number
  gis_import_id?: string
  created_by?: number
  approval_status: 0 | 1 | 2 | 3 | 4 | 5
  first_approver_id?: number
  first_approval_at?: string
  first_approval_note?: string
  second_approver_id?: number
  second_approval_at?: string
  second_approval_note?: string
  status: number
  created_at: string
  updated_at: string
}

export interface RestrictedAreaTemplate {
  id: number
  template_name: string
  template_category: 'hospital' | 'school' | 'mall' | 'water_source' | 'custom'
  area_type: string
  level: number
  default_radius: number
  restrict_hazard_classes?: string
  restrict_vehicle_types?: string
  height_limit?: number
  weight_limit?: number
  time_rules?: TimeScheduleRule[]
  description?: string
  is_builtin: number
  is_enabled: number
  created_by?: number
  created_at: string
  updated_at: string
}

export interface ApprovalRecord {
  id: number
  area_id: number
  approval_level: number
  approver_id: number
  approver_name: string
  approval_action: string
  approval_note?: string
  old_status: number
  new_status: number
  created_at: string
}

export interface GisImportRecord {
  id: number
  import_batch_no: string
  file_name?: string
  source_type: string
  total_count: number
  success_count: number
  failed_count: number
  failed_details?: any
  import_status: string
  imported_by?: number
  created_at: string
}

export const restrictedAreaApi = {
  list: (params?: PageParams & {
    area_type?: string
    is_temporary?: number
    approval_status?: number
    keyword?: string
  }) => api.getPage<RestrictedAreaItem>('/restricted-areas', params),
  get: (id: number) => api.get<RestrictedAreaItem>(`/restricted-areas/${id}`),
  create: (data: Partial<RestrictedAreaItem> & { submit_approval?: boolean; time_schedule?: TimeScheduleRule[] }) =>
    api.post<RestrictedAreaItem>('/restricted-areas', data),
  update: (id: number, data: Partial<RestrictedAreaItem> & { time_schedule?: TimeScheduleRule[] }) =>
    api.put<RestrictedAreaItem>(`/restricted-areas/${id}`, data),
  delete: (id: number) => api.delete(`/restricted-areas/${id}`),
  submitApproval: (id: number) => api.post<{ success: boolean }>(`/restricted-areas/${id}/submit`),
  approveFirst: (id: number, data: { approval_note?: string }) =>
    api.post<{ success: boolean }>(`/restricted-areas/${id}/approve/first`, data),
  approveSecond: (id: number, data: { approval_note?: string }) =>
    api.post<{ success: boolean }>(`/restricted-areas/${id}/approve/second`, data),
  reject: (id: number, level: number, data: { approval_note?: string }) =>
    api.post<{ success: boolean }>(`/restricted-areas/${id}/reject?level=${level}`, data),
  revoke: (id: number, data: { approval_note?: string }) =>
    api.post<{ success: boolean }>(`/restricted-areas/${id}/revoke`, data),
  getApprovalHistory: (id: number) =>
    api.get<{ list: ApprovalRecord[]; total: number }>(`/restricted-areas/${id}/approvals`),
  listPendingApprovals: (params?: { level?: number; page?: number; page_size?: number }) =>
    api.getPage<RestrictedAreaItem>('/restricted-areas/approvals/pending', params),

  listTemplates: (params?: PageParams & { category?: string }) =>
    api.getPage<RestrictedAreaTemplate>('/restricted-areas/templates', params),
  getTemplate: (id: number) => api.get<RestrictedAreaTemplate>(`/restricted-areas/templates/${id}`),
  createTemplate: (data: Partial<RestrictedAreaTemplate> & { time_rules?: TimeScheduleRule[] }) =>
    api.post<RestrictedAreaTemplate>('/restricted-areas/templates', data),
  updateTemplate: (id: number, data: Partial<RestrictedAreaTemplate> & { time_rules?: TimeScheduleRule[] }) =>
    api.put<RestrictedAreaTemplate>(`/restricted-areas/templates/${id}`, data),
  deleteTemplate: (id: number) => api.delete(`/restricted-areas/templates/${id}`),
  applyTemplate: (templateId: number, data: { center_lat: number; center_lng: number; name?: string; address?: string }) =>
    api.post<RestrictedAreaItem>(`/restricted-areas/templates/${templateId}/apply`, data),

  importGis: (data: { source_type: string; file_name?: string; features: any }) =>
    api.post<GisImportRecord>('/restricted-areas/gis/import-json', data),
  importGisFile: (file: File, sourceType: string, onProgress?: (p: number) => void) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('source_type', sourceType)
    return api.post<GisImportRecord>('/restricted-areas/gis/import', formData, {
      onUploadProgress: (e) => onProgress?.(e.total ? Math.round((e.loaded / e.total) * 100) : 0),
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  listGisImports: (params?: PageParams) =>
    api.getPage<GisImportRecord>('/restricted-areas/gis/imports', params),

  pullActiveAreas: (params?: { since_version?: number; hazard_class?: string }) =>
    api.get<{ list: RestrictedAreaItem[]; total: number; latest_version: number }>('/restricted-areas/sync/pull', params),
}

export const minioApi = {
  getConfig: () => api.get<{ endpoint: string; bucket: string; region: string; access_key: string }>('/minio/config'),
  getPresignedUrl: (object_name: string, expires_in?: number) =>
    api.get<{ url: string; expires_at: string }>('/minio/presigned-url', { object_name, expires_in }),
  uploadFile: (bucket: string, file: File, onProgress?: (p: number) => void) => {
    const formData = new FormData()
    formData.append('bucket', bucket)
    formData.append('file', file)
    return api.post<{ url: string; object_name: string; size: number }>('/minio/upload', formData, {
      onUploadProgress: (e) => onProgress?.(e.total ? Math.round((e.loaded / e.total) * 100) : 0),
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  deleteFile: (object_name: string) => api.delete(`/minio/delete`, { data: { object_name } }),
  listFiles: (params?: { bucket?: string; prefix?: string }) =>
    api.get<Array<{ name: string; size: number; last_modified: string; url: string }>>('/minio/files', params),
}

export interface ServiceAreaItem {
  id: number
  name: string
  highway_name: string
  direction: string
  province: string
  city: string
  latitude: number
  longitude: number
  distance_from_start_km: number
  has_restaurant: boolean
  has_hotel: boolean
  has_fuel_station: boolean
  has_charging: boolean
  has_rest_room: boolean
  has_maintenance: boolean
  has_danger_goods_parking: boolean
  parking_spaces: number
  danger_parking_spaces: number
  phone: string
  rating: number
  status: number
  created_at: string
  updated_at: string
}

export interface ServiceAreaRealtimeStatus {
  id: number
  service_area_id: number
  total_parking_spaces: number
  available_parking_spaces: number
  total_danger_spaces: number
  available_danger_spaces: number
  has_fuel: boolean
  fuel_price_92: number
  fuel_price_95: number
  fuel_price_diesel: number
  has_charging: boolean
  charging_piles_total: number
  charging_piles_available: number
  has_restaurant: boolean
  restaurant_rating: number
  restaurant_wait_minutes: number
  has_hotel: boolean
  hotel_rating: number
  has_maintenance: boolean
  security_level: number
  security_patrol_interval: number
  crowd_level: number
  weather_condition: string
  update_time: string
  data_source: string
}

export interface ServiceAreaReview {
  id: number
  service_area_id: number
  driver_id: number
  driver_name: string
  waybill_id: number
  vehicle_id: number
  security_score: number
  environment_score: number
  food_score: number
  service_score: number
  overall_score: number
  comment_text: string
  tags_array: string[]
  images_array: string[]
  is_anonymous: boolean
  status: number
  check_in_record_id: number
  created_at: string
  updated_at: string
}

export interface DrivingRestRecord {
  id: number
  driver_id: number
  vehicle_id: number
  waybill_id: number
  record_date: string
  drive_start_time: string
  drive_end_time: string
  continuous_drive_minutes: number
  rest_start_time: string
  rest_end_time: string
  rest_duration_minutes: number
  rest_service_area_id: number
  rest_service_area_name: string
  status: 'driving' | 'resting' | 'completed'
  is_overtime: boolean
  overtime_minutes: number
  check_in_time: string
  check_in_latitude: number
  check_in_longitude: number
  check_out_time: string
  check_out_latitude: number
  check_out_longitude: number
  min_rest_required: number
  max_continuous_drive: number
  remaining_drive_minutes?: number
  rest_progress_percent?: number
  created_at: string
  updated_at: string
}

export interface RestCountdownResponse {
  driver_id: number
  vehicle_id: number
  waybill_id: number
  status: string
  continuous_drive_minutes: number
  remaining_drive_minutes: number
  max_continuous_drive: number
  is_overtime: boolean
  overtime_minutes: number
  min_rest_required: number
  current_rest_minutes: number
  rest_progress_percent: number
  can_continue_driving: boolean
  current_service_area_id: number
  current_service_area_name: string
  next_recommendation_id: number
}

export interface ServiceAreaRecommendation {
  id: number
  recommend_no: string
  driver_id: number
  vehicle_id: number
  waybill_id: number
  current_latitude: number
  current_longitude: number
  current_address: string
  continuous_drive_minutes: number
  remaining_drive_minutes: number
  fatigue_score: number
  hazard_class: string
  recommend_reason: string
  recommended_service_area_id: number
  recommended_service_area_name: string
  distance_km: number
  estimated_arrival_minutes: number
  alternatives_array: Array<{
    service_area_id: number
    service_area_name: string
    distance_km: number
    estimated_arrival_minutes: number
    available_danger_spaces: number
    security_level: number
    restaurant_rating: number
    has_fuel: boolean
    has_charging: boolean
    recommend_reason: string
    match_score: number
  }>
  status: string
  accepted_at: string
  arrived_at: string
  dispatch_source: string
  dispatcher_id: number
  created_at: string
  updated_at: string
}

export const serviceAreaApi = {
  list: (params?: PageParams & { keyword?: string; has_danger_parking?: boolean }) =>
    api.getPage<ServiceAreaItem>('/service-areas', params),
  get: (id: number) => api.get<{ basic_info: ServiceAreaItem; real_status: ServiceAreaRealtimeStatus }>(`/service-areas/${id}`),
  getRealtimeStatus: (id: number) => api.get<ServiceAreaRealtimeStatus>(`/service-areas/${id}/status`),
  updateRealtimeStatus: (data: {
    service_area_id: number
    available_parking_spaces?: number
    available_danger_spaces?: number
    security_level?: number
    restaurant_rating?: number
    crowd_level?: number
    weather_condition?: string
  }) => api.post<{ success: boolean; update_time: string }>('/service-areas/status', data),

  getRestCountdown: (driverId: number, vehicleId?: number) =>
    api.get<RestCountdownResponse>('/service-areas/rest/countdown', { driver_id: driverId, vehicle_id: vehicleId }),
  startDriving: (data: { driver_id: number; vehicle_id: number; waybill_id?: number }) =>
    api.post<DrivingRestRecord>('/service-areas/rest/start', data),
  checkIn: (data: {
    driver_id: number
    vehicle_id: number
    service_area_id: number
    latitude: number
    longitude: number
    waybill_id?: number
  }) => api.post<DrivingRestRecord>('/service-areas/rest/check-in', data),
  checkOut: (data: {
    driver_id: number
    vehicle_id: number
    latitude?: number
    longitude?: number
  }) => api.post<DrivingRestRecord>('/service-areas/rest/check-out', data),
  listRestRecords: (params?: PageParams & { driver_id?: number; start_date?: string; end_date?: string }) =>
    api.getPage<DrivingRestRecord>('/service-areas/rest/records', params),

  recommend: (data: {
    driver_id: number
    vehicle_id: number
    waybill_id?: number
    latitude: number
    longitude: number
    hazard_class?: string
    fatigue_score?: number
    radius_km?: number
  }) => api.post<ServiceAreaRecommendation>('/service-areas/recommend', data),

  acceptRecommendation: (id: number) =>
    api.post<{ success: boolean; accepted_at: string }>(`/service-areas/recommendations/${id}/accept`),
  rejectRecommendation: (id: number, reason?: string) =>
    api.post<{ success: boolean }>(`/service-areas/recommendations/${id}/reject`, { reason }),

  listReviews: (params?: PageParams & { service_area_id?: number }) =>
    api.getPage<ServiceAreaReview>('/service-areas/reviews', params),
  submitReview: (data: {
    service_area_id: number
    driver_id: number
    security_score: number
    environment_score?: number
    food_score?: number
    service_score?: number
    comment_text?: string
    tags?: string[]
    images?: string[]
    is_anonymous?: boolean
    waybill_id?: number
    vehicle_id?: number
  }) => api.post<ServiceAreaReview>('/service-areas/reviews', data),

  getStatistics: () => api.get<{
    total_service_areas: number
    danger_parking_areas: number
    average_rating: number
    today_check_ins: number
    today_reviews: number
  }>('/service-areas/statistics/overview'),
}

export default api
