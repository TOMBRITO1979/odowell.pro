import React, { useState, useEffect } from 'react';
import { Table, Button, Space, Tag, message, Modal, Form, Input, Select } from 'antd';
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
  const [form] = Form.useForm();
  const { isAdmin } = usePermission();

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

      <div style={{ overflowX: 'auto' }}>
        <Table
          columns={columns}
          dataSource={users}
          loading={loading}
          rowKey="id"
          scroll={{ x: 'max-content' }}
        />
      </div>

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
