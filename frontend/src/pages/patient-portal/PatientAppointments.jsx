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
  MedicineBoxOutlined,
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
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
      }}
    >
      <Row gutter={[12, 12]}>
        {/* Data e Hora */}
        <Col xs={24} sm={12}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                width: 48,
                height: 48,
                borderRadius: 12,
                background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <CalendarOutlined style={{ fontSize: 20, color: '#fff' }} />
            </div>
            <div>
              <Text strong style={{ fontSize: 16, display: 'block' }}>
                {dayjs(appointment.start_time).format('DD/MM/YYYY')}
              </Text>
              <Text type="secondary">
                <ClockCircleOutlined style={{ marginRight: 4 }} />
                {dayjs(appointment.start_time).format('HH:mm')} - {dayjs(appointment.end_time).format('HH:mm')}
              </Text>
            </div>
          </div>
        </Col>

        {/* Profissional e Procedimento */}
        <Col xs={24} sm={12}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <Text>
              <UserOutlined style={{ marginRight: 6, color: '#66BB6A' }} />
              {appointment.dentist?.name || 'Nao definido'}
            </Text>
            <Text type="secondary">
              <MedicineBoxOutlined style={{ marginRight: 6 }} />
              {procedureLabels[appointment.procedure] || appointment.procedure}
            </Text>
          </div>
        </Col>

        {/* Status e Ações */}
        <Col xs={24}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8, paddingTop: 12, borderTop: '1px solid #f0f0f0' }}>
            <Tag color={statusColors[appointment.status]} style={{ margin: 0 }}>
              {statusLabels[appointment.status]}
            </Tag>
            {showActions && (appointment.status === 'scheduled' || appointment.status === 'confirmed') && (
              <Popconfirm
                title="Cancelar consulta"
                description="Tem certeza que deseja cancelar?"
                onConfirm={() => handleCancelAppointment(appointment.id)}
                okText="Sim"
                cancelText="Nao"
              >
                <Button danger size="small" icon={<DeleteOutlined />}>
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
      <div>
        {appointments.map((appointment) => (
          <AppointmentCard
            key={appointment.id}
            appointment={appointment}
            showActions={showActions}
          />
        ))}
      </div>
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
