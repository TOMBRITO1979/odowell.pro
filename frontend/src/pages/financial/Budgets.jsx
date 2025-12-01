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
  DownloadOutlined,
  UploadOutlined,
  FilePdfOutlined,
  FileExcelOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { budgetsAPI, patientsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const Budgets = () => {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();
  const [loading, setLoading] = useState(false);
  const [budgets, setBudgets] = useState([]);
  const [patients, setPatients] = useState([]);
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
    status: undefined,
  });

  const statusOptions = [
    { value: 'pending', label: 'Pendente', color: statusColors.pending },
    { value: 'approved', label: 'Aprovado', color: statusColors.approved },
    { value: 'rejected', label: 'Rejeitado', color: statusColors.error },
    { value: 'expired', label: 'Expirado', color: statusColors.cancelled },
    { value: 'cancelled', label: 'Cancelado', color: statusColors.cancelled },
  ];

  useEffect(() => {
    fetchBudgets();
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

  const fetchBudgets = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await budgetsAPI.getAll(params);
      let fetchedBudgets = response.data.budgets || [];

      // Filter out approved budgets unless specifically filtered for them
      // Approved budgets should appear in Payments tab instead
      // Cancelled budgets are shown here to track what was cancelled
      if (!filters.status || (filters.status !== 'approved' && filters.status !== 'cancelled')) {
        fetchedBudgets = fetchedBudgets.filter(budget => budget.status !== 'approved');
      }

      setBudgets(fetchedBudgets);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar orçamentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await budgetsAPI.delete(id);
      message.success('Orçamento excluído com sucesso');
      fetchBudgets();
    } catch (error) {
      message.error('Erro ao excluir orçamento');
    }
  };

  const handleExportCSV = async () => {
    try {
      // Only include defined filter values
      const cleanFilters = Object.fromEntries(
        Object.entries(filters).filter(([_, value]) => value !== undefined && value !== null && value !== '')
      );
      const params = new URLSearchParams(cleanFilters).toString();
      const response = await budgetsAPI.exportCSV(params);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `orcamentos_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
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

  const handleExportPDF = async () => {
    try {
      // Only include defined filter values
      const cleanFilters = Object.fromEntries(
        Object.entries(filters).filter(([_, value]) => value !== undefined && value !== null && value !== '')
      );
      const params = new URLSearchParams(cleanFilters).toString();
      const response = await budgetsAPI.exportPDF(params);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `orcamentos_lista_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
      console.error('PDF error:', error);
    }
  };

  const handleImportCSV = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const response = await budgetsAPI.importCSV(formData);
      message.success(response.data.message);

      if (response.data.errors && response.data.errors.length > 0) {
        Modal.warning({
          title: 'Avisos durante a importação',
          content: (
            <div>
              <p>{response.data.imported} orçamentos importados com sucesso.</p>
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
      fetchBudgets();
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
      title: 'Profissional',
      dataIndex: ['dentist', 'name'],
      key: 'dentist',
      ellipsis: true,
    },
    {
      title: 'Valor Total',
      dataIndex: 'total_value',
      key: 'total_value',
      width: 130,
      render: (value) => formatCurrency(value),
      sorter: true,
    },
    {
      title: 'Valor Pago',
      key: 'paid',
      width: 130,
      render: (_, record) => {
        // Sum all paid payments for this budget
        const paidAmount = (record.payments || [])
          .filter(p => p.status === 'paid')
          .reduce((sum, p) => sum + p.amount, 0);
        return <span style={{ color: '#52c41a', fontWeight: 'bold' }}>{formatCurrency(paidAmount)}</span>;
      },
    },
    {
      title: 'Valor Devido',
      key: 'due',
      width: 130,
      render: (_, record) => {
        // Calculate remaining amount
        const paidAmount = (record.payments || [])
          .filter(p => p.status === 'paid')
          .reduce((sum, p) => sum + p.amount, 0);
        const dueAmount = record.total_value - paidAmount;
        const color = dueAmount > 0 ? '#ff4d4f' : '#52c41a';
        return <span style={{ color, fontWeight: 'bold' }}>{formatCurrency(dueAmount)}</span>;
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
      title: 'Validade',
      dataIndex: 'valid_until',
      key: 'valid_until',
      width: 120,
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/budgets/${record.id}/view`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          {canEdit('budgets') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/budgets/${record.id}/edit`)}
              title="Editar"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('budgets') && (
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
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>Orçamentos</span>
          </Space>
        }
        extra={
          canCreate('budgets') && (
            <Button
              icon={<PlusOutlined />}
              onClick={() => navigate('/budgets/new')}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              Novo Orçamento
            </Button>
          )
        }
        style={{ boxShadow: shadows.small }}
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
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
          <Col xs={24} sm={12} md={8}>
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
          <Col xs={24} sm={12} md={8}>
            <Button onClick={fetchBudgets} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col>
            <Space>
              <Button
                icon={<FileExcelOutlined />}
                onClick={handleExportCSV}
                title="Exportar CSV"
                style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}
              >
                Exportar CSV
              </Button>
              <Button
                icon={<FilePdfOutlined />}
                onClick={handleExportPDF}
                title="Gerar PDF da Lista"
                style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
              >
                Gerar PDF
              </Button>
              {canCreate('budgets') && (
                <Button
                  icon={<UploadOutlined />}
                  onClick={() => setUploadModalVisible(true)}
                  title="Importar CSV"
                  style={{ backgroundColor: actionColors.import, borderColor: actionColors.import, color: '#fff' }}
                >
                  Importar CSV
                </Button>
              )}
            </Space>
          </Col>
        </Row>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={budgets}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={handleTableChange}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>

      <Modal
        title="Importar Orçamentos via CSV"
        open={uploadModalVisible}
        onCancel={() => setUploadModalVisible(false)}
        footer={null}
      >
        <div style={{ marginBottom: 16 }}>
          <p><strong>Formato do CSV:</strong></p>
          <p>O arquivo deve conter as seguintes colunas (sem cabeçalho):</p>
          <ol>
            <li>ID do Paciente (número)</li>
            <li>Descrição do Orçamento</li>
            <li>Valor Total (número decimal, ex: 1500.50)</li>
            <li>Status (pending/approved/rejected/expired)</li>
            <li>Observações (opcional)</li>
          </ol>
          <p><strong>Exemplo:</strong></p>
          <code>1,"Tratamento de canal",1500.50,pending,"Inclui coroa"</code>
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

export default Budgets;
