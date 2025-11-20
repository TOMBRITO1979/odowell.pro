import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Row, Col, Modal, Divider } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined, MailOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { authAPI } from '../services/api';

const Profile = () => {
  const { user, updateUser } = useAuth();
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
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
      console.error('Error:', error);
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
      console.error('Error:', error);
    }
  };

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
