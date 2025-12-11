import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Descriptions,
  Tag,
  message,
  Divider,
  Typography,
  Row,
  Col,
  Modal,
  Form,
  Input,
  Alert,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  FilePdfOutlined,
  SafetyCertificateOutlined,
  CheckCircleOutlined,
  LockOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { prescriptionsAPI, signingAPI, certificatesAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors } from '../../theme/designSystem';

const { Title, Paragraph, Text } = Typography;

const PrescriptionDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [prescription, setPrescription] = useState(null);
  const { canEdit } = usePermission();

  // Digital Signature state
  const [signModalVisible, setSignModalVisible] = useState(false);
  const [signing, setSigning] = useState(false);
  const [hasCertificate, setHasCertificate] = useState(false);
  const [signForm] = Form.useForm();

  useEffect(() => {
    fetchPrescription();
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

  const fetchPrescription = async () => {
    setLoading(true);
    try {
      const response = await prescriptionsAPI.getOne(id);
      setPrescription(response.data.prescription);
    } catch (error) {
      message.error('Erro ao carregar receita');
    } finally {
      setLoading(false);
    }
  };

  const handleDownloadPDF = async () => {
    try {
      // If signed, download signed PDF
      const response = prescription?.is_signed
        ? await signingAPI.downloadSignedPrescriptionPDF(id)
        : await prescriptionsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      const filename = prescription?.is_signed ? `receita_assinada_${id}.pdf` : `receita_${id}.pdf`;
      link.setAttribute('download', filename);
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
      await signingAPI.signPrescription(id, values.password);
      message.success('Receita assinada digitalmente com sucesso!');
      setSignModalVisible(false);
      signForm.resetFields();
      fetchPrescription(); // Reload to show signature info
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao assinar receita');
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

  const getStatusTag = (status) => {
    const colors = {
      draft: 'default',
      issued: 'success',
      cancelled: 'error',
    };
    const labels = {
      draft: 'Rascunho',
      issued: 'Emitido',
      cancelled: 'Cancelado',
    };
    return <Tag color={colors[status]}>{labels[status]}</Tag>;
  };

  const getTypeLabel = (type) => {
    const labels = {
      prescription: 'Receita',
      medical_report: 'Laudo Médico',
      certificate: 'Atestado',
      referral: 'Encaminhamento',
    };
    return labels[type] || type;
  };

  if (!prescription) {
    return <div>Carregando...</div>;
  }

  return (
    <div>
      <Card
        loading={loading}
        title={
          <Space>
            <span>{getTypeLabel(prescription.type)}</span>
            {getStatusTag(prescription.status)}
          </Space>
        }
        extra={
          <Space>
            {prescription.status === 'draft' && canEdit('prescriptions') && (
              <Button
                icon={<EditOutlined />}
                onClick={() => navigate(`/prescriptions/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            {!prescription.is_signed && canEdit('prescriptions') && (
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
              {prescription.is_signed ? 'Baixar PDF Assinado' : 'Gerar PDF'}
            </Button>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/prescriptions')}
            >
              Voltar
            </Button>
          </Space>
        }
      >
        {/* Print view */}
        <div className="print-only" style={{ padding: '40px' }}>
          <div style={{ textAlign: 'center', marginBottom: '30px' }}>
            <Title level={3}>{prescription.clinic_name}</Title>
            <Text>{prescription.clinic_address}</Text>
            <br />
            <Text>Tel: {prescription.clinic_phone}</Text>
          </div>

          <Divider />

          <div style={{ marginBottom: '30px' }}>
            <Title level={4}>{prescription.title || getTypeLabel(prescription.type)}</Title>
          </div>

          <div style={{ marginBottom: '20px' }}>
            <Text strong>Paciente: </Text>
            <Text>{prescription.patient?.name}</Text>
          </div>

          {prescription.diagnosis && (
            <div style={{ marginBottom: '20px' }}>
              <Text strong>Diagnóstico: </Text>
              <Paragraph>{prescription.diagnosis}</Paragraph>
            </div>
          )}

          {prescription.medications && (
            <div style={{ marginBottom: '20px' }}>
              <Text strong>Medicamentos:</Text>
              <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
                {prescription.medications}
              </Paragraph>
            </div>
          )}

          <div style={{ marginBottom: '30px' }}>
            <Text strong>Conteúdo:</Text>
            <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
              {prescription.content}
            </Paragraph>
          </div>

          {prescription.valid_until && (
            <div style={{ marginBottom: '30px' }}>
              <Text>Válido até: {dayjs(prescription.valid_until).format('DD/MM/YYYY')}</Text>
            </div>
          )}

          <div style={{ marginTop: '60px' }}>
            <Row>
              <Col span={12}>
                <div>
                  <Text>{dayjs(prescription.issued_at || prescription.created_at).format('DD/MM/YYYY')}</Text>
                </div>
              </Col>
              <Col span={12} style={{ textAlign: 'right' }}>
                <div style={{ borderTop: '1px solid #000', paddingTop: '5px', marginTop: '40px' }}>
                  <Text strong>{prescription.dentist_name}</Text>
                  <br />
                  <Text>CRO: {prescription.dentist_cro}</Text>
                </div>
              </Col>
            </Row>
          </div>
        </div>

        {/* Screen view */}
        <div className="screen-only">
          <Descriptions bordered column={2}>
            <Descriptions.Item label="Paciente" span={2}>
              {prescription.patient?.name}
            </Descriptions.Item>
            <Descriptions.Item label="Tipo">
              {getTypeLabel(prescription.type)}
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              {getStatusTag(prescription.status)}
            </Descriptions.Item>
            <Descriptions.Item label="Data de Criação">
              {dayjs(prescription.created_at).format('DD/MM/YYYY HH:mm')}
            </Descriptions.Item>
            {prescription.issued_at && (
              <Descriptions.Item label="Data de Emissão">
                {dayjs(prescription.issued_at).format('DD/MM/YYYY HH:mm')}
              </Descriptions.Item>
            )}
            {prescription.valid_until && (
              <Descriptions.Item label="Válido Até" span={2}>
                {dayjs(prescription.valid_until).format('DD/MM/YYYY')}
              </Descriptions.Item>
            )}
            {prescription.title && (
              <Descriptions.Item label="Título" span={2}>
                {prescription.title}
              </Descriptions.Item>
            )}
          </Descriptions>

          <Divider orientation="left">Informações da Clínica</Divider>
          <Descriptions bordered column={1}>
            <Descriptions.Item label="Nome">{prescription.clinic_name}</Descriptions.Item>
            <Descriptions.Item label="Endereço">{prescription.clinic_address}</Descriptions.Item>
            <Descriptions.Item label="Telefone">{prescription.clinic_phone}</Descriptions.Item>
          </Descriptions>

          <Divider orientation="left">Informações do Dentista</Divider>
          <Descriptions bordered column={2}>
            <Descriptions.Item label="Nome">{prescription.dentist_name}</Descriptions.Item>
            <Descriptions.Item label="CRO">{prescription.dentist_cro}</Descriptions.Item>
          </Descriptions>

          {prescription.diagnosis && (
            <>
              <Divider orientation="left">Diagnóstico</Divider>
              <Paragraph>{prescription.diagnosis}</Paragraph>
            </>
          )}

          {prescription.medications && (
            <>
              <Divider orientation="left">Medicamentos</Divider>
              <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
                {prescription.medications}
              </Paragraph>
            </>
          )}

          <Divider orientation="left">Conteúdo</Divider>
          <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
            {prescription.content}
          </Paragraph>

          {prescription.notes && (
            <>
              <Divider orientation="left">Observações Internas</Divider>
              <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
                {prescription.notes}
              </Paragraph>
            </>
          )}

          {prescription.print_count > 0 && (
            <>
              <Divider />
              <Text type="secondary">
                Este documento foi impresso {prescription.print_count} vez(es).
                {prescription.printed_at && ` Última impressão: ${dayjs(prescription.printed_at).format('DD/MM/YYYY HH:mm')}`}
              </Text>
            </>
          )}

          {/* Signature Info */}
          {prescription.is_signed && (
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
                      {prescription.signed_by_name} (CRO: {prescription.signed_by_cro})
                    </Descriptions.Item>
                    <Descriptions.Item label="Data/Hora">
                      {dayjs(prescription.signed_at).format('DD/MM/YYYY HH:mm:ss')}
                    </Descriptions.Item>
                    <Descriptions.Item label="Certificado">
                      {prescription.certificate_thumbprint}
                    </Descriptions.Item>
                    <Descriptions.Item label="Hash SHA-256">
                      <Text code style={{ fontSize: 11 }}>{prescription.signature_hash}</Text>
                    </Descriptions.Item>
                  </Descriptions>
                }
                type="success"
                showIcon
                icon={<SafetyCertificateOutlined />}
              />
            </>
          )}
        </div>
      </Card>

      {/* Sign Modal */}
      <Modal
        title={
          <Space>
            <SafetyCertificateOutlined />
            <span>Assinar Receita Digitalmente</span>
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
          description="Ao assinar, o documento sera marcado com seu certificado digital e nao podera mais ser editado."
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

      <style>{`
        @media print {
          .screen-only { display: none !important; }
          .ant-card-head, .ant-card-extra, button { display: none !important; }
          .print-only { display: block !important; }
        }
        @media screen {
          .print-only { display: none !important; }
          .screen-only { display: block !important; }
        }
      `}</style>
    </div>
  );
};

export default PrescriptionDetails;
