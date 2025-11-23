import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  DatePicker,
  Button,
  Card,
  message,
  Spin,
  Space
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import api from '../../services/api';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;

const WaitingListForm = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();

  const [loading, setLoading] = useState(false);
  const [patients, setPatients] = useState([]);
  const [dentists, setDentists] = useState([]);
  const [loadingData, setLoadingData] = useState(false);

  const isEdit = Boolean(id);

  useEffect(() => {
    fetchPatients();
    fetchDentists();
    if (isEdit) {
      fetchEntry();
    }
  }, [id]);

  const fetchPatients = async () => {
    try {
      const response = await api.get('/patients', {
        params: { page: 1, page_size: 1000 }
      });
      setPatients(response.data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    }
  };

  const fetchDentists = async () => {
    try {
      const response = await api.get('/users');
      setDentists(response.data.users || []);
    } catch (error) {
      console.error('Error fetching dentists:', error);
    }
  };

  const fetchEntry = async () => {
    setLoadingData(true);
    try {
      const response = await api.get(`/waiting-list/${id}`);
      const entry = response.data;

      form.setFieldsValue({
        patient_id: entry.patient_id,
        dentist_id: entry.dentist_id,
        procedure: entry.procedure,
        priority: entry.priority,
        status: entry.status,
        notes: entry.notes
      });
    } catch (error) {
      message.error('Erro ao carregar entrada da lista');
      console.error('Error fetching entry:', error);
    } finally {
      setLoadingData(false);
    }
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      if (isEdit) {
        await api.put(`/waiting-list/${id}`, values);
        message.success('Lista de espera atualizada');
      } else {
        await api.post('/waiting-list', values);
        message.success('Paciente adicionado à lista de espera');
      }
      navigate('/waiting-list');
    } catch (error) {
      message.error(isEdit ? 'Erro ao atualizar' : 'Erro ao adicionar');
      console.error('Error submitting:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: 800, margin: '0 auto' }}>
      <Card title={isEdit ? 'Editar Entrada na Lista de Espera' : 'Adicionar à Lista de Espera'}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            priority: 'normal',
            status: 'waiting'
          }}
        >
          <Form.Item
            name="patient_id"
            label="Paciente"
            rules={[{ required: true, message: 'Selecione o paciente' }]}
          >
            <Select
              showSearch
              placeholder="Selecione o paciente"
              optionFilterProp="children"
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {patients.map(patient => (
                <Option key={patient.id} value={patient.id}>
                  {patient.name} {patient.cpf && `(${patient.cpf})`}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="procedure"
            label="Procedimento Desejado"
            rules={[{ required: true, message: 'Informe o procedimento' }]}
          >
            <Input placeholder="Ex: Limpeza, Canal, Extração..." />
          </Form.Item>

          <Form.Item
            name="dentist_id"
            label="Dentista Preferido (opcional)"
            help="Deixe em branco para qualquer dentista"
          >
            <Select
              showSearch
              placeholder="Qualquer dentista"
              allowClear
              optionFilterProp="children"
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {dentists.map(dentist => (
                <Option key={dentist.id} value={dentist.id}>
                  {dentist.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="priority"
            label="Prioridade"
            rules={[{ required: true, message: 'Selecione a prioridade' }]}
          >
            <Select>
              <Option value="normal">Normal</Option>
              <Option value="urgent">Urgente</Option>
            </Select>
          </Form.Item>

          {isEdit && (
            <Form.Item
              name="status"
              label="Status"
            >
              <Select>
                <Option value="waiting">Aguardando</Option>
                <Option value="contacted">Contatado</Option>
                <Option value="scheduled">Agendado</Option>
                <Option value="cancelled">Cancelado</Option>
              </Select>
            </Form.Item>
          )}

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea
              rows={4}
              placeholder="Informações adicionais, preferências de horário, etc."
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {isEdit ? 'Atualizar' : 'Adicionar'}
              </Button>
              <Button onClick={() => navigate('/waiting-list')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default WaitingListForm;
