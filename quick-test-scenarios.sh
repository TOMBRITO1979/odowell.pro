#!/bin/bash

# Quick test scenarios for Maria

echo "Escolha o cenário:"
echo "1) Dashboard apenas"
echo "2) Dashboard + Agenda"
echo "3) Financeiro (Dashboard + Orçamentos + Pagamentos)"
echo "4) Recepcionista completo"
echo "5) Estoque (Dashboard + Produtos + Fornecedores + Movimentações)"
echo "6) Todos os módulos (readonly)"

read -p "Opção: " OPTION

POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres)

case $OPTION in
  1)
    echo "Configurando: Dashboard apenas..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
    WHERE m.code = 'dashboard' AND p.action = 'view';
    "
    ;;
  2)
    echo "Configurando: Dashboard + Agenda..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
    WHERE (m.code = 'dashboard' AND p.action = 'view')
       OR (m.code = 'appointments' AND p.action IN ('view', 'create', 'edit', 'delete'));
    "
    ;;
  3)
    echo "Configurando: Financeiro..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
    WHERE (m.code = 'dashboard' AND p.action = 'view')
       OR (m.code IN ('budgets', 'payments') AND p.action IN ('view', 'create', 'edit', 'delete'));
    "
    ;;
  4)
    echo "Configurando: Recepcionista completo..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
    WHERE (m.code = 'dashboard' AND p.action = 'view')
       OR (m.code IN ('appointments', 'patients', 'budgets', 'payments') AND p.action IN ('view', 'create', 'edit', 'delete'))
       OR (m.code IN ('exams', 'medical_records') AND p.action = 'view');
    "
    ;;
  5)
    echo "Configurando: Estoque..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
    WHERE (m.code = 'dashboard' AND p.action = 'view')
       OR (m.code IN ('products', 'suppliers', 'stock_movements') AND p.action IN ('view', 'create', 'edit', 'delete'));
    "
    ;;
  6)
    echo "Configurando: Todos os módulos (readonly)..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
    DELETE FROM public.user_permissions WHERE user_id = 7;
    INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
    SELECT 7, p.id, 4, NOW(), NOW()
    FROM public.permissions p
    WHERE p.action = 'view' AND p.deleted_at IS NULL;
    "
    ;;
  *)
    echo "Opção inválida"
    exit 1
    ;;
esac

echo ""
echo "✓ Permissões configuradas!"
echo "Faça logout e login novamente com maria@gmail.com"
echo ""
echo "Permissões atuais:"
docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
SELECT m.code, p.action
FROM public.user_permissions up
INNER JOIN public.permissions p ON p.id = up.permission_id
INNER JOIN public.modules m ON m.id = p.module_id
WHERE up.user_id = 7 AND up.deleted_at IS NULL
ORDER BY m.code, p.action;
"
