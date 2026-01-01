import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Typography,
  Spin,
  message,
} from 'antd';
import {
  HomeOutlined,
  CalendarOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  MedicineBoxOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { patientPortalAPI } from '../../services/api';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const PatientPortalLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [clinicInfo, setClinicInfo] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchClinicInfo();
  }, []);

  const fetchClinicInfo = async () => {
    try {
      const response = await patientPortalAPI.getClinic();
      setClinicInfo(response.data);
    } catch (error) {
      console.error('Error fetching clinic info:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      message.error('Erro ao sair');
    }
  };

  const menuItems = [
    {
      key: '/patient',
      icon: <HomeOutlined />,
      label: 'Inicio',
    },
    {
      key: '/patient/appointments',
      icon: <CalendarOutlined />,
      label: 'Minhas Consultas',
    },
    {
      key: '/patient/book',
      icon: <MedicineBoxOutlined />,
      label: 'Agendar Consulta',
    },
    {
      key: '/patient/medical-records',
      icon: <FileTextOutlined />,
      label: 'Meus Prontuarios',
    },
    {
      key: '/patient/profile',
      icon: <UserOutlined />,
      label: 'Meus Dados',
    },
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Meus Dados',
      onClick: () => navigate('/patient/profile'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Sair',
      onClick: handleLogout,
    },
  ];

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        breakpoint="lg"
        collapsedWidth={80}
        style={{
          background: '#fff',
          boxShadow: '2px 0 8px rgba(0,0,0,0.05)',
        }}
        onBreakpoint={(broken) => {
          setCollapsed(broken);
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #f0f0f0',
            padding: '0 16px',
          }}
        >
          {!collapsed ? (
            <Text strong style={{ fontSize: 16, color: '#66BB6A' }}>
              {clinicInfo?.clinic?.name || 'Portal do Paciente'}
            </Text>
          ) : (
            <MedicineBoxOutlined style={{ fontSize: 24, color: '#66BB6A' }} />
          )}
        </div>

        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ borderRight: 0, marginTop: 16 }}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
              style: { fontSize: 18, cursor: 'pointer' },
              onClick: () => setCollapsed(!collapsed),
            })}
            <Text style={{ marginLeft: 16, fontSize: 16 }}>
              Portal do Paciente
            </Text>
          </div>

          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#66BB6A' }} />
              <Text style={{ marginLeft: 8 }}>{user?.name}</Text>
            </div>
          </Dropdown>
        </Header>

        <Content
          style={{
            margin: 24,
            padding: 24,
            background: '#fff',
            borderRadius: 8,
            minHeight: 280,
          }}
        >
          <Outlet context={{ clinicInfo }} />
        </Content>
      </Layout>
    </Layout>
  );
};

export default PatientPortalLayout;
