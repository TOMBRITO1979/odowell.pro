import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Row, Col, Tabs, TimePicker, Switch } from 'antd';
import { SettingOutlined, ShopOutlined, ClockCircleOutlined, DollarOutlined } from '@ant-design/icons';
import { settingsAPI } from '../services/api';
import dayjs from 'dayjs';

const Settings = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetchingSettings, setFetchingSettings] = useState(true);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    setFetchingSettings(true);
    try {
      const response = await settingsAPI.get();
      const settings = response.data.settings;

      // Parse working hours if they exist
      const formValues = {
        clinic_name: settings.clinic_name || '',
        clinic_cnpj: settings.clinic_cnpj || '',
        clinic_address: settings.clinic_address || '',
        clinic_city: settings.clinic_city || '',
        clinic_state: settings.clinic_state || '',
        clinic_zip: settings.clinic_zip || '',
        clinic_phone: settings.clinic_phone || '',
        clinic_email: settings.clinic_email || '',
        working_hours_start: settings.working_hours_start ? dayjs(settings.working_hours_start, 'HH:mm') : null,
        working_hours_end: settings.working_hours_end ? dayjs(settings.working_hours_end, 'HH:mm') : null,
        default_appointment_duration: settings.default_appointment_duration || 30,
        payment_cash_enabled: settings.payment_cash_enabled ?? true,
        payment_credit_card_enabled: settings.payment_credit_card_enabled ?? true,
        payment_debit_card_enabled: settings.payment_debit_card_enabled ?? true,
        payment_pix_enabled: settings.payment_pix_enabled ?? true,
        payment_transfer_enabled: settings.payment_transfer_enabled ?? false,
        payment_insurance_enabled: settings.payment_insurance_enabled ?? false,
      };

      form.setFieldsValue(formValues);
    } catch (error) {
      // Settings may not exist yet, that's ok
      console.log('Settings not found, using defaults');
    } finally {
      setFetchingSettings(false);
    }
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      // Format time values
      const settingsData = {
        ...values,
        working_hours_start: values.working_hours_start?.format('HH:mm'),
        working_hours_end: values.working_hours_end?.format('HH:mm'),
      };

      await settingsAPI.update(settingsData);
      message.success('Configurações salvas com sucesso');
    } catch (error) {
      message.error('Erro ao salvar configurações');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const clinicInfoTab = (
    <Row gutter={16}>
      <Col xs={24} md={12}>
        <Form.Item
          label="Nome da Clínica"
          name="clinic_name"
          rules={[{ required: true, message: 'Por favor, insira o nome da clínica' }]}
        >
          <Input placeholder="Nome da sua clínica" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="CNPJ/CPF"
          name="clinic_cnpj"
        >
          <Input placeholder="00.000.000/0000-00" />
        </Form.Item>
      </Col>

      <Col xs={24}>
        <Form.Item
          label="Endereço"
          name="clinic_address"
        >
          <Input placeholder="Rua, número" />
        </Form.Item>
      </Col>

      <Col xs={24} md={8}>
        <Form.Item
          label="Cidade"
          name="clinic_city"
        >
          <Input placeholder="Cidade" />
        </Form.Item>
      </Col>

      <Col xs={24} md={4}>
        <Form.Item
          label="Estado"
          name="clinic_state"
        >
          <Input placeholder="UF" maxLength={2} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="CEP"
          name="clinic_zip"
        >
          <Input placeholder="00000-000" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Telefone"
          name="clinic_phone"
        >
          <Input placeholder="(00) 0000-0000" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="E-mail"
          name="clinic_email"
          rules={[{ type: 'email', message: 'E-mail inválido' }]}
        >
          <Input placeholder="contato@clinica.com" />
        </Form.Item>
      </Col>
    </Row>
  );

  const scheduleTab = (
    <Row gutter={16}>
      <Col xs={24} md={12}>
        <Form.Item
          label="Horário de Abertura"
          name="working_hours_start"
        >
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Horário de Fechamento"
          name="working_hours_end"
        >
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Duração Padrão da Consulta (minutos)"
          name="default_appointment_duration"
        >
          <Input type="number" min={15} step={15} placeholder="30" />
        </Form.Item>
      </Col>
    </Row>
  );

  const paymentTab = (
    <Row gutter={16}>
      <Col xs={24}>
        <p style={{ marginBottom: 16, color: '#666' }}>
          Selecione as formas de pagamento aceitas pela clínica:
        </p>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Dinheiro"
          name="payment_cash_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Cartão de Crédito"
          name="payment_credit_card_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Cartão de Débito"
          name="payment_debit_card_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="PIX"
          name="payment_pix_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Transferência Bancária"
          name="payment_transfer_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Convênio"
          name="payment_insurance_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>
    </Row>
  );

  const tabItems = [
    {
      key: 'clinic',
      label: (
        <span>
          <ShopOutlined />
          Dados da Clínica
        </span>
      ),
      children: clinicInfoTab,
    },
    {
      key: 'schedule',
      label: (
        <span>
          <ClockCircleOutlined />
          Agenda
        </span>
      ),
      children: scheduleTab,
    },
    {
      key: 'payment',
      label: (
        <span>
          <DollarOutlined />
          Pagamentos
        </span>
      ),
      children: paymentTab,
    },
  ];

  return (
    <div>
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SettingOutlined />
            <span>Configurações da Clínica</span>
          </div>
        }
        loading={fetchingSettings}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
        >
          <Tabs items={tabItems} />

          <div style={{ marginTop: 24, paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} size="large">
                Salvar Configurações
              </Button>
            </Form.Item>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
