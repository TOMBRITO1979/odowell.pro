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
  Progress,
  Typography,
  Divider,
} from 'antd';
import {
  CalendarOutlined,
  UserOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  DownloadOutlined,
  SyncOutlined,
  TrophyOutlined,
  RiseOutlined,
} from '@ant-design/icons';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import dayjs from 'dayjs';
import { reportsAPI } from '../services/api';
import { actionColors, statusColors, shadows } from '../theme/designSystem';

const { RangePicker } = DatePicker;
const { Title, Text } = Typography;

const COLORS = ['#52c41a', '#ff4d4f', '#faad14', '#1890ff', '#722ed1', '#eb2f96'];

// Cores suaves e foscas para cada profissional (não agridem a vista)
const PROFESSIONAL_COLORS = [
  '#a7d7a7', // verde suave
  '#b8d4e8', // azul suave
  '#d4c4e8', // lilás suave
  '#f5e0a8', // amarelo suave
  '#e8c4d4', // rosa suave
  '#a8d8d8', // teal suave
  '#e8d0b8', // pêssego suave
  '#c4c8e8', // indigo suave
  '#b8e0c8', // menta suave
  '#d4e0a8', // lima suave
];

// Cores específicas para status de orçamentos (tons foscos/suaves)
const BUDGET_STATUS_COLORS = {
  'approved': '#4a8c6f',   // Verde fosco para aprovado
  'pending': '#c9a227',    // Amarelo fosco para pendente
  'rejected': '#b45454',   // Vermelho fosco para rejeitado
  'cancelled': '#8b7355',  // Marrom acinzentado para cancelado
};

// Tradução de status de orçamentos
const translateStatus = (status) => {
  const translations = {
    'approved': 'Aprovado',
    'cancelled': 'Cancelado',
    'pending': 'Pendente',
    'rejected': 'Rejeitado',
  };
  return translations[status] || status;
};

// Função para obter cor do status do orçamento
const getBudgetStatusColor = (status) => {
  return BUDGET_STATUS_COLORS[status] || '#1890ff';
};

const Dashboard = () => {
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState([dayjs().subtract(30, 'days'), dayjs()]);
  const [dashboardData, setDashboardData] = useState(null);

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await reportsAPI.getAdvancedDashboard(params);
      setDashboardData(response.data);
    } catch (error) {
      message.error('Erro ao carregar dados do dashboard');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const handleDateRangeChange = (dates) => {
    setDateRange(dates);
  };

  const handleRefresh = () => {
    fetchDashboardData();
  };

  const exportToCSV = () => {
    if (!dashboardData) return;

    const csvData = [];
    csvData.push(['Dashboard Executivo - OdoWell']);
    csvData.push(['Período', `${dateRange[0].format('DD/MM/YYYY')} - ${dateRange[1].format('DD/MM/YYYY')}`]);
    csvData.push([]);
    csvData.push(['Indicadores Principais']);
    csvData.push(['Total de Agendamentos', dashboardData.total_appointments]);
    csvData.push(['Consultas Concluídas', dashboardData.completed_appointments]);
    csvData.push(['Consultas Canceladas', dashboardData.cancelled_appointments]);
    csvData.push(['Faltas (No-Show)', dashboardData.no_shows]);
    csvData.push(['Remarcações', dashboardData.reschedules]);
    csvData.push(['Taxa de Comparecimento', `${dashboardData.attendance_rate?.toFixed(2)}%`]);
    csvData.push([]);
    csvData.push(['Pacientes']);
    csvData.push(['Total de Pacientes Ativos', dashboardData.total_patients]);
    csvData.push(['Novos Pacientes no Período', dashboardData.new_patients]);
    csvData.push([]);
    csvData.push(['Orçamentos']);
    if (dashboardData.budget_status && dashboardData.budget_status.length > 0) {
      csvData.push(['Status', 'Quantidade']);
      dashboardData.budget_status.forEach((item) => {
        csvData.push([item.status, item.count]);
      });
    }

    const csvContent = csvData.map((row) => row.join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `dashboard_${dayjs().format('YYYY-MM-DD')}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    message.success('CSV gerado com sucesso!');
  };

  const exportToPDF = async () => {
    if (!dashboardData) return;

    try {
      const params = {};
      if (dateRange && dateRange[0] && dateRange[1]) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }

      const response = await reportsAPI.downloadDashboardPDF(params);
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.setAttribute('href', url);
      link.setAttribute('download', `dashboard_${dayjs().format('YYYY-MM-DD')}.pdf`);
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      message.success('PDF gerado com sucesso!');
    } catch (error) {
      message.error('Erro ao gerar PDF');
      console.error('Error:', error);
    }
  };

  return (
    <div style={{ background: '#f0f2f5', minHeight: '100vh', padding: '24px' }}>
      <Card
        style={{
          marginBottom: 24,
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: '#fff',
          boxShadow: shadows.large,
        }}
      >
        <Row align="middle" justify="space-between">
          <Col>
            <Title level={2} style={{ color: '#fff', margin: 0 }}>
              <TrophyOutlined /> Dashboard Executivo
            </Title>
            <Text style={{ color: '#fff', fontSize: 16 }}>
              Visão geral completa do desempenho da clínica
            </Text>
          </Col>
          <Col>
            <Space>
              <RangePicker
                value={dateRange}
                onChange={handleDateRangeChange}
                format="DD/MM/YYYY"
                style={{ width: 280 }}
              />
              <Button type="primary" icon={<SyncOutlined />} onClick={handleRefresh} loading={loading}>
                Atualizar
              </Button>
              <Button
                icon={<DownloadOutlined />}
                onClick={exportToPDF}
                disabled={!dashboardData}
                style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
              >
                Exportar PDF
              </Button>
              <Button
                icon={<DownloadOutlined />}
                onClick={exportToCSV}
                disabled={!dashboardData}
                style={{ backgroundColor: actionColors.exportExcel, borderColor: actionColors.exportExcel, color: '#fff' }}
              >
                Exportar CSV
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Spin spinning={loading}>
        {dashboardData && (
          <>
            {/* KPIs Principais */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={24} sm={12} lg={6}>
                <Card hoverable style={{ boxShadow: shadows.medium, borderLeft: `4px solid ${statusColors.success}` }}>
                  <Statistic
                    title={<Text strong>Total de Consultas</Text>}
                    value={dashboardData.total_appointments}
                    prefix={<CalendarOutlined style={{ color: statusColors.success }} />}
                    valueStyle={{ color: statusColors.success, fontSize: 32 }}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Card hoverable style={{ boxShadow: shadows.medium, borderLeft: `4px solid #1890ff` }}>
                  <Statistic
                    title={<Text strong>Consultas Concluídas</Text>}
                    value={dashboardData.completed_appointments}
                    prefix={<CheckCircleOutlined style={{ color: '#1890ff' }} />}
                    valueStyle={{ color: '#1890ff', fontSize: 32 }}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Card hoverable style={{ boxShadow: shadows.medium, borderLeft: `4px solid ${statusColors.error}` }}>
                  <Statistic
                    title={<Text strong>Faltas (No-Show)</Text>}
                    value={dashboardData.no_shows}
                    prefix={<CloseCircleOutlined style={{ color: statusColors.error }} />}
                    valueStyle={{ color: statusColors.error, fontSize: 32 }}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Card hoverable style={{ boxShadow: shadows.medium, borderLeft: `4px solid ${statusColors.warning}` }}>
                  <Statistic
                    title={<Text strong>Remarcações</Text>}
                    value={dashboardData.reschedules}
                    prefix={<ClockCircleOutlined style={{ color: statusColors.warning }} />}
                    valueStyle={{ color: statusColors.warning, fontSize: 32 }}
                  />
                </Card>
              </Col>
            </Row>

            {/* Taxa de Comparecimento e Pacientes */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={24} lg={8}>
                <Card title="Taxa de Comparecimento" style={{ boxShadow: shadows.medium, height: '100%' }}>
                  <div style={{ textAlign: 'center', padding: '20px 0' }}>
                    <Progress
                      type="circle"
                      percent={parseFloat(dashboardData.attendance_rate?.toFixed(2) || 0)}
                      format={(percent) => `${percent}%`}
                      strokeColor={{
                        '0%': '#ff4d4f',
                        '50%': '#faad14',
                        '100%': '#52c41a',
                      }}
                      width={180}
                    />
                    <div style={{ marginTop: 16 }}>
                      <Text type="secondary">Taxa de presença nas consultas agendadas</Text>
                    </div>
                  </div>
                </Card>
              </Col>
              <Col xs={24} lg={8}>
                <Card title="Status dos Agendamentos" style={{ boxShadow: shadows.medium, height: '100%' }}>
                  <Row gutter={[16, 16]}>
                    <Col span={12}>
                      <Statistic
                        title="Concluídas"
                        value={dashboardData.completed_appointments}
                        valueStyle={{ color: statusColors.success }}
                        prefix={<CheckCircleOutlined />}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="Canceladas"
                        value={dashboardData.cancelled_appointments}
                        valueStyle={{ color: statusColors.error }}
                        prefix={<CloseCircleOutlined />}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="Faltas"
                        value={dashboardData.no_shows}
                        valueStyle={{ color: statusColors.warning }}
                        prefix={<CloseCircleOutlined />}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="Remarcadas"
                        value={dashboardData.reschedules}
                        valueStyle={{ color: '#1890ff' }}
                        prefix={<SyncOutlined />}
                      />
                    </Col>
                  </Row>
                </Card>
              </Col>
              <Col xs={24} lg={8}>
                <Card title="Pacientes" style={{ boxShadow: shadows.medium, height: '100%' }}>
                  <div style={{ padding: '20px 0' }}>
                    <Statistic
                      title="Total de Pacientes Ativos"
                      value={dashboardData.total_patients}
                      prefix={<UserOutlined />}
                      valueStyle={{ color: statusColors.success, fontSize: 40 }}
                    />
                    <Divider />
                    <Statistic
                      title="Novos Pacientes no Período"
                      value={dashboardData.new_patients}
                      prefix={<RiseOutlined />}
                      valueStyle={{ color: '#1890ff', fontSize: 28 }}
                    />
                  </div>
                </Card>
              </Col>
            </Row>

            {/* Gráficos de Tendência */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={24} lg={16}>
                <Card title="Agendamentos por Dia" style={{ boxShadow: shadows.medium }}>
                  {dashboardData.daily_appointments && dashboardData.daily_appointments.length > 0 ? (
                    <ResponsiveContainer width="100%" height={300}>
                      <AreaChart data={dashboardData.daily_appointments}>
                        <defs>
                          <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#1890ff" stopOpacity={0.8} />
                            <stop offset="95%" stopColor="#1890ff" stopOpacity={0.1} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" angle={-45} textAnchor="end" height={80} />
                        <YAxis />
                        <Tooltip />
                        <Area
                          type="monotone"
                          dataKey="count"
                          stroke="#1890ff"
                          fillOpacity={1}
                          fill="url(#colorCount)"
                          name="Agendamentos"
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                      <Text type="secondary">Nenhum agendamento no período selecionado</Text>
                    </div>
                  )}
                </Card>
              </Col>
              <Col xs={24} lg={8}>
                <Card title="Orçamentos por Status" style={{ boxShadow: shadows.medium }}>
                  {dashboardData.budget_status && dashboardData.budget_status.length > 0 ? (
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={dashboardData.budget_status}
                          dataKey="count"
                          nameKey="status"
                          cx="50%"
                          cy="50%"
                          outerRadius={80}
                          label={(entry) => `${translateStatus(entry.status)}: ${entry.count}`}
                        >
                          {dashboardData.budget_status.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={getBudgetStatusColor(entry.status)} />
                          ))}
                        </Pie>
                        <Tooltip formatter={(value, name) => [value, translateStatus(name)]} />
                        <Legend formatter={(value) => translateStatus(value)} />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                      <Text type="secondary">Nenhum orçamento no período</Text>
                    </div>
                  )}
                </Card>
              </Col>
            </Row>

            {/* Gráficos de Performance */}
            <Row gutter={[16, 16]}>
              <Col xs={24} lg={12}>
                <Card title="Consultas por Profissional" style={{ boxShadow: shadows.medium }}>
                  {dashboardData.professional_appointments && dashboardData.professional_appointments.length > 0 ? (
                    <ResponsiveContainer width="100%" height={350}>
                      <BarChart data={dashboardData.professional_appointments}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="professional" angle={-45} textAnchor="end" height={120} />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Bar dataKey="count" name="Consultas Realizadas">
                          {dashboardData.professional_appointments.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={PROFESSIONAL_COLORS[index % PROFESSIONAL_COLORS.length]} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                      <Text type="secondary">Nenhuma consulta concluída no período</Text>
                    </div>
                  )}
                </Card>
              </Col>
              <Col xs={24} lg={12}>
                <Card title="Orçamentos por Dia" style={{ boxShadow: shadows.medium }}>
                  {dashboardData.daily_budgets && dashboardData.daily_budgets.length > 0 ? (
                    <ResponsiveContainer width="100%" height={350}>
                      <LineChart data={dashboardData.daily_budgets}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" angle={-45} textAnchor="end" height={120} />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Line type="monotone" dataKey="count" stroke="#722ed1" name="Orçamentos Criados" strokeWidth={2} />
                      </LineChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                      <Text type="secondary">Nenhum orçamento no período</Text>
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </>
        )}
      </Spin>
    </div>
  );
};

export default Dashboard;
