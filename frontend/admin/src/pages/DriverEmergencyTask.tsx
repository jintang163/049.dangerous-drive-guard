import React, { useState, useEffect, useCallback } from 'react'
import {
  Card, Tag, Button, Space, List, Drawer, Descriptions, Badge,
  Radio, Empty, Spin, Typography, message, Row, Col, Statistic,
  Divider, Alert,
} from 'antd'
import {
  SafetyCertificateOutlined, CheckCircleOutlined,
  ClockCircleOutlined, CarOutlined, ExclamationCircleOutlined,
  WarningOutlined, ExperimentOutlined, MedicineBoxOutlined,
  EnvironmentOutlined, FireOutlined, PhoneOutlined,
  FileTextOutlined, BellOutlined, UserOutlined,
} from '@ant-design/icons'
import { emergencyApi } from '@/services/api'
import type { EmergencyTaskCard, EmergencyPlanItem } from '@/services/api'
import { useAppStore } from '@/store/app'
import WebSocketManager from '@/services/ws'
import dayjs from 'dayjs'

const { Text, Paragraph, Title } = Typography

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

const CARD_STATUS_MAP: Record<string, { label: string; color: string }> = {
  active: { label: '进行中', color: 'blue' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'default' },
  expired: { label: '已过期', color: 'red' },
}

const PUSH_STATUS_MAP: Record<string, { label: string; color: string }> = {
  pending: { label: '待推送', color: 'default' },
  pushed: { label: '已推送', color: 'blue' },
  acknowledged: { label: '已确认', color: 'green' },
  expired: { label: '已过期', color: 'red' },
}

const FILTER_OPTIONS = [
  { label: '全部', value: 'all' },
  { label: '待确认', value: 'pending_ack' },
  { label: '已确认', value: 'acknowledged' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
]

const parseJsonField = (val: any): any[] => {
  if (!val) return []
  if (Array.isArray(val)) return val
  try { return JSON.parse(val) } catch { return [] }
}

const formatSteps = (text: string) => {
  if (!text) return text
  return text.split(/(?=\d+[.、)])/g).filter(Boolean)
}

const DriverEmergencyTask: React.FC = () => {
  const { user } = useAppStore()
  const [loading, setLoading] = useState(false)
  const [taskCards, setTaskCards] = useState<EmergencyTaskCard[]>([])
  const [selectedCard, setSelectedCard] = useState<EmergencyTaskCard | null>(null)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [filterStatus, setFilterStatus] = useState('all')
  const [unreadCount, setUnreadCount] = useState(0)
  const [acknowledging, setAcknowledging] = useState(false)
  const [hasNewTask, setHasNewTask] = useState(false)

  const fetchTaskCards = useCallback(async () => {
    if (!user?.id) return
    setLoading(true)
    try {
      const data = await emergencyApi.listTaskCards({
        driver_id: user.id,
        page: 1,
        page_size: 50,
      })
      const list = data?.list || []
      setTaskCards(list)
      const unread = list.filter(
        (c: EmergencyTaskCard) => c.push_status === 'pushed' && c.card_status === 'active'
      ).length
      setUnreadCount(unread)
    } catch {
      message.error('获取任务卡列表失败')
    } finally {
      setLoading(false)
    }
  }, [user?.id])

  useEffect(() => {
    fetchTaskCards()
  }, [fetchTaskCards])

  useEffect(() => {
    const ws = WebSocketManager.getInstance()
    ws.connect()

    const handleNewTaskCard = (data: any) => {
      if (data && data.driver_id === user?.id) {
        setTaskCards(prev => {
          const exists = prev.some(c => c.id === data.id)
          if (exists) {
            return prev.map(c => c.id === data.id ? { ...c, ...data } : c)
          }
          return [data, ...prev]
        })
        setUnreadCount(prev => prev + 1)
        setHasNewTask(true)
        message.info('您有新的应急任务卡，请及时查看')
      }
    }

    const unsub = ws.on('emergency_task_card', handleNewTaskCard)
    return () => { unsub() }
  }, [user?.id])

  const filteredCards = taskCards.filter(card => {
    if (filterStatus === 'all') return true
    if (filterStatus === 'pending_ack') {
      return card.push_status === 'pushed' && card.card_status === 'active'
    }
    if (filterStatus === 'acknowledged') {
      return card.push_status === 'acknowledged' && card.card_status === 'active'
    }
    return card.card_status === filterStatus
  })

  const handleViewDetail = (card: EmergencyTaskCard) => {
    setSelectedCard(card)
    setDrawerOpen(true)
    if (hasNewTask) {
      setHasNewTask(false)
    }
  }

  const handleAcknowledge = async () => {
    if (!selectedCard) return
    setAcknowledging(true)
    try {
      await emergencyApi.ackTaskCard(selectedCard.id)
      message.success('任务卡确认成功')
      setTaskCards(prev =>
        prev.map(c =>
          c.id === selectedCard.id
            ? { ...c, push_status: 'acknowledged' as const, acknowledged_at: new Date().toISOString() }
            : c
        )
      )
      setSelectedCard(prev =>
        prev ? { ...prev, push_status: 'acknowledged' as const, acknowledged_at: new Date().toISOString() } : null
      )
      setUnreadCount(prev => Math.max(0, prev - 1))
    } catch (e: any) {
      message.error(e?.message || '确认失败')
    } finally {
      setAcknowledging(false)
    }
  }

  const renderCardItem = (card: EmergencyTaskCard) => {
    const dangerInfo = DANGER_CLASS_MAP[card.plan_snapshot?.danger_class || ''] || { label: '未知', color: 'default' }
    const isPendingAck = card.push_status === 'pushed' && card.card_status === 'active'
    const cardStatus = CARD_STATUS_MAP[card.card_status] || { label: card.card_status, color: 'default' }
    const pushStatus = PUSH_STATUS_MAP[card.push_status] || { label: card.push_status, color: 'default' }

    return (
      <Card
        size="small"
        style={{
          borderRadius: 12,
          marginBottom: 12,
          border: isPendingAck ? '2px solid #ff4d4f' : '1px solid #f0f0f0',
          boxShadow: isPendingAck ? '0 2px 8px rgba(255,77,79,0.15)' : 'none',
          cursor: 'pointer',
          transition: 'all 0.2s',
        }}
        bodyStyle={{ padding: '12px 16px' }}
        onClick={() => handleViewDetail(card)}
        hoverable
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
          <Space size={8} wrap>
            <Tag color={dangerInfo.color} style={{ margin: 0, fontWeight: 600 }}>
              UN {card.un_number}
            </Tag>
            <Tag color={dangerInfo.color} style={{ margin: 0 }}>
              {dangerInfo.label}
            </Tag>
          </Space>
          <Space size={8}>
            {isPendingAck && (
              <Badge dot color="#ff4d4f" />
            )}
            <Tag color={cardStatus.color} style={{ margin: 0 }}>
              {cardStatus.label}
            </Tag>
          </Space>
        </div>

        <Title level={5} style={{ margin: '0 0 8px 0', fontSize: 15 }} ellipsis={{ rows: 1 }}>
          {card.title}
        </Title>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: '#8c8c8c' }}>
          <Space size={12}>
            <span>
              <CarOutlined style={{ marginRight: 4 }} />
              {card.plate_number}
            </span>
            <span>
              <ClockCircleOutlined style={{ marginRight: 4 }} />
              {dayjs(card.pushed_at || card.created_at).format('MM-DD HH:mm')}
            </span>
          </Space>
          <Tag color={pushStatus.color} style={{ margin: 0, fontSize: 11 }} size="small">
            {pushStatus.label}
          </Tag>
        </div>

        {card.plan_snapshot?.hazard_summary && (
          <div
            style={{
              marginTop: 10,
              padding: '8px 10px',
              background: '#fffbe6',
              borderRadius: 6,
              fontSize: 12,
              color: '#d46b08',
              lineHeight: 1.5,
            }}
          >
            <ExclamationCircleOutlined style={{ marginRight: 4 }} />
            <Text ellipsis={{ rows: 2 }} style={{ color: '#d46b08', fontSize: 12 }}>
              {card.plan_snapshot.hazard_summary}
            </Text>
          </div>
        )}

        {isPendingAck && (
          <div style={{ marginTop: 10, textAlign: 'right' }}>
            <Button
              type="primary"
              danger
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={e => {
                e.stopPropagation()
                setSelectedCard(card)
                setDrawerOpen(true)
              }}
            >
              确认接收
            </Button>
          </div>
        )}
      </Card>
    )
  }

  const renderPlanDetail = (plan: Partial<EmergencyPlanItem>) => {
    const equipList = parseJsonField(plan.protective_equipment)
    const contactsList = parseJsonField(plan.emergency_contacts)
    const leakSteps = formatSteps(plan.leak_disposal || '')

    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {plan.hazard_summary && (
          <Card size="small" title={<Space><WarningOutlined style={{ color: '#ff4d4f' }} />危险性概述</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.hazard_summary}</Paragraph>
          </Card>
        )}

        {plan.leak_disposal && (
          <Card size="small" title={<Space><ExperimentOutlined style={{ color: '#fa8c16' }} />泄漏处置方法</Space>} style={{ borderRadius: 8 }}>
            {Array.isArray(leakSteps) && leakSteps.length > 1 ? (
              <List size="small" dataSource={leakSteps} renderItem={(step, idx) => (
                <List.Item style={{ padding: '4px 0', border: 'none' }}>
                  <Space align="start" size={8}>
                    <Tag color="orange" style={{ borderRadius: '50%', width: 20, height: 20, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', padding: 0, fontSize: 11, flexShrink: 0 }}>{idx + 1}</Tag>
                    <Text style={{ whiteSpace: 'pre-wrap', fontSize: 13 }}>{step.replace(/^\d+[.、)]\s*/, '')}</Text>
                  </Space>
                </List.Item>
              )} />
            ) : <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.leak_disposal}</Paragraph>}
          </Card>
        )}

        {(plan.neutralizer || plan.neutralizer_usage) && (
          <Card size="small" title={<Space><ExperimentOutlined style={{ color: '#722ed1' }} />中和剂/吸附剂</Space>} style={{ borderRadius: 8 }}>
            <Descriptions column={1} size="small" bordered>
              {plan.neutralizer && <Descriptions.Item label="中和剂/吸附剂" contentStyle={{ fontSize: 13 }}>{plan.neutralizer}</Descriptions.Item>}
              {plan.neutralizer_usage && <Descriptions.Item label="使用方法"><Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.neutralizer_usage}</Paragraph></Descriptions.Item>}
            </Descriptions>
          </Card>
        )}

        {equipList.length > 0 && (
          <Card size="small" title={<Space><SafetyCertificateOutlined style={{ color: '#1677ff' }} />防护装备清单</Space>} style={{ borderRadius: 8 }}>
            <List size="small" dataSource={equipList} renderItem={(item: any) => (
              <List.Item style={{ padding: '4px 0', border: 'none' }}>
                <Space size={8}>
                  <SafetyCertificateOutlined style={{ color: '#1677ff' }} />
                  <Text style={{ fontSize: 13 }}>{typeof item === 'string' ? item : item.name || item.label || JSON.stringify(item)}</Text>
                </Space>
              </List.Item>
            )} />
          </Card>
        )}

        {plan.first_aid && (
          <Card size="small" title={<Space><MedicineBoxOutlined style={{ color: '#52c41a' }} />急救措施</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.first_aid}</Paragraph>
          </Card>
        )}

        {plan.fire_fighting && (
          <Card size="small" title={<Space><FireOutlined style={{ color: '#ff4d4f' }} />灭火方法</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.fire_fighting}</Paragraph>
          </Card>
        )}

        {(plan.evacuation_distance || plan.isolation_distance) && (
          <Card size="small" title={<Space><EnvironmentOutlined style={{ color: '#f5222d' }} />疏散与隔离距离</Space>} style={{ borderRadius: 8 }}>
            <Descriptions column={2} size="small" bordered>
              {plan.evacuation_distance && <Descriptions.Item label="疏散距离"><Text type="danger" strong style={{ fontSize: 13 }}>{plan.evacuation_distance}</Text></Descriptions.Item>}
              {plan.isolation_distance && <Descriptions.Item label="隔离距离"><Text type="warning" strong style={{ fontSize: 13 }}>{plan.isolation_distance}</Text></Descriptions.Item>}
            </Descriptions>
          </Card>
        )}

        {plan.environmental_protection && (
          <Card size="small" title={<Space><EnvironmentOutlined style={{ color: '#13c2c2' }} />环保措施</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.environmental_protection}</Paragraph>
          </Card>
        )}

        {plan.special_precautions && (
          <Card size="small" title={<Space><WarningOutlined style={{ color: '#faad14' }} />特殊注意事项</Space>} style={{ borderRadius: 8 }}>
            <Alert type="warning" message={<Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.special_precautions}</Paragraph>} style={{ borderRadius: 6 }} />
          </Card>
        )}

        {contactsList.length > 0 && (
          <Card size="small" title={<Space><PhoneOutlined style={{ color: '#1677ff' }} />应急联系方式</Space>} style={{ borderRadius: 8 }}>
            <List size="small" dataSource={contactsList} renderItem={(item: any) => (
              <List.Item style={{ padding: '4px 0', border: 'none' }}>
                <Space size={8}>
                  <PhoneOutlined style={{ color: '#1677ff' }} />
                  <Text style={{ fontSize: 13 }}>{typeof item === 'string' ? item : `${item.name || item.title || ''}${item.phone ? ` - ${item.phone}` : ''}${item.role ? ` (${item.role})` : ''}`}</Text>
                </Space>
              </List.Item>
            )} />
          </Card>
        )}

        {plan.reference_standards && (
          <Card size="small" title={<Space><FileTextOutlined style={{ color: '#8c8c8c' }} />参考标准</Space>} style={{ borderRadius: 8 }}>
            <Paragraph style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 13 }}>{plan.reference_standards}</Paragraph>
          </Card>
        )}
      </div>
    )
  }

  return (
    <div style={{ padding: '16px', maxWidth: 600, margin: '0 auto' }}>
      <Card
        bordered={false}
        style={{ borderRadius: 12, marginBottom: 16 }}
        bodyStyle={{ padding: '16px 20px' }}
      >
        <Row gutter={12} align="middle">
          <Col flex="auto">
            <Space size={12} align="center">
              <div style={{
                width: 48, height: 48, borderRadius: '50%',
                background: 'linear-gradient(135deg, #1677ff, #69b1ff)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#fff', fontSize: 20,
              }}>
                <UserOutlined />
              </div>
              <div>
                <div style={{ fontSize: 16, fontWeight: 600 }}>{user?.real_name || user?.username || '司机'}</div>
                <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                  <CarOutlined style={{ marginRight: 4 }} />
                  应急任务卡
                </div>
              </div>
            </Space>
          </Col>
          <Col>
            <Badge count={unreadCount} size="small" offset={[-4, 4]}>
              <Statistic
                title={<span style={{ fontSize: 12 }}>待处理任务</span>}
                value={unreadCount}
                valueStyle={{ color: unreadCount > 0 ? '#ff4d4f' : '#8c8c8c', fontSize: 24 }}
                prefix={<BellOutlined />}
              />
            </Badge>
          </Col>
        </Row>

        {hasNewTask && (
          <Alert
            type="warning"
            showIcon
            message="您有新的应急任务卡"
            style={{ marginTop: 12, borderRadius: 8 }}
            action={
              <Button size="small" type="primary" danger onClick={fetchTaskCards}>
                刷新查看
              </Button>
            }
          />
        )}
      </Card>

      <Card
        bordered={false}
        style={{ borderRadius: 12, marginBottom: 16 }}
        bodyStyle={{ padding: '12px 16px' }}
      >
        <Radio.Group
          value={filterStatus}
          onChange={e => setFilterStatus(e.target.value)}
          optionType="button"
          buttonStyle="solid"
          size="small"
          style={{ width: '100%', display: 'flex', flexWrap: 'wrap', gap: 4 }}
        >
          {FILTER_OPTIONS.map(opt => (
            <Radio.Button key={opt.value} value={opt.value} style={{ flex: 1, minWidth: 60, textAlign: 'center' }}>
              {opt.label}
            </Radio.Button>
          ))}
        </Radio.Group>
      </Card>

      <Spin spinning={loading} tip="加载中...">
        {filteredCards.length > 0 ? (
          <div>
            {filteredCards.map(card => (
              <div key={card.id}>
                {renderCardItem(card)}
              </div>
            ))}
          </div>
        ) : (
          <Empty
            description="暂无任务卡"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            style={{ padding: '40px 0' }}
          />
        )}
      </Spin>

      <Drawer
        title={
          <Space>
            <SafetyCertificateOutlined style={{ color: '#1677ff' }} />
            <Text strong>应急任务卡详情</Text>
            {selectedCard && (
              <Tag color={DANGER_CLASS_MAP[selectedCard.plan_snapshot?.danger_class || '']?.color || 'default'}>
                UN {selectedCard.un_number}
              </Tag>
            )}
          </Space>
        }
        open={drawerOpen}
        onClose={() => { setDrawerOpen(false); setSelectedCard(null) }}
        width={360}
        placement="right"
        extra={
          selectedCard && selectedCard.push_status === 'pushed' && selectedCard.card_status === 'active' ? (
            <Button
              type="primary"
              danger
              icon={<CheckCircleOutlined />}
              onClick={handleAcknowledge}
              loading={acknowledging}
            >
              确认接收
            </Button>
          ) : null
        }
      >
        {selectedCard && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="任务卡编号">
                <Text copyable style={{ fontSize: 13 }}>{selectedCard.card_no}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="标题">
                <Text strong style={{ fontSize: 13 }}>{selectedCard.title}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="UN编号">
                <Tag color={DANGER_CLASS_MAP[selectedCard.plan_snapshot?.danger_class || '']?.color || 'default'}>
                  UN {selectedCard.un_number}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="危险类别">
                {DANGER_CLASS_MAP[selectedCard.plan_snapshot?.danger_class || '']?.label || selectedCard.plan_snapshot?.danger_class || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="车牌号">
                <Tag color="blue">{selectedCard.plate_number}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="司机">{selectedCard.driver_name}</Descriptions.Item>
              {selectedCard.waybill_no && <Descriptions.Item label="运单号">{selectedCard.waybill_no}</Descriptions.Item>}
              <Descriptions.Item label="推送状态">
                {(() => {
                  const m = PUSH_STATUS_MAP[selectedCard.push_status]
                  return m ? <Tag color={m.color}>{m.label}</Tag> : selectedCard.push_status
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="任务卡状态">
                {(() => {
                  const m = CARD_STATUS_MAP[selectedCard.card_status]
                  return m ? <Tag color={m.color}>{m.label}</Tag> : selectedCard.card_status
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {dayjs(selectedCard.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              {selectedCard.pushed_at && (
                <Descriptions.Item label="推送时间">
                  {dayjs(selectedCard.pushed_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
              {selectedCard.acknowledged_at && (
                <Descriptions.Item label="确认时间">
                  {dayjs(selectedCard.acknowledged_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
              {selectedCard.completed_at && (
                <Descriptions.Item label="完成时间">
                  {dayjs(selectedCard.completed_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
            </Descriptions>

            {selectedCard.plan_snapshot && (
              <>
                <Divider orientation="left" style={{ margin: '8px 0' }}>预案快照</Divider>
                {renderPlanDetail(selectedCard.plan_snapshot)}
              </>
            )}
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default DriverEmergencyTask
