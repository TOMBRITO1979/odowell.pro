import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Card,
  message,
  Popconfirm,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  FileTextOutlined,
  FilePdfOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { prescriptionsAPI, patientsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const Prescriptions = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [prescriptions, setPrescriptions] = useState([]);
  const [patients, setPatients] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // Filters
  const [filters, setFilters] = useState({
    patient_id: undefined,
    type: undefined,
    status: undefined,
    dentist_id: undefined,
  });

  const prescriptionTypes = [
    { value: 'prescription', label: 'Receita', color: statusColors.inProgress },
    { value: 'medical_report', label: 'Laudo Médico', color: statusColors.success },
    { value: 'certificate', label: 'Atestado', color: statusColors.pending },
    { value: 'referral', label: 'Encaminhamento', color: '#a855f7' },
  ];

  const statusOptions = [
    { value: 'draft', label: 'Rascunho', color: statusColors.cancelled },
    { value: 'issued', label: 'Emitido', color: statusColors.success },
    { value: 'cancelled', label: 'Cancelado', color: statusColors.error },
  ];

  useEffect(() => {
    fetchPrescriptions();
    fetchPatients();
  }, [pagination.current, filters]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchPrescriptions = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await prescriptionsAPI.getAll(params);
      setPrescriptions(response.data.prescriptions || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar receituário');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await prescriptionsAPI.delete(id);
      message.success('Receita excluída com sucesso');
      fetchPrescriptions();
    } catch (error) {
      message.error('Erro ao excluir receita');
    }
  };

  const handleDownloadPDF = async (id) => {
    try {
      const response = await prescriptionsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `receita_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
    }
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getTypeTag = (type) => {
    const typeObj = prescriptionTypes.find((t) => t.value === type);
    return typeObj ? (
      <Tag color={typeObj.color}>{typeObj.label}</Tag>
    ) : (
      <Tag>{type}</Tag>
    );
  };

  const getStatusTag = (status) => {
    const statusObj = statusOptions.find((s) => s.value === status);
    return statusObj ? (
      <Tag color={statusObj.color}>{statusObj.label}</Tag>
    ) : (
      <Tag>{status}</Tag>
    );
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 100,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
      sorter: true,
    },
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient',
      ellipsis: true,
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: 130,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'Título',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: 'Dentista',
      dataIndex: 'dentist_name',
      key: 'dentist_name',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 130,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/prescriptions/${record.id}`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/prescriptions/${record.id}/edit`)}
            title="Editar"
            style={{ color: actionColors.edit }}
          />
          <Button
            type="text"
            icon={<FilePdfOutlined />}
            onClick={() => handleDownloadPDF(record.id)}
            title="Gerar PDF"
            style={{ color: actionColors.exportPDF }}
          />
          <Popconfirm
            title="Tem certeza que deseja excluir?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button
              type="text"
              icon={<DeleteOutlined />}
              title="Excluir"
              style={{ color: actionColors.delete }}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            <span>Receituário</span>
          </Space>
        }
        extra={
          <Button
            icon={<PlusOutlined />}
            onClick={() => navigate('/prescriptions/new')}
            style={{
              backgroundColor: actionColors.create,
              borderColor: actionColors.create,
              color: '#fff'
            }}
          >
            Nova Receita
          </Button>
        }
        style={{ boxShadow: shadows.small }}
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Selecione o paciente"
              style={{ width: '100%' }}
              allowClear
              showSearch
              filterOption={(input, option) =>
                option.children.toLowerCase().includes(input.toLowerCase())
              }
              value={filters.patient_id}
              onChange={(value) =>
                setFilters({ ...filters, patient_id: value })
              }
            >
              {patients.map((patient) => (
                <Select.Option key={patient.id} value={patient.id}>
                  {patient.name}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Tipo de documento"
              style={{ width: '100%' }}
              allowClear
              value={filters.type}
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              {prescriptionTypes.map((type) => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Status"
              style={{ width: '100%' }}
              allowClear
              value={filters.status}
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              {statusOptions.map((status) => (
                <Select.Option key={status.value} value={status.value}>
                  {status.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Button onClick={fetchPrescriptions} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={prescriptions}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={handleTableChange}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>
    </div>
  );
};

export default Prescriptions;
