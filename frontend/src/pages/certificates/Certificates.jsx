import React, { useState, useEffect } from 'react';
import {
  Card, Table, Button, Modal, Form, Input, Upload, message, Tag, Space, Typography, Alert, Popconfirm, Tooltip
} from 'antd';
import {
  SafetyCertificateOutlined, UploadOutlined, DeleteOutlined, CheckCircleOutlined,
  CloseCircleOutlined, LockOutlined, KeyOutlined, ExclamationCircleOutlined
} from '@ant-design/icons';
import { certificatesAPI } from '../../services/api';
import dayjs from 'dayjs';

const { Text, Paragraph } = Typography;

const Certificates = () => {
  const [certificates, setCertificates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploadLoading, setUploadLoading] = useState(false);
  const [form] = Form.useForm();
  const [fileList, setFileList] = useState([]);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    fetchCertificates();
  }, []);

  const fetchCertificates = async () => {
    setLoading(true);
    try {
      const response = await certificatesAPI.getAll();
      setCertificates(response.data.certificates || []);
    } catch (error) {
      message.error('Erro ao carregar certificados');
    } finally {
      setLoading(false);
    }
  };

  const handleUpload = async (values) => {
    if (fileList.length === 0) {
      message.error('Selecione um arquivo de certificado');
      return;
    }

    setUploadLoading(true);
    try {
      const formData = new FormData();
      formData.append('certificate', fileList[0].originFileObj);
      formData.append('name', values.name);
      formData.append('password', values.password);

      await certificatesAPI.upload(formData);
      message.success('Certificado cadastrado com sucesso!');
      setUploadModalVisible(false);
      form.resetFields();
      setFileList([]);
      fetchCertificates();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cadastrar certificado');
    } finally {
      setUploadLoading(false);
    }
  };

  const handleActivate = async (id) => {
    try {
      await certificatesAPI.activate(id);
      message.success('Certificado ativado com sucesso!');
      fetchCertificates();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao ativar certificado');
    }
  };

  const handleDelete = async (id) => {
    try {
      await certificatesAPI.delete(id);
      message.success('Certificado excluido com sucesso!');
      fetchCertificates();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao excluir certificado');
    }
  };

  const getValidityStatus = (record) => {
    if (record.is_expired) return { color: 'red', text: 'Expirado' };
    if (record.days_until_expiry <= 30) return { color: 'orange', text: `${record.days_until_expiry} dias` };
    return { color: 'green', text: `${record.days_until_expiry} dias` };
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (certificates.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum certificado cadastrado</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {certificates.map((record) => {
          const validity = getValidityStatus(record);
          return (
            <Card
              key={record.id}
              size="small"
              style={{ borderLeft: `4px solid ${record.active ? '#1890ff' : '#d9d9d9'}` }}
              bodyStyle={{ padding: '12px' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                <div>
                  <div style={{ fontWeight: 600, fontSize: '15px' }}>{record.name}</div>
                  <Text type="secondary" style={{ fontSize: 12 }}>{record.subject_cn}</Text>
                </div>
                {record.active ? (
                  <Tag color="blue" icon={<CheckCircleOutlined />}>Ativo</Tag>
                ) : (
                  <Tag icon={<CloseCircleOutlined />}>Inativo</Tag>
                )}
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
                <div><strong>CPF:</strong> {record.subject_cpf || '-'}</div>
                <div><strong>Validade:</strong><br /><Tag color={validity.color}>{validity.text}</Tag></div>
                <div style={{ gridColumn: '1 / -1' }}><strong>Emissor:</strong> {record.issuer_cn}</div>
                <div><strong>Ultimo uso:</strong> {record.last_used_at ? dayjs(record.last_used_at).format('DD/MM/YYYY') : 'Nunca'}</div>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
                {!record.active && !record.is_expired && (
                  <Button type="primary" size="small" icon={<CheckCircleOutlined />} onClick={() => handleActivate(record.id)}>Ativar</Button>
                )}
                <Popconfirm title="Excluir certificado?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Nao">
                  <Button danger size="small" icon={<DeleteOutlined />}>Excluir</Button>
                </Popconfirm>
              </div>
            </Card>
          );
        })}
      </div>
    );
  };

  const columns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{text}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.subject_cn}</Text>
        </Space>
      ),
    },
    {
      title: 'CPF',
      dataIndex: 'subject_cpf',
      key: 'subject_cpf',
      render: (cpf) => cpf || '-',
    },
    {
      title: 'Emissor',
      dataIndex: 'issuer_cn',
      key: 'issuer_cn',
      ellipsis: true,
    },
    {
      title: 'Validade',
      key: 'validity',
      render: (_, record) => {
        const isExpired = record.is_expired;
        const daysLeft = record.days_until_expiry;

        if (isExpired) {
          return <Tag color="red">Expirado</Tag>;
        }
        if (daysLeft <= 30) {
          return (
            <Tooltip title={`Expira em ${dayjs(record.not_after).format('DD/MM/YYYY')}`}>
              <Tag color="orange">{daysLeft} dias restantes</Tag>
            </Tooltip>
          );
        }
        return (
          <Tooltip title={`Expira em ${dayjs(record.not_after).format('DD/MM/YYYY')}`}>
            <Tag color="green">{daysLeft} dias restantes</Tag>
          </Tooltip>
        );
      },
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => (
        <Space>
          {record.active ? (
            <Tag color="blue" icon={<CheckCircleOutlined />}>Ativo</Tag>
          ) : (
            <Tag icon={<CloseCircleOutlined />}>Inativo</Tag>
          )}
          {!record.is_valid && <Tag color="red">Invalido</Tag>}
        </Space>
      ),
    },
    {
      title: 'Ultimo Uso',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY HH:mm') : 'Nunca',
    },
    {
      title: 'Acoes',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {!record.active && !record.is_expired && (
            <Button
              type="primary"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => handleActivate(record.id)}
            >
              Ativar
            </Button>
          )}
          <Popconfirm
            title="Excluir certificado?"
            description="Esta acao nao pode ser desfeita."
            onConfirm={() => handleDelete(record.id)}
            okText="Excluir"
            cancelText="Cancelar"
            okButtonProps={{ danger: true }}
          >
            <Button danger size="small" icon={<DeleteOutlined />}>
              Excluir
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const uploadProps = {
    maxCount: 1,
    accept: '.pfx,.p12',
    beforeUpload: (file) => {
      const isValid = file.name.endsWith('.pfx') || file.name.endsWith('.p12');
      if (!isValid) {
        message.error('Apenas arquivos .pfx ou .p12 sao aceitos');
        return Upload.LIST_IGNORE;
      }
      return false;
    },
    fileList,
    onChange: ({ fileList: newFileList }) => {
      setFileList(newFileList);
    },
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <SafetyCertificateOutlined />
            <span>Certificados Digitais</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<UploadOutlined />}
            onClick={() => setUploadModalVisible(true)}
          >
            Adicionar Certificado
          </Button>
        }
      >
        <Alert
          message="Assinatura Digital ICP-Brasil"
          description={
            <Paragraph style={{ marginBottom: 0 }}>
              Cadastre seu certificado digital A1 (arquivo .pfx/.p12) para assinar digitalmente receitas e prontuarios.
              O certificado e armazenado de forma criptografada e a senha e solicitada apenas no momento da assinatura.
            </Paragraph>
          }
          type="info"
          showIcon
          icon={<KeyOutlined />}
          style={{ marginBottom: 24 }}
        />

        {isMobile ? renderMobileCards() : (
          <Table
            columns={columns}
            dataSource={certificates}
            rowKey="id"
            loading={loading}
            locale={{ emptyText: 'Nenhum certificado cadastrado' }}
            pagination={false}
          />
        )}
      </Card>

      {/* Upload Modal */}
      <Modal
        title={
          <Space>
            <SafetyCertificateOutlined />
            <span>Adicionar Certificado Digital</span>
          </Space>
        }
        open={uploadModalVisible}
        onCancel={() => {
          setUploadModalVisible(false);
          form.resetFields();
          setFileList([]);
        }}
        footer={null}
        width={500}
      >
        <Alert
          message="Informacoes Importantes"
          description={
            <ul style={{ marginBottom: 0, paddingLeft: 20 }}>
              <li>Use certificados A1 ICP-Brasil validos (.pfx ou .p12)</li>
              <li>A senha sera usada para validar e criptografar o certificado</li>
              <li>O certificado sera armazenado de forma segura (AES-256)</li>
              <li>A senha nao e armazenada - sera solicitada a cada assinatura</li>
            </ul>
          }
          type="warning"
          showIcon
          icon={<ExclamationCircleOutlined />}
          style={{ marginBottom: 24 }}
        />

        <Form
          form={form}
          layout="vertical"
          onFinish={handleUpload}
        >
          <Form.Item
            label="Nome do Certificado"
            name="name"
            rules={[{ required: true, message: 'Informe um nome para identificar o certificado' }]}
          >
            <Input placeholder="Ex: Certificado Dr. Joao Silva" />
          </Form.Item>

          <Form.Item
            label="Arquivo do Certificado (.pfx ou .p12)"
            required
          >
            <Upload.Dragger {...uploadProps}>
              <p className="ant-upload-drag-icon">
                <SafetyCertificateOutlined style={{ fontSize: 48, color: '#1890ff' }} />
              </p>
              <p className="ant-upload-text">Clique ou arraste o arquivo do certificado</p>
              <p className="ant-upload-hint">Apenas arquivos .pfx ou .p12</p>
            </Upload.Dragger>
          </Form.Item>

          <Form.Item
            label="Senha do Certificado"
            name="password"
            rules={[{ required: true, message: 'Informe a senha do certificado' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Senha do certificado A1"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 24 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => {
                setUploadModalVisible(false);
                form.resetFields();
                setFileList([]);
              }}>
                Cancelar
              </Button>
              <Button type="primary" htmlType="submit" loading={uploadLoading}>
                Cadastrar Certificado
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Certificates;
