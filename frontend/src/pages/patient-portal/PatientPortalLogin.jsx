import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Spin, Alert } from 'antd';
import { UserOutlined, LockOutlined, MedicineBoxOutlined } from '@ant-design/icons';
import { patientPortalPublicAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

const { Title, Text } = Typography;

// Extract clinic slug from subdomain
const getClinicSlug = () => {
  const hostname = window.location.hostname;

  // For local development (localhost)
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    // Check URL params for testing
    const params = new URLSearchParams(window.location.search);
    return params.get('slug') || null;
  }

  // For production (xxx.odowell.pro)
  const parts = hostname.split('.');
  if (parts.length >= 3 && parts[1] === 'odowell' && parts[2] === 'pro') {
    const slug = parts[0];
    // Skip reserved subdomains
    if (slug === 'app' || slug === 'api' || slug === 'www') {
      return null;
    }
    return slug;
  }

  return null;
};

const PatientPortalLogin = () => {
  const [loading, setLoading] = useState(false);
  const [loadingClinic, setLoadingClinic] = useState(true);
  const [clinicName, setClinicName] = useState(null);
  const [clinicSlug, setClinicSlug] = useState(null);
  const [error, setError] = useState(null);
  const navigate = useNavigate();
  const { loginWithData } = useAuth();

  useEffect(() => {
    const slug = getClinicSlug();
    if (!slug) {
      setError('URL inválida. Acesse através do endereço fornecido pela sua clínica.');
      setLoadingClinic(false);
      return;
    }

    setClinicSlug(slug);
    loadClinicInfo(slug);
  }, []);

  const loadClinicInfo = async (slug) => {
    try {
      const response = await patientPortalPublicAPI.getClinicInfo(slug);
      setClinicName(response.data.clinic_name);
      setError(null);
    } catch (err) {
      if (err.response?.status === 404) {
        setError('Clínica não encontrada. Verifique o endereço.');
      } else {
        setError('Erro ao carregar informações da clínica.');
      }
    } finally {
      setLoadingClinic(false);
    }
  };

  const onFinish = async (values) => {
    if (!clinicSlug) return;

    setLoading(true);
    try {
      const response = await patientPortalPublicAPI.login(clinicSlug, values.email, values.password);

      // Update auth context with login data
      loginWithData(response.data);

      message.success('Login realizado com sucesso!');
      navigate('/patient');
    } catch (err) {
      const errorMsg = err.response?.data?.error || 'Erro ao fazer login';
      if (err.response?.status === 402) {
        message.error('A clínica está com a assinatura inativa. Entre em contato.');
      } else {
        message.error(errorMsg);
      }
    } finally {
      setLoading(false);
    }
  };

  if (loadingClinic) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
      }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
      }}>
        <Card style={{ width: 400, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
          <Alert
            message="Erro"
            description={error}
            type="error"
            showIcon
          />
        </Card>
      </div>
    );
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)'
    }}>
      <Card style={{ width: 400, boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <div style={{
            width: 80,
            height: 80,
            margin: '0 auto 16px',
            background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
            borderRadius: 16,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(102, 187, 106, 0.3)',
          }}>
            <MedicineBoxOutlined style={{ fontSize: 40, color: '#fff' }} />
          </div>
          <Title level={3} style={{ color: '#4CAF50', marginTop: 0, marginBottom: 8 }}>
            {clinicName || 'Portal do Paciente'}
          </Title>
          <Text type="secondary">Acesse sua area do paciente</Text>
        </div>

        <Form
          name="patient-portal-login"
          onFinish={onFinish}
          size="large"
          layout="vertical"
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: 'Por favor, insira seu email!' },
              { type: 'email', message: 'Email invalido!' }
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="Email"
              autoComplete="email"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Por favor, insira sua senha!' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Senha"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Entrar
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 24, paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Para redefinir sua senha, entre em contato com a clinica.
          </Text>
        </div>
      </Card>
    </div>
  );
};

export default PatientPortalLogin;
