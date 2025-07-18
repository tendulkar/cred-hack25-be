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
NAMESPACE="code-analyser"

echo -e "${YELLOW}Code Analyser Backend - Build and Deploy (Prebuilt Binary)${NC}"
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

if ! command_exists go; then
    echo -e "${RED}Error: Go is not installed.${NC}"
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

# STEP 1: Build the binary locally
echo -e "\n${YELLOW}Step 1: Building Go binary locally...${NC}"
echo "This is faster than building in Docker"

# Build with optimizations for Linux
echo "Building binary for Linux..."
GOOS=linux GOARCH=amd64 go mod tidy && go mod vendor
GOOS=linux GOARCH=amd64 go build -o api-server ./cmd/api/main.go

echo -e "${GREEN}Binary built successfully at ./api-server${NC}"

# STEP 2: Build Docker image
echo -e "\n${YELLOW}Step 2: Building Docker image from prebuilt binary...${NC}"

# Full image name
ECR_REPOSITORY_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}"

echo "Building Docker image for API server..."
docker build -t ${ECR_REPOSITORY}:${IMAGE_TAG} -f Dockerfile.prebuilt .

echo "Tagging image for ECR..."
docker tag ${ECR_REPOSITORY}:${IMAGE_TAG} ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

# STEP 3: Push to ECR
echo -e "\n${YELLOW}Step 3: Pushing to ECR...${NC}"

echo "Logging in to AWS ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# Check if repository exists, if not create it
# aws ecr describe-repositories --repository-names ${ECR_REPOSITORY} --region ${AWS_REGION} || \
#     aws ecr create-repository --repository-name ${ECR_REPOSITORY} --region ${AWS_REGION}

echo "Pushing image to ECR..."
docker push ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

echo -e "${GREEN}Image successfully pushed to ${ECR_REPOSITORY_URI}:${IMAGE_TAG}${NC}"

# STEP 4: Set up Kubernetes namespace
# echo -e "\n${YELLOW}Step 4: Setting up Kubernetes namespace...${NC}"

# Check if namespace exists, create if it doesn't
# kubectl get namespace ${NAMESPACE} &> /dev/null || \
#     (echo "Creating namespace ${NAMESPACE}..." && \
#     kubectl create namespace ${NAMESPACE})

# echo "Setting kubectl context to namespace: ${NAMESPACE}"
# kubectl config set-context --current --namespace=${NAMESPACE}


# # STEP 5: Deploy PostgreSQL (optional)
# echo -e "\n${YELLOW}Step 5: Deploying PostgreSQL (optional)...${NC}"
# echo "Do you want to deploy PostgreSQL to Kubernetes?"
# echo "Note: If you already have a PostgreSQL instance, you can skip this step"
# read -p "(y/n) " -n 1 -r
# echo
# if [[ $REPLY =~ ^[Yy]$ ]]; then
#     echo "Deploying PostgreSQL..."
#     kubectl apply -f k8s/postgres-deployment.yaml 
    
#     # Wait for PostgreSQL to be ready
#     echo "Waiting for PostgreSQL to be ready..."
#     kubectl rollout status deployment/postgres 
# fi

# STEP 4: Deploy API server
echo -e "\n${YELLOW}Step 4: Deploying API server...${NC}"
echo "Deploying API server..."
kubectl apply -f k8s/api-deployment.yaml 
kubectl rollout restart deployment/code-analyser-be
kubectl apply -f k8s/api-service.yaml 

# Wait for deployments to be ready
echo "Waiting for API server to be ready..."
kubectl rollout status deployment/code-analyser-be 

# STEP 5: Print summary
echo -e "\n${GREEN}Deployment completed successfully!${NC}"
echo -e "\n${YELLOW}API Server Status:${NC}"
kubectl get pods -l app=code-analyser-be 
echo -e "\n${YELLOW}API Service:${NC}"
kubectl get service code-analyser-be 

echo -e "\n${GREEN}You can access the API at:${NC}"
echo "- Inside the cluster: http://code-analyser-be.${NAMESPACE}.svc.cluster.local"

echo -e "\n${YELLOW}To check logs:${NC}"
echo "kubectl logs -f deployment/code-analyser-be "

echo -e "\n${YELLOW}To port-forward the service for local access:${NC}"
echo "kubectl port-forward service/code-analyser-be 8080:80 "
echo "Then access the API at: http://localhost:8080"

# Clean up the local binary
echo -e "\n${YELLOW}Cleaning up...${NC}"
rm -f ./api-server

echo -e "\n${GREEN}Build and deployment process completed successfully!${NC}"
