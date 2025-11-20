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
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { reportsAPI } from '../../services/api';

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
      >
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Total de Pacientes"
                value={dashboard.total_patients}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Agendamentos"
                value={dashboard.total_appointments}
                prefix={<CalendarOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="Receita Total"
                value={dashboard.total_revenue}
                prefix={<DollarOutlined />}
                valueStyle={{ color: '#52c41a' }}
                formatter={(value) => formatCurrency(value)}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="A Receber"
                value={dashboard.pending_payments}
                prefix={<FileTextOutlined />}
                valueStyle={{ color: '#faad14' }}
                formatter={(value) => formatCurrency(value)}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={24}>
            <Card title="Relatórios Disponíveis" loading={loading}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Button
                  type="primary"
                  block
                  icon={<BarChartOutlined />}
                  onClick={handleRevenueReport}
                  loading={loading}
                >
                  Relatório de Receitas
                </Button>
                <Button
                  type="primary"
                  block
                  icon={<CalendarOutlined />}
                  onClick={handleAttendanceReport}
                  loading={loading}
                >
                  Relatório de Atendimentos
                </Button>
                <Button
                  type="primary"
                  block
                  icon={<FileTextOutlined />}
                  onClick={handleProceduresReport}
                  loading={loading}
                >
                  Relatório de Procedimentos
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
            type="default"
            icon={<FileExcelOutlined />}
            onClick={handleDownloadExcel}
            style={{ backgroundColor: '#217346', color: 'white', borderColor: '#217346' }}
          >
            Baixar Excel
          </Button>,
          <Button
            key="pdf"
            type="primary"
            danger
            icon={<FilePdfOutlined />}
            onClick={handleDownloadPDF}
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
