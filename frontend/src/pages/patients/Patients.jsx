import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Input, Space, Popconfirm, message, Tag, Upload, Modal } from 'antd';
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
  const fileInputRef = useRef(null);
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission();

  useEffect(() => {
    loadPatients();
  }, [pagination.current, search]);

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
      message.error('Erro ao excluir paciente');
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
      console.error('Export error:', error);
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
      console.error('PDF error:', error);
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
      console.error('Import error:', error);
    } finally {
      setUploading(false);
    }

    return false; // Prevent default upload behavior
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
