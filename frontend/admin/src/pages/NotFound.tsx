import React from 'react'
import { Result, Button } from 'antd'
import { HomeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

const NotFound: React.FC = () => {
  const navigate = useNavigate()

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            radial-gradient(circle at 20% 30%, rgba(255,255,255,0.15) 0%, transparent 40%),
            radial-gradient(circle at 80% 70%, rgba(255,255,255,0.1) 0%, transparent 40%),
            url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 600 600"><defs><pattern id="g" width="60" height="60" patternUnits="userSpaceOnUse"><circle cx="30" cy="30" r="1" fill="rgba(255,255,255,0.08)"/></pattern></defs><rect width="100%" height="100%" fill="url(%23g)"/></svg>')
          `,
        }}
      />
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          maxWidth: 520,
          width: '100%',
          background: 'rgba(255,255,255,0.96)',
          borderRadius: 24,
          boxShadow: '0 32px 80px rgba(0,0,0,0.18)',
          backdropFilter: 'blur(20px)',
          padding: '48px 40px',
          textAlign: 'center',
        }}
      >
        <Result
          icon={
            <img
              src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 400 220'><defs><linearGradient id='g1' x1='0%' y1='0%' x2='100%' y2='100%'><stop offset='0%' stop-color='%231677ff'/><stop offset='100%' stop-color='%23722ed1'/></linearGradient></defs><text x='200' y='150' font-family='Segoe UI, Arial, sans-serif' font-size='140' font-weight='800' fill='url(%23g1)' text-anchor='middle'>404</text><path d='M60 190 L340 190' stroke='%23e5e7eb' stroke-width='2' stroke-linecap='round'/><circle cx='80' cy='186' r='8' fill='%23faad14'/><rect x='100' y='160' width='50' height='26' rx='4' fill='%231677ff' opacity='0.6'/><circle cx='180' cy='186' r='6' fill='%2352c41a'/><rect x='198' y='166' width='38' height='20' rx='3' fill='%23eb2f96' opacity='0.5'/><circle cx='270' cy='186' r='7' fill='%23ff4d4f'/><rect x='288' y='162' width='44' height='24' rx='3' fill='%2313c2c2' opacity='0.55'/></svg>"
              alt="404"
              style={{ maxWidth: 360, width: '100%', marginBottom: 8 }}
            />
          }
          status="404"
          title={
            <span
              style={{
                fontSize: 28,
                fontWeight: 700,
                color: '#1f1f1f',
                display: 'block',
                marginBottom: 12,
              }}
            >
              页面走丢了
            </span>
          }
          subTitle={
            <span style={{ fontSize: 14, color: '#8c8c8c', lineHeight: 1.8 }}>
              抱歉，您访问的页面不存在或已被移动。<br />
              请检查URL是否正确，或返回首页继续浏览。
            </span>
          }
          extra={
            <div
              style={{
                display: 'flex',
                gap: 12,
                justifyContent: 'center',
                flexWrap: 'wrap',
                marginTop: 16,
              }}
            >
              <Button
                type="primary"
                size="large"
                icon={<HomeOutlined />}
                onClick={() => navigate('/dashboard', { replace: true })}
                style={{
                  height: 44,
                  minWidth: 160,
                  fontSize: 15,
                  fontWeight: 600,
                  borderRadius: 10,
                  background: 'linear-gradient(135deg, #1677ff 0%, #722ed1 100%)',
                  border: 'none',
                  boxShadow: '0 8px 20px rgba(22,119,255,0.3)',
                }}
              >
                返回首页
              </Button>
              <Button
                size="large"
                onClick={() => navigate(-1)}
                style={{
                  height: 44,
                  minWidth: 120,
                  fontSize: 15,
                  borderRadius: 10,
                }}
              >
                返回上页
              </Button>
            </div>
          }
        />
        <div
          style={{
            marginTop: 32,
            paddingTop: 24,
            borderTop: '1px dashed #e5e7eb',
            display: 'flex',
            justifyContent: 'center',
            gap: 24,
            flexWrap: 'wrap',
          }}
        >
          {[
            { label: '数据总览', path: '/dashboard' },
            { label: '实时监控', path: '/monitor' },
            { label: '路径规划', path: '/route-plan' },
            { label: '疲劳报警', path: '/fatigue-alarms' },
          ].map(item => (
            <a
              key={item.path}
              onClick={() => navigate(item.path)}
              style={{
                color: '#1677ff',
                fontSize: 13,
                cursor: 'pointer',
                textDecoration: 'none',
              }}
            >
              {item.label}
            </a>
          ))}
        </div>
      </div>
    </div>
  )
}

export default NotFound
