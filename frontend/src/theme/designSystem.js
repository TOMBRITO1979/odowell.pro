/**
 * OdoWell Design System
 * Sistema centralizado de cores, estilos e padrões visuais
 * Paleta Suave Moderna - Verde Claro #C8E6C9
 */

// ============================================
// CORES PRIMÁRIAS (Verde Claro Suave)
// ============================================
export const brandColors = {
  primary: '#66BB6A',        // Verde suave principal
  primaryDark: '#4CAF50',    // Verde médio (hover)
  primaryLight: '#E8F5E9',   // Verde muito claro (fundo)
  primary50: '#E8F5E9',
  primary100: '#C8E6C9',
  primary200: '#A5D6A7',
  primary300: '#81C784',
  primary400: '#66BB6A',
  primary500: '#4CAF50',
  primary600: '#43A047',
};

// ============================================
// CORES FUNCIONAIS (Ações - Tons Suaves)
// ============================================
export const actionColors = {
  create: '#7986CB',        // Azul índigo suave - Criar/Adicionar
  edit: '#FFB74D',          // Âmbar suave - Editar/Modificar
  view: '#B39DDB',          // Roxo suave - Visualizar/Ver
  delete: '#E57373',        // Vermelho suave - Excluir/Remover
  save: '#81C784',          // Verde suave - Salvar/Confirmar
  cancel: '#90A4AE',        // Cinza azulado - Cancelar
  exportPDF: '#E57373',     // Vermelho suave - Exportar PDF
  exportExcel: '#81C784',   // Verde suave - Exportar Excel/CSV
  import: '#64B5F6',        // Azul suave - Importar
  print: '#78909C',         // Cinza azulado - Imprimir
  refund: '#FFB74D',        // Âmbar suave - Estornar
  approve: '#81C784',       // Verde suave - Aprovar
  reject: '#E57373',        // Vermelho suave - Rejeitar
};

// ============================================
// CORES DE STATUS (Tons Suaves)
// ============================================
export const statusColors = {
  pending: '#FFD54F',       // Âmbar suave - Pendente/Aguardando
  success: '#81C784',       // Verde suave - Sucesso/Pago/Concluído
  error: '#E57373',         // Vermelho suave - Erro/Faltou
  cancelled: '#90A4AE',     // Cinza azulado - Cancelado
  inProgress: '#64B5F6',    // Azul suave - Em Progresso
  approved: '#81C784',      // Verde suave - Aprovado
  rejected: '#E57373',      // Vermelho suave - Rejeitado
  overdue: '#E57373',       // Vermelho suave - Atrasado
  refunded: '#B39DDB',      // Roxo suave - Estornado
  waiting: '#FFD54F',       // Âmbar suave - Aguardando
  scheduled: '#64B5F6',     // Azul suave - Agendado
  completed: '#81C784',     // Verde suave - Concluído
  noShow: '#EF9A9A',        // Vermelho claro - Faltou
};

// ============================================
// CORES DE FUNDO
// ============================================
export const backgroundColors = {
  primary: '#FAFAFA',
  secondary: '#F5F5F5',
  tertiary: '#EEEEEE',
  card: '#FFFFFF',
  hover: '#E8F5E9',
};

// ============================================
// CORES DE TEXTO
// ============================================
export const textColors = {
  primary: '#37474F',
  secondary: '#78909C',
  tertiary: '#90A4AE',
  muted: '#B0BEC5',
  disabled: '#CFD8DC',
  inverse: '#FFFFFF',
  link: '#5C6BC0',
  linkHover: '#3F51B5',
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
