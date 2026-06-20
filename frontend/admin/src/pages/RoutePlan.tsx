import React, { useState } from 'react'
import {
  Row,
  Col,
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  Typography,
  Tag,
  Divider,
  Statistic,
  Progress,
  Tabs,
  List,
  Descriptions,
  Tooltip,
  message,
  Radio,
  Empty,
  Spin,
  Alert,
} from 'antd'
import {
  EnvironmentOutlined,
  CarOutlined,
  RouteOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
  BankOutlined,
  ExperimentOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  ZoomInOutlined,
} from '@ant-design/icons'
import AMap from '@/components/AMap'
import api from '@/services/api'
import { formatDistance, formatDuration } from '@/utils/auth'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs

interface RestrictedArea {
  id: number
  name: string
  area_type: string
  level: number
  center_latitude: number
  center_longitude: number
  radius: number
}

interface RouteResult {
  id?: number
  plan_no: string
  strategy: string
  origin: { latitude: number; longitude: number; address: string }
  destination: { latitude: number; longitude: number; address: string }
  route_path: Array<{ lat: number; lng: number }>
  total_distance: number
  estimated_duration: number
  expected_speed: number
  toll_fee: number
  fuel_cost: number
  avoid_tunnels: number
  avoid_bridges: number
  avoid_populated: number
  avoid_water_protection: number
  restricted_segments: Array<{
    area_id: number
    area_name: string
    area_type: string
    level: number
    distance: number
    reason: string
    suggestion: string
  }>
  safety_score: number
}

type StrategyType = 'shortest' | 'safest' | 'economic'

const RoutePlan: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [routes, setRoutes] = useState<Record<StrategyType, RouteResult | null>>({
    shortest: null,
    safest: null,
    economic: null,
  })
  const [activeStrategy, setActiveStrategy] = useState<StrategyType>('safest')
  const [restrictedAreas, setRestrictedAreas] = useState<RestrictedArea[]>([])
  const [originPicked, setOriginPicked] = useState<[number, number] | null>(null)
  const [destPicked, setDestPicked] = useState<[number, number] | null>(null)
  const [pickingMode, setPickingMode] = useState<'origin' | 'dest' | null>(null)

  const fetchRestricted = async () => {
    try {
      const res: any = await api.get('/routes/restricted-areas')
      setRestrictedAreas(res?.list || [])
    } catch (e) { }
  }

  React.useEffect(() => {
    fetchRestricted()
  }, [])

  const handleSubmit = async (values: any) => {
    setLoading(true)
    try {
      const payload = {
        origin: {
          address: values.origin_address,
          latitude: values.origin_lat,
          longitude: values.origin_lng,
        },
        destination: {
          address: values.dest_address,
          latitude: values.dest_lat,
          longitude: values.dest_lng,
        },
        vehicle_type: values.vehicle_type,
        vehicle_height: values.vehicle_height,
        vehicle_weight: values.vehicle_weight,
        hazard_class: values.hazard_class,
      }

      const strategies: StrategyType[] = ['shortest', 'safest', 'economic']
      const results = await Promise.all(
        strategies.map(s =>
          api.post<RouteResult>('/routes/plan', {
            ...payload,
            strategy: s,
          }).catch(() => null)
        )
      )

      setRoutes({
        shortest: results[0] as any,
        safest: results[1] as any,
        economic: results[2] as any,
      })
      message.success('路径规划完成')
    } catch (e: any) {
      message.error(e.message || '规划失败')
    } finally {
      setLoading(false)
    }
  }

  const handleMapClick = (lng: number, lat: number) => {
    if (pickingMode === 'origin') {
      form.setFieldsValue({
        origin_lat: Number(lat.toFixed(6)),
        origin_lng: Number(lng.toFixed(6)),
        origin_address: `坐标 (${lat.toFixed(4)}, ${lng.toFixed(4)})`,
      })
      setOriginPicked([lng, lat])
      setPickingMode(null)
    } else if (pickingMode === 'dest') {
      form.setFieldsValue({
        dest_lat: Number(lat.toFixed(6)),
        dest_lng: Number(lng.toFixed(6)),
        dest_address: `坐标 (${lat.toFixed(4)}, ${lng.toFixed(4)})`,
      })
      setDestPicked([lng, lat])
      setPickingMode(null)
    }
  }

  const currentRoute = routes[activeStrategy]

  const routeMarkers = [
    ...(originPicked ? [{
      position: originPicked as [number, number],
      title: '起点',
      color: '#52c41a',
    }] : []),
    ...(destPicked ? [{
      position: destPicked as [number, number],
      title: '终点',
      color: '#1677ff',
    }] : []),
    ...(currentRoute ? currentRoute.restricted_segments.map((seg, i) => {
      const area = restrictedAreas.find(a => a.id === seg.area_id)
      return {
        position: [
          area?.center_longitude || 116.4 + i * 0.05,
          area?.center_latitude || 39.9 + i * 0.03,
        ] as [number, number],
        title: seg.area_name,
        color: seg.level === 2 ? '#ff4d4f' : '#fa8c16',
        info: seg,
      }
    }) : []),
  ]

  const routePolylines = currentRoute?.route_path ? [{
    path: currentRoute.route_path.map(p => [p.lng, p.lat]) as [number, number][],
    color: activeStrategy === 'shortest' ? '#1677ff' :
      activeStrategy === 'safest' ? '#52c41a' : '#722ed1',
    weight: 7,
  }] : []

  const restrictedPolygons = restrictedAreas.slice(0, 50).map(area => {
    const r = (area.radius || 500) / 111000
    return {
      path: Array.from({ length: 24 }, (_, i) => {
        const a = (i / 24) * Math.PI * 2
        return [area.center_longitude + r * Math.cos(a), area.center_latitude + r * Math.sin(a)] as [number, number]
      }),
      fillColor: area.level === 2 ? '#ff4d4f' : '#fa8c16',
      strokeColor: area.level === 2 ? '#ff4d4f' : '#fa8c16',
      fillOpacity: 0.08,
      strokeWeight: 1,
    }
  })

  const strategyOptions: Array<{ key: StrategyType; label: string; color: string; desc: string; icon: any }> = [
    { key: 'shortest', label: '最短路径', color: '#1677ff', desc: '优先距离最短', icon: <ThunderboltOutlined /> },
    { key: 'safest', label: '最安全路径', color: '#52c41a', desc: '最大限度绕行，避开所有风险', icon: <SafetyCertificateOutlined /> },
    { key: 'economic', label: '经济路径', color: '#722ed1', desc: '高速优先，省时省油费', icon: <BankOutlined /> },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <RouteOutlined style={{ color: '#1677ff', fontSize: 18 }} />
            <Text strong style={{ fontSize: 16 }}>危险品运输路径规划</Text>
            <Tag color="blue">A* 算法 + 路权权重</Tag>
          </Space>
        }
      >
        <Alert
          type="info"
          showIcon
          message="系统自动避开：人口密集区（学校/医院/商圈）、隧道、桥梁、水源保护区、限高限重路段"
          style={{ marginBottom: 24, borderRadius: 8 }}
        />
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            vehicle_type: 'tanker',
            hazard_class: '3',
            origin_address: '北京市朝阳区化工路',
            origin_lat: 39.8726,
            origin_lng: 116.4888,
            dest_address: '天津市滨海新区化工园区',
            dest_lat: 39.0275,
            dest_lng: 117.6394,
          }}
        >
          <Row gutter={24}>
            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #e6f4ff' }} title={<Space><EnvironmentOutlined style={{ color: '#52c41a' }} /> 起点</Card>}>
                <Form.Item label="起点地址" name="origin_address" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input placeholder="输入起点地址或点击地图选点" />
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="纬度" name="origin_lat" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="经度" name="origin_lng" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                </Row>
                <Button
                  type={pickingMode === 'origin' ? 'primary' : 'dashed'}
                  size="small"
                  block
                  style={{ marginTop: 12 }}
                  icon={<ZoomInOutlined />}
                  onClick={() => setPickingMode(pickingMode === 'origin' ? null : 'origin')}
                >
                  {pickingMode === 'origin' ? '正在地图上选点...' : '在地图上选择起点'}
                </Button>
              </Card>
            </Col>

            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #ffccc7' }} title={<Space><EnvironmentOutlined style={{ color: '#ff4d4f' }} /> 终点</Card>}>
                <Form.Item label="终点地址" name="dest_address" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Input placeholder="输入终点地址或点击地图选点" />
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="纬度" name="dest_lat" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="经度" name="dest_lng" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                      <Input step={0.000001} />
                    </Form.Item>
                  </Col>
                </Row>
                <Button
                  type={pickingMode === 'dest' ? 'primary' : 'dashed'}
                  size="small"
                  block
                  style={{ marginTop: 12 }}
                  icon={<ZoomInOutlined />}
                  onClick={() => setPickingMode(pickingMode === 'dest' ? null : 'dest')}
                >
                  {pickingMode === 'dest' ? '正在地图上选点...' : '在地图上选择终点'}
                </Button>
              </Card>
            </Col>

            <Col xs={24} md={8}>
              <Card size="small" style={{ borderRadius: 8, border: '1px solid #ffe7ba' }} title={<Space><CarOutlined style={{ color: '#fa8c16' }} /> 车辆 & 危险品</Card>}>
                <Form.Item label="车辆类型" name="vehicle_type" rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                  <Select>
                    <Option value="tanker">🏺 罐车</Option>
                    <Option value="van">📦 厢式货车</Option>
                    <Option value="flatbed">🪵 平板车</Option>
                    <Option value="other">🚛 其他</Option>
                  </Select>
                </Form.Item>
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item label="车高(m)" name="vehicle_height" initialValue={3.9} style={{ marginBottom: 0 }}>
                      <Input type="number" step={0.1} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="总重(吨)" name="vehicle_weight" initialValue={32.5} style={{ marginBottom: 0 }}>
                      <Input type="number" step={0.5} />
                    </Form.Item>
                  </Col>
                </Row>
                <Form.Item label="危险品类别" name="hazard_class" rules={[{ required: true }]} style={{ marginTop: 8, marginBottom: 0 }}>
                  <Select>
                    <Option value="1">1类 - 爆炸品</Option>
                    <Option value="2">2类 - 压缩气体</Option>
                    <Option value="3">3类 - 易燃液体 ⭐</Option>
                    <Option value="4">4类 - 易燃固体</Option>
                    <Option value="5">5类 - 氧化剂</Option>
                    <Option value="6">6类 - 毒害品</Option>
                    <Option value="8">8类 - 腐蚀品</Option>
                    <Option value="9">9类 - 杂项</Option>
                  </Select>
                </Form.Item>
              </Card>
            </Col>
          </Row>

          <Divider style={{ margin: '20px 0 16px' }} />

          <Row gutter={16} style={{ marginBottom: 16 }}>
            {strategyOptions.map(s => (
              <Col xs={24} md={8} key={s.key}>
                <div
                  onClick={() => setActiveStrategy(s.key)}
                  style={{
                    padding: 16,
                    borderRadius: 10,
                    border: `2px solid ${activeStrategy === s.key ? s.color : '#f0f0f0'}`,
                    background: activeStrategy === s.key ? `${s.color}08` : '#fafafa',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                  }}
                >
                  <Space style={{ fontSize: 18, color: s.color }}>
                    {s.icon}
                    <Text strong style={{ fontSize: 16, color: '#1f1f1f' }}>{s.label}</Text>
                  </Space>
                  <div style={{ marginTop: 4, fontSize: 12, color: '#8c8c8c' }}>{s.desc}</div>
                </div>
              </Col>
            ))}
          </Row>

          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading} size="large" icon={<RouteOutlined />}>
                {loading ? '规划中...' : '开始三策略规划'}
              </Button>
              <Button
                size="large"
                icon={<ReloadOutlined />}
                onClick={() => {
                  form.resetFields()
                  setRoutes({ shortest: null, safest: null, economic: null })
                  setOriginPicked(null)
                  setDestPicked(null)
                }}
              >
                重置
              </Button>
              {pickingMode && (
                <Tag color="red">
                  📍 点击地图选择{pickingMode === 'origin' ? '起点' : '终点'}位置
                </Tag>
              )}
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            bodyStyle={{ padding: 0 }}
            title={
              <Space>
                <ExperimentOutlined style={{ color: '#1677ff' }} />
                <Text strong style={{ fontSize: 15 }}>路径规划地图</Text>
                <Tag color="geekblue">高德高精地图 · 货车模式</Tag>
              </Space>
            }
          >
            <AMap
              style={{ height: 560, borderRadius: '0 0 12px 12px' }}
              markers={routeMarkers}
              polylines={routePolylines}
              polygons={restrictedPolygons}
              center={[116.8, 39.5]}
              zoom={9}
              onMapClick={handleMapClick}
              showTraffic
            />
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            bordered={false}
            style={{ borderRadius: 12, height: 560, display: 'flex', flexDirection: 'column' }}
            bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}
          >
            <Tabs
              activeKey={activeStrategy}
              onChange={k => setActiveStrategy(k as StrategyType)}
              size="large"
              style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
              items={strategyOptions.map(s => ({
                key: s.key,
                label: (
                  <span style={{ color: activeStrategy === s.key ? s.color : undefined }}>
                    {s.icon} {s.label}
                  </span>
                ),
                children: (() => {
                  const r = routes[s.key]
                  if (!r) {
                    return (
                      <div style={{ padding: 40, textAlign: 'center' }}>
                        <Empty description="请先点击「开始三策略规划」" />
                      </div>
                    )
                  }
                  return (
                    <div style={{ padding: '0 20px 20px', overflow: 'auto', height: '100%' }}>
                      <Row gutter={12} style={{ marginBottom: 16 }}>
                        <Col span={12}>
                          <Card size="small" style={{ borderRadius: 8 }}>
                            <Statistic
                              title="总里程"
                              value={formatDistance(r.total_distance)}
                              valueStyle={{ fontSize: 18, color: s.color }}
                            />
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card size="small" style={{ borderRadius: 8 }}>
                            <Statistic
                              title="预计时长"
                              value={formatDuration(r.estimated_duration)}
                              valueStyle={{ fontSize: 18, color: s.color }}
                            />
                          </Card>
                        </Col>
                      </Row>

                      <Row gutter={12} style={{ marginBottom: 16 }}>
                        <Col span={12}>
                          <Card size="small" style={{ borderRadius: 8 }}>
                            <Statistic
                              title="过路费"
                              value={r.toll_fee?.toFixed?.(2) || 0}
                              prefix="¥"
                              valueStyle={{ fontSize: 16 }}
                            />
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card size="small" style={{ borderRadius: 8 }}>
                            <Statistic
                              title="油费估算"
                              value={r.fuel_cost?.toFixed?.(2) || 0}
                              prefix="¥"
                              valueStyle={{ fontSize: 16 }}
                            />
                          </Card>
                        </Col>
                      </Row>

                      <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} title="安全评分">
                        <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
                          <Progress
                            type="circle"
                            percent={r.safety_score || 0}
                            size={80}
                            strokeColor={
                              (r.safety_score || 0) >= 90 ? '#52c41a' :
                                (r.safety_score || 0) >= 70 ? '#faad14' : '#ff4d4f'
                            }
                          />
                          <div>
                            <Space direction="vertical" size={4}>
                              <Space>
                                <Tag color="green">🚇 避开隧道 ×{r.avoid_tunnels || 0}</Tag>
                              </Space>
                              <Space>
                                <Tag color="cyan">🌉 避开桥梁 ×{r.avoid_bridges || 0}</Tag>
                              </Space>
                              <Space>
                                <Tag color="purple">🏘️ 避开密集区 ×{r.avoid_populated || 0}</Tag>
                              </Space>
                              <Space>
                                <Tag color="blue">💧 避开水源地 ×{r.avoid_water_protection || 0}</Tag>
                              </Space>
                            </Space>
                          </div>
                        </div>
                      </Card>

                      <Divider style={{ margin: '12px 0' }} />

                      <Title level={5} style={{ marginTop: 0 }}>
                        <WarningOutlined style={{ color: '#fa8c16' }} /> 禁行/限行路段
                      </Title>
                      {r.restricted_segments?.length ? (
                        <List
                          size="small"
                          dataSource={r.restricted_segments}
                          renderItem={(seg) => (
                            <List.Item style={{ padding: '10px 0', borderBottom: '1px dashed #f0f0f0' }}>
                              <div style={{ width: '100%' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                                  <Tag color={seg.level === 2 ? 'red' : 'orange'} style={{ margin: 0 }}>
                                    {seg.level === 2 ? '必须绕行' : '建议绕行'}
                                  </Tag>
                                  <Text strong style={{ fontSize: 13 }}>{seg.area_name}</Text>
                                  <Tag color="blue">{seg.area_type}</Tag>
                                </div>
                                <Paragraph type="secondary" style={{ margin: '4px 0', fontSize: 12 }}>
                                  <ExclamationCircleOutlined /> {seg.reason}
                                </Paragraph>
                                <Paragraph style={{ margin: '4px 0', fontSize: 12, color: '#52c41a' }}>
                                  <CheckCircleOutlined /> {seg.suggestion}
                                </Paragraph>
                              </div>
                            </List.Item>
                          )}
                        />
                      ) : (
                        <Empty description="无禁行路段，路径非常安全！" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                      )}
                    </div>
                  )
                })(),
              }))}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default RoutePlan
