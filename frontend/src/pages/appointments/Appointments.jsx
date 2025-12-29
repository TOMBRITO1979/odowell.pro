import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Table, Button, message, Tag, Space, Popconfirm, Card, Row, Col, Select, DatePicker, Segmented, Tooltip } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
  ClearOutlined,
  UnorderedListOutlined,
  CalendarOutlined,
  LeftOutlined,
  RightOutlined
} from '@ant-design/icons';
import { appointmentsAPI, usersAPI, settingsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, spacing, shadows } from '../../theme/designSystem';
import { getHolidayInfo } from '../../utils/brazilianHolidays';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import 'dayjs/locale/pt-br';

dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.locale('pt-br');

// Função para formatar data/hora no timezone de São Paulo
const formatDateTime = (dateTime, format = 'DD/MM/YYYY HH:mm') => {
  if (!dateTime) return 'N/A';
  return dayjs(dateTime).tz('America/Sao_Paulo').format(format);
};

const { RangePicker } = DatePicker;

const Appointments = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dentists, setDentists] = useState([]);
  const [viewMode, setViewMode] = useState('list'); // 'list' or 'calendar'
  const [currentWeekStart, setCurrentWeekStart] = useState(dayjs().startOf('week'));
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [lunchBreak, setLunchBreak] = useState({
    enabled: false,
    start: null,
    end: null,
  });
  const [filters, setFilters] = useState({
    dentist_id: null,
    procedure: null,
    status: null,
    dateRange: null,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const navigate = useNavigate();
  const location = useLocation();
  const { canCreate, canEdit, canDelete } = usePermission();

  // Detectar mudança de tamanho da tela
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth <= 768);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    let mounted = true;

    const fetchDentists = async () => {
      try {
        const response = await usersAPI.getAll();
        if (mounted) {
          const professionals = (response.data.users || []).filter(
            user => user.role === 'dentist' || user.role === 'admin'
          );
          setDentists(professionals);
        }
      } catch (error) {
      }
    };

    const fetchSettings = async () => {
      try {
        const response = await settingsAPI.get();
        if (mounted && response.data?.settings) {
          const settings = response.data.settings;
          if (settings.lunch_break_enabled && settings.lunch_break_start && settings.lunch_break_end) {
            setLunchBreak({
              enabled: true,
              start: parseInt(settings.lunch_break_start.split(':')[0]),
              end: parseInt(settings.lunch_break_end.split(':')[0]),
            });
          }
        }
      } catch (error) {
        // Ignora erro se não conseguir buscar configurações
      }
    };

    fetchDentists();
    fetchSettings();

    return () => {
      mounted = false;
    };
  }, []);

  // Função para buscar agendamentos - reutilizável
  const fetchAppointments = React.useCallback(() => {
    setLoading(true);
    const params = {
      page: viewMode === 'calendar' ? 1 : pagination.current,
      page_size: viewMode === 'calendar' ? 500 : pagination.pageSize
    };

    if (filters.dentist_id) params.dentist_id = filters.dentist_id;
    if (filters.procedure) params.procedure = filters.procedure;
    if (filters.status) params.status = filters.status;

    if (viewMode === 'calendar') {
      // Para o calendário, buscar a semana inteira
      params.start_date = currentWeekStart.startOf('day').toISOString();
      params.end_date = currentWeekStart.add(6, 'day').endOf('day').toISOString();
    } else {
      // Para lista, usar filtros normais
      if (filters.dateRange && filters.dateRange[0]) {
        params.start_date = filters.dateRange[0].startOf('day').toISOString();
      } else {
        params.start_date = dayjs().startOf('day').toISOString();
      }
      if (filters.dateRange && filters.dateRange[1]) {
        params.end_date = filters.dateRange[1].endOf('day').toISOString();
      }
    }

    appointmentsAPI.getAll(params)
      .then(res => {
        setData(res.data.appointments || []);
        setPagination(prev => ({
          ...prev,
          total: res.data.total || 0,
        }));
      })
      .catch(() => {
        message.error('Erro ao carregar agendamentos');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [pagination.current, pagination.pageSize, filters, viewMode, currentWeekStart]);

  useEffect(() => {
    fetchAppointments();
  }, [fetchAppointments, location.key]);

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  const clearFilters = () => {
    setFilters({
      dentist_id: null,
      procedure: null,
      status: null,
      dateRange: null,
    });
    setPagination(prev => ({ ...prev, current: 1 }));
  };

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

  const statusOptions = [
    { value: 'scheduled', label: 'Agendado' },
    { value: 'confirmed', label: 'Confirmado' },
    { value: 'in_progress', label: 'Em Atendimento' },
    { value: 'completed', label: 'Concluído' },
    { value: 'cancelled', label: 'Cancelado' },
    { value: 'no_show', label: 'Faltou' },
  ];

  const handleDelete = async (id) => {
    try {
      await appointmentsAPI.delete(id);
      message.success('Agendamento deletado com sucesso');
      fetchAppointments();
    } catch (error) {
      message.error('Erro ao deletar agendamento');
    }
  };

  const handleExportCSV = async () => {
    try {
      const response = await appointmentsAPI.exportCSV('');
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `agendamentos_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
    }
  };

  const handleExportPDF = async () => {
    try {
      const response = await appointmentsAPI.exportPDF('');
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `agendamentos_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
    }
  };

  const getStatusConfig = (status) => {
    const statusConfig = {
      scheduled: { color: statusColors.pending, text: 'Agendado', bg: '#fff7e6' },
      confirmed: { color: statusColors.approved, text: 'Confirmado', bg: '#e6f7ff' },
      in_progress: { color: statusColors.inProgress, text: 'Em Atendimento', bg: '#e6fffb' },
      completed: { color: statusColors.success, text: 'Concluído', bg: '#f6ffed' },
      cancelled: { color: statusColors.cancelled, text: 'Cancelado', bg: '#fff1f0' },
      no_show: { color: statusColors.error, text: 'Faltou', bg: '#fff2e8' },
    };
    return statusConfig[status] || { color: statusColors.pending, text: status, bg: '#fafafa' };
  };

  const getStatusTag = (status) => {
    const config = getStatusConfig(status);
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

  // Navegação do calendário
  const goToPreviousWeek = () => {
    setCurrentWeekStart(prev => prev.subtract(7, 'day'));
  };

  const goToNextWeek = () => {
    setCurrentWeekStart(prev => prev.add(7, 'day'));
  };

  const goToToday = () => {
    setCurrentWeekStart(dayjs().startOf('week'));
  };

  // Gerar dias da semana
  const getWeekDays = () => {
    const days = [];
    for (let i = 0; i < 7; i++) {
      days.push(currentWeekStart.add(i, 'day'));
    }
    return days;
  };

  // Gerar horários do dia (08:00 às 20:00)
  const getTimeSlots = () => {
    const slots = [];
    for (let hour = 8; hour <= 20; hour++) {
      slots.push(`${hour.toString().padStart(2, '0')}:00`);
    }
    return slots;
  };

  // Agrupar agendamentos por dia e hora
  const getAppointmentsForSlot = (day, timeSlot) => {
    const slotHour = parseInt(timeSlot.split(':')[0]);
    return data.filter(apt => {
      const aptDate = dayjs(apt.start_time);
      return aptDate.format('YYYY-MM-DD') === day.format('YYYY-MM-DD') &&
             aptDate.hour() === slotHour;
    });
  };

  // Renderizar card de agendamento no calendário
  const renderAppointmentCard = (appointment) => {
    const config = getStatusConfig(appointment.status);
    return (
      <Tooltip
        key={appointment.id}
        title={
          <div>
            <div><strong>{appointment.patient?.name}</strong></div>
            <div>{getProcedureText(appointment.procedure)}</div>
            <div>{formatDateTime(appointment.start_time, 'HH:mm')} - {formatDateTime(appointment.end_time, 'HH:mm')}</div>
            <div>Prof: {appointment.dentist?.name}</div>
            <div>Status: {config.text}</div>
          </div>
        }
      >
        <div
          onClick={() => navigate(`/appointments/${appointment.id}`)}
          style={{
            backgroundColor: config.bg,
            borderLeft: `3px solid ${config.color}`,
            padding: '4px 6px',
            marginBottom: '2px',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '11px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          <div style={{ fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {appointment.patient?.name?.split(' ')[0] || 'N/A'}
          </div>
          <div style={{ fontSize: '10px', color: '#666' }}>
            {formatDateTime(appointment.start_time, 'HH:mm')}
          </div>
        </div>
      </Tooltip>
    );
  };

  // Renderizar visualização de calendário
  const renderCalendarView = () => {
    const weekDays = getWeekDays();
    const timeSlots = getTimeSlots();
    const dayNames = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];

    return (
      <div>
        {/* Navegação da semana */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 16,
          padding: '12px 16px',
          background: '#fafafa',
          borderRadius: '8px'
        }}>
          <Button icon={<LeftOutlined />} onClick={goToPreviousWeek}>
            Anterior
          </Button>
          <div style={{ textAlign: 'center' }}>
            <Button type="link" onClick={goToToday} style={{ marginBottom: 4 }}>
              Hoje
            </Button>
            <div style={{ fontWeight: 600, fontSize: 16 }}>
              {currentWeekStart.format('DD MMM')} - {currentWeekStart.add(6, 'day').format('DD MMM YYYY')}
            </div>
          </div>
          <Button icon={<RightOutlined />} onClick={goToNextWeek} iconPosition="end">
            Próxima
          </Button>
        </div>

        {/* Grid do calendário */}
        <div style={{ overflowX: 'auto' }}>
          <div style={{
            display: 'grid',
            gridTemplateColumns: '60px repeat(7, 1fr)',
            minWidth: '800px',
            border: '1px solid #e8e8e8',
            borderRadius: '8px',
            overflow: 'hidden'
          }}>
            {/* Header com dias da semana */}
            <div style={{
              backgroundColor: '#fafafa',
              padding: '8px',
              fontWeight: 600,
              borderBottom: '1px solid #e8e8e8',
              borderRight: '1px solid #e8e8e8'
            }}>
              Hora
            </div>
            {weekDays.map((day, index) => {
              const isToday = day.format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD');
              const holiday = getHolidayInfo(day);
              const isHoliday = holiday !== null;

              return (
                <Tooltip
                  key={index}
                  title={isHoliday ? holiday.name : null}
                  placement="top"
                >
                  <div
                    style={{
                      backgroundColor: isHoliday ? '#fff1f0' : isToday ? '#e6f7ff' : '#fafafa',
                      padding: '8px',
                      textAlign: 'center',
                      fontWeight: 600,
                      borderBottom: '1px solid #e8e8e8',
                      borderRight: index < 6 ? '1px solid #e8e8e8' : 'none',
                      position: 'relative'
                    }}
                  >
                    <div style={{ color: isHoliday ? '#cf1322' : 'inherit' }}>
                      {dayNames[index]}
                    </div>
                    <div style={{
                      fontSize: '18px',
                      color: isHoliday ? '#cf1322' : isToday ? '#1890ff' : 'inherit'
                    }}>
                      {day.format('DD')}
                    </div>
                    {isHoliday && (
                      <div style={{
                        fontSize: '9px',
                        color: '#cf1322',
                        marginTop: '2px',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        maxWidth: '100%'
                      }}>
                        {holiday.name}
                      </div>
                    )}
                  </div>
                </Tooltip>
              );
            })}

            {/* Slots de horário */}
            {timeSlots.map((timeSlot, slotIndex) => {
              const slotHour = parseInt(timeSlot.split(':')[0]);
              const isLunchHour = lunchBreak.enabled && slotHour >= lunchBreak.start && slotHour < lunchBreak.end;

              return (
              <React.Fragment key={timeSlot}>
                <div style={{
                  padding: '4px 8px',
                  backgroundColor: isLunchHour ? '#f5f5f5' : '#fafafa',
                  borderBottom: slotIndex < timeSlots.length - 1 ? '1px solid #e8e8e8' : 'none',
                  borderRight: '1px solid #e8e8e8',
                  fontSize: '12px',
                  fontWeight: 500,
                  display: 'flex',
                  alignItems: 'flex-start',
                  minHeight: '60px',
                  color: isLunchHour ? '#999' : 'inherit'
                }}>
                  <Tooltip title={isLunchHour ? 'Horário de Almoço' : null}>
                    <span>{timeSlot}</span>
                  </Tooltip>
                </div>
                {weekDays.map((day, dayIndex) => {
                  const appointments = getAppointmentsForSlot(day, timeSlot);
                  const isToday = day.format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD');
                  const dayHoliday = getHolidayInfo(day);
                  const isDayHoliday = dayHoliday !== null;

                  // Prioridade de cor: almoço > feriado > hoje > normal
                  let bgColor = '#fff';
                  if (isLunchHour) {
                    bgColor = '#f0f0f0'; // Cinza para almoço
                  } else if (isDayHoliday) {
                    bgColor = '#fff7f6'; // Vermelho claro para feriados
                  } else if (isToday) {
                    bgColor = '#f6ffed'; // Verde claro para hoje
                  }

                  return (
                    <Tooltip
                      key={dayIndex}
                      title={isLunchHour ? 'Horário de Almoço' : null}
                    >
                      <div
                        style={{
                          padding: '4px',
                          borderBottom: slotIndex < timeSlots.length - 1 ? '1px solid #e8e8e8' : 'none',
                          borderRight: dayIndex < 6 ? '1px solid #e8e8e8' : 'none',
                          minHeight: '60px',
                          backgroundColor: bgColor,
                          position: 'relative'
                        }}
                      >
                        {isLunchHour && (
                          <div style={{
                            position: 'absolute',
                            top: '50%',
                            left: '50%',
                            transform: 'translate(-50%, -50%)',
                            color: '#bbb',
                            fontSize: '10px',
                            textAlign: 'center',
                            zIndex: 0,
                            pointerEvents: 'none'
                          }}>
                            Almoço
                          </div>
                        )}
                        <div style={{ position: 'relative', zIndex: 1 }}>
                          {appointments.map(apt => renderAppointmentCard(apt))}
                        </div>
                      </div>
                    </Tooltip>
                  );
                })}
              </React.Fragment>
              );
            })}
          </div>
        </div>
      </div>
    );
  };

  // Renderizar cards para versão mobile
  const renderMobileCards = () => {
    if (loading) {
      return (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <span>Carregando...</span>
        </div>
      );
    }

    if (data.length === 0) {
      return (
        <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
          Nenhum agendamento encontrado
        </div>
      );
    }

    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {data.map((record) => {
          const statusConfig = getStatusConfig(record.status);
          return (
            <Card
              key={record.id}
              size="small"
              style={{
                borderLeft: `4px solid ${statusConfig.color}`,
                backgroundColor: statusConfig.bg,
              }}
              bodyStyle={{ padding: '12px' }}
            >
              {/* Cabeçalho com Paciente e Status */}
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: '8px'
              }}>
                <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>
                  {record.patient?.name || 'N/A'}
                </div>
                {getStatusTag(record.status)}
              </div>

              {/* Informações do agendamento */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '6px',
                fontSize: '13px',
                color: '#555'
              }}>
                <div>
                  <strong>Data/Hora:</strong><br />
                  {formatDateTime(record.start_time)}
                </div>
                <div>
                  <strong>Procedimento:</strong><br />
                  {getProcedureText(record.procedure)}
                </div>
                <div>
                  <strong>Profissional:</strong><br />
                  {record.dentist?.name || 'N/A'}
                </div>
                <div>
                  <strong>Sala:</strong><br />
                  {record.room || '-'}
                </div>
              </div>

              {/* Botões de ação */}
              <div style={{
                display: 'flex',
                justifyContent: 'flex-end',
                gap: '8px',
                marginTop: '12px',
                paddingTop: '8px',
                borderTop: '1px solid rgba(0,0,0,0.06)'
              }}>
                <Button
                  type="text"
                  size="small"
                  icon={<EyeOutlined />}
                  onClick={() => navigate(`/appointments/${record.id}`)}
                  style={{ color: actionColors.view }}
                >
                  Ver
                </Button>
                {canEdit('appointments') && (
                  <Button
                    type="text"
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => navigate(`/appointments/${record.id}/edit`)}
                    style={{ color: actionColors.edit }}
                  >
                    Editar
                  </Button>
                )}
                {canDelete('appointments') && (
                  <Popconfirm
                    title="Tem certeza que deseja deletar este agendamento?"
                    onConfirm={() => handleDelete(record.id)}
                    okText="Sim"
                    cancelText="Não"
                  >
                    <Button
                      type="text"
                      size="small"
                      icon={<DeleteOutlined />}
                      style={{ color: actionColors.delete }}
                    >
                      Excluir
                    </Button>
                  </Popconfirm>
                )}
              </div>
            </Card>
          );
        })}

        {/* Paginação para mobile */}
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: '16px',
          marginTop: '16px',
          padding: '12px',
          background: '#fafafa',
          borderRadius: '8px'
        }}>
          <Button
            disabled={pagination.current === 1}
            onClick={() => setPagination(prev => ({ ...prev, current: prev.current - 1 }))}
          >
            Anterior
          </Button>
          <span style={{ fontSize: '13px' }}>
            Página {pagination.current} de {Math.ceil(pagination.total / pagination.pageSize) || 1}
          </span>
          <Button
            disabled={pagination.current >= Math.ceil(pagination.total / pagination.pageSize)}
            onClick={() => setPagination(prev => ({ ...prev, current: prev.current + 1 }))}
          >
            Próxima
          </Button>
        </div>
      </div>
    );
  };

  const columns = [
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient_name',
      render: (text) => text || 'N/A',
    },
    {
      title: 'Profissional',
      dataIndex: ['dentist', 'name'],
      key: 'dentist_name',
      render: (text) => text || 'N/A',
    },
    {
      title: 'Data/Hora',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (text) => formatDateTime(text),
      sorter: (a, b) => dayjs(a.start_time).unix() - dayjs(b.start_time).unix(),
      defaultSortOrder: 'ascend',
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
      title: 'Sala',
      dataIndex: 'room',
      key: 'room',
      render: (text) => text || '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/appointments/${record.id}`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          {canEdit('appointments') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/appointments/${record.id}/edit`)}
              title="Editar"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('appointments') && (
            <Popconfirm
              title="Tem certeza que deseja deletar este agendamento?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                icon={<DeleteOutlined />}
                title="Deletar"
                style={{ color: actionColors.delete }}
              />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      {isMobile ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginBottom: 16 }}>
          <h1 style={{ margin: 0, fontSize: '18px', textAlign: 'left' }}>Agendamentos</h1>
          <Segmented
            value={viewMode}
            onChange={setViewMode}
            options={[
              { value: 'list', icon: <UnorderedListOutlined />, label: 'Lista' },
              { value: 'calendar', icon: <CalendarOutlined />, label: 'Calendário' },
            ]}
          />
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '8px', width: '100%' }}>
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}
            >
              CSV
            </Button>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleExportPDF}
              style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
            >
              PDF
            </Button>
            {canCreate('appointments') && (
              <Button
                icon={<PlusOutlined />}
                onClick={() => navigate('/appointments/new')}
                style={{ backgroundColor: actionColors.create, borderColor: actionColors.create, color: '#fff' }}
              >
                Novo
              </Button>
            )}
          </div>
        </div>
      ) : (
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16, alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
          <h1 style={{ margin: 0 }}>Agendamentos</h1>
          <Space wrap>
            <Segmented
              value={viewMode}
              onChange={setViewMode}
              options={[
                { value: 'list', icon: <UnorderedListOutlined />, label: 'Lista' },
                { value: 'calendar', icon: <CalendarOutlined />, label: 'Calendário' },
              ]}
            />
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              style={{
                backgroundColor: actionColors.exportExcel,
                borderColor: actionColors.exportExcel,
                color: '#fff'
              }}
            >
              Exportar CSV
            </Button>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleExportPDF}
              style={{
                backgroundColor: actionColors.exportPDF,
                borderColor: actionColors.exportPDF,
                color: '#fff'
              }}
            >
              Gerar PDF
            </Button>
            {canCreate('appointments') && (
              <Button
                icon={<PlusOutlined />}
                onClick={() => navigate('/appointments/new')}
                style={{
                  backgroundColor: actionColors.create,
                  borderColor: actionColors.create,
                  color: '#fff'
                }}
              >
                Novo Agendamento
              </Button>
            )}
          </Space>
        </div>
      )}

      <Card style={{ boxShadow: shadows.small }}>
        {/* Filtros - mostrar apenas no modo lista ou filtros simplificados no calendário */}
        <div style={{ marginBottom: 16, padding: '16px', background: '#fafafa', borderRadius: '8px' }}>
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} sm={12} md={6}>
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
            <Col xs={24} sm={12} md={5}>
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
            <Col xs={24} sm={12} md={5}>
              <Select
                placeholder="Status"
                allowClear
                style={{ width: '100%' }}
                value={filters.status}
                onChange={(value) => handleFilterChange('status', value)}
              >
                {statusOptions.map(s => (
                  <Select.Option key={s.value} value={s.value}>{s.label}</Select.Option>
                ))}
              </Select>
            </Col>
            {viewMode === 'list' && (
              <Col xs={24} sm={12} md={6}>
                <RangePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  value={filters.dateRange}
                  onChange={(dates) => handleFilterChange('dateRange', dates)}
                  placeholder={['Hoje (padrão)', 'Data Fim']}
                />
              </Col>
            )}
            <Col xs={24} sm={24} md={viewMode === 'list' ? 2 : 8}>
              <Button
                icon={<ClearOutlined />}
                onClick={clearFilters}
                style={{ width: viewMode === 'list' ? '100%' : 'auto' }}
                title="Limpar Filtros"
              >
                Limpar
              </Button>
            </Col>
          </Row>
        </div>

        {/* Renderizar visualização baseada no modo */}
        {viewMode === 'list' ? (
          isMobile ? (
            // Versão mobile: cards
            renderMobileCards()
          ) : (
            // Versão desktop: tabela
            <div style={{ overflowX: 'auto' }}>
              <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                loading={loading}
                pagination={{
                  ...pagination,
                  showSizeChanger: true,
                  pageSizeOptions: ['10', '20', '50', '100'],
                  onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
                }}
                scroll={{ x: 'max-content' }}
              />
            </div>
          )
        ) : (
          renderCalendarView()
        )}
      </Card>
    </div>
  );
};

export default Appointments;
