-- Create the database user if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'code_analyser_user') THEN
        CREATE USER code_analyser_user WITH PASSWORD 'code_analyser_password';
    END IF;
END
$$;

-- Create the database if it doesn't exist
CREATE DATABASE code_analyser WITH OWNER = code_analyser_user;

-- Grant privileges to the user
GRANT CONNECT ON DATABASE code_analyser TO code_analyser_user;
GRANT ALL PRIVILEGES ON DATABASE code_analyser TO code_analyser_user;

-- Connect to the database
\c code_analyser

-- Create the UUID extension if it doesn't exist
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
