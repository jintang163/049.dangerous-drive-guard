const TOKEN_KEY = 'ddg_access_token'
const USER_KEY = 'ddg_user_info'
const EXPIRE_KEY = 'ddg_token_expire'
const PERMS_KEY = 'ddg_permissions'

export const getToken = () => localStorage.getItem(TOKEN_KEY) || ''

export const setToken = (token: string, expireAt?: Date) => {
  localStorage.setItem(TOKEN_KEY, token)
  if (expireAt) {
    localStorage.setItem(EXPIRE_KEY, expireAt.toISOString())
  }
}

export const getTokenExpire = () => {
  const raw = localStorage.getItem(EXPIRE_KEY)
  return raw ? new Date(raw) : null
}

export const isTokenExpired = () => {
  const expire = getTokenExpire()
  return !expire || expire.getTime() < Date.now()
}

export const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(EXPIRE_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(PERMS_KEY)
}

export const getUserInfo = () => {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? JSON.parse(raw) : null
}

export const setUserInfo = (user: any) => {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export const getPermissions = () => {
  const raw = localStorage.getItem(PERMS_KEY)
  return raw ? JSON.parse(raw) : []
}

export const setPermissions = (perms: string[]) => {
  localStorage.setItem(PERMS_KEY, JSON.stringify(perms))
}

export const hasPermission = (perm: string) => {
  const perms = getPermissions()
  return perms.includes(perm) || perms.includes('*')
}

export const formatDateTime = (date: string | Date, fmt = 'YYYY-MM-DD HH:mm:ss') => {
  if (!date) return ''
  const d = typeof date === 'string' ? new Date(date) : date
  const pad = (n: number) => n.toString().padStart(2, '0')
  return fmt
    .replace('YYYY', d.getFullYear().toString())
    .replace('MM', pad(d.getMonth() + 1))
    .replace('DD', pad(d.getDate()))
    .replace('HH', pad(d.getHours()))
    .replace('mm', pad(d.getMinutes()))
    .replace('ss', pad(d.getSeconds()))
}

export const formatDistance = (meters: number) => {
  if (!meters) return '0 km'
  if (meters < 1000) return `${Math.round(meters)} m`
  return `${(meters / 1000).toFixed(2)} km`
}

export const formatDuration = (seconds: number) => {
  if (!seconds) return '0分钟'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h) return `${h}小时${m}分钟`
  return `${m}分钟`
}

export const debounce = <T extends (...args: any[]) => any>(fn: T, delay = 300) => {
  let timer: NodeJS.Timeout
  return (...args: Parameters<T>) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

export const throttle = <T extends (...args: any[]) => any>(fn: T, delay = 300) => {
  let last = 0
  return (...args: Parameters<T>) => {
    const now = Date.now()
    if (now - last >= delay) {
      last = now
      fn(...args)
    }
  }
}

export const generateId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
