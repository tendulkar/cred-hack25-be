#!/bin/bash
set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Code Analyser Backend Kubernetes Deployment${NC}"
echo "-------------------------------------"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed. Please install kubectl and try again.${NC}"
    exit 1
fi

# Check if the current kubectl context is correct
CURRENT_CONTEXT=$(kubectl config current-context)
echo -e "Current Kubernetes context: ${GREEN}${CURRENT_CONTEXT}${NC}"
read -p "Continue with this context? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment aborted."
    exit 1
fi


# Update the secrets
echo -e "\n${YELLOW}Setting up secrets...${NC}"

# 1. Check if secret code-analyser-db-credentials exists
if kubectl get secret code-analyser-db-credentials &> /dev/null; then
    echo "Secret code-analyser-db-credentials already exists. Do you want to update it?"
    read -p "(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Read DB credentials
        read -p "DB Host (default: postgres-service): " DB_HOST
        DB_HOST=${DB_HOST:-postgres-service}
        read -p "DB Port (default: 5432): " DB_PORT
        DB_PORT=${DB_PORT:-5432}
        read -p "DB Username (default: code_analyser_user): " DB_USERNAME
        DB_USERNAME=${DB_USERNAME:-code_analyser_user}
        read -p "DB Password (default: code_analyser_password): " DB_PASSWORD
        DB_PASSWORD=${DB_PASSWORD:-code_analyser_password}
        read -p "DB Name (default: code_analyser): " DB_NAME
        DB_NAME=${DB_NAME:-code_analyser}
        
        # Update DB credentials secret
        kubectl create secret generic code-analyser-db-credentials \
            --from-literal=host=${DB_HOST} \
            --from-literal=port=${DB_PORT} \
            --from-literal=username=${DB_USERNAME} \
            --from-literal=password=${DB_PASSWORD} \
            --from-literal=dbname=${DB_NAME} \
            \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
else
    # Read DB credentials
    read -p "DB Host (default: postgres-service): " DB_HOST
    DB_HOST=${DB_HOST:-postgres-service}
    read -p "DB Port (default: 5432): " DB_PORT
    DB_PORT=${DB_PORT:-5432}
    read -p "DB Username (default: code_analyser_user): " DB_USERNAME
    DB_USERNAME=${DB_USERNAME:-code_analyser_user}
    read -p "DB Password (default: code_analyser_password): " DB_PASSWORD
    DB_PASSWORD=${DB_PASSWORD:-code_analyser_password}
    read -p "DB Name (default: code_analyser): " DB_NAME
    DB_NAME=${DB_NAME:-code_analyser}
    
    # Create DB credentials secret
    kubectl create secret generic code-analyser-db-credentials \
        --from-literal=host=${DB_HOST} \
        --from-literal=port=${DB_PORT} \
        --from-literal=username=${DB_USERNAME} \
        --from-literal=password=${DB_PASSWORD} \
        --from-literal=dbname=${DB_NAME} \
       
fi

# 2. Check if secret code-analyser-api-secrets exists
if kubectl get secret code-analyser-api-secrets &> /dev/null; then
    echo "Secret code-analyser-api-secrets already exists. Do you want to update it?"
    read -p "(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Read API secrets
        read -p "JWT Secret (default: your-secret-key-change-in-production): " JWT_SECRET
        JWT_SECRET=${JWT_SECRET:-your-secret-key-change-in-production}
        read -p "OpenAI API Key: " OPENAI_API_KEY
        read -p "Gemini API Key: " GEMINI_API_KEY
        read -p "Sonnet API Key: " SONNET_API_KEY
        
        # Update API secrets
        kubectl create secret generic code-analyser-api-secrets \
            --from-literal=jwt-secret=${JWT_SECRET} \
            --from-literal=openai-api-key=${OPENAI_API_KEY} \
            --from-literal=gemini-api-key=${GEMINI_API_KEY} \
            --from-literal=sonnet-api-key=${SONNET_API_KEY} \
            \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
else
    # Read API secrets
    read -p "JWT Secret (default: your-secret-key-change-in-production): " JWT_SECRET
    JWT_SECRET=${JWT_SECRET:-your-secret-key-change-in-production}
    read -p "OpenAI API Key: " OPENAI_API_KEY
    read -p "Gemini API Key: " GEMINI_API_KEY
    read -p "Sonnet API Key: " SONNET_API_KEY
    
    # Create API secrets
    kubectl create secret generic code-analyser-api-secrets \
        --from-literal=jwt-secret=${JWT_SECRET} \
        --from-literal=openai-api-key=${OPENAI_API_KEY} \
        --from-literal=gemini-api-key=${GEMINI_API_KEY} \
        --from-literal=sonnet-api-key=${SONNET_API_KEY} \
       
fi

# Deploy PostgreSQL if needed
echo -e "\n${YELLOW}Do you want to deploy PostgreSQL to Kubernetes?${NC}"
echo "Note: If you already have a PostgreSQL instance, you can skip this step"
read -p "(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Deploying PostgreSQL...${NC}"
    kubectl apply -f k8s/postgres-deployment.yaml
fi

# Deploy API server
echo -e "\n${YELLOW}Deploying API server...${NC}"
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/api-service.yaml

# Wait for deployments to be ready
echo -e "\n${YELLOW}Waiting for deployments to be ready...${NC}"
kubectl rollout status deployment/code-analyser-be

# Print summary
echo -e "\n${GREEN}Deployment completed successfully!${NC}"
echo -e "\n${YELLOW}API Server Status:${NC}"
kubectl get pods -l app=code-analyser-be
echo -e "\n${YELLOW}API Service:${NC}"
kubectl get service code-analyser-be

echo -e "\n${GREEN}You can access the API at:${NC}"
echo "- Inside the cluster: http://code-analyser-be.${NAMESPACE}.svc.cluster.local"

echo -e "\n${YELLOW}To check logs:${NC}"
echo "kubectl logs -f deployment/code-analyser-be"

echo -e "\n${YELLOW}To port-forward the service for local access:${NC}"
echo "kubectl port-forward service/code-analyser-be 8080:80"
echo "Then access the API at: http://localhost:8080"
