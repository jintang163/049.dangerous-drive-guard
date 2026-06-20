import React, { useEffect } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Button,
  Badge,
  Space,
  Typography,
  Tooltip,
  Divider,
} from 'antd'
import {
  DashboardOutlined,
  CarOutlined,
  RouteOutlined,
  AlertOutlined,
  TruckOutlined,
  UserOutlined,
  FileTextOutlined,
  SafetyCertificateOutlined,
  WarningOutlined,
  CloudOutlined,
  LinkOutlined,
  BellOutlined,
  SettingOutlined,
  LogoutOutlined,
  ExpandOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FundProjectionScreenOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAppStore } from '@/store/app'
import { getUserInfo, clearToken, hasPermission } from '@/utils/auth'
import WebSocketManager from '@/services/ws'

const { Header, Sider, Content, Footer } = Layout
const { Title } = Typography

type MenuItem = Required<MenuProps>['items'][number]

const MainLayout: React.FC = () => {
  const navigate = useNavigate()
  const {
    user,
    sidebarCollapsed,
    setSidebarCollapsed,
    unreadAlarmCount,
    setUser,
    setPermissions,
    logout,
  } = useAppStore()

  useEffect(() => {
    const savedUser = getUserInfo()
    if (savedUser && !user) {
      setUser(savedUser)
    }
    const ws = WebSocketManager.getInstance()
    ws.connect()
    return () => {
      ws.disconnect()
    }
  }, [])

  const menuItems: MenuItem[] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '数据总览',
    },
    {
      key: '/monitor',
      icon: <FundProjectionScreenOutlined />,
      label: '实时监控',
    },
    {
      key: '/route-plan',
      icon: <RouteOutlined />,
      label: '路径规划',
    },
    {
      key: '/fatigue-alarms',
      icon: <AlertOutlined />,
      label: '疲劳报警',
    },
    {
      key: '/vehicles',
      icon: <TruckOutlined />,
      label: '车辆管理',
    },
    {
      key: '/drivers',
      icon: <UserOutlined />,
      label: '驾驶员管理',
    },
    {
      key: '/waybills',
      icon: <FileTextOutlined />,
      label: '电子运单',
    },
    {
      key: '/escort',
      icon: <SafetyCertificateOutlined />,
      label: '电子押运',
    },
    {
      key: '/rescue',
      icon: <WarningOutlined />,
      label: '紧急救援',
    },
    {
      key: '/weather',
      icon: <CloudOutlined />,
      label: '天气预警',
    },
    {
      key: '/blockchain',
      icon: <LinkOutlined />,
      label: '区块链存证',
    },
  ]

  const filteredMenuItems = menuItems.filter(item => {
    if (!item) return false
    const key = (item as any).key as string
    const permMap: Record<string, string> = {
      '/dashboard': 'dashboard:view',
      '/monitor': 'monitor:view',
      '/route-plan': 'route:plan',
      '/fatigue-alarms': 'alarm:handle',
      '/vehicles': 'vehicle:view',
      '/drivers': 'vehicle:view',
      '/waybills': 'waybill:manage',
      '/escort': 'escort:event_report',
    }
    if (permMap[key]) return hasPermission(permMap[key])
    return true
  })

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'setting',
      icon: <SettingOutlined />,
      label: '系统设置',
      disabled: true,
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout()
        clearToken()
        navigate('/login', { replace: true })
      },
    },
  ]

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        theme="dark"
        width={240}
        style={{
          background: 'linear-gradient(180deg, #001529 0%, #002140 100%)',
          position: 'sticky',
          top: 0,
          height: '100vh',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: sidebarCollapsed ? '0 12px' : '0 20px',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          {sidebarCollapsed ? (
            <CarOutlined style={{ fontSize: 28, color: '#1677ff' }} />
          ) : (
            <Title level={5} style={{ color: '#fff', margin: 0, whiteSpace: 'nowrap' }}>
              <SafetyCertificateOutlined style={{ color: '#1677ff', marginRight: 8 }} />
              危运安全监控平台
            </Title>
          )}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          items={filteredMenuItems}
          selectedKeys={[useAppStore.getState().currentPath]}
          onClick={handleMenuClick}
          style={{
            borderRight: 'none',
            marginTop: 12,
          }}
        />
        <div
          style={{
            position: 'absolute',
            bottom: 16,
            left: sidebarCollapsed ? 20 : 24,
            right: sidebarCollapsed ? 20 : 24,
            opacity: 0.6,
            fontSize: 12,
            color: '#fff',
            textAlign: 'center',
          }}
        >
          v1.0.0 · DDG System
        </div>
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
            position: 'sticky',
            top: 0,
            zIndex: 10,
          }}
        >
          <Space>
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              style={{ fontSize: 16 }}
            />
            <Title level={4} style={{ margin: 0, color: '#1f1f1f' }}>
              {filteredMenuItems.find(i => i && (i as any).key === useAppStore.getState().currentPath) as any
                ? (filteredMenuItems.find(i => i && (i as any).key === useAppStore.getState().currentPath) as any).label
                : '危险品运输安全监控平台'}
            </Title>
          </Space>
          <Space size={16}>
            <Tooltip title="打开大屏">
              <Button
                type="text"
                icon={<ExpandOutlined />}
                onClick={() => window.open('/big-screen', '_blank')}
              />
            </Tooltip>
            <Badge count={unreadAlarmCount} overflowCount={99} size="small">
              <Tooltip title="报警通知">
                <Button
                  type="text"
                  icon={<BellOutlined style={{ color: '#fa8c16', fontSize: 18 }} />}
                  onClick={() => navigate('/fatigue-alarms')}
                />
              </Tooltip>
            </Badge>
            <Divider type="vertical" />
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Space style={{ cursor: 'pointer', padding: '4px 8px' }}>
                <Avatar
                  size={36}
                  style={{ backgroundColor: '#1677ff' }}
                  src={user?.avatar_url}
                >
                  {user?.real_name?.[0] || 'U'}
                </Avatar>
                <div style={{ textAlign: 'left' }}>
                  <div style={{ fontSize: 14, fontWeight: 500, lineHeight: 1.2 }}>
                    {user?.real_name || '未登录'}
                  </div>
                  <div style={{ fontSize: 12, color: '#8c8c8c', lineHeight: 1.4 }}>
                    {getRoleLabel(user?.role)}
                  </div>
                </div>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content
          style={{
            margin: 0,
            padding: 16,
            background: '#f0f2f5',
            minHeight: 'calc(100vh - 64px)',
          }}
        >
          <Outlet />
        </Content>
        <Footer style={{ textAlign: 'center', color: '#8c8c8c' }}>
          Dangerous Drive Guard System ©{new Date().getFullYear()} | 危险品运输安全监控平台
        </Footer>
      </Layout>
    </Layout>
  )
}

const getRoleLabel = (role?: string) => {
  const map: Record<string, string> = {
    admin: '系统管理员',
    dispatcher: '调度员',
    driver: '驾驶员',
    escort: '押运员',
    viewer: '访客',
  }
  return map[role || ''] || role
}

export default MainLayout
