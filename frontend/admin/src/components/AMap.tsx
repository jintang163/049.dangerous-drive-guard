import React, { useEffect, useRef, useState } from 'react'
import AMapLoader from '@amap/amap-jsapi-loader'
import { Spin, Empty, Tooltip } from 'antd'

interface MarkerData {
  position: [number, number]
  title: string
  color?: string
  icon?: string
  info?: any
  onClick?: () => void
}

interface PolylineData {
  path: [number, number][]
  color?: string
  weight?: number
  opacity?: number
  style?: 'solid' | 'dashed'
}

interface PolygonData {
  path: [number, number][]
  fillColor?: string
  strokeColor?: string
  strokeWeight?: number
  fillOpacity?: number
}

interface Props {
  style?: React.CSSProperties
  className?: string
  center?: [number, number]
  zoom?: number
  markers?: MarkerData[]
  polylines?: PolylineData[]
  polygons?: PolygonData[]
  showTraffic?: boolean
  showScale?: boolean
  showToolBar?: boolean
  onMarkerClick?: (marker: MarkerData) => void
  onMapClick?: (lng: number, lat: number) => void
  onMapLoaded?: (map: any, AMap: any) => void
}

const AMAP_KEY = import.meta.env.VITE_AMAP_KEY || 'demo_key'

declare global {
  interface Window {
    AMap: any
    _AMapSecurityConfig: any
  }
}

const AMap: React.FC<Props> = ({
  style,
  className,
  center = [116.4074, 39.9042],
  zoom = 11,
  markers = [],
  polylines = [],
  polygons = [],
  showTraffic = false,
  showScale = true,
  showToolBar = true,
  onMarkerClick,
  onMapClick,
  onMapLoaded,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<any>(null)
  const markersRef = useRef<any[]>([])
  const polylinesRef = useRef<any[]>([])
  const polygonsRef = useRef<any[]>([])
  const amapInstanceRef = useRef<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    window._AMapSecurityConfig = {
      securityJsCode: '',
    }

    let disposed = false
    AMapLoader.load({
      key: AMAP_KEY,
      version: '2.0',
      plugins: ['AMap.Scale', 'AMap.ToolBar', 'AMap.TrafficLayer', 'AMap.Marker', 'AMap.InfoWindow'],
    })
      .then((AMap) => {
        if (disposed || !containerRef.current) return
        amapInstanceRef.current = AMap
        const map = new AMap.Map(containerRef.current, {
          viewMode: '2D',
          zoom,
          center,
          resizeEnable: true,
          mapStyle: 'amap://styles/whitesmoke',
        })
        mapRef.current = map

        if (showScale) {
          map.addControl(new AMap.Scale())
        }
        if (showToolBar) {
          map.addControl(new AMap.ToolBar({ position: 'RB' }))
        }
        if (showTraffic) {
          const trafficLayer = new AMap.TrafficLayer()
          trafficLayer.setMap(map)
        }

        if (onMapClick) {
          map.on('click', (e: any) => {
            onMapClick(e.lnglat.getLng(), e.lnglat.getLat())
          })
        }

        setLoading(false)
        onMapLoaded?.(map, AMap)
      })
      .catch((err) => {
        console.error('[AMap] load error:', err)
        if (!disposed) {
          setError(true)
          setLoading(false)
        }
      })

    return () => {
      disposed = true
      if (mapRef.current) {
        mapRef.current.destroy()
        mapRef.current = null
      }
    }
  }, [])

  const clearMarkers = () => {
    markersRef.current.forEach(m => m.setMap(null))
    markersRef.current = []
  }

  const clearPolylines = () => {
    polylinesRef.current.forEach(p => p.setMap(null))
    polylinesRef.current = []
  }

  const clearPolygons = () => {
    polygonsRef.current.forEach(p => p.setMap(null))
    polygonsRef.current = []
  }

  useEffect(() => {
    if (!mapRef.current || !amapInstanceRef.current) return
    const AMap = amapInstanceRef.current
    clearMarkers()

    markers.forEach((m, idx) => {
      const marker = new AMap.Marker({
        position: m.position,
        title: m.title,
        anchor: 'bottom-center',
        content: createMarkerHTML(m, idx),
        offset: new AMap.Pixel(0, 0),
      })
      if (onMarkerClick || m.onClick) {
        marker.on('click', () => {
          if (m.onClick) m.onClick()
          onMarkerClick?.(m)
        })
      }
      marker.setMap(mapRef.current)
      markersRef.current.push(marker)
    })
  }, [markers])

  useEffect(() => {
    if (!mapRef.current || !amapInstanceRef.current) return
    const AMap = amapInstanceRef.current
    clearPolylines()

    polylines.forEach((pl) => {
      const polyline = new AMap.Polyline({
        path: pl.path,
        strokeColor: pl.color || '#1677ff',
        strokeWeight: pl.weight || 6,
        strokeOpacity: pl.opacity || 0.9,
        lineJoin: 'round',
        lineCap: 'round',
        strokeStyle: pl.style || 'solid',
        showDir: true,
      })
      polyline.setMap(mapRef.current)
      polylinesRef.current.push(polyline)
    })
  }, [polylines])

  useEffect(() => {
    if (!mapRef.current || !amapInstanceRef.current) return
    const AMap = amapInstanceRef.current
    clearPolygons()

    polygons.forEach((pg) => {
      const polygon = new AMap.Polygon({
        path: pg.path,
        strokeColor: pg.strokeColor || '#ff4d4f',
        strokeWeight: pg.strokeWeight || 2,
        strokeOpacity: 0.8,
        fillColor: pg.fillColor || '#ff4d4f',
        fillOpacity: pg.fillOpacity || 0.15,
      })
      polygon.setMap(mapRef.current)
      polygonsRef.current.push(polygon)
    })
  }, [polygons])

  useEffect(() => {
    if (mapRef.current) {
      mapRef.current.setZoomAndCenter(zoom, center)
    }
  }, [center, zoom])

  const createMarkerHTML = (m: MarkerData, idx: number) => {
    const color = m.color || '#1677ff'
    const statusLabel = m.info?.fatigue_level === 'fatigue' ? '疲劳' :
      m.info?.fatigue_level === 'warning' ? '预警' :
        m.info?.status === 'offline' ? '离线' : '正常'
    const statusColor = m.info?.fatigue_level === 'fatigue' ? '#ff4d4f' :
      m.info?.fatigue_level === 'warning' ? '#faad14' :
        m.info?.status === 'offline' ? '#8c8c8c' : '#52c41a'
    const plate = m.title || '车辆'
    return `
      <div style="position: relative; transform: translate(-50%, -100%); cursor: pointer; white-space: nowrap;">
        <div style="
          background: ${color};
          color: #fff;
          padding: 4px 10px;
          border-radius: 6px;
          font-size: 12px;
          font-weight: 600;
          box-shadow: 0 2px 8px rgba(0,0,0,0.2);
          display: flex;
          align-items: center;
          gap: 6px;
          margin-bottom: 6px;
        ">
          <span style="
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: ${statusColor};
            box-shadow: 0 0 0 2px rgba(255,255,255,0.3);
            animation: pulse 2s infinite;
          "></span>
          ${plate}
        </div>
        <div style="
          position: absolute;
          bottom: -8px;
          left: 50%;
          transform: translateX(-50%);
          width: 0;
          height: 0;
          border-left: 8px solid transparent;
          border-right: 8px solid transparent;
          border-top: 10px solid ${color};
        "></div>
        <div style="
          position: absolute;
          bottom: -14px;
          left: 50%;
          transform: translateX(-50%);
          width: 10px;
          height: 10px;
          background: ${color};
          border-radius: 50%;
          border: 2px solid #fff;
          box-shadow: 0 2px 4px rgba(0,0,0,0.2);
        "></div>
      </div>
    `
  }

  if (error) {
    return (
      <div
        ref={containerRef}
        style={{
          ...style,
          minHeight: 400,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: '#fafafa',
          borderRadius: 12,
          border: '1px dashed #d9d9d9',
          ...className ? {} : {},
        }}
        className={className}
      >
        <Empty description="地图加载失败，请检查AMap Key配置" />
      </div>
    )
  }

  return (
    <div style={{ position: 'relative', ...style }} className={className}>
      <div
        ref={containerRef}
        style={{
          width: '100%',
          height: '100%',
          minHeight: 300,
          borderRadius: 'inherit',
        }}
      />
      {loading && (
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'rgba(255,255,255,0.8)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 'inherit',
            zIndex: 10,
          }}
        >
          <Spin size="large" tip="地图加载中..." />
        </div>
      )}
      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; transform: scale(1); }
          50% { opacity: 0.6; transform: scale(1.5); }
        }
      `}</style>
    </div>
  )
}

export default AMap
