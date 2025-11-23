#!/bin/bash

# Create a test user for baseline tests
# Password: testpassword123

set -e

echo "Creating test user for baseline tests..."

# Get PostgreSQL container ID
POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres)

if [ -z "$POSTGRES_CONTAINER" ]; then
    echo "Error: PostgreSQL container not found"
    exit 1
fi

# Create test user with known password (bcrypt hash of "testpassword123")
# Generated using Python bcrypt
HASHED_PASSWORD='$2b$12$xfHf5LqexRkGjKzKwFS8xOpN2eHKe4SUsSjVY93DDn3.KRfeIbf5i'

# Check if test user already exists
EXISTING=$(docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -t -c "SELECT email FROM public.users WHERE email = 'test@baseline.com';" | xargs)

if [ "$EXISTING" = "test@baseline.com" ]; then
    echo "Test user already exists. Updating password..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
        UPDATE public.users
        SET password = '$HASHED_PASSWORD',
            role = 'admin',
            active = true
        WHERE email = 'test@baseline.com';
    "
else
    echo "Creating new test user..."
    docker exec $POSTGRES_CONTAINER psql -U drcrwell_user -d drcrwell_db -c "
        INSERT INTO public.users (tenant_id, email, password, name, role, active, created_at, updated_at)
        VALUES (1, 'test@baseline.com', '$HASHED_PASSWORD', 'Baseline Test User', 'admin', true, NOW(), NOW());
    "
fi

echo "✓ Test user created/updated successfully"
echo "  Email: test@baseline.com"
echo "  Password: testpassword123"
echo "  Role: admin"
echo "  Tenant: 1"
