import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Form, Input, Button, Select, DatePicker, Card, message } from 'antd';
import { tasksAPI, usersAPI, patientsAPI } from '../../services/api';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;

const TaskForm = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [patients, setPatients] = useState([]);
  const navigate = useNavigate();
  const { id } = useParams();
  const isEditing = !!id;

  useEffect(() => {
    loadUsers();
    loadPatients();
    if (isEditing) {
      loadTask();
    }
  }, [id]);

  const loadUsers = async () => {
    try {
      const response = await usersAPI.getAll();
      setUsers(response.data.users || []);
    } catch (error) {
    }
  };

  const loadPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page_size: 100 });
      setPatients(response.data.patients || []);
    } catch (error) {
    }
  };

  const loadTask = async () => {
    setLoading(true);
    try {
      const response = await tasksAPI.getOne(id);
      const task = response.data.task;
      form.setFieldsValue({
        title: task.title,
        description: task.description,
        priority: task.priority,
        status: task.status,
        due_date: task.due_date ? dayjs(task.due_date) : null,
        responsible_ids: task.responsibles?.map(r => r.user_id) || [],
        patient_id: task.assignments?.find(a => a.assignable_type === 'patient')?.assignable_id,
      });
    } catch (error) {
      message.error('Erro ao carregar tarefa');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        title: values.title,
        description: values.description,
        priority: values.priority,
        status: values.status,
        due_date: values.due_date ? values.due_date.toISOString() : null,
        responsible_ids: values.responsible_ids || [],
        assignments: values.patient_id ? [
          {
            assignable_type: 'patient',
            assignable_id: values.patient_id,
          }
        ] : [],
      };

      if (isEditing) {
        await tasksAPI.update(id, data);
        message.success('Tarefa atualizada com sucesso');
      } else {
        await tasksAPI.create(data);
        message.success('Tarefa criada com sucesso');
      }
      navigate('/tasks');
    } catch (error) {
      message.error(isEditing ? 'Erro ao atualizar tarefa' : 'Erro ao criar tarefa');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card title={isEditing ? 'Editar Tarefa' : 'Nova Tarefa'}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            priority: 'medium',
            status: 'pending',
          }}
        >
          <Form.Item
            label="Título"
            name="title"
            rules={[{ required: true, message: 'Por favor, informe o título da tarefa' }]}
          >
            <Input placeholder="Título da tarefa" />
          </Form.Item>

          <Form.Item
            label="Descrição"
            name="description"
          >
            <TextArea rows={4} placeholder="Descrição detalhada da tarefa" />
          </Form.Item>

          <Form.Item
            label="Prioridade"
            name="priority"
            rules={[{ required: true, message: 'Por favor, selecione a prioridade' }]}
          >
            <Select placeholder="Selecione a prioridade">
              <Option value="low">Baixa</Option>
              <Option value="medium">Média</Option>
              <Option value="high">Alta</Option>
              <Option value="urgent">Urgente</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Status"
            name="status"
            rules={[{ required: true, message: 'Por favor, selecione o status' }]}
          >
            <Select placeholder="Selecione o status">
              <Option value="pending">Pendente</Option>
              <Option value="in_progress">Em Andamento</Option>
              <Option value="completed">Concluída</Option>
              <Option value="cancelled">Cancelada</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Data de Vencimento"
            name="due_date"
          >
            <DatePicker
              format="DD/MM/YYYY"
              style={{ width: '100%' }}
              placeholder="Selecione a data"
            />
          </Form.Item>

          <Form.Item
            label="Responsáveis"
            name="responsible_ids"
          >
            <Select
              mode="multiple"
              placeholder="Selecione os responsáveis"
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {users.map(user => (
                <Option key={user.id} value={user.id}>
                  {user.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="Paciente (Opcional)"
            name="patient_id"
          >
            <Select
              placeholder="Selecione um paciente"
              allowClear
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {patients.map(patient => (
                <Option key={patient.id} value={patient.id}>
                  {patient.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} style={{ marginRight: 8 }}>
              {isEditing ? 'Atualizar' : 'Criar'}
            </Button>
            <Button onClick={() => navigate('/tasks')}>
              Cancelar
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default TaskForm;
