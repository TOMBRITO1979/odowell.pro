import { useState, useEffect } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Switch, Space,
  message, Popconfirm, Tag, Typography, Card
} from 'antd';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, FileTextOutlined, EyeOutlined
} from '@ant-design/icons';
import { consentTemplatesAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { TextArea } = Input;
const { Title } = Typography;

const ConsentTemplates = () => {
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [previewModalVisible, setPreviewModalVisible] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState(null);
  const [previewTemplate, setPreviewTemplate] = useState(null);
  const [form] = Form.useForm();

  const { canCreate, canEdit, canDelete } = usePermission();

  const consentTypes = [
    { value: 'treatment', label: 'Tratamento', color: statusColors.inProgress },
    { value: 'procedure', label: 'Procedimento', color: statusColors.success },
    { value: 'anesthesia', label: 'Anestesia', color: statusColors.pending },
    { value: 'data_privacy', label: 'LGPD/Privacidade', color: '#a855f7' },
    { value: 'general', label: 'Geral', color: statusColors.cancelled },
  ];

  useEffect(() => {
    fetchTemplates();
  }, []);

  const fetchTemplates = async () => {
    try {
      setLoading(true);
      const response = await consentTemplatesAPI.getAll();
      setTemplates(response.data.templates || []);
    } catch (error) {
      message.error('Erro ao carregar templates');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingTemplate(null);
    form.resetFields();
    form.setFieldsValue({ active: true, is_default: false });
    setModalVisible(true);
  };

  const handleEdit = (record) => {
    setEditingTemplate(record);
    form.setFieldsValue({
      title: record.title,
      type: record.type,
      version: record.version,
      content: record.content,
      description: record.description,
      active: record.active,
      is_default: record.is_default,
    });
    setModalVisible(true);
  };

  const handlePreview = (record) => {
    setPreviewTemplate(record);
    setPreviewModalVisible(true);
  };

  const handleDownloadPDF = async (id) => {
    try {
      const response = await consentTemplatesAPI.getPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `termo_consentimento_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
    }
  };

  const handleDelete = async (id) => {
    try {
      await consentTemplatesAPI.delete(id);
      message.success('Template excluído com sucesso');
      fetchTemplates();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao excluir template');
    }
  };

  const handleSubmit = async (values) => {
    try {
      if (editingTemplate) {
        await consentTemplatesAPI.update(editingTemplate.id, values);
        message.success('Template atualizado com sucesso');
      } else {
        await consentTemplatesAPI.create(values);
        message.success('Template criado com sucesso');
      }
      setModalVisible(false);
      form.resetFields();
      fetchTemplates();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao salvar template');
    }
  };

  const getTypeInfo = (type) => {
    return consentTypes.find(t => t.value === type) || consentTypes[4];
  };

  const columns = [
    {
      title: 'Título',
      dataIndex: 'title',
      key: 'title',
      width: '25%',
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: '15%',
      render: (type) => {
        const typeInfo = getTypeInfo(type);
        return <Tag color={typeInfo.color}>{typeInfo.label}</Tag>;
      },
    },
    {
      title: 'Versão',
      dataIndex: 'version',
      key: 'version',
      width: '10%',
    },
    {
      title: 'Status',
      key: 'status',
      width: '15%',
      render: (_, record) => (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {record.active && <Tag color={statusColors.success} style={{ margin: 0 }}>Ativo</Tag>}
          {!record.active && <Tag color={statusColors.cancelled} style={{ margin: 0 }}>Inativo</Tag>}
          {record.is_default && <Tag color={statusColors.inProgress} style={{ margin: 0 }}>Padrão</Tag>}
        </div>
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: '15%',
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => handlePreview(record)}
            style={{ color: actionColors.view }}
          >
            Ver
          </Button>
          {canEdit('clinical_records') && (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              style={{ color: actionColors.edit }}
            >
              Editar
            </Button>
          )}
          {canDelete('clinical_records') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button type="link" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>
                Excluir
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card style={{ boxShadow: shadows.small }}>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12 }}>
          <Title level={4} style={{ margin: 0 }}>
            <FileTextOutlined /> Termos de Consentimento
          </Title>
          {canCreate('clinical_records') && (
            <Button
              icon={<PlusOutlined />}
              onClick={handleCreate}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              <span className="btn-text-desktop">Novo Template</span>
              <span className="btn-text-mobile">Novo</span>
            </Button>
          )}
        </div>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={templates}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10 }}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>

      <Modal
        title={editingTemplate ? 'Editar Template' : 'Novo Template'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={800}
        okText="Salvar"
        cancelText="Cancelar"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="Título"
            name="title"
            rules={[{ required: true, message: 'Informe o título' }]}
          >
            <Input placeholder="Ex: Consentimento para Tratamento Odontológico" />
          </Form.Item>

          <Form.Item
            label="Tipo"
            name="type"
            rules={[{ required: true, message: 'Selecione o tipo' }]}
          >
            <Select placeholder="Selecione o tipo">
              {consentTypes.map(type => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="Versão"
            name="version"
            rules={[{ required: true, message: 'Informe a versão' }]}
          >
            <Input placeholder="Ex: 1.0.0" />
          </Form.Item>

          <Form.Item
            label="Conteúdo"
            name="content"
            rules={[{ required: true, message: 'Informe o conteúdo' }]}
          >
            <TextArea
              rows={8}
              placeholder="Conteúdo do termo de consentimento..."
            />
          </Form.Item>

          <Form.Item
            label="Descrição"
            name="description"
          >
            <TextArea
              rows={3}
              placeholder="Descrição opcional do template..."
            />
          </Form.Item>

          <Space>
            <Form.Item
              name="active"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <Switch /> <span style={{ marginLeft: 8 }}>Ativo</span>
            </Form.Item>

            <Form.Item
              name="is_default"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <Switch /> <span style={{ marginLeft: 8 }}>Template Padrão</span>
            </Form.Item>
          </Space>
        </Form>
      </Modal>

      <Modal
        title="Visualizar Template"
        open={previewModalVisible}
        onCancel={() => setPreviewModalVisible(false)}
        footer={[
          <Button
            key="pdf"
            onClick={() => handleDownloadPDF(previewTemplate?.id)}
            disabled={!previewTemplate}
            style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
          >
            Gerar PDF
          </Button>,
          <Button key="close" onClick={() => setPreviewModalVisible(false)}>
            Fechar
          </Button>
        ]}
        width={800}
      >
        {previewTemplate && (
          <div>
            <Title level={3}>{previewTemplate.title}</Title>
            <Space style={{ marginBottom: 16 }}>
              <Tag color={getTypeInfo(previewTemplate.type).color}>
                {getTypeInfo(previewTemplate.type).label}
              </Tag>
              <Tag>Versão: {previewTemplate.version}</Tag>
              {previewTemplate.is_default && <Tag color="blue">Padrão</Tag>}
            </Space>
            {previewTemplate.description && (
              <div style={{ marginBottom: 16 }}>
                <Typography.Text type="secondary">
                  {previewTemplate.description}
                </Typography.Text>
              </div>
            )}
            <div style={{
              padding: '16px',
              backgroundColor: '#f5f5f5',
              borderRadius: '4px',
              whiteSpace: 'pre-wrap',
              minHeight: '200px'
            }}>
              {previewTemplate.content}
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ConsentTemplates;
