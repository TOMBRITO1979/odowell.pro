import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Space,
  Button,
  Input,
  Select,
  DatePicker,
  Tag,
  Row,
  Col,
  Statistic,
  message,
  Tooltip,
  Typography,
  Drawer,
} from 'antd';
import {
  SearchOutlined,
  DownloadOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  SafetyOutlined,
  UserOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { auditAPI } from '../../services/api';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Text } = Typography;

const AuditLogs = () => {
  const [logs, setLogs] = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [filters, setFilters] = useState({
    user_email: '',
    action: '',
    resource: '',
    success: '',
    dateRange: null,
  });
  const [selectedLog, setSelectedLog] = useState(null);
  const [drawerVisible, setDrawerVisible] = useState(false);

  useEffect(() => {
    fetchLogs();
    fetchStats();
  }, []);

  const fetchLogs = async (page = 1, pageSize = 20) => {
    setLoading(true);
    try {
      const params = {
        page,
        page_size: pageSize,
        user_email: filters.user_email,
        action: filters.action,
        resource: filters.resource,
        success: filters.success,
      };

      if (filters.dateRange && filters.dateRange[0]) {
        params.start_date = filters.dateRange[0].format('YYYY-MM-DD');
      }
      if (filters.dateRange && filters.dateRange[1]) {
        params.end_date = filters.dateRange[1].format('YYYY-MM-DD');
      }

      const response = await auditAPI.getLogs(params);
      setLogs(response.data.logs || []);
      setPagination({
        current: response.data.page,
        pageSize: response.data.page_size,
        total: response.data.total,
      });
    } catch (error) {
      message.error('Erro ao carregar logs de auditoria');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await auditAPI.getStats();
      setStats(response.data);
    } catch (error) {
      console.error('Erro ao carregar estatisticas:', error);
    }
  };

  const handleTableChange = (pag) => {
    fetchLogs(pag.current, pag.pageSize);
  };

  const handleSearch = () => {
    fetchLogs(1, pagination.pageSize);
  };

  const handleExport = async () => {
    try {
      const params = {};
      if (filters.dateRange && filters.dateRange[0]) {
        params.start_date = filters.dateRange[0].format('YYYY-MM-DD');
      }
      if (filters.dateRange && filters.dateRange[1]) {
        params.end_date = filters.dateRange[1].format('YYYY-MM-DD');
      }

      const response = await auditAPI.exportCSV(params);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8;' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `audit_logs_${dayjs().format('YYYY-MM-DD')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();

      message.success('Exportacao concluida');
    } catch (error) {
      message.error('Erro ao exportar logs');
    }
  };

  const showLogDetails = (record) => {
    setSelectedLog(record);
    setDrawerVisible(true);
  };

  const getActionColor = (action) => {
    const colors = {
      create: 'green',
      update: 'blue',
      delete: 'red',
      login: 'purple',
      logout: 'orange',
      view: 'default',
    };
    return colors[action] || 'default';
  };

  const getActionLabel = (action) => {
    const labels = {
      create: 'Criar',
      update: 'Atualizar',
      delete: 'Excluir',
      login: 'Login',
      logout: 'Logout',
      view: 'Visualizar',
    };
    return labels[action] || action;
  };

  const getResourceLabel = (resource) => {
    const labels = {
      patients: 'Pacientes',
      appointments: 'Agendamentos',
      medical_records: 'Prontuarios',
      prescriptions: 'Receitas',
      budgets: 'Orcamentos',
      payments: 'Pagamentos',
      products: 'Produtos',
      suppliers: 'Fornecedores',
      users: 'Usuarios',
      settings: 'Configuracoes',
      auth: 'Autenticacao',
    };
    return labels[resource] || resource;
  };

  const columns = [
    {
      title: 'Data/Hora',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date) => dayjs(date).format('DD/MM/YYYY HH:mm:ss'),
    },
    {
      title: 'Usuario',
      dataIndex: 'user_email',
      key: 'user_email',
      width: 200,
      render: (email, record) => (
        <Space direction="vertical" size={0}>
          <Text>{email || '-'}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.user_role}</Text>
        </Space>
      ),
    },
    {
      title: 'Acao',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action) => (
        <Tag color={getActionColor(action)}>{getActionLabel(action)}</Tag>
      ),
    },
    {
      title: 'Recurso',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: (resource) => getResourceLabel(resource),
    },
    {
      title: 'Caminho',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (path, record) => (
        <Tooltip title={`${record.method} ${path}`}>
          <Text code style={{ fontSize: 12 }}>{record.method} {path}</Text>
        </Tooltip>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 120,
    },
    {
      title: 'Status',
      dataIndex: 'success',
      key: 'success',
      width: 80,
      render: (success) => (
        success ? (
          <Tag icon={<CheckCircleOutlined />} color="success">OK</Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">Erro</Tag>
        )
      ),
    },
    {
      title: 'Acoes',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button
          type="text"
          icon={<EyeOutlined />}
          onClick={() => showLogDetails(record)}
        />
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Space align="center">
          <SafetyOutlined style={{ fontSize: 24, color: '#4CAF50' }} />
          <div>
            <h2 style={{ margin: 0 }}>Logs de Auditoria</h2>
            <Text type="secondary">Rastreamento de acoes para conformidade LGPD</Text>
          </div>
        </Space>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Total de Registros"
                value={stats.total_logs}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Ultimas 24h"
                value={stats.last_24h}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Usuarios Unicos"
                value={stats.unique_users}
                prefix={<UserOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Acoes com Erro"
                value={stats.failed_actions}
                prefix={<CloseCircleOutlined />}
                valueStyle={{ color: stats.failed_actions > 0 ? '#ff4d4f' : '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Input
              placeholder="Email do usuario"
              value={filters.user_email}
              onChange={(e) => setFilters({ ...filters, user_email: e.target.value })}
              prefix={<SearchOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="Acao"
              allowClear
              style={{ width: '100%' }}
              value={filters.action || undefined}
              onChange={(value) => setFilters({ ...filters, action: value || '' })}
            >
              <Select.Option value="create">Criar</Select.Option>
              <Select.Option value="update">Atualizar</Select.Option>
              <Select.Option value="delete">Excluir</Select.Option>
              <Select.Option value="login">Login</Select.Option>
              <Select.Option value="logout">Logout</Select.Option>
              <Select.Option value="view">Visualizar</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="Recurso"
              allowClear
              style={{ width: '100%' }}
              value={filters.resource || undefined}
              onChange={(value) => setFilters({ ...filters, resource: value || '' })}
            >
              <Select.Option value="patients">Pacientes</Select.Option>
              <Select.Option value="appointments">Agendamentos</Select.Option>
              <Select.Option value="medical_records">Prontuarios</Select.Option>
              <Select.Option value="prescriptions">Receitas</Select.Option>
              <Select.Option value="budgets">Orcamentos</Select.Option>
              <Select.Option value="payments">Pagamentos</Select.Option>
              <Select.Option value="auth">Autenticacao</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="Status"
              allowClear
              style={{ width: '100%' }}
              value={filters.success || undefined}
              onChange={(value) => setFilters({ ...filters, success: value || '' })}
            >
              <Select.Option value="true">Sucesso</Select.Option>
              <Select.Option value="false">Erro</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <RangePicker
              style={{ width: '100%' }}
              value={filters.dateRange}
              onChange={(dates) => setFilters({ ...filters, dateRange: dates })}
              format="DD/MM/YYYY"
            />
          </Col>
        </Row>
        <Row style={{ marginTop: 16 }}>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
                Buscar
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => { fetchLogs(); fetchStats(); }}>
                Atualizar
              </Button>
              <Button icon={<DownloadOutlined />} onClick={handleExport}>
                Exportar CSV
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `Total: ${total} registros`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Details Drawer */}
      <Drawer
        title="Detalhes do Log"
        placement="right"
        width={500}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        {selectedLog && (
          <div>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Text type="secondary">ID</Text>
                <div><Text strong>{selectedLog.id}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Data/Hora</Text>
                <div><Text strong>{dayjs(selectedLog.created_at).format('DD/MM/YYYY HH:mm:ss')}</Text></div>
              </Col>
              <Col span={24}>
                <Text type="secondary">Usuario</Text>
                <div><Text strong>{selectedLog.user_email || '-'}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Funcao</Text>
                <div><Text strong>{selectedLog.user_role || '-'}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">User ID</Text>
                <div><Text strong>{selectedLog.user_id || '-'}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Acao</Text>
                <div><Tag color={getActionColor(selectedLog.action)}>{getActionLabel(selectedLog.action)}</Tag></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Recurso</Text>
                <div><Text strong>{getResourceLabel(selectedLog.resource)}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">ID do Recurso</Text>
                <div><Text strong>{selectedLog.resource_id || '-'}</Text></div>
              </Col>
              <Col span={12}>
                <Text type="secondary">Status</Text>
                <div>
                  {selectedLog.success ? (
                    <Tag icon={<CheckCircleOutlined />} color="success">Sucesso</Tag>
                  ) : (
                    <Tag icon={<CloseCircleOutlined />} color="error">Erro</Tag>
                  )}
                </div>
              </Col>
              <Col span={24}>
                <Text type="secondary">Metodo / Caminho</Text>
                <div><Text code>{selectedLog.method} {selectedLog.path}</Text></div>
              </Col>
              <Col span={24}>
                <Text type="secondary">Endereco IP</Text>
                <div><Text strong>{selectedLog.ip_address}</Text></div>
              </Col>
              <Col span={24}>
                <Text type="secondary">User Agent</Text>
                <div><Text style={{ fontSize: 12, wordBreak: 'break-all' }}>{selectedLog.user_agent}</Text></div>
              </Col>
              {selectedLog.details && (
                <Col span={24}>
                  <Text type="secondary">Detalhes Adicionais</Text>
                  <pre style={{
                    background: '#f5f5f5',
                    padding: 12,
                    borderRadius: 4,
                    fontSize: 12,
                    overflow: 'auto'
                  }}>
                    {JSON.stringify(JSON.parse(selectedLog.details), null, 2)}
                  </pre>
                </Col>
              )}
            </Row>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default AuditLogs;
