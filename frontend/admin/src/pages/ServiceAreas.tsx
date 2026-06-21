import React, { useEffect, useState, useCallback } from 'react'
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
  Form,
  Select,
  Input,
  Statistic,
  Badge,
  Drawer,
  Descriptions,
  Progress,
  message,
  Rate,
  Empty,
  Divider,
  InputNumber,
  Tabs,
  List,
  Avatar,
} from 'antd'
import {
  EnvironmentOutlined,
  CarOutlined,
  SafetyCertificateOutlined,
  StarOutlined,
  ClockCircleOutlined,
  EyeOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CoffeeOutlined,
  SecurityScanOutlined,
  ApartmentOutlined,
  RestOutlined,
  CheckSquareOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { serviceAreaApi } from '@/services/api'
import type {
  ServiceAreaItem,
  ServiceAreaRealtimeStatus,
  ServiceAreaReview,
  DrivingRestRecord,
  RestCountdownResponse,
} from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs

const statusMap: Record<number, { color: string; label: string }> = {
  1: { color: 'green', label: '正常' },
  0: { color: 'default', label: '停用' },
}

const securityLevelMap: Record<number, { color: string; label: string }> = {
  1: { color: 'green', label: 'A级' },
  2: { color: 'blue', label: 'B级' },
  3: { color: 'orange', label: 'C级' },
  4: { color: 'red', label: 'D级' },
}

const crowdLevelMap: Record<number, { color: string; label: string }> = {
  1: { color: 'green', label: '畅通' },
  2: { color: 'blue', label: '正常' },
  3: { color: 'orange', label: '较拥挤' },
  4: { color: 'red', label: '拥挤' },
}

const restStatusMap: Record<string, { color: string; label: string; icon: React.ReactNode }> = {
  driving: { color: 'blue', label: '驾驶中', icon: <CarOutlined /> },
  resting: { color: 'orange', label: '休息中', icon: <RestOutlined /> },
  completed: { color: 'green', label: '已完成', icon: <CheckSquareOutlined /> },
}

const ServiceAreas: React.FC = () => {
  const [activeTab, setActiveTab] = useState('list')
  const [loading, setLoading] = useState(false)
  const [statsLoading, setStatsLoading] = useState(false)

  const [serviceAreaList, setServiceAreaList] = useState<ServiceAreaItem[]>([])
  const [totalAreas, setTotalAreas] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [keyword, setKeyword] = useState('')
  const [dangerParkingFilter, setDangerParkingFilter] = useState<boolean>()

  const [detailDrawer, setDetailDrawer] = useState<ServiceAreaItem | null>(null)
  const [detailStatus, setDetailStatus] = useState<ServiceAreaRealtimeStatus | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  const [restRecords, setRestRecords] = useState<DrivingRestRecord[]>([])
  const [restTotal, setRestTotal] = useState(0)
  const [restPage, setRestPage] = useState(1)
  const [restPageSize, setRestPageSize] = useState(10)
  const [restLoading, setRestLoading] = useState(false)

  const [reviews, setReviews] = useState<ServiceAreaReview[]>([])
  const [reviewTotal, setReviewTotal] = useState(0)
  const [reviewPage, setReviewPage] = useState(1)
  const [reviewPageSize, setReviewPageSize] = useState(10)
  const [reviewLoading, setReviewLoading] = useState(false)

  const [stats, setStats] = useState<{
    total_service_areas: number
    danger_parking_areas: number
    average_rating: number
    today_check_ins: number
    today_reviews: number
  } | null>(null)

  const [statusUpdateModal, setStatusUpdateModal] = useState<ServiceAreaItem | null>(null)
  const [statusForm] = Form.useForm()

  const fetchStatistics = useCallback(async () => {
    setStatsLoading(true)
    try {
      const res = await serviceAreaApi.getStatistics()
      setStats(res)
    } catch (e) {
      // ignore
    } finally {
      setStatsLoading(false)
    }
  }, [])

  const fetchServiceAreas = useCallback(async () => {
    setLoading(true)
    try {
      const res = await serviceAreaApi.list({
        page,
        page_size: pageSize,
        keyword,
        has_danger_parking: dangerParkingFilter,
      })
      setServiceAreaList(res?.list || [])
      setTotalAreas(res?.total || 0)
    } catch (e) {
      message.error('获取服务区列表失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, keyword, dangerParkingFilter])

  const fetchRestRecords = useCallback(async () => {
    setRestLoading(true)
    try {
      const res = await serviceAreaApi.listRestRecords({
        page: restPage,
        page_size: restPageSize,
      })
      setRestRecords(res?.list || [])
      setRestTotal(res?.total || 0)
    } catch (e) {
      message.error('获取休息记录失败')
    } finally {
      setRestLoading(false)
    }
  }, [restPage, restPageSize])

  const fetchReviews = useCallback(async () => {
    setReviewLoading(true)
    try {
      const res = await serviceAreaApi.listReviews({
        page: reviewPage,
        page_size: reviewPageSize,
      })
      setReviews(res?.list || [])
      setReviewTotal(res?.total || 0)
    } catch (e) {
      message.error('获取评价列表失败')
    } finally {
      setReviewLoading(false)
    }
  }, [reviewPage, reviewPageSize])

  const fetchDetail = async (id: number) => {
    setDetailLoading(true)
    try {
      const res = await serviceAreaApi.get(id)
      setDetailStatus(res.real_status || null)
    } catch (e) {
      message.error('获取服务区详情失败')
    } finally {
      setDetailLoading(false)
    }
  }

  const handleViewDetail = (record: ServiceAreaItem) => {
    setDetailDrawer(record)
    fetchDetail(record.id)
  }

  const handleStatusUpdate = (record: ServiceAreaItem) => {
    setStatusUpdateModal(record)
    statusForm.setFieldsValue({
      service_area_id: record.id,
    })
  }

  const handleStatusSubmit = async () => {
    try {
      const values = await statusForm.validateFields()
      await serviceAreaApi.updateRealtimeStatus(values)
      message.success('状态更新成功')
      setStatusUpdateModal(null)
      statusForm.resetFields()
      if (detailDrawer) {
        fetchDetail(detailDrawer.id)
      }
      fetchStatistics()
    } catch (e) {
      // ignore
    }
  }

  useEffect(() => {
    fetchStatistics()
  }, [fetchStatistics])

  useEffect(() => {
    if (activeTab === 'list') {
      fetchServiceAreas()
    }
  }, [activeTab, fetchServiceAreas])

  useEffect(() => {
    if (activeTab === 'records') {
      fetchRestRecords()
    }
  }, [activeTab, fetchRestRecords])

  useEffect(() => {
    if (activeTab === 'reviews') {
      fetchReviews()
    }
  }, [activeTab, fetchReviews])

  const statsCards = stats && [
    {
      title: '服务区总数',
      value: stats.total_service_areas,
      icon: <ApartmentOutlined style={{ fontSize: 32, color: '#1677ff' }} />,
      color: 'blue',
    },
    {
      title: '危险品停靠区',
      value: stats.danger_parking_areas,
      icon: <SafetyCertificateOutlined style={{ fontSize: 32, color: '#fa8c16' }} />,
      color: 'orange',
    },
    {
      title: '平均评分',
      value: stats.average_rating.toFixed(1),
      icon: <StarOutlined style={{ fontSize: 32, color: '#fadb14' }} />,
      color: 'gold',
      suffix: '分',
    },
    {
      title: '今日签到数',
      value: stats.today_check_ins,
      icon: <CheckSquareOutlined style={{ fontSize: 32, color: '#52c41a' }} />,
      color: 'green',
    },
  ]

  const columns = [
    {
      title: '服务区名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: ServiceAreaItem) => (
        <Space>
          <EnvironmentOutlined style={{ color: '#1677ff' }} />
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: 12, color: '#8c8c8c' }}>
              {record.highway_name} · {record.direction}方向
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: '位置',
      dataIndex: 'province',
      key: 'location',
      render: (_: any, record: ServiceAreaItem) => (
        <Text>{record.province} {record.city}</Text>
      ),
    },
    {
      title: '危险品停靠',
      dataIndex: 'has_danger_goods_parking',
      key: 'has_danger_goods_parking',
      render: (val: boolean) => (
        <Tag color={val ? 'green' : 'default'} icon={val ? <CheckCircleOutlined /> : <WarningOutlined />}>
          {val ? '允许' : '不允许'}
        </Tag>
      ),
    },
    {
      title: '车位总数',
      dataIndex: 'parking_spaces',
      key: 'parking_spaces',
      render: (val: number, record: ServiceAreaItem) => (
        <div>
          <div>普通: {val}个</div>
          <div style={{ color: '#fa8c16' }}>危险品: {record.danger_parking_spaces}个</div>
        </div>
      ),
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      render: (val: number) => (
        <Space>
          <Rate disabled value={val} allowHalf style={{ fontSize: 12 }} />
          <Text>{val}</Text>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (val: number) => (
        <Tag color={statusMap[val]?.color || 'default'}>
          {statusMap[val]?.label || '未知'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: ServiceAreaItem) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
            详情
          </Button>
          <Button type="link" size="small" icon={<ReloadOutlined />} onClick={() => handleStatusUpdate(record)}>
            更新状态
          </Button>
        </Space>
      ),
    },
  ]

  const restColumns = [
    {
      title: '驾驶员',
      dataIndex: 'driver_id',
      key: 'driver',
      render: (_: number, record: DrivingRestRecord) => (
        <Space>
          <Avatar size="small" style={{ backgroundColor: '#1677ff' }}>
            {record.driver_id}
          </Avatar>
          <Text>驾驶员 #{record.driver_id}</Text>
        </Space>
      ),
    },
    {
      title: '车辆',
      dataIndex: 'vehicle_id',
      key: 'vehicle',
      render: (val: number) => <Text>车辆 #{val}</Text>,
    },
    {
      title: '服务区',
      dataIndex: 'rest_service_area_name',
      key: 'service_area',
      render: (val: string) => val || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (val: string) => {
        const info = restStatusMap[val] || restStatusMap.completed
        return (
          <Tag color={info.color} icon={info.icon}>
            {info.label}
          </Tag>
        )
      },
    },
    {
      title: '连续驾驶时长',
      dataIndex: 'continuous_drive_minutes',
      key: 'drive_minutes',
      render: (val: number) => (
        <Space>
          <ClockCircleOutlined style={{ color: val >= 240 ? '#ff4d4f' : '#52c41a' }} />
          <Text style={{ color: val >= 240 ? '#ff4d4f' : undefined }}>
            {Math.floor(val / 60)}小时{val % 60}分钟
          </Text>
        </Space>
      ),
    },
    {
      title: '休息时长',
      dataIndex: 'rest_duration_minutes',
      key: 'rest_minutes',
      render: (val: number, record: DrivingRestRecord) => {
        if (record.status === 'driving') return '-'
        const minRequired = record.min_rest_required || 20
        const isEnough = val >= minRequired
        return (
          <div>
            <Progress
              percent={Math.min(Math.round((val / minRequired) * 100), 100)}
              size="small"
              status={isEnough ? 'success' : 'active'}
              strokeColor={isEnough ? '#52c41a' : '#fa8c16'}
            />
            <div style={{ fontSize: 12, color: '#8c8c8c' }}>
              {val} / {minRequired} 分钟
            </div>
          </div>
        )
      },
    },
    {
      title: '是否超时',
      dataIndex: 'is_overtime',
      key: 'is_overtime',
      render: (val: boolean, record: DrivingRestRecord) => (
        <Tag color={val ? 'red' : 'green'} icon={val ? <ExclamationCircleOutlined /> : <CheckCircleOutlined />}>
          {val ? `超时${record.overtime_minutes}分钟` : '正常'}
        </Tag>
      ),
    },
    {
      title: '开始驾驶',
      dataIndex: 'drive_start_time',
      key: 'drive_start',
      render: (val: string) => formatDateTime(val),
    },
  ]

  const reviewColumns = [
    {
      title: '服务区',
      dataIndex: 'service_area_id',
      key: 'service_area',
      render: (val: number) => `服务区 #${val}`,
    },
    {
      title: '驾驶员',
      dataIndex: 'driver_name',
      key: 'driver',
      render: (val: string) => val || '-',
    },
    {
      title: '安全性评分',
      dataIndex: 'security_score',
      key: 'security_score',
      render: (val: number) => (
        <Space>
          <Rate disabled value={val} allowHalf style={{ fontSize: 12 }} />
          <Text strong style={{ color: val >= 4 ? '#52c41a' : val >= 3 ? '#fa8c16' : '#ff4d4f' }}>
            {val}分
          </Text>
        </Space>
      ),
    },
    {
      title: '综合评分',
      dataIndex: 'overall_score',
      key: 'overall_score',
      render: (val: number) => (
        <Badge count={`${val}分`} style={{ backgroundColor: val >= 4 ? '#52c41a' : '#fa8c16' }} />
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags_array',
      key: 'tags',
      render: (tags: string[]) => (
        <Space size={4} wrap>
          {tags?.map((tag, idx) => (
            <Tag key={idx} color="blue" style={{ margin: 2 }}>
              {tag}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '评价内容',
      dataIndex: 'comment_text',
      key: 'comment',
      ellipsis: true,
      render: (val: string) => val || '无评价内容',
    },
    {
      title: '评价时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (val: string) => formatDateTime(val),
    },
  ]

  const securityChartOption = {
    tooltip: {
      trigger: 'axis',
    },
    legend: {
      data: ['安保等级', '餐饮评分'],
    },
    xAxis: {
      type: 'category',
      data: serviceAreaList.slice(0, 6).map(s => s.name?.slice(0, 6) || ''),
    },
    yAxis: [
      {
        type: 'value',
        name: '安保等级',
        min: 0,
        max: 5,
      },
      {
        type: 'value',
        name: '评分',
        min: 0,
        max: 5,
      },
    ],
    series: [
      {
        name: '安保等级',
        type: 'bar',
        data: [4, 4, 3, 5, 4, 3],
        itemStyle: { color: '#1677ff' },
      },
      {
        name: '餐饮评分',
        type: 'line',
        yAxisIndex: 1,
        data: [4.2, 3.8, 4.5, 4.0, 3.5, 4.3],
        itemStyle: { color: '#fa8c16' },
        smooth: true,
      },
    ],
  }

  return (
    <div style={{ padding: 8 }}>
      <Row gutter={[16, 16]}>
        {statsCards?.map((stat, idx) => (
          <Col xs={12} sm={12} md={6} key={idx}>
            <Card loading={statsLoading}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div>
                  <Statistic
                    title={stat.title}
                    value={stat.value}
                    suffix={stat.suffix}
                    valueStyle={{ color: stat.color }}
                  />
                </div>
                {stat.icon}
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Card style={{ marginTop: 16 }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="服务区列表" key="list">
            <Space style={{ marginBottom: 16 }} wrap>
              <Input.Search
                placeholder="搜索服务区名称/高速"
                allowClear
                style={{ width: 240 }}
                value={keyword}
                onChange={e => setKeyword(e.target.value)}
                onSearch={() => {
                  setPage(1)
                  fetchServiceAreas()
                }}
              />
              <Select
                placeholder="危险品停靠"
                style={{ width: 160 }}
                allowClear
                value={dangerParkingFilter}
                onChange={val => {
                  setDangerParkingFilter(val)
                  setPage(1)
                }}
              >
                <Option value={true}>允许危险品停靠</Option>
                <Option value={false}>不允许危险品停靠</Option>
              </Select>
              <Button icon={<ReloadOutlined />} onClick={fetchServiceAreas}>
                刷新
              </Button>
            </Space>

            <Table
              columns={columns}
              dataSource={serviceAreaList}
              rowKey="id"
              loading={loading}
              pagination={{
                current: page,
                pageSize,
                total: totalAreas,
                onChange: (p, ps) => {
                  setPage(p)
                  setPageSize(ps)
                },
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (t) => `共 ${t} 个服务区`,
              }}
            />
          </TabPane>

          <TabPane tab="休息记录" key="records">
            <Space style={{ marginBottom: 16 }}>
              <Button icon={<ReloadOutlined />} onClick={fetchRestRecords}>
                刷新
              </Button>
            </Space>

            <Table
              columns={restColumns}
              dataSource={restRecords}
              rowKey="id"
              loading={restLoading}
              pagination={{
                current: restPage,
                pageSize: restPageSize,
                total: restTotal,
                onChange: (p, ps) => {
                  setRestPage(p)
                  setRestPageSize(ps)
                },
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (t) => `共 ${t} 条记录`,
              }}
            />
          </TabPane>

          <TabPane tab="评价管理" key="reviews">
            <Space style={{ marginBottom: 16 }}>
              <Button icon={<ReloadOutlined />} onClick={fetchReviews}>
                刷新
              </Button>
            </Space>

            <Table
              columns={reviewColumns}
              dataSource={reviews}
              rowKey="id"
              loading={reviewLoading}
              pagination={{
                current: reviewPage,
                pageSize: reviewPageSize,
                total: reviewTotal,
                onChange: (p, ps) => {
                  setReviewPage(p)
                  setReviewPageSize(ps)
                },
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (t) => `共 ${t} 条评价`,
              }}
            />
          </TabPane>

          <TabPane tab="数据分析" key="analysis">
            <Row gutter={[16, 16]}>
              <Col xs={24} md={12}>
                <Card title="安保等级与餐饮评分对比">
                  <ReactECharts option={securityChartOption} style={{ height: 300 }} />
                </Card>
              </Col>
              <Col xs={24} md={12}>
                <Card title="今日休息情况">
                  <List
                    size="small"
                    dataSource={restRecords.slice(0, 5)}
                    locale={{ emptyText: <Empty description="暂无休息记录" /> }}
                    renderItem={(item) => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<Avatar icon={<RestOutlined />} />}
                          title={`驾驶员 #${item.driver_id}`}
                          description={
                            <Space>
                              <Text type="secondary">{item.rest_service_area_name || '未知服务区'}</Text>
                              <Tag color={restStatusMap[item.status]?.color}>
                                {restStatusMap[item.status]?.label}
                              </Tag>
                            </Space>
                          }
                        />
                        <Text strong>{item.rest_duration_minutes}分钟</Text>
                      </List.Item>
                    )}
                  />
                </Card>
              </Col>
            </Row>
          </TabPane>
        </Tabs>
      </Card>

      <Drawer
        title="服务区详情"
        placement="right"
        width={500}
        open={!!detailDrawer}
        onClose={() => setDetailDrawer(null)}
        loading={detailLoading}
      >
        {detailDrawer && (
          <>
            <Descriptions title="基本信息" column={1} bordered size="small">
              <Descriptions.Item label="服务区名称">{detailDrawer.name}</Descriptions.Item>
              <Descriptions.Item label="所属高速">
                {detailDrawer.highway_name}（{detailDrawer.direction}方向）
              </Descriptions.Item>
              <Descriptions.Item label="位置">
                {detailDrawer.province} {detailDrawer.city}
              </Descriptions.Item>
              <Descriptions.Item label="坐标">
                {detailDrawer.latitude}, {detailDrawer.longitude}
              </Descriptions.Item>
              <Descriptions.Item label="联系电话">{detailDrawer.phone || '-'}</Descriptions.Item>
              <Descriptions.Item label="综合评分">
                <Rate disabled value={detailDrawer.rating} allowHalf style={{ fontSize: 14 }} />
              </Descriptions.Item>
              <Descriptions.Item label="服务设施">
                <Space wrap>
                  {detailDrawer.has_restaurant && <Tag color="green">餐饮</Tag>}
                  {detailDrawer.has_hotel && <Tag color="blue">住宿</Tag>}
                  {detailDrawer.has_fuel_station && <Tag color="orange">加油站</Tag>}
                  {detailDrawer.has_charging && <Tag color="cyan">充电桩</Tag>}
                  {detailDrawer.has_maintenance && <Tag color="purple">维修</Tag>}
                  {detailDrawer.has_danger_goods_parking && (
                    <Tag color="red">危险品停靠</Tag>
                  )}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="车位信息">
                <div>普通车位：{detailDrawer.parking_spaces} 个</div>
                <div style={{ color: '#fa8c16' }}>
                  危险品车位：{detailDrawer.danger_parking_spaces} 个
                </div>
              </Descriptions.Item>
            </Descriptions>

            <Divider />

            <Title level={5} style={{ marginTop: 16 }}>
              <SecurityScanOutlined /> 实时状态
            </Title>
            {detailStatus ? (
              <div>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Card size="small" title="车位余量">
                      <Statistic
                        title="危险品可用"
                        value={detailStatus.available_danger_spaces}
                        suffix={`/ ${detailStatus.total_danger_spaces}`}
                        valueStyle={{
                          color: detailStatus.available_danger_spaces > 0 ? '#52c41a' : '#ff4d4f',
                          fontSize: 20,
                        }}
                      />
                      <Progress
                        percent={
                          detailStatus.total_danger_spaces
                            ? Math.round(
                                ((detailStatus.total_danger_spaces - detailStatus.available_danger_spaces) /
                                  detailStatus.total_danger_spaces) *
                                  100
                              )
                            : 0
                        }
                        size="small"
                        status="normal"
                      />
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card size="small" title="安保等级">
                      <Statistic
                        value={securityLevelMap[detailStatus.security_level]?.label || '-'}
                        valueStyle={{
                          color:
                            detailStatus.security_level <= 2
                              ? '#52c41a'
                              : detailStatus.security_level === 3
                              ? '#fa8c16'
                              : '#ff4d4f',
                          fontSize: 20,
                        }}
                      />
                      <Text type="secondary">
                        巡逻间隔：{detailStatus.security_patrol_interval}分钟
                      </Text>
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card size="small" title="餐饮">
                      {detailStatus.has_restaurant ? (
                        <>
                          <Space>
                            <Rate disabled value={detailStatus.restaurant_rating} allowHalf style={{ fontSize: 12 }} />
                            <Text strong>{detailStatus.restaurant_rating}分</Text>
                          </Space>
                          <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 4 }}>
                            预计等待：{detailStatus.restaurant_wait_minutes}分钟
                          </div>
                        </>
                      ) : (
                        <Text type="secondary">暂无餐饮服务</Text>
                      )}
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card size="small" title="拥挤程度">
                      <Tag color={crowdLevelMap[detailStatus.crowd_level]?.color || 'default'}>
                        {crowdLevelMap[detailStatus.crowd_level]?.label || '-'}
                      </Tag>
                      <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 4 }}>
                        天气：{detailStatus.weather_condition || '-'}
                      </div>
                    </Card>
                  </Col>
                </Row>

                <Divider orientation="left">加油充电</Divider>
                <Row gutter={[16, 8]}>
                  <Col span={8}>
                    <Text type="secondary">92#</Text>
                    <div style={{ fontSize: 16, fontWeight: 500, color: '#fa8c16' }}>
                      ¥{detailStatus.fuel_price_92?.toFixed(2) || '-'}
                    </div>
                  </Col>
                  <Col span={8}>
                    <Text type="secondary">95#</Text>
                    <div style={{ fontSize: 16, fontWeight: 500, color: '#fa8c16' }}>
                      ¥{detailStatus.fuel_price_95?.toFixed(2) || '-'}
                    </div>
                  </Col>
                  <Col span={8}>
                    <Text type="secondary">柴油</Text>
                    <div style={{ fontSize: 16, fontWeight: 500, color: '#fa8c16' }}>
                      ¥{detailStatus.fuel_price_diesel?.toFixed(2) || '-'}
                    </div>
                  </Col>
                </Row>
                {detailStatus.has_charging && (
                  <Row style={{ marginTop: 8 }}>
                    <Col span={24}>
                      <Text type="secondary">充电桩：</Text>
                      <Text strong>
                        {detailStatus.charging_piles_available} / {detailStatus.charging_piles_total} 个可用
                      </Text>
                    </Col>
                  </Row>
                )}

                <div style={{ marginTop: 16, fontSize: 12, color: '#8c8c8c', textAlign: 'right' }}>
                  数据更新时间：{formatDateTime(detailStatus.update_time)}
                </div>
              </div>
            ) : (
              <Empty description="暂无实时数据" />
            )}

            <Button type="primary" block style={{ marginTop: 16 }} onClick={() => handleStatusUpdate(detailDrawer)}>
              <ReloadOutlined /> 更新实时状态
            </Button>
          </>
        )}
      </Drawer>

      <Modal
        title="更新服务区状态"
        open={!!statusUpdateModal}
        onOk={handleStatusSubmit}
        onCancel={() => {
          setStatusUpdateModal(null)
          statusForm.resetFields()
        }}
        width={480}
      >
        <Form form={statusForm} layout="vertical">
          <Form.Item name="service_area_id" hidden>
            <Input />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="可用普通车位" name="available_parking_spaces">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="可用危险品车位" name="available_danger_spaces">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="安保等级" name="security_level">
                <Select>
                  <Option value={1}>A级 - 优秀</Option>
                  <Option value={2}>B级 - 良好</Option>
                  <Option value={3}>C级 - 一般</Option>
                  <Option value={4}>D级 - 较差</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="餐饮评分" name="restaurant_rating">
                <InputNumber style={{ width: '100%' }} min={0} max={5} step={0.5} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="拥挤程度" name="crowd_level">
                <Select>
                  <Option value={1}>畅通</Option>
                  <Option value={2}>正常</Option>
                  <Option value={3}>较拥挤</Option>
                  <Option value={4}>拥挤</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="天气状况" name="weather_condition">
                <Input placeholder="如：晴、多云、雨" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  )
}

export default ServiceAreas
