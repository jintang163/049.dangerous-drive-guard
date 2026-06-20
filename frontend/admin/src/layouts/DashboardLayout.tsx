import React, { useEffect, useMemo, useRef, useState } from 'react'
import {
  Typography,
  Space,
  Tag,
  Progress,
  Badge,
  List,
  Avatar,
  Button,
  Tooltip,
  Divider,
  Segmented,
} from 'antd'
import {
  SafetyCertificateOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  CarOutlined,
  FileTextOutlined,
  AlertOutlined,
  SafetyOutlined,
  FireOutlined,
  CloudOutlined,
  ThunderboltOutlined,
  WarningOutlined,
  EnvironmentOutlined,
  RiseOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  GlobalOutlined,
  RobotOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import AMap from '@/components/AMap'
import { useAppStore, VehicleStatus, AlarmItem } from '@/store/app'
import dayjs from 'dayjs'
import '@/styles/index.less'

const { Title, Text } = Typography

interface WeatherInfo {
  city: string
  temperature: number
  condition: string
  humidity: number
  wind: string
  aqi: number
  aqiLevel: string
}

interface WeatherAlert {
  id: number
  type: string
  level: '蓝色' | '黄色' | '橙色' | '红色'
  title: string
  content: string
  time: string
}

const DashboardLayout: React.FC = () => {
  const { vehicles, alarms, stats } = useAppStore()
  const [currentTime, setCurrentTime] = useState<string>(dayjs().format('YYYY年MM月DD日 HH:mm:ss dddd'))
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [vehicleFilter, setVehicleFilter] = useState<string>('all')
  const [systemStartTime] = useState<dayjs.Dayjs>(dayjs().subtract(12, 'day').subtract(8, 'hour'))
  const [newAlarmIds, setNewAlarmIds] = useState<Set<number>>(new Set())
  const containerRef = useRef<HTMLDivElement>(null)
  const lastAlarmCountRef = useRef<number>(0)

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(dayjs().format('YYYY年MM月DD日 HH:mm:ss dddd'))
    }, 1000)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    const curLen = alarms.length
    if (curLen > lastAlarmCountRef.current && lastAlarmCountRef.current > 0) {
      const newIds = new Set<number>()
      for (let i = 0; i < curLen - lastAlarmCountRef.current && i < alarms.length; i++) {
        newIds.add(alarms[i].id)
      }
      setNewAlarmIds(prev => {
        const merged = new Set([...prev, ...newIds])
        return merged
      })
      setTimeout(() => {
        setNewAlarmIds(prev => {
          const next = new Set(prev)
          newIds.forEach(id => next.delete(id))
          return next
        })
      }, 3000)
    }
    lastAlarmCountRef.current = curLen
  }, [alarms])

  const systemUptime = useMemo(() => {
    const diff = dayjs().diff(systemStartTime)
    const d = Math.floor(diff / 86400000)
    const h = Math.floor((diff % 86400000) / 3600000)
    const m = Math.floor((diff % 3600000) / 60000)
    return `${d}天 ${h}时 ${m}分`
  }, [systemStartTime, currentTime])

  const weatherInfo: WeatherInfo = useMemo(() => ({
    city: '北京',
    temperature: 26,
    condition: '多云转晴',
    humidity: 58,
    wind: '东南风3级',
    aqi: 72,
    aqiLevel: '良',
  }), [])

  const weatherAlerts: WeatherAlert[] = useMemo(() => [
    { id: 1, type: '暴雨', level: '黄色', title: '暴雨黄色预警', content: '预计未来6小时内部分地区降雨量将达50毫米以上', time: '2小时前' },
    { id: 2, type: '高温', level: '橙色', title: '高温橙色预警', content: '最高气温将升至37℃以上，请注意防暑降温', time: '5小时前' },
  ], [])

  const handleFullscreen = async () => {
    try {
      if (!document.fullscreenElement && containerRef.current) {
        await containerRef.current.requestFullscreen()
        setIsFullscreen(true)
      } else {
        await document.exitFullscreen()
        setIsFullscreen(false)
      }
    } catch (err) {
      console.error('全屏切换失败:', err)
    }
  }

  useEffect(() => {
    const onFsChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onFsChange)
    return () => document.removeEventListener('fullscreenchange', onFsChange)
  }, [])

  const runningVehicles = useMemo(() => vehicles.filter(v => v.status === 'running').length, [vehicles])
  const inTransitWaybills = useMemo(() => Math.floor(vehicles.length * 0.85), [vehicles])
  const todayAlarms = useMemo(() => alarms.filter(a => dayjs(a.created_at).isSame(dayjs(), 'day')).length, [alarms])
  const avgSafetyScore = useMemo(() => {
    if (!vehicles.length) return 0
    const sum = vehicles.reduce((acc, v) => acc + (v.fatigue_score || 85), 0)
    return Math.round(sum / vehicles.length)
  }, [vehicles])

  const trendOption = useMemo(() => {
    const hours = Array.from({ length: 24 }, (_, i) => `${i.toString().padStart(2, '0')}:00`)
    const gen = (base: number, amp: number) => hours.map((_, i) =>
      Math.max(0, Math.round(base + Math.sin(i / 3) * amp + Math.random() * amp * 0.5))
    )
    return {
      backgroundColor: 'transparent',
      grid: { left: 40, right: 16, top: 28, bottom: 30 },
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(10, 22, 40, 0.92)',
        borderColor: 'rgba(22, 119, 255, 0.4)',
        textStyle: { color: '#fff', fontSize: 12 },
      },
      legend: {
        data: ['在线车辆', '报警数'],
        textStyle: { color: 'rgba(255,255,255,0.7)', fontSize: 11 },
        top: 0,
        right: 0,
        itemWidth: 12,
        itemHeight: 8,
      },
      xAxis: {
        type: 'category',
        data: hours,
        axisLine: { lineStyle: { color: 'rgba(255,255,255,0.15)' } },
        axisLabel: { color: 'rgba(255,255,255,0.5)', fontSize: 10, interval: 3 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)', type: 'dashed' } },
        axisLabel: { color: 'rgba(255,255,255,0.5)', fontSize: 10 },
      },
      series: [
        {
          name: '在线车辆',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 4,
          showSymbol: false,
          data: gen(58, 18),
          lineStyle: { color: '#1677ff', width: 2, shadowColor: 'rgba(22,119,255,0.5)', shadowBlur: 8 },
          itemStyle: { color: '#1677ff' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(22,119,255,0.35)' },
                { offset: 1, color: 'rgba(22,119,255,0)' },
              ],
            },
          },
        },
        {
          name: '报警数',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 4,
          showSymbol: false,
          data: gen(5, 4),
          lineStyle: { color: '#ff4d4f', width: 2, shadowColor: 'rgba(255,77,79,0.5)', shadowBlur: 8 },
          itemStyle: { color: '#ff4d4f' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(255,77,79,0.25)' },
                { offset: 1, color: 'rgba(255,77,79,0)' },
              ],
            },
          },
        },
      ],
    }
  }, [])

  const fatiguePieOption = useMemo(() => {
    return {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(10, 22, 40, 0.92)',
        borderColor: 'rgba(22, 119, 255, 0.4)',
        textStyle: { color: '#fff', fontSize: 12 },
      },
      legend: {
        orient: 'vertical',
        right: 4,
        top: 'center',
        textStyle: { color: 'rgba(255,255,255,0.7)', fontSize: 11 },
        itemWidth: 10,
        itemHeight: 10,
      },
      series: [
        {
          type: 'pie',
          radius: ['42%', '68%'],
          center: ['32%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 6,
            borderColor: '#0a1628',
            borderWidth: 2,
          },
          label: { show: false },
          labelLine: { show: false },
          data: [
            { value: 38, name: '疲劳瞌睡', itemStyle: { color: '#ff4d4f' } },
            { value: 22, name: '频繁打哈欠', itemStyle: { color: '#fa8c16' } },
            { value: 15, name: '异常姿态', itemStyle: { color: '#faad14' } },
            { value: 12, name: '视线偏离', itemStyle: { color: '#a0d911' } },
            { value: 8, name: '使用手机', itemStyle: { color: '#13c2c2' } },
            { value: 5, name: '其他', itemStyle: { color: '#722ed1' } },
          ],
        },
      ],
    }
  }, [])

  const topRiskVehicles = useMemo(() => {
    return [...vehicles]
      .sort((a, b) => (a.fatigue_score || 100) - (b.fatigue_score || 100))
      .slice(0, 5)
  }, [vehicles])

  const vehicleMapData = useMemo(() => {
    let arr = vehicles
    if (vehicleFilter === 'running') arr = arr.filter(v => v.status === 'running')
    else if (vehicleFilter === 'warning') arr = arr.filter(v => v.fatigue_level === 'warning' || v.fatigue_level === 'fatigue')
    else if (vehicleFilter === 'fatigue') arr = arr.filter(v => v.fatigue_level === 'fatigue')
    return arr.map(v => ({
      position: [v.longitude || (116.3 + Math.random() * 0.2), v.latitude || (39.85 + Math.random() * 0.2)],
      title: v.plate_number,
      color: v.fatigue_level === 'fatigue' ? '#ff4d4f' : v.fatigue_level === 'warning' ? '#faad14' : v.status === 'running' ? '#52c41a' : '#8c8c8c',
      status: v.status,
      info: { ...v, fatigueLevel: v.fatigue_level },
    })).filter(v => v.position[0] && v.position[1])
  }, [vehicles, vehicleFilter])

  const routePolylines = useMemo(() => [
    {
      path: [[116.30, 39.86], [116.35, 39.88], [116.40, 39.90], [116.46, 39.92], [116.50, 39.90]] as [number, number][],
      color: '#52c41a',
      weight: 5,
      opacity: 0.85,
    },
    {
      path: [[116.32, 39.94], [116.38, 39.92], [116.42, 39.88], [116.48, 39.85]] as [number, number][],
      color: '#1677ff',
      weight: 5,
      opacity: 0.85,
    },
  ], [])

  const forbiddenZones = useMemo(() => [
    {
      path: [
        [116.36, 39.92], [116.40, 39.92], [116.40, 39.945], [116.36, 39.945],
      ] as [number, number][],
      fillColor: '#ff4d4f',
      strokeColor: '#ff4d4f',
      fillOpacity: 0.12,
      strokeWeight: 2,
    },
    {
      path: [
        [116.44, 39.87], [116.475, 39.87], [116.475, 39.89], [116.44, 39.89],
      ] as [number, number][],
      fillColor: '#faad14',
      strokeColor: '#faad14',
      fillOpacity: 0.1,
      strokeWeight: 2,
    },
  ], [])

  const getLevelColor = (level: number) => level === 3 ? '#ff4d4f' : level === 2 ? '#fa8c16' : '#1677ff'
  const getLevelLabel = (level: number) => level === 3 ? '严重' : level === 2 ? '警告' : '提醒'

  const PanelCard: React.FC<{
    title: React.ReactNode
    icon?: React.ReactNode
    extra?: React.ReactNode
    children: React.ReactNode
    style?: React.CSSProperties
  }> = ({ title, icon, extra, children, style }) => (
    <div className="panel-card" style={style}>
      <div className="panel-card-inner">
        <div className="panel-card-header">
          <div className="panel-card-title">
            {icon && <span className="panel-card-icon">{icon}</span>}
            <span>{title}</span>
          </div>
          {extra && <div className="panel-card-extra">{extra}</div>}
        </div>
        <div className="panel-card-body">{children}</div>
      </div>
    </div>
  )

  const StatBlock: React.FC<{
    label: string
    value: number | string
    suffix?: string
    color: string
    icon: React.ReactNode
    decimal?: number
  }> = ({ label, value, suffix, color, icon, decimal = 0 }) => {
    const [display, setDisplay] = useState<number>(0)
    const targetVal = typeof value === 'string' ? parseFloat(value) || 0 : value
    useEffect(() => {
      let raf: number
      const duration = 1200
      const start = performance.now()
      const startVal = 0
      const step = (ts: number) => {
        const p = Math.min((ts - start) / duration, 1)
        const ease = 1 - Math.pow(1 - p, 3)
        setDisplay(startVal + (targetVal - startVal) * ease)
        if (p < 1) raf = requestAnimationFrame(step)
      }
      raf = requestAnimationFrame(step)
      return () => cancelAnimationFrame(raf)
    }, [targetVal])
    const showVal = decimal > 0 ? display.toFixed(decimal) : Math.round(display).toString()
    return (
      <div className="stat-block">
        <div className="stat-icon" style={{ background: `${color}22`, color }}>
          {icon}
        </div>
        <div className="stat-content">
          <div className="stat-label">{label}</div>
          <div className="stat-value-row">
            <span className="stat-value" style={{ color }}>
              {showVal}
            </span>
            {suffix && <span className="stat-suffix">{suffix}</span>}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="big-screen" ref={containerRef}>
      <div className="big-screen-bg">
        <div className="bg-grid" />
        <div className="bg-scan" />
      </div>

      <header className="big-screen-header">
        <div className="header-left">
          <div className="logo-wrap">
            <SafetyCertificateOutlined className="logo-icon" />
          </div>
          <div className="header-title-wrap">
            <Title level={3} className="header-title">
              危险品运输安全监控 · 指挥大屏
            </Title>
            <Text className="header-subtitle">
              Dangerous Drive Guard · Real-time Command Center
            </Text>
          </div>
        </div>

        <div className="header-center">
          <ClockCircleOutlined />
          <span className="header-time">{currentTime}</span>
        </div>

        <div className="header-right">
          <div className="weather-card">
            <CloudOutlined style={{ color: '#52c41a', marginRight: 6 }} />
            <span className="weather-city">{weatherInfo.city}</span>
            <span className="weather-temp">{weatherInfo.temperature}°C</span>
            <span className="weather-cond">{weatherInfo.condition}</span>
            <Tag color={weatherInfo.aqi < 50 ? 'green' : weatherInfo.aqi < 100 ? 'lime' : weatherInfo.aqi < 150 ? 'orange' : 'red'} style={{ marginLeft: 6 }}>
              AQI {weatherInfo.aqi} {weatherInfo.aqiLevel}
            </Tag>
          </div>
          <Tooltip title={isFullscreen ? '退出全屏' : '全屏显示'}>
            <Button
              type="text"
              icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
              onClick={handleFullscreen}
              className="header-btn"
            />
          </Tooltip>
        </div>
      </header>

      <main className="big-screen-main">
        <aside className="big-screen-col left">
          <PanelCard
            title="今日运行指标"
            icon={<FireOutlined />}
            extra={<Tag color="blue">实时</Tag>}
          >
            <div className="stat-grid">
              <StatBlock
                label="车辆在线数"
                value={runningVehicles || 76}
                suffix={`/ ${stats?.total_vehicles || 92}`}
                color="#1677ff"
                icon={<CarOutlined />}
              />
              <StatBlock
                label="运输中运单"
                value={inTransitWaybills || 83}
                suffix="单"
                color="#52c41a"
                icon={<FileTextOutlined />}
              />
              <StatBlock
                label="今日报警数"
                value={todayAlarms || 14}
                suffix="起"
                color="#ff4d4f"
                icon={<AlertOutlined />}
              />
              <StatBlock
                label="安全评分均值"
                value={avgSafetyScore || 87.6}
                suffix="分"
                color="#faad14"
                icon={<SafetyOutlined />}
                decimal={1}
              />
            </div>
          </PanelCard>

          <PanelCard
            title="近24小时车辆趋势"
            icon={<RiseOutlined />}
            extra={<Tag color="cyan" style={{ margin: 0 }}>24H</Tag>}
            style={{ marginTop: 12 }}
          >
            <ReactECharts
              option={trendOption}
              style={{ height: 210, width: '100%' }}
              notMerge
              lazyUpdate
            />
          </PanelCard>

          <PanelCard
            title="高风险车辆 TOP 5"
            icon={<WarningOutlined />}
            extra={<Tag color="red">风险</Tag>}
            style={{ marginTop: 12, flex: 1, minHeight: 0 }}
          >
            <List
              dataSource={topRiskVehicles.length ? topRiskVehicles : mockTopRisk()}
              renderItem={(v, idx) => {
                const riskPct = 100 - Math.round(v.fatigue_score || 80)
                const riskColor = riskPct >= 60 ? '#ff4d4f' : riskPct >= 40 ? '#fa8c16' : '#faad14'
                return (
                  <List.Item className="top-risk-item">
                    <div className="top-risk-rank" style={{ background: idx < 2 ? 'linear-gradient(135deg, #ff4d4f, #ff7875)' : idx < 4 ? 'linear-gradient(135deg, #fa8c16, #ffa940)' : 'linear-gradient(135deg, #1677ff, #4096ff)' }}>
                      {idx + 1}
                    </div>
                    <div className="top-risk-main">
                      <div className="top-risk-row">
                        <Space size={6}>
                          <Text strong className="top-risk-plate">{v.plate_number}</Text>
                          <Badge
                            status={v.fatigue_level === 'fatigue' ? 'error' : v.fatigue_level === 'warning' ? 'warning' : 'success'}
                            text={<Text type="secondary" style={{ fontSize: 11 }}>{v.driver_name}</Text>}
                          />
                        </Space>
                        <Text strong style={{ color: riskColor, fontSize: 13 }}>
                          {riskPct}%
                        </Text>
                      </div>
                      <Progress
                        percent={riskPct}
                        showInfo={false}
                        strokeColor={riskColor}
                        trailColor="rgba(255,255,255,0.08)"
                        size="small"
                        className="top-risk-progress"
                      />
                    </div>
                  </List.Item>
                )
              }}
            />
          </PanelCard>
        </aside>

        <section className="big-screen-col center">
          <div className="map-top-bar">
            <div className="map-stats">
              <div className="map-stat">
                <span className="map-stat-label">当前车辆总数</span>
                <span className="map-stat-value">{vehicles.length || 92}</span>
                <span className="map-stat-unit">辆</span>
              </div>
              <Divider type="vertical" className="map-stat-divider" />
              <div className="map-stat">
                <span className="map-stat-label">正常</span>
                <span className="map-stat-dot" style={{ background: '#52c41a' }} />
                <span className="map-stat-value" style={{ color: '#52c41a' }}>
                  {vehicles.filter(v => v.fatigue_level === 'normal').length || 74}
                </span>
              </div>
              <div className="map-stat">
                <span className="map-stat-label">预警</span>
                <span className="map-stat-dot" style={{ background: '#faad14' }} />
                <span className="map-stat-value" style={{ color: '#faad14' }}>
                  {vehicles.filter(v => v.fatigue_level === 'warning').length || 13}
                </span>
              </div>
              <div className="map-stat">
                <span className="map-stat-label">疲劳</span>
                <span className="map-stat-dot pulse-dot" style={{ background: '#ff4d4f' }} />
                <span className="map-stat-value" style={{ color: '#ff4d4f' }}>
                  {vehicles.filter(v => v.fatigue_level === 'fatigue').length || 5}
                </span>
              </div>
            </div>
            <Segmented
              value={vehicleFilter}
              onChange={(v) => setVehicleFilter(v as string)}
              size="small"
              options={[
                { label: '全部', value: 'all' },
                { label: '在途', value: 'running' },
                { label: '预警/疲劳', value: 'warning' },
                { label: '仅疲劳', value: 'fatigue' },
              ]}
              className="map-filter"
            />
          </div>

          <div className="map-legend">
            <Space size={20} wrap>
              <Space size={6}><span className="legend-poly" style={{ background: '#52c41a' }} /> 规划路线</Space>
              <Space size={6}><span className="legend-poly" style={{ background: '#1677ff' }} /> 备用路线</Space>
              <Space size={6}><span className="legend-zone" style={{ background: 'rgba(255,77,79,0.25)', borderColor: '#ff4d4f' }} /> 禁行区(核心)</Space>
              <Space size={6}><span className="legend-zone" style={{ background: 'rgba(250,173,20,0.25)', borderColor: '#faad14' }} /> 慎行区(学校)</Space>
            </Space>
          </div>

          <div className="map-container">
            <AMap
              style={{ width: '100%', height: '100%', borderRadius: 0 }}
              center={[116.4074, 39.9042]}
              zoom={11}
              markers={vehicleMapData.length ? vehicleMapData : mockMapMarkers()}
              polylines={routePolylines}
              polygons={forbiddenZones}
              showTraffic
              showScale
              showToolBar
            />
            <div className="map-copyright">
              <EnvironmentOutlined /> 高德地图提供技术支持 · 数据每3秒刷新
            </div>
          </div>
        </section>

        <aside className="big-screen-col right">
          <PanelCard
            title="实时报警滚动"
            icon={<AlertOutlined />}
            extra={
              <Badge
                count={alarms.length || 23}
                showZero
                style={{ backgroundColor: '#ff4d4f', boxShadow: '0 0 8px rgba(255,77,79,0.6)' }}
              />
            }
          >
            <div className="alarm-scroll">
              <List
                dataSource={alarms.length ? alarms.slice(0, 12) : mockAlarms()}
                renderItem={(alarm: AlarmItem & {_mock?: boolean}) => {
                  const isNew = newAlarmIds.has(alarm.id)
                  return (
                    <List.Item
                      className={`alarm-item ${isNew ? 'alarm-new' : ''}`}
                    >
                      <Avatar
                        size={32}
                        style={{
                          background: getLevelColor(alarm.alarm_level),
                          flexShrink: 0,
                          boxShadow: `0 0 10px ${getLevelColor(alarm.alarm_level)}66`,
                        }}
                        icon={<WarningOutlined />}
                      />
                      <div className="alarm-content">
                        <div className="alarm-row1">
                          <Space size={6} wrap>
                            <Text strong className="alarm-plate">{alarm.vehicle_plate}</Text>
                            <Tag color={getLevelColor(alarm.alarm_level)} style={{ margin: 0, fontSize: 11 }}>
                              {getLevelLabel(alarm.alarm_level)}
                            </Tag>
                            <Text className="alarm-type">{getAlarmTypeLabel(alarm.alarm_type)}</Text>
                          </Space>
                        </div>
                        <div className="alarm-row2">
                          <EnvironmentOutlined />
                          <Text ellipsis className="alarm-loc">
                            {alarm.location_address || '北京市朝阳区建国路附近'}
                          </Text>
                        </div>
                        <div className="alarm-row3">
                          <ClockCircleOutlined />
                          <span>{dayjs(alarm.created_at).format('HH:mm:ss')}</span>
                          <span className="alarm-score">
                            疲劳 <Text strong style={{ color: getLevelColor(alarm.alarm_level) }}>
                              {Math.round(alarm.fatigue_score)}
                            </Text>
                          </span>
                        </div>
                      </div>
                    </List.Item>
                  )
                }}
              />
            </div>
          </PanelCard>

          <PanelCard
            title="疲劳检测分布"
            icon={<RobotOutlined />}
            extra={<Tag color="purple">AI识别</Tag>}
            style={{ marginTop: 12 }}
          >
            <ReactECharts
              option={fatiguePieOption}
              style={{ height: 210, width: '100%' }}
              notMerge
            />
          </PanelCard>

          <PanelCard
            title="天气预警"
            icon={<ThunderboltOutlined />}
            extra={<Tag color="orange">气象</Tag>}
            style={{ marginTop: 12 }}
          >
            <div className="weather-alert-list">
              {weatherAlerts.map(w => (
                <div
                  key={w.id}
                  className={`weather-alert-item alert-${w.level}`}
                >
                  <div className="alert-level-tag">
                    {w.type}{w.level}
                  </div>
                  <div className="alert-info">
                    <div className="alert-title">{w.title}</div>
                    <div className="alert-content">{w.content}</div>
                    <div className="alert-time">
                      <ClockCircleOutlined /> {w.time}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </PanelCard>
        </aside>
      </main>

      <footer className="big-screen-footer">
        <div className="footer-item">
          <SyncOutlined spin style={{ color: '#52c41a' }} />
          <span>系统运行：{systemUptime}</span>
        </div>
        <div className="footer-divider" />
        <div className="footer-item">
          <ClockCircleOutlined />
          <span>数据更新：{dayjs().format('YYYY-MM-DD HH:mm:ss')}</span>
        </div>
        <div className="footer-divider" />
        <div className="footer-item footer-stack">
          <GlobalOutlined />
          <span>CloudWeGo · 高德地图</span>
          <span className="footer-sep">·</span>
          <RobotOutlined />
          <span>AI疲劳检测</span>
          <span className="footer-sep">·</span>
          <LinkOutlined />
          <span>区块链存证</span>
        </div>
        <div className="footer-divider" />
        <div className="footer-item">
          <SafetyCertificateOutlined style={{ color: '#1677ff' }} />
          <span>DDG v1.0.0 © 危险品运输安全监控平台</span>
        </div>
      </footer>
    </div>
  )
}

const getAlarmTypeLabel = (type: string): string => {
  const map: Record<string, string> = {
    fatigue_perclos: '瞌睡',
    excessive_yawn: '频繁哈欠',
    abnormal_head_posture: '姿态异常',
    gaze_distraction: '视线偏离',
    phone_usage: '使用手机',
    smoking: '抽烟',
    no_seatbelt: '未系安全带',
    continuous_fatigue: '连续疲劳',
  }
  return map[type] || '异常行为'
}

const mockTopRisk = (): VehicleStatus[] => [
  { vehicle_id: 101, plate_number: '京A·88X01', driver_name: '张建国', fatigue_score: 38, fatigue_level: 'fatigue', status: 'running' } as VehicleStatus,
  { vehicle_id: 102, plate_number: '京B·7F203', driver_name: '李志强', fatigue_score: 45, fatigue_level: 'fatigue', status: 'running' } as VehicleStatus,
  { vehicle_id: 103, plate_number: '京C·3K518', driver_name: '王海峰', fatigue_score: 56, fatigue_level: 'warning', status: 'running' } as VehicleStatus,
  { vehicle_id: 104, plate_number: '京E·6M982', driver_name: '赵立新', fatigue_score: 62, fatigue_level: 'warning', status: 'idle' } as VehicleStatus,
  { vehicle_id: 105, plate_number: '京F·2D456', driver_name: '刘振华', fatigue_score: 71, fatigue_level: 'warning', status: 'running' } as VehicleStatus,
]

const mockMapMarkers = () => [
  { position: [116.325, 39.875] as [number, number], title: '京A·88X01', color: '#ff4d4f', status: 'running', info: { plate_number: '京A·88X01', fatigue_level: 'fatigue', fatigue_score: 38 } },
  { position: [116.365, 39.892] as [number, number], title: '京B·7F203', color: '#ff4d4f', status: 'running', info: { plate_number: '京B·7F203', fatigue_level: 'fatigue', fatigue_score: 45 } },
  { position: [116.388, 39.908] as [number, number], title: '京C·3K518', color: '#faad14', status: 'running', info: { plate_number: '京C·3K518', fatigue_level: 'warning', fatigue_score: 56 } },
  { position: [116.422, 39.885] as [number, number], title: '京E·6M982', color: '#faad14', status: 'idle', info: { plate_number: '京E·6M982', fatigue_level: 'warning', fatigue_score: 62 } },
  { position: [116.455, 39.915] as [number, number], title: '京F·2D456', color: '#faad14', status: 'running', info: { plate_number: '京F·2D456', fatigue_level: 'warning', fatigue_score: 71 } },
  { position: [116.338, 39.928] as [number, number], title: '京G·9H001', color: '#52c41a', status: 'running', info: { plate_number: '京G·9H001', fatigue_level: 'normal', fatigue_score: 92 } },
  { position: [116.400, 39.938] as [number, number], title: '京H·5P332', color: '#52c41a', status: 'running', info: { plate_number: '京H·5P332', fatigue_level: 'normal', fatigue_score: 95 } },
  { position: [116.478, 39.880] as [number, number], title: '京J·8A774', color: '#52c41a', status: 'running', info: { plate_number: '京J·8A774', fatigue_level: 'normal', fatigue_score: 88 } },
  { position: [116.490, 39.902] as [number, number], title: '京K·6C209', color: '#52c41a', status: 'running', info: { plate_number: '京K·6C209', fatigue_level: 'normal', fatigue_score: 90 } },
]

const mockAlarms = (): any[] => {
  const types = ['fatigue_perclos', 'excessive_yawn', 'abnormal_head_posture', 'gaze_distraction', 'phone_usage', 'smoking']
  const plates = ['京A·88X01', '京B·7F203', '京C·3K518', '京E·6M982', '京F·2D456', '京G·9H001', '京H·5P332']
  const drivers = ['张建国', '李志强', '王海峰', '赵立新', '刘振华', '陈文博', '孙晓峰']
  const locs = ['朝阳区建国路88号附近', '海淀区中关村大街', '丰台区南三环西路', '通州区新华大街', '东城区东直门外', '西城区复兴门内', '昌平区回龙观东大街']
  return Array.from({ length: 10 }, (_, i) => {
    const level = (i < 3 ? 3 : i < 7 ? 2 : 1) as 1 | 2 | 3
    const score = level === 3 ? 30 + Math.floor(Math.random() * 20) : level === 2 ? 50 + Math.floor(Math.random() * 25) : 70 + Math.floor(Math.random() * 20)
    return {
      id: 9000 + i,
      alarm_no: `AL${Date.now()}_${i}`,
      vehicle_plate: plates[i % plates.length],
      driver_name: drivers[i % drivers.length],
      alarm_type: types[i % types.length],
      alarm_level: level,
      fatigue_score: score,
      location_address: locs[i % locs.length],
      latitude: 39.85 + Math.random() * 0.1,
      longitude: 116.3 + Math.random() * 0.2,
      vehicle_speed: 30 + Math.floor(Math.random() * 50),
      created_at: dayjs().subtract(i * 5 + Math.floor(Math.random() * 4), 'minute').toISOString(),
      status: i < 2 ? 'pending' : i < 6 ? 'processing' : 'acknowledged',
    }
  })
}

export default DashboardLayout
