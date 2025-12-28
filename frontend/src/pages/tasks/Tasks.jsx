import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Input, Space, Popconfirm, message, Tag, Select, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SearchOutlined } from '@ant-design/icons';
import { tasksAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { Option } = Select;

const Tasks = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    loadTasks();
  }, [pagination.current, pagination.pageSize, search, statusFilter, priorityFilter]);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const response = await tasksAPI.getAll({
        page: pagination.current,
        page_size: pagination.pageSize,
        search,
        status: statusFilter,
        priority: priorityFilter,
      });
      setTasks(response.data.tasks || []);
      setPagination(prev => ({ ...prev, total: response.data.total }));
    } catch (error) {
      message.error('Erro ao carregar tarefas');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await tasksAPI.delete(id);
      message.success('Tarefa excluída com sucesso');
      loadTasks();
    } catch (error) {
      message.error('Erro ao excluir tarefa');
    }
  };

  const getPriorityColor = (priority) => {
    const colors = {
      low: 'blue',
      medium: 'orange',
      high: 'red',
      urgent: 'purple',
    };
    return colors[priority] || 'default';
  };

  const getStatusColor = (status) => {
    const colors = {
      pending: 'default',
      in_progress: 'processing',
      completed: 'success',
      cancelled: 'error',
    };
    return colors[status] || 'default';
  };

  const getStatusLabel = (status) => {
    const labels = {
      pending: 'Pendente',
      in_progress: 'Em Andamento',
      completed: 'Concluída',
      cancelled: 'Cancelada',
    };
    return labels[status] || status;
  };

  const getPriorityLabel = (priority) => {
    const labels = {
      low: 'Baixa',
      medium: 'Média',
      high: 'Alta',
      urgent: 'Urgente',
    };
    return labels[priority] || priority;
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (tasks.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhuma tarefa encontrada</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {tasks.map((record) => (
          <Card key={record.id} size="small" style={{ borderLeft: `4px solid ${getPriorityColor(record.priority) === 'red' ? '#ff4d4f' : getPriorityColor(record.priority) === 'orange' ? '#faad14' : getPriorityColor(record.priority) === 'purple' ? '#722ed1' : '#1890ff'}` }} bodyStyle={{ padding: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
              <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.title}</div>
              <Tag color={getStatusColor(record.status)}>{getStatusLabel(record.status)}</Tag>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
              <div><strong>Prioridade:</strong><br /><Tag color={getPriorityColor(record.priority)}>{getPriorityLabel(record.priority)}</Tag></div>
              <div><strong>Vencimento:</strong><br />{record.due_date ? new Date(record.due_date).toLocaleDateString('pt-BR') : '-'}</div>
              <div><strong>Criado por:</strong><br />{record.creator?.name || '-'}</div>
              <div><strong>Responsáveis:</strong><br />{record.responsibles?.map(r => r.user?.name).filter(Boolean).join(', ') || '-'}</div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/tasks/${record.id}`)} style={{ color: actionColors.view }}>Ver</Button>
              {canEdit('tasks') && <Button type="text" size="small" icon={<EditOutlined />} onClick={() => navigate(`/tasks/${record.id}/edit`)} style={{ color: actionColors.edit }}>Editar</Button>}
              {canDelete('tasks') && (
                <Popconfirm title="Tem certeza?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                  <Button type="text" size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
                </Popconfirm>
              )}
            </div>
          </Card>
        ))}
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '16px', marginTop: '16px', padding: '12px', background: '#fafafa', borderRadius: '8px' }}>
          <Button disabled={pagination.current === 1} onClick={() => setPagination(prev => ({ ...prev, current: prev.current - 1 }))}>Anterior</Button>
          <span style={{ fontSize: '13px' }}>Pág. {pagination.current} de {Math.ceil(pagination.total / pagination.pageSize) || 1}</span>
          <Button disabled={pagination.current >= Math.ceil(pagination.total / pagination.pageSize)} onClick={() => setPagination(prev => ({ ...prev, current: prev.current + 1 }))}>Próxima</Button>
        </div>
      </div>
    );
  };

  const columns = [
    {
      title: 'Título',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: 'Prioridade',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {getPriorityLabel(priority)}
        </Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusLabel(status)}
        </Tag>
      ),
    },
    {
      title: 'Data de Vencimento',
      dataIndex: 'due_date',
      key: 'due_date',
      render: (date) => date ? new Date(date).toLocaleDateString('pt-BR') : '-',
    },
    {
      title: 'Criado por',
      dataIndex: ['creator', 'name'],
      key: 'creator',
      render: (name) => name || '-',
    },
    {
      title: 'Responsáveis',
      dataIndex: 'responsibles',
      key: 'responsibles',
      render: (responsibles) => {
        if (!responsibles || responsibles.length === 0) return '-';
        return responsibles.map(r => r.user?.name).filter(Boolean).join(', ');
      },
    },
    {
      title: 'Ações',
      key: 'actions',
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/tasks/${record.id}`)}
            size="small"
            style={{ color: actionColors.view }}
            title="Visualizar"
          />
          {canEdit('tasks') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/tasks/${record.id}/edit`)}
              size="small"
              style={{ color: actionColors.edit }}
              title="Editar"
            />
          )}
          {canDelete('tasks') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                icon={<DeleteOutlined />}
                size="small"
                style={{ color: actionColors.delete }}
                title="Excluir"
              />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16, alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
        <h1 style={{ margin: 0 }}>Tarefas</h1>
        {canCreate('tasks') && (
          <Button
            icon={<PlusOutlined />}
            onClick={() => navigate('/tasks/new')}
            style={{
              backgroundColor: actionColors.create,
              borderColor: actionColors.create,
              color: '#fff'
            }}
          >
            <span className="btn-text-desktop">Nova Tarefa</span>
            <span className="btn-text-mobile">Nova</span>
          </Button>
        )}
      </div>

      <Space style={{ marginBottom: 16 }} wrap>
        <Input
          placeholder="Buscar por título ou descrição"
          prefix={<SearchOutlined />}
          onChange={(e) => setSearch(e.target.value)}
          style={{ width: 300, maxWidth: '100%' }}
        />
        <Select
          placeholder="Status"
          allowClear
          style={{ width: 150 }}
          onChange={setStatusFilter}
          value={statusFilter || undefined}
        >
          <Option value="pending">Pendente</Option>
          <Option value="in_progress">Em Andamento</Option>
          <Option value="completed">Concluída</Option>
          <Option value="cancelled">Cancelada</Option>
        </Select>
        <Select
          placeholder="Prioridade"
          allowClear
          style={{ width: 150 }}
          onChange={setPriorityFilter}
          value={priorityFilter || undefined}
        >
          <Option value="low">Baixa</Option>
          <Option value="medium">Média</Option>
          <Option value="high">Alta</Option>
          <Option value="urgent">Urgente</Option>
        </Select>
      </Space>

      {isMobile ? renderMobileCards() : (
        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={tasks}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
            }}
            onChange={(newPagination) => setPagination(newPagination)}
            scroll={{ x: 'max-content' }}
          />
        </div>
      )}
    </div>
  );
};

export default Tasks;
