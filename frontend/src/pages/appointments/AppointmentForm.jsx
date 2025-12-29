import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  DatePicker,
  Select,
  message,
  Row,
  Col,
  Space,
  TimePicker,
  Modal,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  CalendarOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { appointmentsAPI, patientsAPI, usersAPI, waitingListAPI, settingsAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

// Configurar plugins de timezone
dayjs.extend(utc);
dayjs.extend(timezone);

const { TextArea } = Input;

const AppointmentForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const [dentists, setDentists] = useState([]);
  const [lunchBreak, setLunchBreak] = useState({
    enabled: false,
    start: null,
    end: null,
  });
  const isEditing = !!id;

  // Get patient_id from URL params (from waiting list)
  const patientIdFromUrl = searchParams.get('patient_id');
  const waitingListId = searchParams.get('waiting_list_id');

  const statusOptions = [
    { value: 'scheduled', label: 'Agendado' },
    { value: 'confirmed', label: 'Confirmado' },
    { value: 'in_progress', label: 'Em Atendimento' },
    { value: 'completed', label: 'Concluído' },
    { value: 'cancelled', label: 'Cancelado' },
    { value: 'no_show', label: 'Faltou' },
  ];

  const roomOptions = [
    { value: 'Sala 1', label: 'Sala 1' },
    { value: 'Sala 2', label: 'Sala 2' },
    { value: 'Sala 3', label: 'Sala 3' },
    { value: 'Sala 4', label: 'Sala 4' },
    { value: 'Sala 5', label: 'Sala 5' },
  ];

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

  useEffect(() => {
    fetchPatients();
    fetchDentists();
    fetchSettings();
    if (isEditing) {
      fetchAppointment();
    }
  }, [id]);

  const fetchSettings = async () => {
    try {
      const response = await settingsAPI.get();
      if (response.data?.settings) {
        const settings = response.data.settings;
        if (settings.lunch_break_enabled && settings.lunch_break_start && settings.lunch_break_end) {
          setLunchBreak({
            enabled: true,
            start: settings.lunch_break_start,
            end: settings.lunch_break_end,
          });
        }
      }
    } catch (error) {
      // Ignora erro se não conseguir buscar configurações
    }
  };

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);

      // Pre-select patient if coming from waiting list
      if (patientIdFromUrl && !isEditing) {
        const patientId = parseInt(patientIdFromUrl, 10);
        form.setFieldValue('patient_id', patientId);
      }
    } catch (error) {
      // Silently fail
    }
  };

  const fetchDentists = async () => {
    try {
      const response = await usersAPI.getAll();
      // Filtrar apenas dentistas e admins (profissionais que atendem)
      const professionals = (response.data.users || []).filter(
        u => u.role === 'dentist' || u.role === 'admin'
      );
      setDentists(professionals);
    } catch (error) {
      // Silently fail
    }
  };

  const fetchAppointment = async () => {
    setLoading(true);
    try {
      const response = await appointmentsAPI.getOne(id);
      const appointment = response.data.appointment;

      form.setFieldsValue({
        ...appointment,
        dentist_id: appointment.dentist_id,
        room: appointment.room,
        date: appointment.start_time ? dayjs(appointment.start_time) : null,
        start_time: appointment.start_time ? dayjs(appointment.start_time) : null,
        end_time: appointment.end_time ? dayjs(appointment.end_time) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar agendamento');
    } finally {
      setLoading(false);
    }
  };

  // Verifica se o horário está dentro do período de almoço
  const isLunchTime = (startTime) => {
    if (!lunchBreak.enabled || !lunchBreak.start || !lunchBreak.end) {
      return false;
    }

    const slotHour = startTime.hour();
    const slotMinute = startTime.minute();
    const slotMinutes = slotHour * 60 + slotMinute;

    const [lunchStartHour, lunchStartMinute] = lunchBreak.start.split(':').map(Number);
    const [lunchEndHour, lunchEndMinute] = lunchBreak.end.split(':').map(Number);
    const lunchStartMinutes = lunchStartHour * 60 + lunchStartMinute;
    const lunchEndMinutes = lunchEndHour * 60 + lunchEndMinute;

    return slotMinutes >= lunchStartMinutes && slotMinutes < lunchEndMinutes;
  };

  const saveAppointment = async (values) => {
    setLoading(true);
    try {
      // Combinar data com horários usando timezone de São Paulo
      const date = values.date;
      const startTime = values.start_time;
      const endTime = values.end_time;

      // Criar datetime com timezone explícito de São Paulo
      const start_time = dayjs.tz(
        `${date.format('YYYY-MM-DD')} ${startTime.format('HH:mm')}:00`,
        'America/Sao_Paulo'
      ).toISOString();

      const end_time = dayjs.tz(
        `${date.format('YYYY-MM-DD')} ${endTime.format('HH:mm')}:00`,
        'America/Sao_Paulo'
      ).toISOString();

      const data = {
        patient_id: values.patient_id,
        dentist_id: values.dentist_id,
        start_time,
        end_time,
        status: values.status || 'scheduled',
        procedure: values.procedure,
        room: values.room,
        notes: values.notes,
      };

      if (isEditing) {
        await appointmentsAPI.update(id, data);
        message.success('Agendamento atualizado com sucesso!');
      } else {
        const response = await appointmentsAPI.create(data);
        message.success('Agendamento criado com sucesso!');

        // If coming from waiting list, update the waiting list entry
        if (waitingListId && response.data?.appointment?.id) {
          try {
            await waitingListAPI.schedule(waitingListId, response.data.appointment.id);
          } catch (err) {
            // Don't fail the whole operation if this fails
          }
        }
      }

      navigate('/appointments');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar agendamento'
      );
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    const startTime = values.start_time;

    // Verificar se está tentando agendar no horário de almoço
    if (isLunchTime(startTime)) {
      Modal.confirm({
        title: 'Horário de Almoço',
        icon: <ExclamationCircleOutlined />,
        content: `Este horário (${startTime.format('HH:mm')}) está dentro do período de almoço (${lunchBreak.start} - ${lunchBreak.end}). Deseja continuar mesmo assim?`,
        okText: 'Sim, continuar',
        cancelText: 'Não, escolher outro horário',
        onOk: () => saveAppointment(values),
      });
      return;
    }

    // Se não for horário de almoço, salvar normalmente
    saveAppointment(values);
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <CalendarOutlined />
            <span>{isEditing ? 'Editar Agendamento' : 'Novo Agendamento'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/appointments')}
          >
            Voltar
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            status: 'scheduled',
            procedure: 'consultation',
          }}
        >
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="patient_id"
                label="Paciente"
                rules={[
                  { required: true, message: 'Selecione o paciente' },
                ]}
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {patients.map((patient) => (
                    <Select.Option key={patient.id} value={patient.id}>
                      {patient.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="dentist_id"
                label="Profissional"
                rules={[
                  { required: true, message: 'Selecione o profissional' },
                ]}
              >
                <Select
                  placeholder="Selecione o profissional"
                  showSearch
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {dentists.map((dentist) => (
                    <Select.Option key={dentist.id} value={dentist.id}>
                      {dentist.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="procedure"
                label="Procedimento"
                rules={[
                  { required: true, message: 'Selecione o procedimento' },
                ]}
              >
                <Select placeholder="Selecione o procedimento">
                  {procedureOptions.map((proc) => (
                    <Select.Option key={proc.value} value={proc.value}>
                      {proc.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="status"
                label="Status"
                rules={[
                  { required: true, message: 'Selecione o status' },
                ]}
              >
                <Select placeholder="Selecione o status">
                  {statusOptions.map((status) => (
                    <Select.Option key={status.value} value={status.value}>
                      {status.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={6}>
              <Form.Item
                name="room"
                label="Sala"
              >
                <Select placeholder="Selecione a sala" allowClear>
                  {roomOptions.map((room) => (
                    <Select.Option key={room.value} value={room.value}>
                      {room.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="date"
                label="Data do Agendamento"
                rules={[
                  { required: true, message: 'Selecione a data' },
                ]}
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="start_time"
                label="Horário de Início"
                rules={[
                  { required: true, message: 'Selecione o horário de início' },
                ]}
              >
                <TimePicker
                  style={{ width: '100%' }}
                  format="HH:mm"
                  minuteStep={15}
                  placeholder="Selecione o horário"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="end_time"
                label="Horário de Término"
                rules={[
                  { required: true, message: 'Selecione o horário de término' },
                ]}
              >
                <TimePicker
                  style={{ width: '100%' }}
                  format="HH:mm"
                  minuteStep={15}
                  placeholder="Selecione o horário"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="notes" label="Observações">
            <TextArea
              rows={4}
              placeholder="Observações sobre o agendamento..."
              maxLength={1000}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                icon={<SaveOutlined />}
              >
                Salvar
              </Button>
              <Button onClick={() => navigate('/appointments')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default AppointmentForm;
