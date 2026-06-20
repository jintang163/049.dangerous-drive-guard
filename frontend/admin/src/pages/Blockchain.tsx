import React, { useState, useRef } from 'react'
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
  Statistic,
  List,
  Divider,
  Descriptions,
  message,
  Select,
  Input,
  Form,
  Upload,
  Progress,
  Alert,
  Tooltip,
  Badge,
} from 'antd'
import {
  SafetyCertificateOutlined,
  LinkOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
  RiseOutlined,
  BlockOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  FileSearchOutlined,
  UploadOutlined,
  ReloadOutlined,
  FilterOutlined,
  ExportOutlined,
  InfoCircleOutlined,
  EyeOutlined,
  FileTextOutlined,
  WarningOutlined,
  FileProtectOutlined,
  AuditOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { formatDateTime } from '@/utils/auth'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select

interface Block {
  id: string
  block_height: number
  block_hash: string
  prev_hash: string
  transaction_count: number
  timestamp: string
  miner: string
  size: string
  nonce: string
  difficulty: string
}

interface EvidenceRecord {
  id: string
  evidence_id: string
  business_type: 'waybill' | 'alarm' | 'score' | 'event'
  business_no: string
  data_hash: string
  chain_time: string
  block_height: number
  operator: string
  data_size: string
  verified?: boolean
}

const businessTypeMap: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  waybill: { label: '运单存证', color: 'blue', icon: <FileTextOutlined /> },
  alarm: { label: '报警存证', color: 'red', icon: <WarningOutlined /> },
  score: { label: '评分存证', color: 'gold', icon: <AuditOutlined /> },
  event: { label: '事件存证', color: 'purple', icon: <FileProtectOutlined /> },
}

const generateHash = () =>
  '0x' + Array.from({ length: 64 }, () => '0123456789abcdef'[Math.floor(Math.random() * 16)]).join('')

const mockBlocks: Block[] = Array.from({ length: 12 }, (_, i) => ({
  id: String(i + 1),
  block_height: 1256893 - i,
  block_hash: generateHash(),
  prev_hash: i === 11 ? '0x0000000000000000000000000000000000000000000000000000000000000000' : generateHash(),
  transaction_count: Math.floor(Math.random() * 120) + 30,
  timestamp: dayjs().subtract(i * 3, 'minute').format('YYYY-MM-DD HH:mm:ss'),
  miner: `Node-${Math.floor(Math.random() * 20) + 1}`,
  size: `${(Math.random() * 2 + 0.5).toFixed(2)} MB`,
  nonce: Math.floor(Math.random() * 999999999).toString(),
  difficulty: (Math.random() * 10 + 15).toFixed(4) + ' T',
}))

const mockEvidences: EvidenceRecord[] = [
  {
    id: '1',
    evidence_id: 'EVI202606200001',
    business_type: 'waybill',
    business_no: 'WB2026062000887',
    data_hash: generateHash(),
    chain_time: '2026-06-20 10:23:45',
    block_height: 1256893,
    operator: '调度员-张三',
    data_size: '2.34 KB',
    verified: true,
  },
  {
    id: '2',
    evidence_id: 'EVI202606200002',
    business_type: 'alarm',
    business_no: 'ALM2026062000156',
    data_hash: generateHash(),
    chain_time: '2026-06-20 10:18:32',
    block_height: 1256892,
    operator: '系统自动',
    data_size: '5.12 KB',
    verified: true,
  },
  {
    id: '3',
    evidence_id: 'EVI202606200003',
    business_type: 'score',
    business_no: 'SCR202606-DR00128',
    data_hash: generateHash(),
    chain_time: '2026-06-20 10:05:11',
    block_height: 1256891,
    operator: '系统自动',
    data_size: '1.87 KB',
  },
  {
    id: '4',
    evidence_id: 'EVI202606200004',
    business_type: 'event',
    business_no: 'EVT2026062000042',
    data_hash: generateHash(),
    chain_time: '2026-06-20 09:52:08',
    block_height: 1256890,
    operator: '安检员-李四',
    data_size: '3.56 KB',
    verified: true,
  },
  {
    id: '5',
    evidence_id: 'EVI202606200005',
    business_type: 'waybill',
    business_no: 'WB2026062000886',
    data_hash: generateHash(),
    chain_time: '2026-06-20 09:40:22',
    block_height: 1256888,
    operator: '调度员-王五',
    data_size: '2.18 KB',
  },
  {
    id: '6',
    evidence_id: 'EVI202606200006',
    business_type: 'alarm',
    business_no: 'ALM2026062000155',
    data_hash: generateHash(),
    chain_time: '2026-06-20 09:28:17',
    block_height: 1256887,
    operator: '系统自动',
    data_size: '4.89 KB',
    verified: true,
  },
  {
    id: '7',
    evidence_id: 'EVI202606190098',
    business_type: 'score',
    business_no: 'SCR202606-DR00089',
    data_hash: generateHash(),
    chain_time: '2026-06-19 22:15:33',
    block_height: 1256521,
    operator: '系统自动',
    data_size: '1.92 KB',
  },
  {
    id: '8',
    evidence_id: 'EVI202606190097',
    business_type: 'event',
    business_no: 'EVT2026061900108',
    data_hash: generateHash(),
    chain_time: '2026-06-19 20:48:59',
    block_height: 1256487,
    operator: '调度员-赵六',
    data_size: '3.21 KB',
    verified: true,
  },
  {
    id: '9',
    evidence_id: 'EVI202606190096',
    business_type: 'waybill',
    business_no: 'WB2026061900756',
    data_hash: generateHash(),
    chain_time: '2026-06-19 18:32:14',
    block_height: 1256421,
    operator: '调度员-张三',
    data_size: '2.45 KB',
  },
  {
    id: '10',
    evidence_id: 'EVI202606190095',
    business_type: 'alarm',
    business_no: 'ALM2026061900234',
    data_hash: generateHash(),
    chain_time: '2026-06-19 16:12:08',
    block_height: 1256358,
    operator: '系统自动',
    data_size: '5.67 KB',
    verified: true,
  },
]

const Blockchain: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [blocks, setBlocks] = useState<Block[]>(mockBlocks)
  const [evidences, setEvidences] = useState<EvidenceRecord[]>(mockEvidences)
  const [blockDetail, setBlockDetail] = useState<Block | null>(null)
  const [verifyModal, setVerifyModal] = useState(false)
  const [verifyForm] = Form.useForm()
  const [verifyResult, setVerifyResult] = useState<{
    success: boolean
    message: string
    record?: EvidenceRecord
    hashMatch?: boolean
    timestamp?: string
    blockHeight?: number
  } | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [businessFilter, setBusinessFilter] = useState<string>()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [blockPage, setBlockPage] = useState(1)
  const fileInputRef = useRef<any>(null)

  const chainStats = {
    blockHeight: blocks[0]?.block_height || 0,
    txCount: blocks.reduce((a, b) => a + b.transaction_count, 0),
    dataSize: (blocks.length * 1.2 + evidences.length * 0.003).toFixed(2),
    todayNew: evidences.filter(e => dayjs(e.chain_time).isSame(dayjs(), 'day')).length,
  }

  const sha256 = async (str: string) => {
    const buf = new TextEncoder().encode(str)
    const hashBuf = await crypto.subtle.digest('SHA-256', buf)
    return '0x' + Array.from(new Uint8Array(hashBuf)).map(b => b.toString(16).padStart(2, '0')).join('')
  }

  const handleVerify = async (values: any) => {
    setVerifying(true)
    setVerifyResult(null)
    try {
      await new Promise(r => setTimeout(r, 1200))
      const inputHash = values.input_hash?.trim()
      const targetHash = values.file_content ? await sha256(values.file_content) : inputHash

      if (!targetHash || targetHash === '0x') {
        setVerifyResult({
          success: false,
          message: '请输入有效的Hash值或上传文件',
        })
        return
      }

      const record = evidences.find(e => e.data_hash.toLowerCase() === targetHash.toLowerCase())
      if (record) {
        setEvidences(prev => prev.map(e => e.id === record.id ? { ...e, verified: true } : e))
        setVerifyResult({
          success: true,
          message: '核验通过：数据Hash与链上存证一致',
          record,
          hashMatch: true,
          timestamp: record.chain_time,
          blockHeight: record.block_height,
        })
      } else {
        const partialMatch = evidences.find(e =>
          e.data_hash.toLowerCase().startsWith(targetHash.slice(0, 16).toLowerCase()) ||
          targetHash.toLowerCase().startsWith(e.data_hash.slice(0, 16).toLowerCase())
        )
        setVerifyResult({
          success: false,
          message: partialMatch
            ? `未找到匹配的存证记录。注意：输入Hash与存证ID "${partialMatch.evidence_id}" 的Hash前缀相似，可能输入有误`
            : '未在链上找到对应的存证记录，该数据可能未上链或已被篡改',
        })
      }
    } finally {
      setVerifying(false)
    }
  }

  const handleFileUpload = (info: any) => {
    const file = info.file
    const reader = new FileReader()
    reader.onload = (e) => {
      const content = e.target?.result as string
      verifyForm.setFieldsValue({ file_content: content })
      message.success(`文件 "${file.name}" 已加载，将自动计算SHA256摘要`)
    }
    reader.readAsText(file)
  }

  const columns = [
    {
      title: '存证ID',
      dataIndex: 'evidence_id',
      width: 160,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '业务类型',
      dataIndex: 'business_type',
      width: 130,
      render: (v: string) => {
        const t = businessTypeMap[v] || { label: v, color: 'default', icon: null }
        return <Tag color={t.color} icon={t.icon} style={{ fontSize: 12 }}>{t.label}</Tag>
      },
    },
    {
      title: '关联业务号',
      dataIndex: 'business_no',
      width: 170,
      render: (v: string) => <Text copyable style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '数据Hash',
      dataIndex: 'data_hash',
      width: 240,
      ellipsis: true,
      render: (v: string) => (
        <Tooltip title={v}>
          <Text code style={{ fontSize: 11 }}>
            {v.slice(0, 10)}...{v.slice(-8)}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '上链时间',
      dataIndex: 'chain_time',
      width: 170,
      render: (v: string) => <Text type="secondary" style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: '区块高度',
      dataIndex: 'block_height',
      width: 110,
      render: (v: number) => <Text strong style={{ fontFamily: 'monospace' }}>#{v.toLocaleString()}</Text>,
    },
    {
      title: '核验状态',
      dataIndex: 'verified',
      width: 100,
      render: (v: boolean) => v ? (
        <Badge status="success" text={<Tag color="green">已核验</Tag>} />
      ) : (
        <Badge status="default" text={<Tag>待核验</Tag>} />
      ),
    },
    {
      title: '操作',
      width: 140,
      fixed: 'right' as const,
      render: (_: any, record: EvidenceRecord) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => {
            setBlockDetail(blocks.find(b => b.block_height === record.block_height) || null)
          }}>区块</Button>
          <Button
            type="link"
            size="small"
            type="primary"
            icon={<FileSearchOutlined />}
            onClick={() => {
              setVerifyModal(true)
              verifyForm.setFieldsValue({ input_hash: record.data_hash })
              setVerifyResult(null)
            }}
          >
            核验
          </Button>
        </Space>
      ),
    },
  ]

  const chainTrendChart = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['上链笔数', '交易总数'], bottom: 0 },
    grid: { left: 40, right: 40, top: 20, bottom: 40 },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 14 }, (_, i) => dayjs().subtract(13 - i, 'day').format('MM-DD')),
    },
    yAxis: [
      { type: 'value', name: '存证' },
      { type: 'value', name: '交易' },
    ],
    series: [
      {
        name: '上链笔数',
        type: 'bar',
        data: Array.from({ length: 14 }, () => Math.floor(Math.random() * 200) + 50),
        itemStyle: { color: 'rgba(22,119,255,0.8)', borderRadius: [4, 4, 0, 0] },
      },
      {
        name: '交易总数',
        type: 'line',
        smooth: true,
        yAxisIndex: 1,
        data: Array.from({ length: 14 }, (_, i) => 4500 + i * 200 + Math.floor(Math.random() * 500)),
        itemStyle: { color: '#52c41a' },
        areaStyle: { color: 'rgba(82,196,26,0.1)' },
      },
    ],
  }

  const businessDistChart = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, type: 'scroll' },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}: {c}' },
      data: Object.entries(businessTypeMap).map(([k, v], i) => ({
        name: v.label,
        value: evidences.filter(e => e.business_type === k).length * 12 + Math.floor(Math.random() * 50) + 10,
        itemStyle: {
          color: ['#1677ff', '#ff4d4f', '#faad14', '#722ed1'][i % 4],
        },
      })),
    }],
  }

  const copyToClipboard = (text: string, label: string = '内容') => {
    navigator.clipboard.writeText(text)
    message.success(`${label}已复制`)
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="区块高度"
              value={chainStats.blockHeight}
              valueStyle={{ color: '#1677ff' }}
              prefix={<BlockOutlined />}
              formatter={(v) => <span>#{Number(v).toLocaleString()}</span>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="交易总数"
              value={chainStats.txCount}
              valueStyle={{ color: '#52c41a' }}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="存证数据量"
              value={chainStats.dataSize}
              suffix="GB"
              valueStyle={{ color: '#722ed1' }}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }}>
            <Statistic
              title="今日新增"
              value={chainStats.todayNew}
              valueStyle={{ color: '#fa8c16' }}
              prefix={<RiseOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={<Space><BlockOutlined style={{ color: '#1677ff' }} /> 区块浏览器</Space>}
        extra={
          <Space>
            <Button size="small" icon={<ReloadOutlined />}>刷新</Button>
          </Space>
        }
      >
        <List
          grid={{ gutter: 12, xs: 1, sm: 2, md: 3, lg: 4, xl: 6 }}
          dataSource={blocks.slice(0, Math.min(blockPage * 6, blocks.length))}
          renderItem={(block) => (
            <List.Item key={block.id}>
              <Card
                size="small"
                style={{
                  borderRadius: 10,
                  cursor: 'pointer',
                  background: 'linear-gradient(135deg, #f0f5ff 0%, #e6f4ff 100%)',
                  border: '1px solid #d6e4ff',
                }}
                hoverable
                onClick={() => setBlockDetail(block)}
                bodyStyle={{ padding: 14 }}
              >
                <Space direction="vertical" size={6} style={{ width: '100%' }}>
                  <Space size={6}>
                    <BlockOutlined style={{ color: '#1677ff' }} />
                    <Text strong style={{ fontFamily: 'monospace', color: '#1677ff' }}>
                      #{block.block_height.toLocaleString()}
                    </Text>
                  </Space>
                  <Tooltip title={block.block_hash}>
                    <Text code style={{ fontSize: 10, wordBreak: 'break-all', lineHeight: 1.4 }}>
                      {block.block_hash.slice(0, 14)}...
                    </Text>
                  </Tooltip>
                  <Divider style={{ margin: '4px 0' }} />
                  <Space size={8} wrap>
                    <Tag color="blue" style={{ margin: 0, fontSize: 11 }}>
                      <ThunderboltOutlined /> {block.transaction_count}
                    </Tag>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      <ClockCircleOutlined /> {dayjs(block.timestamp).fromNow()}
                    </Text>
                  </Space>
                </Space>
              </Card>
            </List.Item>
          )}
        />
        {blockPage * 6 < blocks.length && (
          <div style={{ textAlign: 'center', marginTop: 12 }}>
            <Button onClick={() => setBlockPage(p => p + 1)}>加载更多区块</Button>
          </div>
        )}
      </Card>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><ThunderboltOutlined /> 近14日上链趋势</Space>}>
            <ReactECharts option={chainTrendChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card bordered={false} style={{ borderRadius: 12 }} title={<Space><FileProtectOutlined /> 业务类型分布</Space>}>
            <ReactECharts option={businessDistChart} style={{ height: 220 }} notMerge />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ borderRadius: 12 }}
        title={
          <Space>
            <SafetyCertificateOutlined style={{ color: '#52c41a' }} />
            <Text strong style={{ fontSize: 15 }}>存证记录</Text>
            <Tag color="green">{evidences.length}</Tag>
          </Space>
        }
        extra={
          <Space wrap>
            <Select allowClear placeholder="业务类型" style={{ width: 140 }} value={businessFilter} onChange={setBusinessFilter}>
              {Object.entries(businessTypeMap).map(([k, v]) => (
                <Option key={k} value={k}>{v.label}</Option>
              ))}
            </Select>
            <Button icon={<FilterOutlined />} onClick={() => setBusinessFilter(undefined)}>重置</Button>
            <Button icon={<ReloadOutlined />}>刷新</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
            <Button
              type="primary"
              icon={<FileSearchOutlined />}
              onClick={() => { setVerifyModal(true); verifyForm.resetFields(); setVerifyResult(null) }}
            >
              Hash核验
            </Button>
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          columns={columns as any}
          dataSource={evidences.filter(e => !businessFilter || e.business_type === businessFilter)}
          pagination={{
            current: page,
            pageSize,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) },
          }}
          scroll={{ x: 1280 }}
        />
      </Card>

      <Modal
        title={
          <Space>
            <BlockOutlined style={{ color: '#1677ff' }} />
            <Text strong>区块详情</Text>
            <Tag color="blue">#{blockDetail?.block_height.toLocaleString()}</Tag>
          </Space>
        }
        open={!!blockDetail}
        onCancel={() => setBlockDetail(null)}
        width={720}
        footer={[
          <Button key="close" onClick={() => setBlockDetail(null)}>关闭</Button>,
        ]}
      >
        {blockDetail && (
          <div>
            <Alert
              type="info"
              showIcon
              icon={<SafetyCertificateOutlined />}
              message={`区块已确认：${blockDetail.transaction_count} 笔交易已永久写入链上，不可篡改`}
              style={{ borderRadius: 8, marginBottom: 16 }}
            />

            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="区块高度" span={1}>
                <Space>
                  <Text strong style={{ fontFamily: 'monospace', fontSize: 16 }}>
                    #{blockDetail.block_height.toLocaleString()}
                  </Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="出块时间" span={1}>{blockDetail.timestamp}</Descriptions.Item>
              <Descriptions.Item label="区块Hash" span={2}>
                <Space style={{ width: '100%' }}>
                  <Text code style={{ wordBreak: 'break-all', fontSize: 11, flex: 1 }}>
                    {blockDetail.block_hash}
                  </Text>
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(blockDetail.block_hash, '区块Hash')}
                  />
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="前块Hash" span={2}>
                <Space style={{ width: '100%' }}>
                  <Text code style={{ wordBreak: 'break-all', fontSize: 11, flex: 1 }}>
                    {blockDetail.prev_hash}
                  </Text>
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(blockDetail.prev_hash, '前块Hash')}
                  />
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="交易数量" span={1}>
                <Tag color="blue"><ThunderboltOutlined /> {blockDetail.transaction_count} 笔</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="区块大小" span={1}>{blockDetail.size}</Descriptions.Item>
              <Descriptions.Item label="出块节点" span={1}>{blockDetail.miner}</Descriptions.Item>
              <Descriptions.Item label="Nonce" span={1} style={{ fontFamily: 'monospace' }}>{blockDetail.nonce}</Descriptions.Item>
              <Descriptions.Item label="挖矿难度" span={2}>{blockDetail.difficulty}</Descriptions.Item>
            </Descriptions>

            <Divider />

            <Card size="small" style={{ borderRadius: 8 }} title={<Space><LinkOutlined /> 包含的存证记录</Space>}>
              {evidences.filter(e => e.block_height === blockDetail.block_height).length === 0 ? (
                <div style={{ color: '#8c8c8c', padding: '16px 0', textAlign: 'center' }}>
                  <InfoCircleOutlined /> 该区块暂无系统相关的业务存证
                </div>
              ) : (
                <List
                  size="small"
                  dataSource={evidences.filter(e => e.block_height === blockDetail.block_height)}
                  renderItem={(e) => {
                    const t = businessTypeMap[e.business_type]
                    return (
                      <List.Item key={e.id}>
                        <List.Item.Meta
                          avatar={<Tag color={t.color} icon={t.icon}>{t.label}</Tag>}
                          title={<Text strong>{e.evidence_id}</Text>}
                          description={<Text code style={{ fontSize: 11 }}>{e.business_no}</Text>}
                        />
                      </List.Item>
                    )
                  }}
                />
              )}
            </Card>
          </div>
        )}
      </Modal>

      <Modal
        title={
          <Space>
            <FileSearchOutlined style={{ color: '#52c41a' }} />
            <Text strong>SHA256 数据核验</Text>
          </Space>
        }
        open={verifyModal}
        onCancel={() => { setVerifyModal(false); verifyForm.resetFields(); setVerifyResult(null) }}
        width={620}
        footer={null}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          icon={<InfoCircleOutlined />}
          message="支持两种核验方式：直接粘贴Hash值，或上传原始文件自动计算摘要后校验"
          style={{ borderRadius: 8, marginBottom: 16 }}
        />

        <Form
          form={verifyForm}
          layout="vertical"
          onFinish={handleVerify}
        >
          <Form.Item label="方式一：输入Hash值" name="input_hash">
            <Input.TextArea
              rows={3}
              placeholder="请粘贴SHA256哈希值 (以0x开头的66位字符，如 0x1234abcd...)"
              style={{ fontFamily: 'monospace', fontSize: 12 }}
              allowClear
            />
          </Form.Item>

          <div style={{ textAlign: 'center', color: '#8c8c8c', margin: '4px 0 12px' }}>
            ——— 或 ———
          </div>

          <Form.Item label="方式二：上传原始文件" name="file_content" style={{ display: 'none' }} />
          <Form.Item label="上传文件">
            <Upload.Dragger
              ref={fileInputRef}
              beforeUpload={(file) => {
                handleFileUpload({ file })
                return false
              }}
              maxCount={1}
              accept="*"
            >
              <p className="ant-upload-drag-icon"><UploadOutlined /></p>
              <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
              <p className="ant-upload-hint">
                支持任意类型文件，系统将自动计算SHA256摘要并与链上比对
              </p>
            </Upload.Dragger>
          </Form.Item>

          <Space style={{ width: '100%', justifyContent: 'center' }}>
            <Button onClick={() => { verifyForm.resetFields(); setVerifyResult(null) }}>
              清空
            </Button>
            <Button
              type="primary"
              icon={<FileSearchOutlined />}
              htmlType="submit"
              loading={verifying}
            >
              开始核验
            </Button>
          </Space>
        </Form>

        {verifyResult && (
          <>
            <Divider style={{ margin: '20px 0 12px' }} />

            <Card
              size="small"
              style={{
                borderRadius: 8,
                background: verifyResult.success ? '#f6ffed' : '#fff1f0',
                border: `1px solid ${verifyResult.success ? '#b7eb8f' : '#ffa39e'}`,
              }}
              title={
                <Space>
                  {verifyResult.success ? (
                    <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 18 }} />
                  ) : (
                    <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 18 }} />
                  )}
                  <Text strong style={{ color: verifyResult.success ? '#52c41a' : '#ff4d4f' }}>
                    {verifyResult.success ? '核验通过 ✓' : '核验未通过 ✗'}
                  </Text>
                </Space>
              }
            >
              <Paragraph style={{ margin: 0, marginBottom: 12 }}>
                {verifyResult.message}
              </Paragraph>

              {verifyResult.success && verifyResult.record && (
                <Descriptions column={2} bordered size="small">
                  <Descriptions.Item label="存证ID">{verifyResult.record.evidence_id}</Descriptions.Item>
                  <Descriptions.Item label="业务类型">
                    {businessTypeMap[verifyResult.record.business_type]?.label}
                  </Descriptions.Item>
                  <Descriptions.Item label="关联业务号">{verifyResult.record.business_no}</Descriptions.Item>
                  <Descriptions.Item label="操作人">{verifyResult.record.operator}</Descriptions.Item>
                  <Descriptions.Item label="上链时间" span={1}>{verifyResult.timestamp}</Descriptions.Item>
                  <Descriptions.Item label="区块高度" span={1}>
                    #{verifyResult.blockHeight?.toLocaleString()}
                  </Descriptions.Item>
                  <Descriptions.Item label="链上Hash" span={2}>
                    <Tooltip title={verifyResult.record.data_hash}>
                      <Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
                        {verifyResult.record.data_hash.slice(0, 24)}...{verifyResult.record.data_hash.slice(-16)}
                      </Text>
                    </Tooltip>
                  </Descriptions.Item>
                </Descriptions>
              )}

              {!verifyResult.success && (
                <div style={{ marginTop: 12 }}>
                  <Progress
                    percent={0}
                    showInfo={false}
                    strokeColor="#ff4d4f"
                    style={{ marginBottom: 8 }}
                  />
                  <Alert
                    type="warning"
                    showIcon
                    message="建议操作"
                    description={
                      <ul style={{ margin: 0, paddingLeft: 18 }}>
                        <li>检查输入的Hash值是否完整（应为64或66位十六进制）</li>
                        <li>确认上传的文件与存证时使用的完全一致</li>
                        <li>核实业务是否已成功上链（可联系管理员）</li>
                      </ul>
                    }
                    style={{ borderRadius: 6 }}
                  />
                </div>
              )}
            </Card>
          </>
        )}
      </Modal>
    </div>
  )
}

export default Blockchain
