import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Form,
  Input,
  Space,
  Tag,
  Typography,
  message,
  Modal,
  Popconfirm,
  Spin,
  Alert,
} from 'antd';
import {
  KeyOutlined,
  UserOutlined,
  LockOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { patientPortalAdminAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const { Text } = Typography;

const PatientPortalAccess = ({ patient }) => {
  const [loading, setLoading] = useState(true);
  const [hasAccess, setHasAccess] = useState(false);
  const [accessInfo, setAccessInfo] = useState(null);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [createForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const { canCreate, canEdit, canDelete, canView } = usePermission();

  useEffect(() => {
    if (patient?.id) {
      checkAccess();
    }
  }, [patient?.id]);

  const checkAccess = async () => {
    setLoading(true);
    try {
      const response = await patientPortalAdminAPI.getAccess(patient.id);
      setHasAccess(response.data.has_access);
      setAccessInfo(response.data.user);
    } catch (error) {
      console.error('Error checking portal access:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAccess = async (values) => {
    setSubmitting(true);
    try {
      await patientPortalAdminAPI.createAccess({
        patient_id: patient.id,
        email: values.email,
        password: values.password,
      });
      message.success('Acesso ao portal criado com sucesso');
      setCreateModalVisible(false);
      createForm.resetFields();
      checkAccess();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar acesso');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdatePassword = async (values) => {
    setSubmitting(true);
    try {
      await patientPortalAdminAPI.updatePassword(patient.id, values.password);
      message.success('Senha atualizada com sucesso');
      setPasswordModalVisible(false);
      passwordForm.resetFields();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao atualizar senha');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteAccess = async () => {
    try {
      await patientPortalAdminAPI.deleteAccess(patient.id);
      message.success('Acesso ao portal removido');
      setHasAccess(false);
      setAccessInfo(null);
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao remover acesso');
    }
  };

  if (!canView('patient_portal')) {
    return null;
  }

  if (loading) {
    return (
      <Card
        type="inner"
        title={
          <Space>
            <KeyOutlined />
            <span>Portal do Paciente</span>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <div style={{ textAlign: 'center', padding: 20 }}>
          <Spin />
        </div>
      </Card>
    );
  }

  return (
    <>
      <Card
        type="inner"
        title={
          <Space>
            <KeyOutlined />
            <span>Portal do Paciente</span>
            {hasAccess ? (
              <Tag color="success" icon={<CheckCircleOutlined />}>
                Ativo
              </Tag>
            ) : (
              <Tag color="default" icon={<CloseCircleOutlined />}>
                Sem acesso
              </Tag>
            )}
          </Space>
        }
        style={{ marginBottom: 16 }}
        extra={
          hasAccess ? (
            <Space>
              {canEdit('patient_portal') && (
                <Button
                  icon={<LockOutlined />}
                  onClick={() => setPasswordModalVisible(true)}
                >
                  Alterar Senha
                </Button>
              )}
              {canDelete('patient_portal') && (
                <Popconfirm
                  title="Remover acesso"
                  description="Tem certeza que deseja remover o acesso deste paciente ao portal?"
                  onConfirm={handleDeleteAccess}
                  okText="Sim"
                  cancelText="Nao"
                  okButtonProps={{ danger: true }}
                >
                  <Button danger icon={<DeleteOutlined />}>
                    Remover
                  </Button>
                </Popconfirm>
              )}
            </Space>
          ) : (
            canCreate('patient_portal') && (
              <Button
                type="primary"
                icon={<KeyOutlined />}
                onClick={() => {
                  createForm.setFieldsValue({
                    email: patient.email || '',
                  });
                  setCreateModalVisible(true);
                }}
              >
                Criar Acesso
              </Button>
            )
          )
        }
      >
        {hasAccess ? (
          <Space direction="vertical">
            <Text>
              <UserOutlined style={{ marginRight: 8 }} />
              <strong>Email de acesso:</strong> {accessInfo?.email}
            </Text>
            <Text type="secondary">
              O paciente pode acessar o portal usando este email e a senha definida pela clinica.
            </Text>
          </Space>
        ) : (
          <Alert
            type="info"
            message="Este paciente ainda nao possui acesso ao portal"
            description="Clique em 'Criar Acesso' para permitir que o paciente acesse o portal e agende consultas online."
          />
        )}
      </Card>

      {/* Modal para criar acesso */}
      <Modal
        title="Criar Acesso ao Portal"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          createForm.resetFields();
        }}
        footer={null}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreateAccess}
        >
          <Form.Item
            name="email"
            label="Email de acesso"
            rules={[
              { required: true, message: 'Informe o email' },
              { type: 'email', message: 'Email invalido' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="Email do paciente" />
          </Form.Item>

          <Form.Item
            name="password"
            label="Senha"
            rules={[
              { required: true, message: 'Informe a senha' },
              { min: 6, message: 'A senha deve ter no minimo 6 caracteres' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Senha de acesso" />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            label="Confirmar Senha"
            dependencies={['password']}
            rules={[
              { required: true, message: 'Confirme a senha' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('As senhas nao coincidem'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Confirmar senha" />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setCreateModalVisible(false)}>
                Cancelar
              </Button>
              <Button type="primary" htmlType="submit" loading={submitting}>
                Criar Acesso
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Modal para alterar senha */}
      <Modal
        title="Alterar Senha do Portal"
        open={passwordModalVisible}
        onCancel={() => {
          setPasswordModalVisible(false);
          passwordForm.resetFields();
        }}
        footer={null}
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handleUpdatePassword}
        >
          <Form.Item
            name="password"
            label="Nova Senha"
            rules={[
              { required: true, message: 'Informe a nova senha' },
              { min: 6, message: 'A senha deve ter no minimo 6 caracteres' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Nova senha" />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            label="Confirmar Nova Senha"
            dependencies={['password']}
            rules={[
              { required: true, message: 'Confirme a nova senha' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('As senhas nao coincidem'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Confirmar nova senha" />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setPasswordModalVisible(false)}>
                Cancelar
              </Button>
              <Button type="primary" htmlType="submit" loading={submitting}>
                Alterar Senha
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default PatientPortalAccess;
