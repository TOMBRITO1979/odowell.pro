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
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { stockMovementsAPI, productsAPI } from '../../services/api';
import { actionColors } from '../../theme/designSystem';

const { TextArea } = Input;
const { Text } = Typography;

const StockMovements = () => {
  const [loading, setLoading] = useState(false);
  const [movements, setMovements] = useState([]);
  const [products, setProducts] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
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
      width: 100,
      render: (_, record) => {
        if (record.reason === 'sale') {
          return (
            <Tooltip title="Gerar Recibo">
              <Button
                type="text"
                icon={<PrinterOutlined />}
                onClick={() => handleDownloadSaleReceipt(record.id)}
              />
            </Tooltip>
          );
        }
        return null;
      },
    },
  ];

  return (
    <div>
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
    </div>
  );
};

export default StockMovements;
