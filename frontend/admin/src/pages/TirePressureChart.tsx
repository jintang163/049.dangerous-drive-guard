import React, { useEffect, useState, useCallback, useMemo } from 'react'
import {
  Row,
  Col,
  Card,
  Select,
  DatePicker,
  Radio,
  Button,
  Space,
  Typography,
  Statistic,
  Tag,
  Spin,
  message,
  Empty,
} from 'antd'
import {
  CarOutlined,
  DownloadOutlined,
  ReloadOutlined,
  DashboardOutlined,
  FireOutlined,
  WarningOutlined,
  MinusOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs, { Dayjs } from 'dayjs'
import { vehicleApi, WheelChartData } from '@/services/api'
import { useAppStore, VehicleItem } from '@/store/app'

const { Title, Text } = Typography
const { Option } = Select
const { RangePicker } = DatePicker

type Granularity = 'raw' | 'minute' | 'hour' | 'day'

const WHEEL_NAMES: Record<string, { label: string; color: string }> = {
  fl: { label: '前轮左 (FL)', color: '#1677ff' },
  fr: { label: '前轮右 (FR)', color: '#52c41a' },
  rl: { label: '后轮左 (RL)', color: '#fa8c16' },
  rr: { label: '后轮右 (RR)', color: '#722ed1' },
}

const WHEEL_KEYS = ['fl', 'fr', 'rl', 'rr'] as const

const toSeriesData = (arr?: { time: string; value: number }[]) => {
  if (!arr || !Array.isArray(arr)) return []
  return arr
    .filter(p => p && p.time && typeof p.value === 'number')
    .map(p => [p.time, p.value])
}

interface ChartConfig {
  title: string
  icon: React.ReactNode
  yLabel: string
  unit: string
  markLines?: Array<{
    yAxis: number
    label: string
    color: string
    lineStyle?: any
  }>
  warnCondition?: (v: number) => boolean
}

const TirePressureChart: React.FC = () => {
  const { vehicleList, fetchVehiclesList } = useAppStore()
  const [vehicleId, setVehicleId] = useState<number | undefined>()
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null]>([
    dayjs().subtract(7, 'day'),
    dayjs(),
  ])
  const [granularity, setGranularity] = useState<Granularity>('hour')

  const [pressureData, setPressureData] = useState<WheelChartData | null>(null)
  const [tempData, setTempData] = useState<WheelChartData | null>(null)
  const [brakeData, setBrakeData] = useState<WheelChartData | null>(null)
  const [loading, setLoading] = useState(false)
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (vehicleList.length === 0) {
      fetchVehiclesList({ page_size: 1000 }).catch(() => {})
    }
  }, [fetchVehiclesList, vehicleList.length])

  useEffect(() => {
    if (!vehicleId && vehicleList.length > 0) {
      const available = vehicleList.find((v: VehicleItem) => v.status === 'online') || vehicleList[0]
      if (available) setVehicleId(available.id)
    }
  }, [vehicleList, vehicleId])

  const buildParams = useCallback(() => {
    const [start, end] = dateRange
    return {
      start_time: start ? start.toISOString() : undefined,
      end_time: end ? end.toISOString() : undefined,
      granularity,
    }
  }, [dateRange, granularity])

  const fetchCharts = useCallback(async () => {
    if (!vehicleId) return
    setLoading(true)
    const params = buildParams()
    try {
      const [p, t, b] = await Promise.all([
        vehicleApi.tirePressureChart(vehicleId, params).catch(() => null),
        vehicleApi.tireTempChart(vehicleId, params).catch(() => null),
        vehicleApi.brakeTempChart(vehicleId, params).catch(() => null),
      ])
      setPressureData(p as any)
      setTempData(t as any)
      setBrakeData(b as any)
    } catch (e: any) {
      message.error(e?.message || '加载图表数据失败')
    } finally {
      setLoading(false)
    }
  }, [vehicleId, buildParams])

  useEffect(() => {
    if (vehicleId) {
      fetchCharts()
    }
  }, [fetchCharts, vehicleId])

  const handleExport = async () => {
    if (!vehicleId) {
      message.warning('请先选择车辆')
      return
    }
    setExporting(true)
    try {
      const [start, end] = dateRange
      const blob = await vehicleApi.exportDiagnostics(vehicleId, {
        start_time: start ? start.toISOString() : undefined,
        end_time: end ? end.toISOString() : undefined,
        type: 'tire_brake',
      })
      const url = window.URL.createObjectURL(blob as Blob)
      const a = document.createElement('a')
      a.href = url
      const plate = vehicleList.find((v: VehicleItem) => v.id === vehicleId)?.plate_number || vehicleId
      a.download = `diagnostics_${plate}_${dayjs().format('YYYYMMDD_HHmmss')}.csv`
      a.click()
      window.URL.revokeObjectURL(url)
      message.success('导出成功')
    } catch (e: any) {
      message.error(e?.message || '导出失败')
    } finally {
      setExporting(false)
    }
  }

  const buildChartOption = (data: WheelChartData | null, config: ChartConfig) => {
    const series = WHEEL_KEYS.map(key => {
      const meta = WHEEL_NAMES[key]
      return {
        name: meta.label,
        type: 'line',
        showSymbol: false,
        smooth: true,
        sampling: 'lttb',
        lineStyle: { width: 2, color: meta.color },
        itemStyle: { color: meta.color },
        data: toSeriesData(data?.[key]),
      }
    })

    const markLine = config.markLines && config.markLines.length > 0 ? {
      silent: true,
      symbol: ['none', 'none'],
      data: config.markLines.map(ml => ({
        yAxis: ml.yAxis,
        label: {
          formatter: ml.label,
          color: ml.color,
          fontSize: 11,
        },
        lineStyle: {
          color: ml.color,
          type: 'dashed',
          width: 1.5,
          ...ml.lineStyle,
        },
      })),
    } : undefined

    if (markLine) {
      series[0] = { ...series[0], markLine }
    }

    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
        valueFormatter: (v: any) => (typeof v === 'number' ? `${v.toFixed(2)} ${config.unit}` : v),
      },
      legend: {
        top: 0,
        type: 'scroll',
      },
      grid: {
        left: 60,
        right: 30,
        top: 40,
        bottom: 50,
      },
      xAxis: {
        type: 'time',
        axisLabel: { fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        name: `${config.yLabel} (${config.unit})`,
        nameTextStyle: { fontSize: 11, color: '#8c8c8c' },
        axisLabel: { fontSize: 11 },
        splitLine: { lineStyle: { type: 'dashed', color: '#f0f0f0' } },
      },
      dataZoom: [
        { type: 'inside', start: 0, end: 100 },
        { type: 'slider', start: 0, end: 100, height: 20, bottom: 10 },
      ],
      series,
    }
  }

  const pressureChart = useMemo(() => buildChartOption(pressureData, {
    title: '胎压历史曲线',
    icon: <DashboardOutlined style={{ color: '#1677ff' }} />,
    yLabel: '胎压',
    unit: 'bar',
    markLines: [
      { yAxis: 7, label: '下限 7bar', color: '#fa8c16' },
      { yAxis: 10, label: '上限 10bar', color: '#ff4d4f' },
    ],
    warnCondition: (v) => v < 7 || v > 10,
  }), [pressureData])

  const tempChart = useMemo(() => buildChartOption(tempData, {
    title: '胎温历史曲线',
    icon: <WarningOutlined style={{ color: '#fa8c16' }} />,
    yLabel: '胎温',
    unit: '℃',
    markLines: [
      { yAxis: 75, label: '告警线 75℃', color: '#ff4d4f' },
    ],
    warnCondition: (v) => v >= 75,
  }), [tempData])

  const brakeChart = useMemo(() => buildChartOption(brakeData, {
    title: '刹车片温度历史曲线',
    icon: <FireOutlined style={{ color: '#ff4d4f' }} />,
    yLabel: '刹车片温度',
    unit: '℃',
    markLines: [
      { yAxis: 250, label: '告警线 250℃', color: '#ff4d4f' },
    ],
    warnCondition: (v) => v >= 250,
  }), [brakeData])

  const renderStats = (data: WheelChartData | null, unit: string, warnCondition?: (v: number) => boolean) => {
    const stats = data?.stats
    const allValues: number[] = []
    WHEEL_KEYS.forEach(k => {
      const arr = data?.[k]
      if (Array.isArray(arr)) {
        arr.forEach(p => {
          if (p && typeof p.value === 'number') allValues.push(p.value)
        })
      }
    })

    const fallbackMax = allValues.length > 0 ? Math.max(...allValues) : undefined
    const fallbackMin = allValues.length > 0 ? Math.min(...allValues) : undefined
    const fallbackAvg = allValues.length > 0 ? allValues.reduce((a, b) => a + b, 0) / allValues.length : undefined
    const fallbackWarn = warnCondition && allValues.length > 0 ? allValues.filter(warnCondition).length : undefined

    const max = stats?.max ?? fallbackMax
    const min = stats?.min ?? fallbackMin
    const avg = stats?.avg ?? fallbackAvg
    const warnCount = stats?.warn_count ?? fallbackWarn

    return (
      <Row gutter={[12, 12]} style={{ marginTop: 12 }}>
        <Col xs={12} sm={6}>
          <Card size="small" style={{ borderRadius: 8, border: 'none', background: '#f6ffed' }}>
            <Statistic
              title={<Text type="secondary" style={{ fontSize: 12 }}>最大值</Text>}
              value={max}
              precision={2}
              suffix={unit}
              valueStyle={{ fontSize: 18, color: '#389e0d' }}
              prefix={<PlusOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" style={{ borderRadius: 8, border: 'none', background: '#e6f4ff' }}>
            <Statistic
              title={<Text type="secondary" style={{ fontSize: 12 }}>最小值</Text>}
              value={min}
              precision={2}
              suffix={unit}
              valueStyle={{ fontSize: 18, color: '#0958d9' }}
              prefix={<MinusOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" style={{ borderRadius: 8, border: 'none', background: '#f9f0ff' }}>
            <Statistic
              title={<Text type="secondary" style={{ fontSize: 12 }}>平均值</Text>}
              value={avg}
              precision={2}
              suffix={unit}
              valueStyle={{ fontSize: 18, color: '#531dab' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" style={{ borderRadius: 8, border: 'none', background: '#fff1f0' }}>
            <Statistic
              title={<Text type="secondary" style={{ fontSize: 12 }}>告警次数</Text>}
              value={warnCount}
              valueStyle={{ fontSize: 18, color: '#cf1322' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
      </Row>
    )
  }

  const renderChartCard = (title: string, icon: React.ReactNode, option: any, data: WheelChartData | null, unit: string, warnCondition?: (v: number) => boolean) => {
    const hasData = WHEEL_KEYS.some(k => Array.isArray(data?.[k]) && (data?.[k]?.length ?? 0) > 0)
    return (
      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            {icon}
            <Text strong style={{ fontSize: 15 }}>{title}</Text>
          </Space>
        }
      >
        <Spin spinning={loading}>
          {hasData || data?.stats ? (
            <>
              <ReactECharts option={option} style={{ height: 320 }} notMerge />
              {renderStats(data, unit, warnCondition)}
            </>
          ) : (
            <div style={{ padding: '60px 0' }}>
              <Empty description="暂无数据" />
            </div>
          )}
        </Spin>
      </Card>
    )
  }

  const currentVehicle = vehicleList.find((v: VehicleItem) => v.id === vehicleId)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <CarOutlined style={{ color: '#1677ff', fontSize: 18 }} />
            <Title level={4} style={{ margin: 0 }}>轮胎与刹车监控</Title>
            {currentVehicle && (
              <Tag color="blue" style={{ marginLeft: 8 }}>
                {currentVehicle.plate_number}
                {currentVehicle.status && (
                  <Tag
                    color={currentVehicle.status === 'online' ? 'green' : currentVehicle.status === 'offline' ? 'default' : 'orange'}
                    style={{ marginLeft: 6 }}
                  >
                    {currentVehicle.status === 'online' ? '在线' : currentVehicle.status === 'offline' ? '离线' : currentVehicle.status === 'maintenance' ? '维护中' : '停运'}
                  </Tag>
                )}
              </Tag>
            )}
          </Space>
        }
        extra={
          <Space wrap>
            <Select
              placeholder="选择车辆"
              style={{ minWidth: 200 }}
              value={vehicleId}
              onChange={setVehicleId}
              showSearch
              optionFilterProp="children"
              loading={vehicleList.length === 0}
            >
              {vehicleList.map((v: VehicleItem) => (
                <Option key={v.id} value={v.id}>
                  <Space>
                    <Tag color={v.status === 'online' ? 'green' : 'default'}>{v.plate_number}</Tag>
                    <Text type="secondary" style={{ fontSize: 12 }}>{v.vehicle_type}</Text>
                  </Space>
                </Option>
              ))}
            </Select>

            <RangePicker
              showTime
              value={dateRange as any}
              onChange={(v: any) => setDateRange(v)}
              style={{ minWidth: 340 }}
            />

            <Radio.Group value={granularity} onChange={e => setGranularity(e.target.value)}>
              <Radio.Button value="raw">原始</Radio.Button>
              <Radio.Button value="minute">分钟</Radio.Button>
              <Radio.Button value="hour">小时</Radio.Button>
              <Radio.Button value="day">天</Radio.Button>
            </Radio.Group>

            <Button icon={<ReloadOutlined />} onClick={fetchCharts} loading={loading}>刷新</Button>
            <Button type="primary" icon={<DownloadOutlined />} onClick={handleExport} loading={exporting}>
              导出CSV
            </Button>
          </Space>
        }
      >
        <Row gutter={[12, 12]}>
          <Col xs={24} md={6}>
            <Card size="small" style={{ borderRadius: 8, border: 'none', background: 'linear-gradient(135deg,#e6f4ff,#f0f5ff)' }}>
              <Statistic
                title={<Text type="secondary" style={{ fontSize: 12 }}>胎压告警(下限7/上限10bar)</Text>}
                value={pressureData?.stats?.warn_count ?? 0}
                valueStyle={{ fontSize: 20, color: '#0958d9' }}
                prefix={<DashboardOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card size="small" style={{ borderRadius: 8, border: 'none', background: 'linear-gradient(135deg,#fff7e6,#fffbe6)' }}>
              <Statistic
                title={<Text type="secondary" style={{ fontSize: 12 }}>胎温告警(≥75℃)</Text>}
                value={tempData?.stats?.warn_count ?? 0}
                valueStyle={{ fontSize: 20, color: '#d46b08' }}
                prefix={<WarningOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card size="small" style={{ borderRadius: 8, border: 'none', background: 'linear-gradient(135deg,#fff1f0,#fff2e8)' }}>
              <Statistic
                title={<Text type="secondary" style={{ fontSize: 12 }}>刹车片告警(≥250℃)</Text>}
                value={brakeData?.stats?.warn_count ?? 0}
                valueStyle={{ fontSize: 20, color: '#cf1322' }}
                prefix={<FireOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card size="small" style={{ borderRadius: 8, border: 'none', background: 'linear-gradient(135deg,#f6ffed,#e6fffb)' }}>
              <Statistic
                title={<Text type="secondary" style={{ fontSize: 12 }}>总告警次数</Text>}
                value={(pressureData?.stats?.warn_count ?? 0) + (tempData?.stats?.warn_count ?? 0) + (brakeData?.stats?.warn_count ?? 0)}
                valueStyle={{ fontSize: 20, color: '#389e0d' }}
                prefix={<CarOutlined />}
              />
            </Card>
          </Col>
        </Row>
      </Card>

      {renderChartCard('胎压历史曲线', <DashboardOutlined style={{ color: '#1677ff' }} />, pressureChart, pressureData, 'bar', (v) => v < 7 || v > 10)}
      {renderChartCard('胎温历史曲线', <WarningOutlined style={{ color: '#fa8c16' }} />, tempChart, tempData, '℃', (v) => v >= 75)}
      {renderChartCard('刹车片温度历史曲线', <FireOutlined style={{ color: '#ff4d4f' }} />, brakeChart, brakeData, '℃', (v) => v >= 250)}
    </div>
  )
}

export default TirePressureChart
