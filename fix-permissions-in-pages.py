#!/usr/bin/env python3

import os
import re

# Mapa de páginas e seus códigos de módulo
PAGE_MODULE_MAP = {
    'Appointments.jsx': 'appointments',
    'Budgets.jsx': 'budgets',
    'Payments.jsx': 'payments',
    'Products.jsx': 'products',
    'Suppliers.jsx': 'suppliers',
    'Campaigns.jsx': 'campaigns',
    'Exams.jsx': 'exams',
    'MedicalRecords.jsx': 'medical_records',
    'Prescriptions.jsx': 'prescriptions',
}

def fix_permissions_in_file(filepath, module_code):
    print(f"Processando: {filepath}")

    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.content()

    original_content = content

    # 1. Adicionar import do usePermission se não existir
    if 'usePermission' not in content:
        # Encontrar a linha do último import
        import_lines = []
        for i, line in enumerate(content.split('\n')):
            if line.startswith('import '):
                import_lines.append(i)

        if import_lines:
            lines = content.split('\n')
            # Verificar se já tem import do AuthContext
            has_auth_import = any('AuthContext' in line for line in lines)

            if not has_auth_import:
                # Adicionar novo import
                last_import_idx = import_lines[-1]
                lines.insert(last_import_idx + 1, "import { usePermission } from '../../contexts/AuthContext';")
            else:
                # Modificar import existente
                for i, line in enumerate(lines):
                    if 'AuthContext' in line and 'usePermission' not in line:
                        # Adicionar usePermission ao import existente
                        if 'useAuth' in line:
                            lines[i] = line.replace('useAuth', 'useAuth, usePermission')
                        else:
                            lines[i] = line.replace("from '../../contexts/AuthContext'", "usePermission } from '../../contexts/AuthContext'")
                            lines[i] = line.replace('import {', 'import { usePermission,')

            content = '\n'.join(lines)

    # 2. Adicionar const { canCreate, canEdit, canDelete } = usePermission(); se não existir
    if 'usePermission()' not in content:
        # Encontrar a linha do componente
        component_match = re.search(r'(const \w+ = \(\) => \{)', content)
        if component_match:
            # Adicionar após as declarações de useState
            lines = content.split('\n')
            navigate_idx = -1
            for i, line in enumerate(lines):
                if 'useNavigate()' in line:
                    navigate_idx = i
                    break

            if navigate_idx > 0:
                lines.insert(navigate_idx + 1, "  const { canCreate, canEdit, canDelete } = usePermission();")
                content = '\n'.join(lines)

    # 3. Adicionar verificação de permissão aos botões de criar
    # Padrão: <Button type="primary" icon={<PlusOutlined />} onClick=...>
    create_button_pattern = r'(<Button[^>]*type="primary"[^>]*icon=\{<PlusOutlined[^>]+>[\s\S]*?</Button>)'

    def wrap_with_permission(match, permission_type):
        button = match.group(1)
        if f"can{permission_type.capitalize()}('{module_code}')" not in button:
            indent = '        '  # Ajustar indentação
            return f"{{{permission_type}('{module_code}') && (\n{indent}{button}\n{indent})}}"
        return button

    content = re.sub(create_button_pattern, lambda m: wrap_with_permission(m, 'canCreate'), content, flags=re.MULTILINE)

    # 4. Adicionar verificação aos botões de editar
    edit_button_pattern = r'(<Button[^>]*icon=\{<EditOutlined[^}]+\}[^>]*>[\s\S]*?</Button>|<Button[^>]*icon=\{<EditOutlined[^/]+/>)'

    matches = list(re.finditer(edit_button_pattern, content))
    for match in reversed(matches):  # Reverso para não afetar índices
        button = match.group(0)
        if f"canEdit('{module_code}')" not in button and 'canEdit' not in content[max(0, match.start()-100):match.start()]:
            start, end = match.span()
            indent = '          '
            replacement = f"{{canEdit('{module_code}') && (\n{indent}{button}\n{indent})}}"
            content = content[:start] + replacement + content[end:]

    # 5. Adicionar verificação aos botões/popconfirm de deletar
    delete_pattern = r'(<Popconfirm[\s\S]*?<Button[^>]*icon=\{<DeleteOutlined[^>]+>[\s\S]*?</Popconfirm>)'

    matches = list(re.finditer(delete_pattern, content))
    for match in reversed(matches):
        popconfirm = match.group(0)
        if f"canDelete('{module_code}')" not in popconfirm and 'canDelete' not in content[max(0, match.start()-100):match.start()]:
            start, end = match.span()
            indent = '          '
            replacement = f"{{canDelete('{module_code}') && (\n{indent}{popconfirm}\n{indent})}}"
            content = content[:start] + replacement + content[end:]

    # Salvar apenas se houver mudanças
    if content != original_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"  ✓ Arquivo atualizado: {filepath}")
        return True
    else:
        print(f"  - Sem mudanças necessárias: {filepath}")
        return False

def main():
    frontend_pages = '/root/drcrwell/frontend/src/pages'
    files_updated = 0

    print("="*70)
    print("Corrigindo verificação de permissões em todas as páginas")
    print("="*70)
    print()

    for filename, module_code in PAGE_MODULE_MAP.items():
        # Encontrar o arquivo
        for root, dirs, files in os.walk(frontend_pages):
            if filename in files:
                filepath = os.path.join(root, filename)
                if fix_permissions_in_file(filepath, module_code):
                    files_updated += 1
                break

    print()
    print("="*70)
    print(f"Concluído! {files_updated} arquivos atualizados.")
    print("="*70)

if __name__ == '__main__':
    main()
