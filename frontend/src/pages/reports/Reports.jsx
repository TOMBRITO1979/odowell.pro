import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  DatePicker,
  Space,
  Button,
  message,
  Spin,
  Empty,
  Progress,
  Tag,
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
  DownloadOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  BarChart,
  Bar,
  PieChart,
  Pie,
  AreaChart,
  Area,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { reportsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';

const { RangePicker } = DatePicker;

// Cores suaves e foscas (paleta moderna)
const COLORS = [
  '#A5D6A7', // verde suave
  '#90CAF9', // azul suave
  '#CE93D8', // lilás suave
  '#FFF59D', // amarelo suave
  '#F8BBD9', // rosa suave
  '#80DEEA', // teal suave
  '#FFCC80', // pêssego suave
  '#B39DDB', // indigo suave
];

// Tradução de métodos de pagamento
const translatePaymentMethod = (method) => {
  const translations = {
    'cash': 'Dinheiro',
    'credit_card': 'Cartão de Crédito',
    'debit_card': 'Cartão de Débito',
    'pix': 'PIX',
    'bank_transfer': 'Transferência',
    'check': 'Cheque',
    'insurance': 'Convênio',
    'other': 'Outro',
  };
  return translations[method] || method;
};

// Tradução de status de orçamentos
const translateBudgetStatus = (status) => {
  const translations = {
    'approved': 'Aprovado',
    'cancelled': 'Cancelado',
    'pending': 'Pendente',
    'rejected': 'Rejeitado',
  };
  return translations[status] || status;
};

const Reports = () => {
  const [loading, setLoading] = useState(false);
  const [dashboard, setDashboard] = useState({
    total_patients: 0,
    total_appointments: 0,
    total_revenue: 0,
    pending_payments: 0,
  });

  // Individual date ranges for each report
  const [revenueDateRange, setRevenueDateRange] = useState(null);
  const [attendanceDateRange, setAttendanceDateRange] = useState(null);
  const [budgetConversionDateRange, setBudgetConversionDateRange] = useState(null);

  // Data for each report
  const [revenueData, setRevenueData] = useState(null);
  const [attendanceData, setAttendanceData] = useState(null);
  const [proceduresData, setProceduresData] = useState(null);
  const [budgetConversionData, setBudgetConversionData] = useState(null);
  const [overduePaymentsData, setOverduePaymentsData] = useState(null);

  // Loading states
  const [loadingRevenue, setLoadingRevenue] = useState(false);
  const [loadingAttendance, setLoadingAttendance] = useState(false);
  const [loadingProcedures, setLoadingProcedures] = useState(false);
  const [loadingBudgetConversion, setLoadingBudgetConversion] = useState(false);
  const [loadingOverduePayments, setLoadingOverduePayments] = useState(false);

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

  const fetchRevenueReport = async () => {
    setLoadingRevenue(true);
    try {
      const params = {};
      if (revenueDateRange && revenueDateRange[0] && revenueDateRange[1]) {
        params.start_date = revenueDateRange[0].format('YYYY-MM-DD');
        params.end_date = revenueDateRange[1].format('YYYY-MM-DD');
      }
      const response = await reportsAPI.getRevenue(params);
      setRevenueData(response.data);
    } catch (error) {
      message.error('Erro ao carregar relatório de receitas');
      console.error('Error:', error);
    } finally {
      setLoadingRevenue(false);
    }
  };

  const fetchAttendanceReport = async () => {
    setLoadingAttendance(true);
    try {
      const params = {};
      if (attendanceDateRange && attendanceDateRange[0] && attendanceDateRange[1]) {
        params.start_date = attendanceDateRange[0].format('YYYY-MM-DD');
        params.end_date = attendanceDateRange[1].format('YYYY-MM-DD');
      }
      const response = await reportsAPI.getAttendance(params);
      setAttendanceData(response.data);
    } catch (error) {
      message.error('Erro ao carregar relatório de atendimentos');
      console.error('Error:', error);
    } finally {
      setLoadingAttendance(false);
    }
  };

  const fetchProceduresReport = async () => {
    setLoadingProcedures(true);
    try {
      const response = await reportsAPI.getProcedures();
      setProceduresData(response.data);
    } catch (error) {
      message.error('Erro ao carregar relatório de procedimentos');
      console.error('Error:', error);
    } finally {
      setLoadingProcedures(false);
    }
  };

  const fetchBudgetConversionReport = async () => {
    setLoadingBudgetConversion(true);
    try {
      const params = {};
      if (budgetConversionDateRange && budgetConversionDateRange[0] && budgetConversionDateRange[1]) {
        params.start_date = budgetConversionDateRange[0].format('YYYY-MM-DD');
        params.end_date = budgetConversionDateRange[1].format('YYYY-MM-DD');
      }
      const response = await reportsAPI.getBudgetConversion(params);
      setBudgetConversionData(response.data);
    } catch (error) {
      message.error('Erro ao carregar relatório de conversão');
      console.error('Error:', error);
    } finally {
      setLoadingBudgetConversion(false);
    }
  };

  const fetchOverduePaymentsReport = async () => {
    setLoadingOverduePayments(true);
    try {
      const response = await reportsAPI.getOverduePayments();
      setOverduePaymentsData(response.data);
    } catch (error) {
      message.error('Erro ao carregar relatório de inadimplência');
      console.error('Error:', error);
    } finally {
      setLoadingOverduePayments(false);
    }
  };

  useEffect(() => {
    fetchDashboard();
    fetchRevenueReport();
    fetchAttendanceReport();
    fetchProceduresReport();
    fetchBudgetConversionReport();
    fetchOverduePaymentsReport();
  }, []);

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value || 0);
  };

  const handleDownloadPDF = async (reportType, dateRange = null) => {
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      let response;
      let filename;

      switch (reportType) {
        case 'revenue':
          response = await reportsAPI.downloadRevenuePDF(params);
          filename = 'relatorio_receitas.pdf';
          break;
        case 'attendance':
          response = await reportsAPI.downloadAttendancePDF(params);
          filename = 'relatorio_atendimentos.pdf';
          break;
        case 'procedures':
          response = await reportsAPI.downloadProceduresPDF();
          filename = 'relatorio_procedimentos.pdf';
          break;
        case 'budget-conversion':
          response = await reportsAPI.downloadBudgetConversionPDF(params);
          filename = 'conversao_orcamentos.pdf';
          break;
        case 'overdue-payments':
          response = await reportsAPI.downloadOverduePaymentsPDF();
          filename = 'inadimplencia.pdf';
          break;
        default:
          return;
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

  const handleDownloadExcel = async (reportType, dateRange = null) => {
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      let response;
      let filename;

      switch (reportType) {
        case 'revenue':
          response = await reportsAPI.downloadRevenueExcel(params);
          filename = 'relatorio_receitas.xlsx';
          break;
        case 'attendance':
          response = await reportsAPI.downloadAttendanceExcel(params);
          filename = 'relatorio_atendimentos.xlsx';
          break;
        case 'procedures':
          response = await reportsAPI.downloadProceduresExcel();
          filename = 'relatorio_procedimentos.xlsx';
          break;
        case 'budget-conversion':
          response = await reportsAPI.downloadBudgetConversionExcel(params);
          filename = 'conversao_orcamentos.xlsx';
          break;
        case 'overdue-payments':
          response = await reportsAPI.downloadOverduePaymentsExcel();
          filename = 'inadimplencia.xlsx';
          break;
        default:
          return;
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

  return (
    <div>
      <Card
        title={
          <Space>
            <BarChartOutlined />
            <span>Relatórios e Indicadores</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
      >
        <Row gutter={16}>
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
      </Card>

      {/* Revenue Report */}
      <Card
        title={
          <Space>
            <DollarOutlined style={{ color: actionColors.create }} />
            <span>Relatório de Receitas</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
        extra={
          <Space>
            <RangePicker
              format="DD/MM/YYYY"
              placeholder={['Data Inicial', 'Data Final']}
              onChange={(dates) => {
                setRevenueDateRange(dates);
                if (dates) fetchRevenueReport();
              }}
              value={revenueDateRange}
              size="small"
            />
            <Button
              size="small"
              icon={<BarChartOutlined />}
              onClick={fetchRevenueReport}
              loading={loadingRevenue}
            >
              Atualizar
            </Button>
          </Space>
        }
      >
        <Spin spinning={loadingRevenue}>
          {revenueData ? (
            <div>
              <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col span={24}>
                  <Statistic
                    title="Receita Total"
                    value={revenueData.total_revenue}
                    formatter={(value) => formatCurrency(value)}
                    valueStyle={{ color: statusColors.success, fontSize: 24 }}
                  />
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} lg={12}>
                  <h4>Receita por Mês</h4>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={revenueData.by_month || []}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="month" />
                      <YAxis />
                      <Tooltip formatter={(value) => formatCurrency(value)} />
                      <Legend />
                      <Bar dataKey="total" fill={actionColors.create} name="Receita" />
                    </BarChart>
                  </ResponsiveContainer>
                </Col>
                <Col xs={24} lg={12}>
                  <h4>Receita por Método de Pagamento</h4>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={revenueData.by_method || []}
                        dataKey="total"
                        nameKey="payment_method"
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                        label={(entry) => `${translatePaymentMethod(entry.payment_method)}: ${formatCurrency(entry.total)}`}
                      >
                        {(revenueData.by_method || []).map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value, name) => [formatCurrency(value), translatePaymentMethod(name)]} />
                      <Legend formatter={(value) => translatePaymentMethod(value)} />
                    </PieChart>
                  </ResponsiveContainer>
                </Col>
              </Row>
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button
                      icon={<FilePdfOutlined />}
                      onClick={() => handleDownloadPDF('revenue', revenueDateRange)}
                      style={{
                        backgroundColor: actionColors.exportPDF,
                        borderColor: actionColors.exportPDF,
                        color: '#fff',
                      }}
                    >
                      Baixar PDF
                    </Button>
                    <Button
                      icon={<FileExcelOutlined />}
                      onClick={() => handleDownloadExcel('revenue', revenueDateRange)}
                      style={{
                        backgroundColor: actionColors.exportExcel,
                        borderColor: actionColors.exportExcel,
                        color: '#fff',
                      }}
                    >
                      Baixar Excel
                    </Button>
                  </Space>
                </Col>
              </Row>
            </div>
          ) : (
            <Empty description="Nenhum dado disponível" />
          )}
        </Spin>
      </Card>

      {/* Attendance Report */}
      <Card
        title={
          <Space>
            <CalendarOutlined style={{ color: actionColors.edit }} />
            <span>Relatório de Atendimentos</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
        extra={
          <Space>
            <RangePicker
              format="DD/MM/YYYY"
              placeholder={['Data Inicial', 'Data Final']}
              onChange={(dates) => {
                setAttendanceDateRange(dates);
                if (dates) fetchAttendanceReport();
              }}
              value={attendanceDateRange}
              size="small"
            />
            <Button
              size="small"
              icon={<CalendarOutlined />}
              onClick={fetchAttendanceReport}
              loading={loadingAttendance}
            >
              Atualizar
            </Button>
          </Space>
        }
      >
        <Spin spinning={loadingAttendance}>
          {attendanceData ? (
            <div>
              <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col xs={24} sm={8}>
                  <Statistic title="Total" value={attendanceData.total} />
                </Col>
                <Col xs={24} sm={8}>
                  <Statistic
                    title="Concluídos"
                    value={attendanceData.completed}
                    valueStyle={{ color: statusColors.success }}
                  />
                </Col>
                <Col xs={24} sm={8}>
                  <Statistic
                    title="Taxa de Comparecimento"
                    value={attendanceData.attendance_rate?.toFixed(2)}
                    suffix="%"
                    valueStyle={{ color: statusColors.success }}
                  />
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} lg={12}>
                  <h4>Distribuição de Status</h4>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={[
                          { name: 'Concluídos', value: attendanceData.completed, color: '#81C784' },
                          { name: 'Cancelados', value: attendanceData.cancelled, color: '#E57373' },
                          { name: 'Faltaram', value: attendanceData.no_show, color: '#FFD54F' },
                        ]}
                        dataKey="value"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                        label
                      >
                        {[
                          { name: 'Concluídos', value: attendanceData.completed, color: '#81C784' },
                          { name: 'Cancelados', value: attendanceData.cancelled, color: '#E57373' },
                          { name: 'Faltaram', value: attendanceData.no_show, color: '#FFD54F' },
                        ].map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                </Col>
                <Col xs={24} lg={12}>
                  <h4>Resumo</h4>
                  <div style={{ padding: '20px 0' }}>
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Tag color="success" style={{ width: '100%', padding: '8px', textAlign: 'center' }}>
                          Concluídos: {attendanceData.completed}
                        </Tag>
                      </Col>
                      <Col span={12}>
                        <Tag color="error" style={{ width: '100%', padding: '8px', textAlign: 'center' }}>
                          Cancelados: {attendanceData.cancelled}
                        </Tag>
                      </Col>
                      <Col span={12}>
                        <Tag color="warning" style={{ width: '100%', padding: '8px', textAlign: 'center' }}>
                          Faltaram: {attendanceData.no_show}
                        </Tag>
                      </Col>
                      <Col span={12}>
                        <Tag color="blue" style={{ width: '100%', padding: '8px', textAlign: 'center' }}>
                          Total: {attendanceData.total}
                        </Tag>
                      </Col>
                    </Row>
                  </div>
                </Col>
              </Row>
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button
                      icon={<FilePdfOutlined />}
                      onClick={() => handleDownloadPDF('attendance', attendanceDateRange)}
                      style={{
                        backgroundColor: actionColors.exportPDF,
                        borderColor: actionColors.exportPDF,
                        color: '#fff',
                      }}
                    >
                      Baixar PDF
                    </Button>
                    <Button
                      icon={<FileExcelOutlined />}
                      onClick={() => handleDownloadExcel('attendance', attendanceDateRange)}
                      style={{
                        backgroundColor: actionColors.exportExcel,
                        borderColor: actionColors.exportExcel,
                        color: '#fff',
                      }}
                    >
                      Baixar Excel
                    </Button>
                  </Space>
                </Col>
              </Row>
            </div>
          ) : (
            <Empty description="Nenhum dado disponível" />
          )}
        </Spin>
      </Card>

      {/* Procedures Report */}
      <Card
        title={
          <Space>
            <FileTextOutlined style={{ color: actionColors.view }} />
            <span>Relatório de Procedimentos</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
        extra={
          <Button
            size="small"
            icon={<FileTextOutlined />}
            onClick={fetchProceduresReport}
            loading={loadingProcedures}
          >
            Atualizar
          </Button>
        }
      >
        <Spin spinning={loadingProcedures}>
          {proceduresData && proceduresData.procedures ? (
            <div>
              <ResponsiveContainer width="100%" height={400}>
                <BarChart data={proceduresData.procedures.slice(0, 10)}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="procedure" angle={-45} textAnchor="end" height={100} />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="count" fill={actionColors.view} name="Quantidade" />
                </BarChart>
              </ResponsiveContainer>
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button
                      icon={<FilePdfOutlined />}
                      onClick={() => handleDownloadPDF('procedures')}
                      style={{
                        backgroundColor: actionColors.exportPDF,
                        borderColor: actionColors.exportPDF,
                        color: '#fff',
                      }}
                    >
                      Baixar PDF
                    </Button>
                    <Button
                      icon={<FileExcelOutlined />}
                      onClick={() => handleDownloadExcel('procedures')}
                      style={{
                        backgroundColor: actionColors.exportExcel,
                        borderColor: actionColors.exportExcel,
                        color: '#fff',
                      }}
                    >
                      Baixar Excel
                    </Button>
                  </Space>
                </Col>
              </Row>
            </div>
          ) : (
            <Empty description="Nenhum dado disponível" />
          )}
        </Spin>
      </Card>

      {/* Budget Conversion Report */}
      <Card
        title={
          <Space>
            <FundOutlined style={{ color: actionColors.approve }} />
            <span>Taxa de Conversão de Orçamentos</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
        extra={
          <Space>
            <RangePicker
              format="DD/MM/YYYY"
              placeholder={['Data Inicial', 'Data Final']}
              onChange={(dates) => {
                setBudgetConversionDateRange(dates);
                if (dates) fetchBudgetConversionReport();
              }}
              value={budgetConversionDateRange}
              size="small"
            />
            <Button
              size="small"
              icon={<FundOutlined />}
              onClick={fetchBudgetConversionReport}
              loading={loadingBudgetConversion}
            >
              Atualizar
            </Button>
          </Space>
        }
      >
        <Spin spinning={loadingBudgetConversion}>
          {budgetConversionData ? (
            <div>
              <Row gutter={16} style={{ marginBottom: 24 }}>
                <Col xs={24} md={8}>
                  <Card style={{ textAlign: 'center' }}>
                    <Statistic title="Total de Orçamentos" value={budgetConversionData.total_budgets || 0} />
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ textAlign: 'center' }}>
                    <Statistic
                      title="Orçamentos Aprovados"
                      value={budgetConversionData.approved_budgets || 0}
                      valueStyle={{ color: statusColors.success }}
                    />
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ textAlign: 'center' }}>
                    <Statistic
                      title="Valor Total Aprovado"
                      value={budgetConversionData.total_approved || 0}
                      formatter={(value) => formatCurrency(value)}
                      valueStyle={{ color: statusColors.success }}
                    />
                  </Card>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col xs={24} lg={12}>
                  <h4 style={{ textAlign: 'center' }}>Taxa de Conversão</h4>
                  <div style={{ textAlign: 'center', padding: '40px 0' }}>
                    <Progress
                      type="circle"
                      percent={parseFloat(budgetConversionData.conversion_rate?.toFixed(2) || 0)}
                      format={(percent) => `${percent}%`}
                      strokeColor={{
                        '0%': '#64B5F6',
                        '100%': '#81C784',
                      }}
                      width={200}
                    />
                  </div>
                </Col>
                <Col xs={24} lg={12}>
                  <h4>Distribuição por Status</h4>
                  {budgetConversionData.by_status && budgetConversionData.by_status.length > 0 ? (
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={budgetConversionData.by_status}
                          dataKey="count"
                          nameKey="status"
                          cx="50%"
                          cy="50%"
                          outerRadius={80}
                          label={(entry) => `${translateBudgetStatus(entry.status)}: ${entry.count}`}
                        >
                          {budgetConversionData.by_status.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip formatter={(value, name) => [value, translateBudgetStatus(name)]} />
                        <Legend formatter={(value) => translateBudgetStatus(value)} />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <Empty description="Nenhum dado de status disponível" />
                  )}
                </Col>
              </Row>
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button
                      icon={<FilePdfOutlined />}
                      onClick={() => handleDownloadPDF('budget-conversion', budgetConversionDateRange)}
                      style={{
                        backgroundColor: actionColors.exportPDF,
                        borderColor: actionColors.exportPDF,
                        color: '#fff',
                      }}
                    >
                      Baixar PDF
                    </Button>
                    <Button
                      icon={<FileExcelOutlined />}
                      onClick={() => handleDownloadExcel('budget-conversion', budgetConversionDateRange)}
                      style={{
                        backgroundColor: actionColors.exportExcel,
                        borderColor: actionColors.exportExcel,
                        color: '#fff',
                      }}
                    >
                      Baixar Excel
                    </Button>
                  </Space>
                </Col>
              </Row>
            </div>
          ) : (
            <Empty description="Nenhum dado disponível" />
          )}
        </Spin>
      </Card>

      {/* Overdue Payments Report */}
      <Card
        title={
          <Space>
            <AlertOutlined style={{ color: actionColors.delete }} />
            <span>Controle de Inadimplência</span>
          </Space>
        }
        style={{ boxShadow: shadows.small, marginBottom: 24 }}
        extra={
          <Button
            size="small"
            icon={<AlertOutlined />}
            onClick={fetchOverduePaymentsReport}
            loading={loadingOverduePayments}
          >
            Atualizar
          </Button>
        }
      >
        <Spin spinning={loadingOverduePayments}>
          {overduePaymentsData ? (
            <div>
              <Row gutter={16} style={{ marginBottom: 24 }}>
                <Col xs={24} md={12}>
                  <Card style={{ textAlign: 'center', borderColor: '#E57373' }}>
                    <Statistic
                      title="Total em Atraso"
                      value={overduePaymentsData.total_overdue || 0}
                      formatter={(value) => formatCurrency(value)}
                      valueStyle={{ color: '#E57373', fontSize: 24 }}
                    />
                  </Card>
                </Col>
                <Col xs={24} md={12}>
                  <Card style={{ textAlign: 'center', borderColor: '#E57373' }}>
                    <Statistic
                      title="Quantidade de Pagamentos Atrasados"
                      value={overduePaymentsData.overdue_count || 0}
                      valueStyle={{ color: '#E57373', fontSize: 24 }}
                    />
                  </Card>
                </Col>
              </Row>
              {overduePaymentsData.overdue_by_age && overduePaymentsData.overdue_by_age.length > 0 ? (
                <Row gutter={16}>
                  <Col xs={24} lg={12}>
                    <h4>Valores por Tempo de Atraso</h4>
                    <ResponsiveContainer width="100%" height={300}>
                      <BarChart data={overduePaymentsData.overdue_by_age}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="age_range" />
                        <YAxis />
                        <Tooltip formatter={(value) => formatCurrency(value)} />
                        <Legend />
                        <Bar dataKey="total" fill="#E57373" name="Valor em Atraso" />
                      </BarChart>
                    </ResponsiveContainer>
                  </Col>
                  <Col xs={24} lg={12}>
                    <h4>Quantidade por Tempo de Atraso</h4>
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={overduePaymentsData.overdue_by_age}
                          dataKey="count"
                          nameKey="age_range"
                          cx="50%"
                          cy="50%"
                          outerRadius={80}
                          label
                        >
                          {overduePaymentsData.overdue_by_age.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </Col>
                </Row>
              ) : (
                <div style={{ textAlign: 'center', padding: '40px', color: '#81C784' }}>
                  <h2>✓ Nenhum pagamento em atraso!</h2>
                  <p style={{ fontSize: 16 }}>Todos os pagamentos estão em dia.</p>
                </div>
              )}
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button
                      icon={<FilePdfOutlined />}
                      onClick={() => handleDownloadPDF('overdue-payments')}
                      style={{
                        backgroundColor: actionColors.exportPDF,
                        borderColor: actionColors.exportPDF,
                        color: '#fff',
                      }}
                    >
                      Baixar PDF
                    </Button>
                    <Button
                      icon={<FileExcelOutlined />}
                      onClick={() => handleDownloadExcel('overdue-payments')}
                      style={{
                        backgroundColor: actionColors.exportExcel,
                        borderColor: actionColors.exportExcel,
                        color: '#fff',
                      }}
                    >
                      Baixar Excel
                    </Button>
                  </Space>
                </Col>
              </Row>
            </div>
          ) : (
            <Empty description="Nenhum dado disponível" />
          )}
        </Spin>
      </Card>
    </div>
  );
};

export default Reports;
