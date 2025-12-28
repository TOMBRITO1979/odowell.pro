import React, { useState, useEffect } from 'react';
import { Table, Card, Tag, Button, Space, Typography, Statistic, Row, Col, Switch, Modal, Select, InputNumber, Input, message, Drawer, Alert, List, Badge } from 'antd';
import {
  ShopOutlined,
  UserOutlined,
  TeamOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  SettingOutlined,
  MailOutlined,
  WarningOutlined,
  StopOutlined,
  DeleteOutlined
} from '@ant-design/icons';
import { adminAPI } from '../../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;

const AdminTenants = () => {
  const [tenants, setTenants] = useState([]);
  const [dashboard, setDashboard] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [selectedTenant, setSelectedTenant] = useState(null);
  const [usersDrawerVisible, setUsersDrawerVisible] = useState(false);
  const [tenantUsers, setTenantUsers] = useState([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editForm, setEditForm] = useState({});
  const [unverifiedTenants, setUnverifiedTenants] = useState([]);
  const [expiringTrials, setExpiringTrials] = useState([]);
  const [inactiveTenants, setInactiveTenants] = useState([]);
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [tenantToDelete, setTenantToDelete] = useState(null);
  const [deleteConfirmName, setDeleteConfirmName] = useState('');
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [tenantsRes, dashboardRes, unverifiedRes, expiringRes, inactiveRes] = await Promise.all([
        adminAPI.getAllTenants(),
        adminAPI.getDashboard(),
        adminAPI.getUnverifiedTenants(),
        adminAPI.getExpiringTrials(),
        adminAPI.getInactiveTenants()
      ]);
      setTenants(tenantsRes.data.tenants || []);
      setDashboard(dashboardRes.data);
      setUnverifiedTenants(unverifiedRes.data.tenants || []);
      setExpiringTrials(expiringRes.data.tenants || []);
      setInactiveTenants(inactiveRes.data.tenants || []);
    } catch (error) {
      message.error('Erro ao carregar dados');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleActive = async (tenant, active) => {
    try {
      await adminAPI.updateTenantStatus(tenant.id, { active });
      message.success(`Clínica ${active ? 'ativada' : 'desativada'} com sucesso`);
      loadData();
    } catch (error) {
      message.error('Erro ao atualizar status');
    }
  };

  const handleViewUsers = async (tenant) => {
    setSelectedTenant(tenant);
    setUsersDrawerVisible(true);
    setUsersLoading(true);
    try {
      const response = await adminAPI.getTenantUsers(tenant.id);
      setTenantUsers(response.data.users || []);
    } catch (error) {
      message.error('Erro ao carregar usuários');
    } finally {
      setUsersLoading(false);
    }
  };

  const handleToggleUserActive = async (user, active) => {
    try {
      await adminAPI.updateTenantUserStatus(selectedTenant.id, user.id, { active });
      message.success(`Usuário ${active ? 'ativado' : 'desativado'} com sucesso`);
      // Reload users
      const response = await adminAPI.getTenantUsers(selectedTenant.id);
      setTenantUsers(response.data.users || []);
    } catch (error) {
      message.error('Erro ao atualizar usuário');
    }
  };

  const handleEditTenant = (tenant) => {
    setSelectedTenant(tenant);
    setEditForm({
      subscription_status: tenant.subscription_status,
      plan_type: tenant.plan_type,
      patient_limit: tenant.patient_limit,
    });
    setEditModalVisible(true);
  };

  const handleSaveEdit = async () => {
    try {
      await adminAPI.updateTenantStatus(selectedTenant.id, editForm);
      message.success('Clínica atualizada com sucesso');
      setEditModalVisible(false);
      loadData();
    } catch (error) {
      message.error('Erro ao atualizar clínica');
    }
  };

  const handleOpenDeleteModal = (tenant) => {
    setTenantToDelete(tenant);
    setDeleteConfirmName('');
    setDeleteModalVisible(true);
  };

  const handleDeleteTenant = async () => {
    if (deleteConfirmName !== tenantToDelete?.name) {
      message.error('O nome digitado não confere');
      return;
    }

    try {
      setDeleting(true);
      await adminAPI.deleteTenant(tenantToDelete.id);
      message.success('Clínica deletada com sucesso');
      setDeleteModalVisible(false);
      setTenantToDelete(null);
      setDeleteConfirmName('');
      loadData();
    } catch (error) {
      message.error('Erro ao deletar clínica');
    } finally {
      setDeleting(false);
    }
  };

  const getStatusTag = (status) => {
    const statusConfig = {
      active: { color: 'success', text: 'Ativo', icon: <CheckCircleOutlined /> },
      trialing: { color: 'processing', text: 'Trial', icon: <ClockCircleOutlined /> },
      past_due: { color: 'warning', text: 'Pagamento Pendente', icon: <ExclamationCircleOutlined /> },
      canceled: { color: 'error', text: 'Cancelado', icon: <CloseCircleOutlined /> },
      expired: { color: 'error', text: 'Expirado', icon: <CloseCircleOutlined /> },
    };
    const config = statusConfig[status] || { color: 'default', text: status, icon: null };
    return <Tag color={config.color} icon={config.icon}>{config.text}</Tag>;
  };

  const getPlanTag = (plan) => {
    const planConfig = {
      bronze: { color: '#CD7F32', text: 'Bronze' },
      silver: { color: '#C0C0C0', text: 'Prata' },
      gold: { color: '#FFD700', text: 'Ouro' },
    };
    const config = planConfig[plan] || { color: 'default', text: plan };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const formatPrice = (cents) => {
    if (!cents) return 'R$ 0,00';
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(cents / 100);
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (tenants.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhuma clinica cadastrada</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {tenants.map((record) => (
          <Card
            key={record.id}
            size="small"
            style={{ borderLeft: `4px solid ${record.active ? (record.subscription_status === 'active' ? '#52c41a' : '#faad14') : '#d9d9d9'}` }}
            bodyStyle={{ padding: '12px' }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
              <div>
                <div style={{ fontWeight: 600, fontSize: '15px' }}>{record.name}</div>
                <Text type="secondary" style={{ fontSize: 12 }}>{record.email}</Text>
              </div>
              <Switch
                size="small"
                checked={record.active}
                onChange={(checked) => handleToggleActive(record, checked)}
                checkedChildren="Ativo"
                unCheckedChildren="Inativo"
              />
            </div>
            <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap', marginBottom: '8px' }}>
              {getStatusTag(record.subscription_status)}
              {getPlanTag(record.plan_type)}
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
              <div><strong>Usuarios:</strong> <Tag icon={<TeamOutlined />} size="small">{record.user_count || 0}</Tag></div>
              <div><strong>Pacientes:</strong> {record.patient_count || 0}/{record.patient_limit?.toLocaleString()}</div>
              <div><strong>Criado:</strong> {dayjs(record.created_at).format('DD/MM/YYYY')}</div>
              {record.subscription_details?.price_monthly && (
                <div><strong>Valor:</strong> {formatPrice(record.subscription_details.price_monthly)}</div>
              )}
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)', flexWrap: 'wrap' }}>
              <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => handleViewUsers(record)}>Usuarios</Button>
              <Button type="text" size="small" icon={<SettingOutlined />} onClick={() => handleEditTenant(record)}>Editar</Button>
              <Button type="text" size="small" danger icon={<DeleteOutlined />} onClick={() => handleOpenDeleteModal(record)}>Deletar</Button>
            </div>
          </Card>
        ))}
      </div>
    );
  };

  const columns = [
    {
      title: 'Clínica',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{text}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.email}</Text>
        </Space>
      ),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => (
        <Space direction="vertical" size={4}>
          {getStatusTag(record.subscription_status)}
          {getPlanTag(record.plan_type)}
        </Space>
      ),
    },
    {
      title: 'Usuários',
      dataIndex: 'user_count',
      key: 'user_count',
      render: (count) => <Tag icon={<TeamOutlined />}>{count || 0}</Tag>,
    },
    {
      title: 'Pacientes',
      key: 'patients',
      render: (_, record) => (
        <Text>{record.patient_count || 0} / {record.patient_limit?.toLocaleString()}</Text>
      ),
    },
    {
      title: 'Último Pagamento',
      key: 'last_payment',
      render: (_, record) => {
        const sub = record.subscription_details;
        if (!sub) return <Text type="secondary">-</Text>;
        return (
          <Space direction="vertical" size={0}>
            <Text>{formatPrice(sub.price_monthly)}</Text>
            {sub.current_period_start && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {dayjs(sub.current_period_start).format('DD/MM/YYYY')}
              </Text>
            )}
          </Space>
        );
      },
    },
    {
      title: 'Criado em',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Ativo',
      key: 'active',
      render: (_, record) => (
        <Switch
          checked={record.active}
          onChange={(checked) => handleToggleActive(record, checked)}
          checkedChildren="Sim"
          unCheckedChildren="Não"
        />
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewUsers(record)}
          >
            Usuários
          </Button>
          <Button
            type="text"
            icon={<SettingOutlined />}
            onClick={() => handleEditTenant(record)}
          >
            Editar
          </Button>
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleOpenDeleteModal(record)}
          >
            Deletar
          </Button>
        </Space>
      ),
    },
  ];

  const userColumns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Cargo',
      dataIndex: 'role',
      key: 'role',
      render: (role) => {
        const roleNames = {
          admin: 'Administrador',
          dentist: 'Dentista',
          receptionist: 'Recepcionista',
          user: 'Usuário',
        };
        return roleNames[role] || role;
      },
    },
    {
      title: 'Ativo',
      key: 'active',
      render: (_, record) => (
        <Switch
          checked={record.active}
          onChange={(checked) => handleToggleUserActive(record, checked)}
          checkedChildren="Sim"
          unCheckedChildren="Não"
        />
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>
        <ShopOutlined style={{ marginRight: 8 }} />
        Administração do Sistema
      </Title>

      {/* Dashboard Stats */}
      {dashboard && (
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Total Clínicas"
                value={dashboard.total_tenants}
                prefix={<ShopOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Clínicas Ativas"
                value={dashboard.active_tenants}
                prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Em Trial"
                value={dashboard.trial_tenants}
                prefix={<ClockCircleOutlined style={{ color: '#1890ff' }} />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Assinaturas Ativas"
                value={dashboard.active_subscriptions}
                prefix={<DollarOutlined style={{ color: '#52c41a' }} />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Total Usuários"
                value={dashboard.total_users}
                prefix={<UserOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Receita Mensal"
                value={formatPrice(dashboard.monthly_revenue)}
                prefix={<DollarOutlined style={{ color: '#faad14' }} />}
                valueStyle={{ fontSize: 20 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Não Verificados"
                value={dashboard.unverified_tenants || 0}
                prefix={<MailOutlined style={{ color: dashboard.unverified_tenants > 0 ? '#faad14' : '#52c41a' }} />}
                valueStyle={{ color: dashboard.unverified_tenants > 0 ? '#faad14' : '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Expirando (24h)"
                value={dashboard.trial_stats?.expiring_soon || 0}
                prefix={<WarningOutlined style={{ color: dashboard.trial_stats?.expiring_soon > 0 ? '#ff4d4f' : '#52c41a' }} />}
                valueStyle={{ color: dashboard.trial_stats?.expiring_soon > 0 ? '#ff4d4f' : '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Card>
              <Statistic
                title="Inativos"
                value={dashboard.inactive_stats?.total || 0}
                prefix={<StopOutlined style={{ color: dashboard.inactive_stats?.total > 0 ? '#8c8c8c' : '#52c41a' }} />}
                valueStyle={{ color: dashboard.inactive_stats?.total > 0 ? '#8c8c8c' : '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* Alerts */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {/* Unverified Tenants Alert */}
        {unverifiedTenants.length > 0 && (
          <Col span={12}>
            <Card
              size="small"
              title={
                <Space>
                  <MailOutlined style={{ color: '#faad14' }} />
                  <span>Emails Não Verificados</span>
                  <Badge count={unverifiedTenants.length} style={{ backgroundColor: '#faad14' }} />
                </Space>
              }
            >
              <List
                size="small"
                dataSource={unverifiedTenants.slice(0, 5)}
                renderItem={(item) => (
                  <List.Item>
                    <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                      <div>
                        <Text strong>{item.name}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{item.email}</Text>
                      </div>
                      <Tag color="warning">
                        {Math.round(item.hours_since_creation)}h sem verificar
                      </Tag>
                    </Space>
                  </List.Item>
                )}
              />
              {unverifiedTenants.length > 5 && (
                <Text type="secondary">... e mais {unverifiedTenants.length - 5} clínica(s)</Text>
              )}
            </Card>
          </Col>
        )}

        {/* Expiring Trials Alert */}
        {expiringTrials.length > 0 && (
          <Col span={12}>
            <Card
              size="small"
              title={
                <Space>
                  <WarningOutlined style={{ color: '#ff4d4f' }} />
                  <span>Trials Expirando (24h)</span>
                  <Badge count={expiringTrials.length} style={{ backgroundColor: '#ff4d4f' }} />
                </Space>
              }
            >
              <List
                size="small"
                dataSource={expiringTrials.slice(0, 5)}
                renderItem={(item) => (
                  <List.Item>
                    <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                      <div>
                        <Text strong>{item.name}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{item.email}</Text>
                      </div>
                      <Tag color="error">
                        {Math.round(item.hours_left)}h restantes
                      </Tag>
                    </Space>
                  </List.Item>
                )}
              />
              {expiringTrials.length > 5 && (
                <Text type="secondary">... e mais {expiringTrials.length - 5} clínica(s)</Text>
              )}
            </Card>
          </Col>
        )}

        {/* Inactive Tenants Alert (expired/canceled/past_due) */}
        {inactiveTenants.length > 0 && (
          <Col span={24}>
            <Card
              size="small"
              title={
                <Space>
                  <CloseCircleOutlined style={{ color: '#8c8c8c' }} />
                  <span>Não Ativaram / Cancelaram</span>
                  <Badge count={inactiveTenants.length} style={{ backgroundColor: '#8c8c8c' }} />
                </Space>
              }
            >
              <List
                size="small"
                dataSource={inactiveTenants.slice(0, 10)}
                renderItem={(item) => (
                  <List.Item>
                    <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                      <div>
                        <Text strong>{item.name}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{item.email}</Text>
                      </div>
                      <Space>
                        <Tag color={
                          item.subscription_status === 'expired' ? 'red' :
                          item.subscription_status === 'canceled' ? 'orange' :
                          item.subscription_status === 'past_due' ? 'gold' : 'default'
                        }>
                          {item.reason}
                        </Tag>
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          {dayjs(item.updated_at).format('DD/MM/YY')}
                        </Text>
                      </Space>
                    </Space>
                  </List.Item>
                )}
              />
              {inactiveTenants.length > 10 && (
                <Text type="secondary">... e mais {inactiveTenants.length - 10} clínica(s)</Text>
              )}
            </Card>
          </Col>
        )}
      </Row>

      {/* Tenants Table */}
      <Card title="Clínicas Cadastradas">
        {isMobile ? renderMobileCards() : (
          <Table
            columns={columns}
            dataSource={tenants}
            rowKey="id"
            loading={loading}
            pagination={{
              defaultPageSize: 20,
              pageSizeOptions: ['20', '50', '100'],
              showSizeChanger: true,
              showTotal: (total, range) => `${range[0]}-${range[1]} de ${total} clínicas`,
            }}
          />
        )}
      </Card>

      {/* Users Drawer */}
      <Drawer
        title={`Usuários - ${selectedTenant?.name}`}
        placement="right"
        width={600}
        onClose={() => setUsersDrawerVisible(false)}
        open={usersDrawerVisible}
      >
        <Table
          columns={userColumns}
          dataSource={tenantUsers}
          rowKey="id"
          loading={usersLoading}
          pagination={false}
        />
      </Drawer>

      {/* Edit Modal */}
      <Modal
        title={`Editar - ${selectedTenant?.name}`}
        open={editModalVisible}
        onOk={handleSaveEdit}
        onCancel={() => setEditModalVisible(false)}
        okText="Salvar"
        cancelText="Cancelar"
      >
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          <div>
            <Text strong>Status da Assinatura</Text>
            <Select
              style={{ width: '100%', marginTop: 8 }}
              value={editForm.subscription_status}
              onChange={(value) => setEditForm({ ...editForm, subscription_status: value })}
            >
              <Option value="active">Ativo</Option>
              <Option value="trialing">Trial</Option>
              <Option value="past_due">Pagamento Pendente</Option>
              <Option value="canceled">Cancelado</Option>
              <Option value="expired">Expirado</Option>
            </Select>
          </div>
          <div>
            <Text strong>Plano</Text>
            <Select
              style={{ width: '100%', marginTop: 8 }}
              value={editForm.plan_type}
              onChange={(value) => {
                const limits = { bronze: 1000, silver: 2500, gold: 5000 };
                setEditForm({ ...editForm, plan_type: value, patient_limit: limits[value] || 1000 });
              }}
            >
              <Option value="bronze">Bronze (1.000 pacientes)</Option>
              <Option value="silver">Prata (2.500 pacientes)</Option>
              <Option value="gold">Ouro (5.000 pacientes)</Option>
            </Select>
          </div>
          <div>
            <Text strong>Limite de Pacientes</Text>
            <InputNumber
              style={{ width: '100%', marginTop: 8 }}
              value={editForm.patient_limit}
              onChange={(value) => setEditForm({ ...editForm, patient_limit: value })}
              min={100}
              max={100000}
            />
          </div>
        </Space>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        title={
          <Space>
            <DeleteOutlined style={{ color: '#ff4d4f' }} />
            <span>Deletar Empresa</span>
          </Space>
        }
        open={deleteModalVisible}
        onCancel={() => {
          setDeleteModalVisible(false);
          setTenantToDelete(null);
          setDeleteConfirmName('');
        }}
        footer={[
          <Button key="cancel" onClick={() => setDeleteModalVisible(false)}>
            Cancelar
          </Button>,
          <Button
            key="delete"
            type="primary"
            danger
            icon={<DeleteOutlined />}
            disabled={deleteConfirmName !== tenantToDelete?.name}
            loading={deleting}
            onClick={handleDeleteTenant}
          >
            Deletar Empresa
          </Button>,
        ]}
      >
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          <Alert
            type="warning"
            showIcon
            message="Esta acao ira desativar permanentemente a empresa e todos os seus usuarios."
            description="Os dados serao mantidos no banco de dados, mas a empresa nao podera mais acessar o sistema."
          />
          <div>
            <Text>Para confirmar, digite o nome da empresa:</Text>
            <Text strong style={{ display: 'block', marginTop: 4, marginBottom: 8 }}>
              {tenantToDelete?.name}
            </Text>
            <Input
              placeholder="Digite o nome da empresa"
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
              status={deleteConfirmName && deleteConfirmName !== tenantToDelete?.name ? 'error' : ''}
            />
          </div>
        </Space>
      </Modal>
    </div>
  );
};

export default AdminTenants;
