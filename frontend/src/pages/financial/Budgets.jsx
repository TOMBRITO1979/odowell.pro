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
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { budgetsAPI, patientsAPI } from '../../services/api';

const Budgets = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [budgets, setBudgets] = useState([]);
  const [patients, setPatients] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // Filters
  const [filters, setFilters] = useState({
    patient_id: undefined,
    status: undefined,
  });

  const statusOptions = [
    { value: 'pending', label: 'Pendente', color: 'warning' },
    { value: 'approved', label: 'Aprovado', color: 'success' },
    { value: 'rejected', label: 'Rejeitado', color: 'error' },
    { value: 'expired', label: 'Expirado', color: 'default' },
  ];

  useEffect(() => {
    fetchBudgets();
    fetchPatients();
  }, [pagination.current, filters]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchBudgets = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await budgetsAPI.getAll(params);
      setBudgets(response.data.budgets || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar orçamentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await budgetsAPI.delete(id);
      message.success('Orçamento excluído com sucesso');
      fetchBudgets();
    } catch (error) {
      message.error('Erro ao excluir orçamento');
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

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
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
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient',
      ellipsis: true,
    },
    {
      title: 'Profissional',
      dataIndex: ['dentist', 'name'],
      key: 'dentist',
      ellipsis: true,
    },
    {
      title: 'Valor Total',
      dataIndex: 'total_value',
      key: 'total_value',
      width: 140,
      render: (value) => formatCurrency(value),
      sorter: true,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Validade',
      dataIndex: 'valid_until',
      key: 'valid_until',
      width: 120,
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/budgets/${record.id}/view`)}
            title="Visualizar"
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/budgets/${record.id}/edit`)}
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
            <DollarOutlined />
            <span>Orçamentos</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/budgets/new')}
          >
            Novo Orçamento
          </Button>
        }
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Selecione o paciente"
              style={{ width: '100%' }}
              allowClear
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().includes(input.toLowerCase())
              }
              value={filters.patient_id}
              onChange={(value) =>
                setFilters({ ...filters, patient_id: value })
              }
            >
              {patients.map((patient) => (
                <Select.Option key={patient.id} value={patient.id}>
                  {patient.name}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Status"
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
            <Button onClick={fetchBudgets} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={budgets}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          scroll={{ x: 1000 }}
        />
      </Card>
    </div>
  );
};

export default Budgets;
