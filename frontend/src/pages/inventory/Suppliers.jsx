import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Card,
  message,
  Popconfirm,
  Modal,
  Form,
  Input,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { suppliersAPI } from '../../services/api';

const { TextArea } = Input;

const Suppliers = () => {
  const [loading, setLoading] = useState(false);
  const [suppliers, setSuppliers] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSupplier, setEditingSupplier] = useState(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchSuppliers();
  }, []);

  const fetchSuppliers = async () => {
    setLoading(true);
    try {
      const response = await suppliersAPI.getAll();
      setSuppliers(response.data.suppliers || []);
    } catch (error) {
      message.error('Erro ao carregar fornecedores');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await suppliersAPI.delete(id);
      message.success('Fornecedor excluído com sucesso');
      fetchSuppliers();
    } catch (error) {
      message.error('Erro ao excluir fornecedor');
    }
  };

  const showModal = (supplier = null) => {
    if (supplier) {
      setEditingSupplier(supplier);
      form.setFieldsValue(supplier);
    } else {
      setEditingSupplier(null);
      form.resetFields();
    }
    setModalVisible(true);
  };

  const handleCancel = () => {
    setModalVisible(false);
    setEditingSupplier(null);
    form.resetFields();
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      if (editingSupplier) {
        await suppliersAPI.update(editingSupplier.id, values);
        message.success('Fornecedor atualizado com sucesso!');
      } else {
        await suppliersAPI.create(values);
        message.success('Fornecedor criado com sucesso!');
      }
      setModalVisible(false);
      form.resetFields();
      fetchSuppliers();
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar fornecedor'
      );
    } finally {
      setLoading(false);
    }
  };

  const formatCNPJ = (cnpj) => {
    if (!cnpj) return '-';
    return cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
  };

  const formatPhone = (phone) => {
    if (!phone) return '-';
    return phone.replace(/^(\d{2})(\d{4,5})(\d{4})$/, '($1) $2-$3');
  };

  const columns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'CNPJ',
      dataIndex: 'cnpj',
      key: 'cnpj',
      render: (cnpj) => formatCNPJ(cnpj),
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: 'Telefone',
      dataIndex: 'phone',
      key: 'phone',
      render: (phone) => formatPhone(phone),
    },
    {
      title: 'Cidade',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => showModal(record)}
            title="Editar"
          />
          <Popconfirm
            title="Tem certeza que deseja excluir?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title="Excluir"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <TeamOutlined />
            <span>Fornecedores</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => showModal()}
          >
            Novo Fornecedor
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={suppliers}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 20 }}
        />
      </Card>

      <Modal
        title={editingSupplier ? 'Editar Fornecedor' : 'Novo Fornecedor'}
        open={modalVisible}
        onCancel={handleCancel}
        footer={null}
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col xs={24} md={16}>
              <Form.Item
                name="name"
                label="Nome do Fornecedor"
                rules={[
                  { required: true, message: 'Informe o nome do fornecedor' },
                  { max: 200, message: 'Nome muito longo' },
                ]}
              >
                <Input placeholder="Ex: Dental Supplies Ltda" />
              </Form.Item>
            </Col>

            <Col xs={24} md={8}>
              <Form.Item
                name="cnpj"
                label="CNPJ"
                rules={[
                  { max: 18, message: 'CNPJ inválido' },
                  {
                    pattern: /^\d{2}\.\d{3}\.\d{3}\/\d{4}-\d{2}$|^\d{14}$/,
                    message: 'Formato inválido (ex: 12.345.678/0001-90)',
                  },
                ]}
              >
                <Input placeholder="12.345.678/0001-90" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="email"
                label="Email"
                rules={[
                  { type: 'email', message: 'Email inválido' },
                  { max: 100, message: 'Email muito longo' },
                ]}
              >
                <Input placeholder="contato@fornecedor.com.br" />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="phone"
                label="Telefone"
                rules={[
                  { max: 20, message: 'Telefone muito longo' },
                ]}
              >
                <Input placeholder="(11) 98765-4321" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="address"
            label="Endereço"
            rules={[
              { max: 500, message: 'Endereço muito longo' },
            ]}
          >
            <TextArea
              rows={2}
              placeholder="Endereço completo, cidade, estado..."
            />
          </Form.Item>

          <Form.Item name="notes" label="Observações">
            <TextArea
              rows={3}
              placeholder="Informações adicionais sobre o fornecedor..."
              maxLength={1000}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
              >
                Salvar
              </Button>
              <Button onClick={handleCancel}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Suppliers;
