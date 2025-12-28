import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import ptBR from 'antd/locale/pt_BR'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import 'dayjs/locale/pt-br'
import App from './App'
import { AuthProvider } from './contexts/AuthContext'
import './index.css'
import './styles/mobile.css'
import './styles/mobile-override.css'

// Limpa Service Workers antigos para evitar problemas de cache
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.getRegistrations().then(registrations => {
    registrations.forEach(registration => {
      registration.unregister()
      console.log('Service Worker removido:', registration.scope)
    })
  })
  // Limpa caches antigos
  if ('caches' in window) {
    caches.keys().then(names => {
      names.forEach(name => {
        caches.delete(name)
        console.log('Cache removido:', name)
      })
    })
  }
}

// Configure dayjs with timezone support
dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.locale('pt-br')
// Set default timezone to America/Sao_Paulo
dayjs.tz.setDefault('America/Sao_Paulo')

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <ConfigProvider
        locale={ptBR}
        theme={{
          token: {
            // Cor primária suave
            colorPrimary: '#66BB6A',
            colorLink: '#5C6BC0',
            colorLinkHover: '#3F51B5',
            // Cores de sucesso/erro/warning/info
            colorSuccess: '#81C784',
            colorWarning: '#FFD54F',
            colorError: '#EF9A9A',
            colorInfo: '#64B5F6',
            // Texto
            colorText: '#37474F',
            colorTextSecondary: '#78909C',
            colorTextTertiary: '#90A4AE',
            colorTextDisabled: '#B0BEC5',
            // Fundos
            colorBgContainer: '#FFFFFF',
            colorBgLayout: '#F5F5F5',
            colorBgElevated: '#FFFFFF',
            // Bordas
            colorBorder: '#E0E0E0',
            colorBorderSecondary: '#EEEEEE',
            // Border radius moderno
            borderRadius: 8,
            borderRadiusLG: 12,
            borderRadiusSM: 4,
            // Sombras suaves
            boxShadow: '0 2px 4px rgba(0, 0, 0, 0.05)',
            boxShadowSecondary: '0 4px 12px rgba(0, 0, 0, 0.08)',
          },
          components: {
            Button: {
              primaryColor: '#FFFFFF',
              colorPrimaryHover: '#4CAF50',
              colorPrimaryActive: '#43A047',
            },
            Table: {
              headerBg: '#FAFAFA',
              rowHoverBg: '#E8F5E9',
              borderColor: '#EEEEEE',
            },
            Card: {
              colorBgContainer: '#FFFFFF',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            },
            Input: {
              activeBorderColor: '#66BB6A',
              hoverBorderColor: '#81C784',
            },
            Menu: {
              itemSelectedBg: '#E8F5E9',
              itemSelectedColor: '#4CAF50',
              itemHoverBg: '#F5F5F5',
            },
          },
        }}
      >
        <AuthProvider>
          <App />
        </AuthProvider>
      </ConfigProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
