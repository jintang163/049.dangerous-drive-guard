import { useAppStore, VehicleStatus, AlarmItem } from '@/store/app'

type WSMessage = {
  type: string
  timestamp: number
  data: any
  trace_id?: string
}

class WebSocketManager {
  private static instance: WebSocketManager
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private reconnectDelay = 3000
  private heartbeatInterval: NodeJS.Timeout | null = null
  private listeners: Map<string, Set<(data: any) => void>> = new Map()
  private url: string
  private connecting = false

  private constructor() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = import.meta.env.VITE_WS_BASE || `${protocol}//${window.location.host}`
    this.url = `${host}/api/v1/ws/monitor`
  }

  static getInstance(): WebSocketManager {
    if (!WebSocketManager.instance) {
      WebSocketManager.instance = new WebSocketManager()
    }
    return WebSocketManager.instance
  }

  connect() {
    if (this.ws || this.connecting) return
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] max reconnect attempts reached')
      return
    }

    this.connecting = true
    const token = localStorage.getItem('ddg_access_token') || ''
    const wsUrl = token
      ? `${this.url}?token=${encodeURIComponent(token)}`
      : this.url

    try {
      this.ws = new WebSocket(wsUrl)
    } catch (e) {
      this.connecting = false
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.connecting = false
      this.reconnectAttempts = 0
      useAppStore.getState().setWsConnected(true)
      this.startHeartbeat()
      this.emit('connected', null)
      console.log('[WS] connected')
    }

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.handleMessage(msg)
      } catch (e) {
        console.warn('[WS] parse error', e)
      }
    }

    this.ws.onerror = (err) => {
      console.error('[WS] error', err)
      this.connecting = false
    }

    this.ws.onclose = (event) => {
      this.connecting = false
      useAppStore.getState().setWsConnected(false)
      this.stopHeartbeat()
      this.emit('disconnected', event)
      console.log(`[WS] disconnected, code=${event.code}, reason=${event.reason}`)
      if (event.code !== 1000) {
        this.scheduleReconnect()
      }
    }
  }

  disconnect() {
    this.reconnectAttempts = this.maxReconnectAttempts
    this.stopHeartbeat()
    if (this.ws) {
      this.ws.close(1000, 'client disconnect')
      this.ws = null
    }
  }

  private handleMessage(msg: WSMessage) {
    const { type, data } = msg
    this.emit(type, data)

    switch (type) {
      case 'vehicle_status': {
        const v = data as VehicleStatus
        useAppStore.getState().upsertVehicle(v)
        break
      }
      case 'vehicle_track': {
        const track = data
        const vs: VehicleStatus | undefined = useAppStore.getState().vehicles.find(x => x.vehicle_id === track.vehicle_id)
        if (vs) {
          useAppStore.getState().upsertVehicle({
            ...vs,
            latitude: track.latitude,
            longitude: track.longitude,
            speed: track.speed,
            direction: track.direction,
            last_update_time: new Date().toISOString(),
          })
        }
        break
      }
      case 'alarm_notify': {
        const alarm = data as AlarmItem
        useAppStore.getState().addAlarm(alarm)
        this.emit('new_alarm', alarm)
        break
      }
      case 'driver_fatigue': {
        break
      }
      case 'sos_alert': {
        this.emit('sos_alert', data)
        break
      }
      case 'escort_polling': {
        this.emit('escort_polling', data)
        break
      }
      case 'heartbeat':
        break
      case 'error':
        console.error('[WS] server error', data)
        break
    }
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send('heartbeat', { time: Date.now() })
      }
    }, 30000)
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  private scheduleReconnect() {
    this.reconnectAttempts++
    const delay = Math.min(this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1), 60000)
    console.log(`[WS] reconnect in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
    setTimeout(() => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.connect()
      }
    }, delay)
  }

  send(type: string, data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, timestamp: Date.now(), data }))
    }
  }

  on(type: string, callback: (data: any) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set())
    }
    this.listeners.get(type)!.add(callback)
    return () => this.off(type, callback)
  }

  off(type: string, callback: (data: any) => void) {
    this.listeners.get(type)?.delete(callback)
  }

  private emit(type: string, data: any) {
    this.listeners.get(type)?.forEach(cb => {
      try {
        cb(data)
      } catch (e) {
        console.error(`[WS] listener error ${type}`, e)
      }
    })
  }
}

export default WebSocketManager
