import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  message,
  Space,
  Tag,
  Spin,
  Popconfirm,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  CalendarOutlined,
  WhatsAppOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { appointmentsAPI, whatsappBusinessAPI } from '../../services/api';

import { usePermission } from '../../contexts/AuthContext';

// Configurar plugins de timezone
dayjs.extend(utc);
dayjs.extend(timezone);

// Função para formatar data/hora no timezone de São Paulo
const formatDateTime = (dateTime) => {
  if (!dateTime) return 'N/A';
  return dayjs(dateTime).tz('America/Sao_Paulo').format('DD/MM/YYYY HH:mm');
};

const AppointmentDetails = () => {
  const [appointment, setAppointment] = useState(null);
  const [loading, setLoading] = useState(false);
  const [sendingWhatsApp, setSendingWhatsApp] = useState(false);
  const navigate = useNavigate();
  const { id } = useParams();
  const { canEdit, canDelete } = usePermission();

  useEffect(() => {
    fetchAppointment();
  }, [id]);

  const fetchAppointment = async () => {
    setLoading(true);
    try {
      const response = await appointmentsAPI.getOne(id);
      setAppointment(response.data.appointment);
    } catch (error) {
      message.error('Erro ao carregar agendamento');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      await appointmentsAPI.delete(id);
      message.success('Agendamento deletado com sucesso');
      navigate('/appointments');
    } catch (error) {
      message.error('Erro ao deletar agendamento');
    }
  };

  const handleSendWhatsAppConfirmation = async () => {
    setSendingWhatsApp(true);
    try {
      const response = await whatsappBusinessAPI.sendConfirmation(id);
      if (response.data.success) {
        message.success(`Confirmação enviada para ${response.data.patient_name} via WhatsApp!`);
      }
    } catch (error) {
      const errorMsg = error.response?.data?.error || 'Erro ao enviar confirmação';
      message.error(errorMsg);
    } finally {
      setSendingWhatsApp(false);
    }
  };

  const getStatusTag = (status) => {
    const statusConfig = {
      scheduled: { color: 'blue', text: 'Agendado' },
      confirmed: { color: 'green', text: 'Confirmado' },
      in_progress: { color: 'orange', text: 'Em Atendimento' },
      completed: { color: 'success', text: 'Concluído' },
      cancelled: { color: 'red', text: 'Cancelado' },
      no_show: { color: 'default', text: 'Faltou' },
    };
    const config = statusConfig[status] || { color: 'default', text: status };
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

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!appointment) {
    return null;
  }

  return (
    <div>
      <Card
        title={
          <Space>
            <CalendarOutlined />
            <span>Detalhes do Agendamento</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/appointments')}
            >
              Voltar
            </Button>
            {canEdit('appointments') && appointment.status !== 'cancelled' && appointment.status !== 'completed' && (
              <Button
                type="default"
                icon={<WhatsAppOutlined />}
                onClick={handleSendWhatsAppConfirmation}
                loading={sendingWhatsApp}
                style={{ backgroundColor: '#25D366', borderColor: '#25D366', color: '#fff' }}
              >
                Enviar Confirmação
              </Button>
            )}
            {canEdit('appointments') && (
              <Button
                type="primary"
                icon={<EditOutlined />}
                onClick={() => navigate(`/appointments/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            {canDelete('appointments') && (
              <Popconfirm
                title="Tem certeza que deseja deletar este agendamento?"
                onConfirm={handleDelete}
                okText="Sim"
                cancelText="Não"
              >
                <Button danger icon={<DeleteOutlined />}>
                  Deletar
                </Button>
              </Popconfirm>
            )}
          </Space>
        }
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="ID">
            {appointment.id}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            {getStatusTag(appointment.status)}
          </Descriptions.Item>
          <Descriptions.Item label="Paciente">
            {appointment.patient?.name || 'N/A'}
          </Descriptions.Item>
          <Descriptions.Item label="Procedimento">
            {getProcedureText(appointment.procedure)}
          </Descriptions.Item>
          <Descriptions.Item label="Data/Hora Início">
            {formatDateTime(appointment.start_time)}
          </Descriptions.Item>
          <Descriptions.Item label="Data/Hora Fim">
            {formatDateTime(appointment.end_time)}
          </Descriptions.Item>
          <Descriptions.Item label="Duração">
            {appointment.start_time && appointment.end_time
              ? `${dayjs(appointment.end_time).diff(dayjs(appointment.start_time), 'minute')} minutos`
              : 'N/A'}
          </Descriptions.Item>
          <Descriptions.Item label="Sala">
            {appointment.room || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Criado em">
            {formatDateTime(appointment.created_at)}
          </Descriptions.Item>
          <Descriptions.Item label="Atualizado em">
            {formatDateTime(appointment.updated_at)}
          </Descriptions.Item>
          <Descriptions.Item label="Observações" span={2}>
            {appointment.notes || 'Sem observações'}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
};

export default AppointmentDetails;
