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
import { useNavigate, useLocation } from 'react-router-dom';
import api from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, brandColors, spacing, shadows } from '../../theme/designSystem';

const { Search } = Input;
const { Option } = Select;

const WaitingList = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { canCreate, canEdit, canDelete } = usePermission();

  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({});
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
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
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    fetchWaitingList();
    fetchStats();
  }, [pagination.current, pagination.pageSize, filters, location.key]); // location.key força atualização ao navegar

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
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await api.get('/waiting-list/stats');
      setStats(response.data);
    } catch (error) {
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
    }
  };

  const handleTableChange = (pag) => {
    setPagination({
      ...pagination,
      current: pag.current,
      pageSize: pag.pageSize,
    });
  };

  const getStatusConfig = (status) => {
    const statusMap = {
      waiting: { color: statusColors.pending, label: 'Aguardando' },
      contacted: { color: statusColors.inProgress, label: 'Contatado' },
      scheduled: { color: statusColors.success, label: 'Agendado' },
      cancelled: { color: statusColors.cancelled, label: 'Cancelado' }
    };
    return statusMap[status] || { color: statusColors.pending, label: status };
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (entries.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum paciente na lista de espera</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {entries.map((record) => {
          const statusConfig = getStatusConfig(record.status);
          const isUrgent = record.priority === 'urgent';
          return (
            <Card
              key={record.id}
              size="small"
              style={{ borderLeft: `4px solid ${isUrgent ? statusColors.error : statusColors.success}` }}
              bodyStyle={{ padding: '12px' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.patient?.name}</div>
                <Tag
                  color={isUrgent ? statusColors.error : statusColors.success}
                  icon={isUrgent ? <ExclamationCircleOutlined /> : null}
                >
                  {isUrgent ? 'Urgente' : 'Normal'}
                </Tag>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
                <div><strong>Status:</strong><br /><Tag color={statusConfig.color}>{statusConfig.label}</Tag></div>
                <div><strong>Dentista:</strong><br />{record.dentist?.name || 'Qualquer'}</div>
                <div style={{ gridColumn: '1 / -1' }}><strong>Procedimento:</strong><br />{record.procedure || '-'}</div>
                <div><strong>Cadastro:</strong> {new Date(record.created_at).toLocaleDateString('pt-BR')}</div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '6px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
                {record.status === 'waiting' && canEdit('appointments') && (
                  <Button
                    size="small"
                    icon={<PhoneOutlined />}
                    onClick={() => handleContact(record.id)}
                    style={{ color: actionColors.approve }}
                  >
                    Contatar
                  </Button>
                )}
                {record.status !== 'scheduled' && canEdit('appointments') && (
                  <Button
                    size="small"
                    icon={<CalendarOutlined />}
                    onClick={() => navigate(`/appointments/new?waiting_list_id=${record.id}&patient_id=${record.patient_id}`)}
                    style={{ color: actionColors.create }}
                  >
                    Agendar
                  </Button>
                )}
                {canEdit('appointments') && (
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => navigate(`/waiting-list/${record.id}/edit`)}
                    style={{ color: actionColors.edit }}
                  >
                    Editar
                  </Button>
                )}
                {canDelete('appointments') && (
                  <Popconfirm title="Remover da lista?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                    <Button size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
                  </Popconfirm>
                )}
              </div>
            </Card>
          );
        })}
        <div style={{ textAlign: 'center', padding: '16px' }}>
          <span style={{ color: '#666' }}>
            Mostrando {entries.length} de {pagination.total} registros
          </span>
        </div>
      </div>
    );
  };

  const columns = [
    {
      title: 'Prioridade',
      dataIndex: 'priority',
      key: 'priority',
      width: 130,
      render: (priority) => (
        <Tag
          color={priority === 'urgent' ? statusColors.error : statusColors.success}
          icon={priority === 'urgent' ? <ExclamationCircleOutlined /> : null}
          style={{ margin: 0, whiteSpace: 'nowrap' }}
        >
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
      align: 'center',
      render: (status) => {
        const statusMap = {
          waiting: { color: statusColors.pending, label: 'Aguardando' },
          contacted: { color: statusColors.inProgress, label: 'Contatado' },
          scheduled: { color: statusColors.success, label: 'Agendado' },
          cancelled: { color: statusColors.cancelled, label: 'Cancelado' }
        };
        const config = statusMap[status] || { color: statusColors.pending, label: status };
        return <Tag color={config.color}>{config.label}</Tag>;
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
      align: 'center',
      render: (_, record) => (
        <Space>
          {record.status === 'waiting' && canEdit('appointments') && (
            <Tooltip title="Marcar como contatado">
              <Button
                type="link"
                icon={<PhoneOutlined />}
                onClick={() => handleContact(record.id)}
                style={{ color: actionColors.approve }}
              />
            </Tooltip>
          )}
          {record.status !== 'scheduled' && canEdit('appointments') && (
            <Tooltip title="Agendar consulta">
              <Button
                type="link"
                icon={<CalendarOutlined />}
                onClick={() => navigate(`/appointments/new?waiting_list_id=${record.id}&patient_id=${record.patient_id}`)}
                style={{ color: actionColors.create }}
              />
            </Tooltip>
          )}
          {canEdit('appointments') && (
            <Tooltip title="Editar">
              <Button
                type="link"
                icon={<EditOutlined />}
                onClick={() => navigate(`/waiting-list/${record.id}/edit`)}
                style={{ color: actionColors.edit }}
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
              <Button
                type="link"
                icon={<DeleteOutlined />}
                style={{ color: actionColors.delete }}
              />
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
      <Row gutter={[spacing.md, spacing.md]} style={{ marginBottom: spacing.lg }}>
        <Col xs={12} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }} bodyStyle={{ padding: isMobile ? '12px' : '24px' }}>
            <Statistic
              title="Aguardando"
              value={stats.total_waiting || 0}
              valueStyle={{ color: statusColors.pending, fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }} bodyStyle={{ padding: isMobile ? '12px' : '24px' }}>
            <Statistic
              title="Urgentes"
              value={stats.total_urgent || 0}
              valueStyle={{ color: statusColors.error, fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }} bodyStyle={{ padding: isMobile ? '12px' : '24px' }}>
            <Statistic
              title="Contatados"
              value={stats.total_contacted || 0}
              valueStyle={{ color: statusColors.inProgress, fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }} bodyStyle={{ padding: isMobile ? '12px' : '24px' }}>
            <Statistic
              title="Agendados"
              value={stats.total_scheduled || 0}
              valueStyle={{ color: statusColors.success, fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: spacing.md, boxShadow: shadows.small }}>
        {isMobile ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <Input
              placeholder="Buscar procedimento"
              prefix={<SearchOutlined />}
              value={filters.procedure}
              onChange={(e) => setFilters({ ...filters, procedure: e.target.value })}
              onPressEnter={fetchWaitingList}
              style={{ width: '100%' }}
            />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
              <Select
                placeholder="Status"
                value={filters.status || undefined}
                onChange={(value) => setFilters({ ...filters, status: value })}
                style={{ width: '100%' }}
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
                style={{ width: '100%' }}
                allowClear
              >
                <Option value="normal">Normal</Option>
                <Option value="urgent">Urgente</Option>
              </Select>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
              <Button onClick={fetchWaitingList}>Filtrar</Button>
              {canCreate('appointments') && (
                <Button
                  icon={<PlusOutlined />}
                  onClick={() => navigate('/waiting-list/new')}
                  style={{
                    backgroundColor: actionColors.create,
                    borderColor: actionColors.create,
                    color: '#fff'
                  }}
                >
                  Adicionar
                </Button>
              )}
            </div>
          </div>
        ) : (
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
                icon={<PlusOutlined />}
                onClick={() => navigate('/waiting-list/new')}
                style={{
                  backgroundColor: actionColors.create,
                  borderColor: actionColors.create,
                  color: '#fff'
                }}
              >
                Adicionar à Lista
              </Button>
            )}
          </Space>
        )}
      </Card>

      {/* Table */}
      <Card style={{ boxShadow: shadows.small }}>
        {isMobile ? renderMobileCards() : (
          <Table
            columns={columns}
            dataSource={entries}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
            }}
            onChange={handleTableChange}
            scroll={{ x: 1000 }}
          />
        )}
      </Card>
    </div>
  );
};

export default WaitingList;
