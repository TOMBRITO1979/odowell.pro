import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Table,
  Button,
  message,
  Modal,
  Form,
  Input,
  Select,
  Upload,
  Space,
  Popconfirm,
  DatePicker,
  Tag,
} from 'antd';
import {
  PlusOutlined,
  UploadOutlined,
  DeleteOutlined,
  DownloadOutlined,
  FileOutlined,
  EyeOutlined,
  EditOutlined,
} from '@ant-design/icons';
import { examsAPI, patientsAPI } from '../../services/api';
import { actionColors, statusColors, shadows } from '../../theme/designSystem';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;

const Exams = () => {
  const navigate = useNavigate();
  const [exams, setExams] = useState([]);
  const [patients, setPatients] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [fileList, setFileList] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  useEffect(() => {
    fetchPatients();
  }, []);

  useEffect(() => {
    if (selectedPatient) {
      fetchExams();
    }
  }, [selectedPatient, pagination.current, pagination.pageSize]);

  const fetchPatients = async () => {
    try {
      const response = await patientsAPI.getAll({ page: 1, page_size: 1000 });
      setPatients(response.data.patients || []);
    } catch (error) {
      message.error('Erro ao carregar pacientes');
    }
  };

  const fetchExams = async () => {
    if (!selectedPatient) return;

    setLoading(true);
    try {
      const response = await examsAPI.getAll({
        patient_id: selectedPatient,
        page: pagination.current,
        page_size: pagination.pageSize,
      });
      setExams(response.data.exams || []);
      setPagination({
        ...pagination,
        total: response.data.total || 0,
      });
    } catch (error) {
      message.error('Erro ao carregar exames');
    } finally {
      setLoading(false);
    }
  };

  const handleUploadExam = async (values) => {
    if (fileList.length === 0) {
      message.error('Por favor, selecione um arquivo');
      return;
    }

    // Get the file object - originFileObj is used by Ant Design Upload, fallback to the file itself
    const file = fileList[0].originFileObj || fileList[0];

    if (!file) {
      message.error('Arquivo não encontrado');
      return;
    }

    const formData = new FormData();
    formData.append('file', file);
    formData.append('patient_id', selectedPatient);
    formData.append('name', values.name);
    if (values.description) formData.append('description', values.description);
    if (values.exam_type) formData.append('exam_type', values.exam_type);
    if (values.exam_date) formData.append('exam_date', values.exam_date.format('YYYY-MM-DD'));
    if (values.notes) formData.append('notes', values.notes);

    try {
      setLoading(true);
      await examsAPI.create(formData);
      message.success('Exame enviado com sucesso!');
      setModalVisible(false);
      form.resetFields();
      setFileList([]);
      fetchExams();
    } catch (error) {
      message.error('Erro ao enviar exame');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async (record) => {
    try {
      const response = await examsAPI.getDownloadURL(record.id);
      window.open(response.data.download_url, '_blank');
    } catch (error) {
      message.error('Erro ao gerar link de download');
    }
  };

  const handleDelete = async (id) => {
    try {
      await examsAPI.delete(id);
      message.success('Exame deletado com sucesso');
      fetchExams();
    } catch (error) {
      message.error('Erro ao deletar exame');
    }
  };

  const columns = [
    {
      title: 'Nome do Exame',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Space>
          <FileOutlined />
          {text}
        </Space>
      ),
    },
    {
      title: 'Tipo',
      dataIndex: 'exam_type',
      key: 'exam_type',
      render: (type) => type ? <Tag color="blue">{type}</Tag> : '-',
    },
    {
      title: 'Data do Exame',
      dataIndex: 'exam_date',
      key: 'exam_date',
      render: (date) => date ? dayjs(date).format('DD/MM/YYYY') : '-',
    },
    {
      title: 'Arquivo',
      dataIndex: 'file_name',
      key: 'file_name',
    },
    {
      title: 'Tamanho',
      dataIndex: 'file_size',
      key: 'file_size',
      render: (size) => size ? `${(size / 1024).toFixed(2)} KB` : '-',
    },
    {
      title: 'Upload em',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD/MM/YYYY HH:mm'),
    },
    {
      title: 'Ações',
      key: 'actions',
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/exams/${record.id}`)}
            title="Visualizar"
            style={{ color: actionColors.view }}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => navigate(`/exams/${record.id}/edit`)}
            title="Editar"
            style={{ color: actionColors.edit }}
          />
          <Button
            type="text"
            icon={<DownloadOutlined />}
            onClick={() => handleDownload(record)}
            title="Download"
            style={{ color: actionColors.exportPDF }}
          />
          <Popconfirm
            title="Tem certeza que deseja deletar este exame?"
            onConfirm={() => handleDelete(record.id)}
            okText="Sim"
            cancelText="Não"
          >
            <Button
              type="text"
              icon={<DeleteOutlined />}
              title="Deletar"
              style={{ color: actionColors.delete }}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const uploadProps = {
    onRemove: () => {
      setFileList([]);
    },
    beforeUpload: (file) => {
      setFileList([file]);
      return false; // Prevent auto upload
    },
    fileList,
    maxCount: 1,
  };

  return (
    <div>
      <Card
        title="Exames de Pacientes"
        extra={
          <Space>
            <Select
              showSearch
              style={{ width: 300 }}
              placeholder="Selecione um paciente"
              optionFilterProp="children"
              onChange={(value) => {
                setSelectedPatient(value);
                setPagination({ ...pagination, current: 1 });
              }}
              filterOption={(input, option) => {
                const children = String(option.children || '');
                return children.toLowerCase().includes(input.toLowerCase());
              }}
            >
              {patients.map((patient) => (
                <Option key={patient.id} value={patient.id}>
                  {patient.name} - {patient.cpf}
                </Option>
              ))}
            </Select>
            <Button
              icon={<PlusOutlined />}
              onClick={() => setModalVisible(true)}
              disabled={!selectedPatient}
              style={{
                backgroundColor: actionColors.create,
                borderColor: actionColors.create,
                color: '#fff'
              }}
            >
              Novo Exame
            </Button>
          </Space>
        }
      >
        {!selectedPatient ? (
          <div style={{ textAlign: 'center', padding: '50px 0' }}>
            <FileOutlined style={{ fontSize: 48, color: '#ccc' }} />
            <p style={{ marginTop: 16, color: '#999' }}>
              Selecione um paciente para visualizar seus exames
            </p>
          </div>
        ) : (
          <Table
            dataSource={exams}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              onChange: (page) => setPagination({ ...pagination, current: page }),
            }}
          />
        )}
      </Card>

      <Modal
        title="Upload de Novo Exame"
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
          setFileList([]);
        }}
        onOk={() => form.submit()}
        confirmLoading={loading}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleUploadExam}
        >
          <Form.Item
            name="name"
            label="Nome do Exame"
            rules={[{ required: true, message: 'Por favor, insira o nome do exame' }]}
          >
            <Input placeholder="Ex: Raio-X Panorâmico" />
          </Form.Item>

          <Form.Item
            name="exam_type"
            label="Tipo de Exame"
          >
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

          <Form.Item
            name="exam_date"
            label="Data do Exame"
          >
            <DatePicker style={{ width: '100%' }} format="DD/MM/YYYY" />
          </Form.Item>

          <Form.Item
            name="description"
            label="Descrição"
          >
            <TextArea rows={3} placeholder="Descrição do exame" />
          </Form.Item>

          <Form.Item
            name="notes"
            label="Observações"
          >
            <TextArea rows={2} placeholder="Observações adicionais" />
          </Form.Item>

          <Form.Item
            label="Arquivo"
            required
          >
            <Upload {...uploadProps}>
              <Button icon={<UploadOutlined />}>Selecionar Arquivo</Button>
            </Upload>
            <p style={{ marginTop: 8, color: '#999', fontSize: 12 }}>
              Formatos aceitos: PDF, JPG, PNG, DICOM. Tamanho máximo: 10MB
            </p>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Exams;
