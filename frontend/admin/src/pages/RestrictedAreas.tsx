import React, { useState, useEffect, useRef } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  DatePicker,
  Tabs,
  Drawer,
  Descriptions,
  message,
  Popconfirm,
  Row,
  Col,
  Upload,
  Tooltip,
  Timeline,
  Empty,
  Badge,
  Radio,
  Checkbox,
  TimePicker,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  ExclamationCircleOutlined,
  UploadOutlined,
  FundProjectionScreenOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
  FileTextOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { TabsProps } from 'antd'
import AMap from '@/components/AMap'
import {
  restrictedAreaApi,
  type RestrictedAreaItem,
  type RestrictedAreaTemplate,
  type TimeScheduleRule,
  type ApprovalRecord,
} from '@/services/api'
import { PageResult } from '@/services/api'
import dayjs, { Dayjs } from 'dayjs'
import { getUserInfo, hasPermission } from '@/utils/auth'

const { RangePicker } = DatePicker
const { TextArea } = Input
const { Option } = Select
const { Group: CheckboxGroup } = Checkbox

const AREA_TYPE_OPTIONS = [
  { value: 'school', label: '学校', color: 'blue' },
  { value: 'hospital', label: '医院', color: 'red' },
  { value: 'mall', label: '商圈', color: 'orange' },
  { value: 'water_protection', label: '水源地', color: 'cyan' },
  { value: 'tunnel', label: '隧道', color: 'purple' },
  { value: 'bridge', label: '桥梁', color: 'gold' },
  { value: 'height_limit', label: '限高路段', color: 'magenta' },
  { value: 'weight_limit', label: '限重路段', color: 'geekblue' },
]

const TEMPLATE_CATEGORY_OPTIONS = [
  { value: 'hospital', label: '医院模板', icon: '🏥', color: 'red' },
  { value: 'school', label: '学校模板', icon: '🏫', color: 'blue' },
  { value: 'mall', label: '商圈模板', icon: '🏬', color: 'orange' },
  { value: 'water_source', label: '水源地模板', icon: '💧', color: 'cyan' },
  { value: 'custom', label: '自定义模板', icon: '📋', color: 'default' },
]

const APPROVAL_STATUS_MAP: Record<number, { text: string; color: string }> = {
  0: { text: '待提交', color: 'default' },
  1: { text: '一级审批中', color: 'processing' },
  2: { text: '二级审批中', color: 'processing' },
  3: { text: '已通过', color: 'success' },
  4: { text: '已拒绝', color: 'error' },
  5: { text: '已撤销', color: 'warning' },
}

const WEEKDAYS = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 },
]

type DrawMode = 'none' | 'polygon' | 'circle'

const RestrictedAreas: React.FC = () => {
  const user = getUserInfo()
  const isAdmin = user?.role === 'admin'

  const [activeTab, setActiveTab] = useState('list')
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<RestrictedAreaItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState({
    area_type: '',
    is_temporary: undefined as number | undefined,
    approval_status: undefined as number | undefined,
    keyword: '',
  })

  const [detailVisible, setDetailVisible] = useState(false)
  const [currentItem, setCurrentItem] = useState<RestrictedAreaItem | null>(null)

  const [modalVisible, setModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create')
  const [form] = Form.useForm()

  const [drawMode, setDrawMode] = useState<DrawMode>('none')
  const [polygonPoints, setPolygonPoints] = useState<[number, number][]>([])
  const [circleCenter, setCircleCenter] = useState<[number, number] | null>(null)
  const [circleRadius, setCircleRadius] = useState<number>(500)
  const mapRef = useRef<any>(null)
  const amapRef = useRef<any>(null)
  const drawingRef = useRef<any>(null)

  const [templateData, setTemplateData] = useState<RestrictedAreaTemplate[]>([])
  const [templateLoading, setTemplateLoading] = useState(false)

  const [approvalHistory, setApprovalHistory] = useState<ApprovalRecord[]>([])
  const [approvalLoading, setApprovalLoading] = useState(false)

  const [pendingData, setPendingData] = useState<RestrictedAreaItem[]>([])
  const [pendingTotal, setPendingTotal] = useState(0)
  const [pendingPage, setPendingPage] = useState(1)
  const [pendingLoading, setPendingLoading] = useState(false)

  useEffect(() => {
    if (activeTab === 'list') {
      fetchList()
    } else if (activeTab === 'templates') {
      fetchTemplates()
    } else if (activeTab === 'approvals') {
      fetchPendingApprovals()
    }
  }, [activeTab, page, pageSize, filters, pendingPage])

  const fetchList = async () => {
    setLoading(true)
    try {
      const res = await restrictedAreaApi.list({
        page,
        page_size: pageSize,
        ...filters,
      })
      setData(res.list || [])
      setTotal(res.total || 0)
    } catch (e) {
      message.error('获取禁行区域列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchTemplates = async () => {
    setTemplateLoading(true)
    try {
      const res = await restrictedAreaApi.listTemplates({ page: 1, page_size: 100 })
      setTemplateData(res.list || [])
    } catch (e) {
      message.error('获取模板列表失败')
    } finally {
      setTemplateLoading(false)
    }
  }

  const fetchPendingApprovals = async () => {
    setPendingLoading(true)
    try {
      const level = isAdmin ? 2 : 1
      const res = await restrictedAreaApi.listPendingApprovals({ level, page: pendingPage, page_size: 10 })
      setPendingData(res.list || [])
      setPendingTotal(res.total || 0)
    } catch (e) {
      message.error('获取待审批列表失败')
    } finally {
      setPendingLoading(false)
    }
  }

  const handleCreate = () => {
    setModalMode('create')
    setDrawMode('none')
    setPolygonPoints([])
    setCircleCenter(null)
    form.resetFields()
    form.setFieldsValue({
      level: 2,
      shape_type: 'polygon',
      is_temporary: 0,
      submit_approval: true,
    })
    setModalVisible(true)
  }

  const handleEdit = (record: RestrictedAreaItem) => {
    setModalMode('edit')
    setDrawMode('none')
    setCurrentItem(record)
    setPolygonPoints([])
    setCircleCenter(null)

    const schedule = record.time_schedule || []

    form.setFieldsValue({
      ...record,
      time_schedule: schedule,
      effective_period: record.effective_from && record.effective_to
        ? [dayjs(record.effective_from), dayjs(record.effective_to)]
        : null,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await restrictedAreaApi.delete(id)
      message.success('删除成功')
      fetchList()
    } catch (e) {
      message.error('删除失败')
    }
  }

  const handleDetail = async (record: RestrictedAreaItem) => {
    setCurrentItem(record)
    try {
      approvalLoading = true
      const res = await restrictedAreaApi.getApprovalHistory(record.id)
      setApprovalHistory(res.list || [])
    } catch (e) {
      // ignore
    } finally {
      setApprovalLoading(false)
    }
    setDetailVisible(true)
  }

  const handleSubmitApproval = async (id: number) => {
    try {
      await restrictedAreaApi.submitApproval(id)
      message.success('已提交审批')
      fetchList()
    } catch (e) {
      message.error('提交审批失败')
    }
  }

  const handleApprove = async (record: RestrictedAreaItem, level: number) => {
    Modal.confirm({
      title: level === 1 ? '确认通过一级审批？' : '确认通过二级审批并生效？',
      content: (
        <div>
          <p>区域名称：{record.name}</p>
          <Input.TextArea
            id="approval_note"
            placeholder="请输入审批意见（可选）"
            rows={3}
            style={{ marginTop: 8 }}
          />
        </div>
      ),
      onOk: async () => {
        try {
          const note = (document.getElementById('approval_note') as HTMLTextAreaElement)?.value
          if (level === 1) {
            await restrictedAreaApi.approveFirst(record.id, { approval_note: note })
          } else {
            await restrictedAreaApi.approveSecond(record.id, { approval_note: note })
          }
          message.success('审批通过')
          fetchList()
          fetchPendingApprovals()
        } catch (e) {
          message.error('审批失败')
        }
      },
    })
  }

  const handleReject = async (record: RestrictedAreaItem, level: number) => {
    Modal.confirm({
      title: '确认拒绝审批？',
      content: (
        <div>
          <p>区域名称：{record.name}</p>
          <Input.TextArea
            id="reject_note"
            placeholder="请输入拒绝原因"
            rows={3}
            style={{ marginTop: 8 }}
          />
        </div>
      ),
      onOk: async () => {
        try {
          const note = (document.getElementById('reject_note') as HTMLTextAreaElement)?.value
          await restrictedAreaApi.reject(record.id, level, { approval_note: note })
          message.success('已拒绝')
          fetchList()
          fetchPendingApprovals()
        } catch (e) {
          message.error('操作失败')
        }
      },
    })
  }

  const handleRevoke = async (record: RestrictedAreaItem) => {
    Modal.confirm({
      title: '确认撤销该禁行区域？',
      content: '撤销后该区域将不再生效，此操作不可恢复。',
      okType: 'danger',
      onOk: async () => {
        try {
          await restrictedAreaApi.revoke(record.id, {})
          message.success('已撤销')
          fetchList()
        } catch (e) {
          message.error('撤销失败')
        }
      },
    })
  }

  const handleApplyTemplate = async (tpl: RestrictedAreaTemplate) => {
    Modal.confirm({
      title: `应用模板：${tpl.template_name}`,
      content: (
        <div style={{ marginBottom: 8 }}>
          <p style={{ marginBottom: 8 }}>请在地图上点击选择区域中心点：</p>
          <div style={{ height: 300, width: '100%' }}>
            <AMap
              onMapLoaded={(map, AMap) => {
                mapRef.current = map
                amapRef.current = AMap
                map.on('click', (e: any) => {
                  const lng = e.lnglat.getLng()
                  const lat = e.lnglat.getLat()
                  ;(window as any).__templateCenter = { lat, lng }
                  message.success(`已选择中心点: ${lat.toFixed(6)}, ${lng.toFixed(6)}`)
                })
              }}
              zoom={12}
            />
          </div>
        </div>
      ),
      onOk: async () => {
        const center = (window as any).__templateCenter
        if (!center) {
          message.warning('请先在地图上选择中心点')
          return Promise.reject()
        }
        try {
          const area = await restrictedAreaApi.applyTemplate(tpl.id, {
            center_lat: center.lat,
            center_lng: center.lng,
            name: tpl.template_name,
          })
          message.success('模板应用成功，请在编辑中完善信息')
          setModalMode('edit')
          setCurrentItem(area as any)
          form.setFieldsValue({ ...area })
          setModalVisible(true)
        } catch (e) {
          message.error('应用模板失败')
        }
      },
    })
  }

  const [gisSourceType, setGisSourceType] = useState<string>('geojson')
  const [gisFileList, setGisFileList] = useState<any[]>([])
  const [gisImporting, setGisImporting] = useState(false)

  const handleGisImport = () => {
    Modal.confirm({
      title: '导入官方危险品运输禁行路网',
      width: 520,
      content: (
        <div style={{ marginTop: 12 }}>
          <p style={{ marginBottom: 8, color: '#595959' }}>选择GIS数据源类型：</p>
          <Select
            value={gisSourceType}
            onChange={setGisSourceType}
            style={{ width: '100%', marginBottom: 12 }}
          >
            <Option value="geojson">GeoJSON数据</Option>
            <Option value="shp">Shapefile数据（ZIP压缩包）</Option>
            <Option value="official_road_network">官方危险品禁行路网</Option>
          </Select>
          <Upload
            beforeUpload={(file) => {
              setGisFileList([file])
              return false
            }}
            onRemove={() => { setGisFileList([]) }}
            maxCount={1}
            accept={gisSourceType === 'shp' ? '.zip' : '.geojson,.json'}
            fileList={gisFileList}
          >
            <Button icon={<UploadOutlined />}>选择GIS数据文件</Button>
          </Upload>
          {gisSourceType === 'shp' && (
            <p style={{ marginTop: 8, fontSize: 12, color: '#8c8c8c' }}>
              Shapefile需以ZIP压缩包形式上传（含.shp/.dbf/.shx），建议先通过ogr2ogr转换为GeoJSON
            </p>
          )}
          <p style={{ marginTop: 8, fontSize: 12, color: '#fa8c16' }}>
            ⚠️ 导入的区域需经二级审批后方可生效
          </p>
        </div>
      ),
      okText: '开始导入',
      confirmLoading: gisImporting,
      onOk: async () => {
        if (gisFileList.length === 0) {
          message.warning('请先选择GIS数据文件')
          return Promise.reject()
        }
        setGisImporting(true)
        try {
          await restrictedAreaApi.importGisFile(gisFileList[0] as any, gisSourceType)
          message.success('GIS数据导入成功，已创建待审批禁行区域')
          setGisFileList([])
          fetchList()
        } catch (e: any) {
          message.error(e?.message || '导入失败')
          return Promise.reject()
        } finally {
          setGisImporting(false)
        }
      },
    })
  }

  const handleMapLoaded = (map: any, AMap: any) => {
    mapRef.current = map
    amapRef.current = AMap
  }

  const handleMapClick = (lng: number, lat: number) => {
    if (drawMode === 'polygon') {
      const newPoints = [...polygonPoints, [lng, lat] as [number, number]]
      setPolygonPoints(newPoints)
    } else if (drawMode === 'circle') {
      setCircleCenter([lng, lat])
    }
  }

  const clearDrawing = () => {
    setDrawMode('none')
    setPolygonPoints([])
    setCircleCenter(null)
  }

  const handleFormSubmit = async (values: any) => {
    const shapeType = values.shape_type

    let boundaryPolygon: any = null
    if (shapeType === 'polygon') {
      if (polygonPoints.length < 3) {
        message.error('请至少绘制3个点的多边形区域')
        return
      }
      const closedPoints = [...polygonPoints, polygonPoints[0]]
      boundaryPolygon = {
        type: 'Polygon',
        coordinates: [closedPoints],
      }
      values.center_latitude = polygonPoints.reduce((s, p) => s + p[1], 0) / polygonPoints.length
      values.center_longitude = polygonPoints.reduce((s, p) => s + p[0], 0) / polygonPoints.length
    } else if (shapeType === 'circle') {
      if (!circleCenter) {
        message.error('请在地图上点击选择圆心位置')
        return
      }
      values.center_latitude = circleCenter[1]
      values.center_longitude = circleCenter[0]
      values.radius = circleRadius
      boundaryPolygon = generateCircleGeoJSON(circleCenter[1], circleCenter[0], circleRadius)
    }

    values.boundary_polygon = boundaryPolygon

    if (values.effective_period) {
      values.effective_from = values.effective_period[0]?.toISOString()
      values.effective_to = values.effective_period[1]?.toISOString()
      delete values.effective_period
    }

    try {
      if (modalMode === 'create') {
        await restrictedAreaApi.create(values)
        message.success('创建成功')
      } else {
        await restrictedAreaApi.update(currentItem!.id, values)
        message.success('更新成功')
      }
      setModalVisible(false)
      clearDrawing()
      fetchList()
    } catch (e) {
      message.error(modalMode === 'create' ? '创建失败' : '更新失败')
    }
  }

  const generateCircleGeoJSON = (lat: number, lng: number, radius: number) => {
    const points: [number, number][] = []
    const sides = 36
    for (let i = 0; i < sides; i++) {
      const angle = (i * 2 * Math.PI) / sides
      const dx = (radius * Math.cos(angle)) / (111320 * Math.cos((lat * Math.PI) / 180))
      const dy = (radius * Math.sin(angle)) / 110540
      points.push([lng + dx, lat + dy])
    }
    points.push(points[0])
    return {
      type: 'Polygon',
      coordinates: [points],
    }
  }

  const displayPolygons = data
    .filter(item => item.approval_status === 3 && item.status === 1)
    .map(item => {
      const coords = item.boundary_polygon?.coordinates?.[0] || []
      const areaColor = AREA_TYPE_OPTIONS.find(o => o.value === item.area_type)?.color || '#ff4d4f'
      return {
        path: coords,
        fillColor: areaColor,
        strokeColor: areaColor,
        fillOpacity: 0.15,
        strokeWeight: 2,
      }
    })

  if (drawMode === 'polygon' && polygonPoints.length > 0) {
    displayPolygons.push({
      path: [...polygonPoints, polygonPoints[0]] as [number, number][],
      fillColor: '#1677ff',
      strokeColor: '#1677ff',
      fillOpacity: 0.3,
      strokeWeight: 3,
    })
  }

  if (drawMode === 'circle' && circleCenter) {
    const circleCoords = generateCircleGeoJSON(circleCenter[1], circleCenter[0], circleRadius).coordinates[0]
    displayPolygons.push({
      path: circleCoords,
      fillColor: '#faad14',
      strokeColor: '#faad14',
      fillOpacity: 0.3,
      strokeWeight: 3,
    })
  }

  const areaColumns: ColumnsType<RestrictedAreaItem> = [
    {
      title: '区域名称',
      dataIndex: 'name',
      width: 160,
      render: (text, record) => (
        <Space>
          <SafetyCertificateOutlined style={{ color: AREA_TYPE_OPTIONS.find(o => o.value === record.area_type)?.color }} />
          {text}
          {record.is_temporary === 1 && <Tag color="orange">临时</Tag>}
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'area_type',
      width: 100,
      render: (type) => {
        const opt = AREA_TYPE_OPTIONS.find(o => o.value === type)
        return <Tag color={opt?.color as any}>{opt?.label || type}</Tag>
      },
    },
    {
      title: '形状',
      dataIndex: 'shape_type',
      width: 80,
      render: (type) => (type === 'circle' ? '圆形' : '多边形'),
    },
    {
      title: '半径',
      dataIndex: 'radius',
      width: 80,
      render: (v) => (v ? `${v}米` : '-'),
    },
    {
      title: '地址',
      dataIndex: 'address',
      width: 180,
      ellipsis: true,
    },
    {
      title: '审批状态',
      dataIndex: 'approval_status',
      width: 110,
      render: (status) => {
        const info = APPROVAL_STATUS_MAP[status] || { text: '未知', color: 'default' }
        return <Tag color={info.color as any}>{info.text}</Tag>
      },
    },
    {
      title: '生效状态',
      dataIndex: 'status',
      width: 80,
      render: (status, record) => (
        <Badge
          status={record.approval_status === 3 && status === 1 ? 'success' : 'default'}
          text={record.approval_status === 3 && status === 1 ? '生效中' : '未生效'}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 160,
      render: (t) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 280,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleDetail(record)}>
            详情
          </Button>
          {record.approval_status === 0 && (
            <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
          )}
          {record.approval_status === 0 && (
            <Button type="link" size="small" onClick={() => handleSubmitApproval(record.id)}>
              提交审批
            </Button>
          )}
          {record.approval_status === 3 && (
            <Button type="link" size="small" danger onClick={() => handleRevoke(record)}>
              撤销
            </Button>
          )}
          {isAdmin && (
            <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  const pendingColumns: ColumnsType<RestrictedAreaItem> = [
    { title: '区域名称', dataIndex: 'name', width: 160 },
    {
      title: '类型',
      dataIndex: 'area_type',
      width: 100,
      render: (type) => {
        const opt = AREA_TYPE_OPTIONS.find(o => o.value === type)
        return <Tag color={opt?.color as any}>{opt?.label || type}</Tag>
      },
    },
    {
      title: '临时/永久',
      dataIndex: 'is_temporary',
      width: 100,
      render: (v) => (v === 1 ? <Tag color="orange">临时</Tag> : <Tag>永久</Tag>),
    },
    { title: '地址', dataIndex: 'address', ellipsis: true },
    { title: '创建时间', dataIndex: 'created_at', width: 160, render: (t) => dayjs(t).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作',
      width: 200,
      fixed: 'right',
      render: (_, record) => {
        const level = isAdmin ? 2 : 1
        return (
          <Space size="small">
            <Button type="link" size="small" icon={<CheckOutlined />} onClick={() => handleApprove(record, level)}>
              通过
            </Button>
            <Button type="link" size="small" danger icon={<CloseOutlined />} onClick={() => handleReject(record, level)}>
              拒绝
            </Button>
          </Space>
        )
      },
    },
  ]

  const tabItems: TabsProps['items'] = [
    {
      key: 'list',
      label: (
        <span>
          <FundProjectionScreenOutlined /> 禁行区域管理
        </span>
      ),
      children: (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Card size="small" title="区域总览">
            <AMap
              style={{ height: 380, borderRadius: 8 }}
              polygons={displayPolygons}
              zoom={11}
              center={[116.4074, 39.9042]}
              onMapLoaded={handleMapLoaded}
            />
          </Card>

          <Card
            size="small"
            title="区域列表"
            extra={
              <Space>
                <Select
                  placeholder="区域类型"
                  allowClear
                  style={{ width: 120 }}
                  value={filters.area_type || undefined}
                  onChange={(v) => setFilters({ ...filters, area_type: v || '' })}
                >
                  {AREA_TYPE_OPTIONS.map(o => (
                    <Option key={o.value} value={o.value}>{o.label}</Option>
                  ))}
                </Select>
                <Select
                  placeholder="临时/永久"
                  allowClear
                  style={{ width: 120 }}
                  value={filters.is_temporary}
                  onChange={(v) => setFilters({ ...filters, is_temporary: v })}
                >
                  <Option value={0}>永久</Option>
                  <Option value={1}>临时</Option>
                </Select>
                <Select
                  placeholder="审批状态"
                  allowClear
                  style={{ width: 130 }}
                  value={filters.approval_status}
                  onChange={(v) => setFilters({ ...filters, approval_status: v })}
                >
                  {Object.entries(APPROVAL_STATUS_MAP).map(([k, v]) => (
                    <Option key={k} value={Number(k)}>{v.text}</Option>
                  ))}
                </Select>
                <Input.Search
                  placeholder="搜索名称/地址"
                  style={{ width: 180 }}
                  allowClear
                  onSearch={(v) => setFilters({ ...filters, keyword: v })}
                />
                <Button icon={<ReloadOutlined />} onClick={fetchList}>刷新</Button>
                <Button type="primary" icon={<UploadOutlined />} onClick={handleGisImport}>
                  导入GIS
                </Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                  新建区域
                </Button>
              </Space>
            }
          >
            <Table
              rowKey="id"
              loading={loading}
              columns={areaColumns}
              dataSource={data}
              pagination={{
                current: page,
                pageSize,
                total,
                showSizeChanger: true,
                onChange: (p, ps) => { setPage(p); setPageSize(ps) },
              }}
              scroll={{ x: 1200 }}
            />
          </Card>
        </Space>
      ),
    },
    {
      key: 'templates',
      label: (
        <span>
          <FileTextOutlined /> 禁行模板库
        </span>
      ),
      children: (
        <Card
          size="small"
          title="模板列表（预设医院、学校、商圈、水源地四类模板）"
          extra={
            isAdmin && (
              <Button type="primary" icon={<PlusOutlined />}>
                新建模板
              </Button>
            )
          }
        >
          <Row gutter={[16, 16]}>
            {templateData.map((tpl) => {
              const catOpt = TEMPLATE_CATEGORY_OPTIONS.find(c => c.value === tpl.template_category)
              return (
                <Col xs={24} sm={12} md={8} lg={6} key={tpl.id}>
                  <Card
                    hoverable
                    size="small"
                    style={{
                      borderLeft: `4px solid ${catOpt?.color === 'default' ? '#d9d9d9' : ''}`,
                    }}
                    styles={{ body: { padding: 16 } }}
                  >
                    <Space direction="vertical" size={8} style={{ width: '100%' }}>
                      <Space>
                        <span style={{ fontSize: 20 }}>{catOpt?.icon || '📋'}</span>
                        <strong>{tpl.template_name}</strong>
                        {tpl.is_builtin === 1 && <Tag color="blue">内置</Tag>}
                      </Space>
                      <Descriptions size="small" column={1} colon={false}>
                        <Descriptions.Item label="类型">{AREA_TYPE_OPTIONS.find(o => o.value === tpl.area_type)?.label || tpl.area_type}</Descriptions.Item>
                        <Descriptions.Item label="默认半径">{tpl.default_radius}米</Descriptions.Item>
                        <Descriptions.Item label="危险等级">{tpl.level === 1 ? '建议绕行' : '必须避开'}</Descriptions.Item>
                      </Descriptions>
                      <div style={{ fontSize: 12, color: '#8c8c8c', minHeight: 32 }}>
                        {tpl.description}
                      </div>
                      <Button
                        type="primary"
                        block
                        icon={<PlusOutlined />}
                        onClick={() => handleApplyTemplate(tpl)}
                      >
                        应用模板创建区域
                      </Button>
                    </Space>
                  </Card>
                </Col>
              )
            })}
          </Row>
          {templateData.length === 0 && !templateLoading && <Empty description="暂无模板" />}
        </Card>
      ),
    },
    {
      key: 'approvals',
      label: (
        <span>
          <SafetyCertificateOutlined /> 审批中心
          {pendingTotal > 0 && <Badge count={pendingTotal} style={{ marginLeft: 8 }} />}
        </span>
      ),
      children: (
        <Card size="small" title={`待我审批（${isAdmin ? '二级审批' : '一级审批'}）`}>
          <Table
            rowKey="id"
            loading={pendingLoading}
            columns={pendingColumns}
            dataSource={pendingData}
            pagination={{
              current: pendingPage,
              pageSize: 10,
              total: pendingTotal,
              onChange: setPendingPage,
            }}
            scroll={{ x: 1000 }}
          />
          {pendingData.length === 0 && !pendingLoading && <Empty description="暂无待审批事项" />}
        </Card>
      ),
    },
  ]

  return (
    <div>
      <Card bodyStyle={{ padding: 0 }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ padding: '0 16px' }}
        />
      </Card>

      <Drawer
        title="禁行区域详情"
        width={560}
        open={detailVisible}
        onClose={() => setDetailVisible(false)}
      >
        {currentItem && (
          <Space direction="vertical" size={24} style={{ width: '100%' }}>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="区域名称">{currentItem.name}</Descriptions.Item>
              <Descriptions.Item label="区域类型">
                <Tag color={AREA_TYPE_OPTIONS.find(o => o.value === currentItem.area_type)?.color as any}>
                  {AREA_TYPE_OPTIONS.find(o => o.value === currentItem.area_type)?.label || currentItem.area_type}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="区域形状">
                {currentItem.shape_type === 'circle' ? '圆形' : '多边形'}
              </Descriptions.Item>
              <Descriptions.Item label="半径">{currentItem.radius ? `${currentItem.radius}米` : '-'}</Descriptions.Item>
              <Descriptions.Item label="危险等级">
                {currentItem.level === 1 ? '建议绕行' : '必须避开'}
              </Descriptions.Item>
              <Descriptions.Item label="是否临时">
                {currentItem.is_temporary === 1 ? <Tag color="orange">是</Tag> : '否'}
              </Descriptions.Item>
              {currentItem.is_temporary === 1 && (
                <Descriptions.Item label="临时原因">{currentItem.temp_reason || '-'}</Descriptions.Item>
              )}
              <Descriptions.Item label="地址">{currentItem.address || '-'}</Descriptions.Item>
              <Descriptions.Item label="中心坐标">
                {currentItem.center_latitude?.toFixed(6)}, {currentItem.center_longitude?.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="限制危险品类别">{currentItem.restrict_hazard_classes || '全部'}</Descriptions.Item>
              <Descriptions.Item label="生效时间段">
                {currentItem.time_schedule && currentItem.time_schedule.length > 0
                  ? currentItem.time_schedule.map((r, i) => (
                      <div key={i}>
                        {r.weekdays?.map(w => WEEKDAYS.find(d => d.value === w)?.label).join('、') || '每天'}
                        ：{r.start_time} - {r.end_time}
                      </div>
                    ))
                  : '全天生效'}
              </Descriptions.Item>
              <Descriptions.Item label="有效期">
                {currentItem.effective_from && currentItem.effective_to
                  ? `${dayjs(currentItem.effective_from).format('YYYY-MM-DD')} 至 ${dayjs(currentItem.effective_to).format('YYYY-MM-DD')}`
                  : '长期有效'}
              </Descriptions.Item>
              <Descriptions.Item label="审批状态">
                <Tag color={APPROVAL_STATUS_MAP[currentItem.approval_status]?.color as any}>
                  {APPROVAL_STATUS_MAP[currentItem.approval_status]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="生效状态">
                <Badge
                  status={currentItem.approval_status === 3 && currentItem.status === 1 ? 'success' : 'default'}
                  text={currentItem.approval_status === 3 && currentItem.status === 1 ? '生效中' : '未生效'}
                />
              </Descriptions.Item>
              <Descriptions.Item label="来源">
                {currentItem.source === 'official' ? '官方数据' : currentItem.source === 'template' ? '模板创建' : '手动创建'}
              </Descriptions.Item>
            </Descriptions>

            <div>
              <h4 style={{ marginBottom: 12 }}>审批流程</h4>
              <Timeline
                items={approvalHistory.map(r => ({
                  color: r.approval_action === 'approve' ? 'green' : r.approval_action === 'reject' ? 'red' : 'blue',
                  children: (
                    <div>
                      <div>
                        <strong>{r.approver_name}</strong>
                        <Tag style={{ marginLeft: 8 }}>
                          {r.approval_action === 'submit' ? '提交' : r.approval_action === 'approve' ? '通过' : r.approval_action === 'reject' ? '拒绝' : '撤销'}
                        </Tag>
                        <span style={{ color: '#8c8c8c', marginLeft: 8, fontSize: 12 }}>
                          {dayjs(r.created_at).format('YYYY-MM-DD HH:mm')}
                        </span>
                      </div>
                      {r.approval_note && <div style={{ color: '#595959', marginTop: 4 }}>{r.approval_note}</div>}
                    </div>
                  ),
                }))}
              />
              {approvalHistory.length === 0 && <Empty description="暂无审批记录" />}
            </div>

            <div style={{ height: 280 }}>
              <AMap
                style={{ height: '100%', borderRadius: 8 }}
                center={[currentItem.center_longitude, currentItem.center_latitude]}
                zoom={14}
                polygons={
                  currentItem.boundary_polygon?.coordinates?.[0]
                    ? [{
                        path: currentItem.boundary_polygon.coordinates[0],
                        fillColor: AREA_TYPE_OPTIONS.find(o => o.value === currentItem.area_type)?.color || '#ff4d4f',
                        strokeColor: AREA_TYPE_OPTIONS.find(o => o.value === currentItem.area_type)?.color || '#ff4d4f',
                        fillOpacity: 0.25,
                      }]
                    : []
                }
              />
            </div>
          </Space>
        )}
      </Drawer>

      <Modal
        title={modalMode === 'create' ? '新建禁行区域' : '编辑禁行区域'}
        open={modalVisible}
        width={880}
        onCancel={() => { setModalVisible(false); clearDrawing() }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFormSubmit}
          initialValues={{ level: 2, shape_type: 'polygon', is_temporary: 0, submit_approval: true }}
        >
          <Row gutter={16}>
            <Col span={24}>
              <Card
                size="small"
                title={
                  <Space>
                    <span>地图绘制区域</span>
                    <Radio.Group
                      value={drawMode}
                      onChange={(e) => {
                        setDrawMode(e.target.value)
                        if (e.target.value !== 'polygon') setPolygonPoints([])
                        if (e.target.value !== 'circle') setCircleCenter(null)
                      }}
                    >
                      <Radio.Button value="none">浏览</Radio.Button>
                      <Radio.Button value="polygon">绘制多边形</Radio.Button>
                      <Radio.Button value="circle">绘制圆形</Radio.Button>
                    </Radio.Group>
                    <Button size="small" onClick={clearDrawing} disabled={drawMode === 'none'}>
                      清除绘制
                    </Button>
                  </Space>
                }
                style={{ marginBottom: 16 }}
              >
                <div style={{ height: 320 }}>
                  <AMap
                    style={{ height: '100%', borderRadius: 8 }}
                    polygons={displayPolygons.filter(p => p !== undefined) as any[]}
                    zoom={12}
                    onMapLoaded={handleMapLoaded}
                    onMapClick={drawMode !== 'none' ? handleMapClick : undefined}
                  />
                </div>
                {drawMode === 'polygon' && (
                  <div style={{ marginTop: 8, color: '#8c8c8c', fontSize: 12 }}>
                    💡 提示：在地图上点击添加多边形顶点（至少3个点），当前已绘制 {polygonPoints.length} 个点
                  </div>
                )}
                {drawMode === 'circle' && (
                  <Space style={{ marginTop: 8 }}>
                    <span style={{ color: '#8c8c8c', fontSize: 12 }}>
                      💡 提示：在地图上点击选择圆心
                      {circleCenter && `（已选: ${circleCenter[1].toFixed(6)}, ${circleCenter[0].toFixed(6)}）`}
                    </span>
                    <Form.Item label="半径(米)" style={{ marginBottom: 0 }}>
                      <Form.Item name="radius" noStyle>
                        <InputNumber
                          min={50}
                          max={20000}
                          step={50}
                          style={{ width: 140 }}
                          value={circleRadius}
                          onChange={(v) => setCircleRadius(v || 500)}
                        />
                      </Form.Item>
                    </Form.Item>
                  </Space>
                )}
              </Card>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="区域名称" name="name" rules={[{ required: true, message: '请输入区域名称' }]}>
                <Input placeholder="如：北京市第一中学" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item label="区域类型" name="area_type" rules={[{ required: true }]}>
                <Select>
                  {AREA_TYPE_OPTIONS.map(o => (
                    <Option key={o.value} value={o.value}>{o.label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item label="区域形状" name="shape_type">
                <Select>
                  <Option value="polygon">多边形</Option>
                  <Option value="circle">圆形</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item label="危险等级" name="level">
                <Select>
                  <Option value={1}>1级 - 建议绕行</Option>
                  <Option value={2}>2级 - 必须避开</Option>
                </Select>
              </Form.Item>
            </Col>

            <Col span={24}>
              <Form.Item label="地址" name="address">
                <Input placeholder="详细地址（可选）" />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="是否临时禁行" name="is_temporary">
                <Radio.Group>
                  <Radio value={0}>永久禁行</Radio>
                  <Radio value={1}>
                    <Space>
                      临时禁行
                      <ThunderboltOutlined style={{ color: '#faad14' }} />
                    </Space>
                  </Radio>
                </Radio.Group>
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item noStyle shouldUpdate={(prev, cur) => prev.is_temporary !== cur.is_temporary}>
                {({ getFieldValue }) =>
                  getFieldValue('is_temporary') === 1 ? (
                    <Form.Item label="临时原因" name="temp_reason">
                      <Select>
                        <Option value="accident">交通事故</Option>
                        <Option value="construction">道路施工</Option>
                        <Option value="emergency">突发事件</Option>
                        <Option value="other">其他</Option>
                      </Select>
                    </Form.Item>
                  ) : null
                }
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="限制危险品类别" name="restrict_hazard_classes" tooltip="留空表示限制全部类别">
                <Input placeholder="如：3,8 （多个用逗号分隔）" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item label="有效期" name="effective_period">
                <RangePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>

            <Col span={24}>
              <Form.Item label="生效时间段" name="time_schedule">
                <div style={{ padding: 12, background: '#fafafa', borderRadius: 8 }}>
                  <ScheduleEditor />
                </div>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="限高(米)" name="height_limit">
                <InputNumber min={0} max={10} step={0.1} style={{ width: '100%' }} placeholder="不限请留空" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item label="限重(吨)" name="weight_limit">
                <InputNumber min={0} max={200} step={1} style={{ width: '100%' }} placeholder="不限请留空" />
              </Form.Item>
            </Col>

            {modalMode === 'create' && (
              <Col span={24}>
                <Form.Item name="submit_approval" valuePropName="checked">
                  <Checkbox>创建后立即提交审批</Checkbox>
                </Form.Item>
              </Col>
            )}
          </Row>

          <Form.Item style={{ textAlign: 'right', marginTop: 16 }}>
            <Space>
              <Button onClick={() => { setModalVisible(false); clearDrawing() }}>取消</Button>
              <Button type="primary" htmlType="submit">
                {modalMode === 'create' ? '创建' : '保存'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

const ScheduleEditor: React.FC = () => {
  const [rules, setRules] = useState<TimeScheduleRule[]>([
    { weekdays: [1, 2, 3, 4, 5], start_time: '07:00', end_time: '09:00', description: '早高峰' },
    { weekdays: [1, 2, 3, 4, 5], start_time: '17:00', end_time: '19:00', description: '晚高峰' },
  ])

  const addRule = () => {
    setRules([...rules, { weekdays: [1, 2, 3, 4, 5], start_time: '00:00', end_time: '23:59' }])
  }

  const removeRule = (idx: number) => {
    setRules(rules.filter((_, i) => i !== idx))
  }

  const updateRule = (idx: number, field: keyof TimeScheduleRule, value: any) => {
    const newRules = [...rules]
    newRules[idx] = { ...newRules[idx], [field]: value }
    setRules(newRules)
  }

  return (
    <div>
      <Space direction="vertical" size={8} style={{ width: '100%' }}>
        {rules.map((rule, idx) => (
          <Row key={idx} gutter={8} align="middle">
            <Col flex="auto">
              <CheckboxGroup
                value={rule.weekdays}
                onChange={(v) => updateRule(idx, 'weekdays', v as number[])}
                options={WEEKDAYS}
              />
            </Col>
            <Col>
              <TimePicker.RangePicker
                format="HH:mm"
                minuteStep={15}
                value={rule.start_time && rule.end_time ? [dayjs(rule.start_time, 'HH:mm'), dayjs(rule.end_time, 'HH:mm')] : null}
                onChange={(vals) => {
                  if (vals && vals[0] && vals[1]) {
                    updateRule(idx, 'start_time', vals[0].format('HH:mm'))
                    updateRule(idx, 'end_time', vals[1].format('HH:mm'))
                  }
                }}
              />
            </Col>
            <Col>
              <Input
                placeholder="备注"
                value={rule.description}
                onChange={(e) => updateRule(idx, 'description', e.target.value)}
                style={{ width: 100 }}
              />
            </Col>
            <Col>
              <Button type="text" danger size="small" onClick={() => removeRule(idx)}>
                删除
              </Button>
            </Col>
          </Row>
        ))}
        <Button type="dashed" size="small" icon={<PlusOutlined />} onClick={addRule}>
          添加时间段
        </Button>
      </Space>
    </div>
  )
}

export default RestrictedAreas
