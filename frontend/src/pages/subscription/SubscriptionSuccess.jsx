import React, { useEffect } from 'react';
import { Result, Button, Card } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const SubscriptionSuccess = () => {
  const navigate = useNavigate();
  const { refreshUser } = useAuth();

  useEffect(() => {
    // Refresh user data to get updated subscription status
    refreshUser();
  }, []);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #E8F5E9 0%, #C8E6C9 100%)',
        padding: 24,
      }}
    >
      <Card
        style={{
          maxWidth: 500,
          width: '100%',
          textAlign: 'center',
          boxShadow: '0 8px 32px rgba(0,0,0,0.1)',
          borderRadius: 16,
        }}
      >
        <Result
          icon={<CheckCircleOutlined style={{ color: '#66BB6A' }} />}
          title="Assinatura Ativada!"
          subTitle="Sua assinatura foi processada com sucesso. Você já pode usar todas as funcionalidades do sistema."
          extra={[
            <Button
              type="primary"
              key="dashboard"
              size="large"
              onClick={() => navigate('/')}
            >
              Ir para o Dashboard
            </Button>,
            <Button
              key="subscription"
              onClick={() => navigate('/subscription')}
            >
              Ver Detalhes da Assinatura
            </Button>,
          ]}
        />
      </Card>
    </div>
  );
};

export default SubscriptionSuccess;
