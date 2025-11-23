import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Input, Space, Popconfirm, message, Tag, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SearchOutlined } from '@ant-design/icons';
import { tasksAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const { Option } = Select;

const Tasks = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  useEffect(() => {
    loadTasks();
  }, [pagination.current, search, statusFilter, priorityFilter]);

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
      render: (_, record) => (
        <Space>
          <Button
            icon={<EyeOutlined />}
            onClick={() => navigate(`/tasks/${record.id}`)}
            size="small"
          />
          {canEdit('tasks') && (
            <Button
              icon={<EditOutlined />}
              onClick={() => navigate(`/tasks/${record.id}/edit`)}
              size="small"
            />
          )}
          {canDelete('tasks') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button icon={<DeleteOutlined />} danger size="small" />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16, alignItems: 'center' }}>
        <h1>Tarefas</h1>
        {canCreate('tasks') && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/tasks/new')}
            className="mobile-compact-btn"
            style={{ whiteSpace: 'nowrap' }}
          >
            <span className="desktop-text">Nova Tarefa</span>
            <span className="mobile-text">Nova</span>
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

      <div style={{ overflowX: 'auto' }}>
        <Table
          columns={columns}
          dataSource={tasks}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={(newPagination) => setPagination(newPagination)}
          scroll={{ x: 'max-content' }}
        />
      </div>
    </div>
  );
};

export default Tasks;
