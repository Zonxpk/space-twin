#!/bin/bash

# Configuration
PROJECT_ID="floorplan-digital-twin"
REGION="asia-northeast1"
VERTEX_LOCATION="global"
SERVICE_NAME="floorplan-backend"
REPO_NAME="floorplan-repo"
IMAGE_TAG="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/$SERVICE_NAME"

# Ensure gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "gcloud CLI is not installed. Please install Google Cloud SDK."
    exit 1
fi

# Create Artifact Registry repo if it doesn't exist
echo "Checking/Creating Artifact Registry repository..."
gcloud artifacts repositories create $REPO_NAME \
    --repository-format=docker \
    --location=$REGION \
    --description="Docker repository for Floorplan Backend" \
    --quiet 2>/dev/null

# Build the container image using Cloud Build
echo "Building container image..."
gcloud builds submit --tag $IMAGE_TAG .

if [ $? -ne 0 ]; then
    echo "Build failed."
    exit 1
fi

# Deploy to Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image $IMAGE_TAG \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
    --set-env-vars GCP_PROJECT_ID=$PROJECT_ID,GCP_LOCATION=$VERTEX_LOCATION,GIN_MODE=release

if [ $? -ne 0 ]; then
    echo "Deployment failed."
    exit 1
fi

echo "Deployment successful!"
