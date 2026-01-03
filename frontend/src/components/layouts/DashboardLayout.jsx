import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Typography, Badge, Drawer, Button, Alert } from 'antd';
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
  CheckSquareOutlined,
  MenuOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
  AppstoreOutlined,
  CrownOutlined,
  CreditCardOutlined,
  TagsOutlined,
  WalletOutlined,
  UsergroupAddOutlined,
  SafetyOutlined,
  SafetyCertificateOutlined,
  WarningOutlined,
  BellOutlined,
} from '@ant-design/icons';
import { useAuth, usePermission } from '../../contexts/AuthContext';
import { tasksAPI, paymentsAPI, portalNotificationsAPI } from '../../services/api';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const API_URL = import.meta.env.VITE_API_URL ;

const DashboardLayout = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [pendingTasksCount, setPendingTasksCount] = useState(0);
  const [overduePaymentsCount, setOverduePaymentsCount] = useState(0);
  const [portalNotificationsCount, setPortalNotificationsCount] = useState(0);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, tenant, logout } = useAuth();
  const { canView, isAdmin, isSuperAdmin } = usePermission();

  // Check if sidebar should be hidden for this user
  const hideSidebar = user?.hide_sidebar || false;

  // Handle window resize for mobile detection
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 768;
      setIsMobile(mobile);
      if (!mobile && mobileMenuVisible) {
        setMobileMenuVisible(false);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [mobileMenuVisible]);

  useEffect(() => {
    let mounted = true;

    const loadPendingTasksCount = async () => {
      try {
        const response = await tasksAPI.getPendingCount();
        if (mounted) {
          setPendingTasksCount(response.data.count || 0);
        }
      } catch (error) {
      }
    };

    if (canView('tasks')) {
      loadPendingTasksCount();
      // Reload count every 2 minutes
      const interval = setInterval(loadPendingTasksCount, 120000);
      return () => {
        mounted = false;
        clearInterval(interval);
      };
    }

    return () => {
      mounted = false;
    };
  }, [canView]);

  useEffect(() => {
    let mounted = true;

    const loadOverduePaymentsCount = async () => {
      try {
        const response = await paymentsAPI.getOverdueCount();
        if (mounted) {
          setOverduePaymentsCount(response.data.count || 0);
        }
      } catch (error) {
      }
    };

    if (canView('payments')) {
      loadOverduePaymentsCount();
      // Reload count every 2 minutes
      const interval = setInterval(loadOverduePaymentsCount, 120000);
      return () => {
        mounted = false;
        clearInterval(interval);
      };
    }

    return () => {
      mounted = false;
    };
  }, [canView]);

  useEffect(() => {
    let mounted = true;

    const loadPortalNotificationsCount = async () => {
      try {
        const response = await portalNotificationsAPI.getCount();
        if (mounted) {
          setPortalNotificationsCount(response.data.count || 0);
        }
      } catch (error) {
      }
    };

    if (canView('appointments')) {
      loadPortalNotificationsCount();
      // Reload count every 2 minutes
      const interval = setInterval(loadPortalNotificationsCount, 120000);
      return () => {
        mounted = false;
        clearInterval(interval);
      };
    }

    return () => {
      mounted = false;
    };
  }, [canView]);

  // Calculate trial days remaining
  const getTrialInfo = () => {
    if (!tenant) return null;

    if (tenant.subscription_status === 'trialing' && tenant.trial_ends_at) {
      const trialEnd = new Date(tenant.trial_ends_at);
      const now = new Date();
      const diffTime = trialEnd - now;
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

      if (diffDays > 0) {
        return { status: 'trialing', daysRemaining: diffDays };
      } else {
        return { status: 'expired', daysRemaining: 0 };
      }
    }

    if (tenant.subscription_status === 'past_due') {
      return { status: 'past_due', daysRemaining: 0 };
    }

    return null;
  };

  const trialInfo = getTrialInfo();

  // Filter menu items based on permissions
  const allMenuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
      permission: 'dashboard',
    },
    {
      key: '/attendance',
      icon: <AppstoreOutlined />,
      label: 'Atendimento',
      permission: 'appointments',
    },
    {
      key: '/appointments',
      icon: <CalendarOutlined />,
      label: 'Agenda',
      permission: 'appointments',
    },
    {
      key: '/portal-notifications',
      icon: <BellOutlined />,
      label: (
        <Badge count={portalNotificationsCount} offset={[10, 0]} size="small">
          Notif. Portal
        </Badge>
      ),
      permission: 'appointments',
    },
    {
      key: '/waiting-list',
      icon: <ClockCircleOutlined />,
      label: 'Lista de Espera',
      permission: 'appointments',
    },
    {
      key: '/patients',
      icon: <UserOutlined />,
      label: 'Pacientes',
      permission: 'patients',
    },
    {
      key: '/leads',
      icon: <UsergroupAddOutlined />,
      label: 'Leads',
      permission: 'leads',
    },
    {
      key: '/medical-records',
      icon: <MedicineBoxOutlined />,
      label: 'Prontuários',
      permission: 'medical_records',
    },
    {
      key: '/prescriptions',
      icon: <FormOutlined />,
      label: 'Receituário',
      permission: 'prescriptions',
    },
    {
      key: '/exams',
      icon: <FileOutlined />,
      label: 'Exames',
      permission: 'exams',
    },
    {
      key: '/consent-templates',
      icon: <FileTextOutlined />,
      label: 'Termos de Consentimento',
      permission: 'clinical_records',
    },
    {
      key: '/tasks',
      icon: <CheckSquareOutlined />,
      label: (
        <Badge count={pendingTasksCount} offset={[10, 0]} size="small">
          Tarefas
        </Badge>
      ),
      permission: 'tasks',
    },
    {
      key: 'financial',
      icon: <DollarOutlined />,
      label: 'Financeiro',
      children: [
        { key: '/budgets', label: 'Orçamentos', permission: 'budgets' },
        { key: '/treatments', label: 'Tratamentos', permission: 'budgets' },
        { key: '/payments', label: 'Pagamentos', permission: 'payments' },
        { key: '/expenses', label: (
          <Badge count={overduePaymentsCount} offset={[10, 0]} size="small">
            Contas a Pagar
          </Badge>
        ), permission: 'payments' },
        { key: '/plans', label: 'Planos', permission: 'plans' },
      ],
    },
    {
      key: 'inventory',
      icon: <ShoppingOutlined />,
      label: 'Estoque',
      children: [
        { key: '/products', label: 'Produtos', permission: 'products' },
        { key: '/suppliers', label: 'Fornecedores', permission: 'suppliers' },
        { key: '/stock-movements', label: 'Movimentações', permission: 'stock_movements' },
      ],
    },
    {
      key: '/campaigns',
      icon: <MessageOutlined />,
      label: 'Campanhas',
      permission: 'campaigns',
    },
    {
      key: '/reports',
      icon: <BarChartOutlined />,
      label: 'Relatórios',
      permission: 'reports',
    },
    // Admin only
    ...(isAdmin ? [{
      key: '/users',
      icon: <UserOutlined />,
      label: 'Usuários',
      adminOnly: true,
    }] : []),
    // Subscription - available to admins
    ...(isAdmin ? [{
      key: '/subscription',
      icon: <CreditCardOutlined />,
      label: 'Assinatura',
      adminOnly: true,
    }] : []),
    // LGPD Compliance (permissão baseada em módulos)
    {
      key: 'lgpd',
      icon: <SafetyOutlined />,
      label: 'LGPD',
      children: [
        { key: '/admin/data-requests', label: 'Solicitações', permission: 'data_requests' },
        { key: '/admin/audit-logs', label: 'Logs de Auditoria', permission: 'audit_logs' },
      ],
    },
    // Certificados Digitais (permissão baseada em módulo)
    {
      key: '/certificates',
      icon: <SafetyCertificateOutlined />,
      label: 'Certificado Digital',
      permission: 'certificates',
    },
    // Super Admin only - Platform Administration
    ...(isSuperAdmin ? [{
      key: '/admin/tenants',
      icon: <CrownOutlined />,
      label: 'Adm Empresas',
      superAdminOnly: true,
    }] : []),
  ];

  // Filter items based on permissions
  const menuItems = allMenuItems.filter(item => {
    // Super Admin-only items
    if (item.superAdminOnly) return isSuperAdmin;

    // Admin-only items
    if (item.adminOnly) return isAdmin;

    // Items with children - filter children first, then show parent only if any child is accessible
    if (item.children) {
      item.children = item.children.filter(child => {
        return child.permission ? canView(child.permission) : true;
      });
      return item.children.length > 0;
    }

    // Items with permission - check if user can view
    if (item.permission) {
      return canView(item.permission);
    }

    // Items without permission and without children are always visible (none currently)
    return true;
  });

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
    if (isMobile) {
      setMobileMenuVisible(false);
    }
  };

  const siderStyle = {
    overflow: 'auto',
    height: '100vh',
    position: 'fixed',
    left: 0,
    top: 0,
    bottom: 0,
    boxShadow: '2px 0 8px rgba(0, 0, 0, 0.08)',
  };

  const logoStyle = {
    height: 64,
    margin: 16,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: collapsed ? 18 : 20,
    fontWeight: 700,
    color: '#66BB6A',
    background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
    WebkitBackgroundClip: 'text',
    WebkitTextFillColor: 'transparent',
    transition: 'all 0.3s',
  };

  const headerStyle = {
    padding: '0 24px',
    background: '#fff',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
    zIndex: 100,
    position: 'sticky',
    top: 0,
  };

  const contentStyle = {
    margin: '24px 16px',
    padding: 24,
    background: 'transparent',
    minHeight: 280,
  };

  // Sidebar content
  const sidebarContent = (
    <>
      <div style={logoStyle}>
        {collapsed ? '🦷' : '🦷 OdoWell'}
      </div>
      <Menu
        theme="light"
        mode="inline"
        selectedKeys={[location.pathname]}
        defaultOpenKeys={['financial', 'inventory']}
        items={menuItems}
        onClick={handleMenuClick}
        style={{
          borderRight: 'none',
        }}
      />
    </>
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* Desktop Sidebar - hidden if hideSidebar is true */}
      {!isMobile && !hideSidebar && (
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          theme="light"
          breakpoint="lg"
          trigger={null}
          style={siderStyle}
          width={240}
        >
          <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            <div style={{ flex: 1, overflowY: 'auto', overflowX: 'hidden' }}>
              {sidebarContent}
            </div>
            <div
              style={{
                height: 48,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderTop: '1px solid #f0f0f0',
                cursor: 'pointer',
                transition: 'all 0.3s',
                backgroundColor: '#fff',
              }}
              onClick={() => setCollapsed(!collapsed)}
            >
              {collapsed ? <MenuUnfoldOutlined style={{ fontSize: 16, color: '#66BB6A' }} /> : <MenuFoldOutlined style={{ fontSize: 16, color: '#66BB6A' }} />}
            </div>
          </div>
        </Sider>
      )}

      {/* Mobile Drawer - also hidden if hideSidebar is true */}
      {isMobile && !hideSidebar && (
        <Drawer
          placement="left"
          onClose={() => setMobileMenuVisible(false)}
          open={mobileMenuVisible}
          bodyStyle={{ padding: 0 }}
          width={240}
        >
          {sidebarContent}
        </Drawer>
      )}

      <Layout style={{ marginLeft: (isMobile || hideSidebar) ? 0 : (collapsed ? 80 : 240), transition: 'all 0.3s' }}>
        <Header style={headerStyle}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {isMobile && !hideSidebar && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                onClick={() => setMobileMenuVisible(true)}
                style={{ fontSize: 18 }}
              />
            )}
            <Text strong style={{ fontSize: isMobile ? 14 : 16 }}>
              {tenant?.name || 'OdoWell'}
            </Text>
          </div>

          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight">
            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', gap: 8 }}>
              <Avatar
                icon={<UserOutlined />}
                src={user?.profile_picture ? `${API_URL}/${user.profile_picture}` : null}
                style={{ backgroundColor: '#66BB6A' }}
                size={isMobile ? 32 : 40}
              />
              {!isMobile && <Text>{user?.name}</Text>}
            </div>
          </Dropdown>
        </Header>

        <Content style={contentStyle}>
          {/* Trial/Subscription Warning Banner */}
          {trialInfo && isAdmin && (
            <Alert
              message={
                trialInfo.status === 'trialing'
                  ? `Periodo de teste: ${trialInfo.daysRemaining} dia${trialInfo.daysRemaining > 1 ? 's' : ''} restante${trialInfo.daysRemaining > 1 ? 's' : ''}`
                  : trialInfo.status === 'expired'
                  ? 'Seu periodo de teste expirou'
                  : 'Pagamento pendente'
              }
              description={
                trialInfo.status === 'trialing'
                  ? 'Assine agora para continuar usando o sistema apos o periodo de teste.'
                  : 'Assine ou regularize o pagamento para continuar usando o sistema.'
              }
              type={trialInfo.status === 'trialing' ? 'warning' : 'error'}
              showIcon
              icon={<WarningOutlined />}
              action={
                <Button
                  size="small"
                  type="primary"
                  onClick={() => navigate('/subscription')}
                >
                  Ver Planos
                </Button>
              }
              style={{ marginBottom: 16 }}
              closable
            />
          )}
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default DashboardLayout;
