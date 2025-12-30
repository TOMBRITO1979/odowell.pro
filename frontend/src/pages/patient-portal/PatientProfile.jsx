import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Typography,
  Spin,
  Tag,
  Divider,
  Row,
  Col,
  message,
} from 'antd';
import {
  UserOutlined,
  PhoneOutlined,
  MailOutlined,
  HomeOutlined,
  MedicineBoxOutlined,
  IdcardOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { patientPortalAPI } from '../../services/api';

const { Title, Text } = Typography;

const PatientProfile = () => {
  const [loading, setLoading] = useState(true);
  const [patient, setPatient] = useState(null);

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    try {
      const response = await patientPortalAPI.getProfile();
      setPatient(response.data.patient);
    } catch (error) {
      message.error('Erro ao carregar dados do perfil');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!patient) {
    return (
      <Card>
        <Text type="secondary">Dados do paciente nao encontrados</Text>
      </Card>
    );
  }

  const formatCPF = (cpf) => {
    if (!cpf) return 'Nao informado';
    const cleaned = cpf.replace(/\D/g, '');
    if (cleaned.length === 11) {
      return `${cleaned.slice(0, 3)}.${cleaned.slice(3, 6)}.${cleaned.slice(6, 9)}-${cleaned.slice(9)}`;
    }
    return cpf;
  };

  const formatPhone = (phone) => {
    if (!phone) return 'Nao informado';
    const cleaned = phone.replace(/\D/g, '');
    if (cleaned.length === 11) {
      return `(${cleaned.slice(0, 2)}) ${cleaned.slice(2, 7)}-${cleaned.slice(7)}`;
    }
    if (cleaned.length === 10) {
      return `(${cleaned.slice(0, 2)}) ${cleaned.slice(2, 6)}-${cleaned.slice(6)}`;
    }
    return phone;
  };

  return (
    <div>
      <Title level={4}>
        <UserOutlined /> Meus Dados
      </Title>

      <Row gutter={[24, 24]}>
        {/* Personal Info */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <>
                <IdcardOutlined /> Dados Pessoais
              </>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Nome">{patient.name}</Descriptions.Item>
              <Descriptions.Item label="CPF">{formatCPF(patient.cpf)}</Descriptions.Item>
              <Descriptions.Item label="RG">{patient.rg || 'Nao informado'}</Descriptions.Item>
              <Descriptions.Item label="Data de Nascimento">
                {patient.birth_date
                  ? dayjs(patient.birth_date).format('DD/MM/YYYY')
                  : 'Nao informada'}
              </Descriptions.Item>
              <Descriptions.Item label="Genero">
                {patient.gender === 'M' ? 'Masculino' : patient.gender === 'F' ? 'Feminino' : patient.gender || 'Nao informado'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Contact Info */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <>
                <PhoneOutlined /> Contato
              </>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Email">
                <MailOutlined style={{ marginRight: 8 }} />
                {patient.email || 'Nao informado'}
              </Descriptions.Item>
              <Descriptions.Item label="Telefone">
                <PhoneOutlined style={{ marginRight: 8 }} />
                {formatPhone(patient.phone)}
              </Descriptions.Item>
              <Descriptions.Item label="Celular">
                <PhoneOutlined style={{ marginRight: 8 }} />
                {formatPhone(patient.cell_phone)}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Address */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <>
                <HomeOutlined /> Endereco
              </>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Endereco">
                {patient.address
                  ? `${patient.address}${patient.number ? `, ${patient.number}` : ''}${patient.complement ? ` - ${patient.complement}` : ''}`
                  : 'Nao informado'}
              </Descriptions.Item>
              <Descriptions.Item label="Bairro">
                {patient.district || 'Nao informado'}
              </Descriptions.Item>
              <Descriptions.Item label="Cidade/Estado">
                {patient.city && patient.state
                  ? `${patient.city} - ${patient.state}`
                  : patient.city || patient.state || 'Nao informado'}
              </Descriptions.Item>
              <Descriptions.Item label="CEP">
                {patient.zip_code || 'Nao informado'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Medical Info */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <>
                <MedicineBoxOutlined /> Informacoes Medicas
              </>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Tipo Sanguineo">
                {patient.blood_type ? (
                  <Tag color="red">{patient.blood_type}</Tag>
                ) : (
                  'Nao informado'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Alergias">
                {patient.allergies || 'Nenhuma informada'}
              </Descriptions.Item>
              <Descriptions.Item label="Medicamentos em Uso">
                {patient.medications || 'Nenhum informado'}
              </Descriptions.Item>
              <Descriptions.Item label="Doencas Sistemicas">
                {patient.systemic_diseases || 'Nenhuma informada'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Insurance */}
        {patient.has_insurance && (
          <Col xs={24}>
            <Card title="Convenio">
              <Descriptions column={{ xs: 1, sm: 2 }} size="small">
                <Descriptions.Item label="Convenio">
                  {patient.insurance_name || 'Nao informado'}
                </Descriptions.Item>
                <Descriptions.Item label="Numero da Carteirinha">
                  {patient.insurance_number || 'Nao informado'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
        )}
      </Row>

      <Divider />

      <Text type="secondary">
        Para atualizar seus dados, entre em contato com a clinica.
      </Text>
    </div>
  );
};

export default PatientProfile;
