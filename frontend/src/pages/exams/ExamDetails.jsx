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
  DownloadOutlined,
  FileOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { examsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const ExamDetails = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [exam, setExam] = useState(null);
  const [loading, setLoading] = useState(false);
  const { canEdit, canDelete } = usePermission();

  useEffect(() => {
    fetchExam();
  }, [id]);

  const fetchExam = async () => {
    setLoading(true);
    try {
      const response = await examsAPI.getOne(id);
      setExam(response.data.exam);
    } catch (error) {
      message.error('Erro ao carregar exame');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    try {
      const response = await examsAPI.getDownloadURL(id);
      window.open(response.data.download_url, '_blank');
    } catch (error) {
      message.error('Erro ao gerar link de download');
    }
  };

  const handleDelete = async () => {
    try {
      await examsAPI.delete(id);
      message.success('Exame deletado com sucesso');
      navigate('/exams');
    } catch (error) {
      message.error('Erro ao deletar exame');
    }
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return '-';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!exam) {
    return null;
  }

  return (
    <div>
      <Card
        title={
          <Space>
            <FileOutlined />
            <span>Detalhes do Exame</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/exams')}
            >
              Voltar
            </Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              onClick={handleDownload}
            >
              Download
            </Button>
            {canEdit('exams') && (
              <Button
                icon={<EditOutlined />}
                onClick={() => navigate(`/exams/${id}/edit`)}
              >
                Editar
              </Button>
            )}
            {canDelete('exams') && (
              <Popconfirm
                title="Tem certeza que deseja deletar este exame?"
                onConfirm={handleDelete}
                okText="Sim"
                cancelText="Não"
              >
                <Button danger icon={<DeleteOutlined />}>
                  Deletar
                </Button>
              </Popconfirm>
            )}
          </Space>
        }
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Nome do Exame" span={2}>
            {exam.name}
          </Descriptions.Item>

          <Descriptions.Item label="Tipo">
            {exam.exam_type ? <Tag color="blue">{exam.exam_type}</Tag> : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Data do Exame">
            {exam.exam_date ? dayjs(exam.exam_date).format('DD/MM/YYYY') : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Paciente" span={2}>
            {exam.patient_name || '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Nome do Arquivo" span={2}>
            {exam.file_name || '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Tamanho do Arquivo">
            {formatFileSize(exam.file_size)}
          </Descriptions.Item>

          <Descriptions.Item label="Data de Upload">
            {exam.created_at ? dayjs(exam.created_at).format('DD/MM/YYYY HH:mm') : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Última Atualização">
            {exam.updated_at ? dayjs(exam.updated_at).format('DD/MM/YYYY HH:mm') : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="Enviado por">
            {exam.uploaded_by_name || '-'}
          </Descriptions.Item>

          {exam.description && (
            <Descriptions.Item label="Descrição" span={2}>
              {exam.description}
            </Descriptions.Item>
          )}

          {exam.notes && (
            <Descriptions.Item label="Observações" span={2}>
              {exam.notes}
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>
    </div>
  );
};

export default ExamDetails;
