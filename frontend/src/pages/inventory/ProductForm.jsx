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
  InputNumber,
  Switch,
  DatePicker,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  InboxOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { productsAPI, suppliersAPI } from '../../services/api';

const { TextArea } = Input;

const ProductForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [suppliers, setSuppliers] = useState([]);
  const [active, setActive] = useState(true);

  const categoryOptions = [
    { value: 'material', label: 'Material' },
    { value: 'medicine', label: 'Medicamento' },
    { value: 'equipment', label: 'Equipamento' },
    { value: 'consumable', label: 'Consumível' },
  ];

  const unitOptions = [
    { value: 'un', label: 'Unidade (un)' },
    { value: 'cx', label: 'Caixa (cx)' },
    { value: 'ml', label: 'Mililitro (ml)' },
    { value: 'g', label: 'Grama (g)' },
    { value: 'kg', label: 'Quilograma (kg)' },
    { value: 'l', label: 'Litro (l)' },
    { value: 'pct', label: 'Pacote (pct)' },
  ];

  useEffect(() => {
    fetchSuppliers();
    if (id) {
      fetchProduct();
    }
  }, [id]);

  const fetchSuppliers = async () => {
    try {
      const response = await suppliersAPI.getAll();
      setSuppliers(response.data.suppliers || []);
    } catch (error) {
      console.error('Error fetching suppliers:', error);
    }
  };

  const fetchProduct = async () => {
    setLoading(true);
    try {
      const response = await productsAPI.getOne(id);
      const product = response.data.product;

      setActive(product.active);

      form.setFieldsValue({
        ...product,
        expiration_date: product.expiration_date ? dayjs(product.expiration_date) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar produto');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        active: active,
        expiration_date: values.expiration_date
          ? values.expiration_date.format('YYYY-MM-DD')
          : null,
      };

      if (id) {
        await productsAPI.update(id, data);
        message.success('Produto atualizado com sucesso!');
      } else {
        await productsAPI.create(data);
        message.success('Produto criado com sucesso!');
      }
      navigate('/products');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar produto'
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
            <InboxOutlined />
            <span>{id ? 'Editar Produto' : 'Novo Produto'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/products')}
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
                name="name"
                label="Nome do Produto"
                rules={[
                  { required: true, message: 'Informe o nome do produto' },
                  { max: 200, message: 'Nome muito longo' },
                ]}
              >
                <Input placeholder="Ex: Resina composta Z350" />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="code"
                label="Código"
                rules={[
                  { max: 50, message: 'Código muito longo' },
                ]}
              >
                <Input placeholder="Ex: RES-001" />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="barcode"
                label="Código de Barras"
                rules={[
                  { max: 50, message: 'Código de barras muito longo' },
                ]}
              >
                <Input placeholder="EAN-13" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item
                name="category"
                label="Categoria"
                rules={[
                  { required: true, message: 'Selecione a categoria' },
                ]}
              >
                <Select placeholder="Selecione">
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
                name="supplier_id"
                label="Fornecedor"
              >
                <Select
                  placeholder="Selecione o fornecedor"
                  showSearch
                  allowClear
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {suppliers.map((supplier) => (
                    <Select.Option key={supplier.id} value={supplier.id}>
                      {supplier.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="unit"
                label="Unidade de Medida"
                rules={[
                  { required: true, message: 'Selecione a unidade' },
                ]}
              >
                <Select placeholder="Selecione">
                  {unitOptions.map((unit) => (
                    <Select.Option key={unit.value} value={unit.value}>
                      {unit.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="Descrição">
            <TextArea
              rows={3}
              placeholder="Descrição detalhada do produto..."
              maxLength={1000}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col xs={24} md={6}>
              <Form.Item
                name="quantity"
                label="Quantidade em Estoque"
                rules={[
                  { required: true, message: 'Informe a quantidade' },
                ]}
                initialValue={0}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  precision={0}
                  placeholder="0"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="minimum_stock"
                label="Estoque Mínimo"
                rules={[
                  { required: true, message: 'Informe o estoque mínimo' },
                ]}
                initialValue={0}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  precision={0}
                  placeholder="0"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="cost_price"
                label="Preço de Custo (R$)"
                rules={[
                  { required: true, message: 'Informe o preço de custo' },
                ]}
                initialValue={0}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={0.01}
                  precision={2}
                  placeholder="0.00"
                  prefix="R$"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={6}>
              <Form.Item
                name="sale_price"
                label="Preço de Venda (R$)"
                rules={[
                  { required: true, message: 'Informe o preço de venda' },
                ]}
                initialValue={0}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={0.01}
                  precision={2}
                  placeholder="0.00"
                  prefix="R$"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="expiration_date"
                label="Data de Validade"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item label="Produto Ativo">
                <Switch
                  checked={active}
                  onChange={setActive}
                  checkedChildren="Sim"
                  unCheckedChildren="Não"
                />
              </Form.Item>
            </Col>
          </Row>

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
              <Button onClick={() => navigate('/products')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default ProductForm;
