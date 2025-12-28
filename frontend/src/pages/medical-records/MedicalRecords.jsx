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
  DatePicker,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { medicalRecordsAPI, patientsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { RangePicker } = DatePicker;

const MedicalRecords = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState([]);
  const [patients, setPatients] = useState([]);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // Filters
  const [filters, setFilters] = useState({
    patient_id: undefined,
    type: undefined,
    search: '',
    date_range: null,
  });

  const recordTypes = [
    { value: 'anamnesis', label: 'Anamnese', color: 'blue' },
    { value: 'treatment', label: 'Tratamento', color: 'green' },
    { value: 'procedure', label: 'Procedimento', color: 'purple' },
    { value: 'prescription', label: 'Receita', color: 'orange' },
    { value: 'certificate', label: 'Atestado', color: 'red' },
  ];

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    fetchRecords();
    fetchPatients();
  }, [pagination.current, pagination.pageSize, filters]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
    }
  };

  const fetchRecords = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await medicalRecordsAPI.getAll(params);
      setRecords(response.data.records || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar prontuários');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await medicalRecordsAPI.delete(id);
      message.success('Prontuário excluído com sucesso');
      fetchRecords();
    } catch (error) {
      message.error('Erro ao excluir prontuário');
    }
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getTypeTag = (type) => {
    const typeObj = recordTypes.find((t) => t.value === type);
    return typeObj ? (
      <Tag color={typeObj.color}>{typeObj.label}</Tag>
    ) : (
      <Tag>{type}</Tag>
    );
  };

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (records.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum prontuário encontrado</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {records.map((record) => {
          const typeObj = recordTypes.find((t) => t.value === record.type);
          return (
            <Card
              key={record.id}
              size="small"
              style={{ borderLeft: `4px solid ${typeObj?.color || '#1890ff'}` }}
              bodyStyle={{ padding: '12px' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.patient?.name}</div>
                {getTypeTag(record.type)}
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
                <div><strong>Data:</strong> {dayjs(record.created_at).format('DD/MM/YYYY')}</div>
                <div><strong>Profissional:</strong> {record.dentist?.name || '-'}</div>
                {record.diagnosis && (
                  <div style={{ gridColumn: '1 / -1' }}><strong>Diagnóstico:</strong> {record.diagnosis}</div>
                )}
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
                <Button
                  type="text"
                  size="small"
                  icon={<EyeOutlined />}
                  onClick={() => navigate(`/medical-records/${record.id}/view`)}
                  style={{ color: actionColors.view }}
                >
                  Ver
                </Button>
                <Button
                  type="text"
                  size="small"
                  icon={<EditOutlined />}
                  onClick={() => navigate(`/medical-records/${record.id}/edit`)}
                  style={{ color: actionColors.edit }}
                >
                  Editar
                </Button>
                <Popconfirm title="Excluir prontuário?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                  <Button type="text" size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
                </Popconfirm>
              </div>
            </Card>
          );
        })}
        <div style={{ textAlign: 'center', padding: '16px' }}>
          <span style={{ color: '#666' }}>Mostrando {records.length} de {pagination.total} prontuários</span>
        </div>
      </div>
    );
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
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
      width: 140,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'Profissional',
      dataIndex: ['dentist', 'name'],
      key: 'dentist',
      ellipsis: true,
    },
    {
      title: 'Diagnóstico',
      dataIndex: 'diagnosis',
      key: 'diagnosis',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 150,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/medical-records/${record.id}/view`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/medical-records/${record.id}/edit`)}
            title="Editar"
            style={{ color: actionColors.edit }}
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
            <span>Prontuários Médicos</span>
          </Space>
        }
        extra={
          <Button
            icon={<PlusOutlined />}
            onClick={() => navigate('/medical-records/new')}
            style={{
              backgroundColor: actionColors.create,
              borderColor: actionColors.create,
              color: '#fff'
            }}
          >
            <span className="btn-text-desktop">Novo Prontuário</span>
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
              placeholder="Tipo de registro"
              style={{ width: '100%' }}
              allowClear
              value={filters.type}
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              {recordTypes.map((type) => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Input
              placeholder="Buscar por diagnóstico..."
              prefix={<SearchOutlined />}
              allowClear
              value={filters.search}
              onChange={(e) =>
                setFilters({ ...filters, search: e.target.value })
              }
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Button onClick={fetchRecords} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        {isMobile ? renderMobileCards() : (
          <div style={{ overflowX: 'auto' }}>
            <Table
              columns={columns}
              dataSource={records}
              rowKey="id"
              loading={loading}
              pagination={pagination}
              onChange={handleTableChange}
              scroll={{ x: 'max-content' }}
            />
          </div>
        )}
      </Card>
    </div>
  );
};

export default MedicalRecords;
