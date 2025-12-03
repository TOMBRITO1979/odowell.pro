import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Form,
  Select,
  Button,
  Card,
  message,
  Space,
  Typography,
  Alert,
  Spin,
  Row,
  Col,
  Divider,
  Input,
  Modal,
} from 'antd';
import {
  ArrowLeftOutlined,
  CreditCardOutlined,
  UserOutlined,
  LinkOutlined,
  SendOutlined,
} from '@ant-design/icons';
import { patientSubscriptionsAPI, patientsAPI, settingsAPI } from '../../services/api';
import { shadows } from '../../theme/designSystem';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const PlanForm = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(true);
  const [patients, setPatients] = useState([]);
  const [products, setProducts] = useState([]);
  const [stripeConnected, setStripeConnected] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [selectedPrice, setSelectedPrice] = useState(null);
  const [checkoutUrl, setCheckoutUrl] = useState(null);

  useEffect(() => {
    loadInitialData();
  }, []);

  const loadInitialData = async () => {
    setLoadingData(true);
    try {
      // Check Stripe connection
      const stripeRes = await settingsAPI.getStripeSettings();
      const isConnected = stripeRes.data.stripe_connected || false;
      setStripeConnected(isConnected);

      if (!isConnected) {
        setLoadingData(false);
        return;
      }

      // Load patients and products in parallel
      const [patientsRes, productsRes] = await Promise.all([
        patientsAPI.getAll({ page: 1, page_size: 1000 }),
        patientSubscriptionsAPI.getStripeProducts(),
      ]);

      setPatients(patientsRes.data.patients || []);
      setProducts(productsRes.data || []);
    } catch (error) {
      console.error('Error loading data:', error);
      message.error('Erro ao carregar dados');
    } finally {
      setLoadingData(false);
    }
  };

  const handleProductChange = (productId) => {
    const product = products.find(p => p.id === productId);
    setSelectedProduct(product);
    setSelectedPrice(null);
    form.setFieldValue('stripe_price_id', undefined);
  };

  const handlePriceChange = (priceId) => {
    if (!selectedProduct) return;
    const price = selectedProduct.prices.find(p => p.id === priceId);
    setSelectedPrice(price);
  };

  const formatCurrency = (amount, currency = 'BRL') => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(amount / 100);
  };

  const formatInterval = (interval, count) => {
    if (count === 1) {
      const labels = { month: 'mês', year: 'ano', week: 'semana', day: 'dia' };
      return labels[interval] || interval;
    }
    const labels = { month: 'meses', year: 'anos', week: 'semanas', day: 'dias' };
    return `${count} ${labels[interval] || interval}`;
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const response = await patientSubscriptionsAPI.create(values);
      const url = response.data.checkout_url;
      setCheckoutUrl(url);

      Modal.success({
        title: 'Assinatura Criada!',
        content: (
          <div>
            <Paragraph>
              A assinatura foi criada com sucesso. Envie o link de pagamento abaixo para o paciente:
            </Paragraph>
            <TextArea
              value={url}
              readOnly
              autoSize={{ minRows: 2, maxRows: 4 }}
              style={{ marginBottom: 16 }}
            />
            <Space>
              <Button
                type="primary"
                icon={<LinkOutlined />}
                onClick={() => {
                  navigator.clipboard.writeText(url);
                  message.success('Link copiado para a área de transferência!');
                }}
              >
                Copiar Link
              </Button>
              <Button
                icon={<SendOutlined />}
                onClick={() => window.open(url, '_blank')}
              >
                Abrir Link
              </Button>
            </Space>
          </div>
        ),
        okText: 'Ir para Planos',
        onOk: () => navigate('/plans'),
        width: 500,
      });
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar assinatura');
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
        <p>Carregando...</p>
      </div>
    );
  }

  if (!stripeConnected) {
    return (
      <Card style={{ boxShadow: shadows.card }}>
        <Alert
          message="Stripe não configurado"
          description={
            <div>
              <Paragraph>
                Para criar assinaturas recorrentes, você precisa configurar suas credenciais do Stripe.
              </Paragraph>
              <Button type="primary" onClick={() => navigate('/settings')}>
                Ir para Configurações
              </Button>
            </div>
          }
          type="warning"
          showIcon
        />
      </Card>
    );
  }

  if (products.length === 0) {
    return (
      <Card style={{ boxShadow: shadows.card }}>
        <Alert
          message="Nenhum produto encontrado"
          description={
            <div>
              <Paragraph>
                Você precisa criar produtos recorrentes no Stripe antes de criar assinaturas.
              </Paragraph>
              <Paragraph type="secondary">
                Acesse o painel do Stripe e crie produtos com preços recorrentes (mensal, anual, etc).
              </Paragraph>
              <Button
                type="primary"
                onClick={() => window.open('https://dashboard.stripe.com/products', '_blank')}
              >
                Abrir Stripe Dashboard
              </Button>
            </div>
          }
          type="info"
          showIcon
        />
        <div style={{ marginTop: 16 }}>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/plans')}>
            Voltar
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div>
      <Card style={{ marginBottom: 16, boxShadow: shadows.card }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/plans')}>
            Voltar
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            <CreditCardOutlined /> Nova Assinatura
          </Title>
        </Space>
      </Card>

      <Card style={{ boxShadow: shadows.card }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
        >
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <Form.Item
                name="patient_id"
                label="Paciente"
                rules={[{ required: true, message: 'Selecione um paciente' }]}
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
                  optionFilterProp="children"
                  suffixIcon={<UserOutlined />}
                >
                  {patients.map(p => (
                    <Select.Option key={p.id} value={p.id}>
                      {p.name} {p.email && `(${p.email})`}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Divider>Plano</Divider>

          <Row gutter={24}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Produto"
                rules={[{ required: true, message: 'Selecione um produto' }]}
              >
                <Select
                  placeholder="Selecione o produto"
                  onChange={handleProductChange}
                  value={selectedProduct?.id}
                >
                  {products.map(p => (
                    <Select.Option key={p.id} value={p.id}>
                      {p.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>

              {selectedProduct && selectedProduct.description && (
                <Alert
                  message={selectedProduct.description}
                  type="info"
                  style={{ marginBottom: 16 }}
                />
              )}
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="stripe_price_id"
                label="Preço"
                rules={[{ required: true, message: 'Selecione um preço' }]}
              >
                <Select
                  placeholder="Selecione o preço"
                  onChange={handlePriceChange}
                  disabled={!selectedProduct}
                >
                  {selectedProduct?.prices?.map(p => (
                    <Select.Option key={p.id} value={p.id}>
                      {formatCurrency(p.unit_amount, p.currency)} / {formatInterval(p.interval, p.interval_count)}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>

              {selectedPrice && (
                <Card size="small" style={{ backgroundColor: '#f6ffed', border: '1px solid #b7eb8f' }}>
                  <Text strong style={{ fontSize: 20, color: '#52c41a' }}>
                    {formatCurrency(selectedPrice.unit_amount, selectedPrice.currency)}
                  </Text>
                  <Text type="secondary">
                    {' '}/ {formatInterval(selectedPrice.interval, selectedPrice.interval_count)}
                  </Text>
                </Card>
              )}
            </Col>
          </Row>

          <Divider>Observações (opcional)</Divider>

          <Row gutter={24}>
            <Col xs={24}>
              <Form.Item
                name="notes"
                label="Notas internas"
              >
                <TextArea
                  rows={3}
                  placeholder="Adicione observações sobre esta assinatura (visível apenas para a equipe)"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading} size="large">
                Criar Assinatura e Gerar Link
              </Button>
              <Button onClick={() => navigate('/plans')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default PlanForm;
