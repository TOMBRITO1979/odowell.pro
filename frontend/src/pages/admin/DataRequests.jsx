import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Space,
  Button,
  Select,
  Tag,
  Row,
  Col,
  Statistic,
  message,
  Modal,
  Form,
  Input,
  Typography,
  Drawer,
  Timeline,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SafetyOutlined,
  UserOutlined,
  FileTextOutlined,
  EditOutlined,
} from '@ant-design/icons';
import { dataRequestAPI, patientsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import dayjs from 'dayjs';

const { Text, Title } = Typography;
const { TextArea } = Input;

const DataRequests = () => {
  const [requests, setRequests] = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [filters, setFilters] = useState({ status: '', type: '' });
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [patients, setPatients] = useState([]);
  const [searchingPatients, setSearchingPatients] = useState(false);
  const [form] = Form.useForm();
  const [statusForm] = Form.useForm();
  const { canEdit } = usePermission();

  useEffect(() => {
    fetchRequests();
    fetchStats();
  }, []);

  const fetchRequests = async (page = 1, pageSize = 20) => {
    setLoading(true);
    try {
      const params = {
        page,
        page_size: pageSize,
        status: filters.status,
        type: filters.type,
      };
      const response = await dataRequestAPI.getAll(params);
      setRequests(response.data.requests || []);
      setPagination({
        current: response.data.page,
        pageSize: response.data.page_size,
        total: response.data.total,
      });
    } catch (error) {
      message.error('Erro ao carregar solicitacoes');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await dataRequestAPI.getStats();
      setStats(response.data);
    } catch (error) {
      console.error('Erro ao carregar estatisticas:', error);
    }
  };

  const handleTableChange = (pag) => {
    fetchRequests(pag.current, pag.pageSize);
  };

  const handleSearch = () => {
    fetchRequests(1, pagination.pageSize);
  };

  const searchPatients = async (value) => {
    if (value.length < 2) return;
    setSearchingPatients(true);
    try {
      const response = await patientsAPI.getAll({ search: value, page_size: 10 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Erro ao buscar pacientes:', error);
    } finally {
      setSearchingPatients(false);
    }
  };

  const handleCreateRequest = async (values) => {
    try {
      await dataRequestAPI.create(values);
      message.success('Solicitacao criada com sucesso');
      setCreateModalVisible(false);
      form.resetFields();
      fetchRequests();
      fetchStats();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao criar solicitacao');
    }
  };

  const handleUpdateStatus = async (values) => {
    try {
      await dataRequestAPI.updateStatus(selectedRequest.id, values);
      message.success('Status atualizado com sucesso');
      setStatusModalVisible(false);
      statusForm.resetFields();
      fetchRequests();
      fetchStats();
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao atualizar status');
    }
  };

  const showDetails = (record) => {
    setSelectedRequest(record);
    setDrawerVisible(true);
  };

  const showStatusModal = (record) => {
    setSelectedRequest(record);
    statusForm.setFieldsValue({
      status: record.status,
      response_notes: record.response_notes,
      rejection_reason: record.rejection_reason,
    });
    setStatusModalVisible(true);
  };

  const getTypeLabel = (type) => {
    const labels = {
      access: 'Acesso aos Dados',
      portability: 'Portabilidade',
      correction: 'Correcao',
      deletion: 'Exclusao',
      revocation: 'Revogacao de Consentimento',
    };
    return labels[type] || type;
  };

  const getTypeColor = (type) => {
    const colors = {
      access: 'blue',
      portability: 'cyan',
      correction: 'orange',
      deletion: 'red',
      revocation: 'purple',
    };
    return colors[type] || 'default';
  };

  const getStatusLabel = (status) => {
    const labels = {
      pending: 'Pendente',
      in_progress: 'Em Andamento',
      completed: 'Concluido',
      rejected: 'Rejeitado',
    };
    return labels[status] || status;
  };

  const getStatusColor = (status) => {
    const colors = {
      pending: 'orange',
      in_progress: 'blue',
      completed: 'green',
      rejected: 'red',
    };
    return colors[status] || 'default';
  };

  const columns = [
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Paciente',
      dataIndex: 'patient_name',
      key: 'patient_name',
      render: (name, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{name}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.patient_cpf}</Text>
        </Space>
      ),
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: 150,
      render: (type) => <Tag color={getTypeColor(type)}>{getTypeLabel(type)}</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (status) => <Tag color={getStatusColor(status)}>{getStatusLabel(status)}</Tag>,
    },
    {
      title: 'Processado em',
      dataIndex: 'processed_at',
      key: 'processed_at',
      width: 120,
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Acoes',
      key: 'actions',
      width: 100,
      render: (_, record) => (
        <Space>
          <Button type="text" icon={<EyeOutlined />} onClick={() => showDetails(record)} />
          {canEdit('patients') && record.status !== 'completed' && record.status !== 'rejected' && (
            <Button type="text" icon={<EditOutlined />} onClick={() => showStatusModal(record)} />
          )}
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Space align="center">
          <SafetyOutlined style={{ fontSize: 24, color: '#4CAF50' }} />
          <div>
            <Title level={4} style={{ margin: 0 }}>Solicitacoes LGPD</Title>
            <Text type="secondary">Gerenciamento de solicitacoes dos titulares de dados</Text>
          </div>
        </Space>
      </div>

      {/* Statistics */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Total de Solicitacoes"
                value={stats.total}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Pendentes"
                value={stats.pending}
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: stats.pending > 0 ? '#faad14' : '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Text type="secondary">Por Tipo</Text>
              <div style={{ marginTop: 8 }}>
                {stats.by_type?.map(item => (
                  <Tag key={item.type} color={getTypeColor(item.type)} style={{ marginBottom: 4 }}>
                    {getTypeLabel(item.type)}: {item.count}
                  </Tag>
                ))}
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Text type="secondary">Por Status</Text>
              <div style={{ marginTop: 8 }}>
                {stats.by_status?.map(item => (
                  <Tag key={item.status} color={getStatusColor(item.status)} style={{ marginBottom: 4 }}>
                    {getStatusLabel(item.status)}: {item.count}
                  </Tag>
                ))}
              </div>
            </Card>
          </Col>
        </Row>
      )}

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Tipo"
              allowClear
              style={{ width: '100%' }}
              value={filters.type || undefined}
              onChange={(value) => setFilters({ ...filters, type: value || '' })}
            >
              <Select.Option value="access">Acesso aos Dados</Select.Option>
              <Select.Option value="portability">Portabilidade</Select.Option>
              <Select.Option value="correction">Correcao</Select.Option>
              <Select.Option value="deletion">Exclusao</Select.Option>
              <Select.Option value="revocation">Revogacao</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Status"
              allowClear
              style={{ width: '100%' }}
              value={filters.status || undefined}
              onChange={(value) => setFilters({ ...filters, status: value || '' })}
            >
              <Select.Option value="pending">Pendente</Select.Option>
              <Select.Option value="in_progress">Em Andamento</Select.Option>
              <Select.Option value="completed">Concluido</Select.Option>
              <Select.Option value="rejected">Rejeitado</Select.Option>
            </Select>
          </Col>
          <Col>
            <Space>
              <Button type="primary" onClick={handleSearch}>Buscar</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { fetchRequests(); fetchStats(); }}>
                Atualizar
              </Button>
              {canEdit('patients') && (
                <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
                  Nova Solicitacao
                </Button>
              )}
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={requests}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `Total: ${total} solicitacoes`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      {/* Create Modal */}
      <Modal
        title="Nova Solicitacao LGPD"
        open={createModalVisible}
        onCancel={() => { setCreateModalVisible(false); form.resetFields(); }}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateRequest}>
          <Form.Item
            name="patient_id"
            label="Paciente"
            rules={[{ required: true, message: 'Selecione um paciente' }]}
          >
            <Select
              showSearch
              placeholder="Digite para buscar..."
              filterOption={false}
              onSearch={searchPatients}
              loading={searchingPatients}
              notFoundContent={searchingPatients ? 'Buscando...' : 'Nenhum paciente encontrado'}
            >
              {patients.map(p => (
                <Select.Option key={p.id} value={p.id}>
                  {p.name} - {p.cpf}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            name="type"
            label="Tipo de Solicitacao"
            rules={[{ required: true, message: 'Selecione o tipo' }]}
          >
            <Select placeholder="Selecione o tipo">
              <Select.Option value="access">Acesso aos Dados</Select.Option>
              <Select.Option value="portability">Portabilidade dos Dados</Select.Option>
              <Select.Option value="correction">Correcao de Dados</Select.Option>
              <Select.Option value="deletion">Exclusao de Dados</Select.Option>
              <Select.Option value="revocation">Revogacao de Consentimento</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="description" label="Descricao">
            <TextArea rows={3} placeholder="Descreva a solicitacao do paciente..." />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">Criar Solicitacao</Button>
              <Button onClick={() => { setCreateModalVisible(false); form.resetFields(); }}>Cancelar</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Status Update Modal */}
      <Modal
        title="Atualizar Status"
        open={statusModalVisible}
        onCancel={() => { setStatusModalVisible(false); statusForm.resetFields(); }}
        footer={null}
      >
        <Form form={statusForm} layout="vertical" onFinish={handleUpdateStatus}>
          <Form.Item
            name="status"
            label="Novo Status"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="pending">Pendente</Select.Option>
              <Select.Option value="in_progress">Em Andamento</Select.Option>
              <Select.Option value="completed">Concluido</Select.Option>
              <Select.Option value="rejected">Rejeitado</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="response_notes" label="Notas de Resposta">
            <TextArea rows={3} placeholder="Descreva as acoes tomadas..." />
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) => prev.status !== curr.status}
          >
            {({ getFieldValue }) =>
              getFieldValue('status') === 'rejected' && (
                <Form.Item name="rejection_reason" label="Motivo da Rejeicao">
                  <TextArea rows={2} placeholder="Explique o motivo..." />
                </Form.Item>
              )
            }
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">Atualizar</Button>
              <Button onClick={() => { setStatusModalVisible(false); statusForm.resetFields(); }}>Cancelar</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Details Drawer */}
      <Drawer
        title="Detalhes da Solicitacao"
        placement="right"
        width={500}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        {selectedRequest && (
          <div>
            <Row gutter={[16, 16]}>
              <Col span={24}>
                <Card size="small">
                  <Space>
                    <UserOutlined />
                    <div>
                      <Text strong>{selectedRequest.patient_name}</Text>
                      <br />
                      <Text type="secondary">{selectedRequest.patient_cpf}</Text>
                    </div>
                  </Space>
                </Card>
              </Col>
              <Col span={12}>
                <Text type="secondary">Tipo</Text>
                <div><Tag color={getTypeColor(selectedRequest.type)}>{getTypeLabel(selectedRequest.type)}</Tag></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Status</Text>
                <div><Tag color={getStatusColor(selectedRequest.status)}>{getStatusLabel(selectedRequest.status)}</Tag></div>
              </Col>
              <Col span={24}>
                <Text type="secondary">Data da Solicitacao</Text>
                <div><Text strong>{dayjs(selectedRequest.created_at).format('DD/MM/YYYY HH:mm')}</Text></div>
              </Col>
              {selectedRequest.description && (
                <Col span={24}>
                  <Text type="secondary">Descricao</Text>
                  <div><Text>{selectedRequest.description}</Text></div>
                </Col>
              )}
              {selectedRequest.processed_at && (
                <Col span={24}>
                  <Text type="secondary">Processado em</Text>
                  <div><Text strong>{dayjs(selectedRequest.processed_at).format('DD/MM/YYYY HH:mm')}</Text></div>
                </Col>
              )}
              {selectedRequest.response_notes && (
                <Col span={24}>
                  <Text type="secondary">Notas de Resposta</Text>
                  <div><Text>{selectedRequest.response_notes}</Text></div>
                </Col>
              )}
              {selectedRequest.rejection_reason && (
                <Col span={24}>
                  <Text type="secondary">Motivo da Rejeicao</Text>
                  <div><Text type="danger">{selectedRequest.rejection_reason}</Text></div>
                </Col>
              )}
              <Col span={24}>
                <Text type="secondary">IP da Solicitacao</Text>
                <div><Text code>{selectedRequest.request_ip}</Text></div>
              </Col>
            </Row>

            <div style={{ marginTop: 24 }}>
              <Text type="secondary">Historico</Text>
              <Timeline style={{ marginTop: 16 }}>
                <Timeline.Item color="green">
                  Solicitacao criada em {dayjs(selectedRequest.created_at).format('DD/MM/YYYY HH:mm')}
                </Timeline.Item>
                {selectedRequest.status === 'in_progress' && (
                  <Timeline.Item color="blue">Em andamento</Timeline.Item>
                )}
                {selectedRequest.processed_at && (
                  <Timeline.Item color={selectedRequest.status === 'completed' ? 'green' : 'red'}>
                    {selectedRequest.status === 'completed' ? 'Concluido' : 'Rejeitado'} em {dayjs(selectedRequest.processed_at).format('DD/MM/YYYY HH:mm')}
                  </Timeline.Item>
                )}
              </Timeline>
            </div>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default DataRequests;
