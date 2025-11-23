import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Modal,
  message,
  Card,
  Statistic,
  Row,
  Col,
  Popconfirm,
  Tooltip
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  PhoneOutlined,
  CalendarOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const { Search } = Input;
const { Option } = Select;

const WaitingList = () => {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({});
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0
  });

  // Filters
  const [filters, setFilters] = useState({
    status: '',
    priority: '',
    procedure: ''
  });

  useEffect(() => {
    fetchWaitingList();
    fetchStats();
  }, [pagination.current, filters]);

  const fetchWaitingList = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters
      };

      const response = await api.get('/waiting-list', { params });
      setEntries(response.data.entries || []);
      setPagination(prev => ({
        ...prev,
        total: response.data.total
      }));
    } catch (error) {
      message.error('Erro ao carregar lista de espera');
      console.error('Error fetching waiting list:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await api.get('/waiting-list/stats');
      setStats(response.data);
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const handleContact = async (id) => {
    try {
      await api.post(`/waiting-list/${id}/contact`);
      message.success('Paciente marcado como contatado');
      fetchWaitingList();
      fetchStats();
    } catch (error) {
      message.error('Erro ao marcar como contatado');
      console.error('Error contacting:', error);
    }
  };

  const handleDelete = async (id) => {
    try {
      await api.delete(`/waiting-list/${id}`);
      message.success('Removido da lista de espera');
      fetchWaitingList();
      fetchStats();
    } catch (error) {
      message.error('Erro ao remover da lista');
      console.error('Error deleting:', error);
    }
  };

  const handleTableChange = (pag) => {
    setPagination({
      ...pagination,
      current: pag.current
    });
  };

  const columns = [
    {
      title: 'Prioridade',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (priority) => (
        <Tag color={priority === 'urgent' ? 'red' : 'default'} icon={priority === 'urgent' ? <ExclamationCircleOutlined /> : null}>
          {priority === 'urgent' ? 'Urgente' : 'Normal'}
        </Tag>
      )
    },
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient',
      render: (name, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          {record.patient?.email && (
            <div style={{ fontSize: 12, color: '#999' }}>{record.patient.email}</div>
          )}
        </div>
      )
    },
    {
      title: 'Procedimento',
      dataIndex: 'procedure',
      key: 'procedure',
      ellipsis: true
    },
    {
      title: 'Dentista',
      dataIndex: ['dentist', 'name'],
      key: 'dentist',
      render: (name) => name || <Tag>Qualquer dentista</Tag>
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const colors = {
          waiting: 'blue',
          contacted: 'orange',
          scheduled: 'green',
          cancelled: 'red'
        };
        const labels = {
          waiting: 'Aguardando',
          contacted: 'Contatado',
          scheduled: 'Agendado',
          cancelled: 'Cancelado'
        };
        return <Tag color={colors[status]}>{labels[status]}</Tag>;
      }
    },
    {
      title: 'Data Cadastro',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => new Date(date).toLocaleDateString('pt-BR')
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          {record.status === 'waiting' && canEdit('appointments') && (
            <Tooltip title="Marcar como contatado">
              <Button
                type="link"
                icon={<PhoneOutlined />}
                onClick={() => handleContact(record.id)}
              />
            </Tooltip>
          )}
          {record.status !== 'scheduled' && canEdit('appointments') && (
            <Tooltip title="Agendar consulta">
              <Button
                type="link"
                icon={<CalendarOutlined />}
                onClick={() => navigate(`/appointments/new?waiting_list_id=${record.id}&patient_id=${record.patient_id}`)}
              />
            </Tooltip>
          )}
          {canEdit('appointments') && (
            <Tooltip title="Editar">
              <Button
                type="link"
                icon={<EditOutlined />}
                onClick={() => navigate(`/waiting-list/${record.id}/edit`)}
              />
            </Tooltip>
          )}
          {canDelete('appointments') && (
            <Popconfirm
              title="Remover da lista de espera?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button type="link" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          )}
        </Space>
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <h1>Lista de Espera</h1>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Aguardando"
              value={stats.total_waiting || 0}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Urgentes"
              value={stats.total_urgent || 0}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Contatados"
              value={stats.total_contacted || 0}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Agendados"
              value={stats.total_scheduled || 0}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space wrap>
            <Input
              placeholder="Buscar procedimento"
              prefix={<SearchOutlined />}
              value={filters.procedure}
              onChange={(e) => setFilters({ ...filters, procedure: e.target.value })}
              onPressEnter={fetchWaitingList}
              style={{ width: 200 }}
            />
            <Select
              placeholder="Status"
              value={filters.status || undefined}
              onChange={(value) => setFilters({ ...filters, status: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Option value="waiting">Aguardando</Option>
              <Option value="contacted">Contatado</Option>
              <Option value="scheduled">Agendado</Option>
              <Option value="cancelled">Cancelado</Option>
            </Select>
            <Select
              placeholder="Prioridade"
              value={filters.priority || undefined}
              onChange={(value) => setFilters({ ...filters, priority: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Option value="normal">Normal</Option>
              <Option value="urgent">Urgente</Option>
            </Select>
            <Button onClick={fetchWaitingList}>Filtrar</Button>
          </Space>

          {canCreate('appointments') && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => navigate('/waiting-list/new')}
            >
              Adicionar à Lista
            </Button>
          )}
        </Space>
      </Card>

      {/* Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={entries}
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

export default WaitingList;
