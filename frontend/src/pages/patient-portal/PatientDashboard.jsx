import React, { useState, useEffect } from 'react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import {
  Card,
  Row,
  Col,
  Typography,
  Button,
  Space,
  Tag,
  Descriptions,
  Divider,
  Empty,
  Spin,
  message,
} from 'antd';
import {
  CalendarOutlined,
  ClockCircleOutlined,
  UserOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

const { Title, Text, Paragraph } = Typography;

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

const PatientDashboard = () => {
  const navigate = useNavigate();
  const { clinicInfo } = useOutletContext();
  const { user } = useAuth();
  const [upcomingAppointments, setUpcomingAppointments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [hasPending, setHasPending] = useState(false);

  useEffect(() => {
    fetchUpcomingAppointments();
  }, []);

  const fetchUpcomingAppointments = async () => {
    try {
      const response = await patientPortalAPI.getAppointments('upcoming');
      const appointments = response.data.appointments || [];
      setUpcomingAppointments(appointments);
      setHasPending(appointments.length > 0);
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
      fetchUpcomingAppointments();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cancelar consulta');
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Title level={4}>Bem-vindo(a), {user?.name}!</Title>

      <Row gutter={[24, 24]}>
        {/* Clinic Info Card */}
        <Col xs={24} lg={12}>
          <Card title="Dados da Clinica" extra={<EnvironmentOutlined />}>
            {clinicInfo?.clinic ? (
              <Descriptions column={1} size="small">
                <Descriptions.Item label="Nome">
                  {clinicInfo.clinic.name}
                </Descriptions.Item>
                <Descriptions.Item label="Endereco">
                  {clinicInfo.clinic.address}
                  {clinicInfo.clinic.city && `, ${clinicInfo.clinic.city}`}
                  {clinicInfo.clinic.state && ` - ${clinicInfo.clinic.state}`}
                </Descriptions.Item>
                <Descriptions.Item label="Telefone">
                  <PhoneOutlined /> {clinicInfo.clinic.phone || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Horario de Funcionamento">
                  <ClockCircleOutlined /> {clinicInfo.working_hours?.start || '08:00'} - {clinicInfo.working_hours?.end || '18:00'}
                </Descriptions.Item>
              </Descriptions>
            ) : (
              <Empty description="Informacoes nao disponiveis" />
            )}
          </Card>
        </Col>

        {/* Dentists Card */}
        <Col xs={24} lg={12}>
          <Card title="Profissionais" extra={<UserOutlined />}>
            {clinicInfo?.dentists && clinicInfo.dentists.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }}>
                {clinicInfo.dentists.map((dentist) => (
                  <div key={dentist.id} style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
                    <Text strong>{dentist.name}</Text>
                    {dentist.specialty && (
                      <Tag color="blue" style={{ marginLeft: 8 }}>{dentist.specialty}</Tag>
                    )}
                    {dentist.cro && (
                      <Text type="secondary" style={{ display: 'block', fontSize: 12 }}>
                        CRO: {dentist.cro}
                      </Text>
                    )}
                  </div>
                ))}
              </Space>
            ) : (
              <Empty description="Nenhum profissional cadastrado" />
            )}
          </Card>
        </Col>

        {/* Upcoming Appointments */}
        <Col span={24}>
          <Card
            title="Proximas Consultas"
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
            {hasPending && (
              <Paragraph type="warning" style={{ marginBottom: 16 }}>
                Voce ja possui uma consulta agendada. Para agendar outra, aguarde a conclusao ou cancele a consulta atual.
              </Paragraph>
            )}

            {upcomingAppointments.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }} size="large">
                {upcomingAppointments.map((appointment) => (
                  <Card
                    key={appointment.id}
                    size="small"
                    style={{ background: '#fafafa' }}
                  >
                    <Row gutter={16} align="middle">
                      <Col xs={24} sm={8}>
                        <Space>
                          <CalendarOutlined style={{ fontSize: 24, color: '#66BB6A' }} />
                          <div>
                            <Text strong style={{ display: 'block' }}>
                              {dayjs(appointment.start_time).format('DD/MM/YYYY')}
                            </Text>
                            <Text type="secondary">
                              {dayjs(appointment.start_time).format('HH:mm')} - {dayjs(appointment.end_time).format('HH:mm')}
                            </Text>
                          </div>
                        </Space>
                      </Col>
                      <Col xs={24} sm={8}>
                        <Text>
                          <UserOutlined /> {appointment.dentist?.name || 'Nao definido'}
                        </Text>
                        <br />
                        <Text type="secondary">
                          {procedureLabels[appointment.procedure] || appointment.procedure}
                        </Text>
                      </Col>
                      <Col xs={24} sm={4}>
                        <Tag color={statusColors[appointment.status]}>
                          {statusLabels[appointment.status]}
                        </Tag>
                      </Col>
                      <Col xs={24} sm={4} style={{ textAlign: 'right' }}>
                        {(appointment.status === 'scheduled' || appointment.status === 'confirmed') && (
                          <Button
                            danger
                            size="small"
                            onClick={() => handleCancelAppointment(appointment.id)}
                          >
                            Cancelar
                          </Button>
                        )}
                      </Col>
                    </Row>
                  </Card>
                ))}
              </Space>
            ) : (
              <Empty
                description="Voce nao possui consultas agendadas"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              >
                <Button type="primary" onClick={() => navigate('/patient/book')}>
                  Agendar Primeira Consulta
                </Button>
              </Empty>
            )}
          </Card>
        </Col>

        {/* Quick Links */}
        <Col span={24}>
          <Card title="Acesso Rapido">
            <Row gutter={16}>
              <Col xs={24} sm={8}>
                <Button
                  block
                  size="large"
                  icon={<CalendarOutlined />}
                  onClick={() => navigate('/patient/appointments')}
                  style={{ height: 60 }}
                >
                  Historico de Consultas
                </Button>
              </Col>
              <Col xs={24} sm={8}>
                <Button
                  block
                  size="large"
                  icon={<UserOutlined />}
                  onClick={() => navigate('/patient/profile')}
                  style={{ height: 60 }}
                >
                  Meus Dados
                </Button>
              </Col>
              <Col xs={24} sm={8}>
                <Button
                  block
                  size="large"
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => navigate('/patient/book')}
                  disabled={hasPending}
                  style={{ height: 60 }}
                >
                  Nova Consulta
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default PatientDashboard;
