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

export const fatigueApi = {
  detect: (data: { image_base64: string; vehicle_id: number; driver_id: number }) =>
    api.post<{ fatigue_score: number; fatigue_level: string; alarm_triggered: boolean; landmarks: number[] }>('/fatigue/detect', data),
  uploadFrame: (data: FormData) =>
    api.post<{ frame_id: number; fatigue_score: number; processed: boolean }>('/fatigue/frames', data, { headers: { 'Content-Type': 'multipart/form-data' } }),
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
    api.post<GisImportRecord>('/restricted-areas/gis/import', data),
  listGisImports: (params?: PageParams) =>
    api.getPage<GisImportRecord>('/restricted-areas/gis/imports', params),
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

export default api
