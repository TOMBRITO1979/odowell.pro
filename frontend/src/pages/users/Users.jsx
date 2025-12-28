import React, { useState, useEffect } from 'react';
import { Table, Button, Space, Tag, message, Modal, Form, Input, Select, Card } from 'antd';
import { UserOutlined, KeyOutlined, PlusOutlined } from '@ant-design/icons';
import { usersAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors } from '../../theme/designSystem';
import UserPermissions from './UserPermissions';

const { Option } = Select;

const Users = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [permissionsVisible, setPermissionsVisible] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [form] = Form.useForm();
  const { isAdmin } = usePermission();

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await usersAPI.getAll();
      setUsers(response.data.users || []);
    } catch (error) {
      message.error('Erro ao carregar usuários');
    } finally {
      setLoading(false);
    }
  };

  const handleManagePermissions = (user) => {
    setSelectedUser(user);
    setPermissionsVisible(true);
  };

  const handleCreateUser = async (values) => {
    try {
      await usersAPI.create(values);
      message.success('Usuário criado com sucesso');
      setCreateModalVisible(false);
      form.resetFields();
      fetchUsers();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar usuário');
    }
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (users.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum usuário encontrado</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {users.map((record) => (
          <Card key={record.id} size="small" style={{ borderLeft: `4px solid ${record.role === 'admin' ? '#ff4d4f' : record.role === 'dentist' ? '#1890ff' : '#52c41a'}` }} bodyStyle={{ padding: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
              <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.name}</div>
              <Tag color={record.active ? 'success' : 'default'}>{record.active ? 'Ativo' : 'Inativo'}</Tag>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
              <div style={{ gridColumn: '1 / -1' }}><strong>Email:</strong> {record.email}</div>
              <div><strong>Função:</strong><br /><Tag color={record.role === 'admin' ? 'red' : record.role === 'dentist' ? 'blue' : 'green'}>{record.role}</Tag></div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <Button size="small" icon={<KeyOutlined />} onClick={() => handleManagePermissions(record)}>Permissões</Button>
            </div>
          </Card>
        ))}
      </div>
    );
  };

  const columns = [
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
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role) => (
        <Tag color={role === 'admin' ? 'red' : role === 'dentist' ? 'blue' : 'green'}>
          {role}
        </Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      render: (active) => (
        <Tag color={active ? 'success' : 'default'}>
          {active ? 'Ativo' : 'Inativo'}
        </Tag>
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            icon={<KeyOutlined />}
            onClick={() => handleManagePermissions(record)}
          >
            Permissões
          </Button>
        </Space>
      ),
    },
  ];

  if (!isAdmin) {
    return <div>Acesso negado. Apenas administradores podem gerenciar usuários.</div>;
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2><UserOutlined /> Gerenciamento de Usuários</h2>
        <Button
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
          style={{
            backgroundColor: actionColors.create,
            borderColor: actionColors.create,
            color: '#fff'
          }}
        >
          Novo Usuário
        </Button>
      </div>

      {isMobile ? renderMobileCards() : (
        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={users}
            loading={loading}
            rowKey="id"
            scroll={{ x: 'max-content' }}
          />
        </div>
      )}

      <Modal
        title="Novo Usuário"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateUser}>
          <Form.Item name="name" label="Nome" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="Senha" rules={[{ required: true, min: 6 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="Função" rules={[{ required: true }]}>
            <Select>
              <Option value="admin">Admin</Option>
              <Option value="dentist">Dentista</Option>
              <Option value="receptionist">Recepcionista</Option>
              <Option value="user">Usuário</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {permissionsVisible && (
        <UserPermissions
          user={selectedUser}
          visible={permissionsVisible}
          onClose={() => {
            setPermissionsVisible(false);
            setSelectedUser(null);
          }}
        />
      )}
    </div>
  );
};

export default Users;
