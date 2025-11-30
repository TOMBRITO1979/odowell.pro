import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, message, Tag, Space, Popconfirm, Card, Row, Col, Select, DatePicker } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, FileExcelOutlined, FilePdfOutlined, FilterOutlined, ClearOutlined } from '@ant-design/icons';
import { appointmentsAPI, usersAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, spacing, shadows, buttonSizes } from '../../theme/designSystem';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

const Appointments = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dentists, setDentists] = useState([]);
  const [filters, setFilters] = useState({
    dentist_id: null,
    procedure: null,
    status: null,
    dateRange: null,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  useEffect(() => {
    fetchDentists();
  }, []);

  useEffect(() => {
    fetchAppointments();
  }, [pagination.current, filters]);

  const fetchDentists = async () => {
    try {
      const response = await usersAPI.getAll();
      // Filtrar apenas dentistas e admins (profissionais que atendem)
      const professionals = (response.data.users || []).filter(
        user => user.role === 'dentist' || user.role === 'admin'
      );
      setDentists(professionals);
    } catch (error) {
      console.error('Error fetching dentists:', error);
    }
  };

  const fetchAppointments = () => {
    setLoading(true);
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize
    };

    // Adicionar filtros se definidos
    if (filters.dentist_id) params.dentist_id = filters.dentist_id;
    if (filters.procedure) params.procedure = filters.procedure;
    if (filters.status) params.status = filters.status;
    if (filters.dateRange && filters.dateRange[0]) {
      params.start_date = filters.dateRange[0].startOf('day').toISOString();
    }
    if (filters.dateRange && filters.dateRange[1]) {
      params.end_date = filters.dateRange[1].endOf('day').toISOString();
    }

    appointmentsAPI.getAll(params)
      .then(res => {
        setData(res.data.appointments || []);
        setPagination({
          ...pagination,
          total: res.data.total || 0,
        });
      })
      .catch(() => message.error('Erro ao carregar agendamentos'))
      .finally(() => setLoading(false));
  };

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  const clearFilters = () => {
    setFilters({
      dentist_id: null,
      procedure: null,
      status: null,
      dateRange: null,
    });
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  const procedureOptions = [
    { value: 'consultation', label: 'Consulta' },
    { value: 'cleaning', label: 'Limpeza' },
    { value: 'filling', label: 'Restauração' },
    { value: 'extraction', label: 'Extração' },
    { value: 'root_canal', label: 'Canal' },
    { value: 'orthodontics', label: 'Ortodontia' },
    { value: 'whitening', label: 'Clareamento' },
    { value: 'prosthesis', label: 'Prótese' },
    { value: 'implant', label: 'Implante' },
    { value: 'emergency', label: 'Emergência' },
    { value: 'other', label: 'Outro' },
  ];

  const statusOptions = [
    { value: 'scheduled', label: 'Agendado' },
    { value: 'confirmed', label: 'Confirmado' },
    { value: 'in_progress', label: 'Em Atendimento' },
    { value: 'completed', label: 'Concluído' },
    { value: 'cancelled', label: 'Cancelado' },
    { value: 'no_show', label: 'Faltou' },
  ];

  const handleDelete = async (id) => {
    try {
      await appointmentsAPI.delete(id);
      message.success('Agendamento deletado com sucesso');
      fetchAppointments();
    } catch (error) {
      message.error('Erro ao deletar agendamento');
    }
  };

  const handleExportCSV = async () => {
    try {
      const response = await appointmentsAPI.exportCSV('');
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `agendamentos_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
      console.error('Export error:', error);
    }
  };

  const handleExportPDF = async () => {
    try {
      const response = await appointmentsAPI.exportPDF('');
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `agendamentos_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
      console.error('PDF error:', error);
    }
  };

  const getStatusTag = (status) => {
    const statusConfig = {
      scheduled: { color: statusColors.pending, text: 'Agendado' },
      confirmed: { color: statusColors.approved, text: 'Confirmado' },
      in_progress: { color: statusColors.inProgress, text: 'Em Atendimento' },
      completed: { color: statusColors.success, text: 'Concluído' },
      cancelled: { color: statusColors.cancelled, text: 'Cancelado' },
      no_show: { color: statusColors.error, text: 'Faltou' },
    };
    const config = statusConfig[status] || { color: statusColors.pending, text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getProcedureText = (procedure) => {
    const procedures = {
      consultation: 'Consulta',
      cleaning: 'Limpeza',
      filling: 'Restauração',
      extraction: 'Extração',
      root_canal: 'Canal',
      orthodontics: 'Ortodontia',
      whitening: 'Clareamento',
      prosthesis: 'Prótese',
      implant: 'Implante',
      emergency: 'Emergência',
      other: 'Outro',
    };
    return procedures[procedure] || procedure;
  };

  const columns = [
    {
      title: 'Paciente',
      dataIndex: ['patient', 'name'],
      key: 'patient_name',
      render: (text) => text || 'N/A',
    },
    {
      title: 'Profissional',
      dataIndex: ['dentist', 'name'],
      key: 'dentist_name',
      render: (text) => text || 'N/A',
      responsive: ['md', 'lg', 'xl', 'xxl'], // Esconde em mobile
    },
    {
      title: 'Data/Hora',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (text) => text ? dayjs(text).format('DD/MM/YYYY HH:mm') : 'N/A',
    },
    {
      title: 'Procedimento',
      dataIndex: 'procedure',
      key: 'procedure',
      render: (text) => getProcedureText(text),
      responsive: ['sm', 'md', 'lg', 'xl', 'xxl'], // Esconde em xs
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Sala',
      dataIndex: 'room',
      key: 'room',
      render: (text) => text || '-',
      responsive: ['lg', 'xl', 'xxl'], // Esconde em tablet e mobile
    },
    {
      title: 'Ações',
      key: 'actions',
      fixed: 'right',
      width: 100,
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/appointments/${record.id}`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          {canEdit('appointments') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/appointments/${record.id}/edit`)}
              title="Editar"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('appointments') && (
            <Popconfirm
              title="Tem certeza que deseja deletar este agendamento?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                type="text"
                icon={<DeleteOutlined />}
                title="Deletar"
                style={{ color: actionColors.delete }}
              />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="Agendamentos"
        style={{
          boxShadow: shadows.small,
        }}
        extra={
          <div className="appointments-button-group">
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              style={{
                backgroundColor: actionColors.exportExcel,
                borderColor: actionColors.exportExcel,
                color: '#fff'
              }}
              className="appointments-btn"
            >
              <span className="btn-text-desktop">Exportar CSV</span>
              <span className="btn-text-mobile">CSV</span>
            </Button>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleExportPDF}
              style={{
                backgroundColor: actionColors.exportPDF,
                borderColor: actionColors.exportPDF,
                color: '#fff'
              }}
              className="appointments-btn"
            >
              <span className="btn-text-desktop">Gerar PDF</span>
              <span className="btn-text-mobile">PDF</span>
            </Button>
            {canCreate('appointments') && (
              <Button
                icon={<PlusOutlined />}
                onClick={() => navigate('/appointments/new')}
                style={{
                  backgroundColor: actionColors.create,
                  borderColor: actionColors.create,
                  color: '#fff'
                }}
                className="appointments-btn"
              >
                <span className="btn-text-desktop">Novo Agendamento</span>
                <span className="btn-text-mobile">Novo</span>
              </Button>
            )}
          </div>
        }
      >
        {/* Filtros */}
        <div style={{ marginBottom: 16, padding: '16px', background: '#fafafa', borderRadius: '8px' }}>
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} sm={12} md={6}>
              <Select
                placeholder="Profissional"
                allowClear
                style={{ width: '100%' }}
                value={filters.dentist_id}
                onChange={(value) => handleFilterChange('dentist_id', value)}
                showSearch
                filterOption={(input, option) =>
                  option.children.toLowerCase().includes(input.toLowerCase())
                }
              >
                {dentists.map(d => (
                  <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>
                ))}
              </Select>
            </Col>
            <Col xs={24} sm={12} md={5}>
              <Select
                placeholder="Procedimento"
                allowClear
                style={{ width: '100%' }}
                value={filters.procedure}
                onChange={(value) => handleFilterChange('procedure', value)}
              >
                {procedureOptions.map(p => (
                  <Select.Option key={p.value} value={p.value}>{p.label}</Select.Option>
                ))}
              </Select>
            </Col>
            <Col xs={24} sm={12} md={5}>
              <Select
                placeholder="Status"
                allowClear
                style={{ width: '100%' }}
                value={filters.status}
                onChange={(value) => handleFilterChange('status', value)}
              >
                {statusOptions.map(s => (
                  <Select.Option key={s.value} value={s.value}>{s.label}</Select.Option>
                ))}
              </Select>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <RangePicker
                style={{ width: '100%' }}
                format="DD/MM/YYYY"
                value={filters.dateRange}
                onChange={(dates) => handleFilterChange('dateRange', dates)}
                placeholder={['Data Início', 'Data Fim']}
              />
            </Col>
            <Col xs={24} sm={24} md={2}>
              <Button
                icon={<ClearOutlined />}
                onClick={clearFilters}
                style={{ width: '100%' }}
                title="Limpar Filtros"
              >
                Limpar
              </Button>
            </Col>
          </Row>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={data}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              onChange: (page) => setPagination({ ...pagination, current: page }),
            }}
            scroll={{ x: 'max-content' }}
          />
        </div>
      </Card>
    </div>
  );
};

export default Appointments;
