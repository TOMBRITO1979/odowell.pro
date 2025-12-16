import React, { useState, useEffect } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Result, Spin } from 'antd';
import { LockOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import api from '../../services/api';

const { Title } = Typography;

const ResetPassword = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [validating, setValidating] = useState(true);
  const [tokenValid, setTokenValid] = useState(false);
  const [tokenError, setTokenError] = useState('');
  const [success, setSuccess] = useState(false);

  const token = searchParams.get('token');

  useEffect(() => {
    const validateToken = async () => {
      if (!token) {
        setTokenError('Token não fornecido');
        setValidating(false);
        return;
      }

      try {
        const response = await api.get(`/auth/validate-reset-token?token=${token}`);
        if (response.data.valid) {
          setTokenValid(true);
        } else {
          setTokenError(response.data.error || 'Token inválido');
        }
      } catch (error) {
        setTokenError('Erro ao validar token');
      } finally {
        setValidating(false);
      }
    };

    validateToken();
  }, [token]);

  const onFinish = async (values) => {
    if (values.new_password !== values.confirm_password) {
      message.error('As senhas não coincidem!');
      return;
    }

    setLoading(true);
    try {
      await api.post('/auth/reset-password', {
        token: token,
        new_password: values.new_password
      });
      setSuccess(true);
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao redefinir senha');
    } finally {
      setLoading(false);
    }
  };

  // Loading state
  if (validating) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
      }}>
        <Card style={{ width: 400, textAlign: 'center' }}>
          <Spin size="large" />
          <p style={{ marginTop: 16 }}>Validando link...</p>
        </Card>
      </div>
    );
  }

  // Token invalid
  if (!tokenValid) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
      }}>
        <Card style={{ width: 450, boxShadow: '0 8px 32px rgba(0,0,0,0.1)' }}>
          <Result
            status="error"
            title="Link inválido ou expirado"
            subTitle={tokenError || 'Este link de redefinição de senha não é mais válido. Solicite um novo link.'}
            extra={[
              <Link to="/forgot-password" key="forgot">
                <Button type="primary">Solicitar novo link</Button>
              </Link>,
              <Link to="/login" key="login">
                <Button>Voltar ao Login</Button>
              </Link>
            ]}
          />
        </Card>
      </div>
    );
  }

  // Success state
  if (success) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
      }}>
        <Card style={{ width: 450, boxShadow: '0 8px 32px rgba(0,0,0,0.1)' }}>
          <Result
            status="success"
            title="Senha redefinida com sucesso!"
            subTitle="Sua senha foi alterada. Você já pode fazer login com a nova senha."
            extra={[
              <Link to="/login" key="login">
                <Button type="primary" size="large">Fazer Login</Button>
              </Link>
            ]}
          />
        </Card>
      </div>
    );
  }

  // Reset password form
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
          <Title level={3} style={{ color: '#4CAF50', marginTop: 0 }}>Redefinir senha</Title>
          <Typography.Text type="secondary">
            Digite sua nova senha
          </Typography.Text>
        </div>

        <Form
          name="reset-password"
          onFinish={onFinish}
          size="large"
        >
          <Form.Item
            name="new_password"
            rules={[
              { required: true, message: 'Por favor, insira a nova senha!' },
              { min: 12, message: 'A senha deve ter no mínimo 12 caracteres' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Nova senha" />
          </Form.Item>

          <Form.Item
            name="confirm_password"
            rules={[
              { required: true, message: 'Por favor, confirme a senha!' },
              { min: 12, message: 'A senha deve ter no mínimo 12 caracteres' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Confirmar nova senha" />
          </Form.Item>

          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16, fontSize: 12 }}>
            A senha deve conter: mínimo 12 caracteres, 1 maiúscula, 1 minúscula, 1 número e 1 caractere especial.
          </Typography.Text>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Redefinir senha
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Link to="/login">
              <ArrowLeftOutlined /> Voltar ao login
            </Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default ResetPassword;
