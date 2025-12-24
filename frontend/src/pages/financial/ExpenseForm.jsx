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
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  WalletOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { paymentsAPI } from '../../services/api';
import { shadows } from '../../theme/designSystem';

const { TextArea } = Input;

const ExpenseForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [recurrenceDays, setRecurrenceDays] = useState(0);

  const recurrenceOptions = [
    { value: 0, label: 'Não recorrente' },
    { value: 7, label: 'A cada 7 dias (Semanal)' },
    { value: 15, label: 'A cada 15 dias (Quinzenal)' },
    { value: 30, label: 'A cada 30 dias (Mensal)' },
    { value: 180, label: 'A cada 180 dias (Semestral)' },
    { value: 360, label: 'A cada 360 dias (Anual)' },
  ];

  const categoryOptions = [
    { value: 'salario', label: 'Salário / Funcionários' },
    { value: 'aluguel', label: 'Aluguel' },
    { value: 'luz', label: 'Luz / Energia' },
    { value: 'agua', label: 'Água' },
    { value: 'internet', label: 'Internet / Telefone' },
    { value: 'insumos', label: 'Insumos / Materiais' },
    { value: 'equipamentos', label: 'Equipamentos' },
    { value: 'manutencao', label: 'Manutenção' },
    { value: 'limpeza', label: 'Limpeza' },
    { value: 'marketing', label: 'Marketing' },
    { value: 'impostos', label: 'Impostos / Taxas' },
    { value: 'software', label: 'Software / Sistemas' },
    { value: 'outros', label: 'Outros' },
  ];

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'boleto', label: 'Boleto' },
  ];

  const statusOptions = [
    { value: 'pending', label: 'Pendente' },
    { value: 'paid', label: 'Pago' },
    { value: 'overdue', label: 'Atrasado' },
    { value: 'cancelled', label: 'Cancelado' },
  ];

  useEffect(() => {
    if (id) {
      fetchExpense();
    }
  }, [id]);

  const fetchExpense = async () => {
    setLoading(true);
    try {
      const response = await paymentsAPI.getOne(id);
      const expense = response.data.payment;

      setRecurrenceDays(expense.recurrence_days || 0);

      form.setFieldsValue({
        ...expense,
        recurrence_days: expense.recurrence_days || 0,
        due_date: expense.due_date ? dayjs(expense.due_date) : null,
        paid_date: expense.paid_date ? dayjs(expense.paid_date) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar conta');
      navigate('/expenses');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const recDays = values.recurrence_days || 0;
      const data = {
        ...values,
        type: 'expense', // Sempre despesa
        patient_id: null, // Despesas não precisam de paciente
        is_recurring: recDays > 0,
        recurrence_days: recDays,
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
        message.success('Conta atualizada com sucesso!');
      } else {
        await paymentsAPI.create(data);
        message.success('Conta registrada com sucesso!');
      }
      navigate('/expenses');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar conta'
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
            <WalletOutlined />
            <span>{id ? 'Editar Conta' : 'Nova Conta a Pagar'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/expenses')}
          >
            Voltar
          </Button>
        }
        style={{ boxShadow: shadows.small }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
          initialValues={{
            status: 'pending',
            payment_method: 'pix',
          }}
        >
          <Row gutter={16}>
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

            <Col xs={24} md={8}>
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

            <Col xs={24} md={8}>
              <Form.Item
                name="payment_method"
                label="Forma de Pagamento"
                rules={[
                  { required: true, message: 'Selecione a forma' },
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
            <Col xs={24}>
              <Form.Item
                name="description"
                label="Descrição"
                rules={[
                  { required: true, message: 'Informe a descrição' },
                ]}
              >
                <Input placeholder="Ex: Conta de luz referente a Novembro/2025" />
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

          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item
                name="recurrence_days"
                label="Recorrência"
                initialValue={0}
              >
                <Select
                  placeholder="Selecione a recorrência"
                  onChange={(value) => setRecurrenceDays(value)}
                >
                  {recurrenceOptions.map((option) => (
                    <Select.Option key={option.value} value={option.value}>
                      {option.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
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
              <Button onClick={() => navigate('/expenses')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default ExpenseForm;
