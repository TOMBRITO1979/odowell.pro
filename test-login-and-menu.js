#!/usr/bin/env node

// Script automatizado para testar login e verificar JWT/menu

const https = require('https');

// Configuração
const API_URL = 'api.dr.crwell.pro';
const USER_EMAIL = 'maria@gmail.com';
const USER_PASSWORD = 'senha123';

// Decode JWT payload
function decodeJWT(token) {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) throw new Error('Invalid JWT format');
    const payload = Buffer.from(parts[1], 'base64').toString('utf8');
    return JSON.parse(payload);
  } catch (error) {
    console.error('Error decoding JWT:', error.message);
    return null;
  }
}

// Simular as funções do frontend
function hasPermission(permissions, module, action, role) {
  if (role === 'admin') return true;
  return permissions?.[module]?.[action] === true;
}

function hasAnyPermission(permissions, module, role) {
  if (role === 'admin') return true;
  const modulePerms = permissions?.[module];
  if (!modulePerms) return false;
  return Object.values(modulePerms).some(perm => perm === true);
}

function canView(permissions, module, role) {
  return hasPermission(permissions, module, 'view', role) || hasAnyPermission(permissions, module, role);
}

// Menu items (do DashboardLayout.jsx)
function getMenuItems(permissions, role) {
  const isAdmin = role === 'admin';

  const allMenuItems = [
    { key: '/', label: 'Dashboard', permission: 'dashboard' },
    { key: '/appointments', label: 'Agenda', permission: 'appointments' },
    { key: '/patients', label: 'Pacientes', permission: 'patients' },
    { key: '/medical-records', label: 'Prontuários', permission: 'medical_records' },
    { key: '/prescriptions', label: 'Receituário', permission: 'prescriptions' },
    { key: '/exams', label: 'Exames', permission: 'exams' },
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
    { key: '/campaigns', label: 'Campanhas', permission: 'campaigns' },
    { key: '/reports', label: 'Relatórios', permission: 'reports' },
  ];

  const menuItems = allMenuItems.filter(item => {
    if (item.adminOnly) return isAdmin;

    if (item.children) {
      item.children = item.children.filter(child => {
        return child.permission ? canView(permissions, child.permission, role) : true;
      });
      return item.children.length > 0;
    }

    if (item.permission) {
      return canView(permissions, item.permission, role);
    }

    return true;
  });

  return menuItems;
}

// Fazer login via API
function login() {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({
      email: USER_EMAIL,
      password: USER_PASSWORD
    });

    const options = {
      hostname: API_URL,
      port: 443,
      path: '/api/auth/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        if (res.statusCode === 200) {
          try {
            const response = JSON.parse(data);
            resolve(response);
          } catch (error) {
            reject(new Error('Failed to parse response: ' + error.message));
          }
        } else {
          reject(new Error(`Login failed with status ${res.statusCode}: ${data}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.write(postData);
    req.end();
  });
}

// Exibir menu formatado
function displayMenu(menuItems, indent = '') {
  menuItems.forEach(item => {
    if (item.children && item.children.length > 0) {
      console.log(`${indent}✓ ${item.label}`);
      item.children.forEach(child => {
        console.log(`${indent}  ✓ ${child.label}`);
      });
    } else {
      console.log(`${indent}✓ ${item.label}`);
    }
  });
}

// Main
async function main() {
  console.log('='.repeat(70));
  console.log('TESTE AUTOMATIZADO - Login e Verificação de Menu');
  console.log('='.repeat(70));
  console.log(`\nUsuário: ${USER_EMAIL}`);
  console.log('Fazendo login...\n');

  try {
    const response = await login();

    console.log('✓ Login bem-sucedido!\n');
    console.log('='.repeat(70));
    console.log('INFORMAÇÕES DO USUÁRIO');
    console.log('='.repeat(70));
    console.log(`Nome: ${response.user.name}`);
    console.log(`Email: ${response.user.email}`);
    console.log(`Role: ${response.user.role}`);
    console.log(`Tenant: ${response.tenant.name}`);

    // Decodificar JWT
    const decoded = decodeJWT(response.token);

    if (!decoded) {
      console.error('\n❌ Erro ao decodificar JWT');
      return;
    }

    console.log('\n' + '='.repeat(70));
    console.log('PERMISSÕES NO JWT');
    console.log('='.repeat(70));

    const permissions = decoded.permissions || {};
    const permissionCount = Object.keys(permissions).length;

    if (permissionCount === 0) {
      console.log('⚠️  Nenhuma permissão encontrada no JWT!');
    } else {
      console.log(`Total de módulos com permissões: ${permissionCount}\n`);

      Object.keys(permissions).sort().forEach(module => {
        const actions = Object.keys(permissions[module]).filter(action => permissions[module][action]);
        if (actions.length > 0) {
          console.log(`  ${module}: ${actions.join(', ')}`);
        }
      });
    }

    // Calcular menu esperado
    console.log('\n' + '='.repeat(70));
    console.log('ITENS DO MENU (baseado nas permissões)');
    console.log('='.repeat(70));

    const menuItems = getMenuItems(permissions, decoded.role);

    if (menuItems.length === 0) {
      console.log('⚠️  Nenhum item de menu visível!');
    } else {
      console.log(`Total de itens no menu: ${menuItems.length}\n`);
      displayMenu(menuItems);
    }

    console.log('\n' + '='.repeat(70));
    console.log('✓ TESTE CONCLUÍDO');
    console.log('='.repeat(70));

  } catch (error) {
    console.error('\n❌ Erro:', error.message);
    process.exit(1);
  }
}

main();
