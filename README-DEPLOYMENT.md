# Hack25 Backend Deployment Guide

This guide provides instructions for building, deploying, and running the Hack25 backend using Docker and Kubernetes.

## Prerequisites

- Docker installed
- AWS CLI configured with appropriate permissions
- kubectl configured to connect to your Kubernetes cluster
- Access to AWS ECR (Elastic Container Registry)
- PostgreSQL (either installed locally or deployed to Kubernetes)

## Overview

The Hack25 backend consists of two main components:

1. **API Server**: The main backend service that handles API requests
2. **GitHub Client**: A command-line tool for interacting with GitHub repositories

This guide covers deployment for both components.

## Local Development Setup

### Database Setup

Before running the application, you need to set up the PostgreSQL database:

```bash
# Make the script executable (if not already)
chmod +x scripts/db/setup_database.sh

# Run the script
cd scripts/db
./setup_database.sh
```

This will:
1. Create the database and user
2. Create the tables for users and code analysis
3. Seed initial data (admin user and regular user)
4. Update the .env file with database credentials

### Running Locally

To run the API server locally:

```bash
go run cmd/api/main.go
```

To run the GitHub client locally:

```bash
go run cmd/github-client/main.go -repo=https://github.com/your-username/your-repo
```

## Docker Build and Deployment

### API Server

#### Building and Pushing to ECR

```bash
# Make the script executable (if not already)
chmod +x scripts/build-and-push-api.sh

# Run the script
./scripts/build-and-push-api.sh
```

#### Running Locally with Docker

```bash
docker build -t hack25-api:latest -f Dockerfile.api .
docker run -p 6060:6060 hack25-api:latest
```

### GitHub Client

#### Building and Pushing to ECR

```bash
# Make the script executable (if not already)
chmod +x scripts/build-and-push.sh

# Run the script
./scripts/build-and-push.sh
```

#### Running Locally with Docker

```bash
docker build -t github-client:latest .
docker run -it github-client:latest -repo=https://github.com/your-username/your-repo
```

## Kubernetes Deployment

The application can be deployed to Kubernetes using our deployment scripts.

### Automated Deployment

We provide a script to automate the deployment process:

```bash
# Make the script executable (if not already)
chmod +x scripts/deploy-to-kubernetes.sh

# Run the script
./scripts/deploy-to-kubernetes.sh
```

This script will:
1. Confirm the current Kubernetes context
2. Create a namespace for the application if it doesn't exist
3. Set up necessary secrets (database credentials and API keys)
4. Deploy PostgreSQL (optional)
5. Deploy the API server
6. Wait for deployments to be ready
7. Print summary and usage instructions

### Manual Deployment

If you prefer to deploy manually, follow these steps:

1. Create the namespace:
   ```bash
   kubectl create namespace hack25
   ```

2. Create the necessary secrets:
   ```bash
   # Database credentials
   kubectl create secret generic hack25-db-credentials \
       --from-literal=host=postgres-service \
       --from-literal=port=5432 \
       --from-literal=username=hack25_user \
       --from-literal=password=hack25_password \
       --from-literal=dbname=hack25 \
       -n hack25

   # API secrets
   kubectl create secret generic hack25-api-secrets \
       --from-literal=jwt-secret=your-secret-key-change-in-production \
       --from-literal=openai-api-key=your-openai-api-key \
       --from-literal=gemini-api-key=your-gemini-api-key \
       --from-literal=sonnet-api-key=your-sonnet-api-key \
       -n hack25
   ```

3. Deploy PostgreSQL (optional):
   ```bash
   kubectl apply -f k8s/postgres-deployment.yaml -n hack25
   ```

4. Deploy the API server:
   ```bash
   kubectl apply -f k8s/api-deployment.yaml -n hack25
   kubectl apply -f k8s/api-service.yaml -n hack25
   ```

5. Deploy the GitHub client (if needed):
   ```bash
   kubectl apply -f k8s/deployment.yaml -n hack25
   kubectl apply -f k8s/service.yaml -n hack25
   ```

### Verifying Deployment

Check if the deployments and services are running:

```bash
# Check deployment status
kubectl get deployments -n hack25

# Check pod status
kubectl get pods -n hack25

# Check service status
kubectl get services -n hack25

# Check ingress status
kubectl get ingress -n hack25
```

### Accessing the API

The API server is deployed as a ClusterIP service with an optional Ingress for external access.

1. Using port forwarding:
   ```bash
   kubectl port-forward service/hack25-api 8080:80 -n hack25
   ```
   Then access the API at: http://localhost:8080

2. Using the Ingress:
   If the Ingress is configured, you can access the API at: https://api.hack25.example.com

## Managing API Keys and Secrets

To update API keys or secrets after the initial deployment:

```bash
# Update API secrets
kubectl create secret generic hack25-api-secrets \
    --from-literal=jwt-secret=your-new-secret-key \
    --from-literal=openai-api-key=your-new-openai-api-key \
    --from-literal=gemini-api-key=your-new-gemini-api-key \
    --from-literal=sonnet-api-key=your-new-sonnet-api-key \
    -n hack25 \
    --dry-run=client -o yaml | kubectl apply -f -

# Update DB credentials
kubectl create secret generic hack25-db-credentials \
    --from-literal=host=your-new-db-host \
    --from-literal=port=your-new-db-port \
    --from-literal=username=your-new-db-username \
    --from-literal=password=your-new-db-password \
    --from-literal=dbname=your-new-db-name \
    -n hack25 \
    --dry-run=client -o yaml | kubectl apply -f -
```

After updating secrets, you need to restart the deployment:

```bash
kubectl rollout restart deployment/hack25-api -n hack25
```

## Troubleshooting

### Common Issues

1. **Image Pull Errors**: Ensure your Kubernetes cluster has access to the ECR repository.

2. **Permission Issues**: Verify that your AWS credentials have the necessary permissions for ECR.

3. **Resource Constraints**: If pods are not starting, check if your cluster has enough resources.

4. **Database Connection Issues**: Make sure the database is accessible from the Kubernetes cluster.

### Viewing Logs

```bash
# Get the pod name
kubectl get pods -n hack25

# View logs for API server
kubectl logs -f deployment/hack25-api -n hack25

# View logs for PostgreSQL
kubectl logs -f deployment/postgres -n hack25

# View logs for GitHub client
kubectl logs -f deployment/github-client -n hack25
```

## Cleanup

To remove the deployment and services:

```bash
# Remove API server
kubectl delete -f k8s/api-deployment.yaml -n hack25
kubectl delete -f k8s/api-service.yaml -n hack25

# Remove GitHub client
kubectl delete -f k8s/deployment.yaml -n hack25
kubectl delete -f k8s/service.yaml -n hack25

# Remove PostgreSQL
kubectl delete -f k8s/postgres-deployment.yaml -n hack25

# Remove secrets
kubectl delete secret hack25-db-credentials -n hack25
kubectl delete secret hack25-api-secrets -n hack25

# Remove namespace
kubectl delete namespace hack25
```

To remove the ECR repositories:

```bash
aws ecr delete-repository --repository-name hack25-api --force --region ap-south-1
aws ecr delete-repository --repository-name github-client --force --region ap-south-1
```
