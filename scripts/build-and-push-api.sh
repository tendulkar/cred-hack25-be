#!/bin/bash
set -e

# Configuration
AWS_REGION="ap-south-1"
AWS_ACCOUNT_ID="109869387501"
ECR_REPOSITORY="code-analyser-be"
IMAGE_TAG="latest"

# Full image name
ECR_REPOSITORY_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}"

echo "Building Docker image for API server..."
docker build -t ${ECR_REPOSITORY}:${IMAGE_TAG} -f Dockerfile.api .

echo "Tagging image for ECR..."
docker tag ${ECR_REPOSITORY}:${IMAGE_TAG} ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

echo "Logging in to AWS ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# Check if repository exists, if not create it
aws ecr describe-repositories --repository-names ${ECR_REPOSITORY} --region ${AWS_REGION} || \
    aws ecr create-repository --repository-name ${ECR_REPOSITORY} --region ${AWS_REGION}

echo "Pushing image to ECR..."
docker push ${ECR_REPOSITORY_URI}:${IMAGE_TAG}

echo "Image successfully pushed to ${ECR_REPOSITORY_URI}:${IMAGE_TAG}"
