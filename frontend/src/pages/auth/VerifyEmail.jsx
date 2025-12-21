import React, { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { Card, Typography, Spin, Result, Button } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, MailOutlined } from '@ant-design/icons';
import { authAPI } from '../../services/api';

const { Title } = Typography;

const VerifyEmail = () => {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState('loading'); // loading, success, error
  const [errorMessage, setErrorMessage] = useState('');
  const [verifiedEmail, setVerifiedEmail] = useState('');
  const [tenantName, setTenantName] = useState('');

  const token = searchParams.get('token');

  useEffect(() => {
    const verifyEmail = async () => {
      if (!token) {
        setStatus('error');
        setErrorMessage('Token de verificação não encontrado');
        return;
      }

      try {
        const response = await authAPI.verifyEmail(token);
        setVerifiedEmail(response.data.email || '');
        setTenantName(response.data.tenant_name || '');
        setStatus('success');
      } catch (error) {
        setStatus('error');
        setErrorMessage(error.response?.data?.error || 'Erro ao verificar email');
      }
    };

    verifyEmail();
  }, [token]);

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #81C784 0%, #66BB6A 100%)'
    }}>
      <Card style={{ width: 450, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
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
        </div>

        {status === 'loading' && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
            <Typography.Text style={{ display: 'block', marginTop: 16 }}>
              Verificando seu email...
            </Typography.Text>
          </div>
        )}

        {status === 'success' && (
          <Result
            icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            title="Email verificado com sucesso!"
            subTitle={
              <div>
                {tenantName && <div style={{ marginBottom: 8 }}><strong>Empresa:</strong> {tenantName}</div>}
                {verifiedEmail && <div style={{ marginBottom: 8 }}><strong>Email:</strong> {verifiedEmail}</div>}
                <div>Sua conta esta ativa. Agora voce pode fazer login e comecar a usar o sistema.</div>
              </div>
            }
            extra={[
              <Button type="primary" key="login">
                <Link to="/login">Fazer Login</Link>
              </Button>
            ]}
          />
        )}

        {status === 'error' && (
          <Result
            icon={<CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
            title="Erro na verificação"
            subTitle={errorMessage}
            extra={[
              <Button type="primary" key="resend">
                <Link to="/resend-verification">Reenviar email de verificação</Link>
              </Button>,
              <Button key="login">
                <Link to="/login">Voltar ao Login</Link>
              </Button>
            ]}
          />
        )}
      </Card>
    </div>
  );
};

export default VerifyEmail;
