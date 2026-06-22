import React, { useEffect, useState, useMemo } from 'react'
import {
  Row,
  Col,
  Card,
  Statistic,
  Tag,
  Button,
  Space,
  Typography,
  Progress,
  Table,
  Drawer,
  Descriptions,
  Badge,
  Select,
  DatePicker,
  Tabs,
  Tooltip,
  Avatar,
  List,
  Empty,
  Spin,
  message,
  Modal,
  Input,
  InputNumber,
} from 'antd'
import {
  TrophyOutlined,
  RiseOutlined,
  FallOutlined,
  StarOutlined,
  WarningOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  UserOutlined,
  CarOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  SendOutlined,
  GiftOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs'
import { scoreApi, ScoreOverview, ScoreLeaderboardItem, DrivingScoreRecord, ScoreBonus, MonthlyReport, RetrainingTask } from '@/services/api'
import { hasPermission } from '@/utils/auth'

const { Title, Text } = Typography
const { Option } = Select
const { RangePicker } = DatePicker

const scoreLevelMap: Record<string, { label: string; color: string }> = {
  excellent: { label: '优秀', color: '#52c41a' },
  good: { label: '良好', color: '#1677ff' },
  normal: { label: '一般', color: '#faad14' },
  poor: { label: '较差', color: '#fa8c16' },
  danger: { label: '危险', color: '#ff4d4f' },
}

const bonusTypeMap: Record<string, string> = {
  no_violation_30d: '连续30天无违规',
  safety_champion: '安全驾驶标兵',
  continuous_clean: '持续无违规',
}

const retrainingStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待培训', color: 'red' },
  in_progress: { label: '培训中', color: 'orange' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'default' },
}

const taskTypeMap: Record<string, string> = {
  safety_training: '安全培训',
  rule_exam: '交规考核',
  mentor_drive: '导师陪驾',
  observation: '观察学习',
}

const triggerTypeMap: Record<string, string> = {
  low_score: '评分过低',
  serious_violation: '严重违规',
  repeated_violation: '多次违规',
}

const mockOverview: ScoreOverview = {
  total_drivers: 40,
  excellent_count: 15,
  good_count: 12,
  normal_count: 7,
  poor_count: 4,
  danger_count: 2,
  avg_score: 82.5,
  retraining_count: 3,
  today_calculated: 38,
  need_retraining_count: 3,
}

const mockLeaderboard: ScoreLeaderboardItem[] = Array.from({ length: 20 }, (_, i) => ({
  driver_id: i + 1,
  driver_name: ['张建国', '李志强', '王海军', '赵卫东', '陈光明', '刘振华', '杨大勇', '黄伟峰', '周明辉', '吴建军', '郑海涛', '孙文博', '马俊杰', '朱永刚', '胡德明', '林国安', '何建设', '罗正阳', '谢天佑', '邓伟才'][i],
  plate_number: `京A${(10000 + i).toString().padStart(5, '0')}`,
  total_score: Math.max(45, 100 - i * 3 - Math.floor(Math.random() * 5)),
  score_level: i < 5 ? 'excellent' : i < 12 ? 'good' : i < 17 ? 'normal' : i < 19 ? 'poor' : 'danger',
  rank: i + 1,
  avg_score_30d: Math.max(48, 98 - i * 2.5 - Math.floor(Math.random() * 3)),
  bonus_points: i < 5 ? 5 : i < 10 ? 2 : 0,
  org_name: '危险品运输示范企业',
}))

const mockDriverScore: DrivingScoreRecord = {
  id: 1,
  driver_id: 1,
  waybill_id: null,
  vehicle_id: 1,
  trip_date: dayjs().format('YYYY-MM-DD'),
  total_score: 88,
  score_level: 'good',
  fatigue_score: 92,
  fatigue_deduction: 3,
  overspeed_score: 90,
  overspeed_count: 1,
  overspeed_deduction: 2,
  sudden_brake_count: 2,
  sudden_brake_deduction: 2,
  sudden_accel_count: 1,
  sudden_accel_deduction: 1,
  sharp_turn_count: 0,
  sharp_turn_deduction: 0,
  lane_deviation_count: 1,
  lane_deviation_deduction: 1,
  phone_usage_count: 0,
  phone_usage_deduction: 0,
  smoking_count: 0,
  smoking_deduction: 0,
  seatbelt_violation_count: 0,
  seatbelt_violation_deduction: 0,
  route_deviation_count: 0,
  route_deviation_deduction: 0,
  fatigue_alarm_count: 1,
  total_distance: 256.8,
  driving_duration: 320,
  night_driving_duration: 45,
  overspeed_duration: 3.5,
}

const mockHistory: DrivingScoreRecord[] = Array.from({ length: 30 }, (_, i) => ({
  ...mockDriverScore,
  id: i + 1,
  trip_date: dayjs().subtract(29 - i, 'day').format('YYYY-MM-DD'),
  total_score: Math.max(65, 88 - Math.floor(Math.random() * 15) + Math.floor(Math.random() * 10)),
  fatigue_alarm_count: Math.floor(Math.random() * 3),
  overspeed_count: Math.floor(Math.random() * 2),
  sudden_brake_count: Math.floor(Math.random() * 3),
  sudden_accel_count: Math.floor(Math.random() * 2),
  sharp_turn_count: Math.floor(Math.random() * 2),
}))

const mockBonuses: ScoreBonus[] = [
  { id: 1, driver_id: 1, bonus_type: 'no_violation_30d', bonus_points: 5, reason: '连续30天无违规，奖励加分', streak_days: 30, start_date: dayjs().subtract(30, 'day').format('YYYY-MM-DD'), end_date: dayjs().format('YYYY-MM-DD'), awarded_by: 0, status: 1, created_at: dayjs().format('YYYY-MM-DD HH:mm:ss') },
]

const mockMonthlyReports: MonthlyReport[] = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1,
  driver_id: i + 1,
  report_month: dayjs().format('YYYY-MM'),
  avg_score: Math.max(50, 90 - i * 4 + Math.floor(Math.random() * 5)),
  min_score: Math.max(40, 70 - i * 3),
  max_score: Math.min(100, 95 - i),
  total_fatigue_alarms: Math.floor(Math.random() * 8),
  total_sudden_events: Math.floor(Math.random() * 12),
  total_overspeed_duration: Math.floor(Math.random() * 30),
  total_distance: 2000 + Math.floor(Math.random() * 5000),
  total_driving_duration: 4000 + Math.floor(Math.random() * 8000),
  total_bonus_points: i < 3 ? 5 : 0,
  violation_days: Math.floor(Math.random() * 5),
  clean_days: 25 - Math.floor(Math.random() * 5),
  score_trend: Array.from({ length: 30 }, (_, di) => ({
    date: dayjs().subtract(29 - di, 'day').format('MM-DD'),
    score: Math.max(50, 85 - i * 2 + Math.floor(Math.random() * 15)),
  })),
  need_retraining: i >= 7 ? 1 : 0,
  retraining_triggered_at: i >= 7 ? dayjs().format('YYYY-MM-DD HH:mm:ss') : null,
  report_sent: i < 5 ? 1 : 0,
  report_sent_at: i < 5 ? dayjs().subtract(i, 'day').format('YYYY-MM-DD HH:mm:ss') : null,
}))

const mockRetrainingTasks: RetrainingTask[] = [
  { id: 1, driver_id: 18, trigger_score: 52, trigger_type: 'low_score', trigger_month: dayjs().format('YYYY-MM'), task_type: 'safety_training', status: 'pending', assigned_at: null, started_at: null, completed_at: null, result_score: null, result_note: null, created_by: 0, created_at: dayjs().subtract(2, 'day').format('YYYY-MM-DD HH:mm:ss') },
  { id: 2, driver_id: 19, trigger_score: 48, trigger_type: 'low_score', trigger_month: dayjs().format('YYYY-MM'), task_type: 'rule_exam', status: 'in_progress', assigned_at: dayjs().subtract(1, 'day').format('YYYY-MM-DD HH:mm:ss'), started_at: dayjs().subtract(1, 'day').format('YYYY-MM-DD HH:mm:ss'), completed_at: null, result_score: null, result_note: null, created_by: 0, created_at: dayjs().subtract(3, 'day').format('YYYY-MM-DD HH:mm:ss') },
  { id: 3, driver_id: 20, trigger_score: 55, trigger_type: 'repeated_violation', trigger_month: dayjs().subtract(1, 'month').format('YYYY-MM'), task_type: 'mentor_drive', status: 'completed', assigned_at: dayjs().subtract(10, 'day').format('YYYY-MM-DD HH:mm:ss'), started_at: dayjs().subtract(9, 'day').format('YYYY-MM-DD HH:mm:ss'), completed_at: dayjs().subtract(2, 'day').format('YYYY-MM-DD HH:mm:ss'), result_score: 82, result_note: '培训考核通过', created_by: 0, created_at: dayjs().subtract(12, 'day').format('YYYY-MM-DD HH:mm:ss') },
]

const DrivingScore: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [overview, setOverview] = useState<ScoreOverview>(mockOverview)
  const [leaderboard, setLeaderboard] = useState<ScoreLeaderboardItem[]>(mockLeaderboard)
  const [detailDrawer, setDetailDrawer] = useState<ScoreLeaderboardItem | null>(null)
  const [driverScoreData, setDriverScoreData] = useState<{ latest: DrivingScoreRecord; history: DrivingScoreRecord[] } | null>(null)
  const [driverBonuses, setDriverBonuses] = useState<ScoreBonus[]>([])
  const [monthlyReports, setMonthlyReports] = useState<MonthlyReport[]>(mockMonthlyReports)
  const [retrainingTasks, setRetrainingTasks] = useState<RetrainingTask[]>(mockRetrainingTasks)
  const [reportMonth, setReportMonth] = useState(dayjs().format('YYYY-MM'))
  const isAdmin = hasPermission('score:manage')

  useEffect(() => {
    setLoading(true)
    setTimeout(() => {
      setOverview(mockOverview)
      setLeaderboard(mockLeaderboard)
      setMonthlyReports(mockMonthlyReports)
      setRetrainingTasks(mockRetrainingTasks)
      setLoading(false)
    }, 500)
  }, [])

  const handleViewDetail = (item: ScoreLeaderboardItem) => {
    setDetailDrawer(item)
    setDriverScoreData({ latest: { ...mockDriverScore, driver_id: item.driver_id, total_score: item.total_score, score_level: item.score_level }, history: mockHistory.map(h => ({ ...h, driver_id: item.driver_id })) })
    setDriverBonuses(item.bonus_points > 0 ? mockBonuses : [])
  }

  const scoreTrendChart = useMemo(() => {
    if (!driverScoreData?.history?.length) return {}
    return {
      tooltip: { trigger: 'axis' },
      grid: { left: 50, right: 20, top: 30, bottom: 30 },
      xAxis: {
        type: 'category',
        data: driverScoreData.history.map(h => dayjs(h.trip_date).format('MM-DD')),
        axisLabel: { fontSize: 10, interval: 2 },
      },
      yAxis: {
        type: 'value',
        min: 0,
        max: 100,
        splitLine: { lineStyle: { color: '#f0f0f0' } },
      },
      visualMap: {
        show: false,
        pieces: [
          { lte: 60, color: '#ff4d4f' },
          { gt: 60, lte: 75, color: '#fa8c16' },
          { gt: 75, lte: 90, color: '#1677ff' },
          { gt: 90, color: '#52c41a' },
        ],
      },
      series: [{
        type: 'line',
        smooth: true,
        data: driverScoreData.history.map(h => h.total_score),
        markLine: {
          silent: true,
          data: [{ yAxis: 60, lineStyle: { color: '#ff4d4f', type: 'dashed' }, label: { formatter: '再培训线', position: 'end' } }],
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22,119,255,0.3)' },
              { offset: 1, color: 'rgba(22,119,255,0.02)' },
            ],
          },
        },
        lineStyle: { width: 3 },
      }],
    }
  }, [driverScoreData])

  const deductionBreakdownChart = useMemo(() => {
    if (!driverScoreData?.latest) return {}
    const s = driverScoreData.latest
    return {
      tooltip: { trigger: 'item', formatter: '{b}: {c}分 ({d}%)' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}\n{c}分' },
        data: [
          { name: '疲劳扣分', value: s.fatigue_deduction, itemStyle: { color: '#ff4d4f' } },
          { name: '超速扣分', value: s.overspeed_deduction, itemStyle: { color: '#fa8c16' } },
          { name: '急刹车', value: s.sudden_brake_deduction, itemStyle: { color: '#faad14' } },
          { name: '急加速', value: s.sudden_accel_deduction, itemStyle: { color: '#eb2f96' } },
          { name: '急转弯', value: s.sharp_turn_deduction, itemStyle: { color: '#722ed1' } },
          { name: '车道偏离', value: s.lane_deviation_deduction, itemStyle: { color: '#13c2c2' } },
          { name: '手机使用', value: s.phone_usage_deduction, itemStyle: { color: '#a0d911' } },
          { name: '抽烟', value: s.smoking_deduction, itemStyle: { color: '#9254de' } },
        ].filter(d => d.value > 0),
      }],
    }
  }, [driverScoreData])

  const levelDistributionChart = useMemo(() => ({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: ['50%', '75%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 3 },
      label: { show: true, formatter: '{b}\n{c}人', fontSize: 12 },
      data: [
        { name: '优秀(≥90)', value: overview.excellent_count, itemStyle: { color: '#52c41a' } },
        { name: '良好(75-89)', value: overview.good_count, itemStyle: { color: '#1677ff' } },
        { name: '一般(60-74)', value: overview.normal_count, itemStyle: { color: '#faad14' } },
        { name: '较差(40-59)', value: overview.poor_count, itemStyle: { color: '#fa8c16' } },
        { name: '危险(<40)', value: overview.danger_count, itemStyle: { color: '#ff4d4f' } },
      ].filter(d => d.value > 0),
    }],
  }), [overview])

  const leaderboardColumns = [
    {
      title: '排名',
      dataIndex: 'rank',
      width: 70,
      render: (v: number) => {
        if (v === 1) return <Tag color="gold" style={{ fontWeight: 700 }}>🥇 1</Tag>
        if (v === 2) return <Tag color="default" style={{ fontWeight: 700 }}>🥈 2</Tag>
        if (v === 3) return <Tag color="orange" style={{ fontWeight: 700 }}>🥉 3</Tag>
        return <Text type="secondary">#{v}</Text>
      },
    },
    {
      title: '驾驶员',
      dataIndex: 'driver_name',
      width: 140,
      render: (v: string, r: ScoreLeaderboardItem) => (
        <Space>
          <Avatar icon={<UserOutlined />} style={{ backgroundColor: `hsl(${(r.driver_id * 37) % 360}, 60%, 50%)` }} size="small" />
          <div>
            <Text strong>{v}</Text>
            <div><Text type="secondary" style={{ fontSize: 11 }}>{r.plate_number}</Text></div>
          </div>
        </Space>
      ),
    },
    {
      title: '今日评分',
      dataIndex: 'total_score',
      width: 160,
      sorter: (a: ScoreLeaderboardItem, b: ScoreLeaderboardItem) => a.total_score - b.total_score,
      render: (v: number) => (
        <Progress
          percent={v}
          size="small"
          strokeColor={v >= 90 ? '#52c41a' : v >= 75 ? '#1677ff' : v >= 60 ? '#faad14' : '#ff4d4f'}
          format={p => <Text strong style={{ color: v >= 90 ? '#52c41a' : v >= 75 ? '#1677ff' : v >= 60 ? '#faad14' : '#ff4d4f', fontSize: 12 }}>{p}</Text>}
        />
      ),
    },
    {
      title: '等级',
      dataIndex: 'score_level',
      width: 80,
      render: (v: string) => <Tag color={scoreLevelMap[v]?.color}>{scoreLevelMap[v]?.label}</Tag>,
    },
    {
      title: '30日均分',
      dataIndex: 'avg_score_30d',
      width: 100,
      render: (v: number) => <Text style={{ color: v >= 80 ? '#52c41a' : '#fa8c16' }}>{v.toFixed(1)}</Text>,
    },
    {
      title: '加分',
      dataIndex: 'bonus_points',
      width: 70,
      render: (v: number) => v > 0 ? (
        <Tag color="green" icon={<GiftOutlined />}>+{v}</Tag>
      ) : <Text type="secondary">0</Text>,
    },
    {
      title: '操作',
      width: 80,
      render: (_: any, r: ScoreLeaderboardItem) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(r)}>详情</Button>
      ),
    },
  ]

  const monthlyReportColumns = [
    {
      title: '驾驶员ID',
      dataIndex: 'driver_id',
      width: 90,
    },
    {
      title: '月份',
      dataIndex: 'report_month',
      width: 100,
    },
    {
      title: '月均分',
      dataIndex: 'avg_score',
      width: 120,
      sorter: (a: MonthlyReport, b: MonthlyReport) => a.avg_score - b.avg_score,
      render: (v: number) => (
        <Progress
          percent={v}
          size="small"
          strokeColor={v >= 90 ? '#52c41a' : v >= 75 ? '#1677ff' : v >= 60 ? '#faad14' : '#ff4d4f'}
          format={p => <Text strong style={{ fontSize: 11 }}>{p}</Text>}
        />
      ),
    },
    {
      title: '疲劳报警',
      dataIndex: 'total_fatigue_alarms',
      width: 90,
      render: (v: number) => <Badge count={v} style={{ backgroundColor: v > 5 ? '#ff4d4f' : '#fa8c16' }} />,
    },
    {
      title: '急变道事件',
      dataIndex: 'total_sudden_events',
      width: 100,
      render: (v: number) => <Badge count={v} style={{ backgroundColor: v > 8 ? '#ff4d4f' : '#faad14' }} />,
    },
    {
      title: '无违规天',
      dataIndex: 'clean_days',
      width: 90,
      render: (v: number) => <Tag color={v >= 25 ? 'green' : v >= 15 ? 'blue' : 'orange'}>{v}天</Tag>,
    },
    {
      title: '需再培训',
      dataIndex: 'need_retraining',
      width: 90,
      render: (v: number) => v === 1 ? (
        <Tag color="red" icon={<WarningOutlined />}>需要</Tag>
      ) : <Tag color="green" icon={<CheckCircleOutlined />}>无需</Tag>,
    },
    {
      title: '报告状态',
      dataIndex: 'report_sent',
      width: 100,
      render: (v: number) => v === 1 ? (
        <Tag color="green">已推送</Tag>
      ) : <Tag color="orange">待推送</Tag>,
    },
    {
      title: '操作',
      width: 80,
      render: (_: any, r: MonthlyReport) => (
        <Button
          type="link"
          size="small"
          icon={<SendOutlined />}
          disabled={r.report_sent === 1}
          onClick={() => {
            message.success('月报已推送至司机和管理员邮箱')
          }}
        >
          推送
        </Button>
      ),
    },
  ]

  const retrainingColumns = [
    {
      title: '驾驶员ID',
      dataIndex: 'driver_id',
      width: 90,
    },
    {
      title: '触发评分',
      dataIndex: 'trigger_score',
      width: 90,
      render: (v: number) => <Text type="danger" strong>{v}</Text>,
    },
    {
      title: '触发原因',
      dataIndex: 'trigger_type',
      width: 100,
      render: (v: string) => <Tag color="orange">{triggerTypeMap[v] || v}</Tag>,
    },
    {
      title: '培训类型',
      dataIndex: 'task_type',
      width: 100,
      render: (v: string) => taskTypeMap[v] || v,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v: string) => <Tag color={retrainingStatusMap[v]?.color}>{retrainingStatusMap[v]?.label}</Tag>,
    },
    {
      title: '考核分数',
      dataIndex: 'result_score',
      width: 90,
      render: (v: number | null) => v ? <Text type={v >= 80 ? 'success' : 'warning'} strong>{v}</Text> : '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 150,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '操作',
      width: 120,
      render: (_: any, r: RetrainingTask) => (
        <Space size={4}>
          {r.status === 'pending' && (
            <Button type="link" size="small" onClick={() => message.success('已开始培训')}>开始</Button>
          )}
          {r.status === 'in_progress' && (
            <Button type="link" size="small" onClick={() => message.success('培训已完成')}>完成</Button>
          )}
        </Space>
      ),
    },
  ]

  const getScoreColor = (score: number) => {
    if (score >= 90) return '#52c41a'
    if (score >= 75) return '#1677ff'
    if (score >= 60) return '#faad14'
    return '#ff4d4f'
  }

  if (loading) {
    return (
      <div style={{ padding: 40, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="参与评分司机"
              value={overview.total_drivers}
              valueStyle={{ color: '#1677ff' }}
              prefix={<TeamOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>人</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日平均评分"
              value={overview.avg_score}
              precision={1}
              valueStyle={{ color: getScoreColor(overview.avg_score) }}
              prefix={<StarOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ 100</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="优秀+良好"
              value={overview.excellent_count + overview.good_count}
              valueStyle={{ color: '#52c41a' }}
              prefix={<SafetyCertificateOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ {overview.total_drivers}</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="较差+危险"
              value={overview.poor_count + overview.danger_count}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>人</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="待再培训"
              value={overview.need_retraining_count}
              valueStyle={{ color: '#fa8c16' }}
              prefix={<ExclamationCircleOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>人</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日已计算"
              value={overview.today_calculated}
              valueStyle={{ color: '#13c2c2' }}
              prefix={<CheckCircleOutlined />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ {overview.total_drivers}</Text>}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={14}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={
              <Space>
                <TrophyOutlined style={{ color: '#faad14' }} />
                <Text strong style={{ fontSize: 15 }}>评分排行榜</Text>
                <Tag color="blue" style={{ margin: 0 }}>车队内排名</Tag>
              </Space>
            }
            extra={
              <Space>
                <Select defaultValue="total_score" style={{ width: 120 }} size="small">
                  <Option value="total_score">按今日分</Option>
                  <Option value="avg_score_30d">按30日均</Option>
                </Select>
                <Button icon={<ReloadOutlined />} size="small">刷新</Button>
              </Space>
            }
          >
            <Table
              rowKey="driver_id"
              columns={leaderboardColumns}
              dataSource={leaderboard}
              pagination={{ pageSize: 10, showTotal: t => `共 ${t} 人` }}
              size="small"
              scroll={{ y: 420 }}
            />
          </Card>
        </Col>
        <Col xs={24} lg={10} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={<Space><RiseOutlined style={{ color: '#1677ff' }} /> 评分等级分布</Space>}
          >
            <ReactECharts option={levelDistributionChart} style={{ height: 260 }} notMerge />
          </Card>
          <Card
            bordered={false}
            style={{ borderRadius: 12 }}
            title={
              <Space>
                <GiftOutlined style={{ color: '#52c41a' }} />
                <Text strong>加分项规则</Text>
              </Space>
            }
          >
            <List
              size="small"
              dataSource={[
                { title: '连续30天无违规', points: '+5分', desc: '连续30天评分≥60分，无任何违规行为', icon: <SafetyCertificateOutlined style={{ color: '#52c41a', fontSize: 20 }} /> },
                { title: '安全驾驶标兵', points: '+3分', desc: '月度评分排名前3名', icon: <TrophyOutlined style={{ color: '#faad14', fontSize: 20 }} /> },
                { title: '持续无违规', points: '+2分/月', desc: '每连续60天无违规额外加分', icon: <StarOutlined style={{ color: '#1677ff', fontSize: 20 }} /> },
              ]}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={item.icon}
                    title={<Space><Text strong>{item.title}</Text><Tag color="green">{item.points}</Tag></Space>}
                    description={<Text type="secondary" style={{ fontSize: 12 }}>{item.desc}</Text>}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      <Card bordered={false} style={{ borderRadius: 12 }}>
        <Tabs
          defaultActiveKey="monthly"
          items={[
            {
              key: 'monthly',
              label: <Space><ClockCircleOutlined /> 评分月报</Space>,
              children: (
                <div>
                  <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space>
                      <Text type="secondary">选择月份：</Text>
                      <DatePicker
                        picker="month"
                        defaultValue={dayjs()}
                        onChange={d => setReportMonth(d?.format('YYYY-MM') || dayjs().format('YYYY-MM'))}
                        allowClear={false}
                      />
                      <Button
                        type="primary"
                        icon={<SendOutlined />}
                        onClick={() => message.success('所有未推送月报已推送至司机和管理员邮箱')}
                      >
                        批量推送月报
                      </Button>
                    </Space>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      月均分低于60分自动触发再培训任务
                    </Text>
                  </div>
                  <Table
                    rowKey="id"
                    columns={monthlyReportColumns}
                    dataSource={monthlyReports}
                    pagination={{ pageSize: 10, showTotal: t => `共 ${t} 条` }}
                    size="small"
                  />
                </div>
              ),
            },
            {
              key: 'retraining',
              label: <Space><ExclamationCircleOutlined /> 再培训任务</Space>,
              children: (
                <div>
                  <div style={{ marginBottom: 16 }}>
                    <Space>
                      <Text type="secondary">状态筛选：</Text>
                      <Select defaultValue="" style={{ width: 120 }} allowClear placeholder="全部状态">
                        <Option value="pending">待培训</Option>
                        <Option value="in_progress">培训中</Option>
                        <Option value="completed">已完成</Option>
                      </Select>
                    </Space>
                  </div>
                  <Table
                    rowKey="id"
                    columns={retrainingColumns}
                    dataSource={retrainingTasks}
                    pagination={{ pageSize: 10 }}
                    size="small"
                  />
                </div>
              ),
            },
          ]}
        />
      </Card>

      <Drawer
        title={
          detailDrawer ? (
            <Space>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: `hsl(${(detailDrawer.driver_id * 37) % 360}, 60%, 50%)` }} />
              <div>
                <Text strong style={{ fontSize: 16 }}>{detailDrawer.driver_name}</Text>
                <div>
                  <Space size={4}>
                    <Tag color={scoreLevelMap[detailDrawer.score_level]?.color}>
                      {scoreLevelMap[detailDrawer.score_level]?.label}
                    </Tag>
                    <Text type="secondary">{detailDrawer.plate_number}</Text>
                    <Tag color="blue">排名 #{detailDrawer.rank}</Tag>
                  </Space>
                </div>
              </div>
            </Space>
          ) : null
        }
        open={!!detailDrawer}
        onClose={() => { setDetailDrawer(null); setDriverScoreData(null); setDriverBonuses([]) }}
        width={780}
      >
        {detailDrawer && driverScoreData && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Row gutter={16}>
              <Col span={8}>
                <Card size="small" style={{ borderRadius: 8, textAlign: 'center' }}>
                  <div style={{
                    width: 100, height: 100, borderRadius: '50%',
                    border: `4px solid ${getScoreColor(driverScoreData.latest.total_score)}`,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    margin: '0 auto 8px',
                  }}>
                    <span style={{
                      fontSize: 32, fontWeight: 700,
                      color: getScoreColor(driverScoreData.latest.total_score),
                    }}>
                      {Math.round(driverScoreData.latest.total_score)}
                    </span>
                  </div>
                  <Text type="secondary">今日安全评分</Text>
                </Card>
              </Col>
              <Col span={16}>
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><ThunderboltOutlined style={{ color: '#fa8c16' }} /> 扣分明细</Space>}>
                  <Descriptions column={2} size="small">
                    <Descriptions.Item label="疲劳扣分">
                      <Text type="danger">-{driverScoreData.latest.fatigue_deduction}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="超速扣分">
                      <Text type="danger">-{driverScoreData.latest.overspeed_deduction}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="急刹车">
                      <Text type="danger">-{driverScoreData.latest.sudden_brake_deduction}</Text>
                      <Text type="secondary" style={{ fontSize: 11 }}>({driverScoreData.latest.sudden_brake_count}次)</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="急加速">
                      <Text type="danger">-{driverScoreData.latest.sudden_accel_deduction}</Text>
                      <Text type="secondary" style={{ fontSize: 11 }}>({driverScoreData.latest.sudden_accel_count}次)</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="急转弯">
                      <Text type="danger">-{driverScoreData.latest.sharp_turn_deduction}</Text>
                      <Text type="secondary" style={{ fontSize: 11 }}>({driverScoreData.latest.sharp_turn_count}次)</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="车道偏离">
                      <Text type="danger">-{driverScoreData.latest.lane_deviation_deduction}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="疲劳报警">
                      <Badge count={driverScoreData.latest.fatigue_alarm_count} style={{ backgroundColor: '#ff4d4f' }} />
                    </Descriptions.Item>
                    <Descriptions.Item label="超速次数">
                      <Badge count={driverScoreData.latest.overspeed_count} style={{ backgroundColor: '#fa8c16' }} />
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={14}>
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><RiseOutlined style={{ color: '#1677ff' }} /> 近30天评分趋势</Space>}>
                  <ReactECharts option={scoreTrendChart} style={{ height: 260 }} notMerge />
                </Card>
              </Col>
              <Col span={10}>
                <Card size="small" style={{ borderRadius: 8 }} title={<Space><FallOutlined style={{ color: '#ff4d4f' }} /> 扣分构成</Space>}>
                  {driverScoreData.latest.fatigue_deduction + driverScoreData.latest.overspeed_deduction + driverScoreData.latest.sudden_brake_deduction + driverScoreData.latest.sudden_accel_deduction + driverScoreData.latest.sharp_turn_deduction + driverScoreData.latest.lane_deviation_deduction > 0 ? (
                    <ReactECharts option={deductionBreakdownChart} style={{ height: 260 }} notMerge />
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                      <CheckCircleOutlined style={{ fontSize: 40, color: '#52c41a' }} />
                      <div style={{ marginTop: 12 }}><Text type="secondary">今日无扣分记录</Text></div>
                    </div>
                  )}
                </Card>
              </Col>
            </Row>

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><GiftOutlined style={{ color: '#52c41a' }} /> 加分记录</Space>}>
              {driverBonuses.length > 0 ? (
                <List
                  size="small"
                  dataSource={driverBonuses}
                  renderItem={(b) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<GiftOutlined style={{ fontSize: 20, color: '#52c41a' }} />}
                        title={<Space><Text strong>{bonusTypeMap[b.bonus_type] || b.bonus_type}</Text><Tag color="green">+{b.bonus_points}分</Tag></Space>}
                        description={
                          <Space direction="vertical" size={0}>
                            <Text type="secondary">{b.reason}</Text>
                            <Text type="secondary" style={{ fontSize: 11 }}>连续{b.streak_days}天 · {b.start_date} 至 {b.end_date}</Text>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              ) : (
                <Empty description="暂无加分记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
              <div style={{ marginTop: 12, textAlign: 'center' }}>
                <Button
                  type="dashed"
                  icon={<SafetyCertificateOutlined />}
                  onClick={() => message.info('正在检查连续无违规加分资格...')}
                >
                  检查加分资格
                </Button>
              </div>
            </Card>

            {driverScoreData.latest.total_score < 60 && (
              <Card size="small" style={{ borderRadius: 8, borderColor: '#ff4d4f', background: '#fff2f0' }}>
                <Space>
                  <WarningOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
                  <div>
                    <Text strong style={{ color: '#ff4d4f' }}>评分低于60分，已自动触发再培训任务</Text>
                    <div><Text type="secondary">该驾驶员需完成安全培训并通过考核后方可继续执行运输任务</Text></div>
                  </div>
                </Space>
              </Card>
            )}
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default DrivingScore
