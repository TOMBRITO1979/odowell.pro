import React, { useState, useEffect } from 'react';
import {
  Modal,
  Button,
  Typography,
  Space,
  Input,
  Alert,
  Statistic,
  Row,
  Col,
  Card,
  message,
  Divider,
} from 'antd';
import {
  ExclamationCircleOutlined,
  DeleteOutlined,
  WarningOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { lgpdAPI } from '../../services/api';

const { Text, Title, Paragraph } = Typography;
const { TextArea } = Input;

const PermanentDeleteModal = ({ visible, onCancel, onSuccess, patientId, patientName }) => {
  const [loading, setLoading] = useState(false);
  const [preview, setPreview] = useState(null);
  const [step, setStep] = useState(1);
  const [confirmationInput, setConfirmationInput] = useState('');
  const [reason, setReason] = useState('');

  useEffect(() => {
    let mounted = true;

    const fetchPreview = async () => {
      setLoading(true);
      try {
        const response = await lgpdAPI.getDeletionPreview(patientId);
        if (mounted) {
          setPreview(response.data);
        }
      } catch (error) {
        if (mounted) {
          message.error('Erro ao carregar informacoes do paciente');
          onCancel();
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    };

    if (visible && patientId) {
      fetchPreview();
    } else {
      // Reset state when modal closes
      setStep(1);
      setConfirmationInput('');
      setReason('');
      setPreview(null);
    }

    return () => {
      mounted = false;
    };
  }, [visible, patientId, onCancel]);

  const handleDelete = async () => {
    if (!preview) return;

    if (confirmationInput !== preview.confirmation_token) {
      message.error('Token de confirmacao incorreto');
      return;
    }

    if (!reason.trim()) {
      message.error('Informe o motivo da exclusao');
      return;
    }

    setLoading(true);
    try {
      await lgpdAPI.permanentDelete(patientId, {
        confirmation_token: confirmationInput,
        reason: reason,
      });
      message.success('Paciente excluido permanentemente');
      onSuccess();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao excluir paciente');
    } finally {
      setLoading(false);
    }
  };

  const renderStep1 = () => (
    <div>
      <Alert
        message="Atencao: Esta acao e irreversivel!"
        description="A exclusao permanente remove todos os dados do paciente do sistema. Esta acao NAO pode ser desfeita."
        type="error"
        showIcon
        icon={<WarningOutlined />}
        style={{ marginBottom: 24 }}
      />

      {preview && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Space>
              <UserOutlined style={{ fontSize: 24 }} />
              <div>
                <Text strong>{preview.patient.name}</Text>
                <br />
                <Text type="secondary">CPF: {preview.patient.cpf}</Text>
              </div>
            </Space>
          </Card>

          <Title level={5}>Dados que serao excluidos:</Title>
          <Row gutter={[16, 16]}>
            <Col span={8}>
              <Statistic title="Agendamentos" value={preview.data_to_delete.appointments} />
            </Col>
            <Col span={8}>
              <Statistic title="Prontuarios" value={preview.data_to_delete.medical_records} />
            </Col>
            <Col span={8}>
              <Statistic title="Receitas" value={preview.data_to_delete.prescriptions} />
            </Col>
            <Col span={8}>
              <Statistic title="Orcamentos" value={preview.data_to_delete.budgets} />
            </Col>
            <Col span={8}>
              <Statistic title="Pagamentos" value={preview.data_to_delete.payments} />
            </Col>
            <Col span={8}>
              <Statistic title="Exames" value={preview.data_to_delete.exams} />
            </Col>
            <Col span={8}>
              <Statistic title="Consentimentos" value={preview.data_to_delete.consents} />
            </Col>
            <Col span={8}>
              <Statistic title="Tarefas" value={preview.data_to_delete.tasks} />
            </Col>
            <Col span={8}>
              <Statistic title="Anexos" value={preview.data_to_delete.attachments} />
            </Col>
          </Row>
        </>
      )}
    </div>
  );

  const renderStep2 = () => (
    <div>
      <Alert
        message="Confirmacao Final"
        description="Para confirmar a exclusao permanente, digite o token de confirmacao e informe o motivo."
        type="warning"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <div style={{ marginBottom: 16 }}>
        <Text strong>Token de confirmacao:</Text>
        <br />
        <Text code copyable>{preview?.confirmation_token}</Text>
      </div>

      <div style={{ marginBottom: 16 }}>
        <Text strong>Digite o token acima para confirmar:</Text>
        <Input
          placeholder="Digite o token de confirmacao"
          value={confirmationInput}
          onChange={(e) => setConfirmationInput(e.target.value)}
          style={{ marginTop: 8 }}
        />
      </div>

      <div style={{ marginBottom: 16 }}>
        <Text strong>Motivo da exclusao (obrigatorio):</Text>
        <TextArea
          placeholder="Ex: Solicitacao do titular conforme LGPD Art. 18"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={3}
          style={{ marginTop: 8 }}
        />
      </div>
    </div>
  );

  return (
    <Modal
      title={
        <Space>
          <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
          <span>Exclusao Permanente - LGPD</span>
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      width={600}
      footer={
        step === 1 ? [
          <Button key="cancel" onClick={onCancel}>
            Cancelar
          </Button>,
          <Button
            key="next"
            type="primary"
            danger
            onClick={() => setStep(2)}
            loading={loading}
          >
            Continuar
          </Button>,
        ] : [
          <Button key="back" onClick={() => setStep(1)}>
            Voltar
          </Button>,
          <Button
            key="delete"
            type="primary"
            danger
            icon={<DeleteOutlined />}
            onClick={handleDelete}
            loading={loading}
            disabled={!confirmationInput || !reason.trim()}
          >
            Excluir Permanentemente
          </Button>,
        ]
      }
    >
      {step === 1 ? renderStep1() : renderStep2()}
    </Modal>
  );
};

export default PermanentDeleteModal;
