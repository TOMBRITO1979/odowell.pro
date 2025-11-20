import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  message,
  Select,
  DatePicker,
  Spin,
} from 'antd';
import {
  ArrowLeftOutlined,
  SaveOutlined,
  FileOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { examsAPI } from '../../services/api';

const { TextArea } = Input;
const { Option } = Select;

const ExamForm = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const location = useLocation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [exam, setExam] = useState(null);

  // Modo visualização read-only
  const isReadOnly = location.pathname.includes('/view');

  useEffect(() => {
    if (id) {
      fetchExam();
    }
  }, [id]);

  const fetchExam = async () => {
    setLoading(true);
    try {
      const response = await examsAPI.getOne(id);
      const examData = response.data.exam;
      setExam(examData);

      form.setFieldsValue({
        name: examData.name,
        exam_type: examData.exam_type,
        exam_date: examData.exam_date ? dayjs(examData.exam_date) : null,
        description: examData.description,
        notes: examData.notes,
      });
    } catch (error) {
      message.error('Erro ao carregar exame');
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const onFinish = async (values) => {
    if (isReadOnly) return;

    setLoading(true);
    try {
      const data = {
        name: values.name,
        exam_type: values.exam_type,
        exam_date: values.exam_date ? values.exam_date.format('YYYY-MM-DD') : null,
        description: values.description,
        notes: values.notes,
      };

      await examsAPI.update(id, data);
      message.success('Exame atualizado com sucesso!');
      navigate('/exams');
    } catch (error) {
      message.error(
        error.response?.data?.error || 'Erro ao atualizar exame'
      );
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading && !exam) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Card
        title={
          <Space>
            <FileOutlined />
            <span>
              {isReadOnly ? 'Visualizar Exame' : 'Editar Exame'}
            </span>
          </Space>
        }
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/exams')}
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
          disabled={isReadOnly}
        >
          <Form.Item
            name="name"
            label="Nome do Exame"
            rules={[
              { required: true, message: 'Por favor, insira o nome do exame' },
            ]}
          >
            <Input placeholder="Ex: Raio-X Panorâmico" />
          </Form.Item>

          <Form.Item name="exam_type" label="Tipo de Exame">
            <Select placeholder="Selecione o tipo">
              <Option value="raio-x">Raio-X</Option>
              <Option value="tomografia">Tomografia</Option>
              <Option value="foto">Fotografia</Option>
              <Option value="laudo">Laudo</Option>
              <Option value="receita">Receita</Option>
              <Option value="atestado">Atestado</Option>
              <Option value="outro">Outro</Option>
            </Select>
          </Form.Item>

          <Form.Item name="exam_date" label="Data do Exame">
            <DatePicker
              style={{ width: '100%' }}
              format="DD/MM/YYYY"
              placeholder="Selecione a data"
            />
          </Form.Item>

          <Form.Item name="description" label="Descrição">
            <TextArea rows={3} placeholder="Descrição do exame" />
          </Form.Item>

          <Form.Item name="notes" label="Observações">
            <TextArea rows={2} placeholder="Observações adicionais" />
          </Form.Item>

          {exam && (
            <Form.Item label="Arquivo">
              <Input
                value={exam.file_name}
                disabled
                addonAfter={
                  <span style={{ color: '#999' }}>
                    Não é possível alterar o arquivo após o upload
                  </span>
                }
              />
            </Form.Item>
          )}

          {!isReadOnly && (
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
                <Button onClick={() => navigate('/exams')}>Cancelar</Button>
              </Space>
            </Form.Item>
          )}
        </Form>
      </Card>
    </div>
  );
};

export default ExamForm;
