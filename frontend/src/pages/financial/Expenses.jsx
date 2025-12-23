import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Select,
  Card,
  message,
  Popconfirm,
  Row,
  Col,
  DatePicker,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  WalletOutlined,
  ArrowDownOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { paymentsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { RangePicker } = DatePicker;

const Expenses = () => {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [expenses, setExpenses] = useState([]);
  const [statistics, setStatistics] = useState({
    total: 0,
    paid: 0,
    pending: 0,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const [filters, setFilters] = useState({
    category: undefined,
    status: undefined,
    date_range: null,
  });

  const categoryOptions = [
    { value: 'salario', label: 'Salário / Funcionários' },
    { value: 'aluguel', label: 'Aluguel' },
    { value: 'luz', label: 'Luz / Energia' },
    { value: 'agua', label: 'Água' },
    { value: 'internet', label: 'Internet / Telefone' },
    { value: 'insumos', label: 'Insumos / Materiais' },
    { value: 'equipamentos', label: 'Equipamentos' },
    { value: 'manutencao', label: 'Manutenção' },
    { value: 'limpeza', label: 'Limpeza' },
    { value: 'marketing', label: 'Marketing' },
    { value: 'impostos', label: 'Impostos / Taxas' },
    { value: 'software', label: 'Software / Sistemas' },
    { value: 'outros', label: 'Outros' },
  ];

  const statusOptions = [
    { value: 'pending', label: 'Pendente', color: statusColors.pending },
    { value: 'paid', label: 'Pago', color: statusColors.success },
    { value: 'overdue', label: 'Atrasado', color: statusColors.error },
    { value: 'cancelled', label: 'Cancelado', color: statusColors.cancelled },
  ];

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'boleto', label: 'Boleto' },
  ];

  useEffect(() => {
    fetchExpenses();
  }, [pagination.current, pagination.pageSize, filters]);

  const fetchExpenses = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        type: 'expense', // Sempre filtrar por despesas
        ...filters,
      };

      // Convert date_range to start_date and end_date if exists
      if (filters.date_range && filters.date_range.length === 2) {
        params.start_date = filters.date_range[0].format('YYYY-MM-DD');
        params.end_date = filters.date_range[1].format('YYYY-MM-DD');
        delete params.date_range;
      }

      const response = await paymentsAPI.getAll(params);
      const expensesList = response.data.payments || [];
      setExpenses(expensesList);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });

      // Calculate statistics from loaded data
      const total = expensesList.reduce((sum, e) => sum + (e.amount || 0), 0);
      const paid = expensesList.filter(e => e.status === 'paid').reduce((sum, e) => sum + (e.amount || 0), 0);
      const pending = expensesList.filter(e => e.status === 'pending' || e.status === 'overdue').reduce((sum, e) => sum + (e.amount || 0), 0);

      setStatistics({ total, paid, pending });
    } catch (error) {
      message.error('Erro ao carregar contas');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await paymentsAPI.delete(id);
      message.success('Conta excluída com sucesso');
      fetchExpenses();
    } catch (error) {
      message.error('Erro ao excluir conta');
    }
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getStatusTag = (status) => {
    const statusObj = statusOptions.find((s) => s.value === status);
    return statusObj ? (
      <Tag color={statusObj.color}>{statusObj.label}</Tag>
    ) : (
      <Tag>{status}</Tag>
    );
  };

  const getCategoryLabel = (category) => {
    const categoryObj = categoryOptions.find((c) => c.value === category);
    return categoryObj ? categoryObj.label : category || '-';
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value || 0);
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 70,
    },
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 110,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
      sorter: true,
    },
    {
      title: 'Categoria',
      dataIndex: 'category',
      key: 'category',
      width: 160,
      render: (category) => getCategoryLabel(category),
    },
    {
      title: 'Descrição',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Valor',
      dataIndex: 'amount',
      key: 'amount',
      width: 130,
      render: (amount) => (
        <span style={{ color: statusColors.error, fontWeight: 'bold' }}>
          {formatCurrency(amount)}
        </span>
      ),
      sorter: true,
    },
    {
      title: 'Método',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 140,
      render: (method) => {
        if (!method) return '-';
        const methodObj = paymentMethods.find((m) => m.value === method);
        return methodObj ? methodObj.label : method;
      },
    },
    {
      title: 'Vencimento',
      dataIndex: 'due_date',
      key: 'due_date',
      width: 110,
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Space>
          {canEdit('payments') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/expenses/${record.id}/edit`)}
              title="Editar"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('payments') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              description="Esta ação não pode ser desfeita"
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
      {/* Stats Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Total de Contas"
              value={statistics.total}
              precision={2}
              prefix={<WalletOutlined style={{ color: statusColors.error }} />}
              formatter={(value) => formatCurrency(value)}
              valueStyle={{ color: statusColors.error }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Pagas"
              value={statistics.paid}
              precision={2}
              prefix={<CheckCircleOutlined style={{ color: statusColors.success }} />}
              formatter={(value) => formatCurrency(value)}
              valueStyle={{ color: statusColors.success }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Pendentes"
              value={statistics.pending}
              precision={2}
              prefix={<ClockCircleOutlined style={{ color: statusColors.pending }} />}
              formatter={(value) => formatCurrency(value)}
              valueStyle={{ color: statusColors.pending }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <WalletOutlined />
            <span>Contas a Pagar</span>
          </Space>
        }
        extra={
          canCreate('payments') && (
            <Button
              icon={<PlusOutlined />}
              onClick={() => navigate('/expenses/new')}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              Nova Conta
            </Button>
          )
        }
        style={{ boxShadow: shadows.small }}
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Filtrar por categoria"
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
              placeholder="Filtrar por status"
              style={{ width: '100%' }}
              allowClear
              value={filters.status}
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              {statusOptions.map((status) => (
                <Select.Option key={status.value} value={status.value}>
                  {status.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <RangePicker
              style={{ width: '100%' }}
              format="DD/MM/YYYY"
              placeholder={['Data início', 'Data fim']}
              value={filters.date_range}
              onChange={(dates) => setFilters({ ...filters, date_range: dates })}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Button onClick={fetchExpenses} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={expenses}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
            }}
            onChange={handleTableChange}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>
    </div>
  );
};

export default Expenses;
