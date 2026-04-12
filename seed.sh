#!/bin/bash
# seed.sh - Run database migrations and seed data

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "TaskFlow Database Setup"
echo "========================"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql not found. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '#' | xargs)
else
    echo -e "${RED}Error: .env file not found. Please copy .env.example to .env and configure it.${NC}"
    exit 1
fi

# Parse database URL
# Format: postgresql://user:password@host:port/database
if [[ $DATABASE_URL =~ postgresql://([^:]+):([^@]+)@([^:]+):([^/]+)/(.+) ]]; then
    DB_USER="${BASH_REMATCH[1]}"
    DB_PASS="${BASH_REMATCH[2]}"
    DB_HOST="${BASH_REMATCH[3]}"
    DB_PORT="${BASH_REMATCH[4]}"
    DB_NAME="${BASH_REMATCH[5]}"
else
    echo -e "${RED}Error: Invalid DATABASE_URL format${NC}"
    exit 1
fi

echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# Set password for psql
export PGPASSWORD="$DB_PASS"

# Run migrations
echo -e "${GREEN}Running migrations...${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f migrations/001_init_schema.up.sql

echo -e "${GREEN}Seeding test data...${NC}"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f migrations/002_seed_data.up.sql

echo ""
echo -e "${GREEN}Database setup completed successfully!${NC}"
echo ""
echo "Test credentials:"
echo "  Email: test@example.com"
echo "  Password: password123"
