import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Table, Tag } from 'antd';
import {
  UserOutlined,
  CalendarOutlined,
  DollarOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { reportsAPI, appointmentsAPI } from '../services/api';

const Dashboard = () => {
  const [stats, setStats] = useState({});
  const [appointments, setAppointments] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboard();
  }, []);

  const loadDashboard = async () => {
    try {
      const [statsRes, appointmentsRes] = await Promise.all([
        reportsAPI.getDashboard(),
        appointmentsAPI.getAll({ page: 1, page_size: 5, start_date: new Date().toISOString().split('T')[0] }),
      ]);
      setStats(statsRes.data);
      setAppointments(appointmentsRes.data.appointments || []);
    } catch (error) {
      console.error('Error loading dashboard:', error);
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient',
    },
    {
      title: 'Horário',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (text) => new Date(text).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const colors = {
          scheduled: 'blue',
          confirmed: 'green',
          completed: 'success',
          cancelled: 'red',
        };
        return <Tag color={colors[status]}>{status}</Tag>;
      },
    },
  ];

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>Dashboard</h1>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total de Pacientes"
              value={stats.total_patients || 0}
              prefix={<UserOutlined />}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Consultas Hoje"
              value={stats.appointments_today || 0}
              prefix={<CalendarOutlined />}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Faturamento Mensal"
              value={stats.revenue_month || 0}
              prefix={<DollarOutlined />}
              precision={2}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Estoque Baixo"
              value={stats.low_stock_count || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: stats.low_stock_count > 0 ? '#cf1322' : undefined }}
              loading={loading}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Próximos Agendamentos" style={{ marginTop: 24 }}>
        <Table
          columns={columns}
          dataSource={appointments}
          rowKey="id"
          pagination={false}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
