import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Card,
  message,
  Popconfirm,
  Row,
  Col,
  Badge,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  InboxOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { productsAPI } from '../../services/api';

const Products = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [products, setProducts] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // Filters
  const [filters, setFilters] = useState({
    search: '',
    category: undefined,
    active: undefined,
  });

  const categoryOptions = [
    { value: 'material', label: 'Material', color: 'blue' },
    { value: 'medicine', label: 'Medicamento', color: 'green' },
    { value: 'equipment', label: 'Equipamento', color: 'purple' },
    { value: 'consumable', label: 'Consumível', color: 'orange' },
  ];

  useEffect(() => {
    fetchProducts();
  }, [pagination.current, filters]);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await productsAPI.getAll(params);
      setProducts(response.data.products || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar produtos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await productsAPI.delete(id);
      message.success('Produto excluído com sucesso');
      fetchProducts();
    } catch (error) {
      message.error('Erro ao excluir produto');
    }
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getCategoryTag = (category) => {
    const catObj = categoryOptions.find((c) => c.value === category);
    return catObj ? (
      <Tag color={catObj.color}>{catObj.label}</Tag>
    ) : (
      <Tag>{category}</Tag>
    );
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  };

  const getStockStatus = (quantity, minimumStock) => {
    if (quantity === 0) {
      return <Tag color="error">Sem estoque</Tag>;
    } else if (quantity <= minimumStock) {
      return <Tag color="warning" icon={<WarningOutlined />}>Estoque baixo</Tag>;
    } else {
      return <Tag color="success">Em estoque</Tag>;
    }
  };

  const columns = [
    {
      title: 'Código',
      dataIndex: 'code',
      key: 'code',
      width: 120,
    },
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Categoria',
      dataIndex: 'category',
      key: 'category',
      width: 140,
      render: (category) => getCategoryTag(category),
    },
    {
      title: 'Quantidade',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 100,
      align: 'center',
      render: (quantity, record) => (
        <Badge
          count={quantity}
          showZero
          overflowCount={9999}
          style={{
            backgroundColor: quantity <= record.minimum_stock ? '#faad14' : '#52c41a',
          }}
        />
      ),
    },
    {
      title: 'Estoque Min.',
      dataIndex: 'minimum_stock',
      key: 'minimum_stock',
      width: 100,
      align: 'center',
    },
    {
      title: 'Status',
      key: 'stock_status',
      width: 140,
      render: (_, record) => getStockStatus(record.quantity, record.minimum_stock),
    },
    {
      title: 'Preço Custo',
      dataIndex: 'cost_price',
      key: 'cost_price',
      width: 120,
      render: (price) => formatCurrency(price),
    },
    {
      title: 'Preço Venda',
      dataIndex: 'sale_price',
      key: 'sale_price',
      width: 120,
      render: (price) => formatCurrency(price),
    },
    {
      title: 'Ativo',
      dataIndex: 'active',
      key: 'active',
      width: 80,
      render: (active) => (
        <Tag color={active ? 'success' : 'default'}>
          {active ? 'Sim' : 'Não'}
        </Tag>
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/products/${record.id}/edit`)}
            title="Editar"
          />
          <Popconfirm
            title="Tem certeza que deseja excluir?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title="Excluir"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <InboxOutlined />
            <span>Produtos e Estoque</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/products/new')}
          >
            Novo Produto
          </Button>
        }
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
            <Input
              placeholder="Buscar por nome ou código..."
              prefix={<SearchOutlined />}
              allowClear
              value={filters.search}
              onChange={(e) =>
                setFilters({ ...filters, search: e.target.value })
              }
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Categoria"
              style={{ width: '100%' }}
              allowClear
              value={filters.category}
              onChange={(value) => setFilters({ ...filters, category: value })}
            >
              {categoryOptions.map((cat) => (
                <Select.Option key={cat.value} value={cat.value}>
                  {cat.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Status"
              style={{ width: '100%' }}
              allowClear
              value={filters.active}
              onChange={(value) => setFilters({ ...filters, active: value })}
            >
              <Select.Option value={true}>Ativo</Select.Option>
              <Select.Option value={false}>Inativo</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Button onClick={fetchProducts} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={products}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default Products;
