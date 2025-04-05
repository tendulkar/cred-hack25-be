#!/bin/bash

# Exit on error
set -e

# Get the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Setting up the database..."

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "PostgreSQL is not installed. Please install PostgreSQL and try again."
    exit 1
fi

# Run the SQL scripts as the postgres user
echo "Creating database and user..."
psql -d postgres -f "$DIR/01_create_database.sql"

echo "Creating users table..."
psql -d postgres -f "$DIR/02_create_users_table.sql"

echo "Creating code analysis tables..."
psql -d postgres -f "$DIR/03_create_code_analysis_tables.sql"

echo "Seeding initial data..."
psql -d postgres -f "$DIR/04_seed_data.sql"

echo "Database setup complete!"

# Update the .env file with the database credentials
if [ -f "$DIR/../../.env" ]; then
    echo "Updating .env file..."
    # Check if the DB_USER and DB_PASSWORD are already set
    if ! grep -q "DB_USER=" "$DIR/../../.env"; then
        echo "DB_USER=code_analyser_user" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_USER=.*/DB_USER=code_analyser_user/' "$DIR/../../.env"
    fi
    
    if ! grep -q "DB_PASSWORD=" "$DIR/../../.env"; then
        echo "DB_PASSWORD=code_analyser_password" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_PASSWORD=.*/DB_PASSWORD=code_analyser_password/' "$DIR/../../.env"
    fi
    
    if ! grep -q "DB_NAME=" "$DIR/../../.env"; then
        echo "DB_NAME=code_analyser" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_NAME=.*/DB_NAME=code_analyser/' "$DIR/../../.env"
    fi
    
    if ! grep -q "DB_HOST=" "$DIR/../../.env"; then
        echo "DB_HOST=localhost" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_HOST=.*/DB_HOST=localhost/' "$DIR/../../.env"
    fi
    
    if ! grep -q "DB_PORT=" "$DIR/../../.env"; then
        echo "DB_PORT=5432" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_PORT=.*/DB_PORT=5432/' "$DIR/../../.env"
    fi
    
    if ! grep -q "DB_SSL_MODE=" "$DIR/../../.env"; then
        echo "DB_SSL_MODE=disable" >> "$DIR/../../.env"
    else
        sed -i '' 's/DB_SSL_MODE=.*/DB_SSL_MODE=disable/' "$DIR/../../.env"
    fi
    
    echo ".env file updated with database credentials."
else
    echo "Creating .env file with database credentials..."
    cat > "$DIR/../../.env" << EOL
DB_USER=code_analyser_user
DB_PASSWORD=code_analyser_password
DB_NAME=code_analyser
DB_HOST=localhost
DB_PORT=5432
DB_SSL_MODE=disable
EOL
    echo ".env file created with database credentials."
fi

echo "You can now run the application with 'go run cmd/api/main.go'"
