import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Tabs,
  Empty,
  Spin,
  message,
  Popconfirm,
  Row,
  Col,
} from 'antd';
import {
  CalendarOutlined,
  PlusOutlined,
  ClockCircleOutlined,
  UserOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';

const { Title, Text } = Typography;

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

  const AppointmentCard = ({ appointment, showActions = false }) => (
    <Card
      size="small"
      style={{
        marginBottom: 12,
        borderRadius: 12,
        background: 'linear-gradient(135deg, #f8fdf9 0%, #f0f9f2 100%)',
        border: '1px solid #d9f0df',
      }}
    >
      <Row gutter={[12, 12]}>
        {/* Data */}
        <Col xs={12} sm={6}>
          <div style={{ textAlign: 'center' }}>
            <div
              style={{
                width: 50,
                height: 50,
                margin: '0 auto 8px',
                borderRadius: 12,
                background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <CalendarOutlined style={{ fontSize: 22, color: '#fff' }} />
            </div>
            <Text strong style={{ display: 'block', fontSize: 14 }}>
              {dayjs(appointment.start_time).format('DD/MM')}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {dayjs(appointment.start_time).format('dddd')}
            </Text>
          </div>
        </Col>

        {/* Horário */}
        <Col xs={12} sm={6}>
          <div style={{ textAlign: 'center' }}>
            <ClockCircleOutlined style={{ fontSize: 24, color: '#66BB6A', marginBottom: 8 }} />
            <Text strong style={{ display: 'block' }}>
              {dayjs(appointment.start_time).format('HH:mm')}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              ate {dayjs(appointment.end_time).format('HH:mm')}
            </Text>
          </div>
        </Col>

        {/* Profissional */}
        <Col xs={12} sm={6}>
          <div style={{ textAlign: 'center' }}>
            <UserOutlined style={{ fontSize: 24, color: '#66BB6A', marginBottom: 8 }} />
            <Text strong style={{ display: 'block', fontSize: 13 }}>
              {appointment.dentist?.name?.split(' ')[0] || 'Profissional'}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {procedureLabels[appointment.procedure] || 'Consulta'}
            </Text>
          </div>
        </Col>

        {/* Status */}
        <Col xs={12} sm={6}>
          <div style={{ textAlign: 'center' }}>
            <Tag
              color={statusColors[appointment.status]}
              style={{ marginBottom: 8 }}
            >
              {statusLabels[appointment.status]}
            </Tag>
            {showActions && (appointment.status === 'scheduled' || appointment.status === 'confirmed') && (
              <Popconfirm
                title="Cancelar consulta"
                description="Tem certeza?"
                onConfirm={() => handleCancelAppointment(appointment.id)}
                okText="Sim"
                cancelText="Nao"
              >
                <Button
                  danger
                  size="small"
                  icon={<DeleteOutlined />}
                  block
                >
                  Cancelar
                </Button>
              </Popconfirm>
            )}
          </div>
        </Col>
      </Row>
    </Card>
  );

  const renderAppointmentList = (appointments, showActions = false) => {
    if (loading) {
      return (
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin size="large" />
        </div>
      );
    }

    if (appointments.length === 0) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={showActions ? "Nenhuma consulta agendada" : "Nenhuma consulta no historico"}
        >
          {showActions && (
            <Button
              type="primary"
              onClick={() => navigate('/patient/book')}
              disabled={hasPending}
            >
              Agendar Consulta
            </Button>
          )}
        </Empty>
      );
    }

    return (
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        {appointments.map((appointment) => (
          <AppointmentCard
            key={appointment.id}
            appointment={appointment}
            showActions={showActions}
          />
        ))}
      </Space>
    );
  };

  const tabItems = [
    {
      key: 'upcoming',
      label: (
        <span>
          <CalendarOutlined /> Proximas ({upcomingAppointments.length})
        </span>
      ),
      children: renderAppointmentList(upcomingAppointments, true),
    },
    {
      key: 'past',
      label: (
        <span>
          <ClockCircleOutlined /> Historico ({pastAppointments.length})
        </span>
      ),
      children: renderAppointmentList(pastAppointments, false),
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
            Agendar
          </Button>
        }
        bodyStyle={{ padding: '12px 16px' }}
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
