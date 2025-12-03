import React from 'react';
import { Result, Button, Card } from 'antd';
import { CloseCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const SubscriptionCancel = () => {
  const navigate = useNavigate();

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%)',
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
          icon={<CloseCircleOutlined style={{ color: '#ff6b6b' }} />}
          title="Assinatura não concluída"
          subTitle="O processo de assinatura foi cancelado. Você pode tentar novamente quando quiser."
          extra={[
            <Button
              type="primary"
              key="retry"
              size="large"
              onClick={() => navigate('/subscription')}
            >
              Escolher um Plano
            </Button>,
            <Button
              key="dashboard"
              onClick={() => navigate('/')}
            >
              Voltar ao Dashboard
            </Button>,
          ]}
        />
      </Card>
    </div>
  );
};

export default SubscriptionCancel;
