import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  message,
  Spin,
  Space,
  Alert
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { leadsAPI } from '../../services/api';
import { WhatsAppOutlined } from '@ant-design/icons';

const { Option } = Select;
const { TextArea } = Input;

const LeadForm = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();

  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(false);
  const [checkingPhone, setCheckingPhone] = useState(false);
  const [phoneCheck, setPhoneCheck] = useState(null);

  const isEdit = Boolean(id);

  useEffect(() => {
    if (isEdit) {
      fetchLead();
    }
  }, [id]);

  const fetchLead = async () => {
    setLoadingData(true);
    try {
      const response = await leadsAPI.getOne(id);
      const lead = response.data;

      form.setFieldsValue({
        name: lead.name,
        phone: lead.phone,
        email: lead.email,
        source: lead.source,
        contact_reason: lead.contact_reason,
        status: lead.status,
        notes: lead.notes
      });
    } catch (error) {
      message.error('Erro ao carregar lead');
    } finally {
      setLoadingData(false);
    }
  };

  const checkPhoneExists = async (phone) => {
    if (!phone || phone.length < 10) {
      setPhoneCheck(null);
      return;
    }

    setCheckingPhone(true);
    try {
      const response = await leadsAPI.checkByPhone(phone);
      setPhoneCheck(response.data);
    } catch (error) {
      setPhoneCheck(null);
    } finally {
      setCheckingPhone(false);
    }
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      if (isEdit) {
        await leadsAPI.update(id, values);
        message.success('Lead atualizado com sucesso');
      } else {
        await leadsAPI.create(values);
        message.success('Lead cadastrado com sucesso');
      }
      navigate('/leads');
    } catch (error) {
      message.error(isEdit ? 'Erro ao atualizar lead' : 'Erro ao cadastrar lead');
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: 800, margin: '0 auto' }}>
      <Card title={isEdit ? 'Editar Lead' : 'Novo Lead'}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            source: 'whatsapp',
            status: 'new'
          }}
        >
          <Form.Item
            name="name"
            label="Nome"
            rules={[{ required: true, message: 'Informe o nome' }]}
          >
            <Input placeholder="Nome completo do lead" />
          </Form.Item>

          <Form.Item
            name="phone"
            label="Telefone"
            rules={[{ required: true, message: 'Informe o telefone' }]}
            extra={checkingPhone ? 'Verificando...' : null}
          >
            <Input
              placeholder="(00) 00000-0000"
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, '');
                if (value.length >= 10) {
                  checkPhoneExists(value);
                }
              }}
            />
          </Form.Item>

          {phoneCheck && phoneCheck.exists && (
            <Alert
              type={phoneCheck.type === 'patient' ? 'success' : 'warning'}
              message={
                phoneCheck.type === 'patient'
                  ? `Este telefone pertence ao paciente: ${phoneCheck.name}`
                  : `Este telefone já está cadastrado como lead: ${phoneCheck.name} (Status: ${phoneCheck.status})`
              }
              style={{ marginBottom: 16 }}
              showIcon
            />
          )}

          <Form.Item
            name="email"
            label="Email"
            rules={[{ type: 'email', message: 'Email inválido' }]}
          >
            <Input placeholder="email@exemplo.com" />
          </Form.Item>

          <Form.Item
            name="source"
            label="Fonte"
            rules={[{ required: true, message: 'Selecione a fonte' }]}
          >
            <Select>
              <Option value="whatsapp">
                <WhatsAppOutlined style={{ color: '#25D366' }} /> WhatsApp
              </Option>
              <Option value="website">Website</Option>
              <Option value="referral">Indicação</Option>
              <Option value="instagram">Instagram</Option>
              <Option value="facebook">Facebook</Option>
              <Option value="other">Outro</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="contact_reason"
            label="Motivo do Contato"
            help="Resumo da conversa ou do motivo pelo qual o lead entrou em contato"
          >
            <TextArea
              rows={3}
              placeholder="Ex: Paciente interessado em clareamento dental, quer saber valores..."
            />
          </Form.Item>

          {isEdit && (
            <Form.Item
              name="status"
              label="Status"
            >
              <Select>
                <Option value="new">Novo</Option>
                <Option value="contacted">Contatado</Option>
                <Option value="qualified">Qualificado</Option>
                <Option value="converted">Convertido</Option>
                <Option value="lost">Perdido</Option>
              </Select>
            </Form.Item>
          )}

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea
              rows={4}
              placeholder="Informações adicionais sobre o lead..."
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {isEdit ? 'Atualizar' : 'Cadastrar'}
              </Button>
              <Button onClick={() => navigate('/leads')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default LeadForm;
