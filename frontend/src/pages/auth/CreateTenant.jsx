import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Result } from 'antd';
import { ShopOutlined, UserOutlined, LockOutlined, MailOutlined, PhoneOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Title } = Typography;

const CreateTenant = () => {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [verificationEmail, setVerificationEmail] = useState('');
  const navigate = useNavigate();
  const { createTenant } = useAuth();
  const [form] = Form.useForm();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const response = await createTenant(values);
      setVerificationEmail(response.verification_email);
      setSuccess(true);
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar consultório');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)',
        padding: 24
      }}>
        <Card style={{ width: 600, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
          <Result
            icon={<CheckCircleOutlined style={{ color: '#4CAF50' }} />}
            status="success"
            title="Consultório criado com sucesso!"
            subTitle={
              <div>
                <p>Enviamos um email de verificação para:</p>
                <p style={{ fontWeight: 'bold', color: '#4CAF50', fontSize: 16 }}>{verificationEmail}</p>
                <p style={{ marginTop: 16 }}>
                  Por favor, verifique sua caixa de entrada e clique no link de verificação
                  para ativar sua conta.
                </p>
                <p style={{ marginTop: 12, fontSize: 13, color: '#666', background: '#fffbe6', padding: 12, borderRadius: 6 }}>
                  <strong>Dica:</strong> Verifique também sua <strong>caixa de spam</strong>.
                  Se o email estiver lá, marque como "Não é spam" para receber nossos emails normalmente.
                </p>
              </div>
            }
            extra={[
              <Button
                type="primary"
                key="login"
                onClick={() => navigate('/login')}
                size="large"
              >
                Ir para Login
              </Button>,
              <Button
                key="resend"
                onClick={() => navigate('/resend-verification')}
                size="large"
              >
                Não recebeu o email?
              </Button>
            ]}
          />
        </Card>
      </div>
    );
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)',
      padding: 24
    }}>
      <Card style={{ width: 600, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <div style={{
            width: 80,
            height: 80,
            margin: '0 auto 16px',
            background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
            borderRadius: 16,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(102, 187, 106, 0.3)',
          }}>
            <span style={{ fontSize: 48, filter: 'brightness(0) invert(1)' }}>🦷</span>
          </div>
          <Title level={2} style={{ color: '#4CAF50', marginTop: 0 }}>Cadastrar Consultório</Title>
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
