import React, { useEffect, useMemo, useState } from 'react'
import {
  Row,
  Col,
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Modal,
  Form,
  Select,
  Input,
  Statistic,
  Drawer,
  Descriptions,
  Progress,
  message,
  Empty,
  Timeline,
  InputNumber,
  Tooltip,
} from 'antd'
import {
  CarOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
  FilterOutlined,
  ReloadOutlined,
  DashboardOutlined,
  ToolOutlined,
  SendOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined,
  GaugeOutlined,
  FireOutlined,
  GasStationOutlined,
  ExportOutlined,
  DeleteOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { ProTable, ProFormSelect, ProFormText } from '@ant-design/pro-components'
import type { ProColumns } from '@ant-design/pro-components'
import ReactECharts from 'echarts-for-react'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import { useAppStore, VehicleItem } from '@/store/app'

const { Title, Text } = Typography
const { Option } = Select

const vehicleTypeMap: Record<string, { label: string; color: string }> = {
  tanker: { label: '罐车', color: 'blue' },
  van: { label: '厢式', color: 'purple' },
  flatbed: { label: '平板', color: 'cyan' },
  container: { label: '集装箱', color: 'geekblue' },
}

const vehicleStatusMap: Record<string, { label: string; color: string; dot: string }> = {
  online: { label: '在线', color: 'green', dot: 'success' },
  offline: { label: '离线', color: 'default', dot: 'default' },
  maintenance: { label: '维修中', color: 'orange', dot: 'warning' },
  stopped: { label: '停运', color: 'red', dot: 'error' },
}

const dangerLevelMap: Record<number, { label: string; color: string }> = {
  1: { label: '一级(剧毒)', color: 'red' },
  2: { label: '二级(易燃)', color: 'orange' },
  3: { label: '三级(腐蚀)', color: 'gold' },
  4: { label: '四级(普通)', color: 'blue' },
}

const faultLevelMap: Record<string, { color: string }> = {
  low: { color: 'blue' },
  medium: { color: 'orange' },
  high: { color: 'red' },
}

const mockVehicles: VehicleItem[] = Array.from({ length: 35 }, (_, i) => {
  const types: VehicleItem['vehicle_type'][] = ['tanker', 'van', 'flatbed', 'container']
  const statuses: VehicleItem['status'][] = ['online', 'offline', 'maintenance', 'stopped']
  const type = types[i % 4]
  const status = statuses[i % 4]
  return {
    id: i + 1,
    plate_number: `京A${(10000 + i).toString().padStart(5, '0')}`,
    vehicle_type: type,
    load_capacity: [20, 30, 40, 50][i % 4],
    danger_level: ((i % 4) + 1) as 1 | 2 | 3 | 4,
    current_driver_id: (i % 15) + 1,
    current_driver_name: ['张建国', '李志强', '王海军', '赵卫东', '陈光明', '刘振华', '杨大勇', '黄伟峰', '周明辉', '吴建军'][(i % 10)],
    status,
    current_address: [
      '北京市朝阳区建国路88号',
      '上海市浦东新区世纪大道100号',
      '广州市天河区珠江新城',
      '深圳市南山区科技园',
      '杭州市西湖区文三路',
    ][i % 5],
    latitude: 39.9 + (Math.random() * 0.5),
    longitude: 116.3 + (Math.random() * 0.5),
    last_report_time: dayjs().subtract(i * 15, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    obd_speed: status === 'online' ? Math.floor(Math.random() * 90) : 0,
    obd_rpm: status === 'online' ? 800 + Math.floor(Math.random() * 2500) : 0,
    obd_water_temp: status === 'online' ? 75 + Math.floor(Math.random() * 20) : 0,
    obd_fuel_level: 20 + Math.floor(Math.random() * 80),
    obd_engine_status: status === 'online' ? '运行正常' : '停机',
    fault_codes: i % 3 === 0 ? [
      { code: `P0${300 + i}`, desc: '发动机水温过高', level: 'high', time: dayjs().subtract(2, 'hour').format('YYYY-MM-DD HH:mm:ss') },
      { code: `P0${400 + i}`, desc: '氧传感器异常', level: 'medium', time: dayjs().subtract(5, 'hour').format('YYYY-MM-DD HH:mm:ss') },
    ] : [],
    maintenance_records: [
      { id: 1, type: '常规保养', time: dayjs().subtract(15, 'day').format('YYYY-MM-DD'), content: '更换机油、机滤、空滤', cost: 850, operator: '王师傅' },
      { id: 2, type: '轮胎更换', time: dayjs().subtract(45, 'day').format('YYYY-MM-DD'), content: '更换前轴两条轮胎', cost: 3200, operator: '李师傅' },
      { id: 3, type: '刹车检修', time: dayjs().subtract(90, 'day').format('YYYY-MM-DD'), content: '更换刹车片，检查刹车油', cost: 1200, operator: '张师傅' },
    ],
    created_at: dayjs().subtract(365 + i, 'day').format('YYYY-MM-DD HH:mm:ss'),
  }
})

const Vehicles: React.FC = () => {
  const { vehicleList, updateVehicleList, upsertVehicleItem, deleteVehicleItem } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [detailDrawer, setDetailDrawer] = useState<VehicleItem | null>(null)
  const [formModal, setFormModal] = useState<{ open: boolean; edit?: VehicleItem }>({ open: false })
  const [form] = Form.useForm()
  const [searchText, setSearchText] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [dangerFilter, setDangerFilter] = useState<number>()

  useEffect(() => {
    setLoading(true)
    setTimeout(() => {
      updateVehicleList(mockVehicles)
      setLoading(false)
    }, 500)
  }, [])

  const filteredData = useMemo(() => {
    return vehicleList.filter(v => {
      if (searchText && !v.plate_number.includes(searchText)) return false
      if (typeFilter && v.vehicle_type !== typeFilter) return false
      if (statusFilter && v.status !== statusFilter) return false
      if (dangerFilter && v.danger_level !== dangerFilter) return false
      return true
    })
  }, [vehicleList, searchText, typeFilter, statusFilter, dangerFilter])

  const stats = useMemo(() => {
    const total = vehicleList.length
    const online = vehicleList.filter(v => v.status === 'online').length
    const fault = vehicleList.filter(v => v.fault_codes.length > 0).length
    const today = vehicleList.filter(v =>
      dayjs(v.last_report_time).isSame(dayjs(), 'day') && v.status === 'online'
    ).length
    return { total, online, fault, today }
  }, [vehicleList])

  const obdTrendChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['车速(km/h)', '转速(×100rpm)'], top: 0 },
    grid: { left: 40, right: 20, top: 40, bottom: 30 },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 12 }, (_, i) => `${i * 5}分钟前`).reverse(),
    },
    yAxis: [{ type: 'value' }, { type: 'value' }],
    series: [
      {
        name: '车速(km/h)',
        type: 'line',
        smooth: true,
        areaStyle: { color: 'rgba(22,119,255,0.15)' },
        itemStyle: { color: '#1677ff' },
        data: Array.from({ length: 12 }, () => Math.floor(Math.random() * 80) + 20),
      },
      {
        name: '转速(×100rpm)',
        type: 'line',
        yAxisIndex: 1,
        smooth: true,
        itemStyle: { color: '#52c41a' },
        data: Array.from({ length: 12 }, () => Math.floor(Math.random() * 15) + 15),
      },
    ],
  }

  const handleSubmit = async (values: any) => {
    try {
      if (formModal.edit) {
        upsertVehicleItem({ ...formModal.edit, ...values })
        message.success('车辆信息已更新')
      } else {
        const newItem: VehicleItem = {
          ...mockVehicles[0],
          id: Date.now(),
          ...values,
          plate_number: values.plate_number,
          created_at: dayjs().format('YYYY-MM-DD HH:mm:ss'),
          last_report_time: dayjs().format('YYYY-MM-DD HH:mm:ss'),
        }
        upsertVehicleItem(newItem)
        message.success('车辆添加成功')
      }
      setFormModal({ open: false })
      form.resetFields()
    } catch (e) {
      message.error('操作失败')
    }
  }

  const handleDelete = (record: VehicleItem) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除车辆 ${record.plate_number} 吗？`,
      okText: '确定删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: () => {
        deleteVehicleItem(record.id)
        message.success('删除成功')
      },
    })
  }

  const columns: ProColumns<VehicleItem>[] = [
    {
      title: '车牌',
      dataIndex: 'plate_number',
      width: 140,
      render: (v: string) => (
        <Tag color="blue" style={{ fontSize: 13, fontWeight: 600 }}>{v}</Tag>
      ),
    },
    {
      title: '车辆类型',
      dataIndex: 'vehicle_type',
      width: 100,
      valueEnum: Object.fromEntries(
        Object.entries(vehicleTypeMap).map(([k, v]) => [k, { text: <Tag color={v.color}>{v.label}</Tag> }])
      ),
    },
    {
      title: '载重(吨)',
      dataIndex: 'load_capacity',
      width: 100,
      sorter: (a, b) => a.load_capacity - b.load_capacity,
    },
    {
      title: '危险品等级',
      dataIndex: 'danger_level',
      width: 120,
      valueEnum: Object.fromEntries(
        Object.entries(dangerLevelMap).map(([k, v]) => [k, { text: <Tag color={v.color}>{v.label}</Tag> }])
      ),
    },
    {
      title: '当前司机',
      dataIndex: 'current_driver_name',
      width: 110,
      render: (v: string) => <Space><Text>{v}</Text></Space>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: Object.fromEntries(
        Object.entries(vehicleStatusMap).map(([k, v]) => [k, { text: <Tag color={v.color}>{v.label}</Tag> }])
      ),
    },
    {
      title: '最新位置',
      dataIndex: 'current_address',
      ellipsis: true,
      render: (v: string, r) => (
        <Tooltip title={v}>
          <span>
            <EnvironmentOutlined style={{ color: '#1677ff' }} />
            {' '}{v || `${r.latitude?.toFixed(4)}, ${r.longitude?.toFixed(4)}`}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '最近上报',
      dataIndex: 'last_report_time',
      width: 170,
      sorter: (a, b) => dayjs(a.last_report_time).valueOf() - dayjs(b.last_report_time).valueOf(),
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(v)}</Text>,
    },
    {
      title: '操作',
      width: 240,
      fixed: 'right' as const,
      render: (_, record: VehicleItem) => (
        <Space size={2} wrap>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailDrawer(record)}>查看</Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => { setFormModal({ open: true, edit: record }); form.setFieldsValue(record) }}>编辑</Button>
          <Button type="link" size="small" icon={<ToolOutlined />} onClick={() => message.success(`已对 ${record.plate_number} 发起远程诊断`)}>诊断</Button>
          <Button type="link" size="small" type="primary" icon={<SendOutlined />} onClick={() => message.success(`已打开 ${record.plate_number} 派单窗口`)}>派单</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="车辆总数"
              value={stats.total}
              valueStyle={{ color: '#1677ff' }}
              prefix={<CarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="在线车辆"
              value={stats.online}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ {stats.total}</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="故障车辆"
              value={stats.fault}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日出车"
              value={stats.today}
              valueStyle={{ color: '#722ed1' }}
              prefix={<DashboardOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <CarOutlined style={{ color: '#1677ff' }} />
            <Text strong style={{ fontSize: 15 }}>车辆列表</Text>
          </Space>
        }
        extra={
          <Space wrap>
            <Input
              allowClear
              placeholder="搜索车牌号..."
              prefix={<SearchOutlined />}
              style={{ width: 180 }}
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
            />
            <Select
              allowClear
              placeholder="车辆类型"
              style={{ width: 130 }}
              value={typeFilter}
              onChange={setTypeFilter}
            >
              {Object.entries(vehicleTypeMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="车辆状态"
              style={{ width: 130 }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              {Object.entries(vehicleStatusMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="危险品等级"
              style={{ width: 150 }}
              value={dangerFilter}
              onChange={setDangerFilter}
            >
              {Object.entries(dangerLevelMap).map(([k, v]) => (
                <Option key={k} value={Number(k)}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => { setSearchText(''); setTypeFilter(undefined); setStatusFilter(undefined); setDangerFilter(undefined) }}>重置</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setLoading(true); setTimeout(() => setLoading(false), 300) }}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => { setFormModal({ open: true }); form.resetFields() }}
            >
              新建车辆
            </Button>
          </Space>
        }
      >
        <ProTable<VehicleItem>
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={filteredData}
          search={false}
          pagination={{
            defaultPageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条`,
          }}
          scroll={{ x: 1400 }}
          toolBarRender={false}
        />
      </Card>

      <Drawer
        title={
          <Space>
            <CarOutlined style={{ color: '#1677ff' }} />
            <Text strong>车辆详情 - {detailDrawer?.plate_number}</Text>
            {detailDrawer && (
              <Tag color={vehicleStatusMap[detailDrawer.status]?.color}>
                {vehicleStatusMap[detailDrawer.status]?.label}
              </Tag>
            )}
          </Space>
        }
        open={!!detailDrawer}
        onClose={() => setDetailDrawer(null)}
        width={640}
        extra={
          detailDrawer && (
            <Space>
              <Button icon={<ToolOutlined />} onClick={() => message.success(`正在诊断 ${detailDrawer.plate_number}`)}>远程诊断</Button>
              <Button type="primary" icon={<EditOutlined />} onClick={() => { setFormModal({ open: true, edit: detailDrawer }); form.setFieldsValue(detailDrawer); setDetailDrawer(null) }}>编辑信息</Button>
            </Space>
          )
        }
      >
        {detailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Card size="small" style={{ borderRadius: 8 }} title="基本信息">
              <Descriptions column={2} size="small" bordered>
                <Descriptions.Item label="车牌号">{detailDrawer.plate_number}</Descriptions.Item>
                <Descriptions.Item label="车辆类型">
                  <Tag color={vehicleTypeMap[detailDrawer.vehicle_type]?.color}>
                    {vehicleTypeMap[detailDrawer.vehicle_type]?.label}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="载重">{detailDrawer.load_capacity} 吨</Descriptions.Item>
                <Descriptions.Item label="危险品等级">
                  <Tag color={dangerLevelMap[detailDrawer.danger_level]?.color}>
                    {dangerLevelMap[detailDrawer.danger_level]?.label}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="当前司机">{detailDrawer.current_driver_name}</Descriptions.Item>
                <Descriptions.Item label="入库时间">{formatDateTime(detailDrawer.created_at)}</Descriptions.Item>
                <Descriptions.Item label="最新位置" span={2}>
                  <EnvironmentOutlined style={{ color: '#1677ff' }} /> {detailDrawer.current_address}
                </Descriptions.Item>
                <Descriptions.Item label="最后上报" span={2}>{formatDateTime(detailDrawer.last_report_time)}</Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><DashboardOutlined style={{ color: '#1677ff' }} /> OBD实时数据</Space>}>
              <Row gutter={[12, 12]}>
                <Col xs={12} sm={6}>
                  <div style={{ padding: 12, background: '#f0f5ff', borderRadius: 8, textAlign: 'center' }}>
                    <GaugeOutlined style={{ fontSize: 24, color: '#1677ff' }} />
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 12 }}>车速</Text>
                    </div>
                    <div style={{ fontSize: 22, fontWeight: 700, color: '#1677ff' }}>
                      {detailDrawer.obd_speed}<Text style={{ fontSize: 12, fontWeight: 400 }}> km/h</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ padding: 12, background: '#f6ffed', borderRadius: 8, textAlign: 'center' }}>
                    <ThunderboltOutlined style={{ fontSize: 24, color: '#52c41a' }} />
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 12 }}>转速</Text>
                    </div>
                    <div style={{ fontSize: 22, fontWeight: 700, color: '#52c41a' }}>
                      {detailDrawer.obd_rpm}<Text style={{ fontSize: 12, fontWeight: 400 }}> rpm</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ padding: 12, background: '#fff7e6', borderRadius: 8, textAlign: 'center' }}>
                    <FireOutlined style={{ fontSize: 24, color: '#fa8c16' }} />
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 12 }}>水温</Text>
                    </div>
                    <div style={{ fontSize: 22, fontWeight: 700, color: '#fa8c16' }}>
                      {detailDrawer.obd_water_temp}<Text style={{ fontSize: 12, fontWeight: 400 }}> °C</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ padding: 12, background: '#f9f0ff', borderRadius: 8, textAlign: 'center' }}>
                    <GasStationOutlined style={{ fontSize: 24, color: '#722ed1' }} />
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 12 }}>油量</Text>
                    </div>
                    <Progress
                      percent={detailDrawer.obd_fuel_level}
                      size="small"
                      strokeColor={detailDrawer.obd_fuel_level < 30 ? '#ff4d4f' : '#722ed1'}
                      style={{ marginTop: 4 }}
                    />
                    <div style={{ fontSize: 14, fontWeight: 600, color: '#722ed1' }}>
                      {detailDrawer.obd_engine_status}
                    </div>
                  </div>
                </Col>
              </Row>
              <div style={{ marginTop: 16 }}>
                <ReactECharts option={obdTrendChart} style={{ height: 200 }} notMerge />
              </div>
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><WarningOutlined style={{ color: '#ff4d4f' }} /> 故障码列表</Space>}>
              {detailDrawer.fault_codes.length > 0 ? (
                <Row gutter={[8, 8]}>
                  {detailDrawer.fault_codes.map(f => (
                    <Col span={24} key={f.code}>
                      <div style={{
                        padding: 10,
                        border: '1px solid #f0f0f0',
                        borderRadius: 6,
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}>
                        <Space>
                          <Tag color={faultLevelMap[f.level]?.color}>{f.code}</Tag>
                          <Text>{f.desc}</Text>
                        </Space>
                        <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(f.time)}</Text>
                      </div>
                    </Col>
                  ))}
                </Row>
              ) : (
                <Empty description="暂无故障码，车况良好" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><ToolOutlined style={{ color: '#1677ff' }} /> 维修记录</Space>}>
              <Timeline
                items={detailDrawer.maintenance_records.map(m => ({
                  color: 'blue',
                  children: (
                    <div>
                      <Space>
                        <Tag color="geekblue">{m.type}</Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>{m.time}</Text>
                      </Space>
                      <div style={{ marginTop: 4 }}>
                        <Text>{m.content}</Text>
                      </div>
                      <div style={{ marginTop: 2 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          操作人：{m.operator} · 费用：¥{m.cost}
                        </Text>
                      </div>
                    </div>
                  ),
                }))}
              />
            </Card>
          </div>
        )}
      </Drawer>

      <Modal
        title={<Space><PlusOutlined style={{ color: '#1677ff' }} /> {formModal.edit ? '编辑车辆' : '新建车辆'}</Space>}
        open={formModal.open}
        onCancel={() => { setFormModal({ open: false }); form.resetFields() }}
        onOk={() => form.submit()}
        okText="提交"
        width={560}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            danger_level: 4,
            status: 'offline',
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="车牌号"
                name="plate_number"
                rules={[{ required: true, message: '请输入车牌号' }]}
              >
                <Input placeholder="例如：京A12345" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="车辆类型"
                name="vehicle_type"
                rules={[{ required: true, message: '请选择车辆类型' }]}
              >
                <Select placeholder="请选择">
                  {Object.entries(vehicleTypeMap).map(([k, v]) => (
                    <Option key={k} value={k}>{v.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="载重(吨)"
                name="load_capacity"
                rules={[{ required: true, message: '请输入载重' }]}
              >
                <InputNumber min={1} max={100} style={{ width: '100%' }} placeholder="请输入载重吨数" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="危险品等级"
                name="danger_level"
                rules={[{ required: true, message: '请选择危险品等级' }]}
              >
                <Select placeholder="请选择">
                  {Object.entries(dangerLevelMap).map(([k, v]) => (
                    <Option key={k} value={Number(k)}>{v.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="当前司机"
                name="current_driver_name"
              >
                <Input placeholder="请输入司机姓名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="车辆状态"
                name="status"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="请选择">
                  {Object.entries(vehicleStatusMap).map(([k, v]) => (
                    <Option key={k} value={k}>{v.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={24}>
              <Form.Item label="当前位置" name="current_address">
                <Input placeholder="请输入当前位置地址" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  )
}

export default Vehicles
