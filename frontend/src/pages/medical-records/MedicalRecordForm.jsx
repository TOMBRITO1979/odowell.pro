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
  Tabs,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { medicalRecordsAPI, patientsAPI } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';
import Odontogram from '../../components/Odontogram';

const { TextArea } = Input;

const MedicalRecordForm = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);

  const recordTypes = [
    { value: 'anamnesis', label: 'Anamnese' },
    { value: 'treatment', label: 'Tratamento' },
    { value: 'procedure', label: 'Procedimento' },
    { value: 'prescription', label: 'Receita' },
    { value: 'certificate', label: 'Atestado' },
  ];

  useEffect(() => {
    fetchPatients();
    if (id) {
      fetchRecord();
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

  const fetchRecord = async () => {
    setLoading(true);
    try {
      const response = await medicalRecordsAPI.getOne(id);
      form.setFieldsValue(response.data.record);
    } catch (error) {
      message.error('Erro ao carregar prontuário');
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        dentist_id: user.id,
      };

      if (id) {
        await medicalRecordsAPI.update(id, data);
        message.success('Prontuário atualizado com sucesso!');
      } else {
        await medicalRecordsAPI.create(data);
        message.success('Prontuário criado com sucesso!');
      }
      navigate('/medical-records');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao salvar prontuário'
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
            <span>{id ? 'Editar Prontuário' : 'Novo Prontuário'}</span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/medical-records')}
          >
            Voltar
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="patient_id"
                label="Paciente"
                rules={[
                  { required: true, message: 'Selecione o paciente' },
                ]}
              >
                <Select
                  placeholder="Selecione o paciente"
                  showSearch
                  filterOption={(input, option) =>
                    option.children
                      .toLowerCase()
                      .includes(input.toLowerCase())
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
                name="type"
                label="Tipo de Registro"
                rules={[
                  { required: true, message: 'Selecione o tipo' },
                ]}
              >
                <Select placeholder="Selecione o tipo">
                  {recordTypes.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      {type.label}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Tabs defaultActiveKey="1">
            <Tabs.TabPane tab="Diagnóstico e Tratamento" key="1">
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Item
                    name="arlegis"
                    label="Alergias"
                  >
                    <TextArea
                      rows={2}
                      placeholder="Informe as alergias do paciente (medicamentos, materiais, látex, etc.)..."
                      style={{ borderColor: '#ff7875' }}
                    />
                  </Form.Item>
                </Col>

                <Col xs={24}>
                  <Form.Item
                    name="diagnosis"
                    label="Diagnóstico"
                  >
                    <TextArea
                      rows={4}
                      placeholder="Descreva o diagnóstico..."
                    />
                  </Form.Item>
                </Col>

                <Col xs={24}>
                  <Form.Item
                    name="treatment_plan"
                    label="Plano de Tratamento"
                  >
                    <TextArea
                      rows={4}
                      placeholder="Descreva o plano de tratamento proposto..."
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Tabs.TabPane>

            <Tabs.TabPane tab="Procedimento" key="2">
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Item
                    name="procedure_done"
                    label="Procedimento Realizado"
                  >
                    <TextArea
                      rows={4}
                      placeholder="Descreva o procedimento realizado..."
                    />
                  </Form.Item>
                </Col>

                <Col xs={24}>
                  <Form.Item
                    name="materials"
                    label="Materiais Utilizados"
                  >
                    <TextArea
                      rows={3}
                      placeholder="Liste os materiais utilizados..."
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Tabs.TabPane>

            <Tabs.TabPane tab="Prescrição" key="3">
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Item
                    name="prescription"
                    label="Prescrição Médica"
                  >
                    <TextArea
                      rows={6}
                      placeholder="Medicamentos e posologia..."
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Tabs.TabPane>

            <Tabs.TabPane tab="Atestado" key="4">
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Item
                    name="certificate"
                    label="Atestado/Declaração"
                  >
                    <TextArea
                      rows={6}
                      placeholder="Texto do atestado ou declaração..."
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Tabs.TabPane>

            <Tabs.TabPane tab="Evolução" key="5">
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Item
                    name="evolution"
                    label="Evolução do Tratamento"
                  >
                    <TextArea
                      rows={4}
                      placeholder="Descreva a evolução do paciente..."
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Tabs.TabPane>

            <Tabs.TabPane tab="Odontograma" key="6">
              <Form.Item
                name="odontogram"
                label=""
              >
                <Odontogram />
              </Form.Item>
            </Tabs.TabPane>
          </Tabs>

          <Divider />

          <Form.Item name="notes" label="Observações Gerais">
            <TextArea
              rows={4}
              placeholder="Observações adicionais..."
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                icon={<SaveOutlined />}
              >
                Salvar
              </Button>
              <Button onClick={() => navigate('/medical-records')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default MedicalRecordForm;
