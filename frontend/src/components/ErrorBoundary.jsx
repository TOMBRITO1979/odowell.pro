import React from 'react';
import { Result, Button } from 'antd';

/**
 * ErrorBoundary component to catch JavaScript errors in child components
 * Prevents the entire app from crashing when a component error occurs
 */
class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    // Log error information for debugging
    console.error('ErrorBoundary caught an error:', error);
    console.error('Error info:', errorInfo);
    this.setState({ errorInfo });
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
          padding: '20px'
        }}>
          <Result
            status="error"
            title="Ocorreu um erro inesperado"
            subTitle="Desculpe, algo deu errado. Por favor, tente recarregar a página ou voltar para o início."
            extra={[
              <Button type="primary" key="reload" onClick={this.handleReload}>
                Recarregar Página
              </Button>,
              <Button key="home" onClick={this.handleGoHome}>
                Voltar ao Início
              </Button>,
            ]}
          />
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
