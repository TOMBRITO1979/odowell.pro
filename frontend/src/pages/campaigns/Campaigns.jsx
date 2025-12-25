import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Card,
  message,
  Popconfirm,
  Row,
  Col,
  Select,
  Progress,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SendOutlined,
  NotificationOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { campaignsAPI } from '../../services/api';
import { actionColors } from '../../theme/designSystem';

// Configurar timezone para Brasil
dayjs.extend(utc);
dayjs.extend(timezone);
const BRAZIL_TZ = 'America/Sao_Paulo';

const Campaigns = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [campaigns, setCampaigns] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const [filters, setFilters] = useState({
    status: undefined,
    type: undefined,
  });

  const statusOptions = [
    { value: 'draft', label: 'Rascunho', color: 'default' },
    { value: 'scheduled', label: 'Agendada', color: 'processing' },
    { value: 'sending', label: 'Enviando', color: 'warning' },
    { value: 'sent', label: 'Enviada', color: 'success' },
    { value: 'failed', label: 'Falhou', color: 'error' },
  ];

  const typeOptions = [
    { value: 'whatsapp', label: 'WhatsApp', color: 'green' },
    { value: 'email', label: 'Email', color: 'blue' },
    { value: 'sms', label: 'SMS', color: 'orange' },
  ];

  useEffect(() => {
    fetchCampaigns();
  }, [pagination.current, pagination.pageSize, filters]);

  const fetchCampaigns = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      };

      const response = await campaignsAPI.getAll(params);
      setCampaigns(response.data.campaigns || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar campanhas');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await campaignsAPI.delete(id);
      message.success('Campanha excluída com sucesso');
      fetchCampaigns();
    } catch (error) {
      message.error('Erro ao excluir campanha');
    }
  };

  const handleSend = async (id) => {
    try {
      setLoading(true);
      await campaignsAPI.send(id);
      message.success('Campanha enviada com sucesso!');
      fetchCampaigns();
    } catch (error) {
      message.error('Erro ao enviar campanha');
    } finally {
      setLoading(false);
    }
  };

  const getStatusTag = (status) => {
    const statusObj = statusOptions.find((s) => s.value === status);
    return statusObj ? (
      <Tag color={statusObj.color}>{statusObj.label}</Tag>
    ) : (
      <Tag>{status}</Tag>
    );
  };

  const getTypeTag = (type) => {
    const typeObj = typeOptions.find((t) => t.value === type);
    return typeObj ? (
      <Tag color={typeObj.color}>{typeObj.label}</Tag>
    ) : (
      <Tag>{type}</Tag>
    );
  };

  const columns = [
    {
      title: 'Nome',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Tipo',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Dest.',
      dataIndex: 'total_recipients',
      key: 'total_recipients',
      width: 80,
      align: 'center',
    },
    {
      title: 'Progresso',
      key: 'progress',
      width: 120,
      render: (_, record) => {
        if (record.total_recipients === 0) return '-';
        const percent = Math.round((record.sent / record.total_recipients) * 100);
        return <Progress percent={percent} size="small" />;
      },
    },
    {
      title: 'Agendada',
      dataIndex: 'scheduled_at',
      key: 'scheduled_at',
      width: 130,
      render: (date) => date ? dayjs(date).format('DD/MM HH:mm') : '-',
    },
    {
      title: 'Criado',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 100,
      render: (date) => dayjs(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 120,
      align: 'center',
      render: (_, record) => (
        <Space>
          {record.status === 'draft' && (
            <Button
              type="text"
              icon={<SendOutlined />}
              onClick={() => handleSend(record.id)}
              title="Enviar"
              style={{ color: actionColors.save }}
            />
          )}
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/campaigns/${record.id}/edit`)}
            title="Editar"
            style={{ color: actionColors.edit }}
          />
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
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <NotificationOutlined />
            <span>Campanhas de Marketing</span>
          </Space>
        }
        extra={
          <Button
            icon={<PlusOutlined />}
            onClick={() => navigate('/campaigns/new')}
            style={{
              backgroundColor: actionColors.create,
              borderColor: actionColors.create,
              color: '#fff'
            }}
          >
            Nova Campanha
          </Button>
        }
      >
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Tipo"
              style={{ width: '100%' }}
              allowClear
              value={filters.type}
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              {typeOptions.map((type) => (
                <Select.Option key={type.value} value={type.value}>
                  {type.label}
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
            <Button onClick={fetchCampaigns} loading={loading}>
              Atualizar
            </Button>
          </Col>
        </Row>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={campaigns}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={setPagination}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>
    </div>
  );
};

export default Campaigns;
