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
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const [filters, setFilters] = useState({
    status: undefined,
    type: undefined,
  });

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

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

  const renderMobileCards = () => {
    if (loading) return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    if (campaigns.length === 0) return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhuma campanha encontrada</div>;
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {campaigns.map((record) => {
          const statusObj = statusOptions.find(s => s.value === record.status);
          return (
            <Card key={record.id} size="small" style={{ borderLeft: `4px solid ${statusObj?.color === 'success' ? '#52c41a' : statusObj?.color === 'error' ? '#ff4d4f' : statusObj?.color === 'warning' ? '#faad14' : '#1890ff'}` }} bodyStyle={{ padding: '12px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.name}</div>
                {getStatusTag(record.status)}
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
                <div><strong>Tipo:</strong><br />{getTypeTag(record.type)}</div>
                <div><strong>Destinatários:</strong><br />{record.total_recipients || 0}</div>
                <div><strong>Enviados:</strong><br />{record.sent || 0} / {record.total_recipients || 0}</div>
                <div><strong>Agendada:</strong><br />{record.scheduled_at ? (() => { const m = record.scheduled_at.match(/(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/); return m ? `${m[3]}/${m[2]} ${m[4]}:${m[5]}` : '-'; })() : '-'}</div>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
                {record.status === 'draft' && <Button type="text" size="small" icon={<SendOutlined />} onClick={() => handleSend(record.id)} style={{ color: actionColors.save }}>Enviar</Button>}
                <Button type="text" size="small" icon={<EditOutlined />} onClick={() => navigate(`/campaigns/${record.id}/edit`)} style={{ color: actionColors.edit }}>Editar</Button>
                <Popconfirm title="Tem certeza?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                  <Button type="text" size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
                </Popconfirm>
              </div>
            </Card>
          );
        })}
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '16px', marginTop: '16px', padding: '12px', background: '#fafafa', borderRadius: '8px' }}>
          <Button disabled={pagination.current === 1} onClick={() => setPagination(prev => ({ ...prev, current: prev.current - 1 }))}>Anterior</Button>
          <span style={{ fontSize: '13px' }}>Pág. {pagination.current} de {Math.ceil(pagination.total / pagination.pageSize) || 1}</span>
          <Button disabled={pagination.current >= Math.ceil(pagination.total / pagination.pageSize)} onClick={() => setPagination(prev => ({ ...prev, current: prev.current + 1 }))}>Próxima</Button>
        </div>
      </div>
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
      // Extrair horário da string ISO sem conversão de timezone
      render: (date) => {
        if (!date) return '-';
        // A data vem como "2025-12-25T15:25:00-03:00"
        // Extrair diretamente: dia/mês hora:min
        const match = date.match(/(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/);
        if (match) {
          return `${match[3]}/${match[2]} ${match[4]}:${match[5]}`;
        }
        return '-';
      },
    },
    {
      title: 'Criado',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 100,
      render: (date) => {
        if (!date) return '-';
        const match = date.match(/(\d{4})-(\d{2})-(\d{2})/);
        if (match) {
          return `${match[3]}/${match[2]}/${match[1]}`;
        }
        return '-';
      },
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

        {isMobile ? renderMobileCards() : (
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
        )}
      </Card>
    </div>
  );
};

export default Campaigns;
