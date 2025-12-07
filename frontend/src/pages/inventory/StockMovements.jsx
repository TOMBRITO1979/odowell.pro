import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Card,
  message,
  Modal,
  Form,
  Input,
  Select,
  Row,
  Col,
  InputNumber,
  Divider,
  Typography,
  Tooltip,
  Popconfirm,
  Descriptions,
  DatePicker,
  Statistic,
  Spin,
} from 'antd';
import {
  PlusOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  SwapOutlined,
  InboxOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
  PrinterOutlined,
  DollarOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  BarChartOutlined,
  ShoppingOutlined,
  ExportOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer, Legend } from 'recharts';
import dayjs from 'dayjs';
import { stockMovementsAPI, productsAPI } from '../../services/api';
import { actionColors } from '../../theme/designSystem';
import { usePermission } from '../../contexts/AuthContext';

const { TextArea } = Input;
const { Text } = Typography;
const { RangePicker } = DatePicker;

const StockMovements = () => {
  const [loading, setLoading] = useState(false);
  const [movements, setMovements] = useState([]);
  const [products, setProducts] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [viewModalVisible, setViewModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [selectedMovement, setSelectedMovement] = useState(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [filters, setFilters] = useState({
    product_id: undefined,
    type: undefined,
  });

  // State for sale-specific display
  const [isSale, setIsSale] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [saleTotal, setSaleTotal] = useState(0);

  // Dashboard statistics state
  const [stats, setStats] = useState({
    exits_by_reason: [],
    exits_by_product: [],
    exits_by_product_date: [],
    total_sales_revenue: 0,
    total_sales_count: 0,
    total_exits: 0,
    total_entries: 0,
  });
  const [statsLoading, setStatsLoading] = useState(false);
  const [statsDateRange, setStatsDateRange] = useState([
    dayjs().startOf('month'),
    dayjs().endOf('month'),
  ]);
  const [statsProductId, setStatsProductId] = useState(undefined);

  // Permissions
  const { canEdit, canDelete } = usePermission();

  const typeOptions = [
    { value: 'entry', label: 'Entrada', color: 'success', icon: <ArrowUpOutlined /> },
    { value: 'exit', label: 'Saida', color: 'error', icon: <ArrowDownOutlined /> },
    { value: 'adjustment', label: 'Ajuste', color: 'warning', icon: <SwapOutlined /> },
  ];

  const reasonOptions = [
    { value: 'purchase', label: 'Compra' },
    { value: 'sale', label: 'Venda' },
    { value: 'usage', label: 'Uso em Procedimento' },
    { value: 'loss', label: 'Perda/Dano' },
    { value: 'adjustment', label: 'Ajuste de Inventario' },
    { value: 'return', label: 'Devolucao' },
    { value: 'donation', label: 'Doacao' },
    { value: 'other', label: 'Outro' },
  ];

  useEffect(() => {
    fetchMovements();
    fetchProducts();
  }, [pagination.current, pagination.pageSize, filters]);

  const fetchMovements = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await stockMovementsAPI.getAll(params);
      setMovements(response.data.movements || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar movimentacoes');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchProducts = async () => {
    try {
      const response = await productsAPI.getAll({ page: 1, page_size: 1000 });
      setProducts(response.data.products || []);
    } catch (error) {
      console.error('Error fetching products:', error);
    }
  };

  // Fetch stats when date range or product filter changes
  useEffect(() => {
    fetchStats();
  }, [statsDateRange, statsProductId]);

  const fetchStats = async () => {
    setStatsLoading(true);
    try {
      const params = {};
      if (statsDateRange && statsDateRange[0] && statsDateRange[1]) {
        params.start_date = statsDateRange[0].format('YYYY-MM-DD');
        params.end_date = statsDateRange[1].format('YYYY-MM-DD');
      }
      if (statsProductId) {
        params.product_id = statsProductId;
      }

      const response = await stockMovementsAPI.getStats(params);
      setStats(response.data);
    } catch (error) {
      console.error('Error fetching stats:', error);
      message.error('Erro ao carregar estatisticas');
    } finally {
      setStatsLoading(false);
    }
  };

  // Reason labels for chart
  const reasonLabels = {
    sale: 'Venda',
    usage: 'Uso em Procedimento',
    loss: 'Perda/Dano',
    return: 'Devolucao',
    donation: 'Doacao',
    other: 'Outro',
  };

  // Colors for chart bars
  const chartColors = {
    sale: '#52c41a',
    usage: '#1890ff',
    loss: '#ff4d4f',
    return: '#faad14',
    donation: '#722ed1',
    other: '#8c8c8c',
  };

  // Prepare chart data for exits by reason
  const chartData = (stats.exits_by_reason || []).map(item => ({
    reason: reasonLabels[item.reason] || item.reason,
    quantidade: item.total_quantity || 0,
    originalReason: item.reason,
  }));

  // Colors for product lines (10 distinct colors)
  const productLineColors = [
    '#1890ff', // blue
    '#52c41a', // green
    '#ff4d4f', // red
    '#faad14', // gold
    '#722ed1', // purple
    '#13c2c2', // cyan
    '#eb2f96', // magenta
    '#fa8c16', // orange
    '#2f54eb', // geekblue
    '#a0d911', // lime
  ];

  // Get unique product names from exits_by_product
  const productNames = (stats.exits_by_product || []).map(p => p.product_name);

  // Transform exits_by_product_date to multi-line chart format
  // From: [{ product_name, date, total_quantity }, ...]
  // To: [{ date, "Product A": 10, "Product B": 20 }, ...]
  const productChartData = React.useMemo(() => {
    const dataByDate = {};
    (stats.exits_by_product_date || []).forEach(item => {
      const formattedDate = dayjs(item.date).format('DD/MM');
      if (!dataByDate[formattedDate]) {
        dataByDate[formattedDate] = { date: formattedDate };
      }
      dataByDate[formattedDate][item.product_name] = (dataByDate[formattedDate][item.product_name] || 0) + item.total_quantity;
    });
    // Convert to array and sort by date
    return Object.values(dataByDate).sort((a, b) => {
      const dateA = a.date.split('/').reverse().join('');
      const dateB = b.date.split('/').reverse().join('');
      return dateA.localeCompare(dateB);
    });
  }, [stats.exits_by_product_date]);

  const showModal = () => {
    form.resetFields();
    setIsSale(false);
    setSelectedProduct(null);
    setSaleTotal(0);
    setModalVisible(true);
  };

  const handleCancel = () => {
    setModalVisible(false);
    form.resetFields();
    setIsSale(false);
    setSelectedProduct(null);
    setSaleTotal(0);
  };

  const handleReasonChange = (value) => {
    setIsSale(value === 'sale');
    // Auto-set type to 'exit' when sale is selected
    if (value === 'sale') {
      form.setFieldsValue({ type: 'exit' });
    }
    calculateSaleTotal();
  };

  const handleProductChange = (productId) => {
    const product = products.find(p => p.id === productId);
    setSelectedProduct(product);
    calculateSaleTotal();
  };

  const handleQuantityChange = () => {
    calculateSaleTotal();
  };

  const calculateSaleTotal = () => {
    const productId = form.getFieldValue('product_id');
    const quantity = form.getFieldValue('quantity') || 0;
    const product = products.find(p => p.id === productId);

    if (product && quantity > 0) {
      setSaleTotal(product.sale_price * quantity);
    } else {
      setSaleTotal(0);
    }
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      await stockMovementsAPI.create(values);
      message.success('Movimentacao registrada com sucesso!');
      setModalVisible(false);
      form.resetFields();
      setIsSale(false);
      setSelectedProduct(null);
      setSaleTotal(0);
      fetchMovements();
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao registrar movimentacao'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleExportCSV = async () => {
    try {
      const response = await stockMovementsAPI.exportCSV('');
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `movimentacoes_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
      console.error('Export error:', error);
    }
  };

  const handleExportPDF = async () => {
    try {
      const response = await stockMovementsAPI.exportPDF('');
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `movimentacoes_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
      console.error('PDF error:', error);
    }
  };

  const handleDownloadSaleReceipt = async (movementId) => {
    try {
      const response = await stockMovementsAPI.downloadSaleReceipt(movementId);
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `recibo_venda_${movementId}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('Recibo gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar recibo');
      console.error('Receipt error:', error);
    }
  };

  // View movement details
  const handleView = (record) => {
    setSelectedMovement(record);
    setViewModalVisible(true);
  };

  // Edit movement
  const handleEdit = (record) => {
    setSelectedMovement(record);
    editForm.setFieldsValue({
      notes: record.notes,
      buyer_name: record.buyer_name,
      buyer_document: record.buyer_document,
      buyer_phone: record.buyer_phone,
      buyer_street: record.buyer_street,
      buyer_number: record.buyer_number,
      buyer_neighborhood: record.buyer_neighborhood,
      buyer_city: record.buyer_city,
      buyer_state: record.buyer_state,
      buyer_zip_code: record.buyer_zip_code,
    });
    setEditModalVisible(true);
  };

  const handleEditSubmit = async (values) => {
    setLoading(true);
    try {
      await stockMovementsAPI.update(selectedMovement.id, values);
      message.success('Movimentacao atualizada com sucesso!');
      setEditModalVisible(false);
      editForm.resetFields();
      setSelectedMovement(null);
      fetchMovements();
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao atualizar movimentacao'
      );
    } finally {
      setLoading(false);
    }
  };

  // Delete movement
  const handleDelete = async (id) => {
    setLoading(true);
    try {
      await stockMovementsAPI.delete(id);
      message.success('Movimentacao excluida com sucesso!');
      fetchMovements();
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao excluir movimentacao'
      );
    } finally {
      setLoading(false);
    }
  };

  const getTypeTag = (type) => {
    const typeObj = typeOptions.find((t) => t.value === type);
    if (!typeObj) return <Tag>{type}</Tag>;

    return (
      <Tag color={typeObj.color} icon={typeObj.icon}>
        {typeObj.label}
      </Tag>
    );
  };

  const getReasonLabel = (reason) => {
    const reasonObj = reasonOptions.find((r) => r.value === reason);
    return reasonObj ? reasonObj.label : reason;
  };

  const getProductName = (productId) => {
    const product = products.find((p) => p.id === productId);
    return product ? product.name : `Produto #${productId}`;
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value || 0);
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
      sorter: true,
    },
    {
      title: 'Produto',
      dataIndex: 'product_id',
      key: 'product_id',
      ellipsis: true,
      render: (productId) => getProductName(productId),
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: 130,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'Quantidade',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 110,
      align: 'center',
      render: (quantity, record) => {
        const color = record.type === 'entry' ? '#81C784' : '#E57373';
        const prefix = record.type === 'entry' ? '+' : record.type === 'exit' ? '-' : '';
        return <span style={{ color, fontWeight: 'bold' }}>{prefix}{quantity}</span>;
      },
    },
    {
      title: 'Motivo',
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
      render: (reason) => (
        <Space>
          {getReasonLabel(reason)}
          {reason === 'sale' && <DollarOutlined style={{ color: '#52c41a' }} />}
        </Space>
      ),
    },
    {
      title: 'Valor',
      key: 'total_price',
      width: 120,
      render: (_, record) => {
        if (record.reason === 'sale' && record.total_price > 0) {
          return <Text strong style={{ color: '#52c41a' }}>{formatCurrency(record.total_price)}</Text>;
        }
        return '-';
      },
    },
    {
      title: 'Acoes',
      key: 'actions',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="Visualizar">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleView(record)}
              style={{ color: actionColors.view }}
            />
          </Tooltip>
          {canEdit('stock_movements') && (
            <Tooltip title="Editar">
              <Button
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleEdit(record)}
                style={{ color: actionColors.edit }}
              />
            </Tooltip>
          )}
          {record.reason === 'sale' && (
            <Tooltip title="Gerar Recibo">
              <Button
                type="text"
                icon={<PrinterOutlined />}
                onClick={() => handleDownloadSaleReceipt(record.id)}
                style={{ color: actionColors.print }}
              />
            </Tooltip>
          )}
          {canDelete('stock_movements') && record.type !== 'adjustment' && (
            <Popconfirm
              title="Excluir movimentacao?"
              description="O estoque sera revertido automaticamente."
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Nao"
              okButtonProps={{ danger: true }}
            >
              <Tooltip title="Excluir">
                <Button
                  type="text"
                  icon={<DeleteOutlined />}
                  style={{ color: actionColors.delete }}
                />
              </Tooltip>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* Dashboard Statistics */}
      <Card
        title={
          <Space>
            <BarChartOutlined />
            <span>Estatisticas de Saidas</span>
          </Space>
        }
        style={{ marginBottom: 16 }}
        extra={
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchStats}
            loading={statsLoading}
          >
            Atualizar
          </Button>
        }
      >
        {/* Stats Filters */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={8}>
            <Text strong style={{ display: 'block', marginBottom: 8 }}>Periodo:</Text>
            <RangePicker
              value={statsDateRange}
              onChange={(dates) => setStatsDateRange(dates)}
              format="DD/MM/YYYY"
              style={{ width: '100%' }}
              allowClear={false}
              placeholder={['Data inicial', 'Data final']}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Text strong style={{ display: 'block', marginBottom: 8 }}>Produto:</Text>
            <Select
              placeholder="Todos os produtos"
              style={{ width: '100%' }}
              allowClear
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().includes(input.toLowerCase())
              }
              value={statsProductId}
              onChange={(value) => setStatsProductId(value)}
            >
              {products.map((product) => (
                <Select.Option key={product.id} value={product.id}>
                  {product.name}
                </Select.Option>
              ))}
            </Select>
          </Col>
        </Row>

        <Spin spinning={statsLoading}>
          <Row gutter={[16, 16]}>
            {/* Revenue Card */}
            <Col xs={24} sm={12} md={6}>
              <Card style={{ backgroundColor: '#f6ffed', borderColor: '#b7eb8f' }}>
                <Statistic
                  title={<span style={{ color: '#389e0d' }}>Total Faturado (Vendas)</span>}
                  value={stats.total_sales_revenue || 0}
                  precision={2}
                  prefix="R$"
                  valueStyle={{ color: '#52c41a', fontWeight: 'bold' }}
                />
              </Card>
            </Col>
            {/* Sales Count Card */}
            <Col xs={24} sm={12} md={6}>
              <Card style={{ backgroundColor: '#e6f7ff', borderColor: '#91d5ff' }}>
                <Statistic
                  title={<span style={{ color: '#096dd9' }}>Vendas Realizadas</span>}
                  value={stats.total_sales_count || 0}
                  prefix={<ShoppingOutlined />}
                  valueStyle={{ color: '#1890ff', fontWeight: 'bold' }}
                />
              </Card>
            </Col>
            {/* Total Exits Card */}
            <Col xs={24} sm={12} md={6}>
              <Card style={{ backgroundColor: '#fff2e8', borderColor: '#ffbb96' }}>
                <Statistic
                  title={<span style={{ color: '#d4380d' }}>Total de Saidas</span>}
                  value={stats.total_exits || 0}
                  prefix={<ExportOutlined />}
                  valueStyle={{ color: '#fa541c', fontWeight: 'bold' }}
                />
              </Card>
            </Col>
            {/* Total Entries Card */}
            <Col xs={24} sm={12} md={6}>
              <Card style={{ backgroundColor: '#f0f5ff', borderColor: '#adc6ff' }}>
                <Statistic
                  title={<span style={{ color: '#1d39c4' }}>Total de Entradas</span>}
                  value={stats.total_entries || 0}
                  prefix={<ArrowUpOutlined />}
                  valueStyle={{ color: '#2f54eb', fontWeight: 'bold' }}
                />
              </Card>
            </Col>
          </Row>

          {/* Chart */}
          {chartData.length > 0 && (
            <Row style={{ marginTop: 24 }}>
              <Col span={24}>
                <Card title="Quantidade de Saidas por Motivo" size="small">
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="reason" />
                      <YAxis />
                      <RechartsTooltip
                        formatter={(value) => [value, 'Quantidade']}
                        labelFormatter={(label) => `Motivo: ${label}`}
                      />
                      <Line
                        type="monotone"
                        dataKey="quantidade"
                        name="Quantidade"
                        stroke="#52c41a"
                        strokeWidth={2}
                        dot={{ r: 5, fill: '#52c41a', strokeWidth: 2 }}
                        activeDot={{ r: 7, fill: '#389e0d' }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </Card>
              </Col>
            </Row>
          )}

          {/* Multi-line chart: Saidas por Produto ao longo do tempo */}
          {productChartData.length > 0 && productNames.length > 0 && (
            <Row style={{ marginTop: 24 }}>
              <Col span={24}>
                <Card title="Saidas por Produto (por Data)" size="small">
                  <ResponsiveContainer width="100%" height={350}>
                    <LineChart data={productChartData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis />
                      <RechartsTooltip />
                      <Legend />
                      {productNames.map((name, idx) => (
                        <Line
                          key={name}
                          type="monotone"
                          dataKey={name}
                          name={name}
                          stroke={productLineColors[idx % productLineColors.length]}
                          strokeWidth={2}
                          dot={{ r: 4, fill: productLineColors[idx % productLineColors.length] }}
                          activeDot={{ r: 6 }}
                          connectNulls
                        />
                      ))}
                    </LineChart>
                  </ResponsiveContainer>
                </Card>
              </Col>
            </Row>
          )}

          {chartData.length === 0 && !statsLoading && (
            <Row style={{ marginTop: 24 }}>
              <Col span={24}>
                <Card>
                  <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
                    <InboxOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                    <p>Nenhuma saida registrada no periodo selecionado</p>
                  </div>
                </Card>
              </Col>
            </Row>
          )}
        </Spin>
      </Card>

      {/* Main Movements Table */}
      <Card
        title={
          <Space>
            <InboxOutlined />
            <span>Movimentacoes de Estoque</span>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<FileExcelOutlined />} onClick={handleExportCSV} style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}>
              Exportar CSV
            </Button>
            <Button icon={<FilePdfOutlined />} onClick={handleExportPDF} style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}>
              Gerar PDF
            </Button>
            <Button
              icon={<PlusOutlined />}
              onClick={showModal}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              Nova Movimentacao
            </Button>
          </Space>
        }
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Selecione o produto"
              style={{ width: '100%' }}
              allowClear
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().includes(input.toLowerCase())
              }
              value={filters.product_id}
              onChange={(value) =>
                setFilters({ ...filters, product_id: value })
              }
            >
              {products.map((product) => (
                <Select.Option key={product.id} value={product.id}>
                  {product.name}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Tipo de Movimentacao"
              style={{ width: '100%' }}
              allowClear
              value={filters.type}
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              {typeOptions.map((type) => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Button onClick={fetchMovements} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={movements}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={(newPagination) => setPagination(newPagination)}
          scroll={{ x: 1000 }}
        />
      </Card>

      <Modal
        title="Nova Movimentacao de Estoque"
        open={modalVisible}
        onCancel={handleCancel}
        footer={null}
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="product_id"
                label="Produto"
                rules={[
                  { required: true, message: 'Selecione o produto' },
                ]}
              >
                <Select
                  placeholder="Selecione o produto"
                  showSearch
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                  onChange={handleProductChange}
                >
                  {products.map((product) => (
                    <Select.Option key={product.id} value={product.id}>
                      {product.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="type"
                label="Tipo de Movimentacao"
                rules={[
                  { required: true, message: 'Selecione o tipo' },
                ]}
              >
                <Select placeholder="Selecione" disabled={isSale}>
                  {typeOptions.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      <Space>
                        {type.icon}
                        {type.label}
                      </Space>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="quantity"
                label="Quantidade"
                rules={[
                  { required: true, message: 'Informe a quantidade' },
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  precision={0}
                  placeholder="Quantidade"
                  onChange={handleQuantityChange}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="reason"
                label="Motivo"
                rules={[
                  { required: true, message: 'Selecione o motivo' },
                ]}
              >
                <Select placeholder="Selecione o motivo" onChange={handleReasonChange}>
                  {reasonOptions.map((reason) => (
                    <Select.Option key={reason.value} value={reason.value}>
                      {reason.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {/* Sale-specific fields */}
          {isSale && (
            <>
              <Divider>Dados da Venda</Divider>

              {/* Price display */}
              {selectedProduct && (
                <Row gutter={16} style={{ marginBottom: 16 }}>
                  <Col xs={12}>
                    <Card size="small" style={{ backgroundColor: '#f6ffed' }}>
                      <Text type="secondary">Preco Unitario:</Text>
                      <br />
                      <Text strong style={{ fontSize: 18, color: '#52c41a' }}>
                        {formatCurrency(selectedProduct.sale_price)}
                      </Text>
                    </Card>
                  </Col>
                  <Col xs={12}>
                    <Card size="small" style={{ backgroundColor: '#e6f7ff' }}>
                      <Text type="secondary">Total da Venda:</Text>
                      <br />
                      <Text strong style={{ fontSize: 18, color: '#1890ff' }}>
                        {formatCurrency(saleTotal)}
                      </Text>
                    </Card>
                  </Col>
                </Row>
              )}

              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_name"
                    label="Nome do Comprador"
                  >
                    <Input placeholder="Nome completo do comprador" />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_document"
                    label="CPF/CNPJ"
                  >
                    <Input placeholder="000.000.000-00" />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_phone"
                    label="Telefone"
                  >
                    <Input placeholder="(00) 00000-0000" />
                  </Form.Item>
                </Col>
              </Row>

              <Divider orientation="left" plain>Endereco do Comprador</Divider>

              <Row gutter={16}>
                <Col xs={24} md={16}>
                  <Form.Item
                    name="buyer_street"
                    label="Rua"
                  >
                    <Input placeholder="Nome da rua" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={8}>
                  <Form.Item
                    name="buyer_number"
                    label="Numero"
                  >
                    <Input placeholder="123" />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_neighborhood"
                    label="Bairro"
                  >
                    <Input placeholder="Bairro" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_city"
                    label="Cidade"
                  >
                    <Input placeholder="Cidade" />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_state"
                    label="Estado"
                  >
                    <Select placeholder="Selecione o estado">
                      <Select.Option value="AC">Acre</Select.Option>
                      <Select.Option value="AL">Alagoas</Select.Option>
                      <Select.Option value="AP">Amapa</Select.Option>
                      <Select.Option value="AM">Amazonas</Select.Option>
                      <Select.Option value="BA">Bahia</Select.Option>
                      <Select.Option value="CE">Ceara</Select.Option>
                      <Select.Option value="DF">Distrito Federal</Select.Option>
                      <Select.Option value="ES">Espirito Santo</Select.Option>
                      <Select.Option value="GO">Goias</Select.Option>
                      <Select.Option value="MA">Maranhao</Select.Option>
                      <Select.Option value="MT">Mato Grosso</Select.Option>
                      <Select.Option value="MS">Mato Grosso do Sul</Select.Option>
                      <Select.Option value="MG">Minas Gerais</Select.Option>
                      <Select.Option value="PA">Para</Select.Option>
                      <Select.Option value="PB">Paraiba</Select.Option>
                      <Select.Option value="PR">Parana</Select.Option>
                      <Select.Option value="PE">Pernambuco</Select.Option>
                      <Select.Option value="PI">Piaui</Select.Option>
                      <Select.Option value="RJ">Rio de Janeiro</Select.Option>
                      <Select.Option value="RN">Rio Grande do Norte</Select.Option>
                      <Select.Option value="RS">Rio Grande do Sul</Select.Option>
                      <Select.Option value="RO">Rondonia</Select.Option>
                      <Select.Option value="RR">Roraima</Select.Option>
                      <Select.Option value="SC">Santa Catarina</Select.Option>
                      <Select.Option value="SP">Sao Paulo</Select.Option>
                      <Select.Option value="SE">Sergipe</Select.Option>
                      <Select.Option value="TO">Tocantins</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_zip_code"
                    label="CEP"
                  >
                    <Input placeholder="00000-000" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Form.Item
            name="notes"
            label="Observacoes"
          >
            <TextArea
              rows={4}
              placeholder="Descreva detalhes da movimentacao..."
              maxLength={500}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
              >
                Registrar
              </Button>
              <Button onClick={handleCancel}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* View Modal */}
      <Modal
        title="Detalhes da Movimentacao"
        open={viewModalVisible}
        onCancel={() => {
          setViewModalVisible(false);
          setSelectedMovement(null);
        }}
        footer={[
          <Button key="close" onClick={() => {
            setViewModalVisible(false);
            setSelectedMovement(null);
          }}>
            Fechar
          </Button>,
          selectedMovement?.reason === 'sale' && (
            <Button
              key="receipt"
              type="primary"
              icon={<PrinterOutlined />}
              onClick={() => handleDownloadSaleReceipt(selectedMovement.id)}
            >
              Gerar Recibo
            </Button>
          ),
        ]}
        width={700}
      >
        {selectedMovement && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="ID">{selectedMovement.id}</Descriptions.Item>
            <Descriptions.Item label="Data">
              {dayjs(selectedMovement.created_at).format('DD/MM/YYYY HH:mm')}
            </Descriptions.Item>
            <Descriptions.Item label="Produto">
              {selectedMovement.product?.name || getProductName(selectedMovement.product_id)}
            </Descriptions.Item>
            <Descriptions.Item label="Tipo">
              {getTypeTag(selectedMovement.type)}
            </Descriptions.Item>
            <Descriptions.Item label="Quantidade">
              <span style={{
                color: selectedMovement.type === 'entry' ? '#81C784' : '#E57373',
                fontWeight: 'bold'
              }}>
                {selectedMovement.type === 'entry' ? '+' : selectedMovement.type === 'exit' ? '-' : ''}
                {selectedMovement.quantity}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Motivo">
              {getReasonLabel(selectedMovement.reason)}
            </Descriptions.Item>
            {selectedMovement.reason === 'sale' && (
              <>
                <Descriptions.Item label="Preco Unitario">
                  {formatCurrency(selectedMovement.unit_price)}
                </Descriptions.Item>
                <Descriptions.Item label="Total">
                  <Text strong style={{ color: '#52c41a' }}>
                    {formatCurrency(selectedMovement.total_price)}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item label="Comprador" span={2}>
                  {selectedMovement.buyer_name || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="CPF/CNPJ">
                  {selectedMovement.buyer_document || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Telefone">
                  {selectedMovement.buyer_phone || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Endereco" span={2}>
                  {selectedMovement.buyer_street
                    ? `${selectedMovement.buyer_street}${selectedMovement.buyer_number ? ', ' + selectedMovement.buyer_number : ''}`
                    : 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Bairro">
                  {selectedMovement.buyer_neighborhood || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Cidade/UF">
                  {selectedMovement.buyer_city || selectedMovement.buyer_state
                    ? `${selectedMovement.buyer_city || ''}${selectedMovement.buyer_city && selectedMovement.buyer_state ? ' - ' : ''}${selectedMovement.buyer_state || ''}`
                    : 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="CEP">
                  {selectedMovement.buyer_zip_code || 'Nao informado'}
                </Descriptions.Item>
              </>
            )}
            <Descriptions.Item label="Usuario">
              {selectedMovement.user?.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Observacoes" span={2}>
              {selectedMovement.notes || 'Nenhuma observacao'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Edit Modal */}
      <Modal
        title="Editar Movimentacao"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setSelectedMovement(null);
          editForm.resetFields();
        }}
        footer={null}
        width={700}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleEditSubmit}
        >
          <Form.Item
            name="notes"
            label="Observacoes"
          >
            <TextArea
              rows={4}
              placeholder="Descreva detalhes da movimentacao..."
              maxLength={500}
            />
          </Form.Item>

          {selectedMovement?.reason === 'sale' && (
            <>
              <Divider>Dados do Comprador</Divider>
              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_name"
                    label="Nome do Comprador"
                  >
                    <Input placeholder="Nome completo" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_document"
                    label="CPF/CNPJ"
                  >
                    <Input placeholder="000.000.000-00" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_phone"
                    label="Telefone"
                  >
                    <Input placeholder="(00) 00000-0000" />
                  </Form.Item>
                </Col>
              </Row>

              <Divider orientation="left" plain>Endereco</Divider>
              <Row gutter={16}>
                <Col xs={24} md={16}>
                  <Form.Item
                    name="buyer_street"
                    label="Rua"
                  >
                    <Input placeholder="Nome da rua" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={8}>
                  <Form.Item
                    name="buyer_number"
                    label="Numero"
                  >
                    <Input placeholder="123" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_neighborhood"
                    label="Bairro"
                  >
                    <Input placeholder="Bairro" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_city"
                    label="Cidade"
                  >
                    <Input placeholder="Cidade" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_state"
                    label="Estado"
                  >
                    <Select placeholder="Selecione" allowClear>
                      <Select.Option value="AC">AC</Select.Option>
                      <Select.Option value="AL">AL</Select.Option>
                      <Select.Option value="AP">AP</Select.Option>
                      <Select.Option value="AM">AM</Select.Option>
                      <Select.Option value="BA">BA</Select.Option>
                      <Select.Option value="CE">CE</Select.Option>
                      <Select.Option value="DF">DF</Select.Option>
                      <Select.Option value="ES">ES</Select.Option>
                      <Select.Option value="GO">GO</Select.Option>
                      <Select.Option value="MA">MA</Select.Option>
                      <Select.Option value="MT">MT</Select.Option>
                      <Select.Option value="MS">MS</Select.Option>
                      <Select.Option value="MG">MG</Select.Option>
                      <Select.Option value="PA">PA</Select.Option>
                      <Select.Option value="PB">PB</Select.Option>
                      <Select.Option value="PR">PR</Select.Option>
                      <Select.Option value="PE">PE</Select.Option>
                      <Select.Option value="PI">PI</Select.Option>
                      <Select.Option value="RJ">RJ</Select.Option>
                      <Select.Option value="RN">RN</Select.Option>
                      <Select.Option value="RS">RS</Select.Option>
                      <Select.Option value="RO">RO</Select.Option>
                      <Select.Option value="RR">RR</Select.Option>
                      <Select.Option value="SC">SC</Select.Option>
                      <Select.Option value="SP">SP</Select.Option>
                      <Select.Option value="SE">SE</Select.Option>
                      <Select.Option value="TO">TO</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item
                    name="buyer_zip_code"
                    label="CEP"
                  >
                    <Input placeholder="00000-000" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
              >
                Salvar
              </Button>
              <Button onClick={() => {
                setEditModalVisible(false);
                setSelectedMovement(null);
                editForm.resetFields();
              }}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default StockMovements;
