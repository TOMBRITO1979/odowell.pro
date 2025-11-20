import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  message,
  Descriptions,
  Space,
  Button,
  Tag,
  Table,
  Divider,
  Modal,
  Form,
  InputNumber,
  DatePicker,
  Select,
  Row,
  Col,
  Statistic,
  Input,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DollarOutlined,
  FilePdfOutlined,
  PlusOutlined,
  PrinterOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { budgetsAPI, paymentsAPI } from '../../services/api';

const BudgetView = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [budget, setBudget] = useState(null);
  const [items, setItems] = useState([]);
  const [payments, setPayments] = useState([]);
  const [paymentModalVisible, setPaymentModalVisible] = useState(false);
  const [paymentForm] = Form.useForm();

  const statusOptions = {
    pending: { label: 'Pendente', color: 'warning' },
    approved: { label: 'Aprovado', color: 'success' },
    rejected: { label: 'Rejeitado', color: 'error' },
    expired: { label: 'Expirado', color: 'default' },
  };

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'zelle', label: 'Zelle' },
    { value: 'insurance', label: 'Convênio' },
  ];

  useEffect(() => {
    fetchBudget();
  }, [id]);

  const fetchBudget = async () => {
    setLoading(true);
    try {
      const response = await budgetsAPI.getOne(id);
      const budgetData = response.data.budget;
      setBudget(budgetData);

      // Parse items from JSON string
      if (budgetData.items) {
        try {
          const parsedItems = JSON.parse(budgetData.items);
          setItems(parsedItems);
        } catch (e) {
          console.error('Error parsing items:', e);
        }
      }

      // Set payments from budget
      if (budgetData.payments) {
        setPayments(budgetData.payments);
      }
    } catch (error) {
      message.error('Erro ao carregar orçamento');
    } finally {
      setLoading(false);
    }
  };

  const handleAddPayment = () => {
    if (budget.status !== 'approved') {
      message.warning('Apenas orçamentos aprovados podem receber pagamentos');
      return;
    }

    // Calculate next installment number
    const paidPayments = payments.filter(p => p.status === 'paid').length;
    const nextInstallment = paidPayments + 1;

    paymentForm.setFieldsValue({
      installment_number: nextInstallment,
      paid_date: dayjs(),
    });

    setPaymentModalVisible(true);
  };

  const handlePaymentSubmit = async (values) => {
    try {
      const paymentData = {
        budget_id: parseInt(id),
        patient_id: budget.patient_id,
        type: 'income',
        category: 'treatment',
        description: `Pagamento ${values.installment_number}/${values.total_installments || 1} - Orçamento #${id}`,
        amount: parseFloat(values.amount),
        payment_method: values.payment_method,
        is_installment: values.total_installments > 1,
        installment_number: parseInt(values.installment_number),
        total_installments: parseInt(values.total_installments) || 1,
        status: 'paid',
        paid_date: values.paid_date.format('YYYY-MM-DDTHH:mm:ss') + 'Z',
        notes: values.notes || '',
      };

      await paymentsAPI.create(paymentData);
      message.success('Pagamento adicionado com sucesso');
      setPaymentModalVisible(false);
      paymentForm.resetFields();
      fetchBudget(); // Reload to get updated payments
    } catch (error) {
      message.error('Erro ao adicionar pagamento');
      console.error('Error:', error);
      if (error.response?.data?.error) {
        message.error('Detalhes: ' + error.response.data.error);
      }
    }
  };

  const handlePrintReceipt = async (paymentId) => {
    try {
      const response = await paymentsAPI.downloadReceipt(id, paymentId);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `recibo_${paymentId}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('Recibo baixado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar recibo');
      console.error('Error:', error);
    }
  };

  const calculateFinancialSummary = () => {
    const totalValue = budget?.total_value || 0;
    const totalPaid = payments
      .filter(p => p.status === 'paid')
      .reduce((sum, p) => sum + p.amount, 0);
    const remainingBalance = totalValue - totalPaid;

    return { totalValue, totalPaid, remainingBalance };
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  };

  const handleDownloadPDF = async () => {
    try {
      const response = await budgetsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `orcamento_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
      console.error('Error:', error);
    }
  };

  const itemColumns = [
    {
      title: 'Descrição',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Quantidade',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 120,
      align: 'center',
    },
    {
      title: 'Valor Unitário',
      dataIndex: 'unit_price',
      key: 'unit_price',
      width: 150,
      align: 'right',
      render: (value) => formatCurrency(value),
    },
    {
      title: 'Total',
      dataIndex: 'total',
      key: 'total',
      width: 150,
      align: 'right',
      render: (value) => formatCurrency(value),
    },
  ];

  if (!budget) {
    return (
      <Card loading={loading}>
        <p>Carregando...</p>
      </Card>
    );
  }

  const statusInfo = statusOptions[budget.status] || { label: budget.status, color: 'default' };

  return (
    <div>
      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>Visualizar Orçamento</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              type="primary"
              danger
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
            >
              Baixar PDF
            </Button>
            <Button
              icon={<EditOutlined />}
              onClick={() => navigate(`/budgets/${id}/edit`)}
            >
              Editar
            </Button>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/budgets')}
            >
              Voltar
            </Button>
          </Space>
        }
        loading={loading}
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="ID">{budget.id}</Descriptions.Item>
          <Descriptions.Item label="Data de Criação">
            {dayjs(budget.created_at).format('DD/MM/YYYY HH:mm')}
          </Descriptions.Item>
          <Descriptions.Item label="Paciente">
            {budget.patient?.name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Profissional">
            {budget.dentist?.name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag color={statusInfo.color}>{statusInfo.label}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Válido Até">
            {budget.valid_until ? dayjs(budget.valid_until).format('DD/MM/YYYY') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Valor Total" span={2}>
            <span style={{ fontSize: 20, fontWeight: 'bold', color: '#1890ff' }}>
              {formatCurrency(budget.total_value)}
            </span>
          </Descriptions.Item>
          {budget.description && (
            <Descriptions.Item label="Descrição" span={2}>
              {budget.description}
            </Descriptions.Item>
          )}
        </Descriptions>

        <Divider>Itens do Orçamento</Divider>

        <Table
          columns={itemColumns}
          dataSource={items}
          rowKey="id"
          pagination={false}
          locale={{ emptyText: 'Nenhum item' }}
          footer={() => (
            <div style={{ textAlign: 'right', fontSize: 18, fontWeight: 'bold' }}>
              Total: {formatCurrency(budget.total_value)}
            </div>
          )}
        />

        {budget.notes && (
          <>
            <Divider>Observações</Divider>
            <p>{budget.notes}</p>
          </>
        )}

        <Divider>Gerenciamento de Pagamentos</Divider>

        {budget.status !== 'approved' ? (
          <Card style={{ marginBottom: 16 }}>
            <div style={{ textAlign: 'center', padding: '20px' }}>
              <p style={{ fontSize: 16, marginBottom: 16 }}>
                Para gerenciar pagamentos, o orçamento precisa estar com status <strong>"Aprovado"</strong>.
              </p>
              <p style={{ color: '#999' }}>
                Status atual: <Tag color={statusOptions[budget.status]?.color}>{statusOptions[budget.status]?.label}</Tag>
              </p>
            </div>
          </Card>
        ) : (
          <>
            <Divider>Resumo Financeiro</Divider>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Valor Total"
                    value={calculateFinancialSummary().totalValue}
                    precision={2}
                    prefix="R$"
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Total Pago"
                    value={calculateFinancialSummary().totalPaid}
                    precision={2}
                    prefix="R$"
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Saldo Restante"
                    value={calculateFinancialSummary().remainingBalance}
                    precision={2}
                    prefix="R$"
                    valueStyle={{
                      color: calculateFinancialSummary().remainingBalance > 0 ? '#ff4d4f' : '#52c41a'
                    }}
                  />
                </Card>
              </Col>
            </Row>

            <Divider>
              <Space>
                Pagamentos Recebidos
                <Button
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={handleAddPayment}
                >
                  Adicionar Pagamento
                </Button>
              </Space>
            </Divider>

            {payments && payments.length > 0 ? (
              <Table
                dataSource={payments}
                rowKey="id"
                columns={[
                  {
                    title: 'Data',
                    dataIndex: 'paid_date',
                    render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
                    width: 120,
                  },
                  {
                    title: 'Parcela',
                    key: 'installment',
                    render: (_, record) =>
                      record.is_installment
                        ? `${record.installment_number}/${record.total_installments}`
                        : '1/1',
                    width: 80,
                  },
                  {
                    title: 'Forma de Pagamento',
                    dataIndex: 'payment_method',
                    render: (method) => {
                      const methodObj = paymentMethods.find(m => m.value === method);
                      return methodObj ? methodObj.label : method;
                    },
                    width: 150,
                  },
                  {
                    title: 'Valor',
                    dataIndex: 'amount',
                    render: (value) => formatCurrency(value),
                    width: 120,
                  },
                  {
                    title: 'Status',
                    dataIndex: 'status',
                    render: (status) => (
                      <Tag color={status === 'paid' ? 'success' : 'warning'}>
                        {status === 'paid' ? 'Pago' : 'Pendente'}
                      </Tag>
                    ),
                    width: 100,
                  },
                  {
                    title: 'Ações',
                    key: 'actions',
                    render: (_, record) => (
                      <Button
                        size="small"
                        icon={<PrinterOutlined />}
                        onClick={() => handlePrintReceipt(record.id)}
                      >
                        Recibo
                      </Button>
                    ),
                    width: 120,
                  },
                ]}
                pagination={false}
                summary={() => (
                  <Table.Summary>
                    <Table.Summary.Row>
                      <Table.Summary.Cell index={0} colSpan={3}>
                        <strong>Total Pago:</strong>
                      </Table.Summary.Cell>
                      <Table.Summary.Cell index={1}>
                        <strong>{formatCurrency(calculateFinancialSummary().totalPaid)}</strong>
                      </Table.Summary.Cell>
                      <Table.Summary.Cell index={2} colSpan={2} />
                    </Table.Summary.Row>
                  </Table.Summary>
                )}
              />
            ) : (
              <p style={{ textAlign: 'center', padding: '20px', color: '#999' }}>
                Nenhum pagamento registrado ainda
              </p>
            )}
          </>
        )}
      </Card>

      {/* Modal para adicionar pagamento */}
      <Modal
        title="Adicionar Pagamento"
        open={paymentModalVisible}
        onCancel={() => {
          setPaymentModalVisible(false);
          paymentForm.resetFields();
        }}
        onOk={() => paymentForm.submit()}
        okText="Salvar"
        cancelText="Cancelar"
        width={600}
      >
        <Form
          form={paymentForm}
          layout="vertical"
          onFinish={handlePaymentSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="amount"
                label="Valor Recebido"
                rules={[{ required: true, message: 'Informe o valor' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  prefix="R$"
                  min={0}
                  step={0.01}
                  precision={2}
                  placeholder="0,00"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="paid_date"
                label="Data do Pagamento"
                rules={[{ required: true, message: 'Informe a data' }]}
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="payment_method"
            label="Forma de Pagamento"
            rules={[{ required: true, message: 'Selecione a forma de pagamento' }]}
          >
            <Select placeholder="Selecione">
              {paymentMethods.map(method => (
                <Select.Option key={method.value} value={method.value}>
                  {method.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="installment_number"
                label="Número da Parcela"
                rules={[{ required: true, message: 'Informe o número da parcela' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  placeholder="1"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="total_installments"
                label="Total de Parcelas"
                initialValue={1}
                rules={[{ required: true, message: 'Informe o total de parcelas' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  placeholder="1"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="notes"
            label="Observações"
          >
            <Input.TextArea rows={3} placeholder="Observações adicionais (opcional)" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default BudgetView;
