import React from 'react';
import { Link } from 'react-router-dom';
import { Card, Typography, Divider, Button, Steps, Alert } from 'antd';
import {
  ArrowLeftOutlined,
  EyeOutlined,
  EditOutlined,
  DownloadOutlined,
  DeleteOutlined,
  StopOutlined,
  SafetyOutlined
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const LGPDRights = () => {
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

          <Title level={3}>Seus Direitos - LGPD</Title>
          <Text type="secondary">Lei Geral de Protecao de Dados (Lei 13.709/2018)</Text>

          <Divider />

          <Alert
            message="Seus dados estao protegidos"
            description="O OdoWell esta em conformidade com a LGPD. Voce tem direitos sobre seus dados pessoais e pode exerce-los a qualquer momento."
            type="success"
            showIcon
            icon={<SafetyOutlined />}
            style={{ marginBottom: 24 }}
          />

          <Title level={4}>Direitos do Titular dos Dados</Title>
          <Paragraph>
            Conforme o Artigo 18 da LGPD, voce tem os seguintes direitos:
          </Paragraph>

          <div style={{ marginBottom: 24 }}>
            <Card size="small" style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <EyeOutlined style={{ fontSize: 24, color: '#4CAF50', marginRight: 16 }} />
                <div>
                  <Text strong>1. Confirmacao e Acesso</Text>
                  <Paragraph style={{ margin: 0, fontSize: 13 }}>
                    Confirmar a existencia de tratamento e acessar seus dados pessoais.
                  </Paragraph>
                </div>
              </div>
            </Card>

            <Card size="small" style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <EditOutlined style={{ fontSize: 24, color: '#4CAF50', marginRight: 16 }} />
                <div>
                  <Text strong>2. Correcao</Text>
                  <Paragraph style={{ margin: 0, fontSize: 13 }}>
                    Corrigir dados incompletos, inexatos ou desatualizados.
                  </Paragraph>
                </div>
              </div>
            </Card>

            <Card size="small" style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <StopOutlined style={{ fontSize: 24, color: '#4CAF50', marginRight: 16 }} />
                <div>
                  <Text strong>3. Anonimizacao, Bloqueio ou Eliminacao</Text>
                  <Paragraph style={{ margin: 0, fontSize: 13 }}>
                    Solicitar anonimizacao, bloqueio ou eliminacao de dados desnecessarios ou excessivos.
                  </Paragraph>
                </div>
              </div>
            </Card>

            <Card size="small" style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <DownloadOutlined style={{ fontSize: 24, color: '#4CAF50', marginRight: 16 }} />
                <div>
                  <Text strong>4. Portabilidade</Text>
                  <Paragraph style={{ margin: 0, fontSize: 13 }}>
                    Obter seus dados em formato estruturado para transferencia a outro fornecedor.
                  </Paragraph>
                </div>
              </div>
            </Card>

            <Card size="small" style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <DeleteOutlined style={{ fontSize: 24, color: '#4CAF50', marginRight: 16 }} />
                <div>
                  <Text strong>5. Eliminacao</Text>
                  <Paragraph style={{ margin: 0, fontSize: 13 }}>
                    Solicitar a exclusao dos dados tratados com consentimento.
                  </Paragraph>
                </div>
              </div>
            </Card>
          </div>

          <Title level={4}>Como Exercer seus Direitos</Title>

          <Steps
            direction="vertical"
            size="small"
            current={-1}
            items={[
              {
                title: 'Entre em contato',
                description: 'Envie um email para dpo@odowell.pro ou solicite diretamente na clinica.'
              },
              {
                title: 'Identifique-se',
                description: 'Para sua seguranca, precisamos confirmar sua identidade antes de processar a solicitacao.'
              },
              {
                title: 'Especifique o pedido',
                description: 'Informe qual direito deseja exercer e forneça detalhes necessarios.'
              },
              {
                title: 'Aguarde o prazo',
                description: 'Responderemos em ate 15 dias uteis, conforme determina a LGPD.'
              }
            ]}
          />

          <Divider />

          <Title level={4}>Limitacoes ao Direito de Exclusao</Title>
          <Paragraph>
            Alguns dados nao podem ser excluidos devido a obrigacoes legais:
          </Paragraph>
          <ul>
            <li>
              <Text strong>Prontuarios medicos:</Text> Devem ser mantidos por 20 anos apos o
              ultimo atendimento (Resolucao CFO 118/2012)
            </li>
            <li>
              <Text strong>Documentos fiscais:</Text> Devem ser mantidos por 5 anos (legislacao tributaria)
            </li>
            <li>
              <Text strong>Dados para defesa em processos:</Text> Podem ser mantidos ate a prescricao
            </li>
          </ul>

          <Alert
            message="Importante"
            description="A exclusao de dados de saude pode impactar a continuidade do seu tratamento. Recomendamos discutir com seu dentista antes de solicitar a exclusao."
            type="warning"
            showIcon
            style={{ marginBottom: 24 }}
          />

          <Title level={4}>Encarregado de Dados (DPO)</Title>
          <Paragraph>
            Nosso Encarregado de Protecao de Dados esta disponivel para:
          </Paragraph>
          <ul>
            <li>Receber reclamacoes e comunicacoes dos titulares</li>
            <li>Prestar esclarecimentos sobre o tratamento de dados</li>
            <li>Orientar sobre as praticas de privacidade</li>
            <li>Interagir com a Autoridade Nacional de Protecao de Dados (ANPD)</li>
          </ul>

          <Card style={{ background: '#f5f5f5', marginBottom: 24 }}>
            <Text strong>Contato do DPO:</Text><br />
            <Text>Email: dpo@odowell.pro</Text><br />
            <Text>Endereco: Sao Paulo/SP</Text>
          </Card>

          <Title level={4}>Reclamacao a ANPD</Title>
          <Paragraph>
            Se voce acredita que seus direitos nao foram atendidos, pode registrar uma
            reclamacao junto a Autoridade Nacional de Protecao de Dados:
          </Paragraph>
          <Paragraph>
            <Text strong>Site:</Text> <a href="https://www.gov.br/anpd" target="_blank" rel="noopener noreferrer">www.gov.br/anpd</a>
          </Paragraph>

          <Divider />

          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Link to="/login">
              <Button type="primary" icon={<ArrowLeftOutlined />}>
                Voltar para Login
              </Button>
            </Link>
            <div style={{ marginTop: 16 }}>
              <Link to="/termos-de-uso" style={{ marginRight: 16 }}>Termos de Uso</Link>
              <Link to="/politica-de-privacidade" style={{ marginRight: 16 }}>Politica de Privacidade</Link>
              <Link to="/politica-de-cookies">Politica de Cookies</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default LGPDRights;
