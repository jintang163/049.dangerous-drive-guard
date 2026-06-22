import React, { useEffect, useState, useCallback } from 'react'
import {
  Row, Col, Card, Tag, Button, Space, Typography, Select, Input,
  Statistic, Drawer, Descriptions, message, Tabs, Table, Modal,
  Form, Divider, Alert, List,
} from 'antd'
import {
  SafetyOutlined, SearchOutlined, ReloadOutlined, EyeOutlined,
  FileTextOutlined, ThunderboltOutlined, AppstoreOutlined,
  CheckCircleOutlined, CloseCircleOutlined,
  WarningOutlined, ExperimentOutlined, FireOutlined,
  EnvironmentOutlined, PhoneOutlined, MedicineBoxOutlined,
  SafetyCertificateOutlined, EditOutlined, CarOutlined, UserOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { emergencyApi } from '@/services/api'
import type { EmergencyPlanItem, EmergencyTaskCard, EmergencyStats } from '@/services/api'
import { useAppStore } from '@/store/app'
import { formatDateTime } from '@/utils/auth'

const { Text, Paragraph } = Typography

const DANGER_CLASS_MAP: Record<string, { label: string; color: string }> = {
  '1': { label: '爆炸品', color: '#ff4d4f' },
  '2': { label: '气体', color: '#fa8c16' },
  '3': { label: '易燃液体', color: '#faad14' },
  '4': { label: '易燃固体', color: '#eb2f96' },
  '5': { label: '氧化剂', color: '#722ed1' },
  '6': { label: '毒性物质', color: '#13c2c2' },
  '7': { label: '放射性物质', color: '#f5222d' },
  '8': { label: '腐蚀性物质', color: '#1677ff' },
  '9': { label: '杂类危险品', color: '#8c8c8c' },
}

const PUSH_STATUS_MAP: Record<string, { label: string; color: string }> = {
  pending: { label: '待推送', color: 'default' },
  pushed: { label: '已推送', color: 'blue' },
  acknowledged: { label: '已确认', color: 'green' },
  expired: { label: '已过期', color: 'red' },
}

const CARD_STATUS_MAP: Record<string, { label: string; color: string }> = {
  active: { label: '活跃', color: 'blue' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'default' },
  expired: { label: '已过期', color: 'red' },
}

const PLAN_STATUS_MAP: Record<string, { label: string; color: string }> = {
  active: { label: '启用', color: 'green' },
  draft: { label: '草稿', color: 'default' },
  archived: { label: '归档', color: 'orange' },
}

const parseJsonField = (val: any): any[] => {
  if (!val) return []
  if (Array.isArray(val)) return val
  try { return JSON.parse(val) } catch { return [] }
}

const formatSteps = (text: string) => {
  if (!text) return text
  return text.split(/(?=\d+[.、)])/g).filter(Boolean)
}

const EmergencyPlan: React.FC = () => {
  const { vehicleList, driverList, waybillList } = useAppStore()
  const [activeTab, setActiveTab] = useState('plans')
  const [planLoading, setPlanLoading] = useState(false)
  const [plans, setPlans] = useState<EmergencyPlanItem[]>([])
  const [planTotal, setPlanTotal] = useState(0)
  const [planPage, setPlanPage] = useState(1)
  const [planPageSize, setPlanPageSize] = useState(10)
  const [selectedPlan, setSelectedPlan] = useState<EmergencyPlanItem | null>(null)
  const [planDrawerOpen, setPlanDrawerOpen] = useState(false)

  const [taskCardLoading, setTaskCardLoading] = useState(false)
  const [taskCards, setTaskCards] = useState<EmergencyTaskCard[]>([])
  const [taskCardTotal, setTaskCardTotal] = useState(0)
  const [taskCardPage, setTaskCardPage] = useState(1)
  const [taskCardPageSize, setTaskCardPageSize] = useState(10)
  const [selectedTaskCard, setSelectedTaskCard] = useState<EmergencyTaskCard | null>(null)
  const [taskCardDrawerOpen, setTaskCardDrawerOpen] = useState(false)

  const [generateModalOpen, setGenerateModalOpen] = useState(false)
  const [generateForm] = Form.useForm()
  const [generating, setGenerating] = useState(false)

  const [stats, setStats] = useState<EmergencyStats>({
    total_plans: 0, builtin_plans: 0, custom_plans: 0, active_task_cards: 0,
    danger_class_distribution: [],
  })

  const [filters, setFilters] = useState({
    un_number: '', danger_class: undefined as string | undefined, keyword: '',
  })

  const fetchStats = useCallback(async () => {
    try {
      const data = await emergencyApi.getStats()
      setStats(data || { total_plans: 0, builtin_plans: 0, custom_plans: 0, active_task_cards: 0, danger_class_distribution: [] })
    } catch { /* ignore */ }
  }, [])

  const fetchPlans = useCallback(async (page = planPage, pageSize = planPageSize) => {
    setPlanLoading(true)
    try {
      const data = await emergencyApi.listPlans({
        page, page_size: pageSize,
        un_number: filters.un_number || undefined,
        danger_class: filters.danger_class || undefined,
        keyword: filters.keyword || undefined,
      })
      setPlans(data?.list || [])
      setPlanTotal(data?.total || 0)
      setPlanPage(page)
      setPlanPageSize(pageSize)
    } catch { /* ignore */ }
    setPlanLoading(false)
  }, [planPage, planPageSize, filters])

  const fetchTaskCards = useCallback(async (page = taskCardPage, pageSize = taskCardPageSize) => {
    setTaskCardLoading(true)
    try {
      const data = await emergencyApi.listTaskCards({ page, page_size: pageSize })
      setTaskCards(data?.list || [])
      setTaskCardTotal(data?.total || 0)
      setTaskCardPage(page)
      setTaskCardPageSize(pageSize)
    } catch { /* ignore */ }
    setTaskCardLoading(false)
  }, [taskCardPage, taskCardPageSize])

  useEffect(() => { fetchStats() }, [])
  useEffect(() => { if (activeTab === 'plans') fetchPlans() }, [activeTab, fetchPlans])
  useEffect(() => { if (activeTab === 'tasks') fetchTaskCards() }, [activeTab, fetchTaskCards])

  const handleSearch = () => { fetchPlans(1, planPageSize) }
  const handleReset = () => {
    setFilters({ un_number: '', danger_class: undefined, keyword: '' })
    setPlans([])
    setPlanTotal(0)
    setTimeout(() => fetchPlans(1, planPageSize), 0)
  }

  const handleViewPlan = async (record: EmergencyPlanItem) => {
    if (record.id) {
      try {
        const detail = await emergencyApi.getPlan(record.id)
        setSelectedPlan(detail || record)
      } catch { setSelectedPlan(record) }
    } else {
      setSelectedPlan(record)
    }
    setPlanDrawerOpen(true)
  }

  const handleOpenGenerateModal = (plan: EmergencyPlanItem) => {
    generateForm.resetFields()
    generateForm.setFieldsValue({ push_channels: ['app'] })
    setSelectedPlan(plan)
    setGenerateModalOpen(true)
  }

  const handleGenerate = async () => {
    try {
      const values = await generateForm.validateFields()
      setGenerating(true)
      const result = await emergencyApi.generateTaskCard({
        plan_id: selectedPlan!.id,
        vehicle_id: values.vehicle_id,
        driver_id: values.driver_id,
        waybill_id: values.waybill_id || undefined,
        push_channels: values.push_channels || ['app'],
      })
      message.success('任务卡生成成功')
      setGenerateModalOpen(false)
      setPlanDrawerOpen(false)
      setActiveTab('tasks')
      fetchTaskCards(1)
      fetchStats()
    } catch (e: any) {
      if (e?.errorFields) return
      message.error(e?.message || '生成任务卡失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleTaskCardAction = async (id: number, action: 'ack' | 'complete' | 'cancel') => {
    try {
      if (action === 'ack') await emergencyApi.ackTaskCard(id)
      else if (action === 'complete') await emergencyApi.completeTaskCard(id)
      else await emergencyApi.cancelTaskCard(id)
      message.success('操作成功')
      fetchTaskCards()
      fetchStats()
    } catch (e: any) {
      message.error(e?.message || '操作失败')
    }
  }

  const pieOption = {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, type: 'scroll', textStyle: { fontSize: 11 } },
    series: [{
      type: 'pie', radius: ['40%', '70%'], center: ['50%', '45%'],
      label: { show: false },
      data: (stats.danger_class_distribution || []).map((d: any) => ({
        name: DANGER_CLASS_MAP[d.danger_class]?.label || `类${d.danger_class}`,
        value: d.count,
        itemStyle: { color: DANGER_CLASS_MAP[d.danger_class]?.color },
      })),
    }],
  }

  const planColumns = [
    {
      title: 'UN编号', dataIndex: 'un_number', width: 120,
      render: (v: string, r: EmergencyPlanItem) => (
        <Space>
          <Tag color={DANGER_CLASS_MAP[r.danger_class]?.color || 'default'} style={{ margin: 0 }}>{v}</Tag>
        </Space>
      ),
    },
    { title: '正确运输名称', dataIndex: 'proper_shipping_name_cn', width: 200, ellipsis: true },
    {
      title: '危险类别', dataIndex: 'danger_class', width: 120,
      render: (v: string) => {
        const m = DANGER_CLASS_MAP[v]
        return m ? <Tag color={m.color}>{m.label}</Tag> : <Tag>{v}</Tag>
      },
    },
    {
      title: '包装组', dataIndex: 'packing_group', width: 80,
      render: (v: string) => v ? <Tag>{v}</Tag> : <Text type="secondary">-</Text>,
    },
    {
      title: '来源', dataIndex: 'source', width: 90,
      render: (v: string) => v === 'builtin' ? <Tag color="blue">内置</Tag> : <Tag color="orange">自定义</Tag>,
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => {
        const m = PLAN_STATUS_MAP[v]
        return m ? <Tag color={m.color}>{m.label}</Tag> : <Tag>{v}</Tag>
      },
    },
    {
      title: '操作', width: 100, fixed: 'right' as const,
      render: (_: any, record: EmergencyPlanItem) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewPlan(record)}>查看详情</Button>
      ),
    },
  ]

  const taskCardColumns = [
    { title: '任务卡编号', dataIndex: 'card_no', width: 150, render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text> },
    { title: 'UN编号', dataIndex: 'un_number', width: 100 },
    { title: '车牌号', dataIndex: 'plate_number', width: 110, render: (v: string) => <Tag color="blue">{v}</Tag> },
    { title: '司机名称', dataIndex: 'driver_name', width: 100 },
    { title: '标题', dataIndex: 'title', width: 180, ellipsis: true },
    {
      title: '推送状态', dataIndex: 'push_status', width: 100,
      render: (v: string) => { const m = PUSH_STATUS_MAP[v]; return m ? <Tag color={m.color}>{m.label}</Tag> : <Tag>{v}</Tag> },
    },
    {
      title: '任务卡状态', dataIndex: 'card_status', width: 100,
      render: (v: string) => { const m = CARD_STATUS_MAP[v]; return m ? <Tag color={m.color}>{m.label}</Tag> : <Tag>{v}</Tag> },
    },
    {
      title: '创建时间', dataIndex: 'created_at', width: 160,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', width: 180, fixed: 'right' as const,
      render: (_: any, record: EmergencyTaskCard) => (
        <Space size={4}>
          {record.card_status === 'active' && record.push_status === 'pushed' && (
            <Button type="link" size="small" icon={<CheckCircleOutlined />} onClick={() => handleTaskCardAction(record.id, 'ack')}>确认</Button>
          )}
          {(record.card_status === 'active') && (
            <Button type="link" size="small" style={{ color: '#52c41a' }} icon={<CheckCircleOutlined />} onClick={() => handleTaskCardAction(record.id, 'complete')}>完成</Button>
          )}
          {(record.card_status === 'active') && (
            <Button type="link" size="small" danger icon={<CloseCircleOutlined />} onClick={() => handleTaskCardAction(record.id, 'cancel')}>取消</Button>
          )}
        </Space>
      ),
    },
  ]

  const renderPlanDetail = (plan: EmergencyPlanItem) => {
    const equipList = parseJsonField(plan.protective_equipment)
    const contactsList = parseJsonField(plan.emergency_contacts)
    const leakSteps = formatSteps(plan.leak_disposal)

    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Card size="small" title={<Space><SafetyCertificateOutlined />基本信息</Space>} style={{ borderRadius: 8 }}>
          <Descriptions column={2} size="small" bordered>
            <Descriptions.Item label="UN编号"><Text strong>{plan.un_number}</Text></Descriptions.Item>
            <Descriptions.Item label="危险类别">
              {DANGER_CLASS_MAP[plan.danger_class] ? <Tag color={DANGER_CLASS_MAP[plan.danger_class].color}>{DANGER_CLASS_MAP[plan.danger_class].label}</Tag> : plan.danger_class}
            </Descriptions.Item>
            <Descriptions.Item label="运输名称(中)" span={2}>{plan.proper_shipping_name_cn}</Descriptions.Item>
            <Descriptions.Item label="运输名称(英)" span={2}>{plan.proper_shipping_name_en}</Descriptions.Item>
            {plan.subsidiary_danger && <Descriptions.Item label="次要危险" span={2}>{plan.subsidiary_danger}</Descriptions.Item>}
            {plan.packing_group && <Descriptions.Item label="包装组">{plan.packing_group}</Descriptions.Item>}
            <Descriptions.Item label="来源">{plan.source === 'builtin' ? <Tag color="blue">内置</Tag> : <Tag color="orange">自定义</Tag>}</Descriptions.Item>
          </Descriptions>
        </Card>

        {plan.hazard_summary && (
          <Card size="small" title={<Space><WarningOutlined style={{ color: '#ff4d4f' }} />危险性概述</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.hazard_summary}</Paragraph>
          </Card>
        )}

        {plan.leak_disposal && (
          <Card size="small" title={<Space><ExperimentOutlined style={{ color: '#fa8c16' }} />泄漏处置方法</Space>} style={{ borderRadius: 8 }}>
            {leakSteps.length > 1 ? (
              <List size="small" dataSource={leakSteps} renderItem={(step, idx) => (
                <List.Item style={{ padding: '6px 0', border: 'none' }}>
                  <Space align="start">
                    <Tag color="orange" style={{ borderRadius: '50%', width: 24, height: 24, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', padding: 0 }}>{idx + 1}</Tag>
                    <Text style={{ whiteSpace: 'pre-wrap' }}>{step.replace(/^\d+[.、)]\s*/, '')}</Text>
                  </Space>
                </List.Item>
              )} />
            ) : <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.leak_disposal}</Paragraph>}
          </Card>
        )}

        {(plan.neutralizer || plan.neutralizer_usage) && (
          <Card size="small" title={<Space><ExperimentOutlined style={{ color: '#722ed1' }} />中和剂/吸附剂</Space>} style={{ borderRadius: 8 }}>
            <Descriptions column={1} size="small" bordered>
              {plan.neutralizer && <Descriptions.Item label="中和剂/吸附剂">{plan.neutralizer}</Descriptions.Item>}
              {plan.neutralizer_usage && <Descriptions.Item label="使用方法"><Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.neutralizer_usage}</Paragraph></Descriptions.Item>}
            </Descriptions>
          </Card>
        )}

        {equipList.length > 0 && (
          <Card size="small" title={<Space><SafetyOutlined style={{ color: '#1677ff' }} />防护装备清单</Space>} style={{ borderRadius: 8 }}>
            <List size="small" dataSource={equipList} renderItem={(item: any) => (
              <List.Item style={{ padding: '4px 0', border: 'none' }}>
                <Space><SafetyOutlined style={{ color: '#1677ff' }} /><Text>{typeof item === 'string' ? item : item.name || item.label || JSON.stringify(item)}</Text></Space>
              </List.Item>
            )} />
          </Card>
        )}

        {(plan.evacuation_distance || plan.isolation_distance) && (
          <Card size="small" title={<Space><EnvironmentOutlined style={{ color: '#f5222d' }} />疏散与隔离距离</Space>} style={{ borderRadius: 8 }}>
            <Descriptions column={2} size="small" bordered>
              {plan.evacuation_distance && <Descriptions.Item label="疏散距离"><Text type="danger" strong>{plan.evacuation_distance}</Text></Descriptions.Item>}
              {plan.isolation_distance && <Descriptions.Item label="隔离距离"><Text type="warning" strong>{plan.isolation_distance}</Text></Descriptions.Item>}
            </Descriptions>
          </Card>
        )}

        {plan.fire_fighting && (
          <Card size="small" title={<Space><FireOutlined style={{ color: '#ff4d4f' }} />灭火方法</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.fire_fighting}</Paragraph>
          </Card>
        )}

        {plan.first_aid && (
          <Card size="small" title={<Space><MedicineBoxOutlined style={{ color: '#52c41a' }} />急救措施</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.first_aid}</Paragraph>
          </Card>
        )}

        {plan.environmental_protection && (
          <Card size="small" title={<Space><EnvironmentOutlined style={{ color: '#13c2c2' }} />环保措施</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.environmental_protection}</Paragraph>
          </Card>
        )}

        {plan.special_precautions && (
          <Card size="small" title={<Space><WarningOutlined style={{ color: '#faad14' }} />特殊注意事项</Space>} style={{ borderRadius: 8 }}>
            <Alert type="warning" message={<Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.special_precautions}</Paragraph>} style={{ borderRadius: 6 }} />
          </Card>
        )}

        {contactsList.length > 0 && (
          <Card size="small" title={<Space><PhoneOutlined style={{ color: '#1677ff' }} />应急联系方式</Space>} style={{ borderRadius: 8 }}>
            <List size="small" dataSource={contactsList} renderItem={(item: any) => (
              <List.Item style={{ padding: '4px 0', border: 'none' }}>
                <Space>
                  <PhoneOutlined style={{ color: '#1677ff' }} />
                  <Text>{typeof item === 'string' ? item : `${item.name || item.title || ''}${item.phone ? ` - ${item.phone}` : ''}${item.role ? ` (${item.role})` : ''}`}</Text>
                </Space>
              </List.Item>
            )} />
          </Card>
        )}

        {plan.reference_standards && (
          <Card size="small" title={<Space><FileTextOutlined style={{ color: '#8c8c8c' }} />参考标准</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{plan.reference_standards}</Paragraph>
          </Card>
        )}

        <Divider />
        <Button type="primary" block size="large" icon={<ThunderboltOutlined />} onClick={() => handleOpenGenerateModal(plan)}>
          一键生成任务卡
        </Button>
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: '16px 24px' }}>
        <Row gutter={12} align="middle" wrap>
          <Col>
            <Input
              placeholder="UN编号搜索"
              value={filters.un_number}
              onChange={e => setFilters(f => ({ ...f, un_number: e.target.value }))}
              onPressEnter={handleSearch}
              style={{ width: 160 }}
              allowClear
            />
          </Col>
          <Col>
            <Select
              placeholder="危险类别"
              value={filters.danger_class}
              onChange={v => setFilters(f => ({ ...f, danger_class: v }))}
              style={{ width: 140 }}
              allowClear
            >
              {Object.entries(DANGER_CLASS_MAP).map(([k, v]) => (
                <Select.Option key={k} value={k}>
                  <Space><span style={{ display: 'inline-block', width: 8, height: 8, borderRadius: '50%', background: v.color }} />{k}类 - {v.label}</Space>
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col>
            <Input
              placeholder="关键词搜索"
              prefix={<SearchOutlined />}
              value={filters.keyword}
              onChange={e => setFilters(f => ({ ...f, keyword: e.target.value }))}
              onPressEnter={handleSearch}
              style={{ width: 200 }}
              allowClear
            />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={handleReset}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="预案总数" value={stats.total_plans} valueStyle={{ color: '#1677ff' }} prefix={<FileTextOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="内置预案" value={stats.builtin_plans} valueStyle={{ color: '#52c41a' }} prefix={<AppstoreOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="自定义预案" value={stats.custom_plans} valueStyle={{ color: '#fa8c16' }} prefix={<EditOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic title="活跃任务卡" value={stats.active_task_cards} valueStyle={{ color: '#f5222d' }} prefix={<ThunderboltOutlined />} />
          </Card>
        </Col>
      </Row>

      {stats.danger_class_distribution?.length > 0 && (
        <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 16 }}>
          <ReactECharts option={pieOption} style={{ height: 240 }} />
        </Card>
      )}

      <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: 0 }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          style={{ padding: '0 16px' }}
          items={[
            {
              key: 'plans',
              label: <Space><FileTextOutlined />预案知识库</Space>,
              children: (
                <Table<EmergencyPlanItem>
                  rowKey="id"
                  loading={planLoading}
                  columns={planColumns}
                  dataSource={plans}
                  pagination={{
                    current: planPage, pageSize: planPageSize, total: planTotal,
                    showSizeChanger: true, showQuickJumper: true,
                    showTotal: t => `共 ${t} 条预案`,
                    onChange: (p, ps) => fetchPlans(p, ps),
                  }}
                  scroll={{ x: 900 }}
                  size="middle"
                />
              ),
            },
            {
              key: 'tasks',
              label: <Space><ThunderboltOutlined />应急任务卡</Space>,
              children: (
                <Table<EmergencyTaskCard>
                  rowKey="id"
                  loading={taskCardLoading}
                  columns={taskCardColumns}
                  dataSource={taskCards}
                  pagination={{
                    current: taskCardPage, pageSize: taskCardPageSize, total: taskCardTotal,
                    showSizeChanger: true, showQuickJumper: true,
                    showTotal: t => `共 ${t} 条任务卡`,
                    onChange: (p, ps) => fetchTaskCards(p, ps),
                  }}
                  scroll={{ x: 1200 }}
                  size="middle"
                  onRow={(record) => ({ onClick: () => { setSelectedTaskCard(record); setTaskCardDrawerOpen(true) }, style: { cursor: 'pointer' } })}
                />
              ),
            },
          ]}
        />
      </Card>

      <Drawer
        title={<Space><SafetyCertificateOutlined style={{ color: '#1677ff' }} /><Text strong>预案详情</Text>{selectedPlan && <Tag color={DANGER_CLASS_MAP[selectedPlan.danger_class]?.color}>{selectedPlan.un_number}</Tag>}</Space>}
        open={planDrawerOpen}
        onClose={() => { setPlanDrawerOpen(false); setSelectedPlan(null) }}
        width={640}
      >
        {selectedPlan && renderPlanDetail(selectedPlan)}
      </Drawer>

      <Drawer
        title={<Space><ThunderboltOutlined style={{ color: '#f5222d' }} /><Text strong>任务卡详情</Text>{selectedTaskCard && <Tag>{selectedTaskCard.card_no}</Tag>}</Space>}
        open={taskCardDrawerOpen}
        onClose={() => { setTaskCardDrawerOpen(false); setSelectedTaskCard(null) }}
        width={600}
      >
        {selectedTaskCard && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Descriptions column={2} size="small" bordered>
              <Descriptions.Item label="任务卡编号" span={2}><Text copyable strong>{selectedTaskCard.card_no}</Text></Descriptions.Item>
              <Descriptions.Item label="UN编号">{selectedTaskCard.un_number}</Descriptions.Item>
              <Descriptions.Item label="车牌号"><Tag color="blue">{selectedTaskCard.plate_number}</Tag></Descriptions.Item>
              <Descriptions.Item label="司机">{selectedTaskCard.driver_name}</Descriptions.Item>
              <Descriptions.Item label="运单号">{selectedTaskCard.waybill_no || '-'}</Descriptions.Item>
              <Descriptions.Item label="推送状态">{(() => { const m = PUSH_STATUS_MAP[selectedTaskCard.push_status]; return m ? <Tag color={m.color}>{m.label}</Tag> : selectedTaskCard.push_status })()}</Descriptions.Item>
              <Descriptions.Item label="任务卡状态">{(() => { const m = CARD_STATUS_MAP[selectedTaskCard.card_status]; return m ? <Tag color={m.color}>{m.label}</Tag> : selectedTaskCard.card_status })()}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(selectedTaskCard.created_at)}</Descriptions.Item>
              <Descriptions.Item label="推送时间">{selectedTaskCard.pushed_at ? formatDateTime(selectedTaskCard.pushed_at) : '-'}</Descriptions.Item>
              {selectedTaskCard.acknowledged_at && <Descriptions.Item label="确认时间">{formatDateTime(selectedTaskCard.acknowledged_at)}</Descriptions.Item>}
              {selectedTaskCard.completed_at && <Descriptions.Item label="完成时间">{formatDateTime(selectedTaskCard.completed_at)}</Descriptions.Item>}
              <Descriptions.Item label="推送渠道" span={2}>
                <Space>{(selectedTaskCard.push_channels || []).map((ch: string) => <Tag key={ch}>{ch}</Tag>)}</Space>
              </Descriptions.Item>
            </Descriptions>
            {selectedTaskCard.plan_snapshot && (
              <>
                <Divider orientation="left">预案快照</Divider>
                {renderPlanDetail(selectedTaskCard.plan_snapshot as EmergencyPlanItem)}
              </>
            )}
          </div>
        )}
      </Drawer>

      <Modal
        title={<Space><ThunderboltOutlined />生成应急任务卡</Space>}
        open={generateModalOpen}
        onCancel={() => setGenerateModalOpen(false)}
        onOk={handleGenerate}
        confirmLoading={generating}
        okText="确认生成"
        width={480}
      >
        <Form form={generateForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="vehicle_id" label="选择车辆" rules={[{ required: true, message: '请选择车辆' }]}>
            <Select placeholder="请选择车辆" showSearch optionFilterProp="children">
              {vehicleList.map(v => (
                <Select.Option key={v.id} value={v.id}>
                  <Space><CarOutlined />{v.plate_number} ({v.vehicle_type})</Space>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="driver_id" label="选择司机" rules={[{ required: true, message: '请选择司机' }]}>
            <Select placeholder="请选择司机" showSearch optionFilterProp="children">
              {driverList.map(d => (
                <Select.Option key={d.id} value={d.id}>
                  <Space><UserOutlined />{d.name} ({d.employee_no})</Space>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="waybill_id" label="选择运单(可选)">
            <Select placeholder="请选择运单" showSearch optionFilterProp="children" allowClear>
              {waybillList.map(w => (
                <Select.Option key={w.id} value={w.id}>
                  {w.waybill_no} - {w.dangerous_goods_name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="push_channels" label="推送渠道">
            <Select mode="multiple" placeholder="请选择推送渠道">
              <Select.Option value="app">APP推送</Select.Option>
              <Select.Option value="sms">短信</Select.Option>
              <Select.Option value="voice">语音</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default EmergencyPlan
