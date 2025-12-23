import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Descriptions,
  message,
  Popconfirm,
  Tag,
  Spin,
  Modal,
  Form,
  Input,
  Alert,
  Divider,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  FilePdfOutlined,
  SafetyCertificateOutlined,
  CheckCircleOutlined,
  LockOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { medicalRecordsAPI, signingAPI, certificatesAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import Odontogram from '../../components/Odontogram';
import { actionColors } from '../../theme/designSystem';

const { Text } = Typography;

const MedicalRecordDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [record, setRecord] = useState(null);
  const [loading, setLoading] = useState(false);
  const { canEdit, canDelete } = usePermission();

  // Digital Signature state
  const [signModalVisible, setSignModalVisible] = useState(false);
  const [signing, setSigning] = useState(false);
  const [hasCertificate, setHasCertificate] = useState(false);
  const [signForm] = Form.useForm();

  const recordTypes = [
    { value: 'anamnesis', label: 'Anamnese', color: 'blue' },
    { value: 'treatment', label: 'Tratamento', color: 'green' },
    { value: 'procedure', label: 'Procedimento', color: 'purple' },
    { value: 'prescription', label: 'Receita', color: 'orange' },
    { value: 'certificate', label: 'Atestado', color: 'red' },
  ];

  useEffect(() => {
    fetchRecord();
    checkCertificate();
  }, [id]);

  const checkCertificate = async () => {
    try {
      const response = await certificatesAPI.getAll();
      const certs = response.data.certificates || [];
      setHasCertificate(certs.some(c => c.active && !c.is_expired));
    } catch (error) {
      setHasCertificate(false);
    }
  };

  const fetchRecord = async () => {
    setLoading(true);
    try {
      const response = await medicalRecordsAPI.getOne(id);
      setRecord(response.data.record);
    } catch (error) {
      message.error('Erro ao carregar prontuário');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      await medicalRecordsAPI.delete(id);
      message.success('Prontuário excluído com sucesso');
      navigate('/medical-records');
    } catch (error) {
      message.error('Erro ao excluir prontuário');
    }
  };

  const handleDownloadPDF = async () => {
    try {
      const response = await medicalRecordsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `prontuario_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
    }
  };

  const handleSign = async (values) => {
    setSigning(true);
    try {
      await signingAPI.signMedicalRecord(id, values.password);
      message.success('Prontuario assinado digitalmente com sucesso!');
      setSignModalVisible(false);
      signForm.resetFields();
      fetchRecord(); // Reload to show signature info
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao assinar prontuario');
    } finally {
      setSigning(false);
    }
  };

  const openSignModal = () => {
    if (!hasCertificate) {
      Modal.info({
        title: 'Certificado Digital Necessario',
        content: (
          <div>
            <p>Voce precisa cadastrar um certificado digital A1 para assinar documentos.</p>
            <p>Acesse o menu <strong>Certificado Digital</strong> para fazer o upload do seu certificado.</p>
          </div>
        ),
        onOk: () => navigate('/certificates'),
      });
      return;
    }
    setSignModalVisible(true);
  };

  const getTypeTag = (type) => {
    const typeObj = recordTypes.find((t) => t.value === type);
    return typeObj ? (
      <Tag color={typeObj.color}>{typeObj.label}</Tag>
    ) : (
      <Tag>{type}</Tag>
    );
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!record) {
    return null;
  }

  return (
    <div>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            <span>Detalhes do Prontuário</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/medical-records')}
            >
              Voltar
            </Button>
            {!record.is_signed && canEdit('medical_records') && (
              <Button
                type="primary"
                icon={<SafetyCertificateOutlined />}
                onClick={openSignModal}
                style={{ backgroundColor: '#52c41a', borderColor: '#52c41a' }}
              >
                Assinar Digitalmente
              </Button>
            )}
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
              style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
            >
              Baixar PDF
            </Button>
            {!record.is_signed && canEdit('medical_records') && (
              <Button
                type="primary"
                icon={<EditOutlined />}
                onClick={() => navigate(`/medical-records/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            {!record.is_signed && canDelete('medical_records') && (
              <Popconfirm
                title="Tem certeza que deseja excluir este prontuário?"
                onConfirm={handleDelete}
                okText="Sim"
                cancelText="Não"
              >
                <Button danger icon={<DeleteOutlined />}>
                  Excluir
                </Button>
              </Popconfirm>
            )}
          </Space>
        }
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Data do Registro">
            {record.created_at
              ? dayjs(record.created_at).format('DD/MM/YYYY HH:mm')
              : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Última Atualização">
            {record.updated_at
              ? dayjs(record.updated_at).format('DD/MM/YYYY HH:mm')
              : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Paciente" span={2}>
            {record.patient?.name || '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Tipo">
            {getTypeTag(record.type)}
          </Descriptions.Item>

          <Descriptions.Item label="Profissional">
            {record.dentist?.name || '-'}
          </Descriptions.Item>

          {record.diagnosis && (
            <Descriptions.Item label="Diagnóstico" span={2}>
              {record.diagnosis}
            </Descriptions.Item>
          )}

          {record.treatment_plan && (
            <Descriptions.Item label="Plano de Tratamento" span={2}>
              {record.treatment_plan}
            </Descriptions.Item>
          )}

          {record.procedure_done && (
            <Descriptions.Item label="Procedimentos Realizados" span={2}>
              {record.procedure_done}
            </Descriptions.Item>
          )}

          {record.materials && (
            <Descriptions.Item label="Materiais Utilizados" span={2}>
              {record.materials}
            </Descriptions.Item>
          )}

          {record.prescription && (
            <Descriptions.Item label="Prescrição" span={2}>
              {record.prescription}
            </Descriptions.Item>
          )}

          {record.certificate && (
            <Descriptions.Item label="Atestado" span={2}>
              {record.certificate}
            </Descriptions.Item>
          )}

          {record.evolution && (
            <Descriptions.Item label="Evolução" span={2}>
              {record.evolution}
            </Descriptions.Item>
          )}

          {record.odontogram && (
            <>
              {/* Força uma nova linha completa */}
              <Descriptions.Item label="" span={2} style={{ height: 0, padding: 0, margin: 0, lineHeight: 0 }} />
              <Descriptions.Item label="Odontograma" span={2} contentStyle={{ padding: 0, display: 'flex', justifyContent: 'center' }}>
                <Odontogram value={record.odontogram} readOnly={true} />
              </Descriptions.Item>
            </>
          )}

          {record.notes && (
            <Descriptions.Item label="Notas Adicionais" span={2}>
              {record.notes}
            </Descriptions.Item>
          )}
        </Descriptions>

        {/* Signature Info */}
        {record.is_signed && (
          <>
            <Divider orientation="left">
              <Space>
                <CheckCircleOutlined style={{ color: '#52c41a' }} />
                Assinatura Digital
              </Space>
            </Divider>
            <Alert
              message="Documento Assinado Digitalmente"
              description={
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Assinado por">
                    {record.signed_by_name} (CRO: {record.signed_by_cro})
                  </Descriptions.Item>
                  <Descriptions.Item label="Data/Hora">
                    {dayjs(record.signed_at).format('DD/MM/YYYY HH:mm:ss')}
                  </Descriptions.Item>
                  <Descriptions.Item label="Certificado">
                    {record.certificate_thumbprint}
                  </Descriptions.Item>
                  <Descriptions.Item label="Hash SHA-256">
                    <Text code style={{ fontSize: 11 }}>{record.signature_hash}</Text>
                  </Descriptions.Item>
                </Descriptions>
              }
              type="success"
              showIcon
              icon={<SafetyCertificateOutlined />}
            />
          </>
        )}
      </Card>

      {/* Sign Modal */}
      <Modal
        title={
          <Space>
            <SafetyCertificateOutlined />
            <span>Assinar Prontuario Digitalmente</span>
          </Space>
        }
        open={signModalVisible}
        onCancel={() => {
          setSignModalVisible(false);
          signForm.resetFields();
        }}
        footer={null}
        width={450}
      >
        <Alert
          message="Assinatura Digital ICP-Brasil"
          description="Ao assinar, o documento sera marcado com seu certificado digital e nao podera mais ser editado ou excluido."
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Form
          form={signForm}
          layout="vertical"
          onFinish={handleSign}
        >
          <Form.Item
            label="Senha do Certificado"
            name="password"
            rules={[{ required: true, message: 'Informe a senha do certificado' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Digite a senha do seu certificado"
              size="large"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 24 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => {
                setSignModalVisible(false);
                signForm.resetFields();
              }}>
                Cancelar
              </Button>
              <Button
                type="primary"
                htmlType="submit"
                loading={signing}
                icon={<SafetyCertificateOutlined />}
                style={{ backgroundColor: '#52c41a', borderColor: '#52c41a' }}
              >
                Assinar Documento
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default MedicalRecordDetails;
