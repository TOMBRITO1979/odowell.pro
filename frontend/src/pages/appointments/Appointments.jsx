import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, message, Tag, Space, Popconfirm, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { appointmentsAPI } from '../../services/api';
import dayjs from 'dayjs';

const Appointments = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const navigate = useNavigate();

  useEffect(() => {
    fetchAppointments();
  }, [pagination.current]);

  const fetchAppointments = () => {
    setLoading(true);
    appointmentsAPI.getAll({ page: pagination.current, page_size: pagination.pageSize })
      .then(res => {
        setData(res.data.appointments || []);
        setPagination({
          ...pagination,
          total: res.data.total || 0,
        });
      })
      .catch(() => message.error('Erro ao carregar agendamentos'))
      .finally(() => setLoading(false));
  };

  const handleDelete = async (id) => {
    try {
      await appointmentsAPI.delete(id);
      message.success('Agendamento deletado com sucesso');
      fetchAppointments();
    } catch (error) {
      message.error('Erro ao deletar agendamento');
    }
  };

  const getStatusTag = (status) => {
    const statusConfig = {
      scheduled: { color: 'blue', text: 'Agendado' },
      confirmed: { color: 'green', text: 'Confirmado' },
      in_progress: { color: 'orange', text: 'Em Atendimento' },
      completed: { color: 'success', text: 'Concluído' },
      cancelled: { color: 'red', text: 'Cancelado' },
      no_show: { color: 'default', text: 'Faltou' },
    };
    const config = statusConfig[status] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getProcedureText = (procedure) => {
    const procedures = {
      consultation: 'Consulta',
      cleaning: 'Limpeza',
      filling: 'Restauração',
      extraction: 'Extração',
      root_canal: 'Canal',
      orthodontics: 'Ortodontia',
      whitening: 'Clareamento',
      prosthesis: 'Prótese',
      implant: 'Implante',
      emergency: 'Emergência',
      other: 'Outro',
    };
    return procedures[procedure] || procedure;
  };

  const columns = [
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient_name',
      render: (text) => text || 'N/A',
    },
    {
      title: 'Data/Hora',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (text) => text ? dayjs(text).format('DD/MM/YYYY HH:mm') : 'N/A',
    },
    {
      title: 'Procedimento',
      dataIndex: 'procedure',
      key: 'procedure',
      render: (text) => getProcedureText(text),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/appointments/${record.id}`)}
            title="Visualizar"
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/appointments/${record.id}/edit`)}
            title="Editar"
          />
          <Popconfirm
            title="Tem certeza que deseja deletar este agendamento?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title="Deletar"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="Agendamentos"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/appointments/new')}
          >
            Novo Agendamento
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            onChange: (page) => setPagination({ ...pagination, current: page }),
          }}
        />
      </Card>
    </div>
  );
};

export default Appointments;
