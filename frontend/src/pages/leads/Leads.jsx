import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Modal,
  message,
  Card,
  Statistic,
  Row,
  Col,
  Popconfirm,
  Tooltip,
  Form,
  Descriptions,
  Divider,
  DatePicker
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  PhoneOutlined,
  UserSwitchOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  WhatsAppOutlined,
  MailOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { leadsAPI, patientsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, brandColors, spacing, shadows } from '../../theme/designSystem';
import dayjs from 'dayjs';

const { Search } = Input;
const { Option } = Select;

const Leads = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { canCreate, canEdit, canDelete } = usePermission();

  const [leads, setLeads] = useState([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({});
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0
  });

  // Filters
  const [filters, setFilters] = useState({
    status: '',
    source: '',
    search: ''
  });

  // Modals
  const [viewModalVisible, setViewModalVisible] = useState(false);
  const [convertModalVisible, setConvertModalVisible] = useState(false);
  const [selectedLead, setSelectedLead] = useState(null);
  const [convertForm] = Form.useForm();
  const [converting, setConverting] = useState(false);

  useEffect(() => {
    fetchLeads();
    fetchStats();
  }, [pagination.current, pagination.pageSize, filters, location.key]);

  const fetchLeads = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters
      };

      const response = await leadsAPI.getAll(params);
      setLeads(response.data.data || []);
      setPagination(prev => ({
        ...prev,
        total: response.data.total
      }));
    } catch (error) {
      message.error('Erro ao carregar leads');
      console.error('Error fetching leads:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await leadsAPI.getStats();
      setStats(response.data);
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const handleDelete = async (id) => {
    try {
      await leadsAPI.delete(id);
      message.success('Lead removido com sucesso');
      fetchLeads();
      fetchStats();
    } catch (error) {
      message.error('Erro ao remover lead');
      console.error('Error deleting:', error);
    }
  };

  const handleConvert = async (values) => {
    if (!selectedLead) return;

    setConverting(true);
    try {
      const response = await leadsAPI.convert(selectedLead.id, values);
      message.success('Lead convertido para paciente com sucesso!');
      setConvertModalVisible(false);
      convertForm.resetFields();
      fetchLeads();
      fetchStats();

      // Offer to view the new patient
      Modal.confirm({
        title: 'Conversão realizada!',
        content: `O lead "${selectedLead.name}" foi convertido para paciente. Deseja visualizar o cadastro do paciente?`,
        okText: 'Ver Paciente',
        cancelText: 'Ficar aqui',
        onOk: () => navigate(`/patients/${response.data.patient_id}`)
      });
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao converter lead');
      console.error('Error converting lead:', error);
    } finally {
      setConverting(false);
    }
  };

  const handleTableChange = (pag) => {
    setPagination({
      ...pagination,
      current: pag.current,
      pageSize: pag.pageSize,
    });
  };

  const getStatusByCode = (byStatus, statusCode) => {
    if (!byStatus || !Array.isArray(byStatus)) return 0;
    const found = byStatus.find(s => s.status === statusCode);
    return found ? found.count : 0;
  };

  const columns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
      render: (name, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          <div style={{ fontSize: 12, color: '#999' }}>
            <PhoneOutlined /> {record.phone}
          </div>
        </div>
      )
    },
    {
      title: 'Contato',
      dataIndex: 'email',
      key: 'email',
      render: (email, record) => (
        <div>
          {email && (
            <div style={{ fontSize: 12 }}>
              <MailOutlined /> {email}
            </div>
          )}
          {record.source && (
            <Tag color={record.source === 'whatsapp' ? 'green' : 'blue'}>
              {record.source === 'whatsapp' ? <WhatsAppOutlined /> : null} {record.source}
            </Tag>
          )}
        </div>
      )
    },
    {
      title: 'Motivo do Contato',
      dataIndex: 'contact_reason',
      key: 'contact_reason',
      ellipsis: true,
      width: 200,
      render: (reason) => reason || '-'
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      align: 'center',
      render: (status) => {
        const statusMap = {
          new: { color: statusColors.pending, label: 'Novo' },
          contacted: { color: statusColors.inProgress, label: 'Contatado' },
          qualified: { color: 'purple', label: 'Qualificado' },
          converted: { color: statusColors.success, label: 'Convertido' },
          lost: { color: statusColors.cancelled, label: 'Perdido' }
        };
        const config = statusMap[status] || { color: statusColors.pending, label: status };
        return <Tag color={config.color}>{config.label}</Tag>;
      }
    },
    {
      title: 'Data',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => new Date(date).toLocaleDateString('pt-BR')
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 200,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Tooltip title="Visualizar">
            <Button
              type="link"
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedLead(record);
                setViewModalVisible(true);
              }}
              style={{ color: actionColors.view }}
            />
          </Tooltip>
          {record.status !== 'converted' && canEdit('leads') && (
            <Tooltip title="Editar">
              <Button
                type="link"
                icon={<EditOutlined />}
                onClick={() => navigate(`/leads/${record.id}/edit`)}
                style={{ color: actionColors.edit }}
              />
            </Tooltip>
          )}
          {record.status !== 'converted' && canEdit('leads') && (
            <Tooltip title="Converter em Paciente">
              <Button
                type="link"
                icon={<UserSwitchOutlined />}
                onClick={() => {
                  setSelectedLead(record);
                  convertForm.resetFields();
                  setConvertModalVisible(true);
                }}
                style={{ color: actionColors.approve }}
              />
            </Tooltip>
          )}
          {canDelete('leads') && (
            <Popconfirm
              title="Excluir este lead?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="link"
                icon={<DeleteOutlined />}
                style={{ color: actionColors.delete }}
              />
            </Popconfirm>
          )}
        </Space>
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <h1>Leads</h1>
      <p style={{ color: '#666', marginBottom: 24 }}>
        Gerencie leads e potenciais pacientes vindos do WhatsApp e outras fontes
      </p>

      {/* Statistics */}
      <Row gutter={[spacing.md, spacing.md]} style={{ marginBottom: spacing.lg }}>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Total de Leads"
              value={stats.total || 0}
              valueStyle={{ color: brandColors.primary }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Novos (este mês)"
              value={stats.this_month || 0}
              valueStyle={{ color: statusColors.pending }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Convertidos"
              value={stats.converted || 0}
              valueStyle={{ color: statusColors.success }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable style={{ boxShadow: shadows.small }}>
            <Statistic
              title="Taxa de Conversão"
              value={stats.total > 0 ? ((stats.converted / stats.total) * 100).toFixed(1) : 0}
              suffix="%"
              valueStyle={{ color: statusColors.inProgress }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: spacing.md, boxShadow: shadows.small }}>
        <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space wrap>
            <Input
              placeholder="Buscar nome, telefone, email"
              prefix={<SearchOutlined />}
              value={filters.search}
              onChange={(e) => setFilters({ ...filters, search: e.target.value })}
              onPressEnter={fetchLeads}
              style={{ width: 250 }}
            />
            <Select
              placeholder="Status"
              value={filters.status || undefined}
              onChange={(value) => setFilters({ ...filters, status: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Option value="new">Novo</Option>
              <Option value="contacted">Contatado</Option>
              <Option value="qualified">Qualificado</Option>
              <Option value="converted">Convertido</Option>
              <Option value="lost">Perdido</Option>
            </Select>
            <Select
              placeholder="Fonte"
              value={filters.source || undefined}
              onChange={(value) => setFilters({ ...filters, source: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Option value="whatsapp">WhatsApp</Option>
              <Option value="website">Website</Option>
              <Option value="referral">Indicação</Option>
              <Option value="instagram">Instagram</Option>
              <Option value="facebook">Facebook</Option>
              <Option value="other">Outro</Option>
            </Select>
            <Button onClick={fetchLeads}>Filtrar</Button>
          </Space>

          {canCreate('leads') && (
            <Button
              icon={<PlusOutlined />}
              onClick={() => navigate('/leads/new')}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              Novo Lead
            </Button>
          )}
        </Space>
      </Card>

      {/* Table */}
      <Card style={{ boxShadow: shadows.small }}>
        <Table
          columns={columns}
          dataSource={leads}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* View Modal */}
      <Modal
        title="Detalhes do Lead"
        open={viewModalVisible}
        onCancel={() => setViewModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setViewModalVisible(false)}>
            Fechar
          </Button>,
          selectedLead?.status !== 'converted' && canEdit('leads') && (
            <Button
              key="convert"
              type="primary"
              icon={<UserSwitchOutlined />}
              onClick={() => {
                setViewModalVisible(false);
                convertForm.resetFields();
                setConvertModalVisible(true);
              }}
            >
              Converter em Paciente
            </Button>
          )
        ]}
        width={600}
      >
        {selectedLead && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="Nome">{selectedLead.name}</Descriptions.Item>
            <Descriptions.Item label="Telefone">{selectedLead.phone}</Descriptions.Item>
            <Descriptions.Item label="Email">{selectedLead.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="Fonte">
              <Tag color={selectedLead.source === 'whatsapp' ? 'green' : 'blue'}>
                {selectedLead.source}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              {(() => {
                const statusMap = {
                  new: { color: statusColors.pending, label: 'Novo' },
                  contacted: { color: statusColors.inProgress, label: 'Contatado' },
                  qualified: { color: 'purple', label: 'Qualificado' },
                  converted: { color: statusColors.success, label: 'Convertido' },
                  lost: { color: statusColors.cancelled, label: 'Perdido' }
                };
                const config = statusMap[selectedLead.status] || { color: 'default', label: selectedLead.status };
                return <Tag color={config.color}>{config.label}</Tag>;
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="Motivo do Contato">
              {selectedLead.contact_reason || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Observações">
              {selectedLead.notes || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Data de Cadastro">
              {new Date(selectedLead.created_at).toLocaleString('pt-BR')}
            </Descriptions.Item>
            {selectedLead.converted_at && (
              <Descriptions.Item label="Data de Conversão">
                {new Date(selectedLead.converted_at).toLocaleString('pt-BR')}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* Convert Modal */}
      <Modal
        title="Converter Lead em Paciente"
        open={convertModalVisible}
        onCancel={() => setConvertModalVisible(false)}
        footer={null}
        width={600}
      >
        {selectedLead && (
          <>
            <p style={{ marginBottom: 16 }}>
              Preencha as informações adicionais para criar o cadastro do paciente:
            </p>
            <Descriptions column={1} size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label="Nome">{selectedLead.name}</Descriptions.Item>
              <Descriptions.Item label="Telefone">{selectedLead.phone}</Descriptions.Item>
              <Descriptions.Item label="Email">{selectedLead.email || '-'}</Descriptions.Item>
            </Descriptions>

            <Divider />

            <Form
              form={convertForm}
              layout="vertical"
              onFinish={handleConvert}
            >
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    name="cpf"
                    label="CPF"
                  >
                    <Input placeholder="000.000.000-00" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="birth_date"
                    label="Data de Nascimento"
                  >
                    <DatePicker
                      format="DD/MM/YYYY"
                      style={{ width: '100%' }}
                      placeholder="Selecione"
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item
                name="address"
                label="Endereço"
              >
                <Input placeholder="Rua, número, complemento" />
              </Form.Item>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item name="city" label="Cidade">
                    <Input />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="state" label="Estado">
                    <Input placeholder="SP" maxLength={2} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="zip_code" label="CEP">
                    <Input placeholder="00000-000" />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item
                name="notes"
                label="Observações Adicionais"
              >
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
                <Space>
                  <Button onClick={() => setConvertModalVisible(false)}>
                    Cancelar
                  </Button>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={converting}
                    icon={<UserSwitchOutlined />}
                  >
                    Converter em Paciente
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </>
        )}
      </Modal>
    </div>
  );
};

export default Leads;
