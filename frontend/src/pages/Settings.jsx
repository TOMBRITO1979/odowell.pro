import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Row, Col, Tabs, TimePicker, Switch, Modal, Alert, Typography, Space, Divider, Collapse } from 'antd';
import { SettingOutlined, ShopOutlined, ClockCircleOutlined, DollarOutlined, ApiOutlined, CopyOutlined, ReloadOutlined, DeleteOutlined, EyeOutlined, CheckCircleOutlined, CloseCircleOutlined, QuestionCircleOutlined, CreditCardOutlined, MessageOutlined, LinkOutlined, MailOutlined, WarningOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { settingsAPI, stripeSettingsAPI } from '../services/api';
import { useNavigate } from 'react-router-dom';
import { useAuth, usePermission } from '../contexts/AuthContext';
import dayjs from 'dayjs';

const { Text, Paragraph } = Typography;
const { Panel } = Collapse;

const Settings = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetchingSettings, setFetchingSettings] = useState(true);
  const { tenant, updateTenant, logout } = useAuth();
  const { isAdmin } = usePermission();
  const navigate = useNavigate();

  // API Key state
  const [apiKeyStatus, setApiKeyStatus] = useState({ has_key: false, active: false, masked_key: '' });
  const [newApiKey, setNewApiKey] = useState(null);
  const [apiKeyLoading, setApiKeyLoading] = useState(false);
  const [apiDocs, setApiDocs] = useState(null);
  const [showDocs, setShowDocs] = useState(false);

  // Stripe state
  const [stripeStatus, setStripeStatus] = useState({ stripe_connected: false, has_secret_key: false, has_webhook_secret: false });
  const [stripeSecretKey, setStripeSecretKey] = useState('');
  const [stripePublishableKey, setStripePublishableKey] = useState('');
  const [stripeWebhookSecret, setStripeWebhookSecret] = useState('');
  const [stripeLoading, setStripeLoading] = useState(false);
  const [stripeTesting, setStripeTesting] = useState(false);

  // Chatwell state
  const [chatwellToken, setChatwellToken] = useState(null);
  const [chatwellLoading, setChatwellLoading] = useState(false);
  const [embedUrl, setEmbedUrl] = useState('');

  // SMTP state
  const [smtpTesting, setSmtpTesting] = useState(false);
  const [hasSMTPPassword, setHasSMTPPassword] = useState(false);

  // Delete tenant state
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [deleteConfirmName, setDeleteConfirmName] = useState('');
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    fetchSettings();
    if (isAdmin) {
      fetchApiKeyStatus();
      fetchStripeStatus();
      fetchChatwellStatus();
    }
  }, [isAdmin]);

  const fetchSettings = async () => {
    setFetchingSettings(true);
    try {
      const response = await settingsAPI.get();
      const settings = response.data.settings;

      // Update has_smtp_password state
      setHasSMTPPassword(response.data.has_smtp_password || false);

      // Parse working hours if they exist
      const formValues = {
        clinic_name: settings.clinic_name || '',
        clinic_cnpj: settings.clinic_cnpj || '',
        clinic_address: settings.clinic_address || '',
        clinic_city: settings.clinic_city || '',
        clinic_state: settings.clinic_state || '',
        clinic_zip: settings.clinic_zip || '',
        clinic_phone: settings.clinic_phone || '',
        clinic_email: settings.clinic_email || '',
        working_hours_start: settings.working_hours_start ? dayjs(settings.working_hours_start, 'HH:mm') : null,
        working_hours_end: settings.working_hours_end ? dayjs(settings.working_hours_end, 'HH:mm') : null,
        default_appointment_duration: settings.default_appointment_duration || 30,
        payment_cash_enabled: settings.payment_cash_enabled ?? true,
        payment_credit_card_enabled: settings.payment_credit_card_enabled ?? true,
        payment_debit_card_enabled: settings.payment_debit_card_enabled ?? true,
        payment_pix_enabled: settings.payment_pix_enabled ?? true,
        payment_transfer_enabled: settings.payment_transfer_enabled ?? false,
        payment_insurance_enabled: settings.payment_insurance_enabled ?? false,
        // SMTP fields
        smtp_host: settings.smtp_host || '',
        smtp_port: settings.smtp_port || 587,
        smtp_username: settings.smtp_username || '',
        smtp_from_name: settings.smtp_from_name || '',
        smtp_from_email: settings.smtp_from_email || '',
        smtp_use_tls: settings.smtp_use_tls ?? true,
      };

      form.setFieldsValue(formValues);
    } catch (error) {
      // Settings may not exist yet, that's ok - use defaults
    } finally {
      setFetchingSettings(false);
    }
  };

  const fetchApiKeyStatus = async () => {
    try {
      const response = await settingsAPI.getAPIKeyStatus();
      setApiKeyStatus(response.data);
    } catch (error) {
      // Could not fetch API key status - ignore
    }
  };

  // Stripe functions
  const fetchStripeStatus = async () => {
    try {
      const response = await stripeSettingsAPI.get();
      setStripeStatus(response.data);
      if (response.data.stripe_publishable_key) {
        setStripePublishableKey(response.data.stripe_publishable_key);
      }
    } catch (error) {
      // Could not fetch Stripe status - ignore
    }
  };

  const handleSaveStripe = async () => {
    const values = {
      stripe_secret_key: stripeSecretKey,
      stripe_publishable_key: stripePublishableKey,
      stripe_webhook_secret: stripeWebhookSecret,
    };
    setStripeLoading(true);
    try {
      const response = await stripeSettingsAPI.update(values);
      setStripeStatus(response.data);
      message.success('Credenciais do Stripe salvas com sucesso!');
      // Clear sensitive fields after save
      setStripeSecretKey('');
      setStripeWebhookSecret('');
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao salvar credenciais do Stripe');
    } finally {
      setStripeLoading(false);
    }
  };

  const handleTestStripe = async () => {
    setStripeTesting(true);
    try {
      const response = await stripeSettingsAPI.test();
      if (response.data.connected) {
        message.success(`Conectado com sucesso! Conta: ${response.data.account_name}`);
      } else {
        message.error(response.data.error || 'Falha na conexão');
      }
    } catch (error) {
      message.error('Erro ao testar conexão');
    } finally {
      setStripeTesting(false);
    }
  };

  const handleDisconnectStripe = async () => {
    Modal.confirm({
      title: 'Desconectar Stripe',
      content: 'Tem certeza que deseja remover as credenciais do Stripe? Isso desativará todas as integrações de pagamento.',
      okText: 'Desconectar',
      cancelText: 'Cancelar',
      okType: 'danger',
      onOk: async () => {
        try {
          await stripeSettingsAPI.disconnect();
          setStripeStatus({ stripe_connected: false, has_secret_key: false, has_webhook_secret: false });
          setStripeSecretKey('');
          setStripePublishableKey('');
          setStripeWebhookSecret('');
          message.success('Stripe desconectado com sucesso');
        } catch (error) {
          message.error('Erro ao desconectar Stripe');
        }
      }
    });
  };

  // Chatwell functions
  const fetchChatwellStatus = async () => {
    try {
      const response = await settingsAPI.getEmbedToken();
      if (response.data.token) {
        setChatwellToken(response.data.token);
        setEmbedUrl(response.data.embed_url || '');
      }
    } catch (error) {
      // Token not configured - ignore
    }
  };

  const handleGenerateChatwellToken = async () => {
    setChatwellLoading(true);
    try {
      const response = await settingsAPI.generateEmbedToken();
      setChatwellToken(response.data.token);
      setEmbedUrl(response.data.embed_url || '');
      message.success('Token gerado com sucesso!');
    } catch (error) {
      message.error('Erro ao gerar token');
    } finally {
      setChatwellLoading(false);
    }
  };

  const handleRevokeChatwellToken = async () => {
    Modal.confirm({
      title: 'Revogar Token',
      content: 'Tem certeza que deseja revogar o token? O painel externo perderá acesso imediatamente.',
      okText: 'Revogar',
      cancelText: 'Cancelar',
      okType: 'danger',
      onOk: async () => {
        setChatwellLoading(true);
        try {
          await settingsAPI.revokeEmbedToken();
          setChatwellToken(null);
          setEmbedUrl('');
          message.success('Token revogado com sucesso');
        } catch (error) {
          message.error('Erro ao revogar token');
        } finally {
          setChatwellLoading(false);
        }
      }
    });
  };

  const handleCopyEmbedUrl = () => {
    if (embedUrl) {
      navigator.clipboard.writeText(embedUrl);
      message.success('URL copiada para a área de transferência');
    }
  };

  // SMTP Test function
  const handleTestSMTP = async () => {
    setSmtpTesting(true);
    try {
      const response = await settingsAPI.testSMTP();
      if (response.data.success) {
        message.success(response.data.message || 'Conexão SMTP testada com sucesso!');
      } else {
        message.error(response.data.error || 'Falha no teste de conexão SMTP');
      }
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao testar conexão SMTP');
    } finally {
      setSmtpTesting(false);
    }
  };

  const handleGenerateApiKey = async () => {
    Modal.confirm({
      title: 'Gerar Nova Chave de API',
      content: apiKeyStatus.has_key
        ? 'Isso substituirá a chave existente. Todas as integrações que usam a chave atual deixarão de funcionar. Deseja continuar?'
        : 'Uma nova chave de API será gerada para sua clínica. Você poderá usar esta chave para integrar com WhatsApp e assistentes de IA.',
      okText: 'Gerar',
      cancelText: 'Cancelar',
      onOk: async () => {
        setApiKeyLoading(true);
        try {
          const response = await settingsAPI.generateAPIKey();
          setNewApiKey(response.data.api_key);
          setApiKeyStatus({ has_key: true, active: true, masked_key: response.data.api_key.substring(0, 8) + '...' + response.data.api_key.slice(-4) });
          message.success('Chave de API gerada com sucesso!');
        } catch (error) {
          message.error('Erro ao gerar chave de API');
        } finally {
          setApiKeyLoading(false);
        }
      }
    });
  };

  const handleToggleApiKey = async () => {
    setApiKeyLoading(true);
    try {
      await settingsAPI.toggleAPIKey(!apiKeyStatus.active);
      setApiKeyStatus({ ...apiKeyStatus, active: !apiKeyStatus.active });
      message.success(apiKeyStatus.active ? 'Chave de API desativada' : 'Chave de API ativada');
    } catch (error) {
      message.error('Erro ao alterar status da chave de API');
    } finally {
      setApiKeyLoading(false);
    }
  };

  const handleRevokeApiKey = async () => {
    Modal.confirm({
      title: 'Revogar Chave de API',
      content: 'Isso desativará permanentemente a chave de API. Todas as integrações deixarão de funcionar imediatamente. Esta ação não pode ser desfeita.',
      okText: 'Revogar',
      cancelText: 'Cancelar',
      okType: 'danger',
      onOk: async () => {
        setApiKeyLoading(true);
        try {
          await settingsAPI.revokeAPIKey();
          setApiKeyStatus({ has_key: false, active: false, masked_key: '' });
          setNewApiKey(null);
          message.success('Chave de API revogada com sucesso');
        } catch (error) {
          message.error('Erro ao revogar chave de API');
        } finally {
          setApiKeyLoading(false);
        }
      }
    });
  };

  const handleCopyApiKey = () => {
    if (newApiKey) {
      navigator.clipboard.writeText(newApiKey);
      message.success('Chave de API copiada para a área de transferência');
    }
  };

  const handleViewDocs = async () => {
    try {
      const response = await settingsAPI.getAPIKeyDocs();
      setApiDocs(response.data);
      setShowDocs(true);
    } catch (error) {
      message.error('Erro ao carregar documentação');
    }
  };

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      // Format time values
      const settingsData = {
        ...values,
        working_hours_start: values.working_hours_start?.format('HH:mm'),
        working_hours_end: values.working_hours_end?.format('HH:mm'),
      };

      await settingsAPI.update(settingsData);

      // Update tenant name in context if clinic name changed
      if (values.clinic_name && values.clinic_name !== tenant?.name) {
        updateTenant({
          ...tenant,
          name: values.clinic_name
        });
      }

      message.success('Configurações salvas com sucesso');
    } catch (error) {
      message.error('Erro ao salvar configurações');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteTenant = async () => {
    if (deleteConfirmName !== tenant?.name) {
      message.error('O nome digitado nao confere');
      return;
    }

    try {
      setDeleting(true);
      await settingsAPI.deleteTenant();
      message.success('Empresa deletada com sucesso. Voce sera desconectado.');
      setDeleteModalVisible(false);
      // Logout and redirect to login
      setTimeout(() => {
        logout();
        navigate('/login');
      }, 1500);
    } catch (error) {
      message.error(error.response?.data?.error || 'Erro ao deletar empresa');
    } finally {
      setDeleting(false);
    }
  };

  const clinicInfoTab = (
    <Row gutter={16}>
      <Col xs={24} md={12}>
        <Form.Item
          label="Nome da Clínica"
          name="clinic_name"
          rules={[{ required: true, message: 'Por favor, insira o nome da clínica' }]}
        >
          <Input placeholder="Nome da sua clínica" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="CNPJ/CPF"
          name="clinic_cnpj"
        >
          <Input placeholder="00.000.000/0000-00" />
        </Form.Item>
      </Col>

      <Col xs={24}>
        <Form.Item
          label="Endereço"
          name="clinic_address"
        >
          <Input placeholder="Rua, número" />
        </Form.Item>
      </Col>

      <Col xs={24} md={8}>
        <Form.Item
          label="Cidade"
          name="clinic_city"
        >
          <Input placeholder="Cidade" />
        </Form.Item>
      </Col>

      <Col xs={24} md={4}>
        <Form.Item
          label="Estado"
          name="clinic_state"
        >
          <Input placeholder="UF" maxLength={2} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="CEP"
          name="clinic_zip"
        >
          <Input placeholder="00000-000" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Telefone"
          name="clinic_phone"
        >
          <Input placeholder="(00) 0000-0000" />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="E-mail"
          name="clinic_email"
          rules={[{ type: 'email', message: 'E-mail inválido' }]}
        >
          <Input placeholder="contato@clinica.com" />
        </Form.Item>
      </Col>
    </Row>
  );

  const scheduleTab = (
    <Row gutter={16}>
      <Col xs={24} md={12}>
        <Form.Item
          label="Horário de Abertura"
          name="working_hours_start"
        >
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Horário de Fechamento"
          name="working_hours_end"
        >
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Duração Padrão da Consulta (minutos)"
          name="default_appointment_duration"
        >
          <Input type="number" min={15} step={15} placeholder="30" />
        </Form.Item>
      </Col>
    </Row>
  );

  const paymentTab = (
    <Row gutter={16}>
      <Col xs={24}>
        <p style={{ marginBottom: 16, color: '#666' }}>
          Selecione as formas de pagamento aceitas pela clínica:
        </p>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Dinheiro"
          name="payment_cash_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Cartão de Crédito"
          name="payment_credit_card_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Cartão de Débito"
          name="payment_debit_card_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="PIX"
          name="payment_pix_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Transferência Bancária"
          name="payment_transfer_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>

      <Col xs={24} md={12}>
        <Form.Item
          label="Convênio"
          name="payment_insurance_enabled"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
      </Col>
    </Row>
  );

  const apiTab = (
    <div>
      <Alert
        message="Integração WhatsApp / IA"
        description="Configure uma chave de API para permitir que assistentes de IA e bots do WhatsApp interajam com o sistema da sua clínica. Os pacientes poderão verificar consultas, remarcar e entrar na lista de espera automaticamente."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* API Key Status */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row align="middle" justify="space-between">
          <Col>
            <Space>
              <Text strong>Status da Chave de API:</Text>
              {apiKeyStatus.has_key ? (
                apiKeyStatus.active ? (
                  <Text type="success"><CheckCircleOutlined /> Ativa</Text>
                ) : (
                  <Text type="warning"><CloseCircleOutlined /> Desativada</Text>
                )
              ) : (
                <Text type="secondary">Não configurada</Text>
              )}
            </Space>
          </Col>
          {apiKeyStatus.has_key && (
            <Col>
              <Space>
                <Text type="secondary">Chave: {apiKeyStatus.masked_key}</Text>
              </Space>
            </Col>
          )}
        </Row>
      </Card>

      {/* New API Key Display */}
      {newApiKey && (
        <Alert
          message="Sua Nova Chave de API"
          description={
            <div>
              <Paragraph>
                <Text strong>IMPORTANTE:</Text> Copie e guarde esta chave em local seguro.
                Ela só será exibida uma vez!
              </Paragraph>
              <div style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, marginBottom: 12 }}>
                <Text code copyable={{ text: newApiKey }}>
                  {newApiKey}
                </Text>
              </div>
              <Button icon={<CopyOutlined />} onClick={handleCopyApiKey}>
                Copiar Chave
              </Button>
            </div>
          }
          type="success"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {/* Actions */}
      <Space wrap style={{ marginBottom: 24 }}>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={handleGenerateApiKey}
          loading={apiKeyLoading}
        >
          {apiKeyStatus.has_key ? 'Regenerar Chave' : 'Gerar Chave de API'}
        </Button>

        {apiKeyStatus.has_key && (
          <>
            <Button
              icon={apiKeyStatus.active ? <CloseCircleOutlined /> : <CheckCircleOutlined />}
              onClick={handleToggleApiKey}
              loading={apiKeyLoading}
            >
              {apiKeyStatus.active ? 'Desativar' : 'Ativar'}
            </Button>
            <Button
              danger
              icon={<DeleteOutlined />}
              onClick={handleRevokeApiKey}
              loading={apiKeyLoading}
            >
              Revogar
            </Button>
          </>
        )}

        <Button
          icon={<QuestionCircleOutlined />}
          onClick={handleViewDocs}
        >
          Ver Documentação
        </Button>
      </Space>

      <Divider />

      {/* Usage Instructions */}
      <Collapse>
        <Panel header="Como usar a API" key="1">
          <Paragraph>
            <ol>
              <li>Gere uma chave de API clicando no botão acima</li>
              <li>Copie a chave gerada e configure no seu bot do WhatsApp ou assistente de IA</li>
              <li>Todas as requisições devem incluir o header <Text code>X-API-Key</Text> com sua chave</li>
              <li>Os pacientes podem verificar sua identidade com CPF e data de nascimento</li>
              <li>Após verificado, o paciente pode consultar, cancelar ou remarcar consultas</li>
            </ol>
          </Paragraph>
        </Panel>
        <Panel header="Endpoints Disponíveis" key="2">
          <Paragraph>
            <ul>
              <li><Text code>POST /api/whatsapp/verify</Text> - Verificar identidade do paciente</li>
              <li><Text code>GET /api/whatsapp/appointments</Text> - Listar consultas agendadas</li>
              <li><Text code>POST /api/whatsapp/appointments/cancel</Text> - Cancelar consulta</li>
              <li><Text code>POST /api/whatsapp/appointments/reschedule</Text> - Remarcar consulta</li>
              <li><Text code>GET /api/whatsapp/slots</Text> - Horários disponíveis</li>
              <li><Text code>POST /api/whatsapp/waiting-list</Text> - Entrar na lista de espera</li>
              <li><Text code>GET /api/whatsapp/procedures</Text> - Lista de procedimentos</li>
              <li><Text code>GET /api/whatsapp/dentists</Text> - Lista de profissionais</li>
            </ul>
          </Paragraph>
        </Panel>
      </Collapse>

      {/* API Documentation Modal */}
      <Modal
        title="Documentação da API WhatsApp"
        open={showDocs}
        onCancel={() => setShowDocs(false)}
        footer={null}
        width={800}
      >
        {apiDocs && (
          <div>
            <Paragraph>{apiDocs.description}</Paragraph>

            <Divider>Autenticação</Divider>
            <Paragraph>
              <Text strong>Header:</Text> <Text code>{apiDocs.authentication?.header}</Text>
              <br />
              <Text type="secondary">{apiDocs.authentication?.description}</Text>
            </Paragraph>

            <Divider>URL Base</Divider>
            <Paragraph>
              <Text code>{apiDocs.base_url}</Text>
            </Paragraph>

            <Divider>Endpoints</Divider>
            {apiDocs.endpoints?.map((endpoint, index) => (
              <Card key={index} size="small" style={{ marginBottom: 8 }}>
                <Text strong>{endpoint.method}</Text> <Text code>{endpoint.path}</Text>
                <br />
                <Text type="secondary">{endpoint.description}</Text>
              </Card>
            ))}

            <Divider>Fluxo de Exemplo</Divider>
            <ol>
              {apiDocs.example_flow?.map((step, index) => (
                <li key={index}>{step}</li>
              ))}
            </ol>
          </div>
        )}
      </Modal>
    </div>
  );

  const stripeTab = (
    <div>
      <Alert
        message="Integração Stripe para Assinaturas"
        description="Configure suas credenciais do Stripe para permitir que pacientes assinem planos de pagamento recorrente diretamente pelo sistema."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* Status */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row align="middle" justify="space-between">
          <Col>
            <Space>
              <Text strong>Status:</Text>
              {stripeStatus.stripe_connected ? (
                <Text type="success"><CheckCircleOutlined /> Conectado</Text>
              ) : (
                <Text type="secondary"><CloseCircleOutlined /> Não conectado</Text>
              )}
            </Space>
          </Col>
          {stripeStatus.stripe_account_name && (
            <Col>
              <Text type="secondary">Conta: {stripeStatus.stripe_account_name}</Text>
            </Col>
          )}
        </Row>
      </Card>

      {/* Using div wrapper to avoid nested forms issue */}
      <div>
        <Divider>Credenciais do Stripe</Divider>

        <Row gutter={16}>
          <Col xs={24}>
            <div style={{ marginBottom: 24 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>Secret Key</label>
              <Input.Password
                value={stripeSecretKey}
                onChange={(e) => setStripeSecretKey(e.target.value)}
                placeholder={stripeStatus.has_secret_key ? "••••••••••••••••" : "sk_live_... ou sk_test_..."}
              />
              <div style={{ marginTop: 4, fontSize: 12, color: '#888' }}>Encontre em: Dashboard Stripe &gt; Developers &gt; API Keys</div>
            </div>
          </Col>

          <Col xs={24}>
            <div style={{ marginBottom: 24 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>Publishable Key</label>
              <Input
                value={stripePublishableKey}
                onChange={(e) => setStripePublishableKey(e.target.value)}
                placeholder="pk_live_... ou pk_test_..."
              />
              <div style={{ marginTop: 4, fontSize: 12, color: '#888' }}>Opcional - usada para componentes do frontend</div>
            </div>
          </Col>

          <Col xs={24}>
            <div style={{ marginBottom: 24 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>Webhook Secret</label>
              <Input.Password
                value={stripeWebhookSecret}
                onChange={(e) => setStripeWebhookSecret(e.target.value)}
                placeholder={stripeStatus.has_webhook_secret ? "••••••••••••••••" : "whsec_..."}
              />
              <div style={{ marginTop: 4, fontSize: 12, color: '#888' }}>Configure o webhook no Stripe e cole o secret aqui</div>
            </div>
          </Col>
        </Row>

        <Space wrap>
          <Button type="primary" loading={stripeLoading} onClick={handleSaveStripe}>
            Salvar Credenciais
          </Button>
          {stripeStatus.stripe_connected && (
            <>
              <Button onClick={handleTestStripe} loading={stripeTesting}>
                Testar Conexão
              </Button>
              <Button danger onClick={handleDisconnectStripe}>
                Desconectar
              </Button>
            </>
          )}
        </Space>
      </div>
    </div>
  );

  const smtpTab = (
    <div>
      <Alert
        message="Configuração SMTP para Campanhas de Email"
        description="Configure suas credenciais SMTP para enviar campanhas de email marketing para seus pacientes. Cada clínica pode usar seu próprio servidor de email."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* Status */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row align="middle" justify="space-between">
          <Col>
            <Space>
              <Text strong>Status da Configuração:</Text>
              {hasSMTPPassword ? (
                <Text type="success"><CheckCircleOutlined /> Configurado</Text>
              ) : (
                <Text type="secondary"><CloseCircleOutlined /> Não configurado</Text>
              )}
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Form.Item
            label="Host SMTP"
            name="smtp_host"
            rules={[{ required: false }]}
          >
            <Input placeholder="smtp.exemplo.com" />
          </Form.Item>
        </Col>

        <Col xs={24} md={6}>
          <Form.Item
            label="Porta"
            name="smtp_port"
            rules={[{ required: false }]}
          >
            <Input type="number" placeholder="587" />
          </Form.Item>
        </Col>

        <Col xs={24} md={6}>
          <Form.Item
            label="Usar TLS"
            name="smtp_use_tls"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        </Col>

        <Col xs={24} md={12}>
          <Form.Item
            label="Usuário SMTP"
            name="smtp_username"
          >
            <Input placeholder="usuario@exemplo.com" />
          </Form.Item>
        </Col>

        <Col xs={24} md={12}>
          <Form.Item
            label="Senha SMTP"
            name="smtp_password"
            help={hasSMTPPassword ? "Deixe em branco para manter a senha atual" : ""}
          >
            <Input.Password placeholder={hasSMTPPassword ? "••••••••••••••••" : "Senha do servidor SMTP"} />
          </Form.Item>
        </Col>

        <Col xs={24} md={12}>
          <Form.Item
            label="Nome do Remetente"
            name="smtp_from_name"
          >
            <Input placeholder="Nome da Clínica" />
          </Form.Item>
        </Col>

        <Col xs={24} md={12}>
          <Form.Item
            label="Email do Remetente"
            name="smtp_from_email"
            rules={[{ type: 'email', message: 'E-mail inválido' }]}
          >
            <Input placeholder="contato@clinica.com" />
          </Form.Item>
        </Col>
      </Row>

      <Space wrap style={{ marginTop: 16 }}>
        <Button onClick={handleTestSMTP} loading={smtpTesting} disabled={!hasSMTPPassword}>
          Testar Conexão
        </Button>
      </Space>

      <Divider />

      <Collapse>
        <Panel header="Como configurar" key="1">
          <Paragraph>
            <ol>
              <li>Obtenha as credenciais SMTP do seu provedor de email (Gmail, SendGrid, AWS SES, etc.)</li>
              <li>Preencha os campos acima com as informações do servidor</li>
              <li>Clique em "Salvar Configurações" para salvar</li>
              <li>Use "Testar Conexão" para verificar se está funcionando</li>
              <li>Agora você pode criar campanhas de email em Marketing → Campanhas</li>
            </ol>
          </Paragraph>
        </Panel>
        <Panel header="Configurações comuns" key="2">
          <Paragraph>
            <ul>
              <li><strong>Gmail:</strong> smtp.gmail.com, porta 587, TLS ativado, use senha de app</li>
              <li><strong>SendGrid:</strong> smtp.sendgrid.net, porta 587, TLS ativado</li>
              <li><strong>AWS SES:</strong> email-smtp.[regiao].amazonaws.com, porta 587, TLS ativado</li>
              <li><strong>Mailgun:</strong> smtp.mailgun.org, porta 587, TLS ativado</li>
            </ul>
          </Paragraph>
        </Panel>
      </Collapse>
    </div>
  );

  const chatwellTab = (
    <div>
      <Alert
        message="Integração Chatwell / Painel Externo"
        description="Gere um token para incorporar páginas do OdoWell em painéis externos como o Chatwell. O atendente poderá visualizar agenda, pacientes e outras informações diretamente no painel de atendimento."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* Status */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row align="middle" justify="space-between">
          <Col>
            <Space>
              <Text strong>Status do Token:</Text>
              {chatwellToken ? (
                <Text type="success"><CheckCircleOutlined /> Configurado</Text>
              ) : (
                <Text type="secondary">Não configurado</Text>
              )}
            </Space>
          </Col>
        </Row>
      </Card>

      {chatwellToken && embedUrl && (
        <Alert
          message="URL de Incorporação"
          description={
            <div>
              <Paragraph>
                Use esta URL para incorporar o OdoWell no seu painel externo:
              </Paragraph>
              <div style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, marginBottom: 12 }}>
                <Text code copyable={{ text: embedUrl }}>
                  {embedUrl}
                </Text>
              </div>
              <Button icon={<CopyOutlined />} onClick={handleCopyEmbedUrl}>
                Copiar URL
              </Button>
            </div>
          }
          type="success"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Space wrap>
        <Button
          type="primary"
          icon={<LinkOutlined />}
          onClick={handleGenerateChatwellToken}
          loading={chatwellLoading}
        >
          {chatwellToken ? 'Regenerar Token' : 'Gerar Token'}
        </Button>
        {chatwellToken && (
          <Button danger onClick={handleRevokeChatwellToken} loading={chatwellLoading}>
            Revogar Token
          </Button>
        )}
      </Space>

      <Divider />

      <Collapse>
        <Panel header="Como usar" key="1">
          <Paragraph>
            <ol>
              <li>Clique em "Gerar Token" para criar um token de acesso</li>
              <li>Copie a URL de incorporação gerada</li>
              <li>No Chatwell ou outro painel, adicione um iframe com a URL</li>
              <li>O atendente terá acesso às informações da clínica</li>
            </ol>
          </Paragraph>
        </Panel>
      </Collapse>
    </div>
  );

  const tabItems = [
    {
      key: 'clinic',
      label: (
        <span>
          <ShopOutlined />
          Dados da Clínica
        </span>
      ),
      children: clinicInfoTab,
    },
    {
      key: 'schedule',
      label: (
        <span>
          <ClockCircleOutlined />
          Agenda
        </span>
      ),
      children: scheduleTab,
    },
    {
      key: 'payment',
      label: (
        <span>
          <DollarOutlined />
          Pagamentos
        </span>
      ),
      children: paymentTab,
    },
    // Only show API tab for admins
    ...(isAdmin ? [{
      key: 'api',
      label: (
        <span>
          <ApiOutlined />
          Integração API
        </span>
      ),
      children: apiTab,
    }] : []),
    // Stripe tab for admins
    ...(isAdmin ? [{
      key: 'stripe',
      label: (
        <span>
          <CreditCardOutlined />
          Stripe
        </span>
      ),
      children: stripeTab,
    }] : []),
    // SMTP tab for admins (campaign emails)
    ...(isAdmin ? [{
      key: 'smtp',
      label: (
        <span>
          <MailOutlined />
          Email / SMTP
        </span>
      ),
      children: smtpTab,
    }] : []),
    // Chatwell tab for admins
    ...(isAdmin ? [{
      key: 'chatwell',
      label: (
        <span>
          <MessageOutlined />
          Chatwell
        </span>
      ),
      children: chatwellTab,
    }] : []),
    // Danger Zone tab for admins
    ...(isAdmin ? [{
      key: 'danger',
      label: (
        <span style={{ color: '#ff4d4f' }}>
          <WarningOutlined />
          Zona de Perigo
        </span>
      ),
      children: (
        <div>
          <Alert
            message="Zona de Perigo"
            description="As acoes nesta secao sao irreversiveis. Tenha certeza do que esta fazendo."
            type="error"
            showIcon
            icon={<ExclamationCircleOutlined />}
            style={{ marginBottom: 24 }}
          />

          <Card
            title={
              <Space>
                <DeleteOutlined style={{ color: '#ff4d4f' }} />
                <span>Deletar Empresa</span>
              </Space>
            }
            style={{ borderColor: '#ff4d4f' }}
          >
            <Paragraph>
              Esta acao ira desativar permanentemente sua empresa e todos os usuarios.
              Os dados serao mantidos no banco de dados, mas voce nao podera mais acessar o sistema.
            </Paragraph>
            <Paragraph type="secondary">
              Apos a delecao, entre em contato com o suporte caso precise recuperar a conta.
            </Paragraph>
            <Button
              danger
              type="primary"
              icon={<DeleteOutlined />}
              onClick={() => {
                setDeleteConfirmName('');
                setDeleteModalVisible(true);
              }}
            >
              Deletar Minha Empresa
            </Button>
          </Card>
        </div>
      ),
    }] : []),
  ];

  return (
    <div>
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SettingOutlined />
            <span>Configurações da Clínica</span>
          </div>
        }
        loading={fetchingSettings}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
        >
          <Tabs items={tabItems} />

          <div style={{ marginTop: 24, paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} size="large">
                Salvar Configurações
              </Button>
            </Form.Item>
          </div>
        </Form>
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        title={
          <Space>
            <DeleteOutlined style={{ color: '#ff4d4f' }} />
            <span>Deletar Empresa</span>
          </Space>
        }
        open={deleteModalVisible}
        onCancel={() => {
          setDeleteModalVisible(false);
          setDeleteConfirmName('');
        }}
        footer={[
          <Button key="cancel" onClick={() => setDeleteModalVisible(false)}>
            Cancelar
          </Button>,
          <Button
            key="delete"
            type="primary"
            danger
            icon={<DeleteOutlined />}
            disabled={deleteConfirmName !== tenant?.name}
            loading={deleting}
            onClick={handleDeleteTenant}
          >
            Deletar Empresa
          </Button>,
        ]}
      >
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          <Alert
            type="warning"
            showIcon
            message="Esta acao ira desativar permanentemente sua empresa e todos os usuarios."
            description="Os dados serao mantidos no banco de dados, mas voce nao podera mais acessar o sistema. Voce sera desconectado apos a delecao."
          />
          <div>
            <Text>Para confirmar, digite o nome da sua empresa:</Text>
            <Text strong style={{ display: 'block', marginTop: 4, marginBottom: 8 }}>
              {tenant?.name}
            </Text>
            <Input
              placeholder="Digite o nome da empresa"
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
              status={deleteConfirmName && deleteConfirmName !== tenant?.name ? 'error' : ''}
            />
          </div>
        </Space>
      </Modal>
    </div>
  );
};

export default Settings;
