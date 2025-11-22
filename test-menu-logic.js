// Script de teste para simular a lógica do menu

// Simular permissões de Maria (do banco de dados)
const mariaPermissions = {
  appointments: { view: true, create: true, edit: true, delete: true },
  budgets: { view: true, create: true, edit: true, delete: true },
  dashboard: { view: true },
  exams: { view: true },
  medical_records: { view: true },
  patients: { view: true, create: true, edit: true, delete: true },
  payments: { view: true, create: true, edit: true, delete: true }
};

const user = {
  name: 'Maria',
  role: 'receptionist'
};

// Funções do usePermission
const hasPermission = (module, action) => {
  if (user?.role === 'admin') {
    return true;
  }
  return mariaPermissions?.[module]?.[action] === true;
};

const hasAnyPermission = (module) => {
  if (user?.role === 'admin') {
    return true;
  }
  const modulePerms = mariaPermissions?.[module];
  if (!modulePerms) return false;
  return Object.values(modulePerms).some(perm => perm === true);
};

const canView = (module) => {
  return hasPermission(module, 'view') || hasAnyPermission(module);
};

const isAdmin = user?.role === 'admin';

// Menu items (do DashboardLayout.jsx)
const allMenuItems = [
  {
    key: '/',
    label: 'Dashboard',
    permission: 'dashboard',
  },
  {
    key: '/appointments',
    label: 'Agenda',
    permission: 'appointments',
  },
  {
    key: '/patients',
    label: 'Pacientes',
    permission: 'patients',
  },
  {
    key: '/medical-records',
    label: 'Prontuários',
    permission: 'medical_records',
  },
  {
    key: '/prescriptions',
    label: 'Receituário',
    permission: 'prescriptions',
  },
  {
    key: '/exams',
    label: 'Exames',
    permission: 'exams',
  },
  {
    key: 'financial',
    label: 'Financeiro',
    children: [
      { key: '/budgets', label: 'Orçamentos', permission: 'budgets' },
      { key: '/payments', label: 'Pagamentos', permission: 'payments' },
    ],
  },
  {
    key: 'inventory',
    label: 'Estoque',
    children: [
      { key: '/products', label: 'Produtos', permission: 'products' },
      { key: '/suppliers', label: 'Fornecedores', permission: 'suppliers' },
      { key: '/stock-movements', label: 'Movimentações', permission: 'stock_movements' },
    ],
  },
  {
    key: '/campaigns',
    label: 'Campanhas',
    permission: 'campaigns',
  },
  {
    key: '/reports',
    label: 'Relatórios',
    permission: 'reports',
  },
];

// Lógica de filtragem (NOVA - após correção)
const menuItems = allMenuItems.filter(item => {
  // Admin-only items
  if (item.adminOnly) return isAdmin;

  // Items with children - filter children first, then show parent only if any child is accessible
  if (item.children) {
    item.children = item.children.filter(child => {
      return child.permission ? canView(child.permission) : true;
    });
    return item.children.length > 0;
  }

  // Items with permission - check if user can view
  if (item.permission) {
    return canView(item.permission);
  }

  // Items without permission and without children are always visible (none currently)
  return true;
});

// Exibir resultados
console.log('='.repeat(60));
console.log('TESTE DE LÓGICA DO MENU - Usuário Maria');
console.log('='.repeat(60));
console.log('\nPermissões de Maria:');
console.log(JSON.stringify(mariaPermissions, null, 2));
console.log('\n' + '='.repeat(60));
console.log('Verificação de canView para cada módulo:');
console.log('='.repeat(60));

const modulesToTest = [
  'dashboard', 'appointments', 'patients', 'medical_records',
  'prescriptions', 'exams', 'budgets', 'payments',
  'products', 'suppliers', 'stock_movements', 'campaigns', 'reports'
];

modulesToTest.forEach(module => {
  const result = canView(module);
  console.log(`canView('${module}'): ${result}`);
});

console.log('\n' + '='.repeat(60));
console.log('Menu Items Filtrados (o que Maria DEVERIA ver):');
console.log('='.repeat(60));

menuItems.forEach(item => {
  if (item.children) {
    console.log(`✓ ${item.label}`);
    item.children.forEach(child => {
      console.log(`  ✓ ${child.label}`);
    });
  } else {
    console.log(`✓ ${item.label}`);
  }
});

console.log('\n' + '='.repeat(60));
console.log(`Total de itens no menu: ${menuItems.length}`);
console.log('='.repeat(60));
