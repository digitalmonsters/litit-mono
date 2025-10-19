# Makefile for Azure Container Registry

# Set your Azure Container Registry details
ACR_NAME = lititakscontainer
IMAGE_NAME = music
TAG = v3

# Build and push the Docker image to ACR
prod: login
	docker build --platform linux/amd64 -t $(ACR_NAME).azurecr.io/$(IMAGE_NAME):$(TAG) .
	docker push $(ACR_NAME).azurecr.io/$(IMAGE_NAME):$(TAG)

# Log in to ACR using Azure CLI
login:
	az acr login --name $(ACR_NAME) 