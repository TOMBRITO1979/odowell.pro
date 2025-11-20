import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
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
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { appointmentsAPI, patientsAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

const { TextArea } = Input;

const AppointmentForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const isEditing = !!id;

  const statusOptions = [
    { value: 'scheduled', label: 'Agendado' },
    { value: 'confirmed', label: 'Confirmado' },
    { value: 'in_progress', label: 'Em Atendimento' },
    { value: 'completed', label: 'Concluído' },
    { value: 'cancelled', label: 'Cancelado' },
    { value: 'no_show', label: 'Faltou' },
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
    if (isEditing) {
      fetchAppointment();
    }
  }, [id]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchAppointment = async () => {
    setLoading(true);
    try {
      const response = await appointmentsAPI.getOne(id);
      const appointment = response.data.appointment;

      form.setFieldsValue({
        ...appointment,
        date: appointment.start_time ? dayjs(appointment.start_time) : null,
        start_time: appointment.start_time ? dayjs(appointment.start_time) : null,
        end_time: appointment.end_time ? dayjs(appointment.end_time) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar agendamento');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      // Combinar data com horários
      const date = values.date;
      const startTime = values.start_time;
      const endTime = values.end_time;

      const start_time = dayjs(date)
        .hour(startTime.hour())
        .minute(startTime.minute())
        .second(0)
        .toISOString();

      const end_time = dayjs(date)
        .hour(endTime.hour())
        .minute(endTime.minute())
        .second(0)
        .toISOString();

      const data = {
        patient_id: values.patient_id,
        dentist_id: user.id,
        start_time,
        end_time,
        status: values.status || 'scheduled',
        procedure: values.procedure,
        notes: values.notes,
      };

      if (isEditing) {
        await appointmentsAPI.update(id, data);
        message.success('Agendamento atualizado com sucesso!');
      } else {
        await appointmentsAPI.create(data);
        message.success('Agendamento criado com sucesso!');
      }

      navigate('/appointments');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar agendamento'
      );
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
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
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={8}>
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

            <Col xs={24} md={8}>
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

            <Col xs={24} md={8}>
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

          <Row gutter={16}>
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
