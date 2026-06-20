import React, { useEffect, useState } from 'react'
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
  Image,
  Progress,
  message,
  Radio,
  Empty,
  Divider,
  Alert,
  DatePicker,
} from 'antd'
import {
  AlertOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PhoneOutlined,
  VideoCameraOutlined,
  SafetyCertificateOutlined,
  EyeOutlined,
  FilterOutlined,
  ReloadOutlined,
  WarningOutlined,
  FireOutlined,
  UserOutlined,
  EnvironmentOutlined,
  ExportOutlined,
  BellOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import api from '@/services/api'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'
import { useAppStore, AlarmItem } from '@/store/app'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { RangePicker } = DatePicker

const levelMap = {
  1: { color: 'blue', label: '提醒', badge: 'processing' as const },
  2: { color: 'orange', label: '警告', badge: 'warning' as const },
  3: { color: 'red', label: '严重', badge: 'error' as const },
}

const statusMap = {
  pending: { color: 'red', label: '待处理' },
  processing: { color: 'orange', label: '处理中' },
  acknowledged: { color: 'blue', label: '已确认' },
  resolved: { color: 'green', label: '已关闭' },
  ignored: { color: 'default', label: '忽略' },
}

const alarmTypeMap: Record<string, { label: string; color: string }> = {
  fatigue_perclos: { label: '疲劳瞌睡', color: 'red' },
  continuous_fatigue_perclos: { label: '连续疲劳', color: 'red' },
  excessive_yawn: { label: '频繁打哈欠', color: 'orange' },
  abnormal_head_posture: { label: '异常姿态', color: 'gold' },
  gaze_distraction: { label: '视线偏离', color: 'gold' },
  phone_usage: { label: '使用手机', color: 'purple' },
  smoking: { label: '抽烟', color: 'volcano' },
  no_seatbelt: { label: '未系安全带', color: 'magenta' },
  continuous_fatigue: { label: '连续超时驾驶', color: 'red' },
}

const FatigueAlarms: React.FC = () => {
  const { alarms, addAlarm, updateAlarm } = useAppStore()
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string>()
  const [levelFilter, setLevelFilter] = useState<number>()
  const [typeFilter, setTypeFilter] = useState<string>()
  const [detailDrawer, setDetailDrawer] = useState<AlarmItem | null>(null)
  const [handleModal, setHandleModal] = useState<AlarmItem | null>(null)
  const [handleForm] = Form.useForm()
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    today: 0,
    severe: 0,
  })

  const fetchData = async () => {
    setLoading(true)
    try {
      const res: any = await api.get('/fatigue/alarms', {
        page,
        page_size: pageSize,
        status: statusFilter || '',
        level: levelFilter || 0,
        type: typeFilter || '',
      })
      setData(res?.list || [])
      setTotal(res?.total || 0)
      setStats(s => ({
        ...s,
        total: res?.total || 0,
        pending: res?.list?.filter((a: any) => a.status === 'pending').length || 0,
        severe: res?.list?.filter((a: any) => a.alarm_level === 3).length || 0,
        today: res?.list?.filter((a: any) => dayjs(a.created_at).isSame(dayjs(), 'day')).length || 0,
      }))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [page, pageSize, statusFilter, levelFilter, typeFilter])

  const trendChart = {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 24 }, (_, i) => `${i.toString().padStart(2, '0')}:00`),
    },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar',
      smooth: true,
      areaStyle: { color: 'rgba(255,77,79,0.1)' },
      itemStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: '#ff4d4f' },
            { offset: 1, color: '#ffccc7' },
          ],
        },
        borderRadius: [4, 4, 0, 0],
      },
      data: Array.from({ length: 24 }, () => Math.floor(Math.random() * 8) + 1),
    }],
  }

  const distributionChart = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, type: 'scroll' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}: {c}' },
      data: Object.entries(alarmTypeMap).map(([k, v], i) => ({
        name: v.label,
        value: Math.floor(Math.random() * 20) + 2,
        itemStyle: {
          color: ['#ff4d4f', '#fa8c16', '#faad14', '#a0d911', '#13c2c2', '#1677ff', '#722ed1', '#eb2f96'][i % 8],
        },
      })),
    }],
  }

  const handleAck = async (values: any) => {
    if (!handleModal) return
    try {
      await api.post(`/fatigue/alarms/${handleModal.id}/ack`, {
        handle_type: values.handle_type,
        handle_note: values.handle_note,
      })
      message.success('报警已处理')
      setHandleModal(null)
      handleForm.resetFields()
      fetchData()
    } catch (e) { }
  }

  const handleBatch = (type: string) => {
    message.success(`批量${type === 'ack' ? '确认' : '忽略'}了 ${data.filter(d => d.status === 'pending').length} 条报警`)
  }

  const columns = [
    {
      title: '报警编号',
      dataIndex: 'alarm_no',
      width: 180,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '车辆',
      dataIndex: 'vehicle_plate',
      width: 120,
      render: (v, r) => (
        <Space>
          <Tag color="blue">{v || r.vehicle_id}</Tag>
          <Text type="secondary" style={{ fontSize: 12 }}>{r.driver_name}</Text>
        </Space>
      ),
    },
    {
      title: '报警类型',
      dataIndex: 'alarm_type',
      width: 140,
      render: (v) => {
        const t = alarmTypeMap[v] || { label: v, color: 'default' }
        return <Tag color={t.color} style={{ fontSize: 12 }}>{t.label}</Tag>
      },
    },
    {
      title: '等级',
      dataIndex: 'alarm_level',
      width: 80,
      render: (v) => {
        const l = levelMap[v as keyof typeof levelMap] || levelMap[1]
        return <Tag color={l.color}>{l.label}</Tag>
      },
    },
    {
      title: '疲劳指数',
      dataIndex: 'fatigue_score',
      width: 120,
      render: (v) => (
        <Progress
          percent={100 - Math.round(v || 0)}
          size="small"
          showInfo={false}
          strokeColor={(100 - v) > 40 ? '#ff4d4f' : (100 - v) > 20 ? '#faad14' : '#52c41a'}
          format={() => <Text strong type={v < 70 ? 'danger' : v < 85 ? 'warning' : 'success'}>{Math.round(v)}</Text>}
        />
      ),
    },
    {
      title: '连续疲劳',
      dataIndex: 'continuous_fatigue_minutes',
      width: 100,
      render: (v) => v ? <Badge count={`${v}分钟`} style={{ backgroundColor: v >= 20 ? '#ff4d4f' : '#fa8c16' }} /> : '-',
    },
    {
      title: '位置',
      dataIndex: 'location_address',
      ellipsis: true,
      render: (v, r) => (
        <Tooltip title={v || `${r.latitude}, ${r.longitude}`}>
          <span>
            <EnvironmentOutlined style={{ color: '#1677ff' }} />
            {' '}{v || `${r.latitude?.toFixed(4)}, ${r.longitude?.toFixed(4)}`}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (v) => {
        const s = statusMap[v as keyof typeof statusMap] || statusMap.pending
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    {
      title: '报警时间',
      dataIndex: 'created_at',
      width: 170,
      render: (v) => <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(v)}</Text>,
    },
    {
      title: '操作',
      width: 140,
      fixed: 'right' as const,
      render: (_, record: AlarmItem) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailDrawer(record)}>详情</Button>
          {(record.status === 'pending' || record.status === 'processing') && (
            <Button type="link" size="small" type="primary" danger icon={<CheckCircleOutlined />} onClick={() => setHandleModal(record)}>处理</Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Alert
        type="warning"
        showIcon
        icon={<BellOutlined />}
        message={
          <Space>
            <Text strong>实时报警提醒</Text>
            <Badge count={stats.pending} showZero style={{ backgroundColor: '#ff4d4f' }}>
              <Tag color="red">待处理 {stats.pending} 条</Tag>
            </Badge>
            <Text type="secondary">今日新增 {stats.today} 条，严重级别 {stats.severe} 条</Text>
          </Space>
        }
        style={{ borderRadius: 12 }}
      />

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="累计报警"
              value={stats.total}
              valueStyle={{ color: '#fa8c16' }}
              prefix={<AlertOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="待处理"
              value={stats.pending}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<FireOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日新增"
              value={stats.today}
              valueStyle={{ color: '#722ed1' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="严重等级"
              value={stats.severe}
              valueStyle={{ color: '#f5222d' }}
              prefix={<SafetyCertificateOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><AlertOutlined style={{ color: '#ff4d4f' }} /> 24小时报警趋势</Space>}>
            <ReactECharts option={trendChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><PieChartOutlined /> 报警类型分布</Space>}>
            <ReactECharts option={distributionChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <AlertOutlined style={{ color: '#ff4d4f' }} />
            <Text strong style={{ fontSize: 15 }}>报警列表</Text>
            <Tag color="red">{total}</Tag>
          </Space>
        }
        extra={
          <Space wrap>
            <Select
              allowClear
              placeholder="状态筛选"
              style={{ width: 140 }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              {Object.entries(statusMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="等级筛选"
              style={{ width: 140 }}
              value={levelFilter}
              onChange={setLevelFilter}
            >
              {Object.entries(levelMap).map(([k, v]) => (
                <Option key={k} value={Number(k)}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Select
              allowClear
              placeholder="类型筛选"
              style={{ width: 160 }}
              value={typeFilter}
              onChange={setTypeFilter}
              showSearch
            >
              {Object.entries(alarmTypeMap).map(([k, v]) => (
                <Option key={k} value={k}><Tag color={v.color}>{v.label}</Tag></Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => { setStatusFilter(undefined); setLevelFilter(undefined); setTypeFilter(undefined) }}>重置</Button>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
            <Button
              type="primary"
              icon={<CheckCircleOutlined />}
              onClick={() => handleBatch('ack')}
            >
              批量确认
            </Button>
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          columns={columns as any}
          dataSource={data}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
          scroll={{ x: 1400 }}
          rowClassName={(r) => r.alarm_level === 3 ? '!bg-red-50' : r.alarm_level === 2 ? '!bg-orange-50' : ''}
        />
      </Card>

      <Drawer
        title={
          <Space>
            <AlertOutlined style={{ color: '#ff4d4f' }} />
            <Text strong>报警详情 - {detailDrawer?.alarm_no}</Text>
            {detailDrawer && (
              <Tag color={levelMap[detailDrawer.alarm_level as keyof typeof levelMap]?.color}>
                {levelMap[detailDrawer.alarm_level as keyof typeof levelMap]?.label}
              </Tag>
            )}
          </Space>
        }
        open={!!detailDrawer}
        onClose={() => setDetailDrawer(null)}
        width={520}
        extra={
          detailDrawer && (detailDrawer.status === 'pending' || detailDrawer.status === 'processing') && (
            <Space>
              <Button icon={<PhoneOutlined />} type="primary">语音对讲</Button>
              <Button type="primary" danger icon={<CheckCircleOutlined />} onClick={() => setHandleModal(detailDrawer)}>立即处理</Button>
            </Space>
          )
        }
      >
        {detailDrawer && (
          <div>
            <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }}>
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="报警编号">{detailDrawer.alarm_no}</Descriptions.Item>
                <Descriptions.Item label="车辆/驾驶员">
                  <Space>
                    <Tag color="blue">{detailDrawer.vehicle_plate || detailDrawer.vehicle_id}</Tag>
                    <UserOutlined /> {detailDrawer.driver_name || detailDrawer.driver_id}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="报警类型">
                  {(() => {
                    const t = alarmTypeMap[detailDrawer.alarm_type] || { label: detailDrawer.alarm_type, color: 'default' }
                    return <Tag color={t.color}>{t.label}</Tag>
                  })()}
                </Descriptions.Item>
                <Descriptions.Item label="疲劳指数">
                  <Progress
                    type="circle"
                    size={60}
                    percent={100 - Math.round(detailDrawer.fatigue_score)}
                    strokeColor={detailDrawer.fatigue_score < 70 ? '#ff4d4f' : detailDrawer.fatigue_score < 85 ? '#faad14' : '#52c41a'}
                  />
                </Descriptions.Item>
                <Descriptions.Item label="连续疲劳时间">
                  {detailDrawer.continuous_fatigue_minutes ? `${detailDrawer.continuous_fatigue_minutes} 分钟` : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="发生位置">
                  <Paragraph copyable style={{ margin: 0, fontSize: 13 }}>
                    <EnvironmentOutlined /> {detailDrawer.location_address || `${detailDrawer.latitude?.toFixed(6)}, ${detailDrawer.longitude?.toFixed(6)}`}
                  </Paragraph>
                </Descriptions.Item>
                <Descriptions.Item label="车速">{Math.round(detailDrawer.vehicle_speed)} km/h</Descriptions.Item>
                <Descriptions.Item label="报警时间">{formatDateTime(detailDrawer.created_at)}</Descriptions.Item>
                <Descriptions.Item label="当前状态">
                  <Tag color={statusMap[detailDrawer.status as keyof typeof statusMap]?.color}>
                    {statusMap[detailDrawer.status as keyof typeof statusMap]?.label}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" style={{ borderRadius: 8, marginBottom: 16 }} title={<Space><VideoCameraOutlined /> 报警现场快照</Space>}>
              <Row gutter={8}>
                <Col span={12}>
                  {detailDrawer.snap_image_url ? (
                    <Image src={detailDrawer.snap_image_url} />
                  ) : (
                    <div style={{
                      background: '#f5f5f5',
                      borderRadius: 6,
                      height: 120,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#8c8c8c',
                      fontSize: 12,
                    }}>
                      <EyeOutlined /> 无快照图片
                    </div>
                  )}
                </Col>
                <Col span={12}>
                  {detailDrawer.video_clip_url ? (
                    <video src={detailDrawer.video_clip_url} controls style={{ width: '100%', borderRadius: 6 }} />
                  ) : (
                    <div style={{
                      background: '#f5f5f5',
                      borderRadius: 6,
                      height: 120,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#8c8c8c',
                      fontSize: 12,
                    }}>
                      <VideoCameraOutlined /> 无视频片段(前10秒)
                    </div>
                  )}
                </Col>
              </Row>
            </Card>

            {detailDrawer.status !== 'pending' && (
              <Card size="small" style={{ borderRadius: 8 }} title={<Space><CheckCircleOutlined /> 处理记录</Space>}>
                <Descriptions column={1} size="small" bordered>
                  <Descriptions.Item label="处理方式">
                    {detailDrawer.handle_type ? <Tag color="blue">{detailDrawer.handle_type}</Tag> : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="处理人ID">{detailDrawer.dispatcher_id || '-'}</Descriptions.Item>
                  <Descriptions.Item label="处理时间">{formatDateTime(detailDrawer.handled_at as any) || '-'}</Descriptions.Item>
                  <Descriptions.Item label="处理备注">{detailDrawer.handle_note || '-'}</Descriptions.Item>
                </Descriptions>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      <Modal
        title={<Space><CheckCircleOutlined style={{ color: '#52c41a' }} /> 处理报警 - {handleModal?.alarm_no}</Space>}
        open={!!handleModal}
        onCancel={() => { setHandleModal(null); handleForm.resetFields() }}
        onOk={() => handleForm.submit()}
        okText="确认处理"
        okButtonProps={{ type: 'primary', danger: false }}
        width={520}
      >
        {handleModal && (
          <div style={{ marginBottom: 16 }}>
            <Alert
              type={handleModal.alarm_level === 3 ? 'error' : 'warning'}
              showIcon
              message={
                <Space>
                  <Tag color={levelMap[handleModal.alarm_level as keyof typeof levelMap]?.color}>
                    {levelMap[handleModal.alarm_level as keyof typeof levelMap]?.label}
                  </Tag>
                  {alarmTypeMap[handleModal.alarm_type]?.label || handleModal.alarm_type}
                  · 车辆 {handleModal.vehicle_plate} · 驾驶员 {handleModal.driver_name}
                </Space>
              }
              description={`疲劳指数 ${Math.round(handleModal.fatigue_score)}${handleModal.continuous_fatigue_minutes ? `，连续疲劳 ${handleModal.continuous_fatigue_minutes} 分钟` : ''}`}
              style={{ borderRadius: 8 }}
            />
          </div>
        )}
        <Form form={handleForm} layout="vertical" onFinish={handleAck}>
          <Form.Item label="处理方式" name="handle_type" rules={[{ required: true, message: '请选择处理方式' }]}>
            <Radio.Group>
              <Radio value="voice_remind">🎙️ 语音提醒</Radio>
              <Radio value="dispatch_rest">☕ 调度服务区停靠</Radio>
              <Radio value="intervene_legal">🚓 通知执法站</Radio>
              <Radio value="escalate">📢 升级上报主管</Radio>
              <Radio value="other">📝 其他</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item label="处理备注" name="handle_note" rules={[{ required: true, message: '请输入处理备注' }]}>
            <Input.TextArea rows={3} placeholder="请描述处理情况..." />
          </Form.Item>
          <Divider style={{ margin: '8px 0 16px' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>
            <WarningOutlined /> 处理后系统将自动通知车载终端语音播报，并记录于驾驶员考核档案
          </Text>
        </Form>
      </Modal>
    </div>
  )
}

export default FatigueAlarms

function PieChartOutlined() {
  return <span>📊</span>
}
