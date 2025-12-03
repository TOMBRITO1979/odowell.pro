import React, { useState, useEffect } from 'react';
import { Card, Button, Tag, Row, Col, Typography, Spin, message, Alert, Progress } from 'antd';
import { CrownOutlined, CheckCircleOutlined, ClockCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import api from '../../services/api';

const { Title, Text, Paragraph } = Typography;

const Subscription = () => {
  const { user, subscription, refreshUser } = useAuth();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState(null);
  const [portalLoading, setPortalLoading] = useState(false);
  const [expiredAlert, setExpiredAlert] = useState(null);

  useEffect(() => {
    loadPlans();
    // Check if user was redirected due to subscription expiration
    const expiredInfo = localStorage.getItem('subscription_expired');
    if (expiredInfo) {
      try {
        setExpiredAlert(JSON.parse(expiredInfo));
      } catch (e) {
        // ignore parse error
      }
      localStorage.removeItem('subscription_expired');
    }
  }, []);

  const loadPlans = async () => {
    try {
      const response = await api.get('/subscription/plans');
      setPlans(response.data.plans);
    } catch (error) {
      message.error('Erro ao carregar planos');
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

  const handleManageSubscription = async () => {
    setPortalLoading(true);
    try {
      const response = await api.post('/subscription/portal');
      window.location.href = response.data.portal_url;
    } catch (error) {
      message.error('Erro ao abrir portal de assinatura');
      setPortalLoading(false);
    }
  };

  const getStatusTag = () => {
    if (!subscription) return null;

    const statusConfig = {
      active: { color: 'success', text: 'Ativo', icon: <CheckCircleOutlined /> },
      trialing: { color: 'processing', text: 'Período de Avaliação', icon: <ClockCircleOutlined /> },
      past_due: { color: 'warning', text: 'Pagamento Pendente', icon: <ExclamationCircleOutlined /> },
      canceled: { color: 'error', text: 'Cancelado', icon: <ExclamationCircleOutlined /> },
      expired: { color: 'error', text: 'Expirado', icon: <ExclamationCircleOutlined /> },
    };

    const config = statusConfig[subscription.status] || statusConfig.expired;
    return <Tag color={config.color} icon={config.icon}>{config.text}</Tag>;
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

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: 1200, margin: '0 auto' }}>
      <Title level={2}>
        <CrownOutlined style={{ marginRight: 8 }} />
        Assinatura
      </Title>

      {/* Subscription Expired Alert (when redirected from blocked route) */}
      {expiredAlert && (
        <Alert
          type="error"
          showIcon
          icon={<ExclamationCircleOutlined />}
          message="Acesso Bloqueado - Assinatura Necessária"
          description={
            <div>
              <p>{expiredAlert.message}</p>
              <p>Seu período de avaliação expirou{expiredAlert.days_expired > 0 ? ` há ${expiredAlert.days_expired} dia(s)` : ''}.
                 Para continuar usando todas as funcionalidades do sistema, escolha um plano abaixo.</p>
            </div>
          }
          style={{ marginBottom: 24 }}
          closable
          onClose={() => setExpiredAlert(null)}
        />
      )}

      {/* Current Status Card */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]} align="middle">
          <Col xs={24} md={16}>
            <Title level={4} style={{ marginBottom: 8 }}>Status Atual</Title>
            <div style={{ marginBottom: 16 }}>
              {getStatusTag()}
              {subscription?.plan_type && (
                <Tag style={{ marginLeft: 8 }}>
                  Plano {subscription.plan_type.charAt(0).toUpperCase() + subscription.plan_type.slice(1)}
                </Tag>
              )}
            </div>

            {subscription?.status === 'trialing' && subscription?.trial_days_remaining > 0 && (
              <Alert
                type="info"
                showIcon
                message={`Restam ${subscription.trial_days_remaining} dias de avaliação gratuita`}
                description="Assine agora para não perder acesso ao sistema após o período de avaliação."
                style={{ marginBottom: 16 }}
              />
            )}

            {subscription?.status === 'trialing' && subscription?.trial_days_remaining <= 0 && (
              <Alert
                type="error"
                showIcon
                message="Seu período de avaliação expirou"
                description="Assine agora para continuar usando o sistema."
                style={{ marginBottom: 16 }}
              />
            )}

            {subscription?.patient_limit && (
              <div>
                <Text type="secondary">Limite de pacientes: </Text>
                <Text strong>{subscription.patient_limit.toLocaleString()}</Text>
              </div>
            )}
          </Col>
          <Col xs={24} md={8} style={{ textAlign: 'right' }}>
            {subscription?.status === 'active' && (
              <Button
                type="primary"
                size="large"
                onClick={handleManageSubscription}
                loading={portalLoading}
              >
                Gerenciar Assinatura
              </Button>
            )}
          </Col>
        </Row>
      </Card>

      {/* Plans */}
      <Title level={3} style={{ marginBottom: 24 }}>Escolha seu Plano</Title>

      <Row gutter={[24, 24]}>
        {plans.map((plan) => {
          const isCurrentPlan = subscription?.plan_type === plan.id && subscription?.status === 'active';
          const isRecommended = plan.id === 'silver';

          return (
            <Col xs={24} md={8} key={plan.id}>
              <Card
                hoverable
                style={{
                  height: '100%',
                  border: isRecommended ? '2px solid #66BB6A' : undefined,
                  position: 'relative',
                }}
              >
                {isRecommended && (
                  <div
                    style={{
                      position: 'absolute',
                      top: -12,
                      left: '50%',
                      transform: 'translateX(-50%)',
                      background: '#66BB6A',
                      color: 'white',
                      padding: '4px 16px',
                      borderRadius: 12,
                      fontSize: 12,
                      fontWeight: 'bold',
                    }}
                  >
                    RECOMENDADO
                  </div>
                )}

                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                  <div style={{ fontSize: 48, marginBottom: 8 }}>{getPlanIcon(plan.id)}</div>
                  <Title level={3} style={{ marginBottom: 0 }}>{plan.name}</Title>
                  <div style={{ marginTop: 16 }}>
                    <Text style={{ fontSize: 36, fontWeight: 'bold', color: '#66BB6A' }}>
                      {formatPrice(plan.price_monthly)}
                    </Text>
                    <Text type="secondary">/mês</Text>
                  </div>
                </div>

                <div style={{ marginBottom: 24 }}>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Até <strong>{plan.patient_limit.toLocaleString()}</strong> pacientes
                  </Paragraph>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Agenda ilimitada
                  </Paragraph>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Prontuário eletrônico
                  </Paragraph>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Orçamentos e financeiro
                  </Paragraph>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Relatórios completos
                  </Paragraph>
                  <Paragraph>
                    <CheckCircleOutlined style={{ color: '#66BB6A', marginRight: 8 }} />
                    Suporte por email
                  </Paragraph>
                </div>

                <Button
                  type={isCurrentPlan ? 'default' : 'primary'}
                  size="large"
                  block
                  disabled={isCurrentPlan}
                  loading={checkoutLoading === plan.id}
                  onClick={() => handleCheckout(plan.id)}
                >
                  {isCurrentPlan ? 'Plano Atual' : 'Assinar'}
                </Button>
              </Card>
            </Col>
          );
        })}
      </Row>

      {/* FAQ Section */}
      <Card style={{ marginTop: 24 }}>
        <Title level={4}>Perguntas Frequentes</Title>

        <Paragraph>
          <Text strong>Posso mudar de plano depois?</Text>
          <br />
          <Text type="secondary">
            Sim! Você pode fazer upgrade ou downgrade do seu plano a qualquer momento
            através do portal de assinatura.
          </Text>
        </Paragraph>

        <Paragraph>
          <Text strong>Como funciona o cancelamento?</Text>
          <br />
          <Text type="secondary">
            Você pode cancelar a qualquer momento. O acesso permanece até o fim do
            período já pago.
          </Text>
        </Paragraph>

        <Paragraph>
          <Text strong>Quais formas de pagamento são aceitas?</Text>
          <br />
          <Text type="secondary">
            Aceitamos cartões de crédito (Visa, Mastercard, American Express) e
            boleto bancário.
          </Text>
        </Paragraph>
      </Card>
    </div>
  );
};

export default Subscription;
