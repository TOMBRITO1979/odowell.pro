import React, { useState, useEffect, useRef } from 'react';
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
  Statistic,
  Upload,
  Modal,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  DollarOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  FilePdfOutlined,
  FileExcelOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { paymentsAPI, patientsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';

const { RangePicker } = DatePicker;

const Payments = () => {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [payments, setPayments] = useState([]);
  const [patients, setPatients] = useState([]);
  const [statistics, setStatistics] = useState({
    income: 0,
    expenses: 0,
    balance: 0,
    pending: 0,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef(null);

  // Filters
  const [filters, setFilters] = useState({
    patient_id: undefined,
    type: undefined,
    status: undefined,
    date_range: null,
  });

  const paymentTypes = [
    { value: 'income', label: 'Receita', color: 'success' },
    { value: 'expense', label: 'Despesa', color: 'error' },
  ];

  const statusOptions = [
    { value: 'pending', label: 'Pendente', color: 'warning' },
    { value: 'paid', label: 'Pago', color: 'success' },
    { value: 'overdue', label: 'Atrasado', color: 'error' },
    { value: 'cancelled', label: 'Cancelado', color: 'default' },
  ];

  const paymentMethods = [
    { value: 'cash', label: 'Dinheiro' },
    { value: 'credit_card', label: 'Cartão de Crédito' },
    { value: 'debit_card', label: 'Cartão de Débito' },
    { value: 'pix', label: 'PIX' },
    { value: 'transfer', label: 'Transferência' },
    { value: 'insurance', label: 'Convênio' },
  ];

  useEffect(() => {
    fetchPayments();
    fetchPatients();
    fetchStatistics();
  }, [pagination.current, filters]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchStatistics = async () => {
    try {
      const response = await paymentsAPI.getCashFlow();
      setStatistics(response.data);
    } catch (error) {
      console.error('Error fetching statistics:', error);
    }
  };

  const fetchPayments = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await paymentsAPI.getAll(params);
      setPayments(response.data.payments || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar pagamentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await paymentsAPI.delete(id);
      message.success('Pagamento excluído com sucesso');
      fetchPayments();
      fetchStatistics();
    } catch (error) {
      message.error('Erro ao excluir pagamento');
    }
  };

  const handleDownloadPDF = async () => {
    try {
      const params = { ...filters };

      // Convert date_range to start_date and end_date if exists
      if (filters.date_range && filters.date_range.length === 2) {
        params.start_date = filters.date_range[0].format('YYYY-MM-DD');
        params.end_date = filters.date_range[1].format('YYYY-MM-DD');
        delete params.date_range;
      }

      const response = await paymentsAPI.downloadPDF(params);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'relatorio_pagamentos.pdf');
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
      console.error('Error:', error);
    }
  };

  const handleExportCSV = async () => {
    try {
      const params = { ...filters };

      // Convert date_range to start_date and end_date if exists
      if (filters.date_range && filters.date_range.length === 2) {
        params.start_date = filters.date_range[0].format('YYYY-MM-DD');
        params.end_date = filters.date_range[1].format('YYYY-MM-DD');
        delete params.date_range;
      }

      // Only include defined filter values
      const cleanFilters = Object.fromEntries(
        Object.entries(params).filter(([_, value]) => value !== undefined && value !== null && value !== '')
      );
      const queryString = new URLSearchParams(cleanFilters).toString();
      const response = await paymentsAPI.exportCSV(queryString);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `pagamentos_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
      console.error('Export error:', error);
    }
  };

  const handleImportCSV = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const response = await paymentsAPI.importCSV(formData);
      message.success(response.data.message);

      if (response.data.errors && response.data.errors.length > 0) {
        Modal.warning({
          title: 'Avisos durante a importação',
          content: (
            <div>
              <p>{response.data.imported} pagamentos importados com sucesso.</p>
              <p>Erros encontrados:</p>
              <ul>
                {response.data.errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            </div>
          ),
          width: 600,
        });
      }

      setUploadModalVisible(false);
      fetchPayments();
    } catch (error) {
      message.error('Erro ao importar CSV');
      console.error('Import error:', error);
    } finally {
      setUploading(false);
    }

    return false; // Prevent default upload behavior
  };

  const handleTableChange = (newPagination) => {
    setPagination(newPagination);
  };

  const getTypeTag = (type) => {
    const typeObj = paymentTypes.find((t) => t.value === type);
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

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
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
      title: 'Descrição',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'Valor',
      dataIndex: 'amount',
      key: 'amount',
      width: 130,
      render: (value, record) => {
        const color = record.type === 'income' ? '#52c41a' : '#ff4d4f';
        return <span style={{ color, fontWeight: 'bold' }}>{formatCurrency(value)}</span>;
      },
      sorter: true,
    },
    {
      title: 'Método',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 140,
      render: (method) => {
        const methodObj = paymentMethods.find((m) => m.value === method);
        return methodObj ? methodObj.label : method;
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Vencimento',
      dataIndex: 'due_date',
      key: 'due_date',
      width: 120,
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          {canEdit('payments') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/payments/${record.id}/edit`)}
              title="Editar"
            />
          )}
          {canDelete('payments') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                title="Excluir"
              />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Receitas"
              value={statistics.income}
              precision={2}
              valueStyle={{ color: '#3f8600' }}
              prefix={<ArrowUpOutlined />}
              suffix="R$"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Despesas"
              value={statistics.expenses}
              precision={2}
              valueStyle={{ color: '#cf1322' }}
              prefix={<ArrowDownOutlined />}
              suffix="R$"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Saldo"
              value={statistics.balance}
              precision={2}
              valueStyle={{ color: statistics.balance >= 0 ? '#3f8600' : '#cf1322' }}
              prefix="R$"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="A Receber"
              value={statistics.pending}
              precision={2}
              valueStyle={{ color: '#1890ff' }}
              prefix="R$"
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>Pagamentos</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
              style={{ backgroundColor: '#ef4444', borderColor: '#ef4444', color: '#fff' }}
            >
              Baixar PDF
            </Button>
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              title="Exportar CSV"
              style={{ backgroundColor: '#22c55e', borderColor: '#22c55e', color: '#fff' }}
            >
              Exportar CSV
            </Button>
            {canCreate('payments') && (
              <Button
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
                title="Importar CSV"
                style={{ backgroundColor: '#3b82f6', borderColor: '#3b82f6', color: '#fff' }}
              >
                Importar CSV
              </Button>
            )}
            {canCreate('payments') && (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate('/payments/new')}
              >
                Novo Pagamento
              </Button>
            )}
          </Space>
        }
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
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="Tipo"
              style={{ width: '100%' }}
              allowClear
              value={filters.type}
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              {paymentTypes.map((type) => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
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
            <Button onClick={fetchPayments} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={payments}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title="Importar Pagamentos via CSV"
        open={uploadModalVisible}
        onCancel={() => setUploadModalVisible(false)}
        footer={null}
      >
        <div style={{ marginBottom: 16 }}>
          <p><strong>Formato do CSV:</strong></p>
          <p>O arquivo deve conter as seguintes colunas (sem cabeçalho):</p>
          <ol>
            <li>ID do Paciente (número)</li>
            <li>Tipo (income/expense)</li>
            <li>Categoria (texto)</li>
            <li>Descrição</li>
            <li>Método de Pagamento (cash/credit_card/debit_card/pix/transfer/insurance)</li>
            <li>Valor (número decimal, ex: 150.50)</li>
            <li>ID do Orçamento (opcional, número)</li>
          </ol>
          <p><strong>Exemplo:</strong></p>
          <code>1,income,treatment,"Consulta",cash,150.50,5</code>
        </div>

        <Upload
          accept=".csv"
          beforeUpload={handleImportCSV}
          showUploadList={false}
        >
          <Button
            icon={<UploadOutlined />}
            loading={uploading}
            block
            type="primary"
          >
            {uploading ? 'Importando...' : 'Selecionar arquivo CSV'}
          </Button>
        </Upload>
      </Modal>
    </div>
  );
};

export default Payments;
