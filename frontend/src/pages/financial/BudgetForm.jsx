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
  Divider,
  Space,
  DatePicker,
  InputNumber,
  Table,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  DollarOutlined,
  PlusOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { budgetsAPI, patientsAPI, usersAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

const { TextArea } = Input;

const BudgetForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const [dentists, setDentists] = useState([]);
  const [items, setItems] = useState([]);
  const [totalValue, setTotalValue] = useState(0);

  const statusOptions = [
    { value: 'pending', label: 'Pendente' },
    { value: 'approved', label: 'Aprovado' },
    { value: 'rejected', label: 'Rejeitado' },
    { value: 'expired', label: 'Expirado' },
  ];

  useEffect(() => {
    fetchPatients();
    fetchDentists();
    if (id) {
      fetchBudget();
    }
  }, [id]);

  useEffect(() => {
    calculateTotal();
  }, [items]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    }
  };

  const fetchDentists = async () => {
    try {
      const response = await usersAPI.getAll();
      // Filtrar apenas dentistas e admins (profissionais que podem ser responsáveis)
      const professionals = (response.data.users || []).filter(
        u => u.role === 'dentist' || u.role === 'admin'
      );
      setDentists(professionals);
    } catch (error) {
    }
  };

  const fetchBudget = async () => {
    setLoading(true);
    try {
      const response = await budgetsAPI.getOne(id);
      const budget = response.data.budget;

      form.setFieldsValue({
        ...budget,
        valid_until: budget.valid_until ? dayjs(budget.valid_until) : null,
      });

      // Parse items from JSON string
      if (budget.items) {
        try {
          const parsedItems = JSON.parse(budget.items);
          setItems(parsedItems);
        } catch (e) {
        }
      }
    } catch (error) {
      message.error('Erro ao carregar orçamento');
    } finally {
      setLoading(false);
    }
  };

  const calculateTotal = () => {
    const total = items.reduce((sum, item) => {
      return sum + (parseFloat(item.quantity || 0) * parseFloat(item.unit_price || 0));
    }, 0);
    setTotalValue(total);
    form.setFieldValue('total_value', total);
  };

  const addItem = () => {
    setItems([
      ...items,
      {
        id: Date.now(),
        description: '',
        quantity: 1,
        unit_price: 0,
        total: 0,
      },
    ]);
  };

  const removeItem = (itemId) => {
    setItems(items.filter((item) => item.id !== itemId));
  };

  const updateItem = (itemId, field, value) => {
    const updatedItems = items.map((item) => {
      if (item.id === itemId) {
        const updated = { ...item, [field]: value };
        updated.total = parseFloat(updated.quantity || 0) * parseFloat(updated.unit_price || 0);
        return updated;
      }
      return item;
    });
    setItems(updatedItems);
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  };

  const itemColumns = [
    {
      title: 'Descrição do Procedimento/Item',
      dataIndex: 'description',
      key: 'description',
      render: (text, record) => (
        <Input
          value={text}
          onChange={(e) => updateItem(record.id, 'description', e.target.value)}
          placeholder="Ex: Limpeza, Restauração, etc."
        />
      ),
    },
    {
      title: 'Qtd',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 100,
      render: (value, record) => (
        <InputNumber
          value={value}
          onChange={(val) => updateItem(record.id, 'quantity', val)}
          min={1}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: 'Valor Unit.',
      dataIndex: 'unit_price',
      key: 'unit_price',
      width: 130,
      render: (value, record) => (
        <InputNumber
          value={value}
          onChange={(val) => updateItem(record.id, 'unit_price', val)}
          min={0}
          step={0.01}
          prefix="R$"
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: 'Total',
      dataIndex: 'total',
      key: 'total',
      width: 130,
      render: (value) => formatCurrency(value),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 80,
      align: 'center',
      render: (_, record) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() => removeItem(record.id)}
          title="Remover"
        />
      ),
    },
  ];

  const onFinish = async (values) => {
    if (items.length === 0) {
      message.error('Adicione pelo menos um item ao orçamento');
      return;
    }

    setLoading(true);
    try {
      const data = {
        ...values,
        total_value: totalValue,
        items: JSON.stringify(items),
      };

      // Converter data para formato ISO
      if (values.valid_until) {
        data.valid_until = values.valid_until.toISOString();
      }

      if (id) {
        await budgetsAPI.update(id, data);
        message.success('Orçamento atualizado com sucesso!');
      } else {
        await budgetsAPI.create(data);
        message.success('Orçamento criado com sucesso!');
      }
      navigate('/budgets');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar orçamento'
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
            <span>{id ? 'Editar Orçamento' : 'Novo Orçamento'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/budgets')}
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
            <Col xs={24} md={12}>
              <Form.Item
                name="patient_id"
                label="Paciente"
                rules={[
                  { required: true, message: 'Selecione o paciente' },
                ]}
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
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

            <Col xs={24} md={12}>
              <Form.Item
                name="dentist_id"
                label="Profissional Responsável"
                rules={[
                  { required: true, message: 'Selecione o profissional' },
                ]}
                initialValue={user?.id}
              >
                <Select
                  placeholder="Selecione o profissional"
                  showSearch
                  filterOption={(input, option) =>
                    option.children
                      .toLowerCase()
                      .includes(input.toLowerCase())
                  }
                >
                  {dentists.map((dentist) => (
                    <Select.Option key={dentist.id} value={dentist.id}>
                      {dentist.name}
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
                name="valid_until"
                label="Válido Até"
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
              placeholder="Descrição geral do orçamento..."
            />
          </Form.Item>

          <Divider>Itens do Orçamento</Divider>

          <Table
            columns={itemColumns}
            dataSource={items}
            rowKey="id"
            pagination={false}
            locale={{ emptyText: 'Nenhum item adicionado' }}
            footer={() => (
              <div>
                <Button
                  type="dashed"
                  onClick={addItem}
                  icon={<PlusOutlined />}
                  block
                >
                  Adicionar Item
                </Button>
                <div style={{ marginTop: 16, textAlign: 'right', fontSize: 18, fontWeight: 'bold' }}>
                  Valor Total: {formatCurrency(totalValue)}
                </div>
              </div>
            )}
            scroll={{ x: 800 }}
          />

          <Divider />

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
              <Button onClick={() => navigate('/budgets')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default BudgetForm;
