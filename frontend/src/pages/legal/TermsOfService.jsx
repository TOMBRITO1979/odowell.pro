import React from 'react';
import { Link } from 'react-router-dom';
import { Card, Typography, Divider, Button } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const TermsOfService = () => {
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

          <Title level={3}>Termos de Uso</Title>
          <Text type="secondary">Ultima atualizacao: {new Date().toLocaleDateString('pt-BR')}</Text>

          <Divider />

          <Title level={4}>1. Aceitacao dos Termos</Title>
          <Paragraph>
            Ao acessar e utilizar o sistema OdoWell, voce concorda em cumprir e estar vinculado a estes
            Termos de Uso. Se voce nao concordar com qualquer parte destes termos, nao podera acessar o servico.
          </Paragraph>

          <Title level={4}>2. Descricao do Servico</Title>
          <Paragraph>
            O OdoWell e um sistema de gestao para clinicas odontologicas que oferece:
          </Paragraph>
          <ul>
            <li>Gestao de pacientes e prontuarios</li>
            <li>Agendamento de consultas</li>
            <li>Controle financeiro (orcamentos, pagamentos, despesas)</li>
            <li>Gestao de estoque e fornecedores</li>
            <li>Emissao de receitas e laudos</li>
            <li>Campanhas de marketing</li>
            <li>Relatorios gerenciais</li>
          </ul>

          <Title level={4}>3. Conta de Usuario</Title>
          <Paragraph>
            Para utilizar o servico, voce deve criar uma conta fornecendo informacoes precisas e completas.
            Voce e responsavel por manter a confidencialidade de sua senha e por todas as atividades
            realizadas em sua conta.
          </Paragraph>

          <Title level={4}>4. Uso Aceitavel</Title>
          <Paragraph>
            Voce concorda em nao:
          </Paragraph>
          <ul>
            <li>Usar o servico para fins ilegais ou nao autorizados</li>
            <li>Violar leis aplicaveis, incluindo a LGPD</li>
            <li>Transmitir virus ou codigo malicioso</li>
            <li>Tentar acessar dados de outros usuarios sem autorizacao</li>
            <li>Interferir no funcionamento do servico</li>
          </ul>

          <Title level={4}>5. Propriedade Intelectual</Title>
          <Paragraph>
            O servico e seu conteudo original, recursos e funcionalidades sao de propriedade exclusiva
            do OdoWell e estao protegidos por leis de direitos autorais, marcas registradas e outras
            leis de propriedade intelectual.
          </Paragraph>

          <Title level={4}>6. Protecao de Dados</Title>
          <Paragraph>
            O tratamento de dados pessoais e realizado em conformidade com a Lei Geral de Protecao de
            Dados (LGPD - Lei 13.709/2018). Consulte nossa <Link to="/politica-de-privacidade">Politica de Privacidade</Link> para
            mais informacoes sobre como coletamos, usamos e protegemos seus dados.
          </Paragraph>

          <Title level={4}>7. Dados de Saude</Title>
          <Paragraph>
            Reconhecemos que os dados de saude sao dados sensiveis nos termos da LGPD. Implementamos
            medidas tecnicas e organizacionais apropriadas para proteger esses dados, incluindo:
          </Paragraph>
          <ul>
            <li>Criptografia de dados em transito e em repouso</li>
            <li>Controle de acesso baseado em funcoes</li>
            <li>Logs de auditoria de todas as acoes</li>
            <li>Isolamento de dados por tenant (clinica)</li>
          </ul>

          <Title level={4}>8. Limitacao de Responsabilidade</Title>
          <Paragraph>
            O OdoWell nao sera responsavel por quaisquer danos indiretos, incidentais, especiais,
            consequenciais ou punitivos, incluindo perda de lucros, dados, uso ou outras perdas
            intangiveis resultantes do uso ou incapacidade de uso do servico.
          </Paragraph>

          <Title level={4}>9. Alteracoes nos Termos</Title>
          <Paragraph>
            Reservamo-nos o direito de modificar estes termos a qualquer momento. Notificaremos sobre
            alteracoes significativas atraves do email cadastrado ou por aviso no sistema. O uso
            continuado do servico apos alteracoes constitui aceitacao dos novos termos.
          </Paragraph>

          <Title level={4}>10. Rescisao</Title>
          <Paragraph>
            Podemos encerrar ou suspender sua conta imediatamente, sem aviso previo, por qualquer
            motivo, incluindo violacao destes Termos de Uso. Apos a rescisao, seu direito de usar
            o servico cessara imediatamente.
          </Paragraph>

          <Title level={4}>11. Lei Aplicavel</Title>
          <Paragraph>
            Estes termos serao regidos e interpretados de acordo com as leis do Brasil, sem considerar
            suas disposicoes sobre conflitos de leis. Qualquer disputa sera submetida ao foro da
            comarca de Sao Paulo/SP.
          </Paragraph>

          <Title level={4}>12. Contato</Title>
          <Paragraph>
            Para questoes sobre estes Termos de Uso, entre em contato atraves do email:
            <Text strong> suporte@odowell.pro</Text>
          </Paragraph>

          <Divider />

          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Link to="/login">
              <Button type="primary" icon={<ArrowLeftOutlined />}>
                Voltar para Login
              </Button>
            </Link>
            <div style={{ marginTop: 16 }}>
              <Link to="/politica-de-privacidade" style={{ marginRight: 16 }}>Politica de Privacidade</Link>
              <Link to="/seus-direitos-lgpd" style={{ marginRight: 16 }}>Direitos LGPD</Link>
              <Link to="/politica-de-cookies">Politica de Cookies</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default TermsOfService;
