import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Input, Space, Popconfirm, message, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SearchOutlined } from '@ant-design/icons';
import { patientsAPI } from '../../services/api';

const Patients = () => {
  const [patients, setPatients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    loadPatients();
  }, [pagination.current, search]);

  const loadPatients = async () => {
    setLoading(true);
    try {
      const response = await patientsAPI.getAll({
        page: pagination.current,
        page_size: pagination.pageSize,
        search,
      });
      setPatients(response.data.patients || []);
      setPagination(prev => ({ ...prev, total: response.data.total }));
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await patientsAPI.delete(id);
      message.success('Paciente excluído com sucesso');
      loadPatients();
    } catch (error) {
      message.error('Erro ao excluir paciente');
    }
  };

  const columns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'CPF',
      dataIndex: 'cpf',
      key: 'cpf',
    },
    {
      title: 'Telefone',
      dataIndex: 'cell_phone',
      key: 'cell_phone',
      render: (text, record) => text || record.phone,
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      render: (active) => (
        <Tag color={active ? 'green' : 'red'}>
          {active ? 'Ativo' : 'Inativo'}
        </Tag>
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            icon={<EyeOutlined />}
            onClick={() => navigate(`/patients/${record.id}`)}
            size="small"
          />
          <Button
            icon={<EditOutlined />}
            onClick={() => navigate(`/patients/${record.id}/edit`)}
            size="small"
          />
          <Popconfirm
            title="Tem certeza que deseja excluir?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button icon={<DeleteOutlined />} danger size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h1>Pacientes</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/patients/new')}>
          Novo Paciente
        </Button>
      </div>

      <Input
        placeholder="Buscar por nome, CPF, email ou telefone"
        prefix={<SearchOutlined />}
        onChange={(e) => setSearch(e.target.value)}
        style={{ marginBottom: 16, maxWidth: 400 }}
      />

      <Table
        columns={columns}
        dataSource={patients}
        rowKey="id"
        loading={loading}
        pagination={pagination}
        onChange={(newPagination) => setPagination(newPagination)}
      />
    </div>
  );
};

export default Patients;
