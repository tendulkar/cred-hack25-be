#!/bin/bash
set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Hack25 Backend - Full Deployment Process${NC}"
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

# Set up the database
echo -e "\n${YELLOW}Do you want to set up the local PostgreSQL database?${NC}"
echo "This is required for local development but not for Kubernetes deployment."
read -p "(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Setting up the database...${NC}"
    chmod +x scripts/db/setup_database.sh
    (cd scripts/db && ./setup_database.sh)
fi

# Build and push Docker images
echo -e "\n${YELLOW}Do you want to build and push the Docker images to ECR?${NC}"
read -p "(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Building and pushing API server image...${NC}"
    chmod +x scripts/build-and-push-api.sh
    ./scripts/build-and-push-api.sh
    
    echo -e "\n${YELLOW}Building and pushing GitHub client image...${NC}"
    chmod +x scripts/build-and-push.sh
    ./scripts/build-and-push.sh
fi

# Deploy to Kubernetes
echo -e "\n${YELLOW}Do you want to deploy to Kubernetes?${NC}"
read -p "(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Deploying to Kubernetes...${NC}"
    chmod +x scripts/deploy-to-kubernetes.sh
    ./scripts/deploy-to-kubernetes.sh
fi

echo -e "\n${GREEN}Deployment process completed!${NC}"
echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Update the environment variables in the Kubernetes deployment if needed"
echo "2. Check the deployment status with 'kubectl get pods -n hack25'"
echo "3. Access the API using the methods described in README-DEPLOYMENT.md"
echo -e "\n${GREEN}For more information, refer to README-DEPLOYMENT.md${NC}"
