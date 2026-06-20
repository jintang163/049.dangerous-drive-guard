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
  List,
  InputNumber,
  Tooltip,
  Divider,
  Avatar,
  Badge,
  DatePicker,
  Rate,
} from 'antd'
import {
  UserOutlined,
  TeamOutlined,
  WarningOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
  FilterOutlined,
  ReloadOutlined,
  StarOutlined,
  SearchOutlined,
  SafetyCertificateOutlined,
  IdcardOutlined,
  PhoneOutlined,
  CarOutlined,
  CoffeeOutlined,
  ExportOutlined,
  DeleteOutlined,
  FileTextOutlined,
  RiseOutlined,
  ClockCircleOutlined,
  EnvironmentOutlined,
} from '@ant-design/icons'
import { ProTable } from '@ant-design/pro-components'
import type { ProColumns } from '@ant-design/pro-components'
import ReactECharts from 'echarts-for-react'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import { useAppStore, DriverItem } from '@/store/app'

const { Title, Text, Paragraph } = Typography
const { Option } = Select

const driverStatusMap: Record<string, { label: string; color: string }> = {
  on_duty: { label: '在岗', color: 'green' },
  rest: { label: '休息', color: 'blue' },
  off_duty: { label: '离岗', color: 'default' },
}

const riskLevelMap: Record<string, { label: string; color: string }> = {
  low: { label: '低风险', color: 'green' },
  medium: { label: '中风险', color: 'orange' },
  high: { label: '高风险', color: 'red' },
}

const licenseTypes = ['A1', 'A2', 'A3', 'B1', 'B2', 'C1', 'C2']

const waybillStatusMap: Record<string, { label: string; color: string }> = {
  completed: { label: '已完成', color: 'green' },
  in_transit: { label: '运输中', color: 'blue' },
  pending: { label: '待出发', color: 'orange' },
}

const mockDrivers: DriverItem[] = Array.from({ length: 40 }, (_, i) => {
  const statuses: DriverItem['status'][] = ['on_duty', 'rest', 'off_duty']
  const risks: DriverItem['risk_level'][] = ['low', 'medium', 'high']
  const status = statuses[i % 3]
  const risk = risks[i % 3]
  const fatigueBase = risk === 'high' ? 12 : risk === 'medium' ? 6 : 2
  const scoreBase = risk === 'high' ? 68 : risk === 'medium' ? 82 : 92
  return {
    id: i + 1,
    name: ['张建国', '李志强', '王海军', '赵卫东', '陈光明', '刘振华', '杨大勇', '黄伟峰', '周明辉', '吴建军', '郑海涛', '孙文博'][(i % 12)],
    employee_no: `DR${(2024000 + i).toString().padStart(7, '0')}`,
    license_type: licenseTypes[i % 7],
    driving_years: 3 + (i % 18),
    phone: `138${(10000000 + i * 137).toString().slice(0, 8)}`,
    linked_vehicle_id: (i % 25) + 1,
    linked_vehicle_plate: `京A${(10000 + (i % 25)).toString().padStart(5, '0')}`,
    status,
    risk_level: risk,
    fatigue_count_30d: fatigueBase + Math.floor(Math.random() * 5),
    driving_score: scoreBase - Math.floor(Math.random() * 8),
    id_card_no: `110101${1980 + (i % 20)}${(i * 37 % 10000).toString().padStart(4, '0')}${(i * 17 % 10000).toString().padStart(4, '0')}`,
    license_no: `110101${1980 + (i % 20)}${(i * 23 % 1000000).toString().padStart(6, '0')}`,
    license_expire_date: dayjs().add(3 + (i % 5), 'year').format('YYYY-MM-DD'),
    qualification_cert_no: `危运证字${(2020000 + i).toString()}号`,
    qualification_cert_expire: dayjs().add(1 + (i % 3), 'year').format('YYYY-MM-DD'),
    score_radar: {
      safety: scoreBase - Math.floor(Math.random() * 10) + 5,
      fatigue: scoreBase - Math.floor(Math.random() * 15),
      speed: scoreBase - Math.floor(Math.random() * 8) + 3,
      lane: scoreBase - Math.floor(Math.random() * 5) + 5,
      focus: scoreBase - Math.floor(Math.random() * 12),
      compliance: scoreBase - Math.floor(Math.random() * 6) + 4,
    },
    fatigue_trend_30d: Array.from({ length: 30 }, (_, di) => ({
      date: dayjs().subtract(29 - di, 'day').format('MM-DD'),
      count: Math.max(0, Math.floor(Math.random() * 3) + (risk === 'high' ? 1 : 0) - (di % 7 === 0 ? 1 : 0)),
    })),
    waybills: [
      { id: 1, waybill_no: `WB${dayjs().format('YYYYMMDD')}${(1000 + i).toString()}`, from: '北京市朝阳区化工园区', to: '上海市浦东新区危险品仓库', date: dayjs().subtract(i, 'day').format('YYYY-MM-DD'), status: i % 3 === 0 ? 'in_transit' : 'completed' },
      { id: 2, waybill_no: `WB${dayjs().subtract(3, 'day').format('YYYYMMDD')}${(2000 + i).toString()}`, from: '广州市天河区物流中心', to: '深圳市南山区港口', date: dayjs().subtract(i + 4, 'day').format('YYYY-MM-DD'), status: 'completed' },
      { id: 3, waybill_no: `WB${dayjs().subtract(7, 'day').format('YYYYMMDD')}${(3000 + i).toString()}`, from: '杭州市萧山区化工园', to: '南京市江宁区仓储基地', date: dayjs().subtract(i + 8, 'day').format('YYYY-MM-DD'), status: 'completed' },
    ],
    created_at: dayjs().subtract(500 + i * 15, 'day').format('YYYY-MM-DD HH:mm:ss'),
  }
})

const Drivers: React.FC = () => {
  const { driverList, updateDriverList, upsertDriverItem, deleteDriverItem } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [detailDrawer, setDetailDrawer] = useState<DriverItem | null>(null)
  const [formModal, setFormModal] = useState<{ open: boolean; edit?: DriverItem }>({ open: false })
  const [form] = Form.useForm()
  const [searchText, setSearchText] = useState('')
  const [licenseFilter, setLicenseFilter] = useState<string>()
  const [statusFilter, setStatusFilter] = useState<string>()
  const [riskFilter, setRiskFilter] = useState<string>()

  useEffect(() => {
    setLoading(true)
    setTimeout(() => {
      updateDriverList(mockDrivers)
      setLoading(false)
    }, 500)
  }, [])

  const filteredData = useMemo(() => {
    return driverList.filter(d => {
      if (searchText) {
        const keyword = searchText.toLowerCase()
        const matchName = d.name.toLowerCase().includes(keyword)
        const matchLicense = d.license_no.toLowerCase().includes(keyword)
        if (!matchName && !matchLicense) return false
      }
      if (licenseFilter && d.license_type !== licenseFilter) return false
      if (statusFilter && d.status !== statusFilter) return false
      if (riskFilter && d.risk_level !== riskFilter) return false
      return true
    })
  }, [driverList, searchText, licenseFilter, statusFilter, riskFilter])

  const stats = useMemo(() => {
    const total = driverList.length
    const onDuty = driverList.filter(d => d.status === 'on_duty').length
    const highRisk = driverList.filter(d => d.risk_level === 'high').length
    const avgScore = total > 0
      ? Math.round(driverList.reduce((s, d) => s + d.driving_score, 0) / total)
      : 0
    return { total, onDuty, highRisk, avgScore }
  }, [driverList])

  const radarOption = (data: DriverItem['score_radar']) => ({
    tooltip: {},
    radar: {
      indicator: [
        { name: '安全驾驶', max: 100 },
        { name: '防疲劳', max: 100 },
        { name: '限速遵守', max: 100 },
        { name: '车道保持', max: 100 },
        { name: '注意力', max: 100 },
        { name: '合规性', max: 100 },
      ],
      radius: 90,
    },
    series: [{
      type: 'radar',
      data: [{
        value: [data.safety, data.fatigue, data.speed, data.lane, data.focus, data.compliance],
        name: '驾驶评分',
        areaStyle: { color: 'rgba(22,119,255,0.2)' },
        lineStyle: { color: '#1677ff' },
        itemStyle: { color: '#1677ff' },
      }],
    }],
  })

  const fatigueTrendOption = (data: DriverItem['fatigue_trend_30d']) => ({
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: data.map(d => d.date),
      axisLabel: { fontSize: 10, interval: 3 },
    },
    yAxis: { type: 'value', max: 5 },
    series: [{
      type: 'line',
      smooth: true,
      areaStyle: { color: 'rgba(250,140,22,0.2)' },
      itemStyle: { color: '#fa8c16' },
      lineStyle: { color: '#fa8c16' },
      data: data.map(d => d.count),
    }],
  })

  const handleSubmit = async (values: any) => {
    try {
      if (formModal.edit) {
        upsertDriverItem({ ...formModal.edit, ...values })
        message.success('驾驶员信息已更新')
      } else {
        const newItem: DriverItem = {
          ...mockDrivers[0],
          id: Date.now(),
          ...values,
          name: values.name,
          created_at: dayjs().format('YYYY-MM-DD HH:mm:ss'),
          score_radar: mockDrivers[0].score_radar,
          fatigue_trend_30d: mockDrivers[0].fatigue_trend_30d,
          waybills: [],
        }
        upsertDriverItem(newItem)
        message.success('驾驶员添加成功')
      }
      setFormModal({ open: false })
      form.resetFields()
    } catch (e) {
      message.error('操作失败')
    }
  }

  const handleDelete = (record: DriverItem) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除驾驶员 ${record.name} 吗？`,
      okText: '确定删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: () => {
        deleteDriverItem(record.id)
        message.success('删除成功')
      },
    })
  }

  const columns: ProColumns<DriverItem>[] = [
    {
      title: '姓名',
      dataIndex: 'name',
      width: 110,
      render: (v: string, r) => (
        <Space>
          <Avatar icon={<UserOutlined />} style={{ backgroundColor: `hsl(${(r.id * 37) % 360}, 60%, 50%)` }} />
          <div>
            <Text strong>{v}</Text>
            <div>
              <Text type="secondary" style={{ fontSize: 11 }}>{r.employee_no}</Text>
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: '工号',
      dataIndex: 'employee_no',
      width: 130,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '准驾车型',
      dataIndex: 'license_type',
      width: 100,
      render: (v: string) => <Tag color="geekblue" style={{ fontWeight: 600 }}>{v}</Tag>,
    },
    {
      title: '驾龄',
      dataIndex: 'driving_years',
      width: 80,
      sorter: (a, b) => a.driving_years - b.driving_years,
      render: (v: number) => `${v} 年`,
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      width: 130,
      render: (v: string) => (
        <Space size={4}>
          <PhoneOutlined style={{ color: '#52c41a', fontSize: 12 }} />
          <Text copyable style={{ fontSize: 12 }}>{v}</Text>
        </Space>
      ),
    },
    {
      title: '关联车辆',
      dataIndex: 'linked_vehicle_plate',
      width: 120,
      render: (v: string) => v ? <Tag color="blue">{v}</Tag> : <Text type="secondary">未分配</Text>,
    },
    {
      title: '近30天疲劳',
      dataIndex: 'fatigue_count_30d',
      width: 110,
      sorter: (a, b) => a.fatigue_count_30d - b.fatigue_count_30d,
      render: (v: number) => (
        <Badge
          count={`${v}次`}
          style={{ backgroundColor: v >= 10 ? '#ff4d4f' : v >= 5 ? '#fa8c16' : '#52c41a' }}
        />
      ),
    },
    {
      title: '驾驶评分',
      dataIndex: 'driving_score',
      width: 140,
      sorter: (a, b) => a.driving_score - b.driving_score,
      render: (v: number) => (
        <Progress
          percent={v}
          size="small"
          showInfo
          strokeColor={v >= 90 ? '#52c41a' : v >= 75 ? '#faad14' : '#ff4d4f'}
          format={p => (
            <Space size={2}>
              <Text strong style={{
                color: v >= 90 ? '#52c41a' : v >= 75 ? '#faad14' : '#ff4d4f',
                fontSize: 12,
              }}>{p}</Text>
              <StarOutlined style={{ color: '#faad14', fontSize: 11 }} />
            </Space>
          )}
        />
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      valueEnum: Object.fromEntries(
        Object.entries(driverStatusMap).map(([k, v]) => [k, { text: <Tag color={v.color}>{v.label}</Tag> }])
      ),
    },
    {
      title: '操作',
      width: 140,
      fixed: 'right' as const,
      render: (_, record: DriverItem) => (
        <Space size={2}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailDrawer(record)}>详情</Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => { setFormModal({ open: true, edit: record }); form.setFieldsValue(record) }}>编辑</Button>
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
              title="司机总数"
              value={stats.total}
              valueStyle={{ color: '#1677ff' }}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="在岗人数"
              value={stats.onDuty}
              valueStyle={{ color: '#52c41a' }}
              prefix={<UserOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ {stats.total}</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="高风险司机"
              value={stats.highRisk}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日平均评分"
              value={stats.avgScore}
              valueStyle={{ color: '#faad14' }}
              prefix={<StarOutlined />}
              suffix={
                <Text type="secondary" style={{ fontSize: 14 }}>
                  <Rate disabled allowHalf defaultValue={stats.avgScore / 20} style={{ fontSize: 12 }} />
                </Text>
              }
            />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <TeamOutlined style={{ color: '#1677ff' }} />
            <Text strong style={{ fontSize: 15 }}>驾驶员列表</Text>
          </Space>
        }
        extra={
          <Space wrap>
            <Input
              allowClear
              placeholder="姓名/驾驶证号"
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
            />
            <Select
              allowClear
              placeholder="准驾车型"
              style={{ width: 120 }}
              value={licenseFilter}
              onChange={setLicenseFilter}
            >
              {licenseTypes.map(t => (
                <Option key={t} value={t}><Tag color="geekblue">{t}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="状态筛选"
              style={{ width: 120 }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              {Object.entries(driverStatusMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="风险等级"
              style={{ width: 120 }}
              value={riskFilter}
              onChange={setRiskFilter}
            >
              {Object.entries(riskLevelMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => { setSearchText(''); setLicenseFilter(undefined); setStatusFilter(undefined); setRiskFilter(undefined) }}>重置</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setLoading(true); setTimeout(() => setLoading(false), 300) }}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => { setFormModal({ open: true }); form.resetFields() }}
            >
              新建驾驶员
            </Button>
          </Space>
        }
      >
        <ProTable<DriverItem>
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
          scroll={{ x: 1500 }}
          toolBarRender={false}
          rowClassName={(r) => r.risk_level === 'high' ? '!bg-red-50' : r.risk_level === 'medium' ? '!bg-orange-50' : ''}
        />
      </Card>

      <Drawer
        title={
          <Space>
            <Avatar icon={<UserOutlined />} style={{ backgroundColor: `hsl(${(detailDrawer?.id || 0) * 37 % 360}, 60%, 50%)` }} />
            <div>
              <Text strong style={{ fontSize: 16 }}>{detailDrawer?.name}</Text>
              <div>
                {detailDrawer && (
                  <Space size={4}>
                    <Tag color={driverStatusMap[detailDrawer.status]?.color}>
                      {driverStatusMap[detailDrawer.status]?.label}
                    </Tag>
                    <Tag color={riskLevelMap[detailDrawer.risk_level]?.color}>
                      {riskLevelMap[detailDrawer.risk_level]?.label}
                    </Tag>
                  </Space>
                )}
              </div>
            </div>
          </Space>
        }
        open={!!detailDrawer}
        onClose={() => setDetailDrawer(null)}
        width={680}
        extra={
          detailDrawer && (
            <Space>
              <Button icon={<PhoneOutlined />} onClick={() => message.success(`正在拨打 ${detailDrawer.phone}`)}>电话联系</Button>
              <Button type="primary" icon={<EditOutlined />} onClick={() => { setFormModal({ open: true, edit: detailDrawer }); form.setFieldsValue(detailDrawer); setDetailDrawer(null) }}>编辑</Button>
            </Space>
          )
        }
      >
        {detailDrawer && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Card size="small" style={{ borderRadius: 8 }} title={<Space><IdcardOutlined style={{ color: '#1677ff' }} /> 基本信息</Space>}>
              <Descriptions column={2} size="small" bordered>
                <Descriptions.Item label="工号">{detailDrawer.employee_no}</Descriptions.Item>
                <Descriptions.Item label="手机号">
                  <Space>
                    <PhoneOutlined style={{ color: '#52c41a' }} />
                    <Text copyable>{detailDrawer.phone}</Text>
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="准驾车型">
                  <Tag color="geekblue" style={{ fontWeight: 600 }}>{detailDrawer.license_type}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="驾龄">{detailDrawer.driving_years} 年</Descriptions.Item>
                <Descriptions.Item label="关联车辆">
                  {detailDrawer.linked_vehicle_plate
                    ? <Tag color="blue"><CarOutlined /> {detailDrawer.linked_vehicle_plate}</Tag>
                    : <Text type="secondary">未分配</Text>}
                </Descriptions.Item>
                <Descriptions.Item label="入职时间">{formatDateTime(detailDrawer.created_at)}</Descriptions.Item>
                <Descriptions.Item label="驾驶评分" span={2}>
                  <Space direction="vertical" style={{ width: '100%' }} size={0}>
                    <Progress
                      percent={detailDrawer.driving_score}
                      strokeColor={detailDrawer.driving_score >= 90 ? '#52c41a' : detailDrawer.driving_score >= 75 ? '#faad14' : '#ff4d4f'}
                    />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      <StarOutlined style={{ color: '#faad14' }} /> 近30天综合评分
                    </Text>
                  </Space>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><SafetyCertificateOutlined style={{ color: '#52c41a' }} /> 资质证件</Space>}>
              <Row gutter={[12, 12]}>
                <Col xs={24} sm={12}>
                  <div style={{ padding: 12, border: '1px solid #f0f0f0', borderRadius: 8 }}>
                    <Space>
                      <IdcardOutlined style={{ fontSize: 20, color: '#1677ff' }} />
                      <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>驾驶证号</Text>
                        <div><Text copyable>{detailDrawer.license_no}</Text></div>
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          有效期至 {detailDrawer.license_expire_date}
                        </Text>
                      </div>
                    </Space>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div style={{ padding: 12, border: '1px solid #f0f0f0', borderRadius: 8 }}>
                    <Space>
                      <SafetyCertificateOutlined style={{ fontSize: 20, color: '#52c41a' }} />
                      <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>危险品从业资格证</Text>
                        <div><Text copyable>{detailDrawer.qualification_cert_no}</Text></div>
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          有效期至 {detailDrawer.qualification_cert_expire}
                        </Text>
                      </div>
                    </Space>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div style={{ padding: 12, border: '1px solid #f0f0f0', borderRadius: 8 }}>
                    <Space>
                      <FileTextOutlined style={{ fontSize: 20, color: '#722ed1' }} />
                      <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>身份证号</Text>
                        <div><Text copyable>{detailDrawer.id_card_no}</Text></div>
                      </div>
                    </Space>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div style={{ padding: 12, border: '1px solid #f0f0f0', borderRadius: 8 }}>
                    <Space>
                      <CoffeeOutlined style={{ fontSize: 20, color: '#fa8c16' }} />
                      <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>近30天疲劳预警</Text>
                        <div style={{ fontSize: 20, fontWeight: 700, color: detailDrawer.fatigue_count_30d >= 10 ? '#ff4d4f' : '#fa8c16' }}>
                          {detailDrawer.fatigue_count_30d} <Text style={{ fontSize: 12, fontWeight: 400 }}>次</Text>
                        </div>
                      </div>
                    </Space>
                  </div>
                </Col>
              </Row>
            </Card>

            <Row gutter={16}>
              <Col xs={24} lg={12}>
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><RiseOutlined style={{ color: '#722ed1' }} /> 驾驶评分雷达图</Space>}>
                  <ReactECharts option={radarOption(detailDrawer.score_radar)} style={{ height: 280 }} notMerge />
                </Card>
              </Col>
              <Col xs={24} lg={12}>
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><ClockCircleOutlined style={{ color: '#fa8c16' }} /> 近30天疲劳趋势</Space>}>
                  <ReactECharts option={fatigueTrendOption(detailDrawer.fatigue_trend_30d)} style={{ height: 280 }} notMerge />
                </Card>
              </Col>
            </Row>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><CarOutlined style={{ color: '#1677ff' }} /> 历史运单</Space>}>
              {detailDrawer.waybills.length > 0 ? (
                <List
                  size="small"
                  dataSource={detailDrawer.waybills}
                  renderItem={(item) => (
                    <List.Item
                      actions={[
                        <Tag key="status" color={waybillStatusMap[item.status]?.color}>
                          {waybillStatusMap[item.status]?.label}
                        </Tag>,
                      ]}
                    >
                      <List.Item.Meta
                        avatar={<CarOutlined style={{ fontSize: 20, color: '#1677ff' }} />}
                        title={
                          <Space>
                            <Text strong>{item.waybill_no}</Text>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              <ClockCircleOutlined /> {item.date}
                            </Text>
                          </Space>
                        }
                        description={
                          <div>
                            <Space direction="vertical" size={0}>
                              <Text>
                                <EnvironmentOutlined style={{ color: '#52c41a' }} /> {item.from}
                              </Text>
                              <Text type="secondary">
                                <EnvironmentOutlined style={{ color: '#ff4d4f' }} /> {item.to}
                              </Text>
                            </Space>
                          </div>
                        }
                      />
                    </List.Item>
                  )}
                />
              ) : (
                <Empty description="暂无运单记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </div>
        )}
      </Drawer>

      <Modal
        title={<Space><PlusOutlined style={{ color: '#1677ff' }} /> {formModal.edit ? '编辑驾驶员' : '新建驾驶员'}</Space>}
        open={formModal.open}
        onCancel={() => { setFormModal({ open: false }); form.resetFields() }}
        onOk={() => form.submit()}
        okText="提交"
        width={640}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            status: 'rest',
            risk_level: 'low',
          }}
        >
          <Divider orientation="left" plain style={{ margin: '4px 0 8px' }}>基本信息</Divider>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="姓名"
                name="name"
                rules={[{ required: true, message: '请输入姓名' }]}
              >
                <Input placeholder="请输入驾驶员姓名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="工号"
                name="employee_no"
                rules={[{ required: true, message: '请输入工号' }]}
              >
                <Input placeholder="自动生成或手动输入" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="手机号"
                name="phone"
                rules={[
                  { required: true, message: '请输入手机号' },
                  { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
                ]}
              >
                <Input placeholder="请输入11位手机号" maxLength={11} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="身份证号"
                name="id_card_no"
                rules={[{ required: true, message: '请输入身份证号' }]}
              >
                <Input placeholder="请输入18位身份证号" maxLength={18} />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left" plain style={{ margin: '4px 0 8px' }}>驾驶资质</Divider>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="准驾车型"
                name="license_type"
                rules={[{ required: true, message: '请选择准驾车型' }]}
              >
                <Select placeholder="请选择">
                  {licenseTypes.map(t => <Option key={t} value={t}>{t}</Option>)}
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="驾驶证号"
                name="license_no"
                rules={[{ required: true, message: '请输入驾驶证号' }]}
              >
                <Input placeholder="请输入驾驶证号" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="有效期至"
                name="license_expire_date"
                rules={[{ required: true, message: '请选择有效期' }]}
              >
                <DatePicker style={{ width: '100%' }} placeholder="选择日期" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="危险品从业资格证号"
                name="qualification_cert_no"
              >
                <Input placeholder="请输入资格证号" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="资格证有效期"
                name="qualification_cert_expire"
              >
                <DatePicker style={{ width: '100%' }} placeholder="选择日期" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="驾龄(年)"
                name="driving_years"
                rules={[{ required: true, message: '请输入驾龄' }]}
              >
                <InputNumber min={0} max={50} style={{ width: '100%' }} placeholder="请输入" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="在岗状态"
                name="status"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="请选择">
                  {Object.entries(driverStatusMap).map(([k, v]) => (
                    <Option key={k} value={k}>{v.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="风险等级"
                name="risk_level"
                rules={[{ required: true, message: '请选择风险等级' }]}
              >
                <Select placeholder="请选择">
                  {Object.entries(riskLevelMap).map(([k, v]) => (
                    <Option key={k} value={k}>{v.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left" plain style={{ margin: '4px 0 8px' }}>车辆分配</Divider>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Item label="关联车辆" name="linked_vehicle_plate">
                <Select allowClear placeholder="请选择关联的车辆">
                  {Array.from({ length: 25 }, (_, i) => (
                    <Option key={i} value={`京A${(10000 + i).toString().padStart(5, '0')}`}>
                      京A{(10000 + i).toString().padStart(5, '0')}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  )
}

export default Drivers
