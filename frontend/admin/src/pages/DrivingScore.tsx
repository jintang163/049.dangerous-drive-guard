import React, { useEffect, useState, useMemo, useCallback } from 'react'
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
  Alert,
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
  LockOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs'
import { scoreApi, ScoreOverview, ScoreLeaderboardItem, DrivingScoreRecord, ScoreBonus, MonthlyReport, RetrainingTask } from '@/services/api'
import { hasPermission, getUserInfo } from '@/utils/auth'

const { Title, Text } = Typography
const { Option } = Select

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

const defaultOverview: ScoreOverview = {
  total_drivers: 0,
  excellent_count: 0,
  good_count: 0,
  normal_count: 0,
  poor_count: 0,
  danger_count: 0,
  avg_score: 0,
  retraining_count: 0,
  today_calculated: 0,
  need_retraining_count: 0,
}

const DrivingScore: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [overview, setOverview] = useState<ScoreOverview>(defaultOverview)
  const [leaderboard, setLeaderboard] = useState<ScoreLeaderboardItem[]>([])
  const [leaderboardLoading, setLeaderboardLoading] = useState(false)
  const [detailDrawer, setDetailDrawer] = useState<ScoreLeaderboardItem | null>(null)
  const [driverScoreData, setDriverScoreData] = useState<{ latest: DrivingScoreRecord; history: DrivingScoreRecord[] } | null>(null)
  const [driverScoreLoading, setDriverScoreLoading] = useState(false)
  const [driverBonuses, setDriverBonuses] = useState<ScoreBonus[]>([])
  const [monthlyReports, setMonthlyReports] = useState<MonthlyReport[]>([])
  const [monthlyReportsLoading, setMonthlyReportsLoading] = useState(false)
  const [retrainingTasks, setRetrainingTasks] = useState<RetrainingTask[]>([])
  const [retrainingLoading, setRetrainingLoading] = useState(false)
  const [reportMonth, setReportMonth] = useState(dayjs().format('YYYY-MM'))
  const [orderBy, setOrderBy] = useState('total_score')
  const isAdmin = hasPermission('admin') || hasPermission('score:manage')
  const currentUser = getUserInfo()

  const fetchOverview = useCallback(async () => {
    try {
      const res = await scoreApi.getOverview()
      setOverview(res)
    } catch (e) {
      message.error('获取评分概览失败')
    }
  }, [])

  const fetchLeaderboard = useCallback(async () => {
    if (!isAdmin) return
    setLeaderboardLoading(true)
    try {
      const res = await scoreApi.getLeaderboard({ top: 20, order_by: orderBy })
      setLeaderboard(res)
    } catch (e) {
      message.error('获取排行榜失败')
    } finally {
      setLeaderboardLoading(false)
    }
  }, [isAdmin, orderBy])

  const fetchMonthlyReports = useCallback(async () => {
    if (!isAdmin) return
    setMonthlyReportsLoading(true)
    try {
      const res = await scoreApi.listMonthlyReports({ month: reportMonth, page: 1, page_size: 20 })
      setMonthlyReports(res.list || [])
    } catch (e) {
      message.error('获取月报列表失败')
    } finally {
      setMonthlyReportsLoading(false)
    }
  }, [isAdmin, reportMonth])

  const fetchRetrainingTasks = useCallback(async () => {
    if (!isAdmin) return
    setRetrainingLoading(true)
    try {
      const res = await scoreApi.listRetrainingTasks({ page: 1, page_size: 20 })
      setRetrainingTasks(res.list || [])
    } catch (e) {
      message.error('获取再培训任务失败')
    } finally {
      setRetrainingLoading(false)
    }
  }, [isAdmin])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      await fetchOverview()
      setLoading(false)
    }
    init()
  }, [fetchOverview])

  useEffect(() => {
    if (isAdmin) {
      fetchLeaderboard()
      fetchMonthlyReports()
      fetchRetrainingTasks()
    }
  }, [isAdmin, fetchLeaderboard, fetchMonthlyReports, fetchRetrainingTasks])

  const handleViewDetail = async (item: ScoreLeaderboardItem) => {
    setDetailDrawer(item)
    setDriverScoreLoading(true)
    try {
      const scoreRes = await scoreApi.getDriverScore(item.driver_id, { days: 30 })
      setDriverScoreData(scoreRes)
      const bonusRes = await scoreApi.getDriverBonuses(item.driver_id)
      setDriverBonuses(bonusRes)
    } catch (e) {
      message.error('获取驾驶员评分详情失败')
    } finally {
      setDriverScoreLoading(false)
    }
  }

  const handleCheckBonus = async (driverId: number) => {
    try {
      const res = await scoreApi.checkBonus(driverId)
      if (res.awarded) {
        message.success(`获得加分：+${res.bonus?.bonus_points}分`)
        const bonusRes = await scoreApi.getDriverBonuses(driverId)
        setDriverBonuses(bonusRes)
      } else {
        message.info(res.message || '未达到连续30天无违规标准')
      }
    } catch (e) {
      message.error('检查加分资格失败')
    }
  }

  const handleSendReport = async (reportId: number) => {
    try {
      await scoreApi.sendMonthlyReport(reportId)
      message.success('月报已推送至司机和管理员邮箱')
      fetchMonthlyReports()
    } catch (e) {
      message.error('推送月报失败')
    }
  }

  const handleBatchSend = async () => {
    try {
      const res = await scoreApi.batchSendMonthlyReports({ month: reportMonth })
      message.success(`批量推送完成：成功 ${res.sent} 封，失败 ${res.failed} 封`)
      fetchMonthlyReports()
    } catch (e) {
      message.error('批量推送月报失败')
    }
  }

  const handleUpdateRetraining = async (taskId: number, status: string) => {
    try {
      await scoreApi.updateRetrainingTask(taskId, { status })
      message.success(status === 'in_progress' ? '已开始培训' : '培训已完成')
      fetchRetrainingTasks()
    } catch (e) {
      message.error('更新培训状态失败')
    }
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
          onClick={() => handleSendReport(r.id)}
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
            <Button type="link" size="small" onClick={() => handleUpdateRetraining(r.id, 'in_progress')}>开始</Button>
          )}
          {r.status === 'in_progress' && (
            <Button type="link" size="small" onClick={() => handleUpdateRetraining(r.id, 'completed')}>完成</Button>
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
        {isAdmin && (
          <Col xs={24} lg={14}>
            <Card
              bordered={false}
              style={{ borderRadius: 12 }}
              title={
                <Space>
                  <TrophyOutlined style={{ color: '#faad14' }} />
                  <Text strong style={{ fontSize: 15 }}>评分排行榜</Text>
                  <Tag color="blue" style={{ margin: 0 }}>车队内排名</Tag>
                  <Tag color="red" style={{ margin: 0 }} icon={<LockOutlined />}>仅管理员可见</Tag>
                </Space>
              }
              extra={
                <Space>
                  <Select value={orderBy} style={{ width: 120 }} size="small" onChange={v => setOrderBy(v)}>
                    <Option value="total_score">按今日分</Option>
                    <Option value="avg_score_30d">按30日均</Option>
                  </Select>
                  <Button icon={<ReloadOutlined />} size="small" onClick={() => fetchLeaderboard()}>刷新</Button>
                </Space>
              }
            >
              <Table
                rowKey="driver_id"
                columns={leaderboardColumns}
                dataSource={leaderboard}
                loading={leaderboardLoading}
                pagination={{ pageSize: 10, showTotal: t => `共 ${t} 人` }}
                size="small"
                scroll={{ y: 420 }}
                locale={{ emptyText: <Empty description="暂无评分数据" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
              />
            </Card>
          </Col>
        )}
        <Col xs={24} lg={isAdmin ? 10 : 24} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
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

      {isAdmin && (
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
                          onClick={handleBatchSend}
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
                      loading={monthlyReportsLoading}
                      pagination={{ pageSize: 10, showTotal: t => `共 ${t} 条` }}
                      size="small"
                      locale={{ emptyText: <Empty description="暂无月报数据" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
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
                        <Button icon={<ReloadOutlined />} size="small" onClick={() => fetchRetrainingTasks()}>刷新</Button>
                      </Space>
                    </div>
                    <Table
                      rowKey="id"
                      columns={retrainingColumns}
                      dataSource={retrainingTasks}
                      loading={retrainingLoading}
                      pagination={{ pageSize: 10 }}
                      size="small"
                      locale={{ emptyText: <Empty description="暂无再培训任务" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                    />
                  </div>
                ),
              },
            ]}
          />
        </Card>
      )}

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
                  onClick={() => detailDrawer && handleCheckBonus(detailDrawer.driver_id)}
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
