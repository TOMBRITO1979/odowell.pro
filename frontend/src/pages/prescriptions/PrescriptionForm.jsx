import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  message,
  Select,
  Row,
  Col,
  Divider,
  Space,
  DatePicker,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { prescriptionsAPI, patientsAPI, usersAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

const { TextArea } = Input;

const PrescriptionForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const [professionals, setProfessionals] = useState([]);

  const prescriptionTypes = [
    { value: 'prescription', label: 'Receita' },
    { value: 'medical_report', label: 'Laudo Médico' },
    { value: 'certificate', label: 'Atestado' },
    { value: 'referral', label: 'Encaminhamento' },
  ];

  useEffect(() => {
    fetchPatients();
    fetchProfessionals();
    if (id) {
      fetchPrescription();
    }
  }, [id]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    }
  };

  const fetchProfessionals = async () => {
    try {
      const response = await usersAPI.getAll({ page: 1, page_size: 100 });
      // Filter only active users who can sign (dentists)
      const users = response.data.users || [];
      setProfessionals(users.filter(u => u.active));
    } catch (error) {
    }
  };

  const fetchPrescription = async () => {
    setLoading(true);
    try {
      const response = await prescriptionsAPI.getOne(id);
      const prescription = response.data.prescription;

      form.setFieldsValue({
        ...prescription,
        valid_until: prescription.valid_until ? dayjs(prescription.valid_until) : null,
        prescription_date: prescription.prescription_date ? dayjs(prescription.prescription_date) : dayjs(),
        signer_id: prescription.signer_id || null,
      });
    } catch (error) {
      message.error('Erro ao carregar receita');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        valid_until: values.valid_until ? values.valid_until.toISOString() : null,
        prescription_date: values.prescription_date ? values.prescription_date.toISOString() : null,
      };

      if (id) {
        await prescriptionsAPI.update(id, data);
        message.success('Receita atualizada com sucesso!');
      } else {
        await prescriptionsAPI.create(data);
        message.success('Receita criada com sucesso!');
      }
      navigate('/prescriptions');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar receita'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            <span>{id ? 'Editar Receita' : 'Nova Receita'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/prescriptions')}
          >
            Voltar
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            type: 'prescription',
          }}
        >
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Paciente"
                name="patient_id"
                rules={[{ required: true, message: 'Selecione o paciente' }]}
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {patients.map((patient) => (
                    <Select.Option key={patient.id} value={patient.id}>
                      {patient.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                label="Tipo de Documento"
                name="type"
                rules={[{ required: true, message: 'Selecione o tipo' }]}
              >
                <Select placeholder="Selecione o tipo">
                  {prescriptionTypes.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      {type.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={16}>
              <Form.Item
                label="Título"
                name="title"
              >
                <Input placeholder="Ex: Receita Médica, Atestado de Comparecimento" />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item
                label="Válido Até"
                name="valid_until"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left">Medicamentos (para receitas)</Divider>

          <Form.Item
            label="Medicamentos"
            name="medications"
          >
            <TextArea
              rows={4}
              placeholder="Liste os medicamentos, dosagens e orientações de uso"
            />
          </Form.Item>

          <Divider orientation="left">Conteúdo Principal</Divider>

          <Form.Item
            label="Conteúdo"
            name="content"
            rules={[{ required: true, message: 'Digite o conteúdo' }]}
          >
            <TextArea
              rows={8}
              placeholder="Conteúdo principal do documento (instruções, observações, etc.)"
            />
          </Form.Item>

          <Form.Item
            label="Diagnóstico"
            name="diagnosis"
          >
            <TextArea
              rows={3}
              placeholder="Diagnóstico ou observações clínicas"
            />
          </Form.Item>

          <Form.Item
            label="Observações Internas"
            name="notes"
          >
            <TextArea
              rows={2}
              placeholder="Observações internas (não aparecerão no documento impresso)"
            />
          </Form.Item>

          <Divider orientation="left">Assinatura do Documento</Divider>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                label="Data do Documento"
                name="prescription_date"
                initialValue={dayjs()}
              >
                <DatePicker
                  style={{ width: '100%' }}
                  format="DD/MM/YYYY"
                  placeholder="Selecione a data"
                />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                label="Profissional Assinante"
                name="signer_id"
              >
                <Select
                  placeholder="Selecione o profissional que vai assinar"
                  showSearch
                  allowClear
                  filterOption={(input, option) =>
                    option.children.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {professionals.map((prof) => (
                    <Select.Option key={prof.id} value={prof.id}>
                      {prof.name} {prof.cro ? `(CRO: ${prof.cro})` : ''}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={loading}
              >
                {id ? 'Atualizar' : 'Salvar Rascunho'}
              </Button>
              <Button onClick={() => navigate('/prescriptions')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default PrescriptionForm;
