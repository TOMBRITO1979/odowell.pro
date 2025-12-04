import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  message,
  Descriptions,
  Space,
  Button,
  Tag,
  Table,
  Divider,
  Alert,
  Result,
  Spin,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DollarOutlined,
  FilePdfOutlined,
  MedicineBoxOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { budgetsAPI, treatmentsAPI } from '../../services/api';
import { actionColors } from '../../theme/designSystem';

const BudgetView = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [budget, setBudget] = useState(null);
  const [items, setItems] = useState([]);
  const [treatment, setTreatment] = useState(null);
  const [loadingTreatment, setLoadingTreatment] = useState(false);

  const statusOptions = {
    pending: { label: 'Pendente', color: 'warning' },
    approved: { label: 'Aprovado', color: 'success' },
    rejected: { label: 'Rejeitado', color: 'error' },
    expired: { label: 'Expirado', color: 'default' },
    cancelled: { label: 'Cancelado', color: 'default' },
  };

  useEffect(() => {
    fetchBudget();
  }, [id]);

  const fetchBudget = async () => {
    setLoading(true);
    try {
      const response = await budgetsAPI.getOne(id);
      const budgetData = response.data.budget;
      setBudget(budgetData);

      // Parse items from JSON string
      if (budgetData.items) {
        try {
          const parsedItems = JSON.parse(budgetData.items);
          setItems(parsedItems);
        } catch (e) {
          console.error('Error parsing items:', e);
        }
      }

      // If budget is approved, fetch the associated treatment
      if (budgetData.status === 'approved') {
        fetchTreatment(budgetData.id);
      }
    } catch (error) {
      message.error('Erro ao carregar orçamento');
    } finally {
      setLoading(false);
    }
  };

  const fetchTreatment = async (budgetId) => {
    setLoadingTreatment(true);
    try {
      const response = await treatmentsAPI.getByBudgetId(budgetId);
      if (response.data.treatments && response.data.treatments.length > 0) {
        setTreatment(response.data.treatments[0]);
      }
    } catch (error) {
      console.error('Error fetching treatment:', error);
    } finally {
      setLoadingTreatment(false);
    }
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  };

  const handleDownloadPDF = async () => {
    try {
      const response = await budgetsAPI.downloadPDF(id);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `orcamento_${id}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      message.success('PDF baixado com sucesso');
    } catch (error) {
      message.error('Erro ao baixar PDF');
      console.error('Error:', error);
    }
  };

  const itemColumns = [
    {
      title: 'Descrição',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Quantidade',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 120,
      align: 'center',
    },
    {
      title: 'Valor Unitário',
      dataIndex: 'unit_price',
      key: 'unit_price',
      width: 150,
      align: 'right',
      render: (value) => formatCurrency(value),
    },
    {
      title: 'Total',
      dataIndex: 'total',
      key: 'total',
      width: 150,
      align: 'right',
      render: (value) => formatCurrency(value),
    },
  ];

  if (!budget) {
    return (
      <Card loading={loading}>
        <p>Carregando...</p>
      </Card>
    );
  }

  const statusInfo = statusOptions[budget.status] || { label: budget.status, color: 'default' };

  return (
    <div>
      <Card
        title={
          <Space>
            <DollarOutlined />
            <span>Visualizar Orçamento</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleDownloadPDF}
              style={{ backgroundColor: actionColors.exportPDF, borderColor: actionColors.exportPDF, color: '#fff' }}
            >
              Baixar PDF
            </Button>
            <Button
              icon={<EditOutlined />}
              onClick={() => navigate(`/budgets/${id}/edit`)}
            >
              Editar
            </Button>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/budgets')}
            >
              Voltar
            </Button>
          </Space>
        }
        loading={loading}
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="ID">{budget.id}</Descriptions.Item>
          <Descriptions.Item label="Data de Criação">
            {dayjs(budget.created_at).format('DD/MM/YYYY HH:mm')}
          </Descriptions.Item>
          <Descriptions.Item label="Paciente">
            {budget.patient?.name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Profissional">
            {budget.dentist?.name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag color={statusInfo.color}>{statusInfo.label}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Válido Até">
            {budget.valid_until ? dayjs(budget.valid_until).format('DD/MM/YYYY') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Valor Total" span={2}>
            <span style={{ fontSize: 20, fontWeight: 'bold', color: '#64B5F6' }}>
              {formatCurrency(budget.total_value)}
            </span>
          </Descriptions.Item>
          {budget.description && (
            <Descriptions.Item label="Descrição" span={2}>
              {budget.description}
            </Descriptions.Item>
          )}
        </Descriptions>

        <Divider>Itens do Orçamento</Divider>

        <Table
          columns={itemColumns}
          dataSource={items}
          rowKey="id"
          pagination={false}
          locale={{ emptyText: 'Nenhum item' }}
          footer={() => (
            <div style={{ textAlign: 'right', fontSize: 18, fontWeight: 'bold' }}>
              Total: {formatCurrency(budget.total_value)}
            </div>
          )}
        />

        {budget.notes && (
          <>
            <Divider>Observações</Divider>
            <p>{budget.notes}</p>
          </>
        )}

        {/* Seção de Tratamento/Pagamentos */}
        {budget.status === 'approved' && (
          <>
            <Divider
              orientation="left"
              style={{
                fontSize: '18px',
                fontWeight: 'bold',
                color: '#66BB6A',
                marginTop: '32px',
                marginBottom: '24px',
                borderColor: '#66BB6A'
              }}
            >
              <Space>
                <MedicineBoxOutlined style={{ fontSize: '20px' }} />
                Tratamento
              </Space>
            </Divider>

            {loadingTreatment ? (
              <div style={{ textAlign: 'center', padding: '40px' }}>
                <Spin size="large" />
                <p style={{ marginTop: 16, color: '#666' }}>Carregando tratamento...</p>
              </div>
            ) : treatment ? (
              <Result
                status="success"
                icon={<MedicineBoxOutlined style={{ color: '#66BB6A' }} />}
                title="Tratamento em Andamento"
                subTitle={
                  <div>
                    <p style={{ marginBottom: 8 }}>
                      Este orçamento foi aprovado e possui um tratamento vinculado.
                    </p>
                    <p style={{ marginBottom: 8 }}>
                      <strong>Status:</strong>{' '}
                      <Tag color={treatment.status === 'completed' ? 'success' : treatment.status === 'cancelled' ? 'error' : 'processing'}>
                        {treatment.status === 'completed' ? 'Concluído' : treatment.status === 'cancelled' ? 'Cancelado' : 'Em Andamento'}
                      </Tag>
                    </p>
                    <p style={{ marginBottom: 8 }}>
                      <strong>Valor Total:</strong> {formatCurrency(treatment.total_value)} |{' '}
                      <strong>Pago:</strong> <span style={{ color: '#66BB6A' }}>{formatCurrency(treatment.paid_value)}</span> |{' '}
                      <strong>Restante:</strong> <span style={{ color: treatment.total_value - treatment.paid_value > 0 ? '#E57373' : '#66BB6A' }}>
                        {formatCurrency(treatment.total_value - treatment.paid_value)}
                      </span>
                    </p>
                  </div>
                }
                extra={[
                  <Button
                    key="view"
                    type="primary"
                    size="large"
                    icon={<MedicineBoxOutlined />}
                    onClick={() => navigate(`/treatments/${treatment.id}`)}
                    style={{ backgroundColor: '#66BB6A', borderColor: '#66BB6A' }}
                  >
                    Gerenciar Pagamentos do Tratamento
                  </Button>,
                ]}
              />
            ) : (
              <Alert
                type="warning"
                showIcon
                message="Tratamento não encontrado"
                description={
                  <div>
                    <p>Este orçamento está aprovado, mas não foi encontrado um tratamento vinculado.</p>
                    <p>Isso pode ter ocorrido por um erro no sistema. Por favor, entre em contato com o suporte ou tente aprovar o orçamento novamente.</p>
                  </div>
                }
              />
            )}
          </>
        )}

        {budget.status !== 'approved' && (
          <>
            <Divider
              orientation="left"
              style={{
                fontSize: '18px',
                fontWeight: 'bold',
                color: '#64B5F6',
                marginTop: '32px',
                marginBottom: '24px',
                borderColor: '#64B5F6'
              }}
            >
              <Space>
                <DollarOutlined style={{ fontSize: '20px' }} />
                Pagamentos
              </Space>
            </Divider>

            <Alert
              type="info"
              showIcon
              message="Orçamento aguardando aprovação"
              description={
                <div>
                  <p>Para gerenciar pagamentos, o orçamento precisa ser <strong>aprovado</strong> primeiro.</p>
                  <p>Quando aprovado, um <strong>Tratamento</strong> será criado automaticamente e você poderá registrar os pagamentos por lá.</p>
                  <p style={{ marginTop: 8 }}>
                    Status atual: <Tag color={statusOptions[budget.status]?.color}>{statusOptions[budget.status]?.label}</Tag>
                  </p>
                </div>
              }
            />
          </>
        )}
      </Card>
    </div>
  );
};

export default BudgetView;
