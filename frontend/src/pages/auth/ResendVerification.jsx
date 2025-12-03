import React, { useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Result } from 'antd';
import { MailOutlined, SendOutlined } from '@ant-design/icons';
import { authAPI } from '../../services/api';

const { Title } = Typography;

const ResendVerification = () => {
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const location = useLocation();
  const initialEmail = location.state?.email || '';

  const onFinish = async (values) => {
    setLoading(true);
    try {
      await authAPI.resendVerification(values.email);
      setSent(true);
      message.success('Email de verificação enviado!');
    } catch (error) {
      // Don't reveal if email exists or not for security
      message.info('Se este email estiver cadastrado, você receberá um link de verificação.');
      setSent(true);
    } finally {
      setLoading(false);
    }
  };

  if (sent) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #81C784 0%, #66BB6A 100%)'
      }}>
        <Card style={{ width: 450, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
          <Result
            icon={<MailOutlined style={{ color: '#4CAF50' }} />}
            title="Email enviado!"
            subTitle="Verifique sua caixa de entrada e spam. O link de verificação expira em 24 horas."
            extra={[
              <Button type="primary" key="login">
                <Link to="/login">Voltar ao Login</Link>
              </Button>,
              <Button key="resend" onClick={() => setSent(false)}>
                Enviar novamente
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
      background: 'linear-gradient(135deg, #81C784 0%, #66BB6A 100%)'
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
            <MailOutlined style={{ fontSize: 40, color: '#fff' }} />
          </div>
          <Title level={3} style={{ color: '#4CAF50', marginTop: 0 }}>
            Verificar Email
          </Title>
          <Typography.Text type="secondary">
            Digite seu email para receber um novo link de verificação
          </Typography.Text>
        </div>

        <Form
          name="resend-verification"
          onFinish={onFinish}
          size="large"
          initialValues={{ email: initialEmail }}
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: 'Por favor, insira seu email!' },
              { type: 'email', message: 'Email inválido!' }
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="Email" />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
              icon={<SendOutlined />}
            >
              Enviar Link de Verificação
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Link to="/login">Voltar ao Login</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default ResendVerification;
