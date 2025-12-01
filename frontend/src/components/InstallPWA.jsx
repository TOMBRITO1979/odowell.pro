import React, { useState, useEffect } from 'react';
import { Button, Space, Typography } from 'antd';
import { DownloadOutlined, CloseOutlined } from '@ant-design/icons';

const { Text } = Typography;

const InstallPWA = () => {
  const [deferredPrompt, setDeferredPrompt] = useState(null);
  const [showBanner, setShowBanner] = useState(false);

  useEffect(() => {
    // Verifica se o usuário já fechou o banner antes
    const bannerDismissed = localStorage.getItem('pwa-install-dismissed');

    // Listener para o evento beforeinstallprompt
    const handleBeforeInstallPrompt = (e) => {
      // Previne o mini-infobar do Chrome de aparecer automaticamente
      e.preventDefault();

      // Guarda o evento para poder acionar depois
      setDeferredPrompt(e);

      // Mostra o banner customizado (se o usuário não o fechou antes)
      if (!bannerDismissed) {
        setShowBanner(true);
      }
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    };
  }, []);

  const handleInstallClick = async () => {
    if (!deferredPrompt) return;

    // Mostra o prompt de instalação
    deferredPrompt.prompt();

    // Espera o usuário responder ao prompt
    const { outcome } = await deferredPrompt.userChoice;

    // PWA installation outcome handled silently

    // Limpa o prompt
    setDeferredPrompt(null);
    setShowBanner(false);
  };

  const handleDismiss = () => {
    setShowBanner(false);
    // Salva a preferência do usuário para não mostrar novamente
    localStorage.setItem('pwa-install-dismissed', 'true');
  };

  if (!showBanner) return null;

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 1000,
        background: 'linear-gradient(135deg, #81C784 0%, #66BB6A 100%)',
        padding: '16px',
        boxShadow: '0 -4px 12px rgba(0, 0, 0, 0.15)',
        animation: 'slideUp 0.3s ease-out',
      }}
    >
      <style>
        {`
          @keyframes slideUp {
            from {
              transform: translateY(100%);
            }
            to {
              transform: translateY(0);
            }
          }
        `}
      </style>

      <div
        style={{
          maxWidth: '600px',
          margin: '0 auto',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
        }}
      >
        {/* Ícone do App */}
        <div
          style={{
            width: 48,
            height: 48,
            background: 'white',
            borderRadius: 12,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 28,
            flexShrink: 0,
          }}
        >
          🦷
        </div>

        {/* Texto */}
        <div style={{ flex: 1, color: 'white' }}>
          <Text strong style={{ color: 'white', display: 'block', fontSize: 15 }}>
            Instalar OdoWell
          </Text>
          <Text style={{ color: 'rgba(255, 255, 255, 0.9)', fontSize: 13 }}>
            Acesso rápido e funciona offline
          </Text>
        </div>

        {/* Botões */}
        <Space size="small">
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleInstallClick}
            style={{
              background: 'white',
              color: '#66BB6A',
              border: 'none',
              fontWeight: 600,
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
            }}
          >
            Instalar
          </Button>
          <Button
            type="text"
            icon={<CloseOutlined />}
            onClick={handleDismiss}
            style={{
              color: 'white',
              border: '1px solid rgba(255, 255, 255, 0.3)',
            }}
          />
        </Space>
      </div>
    </div>
  );
};

export default InstallPWA;
