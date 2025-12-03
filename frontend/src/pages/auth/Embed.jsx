import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Spin, Result, Typography } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import { authAPI } from '../../services/api';

const { Text } = Typography;

const Embed = () => {
  const { token, '*': page } = useParams();
  const navigate = useNavigate();
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const authenticate = async () => {
      try {
        // Clear any existing session
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('tenant');
        localStorage.removeItem('subscription');

        // Authenticate with embed token
        const response = await authAPI.embedAuth(token);
        const { token: jwtToken, user, tenant } = response.data;

        // Store session data
        localStorage.setItem('token', jwtToken);
        localStorage.setItem('user', JSON.stringify(user));
        localStorage.setItem('tenant', JSON.stringify(tenant));

        // Redirect to requested page (or dashboard if not specified)
        const targetPage = page || '';
        navigate(`/${targetPage}`, { replace: true });
      } catch (err) {
        console.error('Embed auth error:', err);
        const errorMessage = err.response?.data?.error || 'Erro de autenticação';
        setError(errorMessage);
        setLoading(false);
      }
    };

    if (token) {
      authenticate();
    } else {
      setError('Token não fornecido');
      setLoading(false);
    }
  }, [token, page, navigate]);

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        background: '#f5f5f5'
      }}>
        <Spin indicator={<LoadingOutlined style={{ fontSize: 48, color: '#66BB6A' }} spin />} />
        <Text style={{ marginTop: 16, color: '#666' }}>Autenticando...</Text>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        background: '#f5f5f5'
      }}>
        <Result
          status="error"
          title="Erro de Autenticação"
          subTitle={error}
        />
      </div>
    );
  }

  return null;
};

export default Embed;
