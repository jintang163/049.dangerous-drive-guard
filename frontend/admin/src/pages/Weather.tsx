import React, { useEffect, useState, useRef } from 'react'
import {
  Row,
  Col,
  Card,
  Table,
  Tag,
  Button,
  Space,
  Typography,
  Modal,
  Statistic,
  List,
  Divider,
  Descriptions,
  Empty,
  Tooltip,
  Badge,
  message,
  Select,
  Input,
} from 'antd'
import {
  CloudOutlined,
  CloudTwoTone,
  ThunderboltOutlined,
  ThunderboltTwoTone,
  CloudServerOutlined,
  CloudServerTwoTone,
  SnowflakeOutlined,
  SnowflakeTwoTone,
  EnvironmentOutlined,
  EyeOutlined,
  CarOutlined,
  RouteOutlined,
  ReloadOutlined,
  FilterOutlined,
  ExportOutlined,
  InfoCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import AMapLoader from '@amap/amap-jsapi-loader'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select

interface WeatherWarning {
  id: string
  warning_no: string
  warning_type: 'rainstorm' | 'fog' | 'wind' | 'ice'
  warning_level: 1 | 2 | 3 | 4
  region_name: string
  region_code: string
  polygon?: number[][]
  publish_time: string
  influence_vehicle_count: number
  status: 'active' | 'expired' | 'cancelled'
  suggestion: string
  influence_scope: string
  publisher: string
}

const warningTypeMap: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  rainstorm: { label: '暴雨预警', color: '#1677ff', icon: <CloudOutlined /> },
  fog: { label: '大雾预警', color: '#8c8c8c', icon: <CloudServerOutlined /> },
  wind: { label: '大风预警', color: '#13c2c2', icon: <ThunderboltOutlined /> },
  ice: { label: '冰雪预警', color: '#69c0ff', icon: <SnowflakeOutlined /> },
}

const warningLevelMap: Record<number, { label: string; color: string; bgColor: string }> = {
  1: { label: '一般(蓝色)', color: 'blue', bgColor: 'rgba(22,119,255,0.3)' },
  2: { label: '较重(黄色)', color: 'gold', bgColor: 'rgba(250,173,20,0.3)' },
  3: { label: '严重(橙色)', color: 'orange', bgColor: 'rgba(250,140,22,0.3)' },
  4: { label: '特别严重(红色)', color: 'red', bgColor: 'rgba(255,77,79,0.3)' },
}

const statusMap = {
  active: { color: 'red', label: '生效中' },
  expired: { color: 'default', label: '已过期' },
  cancelled: { color: 'blue', label: '已解除' },
}

const mockWarnings: WeatherWarning[] = [
  {
    id: '1',
    warning_no: 'WX20260620001',
    warning_type: 'rainstorm',
    warning_level: 4,
    region_name: '广东省深圳市南山区',
    region_code: '440305',
    polygon: [
      [113.93, 22.53], [114.05, 22.53], [114.05, 22.62], [113.93, 22.62], [113.93, 22.53]
    ],
    publish_time: '2026-06-20 08:30:00',
    influence_vehicle_count: 1256,
    status: 'active',
    suggestion: '建议暂停危化品运输，已在途车辆就近服务区停靠避险，驾驶员注意路面积水，保持安全车距',
    influence_scope: '南山区全境及宝安区南部，预计持续12小时',
    publisher: '深圳市气象局',
  },
  {
    id: '2',
    warning_no: 'WX20260620002',
    warning_type: 'fog',
    warning_level: 3,
    region_name: '江苏省南京市江宁区',
    region_code: '320115',
    polygon: [
      [118.76, 31.86], [119.03, 31.86], [119.03, 32.08], [118.76, 32.08], [118.76, 31.86]
    ],
    publish_time: '2026-06-20 05:00:00',
    influence_vehicle_count: 892,
    status: 'active',
    suggestion: '能见度不足200米，建议开启雾灯和危险报警闪光灯，限速60km/h，与前车保持100米以上距离',
    influence_scope: '江宁区及周边高速公路网，预计持续至上午10点',
    publisher: '江苏省气象局',
  },
  {
    id: '3',
    warning_no: 'WX20260620003',
    warning_type: 'wind',
    warning_level: 3,
    region_name: '山东省青岛市黄岛区',
    region_code: '370211',
    polygon: [
      [120.05, 35.88], [120.38, 35.88], [120.38, 36.22], [120.05, 36.22], [120.05, 35.88]
    ],
    publish_time: '2026-06-20 06:15:00',
    influence_vehicle_count: 634,
    status: 'active',
    suggestion: '阵风8-10级，空车及罐装车注意横风影响，建议避开跨海大桥和沿海公路',
    influence_scope: '黄岛区全境及胶州湾大桥沿线',
    publisher: '青岛市气象局',
  },
  {
    id: '4',
    warning_no: 'WX20260620004',
    warning_type: 'ice',
    warning_level: 2,
    region_name: '黑龙江省哈尔滨市道里区',
    region_code: '230102',
    polygon: [
      [126.55, 45.70], [126.70, 45.70], [126.70, 45.83], [126.55, 45.83], [126.55, 45.70]
    ],
    publish_time: '2026-06-19 20:00:00',
    influence_vehicle_count: 421,
    status: 'expired',
    suggestion: '道路结冰预警，建议安装防滑链，减速慢行，避免紧急制动和急打方向',
    influence_scope: '道里区及京哈高速哈尔滨段',
    publisher: '哈尔滨市气象局',
  },
  {
    id: '5',
    warning_no: 'WX20260620005',
    warning_type: 'rainstorm',
    warning_level: 2,
    region_name: '浙江省杭州市西湖区',
    region_code: '330106',
    publish_time: '2026-06-20 09:45:00',
    influence_vehicle_count: 567,
    status: 'active',
    suggestion: '短时强降雨，注意低洼路段积水，立交桥下谨慎通行',
    influence_scope: '西湖区及绕城高速西段',
    publisher: '杭州市气象局',
  },
  {
    id: '6',
    warning_no: 'WX20260619008',
    warning_type: 'fog',
    warning_level: 1,
    region_name: '安徽省合肥市蜀山区',
    region_code: '340104',
    publish_time: '2026-06-19 22:30:00',
    influence_vehicle_count: 289,
    status: 'expired',
    suggestion: '能见度500米左右，注意开启雾灯，谨慎驾驶',
    influence_scope: '蜀山区及合肥绕城高速',
    publisher: '合肥市气象局',
  },
  {
    id: '7',
    warning_no: 'WX20260619006',
    warning_type: 'wind',
    warning_level: 4,
    region_name: '福建省厦门市湖里区',
    region_code: '350206',
    publish_time: '2026-06-19 14:20:00',
    influence_vehicle_count: 712,
    status: 'cancelled',
    suggestion: '台风外围影响，阵风11级，建议所有危化品车辆立即停运，寻找安全停车场',
    influence_scope: '厦门全岛及翔安隧道',
    publisher: '厦门市气象局',
  },
  {
    id: '8',
    warning_no: 'WX20260618012',
    warning_type: 'ice',
    warning_level: 1,
    region_name: '吉林省长春市朝阳区',
    region_code: '220104',
    publish_time: '2026-06-18 19:00:00',
    influence_vehicle_count: 198,
    status: 'expired',
    suggestion: '局部道路结冰，减速慢行',
    influence_scope: '朝阳区及京哈高速长春段',
    publisher: '长春市气象局',
  },
]

const Weather: React.FC = () => {
  const mapRef = useRef<HTMLDivElement>(null)
  const mapInstance = useRef<any>(null)
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<WeatherWarning[]>(mockWarnings)
  const [detailModal, setDetailModal] = useState<WeatherWarning | null>(null)
  const [levelFilter, setLevelFilter] = useState<number>()
  const [typeFilter, setTypeFilter] = useState<string>()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const stats = {
    rainstorm: data.filter(d => d.warning_type === 'rainstorm' && d.status === 'active').length,
    fog: data.filter(d => d.warning_type === 'fog' && d.status === 'active').length,
    wind: data.filter(d => d.warning_type === 'wind' && d.status === 'active').length,
    ice: data.filter(d => d.warning_type === 'ice' && d.status === 'active').length,
  }

  const activeWarnings = data.filter(d => d.status === 'active').sort((a, b) => b.warning_level - a.warning_level)

  const initMap = async () => {
    if (!mapRef.current || mapInstance.current) return
    try {
      const AMap = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: [],
      })
      mapInstance.current = new AMap.Map(mapRef.current, {
        zoom: 4,
        center: [108, 34],
        viewMode: '2D',
      })
      activeWarnings.forEach(warning => {
        if (warning.polygon) {
          const level = warningLevelMap[warning.warning_level]
          new AMap.Polygon({
            path: warning.polygon,
            strokeColor: level.color === 'red' ? '#ff4d4f' : level.color === 'orange' ? '#fa8c16' : level.color === 'gold' ? '#faad14' : '#1677ff',
            strokeWeight: 2,
            strokeOpacity: 0.8,
            fillColor: level.bgColor,
            fillOpacity: 0.5,
            map: mapInstance.current,
          })
        }
      })
    } catch (e) {
      console.log('AMap init failed, using fallback')
    }
  }

  useEffect(() => {
    initMap()
  }, [])

  const columns = [
    {
      title: '预警编号',
      dataIndex: 'warning_no',
      width: 160,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '类型',
      dataIndex: 'warning_type',
      width: 120,
      render: (v: string) => {
        const t = warningTypeMap[v] || { label: v, color: 'default', icon: null }
        return (
          <Tag color={t.color} icon={t.icon} style={{ fontSize: 12 }}>
            {t.label}
          </Tag>
        )
      },
    },
    {
      title: '等级',
      dataIndex: 'warning_level',
      width: 120,
      render: (v: number) => {
        const l = warningLevelMap[v] || warningLevelMap[1]
        return <Tag color={l.color}>{l.label}</Tag>
      },
    },
    {
      title: '区域',
      dataIndex: 'region_name',
      ellipsis: true,
      render: (v: string) => (
        <Tooltip title={v}>
          <span><EnvironmentOutlined style={{ color: '#1677ff' }} /> {v}</span>
        </Tooltip>
      ),
    },
    {
      title: '发布时间',
      dataIndex: 'publish_time',
      width: 170,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '影响车辆数',
      dataIndex: 'influence_vehicle_count',
      width: 120,
      render: (v: number) => (
        <Space>
          <CarOutlined />
          <Text strong>{v.toLocaleString()}</Text>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (v: string) => {
        const s = statusMap[v as keyof typeof statusMap] || statusMap.active
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    {
      title: '操作',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: WeatherWarning) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailModal(record)}>
          详情
        </Button>
      ),
    },
  ]

  const handleRoutePlan = (warning: WeatherWarning) => {
    message.success(`已基于${warning.region_name}天气预警生成路径重规划方案，建议相关运营调度员查看`)
  }

  const trendChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['暴雨', '大雾', '大风', '冰雪'], bottom: 0 },
    grid: { left: 40, right: 20, top: 20, bottom: 40 },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 7 }, (_, i) => dayjs().subtract(6 - i, 'day').format('MM-DD')),
    },
    yAxis: { type: 'value' },
    series: [
      { name: '暴雨', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 6)), itemStyle: { color: '#1677ff' } },
      { name: '大雾', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 4)), itemStyle: { color: '#8c8c8c' } },
      { name: '大风', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 5)), itemStyle: { color: '#13c2c2' } },
      { name: '冰雪', type: 'line', smooth: true, data: Array.from({ length: 7 }, () => Math.floor(Math.random() * 3)), itemStyle: { color: '#69c0ff' } },
    ],
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="暴雨预警"
              value={stats.rainstorm}
              valueStyle={{ color: '#1677ff' }}
              prefix={<CloudTwoTone />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="大雾预警"
              value={stats.fog}
              valueStyle={{ color: '#8c8c8c' }}
              prefix={<CloudServerTwoTone />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="大风预警"
              value={stats.wind}
              valueStyle={{ color: '#13c2c2' }}
              prefix={<ThunderboltTwoTone />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="冰雪预警"
              value={stats.ice}
              valueStyle={{ color: '#69c0ff' }}
              prefix={<SnowflakeTwoTone />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={14}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={<Space><EnvironmentOutlined style={{ color: '#1677ff' }} /> 预警区域分布</Space>}
            extra={
              <Space size={8}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span style={{ width: 12, height: 12, background: 'rgba(255,77,79,0.5)', borderRadius: 2 }}></span>
                  <Text type="secondary" style={{ fontSize: 12 }}>严重</Text>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span style={{ width: 12, height: 12, background: 'rgba(250,140,22,0.5)', borderRadius: 2 }}></span>
                  <Text type="secondary" style={{ fontSize: 12 }}>较重</Text>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span style={{ width: 12, height: 12, background: 'rgba(250,173,20,0.5)', borderRadius: 2 }}></span>
                  <Text type="secondary" style={{ fontSize: 12 }}>一般</Text>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span style={{ width: 12, height: 12, background: 'rgba(22,119,255,0.5)', borderRadius: 2 }}></span>
                  <Text type="secondary" style={{ fontSize: 12 }}>轻微</Text>
                </div>
              </Space>
            }
          >
            <div ref={mapRef} style={{ height: 360, borderRadius: 8, background: 'linear-gradient(180deg, #e6f4ff 0%, #f0f5ff 100%)', position: 'relative', overflow: 'hidden' }}>
              <svg viewBox="100 18 40 25" style={{ width: '100%', height: '100%' }}>
                <path d="M115,35 L125,32 L135,35 L140,30 L150,32 L155,28 L165,30 L170,35 L165,40 L155,38 L150,42 L140,40 L130,42 L120,40 L115,35 Z" fill="#d9d9d9" stroke="#bfbfbf" strokeWidth="0.3" />
                {activeWarnings.map((w, i) => {
                  if (!w.polygon) return null
                  const level = warningLevelMap[w.warning_level]
                  const strokeColor = level.color === 'red' ? '#ff4d4f' : level.color === 'orange' ? '#fa8c16' : level.color === 'gold' ? '#faad14' : '#1677ff'
                  const centerLng = w.polygon.reduce((a, b) => a + b[0], 0) / w.polygon.length
                  const centerLat = w.polygon.reduce((a, b) => a + b[1], 0) / w.polygon.length
                  return (
                    <g key={w.id}>
                      <polygon
                        points={w.polygon.map(p => `${p[0]},${p[1]}`).join(' ')}
                        fill={level.bgColor}
                        stroke={strokeColor}
                        strokeWidth="0.4"
                        style={{ cursor: 'pointer' }}
                        onClick={() => setDetailModal(w)}
                      >
                        <title>{`${w.region_name} - ${warningLevelMap[w.warning_level].label}`}</title>
                      </polygon>
                      <circle cx={centerLng} cy={centerLat} r="0.6" fill={strokeColor} />
                    </g>
                  )
                })}
              </svg>
              <div style={{ position: 'absolute', bottom: 8, right: 8, background: 'rgba(255,255,255,0.85)', padding: '4px 10px', borderRadius: 4, fontSize: 11, color: '#8c8c8c' }}>
                共 {activeWarnings.length} 个生效预警区域
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card
            bordered={false}
            style={{ borderRadius: 12, height: '100%' }}
            title={
              <Space>
                <WarningOutlined style={{ color: '#ff4d4f' }} />
                <Text strong style={{ fontSize: 15 }}>实时预警消息</Text>
                <Badge count={activeWarnings.length} showZero style={{ backgroundColor: '#ff4d4f' }} />
              </Space>
            }
            bodyStyle={{ padding: 0, maxHeight: 360, overflow: 'auto' }}
          >
            {activeWarnings.length === 0 ? (
              <Empty description="暂无生效预警" style={{ padding: '40px 0' }} />
            ) : (
              <List
                dataSource={activeWarnings}
                renderItem={(item) => {
                  const t = warningTypeMap[item.warning_type]
                  const l = warningLevelMap[item.warning_level]
                  return (
                    <List.Item
                      key={item.id}
                      style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0', cursor: 'pointer' }}
                      onClick={() => setDetailModal(item)}
                    >
                      <List.Item.Meta
                        avatar={
                          <div style={{
                            width: 40, height: 40, borderRadius: 8,
                            background: l.color === 'red' ? '#fff1f0' : l.color === 'orange' ? '#fff7e6' : l.color === 'gold' ? '#fffbe6' : '#e6f4ff',
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontSize: 20, color: l.color === 'red' ? '#ff4d4f' : l.color === 'orange' ? '#fa8c16' : l.color === 'gold' ? '#faad14' : '#1677ff',
                          }}>
                            {t.icon}
                          </div>
                        }
                        title={
                          <Space size={4} wrap>
                            <Tag color={l.color}>{l.label}</Tag>
                            <Tag color={t.color}>{t.label}</Tag>
                            <Text strong style={{ fontSize: 13 }}>{item.region_name}</Text>
                          </Space>
                        }
                        description={
                          <div style={{ fontSize: 12 }}>
                            <div style={{ marginBottom: 4 }}>
                              <Text type="secondary"><InfoCircleOutlined /> {item.influence_scope}</Text>
                            </div>
                            <Space size={16}>
                              <Text type="secondary">{item.publish_time}</Text>
                              <Text type="secondary"><CarOutlined /> 影响{item.influence_vehicle_count}车</Text>
                            </Space>
                          </div>
                        }
                      />
                    </List.Item>
                  )
                }}
              />
            )}
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <CloudOutlined style={{ color: '#1677ff' }} />
            <Text strong style={{ fontSize: 15 }}>历史预警记录</Text>
            <Tag color="blue">{data.length}</Tag>
          </Space>
        }
        extra={
          <Space wrap>
            <Select allowClear placeholder="类型筛选" style={{ width: 140 }} value={typeFilter} onChange={setTypeFilter}>
              {Object.entries(warningTypeMap).map(([k, v]) => (
                <Option key={k} value={k}>{v.label}</Option>
              ))}
            </Select>
            <Select allowClear placeholder="等级筛选" style={{ width: 140 }} value={levelFilter} onChange={setLevelFilter}>
              {Object.entries(warningLevelMap).map(([k, v]) => (
                <Option key={k} value={Number(k)}>{v.label}</Option>
              ))}
            </Select>
            <Select allowClear placeholder="状态筛选" style={{ width: 140 }} value={statusFilter} onChange={setStatusFilter}>
              {Object.entries(statusMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => { setTypeFilter(undefined); setLevelFilter(undefined); setStatusFilter(undefined) }}>重置</Button>
            <Button icon={<ReloadOutlined />}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          columns={columns as any}
          dataSource={data.filter(d =>
            (!typeFilter || d.warning_type === typeFilter) &&
            (!levelFilter || d.warning_level === levelFilter) &&
            (!statusFilter || d.status === statusFilter)
          )}
          pagination={{
            current: page,
            pageSize,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
          scroll={{ x: 1100 }}
          rowClassName={(r) => r.warning_level === 4 ? '!bg-red-50' : r.warning_level === 3 ? '!bg-orange-50' : ''}
        />
      </Card>

      <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><CloudOutlined /> 近7日预警趋势</Space>}>
        <ReactECharts option={trendChart} style={{ height: 220 }} notMerge />
      </Card>

      <Modal
        title={
          <Space>
            <WarningOutlined style={{ color: '#ff4d4f' }} />
            <Text strong>预警详情</Text>
            {detailModal && (
              <>
                <Tag color={warningLevelMap[detailModal.warning_level]?.color}>
                  {warningLevelMap[detailModal.warning_level]?.label}
                </Tag>
                <Tag color={warningTypeMap[detailModal.warning_type]?.color}>
                  {warningTypeMap[detailModal.warning_type]?.label}
                </Tag>
              </>
            )}
          </Space>
        }
        open={!!detailModal}
        onCancel={() => setDetailModal(null)}
        width={680}
        footer={
          detailModal && detailModal.status === 'active' ? (
            <Space>
              <Button onClick={() => setDetailModal(null)}>关闭</Button>
              <Button type="primary" icon={<RouteOutlined />} onClick={() => handleRoutePlan(detailModal)}>
                路径重规划建议
              </Button>
            </Space>
          ) : null
        }
      >
        {detailModal && (
          <div>
            <Alert
              type={detailModal.warning_level >= 3 ? 'error' : 'warning'}
              showIcon
              icon={<WarningOutlined />}
              message={<Space><Text strong>{detailModal.region_name}</Text><Text type="secondary">{detailModal.publisher}</Text></Space>}
              description={detailModal.suggestion}
              style={{ borderRadius: 8, marginBottom: 16 }}
            />

            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="预警编号" span={1}>
                <Text copyable>{detailModal.warning_no}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="区域编码" span={1}>{detailModal.region_code}</Descriptions.Item>
              <Descriptions.Item label="发布时间" span={1}>{detailModal.publish_time}</Descriptions.Item>
              <Descriptions.Item label="当前状态" span={1}>
                <Tag color={statusMap[detailModal.status as keyof typeof statusMap]?.color}>
                  {statusMap[detailModal.status as keyof typeof statusMap]?.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="影响车辆数" span={1}>
                <Space><CarOutlined /> <Text strong>{detailModal.influence_vehicle_count.toLocaleString()} 辆</Text></Space>
              </Descriptions.Item>
              <Descriptions.Item label="发布机构" span={1}>{detailModal.publisher}</Descriptions.Item>
              <Descriptions.Item label="影响范围" span={2}>{detailModal.influence_scope}</Descriptions.Item>
            </Descriptions>

            <Divider />

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><InfoCircleOutlined /> 处置建议</Space>}>
              <Paragraph style={{ margin: 0, lineHeight: 1.8 }}>
                {detailModal.suggestion}
              </Paragraph>
            </Card>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Weather
