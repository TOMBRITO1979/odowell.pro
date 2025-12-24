import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Tag,
  message,
  Row,
  Col,
  Descriptions,
  Progress,
  Statistic,
  Table,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  DatePicker,
  Popconfirm,
  Divider,
  Typography,
  Alert,
} from 'antd';
import {
  ArrowLeftOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FilePdfOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  UserOutlined,
  CalendarOutlined,
  MedicineBoxOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { treatmentsAPI, treatmentPaymentsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';
import { usePermission } from '../../contexts/AuthContext';

const { Title, Text } = Typography;
const { TextArea } = Input;

const TreatmentDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const { canCreate, canEdit, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [treatment, setTreatment] = useState(null);
  const [payments, setPayments] = useState([]);
  const [paymentModalVisible, setPaymentModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'check', label: 'Cheque' },
  ];

  const statusOptions = [
    { value: 'in_progress', label: 'Em Andamento', color: statusColors.inProgress, icon: <ClockCircleOutlined /> },
    { value: 'completed', label: 'Concluído', color: statusColors.success, icon: <CheckCircleOutlined /> },
    { value: 'cancelled', label: 'Cancelado', color: statusColors.error, icon: <CloseCircleOutlined /> },
  ];

  useEffect(() => {
    if (id) {
      fetchTreatment();
    }
  }, [id]);

  const fetchTreatment = async () => {
    setLoading(true);
    try {
      const response = await treatmentsAPI.getOne(id);
      setTreatment(response.data.treatment);
      setPayments(response.data.treatment.treatment_payments || []);
    } catch (error) {
      message.error('Erro ao carregar tratamento');
    } finally {
      setLoading(false);
    }
  };

  const handleAddPayment = async (values) => {
    try {
      const data = {
        treatment_id: parseInt(id),
        amount: values.amount,
        payment_method: values.payment_method,
        installment_number: values.installment_number,
        notes: values.notes,
        paid_date: values.paid_date ? values.paid_date.format('YYYY-MM-DD') : null,
      };

      await treatmentPaymentsAPI.create(data);
      message.success('Pagamento registrado com sucesso!');
      setPaymentModalVisible(false);
      form.resetFields();
      fetchTreatment();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao registrar pagamento');
    }
  };

  const handleDeletePayment = async (paymentId) => {
    try {
      await treatmentPaymentsAPI.delete(paymentId);
      message.success('Pagamento excluído com sucesso');
      fetchTreatment();
    } catch (error) {
      message.error('Erro ao excluir pagamento');
    }
  };

  const handleDownloadReceipt = async (paymentId) => {
    try {
      const response = await treatmentPaymentsAPI.downloadReceipt(paymentId);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `recibo_${paymentId}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('Recibo baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar recibo');
    }
  };

  const handleUpdateTreatment = async (values) => {
    try {
      await treatmentsAPI.update(id, {
        total_installments: values.total_installments,
        status: values.status,
        notes: values.notes,
        expected_end_date: values.expected_end_date ? values.expected_end_date.format('YYYY-MM-DD') : null,
      });
      message.success('Tratamento atualizado com sucesso!');
      setEditModalVisible(false);
      fetchTreatment();
    } catch (error) {
      message.error('Erro ao atualizar tratamento');
    }
  };

  const openEditModal = () => {
    editForm.setFieldsValue({
      total_installments: treatment.total_installments,
      status: treatment.status,
      notes: treatment.notes,
      expected_end_date: treatment.expected_end_date ? dayjs(treatment.expected_end_date) : null,
    });
    setEditModalVisible(true);
  };

  const openPaymentModal = () => {
    const nextInstallment = payments.length + 1;
    // Calculate remaining value
    const remaining = treatment.total_value - treatment.paid_value;
    // Use remaining value if less than installment value, otherwise use installment value
    const suggestedAmount = remaining > 0
      ? Math.min(remaining, treatment.installment_value || remaining)
      : treatment.installment_value;

    form.setFieldsValue({
      amount: suggestedAmount > 0 ? suggestedAmount : treatment.installment_value,
      installment_number: nextInstallment,
      paid_date: dayjs(),
    });
    setPaymentModalVisible(true);
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value || 0);
  };

  const getStatusTag = (status) => {
    const statusObj = statusOptions.find((s) => s.value === status);
    return statusObj ? (
      <Tag color={statusObj.color} icon={statusObj.icon} style={{ fontSize: 14, padding: '4px 12px' }}>
        {statusObj.label}
      </Tag>
    ) : (
      <Tag>{status}</Tag>
    );
  };

  const getPaymentMethodLabel = (method) => {
    const methodObj = paymentMethods.find((m) => m.value === method);
    return methodObj ? methodObj.label : method;
  };

  const paymentColumns = [
    {
      title: 'Data',
      dataIndex: 'paid_date',
      key: 'paid_date',
      width: 110,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Parcela',
      dataIndex: 'installment_number',
      key: 'installment_number',
      width: 80,
      render: (num) => `${num}/${treatment?.total_installments || '-'}`,
    },
    {
      title: 'Valor',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (value) => (
        <Text strong style={{ color: statusColors.success }}>
          {formatCurrency(value)}
        </Text>
      ),
    },
    {
      title: 'Forma de Pagamento',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 150,
      render: (method) => getPaymentMethodLabel(method),
    },
    {
      title: 'Recibo',
      dataIndex: 'receipt_number',
      key: 'receipt_number',
      width: 120,
    },
    {
      title: 'Recebido por',
      dataIndex: ['received_by', 'name'],
      key: 'received_by',
      ellipsis: true,
    },
    {
      title: 'Observações',
      dataIndex: 'notes',
      key: 'notes',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<FilePdfOutlined />}
            onClick={() => handleDownloadReceipt(record.id)}
            title="Baixar Recibo"
            style={{ color: actionColors.exportPDF }}
          />
          {canDelete('payments') && (
            <Popconfirm
              title="Tem certeza que deseja excluir este pagamento?"
              description="O valor será removido do total pago"
              onConfirm={() => handleDeletePayment(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                icon={<DeleteOutlined />}
                title="Excluir"
                style={{ color: actionColors.delete }}
              />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  if (!treatment) {
    return (
      <Card loading={loading}>
        Carregando...
      </Card>
    );
  }

  const remainingValue = treatment.total_value - treatment.paid_value;
  const progressPercent = treatment.total_value > 0
    ? Math.round((treatment.paid_value / treatment.total_value) * 100)
    : 0;

  return (
    <div>
      <Card
        title={
          <Space>
            <MedicineBoxOutlined />
            <span>Detalhes do Tratamento</span>
          </Space>
        }
        extra={
          <Space>
            {canEdit('budgets') && (
              <Button icon={<EditOutlined />} onClick={openEditModal}>
                Editar
              </Button>
            )}
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/treatments')}
            >
              Voltar
            </Button>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 16 }}
      >
        {/* Patient and Status */}
        <Row gutter={[16, 16]}>
          <Col xs={24} md={16}>
            <Descriptions bordered column={{ xs: 1, sm: 2 }}>
              <Descriptions.Item label={<><UserOutlined /> Paciente</>} span={2}>
                <Text strong style={{ fontSize: 16 }}>{treatment.patient?.name}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="CPF">
                {treatment.patient?.cpf || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Telefone">
                {treatment.patient?.phone || '-'}
              </Descriptions.Item>
              <Descriptions.Item label={<><MedicineBoxOutlined /> Profissional</>}>
                {treatment.dentist?.name || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Descrição" span={2}>
                {treatment.description || '-'}
              </Descriptions.Item>
              <Descriptions.Item label={<><CalendarOutlined /> Data Início</>}>
                {dayjs(treatment.start_date).format('DD/MM/YYYY')}
              </Descriptions.Item>
              <Descriptions.Item label="Previsão de Término">
                {treatment.expected_end_date
                  ? dayjs(treatment.expected_end_date).format('DD/MM/YYYY')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Status" span={2}>
                {getStatusTag(treatment.status)}
              </Descriptions.Item>
            </Descriptions>
          </Col>

          {/* Financial Summary */}
          <Col xs={24} md={8}>
            <Card
              style={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                borderRadius: 12,
              }}
            >
              <Space direction="vertical" style={{ width: '100%' }} size="large">
                <Statistic
                  title={<Text style={{ color: 'rgba(255,255,255,0.8)' }}>Valor Total</Text>}
                  value={treatment.total_value}
                  precision={2}
                  valueStyle={{ color: '#fff', fontSize: 28 }}
                  formatter={(value) => formatCurrency(value)}
                />

                <Progress
                  percent={progressPercent}
                  strokeColor="#81C784"
                  trailColor="rgba(255,255,255,0.3)"
                  format={(percent) => (
                    <span style={{ color: '#fff' }}>{percent}%</span>
                  )}
                />

                <Row gutter={16}>
                  <Col span={12}>
                    <Statistic
                      title={<Text style={{ color: 'rgba(255,255,255,0.8)', fontSize: 12 }}>Pago</Text>}
                      value={treatment.paid_value}
                      precision={2}
                      valueStyle={{ color: '#81C784', fontSize: 18 }}
                      formatter={(value) => formatCurrency(value)}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title={<Text style={{ color: 'rgba(255,255,255,0.8)', fontSize: 12 }}>Restante</Text>}
                      value={remainingValue}
                      precision={2}
                      valueStyle={{ color: remainingValue > 0 ? '#E57373' : '#81C784', fontSize: 18 }}
                      formatter={(value) => formatCurrency(value)}
                    />
                  </Col>
                </Row>

                <Text style={{ color: 'rgba(255,255,255,0.8)' }}>
                  Parcelas: {payments.length} de {treatment.total_installments}
                  {treatment.installment_value > 0 && (
                    <> ({formatCurrency(treatment.installment_value)} cada)</>
                  )}
                </Text>
              </Space>
            </Card>
          </Col>
        </Row>

        {treatment.notes && (
          <>
            <Divider />
            <Descriptions>
              <Descriptions.Item label="Observações">
                {treatment.notes}
              </Descriptions.Item>
            </Descriptions>
          </>
        )}
      </Card>

      {/* Payments List */}
      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>Pagamentos Recebidos</span>
          </Space>
        }
        extra={
          canCreate('payments') && treatment.status === 'in_progress' && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={openPaymentModal}
              style={{
                backgroundColor: statusColors.success,
                borderColor: statusColors.success,
              }}
            >
              Registrar Pagamento
            </Button>
          )
        }
        style={{ boxShadow: shadows.small }}
      >
        {treatment.status === 'completed' && (
          <Alert
            message="Tratamento Quitado"
            description="Este tratamento foi totalmente pago e finalizado."
            type="success"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        {treatment.status === 'cancelled' && (
          <Alert
            message="Tratamento Cancelado"
            description="Este tratamento foi cancelado."
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Table
          columns={paymentColumns}
          dataSource={payments}
          rowKey="id"
          loading={loading}
          pagination={false}
          scroll={{ x: 900 }}
          locale={{ emptyText: 'Nenhum pagamento registrado' }}
          summary={() => (
            payments.length > 0 && (
              <Table.Summary fixed>
                <Table.Summary.Row>
                  <Table.Summary.Cell index={0} colSpan={2}>
                    <Text strong>Total</Text>
                  </Table.Summary.Cell>
                  <Table.Summary.Cell index={2}>
                    <Text strong style={{ color: statusColors.success }}>
                      {formatCurrency(treatment.paid_value)}
                    </Text>
                  </Table.Summary.Cell>
                  <Table.Summary.Cell index={3} colSpan={5} />
                </Table.Summary.Row>
              </Table.Summary>
            )
          )}
        />
      </Card>

      {/* Add Payment Modal */}
      <Modal
        title="Registrar Pagamento"
        open={paymentModalVisible}
        onCancel={() => {
          setPaymentModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddPayment}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="amount"
                label="Valor"
                rules={[{ required: true, message: 'Informe o valor' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  prefix="R$"
                  precision={2}
                  min={0.01}
                  placeholder="0,00"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="installment_number"
                label="Parcela"
                rules={[{ required: true, message: 'Informe a parcela' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  max={treatment?.total_installments || 999}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="payment_method"
                label="Forma de Pagamento"
                rules={[{ required: true, message: 'Selecione a forma de pagamento' }]}
              >
                <Select placeholder="Selecione">
                  {paymentMethods.map((method) => (
                    <Select.Option key={method.value} value={method.value}>
                      {method.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="paid_date"
                label="Data do Pagamento"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea rows={3} placeholder="Observações opcionais..." />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                Registrar
              </Button>
              <Button onClick={() => setPaymentModalVisible(false)}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit Treatment Modal */}
      <Modal
        title="Editar Tratamento"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          editForm.resetFields();
        }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdateTreatment}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="total_installments"
                label="Total de Parcelas"
                rules={[{ required: true, message: 'Informe o total de parcelas' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  max={999}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="status"
                label="Status"
                rules={[{ required: true, message: 'Selecione o status' }]}
              >
                <Select placeholder="Selecione">
                  {statusOptions.map((status) => (
                    <Select.Option key={status.value} value={status.value}>
                      {status.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="expected_end_date"
            label="Previsão de Término"
          >
            <DatePicker
              style={{ width: '100%' }}
              format="DD/MM/YYYY"
              placeholder="Selecione a data"
            />
          </Form.Item>

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea rows={3} placeholder="Observações opcionais..." />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                Salvar
              </Button>
              <Button onClick={() => setEditModalVisible(false)}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TreatmentDetails;
