import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Select } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Title } = Typography;

const Register = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { register } = useAuth();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      await register(values);
      message.success('Cadastro realizado com sucesso! Faça login para continuar.');
      navigate('/login');
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao cadastrar');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    }}>
      <Card style={{ width: 400, boxShadow: '0 8px 32px rgba(0,0,0,0.1)' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Title level={2} style={{ color: '#1890ff' }}>Criar Conta</Title>
        </div>

        <Form name="register" onFinish={onFinish} size="large">
          <Form.Item
            name="tenant_id"
            rules={[{ required: true, message: 'Selecione o consultório!' }]}
          >
            <Input type="number" prefix={<UserOutlined />} placeholder="ID do Consultório" />
          </Form.Item>

          <Form.Item
            name="name"
            rules={[{ required: true, message: 'Digite seu nome!' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="Nome completo" />
          </Form.Item>

          <Form.Item
            name="email"
            rules={[
              { required: true, message: 'Digite seu email!' },
              { type: 'email', message: 'Email inválido!' }
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="Email" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: 'Digite uma senha!' },
              { min: 6, message: 'Senha deve ter no mínimo 6 caracteres!' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Senha" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Cadastrar
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Link to="/login">Já tem conta? Fazer login</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default Register;
