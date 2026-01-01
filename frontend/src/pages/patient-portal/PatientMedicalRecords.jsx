import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
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
} from 'antd';
import {
  FileTextOutlined,
  EyeOutlined,
  UserOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  SafetyCertificateOutlined,
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

  const columns = [
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'date',
      render: (value) => (
        <Space>
          <CalendarOutlined />
          {dayjs(value).format('DD/MM/YYYY')}
        </Space>
      ),
      sorter: (a, b) => dayjs(a.created_at).unix() - dayjs(b.created_at).unix(),
      defaultSortOrder: 'descend',
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag color={typeColors[type] || 'default'}>
          {typeLabels[type] || type}
        </Tag>
      ),
    },
    {
      title: 'Profissional',
      key: 'dentist',
      render: (_, record) => (
        <Space>
          <UserOutlined />
          {record.dentist_name || 'Nao informado'}
          {record.dentist_cro && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              (CRO: {record.dentist_cro})
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: 'Diagnostico',
      dataIndex: 'diagnosis',
      key: 'diagnosis',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: 'Assinado',
      dataIndex: 'is_signed',
      key: 'is_signed',
      width: 100,
      render: (isSigned) =>
        isSigned ? (
          <Tag icon={<SafetyCertificateOutlined />} color="success">
            Sim
          </Tag>
        ) : (
          <Tag color="default">Nao</Tag>
        ),
    },
    {
      title: 'Acoes',
      key: 'actions',
      width: 100,
      render: (_, record) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => handleViewRecord(record.id)}
        >
          Ver
        </Button>
      ),
    },
  ];

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
            placeholder="Filtrar por tipo"
            allowClear
            style={{ width: 180 }}
            value={filterType || undefined}
            onChange={(value) => setFilterType(value || '')}
          >
            {Object.entries(typeLabels).map(([key, label]) => (
              <Select.Option key={key} value={key}>
                {label}
              </Select.Option>
            ))}
          </Select>
        }
      >
        <Table
          dataSource={records}
          columns={columns}
          rowKey="id"
          loading={loading}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="Nenhum prontuario encontrado"
              />
            ),
          }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `Total: ${total} registros`,
          }}
        />
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
            <Descriptions column={2} bordered size="small">
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
