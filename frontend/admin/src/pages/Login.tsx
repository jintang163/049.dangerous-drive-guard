import React from 'react'
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  Checkbox,
  message,
  Spin,
} from 'antd'
import {
  UserOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import api from '@/services/api'
import { setToken, setUserInfo, setPermissions } from '@/utils/auth'
import { useAppStore } from '@/store/app'

const { Title, Paragraph } = Typography

interface LoginForm {
  username: string
  password: string
  remember: boolean
}

const Login: React.FC = () => {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const { setUser, setPermissions: setStorePerms } = useAppStore()

  const handleLogin = async (values: LoginForm) => {
    setLoading(true)
    try {
      const res: any = await api.post('/auth/login', {
        username: values.username,
        password: values.password,
      })
      const { access_token, expires_at, user, permissions } = res
      setToken(access_token, new Date(expires_at))
      setUserInfo(user)
      setPermissions(permissions)
      setUser(user)
      setStorePerms(permissions)
      message.success(`欢迎回来，${user.real_name}！`)
      setTimeout(() => navigate('/dashboard', { replace: true }), 300)
    } catch (err: any) {
      message.error(err.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        background: `
          linear-gradient(135deg, rgba(22,119,255,0.9) 0%, rgba(0,33,64,0.95) 100%),
          url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 600"><defs><pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse"><path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="1"/></pattern></defs><rect width="100%" height="100%" fill="url(%23grid)"/></svg>')
        `,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
      }}
    >
      <div style={{ display: 'flex', maxWidth: 1000, width: '100%', gap: 48, alignItems: 'center' }}>
        <div style={{ flex: 1, color: '#fff', display: 'flex', flexDirection: 'column', gap: 24 }}>
          <div>
            <SafetyCertificateOutlined style={{ fontSize: 72, color: '#faad14', marginBottom: 16 }} />
            <Title level={1} style={{ color: '#fff', margin: 0, fontSize: 42 }}>
              危险品运输安全监控平台
            </Title>
            <Paragraph style={{ color: 'rgba(255,255,255,0.8)', fontSize: 16, marginTop: 16 }}>
              基于CloudWeGo微服务架构 · 集成高德高精地图 · AI疲劳实时检测
              <br />
              路径智能规划避开禁行区 · 区块链存证溯源 · 全链路安全保障
            </Paragraph>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2,1fr)', gap: 16 }}>
            {[
              { icon: '🛣️', title: '智能路径规划', desc: 'A*算法避开人口密集区、隧道、水源地' },
              { icon: '👁️', title: '疲劳AI识别', desc: '人脸关键点 + PERCLOS实时分析' },
              { icon: '📡', title: '实时监控调度', desc: 'WebSocket推送 + 语音对讲干预' },
              { icon: '🔗', title: '区块链存证', desc: '运单、报警数据不可篡改' },
            ].map(item => (
              <div
                key={item.title}
                style={{
                  padding: 16,
                  borderRadius: 12,
                  background: 'rgba(255,255,255,0.08)',
                  backdropFilter: 'blur(10px)',
                }}
              >
                <div style={{ fontSize: 28, marginBottom: 8 }}>{item.icon}</div>
                <div style={{ fontSize: 15, fontWeight: 600, marginBottom: 4 }}>{item.title}</div>
                <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.7)' }}>{item.desc}</div>
              </div>
            ))}
          </div>
        </div>
        <Card
          bordered={false}
          style={{
            width: 420,
            borderRadius: 16,
            boxShadow: '0 24px 80px rgba(0,0,0,0.24)',
          }}
          bodyStyle={{ padding: 36 }}
        >
          <Title level={3} style={{ textAlign: 'center', marginBottom: 8 }}>
            系统登录
          </Title>
          <Paragraph type="secondary" style={{ textAlign: 'center', marginBottom: 32 }}>
            请使用您的账号密码登录
          </Paragraph>
          <Form layout="vertical" onFinish={handleLogin} initialValues={{ remember: true, username: 'admin' }}>
            <Form.Item
              name="username"
              label="账号"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input
                size="large"
                prefix={<UserOutlined />}
                placeholder="请输入用户名"
                autoComplete="username"
              />
            </Form.Item>
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '密码至少6位' }]}
            >
              <Input.Password
                size="large"
                prefix={<LockOutlined />}
                placeholder="请输入密码"
                autoComplete="current-password"
              />
            </Form.Item>
            <Form.Item name="remember" valuePropName="checked">
              <Checkbox>记住我</Checkbox>
            </Form.Item>
            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                type="primary"
                htmlType="submit"
                size="large"
                block
                loading={loading}
                style={{ height: 44, fontWeight: 600 }}
              >
                {loading ? <Spin size="small" /> : '登 录'}
              </Button>
            </Form.Item>
          </Form>
          <div style={{ marginTop: 24, padding: 12, borderRadius: 8, background: '#e6f4ff' }}>
            <div style={{ fontSize: 12, color: '#1677ff', fontWeight: 500, marginBottom: 4 }}>
              测试账号
            </div>
            <div style={{ fontSize: 12, color: '#595959' }}>
              admin / admin123 (管理员) · dispatcher01 / disp123 (调度员)
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default Login
