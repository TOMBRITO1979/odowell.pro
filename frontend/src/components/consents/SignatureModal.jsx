import { useState, useRef } from 'react';
import { Modal, Form, Input, Select, Button, Space, Typography, Divider, message } from 'antd';
import { ClearOutlined } from '@ant-design/icons';
import SignatureCanvas from 'react-signature-canvas';
import { consentsAPI } from '../../services/api';

const { TextArea } = Input;
const { Title, Text } = Typography;

const SignatureModal = ({ visible, onClose, patient, template, onSuccess }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const signaturePadRef = useRef(null);

  const handleClear = () => {
    if (signaturePadRef.current) {
      signaturePadRef.current.clear();
    }
  };

  const handleSubmit = async (values) => {
    try {
      // Check if signature is provided
      if (signaturePadRef.current && signaturePadRef.current.isEmpty()) {
        message.error('Por favor, assine o termo');
        return;
      }

      setLoading(true);

      // Get signature as base64
      const signatureData = signaturePadRef.current.toDataURL();

      // Prepare consent data
      const consentData = {
        patient_id: patient.id,
        template_id: template.id,
        signature_data: signatureData,
        signature_type: 'digital',
        signer_name: values.signer_name,
        signer_relation: values.signer_relation,
        witness_name: values.witness_name || '',
        notes: values.notes || '',
      };

      // Create consent
      await consentsAPI.create(consentData);

      message.success('Termo assinado com sucesso');
      form.resetFields();
      handleClear();

      if (onSuccess) {
        onSuccess();
      }

      onClose();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao assinar termo');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    form.resetFields();
    handleClear();
    onClose();
  };

  return (
    <Modal
      title="Assinatura de Termo de Consentimento"
      open={visible}
      onCancel={handleCancel}
      onOk={() => form.submit()}
      width={800}
      okText="Confirmar Assinatura"
      cancelText="Cancelar"
      confirmLoading={loading}
    >
      {template && (
        <div>
          <Title level={4}>{template.title}</Title>
          <Text type="secondary">Versão: {template.version}</Text>
          <Divider />

          <div style={{
            maxHeight: '200px',
            overflowY: 'auto',
            padding: '12px',
            backgroundColor: '#f5f5f5',
            borderRadius: '4px',
            marginBottom: '16px'
          }}>
            <Text>{template.content}</Text>
          </div>

          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            initialValues={{
              signer_name: patient?.name || '',
              signer_relation: 'patient'
            }}
          >
            <Form.Item
              label="Nome do Assinante"
              name="signer_name"
              rules={[{ required: true, message: 'Informe o nome do assinante' }]}
            >
              <Input placeholder="Nome completo" />
            </Form.Item>

            <Form.Item
              label="Relação com o Paciente"
              name="signer_relation"
              rules={[{ required: true, message: 'Selecione a relação' }]}
            >
              <Select>
                <Select.Option value="patient">Paciente</Select.Option>
                <Select.Option value="guardian">Responsável Legal</Select.Option>
                <Select.Option value="representative">Representante</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="Testemunha (Opcional)"
              name="witness_name"
            >
              <Input placeholder="Nome da testemunha" />
            </Form.Item>

            <Form.Item
              label="Observações (Opcional)"
              name="notes"
            >
              <TextArea rows={2} placeholder="Observações adicionais..." />
            </Form.Item>

            <Form.Item label="Assinatura Digital">
              <div style={{
                border: '2px dashed #d9d9d9',
                borderRadius: '4px',
                padding: '8px',
                backgroundColor: '#fff'
              }}>
                <SignatureCanvas
                  ref={signaturePadRef}
                  canvasProps={{
                    width: 700,
                    height: 200,
                    className: 'signature-canvas',
                    style: { width: '100%', height: '200px' }
                  }}
                  backgroundColor="white"
                />
              </div>
              <div style={{ marginTop: '8px', textAlign: 'right' }}>
                <Button
                  icon={<ClearOutlined />}
                  onClick={handleClear}
                  size="small"
                >
                  Limpar Assinatura
                </Button>
              </div>
            </Form.Item>
          </Form>
        </div>
      )}
    </Modal>
  );
};

export default SignatureModal;
