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
  Progress,
  Statistic,
  Typography,
} from 'antd';
import {
  EyeOutlined,
  DeleteOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  MedicineBoxOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { treatmentsAPI, patientsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';
import { usePermission } from '../../contexts/AuthContext';

const { Text } = Typography;

const Treatments = () => {
  const navigate = useNavigate();
  const { canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [treatments, setTreatments] = useState([]);
  const [patients, setPatients] = useState([]);
  const [stats, setStats] = useState({
    total: 0,
    inProgress: 0,
    completed: 0,
    totalValue: 0,
    totalPaid: 0,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const [filters, setFilters] = useState({
    patient_id: undefined,
    status: undefined,
  });

  const statusOptions = [
    { value: 'in_progress', label: 'Andando', color: statusColors.inProgress, icon: <ClockCircleOutlined /> },
    { value: 'completed', label: 'Concluído', color: statusColors.success, icon: <CheckCircleOutlined /> },
    { value: 'cancelled', label: 'Cancelado', color: statusColors.error, icon: <CloseCircleOutlined /> },
  ];

  useEffect(() => {
    fetchTreatments();
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

  const fetchTreatments = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await treatmentsAPI.getAll(params);
      const treatmentsList = response.data.treatments || [];
      setTreatments(treatmentsList);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });

      // Calculate stats
      const inProgress = treatmentsList.filter(t => t.status === 'in_progress').length;
      const completed = treatmentsList.filter(t => t.status === 'completed').length;
      const totalValue = treatmentsList.reduce((sum, t) => sum + (t.total_value || 0), 0);
      const totalPaid = treatmentsList.reduce((sum, t) => sum + (t.paid_value || 0), 0);

      setStats({
        total: treatmentsList.length,
        inProgress,
        completed,
        totalValue,
        totalPaid,
      });
    } catch (error) {
      message.error('Erro ao carregar tratamentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await treatmentsAPI.delete(id);
      message.success('Tratamento excluído com sucesso');
      fetchTreatments();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao excluir tratamento');
    }
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getStatusTag = (status) => {
    const statusObj = statusOptions.find((s) => s.value === status);
    return statusObj ? (
      <Tag color={statusObj.color} icon={statusObj.icon}>
        {statusObj.label}
      </Tag>
    ) : (
      <Tag>{status}</Tag>
    );
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value || 0);
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'start_date',
      key: 'start_date',
      width: 100,
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
      title: 'Descrição',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: 'Total',
      dataIndex: 'total_value',
      key: 'total_value',
      width: 110,
      render: (value) => formatCurrency(value),
    },
    {
      title: 'Pago',
      dataIndex: 'paid_value',
      key: 'paid_value',
      width: 110,
      render: (value) => (
        <Text type="success">{formatCurrency(value)}</Text>
      ),
    },
    {
      title: 'Restante',
      key: 'remaining',
      width: 110,
      render: (_, record) => {
        const remaining = record.total_value - record.paid_value;
        return (
          <Text type={remaining > 0 ? 'danger' : 'success'}>
            {formatCurrency(remaining)}
          </Text>
        );
      },
    },
    {
      title: 'Progresso',
      key: 'progress',
      width: 100,
      render: (_, record) => {
        const percent = record.total_value > 0
          ? Math.round((record.paid_value / record.total_value) * 100)
          : 0;
        return (
          <Progress
            percent={percent}
            size="small"
            status={percent >= 100 ? 'success' : 'active'}
          />
        );
      },
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
      width: 80,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/treatments/${record.id}`)}
            title="Visualizar / Gerenciar Pagamentos"
            style={{ color: actionColors.view }}
          />
          {canDelete('budgets') && record.status !== 'completed' && (
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
        <Col xs={24} sm={12} md={6}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Total de Tratamentos"
              value={stats.total}
              prefix={<MedicineBoxOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Em Andamento"
              value={stats.inProgress}
              prefix={<ClockCircleOutlined style={{ color: statusColors.inProgress }} />}
              valueStyle={{ color: statusColors.inProgress }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Valor Total"
              value={stats.totalValue}
              precision={2}
              prefix={<DollarOutlined style={{ color: '#722ed1' }} />}
              formatter={(value) => formatCurrency(value)}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Total Recebido"
              value={stats.totalPaid}
              precision={2}
              prefix={<CheckCircleOutlined style={{ color: statusColors.success }} />}
              valueStyle={{ color: statusColors.success }}
              formatter={(value) => formatCurrency(value)}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <MedicineBoxOutlined />
            <span>Tratamentos</span>
          </Space>
        }
        style={{ boxShadow: shadows.small }}
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Filtrar por paciente"
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
            <Button onClick={fetchTreatments} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={treatments}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={handleTableChange}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>
    </div>
  );
};

export default Treatments;
