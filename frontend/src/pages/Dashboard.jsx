import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Table, Tag } from 'antd';
import {
  UserOutlined,
  CalendarOutlined,
  FileTextOutlined,
  CheckSquareOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { reportsAPI, appointmentsAPI } from '../services/api';
import { statusColors, brandColors, spacing, shadows } from '../theme/designSystem';

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
        const statusMap = {
          scheduled: { color: statusColors.pending, text: 'Agendado' },
          confirmed: { color: statusColors.approved, text: 'Confirmado' },
          completed: { color: statusColors.success, text: 'Concluído' },
          cancelled: { color: statusColors.cancelled, text: 'Cancelado' },
        };
        const statusInfo = statusMap[status] || { color: statusColors.pending, text: status };
        return <Tag color={statusInfo.color}>{statusInfo.text}</Tag>;
      },
    },
  ];

  return (
    <div>
      <h1 style={{ marginBottom: spacing.lg }}>Dashboard</h1>

      <Row gutter={[spacing.md, spacing.md]}>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              boxShadow: shadows.small,
              transition: 'all 0.3s ease',
            }}
          >
            <Statistic
              title="Total de Pacientes"
              value={stats.total_patients || 0}
              prefix={<UserOutlined style={{ color: brandColors.primary }} />}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              boxShadow: shadows.small,
              transition: 'all 0.3s ease',
            }}
          >
            <Statistic
              title="Consultas Hoje"
              value={stats.appointments_today || 0}
              prefix={<CalendarOutlined style={{ color: brandColors.primary }} />}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              boxShadow: shadows.small,
              transition: 'all 0.3s ease',
            }}
          >
            <Statistic
              title="Orçamentos Pendentes"
              value={stats.pending_budgets || 0}
              prefix={<FileTextOutlined style={{ color: statusColors.pending }} />}
              valueStyle={{ color: stats.pending_budgets > 0 ? statusColors.pending : undefined }}
              loading={loading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              boxShadow: shadows.small,
              transition: 'all 0.3s ease',
            }}
          >
            <Statistic
              title="Tarefas Pendentes"
              value={stats.pending_tasks || 0}
              prefix={<CheckSquareOutlined style={{ color: statusColors.inProgress }} />}
              valueStyle={{ color: stats.pending_tasks > 0 ? statusColors.inProgress : undefined }}
              loading={loading}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="Próximos Agendamentos"
        style={{
          marginTop: spacing.lg,
          boxShadow: shadows.small,
        }}
      >
        <Table
          columns={columns}
          dataSource={appointments}
          rowKey="id"
          pagination={false}
          loading={loading}
          scroll={{ x: 600 }}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
