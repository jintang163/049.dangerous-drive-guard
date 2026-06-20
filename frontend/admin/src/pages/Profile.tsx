import React, { useState } from 'react'
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
  Input,
  Tabs,
  Avatar,
  Descriptions,
  message,
  Switch,
  Upload,
  Divider,
  Tooltip,
  Badge,
  Progress,
  Select,
  DatePicker,
  Radio,
  Statistic,
} from 'antd'
import {
  UserOutlined,
  SettingOutlined,
  SafetyOutlined,
  LockOutlined,
  PhoneOutlined,
  MailOutlined,
  EditOutlined,
  ReloadOutlined,
  KeyOutlined,
  LogoutOutlined,
  CalendarOutlined,
  TeamOutlined,
  IdcardOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  BellOutlined,
  FileTextOutlined,
  EyeOutlined,
  InboxOutlined,
  SecurityScanOutlined,
  InfoCircleOutlined,
  DashboardOutlined,
  EnvironmentOutlined,
  ExportOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select
const { TabPane } = Tabs

interface UserProfile {
  id: string
  name: string
  role: string
  organization: string
  employee_no: string
  join_date: string
  phone: string
  email: string
  gender: 'male' | 'female'
  avatar?: string
  position: string
  department: string
  address?: string
  id_card?: string
  emergency_contact?: string
  emergency_phone?: string
  education?: string
  license_no?: string
  license_type?: string
}

interface MyTask {
  id: string
  task_no: string
  title: string
  type: 'dispatch' | 'review' | 'training' | 'inspection'
  priority: 'high' | 'medium' | 'low'
  status: 'pending' | 'processing' | 'completed' | 'cancelled'
  assigned_at: string
  deadline: string
  progress: number
  description: string
}

interface OperationLog {
  id: string
  module: string
  action: string
  target: string
  result: 'success' | 'failed' | 'warning'
  operator_ip: string
  operate_time: string
  detail: string
}

const mockUser: UserProfile = {
  id: 'U20230101001',
  name: '张明',
  role: '高级调度员',
  organization: '华东运输分公司',
  employee_no: 'EMP00128',
  join_date: '2019-06-15',
  phone: '138****8888',
  email: 'zhangming@ddg.com',
  gender: 'male',
  position: '调度中心主管',
  department: '运营调度部',
  address: '上海市浦东新区张江高科技园区',
  id_card: '310***********1234',
  emergency_contact: '张太太',
  emergency_phone: '139****6666',
  education: '本科 - 物流管理',
  license_no: '沪交运管字0012345',
  license_type: 'A1A2D',
}

const mockTasks: MyTask[] = [
  {
    id: '1',
    task_no: 'TK20260620001',
    title: '深圳暴雨天气应急调度',
    type: 'dispatch',
    priority: 'high',
    status: 'processing',
    assigned_at: '2026-06-20 08:35:00',
    deadline: '2026-06-20 18:00:00',
    progress: 60,
    description: '对南山区受暴雨影响的126台危化品车辆进行路径重规划',
  },
  {
    id: '2',
    task_no: 'TK20260620002',
    title: '5月份驾驶员评分复核',
    type: 'review',
    priority: 'medium',
    status: 'pending',
    assigned_at: '2026-06-20 09:00:00',
    deadline: '2026-06-22 18:00:00',
    progress: 0,
    description: '复核系统自动生成的238名驾驶员月度安全评分',
  },
  {
    id: '3',
    task_no: 'TK20260619005',
    title: '新入职驾驶员安全培训',
    type: 'training',
    priority: 'medium',
    status: 'processing',
    assigned_at: '2026-06-19 14:00:00',
    deadline: '2026-06-25 18:00:00',
    progress: 45,
    description: '18名新入职驾驶员危化品运输安全法规培训',
  },
  {
    id: '4',
    task_no: 'TK20260619003',
    title: '车载设备周度巡检',
    type: 'inspection',
    priority: 'low',
    status: 'completed',
    assigned_at: '2026-06-19 10:00:00',
    deadline: '2026-06-21 18:00:00',
    progress: 100,
    description: '对86台车辆的ADAS摄像头、DMS系统进行功能性巡检',
  },
  {
    id: '5',
    task_no: 'TK20260618012',
    title: '青岛大风预警响应',
    type: 'dispatch',
    priority: 'high',
    status: 'completed',
    assigned_at: '2026-06-18 06:20:00',
    deadline: '2026-06-18 20:00:00',
    progress: 100,
    description: '对胶州湾大桥沿线的78台车辆进行绕行调度',
  },
  {
    id: '6',
    task_no: 'TK20260618008',
    title: '报警处理工单审核',
    type: 'review',
    priority: 'low',
    status: 'cancelled',
    assigned_at: '2026-06-18 11:30:00',
    deadline: '2026-06-20 18:00:00',
    progress: 20,
    description: '审核上周的152条人工处理报警记录',
  },
]

const mockLogs: OperationLog[] = [
  {
    id: '1',
    module: '调度管理',
    action: '发布调度指令',
    target: 'WB2026062000887',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 10:23:45',
    detail: '对深圳-上海的运单发布暴雨天气绕行指令',
  },
  {
    id: '2',
    module: '用户中心',
    action: '修改个人资料',
    target: '个人设置',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 10:12:08',
    detail: '更新了邮箱地址和紧急联系人信息',
  },
  {
    id: '3',
    module: '报警管理',
    action: '处理疲劳报警',
    target: 'ALM2026062000156',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 09:45:22',
    detail: '对沪A·88888车辆的连续疲劳报警下达语音提醒',
  },
  {
    id: '4',
    module: '车辆管理',
    action: '导出车辆清单',
    target: '全部车辆(1258)',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 09:18:33',
    detail: '导出Excel格式的华东分公司全部车辆台账',
  },
  {
    id: '5',
    module: '区块链',
    action: '存证核验',
    target: 'EVI202606190097',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 08:55:10',
    detail: '对事件存证记录执行SHA256核验，结果一致',
  },
  {
    id: '6',
    module: '用户中心',
    action: '登录系统',
    target: '管理端控制台',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-20 08:30:00',
    detail: '账号EMP00128成功登录管理后台',
  },
  {
    id: '7',
    module: '路径规划',
    action: '生成路线',
    target: 'RP202606200023',
    result: 'warning',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-19 17:42:18',
    detail: '生成的路线经过两处天气预警区域，已标记提醒',
  },
  {
    id: '8',
    module: '报警管理',
    action: '批量处理',
    target: '待处理报警(42)',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-19 16:28:55',
    detail: '批量确认42条提醒级别报警记录',
  },
  {
    id: '9',
    module: '用户中心',
    action: '修改密码',
    target: '账号安全',
    result: 'success',
    operator_ip: '10.18.32.105',
    operate_time: '2026-06-19 14:05:22',
    detail: '定期密码更换，符合安全策略要求',
  },
  {
    id: '10',
    module: '用户中心',
    action: '登录失败',
    target: '管理端控制台',
    result: 'failed',
    operator_ip: '172.16.5.88',
    operate_time: '2026-06-19 08:22:10',
    detail: '密码错误(第2/5次)，如非本人操作请立即联系管理员',
  },
]

const taskTypeMap: Record<string, { label: string; color: string }> = {
  dispatch: { label: '调度任务', color: 'blue' },
  review: { label: '审核任务', color: 'purple' },
  training: { label: '培训任务', color: 'cyan' },
  inspection: { label: '巡检任务', color: 'green' },
}

const taskPriorityMap: Record<string, { label: string; color: string }> = {
  high: { label: '高优先级', color: 'red' },
  medium: { label: '中优先级', color: 'orange' },
  low: { label: '低优先级', color: 'green' },
}

const taskStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待开始', color: 'default' },
  processing: { label: '进行中', color: 'blue' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'default' },
}

const logResultMap: Record<string, { label: string; color: string }> = {
  success: { label: '成功', color: 'green' },
  failed: { label: '失败', color: 'red' },
  warning: { label: '告警', color: 'orange' },
}

const Profile: React.FC = () => {
  const [user, setUser] = useState<UserProfile>(mockUser)
  const [passwordModal, setPasswordModal] = useState(false)
  const [passwordForm] = Form.useForm()
  const [profileForm] = Form.useForm()
  const [editing, setEditing] = useState(false)
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false)
  const [phoneBound, setPhoneBound] = useState(true)
  const [emailBound, setEmailBound] = useState(true)
  const [taskPage, setTaskPage] = useState(1)
  const [taskPageSize, setTaskPageSize] = useState(10)
  const [logPage, setLogPage] = useState(1)
  const [logPageSize, setLogPageSize] = useState(10)
  const [logModuleFilter, setLogModuleFilter] = useState<string>()
  const [logResultFilter, setLogResultFilter] = useState<string>()

  const handlePasswordChange = async (values: any) => {
    if (values.new_password !== values.confirm_password) {
      message.error('两次输入的新密码不一致')
      return
    }
    await new Promise(r => setTimeout(r, 800))
    message.success('密码修改成功，请使用新密码重新登录')
    setPasswordModal(false)
    passwordForm.resetFields()
  }

  const handleProfileSave = async () => {
    try {
      const values = await profileForm.validateFields()
      await new Promise(r => setTimeout(r, 600))
      setUser(prev => ({ ...prev, ...values }))
      setEditing(false)
      message.success('个人资料已更新')
    } catch {
      message.warning('请检查表单填写')
    }
  }

  const drivingScoreChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['安全评分', '行业均值'], bottom: 0 },
    grid: { left: 40, right: 20, top: 30, bottom: 50 },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 12 }, (_, i) => `${i + 1}月`),
      boundaryGap: false,
    },
    yAxis: {
      type: 'value',
      min: 60,
      max: 100,
      splitLine: { lineStyle: { type: 'dashed' } },
    },
    series: [
      {
        name: '安全评分',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        data: [78, 82, 85, 83, 88, 90, 89, 92, 91, 93, 94, 95],
        lineStyle: { color: '#1677ff', width: 3 },
        itemStyle: { color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22,119,255,0.3)' },
              { offset: 1, color: 'rgba(22,119,255,0.02)' },
            ],
          },
        },
        markLine: {
          silent: true,
          data: [{ type: 'average', name: 'Avg', label: { formatter: '年平均: {c}' } }],
          lineStyle: { color: '#fa8c16' },
        },
      },
      {
        name: '行业均值',
        type: 'line',
        smooth: true,
        symbol: 'diamond',
        symbolSize: 6,
        data: [72, 74, 73, 75, 76, 77, 76, 78, 79, 80, 81, 82],
        lineStyle: { color: '#8c8c8c', width: 2, type: 'dashed' },
        itemStyle: { color: '#8c8c8c' },
      },
    ],
  }

  const scoreDetailChart = {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { bottom: 0 },
    grid: { left: 40, right: 20, top: 20, bottom: 40 },
    xAxis: {
      type: 'category',
      data: ['安全驾驶', '合规操作', '车辆维护', '应急处理', '文明行车', '培训学习'],
      axisLabel: { interval: 0, rotate: 15 },
    },
    yAxis: { type: 'value', max: 100 },
    series: [
      {
        name: '本月得分',
        type: 'bar',
        barWidth: '28%',
        data: [96, 94, 92, 90, 98, 95],
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: '#1677ff' },
              { offset: 1, color: '#69c0ff' },
            ],
          },
          borderRadius: [6, 6, 0, 0],
        },
      },
      {
        name: '上月得分',
        type: 'bar',
        barWidth: '28%',
        data: [92, 90, 88, 85, 95, 92],
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: '#52c41a' },
              { offset: 1, color: '#b7eb8f' },
            ],
          },
          borderRadius: [6, 6, 0, 0],
        },
      },
    ],
  }

  const taskColumns = [
    {
      title: '任务编号',
      dataIndex: 'task_no',
      width: 150,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '任务标题',
      dataIndex: 'title',
      width: 220,
      ellipsis: true,
      render: (v: string, r: MyTask) => (
        <Tooltip title={r.description}>
          <Text strong>{v}</Text>
        </Tooltip>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 100,
      render: (v: string) => {
        const t = taskTypeMap[v] || { label: v, color: 'default' }
        return <Tag color={t.color}>{t.label}</Tag>
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      render: (v: string) => {
        const p = taskPriorityMap[v] || { label: v, color: 'default' }
        return <Tag color={p.color}>{p.label}</Tag>
      },
    },
    {
      title: '进度',
      dataIndex: 'progress',
      width: 180,
      render: (v: number) => (
        <Progress
          percent={v}
          size="small"
          strokeColor={v === 100 ? '#52c41a' : v >= 60 ? '#1677ff' : v >= 30 ? '#faad14' : '#ff4d4f'}
        />
      ),
    },
    {
      title: '截止日期',
      dataIndex: 'deadline',
      width: 160,
      render: (v: string, r: MyTask) => {
        const overdue = dayjs(v).isBefore(dayjs()) && r.status !== 'completed' && r.status !== 'cancelled'
        return (
          <Space>
            <CalendarOutlined style={{ color: overdue ? '#ff4d4f' : '#8c8c8c' }} />
            <Text type={overdue ? 'danger' : 'secondary'} style={{ fontSize: 12 }}>
              {v}{overdue && ' (已超期)'}
            </Text>
          </Space>
        )
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (v: string) => {
        const s = taskStatusMap[v] || { label: v, color: 'default' }
        return <Tag color={s.color}>{s.label}</Tag>
      },
    },
    {
      title: '操作',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, r: MyTask) => (
        <Button type="link" size="small" icon={<EyeOutlined />}>查看</Button>
      ),
    },
  ]

  const logColumns = [
    {
      title: '操作时间',
      dataIndex: 'operate_time',
      width: 170,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '功能模块',
      dataIndex: 'module',
      width: 110,
      render: (v: string) => <Tag color="blue">{v}</Tag>,
    },
    {
      title: '操作类型',
      dataIndex: 'action',
      width: 130,
      render: (v: string) => <Text>{v}</Text>,
    },
    {
      title: '操作对象',
      dataIndex: 'target',
      width: 180,
      ellipsis: true,
      render: (v: string) => <Text code style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '结果',
      dataIndex: 'result',
      width: 90,
      render: (v: string) => {
        const r = logResultMap[v] || { label: v, color: 'default' }
        return <Tag color={r.color}>{r.label}</Tag>
      },
    },
    {
      title: 'IP地址',
      dataIndex: 'operator_ip',
      width: 130,
      render: (v: string) => <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '详细说明',
      dataIndex: 'detail',
      ellipsis: true,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
  ]

  const currentScore = 95
  const scoreRank = 'Top 5%'
  const totalTasks = mockTasks.length
  const completedTasks = mockTasks.filter(t => t.status === 'completed').length

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} lg={8} xl={7}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Card
              bordered={false}
              style={{ borderRadius: 12 }}
              bodyStyle={{ padding: 24, textAlign: 'center' }}
            >
              <div style={{ position: 'relative', display: 'inline-block', marginBottom: 16 }}>
                <Avatar
                  size={96}
                  style={{ background: 'linear-gradient(135deg, #1677ff, #722ed1)', fontSize: 36 }}
                  icon={<UserOutlined />}
                />
                <Badge
                  count={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
                  style={{ right: 0, bottom: 0 }}
                  offset={[-4, -4]}
                />
              </div>
              <Title level={4} style={{ margin: '0 0 4px' }}>{user.name}</Title>
              <Space size={4} wrap style={{ justifyContent: 'center', marginBottom: 16 }}>
                <Tag color="blue" style={{ margin: 0 }}>{user.role}</Tag>
                <Tag color="purple" style={{ margin: 0 }}>{user.position}</Tag>
              </Space>
              <Divider style={{ margin: '12px 0 16px' }} />
              <Descriptions column={1} size="small" labelStyle={{ width: 90, color: '#8c8c8c' }}>
                <Descriptions.Item label={<Space><TeamOutlined /> 所属组织</Space>}>
                  {user.organization}
                </Descriptions.Item>
                <Descriptions.Item label={<Space><DashboardOutlined /> 所属部门</Space>}>
                  {user.department}
                </Descriptions.Item>
                <Descriptions.Item label={<Space><IdcardOutlined /> 工号</Space>}>
                  <Text copyable style={{ fontFamily: 'monospace' }}>{user.employee_no}</Text>
                </Descriptions.Item>
                <Descriptions.Item label={<Space><CalendarOutlined /> 入职时间</Space>}>
                  {user.join_date} <Text type="secondary">({dayjs(user.join_date).fromNow()})</Text>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card
              bordered={false}
              style={{ borderRadius: 12 }}
              title={<Space><SafetyOutlined style={{ color: '#52c41a' }} /> 安全设置</Space>}
              extra={
                <Tooltip title="安全等级">
                  <Tag color="green"><LockOutlined /> 高</Tag>
                </Tooltip>
              }
            >
              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Space>
                    <LockOutlined style={{ color: '#1677ff', fontSize: 16 }} />
                    <div>
                      <div><Text strong>登录密码</Text></div>
                      <Text type="secondary" style={{ fontSize: 12 }}>上次修改: {dayjs().subtract(1, 'day').format('YYYY-MM-DD')}</Text>
                    </div>
                  </Space>
                  <Button type="link" size="small" icon={<KeyOutlined />} onClick={() => setPasswordModal(true)}>
                    修改密码
                  </Button>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Space>
                    <PhoneOutlined style={{ color: '#52c41a', fontSize: 16 }} />
                    <div>
                      <div><Text strong>绑定手机</Text> {phoneBound && <Tag color="green">已验证</Tag>}</div>
                      <Text type="secondary" style={{ fontSize: 12 }}>{user.phone}</Text>
                    </div>
                  </Space>
                  <Space>
                    {phoneBound ? (
                      <>
                        <Button type="link" size="small">更换</Button>
                        <Button type="link" size="small" danger>解绑</Button>
                      </>
                    ) : (
                      <Button type="link" size="small" type="primary">立即绑定</Button>
                    )}
                  </Space>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Space>
                    <MailOutlined style={{ color: '#faad14', fontSize: 16 }} />
                    <div>
                      <div><Text strong>绑定邮箱</Text> {emailBound && <Tag color="green">已验证</Tag>}</div>
                      <Text type="secondary" style={{ fontSize: 12 }}>{user.email}</Text>
                    </div>
                  </Space>
                  <Space>
                    {emailBound ? (
                      <>
                        <Button type="link" size="small">更换</Button>
                        <Button type="link" size="small" danger>解绑</Button>
                      </>
                    ) : (
                      <Button type="link" size="small" type="primary">立即绑定</Button>
                    )}
                  </Space>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0' }}>
                  <Space>
                    <SecurityScanOutlined style={{ color: '#722ed1', fontSize: 16 }} />
                    <div>
                      <div>
                        <Text strong>两步验证</Text>
                        {twoFactorEnabled ? <Tag color="purple">已开启</Tag> : <Tag>未开启</Tag>}
                      </div>
                      <Text type="secondary" style={{ fontSize: 12 }}>登录时需短信/邮箱验证码</Text>
                    </div>
                  </Space>
                  <Switch
                    checked={twoFactorEnabled}
                    onChange={(v) => {
                      if (v) {
                        Modal.confirm({
                          title: '开启两步验证',
                          icon: <SecurityScanOutlined style={{ color: '#722ed1' }} />,
                          content: '开启后每次登录需要输入短信验证码，确定开启吗？',
                          onOk: () => {
                            setTwoFactorEnabled(true)
                            message.success('两步验证已开启')
                          },
                        })
                      } else {
                        setTwoFactorEnabled(false)
                        message.info('两步验证已关闭')
                      }
                    }}
                  />
                </div>
              </Space>
            </Card>
          </div>
        </Col>

        <Col xs={24} lg={16} xl={17}>
          <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
            <Tabs
              defaultActiveKey="profile"
              size="large"
              style={{ padding: '4px 24px 0' }}
              items={[
                {
                  key: 'profile',
                  label: <Space><EditOutlined /> 基本资料</Space>,
                  children: (
                    <div style={{ padding: '8px 4px 24px' }}>
                      <div style={{ textAlign: 'right', marginBottom: 12 }}>
                        {editing ? (
                          <Space>
                            <Button onClick={() => { setEditing(false); profileForm.resetFields() }}>取消</Button>
                            <Button type="primary" icon={<CheckCircleOutlined />} onClick={handleProfileSave}>
                              保存修改
                            </Button>
                          </Space>
                        ) : (
                          <Button type="primary" icon={<EditOutlined />} onClick={() => {
                            profileForm.setFieldsValue(user)
                            setEditing(true)
                          }}>
                            编辑资料
                          </Button>
                        )}
                      </div>

                      {editing ? (
                        <Form
                          form={profileForm}
                          layout="vertical"
                          initialValues={user}
                        >
                          <Row gutter={16}>
                            <Col xs={24} md={12}>
                              <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}>
                                <Input />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="性别" name="gender">
                                <Radio.Group>
                                  <Radio value="male">男</Radio>
                                  <Radio value="female">女</Radio>
                                </Radio.Group>
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="手机号" name="phone">
                                <Input prefix={<PhoneOutlined />} />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="邮箱" name="email">
                                <Input prefix={<MailOutlined />} />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="岗位" name="position">
                                <Select>
                                  <Option value="调度中心主管">调度中心主管</Option>
                                  <Option value="高级调度员">高级调度员</Option>
                                  <Option value="普通调度员">普通调度员</Option>
                                  <Option value="安全管理员">安全管理员</Option>
                                </Select>
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="驾照类型" name="license_type">
                                <Select>
                                  <Option value="A1A2D">A1A2D</Option>
                                  <Option value="A2">A2</Option>
                                  <Option value="A1">A1</Option>
                                  <Option value="B1B2">B1B2</Option>
                                </Select>
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="从业资格证号" name="license_no">
                                <Input />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="入职日期" name="join_date">
                                <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="紧急联系人" name="emergency_contact">
                                <Input />
                              </Form.Item>
                            </Col>
                            <Col xs={24} md={12}>
                              <Form.Item label="紧急联系电话" name="emergency_phone">
                                <Input prefix={<PhoneOutlined />} />
                              </Form.Item>
                            </Col>
                            <Col xs={24}>
                              <Form.Item label="家庭地址" name="address">
                                <Input prefix={<EnvironmentOutlined />} />
                              </Form.Item>
                            </Col>
                          </Row>
                        </Form>
                      ) : (
                        <Descriptions column={2} bordered size="small" labelStyle={{ width: 140, background: '#fafafa' }}>
                          <Descriptions.Item label="姓名" span={1}>{user.name}</Descriptions.Item>
                          <Descriptions.Item label="性别" span={1}>{user.gender === 'male' ? '男' : '女'}</Descriptions.Item>
                          <Descriptions.Item label="手机号" span={1}>
                            <Space><PhoneOutlined style={{ color: '#52c41a' }} /> {user.phone}</Space>
                          </Descriptions.Item>
                          <Descriptions.Item label="邮箱" span={1}>
                            <Space><MailOutlined style={{ color: '#faad14' }} /> {user.email}</Space>
                          </Descriptions.Item>
                          <Descriptions.Item label="岗位职级" span={1}>{user.position}</Descriptions.Item>
                          <Descriptions.Item label="所属部门" span={1}>{user.department}</Descriptions.Item>
                          <Descriptions.Item label="驾照类型" span={1}>
                            <Tag color="purple">{user.license_type}</Tag>
                          </Descriptions.Item>
                          <Descriptions.Item label="从业资格证号" span={1}>
                            <Text copyable style={{ fontFamily: 'monospace' }}>{user.license_no}</Text>
                          </Descriptions.Item>
                          <Descriptions.Item label="学历背景" span={1}>{user.education}</Descriptions.Item>
                          <Descriptions.Item label="入职工龄" span={1}>
                            {dayjs().diff(dayjs(user.join_date), 'year')} 年 {dayjs().diff(dayjs(user.join_date), 'month') % 12} 个月
                          </Descriptions.Item>
                          <Descriptions.Item label="紧急联系人" span={1}>{user.emergency_contact}</Descriptions.Item>
                          <Descriptions.Item label="紧急电话" span={1}>{user.emergency_phone}</Descriptions.Item>
                          <Descriptions.Item label="身份证号" span={1}>
                            <Text style={{ fontFamily: 'monospace' }}>{user.id_card}</Text>
                          </Descriptions.Item>
                          <Descriptions.Item label="家庭地址" span={1}>
                            <EnvironmentOutlined /> {user.address}
                          </Descriptions.Item>
                        </Descriptions>
                      )}
                    </div>
                  ),
                },
                {
                  key: 'tasks',
                  label: <Space><BellOutlined /> 我的任务 <Badge count={mockTasks.filter(t => t.status !== 'completed' && t.status !== 'cancelled').length} offset={[6, -2]} /></Space>,
                  children: (
                    <div style={{ padding: '8px 4px 24px' }}>
                      <Row gutter={12} style={{ marginBottom: 16 }}>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#f0f5ff' }} size="small">
                            <Statistic title="全部任务" value={totalTasks} valueStyle={{ color: '#1677ff', fontSize: 20 }} />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#fffbe6' }} size="small">
                            <Statistic title="待处理" value={totalTasks - completedTasks} valueStyle={{ color: '#faad14', fontSize: 20 }} />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#f6ffed' }} size="small">
                            <Statistic title="已完成" value={completedTasks} valueStyle={{ color: '#52c41a', fontSize: 20 }} />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#f9f0ff' }} size="small">
                            <Statistic title="完成率" value={`${Math.round(completedTasks / totalTasks * 100)}%`} valueStyle={{ color: '#722ed1', fontSize: 20 }} />
                          </Card>
                        </Col>
                      </Row>

                      <Table
                        rowKey="id"
                        columns={taskColumns as any}
                        dataSource={mockTasks}
                        pagination={{
                          current: taskPage,
                          pageSize: taskPageSize,
                          showSizeChanger: true,
                          showQuickJumper: true,
                          showTotal: t => `共 ${t} 条`,
                          onChange: (p, ps) => { setTaskPage(p); setTaskPageSize(ps) },
                        }}
                        scroll={{ x: 1100 }}
                        rowClassName={(r) => r.priority === 'high' ? '!bg-red-50/50' : ''}
                        size="middle"
                      />
                    </div>
                  ),
                },
                {
                  key: 'logs',
                  label: <Space><FileTextOutlined /> 操作日志</Space>,
                  children: (
                    <div style={{ padding: '8px 4px 24px' }}>
                      <div style={{ marginBottom: 12, textAlign: 'right' }}>
                        <Space wrap>
                          <Select allowClear placeholder="模块筛选" style={{ width: 140 }} value={logModuleFilter} onChange={setLogModuleFilter}>
                            {Array.from(new Set(mockLogs.map(l => l.module))).map(m => (
                              <Option key={m} value={m}>{m}</Option>
                            ))}
                          </Select>
                          <Select allowClear placeholder="结果筛选" style={{ width: 120 }} value={logResultFilter} onChange={setLogResultFilter}>
                            {Object.entries(logResultMap).map(([k, v]) => (
                              <Option key={k} value={k}>{v.label}</Option>
                            ))}
                          </Select>
                          <Button icon={<ReloadOutlined />}>刷新</Button>
                          <Button icon={<ExportOutlined />}>导出</Button>
                        </Space>
                      </div>
                      <Table
                        rowKey="id"
                        columns={logColumns as any}
                        dataSource={mockLogs.filter(l =>
                          (!logModuleFilter || l.module === logModuleFilter) &&
                          (!logResultFilter || l.result === logResultFilter)
                        )}
                        pagination={{
                          current: logPage,
                          pageSize: logPageSize,
                          showSizeChanger: true,
                          showQuickJumper: true,
                          showTotal: t => `共 ${t} 条`,
                          onChange: (p, ps) => { setLogPage(p); setLogPageSize(ps) },
                        }}
                        scroll={{ x: 1050 }}
                        size="middle"
                        rowClassName={(r) => r.result === 'failed' ? '!bg-red-50/50' : r.result === 'warning' ? '!bg-orange-50/50' : ''}
                      />
                    </div>
                  ),
                },
                {
                  key: 'score',
                  label: <Space><SafetyCertificateOutlined /> 驾驶评分统计</Space>,
                  children: (
                    <div style={{ padding: '8px 4px 24px' }}>
                      <Row gutter={12} style={{ marginBottom: 16 }}>
                        <Col xs={12} sm={6}>
                          <Card
                            bordered={false}
                            style={{ borderRadius: 8, background: 'linear-gradient(135deg, #1677ff, #722ed1)', color: '#fff' }}
                            size="small"
                          >
                            <Statistic
                              title={<span style={{ color: 'rgba(255,255,255,0.8)' }}>当前安全评分</span>}
                              value={currentScore}
                              suffix="/ 100"
                              valueStyle={{ color: '#fff', fontSize: 28, fontWeight: 700 }}
                            />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#f6ffed' }} size="small">
                            <Statistic
                              title="分公司排名"
                              value={scoreRank}
                              valueStyle={{ color: '#52c41a', fontSize: 20, fontWeight: 600 }}
                              prefix={<TrophyIcon />}
                            />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#fff7e6' }} size="small">
                            <Statistic
                              title="超行业均值"
                              value={`+${currentScore - 82}`}
                              valueStyle={{ color: '#fa8c16', fontSize: 20, fontWeight: 600 }}
                              prefix={<RiseOutlined />}
                            />
                          </Card>
                        </Col>
                        <Col xs={12} sm={6}>
                          <Card bordered={false} style={{ borderRadius: 8, background: '#f9f0ff' }} size="small">
                            <Statistic
                              title="累计运输安全"
                              value="1286"
                              suffix="天"
                              valueStyle={{ color: '#722ed1', fontSize: 20, fontWeight: 600 }}
                              prefix={<CheckCircleOutlined />}
                            />
                          </Card>
                        </Col>
                      </Row>

                      <Row gutter={16}>
                        <Col xs={24} lg={14}>
                          <Card
                            bordered={false}
                            style={{ borderRadius: 12 }}
                            size="small"
                            title={<Space><InfoCircleOutlined style={{ color: '#1677ff' }} /> 12个月安全评分趋势</Space>}
                          >
                            <ReactECharts option={drivingScoreChart} style={{ height: 280 }} notMerge />
                          </Card>
                        </Col>
                        <Col xs={24} lg={10}>
                          <Card
                            bordered={false}
                            style={{ borderRadius: 12 }}
                            size="small"
                            title={<Space><BarChartOutlined style={{ color: '#52c41a' }} /> 各维度得分对比</Space>}
                          >
                            <ReactECharts option={scoreDetailChart} style={{ height: 280 }} notMerge />
                          </Card>
                        </Col>
                      </Row>

                      <Divider style={{ margin: '20px 0 12px' }} />

                      <Card
                        bordered={false}
                        style={{ borderRadius: 12 }}
                        size="small"
                        title={<Space><WarningOutlined style={{ color: '#fa8c16' }} /> 评分明细记录</Space>}
                      >
                        <Descriptions column={3} bordered size="small" labelStyle={{ width: 120, background: '#fafafa' }}>
                          <Descriptions.Item label="安全驾驶">
                            <Progress percent={96} size="small" strokeColor="#1677ff" />
                          </Descriptions.Item>
                          <Descriptions.Item label="合规操作">
                            <Progress percent={94} size="small" strokeColor="#52c41a" />
                          </Descriptions.Item>
                          <Descriptions.Item label="车辆维护">
                            <Progress percent={92} size="small" strokeColor="#13c2c2" />
                          </Descriptions.Item>
                          <Descriptions.Item label="应急处理">
                            <Progress percent={90} size="small" strokeColor="#722ed1" />
                          </Descriptions.Item>
                          <Descriptions.Item label="文明行车">
                            <Progress percent={98} size="small" strokeColor="#fa8c16" />
                          </Descriptions.Item>
                          <Descriptions.Item label="培训学习">
                            <Progress percent={95} size="small" strokeColor="#eb2f96" />
                          </Descriptions.Item>
                        </Descriptions>
                      </Card>
                    </div>
                  ),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title={<Space><LockOutlined style={{ color: '#1677ff' }} /> 修改登录密码</Space>}
        open={passwordModal}
        onCancel={() => { setPasswordModal(false); passwordForm.resetFields() }}
        onOk={() => passwordForm.submit()}
        okText="确认修改"
        okButtonProps={{ type: 'primary' }}
        width={480}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          icon={<InfoCircleOutlined />}
          message="密码安全要求"
          description={
            <ul style={{ margin: 0, paddingLeft: 18, fontSize: 12 }}>
              <li>长度为 8-20 位</li>
              <li>必须包含大小写字母、数字和特殊符号中的至少3种</li>
              <li>不得与最近3次使用的密码相同</li>
            </ul>
          }
          style={{ borderRadius: 8, marginBottom: 16 }}
        />

        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handlePasswordChange}
        >
          <Form.Item
            label="当前密码"
            name="old_password"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入当前登录密码" />
          </Form.Item>
          <Form.Item
            label="新密码"
            name="new_password"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, max: 20, message: '密码长度应为8-20位' },
              {
                pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)|(?=.*[a-z])(?=.*[A-Z])(?=.*[^A-Za-z0-9])|(?=.*[a-z])(?=.*\d)(?=.*[^A-Za-z0-9])|(?=.*[A-Z])(?=.*\d)(?=.*[^A-Za-z0-9])/,
                message: '需包含大小写字母、数字、特殊符号中至少3种',
              },
            ]}
          >
            <Input.Password prefix={<KeyOutlined />} placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            label="确认新密码"
            name="confirm_password"
            rules={[
              { required: true, message: '请再次输入新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password prefix={<KeyOutlined />} placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

function SafetyCertificateOutlined(props: any) {
  return <span {...props} style={{ display: 'inline-flex', alignItems: 'center' }}>🛡️</span>
}

function TrophyIcon() {
  return <span>🏆</span>
}

function RiseOutlined() {
  return <span>📈</span>
}

function BarChartOutlined(props: any) {
  return <span {...props}>📊</span>
}

export default Profile
