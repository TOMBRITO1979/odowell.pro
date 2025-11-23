import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import ptBR from 'antd/locale/pt_BR'
import App from './App'
import { AuthProvider } from './contexts/AuthContext'
import './index.css'
import './styles/mobile.css'
import './styles/mobile-override.css'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <ConfigProvider
        locale={ptBR}
        theme={{
          token: {
            colorPrimary: '#16a34a',
            colorLink: '#16a34a',
            borderRadius: 6,
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
