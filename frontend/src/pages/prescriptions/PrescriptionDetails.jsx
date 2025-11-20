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
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  FilePdfOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { prescriptionsAPI } from '../../services/api';

const { Title, Paragraph, Text } = Typography;

const PrescriptionDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [prescription, setPrescription] = useState(null);

  useEffect(() => {
    fetchPrescription();
  }, [id]);

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
      const response = await prescriptionsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `receita_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
    }
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
            {prescription.status === 'draft' && (
              <Button
                icon={<EditOutlined />}
                onClick={() => navigate(`/prescriptions/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            <Button
              type="primary"
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
            >
              Gerar PDF
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
        </div>
      </Card>

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
