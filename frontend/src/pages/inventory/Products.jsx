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
  Upload,
  Modal,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  InboxOutlined,
  WarningOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { productsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const Products = () => {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [products, setProducts] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploading, setUploading] = useState(false);

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
  }, [pagination.current, pagination.pageSize, filters]);

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

  const handleExportCSV = async () => {
    try {
      const response = await productsAPI.exportCSV('');
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `produtos_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
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
      const response = await productsAPI.exportPDF('');
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `produtos_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
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

  const handleImportCSV = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const response = await productsAPI.importCSV(formData);
      message.success(response.data.message);

      if (response.data.errors && response.data.errors.length > 0) {
        Modal.warning({
          title: 'Avisos durante a importação',
          content: (
            <div>
              <p>{response.data.imported} produtos importados com sucesso.</p>
              <p>Erros encontrados:</p>
              <ul>
                {response.data.errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            </div>
          ),
          width: 600,
        });
      }

      setUploadModalVisible(false);
      fetchProducts();
    } catch (error) {
      message.error('Erro ao importar CSV');
      console.error('Import error:', error);
    } finally {
      setUploading(false);
    }

    return false;
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
      title: 'Qtd',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 80,
      align: 'center',
      render: (quantity, record) => (
        <Badge
          count={quantity}
          showZero
          overflowCount={9999}
          style={{
            backgroundColor: quantity <= record.minimum_stock ? '#FFD54F' : '#81C784',
          }}
        />
      ),
    },
    {
      title: 'Est. Min.',
      dataIndex: 'minimum_stock',
      key: 'minimum_stock',
      width: 80,
      align: 'center',
    },
    {
      title: 'Status',
      key: 'stock_status',
      width: 120,
      render: (_, record) => getStockStatus(record.quantity, record.minimum_stock),
    },
    {
      title: 'P. Custo',
      dataIndex: 'cost_price',
      key: 'cost_price',
      width: 100,
      render: (price) => formatCurrency(price),
    },
    {
      title: 'P. Venda',
      dataIndex: 'sale_price',
      key: 'sale_price',
      width: 100,
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
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Space>
          {canEdit('products') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/products/${record.id}/edit`)}
              title="Editar"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('products') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                icon={<DeleteOutlined />}
                title="Excluir"
                style={{ color: actionColors.delete }}
              />
            </Popconfirm>
          )}
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
          <Space>
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              style={{
                backgroundColor: actionColors.exportExcel,
                borderColor: actionColors.exportExcel,
                color: '#fff'
              }}
            >
              Exportar CSV
            </Button>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleExportPDF}
              style={{
                backgroundColor: actionColors.exportPDF,
                borderColor: actionColors.exportPDF,
                color: '#fff'
              }}
            >
              Gerar PDF
            </Button>
            {canCreate('products') && (
              <Button
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
                style={{
                  backgroundColor: actionColors.view,
                  borderColor: actionColors.view,
                  color: '#fff'
                }}
              >
                Importar CSV
              </Button>
            )}
            {canCreate('products') && (
              <Button
                icon={<PlusOutlined />}
                onClick={() => navigate('/products/new')}
                style={{
                  backgroundColor: actionColors.create,
                  borderColor: actionColors.create,
                  color: '#fff'
                }}
              >
                Novo Produto
              </Button>
            )}
          </Space>
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

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={products}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={handleTableChange}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>

      <Modal
        title="Importar Produtos via CSV"
        open={uploadModalVisible}
        onCancel={() => setUploadModalVisible(false)}
        footer={null}
      >
        <div style={{ marginBottom: 16 }}>
          <p><strong>Formato do CSV:</strong></p>
          <p>O arquivo deve conter as seguintes colunas (COM cabeçalho):</p>
          <ol>
            <li>Nome (obrigatório)</li>
            <li>Código</li>
            <li>Descrição</li>
            <li>Categoria (material/medicine/equipment/consumable)</li>
            <li>Quantidade</li>
            <li>Estoque Mínimo</li>
            <li>Unidade</li>
            <li>Preço Custo</li>
            <li>Preço Venda</li>
          </ol>
          <p><strong>Exemplo:</strong></p>
          <code>Nome,Código,Descrição,Categoria,Quantidade,Estoque Mínimo,Unidade,Preço Custo,Preço Venda</code>
          <br />
          <code>Luva Procedimento,LUV001,Luva descartável,material,100,20,cx,15.50,25.00</code>
        </div>

        <Upload
          accept=".csv"
          beforeUpload={handleImportCSV}
          showUploadList={false}
        >
          <Button
            icon={<UploadOutlined />}
            loading={uploading}
            block
            type="primary"
          >
            {uploading ? 'Importando...' : 'Selecionar arquivo CSV'}
          </Button>
        </Upload>
      </Modal>
    </div>
  );
};

export default Products;
