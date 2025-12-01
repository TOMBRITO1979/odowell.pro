import { useState, useEffect } from 'react';
import {
  Card, Table, Button, Space, Tag, message, Popconfirm, Select, Typography
} from 'antd';
import {
  FileTextOutlined, FilePdfOutlined, DeleteOutlined, PlusOutlined
} from '@ant-design/icons';
import { consentsAPI, consentTemplatesAPI } from '../../services/api';
import SignatureModal from './SignatureModal';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors } from '../../theme/designSystem';
import dayjs from 'dayjs';

const { Title } = Typography;

const PatientConsents = ({ patient }) => {
  const [consents, setConsents] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [signatureModalVisible, setSignatureModalVisible] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState(null);

  const { canCreate, canDelete } = usePermission();

  useEffect(() => {
    if (patient?.id) {
      fetchConsents();
      fetchTemplates();
    }
  }, [patient]);

  const fetchConsents = async () => {
    try {
      setLoading(true);
      const response = await consentsAPI.getByPatient(patient.id);
      setConsents(response.data.consents || []);
    } catch (error) {
      message.error('Erro ao carregar consentimentos');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const fetchTemplates = async () => {
    try {
      const response = await consentTemplatesAPI.getAll({ active: true });
      setTemplates(response.data.templates || []);
    } catch (error) {
      console.error('Erro ao carregar templates:', error);
    }
  };

  const handleSign = (template) => {
    setSelectedTemplate(template);
    setSignatureModalVisible(true);
  };

  const handleDownloadPDF = async (consentId) => {
    try {
      const response = await consentsAPI.getPDF(consentId);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `termo_consentimento_${consentId}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      message.error('Erro ao baixar PDF');
      console.error(error);
    }
  };

  const handleDelete = async (id) => {
    try {
      await consentsAPI.delete(id);
      message.success('Consentimento excluído');
      fetchConsents();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao excluir consentimento');
    }
  };

  const handleSignatureSuccess = () => {
    fetchConsents();
  };

  const columns = [
    {
      title: 'Template',
      dataIndex: 'template_title',
      key: 'template_title',
      width: '30%',
    },
    {
      title: 'Versão',
      dataIndex: 'template_version',
      key: 'template_version',
      width: '10%',
    },
    {
      title: 'Assinante',
      dataIndex: 'signer_name',
      key: 'signer_name',
      width: '20%',
    },
    {
      title: 'Data',
      dataIndex: 'signed_at',
      key: 'signed_at',
      width: '15%',
      render: (date) => dayjs(date).format('DD/MM/YYYY HH:mm'),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: '10%',
      render: (status) => {
        const colors = {
          active: 'success',
          revoked: 'error',
          expired: 'warning',
        };
        const labels = {
          active: 'Ativo',
          revoked: 'Revogado',
          expired: 'Expirado',
        };
        return <Tag color={colors[status]}>{labels[status]}</Tag>;
      },
    },
    {
      title: 'Ações',
      key: 'actions',
      width: '15%',
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<FilePdfOutlined />}
            onClick={() => handleDownloadPDF(record.id)}
            style={{ color: actionColors.exportPDF }}
          >
            PDF
          </Button>
          {canDelete('clinical_records') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button type="link" danger icon={<DeleteOutlined />}>
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
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={4} style={{ margin: 0 }}>
            <FileTextOutlined /> Termos de Consentimento
          </Title>
          {canCreate('clinical_records') && templates.length > 0 && (
            <Space>
              <Select
                placeholder="Selecione um template"
                style={{ width: 300 }}
                onChange={(value) => {
                  const template = templates.find(t => t.id === value);
                  handleSign(template);
                }}
                value={null}
              >
                {templates.map(template => (
                  <Select.Option key={template.id} value={template.id}>
                    {template.title}
                  </Select.Option>
                ))}
              </Select>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  if (templates.length > 0) {
                    handleSign(templates[0]);
                  }
                }}
                disabled={templates.length === 0}
              >
                Novo Termo
              </Button>
            </Space>
          )}
        </div>

        <Table
          columns={columns}
          dataSource={consents}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 5 }}
          locale={{
            emptyText: 'Nenhum termo de consentimento assinado'
          }}
        />
      </Card>

      <SignatureModal
        visible={signatureModalVisible}
        onClose={() => setSignatureModalVisible(false)}
        patient={patient}
        template={selectedTemplate}
        onSuccess={handleSignatureSuccess}
      />
    </div>
  );
};

export default PatientConsents;
