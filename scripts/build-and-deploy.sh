#!/bin/bash
set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
AWS_REGION="ap-south-1"
AWS_ACCOUNT_ID="109869387501"
ECR_REPOSITORY="code-analyser-be"
IMAGE_TAG="latest"

echo -e "${YELLOW}Code Analyser Backend - Build and Deploy${NC}"
echo "=========================================="

# Function to check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"
PREREQS_MET=true

if ! command_exists docker; then
    echo -e "${RED}Error: Docker is not installed.${NC}"
    PREREQS_MET=false
fi

if ! command_exists aws; then
    echo -e "${RED}Error: AWS CLI is not installed.${NC}"
    PREREQS_MET=false
fi

if ! command_exists kubectl; then
    echo -e "${RED}Error: kubectl is not installed.${NC}"
    PREREQS_MET=false
fi

if [ "$PREREQS_MET" = false ]; then
    echo -e "${RED}Please install the missing prerequisites and try again.${NC}"
    exit 1
fi

# Check AWS credentials
echo -e "\n${YELLOW}Checking AWS credentials...${NC}"
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}Error: AWS credentials not configured or invalid.${NC}"
    echo "Please run 'aws configure' to set up your AWS credentials."
    exit 1
fi
echo -e "${GREEN}AWS credentials are valid.${NC}"

# Check kubectl context
CURRENT_CONTEXT=$(kubectl config current-context)
echo -e "\n${YELLOW}Current Kubernetes context: ${GREEN}${CURRENT_CONTEXT}${NC}"
echo "This script will build and deploy the code-analyser-be application to this context."
# read -p "Continue with this context? (y/n) " -n 1 -r
# echo
# if [[ ! $REPLY =~ ^[Yy]$ ]]; then
#     echo "Build and deployment aborted."
#     exit 1
# fi

# STEP 1: Build Docker image
echo -e "\n${YELLOW}Step 1: Building Docker image...${NC}"

# Full image name
ECR_REPOSITORY_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}"

echo "Building Docker image for API server..."
docker build --platform linux/amd64 -t ${ECR_REPOSITORY}:${IMAGE_TAG} -f Dockerfile.api .

echo "Tagging image for ECR..."
docker tag ${ECR_REPOSITORY}:${IMAGE_TAG} ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

# STEP 2: Push to ECR
echo -e "\n${YELLOW}Step 2: Pushing to ECR...${NC}"

echo "Logging in to AWS ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# Check if repository exists, if not create it
aws ecr describe-repositories --repository-names ${ECR_REPOSITORY} --region ${AWS_REGION} || \
    aws ecr create-repository --repository-name ${ECR_REPOSITORY} --region ${AWS_REGION}

echo "Pushing image to ECR..."
docker push ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

echo -e "${GREEN}Image successfully pushed to ${ECR_REPOSITORY_URI}:${IMAGE_TAG}${NC}"

# STEP 4: Set up Kubernetes secrets
echo -e "\n${YELLOW}Step 4: Setting up Kubernetes secrets...${NC}"

# Database credentials
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

# Create or update DB credentials secret
kubectl create secret generic code-analyser-db-credentials \
    --from-literal=host=${DB_HOST} \
    --from-literal=port=${DB_PORT} \
    --from-literal=username=${DB_USERNAME} \
    --from-literal=password=${DB_PASSWORD} \
    --from-literal=dbname=${DB_NAME} \
    --dry-run=client -o yaml | kubectl apply -f -

# API secrets
read -p "JWT Secret (default: your-secret-key-change-in-production): " JWT_SECRET
JWT_SECRET=${JWT_SECRET:-your-secret-key-change-in-production}
read -p "OpenAI API Key: " OPENAI_API_KEY
read -p "Gemini API Key: " GEMINI_API_KEY
read -p "Sonnet API Key: " SONNET_API_KEY

# Create or update API secrets
kubectl create secret generic code-analyser-api-secrets \
    --from-literal=jwt-secret=${JWT_SECRET} \
    --from-literal=openai-api-key=${OPENAI_API_KEY} \
    --from-literal=gemini-api-key=${GEMINI_API_KEY} \
    --from-literal=sonnet-api-key=${SONNET_API_KEY} \
    --dry-run=client -o yaml | kubectl apply -f -

# STEP 5: Deploy PostgreSQL (optional)
echo -e "\n${YELLOW}Step 5: Deploying PostgreSQL (optional)...${NC}"
echo "Do you want to deploy PostgreSQL to Kubernetes?"
echo "Note: If you already have a PostgreSQL instance, you can skip this step"
read -p "(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deploying PostgreSQL..."
    kubectl apply -f k8s/postgres-deployment.yaml 
    
    # Wait for PostgreSQL to be ready
    echo "Waiting for PostgreSQL to be ready..."
    kubectl rollout status deployment/postgres 
fi

# STEP 6: Deploy API server
echo -e "\n${YELLOW}Step 6: Deploying API server...${NC}"
echo "Deploying API server..."
kubectl apply -f k8s/api-deployment.yaml 
kubectl apply -f k8s/api-service.yaml 

# Wait for deployments to be ready
echo "Waiting for API server to be ready..."
kubectl rollout status deployment/code-analyser-be 

# STEP 7: Print summary
echo -e "\n${GREEN}Deployment completed successfully!${NC}"
echo -e "\n${YELLOW}API Server Status:${NC}"
kubectl get pods -l app=code-analyser-be 
echo -e "\n${YELLOW}API Service:${NC}"
kubectl get service code-analyser-be 

echo -e "\n${GREEN}You can access the API at:${NC}"
echo "- Inside the cluster: http://code-analyser-be.code-analyser.svc.cluster.local"

echo -e "\n${YELLOW}To check logs:${NC}"
echo "kubectl logs -f deployment/code-analyser-be "

echo -e "\n${YELLOW}To port-forward the service for local access:${NC}"
echo "kubectl port-forward service/code-analyser-be 8080:80 "
echo "Then access the API at: http://localhost:8080"

echo -e "\n${GREEN}Build and deployment process completed successfully!${NC}"
