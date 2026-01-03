import React, { useState, useEffect } from 'react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import {
  Card,
  Form,
  Select,
  DatePicker,
  Button,
  Typography,
  Space,
  Tag,
  Empty,
  Spin,
  message,
  Result,
  Row,
  Col,
  Radio,
  Alert,
} from 'antd';
import {
  CalendarOutlined,
  ClockCircleOutlined,
  UserOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';

const { Title, Text } = Typography;

const procedureOptions = [
  { value: 'consultation', label: 'Consulta' },
  { value: 'cleaning', label: 'Limpeza' },
  { value: 'filling', label: 'Restauracao' },
  { value: 'extraction', label: 'Extracao' },
  { value: 'root_canal', label: 'Canal' },
  { value: 'orthodontics', label: 'Ortodontia' },
  { value: 'whitening', label: 'Clareamento' },
  { value: 'prosthesis', label: 'Protese' },
  { value: 'implant', label: 'Implante' },
  { value: 'emergency', label: 'Emergencia' },
  { value: 'other', label: 'Outro' },
];

const PatientBookAppointment = () => {
  const navigate = useNavigate();
  const { clinicInfo } = useOutletContext();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [checkingPending, setCheckingPending] = useState(true);
  const [hasPending, setHasPending] = useState(false);
  const [availableSlots, setAvailableSlots] = useState([]);
  const [loadingSlots, setLoadingSlots] = useState(false);
  const [selectedSlot, setSelectedSlot] = useState(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    checkPendingAppointment();
  }, []);

  const checkPendingAppointment = async () => {
    try {
      const response = await patientPortalAPI.getAppointments('upcoming');
      const appointments = response.data.appointments || [];
      setHasPending(appointments.length > 0);
    } catch (error) {
      console.error('Error checking pending appointments:', error);
    } finally {
      setCheckingPending(false);
    }
  };

  const handleDentistOrDateChange = async () => {
    const dentistId = form.getFieldValue('dentist_id');
    const date = form.getFieldValue('date');

    if (dentistId && date) {
      setLoadingSlots(true);
      setSelectedSlot(null);
      try {
        const response = await patientPortalAPI.getAvailableSlots(
          dentistId,
          date.format('YYYY-MM-DD')
        );
        setAvailableSlots(response.data.available_slots || []);
      } catch (error) {
        message.error('Erro ao buscar horarios disponiveis');
        setAvailableSlots([]);
      } finally {
        setLoadingSlots(false);
      }
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      if (!selectedSlot) {
        message.error('Por favor, selecione um horario');
        return;
      }

      setLoading(true);

      const date = values.date.format('YYYY-MM-DD');
      const [startHour, startMin] = selectedSlot.start_time.split(':');
      const [endHour, endMin] = selectedSlot.end_time.split(':');

      const startTime = dayjs(date).hour(parseInt(startHour)).minute(parseInt(startMin)).second(0);
      const endTime = dayjs(date).hour(parseInt(endHour)).minute(parseInt(endMin)).second(0);

      await patientPortalAPI.createAppointment({
        dentist_id: values.dentist_id,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        procedure: values.procedure,
        notes: values.notes,
      });

      setSuccess(true);
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao agendar consulta');
    } finally {
      setLoading(false);
    }
  };

  if (checkingPending) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (hasPending) {
    return (
      <Result
        status="warning"
        title="Voce ja possui uma consulta agendada"
        subTitle="Para agendar uma nova consulta, aguarde a conclusao da consulta atual ou cancele-a primeiro."
        extra={[
          <Button type="primary" key="appointments" onClick={() => navigate('/patient/appointments')}>
            Ver Minhas Consultas
          </Button>,
          <Button key="home" onClick={() => navigate('/patient')}>
            Voltar ao Inicio
          </Button>,
        ]}
      />
    );
  }

  if (success) {
    return (
      <Result
        status="success"
        title="Consulta agendada com sucesso!"
        subTitle="Voce recebera uma confirmacao em breve. Lembre-se de comparecer no horario marcado."
        extra={[
          <Button type="primary" key="appointments" onClick={() => navigate('/patient/appointments')}>
            Ver Minhas Consultas
          </Button>,
          <Button key="home" onClick={() => navigate('/patient')}>
            Voltar ao Inicio
          </Button>,
        ]}
      />
    );
  }

  return (
    <div>
      <Title level={4}>
        <CalendarOutlined /> Agendar Nova Consulta
      </Title>

      <Card>
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            procedure: 'consultation',
          }}
        >
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <Form.Item
                name="dentist_id"
                label="Profissional"
                rules={[{ required: true, message: 'Selecione o profissional' }]}
              >
                <Select
                  placeholder="Selecione o profissional"
                  onChange={handleDentistOrDateChange}
                >
                  {clinicInfo?.dentists?.map((dentist) => (
                    <Select.Option key={dentist.id} value={dentist.id}>
                      {dentist.name}
                      {dentist.specialty && ` - ${dentist.specialty}`}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="procedure"
                label="Tipo de Procedimento"
                rules={[{ required: true, message: 'Selecione o procedimento' }]}
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

          <Form.Item
            name="date"
            label="Data da Consulta"
            rules={[{ required: true, message: 'Selecione a data' }]}
          >
            <DatePicker
              style={{ width: '100%' }}
              format="DD/MM/YYYY"
              placeholder="Selecione a data"
              disabledDate={(current) => current && current < dayjs().startOf('day')}
              onChange={handleDentistOrDateChange}
            />
          </Form.Item>

          {/* Available Slots */}
          <Form.Item label="Horarios Disponiveis">
            {loadingSlots ? (
              <div style={{ textAlign: 'center', padding: 20 }}>
                <Spin />
                <Text type="secondary" style={{ marginLeft: 8 }}>
                  Buscando horarios...
                </Text>
              </div>
            ) : availableSlots.length > 0 ? (
              <Radio.Group
                value={selectedSlot?.start_time}
                onChange={(e) => {
                  const slot = availableSlots.find(s => s.start_time === e.target.value);
                  setSelectedSlot(slot);
                }}
                style={{ width: '100%' }}
              >
                <Row gutter={[8, 8]}>
                  {availableSlots.map((slot) => (
                    <Col key={slot.start_time} xs={12} sm={8} md={6} lg={4}>
                      <Radio.Button
                        value={slot.start_time}
                        style={{
                          width: '100%',
                          textAlign: 'center',
                          height: 40,
                          lineHeight: '40px',
                        }}
                      >
                        <ClockCircleOutlined /> {slot.start_time}
                      </Radio.Button>
                    </Col>
                  ))}
                </Row>
              </Radio.Group>
            ) : form.getFieldValue('dentist_id') && form.getFieldValue('date') ? (
              <Alert
                type="info"
                message="Nenhum horario disponivel para esta data"
                description="Por favor, selecione outra data ou profissional."
              />
            ) : (
              <Alert
                type="info"
                message="Selecione um profissional e uma data para ver os horarios disponiveis"
              />
            )}
          </Form.Item>

          {selectedSlot && (
            <Alert
              type="success"
              style={{ marginBottom: 24 }}
              message={
                <Space>
                  <CheckCircleOutlined />
                  <span>
                    Horario selecionado: <strong>{selectedSlot.start_time}</strong> - <strong>{selectedSlot.end_time}</strong>
                  </span>
                </Space>
              }
            />
          )}

          <Form.Item>
            <Space>
              <Button
                type="primary"
                onClick={handleSubmit}
                loading={loading}
                disabled={!selectedSlot}
                size="large"
              >
                Confirmar Agendamento
              </Button>
              <Button onClick={() => navigate('/patient')} size="large">
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default PatientBookAppointment;
