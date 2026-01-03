import React, { useState, useEffect } from 'react';
import {
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Empty,
  Spin,
  message,
  Modal,
  Descriptions,
  Divider,
  Select,
  Row,
  Col,
} from 'antd';
import {
  FileTextOutlined,
  EyeOutlined,
  UserOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  SafetyCertificateOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';

const { Title, Text, Paragraph } = Typography;

const typeLabels = {
  anamnesis: 'Anamnese',
  treatment: 'Tratamento',
  procedure: 'Procedimento',
  prescription: 'Prescricao',
  certificate: 'Atestado',
  evolution: 'Evolucao',
};

const typeColors = {
  anamnesis: 'purple',
  treatment: 'blue',
  procedure: 'green',
  prescription: 'orange',
  certificate: 'cyan',
  evolution: 'geekblue',
};

const PatientMedicalRecords = () => {
  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState([]);
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [filterType, setFilterType] = useState('');

  useEffect(() => {
    fetchRecords();
  }, [filterType]);

  const fetchRecords = async () => {
    setLoading(true);
    try {
      const response = await patientPortalAPI.getMedicalRecords(filterType);
      setRecords(response.data.medical_records || []);
    } catch (error) {
      message.error('Erro ao carregar prontuarios');
    } finally {
      setLoading(false);
    }
  };

  const handleViewRecord = async (id) => {
    setLoadingDetail(true);
    setDetailModalVisible(true);
    try {
      const response = await patientPortalAPI.getMedicalRecordDetail(id);
      setSelectedRecord(response.data.medical_record);
    } catch (error) {
      message.error('Erro ao carregar detalhes do prontuario');
      setDetailModalVisible(false);
    } finally {
      setLoadingDetail(false);
    }
  };

  const RecordCard = ({ record }) => (
    <Card
      size="small"
      style={{
        marginBottom: 12,
        borderRadius: 12,
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
      }}
      hoverable
      onClick={() => handleViewRecord(record.id)}
    >
      <Row gutter={[12, 8]} align="middle">
        {/* Icone e Tipo */}
        <Col xs={24} sm={8}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                width: 44,
                height: 44,
                borderRadius: 10,
                background: `linear-gradient(135deg, ${typeColors[record.type] === 'purple' ? '#9c27b0' : typeColors[record.type] === 'blue' ? '#2196f3' : typeColors[record.type] === 'green' ? '#4caf50' : typeColors[record.type] === 'orange' ? '#ff9800' : typeColors[record.type] === 'cyan' ? '#00bcd4' : '#3f51b5'} 0%, ${typeColors[record.type] === 'purple' ? '#7b1fa2' : typeColors[record.type] === 'blue' ? '#1976d2' : typeColors[record.type] === 'green' ? '#388e3c' : typeColors[record.type] === 'orange' ? '#f57c00' : typeColors[record.type] === 'cyan' ? '#0097a7' : '#303f9f'} 100%)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <FileTextOutlined style={{ fontSize: 20, color: '#fff' }} />
            </div>
            <div>
              <Tag color={typeColors[record.type] || 'default'} style={{ margin: 0 }}>
                {typeLabels[record.type] || record.type}
              </Tag>
              <Text type="secondary" style={{ display: 'block', fontSize: 12, marginTop: 4 }}>
                <CalendarOutlined style={{ marginRight: 4 }} />
                {dayjs(record.created_at).format('DD/MM/YYYY')}
              </Text>
            </div>
          </div>
        </Col>

        {/* Profissional */}
        <Col xs={24} sm={8}>
          <Text>
            <UserOutlined style={{ marginRight: 6, color: '#66BB6A' }} />
            {record.dentist_name || 'Nao informado'}
          </Text>
          {record.dentist_cro && (
            <Text type="secondary" style={{ display: 'block', fontSize: 12 }}>
              CRO: {record.dentist_cro}
            </Text>
          )}
        </Col>

        {/* Diagnóstico e Assinatura */}
        <Col xs={24} sm={8}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div style={{ flex: 1 }}>
              {record.diagnosis && (
                <Text type="secondary" ellipsis style={{ maxWidth: 150 }}>
                  {record.diagnosis}
                </Text>
              )}
            </div>
            <Space>
              {record.is_signed && (
                <Tag icon={<SafetyCertificateOutlined />} color="success">
                  Assinado
                </Tag>
              )}
              <Button
                type="primary"
                size="small"
                icon={<EyeOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleViewRecord(record.id);
                }}
              >
                Ver
              </Button>
            </Space>
          </div>
        </Col>
      </Row>
    </Card>
  );

  return (
    <div>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            <Title level={4} style={{ margin: 0 }}>
              Meus Prontuarios
            </Title>
          </Space>
        }
        extra={
          <Select
            placeholder="Filtrar"
            allowClear
            style={{ width: 140 }}
            value={filterType || undefined}
            onChange={(value) => setFilterType(value || '')}
            suffixIcon={<FilterOutlined />}
          >
            {Object.entries(typeLabels).map(([key, label]) => (
              <Select.Option key={key} value={key}>
                {label}
              </Select.Option>
            ))}
          </Select>
        }
        bodyStyle={{ padding: '12px 16px' }}
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : records.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="Nenhum prontuario encontrado"
          />
        ) : (
          <div>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              Total: {records.length} registro(s)
            </Text>
            {records.map((record) => (
              <RecordCard key={record.id} record={record} />
            ))}
          </div>
        )}
      </Card>

      {/* Detail Modal */}
      <Modal
        title={
          <Space>
            <FileTextOutlined />
            Detalhes do Prontuario
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => {
          setDetailModalVisible(false);
          setSelectedRecord(null);
        }}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            Fechar
          </Button>,
        ]}
        width={700}
      >
        {loadingDetail ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : selectedRecord ? (
          <div>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="Data">
                {dayjs(selectedRecord.created_at).format('DD/MM/YYYY HH:mm')}
              </Descriptions.Item>
              <Descriptions.Item label="Tipo">
                <Tag color={typeColors[selectedRecord.type] || 'default'}>
                  {typeLabels[selectedRecord.type] || selectedRecord.type}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Profissional" span={2}>
                {selectedRecord.dentist_name || 'Nao informado'}
                {selectedRecord.dentist_cro && ` (CRO: ${selectedRecord.dentist_cro})`}
                {selectedRecord.dentist_specialty && ` - ${selectedRecord.dentist_specialty}`}
              </Descriptions.Item>
            </Descriptions>

            {selectedRecord.diagnosis && (
              <>
                <Divider orientation="left">Diagnostico</Divider>
                <Paragraph>{selectedRecord.diagnosis}</Paragraph>
              </>
            )}

            {selectedRecord.treatment_plan && (
              <>
                <Divider orientation="left">Plano de Tratamento</Divider>
                <Paragraph>{selectedRecord.treatment_plan}</Paragraph>
              </>
            )}

            {selectedRecord.procedure_done && (
              <>
                <Divider orientation="left">Procedimento Realizado</Divider>
                <Paragraph>{selectedRecord.procedure_done}</Paragraph>
              </>
            )}

            {selectedRecord.materials && (
              <>
                <Divider orientation="left">Materiais Utilizados</Divider>
                <Paragraph>{selectedRecord.materials}</Paragraph>
              </>
            )}

            {selectedRecord.evolution && (
              <>
                <Divider orientation="left">Evolucao</Divider>
                <Paragraph>{selectedRecord.evolution}</Paragraph>
              </>
            )}

            {selectedRecord.notes && (
              <>
                <Divider orientation="left">Observacoes</Divider>
                <Paragraph>{selectedRecord.notes}</Paragraph>
              </>
            )}

            {selectedRecord.is_signed && (
              <>
                <Divider />
                <div
                  style={{
                    background: '#f6ffed',
                    border: '1px solid #b7eb8f',
                    borderRadius: 8,
                    padding: 16,
                  }}
                >
                  <Space>
                    <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
                    <div>
                      <Text strong style={{ color: '#52c41a' }}>
                        Documento Assinado Digitalmente
                      </Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        Assinado por: {selectedRecord.signed_by_name}
                        {selectedRecord.signed_by_cro && ` (CRO: ${selectedRecord.signed_by_cro})`}
                        {selectedRecord.signed_at &&
                          ` em ${dayjs(selectedRecord.signed_at).format('DD/MM/YYYY HH:mm')}`}
                      </Text>
                    </div>
                  </Space>
                </div>
              </>
            )}
          </div>
        ) : null}
      </Modal>
    </div>
  );
};

export default PatientMedicalRecords;
