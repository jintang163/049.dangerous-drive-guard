import React, { useEffect, useState, useCallback } from 'react'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tabs,
  Table,
  Tag,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  Upload,
  Slider,
  Switch,
  InputNumber,
  Typography,
  message,
  Drawer,
  Descriptions,
  Empty,
  Tooltip,
  Divider,
  Alert,
  Spin,
} from 'antd'
import {
  SoundOutlined,
  TeamOutlined,
  SafetyCertificateOutlined,
  PlayCircleOutlined,
  UploadOutlined,
  StarOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  AudioOutlined,
  FireOutlined,
  ThunderboltOutlined,
  HeartOutlined,
  SettingOutlined,
  HistoryOutlined,
  PlusOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  VolumeUpOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs'
import {
  voiceApi,
  VoiceAudioItem,
  VoiceStrategyItem,
  VoiceInterventionLog,
  VoiceInterventionStats,
  AudioCategory,
  InterventionStrategyType,
} from '@/services/api'
import { hasPermission } from '@/utils/auth'

const { Title, Text } = Typography
const { Option } = Select
const { TextArea } = Input

const audioCategoryMap: Record<AudioCategory, { label: string; color: string; icon: React.ReactNode }> = {
  family: { label: '家人录音', color: '#eb2f96', icon: <HeartOutlined /> },
  custom: { label: '定制音频', color: '#1677ff', icon: <AudioOutlined /> },
  system: { label: '系统默认', color: '#8c8c8c', icon: <SoundOutlined /> },
  emergency: { label: '紧急报警', color: '#f5222d', icon: <ThunderboltOutlined /> },
}

const strategyTypeMap: Record<InterventionStrategyType, { label: string; color: string; desc: string }> = {
  normal: { label: '普通提醒', color: 'blue', desc: '首次检测到疲劳时播放' },
  continuous: { label: '连续疲劳', color: 'orange', desc: '连续疲劳超过阈值触发' },
  severe: { label: '严重疲劳', color: 'red', desc: '疲劳评分过低或危险驾驶触发' },
  emotional: { label: '情感干预', color: 'magenta', desc: '优先播放家人温馨录音' },
}

const playStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待下发', color: 'default' },
  sent: { label: '已下发', color: 'blue' },
  playing: { label: '播放中', color: 'cyan' },
  completed: { label: '已完成', color: 'green' },
  failed: { label: '失败', color: 'red' },
}

const VoiceIntervention: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<VoiceInterventionStats | null>(null)
  const [audioList, setAudioList] = useState<VoiceAudioItem[]>([])
  const [audioTotal, setAudioTotal] = useState(0)
  const [audioLoading, setAudioLoading] = useState(false)
  const [audioPage, setAudioPage] = useState(1)
  const [audioCategory, setAudioCategory] = useState<AudioCategory | ''>('')
  const [audioModalOpen, setAudioModalOpen] = useState(false)
  const [editingAudio, setEditingAudio] = useState<VoiceAudioItem | null>(null)
  const [audioForm] = Form.useForm()

  const [strategyList, setStrategyList] = useState<VoiceStrategyItem[]>([])
  const [strategyTotal, setStrategyTotal] = useState(0)
  const [strategyLoading, setStrategyLoading] = useState(false)
  const [strategyPage, setStrategyPage] = useState(1)
  const [strategyModalOpen, setStrategyModalOpen] = useState(false)
  const [editingStrategy, setEditingStrategy] = useState<VoiceStrategyItem | null>(null)
  const [strategyForm] = Form.useForm()

  const [logList, setLogList] = useState<VoiceInterventionLog[]>([])
  const [logTotal, setLogTotal] = useState(0)
  const [logLoading, setLogLoading] = useState(false)
  const [logPage, setLogPage] = useState(1)
  const [logStatus, setLogStatus] = useState<string>('')
  const [logDrawer, setLogDrawer] = useState<VoiceInterventionLog | null>(null)

  const [testModalOpen, setTestModalOpen] = useState(false)
  const [testForm] = Form.useForm()

  const isAdmin = hasPermission('admin')

  const fetchStats = useCallback(async () => {
    try {
      const res = await voiceApi.getStatistics(30)
      setStats(res)
    } catch (e) {}
  }, [])

  const fetchAudios = useCallback(async () => {
    setAudioLoading(true)
    try {
      const res = await voiceApi.listAudios({
        page: audioPage,
        page_size: 10,
        category: audioCategory || undefined,
      })
      setAudioList(res.list || [])
      setAudioTotal(res.total || 0)
    } catch (e) {} finally {
      setAudioLoading(false)
    }
  }, [audioPage, audioCategory])

  const fetchStrategies = useCallback(async () => {
    setStrategyLoading(true)
    try {
      const res = await voiceApi.listStrategies({ page: strategyPage, page_size: 10 })
      setStrategyList(res.list || [])
      setStrategyTotal(res.total || 0)
    } catch (e) {} finally {
      setStrategyLoading(false)
    }
  }, [strategyPage])

  const fetchLogs = useCallback(async () => {
    setLogLoading(true)
    try {
      const res = await voiceApi.listLogs({
        page: logPage,
        page_size: 10,
        status: logStatus || undefined,
      })
      setLogList(res.list || [])
      setLogTotal(res.total || 0)
    } catch (e) {} finally {
      setLogLoading(false)
    }
  }, [logPage, logStatus])

  useEffect(() => {
    ;(async () => {
      setLoading(true)
      await Promise.all([fetchStats(), fetchAudios(), fetchStrategies(), fetchLogs()])
      setLoading(false)
    })()
  }, [fetchStats, fetchAudios, fetchStrategies, fetchLogs])

  useEffect(() => { fetchAudios() }, [fetchAudios])
  useEffect(() => { fetchStrategies() }, [fetchStrategies])
  useEffect(() => { fetchLogs() }, [fetchLogs])

  const openCreateAudio = () => {
    setEditingAudio(null)
    audioForm.resetFields()
    setAudioModalOpen(true)
  }
  const openEditAudio = (a: VoiceAudioItem) => {
    setEditingAudio(a)
    audioForm.setFieldsValue(a)
    setAudioModalOpen(true)
  }
  const handleSaveAudio = async () => {
    try {
      const values = await audioForm.validateFields()
      if (editingAudio) {
        await voiceApi.updateAudio(editingAudio.id, values)
        message.success('音频更新成功')
      } else {
        await voiceApi.createAudio(values)
        message.success('音频创建成功')
      }
      setAudioModalOpen(false)
      fetchAudios()
    } catch (e) {}
  }
  const handleDeleteAudio = (id: number) => {
    Modal.confirm({
      title: '删除音频',
      content: '确认删除该音频吗？此操作不可恢复',
      okType: 'danger',
      onOk: async () => {
        await voiceApi.deleteAudio(id)
        message.success('删除成功')
        fetchAudios()
      },
    })
  }
  const handleSetDefault = async (a: VoiceAudioItem) => {
    await voiceApi.setDefaultAudio(a.id, { category: a.category })
    message.success('已设为默认')
    fetchAudios()
  }
  const handleTestPlay = (a: VoiceAudioItem) => {
    setEditingAudio(a)
    testForm.resetFields()
    testForm.setFieldsValue({ audio_id: a.id, audio_name: a.name, volume: a.volume })
    setTestModalOpen(true)
  }

  const openCreateStrategy = () => {
    setEditingStrategy(null)
    strategyForm.resetFields()
    strategyForm.setFieldsValue({
      strategy_type: 'normal',
      priority: 10,
      is_enabled: true,
      force_high_volume: false,
      force_volume_percent: 100,
      play_times: 1,
      play_interval_sec: 5,
      shuffle_audios: false,
      emotional_mode: false,
      cooldown_seconds: 30,
    })
    setStrategyModalOpen(true)
  }
  const openEditStrategy = (s: VoiceStrategyItem) => {
    setEditingStrategy(s)
    strategyForm.setFieldsValue({
      ...s,
      alarm_levels: s.alarm_trigger?.alarm_levels,
      alarm_types: s.alarm_trigger?.alarm_types,
      min_continuous_minutes: s.alarm_trigger?.min_continuous_minutes,
      min_fatigue_score: s.alarm_trigger?.min_fatigue_score,
    })
    setStrategyModalOpen(true)
  }
  const handleSaveStrategy = async () => {
    try {
      const values = await strategyForm.validateFields()
      const { alarm_levels, alarm_types, min_continuous_minutes, min_fatigue_score, audio_ids, ...rest } = values
      const payload: any = {
        ...rest,
        alarm_trigger: { alarm_levels, alarm_types, min_continuous_minutes, min_fatigue_score },
        audio_ids: audio_ids || [],
      }
      if (editingStrategy) {
        await voiceApi.updateStrategy(editingStrategy.id, payload)
        message.success('策略更新成功')
      } else {
        await voiceApi.createStrategy(payload)
        message.success('策略创建成功')
      }
      setStrategyModalOpen(false)
      fetchStrategies()
    } catch (e) {}
  }
  const handleDeleteStrategy = (id: number) => {
    Modal.confirm({
      title: '删除策略',
      content: '确认删除该策略吗？',
      okType: 'danger',
      onOk: async () => {
        await voiceApi.deleteStrategy(id)
        message.success('删除成功')
        fetchStrategies()
      },
    })
  }

  const handleDoTestPlay = async () => {
    try {
      const values = await testForm.validateFields()
      await voiceApi.testPlay(values)
      message.success('已下发播放指令到车端')
      setTestModalOpen(false)
    } catch (e) {}
  }

  const audioColumns = [
    { title: '音频名称', dataIndex: 'name', key: 'name', render: (n: string, r: VoiceAudioItem) => (
      <Space>
        {audioCategoryMap[r.category].icon}
        <Text strong>{n}</Text>
        {r.is_default && <Tag color="gold" icon={<StarOutlined />}>默认</Tag>}
      </Space>
    )},
    { title: '分类', dataIndex: 'category', key: 'category', render: (c: AudioCategory) => (
      <Tag color={audioCategoryMap[c].color}>{audioCategoryMap[c].label}</Tag>
    )},
    { title: '默认音量', dataIndex: 'volume', key: 'volume', render: (v: number) => (
      <Space><VolumeUpOutlined />{v}%</Space>
    )},
    { title: '时长', dataIndex: 'duration_sec', key: 'duration_sec', render: (s: number) => s ? `${s}s` : '-' },
    { title: '播放次数', dataIndex: 'play_count', key: 'play_count' },
    { title: '状态', dataIndex: 'is_enabled', key: 'is_enabled', render: (v: boolean) => v ? <Tag color="green">启用</Tag> : <Tag color="default">禁用</Tag> },
    {
      title: '操作', key: 'action', render: (_: any, r: VoiceAudioItem) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<PlayCircleOutlined />} onClick={() => handleTestPlay(r)}>试播</Button>
          {!r.is_default && <Button type="link" size="small" icon={<StarOutlined />} onClick={() => handleSetDefault(r)}>设默认</Button>}
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEditAudio(r)}>编辑</Button>
          {isAdmin && <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDeleteAudio(r.id)}>删除</Button>}
        </Space>
      ),
    },
  ]

  const strategyColumns = [
    { title: '策略名称', dataIndex: 'name', key: 'name', render: (n: string, r: VoiceStrategyItem) => (
      <Space>
        <Text strong>{n}</Text>
        {r.is_default && <Tag color="gold" icon={<StarOutlined />}>默认</Tag>}
      </Space>
    )},
    { title: '类型', dataIndex: 'strategy_type', key: 'strategy_type', render: (t: InterventionStrategyType) => (
      <Tag color={strategyTypeMap[t].color}>{strategyTypeMap[t].label}</Tag>
    )},
    { title: '优先级', dataIndex: 'priority', key: 'priority' },
    { title: '强制音量', key: 'vol', render: (_: any, r: VoiceStrategyItem) => (
      r.force_high_volume
        ? <Tag color="red">强制 {r.force_volume_percent}%</Tag>
        : <Tag color="default">跟随车机</Tag>
    )},
    { title: '播放次数', dataIndex: 'play_times', key: 'play_times', render: (t: number) => `${t}次` },
    { title: '情感模式', dataIndex: 'emotional_mode', key: 'emotional_mode', render: (v: boolean) => v ? <Tag color="magenta">家人优先</Tag> : '-' },
    { title: '状态', dataIndex: 'is_enabled', key: 'is_enabled', render: (v: boolean) => v ? <Tag color="green">启用</Tag> : <Tag color="default">禁用</Tag> },
    {
      title: '操作', key: 'action', render: (_: any, r: VoiceStrategyItem) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => openEditStrategy(r)}>详情</Button>
          {isAdmin && (
            <>
              <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEditStrategy(r)}>编辑</Button>
              <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDeleteStrategy(r.id)}>删除</Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  const logColumns = [
    { title: '时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => dayjs(t).format('MM-DD HH:mm:ss') },
    { title: '车辆', dataIndex: 'vehicle_plate', key: 'vehicle_plate', render: (p, r: VoiceInterventionLog) => p || `#${r.vehicle_id}` },
    { title: '司机', dataIndex: 'driver_name', key: 'driver_name', render: (n, r: VoiceInterventionLog) => n || `#${r.driver_id}` },
    { title: '策略类型', dataIndex: 'strategy_type', key: 'strategy_type', render: (t: InterventionStrategyType) => t ? <Tag color={strategyTypeMap[t].color}>{strategyTypeMap[t].label}</Tag> : '-' },
    { title: '播放音频', dataIndex: 'audio_name', key: 'audio_name', render: (n, r: VoiceInterventionLog) => (
      <Space>
        {audioCategoryMap[r.category]?.icon}
        {n}
      </Space>
    )},
    { title: '音量', key: 'vol', render: (_: any, r: VoiceInterventionLog) => (
      <Space>
        {r.is_high_volume && <Tag color="red">强制</Tag>}
        <Text>{r.actual_volume_percent}%</Text>
      </Space>
    )},
    { title: '播放次数', dataIndex: 'play_times', key: 'play_times' },
    { title: '疲劳评分', dataIndex: 'fatigue_score', key: 'fatigue_score', render: (s: number) => (
      <Text type={s < 60 ? 'danger' : s < 80 ? 'warning' : 'success'}>{s?.toFixed(1)}</Text>
    )},
    { title: '连续疲劳(分)', dataIndex: 'continuous_minutes', key: 'continuous_minutes' },
    { title: '司机确认', dataIndex: 'driver_ack', key: 'driver_ack', render: (v: boolean) => v ? <CheckCircleOutlined style={{ color: '#52c41a' }} /> : <CloseCircleOutlined style={{ color: '#bfbfbf' }} /> },
    { title: '状态', dataIndex: 'play_status', key: 'play_status', render: (s: string) => (
      <Tag color={playStatusMap[s]?.color || 'default'}>{playStatusMap[s]?.label || s}</Tag>
    )},
    {
      title: '操作', key: 'action', render: (_: any, r: VoiceInterventionLog) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setLogDrawer(r)}>详情</Button>
      ),
    },
  ]

  const trendChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['总干预', '高音量', '家人音频'] },
    grid: { left: 40, right: 20, top: 40, bottom: 30 },
    xAxis: { type: 'category', data: Array.from({ length: 7 }, (_, i) => dayjs().subtract(6 - i, 'day').format('MM-DD')) },
    yAxis: { type: 'value' },
    series: [
      { name: '总干预', type: 'line', smooth: true, data: [12, 18, 15, 22, 28, 19, 24], itemStyle: { color: '#1677ff' } },
      { name: '高音量', type: 'line', smooth: true, data: [3, 5, 4, 8, 12, 6, 9], itemStyle: { color: '#f5222d' } },
      { name: '家人音频', type: 'line', smooth: true, data: [5, 7, 8, 10, 14, 9, 12], itemStyle: { color: '#eb2f96' } },
    ],
  }

  const categoryPieChart = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      avoidLabelOverlap: false,
      label: { show: false },
      data: [
        { value: stats?.family_audio_count || 0, name: '家人录音', itemStyle: { color: '#eb2f96' } },
        { value: (stats?.total_count || 0) - (stats?.family_audio_count || 0) - (stats?.high_volume_count || 0), name: '定制音频', itemStyle: { color: '#1677ff' } },
        { value: stats?.high_volume_count || 0, name: '强制高音量', itemStyle: { color: '#f5222d' } },
      ],
    }],
  }

  return (
    <div style={{ padding: 16 }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Card bordered={false} style={{ borderRadius: 12 }}>
          <Row gutter={16} align="middle">
            <Col>
              <Space size={12}>
                <div style={{ width: 48, height: 48, borderRadius: 12, background: 'linear-gradient(135deg,#1677ff,#722ed1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <SoundOutlined style={{ color: '#fff', fontSize: 24 }} />
                </div>
                <div>
                  <Title level={4} style={{ margin: 0 }}>语音疲劳干预</Title>
                  <Text type="secondary">个性化语音提醒 + 连续疲劳强制高音量报警</Text>
                </div>
              </Space>
            </Col>
            <Col flex="auto" />
            <Col>
              <Space>
                <Button icon={<ReloadOutlined />} onClick={() => { fetchStats(); fetchAudios(); fetchStrategies(); fetchLogs() }}>刷新</Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={openCreateAudio}>上传音频</Button>
                {isAdmin && <Button icon={<SettingOutlined />} onClick={openCreateStrategy}>新建策略</Button>}
              </Space>
            </Col>
          </Row>
        </Card>

        {loading ? (
          <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>
        ) : (
          <>
            <Row gutter={16}>
              <Col xs={12} md={6}>
                <Card bordered={false} style={{ borderRadius: 12 }}>
                  <Statistic title="近30天干预次数" value={stats?.total_count || 0} prefix={<SoundOutlined />} />
                </Card>
              </Col>
              <Col xs={12} md={6}>
                <Card bordered={false} style={{ borderRadius: 12 }}>
                  <Statistic title="强制高音量触发" value={stats?.high_volume_count || 0} prefix={<FireOutlined style={{ color: '#f5222d' }} />} valueStyle={{ color: '#f5222d' }} />
                </Card>
              </Col>
              <Col xs={12} md={6}>
                <Card bordered={false} style={{ borderRadius: 12 }}>
                  <Statistic title="家人音频使用" value={stats?.family_audio_count || 0} prefix={<HeartOutlined style={{ color: '#eb2f96' }} />} valueStyle={{ color: '#eb2f96' }} />
                </Card>
              </Col>
              <Col xs={12} md={6}>
                <Card bordered={false} style={{ borderRadius: 12 }}>
                  <Statistic title="司机确认率" value={stats?.driver_ack_rate || '0%'} prefix={<SafetyCertificateOutlined />} />
                </Card>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col xs={24} lg={14}>
                <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><SoundOutlined style={{ color: '#1677ff' }} /> 干预趋势（近7天）</Space>}>
                  <ReactECharts option={trendChart} style={{ height: 280 }} />
                </Card>
              </Col>
              <Col xs={24} lg={10}>
                <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><SafetyCertificateOutlined /> 音频分布</Space>}>
                  <ReactECharts option={categoryPieChart} style={{ height: 280 }} />
                </Card>
              </Col>
            </Row>

            <Alert
              type="info"
              showIcon
              icon={<ThunderboltOutlined />}
              message="干预策略说明"
              description={
                <ul style={{ margin: '4px 0 0 20px', padding: 0 }}>
                  <li><b>普通提醒</b>：首次检测疲劳，播放系统或家人标准提醒（音量70%）</li>
                  <li><b>连续疲劳</b>：连续疲劳超过10分钟，<b>强制最高音量</b>反复播放家人+系统报警（音量100%）</li>
                  <li><b>严重疲劳</b>：评分低于40，强制最高音量重复播放刺耳报警（100%音量，重复5次）</li>
                  <li><b>情感模式</b>：优先匹配家人（妻子/孩子/父母）录制的温馨语音，提升干预效果</li>
                </ul>
              }
              style={{ borderRadius: 12 }}
            />

            <Card bordered={false} style={{ borderRadius: 12 }}>
              <Tabs
                defaultActiveKey="audios"
                items={[
                  {
                    key: 'audios',
                    label: <Space><AudioOutlined /> 个性化音频库</Space>,
                    children: (
                      <div>
                        <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Space>
                            <Text type="secondary">分类筛选：</Text>
                            <Select value={audioCategory} style={{ width: 140 }} allowClear placeholder="全部分类" onChange={v => { setAudioCategory(v); setAudioPage(1) }}>
                              {(Object.keys(audioCategoryMap) as AudioCategory[]).map(k => (
                                <Option key={k} value={k}>{audioCategoryMap[k].icon} {audioCategoryMap[k].label}</Option>
                              ))}
                            </Select>
                          </Space>
                          <Space>
                            <Upload
                              showUploadList={false}
                              accept=".mp3,.wav,.m4a,.ogg"
                              customRequest={async ({ file }) => {
                                try {
                                  const extra = { name: (file as File).name.replace(/\.[^.]+$/, ''), category: 'custom' as AudioCategory }
                                  await voiceApi.uploadAudio(file as File, extra)
                                  message.success('音频上传成功')
                                  fetchAudios()
                                } catch (e) {}
                              }}
                            >
                              <Button type="primary" icon={<UploadOutlined />}>上传音频文件</Button>
                            </Upload>
                            <Button icon={<PlusOutlined />} onClick={openCreateAudio}>手动创建</Button>
                          </Space>
                        </div>
                        <Table
                          rowKey="id"
                          columns={audioColumns}
                          dataSource={audioList}
                          loading={audioLoading}
                          pagination={{ current: audioPage, total: audioTotal, pageSize: 10, onChange: p => setAudioPage(p), showTotal: t => `共 ${t} 条音频` }}
                          locale={{ emptyText: <Empty description="暂无音频，请上传家人录音或定制音频" /> }}
                        />
                      </div>
                    ),
                  },
                  {
                    key: 'strategies',
                    label: <Space><SettingOutlined /> 干预策略</Space>,
                    children: (
                      <div>
                        <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Text type="secondary">按优先级匹配，数字越小越优先</Text>
                          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateStrategy}>新建策略</Button>
                        </div>
                        <Table
                          rowKey="id"
                          columns={strategyColumns}
                          dataSource={strategyList}
                          loading={strategyLoading}
                          pagination={{ current: strategyPage, total: strategyTotal, pageSize: 10, onChange: p => setStrategyPage(p), showTotal: t => `共 ${t} 条策略` }}
                          locale={{ emptyText: <Empty description="暂无策略" /> }}
                        />
                      </div>
                    ),
                  },
                  {
                    key: 'logs',
                    label: <Space><HistoryOutlined /> 干预日志</Space>,
                    children: (
                      <div>
                        <div style={{ marginBottom: 12 }}>
                          <Space>
                            <Text type="secondary">状态：</Text>
                            <Select value={logStatus} style={{ width: 140 }} allowClear placeholder="全部" onChange={v => { setLogStatus(v); setLogPage(1) }}>
                              {Object.keys(playStatusMap).map(k => (
                                <Option key={k} value={k}>{playStatusMap[k].label}</Option>
                              ))}
                            </Select>
                          </Space>
                        </div>
                        <Table
                          rowKey="id"
                          columns={logColumns}
                          dataSource={logList}
                          loading={logLoading}
                          pagination={{ current: logPage, total: logTotal, pageSize: 10, onChange: p => setLogPage(p), showTotal: t => `共 ${t} 条` }}
                          scroll={{ x: 1400 }}
                          locale={{ emptyText: <Empty description="暂无干预记录" /> }}
                        />
                      </div>
                    ),
                  },
                ]}
              />
            </Card>
          </>
        )}
      </Space>

      <Modal
        title={editingAudio ? '编辑音频' : '新建音频'}
        open={audioModalOpen}
        onOk={handleSaveAudio}
        onCancel={() => setAudioModalOpen(false)}
        width={560}
        destroyOnClose
      >
        <Form form={audioForm} layout="vertical">
          <Form.Item name="name" label="音频名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：家人提醒-妻子、系统报警音" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="category" label="分类" rules={[{ required: true }]}>
                <Select>
                  {(Object.keys(audioCategoryMap) as AudioCategory[]).map(k => (
                    <Option key={k} value={k}>{audioCategoryMap[k].label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="volume" label="默认音量(%)" rules={[{ required: true }]}>
                <Slider min={0} max={100} marks={{ 0: '0', 50: '50', 100: '100' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="audio_url" label="音频URL" rules={[{ required: true }]}>
            <Input placeholder="/audios/family/wife_reminder.mp3" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="duration_sec" label="时长(秒)">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="is_enabled" label="是否启用" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="录制人、适用场景等说明" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingStrategy ? '编辑干预策略' : '新建干预策略'}
        open={strategyModalOpen}
        onOk={handleSaveStrategy}
        onCancel={() => setStrategyModalOpen(false)}
        width={720}
        destroyOnClose
      >
        <Form form={strategyForm} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="策略名称" rules={[{ required: true }]}>
                <Input placeholder="如：连续疲劳强制高音量" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="strategy_type" label="策略类型" rules={[{ required: true }]}>
                <Select>
                  {(Object.keys(strategyTypeMap) as InterventionStrategyType[]).map(k => (
                    <Option key={k} value={k}>{strategyTypeMap[k].label} - {strategyTypeMap[k].desc}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Divider orientation="left">触发条件</Divider>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="alarm_levels" label="报警等级">
                <Select mode="multiple" placeholder="不选=全部等级">
                  <Option value={1}>1-提醒</Option>
                  <Option value={2}>2-警告</Option>
                  <Option value={3}>3-严重</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="alarm_types" label="报警类型">
                <Select mode="multiple" placeholder="不选=全部类型">
                  <Option value="fatigue_perclos">疲劳瞌睡</Option>
                  <Option value="continuous_fatigue_perclos">连续疲劳</Option>
                  <Option value="excessive_yawn">频繁打哈欠</Option>
                  <Option value="abnormal_head_posture">异常姿态</Option>
                  <Option value="eyes_closed">闭眼检测</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="min_continuous_minutes" label="最小连续疲劳(分钟)">
                <InputNumber min={0} style={{ width: '100%' }} placeholder="0=不限制" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="min_fatigue_score" label="疲劳评分低于">
                <InputNumber min={0} max={100} style={{ width: '100%' }} placeholder="例如：40，表示评分<40时触发" />
              </Form.Item>
            </Col>
          </Row>
          <Divider orientation="left">播放配置</Divider>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Item name="audio_ids" label="要播放的音频">
                <Select mode="multiple" placeholder="选择音频，支持多选随机">
                  {audioList.map(a => (
                    <Option key={a.id} value={a.id}>{audioCategoryMap[a.category].label} - {a.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="play_times" label="重复播放次数">
                <InputNumber min={1} max={10} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="play_interval_sec" label="重复间隔(秒)">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="shuffle_audios" label="随机播放" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="force_high_volume" label="强制高音量" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="force_volume_percent" label="强制音量(%)">
                <Slider min={50} max={100} marks={{ 50: '50', 75: '75', 100: '100' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="emotional_mode" label="情感模式(家人优先)" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="cooldown_seconds" label="冷却时间(秒)">
                <InputNumber min={0} max={3600} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
                <InputNumber min={1} max={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="is_enabled" label="启用" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="is_default" label="默认策略" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="说明">
            <TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title="干预详情"
        width={560}
        open={!!logDrawer}
        onClose={() => setLogDrawer(null)}
        destroyOnClose
      >
        {logDrawer && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="干预时间">{dayjs(logDrawer.created_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
              <Descriptions.Item label="车辆">{logDrawer.vehicle_plate || `#${logDrawer.vehicle_id}`}</Descriptions.Item>
              <Descriptions.Item label="司机">{logDrawer.driver_name || `#${logDrawer.driver_id}`}</Descriptions.Item>
              <Descriptions.Item label="策略类型">
                <Tag color={strategyTypeMap[logDrawer.strategy_type]?.color}>
                  {strategyTypeMap[logDrawer.strategy_type]?.label || logDrawer.strategy_type}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="播放状态">
                <Tag color={playStatusMap[logDrawer.play_status]?.color}>
                  {playStatusMap[logDrawer.play_status]?.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="播放音频">
                <Space>
                  {audioCategoryMap[logDrawer.category]?.icon}
                  {logDrawer.audio_name}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="音量">
                <Space>
                  {logDrawer.is_high_volume && <Tag color="red">强制</Tag>}
                  {logDrawer.actual_volume_percent}%
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="播放次数">{logDrawer.play_times} 次</Descriptions.Item>
              <Descriptions.Item label="总播放时长">{(logDrawer.total_play_duration_ms / 1000).toFixed(1)} 秒</Descriptions.Item>
              <Descriptions.Item label="疲劳评分">
                <Text type={logDrawer.fatigue_score < 60 ? 'danger' : 'success'}>{logDrawer.fatigue_score.toFixed(1)}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="连续疲劳">{logDrawer.continuous_minutes} 分钟</Descriptions.Item>
              <Descriptions.Item label="报警类型">{logDrawer.alarm_type}</Descriptions.Item>
              <Descriptions.Item label="司机确认">
                {logDrawer.driver_ack
                  ? <Space><CheckCircleOutlined style={{ color: '#52c41a' }} /> 已确认 ({dayjs(logDrawer.ack_at).format('HH:mm:ss')})</Space>
                  : <Space><CloseCircleOutlined style={{ color: '#bfbfbf' }} /> 未确认</Space>}
              </Descriptions.Item>
              {logDrawer.error_msg && <Descriptions.Item label="错误信息"><Text type="danger">{logDrawer.error_msg}</Text></Descriptions.Item>}
            </Descriptions>
            <Tooltip title="点击播放预览">
              <Button block type="dashed" icon={<PlayCircleOutlined />}>
                播放：{logDrawer.audio_name}
              </Button>
            </Tooltip>
          </Space>
        )}
      </Drawer>

      <Modal title="车端试播" open={testModalOpen} onOk={handleDoTestPlay} onCancel={() => setTestModalOpen(false)} destroyOnClose>
        <Form form={testForm} layout="vertical">
          <Form.Item name="audio_name" label="音频">
            <Input disabled />
          </Form.Item>
          <Form.Item name="audio_id" hidden><Input /></Form.Item>
          <Form.Item name="vehicle_id" label="目标车辆" rules={[{ required: true, message: '请选择车辆' }]}>
            <Select placeholder="选择要试播的车辆" showSearch optionFilterProp="label">
              <Option value={1} label="京A·危12345">京A·危12345</Option>
              <Option value={2} label="京B·危67890">京B·危67890</Option>
              <Option value={3} label="京C·危11111">京C·危11111</Option>
            </Select>
          </Form.Item>
          <Form.Item name="volume" label="播放音量(%)" rules={[{ required: true }]}>
            <Slider min={0} max={100} marks={{ 0: '0', 50: '50', 100: '100' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default VoiceIntervention
