import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  message,
  Select,
  Row,
  Col,
  Space,
  DatePicker,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  NotificationOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { campaignsAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

// Configurar timezone para Brasil
dayjs.extend(utc);
dayjs.extend(timezone);
const BRAZIL_TZ = 'America/Sao_Paulo';

const { TextArea } = Input;

const CampaignForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);

  const typeOptions = [
    { value: 'whatsapp', label: 'WhatsApp' },
    { value: 'email', label: 'Email' },
    { value: 'sms', label: 'SMS' },
  ];

  const segmentOptions = [
    { value: 'all', label: 'Todos os Pacientes' },
    { value: 'tags', label: 'Por Tags' },
    { value: 'custom', label: 'Personalizado' },
  ];

  useEffect(() => {
    if (id) {
      fetchCampaign();
    }
  }, [id]);

  const fetchCampaign = async () => {
    setLoading(true);
    try {
      const response = await campaignsAPI.getOne(id);
      const campaign = response.data.campaign;

      form.setFieldsValue({
        ...campaign,
        // A data já vem no horário de Brasília do banco
        scheduled_at: campaign.scheduled_at ? dayjs(campaign.scheduled_at) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar campanha');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        created_by_id: user.id,
      };

      // Enviar data no horário de Brasília com timezone explícito
      if (values.scheduled_at) {
        // Formatar como ISO 8601 com timezone de Brasília (-03:00)
        data.scheduled_at = values.scheduled_at.format('YYYY-MM-DDTHH:mm:ss') + '-03:00';
        data.status = 'scheduled';
      } else {
        data.status = 'draft';
      }

      if (id) {
        await campaignsAPI.update(id, data);
        message.success('Campanha atualizada com sucesso!');
      } else {
        await campaignsAPI.create(data);
        message.success('Campanha criada com sucesso!');
      }
      navigate('/campaigns');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar campanha'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <NotificationOutlined />
            <span>{id ? 'Editar Campanha' : 'Nova Campanha'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/campaigns')}
          >
            Voltar
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Row gutter={16}>
            <Col xs={24} md={16}>
              <Form.Item
                name="name"
                label="Nome da Campanha"
                rules={[
                  { required: true, message: 'Informe o nome da campanha' },
                  { max: 200, message: 'Nome muito longo' },
                ]}
              >
                <Input placeholder="Ex: Campanha de Retorno" />
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="type"
                label="Tipo"
                rules={[
                  { required: true, message: 'Selecione o tipo' },
                ]}
              >
                <Select placeholder="Selecione">
                  {typeOptions.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      {type.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="segment_type"
                label="Segmentação"
                rules={[
                  { required: true, message: 'Selecione a segmentação' },
                ]}
                initialValue="all"
              >
                <Select>
                  {segmentOptions.map((seg) => (
                    <Select.Option key={seg.value} value={seg.value}>
                      {seg.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="scheduled_at"
                label="Agendar Para (Horário de Brasília)"
              >
                <DatePicker
                  showTime
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY HH:mm"
                  placeholder="Deixe vazio para enviar imediatamente"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="subject"
            label="Assunto (apenas para Email)"
            rules={[
              { max: 200, message: 'Assunto muito longo' },
            ]}
          >
            <Input placeholder="Assunto do email" />
          </Form.Item>

          <Form.Item
            name="message"
            label="Mensagem"
            rules={[
              { required: true, message: 'Informe a mensagem' },
              { max: 5000, message: 'Mensagem muito longa' },
            ]}
          >
            <TextArea
              rows={8}
              placeholder="Digite a mensagem da campanha..."
              showCount
              maxLength={5000}
            />
          </Form.Item>

          <Form.Item name="tags" label="Tags (separadas por vírgula)">
            <Input placeholder="Ex: ortodontia, clareamento" />
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
              <Button onClick={() => navigate('/campaigns')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default CampaignForm;
