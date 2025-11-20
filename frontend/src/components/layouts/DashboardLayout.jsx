import React, { useState } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Typography } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  CalendarOutlined,
  MedicineBoxOutlined,
  DollarOutlined,
  ShoppingOutlined,
  MessageOutlined,
  FileOutlined,
  BarChartOutlined,
  LogoutOutlined,
  SettingOutlined,
  FormOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const DashboardLayout = () => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const { user, tenant, logout } = useAuth();

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      key: '/appointments',
      icon: <CalendarOutlined />,
      label: 'Agenda',
    },
    {
      key: '/patients',
      icon: <UserOutlined />,
      label: 'Pacientes',
    },
    {
      key: '/medical-records',
      icon: <MedicineBoxOutlined />,
      label: 'Prontuários',
    },
    {
      key: '/prescriptions',
      icon: <FormOutlined />,
      label: 'Receituário',
    },
    {
      key: '/exams',
      icon: <FileOutlined />,
      label: 'Exames',
    },
    {
      key: 'financial',
      icon: <DollarOutlined />,
      label: 'Financeiro',
      children: [
        { key: '/budgets', label: 'Orçamentos' },
        { key: '/payments', label: 'Pagamentos' },
      ],
    },
    {
      key: 'inventory',
      icon: <ShoppingOutlined />,
      label: 'Estoque',
      children: [
        { key: '/products', label: 'Produtos' },
        { key: '/suppliers', label: 'Fornecedores' },
        { key: '/stock-movements', label: 'Movimentações' },
      ],
    },
    {
      key: '/campaigns',
      icon: <MessageOutlined />,
      label: 'Campanhas',
    },
    {
      key: '/reports',
      icon: <BarChartOutlined />,
      label: 'Relatórios',
    },
  ];

  const handleUserMenuClick = ({ key }) => {
    if (key === 'profile') {
      navigate('/profile');
    } else if (key === 'settings') {
      navigate('/settings');
    } else if (key === 'logout') {
      logout();
      navigate('/login');
    }
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Perfil',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: 'Configurações',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Sair',
      danger: true,
    },
  ];

  const handleMenuClick = ({ key }) => {
    if (!key.startsWith('/')) return;
    navigate(key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="light"
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div style={{
          height: 64,
          margin: 16,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: collapsed ? 16 : 20,
          fontWeight: 'bold',
          color: '#1890ff'
        }}>
          {collapsed ? 'DC' : 'Dr. Crwell'}
        </div>
        <Menu
          theme="light"
          mode="inline"
          defaultSelectedKeys={['/']}
          selectedKeys={[window.location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>

      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'all 0.2s' }}>
        <Header style={{
          padding: '0 24px',
          background: '#fff',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: '1px solid #f0f0f0'
        }}>
          <div>
            <Text strong>{tenant?.name || 'Dr. Crwell'}</Text>
          </div>

          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight">
            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} style={{ marginRight: 8 }} />
              <Text>{user?.name}</Text>
            </div>
          </Dropdown>
        </Header>

        <Content style={{ margin: '24px 16px', padding: 24, background: '#fff', minHeight: 280 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default DashboardLayout;
