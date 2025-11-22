import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Steps } from 'antd';
import { ShopOutlined, UserOutlined, LockOutlined, MailOutlined, PhoneOutlined, MedicineBoxOutlined } from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Title } = Typography;

const CreateTenant = () => {
  const [loading, setLoading] = useState(false);
  const [current, setCurrent] = useState(0);
  const navigate = useNavigate();
  const { createTenant } = useAuth();
  const [form] = Form.useForm();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      await createTenant(values);
      message.success('Consultório criado com sucesso!');
      navigate('/');
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar consultório');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #16a34a 0%, #15803d 100%)',
      padding: 24
    }}>
      <Card style={{ width: 600, boxShadow: '0 8px 32px rgba(0,0,0,0.1)' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <MedicineBoxOutlined style={{ fontSize: 48, color: '#16a34a', marginBottom: 16 }} />
          <Title level={2} style={{ color: '#16a34a', marginTop: 0 }}>Cadastrar Consultório</Title>
        </div>

        <Form form={form} name="create-tenant" onFinish={onFinish} layout="vertical">
          <Title level={4}>Dados do Consultório</Title>
          <Form.Item
            name="name"
            label="Nome do Consultório"
            rules={[{ required: true }]}
          >
            <Input prefix={<ShopOutlined />} placeholder="Ex: Clínica Dental Sorriso" />
          </Form.Item>

          <Form.Item
            name="subdomain"
            label="Subdomínio"
            rules={[{ required: true }]}
          >
            <Input prefix={<ShopOutlined />} placeholder="Ex: clinicasorriso" />
          </Form.Item>

          <Form.Item name="email" label="Email" rules={[{ type: 'email' }]}>
            <Input prefix={<MailOutlined />} />
          </Form.Item>

          <Form.Item name="phone" label="Telefone">
            <Input prefix={<PhoneOutlined />} />
          </Form.Item>

          <Title level={4} style={{ marginTop: 24 }}>Administrador</Title>
          <Form.Item
            name="admin_name"
            label="Nome"
            rules={[{ required: true }]}
          >
            <Input prefix={<UserOutlined />} />
          </Form.Item>

          <Form.Item
            name="admin_email"
            label="Email"
            rules={[{ required: true, type: 'email' }]}
          >
            <Input prefix={<MailOutlined />} />
          </Form.Item>

          <Form.Item
            name="admin_password"
            label="Senha"
            rules={[{ required: true, min: 6 }]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading} size="large">
              Criar Consultório
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Link to="/login">Já tem conta? Fazer login</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default CreateTenant;
