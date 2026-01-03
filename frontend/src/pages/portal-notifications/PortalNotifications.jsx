import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Tag,
  Space,
  Typography,
  Select,
  Empty,
  Spin,
  Badge,
  Tooltip,
  Row,
  Col,
  Statistic,
} from 'antd';
import {
  BellOutlined,
  CalendarOutlined,
  UserOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlusCircleOutlined,
  MinusCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/pt-br';
import { portalNotificationsAPI } from '../../services/api';

dayjs.extend(relativeTime);
dayjs.locale('pt-br');

const { Title, Text } = Typography;

const actionLabels = {
  create: 'Agendamento',
  cancel: 'Cancelamento',
};

const actionColors = {
  create: 'green',
  cancel: 'red',
};

const actionIcons = {
  create: <PlusCircleOutlined />,
  cancel: <MinusCircleOutlined />,
};

const procedureLabels = {
  consultation: 'Consulta',
  cleaning: 'Limpeza',
  filling: 'Restauracao',
  extraction: 'Extracao',
  root_canal: 'Canal',
  orthodontics: 'Ortodontia',
  whitening: 'Clareamento',
  prosthesis: 'Protese',
  implant: 'Implante',
  emergency: 'Emergencia',
  other: 'Outro',
};

const PortalNotifications = () => {
  const [loading, setLoading] = useState(true);
  const [notifications, setNotifications] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [filterAction, setFilterAction] = useState('');
  const [stats, setStats] = useState({ creates: 0, cancels: 0 });

  useEffect(() => {
    fetchNotifications();
  }, [page, pageSize, filterAction]);

  const fetchNotifications = async () => {
    setLoading(true);
    try {
      const params = {
        page,
        page_size: pageSize,
      };
      if (filterAction) {
        params.action = filterAction;
      }

      const response = await portalNotificationsAPI.getAll(params);
      const data = response.data.notifications || [];
      setNotifications(data);
      setTotal(response.data.total || 0);

      // Calculate stats
      const creates = data.filter(n => n.action === 'create').length;
      const cancels = data.filter(n => n.action === 'cancel').length;
      setStats({ creates, cancels });
    } catch (error) {
      console.error('Error fetching notifications:', error);
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: 'Acao',
      dataIndex: 'action',
      key: 'action',
      width: 130,
      render: (action) => (
        <Tag icon={actionIcons[action]} color={actionColors[action]}>
          {actionLabels[action] || action}
        </Tag>
      ),
    },
    {
      title: 'Paciente',
      key: 'patient',
      render: (_, record) => (
        <Space>
          <UserOutlined />
          <Text strong>{record.appointment?.patient_name || 'N/A'}</Text>
        </Space>
      ),
    },
    {
      title: 'Consulta Agendada',
      key: 'appointment',
      render: (_, record) => {
        if (!record.appointment) return <Text type="secondary">-</Text>;
        return (
          <Space direction="vertical" size={0}>
            <Space>
              <CalendarOutlined />
              <Text>{dayjs(record.appointment.start_time).format('DD/MM/YYYY')}</Text>
            </Space>
            <Space>
              <ClockCircleOutlined />
              <Text type="secondary">
                {dayjs(record.appointment.start_time).format('HH:mm')} - {dayjs(record.appointment.end_time).format('HH:mm')}
              </Text>
            </Space>
          </Space>
        );
      },
    },
    {
      title: 'Procedimento',
      key: 'procedure',
      render: (_, record) => (
        <Text>{procedureLabels[record.appointment?.procedure] || record.appointment?.procedure || '-'}</Text>
      ),
    },
    {
      title: 'Profissional',
      key: 'dentist',
      render: (_, record) => (
        <Text>{record.appointment?.dentist_name || '-'}</Text>
      ),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => {
        const status = record.appointment?.status;
        const colors = {
          scheduled: 'blue',
          confirmed: 'green',
          cancelled: 'red',
          completed: 'default',
        };
        const labels = {
          scheduled: 'Agendado',
          confirmed: 'Confirmado',
          cancelled: 'Cancelado',
          completed: 'Concluido',
        };
        return status ? (
          <Tag color={colors[status]}>{labels[status] || status}</Tag>
        ) : '-';
      },
    },
    {
      title: 'Quando',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date) => (
        <Tooltip title={dayjs(date).format('DD/MM/YYYY HH:mm:ss')}>
          <Text type="secondary">{dayjs(date).fromNow()}</Text>
        </Tooltip>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <BellOutlined />
            <Title level={4} style={{ margin: 0 }}>
              Notificacoes do Portal
            </Title>
          </Space>
        }
        extra={
          <Select
            placeholder="Filtrar por acao"
            allowClear
            style={{ width: 180 }}
            value={filterAction || undefined}
            onChange={(value) => {
              setFilterAction(value || '');
              setPage(1);
            }}
          >
            <Select.Option value="create">
              <Tag color="green" icon={<PlusCircleOutlined />}>Agendamentos</Tag>
            </Select.Option>
            <Select.Option value="cancel">
              <Tag color="red" icon={<MinusCircleOutlined />}>Cancelamentos</Tag>
            </Select.Option>
          </Select>
        }
      >
        {/* Stats Row */}
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={12} sm={8} md={6}>
            <Card size="small" style={{ background: '#f6ffed', borderColor: '#b7eb8f' }}>
              <Statistic
                title="Agendamentos"
                value={stats.creates}
                prefix={<PlusCircleOutlined style={{ color: '#52c41a' }} />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6}>
            <Card size="small" style={{ background: '#fff2f0', borderColor: '#ffccc7' }}>
              <Statistic
                title="Cancelamentos"
                value={stats.cancels}
                prefix={<MinusCircleOutlined style={{ color: '#ff4d4f' }} />}
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8} md={12}>
            <Card size="small">
              <Text type="secondary">
                Atividades recentes do portal do paciente. Aqui voce ve quando pacientes agendam ou cancelam consultas pelo portal online.
              </Text>
            </Card>
          </Col>
        </Row>

        {loading ? (
          <div style={{ textAlign: 'center', padding: 50 }}>
            <Spin size="large" />
          </div>
        ) : notifications.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="Nenhuma atividade do portal ainda"
          />
        ) : (
          <Table
            columns={columns}
            dataSource={notifications}
            rowKey="id"
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              showTotal: (total) => `Total: ${total} notificacoes`,
              onChange: (p, ps) => {
                setPage(p);
                setPageSize(ps);
              },
            }}
            scroll={{ x: 900 }}
          />
        )}
      </Card>
    </div>
  );
};

export default PortalNotifications;
