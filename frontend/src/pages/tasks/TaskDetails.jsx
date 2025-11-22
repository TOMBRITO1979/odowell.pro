import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Descriptions, Button, Space, Tag, message, Spin } from 'antd';
import { EditOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { tasksAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const TaskDetails = () => {
  const [task, setTask] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { id } = useParams();
  const { canEdit } = usePermission();

  useEffect(() => {
    loadTask();
  }, [id]);

  const loadTask = async () => {
    setLoading(true);
    try {
      const response = await tasksAPI.getOne(id);
      setTask(response.data.task);
    } catch (error) {
      message.error('Erro ao carregar tarefa');
      navigate('/tasks');
    } finally {
      setLoading(false);
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

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!task) {
    return null;
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>
          Voltar
        </Button>
        {canEdit('tasks') && (
          <Button type="primary" icon={<EditOutlined />} onClick={() => navigate(`/tasks/${id}/edit`)}>
            Editar
          </Button>
        )}
      </Space>

      <Card title="Detalhes da Tarefa">
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Título" span={2}>
            {task.title}
          </Descriptions.Item>

          <Descriptions.Item label="Descrição" span={2}>
            {task.description || '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Status">
            <Tag color={getStatusColor(task.status)}>
              {getStatusLabel(task.status)}
            </Tag>
          </Descriptions.Item>

          <Descriptions.Item label="Prioridade">
            <Tag color={getPriorityColor(task.priority)}>
              {getPriorityLabel(task.priority)}
            </Tag>
          </Descriptions.Item>

          <Descriptions.Item label="Data de Vencimento">
            {task.due_date ? new Date(task.due_date).toLocaleDateString('pt-BR') : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Criado por">
            {task.creator?.name || '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Responsáveis" span={2}>
            {task.responsibles && task.responsibles.length > 0 ? (
              <Space wrap>
                {task.responsibles.map((responsible) => (
                  <Tag key={responsible.id}>{responsible.user?.name}</Tag>
                ))}
              </Space>
            ) : (
              '-'
            )}
          </Descriptions.Item>

          <Descriptions.Item label="Vinculado a" span={2}>
            {task.assignments && task.assignments.length > 0 ? (
              <Space wrap>
                {task.assignments.map((assignment) => (
                  <Tag key={assignment.id}>
                    {assignment.assignable_type === 'patient' ? 'Paciente' : assignment.assignable_type} #{assignment.assignable_id}
                  </Tag>
                ))}
              </Space>
            ) : (
              '-'
            )}
          </Descriptions.Item>

          <Descriptions.Item label="Criado em">
            {new Date(task.created_at).toLocaleString('pt-BR')}
          </Descriptions.Item>

          <Descriptions.Item label="Atualizado em">
            {new Date(task.updated_at).toLocaleString('pt-BR')}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
};

export default TaskDetails;
