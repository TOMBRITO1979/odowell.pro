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
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  FilePdfOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { medicalRecordsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const MedicalRecordDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [record, setRecord] = useState(null);
  const [loading, setLoading] = useState(false);
  const { canEdit, canDelete } = usePermission();

  const recordTypes = [
    { value: 'anamnesis', label: 'Anamnese', color: 'blue' },
    { value: 'treatment', label: 'Tratamento', color: 'green' },
    { value: 'procedure', label: 'Procedimento', color: 'purple' },
    { value: 'prescription', label: 'Receita', color: 'orange' },
    { value: 'certificate', label: 'Atestado', color: 'red' },
  ];

  useEffect(() => {
    fetchRecord();
  }, [id]);

  const fetchRecord = async () => {
    setLoading(true);
    try {
      const response = await medicalRecordsAPI.getOne(id);
      setRecord(response.data.record);
    } catch (error) {
      message.error('Erro ao carregar prontuário');
      console.error('Error:', error);
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
      console.error('Error:', error);
    }
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
            <Button
              type="primary"
              danger
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
            >
              Baixar PDF
            </Button>
            {canEdit('medical_records') && (
              <Button
                type="primary"
                icon={<EditOutlined />}
                onClick={() => navigate(`/medical-records/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            {canDelete('medical_records') && (
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
            <Descriptions.Item label="Odontograma" span={2}>
              <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                {JSON.stringify(JSON.parse(record.odontogram), null, 2)}
              </pre>
            </Descriptions.Item>
          )}

          {record.notes && (
            <Descriptions.Item label="Notas Adicionais" span={2}>
              {record.notes}
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>
    </div>
  );
};

export default MedicalRecordDetails;
