# Database Setup Scripts

This directory contains scripts to set up the PostgreSQL database for the Hack25 backend application.

## Prerequisites

- PostgreSQL installed and running
- `sudo` access to run commands as the postgres user

## Scripts

1. `01_create_database.sql`: Creates the database and user
2. `02_create_users_table.sql`: Creates the users table
3. `03_create_code_analysis_tables.sql`: Creates tables for code analysis
4. `04_seed_data.sql`: Seeds initial data (admin and regular users)
5. `setup_database.sh`: Main script to run all SQL scripts

## Usage

To set up the database, run:

```bash
./setup_database.sh
```

This script will:
1. Check if PostgreSQL is installed
2. Run all SQL scripts in order
3. Create or update the `.env` file with database credentials

## Default Credentials

The setup creates the following:

- Database: `code_analyser`
- Database User: `code_analyser_user`
- Database Password: `code_analyser_password`

## Initial Users

The seed script creates two users:

1. Admin User:
   - Email: `admin@example.com`
   - Password: `admin123`
   - Role: `admin`

2. Regular User:
   - Email: `user@example.com`
   - Password: `user123`
   - Role: `user`

## Database Schema

### Users Table
- `id`: UUID primary key
- `email`: Unique email address
- `password`: Bcrypt hashed password
- `first_name`: User's first name
- `last_name`: User's last name
- `active`: Boolean indicating if the user is active
- `role`: User role (e.g., "user", "admin")
- `created_at`: Timestamp when the user was created
- `updated_at`: Timestamp when the user was last updated
- `deleted_at`: Soft delete timestamp

### Code Analysis Tables
- `repositories`: Stores information about GitHub repositories
- `files`: Stores information about files in repositories
- `dependencies`: Stores dependencies for each file
- `global_vars`: Stores global variables for each file
- `constants`: Stores constants for each file
- `init_functions`: Stores init functions for each file
- `structs`: Stores structs for each file
- `struct_fields`: Stores fields for each struct
- `methods`: Stores methods for each file
- `method_params`: Stores parameters for each method
- `workflow_steps`: Stores workflow steps for each file
- `workflow_step_dependencies`: Stores dependencies for each workflow step
- `workflow_step_variables`: Stores variables for each workflow step

## Troubleshooting

If you encounter any issues:

1. Make sure PostgreSQL is installed and running
2. Check if the postgres user has the necessary permissions
3. Check if the database already exists
4. Check if the user already exists

For permission issues, you may need to modify the PostgreSQL configuration file (`pg_hba.conf`) to allow local connections with password authentication.
