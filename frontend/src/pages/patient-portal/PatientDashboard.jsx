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
  Empty,
  Spin,
  message,
  Popconfirm,
} from 'antd';
import {
  CalendarOutlined,
  ClockCircleOutlined,
  UserOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  PlusOutlined,
  MedicineBoxOutlined,
  DeleteOutlined,
  FileTextOutlined,
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
      <Title level={4} style={{ marginBottom: 16 }}>Bem-vindo(a), {user?.name}!</Title>

      <Row gutter={[16, 16]}>
        {/* Clinic Info Card */}
        <Col xs={24} lg={12}>
          <Card
            title={<><EnvironmentOutlined style={{ marginRight: 8 }} />Dados da Clinica</>}
            size="small"
          >
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
                <Descriptions.Item label="Horario">
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
          <Card
            title={<><UserOutlined style={{ marginRight: 8 }} />Profissionais</>}
            size="small"
          >
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
            title={<><CalendarOutlined style={{ marginRight: 8 }} />Proximas Consultas</>}
            size="small"
          >
            {/* Aviso e Botão em linhas separadas */}
            {hasPending && (
              <Paragraph type="warning" style={{ marginBottom: 12 }}>
                Voce ja possui uma consulta agendada. Para agendar outra, aguarde a conclusao ou cancele a atual.
              </Paragraph>
            )}

            <div style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate('/patient/book')}
                disabled={hasPending}
                block
                size="large"
              >
                Agendar Nova Consulta
              </Button>
            </div>

            {upcomingAppointments.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                {upcomingAppointments.map((appointment) => (
                  <Card
                    key={appointment.id}
                    size="small"
                    style={{
                      borderRadius: 12,
                      background: 'linear-gradient(135deg, #f8fdf9 0%, #f0f9f2 100%)',
                      border: '1px solid #d9f0df',
                    }}
                  >
                    <Row gutter={[12, 12]}>
                      {/* Data e Hora */}
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
                          {(appointment.status === 'scheduled' || appointment.status === 'confirmed') && (
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
                ))}
              </Space>
            ) : (
              <Empty
                description="Voce nao possui consultas agendadas"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Card>
        </Col>

        {/* Quick Links */}
        <Col span={24}>
          <Card title="Acesso Rapido" size="small">
            <Row gutter={[12, 12]}>
              <Col xs={12} sm={8}>
                <Button
                  block
                  size="large"
                  icon={<CalendarOutlined />}
                  onClick={() => navigate('/patient/appointments')}
                  style={{ height: 56, whiteSpace: 'normal', lineHeight: 1.2 }}
                >
                  Historico
                </Button>
              </Col>
              <Col xs={12} sm={8}>
                <Button
                  block
                  size="large"
                  icon={<FileTextOutlined />}
                  onClick={() => navigate('/patient/medical-records')}
                  style={{ height: 56, whiteSpace: 'normal', lineHeight: 1.2 }}
                >
                  Prontuarios
                </Button>
              </Col>
              <Col xs={24} sm={8}>
                <Button
                  block
                  size="large"
                  icon={<UserOutlined />}
                  onClick={() => navigate('/patient/profile')}
                  style={{ height: 56, whiteSpace: 'normal', lineHeight: 1.2 }}
                >
                  Meus Dados
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
