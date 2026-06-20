import React, { useEffect, useState, useRef } from 'react'
import {
  Row,
  Col,
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Form,
  Select,
  Input,
  Statistic,
  Drawer,
  Descriptions,
  Image,
  Progress,
  message,
  DatePicker,
  Timeline,
  Steps,
  Divider,
  Badge,
  Tooltip,
  Empty,
  Tabs,
} from 'antd'
import {
  InboxOutlined,
  CheckCircleOutlined,
  EnvironmentOutlined,
  CarOutlined,
  UserOutlined,
  CalendarOutlined,
  SearchOutlined,
  EyeOutlined,
  FilterOutlined,
  ReloadOutlined,
  ExportOutlined,
  TruckOutlined,
  WarningOutlined,
  SafetyCertificateOutlined,
  MapOutlined,
  FileTextOutlined,
  SignatureOutlined,
  LinkOutlined,
  EditOutlined,
  CloseOutlined,
  PlayCircleOutlined,
  SendOutlined,
} from '@ant-design/icons'
import {
  ProTable,
  ProFormText,
  ProFormSelect,
  ProFormDateRangePicker,
  ProCard,
  ActionType,
  RequestData,
} from '@ant-design/pro-components'
import type { ProColumns, ActionType as ProActionType } from '@ant-design/pro-components'
import api from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import AMapLoader from '@amap/amap-jsapi-loader'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { RangePicker } = DatePicker
const { TabPane } = Tabs

type WaybillStatus = 'pending' | 'transit' | 'signed' | 'abnormal' | 'cancelled'
type DangerLevel = '1.1' | '1.2' | '1.3' | '1.4' | '2.1' | '2.2' | '2.3' | '3' | '4.1' | '4.2' | '4.3' | '5.1' | '5.2' | '6.1' | '6.2' | '7' | '8' | '9'

interface WaybillItem {
  id: string
  waybill_no: string
  status: WaybillStatus
  danger_goods_name: string
  un_number: string
  danger_level: DangerLevel
  origin: string
  destination: string
  origin_lng: number
  origin_lat: number
  dest_lng: number
  dest_lat: number
  vehicle_plate: string
  driver_name: string
  driver_phone: string
  escort_name: string
  planned_departure: string
  expected_arrival: string
  actual_departure?: string
  actual_arrival?: string
  progress: number
  current_location?: string
  current_lng?: number
  current_lat?: number
  cargo_weight: string
  cargo_quantity: string
  packaging_type: string
  emergency_contact: string
  emergency_phone: string
  created_at: string
  block_hash?: string
  block_height?: number
  status_timeline: { status: string; time: string; operator: string; note?: string }[]
  route_points?: { lng: number; lat: number; time: string }[]
  receiver_signature?: string
  receiver_name?: string
  sign_time?: string
  transport_document_url?: string
  emergency_document_url?: string
}

const statusMap: Record<WaybillStatus, { label: string; color: string; stepStatus: 'wait' | 'process' | 'finish' | 'error' }> = {
  pending: { label: '待调度', color: 'default', stepStatus: 'wait' },
  transit: { label: '运输中', color: 'processing', stepStatus: 'process' },
  signed: { label: '已签收', color: 'success', stepStatus: 'finish' },
  abnormal: { label: '异常', color: 'error', stepStatus: 'error' },
  cancelled: { label: '取消', color: 'warning', stepStatus: 'error' },
}

const dangerLevelColorMap: Record<string, string> = {
  '1.1': 'red', '1.2': 'red', '1.3': 'red', '1.4': 'red',
  '2.1': 'volcano', '2.2': 'orange', '2.3': 'red',
  '3': 'red',
  '4.1': 'orange', '4.2': 'gold', '4.3': 'volcano',
  '5.1': 'purple', '5.2': 'magenta',
  '6.1': 'red', '6.2': 'magenta',
  '7': 'geekblue',
  '8': 'purple',
  '9': 'default',
}

const mockWaybills: WaybillItem[] = Array.from({ length: 38 }, (_, i) => {
  const statuses: WaybillStatus[] = ['pending', 'transit', 'signed', 'abnormal', 'cancelled']
  const status = statuses[i % 5]
  const goodsList = [
    { name: '汽油', un: 'UN1203', level: '3' },
    { name: '液化石油气', un: 'UN1075', level: '2.1' },
    { name: '硫酸', un: 'UN2790', level: '8' },
    { name: '液氯', un: 'UN1017', level: '2.3' },
    { name: '烟花爆竹', un: 'UN0336', level: '1.4' },
    { name: '氰化钠', un: 'UN1689', level: '6.1' },
    { name: '黄磷', un: 'UN1381', level: '4.2' },
    { name: '硝酸', un: 'UN2031', level: '8' },
  ]
  const goods = goodsList[i % goodsList.length]
  const cities = [
    { name: '上海市浦东新区', lng: 121.4737, lat: 31.2304 },
    { name: '北京市朝阳区', lng: 116.4074, lat: 39.9042 },
    { name: '广州市天河区', lng: 113.3245, lat: 23.1291 },
    { name: '深圳市南山区', lng: 113.9304, lat: 22.5333 },
    { name: '杭州市西湖区', lng: 120.1551, lat: 30.2741 },
    { name: '成都市武侯区', lng: 104.0668, lat: 30.5728 },
    { name: '武汉市江汉区', lng: 114.3055, lat: 30.5931 },
    { name: '南京市鼓楼区', lng: 118.778, lat: 32.0617 },
  ]
  const origin = cities[i % cities.length]
  const dest = cities[(i + 3) % cities.length]
  const plates = ['沪A12345', '京B67890', '粤C11111', '粤D22222', '浙E33333', '川F44444', '鄂G55555', '苏H66666']
  const drivers = ['张建国', '李明辉', '王志强', '刘文华', '陈晓峰', '赵大海', '孙德龙', '周永刚']

  const planned = dayjs().add(i * 2 - 30, 'hour')
  const progress = status === 'signed' ? 100 : status === 'transit' ? Math.floor(Math.random() * 70) + 15 : status === 'pending' ? 0 : Math.floor(Math.random() * 50)

  return {
    id: `wb_${10000 + i}`,
    waybill_no: `DDG${dayjs().format('YYYYMM')}${String(1000 + i).padStart(4, '0')}`,
    status,
    danger_goods_name: goods.name,
    un_number: goods.un,
    danger_level: goods.level as DangerLevel,
    origin: origin.name,
    destination: dest.name,
    origin_lng: origin.lng,
    origin_lat: origin.lat,
    dest_lng: dest.lng,
    dest_lat: dest.lat,
    vehicle_plate: plates[i % plates.length],
    driver_name: drivers[i % drivers.length],
    driver_phone: `138${String(10000000 + i * 137).slice(0, 8)}`,
    escort_name: `押运员${i + 1}`,
    planned_departure: planned.toISOString(),
    expected_arrival: planned.add(8 + Math.random() * 20, 'hour').toISOString(),
    actual_departure: status !== 'pending' ? planned.add(10 + Math.random() * 30, 'minute').toISOString() : undefined,
    actual_arrival: status === 'signed' ? planned.add(10 + Math.random() * 30, 'hour').toISOString() : undefined,
    progress,
    current_location: status === 'transit' ? `${origin.name} → ${dest.name} 途中` : undefined,
    current_lng: status === 'transit' ? (origin.lng + dest.lng) / 2 + (Math.random() - 0.5) * 0.5 : undefined,
    current_lat: status === 'transit' ? (origin.lat + dest.lat) / 2 + (Math.random() - 0.5) * 0.5 : undefined,
    cargo_weight: `${(15 + Math.random() * 20).toFixed(1)} 吨`,
    cargo_quantity: `${Math.floor(Math.random() * 200) + 50} 桶`,
    packaging_type: ['钢瓶', '集装箱', '槽罐车', '铁桶'][i % 4],
    emergency_contact: '安全应急部',
    emergency_phone: '400-888-110-',
    created_at: planned.subtract(3 + Math.random() * 12, 'hour').toISOString(),
    block_hash: status === 'signed' ? `0x${Array.from({ length: 64 }, () => '0123456789abcdef'[Math.floor(Math.random() * 16)]).join('')}` : undefined,
    block_height: status === 'signed' ? Math.floor(Math.random() * 100000) + 1000000 : undefined,
    status_timeline: [
      { status: '运单创建', time: planned.subtract(3 + Math.random() * 12, 'hour').toISOString(), operator: '调度员', note: '危险品运输许可已审核通过' },
      ...(status !== 'pending' ? [{ status: '车辆发车', time: planned.add(10 + Math.random() * 30, 'minute').toISOString(), operator: '车载终端', note: '已通过发车前安全检查' }] : []),
      ...(progress >= 30 && status !== 'pending' && status !== 'signed' ? [{ status: '途经检查点A', time: planned.add(3 + Math.random() * 2, 'hour').toISOString(), operator: '北斗定位', note: '行驶正常，未偏离路线' }] : []),
      ...(progress >= 60 && status !== 'pending' && status !== 'signed' ? [{ status: '途经检查点B', time: planned.add(6 + Math.random() * 3, 'hour').toISOString(), operator: '北斗定位', note: '驾驶员未超时驾驶' }] : []),
      ...(status === 'signed' ? [{ status: '货物签收', time: planned.add(10 + Math.random() * 30, 'hour').toISOString(), operator: '收货方', note: '货物完好，数量正确' }] : []),
      ...(status === 'abnormal' ? [{ status: '异常报告', time: planned.add(4, 'hour').toISOString(), operator: '押运员', note: '车辆例行检查发现轻微异常，已处理继续运输' }] : []),
      ...(status === 'cancelled' ? [{ status: '运单取消', time: planned.subtract(1, 'hour').toISOString(), operator: '调度主管', note: '客户取消订单' }] : []),
    ],
    route_points: undefined,
    receiver_signature: status === 'signed' ? 'https://gw.alipayobjects.com/zos/antfincdn/R8sN%24GNdh6/language.svg' : undefined,
    receiver_name: status === 'signed' ? ['王先生', '李女士', '张总', '陈经理'][i % 4] : undefined,
    sign_time: status === 'signed' ? planned.add(10 + Math.random() * 30, 'hour').toISOString() : undefined,
  }
})

const Waybills: React.FC = () => {
  const actionRef = useRef<ActionType>()
  const [loading, setLoading] = useState(false)
  const [detailDrawer, setDetailDrawer] = useState<WaybillItem | null>(null)
  const [mapReady, setMapReady] = useState(false)
  const mapContainerRef = useRef<HTMLDivElement>(null)
  const mapInstanceRef = useRef<any>(null)
  const [searchForm] = Form.useForm()

  const [stats, setStats] = useState({
    total: 38,
    transit: 12,
    todaySigned: 5,
    abnormal: 3,
  })

  useEffect(() => {
    const pending = mockWaybills.filter(w => w.status === 'pending').length
    const transit = mockWaybills.filter(w => w.status === 'transit').length
    const signed = mockWaybills.filter(w => w.status === 'signed').length
    const abnormal = mockWaybills.filter(w => w.status === 'abnormal').length
    const todaySigned = mockWaybills.filter(w => w.status === 'signed' && dayjs(w.actual_arrival).isSame(dayjs(), 'day')).length || 5
    setStats({
      total: mockWaybills.length,
      transit,
      todaySigned,
      abnormal,
    })
  }, [])

  const initMap = async (waybill: WaybillItem) => {
    if (!mapContainerRef.current) return
    try {
      const AMap: any = await AMapLoader.load({
        key: 'your-amap-key',
        version: '2.0',
        plugins: ['AMap.Scale', 'AMap.ToolBar'],
      })
      const centerLng = (waybill.origin_lng + waybill.dest_lng) / 2
      const centerLat = (waybill.origin_lat + waybill.dest_lat) / 2
      const map = new AMap.Map(mapContainerRef.current, {
        zoom: 7,
        center: [centerLng, centerLat],
        mapStyle: 'amap://styles/light',
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar())
      new AMap.Marker({
        position: [waybill.origin_lng, waybill.origin_lat],
        label: { content: `<div style="padding:2px 8px;background:#52c41a;color:#fff;border-radius:4px;font-size:12px">起点</div>`, direction: 'top' },
        icon: new AMap.Icon({ size: new AMap.Size(25, 34), image: 'https://webapi.amap.com/theme/v1.3/markers/n/start.png' }),
        map,
      })
      new AMap.Marker({
        position: [waybill.dest_lng, waybill.dest_lat],
        label: { content: `<div style="padding:2px 8px;background:#ff4d4f;color:#fff;border-radius:4px;font-size:12px">终点</div>`, direction: 'top' },
        icon: new AMap.Icon({ size: new AMap.Size(25, 34), image: 'https://webapi.amap.com/theme/v1.3/markers/n/end.png' }),
        map,
      })
      if (waybill.status === 'transit' && waybill.current_lng && waybill.current_lat) {
        new AMap.Marker({
          position: [waybill.current_lng, waybill.current_lat],
          label: { content: `<div style="padding:2px 8px;background:#1677ff;color:#fff;border-radius:4px;font-size:12px">${waybill.vehicle_plate}</div>`, direction: 'top' },
          icon: new AMap.Icon({ size: new AMap.Size(32, 32), image: 'https://webapi.amap.com/images/car.png' }),
          map,
        })
      }
      const path = [
        [waybill.origin_lng, waybill.origin_lat],
        [(waybill.origin_lng + waybill.dest_lng) / 2 + 0.2, (waybill.origin_lat + waybill.dest_lat) / 2 - 0.3],
        [(waybill.origin_lng + waybill.dest_lng) / 2 - 0.15, (waybill.origin_lat + waybill.dest_lat) / 2 + 0.2],
        [waybill.dest_lng, waybill.dest_lat],
      ]
      new AMap.Polyline({
        path,
        strokeColor: '#1677ff',
        strokeWeight: 5,
        strokeOpacity: 0.8,
        lineJoin: 'round',
        map,
      })
      mapInstanceRef.current = map
      setMapReady(true)
    } catch (e) {
    }
  }

  useEffect(() => {
    if (detailDrawer && !mapReady) {
      const timer = setTimeout(() => initMap(detailDrawer), 300)
      return () => { clearTimeout(timer); if (mapInstanceRef.current) { mapInstanceRef.current.destroy(); mapInstanceRef.current = null; setMapReady(false) } }
    }
    return () => {
      if (mapInstanceRef.current && !detailDrawer) { mapInstanceRef.current.destroy(); mapInstanceRef.current = null; setMapReady(false) }
    }
  }, [detailDrawer])

  const columns: ProColumns<WaybillItem>[] = [
    {
      title: '运单号',
      dataIndex: 'waybill_no',
      width: 170,
      fixed: 'left',
      render: (v: string) => (
        <Space>
          <FileTextOutlined style={{ color: '#1677ff' }} />
          <Text copyable strong style={{ fontSize: 13 }}>{v}</Text>
        </Space>
      ),
    },
    {
      title: '危险品信息',
      width: 200,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Space size={4}>
            <Tag color={dangerLevelColorMap[r.danger_level] || 'default'} style={{ margin: 0 }}>
              <strong>类 {r.danger_level}</strong>
            </Tag>
            <Text strong style={{ fontSize: 13 }}>{r.danger_goods_name}</Text>
          </Space>
          <Text type="secondary" style={{ fontSize: 11 }}>{r.un_number} · {r.packaging_type} · {r.cargo_quantity}</Text>
        </Space>
      ),
    },
    {
      title: '运输路线',
      width: 220,
      render: (_, r) => (
        <Space direction="vertical" size={2} style={{ width: '100%' }}>
          <Space align="start">
            <EnvironmentOutlined style={{ color: '#52c41a', marginTop: 4 }} />
            <div>
              <Text style={{ fontSize: 12 }}>{r.origin}</Text>
            </div>
          </Space>
          <div style={{ paddingLeft: 20 }}>
            <div style={{ width: 2, height: 12, background: '#d9d9d9', marginLeft: 5 }} />
          </div>
          <Space align="start">
            <EnvironmentOutlined style={{ color: '#ff4d4f', marginTop: 4 }} />
            <div>
              <Text style={{ fontSize: 12 }}>{r.destination}</Text>
            </div>
          </Space>
        </Space>
      ),
    },
    {
      title: '车辆/司机',
      width: 160,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Space size={4}>
            <CarOutlined style={{ color: '#1677ff' }} />
            <Tag color="blue" style={{ margin: 0 }}>{r.vehicle_plate}</Tag>
          </Space>
          <Space size={4}>
            <UserOutlined style={{ color: '#fa8c16', fontSize: 11 }} />
            <Text style={{ fontSize: 12 }}>{r.driver_name}</Text>
            <Text type="secondary" style={{ fontSize: 11 }}>/ {r.escort_name}</Text>
          </Space>
        </Space>
      ),
    },
    {
      title: '计划发车/预计到达',
      width: 180,
      render: (_, r) => (
        <Space direction="vertical" size={2}>
          <Space size={4}>
            <PlayCircleOutlined style={{ color: '#52c41a', fontSize: 11 }} />
            <Text style={{ fontSize: 12 }}>{formatDateTime(r.planned_departure, 'MM-DD HH:mm')}</Text>
          </Space>
          <Space size={4}>
            <CalendarOutlined style={{ color: '#722ed1', fontSize: 11 }} />
            <Text style={{ fontSize: 12 }}>{formatDateTime(r.expected_arrival, 'MM-DD HH:mm')}</Text>
          </Space>
        </Space>
      ),
    },
    {
      title: '运输状态',
      width: 260,
      render: (_, r) => (
        <Space direction="vertical" size={6} style={{ width: 240 }}>
          <Steps
            size="small"
            current={r.status === 'signed' ? 3 : r.status === 'transit' ? 2 : r.status === 'pending' ? 0 : r.status === 'abnormal' ? 2 : 1}
            status={r.status === 'abnormal' ? 'error' : undefined}
            items={[
              { title: '待调度', status: r.status === 'pending' ? 'process' : 'finish' },
              { title: '发车', status: r.status === 'pending' ? 'wait' : 'finish' },
              { title: '运输中', status: r.status === 'transit' ? 'process' : r.status === 'signed' ? 'finish' : r.status === 'abnormal' ? 'error' : r.status === 'cancelled' ? 'error' : 'wait' },
              { title: '签收', status: r.status === 'signed' ? 'finish' : 'wait' },
            ]}
          />
          <div>
            <Progress
              percent={r.progress}
              size="small"
              showInfo={false}
              strokeColor={r.status === 'abnormal' ? '#ff4d4f' : r.status === 'cancelled' ? '#faad14' : '#1677ff'}
              style={{ marginBottom: 4 }}
            />
            <Space>
              <Tag color={statusMap[r.status].color}>{statusMap[r.status].label}</Tag>
              <Text type="secondary" style={{ fontSize: 11 }}>{r.progress}%</Text>
            </Space>
          </div>
        </Space>
      ),
    },
    {
      title: '操作',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailDrawer(record)}>详情</Button>
          {record.status === 'pending' && (
            <Button type="link" size="small" type="primary" icon={<SendOutlined />}>调度</Button>
          )}
          {record.status === 'transit' && (
            <Button type="link" size="small" danger icon={<WarningOutlined />}>异常报备</Button>
          )}
          {record.status === 'pending' && (
            <Button type="link" size="small" icon={<CloseOutlined />}>取消</Button>
          )}
        </Space>
      ),
    },
  ]

  const requestData = async (params: any, sort: any, filter: any): Promise<RequestData<WaybillItem>> => {
    setLoading(true)
    await new Promise(r => setTimeout(r, 400))
    const { keyword, status, danger_level, date_range } = params
    let list = [...mockWaybills]
    if (keyword) {
      list = list.filter(w =>
        w.waybill_no.includes(keyword) ||
        w.origin.includes(keyword) ||
        w.destination.includes(keyword) ||
        w.danger_goods_name.includes(keyword)
      )
    }
    if (status) list = list.filter(w => w.status === status)
    if (danger_level) list = list.filter(w => w.danger_level === danger_level)
    if (date_range?.length === 2) {
      const [start, end] = date_range
      list = list.filter(w => dayjs(w.created_at).isBetween(dayjs(start), dayjs(end), 'day', '[]'))
    }
    const { current = 1, pageSize = 10 } = params
    const total = list.length
    const pageList = list.slice((current - 1) * pageSize, current * pageSize)
    setLoading(false)
    return { data: pageList, success: true, total }
  }

  const handleExport = () => {
    message.success('运单数据导出中，请稍候...')
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="运单总数"
              value={stats.total}
              valueStyle={{ color: '#1677ff' }}
              prefix={<InboxOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="运输中"
              value={stats.transit}
              valueStyle={{ color: '#13c2c2' }}
              prefix={<TruckOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日完成"
              value={stats.todaySigned}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="异常运单"
              value={stats.abnormal}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <ProCard bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
        <ProTable<WaybillItem>
          actionRef={actionRef}
          rowKey="id"
          loading={loading}
          columns={columns}
          request={requestData}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条运单`,
          }}
          scroll={{ x: 1400 }}
          search={{
            labelWidth: 'auto',
            defaultCollapsed: false,
            span: { xs: 24, sm: 12, md: 8, lg: 6 },
          }}
          form={{
            syncToUrl: false,
          }}
          toolBarRender={() => [
            <Button key="reload" icon={<ReloadOutlined />} onClick={() => actionRef.current?.reload()}>刷新</Button>,
            <Button key="export" icon={<ExportOutlined />} onClick={handleExport}>导出</Button>,
          ]}
          columnsState={{
            persistenceKey: 'waybills-columns',
            persistenceType: 'localStorage',
          }}
          formItemProps={{ style: { marginBottom: 12 } }}
          params={{}}
          dataSource={undefined}
          headerTitle={
            <Space>
              <FileTextOutlined style={{ color: '#1677ff' }} />
              <Text strong style={{ fontSize: 15 }}>电子运单管理</Text>
              <Tag color="blue">危险品运输</Tag>
            </Space>
          }
          extra={
            <Space wrap>
              <Text type="secondary" style={{ fontSize: 12 }}>
                <SafetyCertificateOutlined /> 所有运单已通过危运许可审核，全程区块链存证
              </Text>
            </Space>
          }
        >
        </ProTable>
      </ProCard>

      <Drawer
        title={
          <Space>
            <FileTextOutlined style={{ color: '#1677ff' }} />
            <Text strong>运单详情 - {detailDrawer?.waybill_no}</Text>
            {detailDrawer && (
              <Tag color={statusMap[detailDrawer.status].color}>
                {statusMap[detailDrawer.status].label}
              </Tag>
            )}
          </Space>
        }
        open={!!detailDrawer}
        onClose={() => { setDetailDrawer(null); setMapReady(false); if (mapInstanceRef.current) { mapInstanceRef.current.destroy(); mapInstanceRef.current = null } }}
        width={720}
        extra={
          detailDrawer && (
            <Space>
              <Button icon={<ExportOutlined />}>导出运单</Button>
              {detailDrawer.status === 'transit' && (
                <Button type="primary" icon={<CarOutlined />}>实时追踪</Button>
              )}
            </Space>
          )
        }
      >
        {detailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Tabs defaultActiveKey="basic" size="small">
              <TabPane tab={<Space><FileTextOutlined />基本信息</Space>} key="basic">
                <Card size="small" style={{ borderRadius: 8 }}>
                  <Descriptions column={2} size="small" bordered>
                    <Descriptions.Item label="运单号" span={2}>
                      <Text copyable strong>{detailDrawer.waybill_no}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="危险品名称">
                      <Space>
                        <Text strong>{detailDrawer.danger_goods_name}</Text>
                        <Tag color="red">{detailDrawer.un_number}</Tag>
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="危险等级">
                      <Tag color={dangerLevelColorMap[detailDrawer.danger_level] || 'default'} style={{ padding: '4px 12px' }}>
                        <strong>类 {detailDrawer.danger_level}</strong>
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="货物重量">{detailDrawer.cargo_weight}</Descriptions.Item>
                    <Descriptions.Item label="货物数量">{detailDrawer.cargo_quantity}</Descriptions.Item>
                    <Descriptions.Item label="包装方式">{detailDrawer.packaging_type}</Descriptions.Item>
                    <Descriptions.Item label="起点">
                      <Space><EnvironmentOutlined style={{ color: '#52c41a' }} />{detailDrawer.origin}</Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="终点">
                      <Space><EnvironmentOutlined style={{ color: '#ff4d4f' }} />{detailDrawer.destination}</Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="车牌号">
                      <Tag color="blue" style={{ padding: '4px 12px' }}>{detailDrawer.vehicle_plate}</Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="司机/押运">
                      <Space direction="vertical" size={0}>
                        <Text><UserOutlined /> {detailDrawer.driver_name}</Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>{detailDrawer.driver_phone}</Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>押运：{detailDrawer.escort_name}</Text>
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="计划发车">{formatDateTime(detailDrawer.planned_departure)}</Descriptions.Item>
                    <Descriptions.Item label="预计到达">{formatDateTime(detailDrawer.expected_arrival)}</Descriptions.Item>
                    {detailDrawer.actual_departure && (
                      <Descriptions.Item label="实际发车">{formatDateTime(detailDrawer.actual_departure)}</Descriptions.Item>
                    )}
                    {detailDrawer.actual_arrival && (
                      <Descriptions.Item label="实际到达">{formatDateTime(detailDrawer.actual_arrival)}</Descriptions.Item>
                    )}
                    <Descriptions.Item label="应急联系人" span={2}>
                      <Space>
                        <Tag color="warning">{detailDrawer.emergency_contact}</Tag>
                        <Text type="danger">{detailDrawer.emergency_phone}</Text>
                      </Space>
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </TabPane>

              <TabPane tab={<Space><MapOutlined />运输路径</Space>} key="map">
                <Card size="small" style={{ borderRadius: 8 }}>
                  <div
                    ref={mapContainerRef}
                    style={{
                      height: 320,
                      borderRadius: 6,
                      background: '#f0f5ff',
                      display: mapReady ? 'block' : 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    {!mapReady && (
                      <div style={{ textAlign: 'center', color: '#8c8c8c' }}>
                        <MapOutlined style={{ fontSize: 36, opacity: 0.4 }} />
                        <div style={{ marginTop: 8, fontSize: 12 }}>
                          路径信息：{detailDrawer.origin} → {detailDrawer.destination}
                        </div>
                        {detailDrawer.current_location && (
                          <div style={{ marginTop: 4, fontSize: 12, color: '#1677ff' }}>
                            当前位置：{detailDrawer.current_location}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                  <Divider style={{ margin: '12px 0' }} />
                  <Space direction="vertical" size={8} style={{ width: '100%' }}>
                    <Space>
                      <Tag color="green">起点</Tag>
                      <Text>{detailDrawer.origin}</Text>
                      <Text type="secondary">({detailDrawer.origin_lng.toFixed(4)}, {detailDrawer.origin_lat.toFixed(4)})</Text>
                    </Space>
                    {detailDrawer.status === 'transit' && detailDrawer.current_lng && (
                      <Space>
                        <Tag color="blue">当前</Tag>
                        <Text>{detailDrawer.current_location}</Text>
                        <Text type="secondary">({detailDrawer.current_lng.toFixed(4)}, {detailDrawer.current_lat.toFixed(4)})</Text>
                      </Space>
                    )}
                    <Space>
                      <Tag color="red">终点</Tag>
                      <Text>{detailDrawer.destination}</Text>
                      <Text type="secondary">({detailDrawer.dest_lng.toFixed(4)}, {detailDrawer.dest_lat.toFixed(4)})</Text>
                    </Space>
                  </Space>
                </Card>
              </TabPane>

              <TabPane tab={<Space><EditOutlined />状态流转</Space>} key="timeline">
                <Card size="small" style={{ borderRadius: 8 }}>
                  <div style={{ padding: '8px 12px' }}>
                    <Progress
                      percent={detailDrawer.progress}
                      strokeColor={detailDrawer.status === 'abnormal' ? '#ff4d4f' : '#1677ff'}
                      style={{ marginBottom: 16 }}
                    />
                    <Timeline
                      mode="left"
                      items={detailDrawer.status_timeline.map((item, idx) => ({
                        color: idx === detailDrawer.status_timeline.length - 1 ? (detailDrawer.status === 'abnormal' ? 'red' : 'blue') : 'gray',
                        label: <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(item.time, 'MM-DD HH:mm:ss')}</Text>,
                        children: (
                          <div>
                            <Space>
                              <Text strong style={{ fontSize: 13 }}>{item.status}</Text>
                              <Tag color="default" style={{ fontSize: 11, padding: '0 6px' }}>{item.operator}</Tag>
                            </Space>
                            {item.note && (
                              <Paragraph style={{ fontSize: 12, marginTop: 4, marginBottom: 0 }} type="secondary">
                                {item.note}
                              </Paragraph>
                            )}
                          </div>
                        ),
                      }))}
                    />
                  </div>
                </Card>
              </TabPane>

              <TabPane tab={<Space><SignatureOutlined />电子签收</Space>} key="sign">
                {detailDrawer.status === 'signed' ? (
                  <Row gutter={16}>
                    <Col span={12}>
                      <Card size="small" style={{ borderRadius: 8, height: '100%' }} title="电子签名">
                        <div style={{ textAlign: 'center', padding: 16 }}>
                          <div style={{
                            height: 140,
                            borderRadius: 6,
                            background: 'linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            border: '2px dashed #95de64',
                          }}>
                            <div style={{ textAlign: 'center' }}>
                              <SignatureOutlined style={{ fontSize: 42, color: '#52c41a' }} />
                              <div style={{ marginTop: 8, fontSize: 14, color: '#389e0d', fontWeight: 600 }}>
                                {detailDrawer.receiver_name}
                              </div>
                            </div>
                          </div>
                          <Divider style={{ margin: '12px 0' }} />
                          <Descriptions column={1} size="small">
                            <Descriptions.Item label="签收人">{detailDrawer.receiver_name}</Descriptions.Item>
                            <Descriptions.Item label="签收时间">{formatDateTime(detailDrawer.sign_time)}</Descriptions.Item>
                          </Descriptions>
                        </div>
                      </Card>
                    </Col>
                    <Col span={12}>
                      <Card size="small" style={{ borderRadius: 8, height: '100%' }} title={<Space><SafetyCertificateOutlined />区块链存证</Space>}>
                        <Alert
                          type="success"
                          showIcon
                          message="运单数据已上链存证，不可篡改"
                          style={{ marginBottom: 12, borderRadius: 6 }}
                        />
                        <Descriptions column={1} size="small" bordered>
                          <Descriptions.Item label="区块高度">
                            <Text copyable>{detailDrawer.block_height}</Text>
                          </Descriptions.Item>
                          <Descriptions.Item label="交易Hash">
                            <Paragraph copyable style={{ fontSize: 11, wordBreak: 'break-all', margin: 0 }}>
                              <LinkOutlined /> {detailDrawer.block_hash}
                            </Paragraph>
                          </Descriptions.Item>
                          <Descriptions.Item label="存证时间">{formatDateTime(detailDrawer.sign_time)}</Descriptions.Item>
                          <Descriptions.Item label="存证节点">
                            <Space>
                              <Badge status="success" />
                              <Text type="secondary" style={{ fontSize: 12 }}>3/5 节点已确认</Text>
                            </Space>
                          </Descriptions.Item>
                        </Descriptions>
                      </Card>
                    </Col>
                  </Row>
                ) : (
                  <Empty description="运单尚未签收，暂无签收信息" />
                )}
              </TabPane>
            </Tabs>
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default Waybills
