import React, { useState, useEffect } from 'react';
import { Card, Button, Typography, Row, Col, Space, message } from 'antd';
import { LockOutlined, CrownOutlined, LogoutOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import api from '../../services/api';

const { Title, Text, Paragraph } = Typography;

const SubscriptionRequired = () => {
  const navigate = useNavigate();
  const { logout, user, subscription } = useAuth();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState(null);

  useEffect(() => {
    loadPlans();
  }, []);

  const loadPlans = async () => {
    try {
      const response = await api.get('/subscription/plans');
      setPlans(response.data.plans);
    } catch (error) {
    } finally {
      setLoading(false);
    }
  };

  const handleCheckout = async (planId) => {
    setCheckoutLoading(planId);
    try {
      const response = await api.post('/subscription/checkout', { plan_id: planId });
      window.location.href = response.data.checkout_url;
    } catch (error) {
      message.error('Erro ao iniciar checkout');
      setCheckoutLoading(null);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const formatPrice = (cents) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(cents / 100);
  };

  const getPlanIcon = (planId) => {
    const icons = {
      bronze: '🥉',
      silver: '🥈',
      gold: '🥇',
    };
    return icons[planId] || '📦';
  };

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
          maxWidth: 900,
          width: '100%',
          textAlign: 'center',
          boxShadow: '0 8px 32px rgba(0,0,0,0.1)',
          borderRadius: 16,
        }}
      >
        <div style={{ marginBottom: 32 }}>
          <div
            style={{
              width: 80,
              height: 80,
              margin: '0 auto 24px',
              background: 'linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 100%)',
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <LockOutlined style={{ fontSize: 40, color: 'white' }} />
          </div>

          <Title level={2} style={{ marginBottom: 8 }}>
            Assinatura Necessária
          </Title>

          <Paragraph style={{ fontSize: 16, color: '#666', maxWidth: 500, margin: '0 auto' }}>
            Seu período de avaliação gratuita terminou. Para continuar usando o
            <strong> OdoWell</strong>, escolha um dos planos abaixo.
          </Paragraph>
        </div>

        {/* Plans */}
        <Row gutter={[16, 16]} style={{ marginBottom: 32 }}>
          {plans.map((plan) => {
            const isRecommended = plan.id === 'silver';

            return (
              <Col xs={24} md={8} key={plan.id}>
                <Card
                  style={{
                    height: '100%',
                    border: isRecommended ? '2px solid #66BB6A' : '1px solid #e8e8e8',
                    position: 'relative',
                  }}
                  bodyStyle={{ padding: 16 }}
                >
                  {isRecommended && (
                    <div
                      style={{
                        position: 'absolute',
                        top: -10,
                        left: '50%',
                        transform: 'translateX(-50%)',
                        background: '#66BB6A',
                        color: 'white',
                        padding: '2px 12px',
                        borderRadius: 10,
                        fontSize: 11,
                        fontWeight: 'bold',
                      }}
                    >
                      POPULAR
                    </div>
                  )}

                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 32, marginBottom: 4 }}>{getPlanIcon(plan.id)}</div>
                    <Title level={4} style={{ marginBottom: 4 }}>{plan.name}</Title>
                    <div>
                      <Text style={{ fontSize: 24, fontWeight: 'bold', color: '#66BB6A' }}>
                        {formatPrice(plan.price_monthly)}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>/mês</Text>
                    </div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      Até {plan.patient_limit.toLocaleString()} pacientes
                    </Text>
                  </div>

                  <Button
                    type={isRecommended ? 'primary' : 'default'}
                    block
                    style={{ marginTop: 16 }}
                    loading={checkoutLoading === plan.id}
                    onClick={() => handleCheckout(plan.id)}
                  >
                    Assinar
                  </Button>
                </Card>
              </Col>
            );
          })}
        </Row>

        <Space direction="vertical" size={16}>
          <Button
            type="link"
            icon={<LogoutOutlined />}
            onClick={handleLogout}
            style={{ color: '#999' }}
          >
            Sair da conta
          </Button>

          <Text type="secondary" style={{ fontSize: 12 }}>
            Pagamento seguro processado pelo Stripe. Cancele quando quiser.
          </Text>
        </Space>
      </Card>
    </div>
  );
};

export default SubscriptionRequired;
