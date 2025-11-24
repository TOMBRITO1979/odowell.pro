import React, { useState } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  DatePicker,
  Space,
  Button,
  message,
  Modal,
  Table,
  Descriptions,
} from 'antd';
import {
  BarChartOutlined,
  DollarOutlined,
  UserOutlined,
  CalendarOutlined,
  FileTextOutlined,
  FilePdfOutlined,
  FileExcelOutlined,
  FundOutlined,
  AlertOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { reportsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { RangePicker } = DatePicker;

const Reports = () => {
  const [loading, setLoading] = useState(false);
  const [dashboard, setDashboard] = useState({
    total_patients: 0,
    total_appointments: 0,
    total_revenue: 0,
    pending_payments: 0,
  });
  const [dateRange, setDateRange] = useState(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [modalContent, setModalContent] = useState(null);
  const [currentReportType, setCurrentReportType] = useState(null);

  const fetchDashboard = async () => {
    setLoading(true);
    try {
      const response = await reportsAPI.getDashboard();
      setDashboard(response.data);
    } catch (error) {
      message.error('Erro ao carregar dados do dashboard');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchDashboard();
  }, []);

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value || 0);
  };

  const handleDownloadPDF = async () => {
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      let response;
      let filename;

      if (currentReportType === 'revenue') {
        response = await reportsAPI.downloadRevenuePDF(params);
        filename = 'relatorio_receitas.pdf';
      } else if (currentReportType === 'attendance') {
        response = await reportsAPI.downloadAttendancePDF(params);
        filename = 'relatorio_atendimentos.pdf';
      } else if (currentReportType === 'procedures') {
        response = await reportsAPI.downloadProceduresPDF();
        filename = 'relatorio_procedimentos.pdf';
      }

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
      console.error('Error:', error);
    }
  };

  const handleDownloadExcel = async () => {
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      let response;
      let filename;

      if (currentReportType === 'revenue') {
        response = await reportsAPI.downloadRevenueExcel(params);
        filename = 'relatorio_receitas.xlsx';
      } else if (currentReportType === 'attendance') {
        response = await reportsAPI.downloadAttendanceExcel(params);
        filename = 'relatorio_atendimentos.xlsx';
      } else if (currentReportType === 'procedures') {
        response = await reportsAPI.downloadProceduresExcel();
        filename = 'relatorio_procedimentos.xlsx';
      }

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('Planilha baixada com sucesso');
    } catch (error) {
      message.error('Erro ao baixar planilha');
      console.error('Error:', error);
    }
  };

  const handleRevenueReport = async () => {
    setLoading(true);
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      const response = await reportsAPI.getRevenue(params);
      const data = response.data;
      setCurrentReportType('revenue');

      const methodColumns = [
        { title: 'Método de Pagamento', dataIndex: 'payment_method', key: 'payment_method' },
        { title: 'Quantidade', dataIndex: 'count', key: 'count' },
        { title: 'Total', dataIndex: 'total', key: 'total', render: (val) => formatCurrency(val) },
      ];

      const monthColumns = [
        { title: 'Mês', dataIndex: 'month', key: 'month' },
        { title: 'Quantidade', dataIndex: 'count', key: 'count' },
        { title: 'Total', dataIndex: 'total', key: 'total', render: (val) => formatCurrency(val) },
      ];

      setModalTitle('Relatório de Receitas');
      setModalContent(
        <div>
          <Descriptions bordered column={1} style={{ marginBottom: 24 }}>
            <Descriptions.Item label="Receita Total">
              {formatCurrency(data.total_revenue)}
            </Descriptions.Item>
          </Descriptions>

          <h3>Por Método de Pagamento</h3>
          <Table
            dataSource={data.by_method || []}
            columns={methodColumns}
            rowKey="payment_method"
            pagination={false}
            style={{ marginBottom: 24 }}
          />

          <h3>Por Mês</h3>
          <Table
            dataSource={data.by_month || []}
            columns={monthColumns}
            rowKey="month"
            pagination={false}
          />
        </div>
      );
      setModalVisible(true);
    } catch (error) {
      message.error('Erro ao carregar relatório de receitas');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAttendanceReport = async () => {
    setLoading(true);
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      const response = await reportsAPI.getAttendance(params);
      const data = response.data;
      setCurrentReportType('attendance');

      setModalTitle('Relatório de Atendimentos');
      setModalContent(
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Total de Agendamentos" span={2}>
            {data.total}
          </Descriptions.Item>
          <Descriptions.Item label="Concluídos">
            {data.completed}
          </Descriptions.Item>
          <Descriptions.Item label="Cancelados">
            {data.cancelled}
          </Descriptions.Item>
          <Descriptions.Item label="Faltaram">
            {data.no_show}
          </Descriptions.Item>
          <Descriptions.Item label="Taxa de Comparecimento">
            {data.attendance_rate?.toFixed(2)}%
          </Descriptions.Item>
        </Descriptions>
      );
      setModalVisible(true);
    } catch (error) {
      message.error('Erro ao carregar relatório de atendimentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleProceduresReport = async () => {
    setLoading(true);
    try {
      const response = await reportsAPI.getProcedures();
      const data = response.data;
      setCurrentReportType('procedures');

      const columns = [
        { title: 'Procedimento', dataIndex: 'procedure', key: 'procedure' },
        { title: 'Quantidade', dataIndex: 'count', key: 'count' },
      ];

      setModalTitle('Relatório de Procedimentos');
      setModalContent(
        <Table
          dataSource={data.procedures || []}
          columns={columns}
          rowKey="procedure"
          pagination={false}
        />
      );
      setModalVisible(true);
    } catch (error) {
      message.error('Erro ao carregar relatório de procedimentos');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleBudgetConversionReport = async () => {
    setLoading(true);
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      const response = await reportsAPI.getBudgetConversion(params);
      const data = response.data;
      setCurrentReportType('budget-conversion');

      setModalTitle('Taxa de Conversão de Orçamentos');
      setModalContent(
        <div>
          <Descriptions bordered column={2}>
            <Descriptions.Item label="Total de Orçamentos" span={2}>
              {data.total_budgets || 0}
            </Descriptions.Item>
            <Descriptions.Item label="Orçamentos Aprovados">
              {data.approved_budgets || 0}
            </Descriptions.Item>
            <Descriptions.Item label="Taxa de Conversão">
              <span style={{ color: '#52c41a', fontWeight: 'bold', fontSize: '16px' }}>
                {data.conversion_rate?.toFixed(2) || 0}%
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Valor Total Aprovado" span={2}>
              {formatCurrency(data.total_approved || 0)}
            </Descriptions.Item>
          </Descriptions>
          {data.by_status && data.by_status.length > 0 && (
            <>
              <h4 style={{ marginTop: 16 }}>Distribuição por Status</h4>
              <Table
                dataSource={data.by_status}
                columns={[
                  { title: 'Status', dataIndex: 'status', key: 'status' },
                  { title: 'Quantidade', dataIndex: 'count', key: 'count' },
                  { title: 'Percentual', dataIndex: 'percentage', key: 'percentage', render: (v) => `${v}%` },
                  { title: 'Valor Total', dataIndex: 'total_amount', key: 'total_amount', render: formatCurrency },
                ]}
                rowKey="status"
                pagination={false}
              />
            </>
          )}
        </div>
      );
      setModalVisible(true);
    } catch (error) {
      message.error('Erro ao carregar relatório de conversão');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOverduePaymentsReport = async () => {
    setLoading(true);
    try {
      const response = await reportsAPI.getOverduePayments();
      const data = response.data;
      setCurrentReportType('overdue-payments');

      setModalTitle('Controle de Inadimplência');
      setModalContent(
        <div>
          <Descriptions bordered column={2}>
            <Descriptions.Item label="Total em Atraso" span={2}>
              <span style={{ color: '#ef4444', fontWeight: 'bold', fontSize: '16px' }}>
                {formatCurrency(data.total_overdue || 0)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Quantidade de Pagamentos Atrasados" span={2}>
              {data.overdue_count || 0}
            </Descriptions.Item>
          </Descriptions>

          {data.overdue_patients && data.overdue_patients.length > 0 && (
            <>
              <h4 style={{ marginTop: 16 }}>Pacientes Inadimplentes</h4>
              <Table
                dataSource={data.overdue_patients}
                columns={[
                  { title: 'Paciente', dataIndex: 'patient_name', key: 'patient_name' },
                  { title: 'Qtd. Atrasados', dataIndex: 'overdue_count', key: 'overdue_count' },
                  { title: 'Total em Atraso', dataIndex: 'total_overdue', key: 'total_overdue', render: formatCurrency },
                  { title: 'Atraso Mais Antigo', dataIndex: 'oldest_due_date', key: 'oldest_due_date' },
                ]}
                rowKey="patient_id"
                pagination={{ pageSize: 10 }}
              />
            </>
          )}

          {data.overdue_by_age && data.overdue_by_age.length > 0 && (
            <>
              <h4 style={{ marginTop: 16 }}>Por Tempo de Atraso</h4>
              <Table
                dataSource={data.overdue_by_age}
                columns={[
                  { title: 'Período', dataIndex: 'age_range', key: 'age_range' },
                  { title: 'Quantidade', dataIndex: 'count', key: 'count' },
                  { title: 'Valor Total', dataIndex: 'total', key: 'total', render: formatCurrency },
                ]}
                rowKey="age_range"
                pagination={false}
              />
            </>
          )}

          {(!data.overdue_patients || data.overdue_patients.length === 0) && (
            <div style={{ textAlign: 'center', padding: '24px', color: '#52c41a' }}>
              <h3>✓ Nenhum pagamento em atraso!</h3>
              <p>Todos os pagamentos estão em dia.</p>
            </div>
          )}
        </div>
      );
      setModalVisible(true);
    } catch (error) {
      message.error('Erro ao carregar relatório de inadimplência');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <BarChartOutlined />
            <span>Relatórios e Indicadores</span>
          </Space>
        }
        extra={
          <RangePicker
            format="DD/MM/YYYY"
            placeholder={['Data Inicial', 'Data Final']}
            onChange={(dates) => setDateRange(dates)}
            value={dateRange}
          />
        }
        style={{ boxShadow: shadows.small }}
      >
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ boxShadow: shadows.small }}>
              <Statistic
                title="Total de Pacientes"
                value={dashboard.total_patients}
                prefix={<UserOutlined />}
                valueStyle={{ color: statusColors.success }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ boxShadow: shadows.small }}>
              <Statistic
                title="Agendamentos"
                value={dashboard.total_appointments}
                prefix={<CalendarOutlined />}
                valueStyle={{ color: statusColors.inProgress }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ boxShadow: shadows.small }}>
              <Statistic
                title="Receita Total"
                value={dashboard.total_revenue}
                prefix={<DollarOutlined />}
                valueStyle={{ color: statusColors.success }}
                formatter={(value) => formatCurrency(value)}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ boxShadow: shadows.small }}>
              <Statistic
                title="A Receber"
                value={dashboard.pending_payments}
                prefix={<FileTextOutlined />}
                valueStyle={{ color: statusColors.pending }}
                formatter={(value) => formatCurrency(value)}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={24}>
            <Card title="Relatórios Disponíveis" loading={loading} style={{ boxShadow: shadows.small }}>
              <Space wrap style={{ width: '100%', justifyContent: 'center' }}>
                <Button
                  icon={<BarChartOutlined />}
                  onClick={handleRevenueReport}
                  loading={loading}
                  style={{
                    backgroundColor: actionColors.create,
                    borderColor: actionColors.create,
                    color: '#fff'
                  }}
                >
                  Relatório de Receitas
                </Button>
                <Button
                  icon={<CalendarOutlined />}
                  onClick={handleAttendanceReport}
                  loading={loading}
                  style={{
                    backgroundColor: actionColors.edit,
                    borderColor: actionColors.edit,
                    color: '#fff'
                  }}
                >
                  Relatório de Atendimentos
                </Button>
                <Button
                  icon={<FileTextOutlined />}
                  onClick={handleProceduresReport}
                  loading={loading}
                  style={{
                    backgroundColor: actionColors.view,
                    borderColor: actionColors.view,
                    color: '#fff'
                  }}
                >
                  Relatório de Procedimentos
                </Button>
                <Button
                  icon={<FundOutlined />}
                  onClick={handleBudgetConversionReport}
                  loading={loading}
                  style={{
                    backgroundColor: actionColors.approve,
                    borderColor: actionColors.approve,
                    color: '#fff'
                  }}
                >
                  Taxa de Conversão de Orçamentos
                </Button>
                <Button
                  icon={<AlertOutlined />}
                  onClick={handleOverduePaymentsReport}
                  loading={loading}
                  style={{
                    backgroundColor: actionColors.delete,
                    borderColor: actionColors.delete,
                    color: '#fff'
                  }}
                >
                  Controle de Inadimplência
                </Button>
              </Space>
            </Card>
          </Col>
        </Row>
      </Card>

      <Modal
        title={modalTitle}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button
            key="excel"
            icon={<FileExcelOutlined />}
            onClick={handleDownloadExcel}
            style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}
          >
            Baixar Excel
          </Button>,
          <Button
            key="pdf"
            icon={<FilePdfOutlined />}
            onClick={handleDownloadPDF}
            style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
          >
            Baixar PDF
          </Button>,
          <Button key="close" onClick={() => setModalVisible(false)}>
            Fechar
          </Button>,
        ]}
        width={800}
      >
        {modalContent}
      </Modal>
    </div>
  );
};

export default Reports;
