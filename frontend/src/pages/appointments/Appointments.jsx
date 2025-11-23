import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, message, Tag, Space, Popconfirm, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, FileExcelOutlined, FilePdfOutlined } from '@ant-design/icons';
import { appointmentsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import dayjs from 'dayjs';

const Appointments = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  useEffect(() => {
    fetchAppointments();
  }, [pagination.current]);

  const fetchAppointments = () => {
    setLoading(true);
    appointmentsAPI.getAll({ page: pagination.current, page_size: pagination.pageSize })
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
      scheduled: { color: 'blue', text: 'Agendado' },
      confirmed: { color: 'green', text: 'Confirmado' },
      in_progress: { color: 'orange', text: 'Em Atendimento' },
      completed: { color: 'success', text: 'Concluído' },
      cancelled: { color: 'red', text: 'Cancelado' },
      no_show: { color: 'default', text: 'Faltou' },
    };
    const config = statusConfig[status] || { color: 'default', text: status };
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
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => getStatusTag(status),
    },
    {
      title: 'Ações',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/appointments/${record.id}`)}
            title="Visualizar"
          />
          {canEdit('appointments') && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/appointments/${record.id}/edit`)}
              title="Editar"
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
                danger
                icon={<DeleteOutlined />}
                title="Deletar"
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
        extra={
          <div className="appointments-button-group">
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleExportCSV}
              style={{ backgroundColor: '#22c55e', borderColor: '#22c55e', color: '#fff' }}
              className="appointments-btn"
            >
              <span className="btn-text-desktop">Exportar CSV</span>
              <span className="btn-text-mobile">CSV</span>
            </Button>
            <Button
              icon={<FilePdfOutlined />}
              onClick={handleExportPDF}
              style={{ backgroundColor: '#ef4444', borderColor: '#ef4444', color: '#fff' }}
              className="appointments-btn"
            >
              <span className="btn-text-desktop">Gerar PDF</span>
              <span className="btn-text-mobile">PDF</span>
            </Button>
            {canCreate('appointments') && (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate('/appointments/new')}
                className="appointments-btn"
              >
                <span className="btn-text-desktop">Novo Agendamento</span>
                <span className="btn-text-mobile">Novo</span>
              </Button>
            )}
          </div>
        }
      >
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
