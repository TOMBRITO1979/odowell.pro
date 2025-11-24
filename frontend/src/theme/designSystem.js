/**
 * OdoWell Design System
 * Sistema centralizado de cores, estilos e padrões visuais
 */

// ============================================
// CORES PRIMÁRIAS (Identidade da Marca)
// ============================================
export const brandColors = {
  primary: '#16a34a',        // Verde principal
  primaryDark: '#15803d',    // Verde escuro (hover, gradientes)
  primaryLight: '#22c55e',   // Verde claro
};

// ============================================
// CORES FUNCIONAIS (Ações)
// ============================================
export const actionColors = {
  create: '#3b82f6',        // Azul suave - Criar/Adicionar
  edit: '#f59e0b',          // Âmbar - Editar/Modificar
  view: '#8b5cf6',          // Roxo suave - Visualizar/Ver
  delete: '#ef4444',        // Vermelho suave - Excluir/Remover
  save: '#10b981',          // Verde esmeralda - Salvar/Confirmar
  cancel: '#6b7280',        // Cinza - Cancelar
  exportPDF: '#dc2626',     // Vermelho PDF - Exportar PDF
  exportExcel: '#059669',   // Verde Excel - Exportar Excel/CSV
  import: '#0284c7',        // Azul céu - Importar
  print: '#4b5563',         // Cinza escuro - Imprimir
  refund: '#f97316',        // Laranja - Estornar
  approve: '#10b981',       // Verde - Aprovar
  reject: '#ef4444',        // Vermelho - Rejeitar
};

// ============================================
// CORES DE STATUS
// ============================================
export const statusColors = {
  pending: '#f59e0b',       // Âmbar - Pendente/Aguardando
  success: '#10b981',       // Verde - Sucesso/Pago/Concluído
  error: '#ef4444',         // Vermelho - Erro/Atrasado
  cancelled: '#6b7280',     // Cinza - Cancelado
  inProgress: '#3b82f6',    // Azul - Em Progresso
  approved: '#10b981',      // Verde - Aprovado
  rejected: '#ef4444',      // Vermelho - Rejeitado
  overdue: '#dc2626',       // Vermelho escuro - Atrasado
  refunded: '#9333ea',      // Roxo - Estornado
};

// ============================================
// ESTILOS DE BOTÕES
// ============================================
export const buttonStyles = {
  // Botão principal (verde da marca)
  primary: {
    backgroundColor: brandColors.primary,
    borderColor: brandColors.primary,
    color: '#fff',
    '&:hover': {
      backgroundColor: brandColors.primaryDark,
      borderColor: brandColors.primaryDark,
    },
  },

  // Ações específicas
  create: {
    backgroundColor: actionColors.create,
    borderColor: actionColors.create,
    color: '#fff',
  },

  edit: {
    backgroundColor: actionColors.edit,
    borderColor: actionColors.edit,
    color: '#fff',
  },

  view: {
    backgroundColor: actionColors.view,
    borderColor: actionColors.view,
    color: '#fff',
  },

  delete: {
    backgroundColor: actionColors.delete,
    borderColor: actionColors.delete,
    color: '#fff',
  },

  save: {
    backgroundColor: actionColors.save,
    borderColor: actionColors.save,
    color: '#fff',
  },

  cancel: {
    backgroundColor: actionColors.cancel,
    borderColor: actionColors.cancel,
    color: '#fff',
  },

  exportPDF: {
    backgroundColor: actionColors.exportPDF,
    borderColor: actionColors.exportPDF,
    color: '#fff',
  },

  exportExcel: {
    backgroundColor: actionColors.exportExcel,
    borderColor: actionColors.exportExcel,
    color: '#fff',
  },

  import: {
    backgroundColor: actionColors.import,
    borderColor: actionColors.import,
    color: '#fff',
  },
};

// ============================================
// TAMANHOS DE BOTÕES
// ============================================
export const buttonSizes = {
  small: 'small',      // 28px - Ações em tabelas
  middle: 'middle',    // 32px - Padrão
  large: 'large',      // 40px - Ações principais
};

// ============================================
// ESPAÇAMENTOS (Padding/Margin)
// ============================================
export const spacing = {
  xs: 8,
  sm: 12,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
};

// ============================================
// BREAKPOINTS (Responsividade)
// ============================================
export const breakpoints = {
  xs: 480,
  sm: 576,
  md: 768,
  lg: 992,
  xl: 1200,
  xxl: 1600,
};

// ============================================
// SOMBRAS
// ============================================
export const shadows = {
  small: '0 1px 3px rgba(0, 0, 0, 0.1)',
  medium: '0 4px 6px rgba(0, 0, 0, 0.1)',
  large: '0 10px 15px rgba(0, 0, 0, 0.1)',
};

// ============================================
// FUNÇÕES AUXILIARES
// ============================================

/**
 * Retorna o estilo inline para um botão de ação específica
 * @param {string} action - Tipo de ação (create, edit, view, delete, etc.)
 * @param {string} size - Tamanho do botão (small, middle, large)
 * @returns {object} Objeto de estilo inline
 */
export const getButtonStyle = (action, size = 'middle') => {
  const baseStyle = buttonStyles[action] || buttonStyles.primary;
  return {
    ...baseStyle,
    size,
  };
};

/**
 * Retorna a cor para um status específico
 * @param {string} status - Status (pending, success, error, etc.)
 * @returns {string} Cor hexadecimal
 */
export const getStatusColor = (status) => {
  return statusColors[status] || statusColors.pending;
};

/**
 * Retorna se estamos em uma tela mobile
 * @returns {boolean}
 */
export const isMobile = () => {
  return window.innerWidth < breakpoints.md;
};

/**
 * Retorna se estamos em uma tela tablet
 * @returns {boolean}
 */
export const isTablet = () => {
  return window.innerWidth >= breakpoints.md && window.innerWidth < breakpoints.lg;
};

/**
 * Retorna se estamos em uma tela desktop
 * @returns {boolean}
 */
export const isDesktop = () => {
  return window.innerWidth >= breakpoints.lg;
};

// ============================================
// COMPONENTES DE BOTÃO PRÉ-CONFIGURADOS
// ============================================

/**
 * Configurações para botões de ação comuns
 */
export const commonButtonConfigs = {
  create: {
    style: buttonStyles.create,
    text: 'Novo',
    icon: 'PlusOutlined',
  },
  edit: {
    style: buttonStyles.edit,
    text: 'Editar',
    icon: 'EditOutlined',
  },
  view: {
    style: buttonStyles.view,
    text: 'Visualizar',
    icon: 'EyeOutlined',
  },
  delete: {
    style: buttonStyles.delete,
    text: 'Excluir',
    icon: 'DeleteOutlined',
  },
  save: {
    style: buttonStyles.save,
    text: 'Salvar',
    icon: 'SaveOutlined',
  },
  cancel: {
    style: buttonStyles.cancel,
    text: 'Cancelar',
    icon: 'CloseOutlined',
  },
  exportPDF: {
    style: buttonStyles.exportPDF,
    text: 'PDF',
    icon: 'FilePdfOutlined',
  },
  exportExcel: {
    style: buttonStyles.exportExcel,
    text: 'Excel',
    icon: 'FileExcelOutlined',
  },
  import: {
    style: buttonStyles.import,
    text: 'Importar',
    icon: 'UploadOutlined',
  },
};

export default {
  brandColors,
  actionColors,
  statusColors,
  buttonStyles,
  buttonSizes,
  spacing,
  breakpoints,
  shadows,
  getButtonStyle,
  getStatusColor,
  isMobile,
  isTablet,
  isDesktop,
  commonButtonConfigs,
};
