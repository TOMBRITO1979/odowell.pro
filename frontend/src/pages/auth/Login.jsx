import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Alert } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { authAPI } from '../../services/api';

const { Title } = Typography;

const Login = () => {
  const [loading, setLoading] = useState(false);
  const [resendingEmail, setResendingEmail] = useState(false);
  const [emailNotVerified, setEmailNotVerified] = useState(null);
  const navigate = useNavigate();
  const { login } = useAuth();

  const onFinish = async (values) => {
    setLoading(true);
    setEmailNotVerified(null);
    try {
      await login(values);
      message.success('Login realizado com sucesso!');
      navigate('/');
    } catch (error) {
      const data = error.response?.data;
      if (data?.email_not_verified) {
        setEmailNotVerified(data.tenant_email);
      } else {
        message.error(data?.error || 'Erro ao fazer login');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleResendVerification = async () => {
    if (!emailNotVerified) return;
    setResendingEmail(true);
    try {
      await authAPI.resendVerification(emailNotVerified);
      message.success('Email de verificação reenviado! Verifique sua caixa de entrada.');
    } catch (error) {
      message.error('Erro ao reenviar email de verificação');
    } finally {
      setResendingEmail(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
    }}>
      <Card style={{ width: 400, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
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
          <Title level={2} style={{ color: '#4CAF50', marginTop: 0 }}>OdoWell</Title>
          <Typography.Text type="secondary">Gestão Odontológica</Typography.Text>
        </div>

        {emailNotVerified && (
          <Alert
            message="Email não verificado"
            description={
              <div>
                <p>Por favor, verifique seu email para ativar a conta.</p>
                <p style={{ fontSize: 12, color: '#666', marginTop: 8 }}>
                  <strong>Dica:</strong> Verifique também sua <strong>caixa de spam</strong>.
                  Se o email estiver lá, marque como "Não é spam" para receber nossos emails normalmente.
                </p>
                <Button
                  type="link"
                  icon={<MailOutlined />}
                  loading={resendingEmail}
                  onClick={handleResendVerification}
                  style={{ padding: 0, marginTop: 8 }}
                >
                  Reenviar email de verificação
                </Button>
              </div>
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={onFinish}
          size="large"
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: 'Por favor, insira seu email!' },
              { type: 'email', message: 'Email inválido!' }
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="Email" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Por favor, insira sua senha!' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Senha" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Entrar
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center', marginBottom: 16 }}>
            <Link to="/forgot-password">Esqueci minha senha</Link>
            <span style={{ margin: '0 8px' }}>|</span>
            <Link to="/resend-verification">Reenviar verificação</Link>
          </div>

          <div style={{ textAlign: 'center' }}>
            <Link to="/register">Criar conta</Link>
            <span style={{ margin: '0 8px' }}>|</span>
            <Link to="/create-tenant">Cadastrar consultório</Link>
          </div>

          <div style={{ textAlign: 'center', marginTop: 24, paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
            <Link to="/termos-de-uso" style={{ fontSize: 12, color: '#888' }}>Termos de Uso</Link>
            <span style={{ margin: '0 8px', color: '#ddd' }}>|</span>
            <Link to="/politica-de-privacidade" style={{ fontSize: 12, color: '#888' }}>Privacidade</Link>
            <span style={{ margin: '0 8px', color: '#ddd' }}>|</span>
            <Link to="/seus-direitos-lgpd" style={{ fontSize: 12, color: '#888' }}>LGPD</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default Login;
