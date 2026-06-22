import React, { useEffect } from 'react'
import {
  notification,
  Modal,
  Button,
  Space,
  Tag,
  Descriptions,
  Typography,
} from 'antd'
import {
  WarningOutlined,
  CheckCircleOutlined,
  EyeOutlined,
  SafetyOutlined,
  EnvironmentOutlined,
  CarOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import WebSocketManager from '@/services/ws'
import { useAppStore, FaultAlert } from '@/store/app'
import { vehicleApi } from '@/services/api'

const { Title, Text, Paragraph } = Typography

const levelConfig: Record<number, { color: string; bgColor: string; label: string; type: 'open' | 'success' | 'info' | 'warning' | 'error' }> = {
  1: { color: '#1677ff', bgColor: '#e6f4ff', label: '提示', type: 'info' },
  2: { color: '#faad14', bgColor: '#fffbe6', label: '警告', type: 'warning' },
  3: { color: '#fa8c16', bgColor: '#fff7e6', label: '严重', type: 'warning' },
  4: { color: '#ff4d4f', bgColor: '#fff1f0', label: '紧急', type: 'error' },
}

const FaultAlertNotification: React.FC = () => {
  const navigate = useNavigate()
  const { faultAlerts, addFaultAlert, ackFaultAlert: ackFaultAlertInStore, resolveFaultAlert: resolveFaultAlertInStore } = useAppStore()
  const [api, contextHolder] = notification.useNotification()
  const [modal, modalContextHolder] = Modal.useModal()

  const showNotification = (alert: FaultAlert) => {
    const config = levelConfig[alert.fault_level] || levelConfig[1]
    const key = `fault-alert-${alert.id}-${Date.now()}`

    api.open({
      key,
      message: (
        <Space>
          <Tag color={config.color} style={{ margin: 0 }}>
            {config.label}
          </Tag>
          <Text strong>{alert.fault_code}</Text>
        </Space>
      ),
      description: (
        <div>
          <div style={{ marginBottom: 4 }}>
            <CarOutlined /> <Text>{alert.plate_number}</Text>
          </div>
          <Paragraph
            ellipsis={{ rows: 2 }}
            style={{ margin: 0, color: 'rgba(0,0,0,0.85)' }}
          >
            {alert.fault_desc}
          </Paragraph>
        </div>
      ),
      icon: <WarningOutlined style={{ color: config.color }} />,
      duration: alert.fault_level >= 3 ? 0 : 8,
      placement: 'topRight',
      style: {
        borderLeft: `4px solid ${config.color}`,
        background: config.bgColor,
      },
      btn: (
        <Space size="small">
          {alert.fault_level >= 3 && (
            <Button
              type="primary"
              size="small"
              danger={alert.fault_level === 4}
              onClick={() => {
                api.destroy(key)
                showDetailModal(alert)
              }}
            >
              查看详情
            </Button>
          )}
          <Button
            size="small"
            onClick={async () => {
              try {
                await vehicleApi.ackFaultAlert(alert.vehicle_id, alert.id)
                ackFaultAlertInStore(alert.id)
                api.destroy(key)
              } catch (e) {
                console.error('ack fault alert failed', e)
              }
            }}
          >
            确认
          </Button>
        </Space>
      ),
    })
  }

  const showDetailModal = (alert: FaultAlert) => {
    const config = levelConfig[alert.fault_level] || levelConfig[1]
    const isLevel4 = alert.fault_level === 4

    modal.confirm({
      title: (
        <Space>
          <WarningOutlined style={{ color: config.color, fontSize: 20 }} />
          <Title level={4} style={{ margin: 0, color: config.color }}>
            故障{config.label} - {alert.fault_code}
          </Title>
        </Space>
      ),
      icon: null,
      width: 640,
      okText: '确认',
      cancelText: '关闭',
      okButtonProps: {
        danger: isLevel4,
        type: 'primary',
      },
      content: (
        <div>
          <Descriptions bordered column={2} size="small" style={{ marginBottom: 16 }}>
            <Descriptions.Item label="车牌号">
              <Space>
                <CarOutlined />
                {alert.plate_number}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="故障等级">
              <Tag color={config.color}>{config.label}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="故障系统">
              {alert.fault_system}
            </Descriptions.Item>
            <Descriptions.Item label="上报时间">
              <Space>
                <ClockCircleOutlined />
                {alert.report_time}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="位置" span={2}>
              {alert.latitude && alert.longitude ? (
                <Space>
                  <EnvironmentOutlined />
                  <Text copyable>
                    {alert.latitude.toFixed(6)}, {alert.longitude.toFixed(6)}
                  </Text>
                </Space>
              ) : (
                <Text type="secondary">暂无位置信息</Text>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="故障描述" span={2}>
              <Paragraph style={{ margin: 0 }}>{alert.fault_desc}</Paragraph>
            </Descriptions.Item>
            {alert.fault_suggestion && (
              <Descriptions.Item label="处理建议" span={2}>
                <Paragraph style={{ margin: 0 }} type="warning">
                  {alert.fault_suggestion}
                </Paragraph>
              </Descriptions.Item>
            )}
          </Descriptions>

          {alert.emergency_action && (
            <div
              style={{
                padding: 16,
                background: isLevel4 ? '#fff1f0' : '#fff7e6',
                borderRadius: 8,
                border: `1px solid ${config.color}33`,
              }}
            >
              <Space style={{ marginBottom: 8 }}>
                <SafetyOutlined style={{ color: config.color }} />
                <Text strong style={{ color: config.color }}>
                  紧急处置步骤
                </Text>
              </Space>
              <Paragraph
                style={{
                  margin: 0,
                  color: 'rgba(0,0,0,0.85)',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {alert.emergency_action}
              </Paragraph>
            </div>
          )}
        </div>
      ),
      footer: (
        <Space style={{ justifyContent: 'flex-end', width: '100%' }}>
          {isLevel4 && (
            <Button
              type="primary"
              danger
              icon={<SafetyOutlined />}
              onClick={async () => {
                try {
                  await vehicleApi.resolveFaultAlert(alert.vehicle_id, alert.id, {
                    resolve_note: '已发起救援请求',
                  })
                  resolveFaultAlertInStore(alert.id)
                  notification.success({
                    message: '救援请求已发起',
                    description: `车辆 ${alert.plate_number} 的救援请求已发送`,
                  })
                  modal.destroy()
                } catch (e) {
                  console.error('rescue request failed', e)
                }
              }}
            >
              救援
            </Button>
          )}
          <Button
            icon={<EyeOutlined />}
            onClick={() => {
              modal.destroy()
              navigate('/vehicles')
            }}
          >
            查看详情
          </Button>
          <Button
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={async () => {
              try {
                await vehicleApi.ackFaultAlert(alert.vehicle_id, alert.id)
                ackFaultAlertInStore(alert.id)
                notification.success({
                  message: '已确认',
                  description: `故障告警 ${alert.fault_code} 已确认`,
                })
                modal.destroy()
              } catch (e) {
                console.error('ack fault alert failed', e)
              }
            }}
          >
            确认
          </Button>
        </Space>
      ),
    })
  }

  useEffect(() => {
    const ws = WebSocketManager.getInstance()

    const unsubscribe = ws.on('new_fault_alert', (data: any) => {
      const alert = data as FaultAlert
      addFaultAlert(alert)

      if (alert.fault_level >= 3) {
        showDetailModal(alert)
      }
      showNotification(alert)
    })

    return () => {
      unsubscribe()
    }
  }, [addFaultAlert])

  return (
    <>
      {contextHolder}
      {modalContextHolder}
    </>
  )
}

export default FaultAlertNotification
