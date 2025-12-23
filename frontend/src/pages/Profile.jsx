import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Row, Col, Modal, Divider, Avatar, Upload } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined, MailOutlined, CameraOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { authAPI } from '../services/api';

const API_URL = import.meta.env.VITE_API_URL ;

const Profile = () => {
  const { user, updateUser } = useAuth();
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [uploadingPicture, setUploadingPicture] = useState(false);
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);

  useEffect(() => {
    if (user) {
      form.setFieldsValue({
        name: user.name,
        email: user.email,
        phone: user.phone || '',
        cro: user.cro || '',
        specialty: user.specialty || '',
      });
    }
  }, [user, form]);

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      const response = await authAPI.updateProfile(values);
      message.success('Perfil atualizado com sucesso');
      updateUser(response.data.user);
    } catch (error) {
      message.error('Erro ao atualizar perfil');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async (values) => {
    try {
      await authAPI.changePassword({
        current_password: values.current_password,
        new_password: values.new_password,
      });
      message.success('Senha alterada com sucesso');
      setPasswordModalVisible(false);
      passwordForm.resetFields();
    } catch (error) {
      message.error('Erro ao alterar senha. Verifique a senha atual.');
    }
  };

  const handleProfilePictureUpload = async (file) => {
    // Validate file type
    const isImage = file.type === 'image/jpeg' || file.type === 'image/png' || file.type === 'image/jpg';
    if (!isImage) {
      message.error('Você só pode fazer upload de arquivos JPG/PNG!');
      return false;
    }

    // Validate file size (5MB)
    const isLt5M = file.size / 1024 / 1024 < 5;
    if (!isLt5M) {
      message.error('A imagem deve ter menos de 5MB!');
      return false;
    }

    setUploadingPicture(true);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await authAPI.uploadProfilePicture(formData);
      message.success('Foto de perfil atualizada com sucesso!');

      // Update user with new profile picture
      const updatedUser = { ...user, profile_picture: response.data.profile_picture };
      updateUser(updatedUser);
    } catch (error) {
      message.error('Erro ao fazer upload da foto de perfil');
    } finally {
      setUploadingPicture(false);
    }

    return false; // Prevent default upload behavior
  };

  const profilePictureUrl = user?.profile_picture
    ? `${API_URL}/${user.profile_picture}`
    : null;

  return (
    <div>
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <UserOutlined />
            <span>Meu Perfil</span>
          </div>
        }
      >
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 24 }}>
          <Upload
            name="file"
            showUploadList={false}
            beforeUpload={handleProfilePictureUpload}
            accept="image/jpeg,image/png,image/jpg"
          >
            <div style={{ position: 'relative', cursor: 'pointer' }}>
              <Avatar
                size={120}
                icon={<UserOutlined />}
                src={profilePictureUrl}
                style={{ backgroundColor: '#66BB6A' }}
              />
              <div
                style={{
                  position: 'absolute',
                  bottom: 0,
                  right: 0,
                  backgroundColor: '#66BB6A',
                  borderRadius: '50%',
                  width: 36,
                  height: 36,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  border: '3px solid white',
                  cursor: 'pointer',
                }}
              >
                <CameraOutlined style={{ color: 'white', fontSize: 16 }} />
              </div>
            </div>
          </Upload>
        </div>

        <Divider />

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
        >
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Nome Completo"
                name="name"
                rules={[{ required: true, message: 'Por favor, insira seu nome' }]}
              >
                <Input prefix={<UserOutlined />} placeholder="Seu nome completo" />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                label="E-mail"
                name="email"
                rules={[
                  { required: true, message: 'Por favor, insira seu e-mail' },
                  { type: 'email', message: 'E-mail inválido' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder="seu@email.com" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Telefone"
                name="phone"
              >
                <Input prefix={<PhoneOutlined />} placeholder="(00) 00000-0000" />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                label="CRO"
                name="cro"
              >
                <Input placeholder="Número do CRO" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Especialidade"
                name="specialty"
              >
                <Input placeholder="Ex: Ortodontia, Implantodontia..." />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Row gutter={16}>
            <Col>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  Salvar Alterações
                </Button>
              </Form.Item>
            </Col>
            <Col>
              <Button
                icon={<LockOutlined />}
                onClick={() => setPasswordModalVisible(true)}
              >
                Alterar Senha
              </Button>
            </Col>
          </Row>
        </Form>
      </Card>

      <Modal
        title="Alterar Senha"
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
          onFinish={handlePasswordChange}
          autoComplete="off"
        >
          <Form.Item
            label="Senha Atual"
            name="current_password"
            rules={[{ required: true, message: 'Por favor, insira sua senha atual' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Senha atual" />
          </Form.Item>

          <Form.Item
            label="Nova Senha"
            name="new_password"
            rules={[
              { required: true, message: 'Por favor, insira a nova senha' },
              { min: 6, message: 'A senha deve ter no mínimo 6 caracteres' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Nova senha" />
          </Form.Item>

          <Form.Item
            label="Confirmar Nova Senha"
            name="confirm_password"
            dependencies={['new_password']}
            rules={[
              { required: true, message: 'Por favor, confirme a nova senha' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('As senhas não conferem'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Confirme a nova senha" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              Alterar Senha
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Profile;
