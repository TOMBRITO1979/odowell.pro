import React from 'react';
import { Link } from 'react-router-dom';
import { Card, Typography, Divider, Button, Table } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const PrivacyPolicy = () => {
  const dataCollectedColumns = [
    { title: 'Tipo de Dado', dataIndex: 'type', key: 'type' },
    { title: 'Finalidade', dataIndex: 'purpose', key: 'purpose' },
    { title: 'Base Legal', dataIndex: 'legal', key: 'legal' },
  ];

  const dataCollectedData = [
    { key: 1, type: 'Nome e CPF', purpose: 'Identificacao do paciente', legal: 'Execucao de contrato' },
    { key: 2, type: 'Email e telefone', purpose: 'Comunicacao e agendamentos', legal: 'Execucao de contrato' },
    { key: 3, type: 'Endereco', purpose: 'Cadastro completo', legal: 'Execucao de contrato' },
    { key: 4, type: 'Dados de saude', purpose: 'Prontuario medico', legal: 'Tutela da saude (Art. 11, II, f)' },
    { key: 5, type: 'Historico medico', purpose: 'Atendimento adequado', legal: 'Tutela da saude (Art. 11, II, f)' },
    { key: 6, type: 'Imagens (radiografias)', purpose: 'Diagnostico', legal: 'Tutela da saude (Art. 11, II, f)' },
  ];

  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #f5fcf7 0%, #e8f8ed 50%, #dff5e5 100%)',
      padding: '40px 20px'
    }}>
      <div style={{ maxWidth: 800, margin: '0 auto' }}>
        <Card style={{ boxShadow: '0 8px 32px rgba(0,0,0,0.1)', borderRadius: 12 }}>
          <div style={{ textAlign: 'center', marginBottom: 24 }}>
            <div style={{
              width: 60,
              height: 60,
              margin: '0 auto 16px',
              background: 'linear-gradient(135deg, #66BB6A 0%, #4CAF50 100%)',
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 4px 12px rgba(102, 187, 106, 0.3)',
            }}>
              <span style={{ fontSize: 32, filter: 'brightness(0) invert(1)' }}>🦷</span>
            </div>
            <Title level={2} style={{ color: '#4CAF50', marginTop: 0 }}>OdoWell</Title>
          </div>

          <Title level={3}>Politica de Privacidade</Title>
          <Text type="secondary">Ultima atualizacao: {new Date().toLocaleDateString('pt-BR')}</Text>

          <Divider />

          <Title level={4}>1. Introducao</Title>
          <Paragraph>
            Esta Politica de Privacidade descreve como o OdoWell coleta, usa, armazena e protege
            seus dados pessoais em conformidade com a Lei Geral de Protecao de Dados (LGPD - Lei 13.709/2018).
          </Paragraph>

          <Title level={4}>2. Controlador de Dados</Title>
          <Paragraph>
            Para fins da LGPD, o <Text strong>controlador</Text> dos dados pessoais e a clinica
            odontologica que utiliza o sistema OdoWell. O OdoWell atua como <Text strong>operador</Text> dos
            dados, processando-os em nome da clinica.
          </Paragraph>

          <Title level={4}>3. Dados Coletados</Title>
          <Paragraph>
            Coletamos os seguintes tipos de dados pessoais:
          </Paragraph>
          <Table
            columns={dataCollectedColumns}
            dataSource={dataCollectedData}
            pagination={false}
            size="small"
            style={{ marginBottom: 16 }}
          />

          <Title level={4}>4. Dados Sensiveis</Title>
          <Paragraph>
            Reconhecemos que dados de saude sao dados pessoais sensiveis conforme o Art. 11 da LGPD.
            O tratamento desses dados e realizado exclusivamente para:
          </Paragraph>
          <ul>
            <li>Tutela da saude, em procedimento realizado por profissionais de saude</li>
            <li>Cumprimento de obrigacao legal ou regulatoria</li>
            <li>Mediante consentimento especifico do titular</li>
          </ul>

          <Title level={4}>5. Finalidades do Tratamento</Title>
          <Paragraph>
            Seus dados sao tratados para as seguintes finalidades:
          </Paragraph>
          <ul>
            <li>Prestacao de servicos odontologicos</li>
            <li>Agendamento e confirmacao de consultas</li>
            <li>Elaboracao de prontuarios e historico medico</li>
            <li>Emissao de receitas e atestados</li>
            <li>Gestao financeira (orcamentos e pagamentos)</li>
            <li>Comunicacao sobre tratamentos e retornos</li>
            <li>Cumprimento de obrigacoes legais e regulatorias</li>
          </ul>

          <Title level={4}>6. Compartilhamento de Dados</Title>
          <Paragraph>
            Seus dados podem ser compartilhados com:
          </Paragraph>
          <ul>
            <li><Text strong>Profissionais de saude:</Text> dentistas e auxiliares envolvidos no tratamento</li>
            <li><Text strong>Laboratorios:</Text> quando necessario para exames ou proteses</li>
            <li><Text strong>Autoridades:</Text> quando exigido por lei ou ordem judicial</li>
            <li><Text strong>Processadores de pagamento:</Text> para transacoes financeiras (ex: Stripe)</li>
          </ul>
          <Paragraph>
            <Text strong>Nao vendemos</Text> seus dados pessoais para terceiros.
          </Paragraph>

          <Title level={4}>7. Retencao de Dados</Title>
          <Paragraph>
            Os dados sao retidos pelos seguintes periodos:
          </Paragraph>
          <ul>
            <li><Text strong>Prontuarios medicos:</Text> 20 anos apos o ultimo atendimento (CFO)</li>
            <li><Text strong>Dados fiscais:</Text> 5 anos (legislacao tributaria)</li>
            <li><Text strong>Logs de auditoria:</Text> 5 anos</li>
            <li><Text strong>Dados de marketing:</Text> ate revogacao do consentimento</li>
          </ul>

          <Title level={4}>8. Seguranca dos Dados</Title>
          <Paragraph>
            Implementamos medidas tecnicas e organizacionais para proteger seus dados:
          </Paragraph>
          <ul>
            <li>Criptografia AES-256 para dados sensiveis</li>
            <li>HTTPS/TLS para transmissao de dados</li>
            <li>Controle de acesso baseado em funcoes (RBAC)</li>
            <li>Isolamento de dados por clinica (schema-per-tenant)</li>
            <li>Logs de auditoria de todas as acoes</li>
            <li>Backups automaticos com retencao de 7 dias</li>
            <li>Senhas criptografadas com bcrypt</li>
          </ul>

          <Title level={4}>9. Seus Direitos (LGPD)</Title>
          <Paragraph>
            Conforme a LGPD, voce tem direito a:
          </Paragraph>
          <ul>
            <li><Text strong>Confirmacao:</Text> saber se tratamos seus dados</li>
            <li><Text strong>Acesso:</Text> obter copia dos seus dados</li>
            <li><Text strong>Correcao:</Text> corrigir dados incorretos</li>
            <li><Text strong>Anonimizacao:</Text> solicitar anonimizacao de dados desnecessarios</li>
            <li><Text strong>Portabilidade:</Text> receber seus dados em formato estruturado</li>
            <li><Text strong>Eliminacao:</Text> solicitar exclusao de dados (respeitando retencao legal)</li>
            <li><Text strong>Revogacao:</Text> revogar consentimento a qualquer momento</li>
          </ul>
          <Paragraph>
            Para exercer seus direitos, acesse nossa pagina de <Link to="/lgpd">Direitos LGPD</Link>.
          </Paragraph>

          <Title level={4}>10. Transferencia Internacional</Title>
          <Paragraph>
            Seus dados podem ser armazenados em servidores localizados fora do Brasil (AWS).
            Garantimos que essas transferencias estao em conformidade com a LGPD, utilizando
            clausulas contratuais padrao e garantias adequadas de protecao.
          </Paragraph>

          <Title level={4}>11. Encarregado de Dados (DPO)</Title>
          <Paragraph>
            Para questoes relacionadas a protecao de dados, entre em contato com nosso
            Encarregado de Dados:
          </Paragraph>
          <Paragraph>
            <Text strong>Email:</Text> dpo@odowell.pro<br />
            <Text strong>Endereco:</Text> Sao Paulo/SP
          </Paragraph>

          <Title level={4}>12. Alteracoes nesta Politica</Title>
          <Paragraph>
            Esta politica pode ser atualizada periodicamente. Notificaremos sobre alteracoes
            significativas atraves do email cadastrado ou aviso no sistema.
          </Paragraph>

          <Divider />

          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Link to="/login">
              <Button type="primary" icon={<ArrowLeftOutlined />}>
                Voltar para Login
              </Button>
            </Link>
            <div style={{ marginTop: 16 }}>
              <Link to="/terms" style={{ marginRight: 16 }}>Termos de Uso</Link>
              <Link to="/lgpd" style={{ marginRight: 16 }}>Direitos LGPD</Link>
              <Link to="/cookies">Politica de Cookies</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default PrivacyPolicy;
