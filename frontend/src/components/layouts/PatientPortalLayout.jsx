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
  Drawer,
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
  MenuOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { patientPortalAPI } from '../../services/api';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const PatientPortalLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, isPatient, loading: authLoading } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [clinicInfo, setClinicInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [drawerVisible, setDrawerVisible] = useState(false);

  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth <= 768;
      setIsMobile(mobile);
      if (!mobile) setDrawerVisible(false);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    // Wait for auth context to finish loading before checking user type
    if (authLoading) return;

    // SECURITY: Verify user is actually a patient before making any API calls
    // This prevents staff tokens from being used on patient portal
    if (!isPatient) {
      console.warn('Non-patient user detected on patient portal, logging out');
      logout();
      navigate('/portal-login');
      return;
    }
    fetchClinicInfo();
  }, [isPatient, authLoading, logout, navigate]);

  const fetchClinicInfo = async () => {
    try {
      const response = await patientPortalAPI.getClinic();
      setClinicInfo(response.data);
    } catch (error) {
      console.error('Error fetching clinic info:', error);
      // If 403, likely wrong user type - logout
      if (error.response?.status === 403) {
        logout();
        navigate('/portal-login');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/portal-login');
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

  if (loading || authLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  const handleMenuClick = ({ key }) => {
    navigate(key);
    if (isMobile) setDrawerVisible(false);
  };

  const menuContent = (
    <>
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
        <Text strong style={{ fontSize: 16, color: '#66BB6A' }}>
          {clinicInfo?.clinic?.name || 'Portal do Paciente'}
        </Text>
      </div>
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={handleMenuClick}
        style={{ borderRight: 0, marginTop: 16 }}
      />
    </>
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* Mobile Drawer Menu */}
      {isMobile && (
        <Drawer
          title={clinicInfo?.clinic?.name || 'Portal do Paciente'}
          placement="left"
          onClose={() => setDrawerVisible(false)}
          open={drawerVisible}
          bodyStyle={{ padding: 0 }}
          width={280}
        >
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={handleMenuClick}
            style={{ borderRight: 0 }}
          />
        </Drawer>
      )}

      {/* Desktop Sider */}
      {!isMobile && (
        <Sider
          trigger={null}
          collapsible
          collapsed={collapsed}
          collapsedWidth={80}
          style={{
            background: '#fff',
            boxShadow: '2px 0 8px rgba(0,0,0,0.05)',
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
            onClick={handleMenuClick}
            style={{ borderRight: 0, marginTop: 16 }}
          />
        </Sider>
      )}

      <Layout>
        <Header
          style={{
            padding: isMobile ? '0 12px' : '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {isMobile ? (
              <MenuOutlined
                style={{ fontSize: 20, cursor: 'pointer' }}
                onClick={() => setDrawerVisible(true)}
              />
            ) : (
              React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
                style: { fontSize: 18, cursor: 'pointer' },
                onClick: () => setCollapsed(!collapsed),
              })
            )}
            <Text style={{ marginLeft: 12, fontSize: isMobile ? 14 : 16 }}>
              Portal do Paciente
            </Text>
          </div>

          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#66BB6A' }} size={isMobile ? 'small' : 'default'} />
              {!isMobile && <Text style={{ marginLeft: 8 }}>{user?.name}</Text>}
            </div>
          </Dropdown>
        </Header>

        <Content
          style={{
            margin: isMobile ? 8 : 24,
            padding: isMobile ? 12 : 24,
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
