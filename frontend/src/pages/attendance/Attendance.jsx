import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Select, DatePicker, Typography, Tag, Spin, Empty } from 'antd';
import { UserOutlined, ClockCircleOutlined, MedicineBoxOutlined } from '@ant-design/icons';
import { Column, Pie } from '@ant-design/plots';
import dayjs from 'dayjs';
import { appointmentsAPI, usersAPI } from '../../services/api';
import { statusColors } from '../../theme/designSystem';

const { Title, Text } = Typography;

// Paleta de cores suaves e calmas (não ofuscam a vista)
const calmColors = {
  // Verde claro suave para gráfico de barras
  chartGreen: '#a7d7a7',
  // Azul claro suave para status
  softBlue: '#b8d4e8',
  // Lilás/roxo claro suave para profissional
  softLilac: '#d4c4e8',
  // Verde para bordas e destaques
  accentGreen: '#7bc47b',
  // Lavanda suave
  lavender: '#c9b8db',
  // Status colors suaves
  statusScheduled: '#b8d4e8',
  statusConfirmed: '#b8e0c8',
  statusInProgress: '#f5e0a8',
  statusCompleted: '#a7d7a7',
  statusCancelled: '#e8b8b8',
  statusNoShow: '#d9d9d9',
};

// Mapeamento de cores específicas por procedimento (verde para Restauração)
const procedureColorMap = {
  'Restauração': '#a7d7a7', // verde suave - destaque para restauração
  'Consulta': '#b8d4e8', // azul suave
  'Limpeza': '#a8d8d8', // teal suave
  'Extração': '#e8c4d4', // rosa suave
  'Canal': '#f5e0a8', // amarelo suave
  'Ortodontia': '#d4c4e8', // lilás suave
  'Clareamento': '#e8d0b8', // pêssego suave
  'Prótese': '#c4c8e8', // indigo suave
  'Implante': '#b8e0c8', // menta suave
  'Emergência': '#e8b8b8', // coral suave
  'Outro': '#d4e0a8', // lima suave
};

// Paleta para gráfico de procedimentos (cores suaves e calmas)
const procedureChartColors = [
  '#b8d4e8', // azul suave
  '#a8d8d8', // teal suave
  '#e8c4d4', // rosa suave
  '#f5e0a8', // amarelo suave
  '#d4c4e8', // lilás suave
  '#e8d0b8', // pêssego suave
  '#c4c8e8', // indigo suave
  '#b8e0c8', // menta suave
  '#e8b8b8', // coral suave
  '#d4e0a8', // lima suave
  '#a7d7a7', // verde suave
];

const Attendance = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [appointments, setAppointments] = useState([]);
  const [dentists, setDentists] = useState([]);
  const [filters, setFilters] = useState({
    date: dayjs(),
    dentist_id: null,
    procedure: null,
  });

  const rooms = ['Sala 1', 'Sala 2', 'Sala 3', 'Sala 4', 'Sala 5'];

  const procedureOptions = [
    { value: 'consultation', label: 'Consulta' },
    { value: 'cleaning', label: 'Limpeza' },
    { value: 'filling', label: 'Restauração' },
    { value: 'extraction', label: 'Extração' },
    { value: 'root_canal', label: 'Canal' },
    { value: 'orthodontics', label: 'Ortodontia' },
    { value: 'whitening', label: 'Clareamento' },
    { value: 'prosthesis', label: 'Prótese' },
    { value: 'implant', label: 'Implante' },
    { value: 'emergency', label: 'Emergência' },
    { value: 'other', label: 'Outro' },
  ];

  const getProcedureLabel = (value) => {
    const proc = procedureOptions.find(p => p.value === value);
    return proc ? proc.label : value;
  };

  const getStatusColor = (status) => {
    const colors = {
      scheduled: calmColors.statusScheduled,
      confirmed: calmColors.statusConfirmed,
      in_progress: calmColors.statusInProgress,
      completed: calmColors.statusCompleted,
      cancelled: calmColors.statusCancelled,
      no_show: calmColors.statusNoShow,
    };
    return colors[status] || calmColors.softBlue;
  };

  const getStatusText = (status) => {
    const texts = {
      scheduled: 'Agendado',
      confirmed: 'Confirmado',
      in_progress: 'Em Atendimento',
      completed: 'Concluído',
      cancelled: 'Cancelado',
      no_show: 'Faltou',
    };
    return texts[status] || status;
  };

  useEffect(() => {
    fetchDentists();
  }, []);

  useEffect(() => {
    fetchAppointments();
  }, [filters]);

  const fetchDentists = async () => {
    try {
      const response = await usersAPI.getAll();
      const professionals = (response.data.users || []).filter(
        u => u.role === 'dentist' || u.role === 'admin'
      );
      setDentists(professionals);
    } catch (error) {
      console.error('Error fetching dentists:', error);
    }
  };

  const fetchAppointments = async () => {
    setLoading(true);
    try {
      const params = {
        page: 1,
        page_size: 500,
      };

      if (filters.date) {
        params.start_date = filters.date.startOf('day').toISOString();
        params.end_date = filters.date.endOf('day').toISOString();
      }
      if (filters.dentist_id) params.dentist_id = filters.dentist_id;
      if (filters.procedure) params.procedure = filters.procedure;

      const response = await appointmentsAPI.getAll(params);
      setAppointments(response.data.appointments || []);
    } catch (error) {
      console.error('Error fetching appointments:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({ ...prev, [key]: value }));
  };

  // Agrupar agendamentos por sala
  const getAppointmentsByRoom = (room) => {
    return appointments
      .filter(apt => apt.room === room)
      .sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
  };

  // Agendamentos sem sala definida
  const getAppointmentsWithoutRoom = () => {
    return appointments
      .filter(apt => !apt.room || apt.room === '')
      .sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
  };

  // Dados para gráfico de pacientes por profissional
  const getChartDataByDentist = () => {
    const countByDentist = {};
    appointments.forEach(apt => {
      const dentistName = apt.dentist?.name || 'Sem profissional';
      countByDentist[dentistName] = (countByDentist[dentistName] || 0) + 1;
    });
    return Object.entries(countByDentist).map(([name, count]) => ({
      profissional: name,
      pacientes: count,
    }));
  };

  // Dados para gráfico de procedimentos
  const getChartDataByProcedure = () => {
    const countByProcedure = {};
    appointments.forEach(apt => {
      const procLabel = getProcedureLabel(apt.procedure);
      countByProcedure[procLabel] = (countByProcedure[procLabel] || 0) + 1;
    });
    return Object.entries(countByProcedure).map(([name, count]) => ({
      procedimento: name,
      quantidade: count,
    }));
  };

  // Cores suaves e foscas para cada profissional (não agridem a vista)
  const professionalColors = [
    '#a7d7a7', // verde suave
    '#b8d4e8', // azul suave
    '#d4c4e8', // lilás suave
    '#f5e0a8', // amarelo suave
    '#e8c4d4', // rosa suave
    '#a8d8d8', // teal suave
    '#e8d0b8', // pêssego suave
    '#c4c8e8', // indigo suave
    '#b8e0c8', // menta suave
    '#d4e0a8', // lima suave
  ];

  const columnConfig = {
    data: getChartDataByDentist(),
    xField: 'profissional',
    yField: 'pacientes',
    colorField: 'profissional',
    color: professionalColors,
    style: {
      radiusTopLeft: 4,
      radiusTopRight: 4,
    },
    label: {
      text: 'pacientes',
      position: 'top',
      style: {
        fill: '#6b7280',
        fontWeight: 600,
      },
    },
    axis: {
      x: {
        label: {
          autoRotate: true,
          style: {
            fill: '#6b7280',
          },
        },
      },
      y: {
        label: {
          style: {
            fill: '#9ca3af',
          },
        },
      },
    },
    legend: {
      color: {
        position: 'bottom',
        layout: { justifyContent: 'center' },
        itemLabelFill: '#4b5563',
      },
    },
  };

  // Dados do gráfico de pizza com cores mapeadas
  const pieData = getChartDataByProcedure();

  const pieConfig = {
    data: pieData,
    angleField: 'quantidade',
    colorField: 'procedimento',
    radius: 0.85,
    innerRadius: 0.55,
    style: {
      fill: ({ procedimento }) => procedureColorMap[procedimento] || '#d9d9d9',
    },
    label: {
      text: (d) => `${d.quantidade}`,
      position: 'outside',
      style: {
        fill: '#374151',
        fontWeight: 500,
      },
    },
    legend: {
      color: {
        position: 'bottom',
        layout: { justifyContent: 'center' },
        itemLabelFill: '#4b5563',
      },
    },
    tooltip: {
      title: 'procedimento',
    },
  };

  const AppointmentCard = ({ appointment }) => (
    <Card
      size="small"
      style={{
        marginBottom: 8,
        cursor: 'pointer',
        borderLeft: `4px solid ${getStatusColor(appointment.status)}`,
        transition: 'all 0.3s',
      }}
      hoverable
      onClick={() => navigate(`/appointments/${appointment.id}/edit`)}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <Text strong style={{ fontSize: 14 }}>
          <UserOutlined style={{ marginRight: 6 }} />
          {appointment.patient?.name || 'Paciente não definido'}
        </Text>
        <Text type="secondary" style={{ fontSize: 12 }}>
          <MedicineBoxOutlined style={{ marginRight: 6 }} />
          {getProcedureLabel(appointment.procedure)}
        </Text>
        <Text type="secondary" style={{ fontSize: 12 }}>
          <ClockCircleOutlined style={{ marginRight: 6 }} />
          {dayjs(appointment.start_time).format('HH:mm')} - {dayjs(appointment.end_time).format('HH:mm')}
        </Text>
        <div>
          <Tag
            style={{
              fontSize: 10,
              backgroundColor: calmColors.softBlue,
              color: '#4a5568',
              border: 'none',
            }}
          >
            {getStatusText(appointment.status)}
          </Tag>
          {appointment.dentist && (
            <Tag
              style={{
                fontSize: 10,
                backgroundColor: calmColors.softLilac,
                color: '#5a4a6a',
                border: 'none',
              }}
            >
              {appointment.dentist.name}
            </Tag>
          )}
        </div>
      </div>
    </Card>
  );

  const RoomColumn = ({ room, appointments }) => (
    <Card
      title={
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ color: '#5a6a7a' }}>{room}</span>
          <Tag
            style={{
              backgroundColor: calmColors.chartGreen,
              color: '#4a5a4a',
              border: 'none',
            }}
          >
            {appointments.length}
          </Tag>
        </div>
      }
      style={{
        height: '100%',
        minHeight: 400,
        backgroundColor: '#fafbfc',
        borderTop: `3px solid ${calmColors.chartGreen}`,
      }}
      headStyle={{
        backgroundColor: '#f5f7f9',
        borderBottom: '1px solid #e8ecf0',
      }}
      bodyStyle={{
        padding: 12,
        maxHeight: 500,
        overflowY: 'auto',
      }}
    >
      {appointments.length === 0 ? (
        <Empty description="Sem agendamentos" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        appointments.map(apt => (
          <AppointmentCard key={apt.id} appointment={apt} />
        ))
      )}
    </Card>
  );

  return (
    <div style={{ padding: '0 0 24px 0' }}>
      <Card style={{ marginBottom: 16 }}>
        <Title level={4} style={{ marginBottom: 16, color: '#4a5a6a' }}>
          <MedicineBoxOutlined style={{ marginRight: 8, color: calmColors.accentGreen }} />
          Painel de Atendimento
        </Title>

        {/* Filtros */}
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={8} md={6}>
            <DatePicker
              style={{ width: '100%' }}
              format="DD/MM/YYYY"
              value={filters.date}
              onChange={(date) => handleFilterChange('date', date)}
              placeholder="Data"
              allowClear={false}
            />
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Profissional"
              allowClear
              style={{ width: '100%' }}
              value={filters.dentist_id}
              onChange={(value) => handleFilterChange('dentist_id', value)}
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().includes(input.toLowerCase())
              }
            >
              {dentists.map(d => (
                <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Procedimento"
              allowClear
              style={{ width: '100%' }}
              value={filters.procedure}
              onChange={(value) => handleFilterChange('procedure', value)}
            >
              {procedureOptions.map(p => (
                <Select.Option key={p.value} value={p.value}>{p.label}</Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={24} md={6}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'flex-end',
              height: '100%',
              gap: 8
            }}>
              <Tag
                style={{
                  margin: 0,
                  fontSize: 13,
                  backgroundColor: calmColors.chartGreen,
                  color: '#4a5a4a',
                  border: 'none',
                }}
              >
                Total: {appointments.length} agendamentos
              </Tag>
            </div>
          </Col>
        </Row>

        {/* Gráficos */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} md={12}>
            <Card
              title={<span style={{ color: '#5a6a7a' }}>Pacientes por Profissional</span>}
              size="small"
              style={{
                height: 300,
                borderTop: `3px solid ${calmColors.chartGreen}`,
              }}
              headStyle={{
                backgroundColor: '#f5f7f9',
                borderBottom: '1px solid #e8ecf0',
              }}
              bodyStyle={{ height: 230, padding: '12px' }}
            >
              {getChartDataByDentist().length > 0 ? (
                <Column {...columnConfig} height={200} />
              ) : (
                <Empty description="Sem dados" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card
              title={<span style={{ color: '#5a6a7a' }}>Procedimentos do Dia</span>}
              size="small"
              style={{
                height: 300,
                borderTop: `3px solid ${calmColors.lavender}`,
              }}
              headStyle={{
                backgroundColor: '#f5f7f9',
                borderBottom: '1px solid #e8ecf0',
              }}
              bodyStyle={{ height: 230, padding: '12px' }}
            >
              {getChartDataByProcedure().length > 0 ? (
                <Pie {...pieConfig} height={200} />
              ) : (
                <Empty description="Sem dados" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
        </Row>
      </Card>

      {/* Colunas por Sala */}
      <Spin spinning={loading}>
        <Row gutter={[16, 16]}>
          {rooms.map(room => (
            <Col xs={24} sm={12} md={8} lg={4} xl={4} key={room}>
              <RoomColumn room={room} appointments={getAppointmentsByRoom(room)} />
            </Col>
          ))}
          <Col xs={24} sm={12} md={8} lg={4} xl={4}>
            <RoomColumn room="Sem Sala" appointments={getAppointmentsWithoutRoom()} />
          </Col>
        </Row>
      </Spin>
    </div>
  );
};

export default Attendance;
