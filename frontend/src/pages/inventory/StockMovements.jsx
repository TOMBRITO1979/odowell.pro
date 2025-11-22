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
  DatePicker,
  InputNumber,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  SwapOutlined,
  InboxOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { stockMovementsAPI, productsAPI } from '../../services/api';

const { TextArea } = Input;

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

  const typeOptions = [
    { value: 'entry', label: 'Entrada', color: 'success', icon: <ArrowUpOutlined /> },
    { value: 'exit', label: 'Saída', color: 'error', icon: <ArrowDownOutlined /> },
    { value: 'adjustment', label: 'Ajuste', color: 'warning', icon: <SwapOutlined /> },
  ];

  const reasonOptions = [
    { value: 'purchase', label: 'Compra' },
    { value: 'sale', label: 'Venda' },
    { value: 'usage', label: 'Uso em Procedimento' },
    { value: 'loss', label: 'Perda/Dano' },
    { value: 'adjustment', label: 'Ajuste de Inventário' },
    { value: 'return', label: 'Devolução' },
    { value: 'donation', label: 'Doação' },
    { value: 'other', label: 'Outro' },
  ];

  useEffect(() => {
    fetchMovements();
    fetchProducts();
  }, [pagination.current, filters]);

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
      message.error('Erro ao carregar movimentações');
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
    setModalVisible(true);
  };

  const handleCancel = () => {
    setModalVisible(false);
    form.resetFields();
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      await stockMovementsAPI.create(values);
      message.success('Movimentação registrada com sucesso!');
      setModalVisible(false);
      form.resetFields();
      fetchMovements();
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao registrar movimentação'
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
        const color = record.type === 'entry' ? '#52c41a' : '#ff4d4f';
        const prefix = record.type === 'entry' ? '+' : record.type === 'exit' ? '-' : '';
        return <span style={{ color, fontWeight: 'bold' }}>{prefix}{quantity}</span>;
      },
    },
    {
      title: 'Motivo',
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
      render: (reason) => getReasonLabel(reason),
    },
    {
      title: 'Observações',
      dataIndex: 'notes',
      key: 'notes',
      ellipsis: true,
      render: (notes) => notes || '-',
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <InboxOutlined />
            <span>Movimentações de Estoque</span>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<FileExcelOutlined />} onClick={handleExportCSV} style={{ backgroundColor: '#22c55e', borderColor: '#22c55e', color: '#fff' }}>
              Exportar CSV
            </Button>
            <Button icon={<FilePdfOutlined />} onClick={handleExportPDF} style={{ backgroundColor: '#ef4444', borderColor: '#ef4444', color: '#fff' }}>
              Gerar PDF
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={showModal}
            >
              Nova Movimentação
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
              placeholder="Tipo de Movimentação"
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
        title="Nova Movimentação de Estoque"
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
                label="Tipo de Movimentação"
                rules={[
                  { required: true, message: 'Selecione o tipo' },
                ]}
              >
                <Select placeholder="Selecione">
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
                <Select placeholder="Selecione o motivo">
                  {reasonOptions.map((reason) => (
                    <Select.Option key={reason.value} value={reason.value}>
                      {reason.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea
              rows={4}
              placeholder="Descreva detalhes da movimentação..."
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
