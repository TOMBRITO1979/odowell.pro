import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Input, Space, Popconfirm, message, Tag, Upload, Modal, Card } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SearchOutlined, FileExcelOutlined, FilePdfOutlined, UploadOutlined } from '@ant-design/icons';
import { patientsAPI } from '../../services/api';
import { usePermission } from '../../contexts/AuthContext';
import { actionColors, statusColors, spacing } from '../../theme/designSystem';
import dayjs from 'dayjs';

const Patients = () => {
  const [patients, setPatients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [search, setSearch] = useState('');
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const fileInputRef = useRef(null);
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  // Detectar mudança de tamanho da tela
  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    let mounted = true;

    const loadPatients = async () => {
      setLoading(true);
      try {
        const response = await patientsAPI.getAll({
          page: pagination.current,
          page_size: pagination.pageSize,
          search,
        });
        if (mounted) {
          setPatients(response.data.patients || []);
          setPagination(prev => ({ ...prev, total: response.data.total }));
        }
      } catch (error) {
        if (mounted) {
          message.error('Erro ao carregar pacientes');
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    };

    loadPatients();

    return () => {
      mounted = false;
    };
  }, [pagination.current, pagination.pageSize, search]);

  const loadPatients = async () => {
    setLoading(true);
    try {
      const response = await patientsAPI.getAll({
        page: pagination.current,
        page_size: pagination.pageSize,
        search,
      });
      setPatients(response.data.patients || []);
      setPagination(prev => ({ ...prev, total: response.data.total }));
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await patientsAPI.delete(id);
      message.success('Paciente excluído com sucesso');
      loadPatients();
    } catch (error) {
      // Verificar se há dependências que impedem a exclusão
      if (error.response?.data?.dependencies) {
        const deps = error.response.data.dependencies;
        const depsList = Object.entries(deps)
          .map(([key, value]) => `${value} ${key}`)
          .join(', ');

        Modal.warning({
          title: 'Não é possível excluir este paciente',
          content: (
            <div>
              <p>Este paciente possui registros relacionados que precisam ser excluídos primeiro:</p>
              <ul style={{ marginTop: 8 }}>
                {Object.entries(deps).map(([key, value]) => (
                  <li key={key}><strong>{value}</strong> {key}</li>
                ))}
              </ul>
              <p style={{ marginTop: 12, color: '#666' }}>
                Exclua ou transfira esses registros antes de excluir o paciente.
              </p>
            </div>
          ),
          width: 500,
        });
      } else {
        message.error(error.response?.data?.error || 'Erro ao excluir paciente');
      }
    }
  };

  const handleExportCSV = async () => {
    try {
      const queryString = search ? `search=${encodeURIComponent(search)}` : '';
      const response = await patientsAPI.exportCSV(queryString);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `pacientes_${dayjs().format('YYYYMMDD_HHmmss')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      message.success('CSV exportado com sucesso');
    } catch (error) {
      message.error('Erro ao exportar CSV');
    }
  };

  const handleExportPDF = async () => {
    try {
      const params = {
        search,
        page: pagination.current,
        page_size: pagination.pageSize,
      };

      const cleanFilters = Object.fromEntries(
        Object.entries(params).filter(([_, value]) => value !== undefined && value !== null && value !== '')
      );
      const queryString = new URLSearchParams(cleanFilters).toString();
      const response = await patientsAPI.exportPDF(queryString);

      // Create blob and download
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `pacientes_lista_${dayjs().format('YYYYMMDD_HHmmss')}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      message.success('PDF gerado com sucesso');
    } catch (error) {
      message.error('Erro ao gerar PDF');
    }
  };

  const handleImportCSV = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const response = await patientsAPI.importCSV(formData);
      message.success(response.data.message);

      if (response.data.errors && response.data.errors.length > 0) {
        Modal.warning({
          title: 'Avisos durante a importação',
          content: (
            <div>
              <p>{response.data.imported} pacientes importados com sucesso.</p>
              <p>Erros encontrados:</p>
              <ul>
                {response.data.errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            </div>
          ),
          width: 600,
        });
      }

      setUploadModalVisible(false);
      loadPatients();
    } catch (error) {
      message.error('Erro ao importar CSV');
    } finally {
      setUploading(false);
    }

    return false; // Prevent default upload behavior
  };

  // Renderizar cards para versão mobile
  const renderMobileCards = () => {
    if (loading) {
      return <div style={{ textAlign: 'center', padding: '40px' }}>Carregando...</div>;
    }
    if (patients.length === 0) {
      return <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>Nenhum paciente encontrado</div>;
    }
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {patients.map((record) => (
          <Card
            key={record.id}
            size="small"
            style={{
              borderLeft: `4px solid ${record.active ? statusColors.success : statusColors.cancelled}`,
            }}
            bodyStyle={{ padding: '12px' }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
              <div style={{ fontWeight: 600, fontSize: '15px', flex: 1 }}>{record.name}</div>
              <Tag color={record.active ? statusColors.success : statusColors.cancelled}>
                {record.active ? 'Ativo' : 'Inativo'}
              </Tag>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px', fontSize: '13px', color: '#555' }}>
              <div><strong>CPF:</strong><br />{record.cpf || '-'}</div>
              <div><strong>Telefone:</strong><br />{record.cell_phone || record.phone || '-'}</div>
              <div style={{ gridColumn: '1 / -1' }}><strong>Email:</strong> {record.email || '-'}</div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '8px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/patients/${record.id}`)} style={{ color: actionColors.view }}>Ver</Button>
              {canEdit('patients') && <Button type="text" size="small" icon={<EditOutlined />} onClick={() => navigate(`/patients/${record.id}/edit`)} style={{ color: actionColors.edit }}>Editar</Button>}
              {canDelete('patients') && (
                <Popconfirm title="Tem certeza que deseja excluir?" onConfirm={() => handleDelete(record.id)} okText="Sim" cancelText="Não">
                  <Button type="text" size="small" icon={<DeleteOutlined />} style={{ color: actionColors.delete }}>Excluir</Button>
                </Popconfirm>
              )}
            </div>
          </Card>
        ))}
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '16px', marginTop: '16px', padding: '12px', background: '#fafafa', borderRadius: '8px' }}>
          <Button disabled={pagination.current === 1} onClick={() => setPagination(prev => ({ ...prev, current: prev.current - 1 }))}>Anterior</Button>
          <span style={{ fontSize: '13px' }}>Página {pagination.current} de {Math.ceil(pagination.total / pagination.pageSize) || 1}</span>
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
    },
    {
      title: 'CPF',
      dataIndex: 'cpf',
      key: 'cpf',
    },
    {
      title: 'Telefone',
      dataIndex: 'cell_phone',
      key: 'cell_phone',
      render: (text, record) => text || record.phone,
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      render: (active) => (
        <Tag color={active ? statusColors.success : statusColors.cancelled}>
          {active ? 'Ativo' : 'Inativo'}
        </Tag>
      ),
    },
    {
      title: 'Ações',
      key: 'actions',
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            icon={<EyeOutlined />}
            onClick={() => navigate(`/patients/${record.id}`)}
            size="small"
            style={{ color: actionColors.view }}
          />
          {canEdit('patients') && (
            <Button
              icon={<EditOutlined />}
              onClick={() => navigate(`/patients/${record.id}/edit`)}
              size="small"
              style={{ color: actionColors.edit }}
            />
          )}
          {canDelete('patients') && (
            <Popconfirm
              title="Tem certeza que deseja excluir?"
              onConfirm={() => handleDelete(record.id)}
              okText="Sim"
              cancelText="Não"
            >
              <Button
                icon={<DeleteOutlined />}
                size="small"
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
      <div style={{ marginBottom: 16 }}>
        <h1 style={{ marginBottom: 12 }}>Pacientes</h1>

        {/* All 4 buttons in same row */}
        <div className="patients-export-row">
          <Button
            icon={<FileExcelOutlined />}
            onClick={handleExportCSV}
            title="Exportar todos os pacientes para CSV"
            style={{
              backgroundColor: actionColors.exportExcel,
              borderColor: actionColors.exportExcel,
              color: '#fff'
            }}
            className="patients-export-btn"
          >
            <span className="btn-text-desktop">Exportar CSV</span>
            <span className="btn-text-mobile">CSV</span>
          </Button>
          <Button
            icon={<FilePdfOutlined />}
            onClick={handleExportPDF}
            title="Gerar PDF da página atual"
            style={{
              backgroundColor: actionColors.exportPDF,
              borderColor: actionColors.exportPDF,
              color: '#fff'
            }}
            className="patients-export-btn"
          >
            <span className="btn-text-desktop">Gerar PDF</span>
            <span className="btn-text-mobile">PDF</span>
          </Button>
          {canCreate('patients') && (
            <Button
              icon={<UploadOutlined />}
              onClick={() => setUploadModalVisible(true)}
              title="Importar pacientes via CSV"
              style={{
                backgroundColor: actionColors.view,
                borderColor: actionColors.view,
                color: '#fff'
              }}
              className="patients-export-btn"
            >
              <span className="btn-text-desktop">Importar CSV</span>
              <span className="btn-text-mobile">Importar</span>
            </Button>
          )}
          {canCreate('patients') && (
            <Button
              icon={<PlusOutlined />}
              onClick={() => navigate('/patients/new')}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
              className="patients-export-btn"
            >
              <span className="btn-text-desktop">Novo Paciente</span>
              <span className="btn-text-mobile">Novo</span>
            </Button>
          )}
        </div>
      </div>

      <Input
        placeholder="Buscar por nome, CPF, email ou telefone"
        prefix={<SearchOutlined />}
        onChange={(e) => setSearch(e.target.value)}
        style={{ marginBottom: 16, maxWidth: 400 }}
      />

      {isMobile ? (
        renderMobileCards()
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={patients}
            rowKey="id"
            loading={loading}
            pagination={pagination}
            onChange={(newPagination) => setPagination(newPagination)}
            scroll={{ x: 'max-content' }}
          />
        </div>
      )}

      <Modal
        title="Importar Pacientes via CSV"
        open={uploadModalVisible}
        onCancel={() => setUploadModalVisible(false)}
        footer={null}
      >
        <div style={{ marginBottom: 16 }}>
          <p><strong>Formato do CSV:</strong></p>
          <p>O arquivo deve conter as seguintes colunas (COM cabeçalho):</p>
          <ol>
            <li>Nome (obrigatório)</li>
            <li>CPF (obrigatório)</li>
            <li>Celular (obrigatório)</li>
            <li>RG</li>
            <li>Data Nascimento (formato: AAAA-MM-DD)</li>
            <li>Gênero (M/F/Other)</li>
            <li>Email</li>
            <li>Telefone</li>
            <li>Endereço</li>
            <li>Número</li>
            <li>Complemento</li>
            <li>Bairro</li>
            <li>Cidade</li>
            <li>Estado</li>
            <li>CEP</li>
            <li>Alergias</li>
            <li>Medicamentos</li>
            <li>Doenças Sistêmicas</li>
            <li>Observações</li>
          </ol>
          <p><strong>Exemplo:</strong></p>
          <code>Nome,CPF,Celular,RG,Data Nascimento,Gênero,Email,Telefone,Endereço,Número,Complemento,Bairro,Cidade,Estado,CEP,Alergias,Medicamentos,Doenças Sistêmicas,Observações</code>
          <br />
          <code>João Silva,12345678900,(11)98765-4321,123456789,1990-01-15,M,joao@email.com,(11)3333-4444,Rua A,100,Apto 10,Centro,São Paulo,SP,01234-567,Nenhuma,Nenhum,Nenhuma,</code>
        </div>

        <Upload
          accept=".csv"
          beforeUpload={handleImportCSV}
          showUploadList={false}
        >
          <Button
            icon={<UploadOutlined />}
            loading={uploading}
            block
            type="primary"
          >
            {uploading ? 'Importando...' : 'Selecionar arquivo CSV'}
          </Button>
        </Upload>
      </Modal>
    </div>
  );
};

export default Patients;
