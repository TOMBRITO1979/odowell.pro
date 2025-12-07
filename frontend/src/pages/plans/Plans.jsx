import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Select,
  Card,
  message,
  Popconfirm,
  Row,
  Col,
  Statistic,
  Typography,
  Modal,
  Tooltip,
  Alert,
  Input,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  StopOutlined,
  SyncOutlined,
  SendOutlined,
  CreditCardOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  LinkOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientSubscriptionsAPI, patientsAPI, stripeSettingsAPI } from '../../services/api';
import { statusColors, shadows, actionColors } from '../../theme/designSystem';
import { usePermission } from '../../contexts/AuthContext';

const { Text } = Typography;

const Plans = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { canCreate, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [subscriptions, setSubscriptions] = useState([]);
  const [patients, setPatients] = useState([]);
  const [stripeConnected, setStripeConnected] = useState(false);
  const [stats, setStats] = useState({
    total: 0,
    active: 0,
    pending: 0,
    canceled: 0,
  });
  const [selectedSubscription, setSelectedSubscription] = useState(null);
  const [detailsModalVisible, setDetailsModalVisible] = useState(false);
  const [payments, setPayments] = useState([]);
  const [loadingPayments, setLoadingPayments] = useState(false);

  const [filters, setFilters] = useState({
    patient_id: searchParams.get('patient_id') || undefined,
    status: undefined,
  });

  const statusOptions = [
    { value: 'active', label: 'Ativo', color: statusColors.success, icon: <CheckCircleOutlined /> },
    { value: 'pending', label: 'Pendente', color: statusColors.warning, icon: <ClockCircleOutlined /> },
    { value: 'past_due', label: 'Atrasado', color: statusColors.error, icon: <ExclamationCircleOutlined /> },
    { value: 'canceled', label: 'Cancelado', color: statusColors.default, icon: <CloseCircleOutlined /> },
    { value: 'trialing', label: 'Trial', color: statusColors.info, icon: <ClockCircleOutlined /> },
  ];

  useEffect(() => {
    checkStripeConnection();
    fetchSubscriptions();
    fetchPatients();

    // Show success/canceled message from Stripe redirect
    if (searchParams.get('success') === 'true') {
      message.success('Assinatura criada com sucesso! O pagamento foi processado.');
    } else if (searchParams.get('canceled') === 'true') {
      message.info('Checkout cancelado. A assinatura permanece pendente.');
    }
  }, [filters]);

  const checkStripeConnection = async () => {
    try {
      const response = await stripeSettingsAPI.get();
      setStripeConnected(response.data.stripe_connected || false);
    } catch (error) {
      console.error('Error checking Stripe connection:', error);
    }
  };

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchSubscriptions = async () => {
    setLoading(true);
    try {
      const response = await patientSubscriptionsAPI.getAll(filters);
      const list = response.data || [];
      setSubscriptions(list);

      // Calculate stats
      const active = list.filter(s => s.status === 'active').length;
      const pending = list.filter(s => s.status === 'pending').length;
      const canceled = list.filter(s => s.status === 'canceled').length;

      setStats({
        total: list.length,
        active,
        pending,
        canceled,
      });
    } catch (error) {
      message.error('Erro ao carregar assinaturas');
      console.error('Error fetching subscriptions:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = async (record) => {
    setSelectedSubscription(record);
    setDetailsModalVisible(true);
    setLoadingPayments(true);

    try {
      const response = await patientSubscriptionsAPI.getOne(record.id);
      setSelectedSubscription(response.data.subscription);
      setPayments(response.data.payments || []);
    } catch (error) {
      message.error('Erro ao carregar detalhes');
    } finally {
      setLoadingPayments(false);
    }
  };

  const handleCancel = async (id) => {
    try {
      await patientSubscriptionsAPI.cancel(id);
      message.success('Assinatura será cancelada ao final do período atual');
      fetchSubscriptions();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cancelar assinatura');
    }
  };

  const handleCancelImmediately = async (id) => {
    try {
      await patientSubscriptionsAPI.cancelImmediately(id);
      message.success('Assinatura cancelada imediatamente');
      fetchSubscriptions();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cancelar assinatura');
    }
  };

  const handleRefresh = async (id) => {
    try {
      await patientSubscriptionsAPI.refresh(id);
      message.success('Status atualizado');
      fetchSubscriptions();
    } catch (error) {
      message.error('Erro ao atualizar status');
    }
  };

  const handleResendLink = async (id) => {
    try {
      const response = await patientSubscriptionsAPI.resendLink(id);
      const checkoutUrl = response.data.checkout_url;

      Modal.success({
        title: 'Link de Pagamento Gerado',
        content: (
          <div>
            <p>Copie o link abaixo e envie para o paciente:</p>
            <Input.TextArea
              value={checkoutUrl}
              readOnly
              autoSize={{ minRows: 2 }}
              style={{ marginBottom: 8 }}
            />
            <Button
              type="primary"
              icon={<LinkOutlined />}
              onClick={() => {
                navigator.clipboard.writeText(checkoutUrl);
                message.success('Link copiado!');
              }}
            >
              Copiar Link
            </Button>
          </div>
        ),
        okText: 'Fechar',
      });

      fetchSubscriptions();
    } catch (error) {
      message.error('Erro ao gerar novo link');
    }
  };

  const getStatusTag = (status) => {
    const option = statusOptions.find(o => o.value === status) || statusOptions[3];
    return (
      <Tag color={option.color} icon={option.icon}>
        {option.label}
      </Tag>
    );
  };

  const formatCurrency = (amount, currency = 'BRL') => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(amount / 100);
  };

  const columns = [
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient',
      render: (text, record) => (
        <a onClick={() => navigate(`/patients/${record.patient_id}`)}>{text || 'N/A'}</a>
      ),
    },
    {
      title: 'Plano',
      dataIndex: 'product_name',
      key: 'product_name',
    },
    {
      title: 'Valor',
      key: 'value',
      render: (_, record) => (
        <Text strong>
          {formatCurrency(record.price_amount, record.price_currency)}
          <Text type="secondary" style={{ fontSize: 12 }}>
            /{record.interval === 'month' ? 'mês' : record.interval === 'year' ? 'ano' : record.interval}
          </Text>
        </Text>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Space direction="vertical" size={0}>
          {getStatusTag(status)}
          {record.cancel_at_period_end && (
            <Text type="secondary" style={{ fontSize: 11 }}>
              Cancela em {dayjs(record.current_period_end).format('DD/MM/YYYY')}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: 'Período Atual',
      key: 'period',
      render: (_, record) => (
        record.current_period_end ? (
          <Text type="secondary">
            Até {dayjs(record.current_period_end).format('DD/MM/YYYY')}
          </Text>
        ) : (
          <Text type="secondary">-</Text>
        )
      ),
    },
    {
      title: 'Criado em',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="Ver Detalhes">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetails(record)}
              style={{ color: actionColors.view }}
            />
          </Tooltip>

          <Tooltip title="Atualizar Status">
            <Button
              type="text"
              icon={<SyncOutlined />}
              onClick={() => handleRefresh(record.id)}
              style={{ color: actionColors.edit }}
            />
          </Tooltip>

          {record.status === 'pending' && record.checkout_url && (
            <Tooltip title="Reenviar Link">
              <Button
                type="text"
                icon={<SendOutlined />}
                onClick={() => handleResendLink(record.id)}
                style={{ color: actionColors.save }}
              />
            </Tooltip>
          )}

          {['active', 'trialing', 'past_due'].includes(record.status) && !record.cancel_at_period_end && (
            <Popconfirm
              title="Cancelar assinatura?"
              description="A assinatura será cancelada ao final do período atual."
              onConfirm={() => handleCancel(record.id)}
              okText="Sim, cancelar"
              cancelText="Não"
            >
              <Tooltip title="Cancelar ao final do período">
                <Button type="text" icon={<StopOutlined />} style={{ color: actionColors.delete }} />
              </Tooltip>
            </Popconfirm>
          )}

          {canDelete('payments') && record.status !== 'canceled' && (
            <Popconfirm
              title="Cancelar imediatamente?"
              description="A assinatura será cancelada imediatamente. Esta ação não pode ser desfeita."
              onConfirm={() => handleCancelImmediately(record.id)}
              okText="Sim, cancelar agora"
              cancelText="Não"
            >
              <Tooltip title="Cancelar imediatamente">
                <Button type="text" icon={<CloseCircleOutlined />} style={{ color: actionColors.delete }} />
              </Tooltip>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  const paymentColumns = [
    {
      title: 'Data',
      dataIndex: 'paid_at',
      key: 'paid_at',
      render: (date, record) => date ? dayjs(date).format('DD/MM/YYYY HH:mm') : dayjs(record.created_at).format('DD/MM/YYYY'),
    },
    {
      title: 'Período',
      key: 'period',
      render: (_, record) => (
        <Text type="secondary">
          {dayjs(record.period_start).format('DD/MM')} - {dayjs(record.period_end).format('DD/MM/YYYY')}
        </Text>
      ),
    },
    {
      title: 'Valor',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount, record) => formatCurrency(amount, record.currency),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const colors = {
          paid: 'green',
          open: 'orange',
          void: 'default',
          uncollectible: 'red',
          draft: 'blue',
        };
        const labels = {
          paid: 'Pago',
          open: 'Aberto',
          void: 'Anulado',
          uncollectible: 'Não Cobrável',
          draft: 'Rascunho',
        };
        return <Tag color={colors[status]}>{labels[status] || status}</Tag>;
      },
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {record.invoice_url && (
            <Tooltip title="Ver Fatura">
              <Button
                type="text"
                icon={<LinkOutlined />}
                onClick={() => window.open(record.invoice_url, '_blank')}
                style={{ color: actionColors.view }}
              />
            </Tooltip>
          )}
          {record.receipt_url && (
            <Tooltip title="Ver Recibo">
              <Button
                type="text"
                icon={<CreditCardOutlined />}
                onClick={() => window.open(record.receipt_url, '_blank')}
                style={{ color: actionColors.print }}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      {!stripeConnected && (
        <Alert
          message="Stripe não configurado"
          description={
            <span>
              Para utilizar assinaturas recorrentes, configure suas credenciais do Stripe nas{' '}
              <a onClick={() => navigate('/settings')}>Configurações</a>.
            </span>
          }
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {/* Stats Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <Card style={{ boxShadow: shadows.card }}>
            <Statistic
              title="Total de Assinaturas"
              value={stats.total}
              prefix={<CreditCardOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card style={{ boxShadow: shadows.card }}>
            <Statistic
              title="Ativas"
              value={stats.active}
              valueStyle={{ color: statusColors.success }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card style={{ boxShadow: shadows.card }}>
            <Statistic
              title="Pendentes"
              value={stats.pending}
              valueStyle={{ color: statusColors.warning }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card style={{ boxShadow: shadows.card }}>
            <Statistic
              title="Canceladas"
              value={stats.canceled}
              valueStyle={{ color: statusColors.default }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16, boxShadow: shadows.card }}>
        <Row gutter={16} align="middle">
          <Col flex="auto">
            <Space wrap>
              <Select
                allowClear
                placeholder="Filtrar por paciente"
                style={{ width: 200 }}
                value={filters.patient_id}
                onChange={(value) => setFilters({ ...filters, patient_id: value })}
                showSearch
                optionFilterProp="children"
              >
                {patients.map(p => (
                  <Select.Option key={p.id} value={p.id}>{p.name}</Select.Option>
                ))}
              </Select>
              <Select
                allowClear
                placeholder="Filtrar por status"
                style={{ width: 150 }}
                value={filters.status}
                onChange={(value) => setFilters({ ...filters, status: value })}
              >
                {statusOptions.map(s => (
                  <Select.Option key={s.value} value={s.value}>
                    <Space>
                      {s.icon}
                      {s.label}
                    </Space>
                  </Select.Option>
                ))}
              </Select>
            </Space>
          </Col>
          <Col>
            {canCreate('payments') && (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate('/plans/new')}
              >
                Nova Assinatura
              </Button>
            )}
          </Col>
        </Row>
      </Card>

      {/* Subscriptions Table */}
      <Card style={{ boxShadow: shadows.card }}>
        <Table
          columns={columns}
          dataSource={subscriptions}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 20 }}
        />
      </Card>

      {/* Details Modal */}
      <Modal
        title={
          <Space>
            <CreditCardOutlined />
            Detalhes da Assinatura
          </Space>
        }
        open={detailsModalVisible}
        onCancel={() => setDetailsModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedSubscription && (
          <div>
            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="Paciente"
                    value={selectedSubscription.patient?.name || 'N/A'}
                    valueStyle={{ fontSize: 16 }}
                  />
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="Plano"
                    value={selectedSubscription.product_name}
                    valueStyle={{ fontSize: 16 }}
                    suffix={
                      <Text type="secondary" style={{ fontSize: 14 }}>
                        {formatCurrency(selectedSubscription.price_amount, selectedSubscription.price_currency)}
                        /{selectedSubscription.interval === 'month' ? 'mês' : 'ano'}
                      </Text>
                    }
                  />
                </Card>
              </Col>
            </Row>

            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <Text type="secondary">Status:</Text>
                <div>{getStatusTag(selectedSubscription.status)}</div>
              </Col>
              <Col span={8}>
                <Text type="secondary">Período Atual:</Text>
                <div>
                  {selectedSubscription.current_period_start && (
                    <Text>
                      {dayjs(selectedSubscription.current_period_start).format('DD/MM/YYYY')} - {dayjs(selectedSubscription.current_period_end).format('DD/MM/YYYY')}
                    </Text>
                  )}
                </div>
              </Col>
              <Col span={8}>
                <Text type="secondary">Criado em:</Text>
                <div>
                  <Text>{dayjs(selectedSubscription.created_at).format('DD/MM/YYYY HH:mm')}</Text>
                </div>
              </Col>
            </Row>

            {selectedSubscription.notes && (
              <div style={{ marginBottom: 24 }}>
                <Text type="secondary">Observações:</Text>
                <div><Text>{selectedSubscription.notes}</Text></div>
              </div>
            )}

            <div>
              <Text strong style={{ fontSize: 16 }}>Últimos Pagamentos</Text>
              <Table
                columns={paymentColumns}
                dataSource={payments}
                rowKey="id"
                loading={loadingPayments}
                pagination={false}
                size="small"
                style={{ marginTop: 8 }}
              />
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Plans;
