import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Table,
  Tag,
  Button,
  Space,
  Typography,
  Tabs,
  Empty,
  Spin,
  message,
  Popconfirm,
} from 'antd';
import {
  CalendarOutlined,
  PlusOutlined,
  ClockCircleOutlined,
  UserOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';

const { Title } = Typography;

const statusColors = {
  scheduled: 'blue',
  confirmed: 'green',
  in_progress: 'orange',
  completed: 'default',
  cancelled: 'red',
  no_show: 'volcano',
};

const statusLabels = {
  scheduled: 'Agendado',
  confirmed: 'Confirmado',
  in_progress: 'Em Atendimento',
  completed: 'Concluido',
  cancelled: 'Cancelado',
  no_show: 'Faltou',
};

const procedureLabels = {
  consultation: 'Consulta',
  cleaning: 'Limpeza',
  filling: 'Restauracao',
  extraction: 'Extracao',
  root_canal: 'Canal',
  orthodontics: 'Ortodontia',
  whitening: 'Clareamento',
  prosthesis: 'Protese',
  implant: 'Implante',
  emergency: 'Emergencia',
  other: 'Outro',
};

const PatientAppointments = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [upcomingAppointments, setUpcomingAppointments] = useState([]);
  const [pastAppointments, setPastAppointments] = useState([]);
  const [activeTab, setActiveTab] = useState('upcoming');
  const [hasPending, setHasPending] = useState(false);

  useEffect(() => {
    fetchAppointments();
  }, []);

  const fetchAppointments = async () => {
    setLoading(true);
    try {
      const [upcomingRes, pastRes] = await Promise.all([
        patientPortalAPI.getAppointments('upcoming'),
        patientPortalAPI.getAppointments('past'),
      ]);

      const upcoming = upcomingRes.data.appointments || [];
      setUpcomingAppointments(upcoming);
      setPastAppointments(pastRes.data.appointments || []);
      setHasPending(upcoming.length > 0);
    } catch (error) {
      message.error('Erro ao carregar consultas');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelAppointment = async (id) => {
    try {
      await patientPortalAPI.cancelAppointment(id);
      message.success('Consulta cancelada com sucesso');
      fetchAppointments();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cancelar consulta');
    }
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'start_time',
      key: 'date',
      render: (value) => (
        <Space>
          <CalendarOutlined />
          {dayjs(value).format('DD/MM/YYYY')}
        </Space>
      ),
      sorter: (a, b) => dayjs(a.start_time).unix() - dayjs(b.start_time).unix(),
    },
    {
      title: 'Horario',
      key: 'time',
      render: (_, record) => (
        <Space>
          <ClockCircleOutlined />
          {dayjs(record.start_time).format('HH:mm')} - {dayjs(record.end_time).format('HH:mm')}
        </Space>
      ),
    },
    {
      title: 'Profissional',
      key: 'dentist',
      render: (_, record) => (
        <Space>
          <UserOutlined />
          {record.dentist?.name || 'Nao definido'}
        </Space>
      ),
    },
    {
      title: 'Procedimento',
      dataIndex: 'procedure',
      key: 'procedure',
      render: (value) => procedureLabels[value] || value,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={statusColors[status]}>{statusLabels[status]}</Tag>
      ),
    },
    {
      title: 'Acoes',
      key: 'actions',
      render: (_, record) => {
        if (record.status === 'scheduled' || record.status === 'confirmed') {
          return (
            <Popconfirm
              title="Cancelar consulta"
              description="Tem certeza que deseja cancelar esta consulta?"
              onConfirm={() => handleCancelAppointment(record.id)}
              okText="Sim"
              cancelText="Nao"
            >
              <Button danger size="small">
                Cancelar
              </Button>
            </Popconfirm>
          );
        }
        return null;
      },
    },
  ];

  const tabItems = [
    {
      key: 'upcoming',
      label: (
        <span>
          <CalendarOutlined /> Proximas ({upcomingAppointments.length})
        </span>
      ),
      children: (
        <Table
          dataSource={upcomingAppointments}
          columns={columns}
          rowKey="id"
          loading={loading}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="Nenhuma consulta agendada"
              >
                <Button
                  type="primary"
                  onClick={() => navigate('/patient/book')}
                  disabled={hasPending}
                >
                  Agendar Consulta
                </Button>
              </Empty>
            ),
          }}
        />
      ),
    },
    {
      key: 'past',
      label: (
        <span>
          <ClockCircleOutlined /> Historico ({pastAppointments.length})
        </span>
      ),
      children: (
        <Table
          dataSource={pastAppointments}
          columns={columns.filter((col) => col.key !== 'actions')}
          rowKey="id"
          loading={loading}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="Nenhuma consulta no historico"
              />
            ),
          }}
        />
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <CalendarOutlined />
            <Title level={4} style={{ margin: 0 }}>
              Minhas Consultas
            </Title>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/patient/book')}
            disabled={hasPending}
          >
            Agendar Consulta
          </Button>
        }
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
        />
      </Card>
    </div>
  );
};

export default PatientAppointments;
