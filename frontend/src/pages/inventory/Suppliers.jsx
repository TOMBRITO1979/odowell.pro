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
  Upload,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TeamOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { suppliersAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { TextArea } = Input;

const Suppliers = () => {
  const [loading, setLoading] = useState(false);
  const [suppliers, setSuppliers] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSupplier, setEditingSupplier] = useState(null);
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [form] = Form.useForm();

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

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

  const handleExportCSV = async () => {
    try {
      const response = await suppliersAPI.exportCSV('');
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `fornecedores_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
    }
  };

  const handleExportPDF = async () => {
    try {
      const response = await suppliersAPI.exportPDF('');
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `fornecedores_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
    }
  };

  const handleImportCSV = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const response = await suppliersAPI.importCSV(formData);
      message.success(response.data.message);

      if (response.data.errors && response.data.errors.length > 0) {
        Modal.warning({
          title: 'Avisos durante a importação',
          content: (
            <div>
              <p>{response.data.imported} fornecedores importados com sucesso.</p>
              <p>Erros encontrados:</p>
              <ul>
                {response.data.errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            </div>
          ),
          width: 600,
        });
      }

      setUploadModalVisible(false);
      fetchSuppliers();
    } catch (error) {
      message.error('Erro ao importar CSV');
    } finally {
      setUploading(false);
    }

    return false;
  };

  const formatCNPJ = (cnpj) => {
    if (!cnpj) return '-';
    return cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
  };

  const formatPhone = (phone) => {
    if (!phone) return '-';
    return phone.replace(/^(\d{2})(\d{4,5})(\d{4})$/, '($1) $2-$3');
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (suppliers.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum fornecedor encontrado</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {suppliers.map((record) => (
          <Card key={record.id} size="small" style={{ borderLeft: '4px solid #1890ff' }} bodyStyle={{ padding: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
              <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.name}</div>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
              <div><strong>CNPJ:</strong><br />{formatCNPJ(record.cnpj)}</div>
              <div><strong>Telefone:</strong><br />{formatPhone(record.phone)}</div>
              <div style={{ gridColumn: '1 / -1' }}><strong>Email:</strong> {record.email || '-'}</div>
              <div style={{ gridColumn: '1 / -1' }}><strong>Cidade:</strong> {record.address || '-'}</div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <Button type="text" size="small" icon={<EditOutlined />} onClick={() => showModal(record)} style={{ color: actionColors.edit }}>Editar</Button>
              <Popconfirm title="Tem certeza?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                <Button type="text" size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
              </Popconfirm>
            </div>
          </Card>
        ))}
      </div>
    );
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
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => showModal(record)}
            title="Editar"
            style={{ color: actionColors.edit }}
          />
          <Popconfirm
            title="Tem certeza que deseja excluir?"
            onConfirm={() => handleDelete(record.id)}
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
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          isMobile ? null : (
            <Space>
              <TeamOutlined />
              <span>Fornecedores</span>
            </Space>
          )
        }
        extra={
          isMobile ? null : (
            <Space>
              <Button
                icon={<FileExcelOutlined />}
                onClick={handleExportCSV}
                style={{
                  backgroundColor: actionColors.exportExcel,
                  borderColor: actionColors.exportExcel,
                  color: '#fff'
                }}
              >
                Exportar CSV
              </Button>
              <Button
                icon={<FilePdfOutlined />}
                onClick={handleExportPDF}
                style={{
                  backgroundColor: actionColors.exportPDF,
                  borderColor: actionColors.exportPDF,
                  color: '#fff'
                }}
              >
                Gerar PDF
              </Button>
              <Button
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
                style={{
                  backgroundColor: actionColors.view,
                  borderColor: actionColors.view,
                  color: '#fff'
                }}
              >
                Importar CSV
              </Button>
              <Button
                icon={<PlusOutlined />}
                onClick={() => showModal()}
                style={{
                  backgroundColor: actionColors.create,
                  borderColor: actionColors.create,
                  color: '#fff'
                }}
              >
                Novo Fornecedor
              </Button>
            </Space>
          )
        }
      >
        {isMobile && (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px', marginBottom: '16px', paddingBottom: '16px', borderBottom: '1px solid #f0f0f0' }}>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '8px', width: '100%', maxWidth: '280px' }}>
              <Button
                icon={<FileExcelOutlined />}
                onClick={handleExportCSV}
                size="small"
                style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}
              >
                CSV
              </Button>
              <Button
                icon={<FilePdfOutlined />}
                onClick={handleExportPDF}
                size="small"
                style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
              >
                PDF
              </Button>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '8px', width: '100%', maxWidth: '280px' }}>
              <Button
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
                size="small"
                style={{ backgroundColor: actionColors.view, borderColor: actionColors.view, color: '#fff' }}
              >
                Importar
              </Button>
              <Button
                icon={<PlusOutlined />}
                onClick={() => showModal()}
                size="small"
                style={{ backgroundColor: actionColors.create, borderColor: actionColors.create, color: '#fff' }}
              >
                Novo
              </Button>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '8px' }}>
              <TeamOutlined style={{ fontSize: '18px' }} />
              <span style={{ fontSize: '16px', fontWeight: 600 }}>Fornecedores</span>
            </div>
          </div>
        )}
        {isMobile ? renderMobileCards() : (
          <div style={{ overflowX: 'auto' }}>
            <Table
              columns={columns}
              dataSource={suppliers}
              rowKey="id"
              loading={loading}
              pagination={{ pageSize: 20 }}
              scroll={{ x: 'max-content' }}
            />
          </div>
        )}
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

      <Modal
        title="Importar Fornecedores via CSV"
        open={uploadModalVisible}
        onCancel={() => setUploadModalVisible(false)}
        footer={null}
      >
        <div style={{ marginBottom: 16 }}>
          <p><strong>Formato do CSV:</strong></p>
          <p>O arquivo deve conter as seguintes colunas (COM cabeçalho):</p>
          <ol>
            <li>Nome (obrigatório)</li>
            <li>CNPJ</li>
            <li>Email</li>
            <li>Telefone</li>
            <li>Endereço</li>
            <li>Cidade</li>
            <li>Estado</li>
            <li>CEP</li>
            <li>Ativo (Sim/Não)</li>
            <li>Observações</li>
          </ol>
          <p><strong>Exemplo:</strong></p>
          <code>Nome,CNPJ,Email,Telefone,Endereço,Cidade,Estado,CEP,Ativo,Observações</code>
          <br />
          <code>Dental Supplies,12.345.678/0001-90,contato@dental.com.br,(11)98765-4321,Rua A 123,São Paulo,SP,01234-567,Sim,Fornecedor principal</code>
        </div>

        <Upload
          accept=".csv"
          beforeUpload={handleImportCSV}
          showUploadList={false}
        >
          <Button
            icon={<UploadOutlined />}
            loading={uploading}
            block
            type="primary"
          >
            {uploading ? 'Importando...' : 'Selecionar arquivo CSV'}
          </Button>
        </Upload>
      </Modal>
    </div>
  );
};

export default Suppliers;
