import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  message,
  Space,
  Spin,
  Tag,
  Row,
  Col,
} from 'antd';
import {
  EditOutlined,
  ArrowLeftOutlined,
  UserOutlined,
  PhoneOutlined,
  MailOutlined,
  HomeOutlined,
  MedicineBoxOutlined,
  SafetyOutlined,
} from '@ant-design/icons';
import { patientsAPI } from '../../services/api';
import dayjs from 'dayjs';
import { usePermission } from '../../contexts/AuthContext';
import PatientConsents from '../../components/consents/PatientConsents';

const PatientDetails = () => {
  const [patient, setPatient] = useState(null);
  const [loading, setLoading] = useState(true);
  const { id } = useParams();
  const navigate = useNavigate();
  const { canEdit } = usePermission();

  useEffect(() => {
    fetchPatient();
  }, [id]);

  const fetchPatient = async () => {
    setLoading(true);
    try {
      const response = await patientsAPI.getOne(id);
      setPatient(response.data.patient);
    } catch (error) {
      message.error('Erro ao carregar dados do paciente');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (date) => {
    if (!date) return '-';
    return dayjs(date).format('DD/MM/YYYY');
  };

  const formatGender = (gender) => {
    if (!gender) return '-';
    const genders = {
      M: 'Masculino',
      F: 'Feminino',
      Other: 'Outro',
    };
    return genders[gender] || gender;
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
        <p style={{ marginTop: 16 }}>Carregando dados do paciente...</p>
      </div>
    );
  }

  if (!patient) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <p>Paciente não encontrado</p>
        <Button onClick={() => navigate('/patients')}>Voltar</Button>
      </div>
    );
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/patients')}
        >
          Voltar
        </Button>
      </Space>

      <Card
        title={
          <Space>
            <UserOutlined />
            <span>{patient.name}</span>
            {patient.active ? (
              <Tag color="success">Ativo</Tag>
            ) : (
              <Tag color="error">Inativo</Tag>
            )}
          </Space>
        }
        extra={
          canEdit('patients') && (
            <Button
              type="primary"
              icon={<EditOutlined />}
              onClick={() => navigate(`/patients/${id}/edit`)}
            >
              Editar
            </Button>
          )
        }
      >
        {/* Informações Pessoais */}
        <Card
          type="inner"
          title={
            <Space>
              <UserOutlined />
              <span>Informações Pessoais</span>
            </Space>
          }
          style={{ marginBottom: 16 }}
        >
          <Descriptions bordered column={{ xs: 1, sm: 2, md: 3 }}>
            <Descriptions.Item label="Nome Completo" span={3}>
              {patient.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="CPF">
              {patient.cpf || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="RG">
              {patient.rg || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Data de Nascimento">
              {formatDate(patient.birth_date)}
            </Descriptions.Item>
            <Descriptions.Item label="Gênero">
              {formatGender(patient.gender)}
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              {patient.active ? (
                <Tag color="success">Ativo</Tag>
              ) : (
                <Tag color="error">Inativo</Tag>
              )}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Contato */}
        <Card
          type="inner"
          title={
            <Space>
              <PhoneOutlined />
              <span>Contato</span>
            </Space>
          }
          style={{ marginBottom: 16 }}
        >
          <Descriptions bordered column={{ xs: 1, sm: 2, md: 3 }}>
            <Descriptions.Item label="Email" span={2}>
              {patient.email ? (
                <a href={`mailto:${patient.email}`}>{patient.email}</a>
              ) : (
                '-'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="Telefone">
              {patient.phone || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Celular" span={2}>
              {patient.cell_phone || '-'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Endereço */}
        <Card
          type="inner"
          title={
            <Space>
              <HomeOutlined />
              <span>Endereço</span>
            </Space>
          }
          style={{ marginBottom: 16 }}
        >
          <Descriptions bordered column={{ xs: 1, sm: 2, md: 3 }}>
            <Descriptions.Item label="Endereço" span={2}>
              {patient.address || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Número">
              {patient.number || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Cidade">
              {patient.city || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Estado">
              {patient.state || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="CEP">
              {patient.zip_code || '-'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Informações Médicas */}
        <Card
          type="inner"
          title={
            <Space>
              <MedicineBoxOutlined />
              <span>Informações Médicas</span>
            </Space>
          }
          style={{ marginBottom: 16 }}
        >
          <Descriptions bordered column={1}>
            <Descriptions.Item label="Alergias">
              {patient.allergies || 'Nenhuma alergia registrada'}
            </Descriptions.Item>
            <Descriptions.Item label="Medicamentos em Uso">
              {patient.medications || 'Nenhum medicamento registrado'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Termos de Consentimento */}
        <div style={{ marginBottom: 16 }}>
          <PatientConsents patient={patient} />
        </div>

        {/* Convênio */}
        {(patient.insurance_name || patient.insurance_number) && (
          <Card
            type="inner"
            title={
              <Space>
                <SafetyOutlined />
                <span>Convênio</span>
              </Space>
            }
            style={{ marginBottom: 16 }}
          >
            <Descriptions bordered column={{ xs: 1, sm: 2 }}>
              <Descriptions.Item label="Nome do Convênio">
                {patient.insurance_name || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Número do Convênio">
                {patient.insurance_number || '-'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        )}

        {/* Tags para Campanhas */}
        {patient.tags && (
          <Card
            type="inner"
            title="Tags (Campanhas)"
            style={{ marginBottom: 16 }}
          >
            <Space size={[0, 8]} wrap>
              {patient.tags.split(',').map((tag, index) => (
                <Tag key={index} color="blue">
                  {tag.trim()}
                </Tag>
              ))}
            </Space>
          </Card>
        )}

        {/* Observações */}
        {patient.notes && (
          <Card type="inner" title="Observações" style={{ marginBottom: 16 }}>
            <p style={{ whiteSpace: 'pre-wrap' }}>{patient.notes}</p>
          </Card>
        )}

        {/* Datas de Registro */}
        <Card
          type="inner"
          title="Informações do Sistema"
          style={{ marginTop: 16 }}
        >
          <Descriptions bordered column={{ xs: 1, sm: 2 }}>
            <Descriptions.Item label="Data de Cadastro">
              {formatDate(patient.created_at)}
            </Descriptions.Item>
            <Descriptions.Item label="Última Atualização">
              {formatDate(patient.updated_at)}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </Card>
    </div>
  );
};

export default PatientDetails;
