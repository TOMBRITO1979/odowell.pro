#!/bin/bash

# Script para configurar permissões de usuário para testes

USER_EMAIL="$1"
PERMISSIONS_PRESET="$2"

if [ -z "$USER_EMAIL" ] || [ -z "$PERMISSIONS_PRESET" ]; then
    echo "Uso: $0 <email> <preset>"
    echo ""
    echo "Presets disponíveis:"
    echo "  all        - Todas as permissões (view, create, edit, delete) em todos os módulos"
    echo "  readonly   - Apenas view em todos os módulos"
    echo "  reception  - Permissões típicas de recepcionista (agenda, pacientes, pagamentos)"
    echo "  dentist    - Permissões típicas de dentista (prontuários, exames, prescrições)"
    echo "  financial  - Apenas módulos financeiros"
    echo "  clear      - Remove todas as permissões"
    exit 1
fi

# Get user ID
USER_ID=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "
SELECT id FROM public.users WHERE email = '$USER_EMAIL' AND deleted_at IS NULL;
" | tr -d ' ')

if [ -z "$USER_ID" ]; then
    echo "Usuário não encontrado: $USER_EMAIL"
    exit 1
fi

echo "Usuário encontrado: $USER_EMAIL (ID: $USER_ID)"

# Clear existing permissions
echo "Removendo permissões existentes..."
docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "
UPDATE public.user_permissions
SET deleted_at = NOW()
WHERE user_id = $USER_ID AND deleted_at IS NULL;
"

# Define permission sets based on preset
case "$PERMISSIONS_PRESET" in
    "all")
        MODULES="appointments budgets campaigns dashboard exams medical_records patients payments prescriptions products reports stock_movements suppliers"
        ACTIONS="view create edit delete"
        ;;
    "readonly")
        MODULES="appointments budgets campaigns dashboard exams medical_records patients payments prescriptions products reports stock_movements suppliers"
        ACTIONS="view"
        ;;
    "reception")
        MODULES="appointments patients budgets payments dashboard"
        ACTIONS="view create edit delete"
        ;;
    "dentist")
        MODULES="medical_records exams prescriptions patients dashboard"
        ACTIONS="view create edit delete"
        ;;
    "financial")
        MODULES="budgets payments dashboard"
        ACTIONS="view create edit delete"
        ;;
    "clear")
        echo "Permissões removidas com sucesso!"
        exit 0
        ;;
    *)
        echo "Preset inválido: $PERMISSIONS_PRESET"
        exit 1
        ;;
esac

# Add permissions
echo "Adicionando permissões: $PERMISSIONS_PRESET"
for MODULE in $MODULES; do
    for ACTION in $ACTIONS; do
        # Get permission ID
        PERM_ID=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "
        SELECT p.id
        FROM public.permissions p
        INNER JOIN public.modules m ON m.id = p.module_id
        WHERE m.code = '$MODULE' AND p.action = '$ACTION'
          AND p.deleted_at IS NULL AND m.deleted_at IS NULL
        LIMIT 1;
        " | tr -d ' ')

        if [ -n "$PERM_ID" ]; then
            docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "
            INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
            VALUES ($USER_ID, $PERM_ID, 4, NOW(), NOW())
            ON CONFLICT (user_id, permission_id) DO UPDATE SET deleted_at = NULL;
            " > /dev/null 2>&1
            echo "  ✓ $MODULE.$ACTION"
        fi
    done
done

echo ""
echo "Permissões configuradas com sucesso!"
echo "Faça logout e login novamente para aplicar as mudanças."
echo ""
echo "Para verificar as permissões, execute:"
echo "  docker exec \$(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c \\"
echo "    SELECT m.code, p.action FROM public.user_permissions up"
echo "    INNER JOIN public.permissions p ON p.id = up.permission_id"
echo "    INNER JOIN public.modules m ON m.id = p.module_id"
echo "    WHERE up.user_id = $USER_ID AND up.deleted_at IS NULL"
echo "    ORDER BY m.code, p.action;\\""
