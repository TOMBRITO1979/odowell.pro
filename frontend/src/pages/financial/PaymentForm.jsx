import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  message,
  Select,
  Row,
  Col,
  Space,
  DatePicker,
  InputNumber,
  Switch,
  Divider,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { paymentsAPI, patientsAPI, budgetsAPI } from '../../services/api';

const { TextArea } = Input;

const PaymentForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const [budgets, setBudgets] = useState([]);
  const [isInstallment, setIsInstallment] = useState(false);
  const [isInsurance, setIsInsurance] = useState(false);

  const paymentTypes = [
    { value: 'income', label: 'Receita' },
    { value: 'expense', label: 'Despesa' },
  ];

  const categoryOptions = [
    { value: 'treatment', label: 'Tratamento' },
    { value: 'material', label: 'Material' },
    { value: 'salary', label: 'Salário' },
    { value: 'rent', label: 'Aluguel' },
    { value: 'utilities', label: 'Utilidades' },
    { value: 'equipment', label: 'Equipamento' },
    { value: 'other', label: 'Outro' },
  ];

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'insurance', label: 'Convênio' },
  ];

  const statusOptions = [
    { value: 'pending', label: 'Pendente' },
    { value: 'paid', label: 'Pago' },
    { value: 'overdue', label: 'Atrasado' },
    { value: 'cancelled', label: 'Cancelado' },
  ];

  useEffect(() => {
    fetchPatients();
    fetchBudgets();
    if (id) {
      fetchPayment();
    }
  }, [id]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    }
  };

  const fetchBudgets = async () => {
    try {
      const response = await budgetsAPI.getAll({ page: 1, page_size: 1000, status: 'approved' });
      setBudgets(response.data.budgets || []);
    } catch (error) {
    }
  };

  const fetchPayment = async () => {
    setLoading(true);
    try {
      const response = await paymentsAPI.getOne(id);
      const payment = response.data.payment;

      setIsInstallment(payment.is_installment);
      setIsInsurance(payment.is_insurance);

      form.setFieldsValue({
        ...payment,
        due_date: payment.due_date ? dayjs(payment.due_date) : null,
        paid_date: payment.paid_date ? dayjs(payment.paid_date) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar pagamento');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        is_installment: isInstallment,
        is_insurance: isInsurance,
      };

      // Converter datas para formato ISO
      if (values.due_date) {
        data.due_date = values.due_date.toISOString();
      }
      if (values.paid_date) {
        data.paid_date = values.paid_date.toISOString();
      }

      if (id) {
        await paymentsAPI.update(id, data);
        message.success('Pagamento atualizado com sucesso!');
      } else {
        await paymentsAPI.create(data);
        message.success('Pagamento criado com sucesso!');
      }
      navigate('/payments');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar pagamento'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>{id ? 'Editar Pagamento' : 'Novo Pagamento'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/payments')}
          >
            Voltar
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item
                name="patient_id"
                label="Paciente (Opcional)"
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
                  allowClear
                  filterOption={(input, option) =>
                    option.children
                      .toLowerCase()
                      .includes(input.toLowerCase())
                  }
                >
                  {patients.map((patient) => (
                    <Select.Option key={patient.id} value={patient.id}>
                      {patient.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="type"
                label="Tipo"
                rules={[
                  { required: true, message: 'Selecione o tipo' },
                ]}
                initialValue="income"
              >
                <Select placeholder="Selecione o tipo">
                  {paymentTypes.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      {type.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="category"
                label="Categoria"
                rules={[
                  { required: true, message: 'Selecione a categoria' },
                ]}
              >
                <Select placeholder="Selecione a categoria">
                  {categoryOptions.map((cat) => (
                    <Select.Option key={cat.value} value={cat.value}>
                      {cat.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="budget_id"
                label="Orçamento Relacionado (Opcional)"
              >
                <Select
                  placeholder="Selecione um orçamento"
                  showSearch
                  allowClear
                  filterOption={(input, option) =>
                    option.children
                      .toLowerCase()
                      .includes(input.toLowerCase())
                  }
                >
                  {budgets.map((budget) => (
                    <Select.Option key={budget.id} value={budget.id}>
                      {budget.patient?.name} - R$ {budget.total_value}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="amount"
                label="Valor"
                rules={[
                  { required: true, message: 'Informe o valor' },
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  prefix="R$"
                  min={0}
                  step={0.01}
                  precision={2}
                  placeholder="0.00"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="payment_method"
                label="Método de Pagamento"
                rules={[
                  { required: true, message: 'Selecione o método' },
                ]}
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
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item
                name="status"
                label="Status"
                rules={[
                  { required: true, message: 'Selecione o status' },
                ]}
                initialValue="pending"
              >
                <Select placeholder="Selecione o status">
                  {statusOptions.map((status) => (
                    <Select.Option key={status.value} value={status.value}>
                      {status.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="due_date"
                label="Data de Vencimento"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="paid_date"
                label="Data de Pagamento"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="Descrição">
            <TextArea
              rows={3}
              placeholder="Descrição do pagamento..."
            />
          </Form.Item>

          <Divider>Opções Adicionais</Divider>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item label="Parcelado">
                <Switch
                  checked={isInstallment}
                  onChange={setIsInstallment}
                  checkedChildren="Sim"
                  unCheckedChildren="Não"
                />
              </Form.Item>

              {isInstallment && (
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="installment_number"
                      label="Parcela Nº"
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
                    >
                      <InputNumber
                        style={{ width: '100%' }}
                        min={1}
                        placeholder="12"
                      />
                    </Form.Item>
                  </Col>
                </Row>
              )}
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="Convênio">
                <Switch
                  checked={isInsurance}
                  onChange={setIsInsurance}
                  checkedChildren="Sim"
                  unCheckedChildren="Não"
                />
              </Form.Item>

              {isInsurance && (
                <Form.Item
                  name="insurance_name"
                  label="Nome do Convênio"
                >
                  <Input placeholder="Nome do convênio..." />
                </Form.Item>
              )}
            </Col>
          </Row>

          <Form.Item name="notes" label="Observações">
            <TextArea
              rows={3}
              placeholder="Observações adicionais..."
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                icon={<SaveOutlined />}
              >
                Salvar
              </Button>
              <Button onClick={() => navigate('/payments')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default PaymentForm;
