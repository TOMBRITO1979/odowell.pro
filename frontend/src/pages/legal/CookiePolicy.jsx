import React from 'react';
import { Link } from 'react-router-dom';
import { Card, Typography, Divider, Button, Table, Tag } from 'antd';
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const CookiePolicy = () => {
  const cookieColumns = [
    { title: 'Tipo', dataIndex: 'type', key: 'type' },
    { title: 'Nome', dataIndex: 'name', key: 'name' },
    { title: 'Finalidade', dataIndex: 'purpose', key: 'purpose' },
    {
      title: 'Obrigatorio',
      dataIndex: 'required',
      key: 'required',
      render: (required) => required ?
        <Tag color="green" icon={<CheckCircleOutlined />}>Sim</Tag> :
        <Tag color="default" icon={<CloseCircleOutlined />}>Nao</Tag>
    },
  ];

  const cookieData = [
    {
      key: 1,
      type: 'LocalStorage',
      name: 'token',
      purpose: 'Autenticacao do usuario (JWT)',
      required: true
    },
    {
      key: 2,
      type: 'LocalStorage',
      name: 'user',
      purpose: 'Dados do usuario logado',
      required: true
    },
    {
      key: 3,
      type: 'LocalStorage',
      name: 'tenant',
      purpose: 'Identificacao da clinica',
      required: true
    },
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

          <Title level={3}>Politica de Cookies</Title>
          <Text type="secondary">Ultima atualizacao: {new Date().toLocaleDateString('pt-BR')}</Text>

          <Divider />

          <Title level={4}>1. O que sao Cookies?</Title>
          <Paragraph>
            Cookies sao pequenos arquivos de texto armazenados no seu navegador quando voce
            visita um site. Eles permitem que o site lembre suas preferencias e melhore
            sua experiencia de navegacao.
          </Paragraph>

          <Title level={4}>2. Uso de Cookies no OdoWell</Title>
          <Paragraph>
            O OdoWell utiliza uma abordagem minimalista em relacao a cookies.
            <Text strong> Nao utilizamos cookies de rastreamento ou publicidade.</Text>
          </Paragraph>
          <Paragraph>
            Em vez de cookies tradicionais, utilizamos <Text strong>LocalStorage</Text> do navegador
            para armazenar informacoes essenciais para o funcionamento do sistema:
          </Paragraph>

          <Table
            columns={cookieColumns}
            dataSource={cookieData}
            pagination={false}
            size="small"
            style={{ marginBottom: 24 }}
          />

          <Title level={4}>3. Tipos de Armazenamento</Title>

          <Card size="small" style={{ marginBottom: 12, background: '#f5f5f5' }}>
            <Text strong>LocalStorage (Utilizado)</Text>
            <Paragraph style={{ margin: '8px 0 0 0', fontSize: 13 }}>
              Armazenamento local do navegador que persiste ate ser explicitamente removido.
              Utilizado para manter sua sessao ativa entre visitas.
            </Paragraph>
          </Card>

          <Card size="small" style={{ marginBottom: 12, background: '#f5f5f5' }}>
            <Text strong>SessionStorage (Nao utilizado)</Text>
            <Paragraph style={{ margin: '8px 0 0 0', fontSize: 13 }}>
              Armazenamento temporario que e limpo quando a aba e fechada.
              Atualmente nao utilizamos este tipo de armazenamento.
            </Paragraph>
          </Card>

          <Card size="small" style={{ marginBottom: 24, background: '#f5f5f5' }}>
            <Text strong>Cookies HTTP (Nao utilizado)</Text>
            <Paragraph style={{ margin: '8px 0 0 0', fontSize: 13 }}>
              Cookies tradicionais enviados automaticamente em requisicoes HTTP.
              O OdoWell nao utiliza cookies HTTP.
            </Paragraph>
          </Card>

          <Title level={4}>4. Cookies de Terceiros</Title>
          <Paragraph>
            O OdoWell pode utilizar servicos de terceiros que possuem suas proprias
            politicas de cookies:
          </Paragraph>
          <ul>
            <li>
              <Text strong>Stripe:</Text> Processamento de pagamentos.
              <a href="https://stripe.com/privacy" target="_blank" rel="noopener noreferrer"> Politica de Privacidade</a>
            </li>
            <li>
              <Text strong>AWS:</Text> Hospedagem de arquivos (S3).
              <a href="https://aws.amazon.com/privacy/" target="_blank" rel="noopener noreferrer"> Politica de Privacidade</a>
            </li>
          </ul>

          <Title level={4}>5. Como Gerenciar</Title>
          <Paragraph>
            Voce pode gerenciar o armazenamento local do navegador de diversas formas:
          </Paragraph>
          <ul>
            <li>
              <Text strong>Logout:</Text> Ao sair do sistema, seus dados de sessao sao removidos
            </li>
            <li>
              <Text strong>Limpar dados do navegador:</Text> Acesse as configuracoes do navegador
              e limpe os dados de site para odowell.pro
            </li>
            <li>
              <Text strong>Modo anonimo:</Text> Use o modo de navegacao privada para nao
              armazenar dados permanentemente
            </li>
          </ul>

          <Title level={4}>6. Impacto da Desativacao</Title>
          <Paragraph>
            Se voce limpar ou bloquear o LocalStorage para o OdoWell:
          </Paragraph>
          <ul>
            <li>Voce sera desconectado automaticamente</li>
            <li>Precisara fazer login novamente a cada visita</li>
            <li>Suas preferencias de interface serao perdidas</li>
          </ul>
          <Paragraph>
            <Text type="secondary">
              Nota: O armazenamento local e necessario para o funcionamento basico do sistema.
              Nao e possivel usar o OdoWell sem ele.
            </Text>
          </Paragraph>

          <Title level={4}>7. Atualizacoes desta Politica</Title>
          <Paragraph>
            Esta politica pode ser atualizada periodicamente. Recomendamos que voce a
            revise regularmente para estar ciente de quaisquer alteracoes.
          </Paragraph>

          <Title level={4}>8. Contato</Title>
          <Paragraph>
            Para duvidas sobre esta politica de cookies, entre em contato:
          </Paragraph>
          <Paragraph>
            <Text strong>Email:</Text> suporte@odowell.pro<br />
            <Text strong>DPO:</Text> dpo@odowell.pro
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
              <Link to="/privacy" style={{ marginRight: 16 }}>Politica de Privacidade</Link>
              <Link to="/lgpd">Direitos LGPD</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default CookiePolicy;
