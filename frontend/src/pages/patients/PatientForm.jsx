import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Form, Input, Button, Card, DatePicker, Select, Switch, message, Row, Col, Space } from 'antd';
import { patientsAPI } from '../../services/api';
import dayjs from 'dayjs';

const { TextArea } = Input;

const PatientForm = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  useEffect(() => {
    if (isEditing) {
      loadPatient();
    }
  }, [id]);

  const loadPatient = async () => {
    try {
      const response = await patientsAPI.getOne(id);
      const patient = response.data.patient;
      form.setFieldsValue({
        ...patient,
        birth_date: patient.birth_date ? dayjs(patient.birth_date) : null,
      });
    } catch (error) {
      message.error('Erro ao carregar paciente');
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        birth_date: values.birth_date ? values.birth_date.toISOString() : null,
      };

      if (isEditing) {
        await patientsAPI.update(id, data);
        message.success('Paciente atualizado com sucesso');
      } else {
        await patientsAPI.create(data);
        message.success('Paciente criado com sucesso');
      }
      navigate('/patients');
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao salvar paciente');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>{isEditing ? 'Editar Paciente' : 'Novo Paciente'}</h1>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ active: true }}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="Nome Completo" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="cpf" label="CPF">
                <Input />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="rg" label="RG">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="birth_date" label="Data de Nascimento">
                <DatePicker style={{ width: '100%' }} format="DD/MM/YYYY" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="gender" label="Gênero">
                <Select>
                  <Select.Option value="M">Masculino</Select.Option>
                  <Select.Option value="F">Feminino</Select.Option>
                  <Select.Option value="Other">Outro</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="active" label="Ativo" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="email" label="Email" rules={[{ type: 'email' }]}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="phone"
                label="Telefone"
                rules={[{ required: true, message: 'Telefone é obrigatório' }]}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="cell_phone" label="Celular">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={18}>
              <Form.Item name="address" label="Endereço">
                <Input />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="number" label="Número">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="city" label="Cidade">
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="state" label="Estado">
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="zip_code" label="CEP">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="allergies" label="Alergias">
                <TextArea rows={3} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="medications" label="Medicamentos em Uso">
                <TextArea rows={3} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="insurance_name" label="Convênio">
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="insurance_number" label="Número do Convênio">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="tags"
            label="Tags (para campanhas)"
            help="Separe as tags por vírgula. Ex: ortodontia, clareamento, vip"
          >
            <Input placeholder="Ex: ortodontia, clareamento, vip" />
          </Form.Item>

          <Form.Item name="notes" label="Observações">
            <TextArea rows={4} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                Salvar
              </Button>
              <Button onClick={() => navigate('/patients')}>
                Cancelar
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default PatientForm;
